import { test, expect, type Page } from "@playwright/test";

/**
 * Sample-data build for the "สินค้าและบาร์โค้ด" menu group (สินค้า /product, บาร์โค้ด /productbarcode,
 * สินค้าชุด /productset), driven entirely through the real UI ("ทำเหมือน user ป้อนเองเลย" — Jead's
 * explicit instruction), reusing every real classification master seeded in the prior phase (brand/
 * class/design/model/pattern/grade/category/group/groupsubone/groupsubtwo — see
 * `.agents/worklog.md`/session report for the exact codes).
 *
 * Theme: small Thai grocery/beverage shop ("ร้านชำ"), same as the master-data phase, so every
 * created record reads as one coherent, hand-entered business, not disconnected random rows.
 *
 * **Key mechanism verified against real source this session (do not re-derive):**
 * - Product classification tab (`tab-product-classification.tsx`) drives 10 MasterPicker fields
 *   (group/groupsubone/groupsubtwo/brand/category/class/design/model/pattern/grade), each opened via
 *   `aria-label="Select <Thai label>"` and each picker row rendered as `<li><button>Name  CODE</button></li>`
 *   (`master-picker.tsx`) — same interaction pattern as the barcode screen's unit picker in
 *   `product-barcode-crud.spec.ts`.
 * - `product-screen.tsx`'s `handlePickerSelect` sets `groupguid` as a frontend-only field for "group" —
 *   the real Go backend model (`backend/internal/product/product/models/product.go`) has NO `GroupGuid`
 *   field on Product (only `GroupCode`/`GroupNames`), so that assignment is silently dropped on save.
 *   This is a harmless pre-existing dead field (same category as the already-flagged
 *   `ProductGroup.ChildCount` dead field) — verification below checks `groupcode`, the field that
 *   actually persists and round-trips.
 * - Barcode's own classification section (`barcode-form.tsx`) is READ-ONLY ("inherited from product"
 *   banner). Linking the barcode to a real Product via the "สินค้าหลัก" `MasterField` picker (the real
 *   UI flow used below, NOT a raw API POST with a bare `itemcode` string) sets `itemguid` on create —
 *   verified live (real MongoDB document + fresh `GET /product/barcode/:guid`) this session: the stored
 *   Mongo doc's own `groupcode`/`brandcode`/`classcode` stay blank, but a `GET` by guid returns them
 *   enriched from the linked product (`productbarcode_http_service.go` lines ~709/1045, keyed on
 *   `ItemGuid`) — i.e. this is a live, on-read computed/denormalized value, not a dead code path, as
 *   long as the barcode was actually linked to its product through this real picker.
 * - Product Set: the real `/productset` screen (`product-set-screen.tsx`) bundles components via
 *   `options[]` -> `choices[].refbarcode` (the same struct as the restaurant-menu-modifier feature,
 *   `ProductOption`/`ProductChoice` in `product.go`), NOT the raw `RefBarcodes[]`/`BOM[]` fields. This is
 *   what a real user actually produces by driving the screen (confirmed via source read of
 *   `product-set-screen.tsx` `addChoiceToGroup`/`handleSave`) — the task briefing's `RefBarcodes`/`BOM`
 *   note describes backend struct shape, but is not the mechanism the live screen exercises. Flagged in
 *   the session report.
 * - "สินค้าในหมวด" (`/productcategorylist`, `product-category-items-editor.tsx`) assigns Product
 *   master codes into a category's `codelist`. Barcode records stay independent and are never used
 *   as category membership.
 *
 * Login/menu-navigation helpers follow the current repo convention (retry-with-deadline, "test" holding
 * excludes "bctest0N" via `test(?!\d)`, dismiss the one-time "ยังไม่มีหน่วยนับสินค้า" modal).
 */

const MAINAPI = "http://localhost:8888";

