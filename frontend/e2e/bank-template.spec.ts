import { execSync } from "node:child_process";
import { expect, test } from "@playwright/test";

function queryMongoBank(code: string): Record<string, unknown> | null {
  try {
    const filter = `{ code: '${code}', deletedat: { '$exists': false } }`;
    const output = execSync(
      `docker exec mongodb mongosh appdb --quiet --eval "JSON.stringify(db.bankmaster.findOne(${filter}))"`,
      { encoding: "utf8" },
    ).trim();
    if (!output || output === "null") return null;
    return JSON.parse(output) as Record<string, unknown>;
  } catch {
    return null;
  }
}

function cleanMongoBanks(codes: string[]) {
  try {
    const list = codes.map((c) => `'${c}'`).join(",");
    execSync(
      `docker exec mongodb mongosh appdb --quiet --eval "db.bankmaster.deleteMany({ code: { '$in': [${list}] } })"`,
      { encoding: "utf8" },
    );
  } catch {
    // ignore
  }
}

test.describe("Thai Bank Template Feature & Live MongoDB Verification", () => {
  const testCodes = ["KBANK", "SCB"];

  test.beforeEach(() => {
    cleanMongoBanks(testCodes);
  });

  test.afterEach(() => {
    cleanMongoBanks(testCodes);
  });

  test("adds Thai banks from template dialog and verifies in MongoDB", async ({ page }) => {
    test.setTimeout(90_000);
    const errors: string[] = [];
    page.on("pageerror", (err) => errors.push(err.message));
    page.on("console", (msg) => {
      if (
        msg.type() === "error" &&
        !msg.text().includes("GSI_LOGGER") &&
        !msg.text().includes("Failed to load resource")
      ) {
        errors.push(msg.text());
      }
    });

    // 1. Sign in via Demo
    await page.goto("/");
    await page.getByRole("button", { name: "ทดลองใช้ระบบ (Demo)", exact: true }).click();
    await page.getByRole("button", { name: /^กลุ่มกิจการรุ่งเรือง/ }).click();
    await page.getByRole("button", { name: /^บริษัท รุ่งเรืองวัสดุ/ }).click();
    await page.getByRole("button", { name: /สำนักงานใหญ่/ }).first().click();
    await expect(page).toHaveURL(/\/menu$/);

    // 2. Open Bank screen from top menu
    await page.getByRole("button", { name: "เมนูบน", exact: true }).click();
    const search = page.getByPlaceholder("ค้นหาเมนู เอกสาร หรือหน้าจอ");
    await search.fill("ธนาคาร");
    const bankItem = page.getByRole("button", { name: /^ธนาคาร/ });
    await expect(bankItem).toBeVisible();
    await bankItem.click();

    // Verify bank screen heading
    await expect(page.getByRole("heading", { name: "ธนาคาร", exact: true }).first()).toBeVisible();

    // 3. Verify "เพิ่มธนาคารไทย" button is visible
    const templateButton = page.getByRole("button", { name: "เพิ่มธนาคารไทย" }).first();
    await expect(templateButton).toBeVisible();
    await templateButton.click();

    // 4. Verify Template Dialog opens
    const dialog = page.getByRole("dialog", { name: "เพิ่มธนาคารไทยจากแม่แบบ" });
    await expect(dialog).toBeVisible();

    // Check search functionality inside dialog
    const searchInput = dialog.getByPlaceholder("ค้นหารหัสธนาคาร หรือชื่อธนาคาร");
    await expect(searchInput).toBeVisible();
    await searchInput.fill("กสิกรไทย");

    // Only KBANK should be listed in filtered view
    await expect(dialog.getByText("KBANK")).toBeVisible();
    await expect(dialog.getByText("Kasikornbank")).toBeVisible();

    // Clear selection, then check only KBANK
    const clearButton = dialog.getByRole("button", { name: "ล้างการเลือก" });
    await clearButton.click();

    const kbankCheckbox = dialog.locator("label").filter({ hasText: "KBANK" }).locator("input[type='checkbox']");
    await kbankCheckbox.check();
    expect(await kbankCheckbox.isChecked()).toBe(true);

    // 5. Click Save: "เพิ่มธนาคารที่เลือก (1)"
    const addSelectedButton = dialog.getByRole("button", { name: /เพิ่มธนาคารที่เลือก/ });
    await expect(addSelectedButton).toBeVisible();
    await addSelectedButton.click();

    // Dialog should close
    await expect(dialog).toHaveCount(0, { timeout: 10_000 });

    // 6. Verify in UI list and detail
    await expect(page.getByText("KBANK").first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText("ธนาคารกสิกรไทย").first()).toBeVisible();

    // 7. STEP 7 VERIFICATION: Check directly in MongoDB
    const mongoKbank = queryMongoBank("KBANK");
    expect(mongoKbank).not.toBeNull();
    expect(mongoKbank?.code).toBe("KBANK");
    expect(mongoKbank?.logo).toBe("/banks/KBANK.png");
    const names = mongoKbank?.names as Array<{ code: string; name: string }>;
    expect(names.some((n) => n.name === "ธนาคารกสิกรไทย")).toBe(true);
    expect(names.some((n) => n.name === "Kasikornbank")).toBe(true);

    // 8. Re-open template dialog to verify KBANK is now marked as "มีในระบบแล้ว"
    await templateButton.click();
    await expect(dialog).toBeVisible();
    await searchInput.fill("KBANK");
    await expect(dialog.getByText("มีในระบบแล้ว").first()).toBeVisible();
    const disabledKbankCheckbox = dialog.locator("label").filter({ hasText: "KBANK" }).locator("input[type='checkbox']");
    expect(await disabledKbankCheckbox.isDisabled()).toBe(true);
    await dialog.getByRole("button", { name: "ยกเลิก" }).click();
    await expect(dialog).toHaveCount(0);

    // 9. Test single Add form auto-fill from template
    const addSingleButton = page.getByRole("button", { name: "เพิ่ม", exact: true });
    await addSingleButton.click();
    const form = page.locator("form").last();
    await expect(form).toBeVisible();

    // Pick SCB from template dropdown
    const templateSelect = form.locator("select").filter({ hasText: "-- เลือกธนาคารไทย --" });
    await expect(templateSelect).toBeVisible();
    await templateSelect.selectOption("SCB");

    // Check code input auto-filled with SCB
    const codeInput = form.locator("label").filter({ hasText: "รหัสธนาคาร" }).locator("input");
    await expect(codeInput).toHaveValue("SCB");

    // Save SCB
    await form.getByRole("button", { name: "บันทึก", exact: true }).click();
    await expect(form).toHaveCount(0, { timeout: 10_000 });

    // Verify SCB in MongoDB
    const mongoScb = queryMongoBank("SCB");
    expect(mongoScb).not.toBeNull();
    expect(mongoScb?.code).toBe("SCB");
    expect(mongoScb?.logo).toBe("/banks/SCB.png");

    expect(errors).toEqual([]);
  });
});
