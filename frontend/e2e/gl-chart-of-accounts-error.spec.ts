import { expect, test, type Page, type TestInfo } from "@playwright/test";

/**
 * Evidence spec for the ผังบัญชี (/gl/chartofaccounts) error UX + premium screenshots.
 *
 * What it proves:
 *  1. A rejected save (duplicate รหัสบัญชี taken from the rendered list) shows exactly ONE visible
 *     Thai `role="alert"` in the editor pane — never English/Mongo/HTTP text.
 *  2. The editor keeps what the user typed and focuses ช่องรหัสบัญชี (no page jump).
 *  3. The rejected save adds ZERO console errors: the BFF relays a user-caused 4xx carrying a
 *     machine `code` as HTTP 200 + `success:false` (frontend/src/lib/workspace-api.ts).
 *  4. Screenshots: light + dark (toggled by the real `button.theme-toggle`) ×
 *     1600/1280/1024/768 portrait, plus hover / focus / disabled / error states.
 *
 * The duplicate attempt is rejected by the API, so this spec never creates a document.
 */

const widths = [1600, 1280, 1024, 768];
// Dev-mode noise that exists before the app is even logged in (React DevTools banner is INFO level,
// these two are ERROR level from the login page): external Google Sign-In iframe + pre-login refresh.
const preExistingNoise = /accounts\.google\.com|GSI_LOGGER|api\/auth\/refresh/;

async function clickByText(page: Page, pattern: RegExp, timeout = 20_000) {
  const deadline = Date.now() + timeout;
  for (;;) {
    const clicked = await page.evaluate((source) => {
      const re = new RegExp(source);
      const el = [...document.querySelectorAll<HTMLElement>("button,a,[role=button]")].find(
        (candidate) => re.test((candidate.textContent ?? "").replace(/\s+/g, " ").trim()) && candidate.offsetParent !== null,
      );
      if (el) { el.click(); return true; }
      return false;
    }, pattern.source);
    if (clicked) return;
    if (Date.now() > deadline) throw new Error(`No clickable element matching ${pattern}`);
    await page.waitForTimeout(300);
  }
}

