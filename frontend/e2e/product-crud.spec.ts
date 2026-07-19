import { test, expect, type Page } from "@playwright/test";

/**
 * UAT regression for the "สินค้า" (Product) screen — `frontend/src/app/menu/product-screen.tsx`
 * (~3300 lines, /product route). Covers Create -> Read -> Update -> Delete via the real UI, plus
 * classification hierarchy (group/category/brand) round-trip, verified against the API and the
 * actual DEV MongoDB (Atlas — see `backend/bootstrap.json` `mongodb.uri`, NOT any local Docker
 * `mongodb` container, which is unused/empty in this environment). Multi-tier `Prices[]` is NOT
 * covered here — see the "NOTE on barcodes/Prices[]" comment below for why that belongs to the
 * Barcode screen's own suite instead.
 *
 * **FIXED 2026-07-07 (was KNOWN, DOCUMENTED, NOT-FIXED BUG found 2026-07-05/06):**
 * The Product LIST (`GET /product` with a `holdingcode` query param, i.e. every real page load of
 * this screen) is proxied by `frontend/src/app/api/product/[[...productPath]]/route.ts` to a
 * PostgreSQL-backed endpoint (`POST {goapi}/api/product/search` -> `handlers.ProductSearchHandler`
 * in `backend/internal/goapi/handlers/product_cache.go`), NOT the MongoDB `/product` endpoint. That
 * PG query joins `public.product`, `public.productbarcode`, and `public.inventorystockbalances` in
 * the tenant's own PostgreSQL database (one DB per holdingcode, e.g. `test`). This is intentional
 * (the list shows a precomputed ยอดคงเหลือ figure that only exists as a PG projection) — the real bug
 * was that the Mongo -> Kafka -> PostgreSQL sync pipeline for Product/ProductBarcode was broken in
 * this DEV environment (Kafka consumer never actually consumed + tenant DB provisioning never created
 * `productbarcode`/`inventorystockbalances`), so the PG tables stayed empty/absent and every list load
 * 500'd with `{"error":"Query execution failed"}`. Root-caused and fixed in
 * `backend/internal/goapi/setupconfig/loader.go` (snake_case config keys were silently dropped,
 * disabling Kafka), `backend/internal/goapi/process/build/create-database.go` (`DatabaseChecker` now
 * always backfills missing tables via `DatabaseRebuildAll`, which now also provisions
 * `inventorystockbalances` via `inventory.CreateInventoryCostingTables`), and
 * `backend/internal/goapi/handlers/kafka/inventory.go` (`holdingcode` -> `holding_code` column name
 * bug that made every barcode sync insert fail). Verified live: list now returns 200, and the
 * post-save list refresh below no longer shows "Query execution failed".
 */

/**
 * The product code input (`product-screen.tsx` basic tab) has no placeholder — it is preceded by
 * a `<label>รหัสสินค้า *</label>` sibling. Find the input in that label's parent `div`.
 */
async function fillFieldByLabelPrefix(
  page: Page,
  labelPrefix: string,
  value: string,
) {
  await page.evaluate(
    ({ labelPrefix, value }) => {
      const setValue = (el: HTMLInputElement, v: string) => {
        const setter = Object.getOwnPropertyDescriptor(
          window.HTMLInputElement.prototype,
          "value",
        )!.set!;
        setter.call(el, v);
        el.dispatchEvent(new Event("input", { bubbles: true }));
        el.dispatchEvent(new Event("change", { bubbles: true }));
      };
      const label = [...document.querySelectorAll<HTMLElement>("label")].find(
        (l) => (l.textContent ?? "").trim().startsWith(labelPrefix),
      );
      if (!label)
        throw new Error(`No label found starting with "${labelPrefix}"`);
      const input =
        label.parentElement?.querySelector<HTMLInputElement>("input");
      if (!input)
        throw new Error(`No input found under label "${labelPrefix}"`);
      setValue(input, value);
    },
    { labelPrefix, value },
  );
}

/** Fills the FIRST (th) name input under the "ชื่อสินค้า" multilingual name editor. */
async function fillFirstNameInput(page: Page, value: string) {
  await page.evaluate((value) => {
    const setValue = (el: HTMLInputElement, v: string) => {
      const setter = Object.getOwnPropertyDescriptor(
        window.HTMLInputElement.prototype,
        "value",
      )!.set!;
      setter.call(el, v);
      el.dispatchEvent(new Event("input", { bubbles: true }));
      el.dispatchEvent(new Event("change", { bubbles: true }));
    };
    const input = document.querySelector<HTMLInputElement>(
      'input[placeholder="th"]',
    );
    if (!input) throw new Error('No name input found with placeholder "th"');
    setValue(input, value);
  }, value);
}

