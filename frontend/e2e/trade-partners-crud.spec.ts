import { test, expect, type Page } from "@playwright/test";

/**
 * Regression test for the "คู่ค้า (ลูกค้า/ผู้ขาย)" (Trade Partners) master screens — all 4 were
 * built fresh in this task (no real frontend screen existed before): /creditorgroup, /debtorgroup,
 * /creditor, /debtor. All 4 are `main-crud` config-driven screens (`system-setting-screens.ts`),
 * same rendering shell as `master-brand-crud.spec.ts` (see that file for the DOM-shape notes this
 * suite reuses: labels wrap inputs via `input.closest("label")`, row action buttons are siblings of
 * the row text not ancestors, delete uses an in-page confirm modal not `window.confirm`).
 *
 * Backend: `backend/internal/debtaccount/{creditor,creditorgroup,debtor,debtorgroup}` — MongoDB +
 * Kafka + PostgreSQL projection pipeline, routes `/debtaccount/creditor(-group)?` /
 * `/debtaccount/debtor(-group)?`. Frontend proxies via
 * `/api/product-barcode/master/[master]/route.ts` (creditorgroup -> /debtaccount/creditor-group,
 * debtorgroup -> /debtaccount/debtor-group) for the `master-multi-picker` group-assignment field,
 * and via the generic `/api/system-settings/[[...settingPath]]/route.ts` proxy for CRUD itself.
 */

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

async function clickButtonByText(page: Page, textPattern: RegExp) {
  // Retry for a few seconds — the async-loaded lists this suite waits on (holdings, companies,
  // branches) occasionally take longer than a fixed timeout to render (flaky under load), and a
  // bounded retry is cheaper and more reliable than guessing a bigger fixed wait everywhere.
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

// Company card button text is composite (icon alt + name + tax/currency badges, e.g. "บริษัท
// บริษัท OWNER TH THB 00000") — much longer than header/nav buttons like "คู่มือ"/"ตั้งค่าระบบ".
// Pick the first sufficiently-long button containing "บริษัท" instead of the first clickable
// element on the page. Retries like clickButtonByText since the company list loads async.
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

async function clickRowActionButton(page: Page, buttonTextPattern: RegExp, rowText: string) {
  const clicked = await page.evaluate(
    ({ pattern, rowText }) => {
      const re = new RegExp(pattern);
      const candidates = [...document.querySelectorAll<HTMLElement>("button")].filter(
        (b) => re.test((b.textContent ?? "").trim()) && b.offsetParent !== null,
      );
      for (const btn of candidates) {
        let el: HTMLElement | null = btn;
        for (let i = 0; i < 6 && el; i++) {
          el = el.parentElement;
          if (el && (el.innerText ?? "").includes(rowText)) {
            btn.click();
            return true;
          }
        }
      }
      return false;
    },
    { pattern: buttonTextPattern.source, rowText },
  );
  if (!clicked) throw new Error(`No "${buttonTextPattern}" action button found in a row containing "${rowText}"`);
}

async function bodyHasText(page: Page, text: string): Promise<boolean> {
  return page.evaluate((t) => document.body.innerText.includes(t), text);
}

async function searchMenuAndOpen(page: Page, searchTerm: string, exactItemPattern: RegExp) {
  // Menu search input renders only after the main-menu shell mounts post-branch-selection;
  // retry rather than assume a fixed wait always covers it (same flakiness class as login).
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
  // "test" (followed by 2 digits), so a bare /test/i can click the wrong holding. Excluding a
  // trailing digit disambiguates "test" from "bctest0N".
  await clickButtonByText(page, /test(?!\d)/i);
  await page.waitForTimeout(2500);
  await clickCompanyCard(page);
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /สำนักงานใหญ่|headquarters/i);
  await page.waitForTimeout(2500);
  // A freshly-created company with zero product units shows a one-time "ยังไม่มีหน่วยนับสินค้า"
  // modal blocking the menu after picking a branch. Dismiss it via "เข้าเมนูก่อน" if present
  // (only companies seeded without product-unit setup hit this — pre-seeded holdings don't).
  const hasProductUnitModal = await bodyHasText(page, "ยังไม่มีหน่วยนับสินค้า");
  if (hasProductUnitModal) {
    await clickButtonByText(page, /^เข้าเมนูก่อน$/);
    await page.waitForTimeout(1500);
  }
}

