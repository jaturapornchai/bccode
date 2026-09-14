import { expect, test } from "@playwright/test";

test("ERP 9-module navigation tree renders uniformly and opens screen", async ({ page }, testInfo) => {
  test.setTimeout(60_000);
  const errors: string[] = [];

  // 1. Sign in via Demo
  await page.goto("/");
  await page.getByRole("button", { name: "ทดลองใช้ระบบ (Demo)", exact: true }).click();
  await page.getByRole("button", { name: /^กลุ่มกิจการรุ่งเรือง/ }).click();
  await page.getByRole("button", { name: /^บริษัท รุ่งเรืองวัสดุ/ }).click();
  await page.getByRole("button", { name: /สำนักงานใหญ่/ }).first().click();
  await expect(page).toHaveURL(/\/menu$/);
  // Scope errors to the menu, as in menu-consistency; Google's login iframe
  // can emit its own report-only CSP message before navigation completes.
  page.on("pageerror", (err) => errors.push(err.message));
  page.on("console", (msg) => {
    if (msg.type() === "error") errors.push(msg.text());
  });
  await expect(page.getByRole("button", { name: "เมนูบน", exact: true })).toHaveAttribute("aria-pressed", "true");
  await page.getByRole("button", { name: "เมนูซ้าย", exact: true }).click();
  await expect(page.getByRole("tree", { name: "เมนู", exact: true })).toBeVisible();
  await expect.poll(() => page.evaluate(() => localStorage.getItem("bc_menu_layout_mode"))).toBe("left");
  await page.reload();
  await expect(page.getByRole("button", { name: "เมนูซ้าย", exact: true })).toHaveAttribute("aria-pressed", "true");
  await expect(page.getByRole("tree", { name: "เมนู", exact: true })).toBeVisible();

  // 2. Verify all 9 ERP module sections are visible in sidebar in Champ order
  const expectedModules = [
    { id: "po", name: /ซื้อ\/สั่งซื้อสินค้า/ },
    { id: "bill", name: /ใบสั่งของ\/ใบกำกับสินค้า/ },
    { id: "ap", name: /เจ้าหนี้/ },
    { id: "ar", name: /ลูกหนี้/ },
    { id: "cash-bank", name: /เงินสดและธนาคาร/ },
    { id: "ic", name: /สินค้าคงคลัง/ },
    { id: "fa", name: /สินทรัพย์และค่าเสื่อมราคา/ },
    { id: "vat", name: /ภาษีมูลค่าเพิ่ม/ },
    { id: "gl", name: /บัญชีแยกประเภท/ },
  ];

  for (const mod of expectedModules) {
    const modButton = page.getByRole("button", { name: mod.name }).first();
    await expect(modButton).toBeVisible();
  }

  // 3. Open "ระบบซื้อ/สั่งซื้อสินค้า (PO)" section in the left sidebar tree
  const poSectionButton = page.getByRole("button", { name: /ระบบซื้อ\/สั่งซื้อสินค้า/ }).first();
  await expect(poSectionButton).toBeVisible();

  const poGroupContainer = page.locator("#menu-tree-po");
  if (!(await poGroupContainer.isVisible())) {
    await poSectionButton.click();
  }
  await expect(poGroupContainer).toBeVisible();

  // 4. Verify groups inside PO (aligned with Champ)
  const poProcureGroup = poGroupContainer.getByRole("button", { name: /^งานจัดซื้อจัดหา/ });
  await expect(poProcureGroup).toBeVisible();
  const poTxGroup = poGroupContainer.getByRole("button", { name: /^บันทึกซื้อและค่าใช้จ่าย/ });
  await expect(poTxGroup).toBeVisible();
  const poPayGroup = poGroupContainer.getByRole("button", { name: /^เงินมัดจำและจ่ายล่วงหน้า/ });
  await expect(poPayGroup).toBeVisible();

  // Click to expand "งานจัดซื้อจัดหา" group
  if ((await poProcureGroup.getAttribute("aria-expanded")) !== "true") {
    await poProcureGroup.click();
  }
  await expect(poProcureGroup).toHaveAttribute("aria-expanded", "true");

  const prItemButton = poGroupContainer.getByRole("button", { name: /^ใบขอซื้อ/ }).first();
  await expect(prItemButton).toBeVisible();

  // Take screenshot of ERP Champ-aligned sidebar tree
  await page.screenshot({ path: testInfo.outputPath("champ-aligned-menu-tree.png"), fullPage: false });

  // 5. Click the PR item to open the screen
  await prItemButton.click();
  await expect(page.getByRole("heading", { name: "ใบขอซื้อ", exact: true }).first()).toBeVisible();

  // 6. Verify "ระบบบัญชีแยกประเภท (GL)" in the left sidebar tree
  const glSectionButton = page.getByRole("button", { name: /บัญชีแยกประเภท/ }).first();
  await expect(glSectionButton).toBeVisible();

  const glGroupContainer = page.locator("#menu-tree-gl");
  if (!(await glGroupContainer.isVisible())) {
    await glSectionButton.click();
  }
  await expect(glGroupContainer).toBeVisible();

  // Verify all 4 GL groups and 35 total items count
  const glMasterGroup = glGroupContainer.getByRole("button", { name: /^ข้อมูลหลักและยอดยกมา/ });
  await expect(glMasterGroup).toBeVisible();
  const glJournalsGroup = glGroupContainer.getByRole("button", { name: /^สมุดรายวันและงานประจำ/ });
  await expect(glJournalsGroup).toBeVisible();
  const glPostingGroup = glGroupContainer.getByRole("button", { name: /^ผ่านบัญชีและประมวลผลสิ้นปี/ });
  await expect(glPostingGroup).toBeVisible();
  const glReportsGroup = glGroupContainer.getByRole("button", { name: /^รายงานการเงินและงบบัญชี/ });
  await expect(glReportsGroup).toBeVisible();
  await expect(glGroupContainer.getByText("35 รายการ")).toBeVisible();

  // Expand "สมุดรายวันและงานประจำ" group
  if ((await glJournalsGroup.getAttribute("aria-expanded")) !== "true") {
    await glJournalsGroup.click();
  }
  await expect(glJournalsGroup).toHaveAttribute("aria-expanded", "true");

  const glJournalsContainer = glGroupContainer.locator('[id="menu-tree-gl:gl-journals"]');
  await expect(glJournalsContainer).toBeVisible();

  // Verify all 9 journal workflow items
  const expectedJournals = [
    "สมุดรายวันขาย",
    "สมุดรายวันซื้อ",
    "สมุดรายวันรับเงิน",
    "สมุดรายวันจ่ายเงิน",
    "สมุดรายวันทั่วไป",
    "กระดาษทำการ",
    "ล็อกงวดบัญชี",
    "ปิดงบบัญชีสิ้นงวด",
    "ตรวจสอบประจำวัน",
  ];
  for (const name of expectedJournals) {
    await expect(glJournalsContainer.getByRole("button", { name: new RegExp(`^${name}`) }).first()).toBeVisible();
  }

  // Scroll GL section into view so the whole tree is captured nicely
  await glJournalsContainer.scrollIntoViewIfNeeded();
  await page.screenshot({ path: testInfo.outputPath("gl-aligned-menu-tree.png"), fullPage: false });

  // 7. Click "สมุดรายวันขาย" to open the GL screen
  const salesJournalButton = glJournalsContainer.getByRole("button", { name: /^สมุดรายวันขาย/ }).first();
  await salesJournalButton.click();
  await expect(page.locator('.gl-workbench[data-gl-route="/gl/journal/uv"]:visible')).toBeVisible();
  await expect(page.locator(".gl-workbench:visible").getByRole("heading", { name: "สมุดรายวันขาย", exact: true })).toBeVisible();
  await page.screenshot({ path: testInfo.outputPath("gl-sales-journal-screen.png"), fullPage: false });

  expect(errors).toEqual([]);
});
