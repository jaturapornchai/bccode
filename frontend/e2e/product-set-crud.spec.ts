import { test, expect, type Page } from "@playwright/test";

/**
 * Regression test for the "สินค้าชุด" (Product Set) screen — /productset.
 *
 * **Key mechanism (verified against source this run, see `.agents/worklog.md`):** ProductSet has NO
 * dedicated backend model or collection. It is a client-side filtered view of the shared `Product`
 * model (`backend/internal/product/product/models/product.go`, collection `products`) — a set is
 * simply a Product document with `itemtype: 2` (ItemTypeSet) and `materialtype: 3` (MaterialTypeSet).
 * The screen calls the same `GET/POST/PUT/DELETE /product[/:guid]` routes as the main Product screen,
 * filtered by `itemtype=2&materialtype=3` query params on list, and always force-sets
 * `itemtype: 2, materialtype: 3` on the save payload (`product-set-screen.tsx` `handleSave`)
 * regardless of what was loaded into the edit form.
 *
 * **Server-side guard (also verified this run):** `product_http_service.go` `Create`/`Update` both
 * call `barcodeModel.ValidateProductClassification(itemType, materialType)`
 * (`productbarcode/models/product_classification.go`), which rejects itemtype=2 paired with any
 * materialtype other than 3 (and rejects any out-of-range itemtype/materialtype value) with a 400 —
 * so the classification is double-guarded (frontend force-set + backend validation), not just a
 * client-side convention.
 *
 * This suite exercises Create/Read/Update/Delete through the real UI and cross-checks the raw API
 * response to confirm itemtype/materialtype actually persist as 2/3, that the backend rejects an
 * invalid itemtype/materialtype combo without corrupting the stored doc, and that a plain product
 * (itemtype 0) created directly via the API never appears in this screen's filtered list.
 *
 * Login/menu-navigation helpers follow the current convention from `trade-partners-crud.spec.ts`
 * (retry-with-deadline instead of fixed waits; holding match excludes "bctest0N" via `test(?!\d)`;
 * dismiss the one-time "ยังไม่มีหน่วยนับสินค้า" modal on freshly-seeded companies).
 */

const MAINAPI = "http://localhost:8888";

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

