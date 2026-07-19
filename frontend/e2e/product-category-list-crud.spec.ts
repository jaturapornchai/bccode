import { test, expect, type Page } from "@playwright/test";

/**
 * Regression coverage for fixes made 2026-07-01 via a uat-crud-mongo sweep across the
 * "จัดกลุ่มสินค้า" menu group:
 *
 * 1. `productcategorylist` ("สินค้าในหมวด") field `codelist` was a raw JSON textarea
 *    (`jsonField`, no structured UI) — violates the project's "no raw {}/[] in user-facing
 *    screens" rule. Fixed by switching it to the existing `stringListField` (chip/tag editor,
 *    same component already used for color/size "aliases"). See system-setting-screens.ts. Live
 *    UI verification of the rendered chip editor is blocked by a SEPARATE, still-open finding
 *    (see below) — a category created on the sibling `productcategorygroupselectscreen` does not
 *    show up on `productcategorylist` even after a full page reload + 30s of polling, for the
 *    same `group-number`/basePath. Root cause not yet isolated (checked: the Next.js API proxy
 *    uses `cache: "no-store"`, so it is not an obvious HTTP cache). Flagging, not silently
 *    working around it. This test instead only covers what's reliably verifiable: the
 *    readOnly-aware empty-state text (below), plus tsc + source confirm that `stringListField`
 *    has no raw-JSON render path anywhere in the codebase.
 * 2. `product-category-tree-view.tsx`'s empty-state hint text unconditionally told the user to
 *    click "เพิ่มหมวดหลัก" even when the tree is rendered `readOnly` (as it is on
 *    `productcategorylist`, where that button never renders) — a dead UI reference that would
 *    confuse a real user. Fixed by branching the hint text on the `readOnly` prop.
 */

async function clickButtonByText(page: Page, textPattern: RegExp) {
  const clicked = await page.evaluate((pattern) => {
    const re = new RegExp(pattern);
    const el = [...document.querySelectorAll<HTMLElement>("button,a,[role=button]")].find(
      (e) => re.test((e.textContent ?? "").trim()) && e.offsetParent !== null,
    );
    if (el) {
      el.click();
      return true;
    }
    return false;
  }, textPattern.source);
  if (!clicked) throw new Error(`No clickable element found matching ${textPattern}`);
}

async function clickContainingText(page: Page, text: string) {
  return page.evaluate((t) => {
    const el = [...document.querySelectorAll<HTMLElement>("button,a,[role=button]")].find(
      (e) => (e.textContent ?? "").includes(t) && e.offsetParent !== null,
    );
    if (el) {
      el.click();
      return true;
    }
    return false;
  }, text);
}

async function bodyHasText(page: Page, text: string): Promise<boolean> {
  return page.evaluate((t) => document.body.innerText.includes(t), text);
}

async function loginAndSearchMenu(page: Page, menuQuery: string) {
  await page.goto("/", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /เข้าทดสอบระบบ/);
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /Test/);
  await page.waitForTimeout(2500);
  await page.evaluate(() => {
    const btn = [...document.querySelectorAll<HTMLElement>("button")].find(
      (b) => (b.textContent ?? "").includes("บริษัท") && (b.textContent ?? "").trim().length > 15,
    );
    if (!btn) throw new Error("No company card button found");
    btn.click();
  });
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /สำนักงานใหญ่|headquarters/i);
  await page.waitForTimeout(2500);
  await page.evaluate((q) => {
    const input = document.querySelector<HTMLInputElement>('input[placeholder*="ค้นหาเมนู"]');
    if (!input) throw new Error("Menu search input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, q);
    input.dispatchEvent(new Event("input", { bubbles: true }));
  }, menuQuery);
  await page.waitForTimeout(1000);
}

test("productcategorylist — readOnly empty-state text no longer references a nonexistent button", async ({
  page,
}) => {
  // Group 20 is the last of the 20 fixed usage/device/channel groups (core-rules.md "Product
  // category group semantics") -- least likely to already hold categories from other manual/E2E
  // testing, so its empty state can be checked directly without creating any data.
  await loginAndSearchMenu(page, "สินค้าในหมวด");
  await clickButtonByText(page, /^สินค้าในหมวด$/);
  await page.waitForTimeout(2000);
  await clickContainingText(page, "20ยังไม่มีหมวดสินค้า");
  await page.waitForTimeout(1500);

  expect(await bodyHasText(page, "กรุณาไปสร้างหมวดสินค้าที่หน้าจอ")).toBe(true);
  expect(await bodyHasText(page, "กดปุ่ม 'เพิ่มหมวดหลัก' ด้านบน")).toBe(false);

  const hasAddRootButton = await page.evaluate(() =>
    [...document.querySelectorAll("button,a")].some(
      (e) => (e.textContent ?? "").includes("เพิ่มหมวดหลัก") && (e as HTMLElement).offsetParent !== null,
    ),
  );
  expect(hasAddRootButton).toBe(false);
});

test("productcategorygroupselectscreen — its OWN empty-state still references the real, working button", async ({
  page,
}) => {
  // Sibling screen keeps the original copy since its "เพิ่มหมวดหลัก" button genuinely renders here
  // (not readOnly) -- guards against the fix accidentally hiding the button/text on the CRUD side.
  await loginAndSearchMenu(page, "จัดหมวดสินค้า");
  await clickButtonByText(page, /^จัดหมวดสินค้า$/);
  await page.waitForTimeout(2000);
  await clickContainingText(page, "20ยังไม่มีหมวดสินค้า");
  await page.waitForTimeout(1500);

  expect(await bodyHasText(page, "กดปุ่ม 'เพิ่มหมวดหลัก' ด้านบน")).toBe(true);
  const hasAddRootButton = await page.evaluate(() =>
    [...document.querySelectorAll("button,a")].some(
      (e) => (e.textContent ?? "").includes("เพิ่มหมวดหลัก") && (e as HTMLElement).offsetParent !== null,
    ),
  );
  expect(hasAddRootButton).toBe(true);
});

const MAINAPI = "http://localhost:8888";

/** Clicks a category-tree row whose OWN text includes `text` (tree rows are
 *  `data-category-row-guid` divs with an onClick on the row itself, not a <button> -- mirrors
 *  `product-barcode-sample-data.spec.ts`'s `clickTreeRowByText`). Prefers the shortest matching
 *  row's own text so an ancestor row's aggregated textContent doesn't win over its child row. */
