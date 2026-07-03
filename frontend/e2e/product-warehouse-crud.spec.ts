import { test, expect, type Page } from "@playwright/test";

/**
 * Regression test for the "คลัง" (Warehouse) screen — /productwarehousescreen.
 *
 * **Rewritten 2026-07-03** for the warehouse -> warehouselocation -> warehousebin master-data
 * redesign (`.agents/worklog.md`), which replaced the old single `warehouse` document with
 * embedded `location[].shelf[]` arrays with THREE separate MongoDB collections and a
 * `GET /warehouse/tree` read model:
 *  - `warehouse` (`Warehouse{Code, Names, Latitude, Longitude, CompanyGuids, Status}`)
 *  - `warehouselocation` ("ที่เก็บสินค้า", references warehouse by guid, own `CompanyGuids` that
 *    must be a subset of its parent warehouse's)
 *  - `warehousebin` ("ที่วางสินค้า", references location by guid, no company scope — inherits)
 *
 * New routes: `POST/GET/PUT/DELETE /warehouse`, `.../warehouse/tree`,
 * `.../warehouse/:warehouseguid/location[...]`, `.../warehouse/:warehouseguid/location/:locationguid/bin[...]`.
 *
 * The frontend tree UI (`warehouse-tree-view.tsx`) self-fetches `GET /warehouse/tree` on mount and
 * renders a 3-level collapsible tree (warehouse -> location -> bin) with an inline right-pane form.
 * Delete goes through the app's shared `useConfirmDialog()` — a custom in-page modal whose confirm
 * button label is "ลบ" (passed as `confirmLabel` on every delete call in this file) — NOT
 * `window.confirm()` (that was true for the old shelf-based UI, no longer applies here).
 *
 * DOM notes learned this run:
 * - The per-row action button `title` attribute "แก้ไข" is REUSED across nesting levels (both a
 *   Location's edit button and a Bin's edit button use `title="แก้ไข"` — only Warehouse edit has
 *   its own distinct title "แก้ไขคลังสินค้า"). Reuses the existing `clickTightestRowAction` fix from
 *   the prior suite: among all candidates, pick the one whose NEAREST matching ancestor contains
 *   the row text, not just the first DOM-order match.
 * - Row click alone selects a node into the read-only-looking right pane (same form, same values);
 *   the explicit pencil icon is what this suite uses to enter edit mode, matching the Data List
 *   Selection vs Editing Rule.
 *
 * A real bug was found and fixed while writing this suite (see `.agents/worklog.md` 2026-07-03):
 * `UpdateWarehouse` / `UpdateLocation` / `UpdateBin` HTTP handlers in `warehouse_http.go` blanket-
 * returned `500 Internal Server Error` for ANY service-layer error, including client-side validation
 * failures like the location -> warehouse company-scope subset check — inconsistent with the sibling
 * `Create*` handlers, which correctly return `400`. Fixed by changing all three `Update*` handlers'
 * error status from `http.StatusInternalServerError` to `http.StatusBadRequest`, matching `Create*`.
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

/**
 * The multilingual name field (`NamesEditor`) has NO static placeholder text — each language input's
 * placeholder is just its language code (e.g. `placeholder="th"`), and the same label
 * ("ชื่อคลังสินค้าหลายภาษา" / "ชื่อที่เก็บสินค้าหลายภาษา") only exists once at a time in the DOM (the
 * right pane renders exactly one active form). Fill the FIRST (primary/th) input under that label.
 * Uses the label div's OWN direct text (not descendant text, which would match many ancestor divs)
 * to find the exact `<div class="text-sm font-medium">{label}</div>` node, then looks for the first
 * `input` among its following siblings inside the shared `space-y-2` wrapper.
 */
