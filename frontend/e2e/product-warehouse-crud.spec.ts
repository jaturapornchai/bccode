import { test, expect, type Page } from "@playwright/test";

/**
 * Regression test for the "คลัง" (Warehouse) screen — /productwarehousescreen.
 * Backing store: MongoDB, single shared `appdb.warehouse` collection, tenant-scoped via a
 * `holdingcode` field on each document (one document per warehouse; zones/shelves are embedded
 * arrays inside it — `location: [{..., shelf: [...]}]`) — matches this project's Data Store Roles
 * rule (MongoDB = raw operational source; PostgreSQL is relational-calculation/projection only).
 *
 * **Migrated 2026-07-01 from a PostgreSQL-direct implementation** (per-tenant DB, 3 relational
 * tables) after a uat-crud-mongo sweep found the whole screen running 100% on Postgres CRUD, a live
 * violation of the Iron MongoDB Source Rule. A correct MongoDB service/repository already existed in
 * the codebase (`internal/warehouse/services/warehouse_http_service.go` +
 * `repositories/warehouse_mongo_repository.go`) but was never wired to the live `/warehouse` routes —
 * this migration rewired `warehouse_http.go` to use it (constructor/route signatures unchanged, so
 * `main.go` needed zero edits) and fixed 2 real bugs found in the process: `CreateShelf` built its
 * update in memory but never called `repo.Update(...)` (silent no-op), and `UpdateShelf` appended the
 * same shelf object twice (copy-paste bug). See `.agents/worklog.md` 2026-07-01 for the full record.
 *
 * **This exact test file was NOT modified for the migration** — it still passes unchanged, which is
 * the strongest evidence the MongoDB-backed API preserved the original PostgreSQL-backed contract
 * exactly (same request/response JSON shapes, same routes).
 *
 * CRITICAL property verified separately (via direct API calls, not by this test): `PUT /warehouse/:id`
 * distinguishes "request omitted `location`/`companyguids` entirely" (preserve existing) from "request
 * included them, even as an empty array" (replace) — a name-only warehouse edit must never wipe an
 * existing warehouse's zones/shelves. The request DTO uses pointer fields
 * (`*[]models.Location`/`*[]string`) specifically so `nil` means "field absent" vs "field present".
 *
 * Earlier PostgreSQL-era history (kept for context; no longer applicable to the current backend):
 * - `UpdateWarehouse` used to implement zone/shelf sync as a **replace-all** pattern (hard-delete +
 *   re-insert on every save) — the MongoDB version instead does a straightforward embedded-array
 *   replace, same net effect from the frontend's point of view.
 * - 3 real bugs were found and fixed on the PostgreSQL implementation before this migration:
 *   1. `formType` defaulted to "editwarehouse" instead of "createwarehouse", so a tenant with zero
 *      live warehouses could never create its first one (silent no-op save, zero error shown).
 *   2. Warehouse create/edit/delete fetched the wrong backend base (goapi instead of mainapi).
 *   3. Backend raw-SQL table/column name typos (`company_warehouses` vs the real `companywarehouses`,
 *      `warehouse_guid` vs the real `warehouseguid`, etc.) broke Update/Delete with 500s.
 *
 * DOM notes learned this run:
 * - Delete uses the browser's native `window.confirm()`, not an in-page modal (different from most
 *   other datacrud screens in this app) — override it before triggering delete.
 * - The per-row action button `title` attribute ("แก้ไข") is REUSED across nesting levels (a
 *   Location's edit button and a Shelf's edit button both have `title="แก้ไข"`). A naive "walk up N
 *   ancestor levels looking for rowText" can match the WRONG button, because a Location's DOM
 *   subtree also contains its nested Shelf rows' text. Fix: among all candidates, pick the one whose
 *   NEAREST (shallowest) matching ancestor contains the row text, not just the first DOM-order match.
 */