async function clickTreeRowByText(page: Page, text: string) {
  const deadline = Date.now() + 8000;
  for (;;) {
    const clicked = await page.evaluate((t) => {
      const rows = [...document.querySelectorAll<HTMLElement>("[data-category-row-guid]")].filter(
        (e) => (e.textContent ?? "").includes(t) && e.offsetParent !== null,
      );
      if (rows.length === 0) return false;
      rows.sort((a, b) => (a.textContent ?? "").length - (b.textContent ?? "").length);
      rows[0].click();
      return true;
    }, text);
    if (clicked) return;
    if (Date.now() > deadline) throw new Error(`No tree row found containing text "${text}"`);
    await page.waitForTimeout(300);
  }
}

/** Clicks the group-number card on the "เลือกชุดหมวดสินค้า" grid whose text starts with `num`
 *  (the card's leading number badge). Matches on the number prefix, not the "ยังไม่มีหมวดสินค้า"
 *  empty-state label, so the test stays idempotent across reruns that leave real data in the group. */
async function clickGroupNumber(page: Page, num: number) {
  const deadline = Date.now() + 8000;
  for (;;) {
    const clicked = await page.evaluate((n) => {
      const selector = document.querySelector<HTMLElement>(
        '[data-testid="product-category-group-selector"]',
      );
      const el = [...(selector?.querySelectorAll<HTMLElement>("button") ?? [])].find(
        (b) => b.offsetParent !== null && (b.textContent ?? "").trim().startsWith(String(n)),
      );
      if (el) {
        el.click();
        return true;
      }
      return false;
    }, num);
    if (clicked) return;
    if (Date.now() > deadline) throw new Error(`No group-number card found starting with "${num}"`);
    await page.waitForTimeout(300);
  }
}

test("productcategorygroupselectscreen — tablet workbench selects read-only detail before edit", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1024, height: 768 });
  await loginAndSearchMenu(page, "จัดหมวดสินค้า");
  await clickButtonByText(page, /^จัดหมวดสินค้า$/);
  await page.waitForTimeout(1500);
  await clickGroupNumber(page, 1);

  const treePane = page.getByTestId("product-category-tree-pane");
  const detailPane = page.getByTestId("product-category-detail-pane");
  const firstRow = treePane.locator("[data-category-row-guid]").first();
  await expect(firstRow).toBeVisible();
  await firstRow.click();

  await expect(detailPane.getByRole("button", { name: "แก้ไข" })).toBeVisible();
  await expect(detailPane.getByText("ลำดับกลุ่ม", { exact: true })).toHaveCount(0);
  await expect(detailPane.getByText("หมวดแม่", { exact: true })).toHaveCount(0);
  await expect(treePane.getByRole("button", { name: "แก้ไข" })).toHaveCount(0);
  await expect(treePane.getByRole("button", { name: "ลบ" })).toHaveCount(0);

  const [treeBox, detailBox] = await Promise.all([
    treePane.boundingBox(),
    detailPane.boundingBox(),
  ]);
  expect(treeBox).not.toBeNull();
  expect(detailBox).not.toBeNull();
  expect(Math.abs(treeBox!.y - detailBox!.y)).toBeLessThan(12);
  expect(treeBox!.x + treeBox!.width).toBeLessThanOrEqual(detailBox!.x);
  expect(detailBox!.width).toBeGreaterThan(treeBox!.width);
  expect(
    await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth),
  ).toBe(true);

  const secondRow = treePane.locator("[data-category-row-guid]").nth(1);
  await expect(secondRow).toBeVisible();
  let delayNextDetail = true;
  await page.route("**/api/system-settings/productcategorygroupselectscreen/**", async (route) => {
    if (delayNextDetail && route.request().method() === "GET") {
      delayNextDetail = false;
      await new Promise((resolve) => setTimeout(resolve, 350));
    }
    await route.continue();
  });
  const transitionProbe = page.evaluate(async () => {
    const pane = document.querySelector<HTMLElement>("[data-testid='product-category-detail-pane']")!;
    let sawEmptyFrame = false;
    let sawLoadingChrome = false;
    const deadline = performance.now() + 450;
    while (performance.now() < deadline) {
      const text = pane.innerText.trim();
      sawEmptyFrame ||= text.length === 0;
      sawLoadingChrome ||= text.includes("กำลังโหลด");
      await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
    }
    return { sawEmptyFrame, sawLoadingChrome };
  });
  await secondRow.click();
  expect(await transitionProbe).toEqual({
    sawEmptyFrame: false,
    sawLoadingChrome: false,
  });
  await expect(detailPane.getByRole("button", { name: "แก้ไข" })).toBeVisible();

  const secondName = (await secondRow.locator("span.basis-40").innerText()).trim();
  delayNextDetail = true;
  await firstRow.click();
  await page.waitForTimeout(25);
  await secondRow.click();
  await page.waitForTimeout(450);
  await expect(secondRow).toHaveAttribute("aria-selected", "true");
  await expect(detailPane.getByText(secondName, { exact: true }).first()).toBeVisible();

  await detailPane.getByRole("button", { name: "แก้ไข" }).click();
  const editForm = detailPane.getByRole("form", { name: "แก้ไข" });
  await expect(editForm).toBeVisible();
  await expect(editForm.getByText("ลำดับกลุ่ม", { exact: true })).toHaveCount(0);
  await expect(editForm.getByText("หมวดแม่", { exact: true })).toHaveCount(0);
  await expect(editForm.getByRole("button", { name: "ยกเลิก" })).toBeVisible();
  await expect(editForm.getByRole("button", { name: "บันทึก" })).toBeVisible();
  await editForm.getByRole("button", { name: "ยกเลิก" }).click();
  await expect(detailPane.getByRole("button", { name: "แก้ไข" })).toBeVisible();
});

