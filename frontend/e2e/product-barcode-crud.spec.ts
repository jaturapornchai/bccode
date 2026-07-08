import { test, expect, type Page } from "@playwright/test";

/**
 * Regression test for the "บาร์โค้ด" (Product Barcode) screen — /productbarcode.
 *
 * Frontend: `frontend/src/app/menu/product-barcode-screen.tsx` (list/detail shell) +
 * `frontend/src/components/product-barcode/barcode-form.tsx` (the actual editor form, tabs:
 * ข้อมูลหลัก / ราคา / ภาพและสี / ขนาด / Marketplace).
 * Backend: `backend/internal/product/productbarcode/{productbarcode_http.go,services/productbarcode_http_service.go}`.
 * Collection: `productbarcodes`. Routes under `/product/barcode/*`.
 *
 * **Bug found + fixed this run (2026-07-06):** the unit picker used by the barcode form's required
 * "หน่วยนับ" field (`MasterPicker` with `master="unit"`, via `/api/product-barcode/master/unit` ->
 * backend `GET /unit?companyguid=...`) returned ZERO results for any company, even though the
 * standalone หน่วยนับ (Unit) master screen showed 109 seeded units. Root cause: every seeded/
 * auto-created Unit document has `companyguids: null` (meaning "available to all companies"), but
 * `UnitHttpService.SearchUnit`/`SearchUnitLimit` (`backend/internal/product/unit/services/
 * unit_http_service.go`) built the company-scope filter as
 * `$or: [{companyguids: {$exists:false}}, {companyguids: {$size:0}}, {companyguids: companyGuid}]`
 * — none of those three branches match a field that is explicitly `null` (MongoDB: `$exists:false`
 * requires the key to be absent, `$size:0` requires an actual empty array; `null` is neither). Fixed
 * by adding a `{companyguids: nil}` branch to both filters. This was a hard blocker: without a unit,
 * the required "หน่วยนับ" field could never be filled and no barcode could ever be created for any
 * company in this environment — this test locks in the fix.
 *
 * Login/menu-navigation helpers follow the current convention from `trade-partners-crud.spec.ts` /
 * `product-set-crud.spec.ts` (retry-with-deadline instead of fixed waits; holding match excludes
 * "bctest0N" via `test(?!\d)`; dismiss the one-time "ยังไม่มีหน่วยนับสินค้า" modal on freshly-seeded
 * companies).
 */

const MAINAPI = "http://localhost:8888";

async function setInputValue(page: Page, input: unknown, value: string) {
  await page.evaluate(
    ({ input, value }) => {
      const el = input as HTMLInputElement;
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
      setter.call(el, value);
      el.dispatchEvent(new Event("input", { bubbles: true }));
      el.dispatchEvent(new Event("change", { bubbles: true }));
    },
    { input, value },
  );
}

async function fillFieldNearLabel(page: Page, labelPattern: RegExp, value: string) {
  await page.evaluate(
    ({ pattern, value }) => {
      const re = new RegExp(pattern);
      const setValue = (el: HTMLInputElement, v: string) => {
        const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
        setter.call(el, v);
        el.dispatchEvent(new Event("input", { bubbles: true }));
        el.dispatchEvent(new Event("change", { bubbles: true }));
      };
      const input = [...document.querySelectorAll<HTMLInputElement>("input")].find((i) =>
        re.test(i.closest("label")?.innerText ?? ""),
      );
      if (!input) throw new Error(`No input found near label matching ${pattern}`);
      setValue(input, value);
    },
    { pattern: labelPattern.source, value },
  );
}

/** barcode-form.tsx's `FieldRow` renders the label as a plain sibling <span>, not a wrapping
 *  <label> (unlike NamesEditor) — `input.closest("label")` never matches for these fields. Walk up
 *  from the label span to its FieldRow container div and find the first <input> inside it instead.
 *  Requires the candidate span's own parent to actually contain an input (rules out unrelated
 *  same-text spans elsewhere on the page, e.g. the left-nav menu item also named "บาร์โค้ด"). */