async function fillFieldByPlaceholder(page: Page, placeholder: string, value: string) {
  await page.evaluate(
    ({ placeholder, value }) => {
      const setValue = (el: HTMLInputElement, v: string) => {
        const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
        setter.call(el, v);
        el.dispatchEvent(new Event("input", { bubbles: true }));
        el.dispatchEvent(new Event("change", { bubbles: true }));
      };
      const input = document.querySelector<HTMLInputElement>(`input[placeholder="${placeholder}"]`);
      if (!input) throw new Error(`No input found with placeholder "${placeholder}"`);
      setValue(input, value);
    },
    { placeholder, value },
  );
}

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

/** Click the row-action button whose NEAREST matching ancestor contains rowText — see file header. */
async function clickTightestRowAction(page: Page, buttonTitle: string, rowText: string) {
  const clicked = await page.evaluate(
    ({ buttonTitle, rowText }) => {
      const buttons = [...document.querySelectorAll<HTMLButtonElement>(`button[title="${buttonTitle}"]`)].filter(
        (b) => b.offsetParent !== null,
      );
      let best: HTMLButtonElement | null = null;
      let bestDepth = Infinity;
      for (const btn of buttons) {
        let el: HTMLElement | null = btn;
        for (let depth = 0; depth < 8 && el; depth++) {
          el = el.parentElement;
          if (el && (el.innerText ?? "").includes(rowText)) {
            if (depth < bestDepth) {
              bestDepth = depth;
              best = btn;
            }
            break;
          }
        }
      }
      if (best) {
        best.click();
        return true;
      }
      return false;
    },
    { buttonTitle, rowText },
  );
  if (!clicked) throw new Error(`No "${buttonTitle}" button found in a row containing "${rowText}"`);
}

async function bodyHasText(page: Page, text: string): Promise<boolean> {
  return page.evaluate((t) => document.body.innerText.includes(t), text);
}

/**
 * Opens the "คลัง" (warehouse) tab via the left menu search. This app manages open tabs as
 * in-memory client state, not real Next.js routes (no `app/productwarehousescreen/` route file
 * exists) — a hard `page.reload()`/`page.goto()` always lands back on the default "ภาพรวม" tab,
 * even though auth/tenant/company/branch selection itself persists across the reload. Any test
 * that reloads mid-flow (e.g. to force a fresh component mount that re-fetches data) MUST call
 * this again afterward to get back to the warehouse screen — reload does not do it automatically.
 */