test("productcategorylist — category editor stays mounted while Mongo detail switches", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1024, height: 768 });
  await loginAndSearchMenu(page, "สินค้าในหมวด");
  await clickButtonByText(page, /^สินค้าในหมวด$/);
  await page.waitForTimeout(1500);
  await clickGroupNumber(page, 1);

  const treePane = page.getByTestId("product-category-tree-pane");
  const detailPane = page.getByTestId("product-category-detail-pane");
  const firstRow = treePane.locator("[data-category-row-guid]").first();
  const secondRow = treePane.locator("[data-category-row-guid]").nth(1);
  await expect(firstRow).toBeVisible();
  await expect(secondRow).toBeVisible();
  await firstRow.click();
  await expect(detailPane.getByText("สินค้าในหมวด", { exact: true })).toBeVisible();
  const categoryActions = detailPane.getByTestId("product-category-item-actions");
  const categoryHeading = detailPane.getByRole("heading", { name: "สินค้าในหมวด", exact: true });
  const categoryCount = detailPane.getByText(/^ทั้งหมด \d+ รายการ$/);
  await expect(categoryActions).toBeInViewport();
  await expect(categoryCount).toBeInViewport();
  const [categoryActionsBox, categoryHeadingBox, categoryCountBox] = await Promise.all([
    categoryActions.boundingBox(),
    categoryHeading.boundingBox(),
    categoryCount.boundingBox(),
  ]);
  expect(categoryActionsBox).not.toBeNull();
  expect(categoryHeadingBox).not.toBeNull();
  expect(categoryCountBox).not.toBeNull();
  expect(categoryActionsBox!.y).toBeLessThan(categoryHeadingBox!.y + 72);
  expect(categoryCountBox!.y).toBeLessThan(categoryHeadingBox!.y + 60);

  let delayNextDetail = true;
  await page.route("**/api/system-settings/productcategorylist/**", async (route) => {
    if (delayNextDetail && route.request().method() === "GET") {
      delayNextDetail = false;
      await new Promise((resolve) => setTimeout(resolve, 350));
    }
    await route.continue();
  });
  const transitionProbe = page.evaluate(async () => {
    const pane = document.querySelector<HTMLElement>("[data-testid='product-category-detail-pane']")!;
    let sawEmptyFrame = false;
    let sawLoadingOrEmptyChrome = false;
    const deadline = performance.now() + 450;
    while (performance.now() < deadline) {
      const text = pane.innerText.trim();
      sawEmptyFrame ||= text.length === 0;
      sawLoadingOrEmptyChrome ||=
        text.includes("กำลังโหลดข้อมูลสินค้า") || text.includes("เลือกหมวดสินค้า");
      await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
    }
    return { sawEmptyFrame, sawLoadingOrEmptyChrome };
  });
  await secondRow.click();
  expect(await transitionProbe).toEqual({
    sawEmptyFrame: false,
    sawLoadingOrEmptyChrome: false,
  });

  const secondName = (await secondRow.locator("span.basis-40").innerText()).trim();
  await expect(detailPane.getByText(secondName, { exact: true })).toBeVisible();
  delayNextDetail = true;
  await firstRow.click();
  await page.waitForTimeout(25);
  await secondRow.click();
  await page.waitForTimeout(450);
  await expect(secondRow).toHaveAttribute("aria-selected", "true");
  await expect(detailPane.getByText(secondName, { exact: true })).toBeVisible();
});

function getAuthHeaders(page: Page) {
  return page.evaluate(() => {
    const auth = JSON.parse(localStorage.getItem("bc_auth")!) as { token: string; backendUrl: string };
    return { Authorization: `Bearer ${auth.token}`, "x-bc-backend-url": auth.backendUrl };
  });
}

/** Fills the primary-language input of the "names" multilingual editor (`NamesEditor` component).
 *  Unlike `product-screen.tsx` (whose primary name input has `placeholder="th"`), THIS screen's
 *  create form resolves the primary language code to an empty string, leaving `placeholder=""` --
 *  confirmed live via the running dev app (`input[placeholder="th"]` never matches here, unlike on
 *  the product screen). The reliable anchor across both cases is the wrapping `<label>` containing
 *  "ภาษาแรก" ("Primary language"), which always renders regardless of the resolved code. */
async function fillPrimaryNameInput(page: Page, value: string) {
  await page.evaluate((v) => {
    const label = [...document.querySelectorAll<HTMLElement>("label")].find(
      (l) => (l.textContent ?? "").includes("ภาษาแรก") && l.offsetParent !== null,
    );
    if (!label) throw new Error('No "ภาษาแรก" (primary language) label found');
    const input = label.querySelector<HTMLInputElement>("input");
    if (!input) throw new Error('No input found under the "ภาษาแรก" label');
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, v);
    input.dispatchEvent(new Event("input", { bubbles: true }));
    input.dispatchEvent(new Event("change", { bubbles: true }));
  }, value);
}