// 6 real sample products: coffee (x2), milk (x1), soda (x1), snack (x1), household (x1) — themed to
// match the seeded classification tree so every relation reads as a coherent grocery-shop catalog.
const PRODUCTS = [
  {
    key: "COFFEE1",
    name: "กาแฟกระป๋องเย็น",
    group: "GRP-COFFEE",
    groupsubone: "GS1-COFFEE",
    groupsubtwo: "GS2-CAN",
    brand: "BR-MAEKONG",
    category: "CT-DRINK",
    class: "CL-DRINK",
    design: "DS-STANDARD",
    model: "MD-STANDARD",
    pattern: "PT-PLAIN",
    grade: "GR-PREMIUM",
    unit: "BOX",
    price: 15,
  },
  {
    key: "COFFEE2",
    name: "กาแฟขวดพรีเมียม",
    group: "GRP-COFFEE",
    groupsubone: "GS1-COFFEE",
    groupsubtwo: "GS2-BOTTLE",
    brand: "BR-JAENOO",
    category: "CT-DRINK",
    class: "CL-DRINK",
    design: "DS-PREMIUM",
    model: "MD-ECONOMY",
    pattern: "PT-CLASSIC",
    grade: "GR-PREMIUM",
    unit: "BOX",
    price: 20,
  },
  {
    key: "MILK1",
    name: "นมสดกล่อง",
    group: "GRP-MILK",
    groupsubone: "GS1-MILK",
    groupsubtwo: "GS2-BOX",
    brand: "BR-THONGSAMUT",
    category: "CT-DRINK",
    class: "CL-DRINK",
    design: "DS-STANDARD",
    model: "MD-STANDARD",
    pattern: "PT-PLAIN",
    grade: "GR-STANDARD",
    unit: "BOX",
    price: 12,
  },
  {
    key: "SODA1",
    name: "น้ำอัดลมกระป๋อง",
    group: "GRP-SODA",
    groupsubone: "GS1-SODA",
    groupsubtwo: "GS2-CAN",
    brand: "BR-MAEKONG",
    category: "CT-DRINK",
    class: "CL-DRINK",
    design: "DS-STANDARD",
    model: "MD-ECONOMY",
    pattern: "PT-PLAIN",
    grade: "GR-STANDARD",
    unit: "BOX",
    price: 10,
  },
  {
    key: "SNACK1",
    name: "ขนมขบเคี้ยวถุง",
    group: "GRP-KITCHEN",
    groupsubone: "GS1-SODA",
    groupsubtwo: "GS2-BOX",
    brand: "BR-JAENOO",
    category: "CT-SNACK",
    class: "CL-SNACK",
    design: "DS-STANDARD",
    model: "MD-STANDARD",
    pattern: "PT-CLASSIC",
    grade: "GR-STANDARD",
    unit: "BOX",
    price: 8,
  },
  {
    key: "HOUSE1",
    name: "น้ำยาล้างจานขวด",
    group: "GRP-KITCHEN",
    groupsubone: "GS1-MILK",
    groupsubtwo: "GS2-BOTTLE",
    brand: "BR-THONGSAMUT",
    category: "CT-HOUSEHOLD",
    class: "CL-HOUSEHOLD",
    design: "DS-PREMIUM",
    model: "MD-ECONOMY",
    pattern: "PT-CLASSIC",
    grade: "GR-PREMIUM",
    unit: "BOX",
    price: 25,
  },
] as const;

const CLASSIFICATION_LABELS: Record<string, string> = {
  group: "กลุ่มสินค้า",
  groupsubone: "กลุ่มย่อย 1",
  groupsubtwo: "กลุ่มย่อย 2",
  brand: "ยี่ห้อ",
  category: "หมวดสินค้า",
  class: "คลาสสินค้า",
  design: "Design",
  model: "Model",
  pattern: "Pattern",
  grade: "เกรด",
};

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

/** Clicks a category-tree row on `product-category-tree-view.tsx` whose row text includes `text` as
 *  its own category name (excludes the "N หมวด/สินค้า" badge counts, matched via the shortest matching
 *  row so a parent row's substring match doesn't win over its own child row). Tree rows are
 *  `data-category-row-guid` divs with an onClick handler on the row itself (not a <button>) —
 *  `clickButtonByText` (button/a/[role=button] only) cannot find them. */