async function openWarehouseScreenFromMenu(page: Page) {
  // "คลัง" search collides with an unrelated field label elsewhere in the app — search by the
  // unique parent section name instead ("คลังและการผลิต"), then click the exact leaf item.
  await page.evaluate(() => {
    const input = document.querySelector<HTMLInputElement>('input[placeholder*="ค้นหาเมนู"]');
    if (!input) throw new Error("Menu search input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, "คลังและการผลิต");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await page.waitForTimeout(1000);
  await clickButtonByText(page, /^คลัง$/);
  await page.waitForTimeout(2500);
  await expect(page.locator("body")).toContainText(/จัดการคลังสินค้า|โซนเก็บสินค้า|รหัสคลังสินค้า/);

  // Delete on this screen uses native window.confirm(), not an in-page modal.
  await page.evaluate(() => {
    window.confirm = () => true;
  });
}

async function loginAndOpenWarehouseScreen(page: Page) {
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
  await openWarehouseScreenFromMenu(page);
}

test("warehouse — create/edit/delete for warehouse, location, and shelf", async ({ page }) => {
  const uid = Date.now().toString().slice(-6);
  const whCode = `E2EW${uid}`;
  const whName = `คลังE2E${uid}`;
  const locCode = `E2EL${uid}`;
  const locName = `โซนE2E${uid}`;
  const shCode = `E2ES${uid}`;
  const shName = `ชั้นE2E${uid}`;

  await loginAndOpenWarehouseScreen(page);

  // WAREHOUSE create (the Add form is open by default when no row is selected)
  await fillFieldByPlaceholder(page, "e.g. 00000", whCode);
  await fillFieldByPlaceholder(page, "เช่น คลังสินค้าหลัก", whName);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, whName)).toBe(true);

  // LOCATION create
  await clickTightestRowAction(page, "เพิ่มโซนเก็บสินค้า", whName);
  await page.waitForTimeout(1500);
  await fillFieldByPlaceholder(page, "e.g. ZONE-A", locCode);
  await fillFieldByPlaceholder(page, "เช่น โซนเอ", locName);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, locName)).toBe(true);

  // SHELF create
  await clickTightestRowAction(page, "เพิ่มชั้นวาง", locName);
  await page.waitForTimeout(1500);
  await fillFieldByPlaceholder(page, "e.g. SH-01", shCode);
  await fillFieldByPlaceholder(page, "e.g. Row A, Tier 1", shName);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, shName)).toBe(true);

  // SHELF edit
  await clickTightestRowAction(page, "แก้ไข", shName);
  await page.waitForTimeout(1500);
  await fillFieldByPlaceholder(page, "e.g. Row A, Tier 1", `${shName}X`);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, `${shName}X`)).toBe(true);

  // SHELF delete
  await clickTightestRowAction(page, "ลบชั้นวาง", `${shName}X`);
  await page.waitForTimeout(1500);
  expect(await bodyHasText(page, `${shName}X`)).toBe(false);

  // LOCATION edit
  await clickTightestRowAction(page, "แก้ไข", locName);
  await page.waitForTimeout(1500);
  await fillFieldByPlaceholder(page, "เช่น โซนเอ", `${locName}X`);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, `${locName}X`)).toBe(true);

  // LOCATION delete
  await clickTightestRowAction(page, "ลบโซนเก็บสินค้า", `${locName}X`);
  await page.waitForTimeout(1500);
  expect(await bodyHasText(page, `${locName}X`)).toBe(false);

  // WAREHOUSE edit
  await clickTightestRowAction(page, "แก้ไขคลังสินค้า", whName);
  await page.waitForTimeout(1500);
  await fillFieldByPlaceholder(page, "เช่น คลังสินค้าหลัก", `${whName}X`);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, `${whName}X`)).toBe(true);

  // WAREHOUSE delete
  await clickTightestRowAction(page, "ลบคลังสินค้า", `${whName}X`);
  await page.waitForTimeout(1500);
  expect(await bodyHasText(page, `${whName}X`)).toBe(false);
});

/**
 * API-level regression for the zone/shelf SUB-ROUTES (`/warehouse/:guid/zone[...]`) — the surface
 * the frontend does NOT use (it does whole-doc PUT), where a 2026-07-01 uat-crud-mongo run found
 * FOUR real bugs hiding behind each other after the MongoDB migration:
 *  1. POST zone panicked (nil-deref) on any warehouse stored with `location: null` — every
 *     warehouse created without a `location` key stored null. Fixed by normalizing Location to []
 *     on all service write paths.
 *  2. POST shelf always returned "document not found": `CreateShelf` looked the warehouse up with
 *     `FindWarehouseByShelf`, which FILTERS ON THE SHELF CODE — a shelf being created can never
 *     match. (Unreachable until bug 1's sibling — CreateShelf never persisting — was fixed.)
 *  3. Zone RENAME left the old-code entry behind and appended a duplicate (`UpdateLocation`
 *     searched for the NEW code it had just proven absent, instead of the old one).
 *  4. Cross-warehouse zone MOVE returned 200 but moved nothing: the HTTP handler overwrote the
 *     request's target `warehousecode` with the source warehouse's code unconditionally.
 * Uses Playwright's Node-side request context (no browser CORS) with a real token captured from a
 * UI login, hitting mainapi directly on :8888.
 */