test("creditorgroup — create, edit, delete", async ({ page }) => {
  const uid = Date.now().toString().slice(-6);
  const code = `E2ECG${uid}`;
  const name = `กลุ่มเจ้าหนี้E2E${uid}`;
  const name2 = `กลุ่มเจ้าหนี้E2Eแก้ไข${uid}`;

  await loginOnly(page);
  await searchMenuAndOpen(page, "กลุ่มเจ้าหนี้", /^กลุ่มเจ้าหนี้$/);
  await expect(page.locator("body")).toContainText(/กลุ่มผู้จำหน่าย/i);

  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัส/, code);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name)).toBe(true);

  await clickRowActionButton(page, /^แก้ไข$/, code);
  await page.waitForTimeout(1500);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name2);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name2)).toBe(true);

  await clickRowActionButton(page, /^ลบ$/, code);
  await page.waitForTimeout(1000);
  expect(await bodyHasText(page, "ต้องการลบ")).toBe(true);
  await clickRowActionButton(page, /^ลบ$/, "ต้องการลบจริงหรือไม่");
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name2)).toBe(false);
});

test("debtorgroup — create, edit, delete", async ({ page }) => {
  const uid = Date.now().toString().slice(-6);
  const code = `E2EDG${uid}`;
  const name = `กลุ่มลูกหนี้E2E${uid}`;
  const name2 = `กลุ่มลูกหนี้E2Eแก้ไข${uid}`;

  await loginOnly(page);
  await searchMenuAndOpen(page, "กลุ่มลูกหนี้", /^กลุ่มลูกหนี้$/);
  await expect(page.locator("body")).toContainText(/กลุ่มลูกค้า/i);

  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัส/, code);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name)).toBe(true);

  await clickRowActionButton(page, /^แก้ไข$/, code);
  await page.waitForTimeout(1500);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name2);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name2)).toBe(true);

  await clickRowActionButton(page, /^ลบ$/, code);
  await page.waitForTimeout(1000);
  expect(await bodyHasText(page, "ต้องการลบ")).toBe(true);
  await clickRowActionButton(page, /^ลบ$/, "ต้องการลบจริงหรือไม่");
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name2)).toBe(false);
});

/**
 * Creditor: full field coverage (tax id, credit terms, WHT, bank account, address) plus an edit
 * that OMITS most fields (only the name is touched) to prove partial-update does not wipe the
 * other fields already saved — the exact bug class fixed in the recent warehouse redesign
 * (fetch-existing-then-merge). Verifies via the backend API directly (not just UI) that omitted
 * fields survive the second PUT.
 */
