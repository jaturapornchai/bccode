import { execSync } from "node:child_process";
import { expect, test } from "@playwright/test";

function queryMongoBank(code: string, includeDeleted = false): Record<string, unknown> | null {
  try {
    const filter = includeDeleted ? `{ code: '${code}' }` : `{ code: '${code}', deletedat: { '$exists': false } }`;
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


function cleanMongoBank(code: string) {
  try {
    execSync(
      `docker exec mongodb mongosh appdb --quiet --eval "db.bankmaster.deleteMany({ code: '${code}' })"`,
      { encoding: "utf8" },
    );
  } catch {
    // ignore
  }
}

test.describe("Bank Master CRUD & MongoDB Verification", () => {
  const testCode = "BTEST";

  test.beforeEach(() => {
    cleanMongoBank(testCode);
  });

  test.afterEach(() => {
    cleanMongoBank(testCode);
  });

  test("creates, reads, updates, and deletes bank master with live MongoDB verification", async ({ page }) => {
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

    // 2. Open Bank screen from menu
    await page.getByRole("button", { name: "เมนูบน", exact: true }).click();
    const search = page.getByPlaceholder("ค้นหาเมนู เอกสาร หรือหน้าจอ");
    await search.fill("ธนาคาร");
    const bankItem = page.getByRole("button", { name: /^ธนาคาร/ });
    await expect(bankItem).toBeVisible();
    await bankItem.click();

    // Verify bank screen heading
    await expect(page.getByRole("heading", { name: "ธนาคาร", exact: true }).first()).toBeVisible();

    // 3. CREATE: Click "เพิ่ม" button
    const addButton = page.getByRole("button", { name: "เพิ่ม", exact: true });
    await expect(addButton).toBeVisible();
    await addButton.click();

    const form = page.locator("form").last();
    await expect(form).toBeVisible();

    // Fill form
    await form.locator("label").filter({ hasText: "รหัสธนาคาร" }).locator("input").fill(testCode);

    // Fill Thai and English names in the names editor
    const nameSection = form.locator("section").filter({ hasText: "ชื่อธนาคาร" });
    const nameInputs = nameSection.locator("input");
    await nameInputs.first().fill("ธนาคารทดสอบระบบ");
    if ((await nameInputs.count()) > 1) {
      await nameInputs.nth(1).fill("Test System Bank");
    }

    // Click Save ("บันทึก")
    const saveButton = form.getByRole("button", { name: "บันทึก", exact: true });
    await expect(saveButton).toBeVisible();
    await saveButton.click();
    await expect(form).toHaveCount(0, { timeout: 10_000 });

    // Wait for detail to show the created record
    await expect(page.getByText(testCode).first()).toBeVisible({ timeout: 10_000 });

    // STEP 3 VERIFICATION: Confirm doc created directly in MongoDB
    const createdDoc = queryMongoBank(testCode);
    expect(createdDoc).not.toBeNull();
    expect(createdDoc?.code).toBe(testCode);
    const names = createdDoc?.names as Array<{ code: string; name: string }>;
    expect(names.some((n) => n.name.includes("ธนาคารทดสอบระบบ"))).toBe(true);
    const originalGuid = createdDoc?.guidfixed;
    expect(originalGuid).toBeTruthy();

    // 4. READ: Verify record is listed and detail is shown
    await expect(page.getByText("ธนาคารทดสอบระบบ").first()).toBeVisible();

    // 5. UPDATE: Click "แก้ไข"
    const editButton = page.getByRole("button", { name: "แก้ไข" }).last();
    await expect(editButton).toBeVisible();
    await editButton.click();

    const editForm = page.locator("form").last();
    await expect(editForm).toBeVisible();

    // Modify Thai name
    const editNameSection = editForm.locator("section").filter({ hasText: "ชื่อธนาคาร" });
    const editNameInput = editNameSection.locator("input").first();
    await expect(editNameInput).toBeVisible();
    await editNameInput.fill("ธนาคารทดสอบระบบ ปรับปรุง");

    // Click Save ("บันทึก")
    await editForm.getByRole("button", { name: "บันทึก", exact: true }).click();
    await expect(editForm).toHaveCount(0, { timeout: 10_000 });
    await expect(page.getByText("ธนาคารทดสอบระบบ ปรับปรุง").first()).toBeVisible({ timeout: 10_000 });

    // STEP 5 VERIFICATION: Confirm doc updated in MongoDB in the same document
    const updatedDoc = queryMongoBank(testCode);
    expect(updatedDoc).not.toBeNull();
    expect(updatedDoc?.guidfixed).toBe(originalGuid);
    const updatedNames = updatedDoc?.names as Array<{ code: string; name: string }>;
    expect(updatedNames.some((n) => n.name === "ธนาคารทดสอบระบบ ปรับปรุง")).toBe(true);

    // 6. DELETE: Click "ลบ"
    const deleteButton = page.getByRole("button", { name: "ลบ" }).last();
    await expect(deleteButton).toBeVisible();
    await deleteButton.click();

    // Confirm dialog
    const confirmDialog = page.getByRole("dialog");
    await expect(confirmDialog).toBeVisible();
    const confirmDeleteButton = confirmDialog.getByRole("button", { name: "ลบ", exact: true });
    await expect(confirmDeleteButton).toBeVisible();
    await confirmDeleteButton.click();
    await expect(confirmDialog).toHaveCount(0, { timeout: 10_000 });

    // Wait for row to disappear from list
    await expect(page.getByText("ธนาคารทดสอบระบบ ปรับปรุง")).toHaveCount(0, { timeout: 10_000 });

    // STEP 6 VERIFICATION: Confirm doc is soft-deleted in MongoDB (active query returns null, deletedat is stamped)
    const activeDoc = queryMongoBank(testCode);
    expect(activeDoc).toBeNull();
    const rawDoc = queryMongoBank(testCode, true);
    expect(rawDoc).not.toBeNull();
    expect(rawDoc?.deletedat).toBeTruthy();
    expect(rawDoc?.deletedby).toBe("demo");

    // Ensure no console errors occurred during the entire test
    expect(errors).toEqual([]);
  });
});
