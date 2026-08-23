import { test, expect, type Page } from "@playwright/test";

/**
 * Regression test for the "ช่องทางขาย ราคา และขนส่ง" (Sales Channel/Pricing/Shipping) menu group.
 *
 * Verified via the uat-crud-mongo skill 2026-07-04 — UI + API + MongoDB (direct collection query,
 * including soft-delete `deletedat` filtering for salechannel/transportchannel, and a real
 * hard-delete for channelprice's atlas-generic screen) all agree for create/update/delete on all 3
 * screens below.
 *
 * The "งาน โครงการ และศูนย์ต้นทุน" (Jobs, Projects & Cost Centers) menu group and its backend
 * modules (`backend/internal/organization/costcenter`, `backend/internal/organization/jobproject`)
 * were fully removed in the 2026-07-05 regression. The former cost-center CRUD test that
 * used to live in this file was removed along with it.
 *
 * `/salechannelscreen` -> backend/internal/channel/salechannel (collection salechannel, soft
 * delete). `/transportchannelscreen` -> backend/internal/channel/transportchannel (collection
 * transportchannel, soft delete). `/channelprice` -> generic "atlas" screen kind (collection
 * productchannelprices, hard delete — pre-existing screen, untouched by this task).
 *
 * DOM notes (same as trade-partners-crud.spec.ts / master-brand-crud.spec.ts): code/name labels are
 * siblings of their inputs, not ancestors; delete uses an in-page confirm modal, not a native dialog;
 * row action buttons are scoped by walking up from the button to the row containing the row's code.
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

async function loginOnly(page: Page) {
  await page.goto("/", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /เข้าทดสอบระบบ/);
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /test(?!\d)/i);
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
}

test("salechannelscreen — create, edit, delete", async ({ page }) => {
  const uid = Date.now().toString().slice(-6);
  const code = `E2ESC${uid}`;
  const name = `ช่องทางขายE2E${uid}`;
  const name2 = `ช่องทางขายE2Eแก้ไข${uid}`;

  await loginOnly(page);
  await page.goto("/salechannelscreen", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await expect(page.locator("body")).toContainText(/ช่องทางขาย/);

  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัสช่องทางขาย/, code);
  await fillFieldNearLabel(page, /^ชื่อช่องทางขาย/, name);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name)).toBe(true);

  await clickRowActionButton(page, /^แก้ไข$/, code);
  await page.waitForTimeout(1500);
  await fillFieldNearLabel(page, /^ชื่อช่องทางขาย/, name2);
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

test("transportchannelscreen — create, edit, delete", async ({ page }) => {
  const uid = Date.now().toString().slice(-6);
  const code = `E2ETC${uid}`;
  const name = `ช่องทางขนส่งE2E${uid}`;
  const name2 = `ช่องทางขนส่งE2Eแก้ไข${uid}`;

  await loginOnly(page);
  await page.goto("/transportchannelscreen", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await expect(page.locator("body")).toContainText(/ช่องทางขนส่ง/);

  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัสช่องทางขนส่ง/, code);
  await fillFieldNearLabel(page, /^ชื่อช่องทางขนส่ง/, name);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name)).toBe(true);

  await clickRowActionButton(page, /^แก้ไข$/, code);
  await page.waitForTimeout(1500);
  await fillFieldNearLabel(page, /^ชื่อช่องทางขนส่ง/, name2);
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

test("channelprice — create, edit, delete (pre-existing atlas screen, hard delete)", async ({ page }) => {
  const uid = Date.now().toString().slice(-6);
  const code = `E2ECP${uid}`;
  const name = `ราคาE2E${uid}`;
  const name2 = `ราคาE2Eแก้ไข${uid}`;

  await loginOnly(page);
  await page.goto("/channelprice", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  await expect(page.locator("body")).toContainText(/ราคาตามช่องทางขาย/);

  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัสราคา/, code);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name);
  await fillFieldNearLabel(page, /^ช่องทางขาย/, "SHOPEE");
  await fillFieldNearLabel(page, /^รหัสสินค้า/, "ITEM001");
  await fillFieldNearLabel(page, /^ราคาขาย/, "99");
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
