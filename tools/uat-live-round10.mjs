import { chromium } from "playwright";
import fs from "fs";
import path from "path";

const BASE_URL = "https://account.bcaicloud.com";
const OUTPUT_DIR = path.resolve("scratch/uat-round10");

if (!fs.existsSync(OUTPUT_DIR)) {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

async function run() {
  console.log("Starting Round 10 UAT on:", BASE_URL);
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

    // Wait until menu
    console.log("Step 4: Checking menu URL:", page.url());

    // 4. Test Cheque Received Report
    console.log("Step 5: Opening รายงานเช็ครับ via เช็ค/บัตร menu...");
    const chequeMenu = page.getByRole("button", { name: "เช็ค/บัตร" });
    await chequeMenu.click();
    await page.waitForTimeout(800);
    const chqRecItem = page.locator('[role="menuitem"]').filter({ hasText: "รายงานเช็ครับ" }).first();
    await chqRecItem.click();
    await page.waitForTimeout(3000);
    const chqRecScreenshot = path.join(OUTPUT_DIR, "cheque-received-report.png");
    await page.screenshot({ path: chqRecScreenshot, fullPage: true });
    console.log("Saved screenshot:", chqRecScreenshot);

    // 5. Test Cheque Issued Report
    console.log("Step 6: Opening รายงานเช็คจ่าย via เช็ค/บัตร menu...");
    await chequeMenu.click();
    await page.waitForTimeout(800);
    const chqIssItem = page.locator('[role="menuitem"]').filter({ hasText: "รายงานเช็คจ่าย" }).first();
    await chqIssItem.click();
    await page.waitForTimeout(3000);
    const chqIssScreenshot = path.join(OUTPUT_DIR, "cheque-issued-report.png");
    await page.screenshot({ path: chqIssScreenshot, fullPage: true });
    console.log("Saved screenshot:", chqIssScreenshot);

    // 6. Test Credit Card Report
    console.log("Step 7: Opening รายงานบัตรเครดิต via เช็ค/บัตร menu...");
    await chequeMenu.click();
    await page.waitForTimeout(800);
    const cardItem = page.locator('[role="menuitem"]').filter({ hasText: "รายงานบัตรเครดิต" }).first();
    await cardItem.click();
    await page.waitForTimeout(3000);
    const cardScreenshot = path.join(OUTPUT_DIR, "credit-card-report.png");
    await page.screenshot({ path: cardScreenshot, fullPage: true });
    console.log("Saved screenshot:", cardScreenshot);

    console.log("Console errors detected:", consoleErrors.length);
    if (consoleErrors.length > 0) {
      console.log("Errors:", consoleErrors.slice(0, 5));
    }
    console.log("✅ Round 10 UAT automation completed successfully!");
  } catch (err) {
    console.error("Round 10 UAT failed:", err);
  } finally {
    await browser.close();
  }
}

run();