test("creditor — create with full fields, partial-update does not wipe fields, delete", async ({
  page,
  request,
}) => {
  const uid = Date.now().toString().slice(-6);
  const code = `E2ECR${uid}`;
  const name = `เจ้าหนี้E2E${uid}`;
  const name2 = `เจ้าหนี้E2Eแก้ไข${uid}`;
  const MAINAPI = "http://localhost:8888";

  await loginOnly(page);
  await searchMenuAndOpen(page, "เจ้าหนี้", /^เจ้าหนี้$/);
  await expect(page.locator("body")).toContainText(/ผู้จำหน่าย/i);

  // CREATE with full field set
  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัสผู้จำหน่าย/, code);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name);
  await fillFieldNearLabel(page, /เลขผู้เสียภาษี/, "1234567890123");
  await fillFieldNearLabel(page, /เครดิต \(วัน\)/, "30");
  await fillFieldNearLabel(page, /วงเงินเครดิต/, "150000");
  await fillFieldNearLabel(page, /อัตราภาษีหัก/, "3");
  // Bank account is a repeatable "bank-accounts" row editor (spec section 3), not a flat field —
  // add a row first, then fill its 3 inputs.
  await clickButtonByText(page, /^เพิ่มบัญชีธนาคาร$/);
  await page.waitForTimeout(500);
  await fillFieldNearLabel(page, /^ธนาคาร$/, "SCB");
  await fillFieldNearLabel(page, /เลขที่บัญชี/, "1112223334");
  await fillFieldNearLabel(page, /ชื่อบัญชี/, name);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, name)).toBe(true);

  // Verify full field set actually persisted via the real backend API (not just UI echo)
  const auth = JSON.parse((await page.evaluate(() => localStorage.getItem("bc_auth")))!) as {
    token: string;
  };
  const H = { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" };
  const listRes = await (await request.get(`${MAINAPI}/debtaccount/creditor?limit=50`, { headers: H })).json();
  const created = (
    listRes.data as {
      guidfixed: string;
      code: string;
      taxid?: string;
      creditday?: number;
      creditlimitsatang?: number;
      whtrate?: number;
      bankaccounts?: { bankcode: string; accountnumber: string }[];
      auth?: unknown;
    }[]
  ).find((c) => c.code === code);
  expect(created).toBeTruthy();
  const guid = created!.guidfixed;
  expect(created!.taxid).toBe("1234567890123");
  expect(created!.creditday).toBe(30);
  expect(created!.creditlimitsatang).toBe(15000000); // 150,000 baht * 100
  expect(created!.whtrate).toBe(3);
  expect(created!.bankaccounts?.[0]?.bankcode).toBe("SCB");
  expect(created!.bankaccounts?.[0]?.accountnumber).toBe("1112223334");
  expect(created!.auth).toBeUndefined(); // Auth field was removed from Creditor model — must not reappear

  // PARTIAL UPDATE via UI: edit only the name, save — other fields must survive
  await clickRowActionButton(page, /^แก้ไข$/, code);
  await page.waitForTimeout(1500);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name2);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, name2)).toBe(true);

  const afterPartialUpdate = await (
    await request.get(`${MAINAPI}/debtaccount/creditor/${guid}`, { headers: H })
  ).json();
  expect(afterPartialUpdate.data.taxid).toBe("1234567890123");
  expect(afterPartialUpdate.data.creditday).toBe(30);
  expect(afterPartialUpdate.data.creditlimitsatang).toBe(15000000);
  expect(afterPartialUpdate.data.whtrate).toBe(3);
  expect(afterPartialUpdate.data.bankaccounts?.[0]?.bankcode).toBe("SCB");

  // DELETE
  await clickRowActionButton(page, /^ลบ$/, code);
  await page.waitForTimeout(1000);
  expect(await bodyHasText(page, "ต้องการลบ")).toBe(true);
  await clickRowActionButton(page, /^ลบ$/, "ต้องการลบจริงหรือไม่");
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name2)).toBe(false);

  const afterDelete = await request.get(`${MAINAPI}/debtaccount/creditor/${guid}`, { headers: H });
  expect([400, 404]).toContain(afterDelete.status());
});

test("debtor — create with full fields, edit, delete", async ({ page, request }) => {
  const uid = Date.now().toString().slice(-6);
  const code = `E2EDR${uid}`;
  const name = `ลูกหนี้E2E${uid}`;
  const name2 = `ลูกหนี้E2Eแก้ไข${uid}`;
  const MAINAPI = "http://localhost:8888";

  await loginOnly(page);
  await searchMenuAndOpen(page, "ลูกหนี้", /^ลูกหนี้$/);
  await expect(page.locator("body")).toContainText(/ลูกค้า/i);

  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัสลูกค้า/, code);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name);
  await fillFieldNearLabel(page, /เครดิต \(วัน\)/, "15");
  await fillFieldNearLabel(page, /วงเงินเครดิต/, "50000");
  await fillFieldNearLabel(page, /ระดับราคา/, "VIP");
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, name)).toBe(true);

  const auth = JSON.parse((await page.evaluate(() => localStorage.getItem("bc_auth")))!) as {
    token: string;
  };
  const H = { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" };
  const listRes = await (await request.get(`${MAINAPI}/debtaccount/debtor?limit=50`, { headers: H })).json();
  const created = (
    listRes.data as {
      guidfixed: string;
      code: string;
      creditday?: number;
      creditlimitsatang?: number;
      pricelevel?: string;
    }[]
  ).find((d) => d.code === code);
  expect(created).toBeTruthy();
  const guid = created!.guidfixed;
  expect(created!.creditday).toBe(15);
  expect(created!.creditlimitsatang).toBe(5000000); // 50,000 baht * 100
  expect(created!.pricelevel).toBe("VIP");

  // EDIT (name only) — same partial-update survival check as creditor
  await clickRowActionButton(page, /^แก้ไข$/, code);
  await page.waitForTimeout(1500);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name2);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, name2)).toBe(true);

  const afterEdit = await (await request.get(`${MAINAPI}/debtaccount/debtor/${guid}`, { headers: H })).json();
  expect(afterEdit.data.creditday).toBe(15);
  expect(afterEdit.data.creditlimitsatang).toBe(5000000);
  expect(afterEdit.data.pricelevel).toBe("VIP");

  // DELETE
  await clickRowActionButton(page, /^ลบ$/, code);
  await page.waitForTimeout(1000);
  expect(await bodyHasText(page, "ต้องการลบ")).toBe(true);
  await clickRowActionButton(page, /^ลบ$/, "ต้องการลบจริงหรือไม่");
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name2)).toBe(false);

  const afterDelete = await request.get(`${MAINAPI}/debtaccount/debtor/${guid}`, { headers: H });
  expect([400, 404]).toContain(afterDelete.status());
});