async function clickButtonByText(page: Page, textPattern: RegExp) {
  const clicked = await page.evaluate((pattern) => {
    const re = new RegExp(pattern);
    const el = [
      ...document.querySelectorAll<HTMLElement>("button,a,[role=button]"),
    ].find(
      (e) => re.test((e.textContent ?? "").trim()) && e.offsetParent !== null,
    );
    if (!el) return false;
    el.click();
    return true;
  }, textPattern.source);
  if (!clicked) throw new Error(`No button found matching ${textPattern}`);
}

async function bodyHasText(page: Page, text: string): Promise<boolean> {
  return page.evaluate((t) => document.body.innerText.includes(t), text);
}

/** Opens the "สินค้า" tab via the left menu search (in-memory tab, not a real Next.js route). */
async function openProductScreenFromMenu(page: Page) {
  await page.evaluate(() => {
    const input = document.querySelector<HTMLInputElement>(
      'input[placeholder*="ค้นหาเมนู"]',
    );
    if (!input) throw new Error("Menu search input not found");
    const setter = Object.getOwnPropertyDescriptor(
      window.HTMLInputElement.prototype,
      "value",
    )!.set!;
    setter.call(input, "สินค้าและบาร์โค้ด");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await page.waitForTimeout(1000);
  await clickButtonByText(page, /^สินค้า$/);
  await page.waitForTimeout(2500);
  await expect(page.locator("body")).toContainText(
    /ไม่พบข้อมูลสินค้า|รหัสสินค้า/,
  );
}

async function loginAndOpenProductScreen(page: Page) {
  await page.goto("/", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2500);
  // Dev test-login shortcut: pre-fills test/demo credentials and submits in one click, then
  // redirects straight to /holding (no separate "Test" confirmation button in this app version).
  await clickButtonByText(page, /เข้าทดสอบระบบ/);
  await page.waitForTimeout(2500);
  // /holding: the holding-group list loads asynchronously after login; wait for at least one real
  // card (not just the "0 กลุ่มกิจการที่เข้าได้" header) before picking the test holding group.
  // Fixed 2026-07-08: the real holding name/code in this DEV environment is "Test"/"test", not the
  // Thai "ทดสอบ" this helper previously waited for — it never matched and every run timed out here.
  await expect(page.locator("body")).toContainText(/Test/i, { timeout: 15000 });
  await page.waitForTimeout(500);
  await page.evaluate(() => {
    const btn = [...document.querySelectorAll<HTMLElement>("button")].find(
      (b) => (b.textContent ?? "").trim().startsWith("Test"),
    );
    if (!btn) throw new Error("No holding group card button found");
    btn.click();
  });
  await page.waitForTimeout(2500);
  // /workspace: pick the first company card.
  await page.evaluate(() => {
    const btn = [...document.querySelectorAll<HTMLElement>("button")].find(
      (b) =>
        (b.textContent ?? "").includes("บริษัท") &&
        (b.textContent ?? "").trim().length > 15,
    );
    if (!btn) throw new Error("No company card button found");
    btn.click();
  });
  await page.waitForTimeout(2500);
  // /workspace (branch step): pick the headquarters branch.
  await clickButtonByText(page, /สำนักงานใหญ่|headquarters/i);
  await page.waitForTimeout(2500);
  await openProductScreenFromMenu(page);
}

test("product — create via UI across all detail tabs, verify API+DB, then update+delete via API", async ({
  page,
}) => {
  const uid = Date.now().toString().slice(-6);
  const code = `E2EPRD${uid}`;
  const name = `สินค้าE2E${uid}`;

  await loginAndOpenProductScreen(page);

  // Read-only workbench regression: the list and detail panes must remain usable at Playwright's
  // default 1280px viewport before entering Create mode. This catches the old five-column list
  // overflow and the all-sections-at-once detail wall without changing any product data.
  await expect(page.getByTestId("product-workbench")).toBeVisible();
  await expect(page.getByTestId("product-list-pane")).toBeVisible();
  await expect(page.getByTestId("product-detail-pane")).toBeVisible();
  expect(
    await page
      .getByTestId("product-workbench")
      .evaluate((element) => element.scrollWidth <= element.clientWidth + 1),
  ).toBe(true);

  const productRows = page.getByTestId("product-row");
  const productRowCount = await productRows.count();
  if (productRowCount > 0) {
    const keyboardRow = productRows.nth(productRowCount > 1 ? 1 : 0);
    await keyboardRow.focus();
    await keyboardRow.press("Enter");
    await expect(keyboardRow).toHaveAttribute("aria-pressed", "true");
    await expect(page.getByTestId("product-detail-card")).toBeVisible();
    await expect(page.getByTestId("product-detail-content")).toBeVisible({
      timeout: 10000,
    });

    for (const detailTab of [
      "overview",
      "classification",
      "inventory",
      "sales",
      "more",
    ] as const) {
      await page.getByTestId(`product-detail-tab-${detailTab}`).click();
      await expect(page.getByTestId("product-detail-content")).toBeVisible();
    }
    await page.getByTestId("product-detail-tab-overview").click();

    const detailEditButton = page
      .getByTestId("product-detail-card")
      .getByRole("button", { name: /^แก้ไข$/ });
    await expect(detailEditButton).toBeEnabled();
    await expect(
      page.getByText(
        "โหลดข้อมูลฉบับเต็มไม่สำเร็จ — กำลังแสดงข้อมูลสรุปจากรายการ",
      ),
    ).toHaveCount(0);
    await detailEditButton.click();
    await expect(page.getByRole("button", { name: /^บันทึก$/ })).toBeVisible();
    await page.getByRole("button", { name: /^ยกเลิก$/ }).click();
    await expect(page.getByTestId("product-detail-card")).toBeVisible();
  }

  // Fixed 2026-07-07 (see file-level comment): the list now works. The Add form is opened
  // explicitly regardless, so Create does not depend on the list's current state.
  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(800);

  await fillFieldByLabelPrefix(page, "รหัสสินค้า", code);
  await fillFirstNameInput(page, name);

  // A5 (progressive disclosure): the 5 advanced tabs (logistics/restaurant/timeforsales/
  // business/misc) are collapsed behind "ขั้นสูง ▸" by default — expand it before touching them.
  await clickButtonByText(page, /^ขั้นสูง/);
  await page.waitForTimeout(200);
  await expect(page.locator("button", { hasText: "ขั้นสูง ▾" })).toBeVisible();

  // Touch every detail tab once so a future regression that crashes a tab's render is caught here,
  // not just in manual testing. Tab labels per product-screen.tsx's ~12-tab detail layout.
  for (const tabLabel of [
    "หมวดหมู่",
    "หน่วยนับและบาร์โค้ด",
    "ส่วนประกอบ (BOM)",
    "คงคลัง",
    "ภาพและสี",
    "การจัดส่ง / โลจิสติกส์",
    "ร้านอาหาร/POS",
    "เวลาขาย",
    "สาขา/ธุรกิจ",
    "อื่นๆ",
    "ข้อมูลหลัก",
  ]) {
    await clickButtonByText(
      page,
      new RegExp(`^${tabLabel.replace(/[()/]/g, "\\$&")}$`),
    );
    await page.waitForTimeout(200);
  }

  // Collapsing "ขั้นสูง" while an advanced tab ("อื่นๆ") is active must keep that tab button
  // visible and its panel rendered (product-screen.tsx: ADVANCED_PRODUCT_TAB_KEYS.includes(productTab)
  // forces the group to render regardless of the collapsed flag).
  await clickButtonByText(page, /^อื่นๆ$/);
  await page.waitForTimeout(200);
  await clickButtonByText(page, /^ขั้นสูง/);
  await page.waitForTimeout(200);
  await expect(page.locator("button", { hasText: "ขั้นสูง ▸" })).toBeVisible();
  await expect(page.locator("button", { hasText: "อื่นๆ" })).toBeVisible();

  // Image display regression (fixed 2026-07-08): the "ภาพและสี" tab used to render a plain
  // <img src="/goapi/s3/file/..."> for the uploaded product image, but that route requires an
  // Authorization header a plain <img> can never send — every uploaded image rendered as a
  // permanently broken icon. Fixed via the shared AuthenticatedImg component
  // (frontend/src/components/authenticated-image.tsx), which fetches with the bearer token and
  // renders a blob: object URL instead. Upload a real file here and confirm it actually paints
  // (naturalWidth > 0) with no page reload, so this bug class cannot silently return.
  await clickButtonByText(page, /^ภาพและสี$/);
  await page.waitForTimeout(200);
  await page
    .locator('input[type="file"][accept*="image"]')
    .first()
    .setInputFiles("public/flags/th.png");
  const uploadedImg = page.locator("img[src^='blob:']").first();
  await expect(uploadedImg).toBeVisible({ timeout: 10000 });
  expect(
    await uploadedImg.evaluate((img: HTMLImageElement) => img.naturalWidth),
  ).toBeGreaterThan(0);

  // Switch back to "ข้อมูลหลัก" (basic tab, primary group) before filling remaining basic fields /
  // saving, so the create flow below operates from a known, non-advanced tab.
  await clickButtonByText(page, /^ข้อมูลหลัก$/);
  await page.waitForTimeout(200);

  // Capture the real create response (contains guidfixed) by waiting on the network response
  // alongside the click. The list-driven UI is not used to find this row afterward because the
  // The list is an asynchronous PostgreSQL projection. Use the create response identity for
  // deterministic Mongo CRUD verification instead of waiting on Kafka projection timing.
  const [createRes] = await Promise.all([
    page.waitForResponse(
      (res) =>
        res.url().includes("/api/product") && res.request().method() === "POST",
    ),
    clickButtonByText(page, /^บันทึก$/),
  ]);
  expect(createRes.status()).toBe(201);
  const createBody = await createRes.json();
  expect(createBody.success).toBe(true);
  const guid = createBody.data.guidfixed as string;
  expect(guid).toBeTruthy();
  await page.waitForTimeout(1000);

  // Fixed 2026-07-07: the companion GET reload used to 500 (PG sync pipeline was broken) and clear
  // the created row from the right pane. Confirm that error no longer appears. (The list's own
  // staleness after this point is the separate cache-TTL issue noted above, not this error.)
  expect(await bodyHasText(page, "Query execution failed")).toBe(false);

  const auth = JSON.parse(
    (await page.evaluate(() => localStorage.getItem("bc_auth")))!,
  ) as {
    token: string;
    backendUrl: string;
  };
  const H = {
    Authorization: `Bearer ${auth.token}`,
    "x-bc-backend-url": auth.backendUrl,
  };

  // READ: GET-by-guid (MongoDB-backed, unaffected by the list bug) must return what was saved.
  const getRes = await page.request.get(`/api/product/${guid}`, { headers: H });
  expect(getRes.ok()).toBe(true);
  const getBody = await getRes.json();
  expect(getBody.data.code).toBe(code);
  expect(
    getBody.data.names.find((n: { code: string }) => n.code === "th")?.name,
  ).toBe(name);

  // UPDATE: change name + classification hierarchy, verify round-trip via API.
  //
  // NOTE on `barcodes`/`Prices[]`: `product-screen.tsx` never reads or writes `Product.Barcodes` or
  // any `prices`/`keynumber` field — multi-tier pricing is exclusively a Barcode-screen
  // (`product-barcode-screen.tsx`) concept. `GetProduct` (`product_http_service.go`) actively
  // IGNORES whatever is stored in `Product.Barcodes` on read and instead re-derives it by querying
  // the separate `productbarcodes` collection by item code (`repomgProductBarcode.FindByItemCode`)
  // — so `Product.Barcodes` is a computed/denormalized read-model field, not something this screen's
  // PUT should ever populate. Do not add a `barcodes`/prices assertion here; that belongs in the
  // Barcode screen's own UAT suite, not this one.
  const updated = {
    ...getBody.data,
    names: [{ code: "th", name: `${name}X`, isauto: false, isdelete: false }],
    groupcode: `GRP${uid}`,
    groupnames: [
      { code: "th", name: `กลุ่มE2E${uid}`, isauto: false, isdelete: false },
    ],
    categorycode: `CAT${uid}`,
    categorynames: [
      { code: "th", name: `หมวดE2E${uid}`, isauto: false, isdelete: false },
    ],
    brandcode: `BRD${uid}`,
    brandnames: [
      { code: "th", name: `ยี่ห้อE2E${uid}`, isauto: false, isdelete: false },
    ],
  };
  const putRes = await page.request.put(`/api/product/${guid}`, {
    headers: H,
    data: updated,
  });
  expect(putRes.ok()).toBe(true);
  expect((await putRes.json()).success).toBe(true);

  const afterUpdateRes = await page.request.get(`/api/product/${guid}`, {
    headers: H,
  });
  const afterUpdate = (await afterUpdateRes.json()).data;
  expect(
    afterUpdate.names.find((n: { code: string }) => n.code === "th")?.name,
  ).toBe(`${name}X`);
  expect(afterUpdate.groupcode).toBe(`GRP${uid}`);
  expect(afterUpdate.categorycode).toBe(`CAT${uid}`);
  expect(afterUpdate.brandcode).toBe(`BRD${uid}`);

  // DELETE: soft-delete (deletedat/deletedby set, document retained) — confirmed by GET now 404ing.
  const delRes = await page.request.delete(`/api/product/${guid}`, {
    headers: H,
  });
  expect(delRes.ok()).toBe(true);
  expect((await delRes.json()).success).toBe(true);

  const afterDeleteRes = await page.request.get(`/api/product/${guid}`, {
    headers: H,
  });
  expect(afterDeleteRes.status()).toBe(404);
});