async function fillFirstNameByLabel(page: Page, labelText: string, value: string) {
  await page.evaluate(
    ({ labelText, value }) => {
      const setValue = (el: HTMLInputElement, v: string) => {
        const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
        setter.call(el, v);
        el.dispatchEvent(new Event("input", { bubbles: true }));
        el.dispatchEvent(new Event("change", { bubbles: true }));
      };
      const ownTextOnly = (el: Element) =>
        [...el.childNodes]
          .filter((n) => n.nodeType === Node.TEXT_NODE)
          .map((n) => n.textContent ?? "")
          .join("")
          .trim();
      const labelEl = [...document.querySelectorAll<HTMLElement>("div")].find((e) =>
        ownTextOnly(e).startsWith(labelText),
      );
      if (!labelEl) throw new Error(`No label found with text "${labelText}"`);
      const wrapper = labelEl.parentElement;
      const input = wrapper?.querySelector<HTMLInputElement>("input");
      if (!input) throw new Error(`No name input found under label "${labelText}"`);
      setValue(input, value);
    },
    { labelText, value },
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
  await expect(page.locator("body")).toContainText(/รหัสคลังสินค้า|ไม่พบข้อมูลคลังสินค้า/);
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

test("warehouse — create/edit/delete for warehouse, location, and bin", async ({ page }) => {
  const uid = Date.now().toString().slice(-6);
  const whCode = `E2EW${uid}`;
  const whName = `คลังE2E${uid}`;
  const locCode = `E2EL${uid}`;
  const locName = `ที่เก็บE2E${uid}`;
  const binCode = `E2EB${uid}`;
  const binName = `ที่วางE2E${uid}`;

  await loginAndOpenWarehouseScreen(page);

  // WAREHOUSE create (the Add form is open by default when no row is selected) — also select 2
  // business types (scopeofwork/warehouse.md "คุณสมบัติ") to cover the businesstypes multi-select,
  // which has no other regression coverage.
  await fillFieldByPlaceholder(page, "e.g. 00000", whCode);
  await fillFirstNameByLabel(page, "ชื่อคลังสินค้าหลายภาษา", whName);
  await page.evaluate(() => {
    const check = (text: string) => {
      const label = [...document.querySelectorAll("label")].find((l) => l.textContent?.trim() === text);
      (label?.querySelector("input[type=checkbox]") as HTMLInputElement | null)?.click();
    };
    check("ฝากขาย");
    check("Drop Ship");
  });
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, whName)).toBe(true);

  // LOCATION create
  await clickTightestRowAction(page, "เพิ่มที่เก็บสินค้า", whName);
  await page.waitForTimeout(1500);
  await fillFieldByPlaceholder(page, "e.g. ZONE-A", locCode);
  await fillFirstNameByLabel(page, "ชื่อที่เก็บสินค้าหลายภาษา", locName);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, locName)).toBe(true);

  // BIN create
  await clickTightestRowAction(page, "เพิ่มที่วางสินค้า", locName);
  await page.waitForTimeout(1500);
  await fillFieldByPlaceholder(page, "e.g. BIN-01", binCode);
  await fillFieldByPlaceholder(page, "e.g. Row A, Tier 1", binName);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, binName)).toBe(true);

  // BIN edit
  await clickTightestRowAction(page, "แก้ไข", binName);
  await page.waitForTimeout(1500);
  await fillFieldByPlaceholder(page, "e.g. Row A, Tier 1", `${binName}X`);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, `${binName}X`)).toBe(true);

  // BIN delete (custom confirm dialog, not window.confirm — confirm button text is "ลบ")
  await clickTightestRowAction(page, "ลบที่วางสินค้า", `${binName}X`);
  await page.waitForTimeout(800);
  await clickButtonByText(page, /^ลบ$/);
  await page.waitForTimeout(1500);
  expect(await bodyHasText(page, `${binName}X`)).toBe(false);

  // LOCATION edit
  await clickTightestRowAction(page, "แก้ไข", locName);
  await page.waitForTimeout(1500);
  await fillFirstNameByLabel(page, "ชื่อที่เก็บสินค้าหลายภาษา", `${locName}X`);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, `${locName}X`)).toBe(true);

  // LOCATION delete
  await clickTightestRowAction(page, "ลบที่เก็บสินค้า", `${locName}X`);
  await page.waitForTimeout(800);
  await clickButtonByText(page, /^ลบ$/);
  await page.waitForTimeout(1500);
  expect(await bodyHasText(page, `${locName}X`)).toBe(false);

  // WAREHOUSE edit — only changes the name, so this also proves businesstypes survives an edit
  // cycle (the fetch-then-merge partial-update fix; see .agents/worklog.md 2026-07-03 for the bug
  // this class of full-document-overwrite would otherwise cause).
  await clickTightestRowAction(page, "แก้ไขคลังสินค้า", whName);
  await page.waitForTimeout(1500);
  const checkedBusinessTypesBeforeEdit = await page.evaluate(() =>
    [...document.querySelectorAll("label")]
      .filter((l) => (l.querySelector("input[type=checkbox]") as HTMLInputElement | null)?.checked)
      .map((l) => l.textContent?.trim()),
  );
  expect(checkedBusinessTypesBeforeEdit.sort()).toEqual(["Drop Ship", "ฝากขาย"].sort());
  await fillFirstNameByLabel(page, "ชื่อคลังสินค้าหลายภาษา", `${whName}X`);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, `${whName}X`)).toBe(true);
  const checkedBusinessTypesAfterEdit = await page.evaluate(() =>
    [...document.querySelectorAll("label")]
      .filter((l) => (l.querySelector("input[type=checkbox]") as HTMLInputElement | null)?.checked)
      .map((l) => l.textContent?.trim()),
  );
  expect(checkedBusinessTypesAfterEdit.sort()).toEqual(["Drop Ship", "ฝากขาย"].sort());

  // WAREHOUSE delete
  await clickTightestRowAction(page, "ลบคลังสินค้า", `${whName}X`);
  await page.waitForTimeout(800);
  await clickButtonByText(page, /^ลบ$/);
  await page.waitForTimeout(1500);
  expect(await bodyHasText(page, `${whName}X`)).toBe(false);
});