/**
 * Regression for the "Trade Partners Redesign" task (2026-07-05): generic `thai-address` field
 * type, repeatable `bank-accounts` field type, and cross-collection taxid linking.
 *
 * Multi-bank round-trip: this is the exact regression test for the previously-verified
 * destroy-bug — the old creditor UI folded `bankaccounts[0]` into 3 flat fields and unconditionally
 * overwrote the whole array on every save (system-settings-screen.tsx, old `buildPayload` creditor
 * branch). A record with 2+ accounts lost accounts 1..n on any save that didn't touch them.
 */
test("creditor — multi-bank round-trip survives an unrelated-field edit (destroy-bug regression)", async ({
  page,
  request,
}) => {
  const uid = Date.now().toString().slice(-6);
  const code = `E2ECRBK${uid}`;
  const name = `เจ้าหนี้บัญชีE2E${uid}`;
  const MAINAPI = "http://localhost:8888";

  await loginOnly(page);
  await searchMenuAndOpen(page, "เจ้าหนี้", /^เจ้าหนี้$/);
  await expect(page.locator("body")).toContainText(/ผู้จำหน่าย/i);

  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัสผู้จำหน่าย/, code);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name);

  // Add 2 bank-account rows via the repeatable "bank-accounts" field editor.
  await clickButtonByText(page, /^เพิ่มบัญชีธนาคาร$/);
  await page.waitForTimeout(300);
  await clickButtonByText(page, /^เพิ่มบัญชีธนาคาร$/);
  await page.waitForTimeout(500);

  await page.evaluate(() => {
    const setValue = (el: HTMLInputElement, v: string) => {
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
      setter.call(el, v);
      el.dispatchEvent(new Event("input", { bubbles: true }));
      el.dispatchEvent(new Event("change", { bubbles: true }));
    };
    const byLabel = (text: string) =>
      [...document.querySelectorAll<HTMLInputElement>("input")].filter(
        (i) => (i.closest("label")?.innerText ?? "").trim() === text,
      );
    const banks = byLabel("ธนาคาร");
    const accnos = byLabel("เลขที่บัญชี");
    const accnames = byLabel("ชื่อบัญชี");
    if (banks.length < 2) throw new Error(`expected 2 bank rows, got ${banks.length}`);
    setValue(banks[0], "KTB");
    setValue(accnos[0], "111-1-11111-1");
    setValue(accnames[0], "ACC ONE");
    setValue(banks[1], "SCB");
    setValue(accnos[1], "222-2-22222-2");
    setValue(accnames[1], "ACC TWO");
  });

  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2500);
  expect(await bodyHasText(page, name)).toBe(true);

  const auth = JSON.parse((await page.evaluate(() => localStorage.getItem("bc_auth")))!) as {
    token: string;
  };
  const H = { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" };
  const listRes = await (await request.get(`${MAINAPI}/debtaccount/creditor?limit=50`, { headers: H })).json();
  const created = (
    listRes.data as { guidfixed: string; code: string; bankaccounts?: { bankcode: string; accountnumber: string; accountname: string }[] }[]
  ).find((c) => c.code === code);
  expect(created).toBeTruthy();
  const guid = created!.guidfixed;
  expect(created!.bankaccounts?.length).toBe(2);
  expect(created!.bankaccounts?.[0]).toMatchObject({ bankcode: "KTB", accountnumber: "111-1-11111-1", accountname: "ACC ONE" });
  expect(created!.bankaccounts?.[1]).toMatchObject({ bankcode: "SCB", accountnumber: "222-2-22222-2", accountname: "ACC TWO" });

  // Edit an UNRELATED field only (email) and save — both bank rows must survive (destroy-bug regression).
  await clickRowActionButton(page, /^แก้ไข$/, code);
  await page.waitForTimeout(1500);
  await fillFieldNearLabel(page, /อีเมล/, "unrelated-edit-test@example.com");
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);

  const afterUnrelatedEdit = await (
    await request.get(`${MAINAPI}/debtaccount/creditor/${guid}`, { headers: H })
  ).json();
  expect(afterUnrelatedEdit.data.email).toBe("unrelated-edit-test@example.com");
  expect(afterUnrelatedEdit.data.bankaccounts?.length).toBe(2);

  // Delete row 1 (KTB) via the per-row "ลบ" button, save, verify length 1 with the correct row left.
  await clickRowActionButton(page, /^แก้ไข$/, code);
  await page.waitForTimeout(1500);
  await page.evaluate(() => {
    const bankLabel = [...document.querySelectorAll("label")].find((l) => l.textContent?.trim() === "ธนาคาร");
    if (!bankLabel) throw new Error("bank row label not found");
    let row: HTMLElement | null = bankLabel.closest("div");
    for (let i = 0; i < 5 && row; i++) {
      const delBtn = [...row.querySelectorAll<HTMLElement>("button")].find((b) => b.textContent?.trim() === "ลบ");
      if (delBtn) {
        delBtn.click();
        return;
      }
      row = row.parentElement;
    }
    throw new Error("delete button not found for bank row");
  });
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);

  const afterRowDelete = await (
    await request.get(`${MAINAPI}/debtaccount/creditor/${guid}`, { headers: H })
  ).json();
  expect(afterRowDelete.data.bankaccounts?.length).toBe(1);
  expect(afterRowDelete.data.bankaccounts?.[0]).toMatchObject({ bankcode: "SCB", accountnumber: "222-2-22222-2" });

  // Cleanup
  await clickRowActionButton(page, /^ลบ$/, code);
  await page.waitForTimeout(1000);
  await clickRowActionButton(page, /^ลบ$/, "ต้องการลบจริงหรือไม่");
  await page.waitForTimeout(1500);
});

