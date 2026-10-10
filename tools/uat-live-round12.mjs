import { chromium } from "playwright";
import fs from "fs";
import path from "path";

const BASE_URL = "https://account.bcaicloud.com";
const OUTPUT_DIR = path.resolve("scratch/uat-round12");

if (!fs.existsSync(OUTPUT_DIR)) {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

async function run() {
  console.log("Starting Round 12 UAT on:", BASE_URL);
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
    const icMenu = page.getByRole("button", { name: "สินค้าคงคลัง" });

    // 4. Test Max Stock Report (สินค้าคงคลัง -> รายงานสินค้า)
    console.log("Step 5: Opening รายงานยอดคงเหลือสินค้าที่ถึงจุดสูงสุด...");
    await icMenu.click();
    await page.waitForTimeout(500);
    const prodReportsSub = page.locator('[role="menuitem"]').filter({ hasText: "รายงานสินค้า" }).first();
    await prodReportsSub.hover();
    await page.waitForTimeout(600);
    const maxStockItem = page.locator('[role="menuitem"]').filter({ hasText: "ถึงจุดสูงสุด" }).first();
    await maxStockItem.waitFor({ state: "visible", timeout: 5000 });
    await maxStockItem.click();
    await page.waitForTimeout(3000);
    const maxStockScreenshot = path.join(OUTPUT_DIR, "max-stock-report.png");
    await page.screenshot({ path: maxStockScreenshot, fullPage: true });
    console.log("Saved screenshot:", maxStockScreenshot);

    // 5. Test Pending Receive Report (สินค้าคงคลัง -> รายงานรายวัน)
    console.log("Step 6: Opening รายงานสรุปยอดสินค้าค้างรับ...");
    await icMenu.click();
    await page.waitForTimeout(500);
    const dailyReportsSub = page.locator('[role="menuitem"]').filter({ hasText: "รายงานรายวัน" }).first();
    await dailyReportsSub.hover();
    await page.waitForTimeout(600);
    const pendingRecItem = page.locator('[role="menuitem"]').filter({ hasText: "ค้างรับ" }).first();
    await pendingRecItem.waitFor({ state: "visible", timeout: 5000 });
    await pendingRecItem.click();
    await page.waitForTimeout(3000);
    const pendingRecScreenshot = path.join(OUTPUT_DIR, "pending-receive-report.png");
    await page.screenshot({ path: pendingRecScreenshot, fullPage: true });
    console.log("Saved screenshot:", pendingRecScreenshot);

    // 6. Test Pending Delivery Report (สินค้าคงคลัง -> รายงานรายวัน)
    console.log("Step 7: Opening รายงานสรุปยอดสินค้าค้างส่ง...");
    await icMenu.click();
    await page.waitForTimeout(500);
    await dailyReportsSub.hover();
    await page.waitForTimeout(600);
    const pendingDelItem = page.locator('[role="menuitem"]').filter({ hasText: "ค้างส่ง" }).first();
    await pendingDelItem.waitFor({ state: "visible", timeout: 5000 });
    await pendingDelItem.click();
    await page.waitForTimeout(3000);
    const pendingDelScreenshot = path.join(OUTPUT_DIR, "pending-delivery-report.png");
    await page.screenshot({ path: pendingDelScreenshot, fullPage: true });
    console.log("Saved screenshot:", pendingDelScreenshot);

    // 7. Test Serial Movement Report (สินค้าคงคลัง -> รายงานรายวัน)
    console.log("Step 8: Opening รายงานเคลื่อนไหว Serial Number...");
    await icMenu.click();
    await page.waitForTimeout(500);
    await dailyReportsSub.hover();
    await page.waitForTimeout(600);
    const serialItem = page.locator('[role="menuitem"]').filter({ hasText: "Serial Number" }).first();
    await serialItem.waitFor({ state: "visible", timeout: 5000 });
    await serialItem.click();
    await page.waitForTimeout(3000);
    const serialScreenshot = path.join(OUTPUT_DIR, "serial-movement-report.png");
    await page.screenshot({ path: serialScreenshot, fullPage: true });
    console.log("Saved screenshot:", serialScreenshot);

    // 8. Test No Movement Stock Report (สินค้าคงคลัง -> รายงานเชิงวิเคราะห์สินค้า)
    console.log("Step 9: Opening รายงานสินค้าไม่มีการเคลื่อนไหว...");
    await icMenu.click();
    await page.waitForTimeout(500);
    const analysisReportsSub = page.locator('[role="menuitem"]').filter({ hasText: "รายงานเชิงวิเคราะห์สินค้า" }).first();
    await analysisReportsSub.hover();
    await page.waitForTimeout(600);
    const noMovementItem = page.locator('[role="menuitem"]').filter({ hasText: "ไม่มีการเคลื่อนไหว" }).first();
    await noMovementItem.waitFor({ state: "visible", timeout: 5000 });
    await noMovementItem.click();
    await page.waitForTimeout(3000);
    const noMovementScreenshot = path.join(OUTPUT_DIR, "no-movement-stock-report.png");
    await page.screenshot({ path: noMovementScreenshot, fullPage: true });
    console.log("Saved screenshot:", noMovementScreenshot);

    // 9. Test Stock Count Variance Report (สินค้าคงคลัง -> รายงานเชิงวิเคราะห์สินค้า)
    console.log("Step 10: Opening รายงานผลต่างจากการตรวจนับสินค้า...");
    await icMenu.click();
    await page.waitForTimeout(500);
    await analysisReportsSub.hover();
    await page.waitForTimeout(600);
    const varianceItem = page.locator('[role="menuitem"]').filter({ hasText: "ผลต่างจากการตรวจนับ" }).first();
    await varianceItem.waitFor({ state: "visible", timeout: 5000 });
    await varianceItem.click();
    await page.waitForTimeout(3000);
    const varianceScreenshot = path.join(OUTPUT_DIR, "stock-count-variance-report.png");
    await page.screenshot({ path: varianceScreenshot, fullPage: true });
    console.log("Saved screenshot:", varianceScreenshot);

    console.log("Console errors detected:", consoleErrors.length);
    if (consoleErrors.length > 0) {
      console.log("Errors:", consoleErrors.slice(0, 5));
    }
    console.log("✅ Round 12 UAT automation completed successfully!");
  } catch (err) {
    console.error("Round 12 UAT failed:", err);
  } finally {
    await browser.close();
  }
}

run();
