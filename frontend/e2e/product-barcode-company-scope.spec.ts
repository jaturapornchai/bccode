import { expect, test, type Page } from "@playwright/test";
import path from "node:path";

const MAINAPI = "http://127.0.0.1:8888";
const UAT_MP4 = Buffer.from(
  "AAAAIGZ0eXBpc29tAAACAGlzb21pc28yYXZjMW1wNDEAAAMPbW9vdgAAAGxtdmhkAAAAAAAAAAAAAAAAAAAD6AAAA+gAAQAAAQAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAjl0cmFrAAAAXHRraGQAAAADAAAAAAAAAAAAAAABAAAAAAAAA+gAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAAAABAAAAAABAAAAAQAAAAAAAkZWR0cwAAABxlbHN0AAAAAAAAAAEAAAPoAAAAAAABAAAAAAGxbWRpYQAAACBtZGhkAAAAAAAAAAAAAAAAAABAAAAAQABVxAAAAAAALWhkbHIAAAAAAAAAAHZpZGUAAAAAAAAAAAAAAABWaWRlb0hhbmRsZXIAAAABXG1pbmYAAAAUdm1oZAAAAAEAAAAAAAAAAAAAACRkaW5mAAAAHGRyZWYAAAAAAAAAAQAAAAx1cmwgAAAAAQAAARxzdGJsAAAAuHN0c2QAAAAAAAAAAQAAAKhhdmMxAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAABAAEABIAAAASAAAAAAAAAABFUxhdmM2Mi4yOC4xMDAgbGlieDI2NAAAAAAAAAAAAAAAGP//AAAALmF2Y0MBQsAK/+EAFmdCwArZHsBEAAADAAQAAAMACDxImSABAAVoy4PLIAAAABBwYXNwAAAAAQAAAAEAAAAUYnRydAAAAAAAABQwAAAAAAAAABhzdHRzAAAAAAAAAAEAAAABAABAAAAAABxzdHNjAAAAAAAAAAEAAAABAAAAAQAAAAEAAAAUc3RzegAAAAAAAAKGAAAAAQAAABRzdGNvAAAAAAAAAAEAAAM/AAAAYnVkdGEAAABabWV0YQAAAAAAAAAhaGRscgAAAAAAAAAAbWRpcmFwcGwAAAAAAAAAAAAAAAAtaWxzdAAAACWpdG9vAAAAHWRhdGEAAAABAAAAAExhdmY2Mi4xMi4xMDAAAAAIZnJlZQAAAo5tZGF0AAACcAYF//9s3EXpvebZSLeWLNgg2SPu73gyNjQgLSBjb3JlIDE2NSByMzIyMyAwNDgwY2IwIC0gSC4yNjQvTVBFRy00IEFWQyBjb2RlYyAtIENvcHlsZWZ0IDIwMDMtMjAyNSAtIGh0dHA6Ly93d3cudmlkZW9sYW4ub3JnL3gyNjQuaHRtbCAtIG9wdGlvbnM6IGNhYmFjPTAgcmVmPTMgZGVibG9jaz0xOjA6MCBhbmFseXNlPTB4MToweDExMSBtZT1oZXggc3VibWU9NyBwc3k9MSBwc3lfcmQ9MS4wMDowLjAwIG1peGVkX3JlZj0xIG1lX3JhbmdlPTE2IGNocm9tYV9tZT0xIHRyZWxsaXM9MSA4eDhkY3Q9MCBjcW09MCBkZWFkem9uZT0yMSwxMSBmYXN0X3Bza2lwPTEgY2hyb21hX3FwX29mZnNldD0tMiB0aHJlYWRzPTEgbG9va2FoZWFkX3RocmVhZHM9MSBzbGljZWRfdGhyZWFkcz0wIG5yPTAgZGVjaW1hdGU9MSBpbnRlcmxhY2VkPTAgYmx1cmF5X2NvbXBhdD0wIGNvbnN0cmFpbmVkX2ludHJhPTAgYmZyYW1lcz0wIHdlaWdodHA9MCBrZXlpbnQ9MjUwIGtleWludF9taW49MSBzY2VuZWN1dD00MCBpbnRyYV9yZWZyZXNoPTAgcmNfbG9va2FoZWFkPTQwIHJjPWNyZiBtYnRyZWU9MSBjcmY9MjMuMCBxY29tcD0wLjYwIHFwbWluPTAgcXBtYXg9NjkgcXBzdGVwPTQgaXBfcmF0aW89MS40MCBhcT0xOjEuMDAAgAAAAA5liIQFf///D0UAAULfgA==",
  "base64",
);

async function clickVisible(page: Page, pattern: RegExp) {
  const target = page
    .getByRole("button", { name: pattern })
    .filter({ visible: true })
    .first();
  await expect(target).toBeVisible({ timeout: 10_000 });
  await target.click();
}

async function pickFirstUnit(page: Page) {
  const trigger = page
    .locator('button[aria-label="หน่วยนับ"]')
    .filter({ visible: true })
    .first();
  await expect(trigger).toBeVisible();
  await trigger.click();
  const firstUnit = page.locator("li button").filter({ visible: true }).first();
  await expect(firstUnit).toBeVisible({ timeout: 10_000 });
  await firstUnit.click();
}