async function fillFieldRowByLabel(page: Page, labelPattern: RegExp, value: string) {
  await page.evaluate(
    ({ pattern, value }) => {
      const re = new RegExp(pattern);
      const setValue = (el: HTMLInputElement, v: string) => {
        const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
        setter.call(el, v);
        el.dispatchEvent(new Event("input", { bubbles: true }));
        el.dispatchEvent(new Event("change", { bubbles: true }));
      };
      const candidates = [...document.querySelectorAll<HTMLElement>("span")].filter((s) => re.test(s.innerText ?? ""));
      const labelSpan = candidates.find((s) => s.parentElement?.querySelector("input"));
      if (!labelSpan) throw new Error(`No FieldRow label span with an input found matching ${pattern}`);
      const input = labelSpan.parentElement!.querySelector<HTMLInputElement>("input")!;
      setValue(input, value);
    },
    { pattern: labelPattern.source, value },
  );
}

async function clickButtonByText(page: Page, textPattern: RegExp) {
  const deadline = Date.now() + 8000;
  for (;;) {
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
    if (clicked) return;
    if (Date.now() > deadline) throw new Error(`No clickable element found matching ${textPattern}`);
    await page.waitForTimeout(300);
  }
}

async function clickCompanyCard(page: Page) {
  const deadline = Date.now() + 8000;
  for (;;) {
    const clicked = await page.evaluate(() => {
      const btn = [...document.querySelectorAll<HTMLElement>("button")].find(
        (b) => (b.textContent ?? "").includes("บริษัท") && (b.textContent ?? "").trim().length > 15,
      );
      if (btn) {
        btn.click();
        return true;
      }
      return false;
    });
    if (clicked) return;
    if (Date.now() > deadline) throw new Error("No company card button found");
    await page.waitForTimeout(300);
  }
}

async function bodyHasText(page: Page, text: string): Promise<boolean> {
  return page.evaluate((t) => document.body.innerText.includes(t), text);
}

async function searchMenuAndOpen(page: Page, searchTerm: string, exactItemPattern: RegExp) {
  const deadline = Date.now() + 8000;
  for (;;) {
    const found = await page.evaluate((term) => {
      const input = document.querySelector<HTMLInputElement>('input[placeholder*="ค้นหาเมนู"]');
      if (!input) return false;
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
      setter.call(input, term);
      input.dispatchEvent(new Event("input", { bubbles: true }));
      return true;
    }, searchTerm);
    if (found) break;
    if (Date.now() > deadline) throw new Error("Menu search input not found");
    await page.waitForTimeout(300);
  }
  await page.waitForTimeout(1000);
  await clickButtonByText(page, exactItemPattern);
  await page.waitForTimeout(2000);
}

async function loginOnly(page: Page) {
  await page.goto("/", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /เข้าทดสอบระบบ/);
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /test(?!\d)/i);
  await page.waitForTimeout(2500);
  await clickCompanyCard(page);
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /สำนักงานใหญ่|headquarters/i);
  await page.waitForTimeout(2500);
  const hasProductUnitModal = await bodyHasText(page, "ยังไม่มีหน่วยนับสินค้า");
  if (hasProductUnitModal) {
    await clickButtonByText(page, /^เข้าเมนูก่อน$/);
    await page.waitForTimeout(1500);
  }
}

async function loginAndOpenBarcodeScreen(page: Page) {
  await loginOnly(page);
  await searchMenuAndOpen(page, "บาร์โค้ด", /^บาร์โค้ด$/);
  await expect(page.locator("body")).toContainText(/จัดการบาร์โค้ด/);
}

/** Open the unit picker on the barcode form and select a unit by its code (e.g. "BOX"). Regression
 *  guard for the companyguids:null filter bug — this must find and select a real unit, not "ไม่พบ
 *  ข้อมูล" (no data found). The MasterField trigger button has `aria-label={label}` ("หน่วยนับ" for
 *  the unit field, see barcode-form.tsx `MasterField`) — much more robust than text/placeholder
 *  matching since the button's visible text is just a "—" placeholder before a unit is picked. */
