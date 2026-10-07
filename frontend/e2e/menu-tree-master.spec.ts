import { expect, test } from "@playwright/test";

// 2026-10-07: rewritten for the Champ-style menu (src/lib/menu-bars.ts). The left tree shows the same heads in the
// same order as the top bar (only the heads that hold screens), with Champ submenus as folders.
test("left menu tree follows the Champ menu bar order and opens a screen", async ({ page }, testInfo) => {
  test.setTimeout(60_000);
  const errors: string[] = [];

  await page.goto("/");
  await page.getByRole("button", { name: "ทดลองใช้ระบบ (Demo)", exact: true }).click();
  await page.getByRole("button", { name: /^กลุ่มกิจการรุ่งเรือง/ }).click();
  await page.getByRole("button", { name: /^บริษัท รุ่งเรืองวัสดุ/ }).click();
  await page.getByRole("button", { name: /สำนักงานใหญ่/ }).first().click();
  await expect(page).toHaveURL(/\/menu$/);
  page.on("pageerror", (err) => errors.push(err.message));
  page.on("console", (msg) => {
    if (msg.type() === "error") errors.push(msg.text());
  });

  // Top bar: overview + 13 Champ heads, each one a menu trigger, on one row at 1024px.
  await page.setViewportSize({ width: 1024, height: 800 });
  await expect(page.getByRole("button", { name: "เมนูบน", exact: true })).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByRole("navigation", { name: "เมนู" }).locator('button[aria-haspopup="menu"]')).toHaveCount(13);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);

  await page.getByRole("button", { name: "เมนูซ้าย", exact: true }).click();
  const tree = page.getByRole("tree", { name: "เมนู", exact: true });
  await expect(tree).toBeVisible();
  await expect.poll(() => page.evaluate(() => localStorage.getItem("bc_menu_layout_mode"))).toBe("left");

  // Heads that hold screens, in Champ order (ไฟล์/จัดการระบบ/หน้าต่าง/ช่วยเหลือ are actions in the top bar only).
  const heads = [/^จัดซื้อ/, /^จัดขาย/, /^เจ้าหนี้/, /^ลูกหนี้/, /^เช็ค\/บัตร/, /^ธนาคาร\/เงินสด/, /^สินค้าคงคลัง/, /^สินทรัพย์/, /^บัญชี/];
  const headTexts = await tree.locator('button[aria-controls^="menu-tree-"]').allInnerTexts();
  expect(headTexts).toHaveLength(heads.length);
  heads.forEach((pattern, index) => expect(headTexts[index]).toMatch(pattern));

  // Accounting head: Champ wording, opens the journal screen.
  await tree.getByRole("button", { name: /^บัญชี/ }).first().click();
  const gl = page.locator("#menu-tree-gl");
  await expect(gl).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath("champ-aligned-menu-tree.png"), fullPage: false });
  await gl.getByRole("button", { name: /^บันทึกข้อมูลรายวัน/ }).first().click();
  await expect(page.locator(".gl-workbench:visible").first()).toBeVisible();

  expect(errors).toEqual([]);
});
