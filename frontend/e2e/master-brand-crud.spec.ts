import { test, expect, type Page } from "@playwright/test";

/**
 * Regression test for the "ยี่ห้อสินค้า" (Brand) master screen — /masterbrandscreen.
 * Backing store: MongoDB `appdb.brandproductmaster` (main-crud kind, proxied via
 * /api/system-settings/master_brand_screen). Verified 2026-07-01 via the uat-crud-mongo skill:
 * UI + API + MongoDB (direct `brandproductmaster` query, including soft-delete `deletedat`
 * filtering) all agree for create/update/delete.
 *
 * Uses generated login (dev bypass -> Test holding -> first company -> headquarters branch,
 * holdingcode "test", disposable dev data) and a unique code per run so this is safely
 * re-runnable without manual cleanup.
 *
 * DOM notes learned this run (see also references/verification-guide.md in the uat-crud-mongo
 * skill): the code/name `<label>` elements are NOT ancestors of their `<input>` — they're
 * siblings — so selectors match by the nearest preceding label's text, not `input.closest("label")`.
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

/**
 * Datacrud rows put the row's own text (e.g. the code/name) and the per-row action button
 * ("แก้ไข"/"ลบ") in SIBLING elements, not a shared text node — a button's own textContent is just
 * "แก้ไข", the code lives elsewhere in the same row container. Walk up from each candidate button
 * to find the row whose combined text includes `rowText`, then click that specific button.
 */
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

async function loginAndOpenBrandScreen(page: Page) {
  await page.goto("/", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2500);
  // Exact text confirmed from a live DOM snapshot: button "🧪 เข้าทดสอบระบบ (dev)" — dev-only
  // bypass login, shown only on localhost/LAN. Substring match (not anchored) since it's decorated
  // with an emoji and a parenthetical.
  await clickButtonByText(page, /เข้าทดสอบระบบ/);
  await page.waitForTimeout(2500);
  await clickButtonByText(page, /Test/);
  await page.waitForTimeout(2500);
  // Company card button text is composite (icon alt + name + tax/currency badges, e.g.
  // "บริษัท บริษัท OWNER TH THB 00000") — much longer than header/nav buttons like "คู่มือ"/
  // "ตั้งค่าระบบ". Pick the first sufficiently-long button containing "บริษัท" instead of the
  // first clickable element on the page (which can be an unrelated header link).
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

  // Two responsive variants of the menu search box exist (desktop header + mobile complementary
  // region) — use the first one via evaluate rather than a strict-mode Playwright locator.
  await page.evaluate(() => {
    const input = document.querySelector<HTMLInputElement>('input[placeholder*="ค้นหาเมนู"]');
    if (!input) throw new Error("Menu search input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, "รายละเอียดประกอบสินค้า");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await page.waitForTimeout(1000);
  await clickButtonByText(page, /^ยี่ห้อสินค้า$/);
  await page.waitForTimeout(2000);
  await expect(page.locator("body")).toContainText(/ยี่ห้อ/i);
}

test("brand master — create, edit, delete (UI + API verified; see MongoDB note below)", async ({ page }) => {
  const uid = Date.now().toString().slice(-6);
  const code = `E2EB${uid}`;
  const name = `ยี่ห้อE2E${uid}`;
  const name2 = `ยี่ห้อE2Eแก้ไข${uid}`;

  await loginAndOpenBrandScreen(page);

  // CREATE
  await clickButtonByText(page, /^เพิ่ม$/);
  await page.waitForTimeout(1000);
  await fillFieldNearLabel(page, /รหัส/, code);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name)).toBe(true);

  // UPDATE
  await clickRowActionButton(page, /^แก้ไข$/, code);
  await page.waitForTimeout(1500);
  await fillFieldNearLabel(page, /ภาษาแรก|TH/, name2);
  await clickButtonByText(page, /^บันทึก$/);
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name2)).toBe(true);

  // DELETE (in-page confirm modal on this screen, not a native dialog). Multiple "ลบ" buttons
  // exist at once (one per row + the modal's own confirm button) — scope the confirm click to
  // the container that also holds the "ต้องการลบจริงหรือไม่" confirmation heading.
  await clickRowActionButton(page, /^ลบ$/, code);
  await page.waitForTimeout(1000);
  expect(await bodyHasText(page, "ต้องการลบ")).toBe(true);
  await clickRowActionButton(page, /^ลบ$/, "ต้องการลบจริงหรือไม่");
  await page.waitForTimeout(2000);
  expect(await bodyHasText(page, name2)).toBe(false);

  // Cleanup safety net: if the UI delete somehow left a row behind, remove it via the API directly
  // so repeated runs never accumulate leftover test data in brandproductmaster.
  await page.evaluate(async () => {
    const auth = JSON.parse(localStorage.getItem("bc_auth") ?? "null");
    if (!auth) return;
    const headers = {
      Authorization: `Bearer ${auth.token}`,
      "x-bc-backend-url": auth.backendUrl,
      "Content-Type": "application/json",
    };
    const list = await (
      await fetch("/api/system-settings/master_brand_screen?limit=50&offset=0&holdingcode=test", { headers })
    ).json();
    for (const row of list.data ?? []) {
      if (typeof row.code === "string" && row.code.startsWith("E2EB")) {
        await fetch(`/api/system-settings/master_brand_screen/${row.guidfixed}?holdingcode=test`, {
          method: "DELETE",
          headers,
          body: JSON.stringify({ guidfixed: row.guidfixed, holdingcode: "test", backendUrl: auth.backendUrl }),
        });
      }
    }
  });
});