async function clickTreeRowByText(page: Page, text: string) {
  const deadline = Date.now() + 8000;
  for (;;) {
    const clicked = await page.evaluate((t) => {
      const rows = [...document.querySelectorAll<HTMLElement>("[data-category-row-guid]")].filter(
        (e) => (e.textContent ?? "").includes(t) && e.offsetParent !== null,
      );
      if (rows.length === 0) return false;
      // Prefer the row with the shortest own text — the exact-match leaf row, not an ancestor whose
      // aggregated textContent happens to also include `t` from a nested child row.
      rows.sort((a, b) => (a.textContent ?? "").length - (b.textContent ?? "").length);
      rows[0].click();
      return true;
    }, text);
    if (clicked) return;
    if (Date.now() > deadline) throw new Error(`No tree row found containing text "${text}"`);
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

function getAuthHeaders(page: Page) {
  return page.evaluate(() => {
    const auth = JSON.parse(localStorage.getItem("bc_auth")!) as { token: string; backendUrl: string };
    return { Authorization: `Bearer ${auth.token}`, "x-bc-backend-url": auth.backendUrl };
  });
}

/** Opens the "สินค้า" tab via the left menu search (matches product-crud.spec.ts convention). */
async function openProductScreen(page: Page) {
  await searchMenuAndOpen(page, "สินค้าและบาร์โค้ด", /^สินค้า$/);
  await expect(page.locator("body")).toContainText(/ไม่พบข้อมูลสินค้า|รหัสสินค้า/);
}

async function fillFieldByLabelPrefix(page: Page, labelPrefix: string, value: string) {
  await page.evaluate(
    ({ labelPrefix, value }) => {
      const setValue = (el: HTMLInputElement, v: string) => {
        const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
        setter.call(el, v);
        el.dispatchEvent(new Event("input", { bubbles: true }));
        el.dispatchEvent(new Event("change", { bubbles: true }));
      };
      const label = [...document.querySelectorAll<HTMLElement>("label")].find((l) =>
        (l.textContent ?? "").trim().startsWith(labelPrefix),
      );
      if (!label) throw new Error(`No label found starting with "${labelPrefix}"`);
      const input = label.parentElement?.querySelector<HTMLInputElement>("input");
      if (!input) throw new Error(`No input found under label "${labelPrefix}"`);
      setValue(input, value);
    },
    { labelPrefix, value },
  );
}

async function fillFirstNameInput(page: Page, value: string) {
  await page.evaluate((value) => {
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    const input = document.querySelector<HTMLInputElement>('input[placeholder="th"]');
    if (!input) throw new Error('No name input found with placeholder "th"');
    setter.call(input, value);
    input.dispatchEvent(new Event("input", { bubbles: true }));
    input.dispatchEvent(new Event("change", { bubbles: true }));
  }, value);
}

/** Opens the classification MasterPicker for `field` (aria-label="Select <label>") and picks the row
 *  whose small code span matches `code` exactly. Mirrors the unit-picker pattern already established
 *  in `product-barcode-crud.spec.ts`'s `pickUnitByCode`, but uses native Playwright locators (not a raw
 *  `page.evaluate` click) — a manual DOM-evaluate click was observed to occasionally hang the CDP
 *  round-trip in headless runs even though the same click works instantly when driven live. */
async function pickClassification(page: Page, field: string, code: string) {
  const label = CLASSIFICATION_LABELS[field];
  await page.locator(`button[aria-label="Select ${label}"]`).click();
  const row = page.locator("li button", { hasText: code }).first();
  await expect(row, `picker row for ${field}=${code}`).toBeVisible({ timeout: 5000 });
  await row.click();
  await page.waitForTimeout(300);
}

test.describe.serial("sample data — สินค้า/บาร์โค้ด/สินค้าชุด (real UI, related grocery-shop theme)", () => {
  const uid = Date.now().toString().slice(-6);
  const productGuids: Record<string, string> = {};
  const barcodeGuids: Record<string, string> = {};
  const barcodeCodes: Record<string, string> = {};

  test("create 6 products via /product UI, each with all 10 classification pickers set", async ({ page }) => {
    await loginOnly(page);
    await openProductScreen(page);

    for (const p of PRODUCTS) {
      const code = `SMP${p.key}${uid}`;
      const name = `${p.name}${uid}`;

      await clickButtonByText(page, /^เพิ่ม$/);
      await page.waitForTimeout(800);

      // The form may reopen on whatever tab was active for the previous product — force back to
      // "ข้อมูลหลัก" (basic tab) first so the code/name labels are actually on-screen.
      await clickButtonByText(page, /^ข้อมูลหลัก$/);
      await page.waitForTimeout(200);

      await fillFieldByLabelPrefix(page, "รหัสสินค้า", code);
      await fillFirstNameInput(page, name);

      // Classification tab lives under the "หมวดหมู่" advanced tab.
      await clickButtonByText(page, /^ขั้นสูง/);
      await page.waitForTimeout(200);
      await clickButtonByText(page, /^หมวดหมู่$/);
      await page.waitForTimeout(300);

      for (const field of ["group", "groupsubone", "groupsubtwo", "brand", "category", "class", "design", "model", "pattern", "grade"] as const) {
        await pickClassification(page, field, p[field]);
      }

      const [createRes] = await Promise.all([
        page.waitForResponse((res) => res.url().includes("/api/product") && res.request().method() === "POST"),
        clickButtonByText(page, /^บันทึก$/),
      ]);
      expect(createRes.status(), `create ${code}`).toBe(201);
      const body = await createRes.json();
      expect(body.success).toBe(true);
      productGuids[p.key] = body.data.guidfixed;
      expect(productGuids[p.key]).toBeTruthy();
      await page.waitForTimeout(800);
    }

    expect(Object.keys(productGuids).length).toBe(PRODUCTS.length);
  });

  test("create 1 barcode per product via /productbarcode UI (real unit + price)", async ({ page }) => {
    await loginOnly(page);
    await searchMenuAndOpen(page, "บาร์โค้ด", /^บาร์โค้ด$/);
    await expect(page.locator("body")).toContainText(/จัดการบาร์โค้ด/);

    for (const p of PRODUCTS) {
      const code = `SMP${p.key}${uid}`;
      const barcode = `SMPBC${p.key}${uid}`;
      barcodeCodes[p.key] = barcode;

      await clickButtonByText(page, /^เพิ่ม$/);
      await page.waitForTimeout(1000);

      // "บาร์โค้ด" is a FieldRow field (plain sibling <span> label, no wrapping <label>).
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
        { pattern: "^บาร์โค้ด", value: barcode },
      );

      // "สินค้าหลัก" (Main Product) MasterField — links this barcode to the real Product just created
      // (sets `itemguid`/`itemcode`/`names` from the picked product, matching a real user's workflow:
      // link the barcode to an existing product rather than typing a free-text code).
      await page.locator('button[aria-label="สินค้าหลัก"]').click();
      const productRow = page.locator("li button", { hasText: code }).first();
      await expect(productRow, `product picker row for ${code}`).toBeVisible({ timeout: 5000 });
      await productRow.click();
      await page.waitForTimeout(400);

      // Unit picker.
      await page.locator('button[aria-label="หน่วยนับ"]').click();
      const unitRow = page.locator("li button", { hasText: p.unit }).first();
      await expect(unitRow, `unit picker row for ${p.unit}`).toBeVisible({ timeout: 5000 });
      await unitRow.click();
      await page.waitForTimeout(400);

      // Price tier 1.
      await clickButtonByText(page, /^ราคา$/);
      await page.waitForTimeout(400);
      const tier1 = await page.evaluateHandle(() =>
        [...document.querySelectorAll<HTMLInputElement>("input")].find((i) => i.value === "0"),
      );
      await page.evaluate(
        ({ input, value }) => {
          const el = input as HTMLInputElement;
          const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
          setter.call(el, value);
          el.dispatchEvent(new Event("input", { bubbles: true }));
          el.dispatchEvent(new Event("change", { bubbles: true }));
        },
        { input: tier1, value: String(p.price) },
      );

      await clickButtonByText(page, /^บันทึก$/);
      await page.waitForTimeout(2000);
      expect(await bodyHasText(page, "บันทึกบาร์โค้ดแล้ว")).toBe(true);
      await page.waitForTimeout(500);
    }

    // Resolve every created barcode's real guidfixed via the same search API the screen uses.
    const H = await getAuthHeaders(page);
    for (const p of PRODUCTS) {
      const listRes = await (
        await page.request.get(`${MAINAPI}/product/barcode?q=${barcodeCodes[p.key]}&limit=5`, { headers: H })
      ).json();
      const found = (listRes.data as { guidfixed: string; barcode: string }[]).find((b) => b.barcode === barcodeCodes[p.key]);
      expect(found, `barcode lookup for ${p.key}`).toBeTruthy();
      barcodeGuids[p.key] = found!.guidfixed;
    }
    expect(Object.keys(barcodeGuids).length).toBe(PRODUCTS.length);
  });

  test("create 2 Product Sets via /productset UI, bundling real barcodes as components", async ({ page }) => {
    await loginOnly(page);
    await searchMenuAndOpen(page, "สินค้าชุด", /^สินค้าชุด$/);
    await expect(page.locator("body")).toContainText(/สินค้าชุดทั้งหมด|ไม่พบข้อมูลสินค้าชุด/);

    const sets = [
      { setKey: "SETCOFFEE", name: `ชุดกาแฟยามเช้า${uid}`, components: ["COFFEE1", "MILK1"] as const },
      { setKey: "SETPARTY", name: `ชุดปาร์ตี้เครื่องดื่ม${uid}`, components: ["SODA1", "SNACK1", "COFFEE2"] as const },
    ];
    const setGuids: Record<string, string> = {};

    for (const set of sets) {
      const setCode = `SMP${set.setKey}${uid}`;

      await clickButtonByText(page, /^สร้างสินค้าชุดใหม่$/);
      await page.waitForTimeout(800);
      await page.evaluate((value) => {
        const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
        const input = document.querySelector<HTMLInputElement>('input[placeholder="เช่น SET-COMBO-01"]');
        if (!input) throw new Error("No set-code input found");
        setter.call(input, value);
        input.dispatchEvent(new Event("input", { bubbles: true }));
        input.dispatchEvent(new Event("change", { bubbles: true }));
      }, setCode);
      await page.evaluate((value) => {
        const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
        const ownTextOnly = (el: Element) =>
          [...el.childNodes]
            .filter((n) => n.nodeType === Node.TEXT_NODE)
            .map((n) => n.textContent ?? "")
            .join("")
            .trim();
        const labelEl = [...document.querySelectorAll<HTMLElement>("span,div")].find((e) =>
          ownTextOnly(e).startsWith("ชื่อสินค้าชุด (รองรับหลายภาษา)"),
        );
        if (!labelEl) throw new Error("No set-name label found");
        const input = labelEl.parentElement?.querySelector<HTMLInputElement>("input");
        if (!input) throw new Error("No set-name input found");
        setter.call(input, value);
        input.dispatchEvent(new Event("input", { bubbles: true }));
        input.dispatchEvent(new Event("change", { bubbles: true }));
      }, set.name);
      await clickButtonByText(page, /^บันทึกข้อมูลชุด$/);
      await page.waitForTimeout(2500);
      expect(await bodyHasText(page, set.name)).toBe(true);

      // Reopen for editing to add real components (create form closes back to the list after save).
      await clickButtonByText(page, new RegExp(setCode.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")));
      await page.waitForTimeout(1200);
      await clickButtonByText(page, /^แก้ไขข้อมูลชุด$/);
      await page.waitForTimeout(1000);
      await clickButtonByText(page, /^สินค้าประกอบชุด/);
      await page.waitForTimeout(500);
      await clickButtonByText(page, /^เพิ่มกลุ่มตัวเลือกใหม่$/);
      await page.waitForTimeout(500);

      for (const compKey of set.components) {
        const compBarcode = barcodeCodes[compKey];
        await clickButtonByText(page, /^ดึงบาร์โค้ดเข้ามา$/);
        await page.waitForTimeout(600);
        await page.fill('input[placeholder="ค้นหาบาร์โค้ด หรือชื่อสินค้า..."]', compBarcode);
        await page.waitForTimeout(800);
        await clickButtonByText(page, new RegExp(compBarcode));
        await page.waitForTimeout(600);
      }

      await clickButtonByText(page, /^บันทึกข้อมูลชุด$/);
      await page.waitForTimeout(2500);

      const H = await getAuthHeaders(page);
      const listRes = await (
        await page.request.get(`${MAINAPI}/product?q=${setCode}&itemtype=2&materialtype=3&limit=20`, { headers: H })
      ).json();
      const created = (listRes.data as { guidfixed: string; code: string; itemtype: number; materialtype: number }[]).find(
        (item) => item.code === setCode,
      );
      expect(created, `set lookup for ${setCode}`).toBeTruthy();
      expect(created!.itemtype).toBe(2);
      expect(created!.materialtype).toBe(3);
      setGuids[set.setKey] = created!.guidfixed;

      const freshGet = await (await page.request.get(`${MAINAPI}/product/${created!.guidfixed}`, { headers: H })).json();
      const options = (freshGet.data.options ?? []) as { choices?: { refbarcode: string }[] }[];
      const allRefBarcodes = options.flatMap((o) => (o.choices ?? []).map((c) => c.refbarcode));
      for (const compKey of set.components) {
        expect(allRefBarcodes, `${setCode} contains ${compKey}`).toContain(barcodeCodes[compKey]);
      }
    }

    expect(Object.keys(setGuids).length).toBe(sets.length);
  });

  test("assign coffee-themed Product masters into the real productcategories 'กาแฟ' node via /productcategorylist UI", async ({
    page,
  }) => {
    await loginOnly(page);
    await page.goto("/productcategorylist", { waitUntil: "domcontentloaded" });
    await page.waitForTimeout(2000);
    await expect(page.locator("body")).toContainText(/สินค้าในหมวด/);

    // Group 1 = "เครื่องดื่ม" (seeded Phase 2). Select it, expand via the real "ลูก N" toggle button
    // (`aria-label` starts with "แสดงหมวดย่อย"/"ซ่อนหมวดย่อย" — product-category-tree-view.tsx), click
    // "กาแฟ" child.
    await clickButtonByText(page, /1.*เครื่องดื่ม|เครื่องดื่ม.*1/);
    await page.waitForTimeout(1200);
    const hasChild = await bodyHasText(page, "กาแฟ");
    if (!hasChild) {
      await page.locator('button[aria-label^="แสดงหมวดย่อย"]').first().click();
      await page.waitForTimeout(600);
    }
    await clickTreeRowByText(page, "กาแฟ");
    await page.waitForTimeout(1000);

    await clickButtonByText(page, /เพิ่มสินค้า/);
    await page.waitForTimeout(800);

    const coffeeComponents = ["COFFEE1", "COFFEE2"] as const;
    for (const key of coffeeComponents) {
      const productCode = `SMP${key}${uid}`;
      await page.fill('input[placeholder*="ค้นหาด้วยรหัสสินค้า"]', productCode);
      await page.waitForTimeout(800);
      await clickButtonByText(page, /^เพิ่ม$/);
      await page.waitForTimeout(400);
    }
    await clickButtonByText(page, /^ปิด$/);
    await page.waitForTimeout(400);
    await clickButtonByText(page, /^บันทึก$/);
    await page.waitForTimeout(1500);
    expect(await bodyHasText(page, "บันทึกข้อมูลเรียบร้อยแล้ว")).toBe(true);

    // Verify via fresh API read: the "กาแฟ" category contains Product codes only.
    const H = await getAuthHeaders(page);
    const catRes = await (
      await page.request.get(`${MAINAPI}/product/category/list?group-number=1`, { headers: H })
    ).json();
    const coffeeCat = (catRes.data as { names: { name: string }[]; codelist?: { code: string; barcode?: unknown }[] }[]).find(
      (c) => c.names.some((n) => n.name === "กาแฟ"),
    );
    expect(coffeeCat, "กาแฟ category node found").toBeTruthy();
    const categoryProducts = coffeeCat!.codelist ?? [];
    const categoryProductCodes = categoryProducts.map((item) => item.code);
    for (const key of coffeeComponents) {
      expect(categoryProductCodes, `กาแฟ codelist contains Product ${key}`).toContain(`SMP${key}${uid}`);
    }
    expect(categoryProducts.every((item) => !("barcode" in item)), "กาแฟ codelist is Product-only").toBe(true);
  });

  test("relation-completeness check: every classification field on every product resolves to a real Phase-2 master doc", async ({
    page,
  }) => {
    await loginOnly(page);
    const H = await getAuthHeaders(page);

    // Pull the real master lists once, to cross-check codes actually exist (not just non-empty).
    const masterCodes: Record<string, Set<string>> = {};
    for (const master of ["brand", "class", "design", "model", "pattern", "grade", "category", "groupsubone", "groupsubtwo"]) {
      const res = await (await page.request.get(`/api/product-barcode/master/${master}?limit=50`, { headers: H })).json();
      masterCodes[master] = new Set((res.data as { code: string }[]).map((d) => d.code));
    }
    const groupRes = await (await page.request.get(`/api/system-settings/productgroup?limit=50`, { headers: H })).json();
    masterCodes["group"] = new Set((groupRes.data as { code: string }[]).map((d) => d.code));

    for (const p of PRODUCTS) {
      const guid = productGuids[p.key];
      const getRes = await (await page.request.get(`${MAINAPI}/product/${guid}`, { headers: H })).json();
      const data = getRes.data;

      expect(data.groupcode, `${p.key}.groupcode`).toBe(p.group);
      expect(masterCodes.group.has(data.groupcode), `${p.key}.groupcode resolves`).toBe(true);

      expect(data.groupsubonecode, `${p.key}.groupsubonecode`).toBe(p.groupsubone);
      expect(masterCodes.groupsubone.has(data.groupsubonecode), `${p.key}.groupsubonecode resolves`).toBe(true);

      expect(data.groupsubtwocode, `${p.key}.groupsubtwocode`).toBe(p.groupsubtwo);
      expect(masterCodes.groupsubtwo.has(data.groupsubtwocode), `${p.key}.groupsubtwocode resolves`).toBe(true);

      expect(data.brandcode, `${p.key}.brandcode`).toBe(p.brand);
      expect(masterCodes.brand.has(data.brandcode), `${p.key}.brandcode resolves`).toBe(true);

      expect(data.categorycode, `${p.key}.categorycode`).toBe(p.category);
      expect(masterCodes.category.has(data.categorycode), `${p.key}.categorycode resolves`).toBe(true);

      expect(data.classcode, `${p.key}.classcode`).toBe(p.class);
      expect(masterCodes.class.has(data.classcode), `${p.key}.classcode resolves`).toBe(true);

      expect(data.designcode, `${p.key}.designcode`).toBe(p.design);
      expect(masterCodes.design.has(data.designcode), `${p.key}.designcode resolves`).toBe(true);

      expect(data.modelcode, `${p.key}.modelcode`).toBe(p.model);
      expect(masterCodes.model.has(data.modelcode), `${p.key}.modelcode resolves`).toBe(true);

      expect(data.patterncode, `${p.key}.patterncode`).toBe(p.pattern);
      expect(masterCodes.pattern.has(data.patterncode), `${p.key}.patterncode resolves`).toBe(true);

      expect(data.gradecode, `${p.key}.gradecode`).toBe(p.grade);
      expect(masterCodes.grade.has(data.gradecode), `${p.key}.gradecode resolves`).toBe(true);
    }
  });
});