/**
 * Generic `thai-address` field cascade on the creditor billing address: forward
 * (province -> district -> subdistrict -> auto zipcode) and reverse (zipcode -> auto province +
 * district, ambiguous-subdistrict narrowing with no crash). Verifies the persisted Mongo-backed
 * document via the API, not just the UI echo.
 */
test("creditor — Thai address cascade forward and reverse, persists correctly", async ({ page, request }) => {
  const uid = Date.now().toString().slice(-6);
  const code = `E2ECRAD${uid}`;
  const name = `เจ้าหนี้ที่อยู่E2E${uid}`;
  const MAINAPI = "http://localhost:8888";

  await loginOnly(page);
  await searchMenuAndOpen(page, "เจ้าหนี้", /^เจ้าหนี้$/);
  await expect(page.locator("body")).toContainText(/ผู้จำหน่าย/i);

  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัสผู้จำหน่าย/, code);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name);

  // Forward cascade: select province -> district -> subdistrict on the billing-address block
  // (first "จังหวัด" select on the page — actual-address is the second).
  const forwardResult = await page.evaluate(async () => {
    const wait = (ms: number) => new Promise((r) => setTimeout(r, ms));
    const provinceSelects = [...document.querySelectorAll<HTMLSelectElement>("select")].filter((s) =>
      /จังหวัด/.test(s.closest("label")?.innerText ?? ""),
    );
    const setSelect = (sel: HTMLSelectElement, value: string) => {
      const setter = Object.getOwnPropertyDescriptor(window.HTMLSelectElement.prototype, "value")!.set!;
      setter.call(sel, value);
      sel.dispatchEvent(new Event("change", { bubbles: true }));
    };
    setSelect(provinceSelects[0], "50"); // เชียงใหม่
    await wait(500);
    const districtSelects = [...document.querySelectorAll<HTMLSelectElement>("select")].filter((s) =>
      /อำเภอ/.test(s.closest("label")?.innerText ?? ""),
    );
    const districtOpt = [...districtSelects[0].options].find((o) => /เมืองเชียงใหม่/.test(o.text));
    if (!districtOpt) throw new Error("district option not found");
    setSelect(districtSelects[0], districtOpt.value);
    await wait(500);
    const subdistrictSelects = [...document.querySelectorAll<HTMLSelectElement>("select")].filter((s) =>
      /ตำบล|แขวง/.test(s.closest("label")?.innerText ?? ""),
    );
    const subOpt = [...subdistrictSelects[0].options].find((o) => /ศรีภูมิ/.test(o.text));
    if (!subOpt) throw new Error("subdistrict option not found");
    setSelect(subdistrictSelects[0], subOpt.value);
    await wait(500);
    const zipInputs = [...document.querySelectorAll<HTMLInputElement>("input")].filter((i) =>
      /รหัสไปรษณีย์/.test(i.closest("label")?.innerText ?? ""),
    );
    return { zip: zipInputs[0].value };
  });
  expect(forwardResult.zip).toBe("50200");

  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);

  const auth = JSON.parse((await page.evaluate(() => localStorage.getItem("bc_auth")))!) as { token: string };
  const H = { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" };
  const listRes = await (await request.get(`${MAINAPI}/debtaccount/creditor?limit=50`, { headers: H })).json();
  const created = (
    listRes.data as {
      guidfixed: string;
      code: string;
      addressforbilling?: { provincecode?: string; districtcode?: string; subdistrictcode?: string; zipcode?: string };
    }[]
  ).find((c) => c.code === code);
  expect(created).toBeTruthy();
  expect(created!.addressforbilling?.provincecode).toBe("50");
  expect(created!.addressforbilling?.districtcode).toBe("5001");
  expect(created!.addressforbilling?.subdistrictcode).toBe("500101");
  expect(created!.addressforbilling?.zipcode).toBe("50200");

  // Reverse cascade: re-open, type an exact-match zipcode, assert province+district auto-fill and
  // the ambiguous-subdistrict hint appears (no crash) rather than picking one silently.
  await clickRowActionButton(page, /^แก้ไข$/, code);
  await page.waitForTimeout(1500);
  const reverseResult = await page.evaluate(async () => {
    const wait = (ms: number) => new Promise((r) => setTimeout(r, ms));
    const zipInputs = [...document.querySelectorAll<HTMLInputElement>("input")].filter((i) =>
      /รหัสไปรษณีย์/.test(i.closest("label")?.innerText ?? ""),
    );
    const setInput = (el: HTMLInputElement, v: string) => {
      const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
      setter.call(el, v);
      el.dispatchEvent(new Event("input", { bubbles: true }));
      el.dispatchEvent(new Event("change", { bubbles: true }));
    };
    // The record already has subdistrict 500101 selected from the earlier forward-cascade step
    // (same zipcode's set) — the app correctly shows "เลือกแล้ว: ..." rather than an ambiguous
    // hint once a matching subdistrict is already selected. Clear the zipcode first so re-typing
    // it re-triggers the ambiguous-match path with no subdistrict pre-selected.
    setInput(zipInputs[0], "");
    await wait(300);
    setInput(zipInputs[0], "50200");
    await wait(600);
    const provinceSelects = [...document.querySelectorAll<HTMLSelectElement>("select")].filter((s) =>
      /จังหวัด/.test(s.closest("label")?.innerText ?? ""),
    );
    const districtSelects = [...document.querySelectorAll<HTMLSelectElement>("select")].filter((s) =>
      /อำเภอ/.test(s.closest("label")?.innerText ?? ""),
    );
    return {
      province: provinceSelects[0].value,
      district: districtSelects[0].value,
      hasAmbiguousHint: /พบ\s*\d+\s*ตำบล/.test(document.body.innerText),
    };
  });
  expect(reverseResult.province).toBe("50");
  expect(reverseResult.district).toBe("5001");
  expect(reverseResult.hasAmbiguousHint).toBe(true);

  // Cleanup — cancel out of edit (no save needed) then delete.
  await page.goto("/creditor", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await clickRowActionButton(page, /^ลบ$/, code);
  await page.waitForTimeout(1000);
  await clickRowActionButton(page, /^ลบ$/, "ต้องการลบจริงหรือไม่");
  await page.waitForTimeout(1500);
});

