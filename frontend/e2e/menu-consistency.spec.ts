import { expect, test } from "@playwright/test";

test("Thai menu names and pending badges remain consistent across themes and viewports", async ({ page }, testInfo) => {
  test.setTimeout(90_000);
  const errors: string[] = [];
  await page.goto("/");
  await page.getByRole("button", { name: "ทดลองใช้ระบบ (Demo)", exact: true }).click();
  await page.getByRole("button", { name: /^กลุ่มกิจการรุ่งเรือง/ }).click();
  await page.getByRole("button", { name: /^บริษัท รุ่งเรืองวัสดุ/ }).click();
  await page.getByRole("button", { name: /สำนักงานใหญ่/ }).first().click();
  await expect(page).toHaveURL(/\/menu$/);
  // Scope console verification to the menu; the external login widget is not under test.
  page.on("pageerror", (error) => errors.push(error.message));
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  await expect(page.locator("[data-menu-pending]").first()).toBeVisible();
  await expect(page.getByRole("button", { name: "เมนูบน", exact: true })).toHaveAttribute("aria-pressed", "true");
  await expect.poll(() => page.evaluate(() => localStorage.getItem("bc_menu_layout_mode"))).toBe("top");
  const search = page.getByPlaceholder("ค้นหาเมนู เอกสาร หรือหน้าจอ");
  await search.fill("ใบสั่งขาย");
  const saleOrder = page.getByRole("button").filter({ has: page.locator('[data-menu-pending="/transaction/saleorder"]') });
  await expect(saleOrder).toBeVisible();
  await expect(saleOrder.locator('[data-menu-pending="/transaction/saleorder"]')).toHaveText("รอพัฒนา");

  for (const width of [1600, 1280, 1024, 768]) {
    await page.setViewportSize({ width, height: width === 768 ? 1024 : 900 });
    for (const theme of ["light", "dark"]) {
      if (await page.locator("html").getAttribute("data-theme") !== theme) {
        await page.locator("button.theme-toggle").click();
      }
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      // Theme colors transition for up to 300 ms; capture the settled palette.
      await page.waitForTimeout(350);
      await saleOrder.hover();
      await saleOrder.focus();
      await expect(saleOrder).toBeFocused();
      const badge = saleOrder.locator("[data-menu-pending]");
      await expect(badge).toBeVisible();
      expect(await badge.evaluate((el) => el.scrollWidth <= el.clientWidth + 1)).toBe(true);
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
      await page.screenshot({ path: testInfo.outputPath(`menu-${width}-${theme}.png`), fullPage: true });
    }
  }

  await saleOrder.click();
  await expect(page.getByRole("heading", { name: "ใบสั่งขาย", exact: true })).toBeVisible();
  await search.fill("ลูกหนี้");
  const debtorItem = page.getByRole("button", { name: /^ลูกหนี้/ }).first();
  await expect(debtorItem).toBeVisible();
  await expect(debtorItem.locator("[data-menu-pending]")).toHaveCount(0);
  await debtorItem.click();
  await expect(page.getByRole("heading", { name: "ลูกหนี้", exact: true }).first()).toBeVisible();
  await search.fill("สมุดบัญชี");
  const bankItem = page.getByRole("button", { name: /^สมุดบัญชี/ }).first();
  await expect(bankItem).toBeVisible();
  await expect(bankItem.locator("[data-menu-pending]")).toHaveCount(0);
  await bankItem.click();
  await expect(page.getByRole("heading", { name: "สมุดบัญชี", exact: true }).first()).toBeVisible();
  expect(errors).toEqual([]);
  await page.getByRole("button", { name: "demo", exact: true }).click();
  await page.getByRole("menuitem", { name: "ออกจากระบบ", exact: true }).click();
  await expect(page).not.toHaveURL(/\/menu$/);
});