async function loginAndOpenChartOfAccounts(page: Page) {
  await page.goto("/", { waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2500);
  await page.getByRole("button", { name: "ทดลองใช้ระบบ (Demo)", exact: true }).click();
  await page.waitForTimeout(3000);
  await page.getByRole("button", { name: /กลุ่มกิจการรุ่งเรือง/ }).first().click();
  await page.waitForTimeout(3000);
  await page.getByRole("button").filter({ hasText: /บริษัท|จำกัด/ }).first().click();
  await page.waitForTimeout(3000);
  await page.getByRole("button").filter({ hasText: /สำนักงาน|สาขา/ }).first().click();
  await page.waitForTimeout(3500);
  await page.evaluate(() => {
    const input = [...document.querySelectorAll<HTMLInputElement>("input")].find((el) => /ค้นหาเมนู/.test(el.placeholder));
    if (!input) throw new Error("menu search input not found");
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")!.set!;
    setter.call(input, "ผังบัญชี");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await page.waitForTimeout(1200);
  await page.getByRole("button").filter({ hasText: /^ผังบัญชี/ }).first().click();
  await page.waitForTimeout(4000);
}

test("ผังบัญชี rejected save: one Thai alert, values kept, zero new console errors", async ({ page }, testInfo: TestInfo) => {
  test.setTimeout(240_000);
  const consoleErrors: string[] = [];
  const pageErrors: string[] = [];
  page.on("console", (message) => { if (message.type() === "error") consoleErrors.push(message.text()); });
  page.on("pageerror", (error) => pageErrors.push(String(error)));

  await loginAndOpenChartOfAccounts(page);
  await expect(page.locator('input[placeholder="ค้นหารหัสหรือชื่อ"]')).toBeVisible({ timeout: 30_000 });

  // 1) read an EXISTING account code from the rendered list (never hardcoded)
  const firstRow = page.locator("table tbody tr").first();
  await firstRow.waitFor({ state: "visible", timeout: 20_000 });
  const existingCode = (await firstRow.locator("td").first().innerText()).trim();
  expect(existingCode.length, "an existing code must be readable from the list").toBeGreaterThan(0);

  // 2) open the add form, type a duplicate code + Thai name, submit
  await clickByText(page, /เพิ่มรายการ/);
  const codeInput = page.locator('input[data-field="accountcode"]');
  await codeInput.waitFor({ state: "visible", timeout: 20_000 });
  await codeInput.fill(existingCode);
  await page.locator('input[data-field="accountnameth"]').fill("ทดสอบรหัสซ้ำ (E2E)");
  await page.waitForTimeout(500);

  const pagerBefore = (await page.getByText(/รายการ · หน้า/).first().innerText()).trim();
  const baseline = consoleErrors.filter((text) => !preExistingNoise.test(text)).length;
  const glErrorsBefore = consoleErrors.filter((text) => /api\/gl/.test(text)).length;
  await clickByText(page, /บันทึกข้อมูล/);
  await page.waitForTimeout(3500);

  // 3) exactly one visible Thai alert (scoped: Next.js also renders `#__next-route-announcer__` with
  // role="alert" and the Thai document title), no provider/English text
  const alerts = page.locator('[role="alert"]:visible:not(#__next-route-announcer__ *)').filter({ hasText: /[\u0E01-\u0E5B]/ });
  await expect(alerts).toHaveCount(1, { timeout: 20_000 });
  const alertText = (await alerts.first().innerText()).replace(/\s+/g, " ").trim();
  expect(alertText).toMatch(/[\u0E01-\u0E5B]/);
  expect(alertText).not.toMatch(/E11000|duplicate key|Mongo|MongoServerError|Failed to load resource|Unhandled/);

  // 4) values kept + focus moved to รหัสบัญชี without scrolling the page
  expect(await codeInput.inputValue()).toBe(existingCode);
  expect(await page.locator('input[data-field="accountnameth"]').inputValue()).toBe("ทดสอบรหัสซ้ำ (E2E)");
  await expect(codeInput, "focus must land on รหัสบัญชี after the rejected save").toBeFocused({ timeout: 10_000 });

  // 5) the rejected save added no console error and no unhandled rejection
  const after = consoleErrors.filter((text) => !preExistingNoise.test(text)).length;
  expect(after, `console errors after the failed save:\n${consoleErrors.join("\n")}`).toBe(baseline);
  expect(consoleErrors.filter((text) => /api\/gl/.test(text)).length).toBe(glErrorsBefore);
  expect(pageErrors).toEqual([]);
  const pagerAfter = (await page.getByText(/รายการ · หน้า/).first().innerText()).trim();
  expect(pagerAfter, "the rejected save must not change the account count").toBe(pagerBefore);
  console.log("OBSERVED alert=" + alertText + " | consoleBaseline=" + baseline + " consoleAfter=" + after + " consoleRaw=" + consoleErrors.length + " pageErrors=" + pageErrors.length + " | pagerBefore=" + pagerBefore + " pagerAfter=" + pagerAfter);

  // 6) screenshots: error state, light + dark via the REAL theme button × 4 viewports
  const shots: string[] = [];
  const shot = async (name: string) => {
    const file = testInfo.outputPath(`${name}.png`);
    await page.screenshot({ path: file });
    shots.push(file);
  };
  for (const dark of [false, true]) {
    const isDark = await page.evaluate(() => document.documentElement.dataset.theme === "dark" || document.documentElement.classList.contains("dark"));
    if (isDark !== dark) {
      await page.locator("button.theme-toggle").first().click();
      await page.waitForTimeout(700);
    }
    const theme = dark ? "dark" : "light";
    for (const width of widths) {
      await page.setViewportSize({ width, height: width === 768 ? 1024 : 900 });
      await page.waitForTimeout(450);
      await shot(`chartofaccounts-error-${theme}-${width}`);
    }
  }

  // 7) hover / focus / disabled states (light, 1280)
  await page.setViewportSize({ width: 1280, height: 900 });
  const nowDark = await page.evaluate(() => document.documentElement.dataset.theme === "dark" || document.documentElement.classList.contains("dark"));
  if (nowDark) { await page.locator("button.theme-toggle").first().click(); await page.waitForTimeout(700); }
  await codeInput.hover();
  await page.waitForTimeout(350);
  await shot("chartofaccounts-state-hover-1280");
  await page.locator('input[data-field="accountnameth"]').focus();
  await page.waitForTimeout(350);
  await shot("chartofaccounts-state-focus-1280");
  await page.evaluate(() => { document.querySelector<HTMLInputElement>('input[data-field="accountcode"]')!.disabled = true; });
  await page.waitForTimeout(350);
  await shot("chartofaccounts-state-disabled-1280");

  // EN name (CHAMP Name2 parity): the visible label text and the accessible name must be identical,
  // and the field must be reachable in the add form.
  const enInput = page.locator('input[data-field="accountnameen"]');
  await expect(enInput).toBeVisible();
  await expect(page.getByRole("textbox", { name: "ชื่อบัญชีภาษาอังกฤษ", exact: true })).toHaveCount(1);
  await expect(page.getByText("ชื่อบัญชีภาษาอังกฤษ", { exact: true })).toBeVisible();
  await enInput.focus();
  await expect(enInput).toBeFocused();

  // the same AccountFields component renders in the view card and in the edit form
  await firstRow.click();
  await page.waitForTimeout(1500);
  await expect(page.locator('input[data-field="accountnameen"]'), "view card").toBeVisible();
  await clickByText(page, /แก้ไข/);
  await page.waitForTimeout(1500);
  await expect(page.locator('input[data-field="accountnameen"]'), "edit form").toBeVisible();
  await expect(page.getByRole("textbox", { name: "ชื่อบัญชีภาษาอังกฤษ", exact: true })).toHaveCount(1);
  console.log(`SCREENSHOTS(${shots.length}):\n${shots.join("\n")}`);
  expect(shots.length).toBe(11);
});