async function pickUnitByCode(page: Page, unitCode: string) {
  const opened = await page.evaluate(() => {
    const trigger = document.querySelector<HTMLButtonElement>('button[aria-label="หน่วยนับ"]');
    if (trigger) {
      trigger.click();
      return true;
    }
    return false;
  });
  expect(opened).toBe(true);
  await page.waitForTimeout(600);

  // The picker modal must show real unit rows, not the empty state — this is the exact bug surface
  // (the underlying barcode LIST also has its own "ไม่พบข้อมูลบาร์โค้ด" empty state elsewhere on the
  // page, so check specifically inside the picker's result list, not document.body as a whole).
  const pickerHasResults = await page.evaluate(() => {
    const rows = document.querySelectorAll("li button");
    return rows.length > 0;
  });
  expect(pickerHasResults).toBe(true);

  // Picker rows are `<li><button onClick=...>` (master-picker.tsx) — the click handler lives on the
  // inner <button>, not the <li>; must click the actual button element for the React handler to fire.
  const picked = await page.evaluate((code) => {
    const row = [...document.querySelectorAll<HTMLButtonElement>("li button")].find(
      (r) => (r.textContent ?? "").includes(code) && r.offsetParent !== null,
    );
    if (row) {
      row.click();
      return true;
    }
    return false;
  }, unitCode);
  expect(picked).toBe(true);
}

