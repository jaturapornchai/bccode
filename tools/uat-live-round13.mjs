import { chromium } from "playwright";
import { mkdirSync } from "node:fs";
import { resolve } from "node:path";

const targetUrl = process.env.UAT_TARGET_URL || "https://account.bcaicloud.com";
const outDir = resolve(process.cwd(), "scratch/uat-round13");
mkdirSync(outDir, { recursive: true });

async function runUAT() {
  console.log(`Starting Round 13 UAT on: ${targetUrl}`);
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const page = await context.newPage();

  let errorCount = 0;
  page.on("pageerror", (err) => {
    console.error("PAGE ERROR:", err.message);
    errorCount++;
  });
  page.on("console", (msg) => {
    if (msg.type() === "error") {
      console.error("CONSOLE ERROR:", msg.text());
      errorCount++;
    }
  });

  console.log("Step 1: Navigating to login...");
  await page.goto(`${targetUrl}/`, { waitUntil: "networkidle" });

  console.log("Step 2: Clicking demo login...");
  const demoBtn = page.getByRole("button", { name: /ทดลองใช้ระบบ/i });
  await demoBtn.click();
  await page.waitForURL("**/holding", { timeout: 15000 });

  console.log("Step 3: Selecting holding RUNGRUENG...");
  const holdingCard = page.locator('button:has-text("RUNGRUENG")').first();
  await holdingCard.click();

  console.log("Step 3.1: Selecting company 01...");
  const compBtn = page.locator('button:has-text("บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด")').first();
  await compBtn.waitFor({ state: "visible", timeout: 15000 });
  await compBtn.click();

  console.log("Step 3.2: Selecting branch 00000...");
  const branchBtn = page.locator('button:has-text("สำนักงานใหญ่")').first();
  await branchBtn.waitFor({ state: "visible", timeout: 15000 });
  await branchBtn.click();

  await page.waitForURL("**/menu", { timeout: 15000 });
  console.log("Step 4: Checking menu URL:", page.url());

  // Function to click an item inside "จัดซื้อ" menu dropdown with optional submenu
  async function openPurchaseMenu(menuLabel, screenshotName, submenuLabel) {
    console.log(`Opening menu: ${menuLabel}${submenuLabel ? ` (in ${submenuLabel})` : ""}...`);
    // Click top menu tab "จัดซื้อ"
    const poTab = page.locator('button:has-text("จัดซื้อ")').first();
    await poTab.click();
    await page.waitForTimeout(600);

    if (submenuLabel) {
      const subTrigger = page.locator(`[role="menuitem"]:has-text("${submenuLabel}")`).first();
      await subTrigger.waitFor({ state: "visible", timeout: 5000 });
      await subTrigger.hover();
      await page.waitForTimeout(600);
    }

    // Look for the menu item
    const item = page.locator(`[role="menuitem"]:has-text("${menuLabel}"), button:has-text("${menuLabel}")`).first();
    await item.waitFor({ state: "visible", timeout: 10000 });
    await item.click();
    await page.waitForTimeout(2000);

    const path = resolve(outDir, `${screenshotName}.png`);
    await page.screenshot({ path, fullPage: false });
    console.log(`Saved screenshot: ${path}`);
  }

  // 1. บันทึกลดหนี้ (/transaction/purchasereducedebt) in "บันทึกซื้อสินค้าเชื่อ"
  await openPurchaseMenu("บันทึกลดหนี้", "purchase-reduce-debt", "บันทึกซื้อสินค้าเชื่อ");

  // 2. บันทึกใบตารางราคารวม (/transaction/rfqpricetable) in "ระบบสืบราคาซื้อสินค้า"
  await openPurchaseMenu("บันทึกใบตารางราคารวม", "rfq-price-table", "ระบบสืบราคาซื้อสินค้า");

  await browser.close();
  console.log(`Console errors detected: ${errorCount}`);
  if (errorCount > 0) {
    console.error("❌ Round 13 UAT failed due to console/page errors!");
    process.exit(1);
  }
  console.log("✅ Round 13 UAT automation completed successfully!");
}

runUAT().catch((err) => {
  console.error("FATAL ERROR in UAT:", err);
  process.exit(1);
});
