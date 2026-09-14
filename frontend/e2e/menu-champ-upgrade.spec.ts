import { expect, test } from "@playwright/test";

test("Champ workflows are searchable, open pending screens, and preserve connected screens", async ({ page }, testInfo) => {
  test.setTimeout(120_000);
  await page.goto("/");
  await page.getByRole("button", { name: "ทดลองใช้ระบบ (Demo)", exact: true }).click();
  await page.getByRole("button", { name: /^กลุ่มกิจการรุ่งเรือง/ }).click();
  await page.getByRole("button", { name: /^บริษัท รุ่งเรืองวัสดุ/ }).click();
  await page.getByRole("button", { name: /สำนักงานใหญ่/ }).first().click();
  await expect(page).toHaveURL(/\/menu$/);
  await page.getByRole("button", { name: "เมนูซ้าย", exact: true }).click();
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  page.on("console", (message) => { if (message.type() === "error") errors.push(message.text()); });
  await page.setViewportSize({ width: 1600, height: 900 });
  const cashSection = page.getByRole("button", { name: /^เงินสดและธนาคาร/ }).first();
  if (!(await page.locator("#menu-tree-cash-bank").isVisible())) await cashSection.click();
  const chequeGroup = page.locator("#menu-tree-cash-bank").getByRole("button", { name: /^เช็ครับ/ }).first();
  if (await chequeGroup.getAttribute("aria-expanded") !== "true") await chequeGroup.click();
  await expect(page.locator("#menu-tree-cash-bank").getByRole("button", { name: /^นำเช็คเข้าใหม่/ })).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath("champ-sidebar-workflow.png"), animations: "disabled", fullPage: true });
  await page.getByRole("button", { name: "เมนูบน", exact: true }).click();
  const search = page.getByPlaceholder("ค้นหาเมนู เอกสาร หรือหน้าจอ");

  for (const [query, title, route] of [
    ["บันทึกใบเสนอซื้อสินค้า", "ใบขอซื้อ", "/transaction/purchaserequisition"],
    ["Weight cost", "บันทึกต้นทุนแฝง", "/transaction/landedcost"],
    ["อนุมัติใบสั่งขาย", "อนุมัติใบสั่งขาย/สั่งจองสินค้า", "/sales/order-approval"],
    ["ใบเสร็จชั่วคราว", "ใบเสร็จชั่วคราว", "/transaction/temporaryreceipt"],
    ["นำฝากเช็ครับ", "นำฝากเช็ครับ", "/banking/cheques/deposit"],
    ["ตรวจสอบรวมสินค้าชุด", "ตรวจสอบรวมสินค้าชุด", "/inventory/set-assembly"],
    ["โอนข้อมูลเข้าสู่ GL", "โอนค่าเสื่อมราคาเข้าบัญชีแยกประเภท", "/asset/post-gl"],
    ["Serial Number", "ทะเบียนเลขเครื่อง", "/productserialregistry"],
  ]) {
    await search.fill(query);
    const entry = page.getByRole("button").filter({ has: page.locator(`[data-menu-pending="${route}"]`) });
    await expect(entry).toBeVisible();
    await expect(entry.locator("[data-menu-pending]")).toHaveText("รอพัฒนา");
    await entry.click();
    await expect(page.getByRole("heading", { name: title, exact: true })).toBeVisible();
    await expect(page.getByText("เมนูในแผนพัฒนา", { exact: true }).last()).toBeVisible();
  }

  await search.fill("เช็ครับ");
  const deposit = page.getByRole("button").filter({ has: page.locator('[data-menu-pending="/banking/cheques/deposit"]') });
  for (const width of [1600, 1280, 1024, 768]) {
    await page.setViewportSize({ width, height: width === 768 ? 1024 : 900 });
    for (const theme of ["light", "dark"]) {
      if (await page.locator("html").getAttribute("data-theme") !== theme) await page.locator("button.theme-toggle").click();
      await expect(page.locator("html")).toHaveAttribute("data-theme", theme);
      await deposit.hover();
      await deposit.focus();
      await expect(deposit).toBeFocused();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
      await page.screenshot({ path: testInfo.outputPath(`champ-cheques-${width}-${theme}.png`), animations: "disabled", fullPage: true });
    }
  }

  // Existing connected screens must stay connected; this test only opens them, without writing data.
  for (const [query, title] of [["โปรโมชั่น", "โปรโมชั่น"], ["Account Mapping", "รูปแบบการเชื่อมโยงบัญชีอัตโนมัติ"]]) {
    await search.fill(query);
    const entry = page.getByRole("button", { name: new RegExp(`^${title}`) }).first();
    await expect(entry).toBeVisible();
    await expect(entry.locator("[data-menu-pending]")).toHaveCount(0);
    await entry.click();
    await expect(page.getByRole("heading", { name: title, exact: true }).first()).toBeVisible();
  }
  expect(errors).toEqual([]);
});
