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
  await clickContainingText(page, "20ยังไม่กำหนดชื่อกลุ่ม");
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
  await clickContainingText(page, "20ยังไม่กำหนดชื่อกลุ่ม");
  await page.waitForTimeout(1500);

  expect(await bodyHasText(page, "กดปุ่ม 'เพิ่มหมวดหลัก' ด้านบน")).toBe(true);
  const hasAddRootButton = await page.evaluate(() =>
    [...document.querySelectorAll("button,a")].some(
      (e) => (e.textContent ?? "").includes("เพิ่มหมวดหลัก") && (e as HTMLElement).offsetParent !== null,
    ),
  );
  expect(hasAddRootButton).toBe(true);
});