/** The NamesEditor label has no htmlFor/id link to its input — find the label's own text node then
 * the first sibling input (same mechanism as product-warehouse-crud.spec.ts `fillFirstNameByLabel`). */
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
      const labelEl = [...document.querySelectorAll<HTMLElement>("span,div")].find((e) =>
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
  // Match the "test" holding specifically — "bctest01".."bctest04" also contain the substring
  // "test" (followed by 2 digits), so a bare /test/i can click the wrong holding.
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

async function loginAndOpenProductSetScreen(page: Page) {
  await loginOnly(page);
  await searchMenuAndOpen(page, "สินค้าชุด", /^สินค้าชุด$/);
  await expect(page.locator("body")).toContainText(/สินค้าชุดทั้งหมด|ไม่พบข้อมูลสินค้าชุด/);
}

test("product set — create/edit/delete via UI, verified against raw API + itemtype/materialtype guard", async ({
  page,
  request,
}) => {
  const uid = Date.now().toString().slice(-6);
  const setCode = `E2ESET${uid}`;
  const setName = `ชุดทดสอบE2E${uid}`;
  const plainCode = `E2EPLAIN${uid}`;
  const ownBarcode = `E2ESETBC${uid}`;

  await loginAndOpenProductSetScreen(page);
  const auth = JSON.parse((await page.evaluate(() => localStorage.getItem("bc_auth")))!) as { token: string };
  const H = { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" };

  let createdGuid = "";
  let plainGuid = "";
  let ownBarcodeGuid = "";

  try {
    // Safety check setup: create a REGULAR product (itemtype 0) directly via API before the set
    // list is (re)loaded, so we can prove it never leaks into the itemtype=2/materialtype=3 view.
    const plainCreate = await request.post(`${MAINAPI}/product`, {
      headers: H,
      data: {
        code: plainCode,
        names: [{ code: "th", name: `สินค้าปกติE2E${uid}` }],
        itemtype: 0,
        materialtype: 0,
      },
    });
    expect(plainCreate.ok()).toBe(true);
    const plainBody = await plainCreate.json();
    plainGuid = plainBody.data?.guidfixed ?? "";
    expect(plainGuid).toBeTruthy();

    // CREATE via UI
    await clickButtonByText(page, /^สร้างสินค้าชุดใหม่$/);
    await page.waitForTimeout(800);
    await fillFieldByPlaceholder(page, "เช่น SET-COMBO-01", setCode);
    await fillFirstNameByLabel(page, "ชื่อสินค้าชุด (รองรับหลายภาษา)", setName);
    await clickButtonByText(page, /^บันทึกข้อมูลชุด$/);
    await page.waitForTimeout(2500);

    // UI: new set appears in the list; the plain product created above must NOT appear (proves the
    // itemtype=2/materialtype=3 filter is not leaking regular products into this screen).
    expect(await bodyHasText(page, setName)).toBe(true);
    expect(await bodyHasText(page, plainCode)).toBe(false);

    // T8 (A6.2): the bundle-vs-BOM clarifier line must render on the list panel.
    expect(await bodyHasText(page, 'สูตรผลิต (BOM)')).toBe(true);

    // API: fetch raw record and confirm itemtype/materialtype are exactly 2/3 (not trusting the
    // 200 response alone — read back a fresh GET by guid).
    const listRes = await (
      await request.get(`${MAINAPI}/product?q=${encodeURIComponent(setCode)}&itemtype=2&materialtype=3&limit=20`, {
        headers: H,
      })
    ).json();
    const created = (
      listRes.data as { guidfixed: string; code: string; itemtype: number; materialtype: number }[]
    ).find((p) => p.code === setCode);
    expect(created).toBeTruthy();
    createdGuid = created!.guidfixed;
    expect(created!.itemtype).toBe(2);
    expect(created!.materialtype).toBe(3);

    // Cross-check: the same itemtype=2/materialtype=3 filtered list must NOT include the plain
    // product created above (server-side filter proof, not just client-side `.filter()` proof).
    expect((listRes.data as { code: string }[]).some((p) => p.code === plainCode)).toBe(false);

    const freshGet = await (await request.get(`${MAINAPI}/product/${createdGuid}`, { headers: H })).json();
    expect(freshGet.data.itemtype).toBe(2);
    expect(freshGet.data.materialtype).toBe(3);
    expect(freshGet.data.code).toBe(setCode);

    // Server-side guard: itemtype=2 (Set) paired with any materialtype other than 3 must be
    // rejected with 400, and must not corrupt the stored document.
    const badUpdate = await request.put(`${MAINAPI}/product/${createdGuid}`, {
      headers: H,
      data: { ...freshGet.data, itemtype: 2, materialtype: 0 },
    });
    expect(badUpdate.status()).toBe(400);
    const badUpdateBody = await badUpdate.json();
    expect(badUpdateBody.success).toBe(false);
    const afterBadUpdate = await (await request.get(`${MAINAPI}/product/${createdGuid}`, { headers: H })).json();
    expect(afterBadUpdate.data.itemtype).toBe(2);
    expect(afterBadUpdate.data.materialtype).toBe(3);
    expect(afterBadUpdate.data.code).toBe(setCode);

    // T7 (A6.1) setup: the set needs its own ProductBarcode record (itemcode === setCode) so it can
    // appear as a candidate row in the component picker (the picker searches ProductBarcode, not
    // Product, directly). Use the real seeded "BOX" unit — an ad-hoc unit code would auto-create a
    // new Unit master record with no Names, which crashes a separate (pre-existing, unrelated)
    // backend GET-by-guid code path; "BOX" avoids that entirely.
    const ownBarcodeCreate = await request.post(`${MAINAPI}/product/barcode`, {
      headers: H,
      data: {
        barcode: ownBarcode,
        itemcode: setCode,
        names: [{ code: "th", name: setName }],
        itemunitcode: "BOX",
        itemunitnames: [{ code: "th", name: "กล่อง" }],
      },
    });
    expect(ownBarcodeCreate.ok()).toBe(true);
    ownBarcodeGuid = (await ownBarcodeCreate.json()).id ?? "";

    // T7 (A6.1): attempting to add the set's own barcode as a component of itself must be rejected
    // client-side with a Thai error toast, and the choice list must remain unchanged (no self-
    // reference corruption). `addChoiceToGroup` guard in product-set-screen.tsx.
    // Reopen the set for editing (the create form closes back to the list view after save).
    await clickButtonByText(page, new RegExp(setCode.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")));
    await page.waitForTimeout(1200);
    await clickButtonByText(page, /^แก้ไขข้อมูลชุด$/);
    await page.waitForTimeout(1000);
    await clickButtonByText(page, /^สินค้าประกอบชุด/);
    await page.waitForTimeout(500);
    await clickButtonByText(page, /^เพิ่มกลุ่มตัวเลือกใหม่$/);
    await page.waitForTimeout(500);
    await clickButtonByText(page, /^ดึงบาร์โค้ดเข้ามา$/);
    await page.waitForTimeout(600);
    await page.fill('input[placeholder="ค้นหาบาร์โค้ด หรือชื่อสินค้า..."]', ownBarcode);
    await page.waitForTimeout(800);
    await clickButtonByText(page, new RegExp(ownBarcode));
    await page.waitForTimeout(600);
    expect(await bodyHasText(page, "ไม่สามารถเพิ่มสินค้าชุดนี้เป็นส่วนประกอบของตัวเองได้")).toBe(true);
    expect(await bodyHasText(page, "ยังไม่มีบาร์โค้ดชิ้นส่วน")).toBe(true);

    // UPDATE via UI: open the set from the list, edit it, rename, save — confirm itemtype/
    // materialtype survive the round-trip (the screen force-sets 2/3 on every save; this proves it
    // doesn't accidentally revert to a regular product on edit).
    await clickButtonByText(page, new RegExp(setCode.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")));
    await page.waitForTimeout(1200);
    await clickButtonByText(page, /^แก้ไขข้อมูลชุด$/);
    await page.waitForTimeout(1000);
    await fillFirstNameByLabel(page, "ชื่อสินค้าชุด (รองรับหลายภาษา)", `${setName}X`);
    await clickButtonByText(page, /^บันทึกข้อมูลชุด$/);
    await page.waitForTimeout(2500);

    expect(await bodyHasText(page, `${setName}X`)).toBe(true);

    const afterUpdate = await (await request.get(`${MAINAPI}/product/${createdGuid}`, { headers: H })).json();
    expect(afterUpdate.data.itemtype).toBe(2);
    expect(afterUpdate.data.materialtype).toBe(3);
    expect(afterUpdate.data.names.some((n: { name: string }) => n.name === `${setName}X`)).toBe(true);

    // DELETE via UI: confirm it only removes this set's record, not the plain product created
    // above (which shares the same underlying `/product` collection and delete endpoint).
    await clickButtonByText(page, /^ลบสินค้าชุด$/);
    await page.waitForTimeout(800);
    await clickButtonByText(page, /^ลบ$/);
    await page.waitForTimeout(1500);
    expect(await bodyHasText(page, `${setName}X`)).toBe(false);

    const afterDelete = await request.get(`${MAINAPI}/product/${createdGuid}`, { headers: H });
    const afterDeleteBody = await afterDelete.json();
    expect(afterDeleteBody.success === false || afterDelete.status() === 404).toBe(true);
    createdGuid = ""; // deleted through the UI already; skip the finally-block cleanup delete

    // The unrelated plain product must still exist untouched after the set delete.
    const plainAfter = await (await request.get(`${MAINAPI}/product/${plainGuid}`, { headers: H })).json();
    expect(plainAfter.success).toBe(true);
    expect(plainAfter.data.code).toBe(plainCode);
    expect(plainAfter.data.itemtype).toBe(0);
  } finally {
    if (createdGuid) await request.delete(`${MAINAPI}/product/${createdGuid}`, { headers: H }).catch(() => {});
    if (plainGuid) await request.delete(`${MAINAPI}/product/${plainGuid}`, { headers: H }).catch(() => {});
    if (ownBarcodeGuid) await request.delete(`${MAINAPI}/product/barcode/${ownBarcodeGuid}`, { headers: H }).catch(() => {});
  }
});