test("Quick Barcode UAT uses the selected Company and persists only required input", async ({
  page,
  request,
}) => {
  test.setTimeout(180_000);
  const barcode = `UAT${Date.now().toString().slice(-9)}`;
  const boxBarcode = `BOX${Date.now().toString().slice(-9)}`;
  let createdGuid = "";
  let boxBarcodeGuid = "";
  let createdProductGuid = "";
  let sessionToken = "";
  const barcodeDescription = `รายละเอียดบาร์โค้ด ${barcode}`;
  const runtimeErrors: string[] = [];
  const privateVideoReads: string[] = [];
  page.on("console", (message) => {
    if (
      message.type() === "error" &&
      !message.text().includes("GSI_LOGGER") &&
      !message.text().includes("Failed to load resource")
    ) {
      runtimeErrors.push(message.text());
    }
  });
  page.on("pageerror", (error) => runtimeErrors.push(error.message));
  page.on("request", (request) => {
    if (
      request.method() === "GET" &&
      request.url().includes("/products/videos/")
    ) {
      privateVideoReads.push(request.url());
    }
  });
  page.on("response", (response) => {
    if (
      response.status() >= 400 &&
      !response.url().includes("accounts.google.com")
    ) {
      runtimeErrors.push(`${response.status()} ${response.url()}`);
    }
  });

  await page.goto("/", { waitUntil: "domcontentloaded" });
  await clickVisible(page, /เข้าทดสอบระบบ/);
  await clickVisible(page, /test(?!\d)/i);

  const companyCard = page
    .locator("button")
    .filter({ hasText: "บริษัท" })
    .last();
  await expect(companyCard).toBeVisible({ timeout: 10_000 });
  const holdingOnlySession = JSON.parse(
    (await page.evaluate(() => localStorage.getItem("bc_auth"))) ?? "{}",
  ) as { token?: string };
  const companyRequired = await request.get(
    `${MAINAPI}/product/barcode?limit=1`,
    {
      headers: { Authorization: `Bearer ${holdingOnlySession.token}` },
    },
  );
  expect(companyRequired.status()).toBe(409);
  await companyCard.click();
  await clickVisible(page, /สำนักงานใหญ่.*00000/);

  const menuSearch = page
    .locator('input[placeholder*="ค้นหาเมนู"]')
    .filter({ visible: true })
    .first();
  await expect(menuSearch).toBeVisible({ timeout: 10_000 });
  const selectedWorkspace = JSON.parse(
    (await page.evaluate(() => localStorage.getItem("bc_workspace"))) ?? "{}",
  ) as { shop?: { holdingcode?: string }; company?: { code?: string } };
  const productScopeRequest = page.waitForRequest(
    (req) =>
      req.method() === "POST" &&
      req.url().includes("/api/workspace/select-holding"),
  );
  await menuSearch.fill("สินค้า");
  await page
    .getByRole("button", { name: /^สินค้า$/ })
    .first()
    .click();
  const productScopePayload = (await productScopeRequest).postDataJSON() as {
    holdingcode?: string;
    businesscode?: string;
  };
  expect(productScopePayload).toMatchObject({
    holdingcode: selectedWorkspace.shop?.holdingcode,
    businesscode: selectedWorkspace.company?.code,
  });

  await menuSearch.fill("บาร์โค้ด");
  await page
    .getByRole("button", { name: /^บาร์โค้ด$/ })
    .first()
    .click();

  await expect(
    page.getByRole("heading", { name: "บาร์โค้ด" }).first(),
  ).toBeVisible({ timeout: 10_000 });
  await expect(
    page.getByText("an active company is required", { exact: true }),
  ).toHaveCount(0);
  await expect(
    page.getByText("กรุณาเลือกบริษัทก่อนใช้งาน", { exact: true }),
  ).toHaveCount(0);
  const firstBarcodeRow = page
    .locator(".bc-list-row")
    .filter({ visible: true })
    .first();
  await expect(firstBarcodeRow).toBeVisible();
  expect(await firstBarcodeRow.locator(":scope > div:visible").count()).toBe(2);
  expect(
    await firstBarcodeRow.evaluate(
      (row) => row.scrollWidth <= row.clientWidth + 1,
    ),
  ).toBe(true);
  await clickVisible(page, /^เพิ่ม$/);
  await expect(page.getByText("ข้อมูลบาร์โค้ดที่จำเป็น")).toBeVisible();
  await expect(
    page.getByText(/สร้างบาร์โค้ดให้ขาย รับสินค้า และเริ่มงานสต๊อกได้ก่อน/),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: /^(ราคา|ต้นทุน|ยอดคงเหลือ|สต๊อก)$/ }),
  ).toHaveCount(0);
  await expect(page.getByText("200 — ใช้ภายในร้าน")).toBeVisible();
  await expect(page.getByText("885 — GS1 Thailand")).toBeVisible();

  await clickVisible(page, /^บันทึก$/);
  await expect(page.getByText("จำเป็นต้องระบุ").first()).toBeVisible();

  const requiredSection = page
    .getByText("ข้อมูลบาร์โค้ดที่จำเป็น")
    .locator("..")
    .locator("..");
  const inputs = requiredSection.locator('input:not([type="radio"])');
  await expect(inputs.nth(0)).toBeVisible();
  await expect(inputs.nth(1)).toBeVisible();
  await page.getByRole("button", { name: "สร้างบาร์โค้ดอัตโนมัติ" }).click();
  await expect(inputs.nth(0)).toHaveValue(/^200\d{10}$/);
  const generatedBarcode = await inputs.nth(0).inputValue();
  await expect(
    page.getByRole("img", { name: `EAN-13 ${generatedBarcode}` }),
  ).toBeVisible();
  await inputs.nth(0).fill(barcode);
  await expect(inputs.nth(1)).toHaveValue(barcode);

  const primaryName = page
    .locator("label")
    .filter({ hasText: /ภาษาแรก|TH/ })
    .locator("input")
    .first();
  await expect(primaryName).toBeVisible();
  await primaryName.fill(`สินค้า ${barcode}`);
  await pickFirstUnit(page);

  const barcodeMainImages = page
    .getByRole("heading", { name: "รูปหลักของบาร์โค้ด" })
    .locator("xpath=ancestor::section[1]");
  await barcodeMainImages
    .locator('input[type="file"]')
    .setInputFiles(path.resolve("public/flags/th.png"));
  await expect(barcodeMainImages.locator("img")).toBeVisible({
    timeout: 10_000,
  });
  await expect(barcodeMainImages).not.toContainText(/\/(?:goapi\/)?s3\/file\//);
  const barcodeGallery = page
    .getByRole("heading", { name: "รูปเพิ่มเติมของบาร์โค้ด" })
    .locator("xpath=ancestor::section[1]");
  await barcodeGallery
    .locator('input[type="file"]')
    .setInputFiles(path.resolve("public/flags/en.png"));
  await expect(barcodeGallery.locator("img")).toHaveCount(1, {
    timeout: 10_000,
  });
  await expect(barcodeGallery).not.toContainText(/\/(?:goapi\/)?s3\/file\//);
  const barcodeVideos = page
    .getByRole("heading", { name: "วิดีโอ", exact: true })
    .locator("xpath=ancestor::section[1]");
  await expect(barcodeVideos).toContainText("MP4 สูงสุด 500 MB");
  await barcodeVideos.locator('input[type="file"]').setInputFiles({
    buffer: UAT_MP4,
    mimeType: "video/mp4",
    name: `${barcode}.mp4`,
  });
  await expect(barcodeVideos.getByTestId("video-poster")).toBeVisible({
    timeout: 10_000,
  });
  expect(privateVideoReads).toHaveLength(0);
  await expect(
    barcodeVideos.getByRole("button", { name: "โหลดและเล่นวิดีโอ" }),
  ).toHaveCount(1, {
    timeout: 10_000,
  });
  await page
    .getByPlaceholder("รายละเอียดเฉพาะของบาร์โค้ดนี้...")
    .fill(barcodeDescription);

  try {
    await clickVisible(page, /^บันทึก$/);
    await expect(page.getByText("บันทึกบาร์โค้ดแล้ว")).toBeVisible({
      timeout: 10_000,
    });
    await page
      .locator(".bc-list-row")
      .filter({ hasText: barcode, visible: true })
      .first()
      .click();
    const detailSummary = page.getByTestId("barcode-detail-summary");
    const detailFields = page.getByTestId("barcode-detail-fields");
    await expect(detailSummary).toBeVisible();
    await expect(detailSummary).toContainText(barcode);
    await expect(detailFields).toBeVisible();
    await expect(detailFields).toContainText("ราคาขาย 1");
    const barcodeMediaDetail = page
      .getByRole("heading", { name: "สื่อบาร์โค้ดและบรรจุภัณฑ์" })
      .locator("..");
    await expect(barcodeMediaDetail.locator("img")).toHaveCount(3, {
      timeout: 10_000,
    });
    await expect(barcodeMediaDetail.getByTestId("video-poster")).toBeVisible();
    expect(privateVideoReads).toHaveLength(0);
    await expect(barcodeMediaDetail).not.toContainText(
      /\/(?:goapi\/)?s3\/file\//,
    );
    await barcodeMediaDetail
      .getByRole("button", { name: /โหลดและเล่น/ })
      .click();
    await expect(barcodeMediaDetail.locator("video")).toHaveCount(1, {
      timeout: 10_000,
    });
    await expect.poll(() => privateVideoReads.length).toBe(1);
    await expect(
      page.getByText(barcodeDescription, { exact: true }),
    ).toBeVisible();
    await expect(page.getByRole("button", { name: "พิมพ์ฉลาก" })).toBeVisible();
    const summaryBox = await detailSummary.boundingBox();
    const fieldsBox = await detailFields.boundingBox();
    expect(summaryBox).not.toBeNull();
    expect(fieldsBox).not.toBeNull();
    expect(summaryBox!.height).toBeLessThan(160);
    expect(Math.abs(fieldsBox!.y - summaryBox!.y)).toBeLessThan(4);
    expect(summaryBox!.x + summaryBox!.width).toBeLessThanOrEqual(
      fieldsBox!.x + 1,
    );

    await page.setViewportSize({ width: 768, height: 1024 });
    await detailSummary.scrollIntoViewIfNeeded();
    await expect
      .poll(() =>
        page.evaluate(
          () =>
            document.documentElement.scrollWidth <=
            document.documentElement.clientWidth + 1,
        ),
      )
      .toBe(true);

    await page.setViewportSize({ width: 390, height: 844 });
    await detailSummary.scrollIntoViewIfNeeded();
    const mobileSummaryBox = await detailSummary.boundingBox();
    const mobileFieldsBox = await detailFields.boundingBox();
    expect(mobileSummaryBox).not.toBeNull();
    expect(mobileFieldsBox).not.toBeNull();
    expect(mobileFieldsBox!.y).toBeGreaterThanOrEqual(
      mobileSummaryBox!.y + mobileSummaryBox!.height + 8,
    );
    expect(
      Math.abs(mobileSummaryBox!.width - mobileFieldsBox!.width),
    ).toBeLessThan(4);
    await page.setViewportSize({ width: 1280, height: 720 });

    const session = JSON.parse(
      (await page.evaluate(() => localStorage.getItem("bc_auth"))) ?? "{}",
    ) as {
      token?: string;
    };
    const workspace = JSON.parse(
      (await page.evaluate(() => localStorage.getItem("bc_workspace"))) ?? "{}",
    ) as {
      company?: { code?: string };
    };
    expect(session.token).toBeTruthy();
    expect(workspace.company?.code).toBeTruthy();
    sessionToken = session.token ?? "";

    const missingUnitProductResponse = await request.post(
      `${MAINAPI}/product`,
      {
        data: {
          code: `NO-UNIT-${Date.now()}`,
          names: [{ code: "th", name: "สินค้าที่ไม่มีหน่วยมาตรฐาน" }],
        },
        headers: { Authorization: `Bearer ${sessionToken}` },
      },
    );
    expect(missingUnitProductResponse.status()).toBe(400);

    const invalidEanResponse = await request.post(
      `${MAINAPI}/product/barcode`,
      {
        data: {
          barcode: "4006381333932",
          itemcode: `INVALID-EAN-${Date.now()}`,
          names: [{ code: "th", name: "ต้องไม่ถูกสร้าง" }],
          itemunitcode: "PCS",
          dividevalue: 1,
          standvalue: 1,
          ismainbarcode: true,
        },
        headers: { Authorization: `Bearer ${sessionToken}` },
      },
    );
    expect(invalidEanResponse.status()).toBe(400);
    expect(await invalidEanResponse.text()).toContain(
      "invalid EAN-13 check digit",
    );

    const response = await request.get(
      `${MAINAPI}/product/barcode?q=${encodeURIComponent(barcode)}&limit=5`,
      {
        headers: { Authorization: `Bearer ${session.token}` },
      },
    );
    expect(response.status()).toBe(200);
    const payload = (await response.json()) as {
      data?: Array<{ guidfixed?: string; barcode?: string; itemcode?: string }>;
    };
    const created = payload.data?.find((item) => item.barcode === barcode);
    expect(created).toBeTruthy();
    createdGuid = created?.guidfixed ?? "";
    expect(created?.itemcode).toBe(barcode);

    const persistedBarcodeResponse = await request.get(
      `${MAINAPI}/product/barcode/${createdGuid}`,
      {
        headers: { Authorization: `Bearer ${session.token}` },
      },
    );
    expect(persistedBarcodeResponse.status()).toBe(200);
    const persistedBarcodePayload = (await persistedBarcodeResponse.json()) as {
      data?: {
        imageuri?: string;
        images?: Array<{ uri?: string }>;
        videos?: Array<{ uri?: string; posteruri?: string }>;
        description?: string;
      };
    };
    expect(persistedBarcodePayload.data?.imageuri).toBeTruthy();
    expect(persistedBarcodePayload.data?.images).toHaveLength(1);
    expect(persistedBarcodePayload.data?.videos).toHaveLength(1);
    expect(persistedBarcodePayload.data?.videos?.[0]?.posteruri).toBeTruthy();
    expect(persistedBarcodePayload.data?.description).toBe(barcodeDescription);
    const barcodeVideoURI =
      persistedBarcodePayload.data?.videos?.[0]?.uri ?? "";
    const barcodePosterURI =
      persistedBarcodePayload.data?.videos?.[0]?.posteruri ?? "";
    expect(barcodeVideoURI).toContain("/companies/");
    expect(barcodePosterURI).toContain("/companies/");
    const barcodeVideoResponse = await request.get(
      `${MAINAPI}${barcodeVideoURI}`,
      {
        headers: { Authorization: `Bearer ${session.token}` },
      },
    );
    expect(barcodeVideoResponse.status()).toBe(200);
    expect(barcodeVideoResponse.headers()["content-type"]).toContain(
      "video/mp4",
    );
    const barcodePosterResponse = await request.get(
      `${MAINAPI}${barcodePosterURI}`,
      {
        headers: { Authorization: `Bearer ${session.token}` },
      },
    );
    expect(barcodePosterResponse.status()).toBe(200);
    expect(barcodePosterResponse.headers()["content-type"]).toContain(
      "image/jpeg",
    );

    const tamperedScopeResponse = await request.get(
      `${MAINAPI}/product/barcode?q=${encodeURIComponent(barcode)}&businesscode=UNTRUSTED&limit=5`,
      { headers: { Authorization: `Bearer ${session.token}` } },
    );
    expect(tamperedScopeResponse.status()).toBe(200);
    const tamperedScopePayload = (await tamperedScopeResponse.json()) as {
      data?: Array<{ barcode?: string }>;
    };
    expect(
      tamperedScopePayload.data?.some((item) => item.barcode === barcode),
    ).toBe(true);

    const createdBarcodeRow = page
      .locator(".bc-list-row")
      .filter({ hasText: barcode, visible: true })
      .first();
    await expect(createdBarcodeRow).toBeVisible();
    await createdBarcodeRow.click();
    await expect(page.getByTestId("barcode-detail-summary")).toContainText(
      barcode,
    );

    const productDetailResponse = page.waitForResponse((response) => {
      const url = new URL(response.url());
      return url.pathname === `/api/product/${encodeURIComponent(barcode)}`;
    });
    await clickVisible(page, /^ไปเติมรายละเอียดสินค้า$/);
    const openedProductResponse = await productDetailResponse;
    expect(openedProductResponse.status()).toBe(200);
    const openedProductPayload = (await openedProductResponse.json()) as {
      data?: { guidfixed?: string };
    };
    createdProductGuid = openedProductPayload.data?.guidfixed ?? "";
    const productDetailCard = page.getByTestId("product-detail-card");
    await expect(productDetailCard).toBeVisible({ timeout: 10_000 });
    await expect(productDetailCard).toContainText(barcode);
    const selectedProductRow = page
      .locator('[data-testid="product-row"][aria-pressed="true"]')
      .filter({ hasText: barcode, visible: true });
    await expect(selectedProductRow).toHaveCount(1);
    const barcodeMediaOnProduct = page
      .getByRole("heading", { name: "สื่อจากบาร์โค้ด" })
      .locator("..");
    await expect(barcodeMediaOnProduct.locator("img")).toHaveCount(3, {
      timeout: 10_000,
    });
    await expect(
      barcodeMediaOnProduct.getByTestId("video-poster"),
    ).toBeVisible();
    await expect(barcodeMediaOnProduct).not.toContainText(
      /\/(?:goapi\/)?s3\/file\//,
    );
    await expect(
      barcodeMediaOnProduct.getByRole("button", { name: /โหลดและเล่น/ }),
    ).toHaveCount(1);

    await clickVisible(page, /^แก้ไข$/);
    await clickVisible(page, /^หน่วยนับและบาร์โค้ด$/);
    const baseUnitSection = page
      .getByRole("heading", { name: "หน่วยนับและอัตราส่วน (Unit conversion)" })
      .locator("xpath=ancestor::section[1]");
    await expect(baseUnitSection.locator("input").first()).not.toHaveValue("");
    const baseRatios = baseUnitSection.locator('input[inputmode="decimal"]');
    await expect(baseRatios).toHaveCount(2);
    await expect(baseRatios.nth(0)).toHaveValue("1");
    await expect(baseRatios.nth(1)).toHaveValue("1");
    await expect(baseRatios.nth(0)).toBeDisabled();
    await expect(baseRatios.nth(1)).toBeDisabled();
    await expect(baseUnitSection.locator("svg.lucide-x")).toHaveCount(0);
    await expect(baseUnitSection.getByTestId("matched-barcodes")).toContainText(
      barcode,
    );
    await expect(page.getByText(/เปิดใช้งานบาร์โค้ดหลายหน่วยนับ/)).toHaveCount(
      0,
    );
    await expect(page.getByPlaceholder("ระบุบาร์โค้ดหน่วยย่อย")).toHaveCount(0);
    await clickVisible(page, /^เพิ่มหน่วยนับ$/);
    const newUnitRow = page.getByTestId("product-unit-conversion").last();
    await expect(newUnitRow).toBeVisible();
    await newUnitRow
      .getByRole("button", {
        name: "หน่วยนับ (Unit Code)",
        exact: true,
      })
      .click();
    const unitPicker = page.getByRole("dialog", { name: /ค้นหา\s+หน่วย/ });
    await expect(unitPicker).toBeVisible();
    await unitPicker.getByPlaceholder("ค้นหา").fill("BOX");
    const boxUnitOption = unitPicker.getByRole("button", { name: /\bBOX\b/ });
    await expect(boxUnitOption).toHaveCount(1);
    await boxUnitOption.click();
    await expect(newUnitRow.locator("input[readonly]").first()).toHaveValue(
      /BOX/,
    );
    const unitRatioInputs = newUnitRow.locator('input[inputmode="decimal"]');
    await expect(unitRatioInputs).toHaveCount(2);
    await unitRatioInputs.nth(0).fill("1");
    await unitRatioInputs.nth(1).fill("12");
    await unitRatioInputs.nth(1).press("Tab");
    await expect(unitRatioInputs.nth(1)).toHaveValue("12");
    await expect(newUnitRow.getByTestId("matched-barcodes")).toContainText(
      "ยังไม่มีบาร์โค้ดที่ใช้หน่วยนี้",
    );
    await clickVisible(page, /^ภาพและสี$/);
    await page.getByRole("radio", { name: "แสดงรูปภาพหลัก" }).check();
    const productMainImages = page
      .getByRole("heading", { name: "รูปภาพหลักของสินค้า" })
      .locator("xpath=ancestor::section[1]");
    await productMainImages
      .locator('input[type="file"]')
      .setInputFiles(path.resolve("public/line_logo.png"));
    await expect(productMainImages.locator("img")).toBeVisible({
      timeout: 10_000,
    });
    await expect(productMainImages).not.toContainText(
      /\/(?:goapi\/)?s3\/file\//,
    );
    const productVideos = page
      .getByRole("heading", { name: "วิดีโอ", exact: true })
      .locator("xpath=ancestor::section[1]");
    await productVideos.locator('input[type="file"]').setInputFiles({
      buffer: UAT_MP4,
      mimeType: "video/mp4",
      name: `${barcode}-product.mp4`,
    });
    await expect(productVideos.getByTestId("video-poster")).toBeVisible({
      timeout: 10_000,
    });
    await expect(
      productVideos.getByRole("button", { name: "โหลดและเล่นวิดีโอ" }),
    ).toHaveCount(1, {
      timeout: 10_000,
    });
    const productUnitUpdate = page.waitForResponse((response) => {
      const url = new URL(response.url());
      return (
        response.request().method() === "PUT" &&
        url.pathname ===
          `/api/product/${encodeURIComponent(createdProductGuid)}`
      );
    });
    await clickVisible(page, /^บันทึก$/);
    expect((await productUnitUpdate).status()).toBe(200);
    await expect(productDetailCard).toBeVisible({ timeout: 10_000 });
    const productMediaDetail = page
      .getByRole("heading", { name: "สื่อสินค้า" })
      .locator("..");
    const barcodeMediaDetailOnProduct = page
      .getByRole("heading", { name: "สื่อจากบาร์โค้ด" })
      .locator("..");
    await expect(productMediaDetail.locator("img")).toHaveCount(2);
    await expect(productMediaDetail.getByTestId("video-poster")).toBeVisible();
    await expect(productMediaDetail).not.toContainText(
      /\/(?:goapi\/)?s3\/file\//,
    );
    await expect(
      productMediaDetail.getByRole("button", { name: /โหลดและเล่น/ }),
    ).toHaveCount(1);
    await expect(barcodeMediaDetailOnProduct.locator("img")).toHaveCount(3);
    await expect(
      barcodeMediaDetailOnProduct.getByTestId("video-poster"),
    ).toBeVisible();
    await expect(barcodeMediaDetailOnProduct).not.toContainText(
      /\/(?:goapi\/)?s3\/file\//,
    );
    await expect(
      barcodeMediaDetailOnProduct.getByRole("button", { name: /โหลดและเล่น/ }),
    ).toHaveCount(1);

    const persistedProductResponse = await request.get(
      `${MAINAPI}/product/${encodeURIComponent(barcode)}`,
      {
        headers: { Authorization: `Bearer ${session.token}` },
      },
    );
    expect(persistedProductResponse.status()).toBe(200);
    const persistedProductPayload = (await persistedProductResponse.json()) as {
      data?: Record<string, unknown> & {
        guidfixed?: string;
        code?: string;
        imageuri?: string;
        unitcode?: string;
        unitnames?: Array<{ code?: string; name?: string }>;
        unitconversions?: Array<{
          unitcode?: string;
          dividevalue?: number;
          standvalue?: number;
        }>;
        dividevalue?: number;
        standvalue?: number;
        videos?: Array<{ uri?: string; posteruri?: string }>;
        barcodes?: Array<{
          barcode?: string;
          imageuri?: string;
          images?: Array<{ uri?: string }>;
          videos?: Array<{ uri?: string; posteruri?: string }>;
        }>;
      };
    };
    createdProductGuid ||= persistedProductPayload.data?.guidfixed ?? "";
    const linkedBarcode = persistedProductPayload.data?.barcodes?.find(
      (item) => item.barcode === barcode,
    );
    expect(persistedProductPayload.data?.unitcode).toBeTruthy();
    expect(persistedProductPayload.data?.dividevalue).toBe(1);
    expect(persistedProductPayload.data?.standvalue).toBe(1);
    expect(persistedProductPayload.data?.unitconversions).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          unitcode: "BOX",
          dividevalue: 1,
          standvalue: 12,
        }),
      ]),
    );
    expect(persistedProductPayload.data?.imageuri).toBeTruthy();
    expect(persistedProductPayload.data?.videos).toHaveLength(1);
    expect(persistedProductPayload.data?.videos?.[0]?.posteruri).toBeTruthy();
    expect(linkedBarcode?.imageuri).toBeTruthy();
    expect(linkedBarcode?.images).toHaveLength(1);
    expect(linkedBarcode?.videos).toHaveLength(1);
    expect(linkedBarcode?.videos?.[0]?.posteruri).toBe(
      persistedBarcodePayload.data?.videos?.[0]?.posteruri,
    );
    expect(persistedProductPayload.data?.videos?.[0]?.uri).not.toBe(
      barcodeVideoURI,
    );
    expect(linkedBarcode?.videos?.[0]?.uri).toBe(barcodeVideoURI);

    const createBoxBarcodeResponse = await request.post(
      `${MAINAPI}/product/barcode`,
      {
        data: {
          barcode: boxBarcode,
          itemcode: barcode,
          names: [{ code: "th", name: `กล่อง ${barcode}` }],
          itemunitcode: "BOX",
          itemunitnames: [{ code: "th", name: "กล่อง" }],
          dividevalue: 99,
          standvalue: 99,
          ismainbarcode: false,
        },
        headers: { Authorization: `Bearer ${sessionToken}` },
      },
    );
    expect(createBoxBarcodeResponse.status()).toBe(201);
    const boxBarcodeListResponse = await request.get(
      `${MAINAPI}/product/barcode?q=${encodeURIComponent(boxBarcode)}&limit=5`,
      { headers: { Authorization: `Bearer ${sessionToken}` } },
    );
    const boxBarcodeList = (await boxBarcodeListResponse.json()) as {
      data?: Array<{ guidfixed?: string; barcode?: string }>;
    };
    boxBarcodeGuid =
      boxBarcodeList.data?.find((item) => item.barcode === boxBarcode)
        ?.guidfixed ?? "";
    expect(boxBarcodeGuid).toBeTruthy();

    const persistedBoxBarcode = await request.get(
      `${MAINAPI}/product/barcode/${encodeURIComponent(boxBarcodeGuid)}`,
      { headers: { Authorization: `Bearer ${sessionToken}` } },
    );
    const persistedBoxPayload = (await persistedBoxBarcode.json()) as {
      data?: { dividevalue?: number; standvalue?: number };
    };
    expect(persistedBoxPayload.data?.dividevalue).toBe(1);
    expect(persistedBoxPayload.data?.standvalue).toBe(12);

    const changeBoxRatioResponse = await request.put(
      `${MAINAPI}/product/${encodeURIComponent(createdProductGuid)}`,
      {
        data: {
          ...persistedProductPayload.data,
          unitconversions: [
            {
              unitcode: "BOX",
              unitnames: [{ code: "th", name: "กล่อง" }],
              dividevalue: 1,
              standvalue: 24,
            },
          ],
        },
        headers: { Authorization: `Bearer ${sessionToken}` },
      },
    );
    expect(changeBoxRatioResponse.status()).toBe(200);
    const syncedBoxBarcode = await request.get(
      `${MAINAPI}/product/barcode/${encodeURIComponent(boxBarcodeGuid)}`,
      { headers: { Authorization: `Bearer ${sessionToken}` } },
    );
    const syncedBoxPayload = (await syncedBoxBarcode.json()) as {
      data?: { dividevalue?: number; standvalue?: number };
    };
    expect(syncedBoxPayload.data).toMatchObject({
      dividevalue: 1,
      standvalue: 24,
    });

    const restoreBoxRatioResponse = await request.put(
      `${MAINAPI}/product/${encodeURIComponent(createdProductGuid)}`,
      {
        data: persistedProductPayload.data,
        headers: { Authorization: `Bearer ${sessionToken}` },
      },
    );
    expect(restoreBoxRatioResponse.status()).toBe(200);
    const restoredBoxBarcode = await request.get(
      `${MAINAPI}/product/barcode/${encodeURIComponent(boxBarcodeGuid)}`,
      { headers: { Authorization: `Bearer ${sessionToken}` } },
    );
    const restoredBoxPayload = (await restoredBoxBarcode.json()) as {
      data?: { dividevalue?: number; standvalue?: number };
    };
    expect(restoredBoxPayload.data).toMatchObject({
      dividevalue: 1,
      standvalue: 12,
    });

    const unknownUnitBarcode = await request.post(
      `${MAINAPI}/product/barcode`,
      {
        data: {
          barcode: `CASE${Date.now().toString().slice(-8)}`,
          itemcode: barcode,
          names: [{ code: "th", name: "หน่วยที่ยังไม่กำหนด" }],
          itemunitcode: "CASE",
          itemunitnames: [{ code: "th", name: "ลัง" }],
        },
        headers: { Authorization: `Bearer ${sessionToken}` },
      },
    );
    expect(unknownUnitBarcode.status()).toBe(400);

    const removeUsedUnitResponse = await request.put(
      `${MAINAPI}/product/${encodeURIComponent(createdProductGuid)}`,
      {
        data: { ...persistedProductPayload.data, unitconversions: [] },
        headers: { Authorization: `Bearer ${sessionToken}` },
      },
    );
    expect(removeUsedUnitResponse.status()).toBe(400);

    const refreshedProduct = page.waitForResponse((response) => {
      const url = new URL(response.url());
      return url.pathname === `/api/product/${encodeURIComponent(barcode)}`;
    });
    await clickVisible(page, /^รีเฟรช$/);
    expect((await refreshedProduct).status()).toBe(200);
    await clickVisible(page, /^แก้ไข$/);
    await clickVisible(page, /^หน่วยนับและบาร์โค้ด$/);
    const boxUnitRow = page
      .getByTestId("product-unit-conversion")
      .filter({ hasText: "BOX" });
    await expect(boxUnitRow).toHaveCount(1);
    await expect(boxUnitRow.getByTestId("matched-barcodes")).toContainText(
      boxBarcode,
    );
    await expect(page.getByPlaceholder("ระบุบาร์โค้ดหน่วยย่อย")).toHaveCount(0);
    await clickVisible(page, /^ยกเลิก$/);

    await menuSearch.fill("บาร์โค้ด");
    await page
      .getByRole("button", { name: /^บาร์โค้ด$/ })
      .first()
      .click();
    await expect(
      page.getByRole("heading", { name: "บาร์โค้ด" }).first(),
    ).toBeVisible();
    await page
      .locator(".bc-list-row")
      .filter({ hasText: barcode, visible: true })
      .first()
      .click();
    await clickVisible(page, /^แก้ไข$/);
    const editInputs = page
      .getByText("ข้อมูลบาร์โค้ดที่จำเป็น")
      .locator("..")
      .locator("..")
      .locator('input:not([type="radio"])');
    await expect(editInputs.nth(0)).toBeDisabled();
    await expect(editInputs.nth(1)).toBeDisabled();
    await expect(
      page
        .locator("label")
        .filter({ hasText: /ภาษาแรก|TH/ })
        .locator("input")
        .first(),
    ).toHaveValue(`สินค้า ${barcode}`);
    const ratios = requiredSection.locator('input[inputmode="decimal"]');
    await expect(ratios.nth(0)).toHaveValue("1");
    await expect(ratios.nth(1)).toHaveValue("1");
    await expect(
      page.getByPlaceholder("รายละเอียดเฉพาะของบาร์โค้ดนี้..."),
    ).toHaveValue(barcodeDescription);
    const retainedBarcodeVideos = page
      .getByRole("heading", { name: "วิดีโอ", exact: true })
      .locator("xpath=ancestor::section[1]");
    await expect(
      retainedBarcodeVideos.getByTestId("video-poster"),
    ).toBeVisible();
    await expect(
      retainedBarcodeVideos.getByRole("button", { name: "โหลดและเล่นวิดีโอ" }),
    ).toHaveCount(1);
  } finally {
    if (boxBarcodeGuid && sessionToken) {
      await request.delete(`${MAINAPI}/product/barcode/${boxBarcodeGuid}`, {
        headers: { Authorization: `Bearer ${sessionToken}` },
      });
    }
    if (createdGuid && sessionToken) {
      await request.delete(`${MAINAPI}/product/barcode/${createdGuid}`, {
        headers: { Authorization: `Bearer ${sessionToken}` },
      });
    }
    if (createdProductGuid && sessionToken) {
      await request.delete(`${MAINAPI}/product/${createdProductGuid}`, {
        headers: { Authorization: `Bearer ${sessionToken}` },
      });
    }
  }

  expect(runtimeErrors).toEqual([]);
});