test("productcategorygroupselectscreen -> productcategorylist — create a 2-level category branch and bind a Product master by code only", async ({
  page,
}) => {
  // Group 19: still-unused slot (1/2 hold real seeded sample data; 20 is reserved for the two
  // empty-state tests above and must stay empty for those to keep passing).
  const uid = Date.now().toString().slice(-6);
  const rootName = `E2Eหมวดหลัก${uid}`;
  const childName = `E2Eหมวดย่อย${uid}`;

  await loginAndSearchMenu(page, "จัดหมวดสินค้า");
  const H = await getAuthHeaders(page);
  const productResponse = await page.request.get(`${MAINAPI}/product?limit=1`, { headers: H });
  expect(productResponse.ok(), "load a real Product master").toBe(true);
  const productPayload = (await productResponse.json()) as {
    data?: { code: string; names?: { code: string; name: string }[] }[];
  };
  const product = (productPayload.data ?? []).find((item) => item.code);
  expect(product, "at least one Product master exists").toBeTruthy();
  const productCode = product!.code;

  const invalidBase = {
    names: [{ code: "th", name: `E2EตรวจProduct${uid}` }],
    xsorts: [],
    groupnumber: 19,
  };
  const missingProduct = await page.request.post(`${MAINAPI}/product/category`, {
    headers: H,
    data: { ...invalidBase, codelist: [{ code: `NOTEXIST${uid}`, xorder: 0 }] },
  });
  expect(missingProduct.status(), "reject a code absent from Product master").toBe(400);
  expect(JSON.stringify(await missingProduct.json())).toContain("ไม่พบสินค้า");

  const logicallySameCode = `${productCode.slice(0, 1)} ${productCode.slice(1)}`.toLowerCase();
  const duplicateProduct = await page.request.post(`${MAINAPI}/product/category`, {
    headers: H,
    data: {
      ...invalidBase,
      codelist: [
        { code: productCode, xorder: 0 },
        { code: logicallySameCode, xorder: 1 },
      ],
    },
  });
  expect(duplicateProduct.status(), "reject logical duplicate Product codes after normalization").toBe(400);
  expect(JSON.stringify(await duplicateProduct.json())).toContain("ซ้ำในหมวด");
  await clickButtonByText(page, /^จัดหมวดสินค้า$/);
  await page.waitForTimeout(2000);
  await clickGroupNumber(page, 19);
  await page.waitForTimeout(1200);

  // Create the root category.
  await clickButtonByText(page, /^เพิ่มหมวดหลัก$/);
  await page.waitForTimeout(500);
  await fillPrimaryNameInput(page, rootName);
  const [rootRes] = await Promise.all([
    page.waitForResponse(
      (res) => res.url().includes("/api/system-settings/productcategorygroupselectscreen") && res.request().method() === "POST",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(rootRes.ok(), "create root category").toBe(true);
  // The system-settings create proxy responds `{success, id}` (not `{data: {...}}`) -- confirmed
  // live via a direct curl to /api/system-settings/productcategorygroupselectscreen.
  const rootBody = (await rootRes.json()) as { id: string };
  const rootGuid = rootBody.id;
  expect(rootGuid).toBeTruthy();
  await page.waitForTimeout(1000);

  // Select the new root row, then create a subcategory under it.
  await clickTreeRowByText(page, rootName);
  await page.waitForTimeout(500);
  await clickButtonByText(page, /^เพิ่มหมวดย่อย$/);
  await page.waitForTimeout(500);
  await fillPrimaryNameInput(page, childName);
  const [childRes] = await Promise.all([
    page.waitForResponse(
      (res) => res.url().includes("/api/system-settings/productcategorygroupselectscreen") && res.request().method() === "POST",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(childRes.ok(), "create subcategory").toBe(true);
  const childBody = (await childRes.json()) as { id: string };
  const childGuid = childBody.id;
  expect(childGuid).toBeTruthy();
  await page.waitForTimeout(1000);

  // Switch to productcategorylist (same group) and bind the Product master into the leaf.
  await page.goto("/productcategorylist", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await clickGroupNumber(page, 19);
  await page.waitForTimeout(1200);
  // Root's subcategories render collapsed by default -- expand before the child row is clickable.
  await page.evaluate(() => {
    [...document.querySelectorAll<HTMLElement>('button[aria-label^="แสดงหมวดย่อย"]')]
      .filter((b) => b.offsetParent !== null)
      .forEach((b) => b.click());
  });
  await page.waitForTimeout(600);
  await clickTreeRowByText(page, childName);
  await page.waitForTimeout(800);

  const pickerRequestUrls: string[] = [];
  page.on("request", (request) => {
    if (request.url().includes("/api/product")) pickerRequestUrls.push(request.url());
  });
  await clickButtonByText(page, /เพิ่มสินค้า/);
  await page.waitForTimeout(800);
  const productDialog = page.getByRole("dialog", { name: "ค้นหาและเพิ่มสินค้า" });
  await expect(productDialog).toBeVisible();
  await expect(productDialog).not.toContainText("บาร์โค้ด");
  await page.fill('input[placeholder*="ค้นหาด้วยรหัสสินค้า"]', productCode);
  await page.waitForTimeout(800);
  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(400);
  expect(pickerRequestUrls.some((url) => url.includes("/api/product?")), "picker reads Product master API").toBe(true);
  expect(pickerRequestUrls.some((url) => url.includes("/api/product-barcode")), "picker never reads ProductBarcode API").toBe(false);
  await clickButtonByText(page, /^ปิด$/);
  await page.waitForTimeout(400);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(1500);
  expect(await bodyHasText(page, "บันทึกข้อมูลเรียบร้อยแล้ว")).toBe(true);

  // Verify via fresh API read: the tree really nests (child.parentguid === root.guidfixed) and the
  // leaf's codelist contains only the bound Product code -- not a barcode relation. Explicit
  // `limit` because group 19 also holds unrelated leftover data seeded in earlier sessions and the
  // backend's default page size is small enough to push this test's own just-created records off
  // the first page once enough other data accumulates.
  const catRes = await (
    await page.request.get(`${MAINAPI}/product/category/list?group-number=19&limit=100000`, { headers: H })
  ).json();
  const records = catRes.data as {
    guidfixed: string;
    parentguid?: string;
    names: { code: string; name: string }[];
    codelist?: { code: string; barcode?: unknown }[];
  }[];
  const root = records.find((r) => r.guidfixed === rootGuid);
  const child = records.find((r) => r.guidfixed === childGuid);
  expect(root, "root category present via API").toBeTruthy();
  expect(child, "subcategory present via API").toBeTruthy();
  expect(child!.parentguid, "subcategory nests under root").toBe(rootGuid);
  const categoryProducts = child!.codelist ?? [];
  expect(categoryProducts.map((item) => item.code), "leaf codelist contains Product code").toContain(productCode);
  expect(categoryProducts.every((item) => !("barcode" in item)), "codelist is Product-only").toBe(true);

  // Remove only the records created by this test; never touch pre-existing debug/orphan data.
  const childDelete = await page.request.delete(`${MAINAPI}/product/category/${childGuid}`, { headers: H });
  expect(childDelete.ok(), "cleanup test subcategory").toBe(true);
  const rootDelete = await page.request.delete(`${MAINAPI}/product/category/${rootGuid}`, { headers: H });
  expect(rootDelete.ok(), "cleanup test root category").toBe(true);
});

/**
 * Regression coverage for Jead's 2026-07-08 report "ลากย้าย น้ำอัดลม ไว้ก่อนกาแฟสำเร็จรูป ทำไม่ได้"
 * on this same screen (`productcategorygroupselectscreen`, sibling drag-to-reorder within a
 * category's children). Root-caused (see `.agents/worklog.md` 2026-07-08 entries) as stale tree
 * data plus a documented short-drag-falls-through-to-click nuance, NOT a broken reorder mechanism
 * -- independently re-verified live via real claude-in-chrome pointer press-move-release drags
 * (multiple successful `xsort` round trips, each confirmed against a fresh MongoDB read and a hard
 * page reload). That same live verification also surfaced a separate, real precision finding: this
 * component's drop-zone thresholds (~28% of a ~45px row) were tight enough that an automated drag
 * landing within a few px of a parent/child or root/child boundary could silently resolve to the
 * WRONG operation (e.g. reparenting instead of reordering) even when "started" (5px-threshold)
 * drag detection fired correctly and the request succeeded with 201.
 *
 * FIXED 2026-07-08 (later): root cause was `getPointerDropCandidate` (product-category-tree-view.tsx,
 * then ~1053-1105) resolving the drop target via `document.elementFromPoint()` + DOM `closest()` --
 * hit-testing the drag's own rendered side effects (the dragged row is `pointer-events-none` while
 * dragging; an inline "วางเป็นลูก" chip could out-compete a neighboring row's own zone) instead of
 * stable row geometry. Rewritten to resolve purely from live `getBoundingClientRect()` containment.
 * Independently re-verified in a fresh browser session (real claude-in-chrome pointer drags, own
 * MongoDB/API reads, not just the report's own claims): a drag landing 4-6px into a first child's
 * "before" zone (right at the parent-row/first-child boundary) now correctly resolves to a same-
 * parent sibling reorder, and a drag landing 2-3px above a last-child/next-root-sibling boundary
 * (within the last child's own "after" zone) now correctly stays a same-parent reorder instead of
 * escaping to root -- see `.agents/worklog.md` for the full pixel-level evidence.
 *
 * A real Chromium pointer drag is NOT simulated here -- Playwright's `mouse.move/down/up` land at
 * exact CSS-pixel targets (no viewport-scale ambiguity), but this component's live FLIP-animation
 * preview reflow during a drag makes a scripted multi-tick drag path flaky to author reliably, and
 * the app-level pointer/preview logic itself is not this test's concern. Per the task's own
 * allowance, this test instead drives the SAME real `xsort` PUT the drag handler calls
 * (`saveXSorts` in `product-category-tree-view.tsx`, same payload shape:
 * `{guidfixed, code, xorder}[]`), then verifies the reorder against a fresh API read AND the real
 * UI after a hard reload -- covering the backend/persistence/display half of the flow. What this
 * test does NOT cover: the pointer-drag gesture itself, and specifically the pixel-to-zone geometry
 * decision inside `getPointerDropCandidate`/`getRowDropPosition` that this fix actually changed --
 * that logic runs entirely client-side before any API call is made, so no API-level assertion can
 * exercise it; only a real (or simulated) pointer drag can. That half was verified manually this
 * session via real claude-in-chrome drags at calibrated boundary pixels, not via CI. The companion
 * test below (`reparent via the real record+xsort API...`) adds data-level coverage for the OTHER
 * half of the same decision -- the "inside" branch (reparent) -- using the exact payload shape
 * `moveCategoryAsChild` sends, complementing this test's "before/after" (reorder) coverage.
 */
test("productcategorygroupselectscreen — sibling reorder via the real xsort API persists and displays in the new order", async ({
  page,
}) => {
  const uid = Date.now().toString().slice(-6);
  const rootName = `E2Eหมวดเรียง${uid}`;
  const childNames = [`E2Eลูกที่1_${uid}`, `E2Eลูกที่2_${uid}`, `E2Eลูกที่3_${uid}`];

  await loginAndSearchMenu(page, "จัดหมวดสินค้า");
  await clickButtonByText(page, /^จัดหมวดสินค้า$/);
  await page.waitForTimeout(2000);
  await clickGroupNumber(page, 19);
  await page.waitForTimeout(1200);

  // Create the root, then 3 children under it (creation order == initial xorder 1,2,3).
  await clickButtonByText(page, /^เพิ่มหมวดหลัก$/);
  await page.waitForTimeout(500);
  await fillPrimaryNameInput(page, rootName);
  const [rootRes] = await Promise.all([
    page.waitForResponse(
      (res) => res.url().includes("/api/system-settings/productcategorygroupselectscreen") && res.request().method() === "POST",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(rootRes.ok(), "create root category").toBe(true);
  const rootGuid = ((await rootRes.json()) as { id: string }).id;
  expect(rootGuid).toBeTruthy();
  await page.waitForTimeout(1000);

  const childGuids: string[] = [];
  for (const childName of childNames) {
    await clickTreeRowByText(page, rootName);
    await page.waitForTimeout(500);
    await clickButtonByText(page, /^เพิ่มหมวดย่อย$/);
    await page.waitForTimeout(500);
    await fillPrimaryNameInput(page, childName);
    const [childRes] = await Promise.all([
      page.waitForResponse(
        (res) => res.url().includes("/api/system-settings/productcategorygroupselectscreen") && res.request().method() === "POST",
      ),
      clickButtonByText(page, /^บันทึก$/),
    ]);
    expect(childRes.ok(), `create subcategory ${childName}`).toBe(true);
    const childGuid = ((await childRes.json()) as { id: string }).id;
    expect(childGuid).toBeTruthy();
    childGuids.push(childGuid);
    await page.waitForTimeout(800);
  }

  const H = await getAuthHeaders(page);
  // Explicit `limit` -- group 19 also holds unrelated leftover data seeded in earlier sessions, and
  // the backend's default page size is small enough to push this test's own records off page 1.
  const listUrl = `${MAINAPI}/product/category/list?group-number=19&limit=100000`;

  // Confirm the initial order really is 1,2,3 in creation order before touching it.
  const before = ((await (await page.request.get(listUrl, { headers: H })).json()).data as {
    guidfixed: string;
    xsorts?: { xorder: number }[];
  }[]);
  const orderOf = (recs: typeof before, guid: string) => recs.find((r) => r.guidfixed === guid)?.xsorts?.[0]?.xorder;
  expect(orderOf(before, childGuids[0]), "child1 starts at xorder 1").toBe(1);
  expect(orderOf(before, childGuids[1]), "child2 starts at xorder 2").toBe(2);
  expect(orderOf(before, childGuids[2]), "child3 starts at xorder 3").toBe(3);

  // Drive the SAME real xsort endpoint the drag handler's `saveXSorts` calls, with the same
  // payload shape, to reorder child3 to before child2 (child1, child3, child2).
  const xsortRes = await page.request.put(
    `/api/system-settings/productcategorygroupselectscreen/xsort?holdingcode=test`,
    {
      headers: { ...H, "Content-Type": "application/json" },
      data: [
        { guidfixed: childGuids[0], code: "X", xorder: 1 },
        { guidfixed: childGuids[2], code: "X", xorder: 2 },
        { guidfixed: childGuids[1], code: "X", xorder: 3 },
      ],
    },
  );
  expect(xsortRes.ok(), "xsort reorder request").toBe(true);

  // Verify via a FRESH API read (not the optimistic response) that the reorder really persisted.
  const after = ((await (await page.request.get(listUrl, { headers: H })).json()).data as typeof before);
  expect(orderOf(after, childGuids[0]), "child1 stays at xorder 1").toBe(1);
  expect(orderOf(after, childGuids[2]), "child3 moved to xorder 2").toBe(2);
  expect(orderOf(after, childGuids[1]), "child2 moved to xorder 3").toBe(3);

  // Verify via a hard reload (fresh JS memory, no in-memory/optimistic state) that the real UI
  // displays the new order too, not just that the API accepted the write.
  await page.goto("/productcategorygroupselectscreen", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await clickGroupNumber(page, 19);
  await page.waitForTimeout(1200);
  await page.evaluate(() => {
    [...document.querySelectorAll<HTMLElement>('button[aria-label^="แสดงหมวดย่อย"]')]
      .filter((b) => b.offsetParent !== null)
      .forEach((b) => b.click());
  });
  await page.waitForTimeout(600);
  const bodyText = await page.evaluate(() => document.body.innerText);
  const idx1 = bodyText.indexOf(childNames[0]);
  const idx3 = bodyText.indexOf(childNames[2]);
  const idx2 = bodyText.indexOf(childNames[1]);
  expect(idx1, "child1 rendered").toBeGreaterThan(-1);
  expect(idx3, "child3 rendered").toBeGreaterThan(-1);
  expect(idx2, "child2 rendered").toBeGreaterThan(-1);
  expect(idx1 < idx3 && idx3 < idx2, "UI displays child1, child3, child2 in that order after reload").toBe(true);
});

/**
 * Companion to the sibling-reorder test above -- covers the OTHER branch of the same fixed
 * decision (`getPointerDropCandidate` choosing "inside" instead of "before"/"after"): a real
 * reparent. Drives the exact record-PUT + xsort-PUT payload shape `moveCategoryAsChild` sends
 * (`product-category-tree-view.tsx`), not a raw arbitrary PUT, so this exercises the real
 * persistence contract the fixed drop-resolution code relies on. Does NOT simulate the pointer
 * drag or the pixel-to-zone geometry decision itself -- see the comment above for why that part
 * can only be covered by a real/simulated drag, not an API-level test.
 */
test("productcategorygroupselectscreen — reparent via the real record+xsort API (moveCategoryAsChild path) persists correct parentguid/parentguidall", async ({
  page,
}) => {
  const uid = Date.now().toString().slice(-6);
  const rootName = `E2Eหมวดย้าย${uid}`;
  const childNames = [`E2Eลูกที่1_${uid}`, `E2Eลูกที่2_${uid}`];

  await loginAndSearchMenu(page, "จัดหมวดสินค้า");
  await clickButtonByText(page, /^จัดหมวดสินค้า$/);
  await page.waitForTimeout(2000);
  await clickGroupNumber(page, 19);
  await page.waitForTimeout(1200);

  await clickButtonByText(page, /^เพิ่มหมวดหลัก$/);
  await page.waitForTimeout(500);
  await fillPrimaryNameInput(page, rootName);
  const [rootRes] = await Promise.all([
    page.waitForResponse(
      (res) => res.url().includes("/api/system-settings/productcategorygroupselectscreen") && res.request().method() === "POST",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(rootRes.ok(), "create root category").toBe(true);
  const rootGuid = ((await rootRes.json()) as { id: string }).id;
  expect(rootGuid).toBeTruthy();
  await page.waitForTimeout(1000);

  const childGuids: string[] = [];
  for (const childName of childNames) {
    await clickTreeRowByText(page, rootName);
    await page.waitForTimeout(500);
    await clickButtonByText(page, /^เพิ่มหมวดย่อย$/);
    await page.waitForTimeout(500);
    await fillPrimaryNameInput(page, childName);
    const [childRes] = await Promise.all([
      page.waitForResponse(
        (res) => res.url().includes("/api/system-settings/productcategorygroupselectscreen") && res.request().method() === "POST",
      ),
      clickButtonByText(page, /^บันทึก$/),
    ]);
    expect(childRes.ok(), `create subcategory ${childName}`).toBe(true);
    const childGuid = ((await childRes.json()) as { id: string }).id;
    expect(childGuid).toBeTruthy();
    childGuids.push(childGuid);
    await page.waitForTimeout(800);
  }
  const [child1Guid, child2Guid] = childGuids;

  const H = await getAuthHeaders(page);
  const recordUrl = (guid: string) => `/api/system-settings/productcategorygroupselectscreen/${guid}?holdingcode=test`;
  // group 19 now also holds unrelated leftover data seeded earlier this session (see
  // .agents/worklog.md) -- the backend default page size is small, so an explicit high `limit` is
  // required or this test's own just-created records can be pushed off the first page.
  const listUrl = `${MAINAPI}/product/category/list?group-number=19&limit=100000`;

  // Reparent child2 to become a child of child1 -- the exact payload shape `moveCategoryAsChild`
  // builds: spread the fetched record, override parentguid/parentguidall/xsorts.
  const child2Before = ((await (await page.request.get(recordUrl(child2Guid), { headers: H })).json()) as {
    data: Record<string, unknown>;
  }).data;
  const newParentGuidAll = `${rootGuid},${child1Guid}`;
  const reparentRes = await page.request.put(recordUrl(child2Guid), {
    headers: { ...H, "Content-Type": "application/json" },
    data: { ...child2Before, parentguid: child1Guid, parentguidall: newParentGuidAll, xsorts: [{ code: "X", xorder: 1 }] },
  });
  expect(reparentRes.ok(), "reparent record PUT").toBe(true);
  const xsortRes = await page.request.put(
    `/api/system-settings/productcategorygroupselectscreen/xsort?holdingcode=test`,
    {
      headers: { ...H, "Content-Type": "application/json" },
      data: [
        { guidfixed: child1Guid, code: "X", xorder: 1 }, // renumbered root-level sibling left behind
        { guidfixed: child2Guid, code: "X", xorder: 1 }, // first (only) child of its new parent
      ],
    },
  );
  expect(xsortRes.ok(), "post-reparent xsort PUT").toBe(true);

  // Verify via a FRESH API read (not the optimistic response) that the reparent really persisted,
  // and that the old parent's remaining child (child1) was left structurally untouched.
  const after = ((await (await page.request.get(listUrl, { headers: H })).json()).data as {
    guidfixed: string;
    parentguid?: string;
    parentguidall?: string;
    xsorts?: { xorder: number }[];
  }[]);
  const child1After = after.find((r) => r.guidfixed === child1Guid);
  const child2After = after.find((r) => r.guidfixed === child2Guid);
  expect(child1After, "child1 present via API").toBeTruthy();
  expect(child2After, "child2 present via API").toBeTruthy();
  expect(child2After!.parentguid, "child2 now nests under child1, not root").toBe(child1Guid);
  expect(child2After!.parentguidall, "child2's ancestor chain includes root then child1").toBe(newParentGuidAll);
  expect(child1After!.parentguid, "child1 stays under root (untouched by the reparent)").toBe(rootGuid);
  expect(child1After!.xsorts?.[0]?.xorder, "child1 keeps xorder 1 under root").toBe(1);
});

/**
 * Regression coverage for the 2026-07-09 port of `product-group-tree-view.tsx`'s per-row up/down
 * move buttons + session-local Undo/Redo stack onto this screen's `product-category-tree-view.tsx`
 * (up/down call the same `reorderCategory()` the drag handler uses; Undo/Redo replay moves through
 * the real `saveCategoryRecord`/`saveXSorts` calls, not a local-only revert). Independently
 * re-verified live this session via a fresh browser session (cleared localStorage/sessionStorage,
 * redone dev-test-login) driving real button clicks and a real scripted pointer-drag reparent on
 * group 1's seeded เครื่องดื่ม/ของใช้ในบ้าน/เครื่องเขียน tree, each step checked against a fresh API
 * read — see `.agents/worklog.md` 2026-07-09. Mirrors `product-group-tree-crud.spec.ts`'s
 * "up/down move buttons + undo/redo" test for the sibling screen.
 */
test("productcategorygroupselectscreen — up/down move buttons + undo/redo reorder and revert via the real xsort API", async ({
  page,
}) => {
  const uid = Date.now().toString().slice(-6);
  const rootName = `E2Eหมวดปุ่ม${uid}`;
  const childNames = [`E2Eปุ่มลูก1_${uid}`, `E2Eปุ่มลูก2_${uid}`];

  await loginAndSearchMenu(page, "จัดหมวดสินค้า");
  await clickButtonByText(page, /^จัดหมวดสินค้า$/);
  await page.waitForTimeout(2000);
  await clickGroupNumber(page, 19);
  await page.waitForTimeout(1200);

  await clickButtonByText(page, /^เพิ่มหมวดหลัก$/);
  await page.waitForTimeout(500);
  await fillPrimaryNameInput(page, rootName);
  const [rootRes] = await Promise.all([
    page.waitForResponse(
      (res) => res.url().includes("/api/system-settings/productcategorygroupselectscreen") && res.request().method() === "POST",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(rootRes.ok(), "create root category").toBe(true);
  const rootGuid = ((await rootRes.json()) as { id: string }).id;
  expect(rootGuid).toBeTruthy();
  await page.waitForTimeout(1000);

  const childGuids: string[] = [];
  for (const childName of childNames) {
    await clickTreeRowByText(page, rootName);
    await page.waitForTimeout(500);
    await clickButtonByText(page, /^เพิ่มหมวดย่อย$/);
    await page.waitForTimeout(500);
    await fillPrimaryNameInput(page, childName);
    const [childRes] = await Promise.all([
      page.waitForResponse(
        (res) => res.url().includes("/api/system-settings/productcategorygroupselectscreen") && res.request().method() === "POST",
      ),
      clickButtonByText(page, /^บันทึก$/),
    ]);
    expect(childRes.ok(), `create subcategory ${childName}`).toBe(true);
    const childGuid = ((await childRes.json()) as { id: string }).id;
    expect(childGuid).toBeTruthy();
    childGuids.push(childGuid);
    await page.waitForTimeout(800);
  }
  const [child1Guid, child2Guid] = childGuids;

  const H = await getAuthHeaders(page);
  const listUrl = `${MAINAPI}/product/category/list?group-number=19&limit=100000`;

  // Reload so the tree view picks up the freshly-created children (fresh JS memory, no stale
  // in-memory overrides), then expand the root to reveal the up/down buttons on its children.
  await page.goto("/productcategorygroupselectscreen", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await clickGroupNumber(page, 19);
  await page.waitForTimeout(1200);
  await page.evaluate((guid) => {
    const row = document.querySelector<HTMLElement>(`[data-category-row-guid="${guid}"]`);
    const toggle = row?.querySelector<HTMLButtonElement>('button[aria-label^="แสดงหมวดย่อย"]');
    if (!toggle) throw new Error("No expand toggle found on root row");
    toggle.click();
  }, rootGuid);
  await page.waitForTimeout(600);

  const orderOf = async () => {
    const data = ((await (await page.request.get(listUrl, { headers: H })).json()).data as {
      guidfixed: string;
      xsorts?: { xorder: number }[];
    }[]);
    return {
      child1: data.find((r) => r.guidfixed === child1Guid)?.xsorts?.[0]?.xorder,
      child2: data.find((r) => r.guidfixed === child2Guid)?.xsorts?.[0]?.xorder,
    };
  };

  // Boundary check: the first sibling's up button and the last sibling's down button must be
  // disabled (can't move a first child further up, or a last child further down).
  const boundaryDisabled = await page.evaluate(({ c1, c2 }) => {
    const row1 = document.querySelector<HTMLElement>(`[data-category-row-guid="${c1}"]`);
    const row2 = document.querySelector<HTMLElement>(`[data-category-row-guid="${c2}"]`);
    const up1 = row1?.querySelector<HTMLButtonElement>('button[aria-label="ย้ายขึ้น"]');
    const down2 = row2?.querySelector<HTMLButtonElement>('button[aria-label="ย้ายลง"]');
    return { firstUpDisabled: up1?.disabled, lastDownDisabled: down2?.disabled };
  }, { c1: child1Guid, c2: child2Guid });
  expect(boundaryDisabled.firstUpDisabled, "first sibling's up button is disabled").toBe(true);
  expect(boundaryDisabled.lastDownDisabled, "last sibling's down button is disabled").toBe(true);

  expect(await orderOf(), "children start in creation order").toEqual({ child1: 1, child2: 2 });

  // Click the "ย้ายลง" (move down) button on child1's row -- calls the exact same `reorderCategory()`
  // the drag handler uses, just with a fixed "before"/"after" target instead of a pointer position.
  const [downRes] = await Promise.all([
    page.waitForResponse((res) => res.url().includes("/xsort") && res.request().method() === "PUT"),
    page.evaluate((guid) => {
      const row = document.querySelector<HTMLElement>(`[data-category-row-guid="${guid}"]`);
      const btn = row?.querySelector<HTMLButtonElement>('button[aria-label="ย้ายลง"]');
      if (!btn) throw new Error("No move-down button found on child1's row");
      btn.click();
    }, child1Guid),
  ]);
  expect(downRes.ok(), "move-down xsort PUT").toBe(true);
  expect(await orderOf(), "down button swapped child1/child2 via the real API").toEqual({ child1: 2, child2: 1 });

  // Undo -- must revert through the real record/xsort API (`applyMoveSnapshot`), not just local state.
  const [undoRes] = await Promise.all([
    page.waitForResponse((res) => res.url().includes("/xsort") && res.request().method() === "PUT"),
    clickButtonByText(page, /^เลิกทำ$/),
  ]);
  expect(undoRes.ok(), "undo xsort PUT").toBe(true);
  expect(await orderOf(), "undo restored creation order via the real API").toEqual({ child1: 1, child2: 2 });

  // Redo -- must reapply the exact same move via the real API.
  const [redoRes] = await Promise.all([
    page.waitForResponse((res) => res.url().includes("/xsort") && res.request().method() === "PUT"),
    clickButtonByText(page, /^ทำซ้ำ$/),
  ]);
  expect(redoRes.ok(), "redo xsort PUT").toBe(true);
  expect(await orderOf(), "redo re-applied the swap via the real API").toEqual({ child1: 2, child2: 1 });
});

/**
 * Companion to the up/down + undo/redo test above -- the existing test only exercises a single
 * linear undo-then-redo sequence, which does not catch a redo tail that fails to clear after a
 * genuinely NEW move is made post-undo (i.e. a stale/leftover "future" branch of history incorrectly
 * staying replayable after being superseded). Standard undo/redo semantics require a new action taken
 * after an undo to discard that redo tail -- confirmed still correct here via live re-verification
 * this session, mirroring `product-group-tree-crud.spec.ts`'s equivalent test for the sibling screen.
 * Uses 3 siblings so "move A" and "move B" are unambiguously different operations (not just the same
 * swap re-applied).
 */
test("productcategorygroupselectscreen — a new move after Undo clears the Redo tail (does not replay the undone move)", async ({
  page,
}) => {
  const uid = Date.now().toString().slice(-6);
  const rootName = `E2Eหมวดล้าง${uid}`;
  const childNames = [`E2Eล้าง1_${uid}`, `E2Eล้าง2_${uid}`, `E2Eล้าง3_${uid}`];

  await loginAndSearchMenu(page, "จัดหมวดสินค้า");
  await clickButtonByText(page, /^จัดหมวดสินค้า$/);
  await page.waitForTimeout(2000);
  await clickGroupNumber(page, 19);
  await page.waitForTimeout(1200);

  await clickButtonByText(page, /^เพิ่มหมวดหลัก$/);
  await page.waitForTimeout(500);
  await fillPrimaryNameInput(page, rootName);
  const [rootRes] = await Promise.all([
    page.waitForResponse(
      (res) => res.url().includes("/api/system-settings/productcategorygroupselectscreen") && res.request().method() === "POST",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(rootRes.ok(), "create root category").toBe(true);
  const rootGuid = ((await rootRes.json()) as { id: string }).id;
  expect(rootGuid).toBeTruthy();
  await page.waitForTimeout(1000);

  const childGuids: string[] = [];
  for (const childName of childNames) {
    await clickTreeRowByText(page, rootName);
    await page.waitForTimeout(500);
    await clickButtonByText(page, /^เพิ่มหมวดย่อย$/);
    await page.waitForTimeout(500);
    await fillPrimaryNameInput(page, childName);
    const [childRes] = await Promise.all([
      page.waitForResponse(
        (res) => res.url().includes("/api/system-settings/productcategorygroupselectscreen") && res.request().method() === "POST",
      ),
      clickButtonByText(page, /^บันทึก$/),
    ]);
    expect(childRes.ok(), `create subcategory ${childName}`).toBe(true);
    const childGuid = ((await childRes.json()) as { id: string }).id;
    expect(childGuid).toBeTruthy();
    childGuids.push(childGuid);
    await page.waitForTimeout(800);
  }
  const [child1Guid, child2Guid, child3Guid] = childGuids;

  const H = await getAuthHeaders(page);
  const listUrl = `${MAINAPI}/product/category/list?group-number=19&limit=100000`;

  await page.goto("/productcategorygroupselectscreen", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await clickGroupNumber(page, 19);
  await page.waitForTimeout(1200);
  await page.evaluate((guid) => {
    const row = document.querySelector<HTMLElement>(`[data-category-row-guid="${guid}"]`);
    const toggle = row?.querySelector<HTMLButtonElement>('button[aria-label^="แสดงหมวดย่อย"]');
    if (!toggle) throw new Error("No expand toggle found on root row");
    toggle.click();
  }, rootGuid);
  await page.waitForTimeout(600);

  const orderOf = async () => {
    const data = ((await (await page.request.get(listUrl, { headers: H })).json()).data as {
      guidfixed: string;
      xsorts?: { xorder: number }[];
    }[]);
    return {
      child1: data.find((r) => r.guidfixed === child1Guid)?.xsorts?.[0]?.xorder,
      child2: data.find((r) => r.guidfixed === child2Guid)?.xsorts?.[0]?.xorder,
      child3: data.find((r) => r.guidfixed === child3Guid)?.xsorts?.[0]?.xorder,
    };
  };
  const clickMoveDown = async (guid: string) => {
    const [res] = await Promise.all([
      page.waitForResponse((r) => r.url().includes("/xsort") && r.request().method() === "PUT"),
      page.evaluate((g) => {
        const row = document.querySelector<HTMLElement>(`[data-category-row-guid="${g}"]`);
        const btn = row?.querySelector<HTMLButtonElement>('button[aria-label="ย้ายลง"]');
        if (!btn) throw new Error("No move-down button found");
        btn.click();
      }, guid),
    ]);
    expect(res.ok(), "move-down xsort PUT").toBe(true);
  };

  expect(await orderOf(), "children start in creation order").toEqual({ child1: 1, child2: 2, child3: 3 });

  // Move A: swap child1/child2 (down-button on child1) -> child2, child1, child3.
  await clickMoveDown(child1Guid);
  expect(await orderOf(), "move A applied").toEqual({ child1: 2, child2: 1, child3: 3 });

  // Undo move A -> back to creation order.
  const [undoRes] = await Promise.all([
    page.waitForResponse((res) => res.url().includes("/xsort") && res.request().method() === "PUT"),
    clickButtonByText(page, /^เลิกทำ$/),
  ]);
  expect(undoRes.ok(), "undo xsort PUT").toBe(true);
  expect(await orderOf(), "undo reverted move A").toEqual({ child1: 1, child2: 2, child3: 3 });

  // Redo button must be enabled here (move A is replayable) before we supersede it.
  const redoEnabledBeforeMoveB = await page.evaluate(() => {
    const btn = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
      (b) => (b.textContent ?? "").trim() === "ทำซ้ำ",
    );
    return btn ? !btn.disabled : null;
  });
  expect(redoEnabledBeforeMoveB, "redo is enabled right after undo, before move B").toBe(true);

  // Move B: a DIFFERENT move (swap child2/child3 via down-button on child2) -> child1, child3, child2.
  await clickMoveDown(child2Guid);
  expect(await orderOf(), "move B applied").toEqual({ child1: 1, child2: 3, child3: 2 });

  // The redo tail (move A) must now be cleared -- the button must be disabled, not just unclicked.
  const redoDisabledAfterMoveB = await page.evaluate(() => {
    const btn = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
      (b) => (b.textContent ?? "").trim() === "ทำซ้ำ",
    );
    return btn ? btn.disabled : null;
  });
  expect(redoDisabledAfterMoveB, "redo is disabled after a new move supersedes the undone one").toBe(true);

  // Order must still reflect move B (not move A) -- fresh API read, not optimistic state.
  expect(await orderOf(), "order reflects move B, unaffected by the cleared redo tail").toEqual({
    child1: 1,
    child2: 3,
    child3: 2,
  });

  // Undo move B, back to creation order, for a clean final state.
  const [undoBRes] = await Promise.all([
    page.waitForResponse((res) => res.url().includes("/xsort") && res.request().method() === "PUT"),
    clickButtonByText(page, /^เลิกทำ$/),
  ]);
  expect(undoBRes.ok(), "undo move B xsort PUT").toBe(true);
  expect(await orderOf(), "undo move B restored creation order").toEqual({ child1: 1, child2: 2, child3: 3 });
});