test("warehouse zone/shelf sub-route API — create/rename/move/delete", async ({ page, request }) => {
  const uid = Date.now().toString().slice(-6);
  const MAINAPI = "http://localhost:8888";

  await loginAndOpenWarehouseScreen(page);
  const auth = JSON.parse((await page.evaluate(() => localStorage.getItem("bc_auth")))!) as { token: string };
  const H = { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" };
  const nameX = (name: string) => [{ code: "th", name, isauto: false, isdelete: false }];

  // Two warehouses (created WITHOUT a location key — bug 1's trigger shape), for the move test.
  const wa = await (await request.post(`${MAINAPI}/warehouse`, { headers: H, data: { code: `SRA${uid}`, names: nameX(`srA${uid}`) } })).json();
  const wb = await (await request.post(`${MAINAPI}/warehouse`, { headers: H, data: { code: `SRB${uid}`, names: nameX(`srB${uid}`) } })).json();
  expect(wa.success).toBe(true);
  expect(wb.success).toBe(true);

  try {
    // Bug 1: zone create on a fresh (location-was-null) warehouse must not 500/panic.
    const zoneRes = await request.post(`${MAINAPI}/warehouse/${wa.id}/zone`, {
      headers: H,
      data: { locationcode: `Z${uid}`, locationnames: nameX(`โซน${uid}`), shelf: [] },
    });
    expect(zoneRes.status()).toBe(201);

    // Bug 2: shelf create via sub-route must persist and return 201.
    const shelfRes = await request.post(`${MAINAPI}/warehouse/${wa.id}/zone/Z${uid}/shelf`, {
      headers: H,
      data: { shelfcode: `SH${uid}`, shelfname: `ชั้น${uid}` },
    });
    expect(shelfRes.status()).toBe(201);

    // Duplicate shelf code must be rejected.
    const dupRes = await request.post(`${MAINAPI}/warehouse/${wa.id}/zone/Z${uid}/shelf`, {
      headers: H,
      data: { shelfcode: `SH${uid}`, shelfname: "dup" },
    });
    expect(dupRes.ok()).toBe(false);

    // Bug 3: rename the zone — old code must vanish, no duplicate zone entry.
    const renameRes = await request.put(`${MAINAPI}/warehouse/${wa.id}/zone/Z${uid}`, {
      headers: H,
      data: { warehousecode: `SRA${uid}`, locationcode: `ZR${uid}`, locationnames: nameX(`โซนใหม่${uid}`), shelf: [{ code: `SH${uid}`, name: `ชั้น${uid}` }] },
    });
    expect(renameRes.ok()).toBe(true);
    const afterRename = await (await request.get(`${MAINAPI}/warehouse/${wa.id}/zone`, { headers: H })).json();
    const zoneCodes = (afterRename.data as { code: string }[]).map((z) => z.code);
    expect(zoneCodes).toEqual([`ZR${uid}`]);

    // Bug 4: cross-warehouse move — zone must actually leave A and appear in B.
    const moveRes = await request.put(`${MAINAPI}/warehouse/${wa.id}/zone/ZR${uid}`, {
      headers: H,
      data: { warehousecode: `SRB${uid}`, locationcode: `ZR${uid}`, locationnames: nameX(`โซนย้าย${uid}`), shelf: [{ code: `SH${uid}`, name: `ชั้น${uid}` }] },
    });
    expect(moveRes.ok()).toBe(true);
    const aZones = await (await request.get(`${MAINAPI}/warehouse/${wa.id}/zone`, { headers: H })).json();
    const bZones = await (await request.get(`${MAINAPI}/warehouse/${wb.id}/zone`, { headers: H })).json();
    expect((aZones.data as unknown[]).length).toBe(0);
    expect((bZones.data as { code: string }[]).map((z) => z.code)).toEqual([`ZR${uid}`]);

    // Shelf + zone delete on the destination.
    expect((await request.delete(`${MAINAPI}/warehouse/${wb.id}/zone/ZR${uid}/shelf/SH${uid}`, { headers: H })).ok()).toBe(true);
    expect((await request.delete(`${MAINAPI}/warehouse/${wb.id}/zone/ZR${uid}`, { headers: H })).ok()).toBe(true);
    const bAfter = await (await request.get(`${MAINAPI}/warehouse/${wb.id}/zone`, { headers: H })).json();
    expect((bAfter.data as unknown[]).length).toBe(0);
  } finally {
    // Cleanup both warehouses regardless of assertion outcomes above.
    await request.delete(`${MAINAPI}/warehouse/${wa.id}`, { headers: H });
    await request.delete(`${MAINAPI}/warehouse/${wb.id}`, { headers: H });
  }
});

/**
 * Regression for the warehouse→location company-scope feature (scopeofwork/warehouse.md):
 * a Warehouse's `companyguids` restricts which companies may use it (empty = all); a Location's
 * own `companyguids` must always be a SUBSET of its parent warehouse's list, enforced server-side
 * by `validateLocationCompanyScope` in `warehouse_http.go`. Covers:
 *  1. UI: setting warehouse-level company scope via the picker persists (API + the picker itself
 *     only offers the warehouse's allowed companies when editing a location under it).
 *  2. API: a location's companyguids that is a subset of the warehouse's is accepted and persists.
 *  3. API: a location's companyguids with a GUID outside the warehouse's list is REJECTED (400,
 *     clear message) and does not corrupt the stored document (re-GET shows unchanged state).
 * Creates a throwaway second company for the subset test (this tenant normally has only one) and
 * deletes it, plus the test warehouse, in `finally`.
 */
test("warehouse — company scope: warehouse restriction + location subset UI and API", async ({ page, request }) => {
  const uid = Date.now().toString().slice(-6);
  const MAINAPI = "http://localhost:8888";
  const nameX = (name: string) => [{ code: "th", name, isauto: false, isdelete: false }];

  // Bootstrap auth via a throwaway page load first, purely to get a bearer token for the setup API
  // calls below — the real page (with companiesList fetched fresh) opens after both companies exist,
  // so the picker's option list legitimately includes c2 and the "location picker excludes c2"
  // check below proves real filtering, not just "c2 was never fetched".
  await loginAndOpenWarehouseScreen(page);
  const auth = JSON.parse((await page.evaluate(() => localStorage.getItem("bc_auth")))!) as { token: string };
  const H = { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" };

  const companiesRes = await (await request.get(`${MAINAPI}/organization/company`, { headers: H })).json();
  const c1 = (companiesRes.data as { guidfixed: string }[])[0].guidfixed;

  const extraCompany = await (
    await request.post(`${MAINAPI}/organization/company`, {
      headers: H,
      data: { code: `SCP${uid}`, names: nameX(`scopeCo${uid}`) },
    })
  ).json();
  expect(extraCompany.success).toBe(true);
  const c2 = extraCompany.id as string;

  const whCode = `SCPWH${uid}`;
  const whName = `คลังสิทธิ์${uid}`;
  let whGuid = "";

  try {
    // Reload now that c2 exists so companiesList (fetched once on mount) includes both companies —
    // makes the later "location picker offers only 1 chip" assertion a real filtering proof. Auth/
    // tenant/company/branch selection survives the reload, but the open "คลัง" tab does not (this
    // app tracks open tabs as in-memory state, not a real route — see openWarehouseScreenFromMenu),
    // so the warehouse screen must be re-opened from the menu afterward.
    await page.reload({ waitUntil: "domcontentloaded" });
    await page.waitForTimeout(2500);
    await openWarehouseScreenFromMenu(page);

    // WAREHOUSE create restricted to c1 only, via the UI picker (not raw API) so the picker's
    // write path is actually exercised.
    await fillFieldByPlaceholder(page, "e.g. 00000", whCode);
    await fillFieldByPlaceholder(page, "เช่น คลังสินค้าหลัก", whName);
    const whChips = page.locator("text=บริษัทที่ใช้คลังนี้ได้").locator("..").locator("input[type=checkbox]");
    await expect(whChips).toHaveCount(2);
    // Check the chip whose row text matches c1's code ("00000"), not a positional guess — company
    // API response order is not a contract this test should depend on.
    await page
      .locator("text=บริษัทที่ใช้คลังนี้ได้")
      .locator("..")
      .locator("label", { hasText: "00000" })
      .locator("input[type=checkbox]")
      .check();
    await clickButtonByText(page, /^บันทึก$/);
    await page.waitForTimeout(2500);
    expect(await bodyHasText(page, whName)).toBe(true);

    const listRes = await (await request.get(`${MAINAPI}/warehouse`, { headers: H })).json();
    const created = (listRes.data as { guidfixed: string; code: string; companyguids?: string[] }[]).find(
      (w) => w.code === whCode,
    );
    expect(created).toBeTruthy();
    whGuid = created!.guidfixed;
    expect(created!.companyguids).toEqual([c1]);

    // Open the warehouse's edit form and add a location — its picker must offer ONLY c1 (the
    // warehouse's allowed company), never c2, even though c2 is a real tenant company the browser
    // session already knows about (fetched in companiesList after the reload above).
    await clickTightestRowAction(page, "เพิ่มโซนเก็บสินค้า", whName);
    await page.waitForTimeout(1500);
    expect(await bodyHasText(page, "แสดงเฉพาะบริษัทที่คลังนี้อนุญาต")).toBe(true);
    const locChips = page.locator("text=บริษัทที่ใช้โซนนี้ได้").locator("..").locator("input[type=checkbox]");
    await expect(locChips).toHaveCount(1);

    // Set the location's own scope to the same subset (c1) via the picker and save.
    const locCode = `SCPZ${uid}`;
    await fillFieldByPlaceholder(page, "e.g. ZONE-A", locCode);
    await fillFieldByPlaceholder(page, "เช่น โซนเอ", `โซนสิทธิ์${uid}`);
    await page.locator("text=บริษัทที่ใช้โซนนี้ได้").locator("..").locator("input[type=checkbox]").first().check();
    await clickButtonByText(page, /^บันทึก$/);
    await page.waitForTimeout(2500);

    const afterLoc = await (await request.get(`${MAINAPI}/warehouse/${whGuid}`, { headers: H })).json();
    const savedLoc = (afterLoc.data.location as { code: string; companyguids?: string[] }[]).find(
      (l) => l.code === locCode,
    );
    expect(savedLoc).toBeTruthy();
    expect(savedLoc!.companyguids).toEqual([c1]);

    // Backend rejection: PUT the warehouse with a location whose companyguids includes c2 (NOT in
    // the warehouse's own [c1]) — must be rejected (not 2xx) with a clear message, and must not
    // corrupt the stored document.
    const badPut = await request.put(`${MAINAPI}/warehouse/${whGuid}`, {
      headers: H,
      data: {
        code: whCode,
        names: nameX(whName),
        companyguids: [c1],
        location: [{ code: locCode, names: nameX(`โซนสิทธิ์${uid}`), companyguids: [c1, c2] }],
      },
    });
    expect(badPut.ok()).toBe(false);
    const badBody = await badPut.json();
    expect(badBody.success).toBe(false);
    expect(typeof badBody.message).toBe("string");
    expect(badBody.message.length).toBeGreaterThan(0);

    // Document must be unchanged from before the rejected attempt.
    const afterReject = await (await request.get(`${MAINAPI}/warehouse/${whGuid}`, { headers: H })).json();
    const locAfterReject = (afterReject.data.location as { code: string; companyguids?: string[] }[]).find(
      (l) => l.code === locCode,
    );
    expect(afterReject.data.companyguids).toEqual([c1]);
    expect(locAfterReject!.companyguids).toEqual([c1]);
  } finally {
    if (whGuid) await request.delete(`${MAINAPI}/warehouse/${whGuid}`, { headers: H });
    await request.delete(`${MAINAPI}/organization/company/${c2}`, { headers: H });
  }
});