test("Quick Barcode pager reaches records after the first 80 and resets on search", async ({
  page,
}) => {
  await page.goto("/", { waitUntil: "domcontentloaded" });
  await clickVisible(page, /เข้าทดสอบระบบ/);
  await clickVisible(page, /test(?!\d)/i);
  await page.locator("button").filter({ hasText: "บริษัท" }).last().click();
  await clickVisible(page, /สำนักงานใหญ่.*00000/);

  const requests: Array<{ keyword?: string; limit?: number; offset?: number }> =
    [];
  const makeRow = (ordinal: number) => {
    const code = `PAGE${String(ordinal).padStart(3, "0")}`;
    return {
      guidfixed: "",
      barcode: code,
      itemcode: code,
      names: [{ code: "th", name: `สินค้า ${code}` }],
      itemunitcode: "PCS",
      itemunitnames: [{ code: "th", name: "ชิ้น" }],
      prices: [{ keynumber: 1, price: "10.00" }],
      dividevalue: 1,
      standvalue: 1,
      ismainbarcode: true,
    };
  };

  await page.route("**/api/product-barcode/list", async (route) => {
    const body = route.request().postDataJSON() as {
      keyword?: string;
      limit?: number;
      offset?: number;
    };
    requests.push(body);
    const offset = body.offset ?? 0;
    const data = body.keyword
      ? [makeRow(81)]
      : Array.from({ length: offset === 0 ? 80 : 1 }, (_, index) =>
          makeRow(offset + index + 1),
        );
    await route.fulfill({
      body: JSON.stringify({
        success: true,
        data,
        total: body.keyword ? 1 : 81,
      }),
      contentType: "application/json",
      status: 200,
    });
  });

  const menuSearch = page
    .locator('input[placeholder*="ค้นหาเมนู"]')
    .filter({ visible: true })
    .first();
  await menuSearch.fill("บาร์โค้ด");
  await page
    .getByRole("button", { name: /^บาร์โค้ด$/ })
    .first()
    .click();

  await expect(page.getByTestId("barcode-pagination")).toContainText(
    "หน้า 1 / 2",
  );
  await page.getByTestId("barcode-page-next").click();
  await expect(
    page.locator(".bc-list-row").filter({ hasText: "PAGE081", visible: true }),
  ).toHaveCount(1);
  await expect(page.getByTestId("barcode-pagination")).toContainText(
    "หน้า 2 / 2",
  );
  expect(
    requests.some((request) => request.limit === 80 && request.offset === 80),
  ).toBe(true);

  const searchResponse = page.waitForResponse((response) =>
    response.url().includes("/api/product-barcode/list"),
  );
  await page
    .locator('input[placeholder*="ค้นหา บาร์โค้ด"]')
    .filter({ visible: true })
    .fill("PAGE081");
  await searchResponse;
  await expect(
    page.locator(".bc-list-row").filter({ hasText: "PAGE081", visible: true }),
  ).toHaveCount(1);
  await expect(page.getByTestId("barcode-pagination")).toHaveCount(0);
  expect(requests.at(-1)).toMatchObject({ keyword: "PAGE081", offset: 0 });
});