/**
 * Taxid cross-link: creating a creditor and a debtor with the same taxid shows an informational
 * link badge on both screens (spec section 1.2). Creating a second creditor with the same
 * taxid+branchnumber shows a non-blocking duplicate warning and still saves successfully.
 */
test("creditor/debtor — taxid link badge and same-collection duplicate warning", async ({ page, request }) => {
  const uid = Date.now().toString().slice(-6);
  const creditorCode = `E2ECRTX${uid}`;
  const debtorCode = `E2EDRTX${uid}`;
  const dupCode = `E2ECRTX2${uid}`;
  const taxid = "1199922223"; // 10-digit placeholder base
  const fullTaxid = (taxid + uid).padEnd(13, "1").slice(0, 13);
  const MAINAPI = "http://localhost:8888";

  await loginOnly(page);
  await searchMenuAndOpen(page, "เจ้าหนี้", /^เจ้าหนี้$/);
  await expect(page.locator("body")).toContainText(/ผู้จำหน่าย/i);

  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัสผู้จำหน่าย/, creditorCode);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, `เจ้าหนี้taxidE2E${uid}`);
  await fillFieldNearLabel(page, /เลขผู้เสียภาษี/, fullTaxid);
  await fillFieldNearLabel(page, /รหัสสาขาภาษี/, "1");
  await page.waitForTimeout(1000);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);

  // Debtor with the same taxid -> expect the cross-collection link badge, save must still succeed.
  await page.goto("/debtor", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัสลูกค้า/, debtorCode);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, `ลูกหนี้taxidE2E${uid}`);
  await fillFieldNearLabel(page, /เลขผู้เสียภาษี/, fullTaxid);
  await page.waitForTimeout(1200);
  expect(await bodyHasText(page, "คู่ค้ารายเดียวกัน")).toBe(true);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, `ลูกหนี้taxidE2E${uid}`)).toBe(true);

  // Second creditor with the same taxid + branchnumber -> same-collection duplicate warning,
  // non-blocking (save still succeeds).
  await page.goto("/creditor", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัสผู้จำหน่าย/, dupCode);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, `เจ้าหนี้ซ้ำE2E${uid}`);
  await fillFieldNearLabel(page, /เลขผู้เสียภาษี/, fullTaxid);
  await fillFieldNearLabel(page, /รหัสสาขาภาษี/, "1");
  await page.waitForTimeout(1200);
  expect(await bodyHasText(page, "ซ้ำกับรหัส")).toBe(true);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, `เจ้าหนี้ซ้ำE2E${uid}`)).toBe(true);

  // Cleanup all 3 records.
  const auth = JSON.parse((await page.evaluate(() => localStorage.getItem("bc_auth")))!) as { token: string };
  const H = { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" };
  const creditors = (
    await (await request.get(`${MAINAPI}/debtaccount/creditor?limit=50`, { headers: H })).json()
  ).data as { guidfixed: string; code: string }[];
  for (const c of creditors.filter((c) => c.code === creditorCode || c.code === dupCode)) {
    await request.delete(`${MAINAPI}/debtaccount/creditor/${c.guidfixed}?holdingcode=bctest01`, { headers: H });
  }
  const debtors = (
    await (await request.get(`${MAINAPI}/debtaccount/debtor?limit=50`, { headers: H })).json()
  ).data as { guidfixed: string; code: string }[];
  for (const d of debtors.filter((d) => d.code === debtorCode)) {
    await request.delete(`${MAINAPI}/debtaccount/debtor/${d.guidfixed}?holdingcode=bctest01`, { headers: H });
  }
});