/**
 * Regression for the warehouse -> location company-scope feature (scopeofwork/warehouse.md):
 * a Warehouse's `companyguids` restricts which companies may use it (empty = all); a Location's
 * own `companyguids` must always be a SUBSET of its parent warehouse's list, enforced server-side
 * by `validateLocationCompanyScope` in `warehouse_location_http_service.go`. Covers:
 *  1. UI: setting warehouse-level company scope via the picker persists (API + the picker itself
 *     only offers the warehouse's allowed companies when adding a location under it).
 *  2. API: a location's companyguids that is a subset of the warehouse's is accepted and persists.
 *  3. API: a location's companyguids with a GUID outside the warehouse's list is REJECTED with 400
 *     on both CREATE and UPDATE, with a clear message, and does not corrupt the stored document on
 *     UPDATE (re-GET shows unchanged state).
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
  let locGuid = "";

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
    await clickButtonByText(page, /^เพิ่มคลังสินค้า$/);
    await page.waitForTimeout(800);
    await fillFieldByPlaceholder(page, "e.g. 00000", whCode);
    await fillFirstNameByLabel(page, "ชื่อคลังสินค้าหลายภาษา", whName);
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

    // Open the warehouse's "add location" form — its picker must offer ONLY c1 (the warehouse's
    // allowed company), never c2, even though c2 is a real tenant company the browser session
    // already knows about (fetched in companiesList after the reload above).
    await clickTightestRowAction(page, "เพิ่มที่เก็บสินค้า", whName);
    await page.waitForTimeout(1500);
    expect(await bodyHasText(page, "แสดงเฉพาะบริษัทที่คลังนี้อนุญาต")).toBe(true);
    const locChips = page.locator("text=บริษัทที่ใช้ที่เก็บสินค้านี้ได้").locator("..").locator("input[type=checkbox]");
    await expect(locChips).toHaveCount(1);

    // Set the location's own scope to the same subset (c1) via the picker and save.
    const locCode = `SCPLOC${uid}`;
    await fillFieldByPlaceholder(page, "e.g. ZONE-A", locCode);
    await fillFirstNameByLabel(page, "ชื่อที่เก็บสินค้าหลายภาษา", `ที่เก็บสิทธิ์${uid}`);
    await page
      .locator("text=บริษัทที่ใช้ที่เก็บสินค้านี้ได้")
      .locator("..")
      .locator("input[type=checkbox]")
      .first()
      .check();
    await clickButtonByText(page, /^บันทึก$/);
    await page.waitForTimeout(2500);

    const locListRes = await (await request.get(`${MAINAPI}/warehouse/${whGuid}/location`, { headers: H })).json();
    const savedLoc = (locListRes.data as { guidfixed: string; code: string; companyguids?: string[] }[]).find(
      (l) => l.code === locCode,
    );
    expect(savedLoc).toBeTruthy();
    locGuid = savedLoc!.guidfixed;
    expect(savedLoc!.companyguids).toEqual([c1]);

    // Backend rejection on CREATE: a location whose companyguids includes c2 (NOT in the
    // warehouse's own [c1]) must be rejected with 400, clear message, and must not persist.
    const badCreate = await request.post(`${MAINAPI}/warehouse/${whGuid}/location`, {
      headers: H,
      data: {
        code: `${locCode}BAD`,
        names: nameX(`badloc${uid}`),
        companyguids: [c1, c2],
      },
    });
    expect(badCreate.status()).toBe(400);
    const badCreateBody = await badCreate.json();
    expect(badCreateBody.success).toBe(false);
    expect(typeof badCreateBody.message).toBe("string");
    expect(badCreateBody.message.length).toBeGreaterThan(0);

    const afterBadCreateList = await (
      await request.get(`${MAINAPI}/warehouse/${whGuid}/location`, { headers: H })
    ).json();
    expect(
      (afterBadCreateList.data as { code: string }[]).some((l) => l.code === `${locCode}BAD`),
    ).toBe(false);

    // Backend rejection on UPDATE: same location, now PUT with companyguids including c2 — must
    // be rejected (400) and must not corrupt the stored document (re-GET shows unchanged state).
    const badUpdate = await request.put(`${MAINAPI}/warehouse/${whGuid}/location/${locGuid}`, {
      headers: H,
      data: {
        code: locCode,
        names: nameX(`ที่เก็บสิทธิ์${uid}`),
        companyguids: [c1, c2],
      },
    });
    expect(badUpdate.status()).toBe(400);
    const badUpdateBody = await badUpdate.json();
    expect(badUpdateBody.success).toBe(false);
    expect(typeof badUpdateBody.message).toBe("string");
    expect(badUpdateBody.message.length).toBeGreaterThan(0);

    const afterReject = await (
      await request.get(`${MAINAPI}/warehouse/${whGuid}/location/${locGuid}`, { headers: H })
    ).json();
    expect(afterReject.data.companyguids).toEqual([c1]);
  } finally {
    if (locGuid) await request.delete(`${MAINAPI}/warehouse/${whGuid}/location/${locGuid}`, { headers: H });
    if (whGuid) await request.delete(`${MAINAPI}/warehouse/${whGuid}`, { headers: H });
    await request.delete(`${MAINAPI}/organization/company/${c2}`, { headers: H });
  }
});
