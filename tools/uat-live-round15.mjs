import { chromium } from "playwright";
import { mkdirSync, readFileSync } from "node:fs";
import { resolve } from "node:path";

const targetUrl = process.env.UAT_TARGET_URL || "https://account.bcaicloud.com";
const outDir = resolve(process.cwd(), "scratch/uat-round15");
mkdirSync(outDir, { recursive: true });

async function runUAT() {
  console.log(`Starting Round 15 UAT on: ${targetUrl}`);
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 }, acceptDownloads: true });
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

  // Click top menu tab "จัดขาย"
  console.log("Step 5: Opening จัดขาย tab...");
  const salesTab = page.locator('button:has-text("จัดขาย")').first();
  await salesTab.click();
  await page.waitForTimeout(600);

  // Click "ใบกำกับภาษีอิเล็กทรอนิกส์"
  console.log("Step 6: Clicking ใบกำกับภาษีอิเล็กทรอนิกส์...");
  const etaxItem = page.locator('[role="menuitem"]:has-text("ใบกำกับภาษีอิเล็กทรอนิกส์"), button:has-text("ใบกำกับภาษีอิเล็กทรอนิกส์")').first();
  await etaxItem.waitFor({ state: "visible", timeout: 10000 });
  await etaxItem.click();
  await page.waitForTimeout(2000);

  // 1. Screenshot e-Tax Portal
  const pathQueue = resolve(outDir, "etax-portal-live.png");
  await page.screenshot({ path: pathQueue, fullPage: false });
  console.log(`Saved screenshot: ${pathQueue}`);

  // 2. Click XML download button on first row (INV6910-0012)
  console.log("Step 7: Testing real XML download from backend endpoint...");
  const downloadPromise = page.waitForEvent("download", { timeout: 15000 });
  const xmlBtn = page.locator('button[title*="Download XML"]').first();
  await xmlBtn.click();

  const download = await downloadPromise;
  const suggestedName = download.suggestedFilename();
  console.log(`Received downloaded file: ${suggestedName}`);

  const savePath = resolve(outDir, suggestedName);
  await download.saveAs(savePath);
  console.log(`Saved downloaded XML to: ${savePath}`);

  // Inspect XML contents
  const xmlContent = readFileSync(savePath, "utf8");
  if (!xmlContent.includes("CrossIndustryInvoice") || !xmlContent.includes("ER3-2560")) {
    throw new Error(`Downloaded XML is invalid or missing ETDA standard header! Content: ${xmlContent.slice(0, 200)}`);
  }
  console.log("✅ Verified: Downloaded XML contains valid ETDA ER3-2560 CrossIndustryInvoice schema!");

  await browser.close();
  console.log(`Console errors detected: ${errorCount}`);
  if (errorCount > 0) {
    console.error("❌ Round 15 UAT failed due to console/page errors!");
    process.exit(1);
  }
  console.log("✅ Round 15 UAT automation completed successfully!");
}

runUAT().catch((err) => {
  console.error("FATAL ERROR in UAT:", err);
  process.exit(1);
});
