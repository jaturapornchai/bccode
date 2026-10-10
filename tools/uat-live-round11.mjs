import { chromium } from "playwright";
import fs from "fs";
import path from "path";

const BASE_URL = "https://account.bcaicloud.com";
const OUTPUT_DIR = path.resolve("scratch/uat-round11");

if (!fs.existsSync(OUTPUT_DIR)) {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

async function run() {
  console.log("Starting Round 11 UAT on:", BASE_URL);
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
  });
  const page = await context.newPage();

  const consoleErrors = [];
  page.on("console", (msg) => {
    if (msg.type() === "error") {
      consoleErrors.push(msg.text());
    }
  });

  try {
    // 1. Visit Login
    console.log("Step 1: Navigating to login...");
    await page.goto(BASE_URL, { waitUntil: "networkidle" });
    await page.waitForTimeout(2000);

    // 2. Click Demo Login
    console.log("Step 2: Clicking demo login...");
    const demoButton = page.locator('button:has-text("ทดลองใช้ระบบ (Demo)"), button:has-text("ทดลองใช้ระบบ")');
    if (await demoButton.isVisible()) {
      await demoButton.click();
    } else {
      await page.click('text="ทดลองใช้ระบบ"');
    }
    await page.waitForURL("**/holding**", { timeout: 15000 }).catch(() => {});
    await page.waitForTimeout(2500);

    // 3. Select Holding -> Company -> Branch -> Menu
    console.log("Step 3: Selecting holding RUNGRUENG...");
    const rungruengCard = page.locator("button").filter({ hasText: "RUNGRUENG" }).first();
    await rungruengCard.waitFor({ state: "visible", timeout: 10000 });
    await rungruengCard.click();
    await page.waitForTimeout(2000);

    console.log("Step 3.1: Selecting company 01...");
    const companyCard = page.locator("button").filter({ hasText: "01" }).first();
    await companyCard.waitFor({ state: "visible", timeout: 10000 });
    await companyCard.click();
    await page.waitForTimeout(2000);

    console.log("Step 3.2: Selecting branch 00000...");
    const branchCard = page.locator("button").filter({ hasText: "00000" }).first();
    await branchCard.waitFor({ state: "visible", timeout: 10000 });
    await branchCard.click();
    await page.waitForTimeout(3000);

    console.log("Step 4: Checking menu URL:", page.url());

    // 4. Test Bank Statement Report
    console.log("Step 5: Opening รายงาน Bank Statement via ธนาคาร/เงินสด -> รายงานต่างๆ...");
    const bankMenu = page.getByRole("button", { name: "ธนาคาร/เงินสด" });
    await bankMenu.click();
    await page.waitForTimeout(500);

    const miscReportsSub1 = page.locator('[role="menuitem"]').filter({ hasText: "รายงานต่างๆ" }).first();
    await miscReportsSub1.hover();
    await page.waitForTimeout(600);

    const bankStatementItem = page.locator('[role="menuitem"]').filter({ hasText: "รายงาน Bank Statement" }).first();
    await bankStatementItem.waitFor({ state: "visible", timeout: 5000 });
    await bankStatementItem.click();
    await page.waitForTimeout(3000);
    const bankStatementScreenshot = path.join(OUTPUT_DIR, "bank-statement-report.png");
    await page.screenshot({ path: bankStatementScreenshot, fullPage: true });
    console.log("Saved screenshot:", bankStatementScreenshot);

    // 5. Test Cash Movement Report
    console.log("Step 6: Opening รายงานเคลื่อนไหวเงินสด via ธนาคาร/เงินสด -> รายงานต่างๆ...");
    await bankMenu.click();
    await page.waitForTimeout(500);

    const miscReportsSub2 = page.locator('[role="menuitem"]').filter({ hasText: "รายงานต่างๆ" }).first();
    await miscReportsSub2.hover();
    await page.waitForTimeout(600);

    const cashMovementItem = page.locator('[role="menuitem"]').filter({ hasText: "รายงานเคลื่อนไหวเงินสด" }).first();
    await cashMovementItem.waitFor({ state: "visible", timeout: 5000 });
    await cashMovementItem.click();
    await page.waitForTimeout(3000);
    const cashMovementScreenshot = path.join(OUTPUT_DIR, "cash-movement-report.png");
    await page.screenshot({ path: cashMovementScreenshot, fullPage: true });
    console.log("Saved screenshot:", cashMovementScreenshot);

    // 6. Test Petty Cash Movement Report
    console.log("Step 7: Opening รายงานเคลื่อนไหวเงินสดย่อย via ธนาคาร/เงินสด -> รายงานต่างๆ...");
    await bankMenu.click();
    await page.waitForTimeout(500);

    const miscReportsSub3 = page.locator('[role="menuitem"]').filter({ hasText: "รายงานต่างๆ" }).first();
    await miscReportsSub3.hover();
    await page.waitForTimeout(600);

    const pettyCashItem = page.locator('[role="menuitem"]').filter({ hasText: "รายงานเคลื่อนไหวเงินสดย่อย" }).first();
    await pettyCashItem.waitFor({ state: "visible", timeout: 5000 });
    await pettyCashItem.click();
    await page.waitForTimeout(3000);
    const pettyCashScreenshot = path.join(OUTPUT_DIR, "petty-cash-movement-report.png");
    await page.screenshot({ path: pettyCashScreenshot, fullPage: true });
    console.log("Saved screenshot:", pettyCashScreenshot);

    // 7. Test Monthly Payment Book Report
    console.log("Step 8: Opening รายงานสมุดจ่ายเงินประจำเดือน via ธนาคาร/เงินสด -> รายงานต่างๆ...");
    await bankMenu.click();
    await page.waitForTimeout(500);

    const miscReportsSub4 = page.locator('[role="menuitem"]').filter({ hasText: "รายงานต่างๆ" }).first();
    await miscReportsSub4.hover();
    await page.waitForTimeout(600);

    const monthlyBookItem = page.locator('[role="menuitem"]').filter({ hasText: "รายงานสมุดจ่ายเงินประจำเดือน" }).first();
    await monthlyBookItem.waitFor({ state: "visible", timeout: 5000 });
    await monthlyBookItem.click();
    await page.waitForTimeout(3000);
    const monthlyBookScreenshot = path.join(OUTPUT_DIR, "monthly-payment-book-report.png");
    await page.screenshot({ path: monthlyBookScreenshot, fullPage: true });
    console.log("Saved screenshot:", monthlyBookScreenshot);

    console.log("Console errors detected:", consoleErrors.length);
    if (consoleErrors.length > 0) {
      console.log("Errors:", consoleErrors.slice(0, 5));
    }
    console.log("✅ Round 11 UAT automation completed successfully!");
  } catch (err) {
    console.error("Round 11 UAT failed:", err);
  } finally {
    await browser.close();
  }
}

run();