test("barcode — create requires a unit (regression: companyguids:null picker bug), duplicate barcode rejected, multi-tier price + price-history, update, delete", async ({
  page,
  request,
}) => {
  const uid = Date.now().toString().slice(-6);
  const barcode = `E2EBC${uid}`;
  const name = `สินค้าUATบาร์โค้ด${uid}`;

  await loginAndOpenBarcodeScreen(page);
  const auth = JSON.parse((await page.evaluate(() => localStorage.getItem("bc_auth")))!) as { token: string };
  const H = { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" };

  let createdGuid = "";

  try {
    // CREATE via UI — barcode, main product name, and a real unit picked from the (previously
    // empty, now fixed) unit picker.
    await clickButtonByText(page, /^เพิ่ม$/);
    await page.waitForTimeout(1000);
    // "บาร์โค้ด" is a FieldRow field (plain sibling <span> label, no wrapping <label>).
    await fillFieldRowByLabel(page, /^บาร์โค้ด/, barcode);
    // The product-name field is rendered by NamesEditor (see names-editor.tsx): the outer "ชื่อสินค้า"
    // text is a heading, not a <label> — the actual per-language <label> reads "ภาษาแรก" / "TH" for
    // the primary language input (same pattern as trade-partners-crud.spec.ts).
    await fillFieldNearLabel(page, /ภาษาแรก|TH/, name);
    await pickUnitByCode(page, "BOX");
    await page.waitForTimeout(500);

    // Image display regression (fixed 2026-07-08, same fix as product-crud.spec.ts): the "ภาพและสี"
    // tab used to render a plain <img src="/goapi/s3/file/..."> for the uploaded barcode image, which
    // a browser can never attach an Authorization header to, so every uploaded image rendered as a
    // permanently broken icon. Fixed via the shared AuthenticatedImg component
    // (frontend/src/components/authenticated-image.tsx). Upload a real file and confirm it actually
    // paints (naturalWidth > 0) with no page reload. NOTE: this only covers the live in-session
    // display mechanism — whether the uploaded imageuri survives the barcode create/update save is a
    // separate, pre-existing, still-open backend bug (imageuri is dropped somewhere in the
    // save/proxy/backend chain for productbarcodes; verified 2026-07-08 that zero records in the
    // `productbarcodes` collection have ever had a non-empty imageuri) — not covered here.
    await clickButtonByText(page, /^ภาพและสี$/);
    await page.waitForTimeout(200);
    await page
      .locator('input[type="file"][accept*="image"]')
      .first()
      .setInputFiles("public/flags/th.png");
    const uploadedBarcodeImg = page.locator("img[src^='blob:']").first();
    await expect(uploadedBarcodeImg).toBeVisible({ timeout: 10000 });
    expect(await uploadedBarcodeImg.evaluate((img: HTMLImageElement) => img.naturalWidth)).toBeGreaterThan(0);
    await clickButtonByText(page, /^ข้อมูลหลัก$/);
    await page.waitForTimeout(200);

    // Multi-tier price: set tier 1 to 199, add tier 2 and set it to 179.
    await clickButtonByText(page, /^ราคา$/);
    await page.waitForTimeout(400);
    const tier1 = await page.evaluateHandle(() =>
      [...document.querySelectorAll<HTMLInputElement>("input")].find((i) => i.value === "0"),
    );
    await setInputValue(page, tier1, "199");
    await clickButtonByText(page, /^เพิ่มระดับราคา$/);
    await page.waitForTimeout(400);
    const tier2 = await page.evaluateHandle(() =>
      [...document.querySelectorAll<HTMLInputElement>("input")].find((i) => i.value === "0"),
    );
    await setInputValue(page, tier2, "179");

    await clickButtonByText(page, /^บันทึก$/);
    await page.waitForTimeout(2500);
    expect(await bodyHasText(page, "บันทึกบาร์โค้ดแล้ว")).toBe(true);

    // Verify via the real backend search API (MongoDB-backed, same service the screen's list uses)
    // — not just the UI echo. Uses MAINAPI directly like the other e2e specs (the Next.js
    // `/api/product-barcode/list` proxy forwards to `serverGoApiBase()`, which defaults to the
    // remote 192.168.2.202 dev server unless `BCAI_LOCAL_BACKEND_URL` is set — not reliable for a
    // deterministic local test run).
    const listRes = await (await request.get(`${MAINAPI}/product/barcode?q=${barcode}&limit=5`, { headers: H })).json();
    const created = (
      listRes.data as {
        guidfixed: string;
        barcode: string;
        itemunitcode: string;
        prices?: { keynumber: number; price: number }[];
      }[]
    ).find((b) => b.barcode === barcode);
    expect(created).toBeTruthy();
    createdGuid = created!.guidfixed;
    expect(created!.itemunitcode).toBe("BOX");
    expect(created!.prices).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ keynumber: 1, price: 199 }),
        expect.objectContaining({ keynumber: 2, price: 179 }),
      ]),
    );

    // Duplicate barcode must be rejected cleanly (400, no corruption) — server-side uniqueness
    // check in `CreateProductBarcode` (`services/productbarcode_http_service.go`).
    const dupCreate = await request.post(`${MAINAPI}/product/barcode`, {
      headers: H,
      data: {
        barcode,
        names: [{ code: "th", name: "ซ้ำ" }],
        itemunitcode: "BOX",
        itemunitnames: [{ code: "th", name: "กล่อง" }],
        prices: [{ keynumber: 1, price: 1 }],
      },
    });
    expect(dupCreate.status()).toBe(400);
    const dupBody = await dupCreate.json();
    expect(dupBody.success).toBe(false);

    // Price history recorded 2 "create" entries (one per tier), tied to this barcode.
    const historyAfterCreate = await (
      await request.get(`${MAINAPI}/product/barcode/price-history/${barcode}?page=1&limit=20`, { headers: H })
    ).json();
    expect(
      (historyAfterCreate.data as { keynumber: number; newprice: number; action: string }[]).filter(
        (h) => h.action === "create",
      ).length,
    ).toBe(2);

    // UPDATE via UI: reopen, bump tier-1 price, save — confirm the DB reflects the new price and a
    // fresh price-history "update" row was appended with the correct old/new price.
    await clickButtonByText(page, new RegExp(barcode));
    await page.waitForTimeout(1200);
    await clickButtonByText(page, /^แก้ไข$/);
    await page.waitForTimeout(1000);
    await clickButtonByText(page, /^ราคา$/);
    await page.waitForTimeout(400);
    const tier1Edit = await page.evaluateHandle(() =>
      [...document.querySelectorAll<HTMLInputElement>("input")].find((i) => i.value === "199"),
    );
    await setInputValue(page, tier1Edit, "249");
    await clickButtonByText(page, /^บันทึก$/);
    await page.waitForTimeout(2000);
    expect(await bodyHasText(page, "บันทึกบาร์โค้ดแล้ว")).toBe(true);

    const listAfterUpdate = await (
      await request.get(`${MAINAPI}/product/barcode?q=${barcode}&limit=5`, { headers: H })
    ).json();
    const afterUpdate = (listAfterUpdate.data as { barcode: string; prices?: { keynumber: number; price: number }[] }[]).find(
      (b) => b.barcode === barcode,
    );
    expect(afterUpdate!.prices).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ keynumber: 1, price: 249 }),
        expect.objectContaining({ keynumber: 2, price: 179 }),
      ]),
    );

    const historyAfterUpdate = await (
      await request.get(`${MAINAPI}/product/barcode/price-history/${barcode}?page=1&limit=20`, { headers: H })
    ).json();
    const updateEntry = (
      historyAfterUpdate.data as { keynumber: number; oldprice: number; newprice: number; action: string }[]
    ).find((h) => h.action === "update");
    expect(updateEntry).toBeTruthy();
    expect(updateEntry!.keynumber).toBe(1);
    expect(updateEntry!.oldprice).toBe(199);
    expect(updateEntry!.newprice).toBe(249);

    // DELETE via UI — confirm modal, then verify the record is actually gone (fresh list + direct
    // GET by guid returns not-found), not just removed from the on-screen list. The detail panel's
    // own "ลบ" trigger button and the confirm-dialog's "ลบ" button share the same text — a bare
    // text match on the SECOND click can re-hit the first-in-DOM-order trigger button behind the
    // overlay instead of the dialog's own confirm button, so scope the second click to a button
    // that is a descendant of the confirm-dialog container (identified by its heading text).
    await clickButtonByText(page, /^ลบ$/);
    await page.waitForTimeout(800);
    expect(await bodyHasText(page, "ต้องการลบบาร์โค้ดจริงหรือไม่")).toBe(true);
    const confirmedDelete = await page.evaluate(() => {
      const heading = [...document.querySelectorAll<HTMLElement>("h1,h2,h3,[role=heading]")].find((h) =>
        (h.textContent ?? "").includes("ต้องการลบบาร์โค้ดจริงหรือไม่"),
      );
      const dialogRoot = heading?.closest('[role="dialog"], [role="alertdialog"]') ?? heading?.parentElement?.parentElement;
      const btn = [...(dialogRoot ?? document).querySelectorAll<HTMLElement>("button")].find(
        (b) => (b.textContent ?? "").trim() === "ลบ",
      );
      if (btn) {
        btn.click();
        return true;
      }
      return false;
    });
    expect(confirmedDelete).toBe(true);
    await expect(page.locator("body")).toContainText("ลบบาร์โค้ดแล้ว", { timeout: 8000 });
    createdGuid = ""; // deleted via UI; skip the finally-block cleanup delete

    const afterDeleteList = await (
      await request.get(`${MAINAPI}/product/barcode?q=${barcode}&limit=5`, { headers: H })
    ).json();
    expect((afterDeleteList.data as unknown[] | undefined)?.length ?? 0).toBe(0);

    const afterDeleteGet = await request.get(`${MAINAPI}/product/barcode/${created!.guidfixed}`, { headers: H });
    expect([400, 404]).toContain(afterDeleteGet.status());
  } finally {
    if (createdGuid) await request.delete(`${MAINAPI}/product/barcode/${createdGuid}`, { headers: H }).catch(() => {});
  }
});
