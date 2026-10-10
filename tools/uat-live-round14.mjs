import { chromium } from "playwright";
import { mkdirSync } from "node:fs";
import { resolve } from "node:path";

const targetUrl = process.env.UAT_TARGET_URL || "https://account.bcaicloud.com";
const outDir = resolve(process.cwd(), "scratch/uat-round14");
mkdirSync(outDir, { recursive: true });

async function runUAT() {
  console.log(`Starting Round 14 UAT on: ${targetUrl}`);
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

  // 1. Screenshot Queue Tab
  const pathQueue = resolve(outDir, "etax-queue.png");
  await page.screenshot({ path: pathQueue, fullPage: false });
  console.log(`Saved screenshot: ${pathQueue}`);

  // 2. Click "ประวัติการนำส่ง" Tab
  const logTabBtn = page.locator('button:has-text("ประวัติการนำส่ง")').first();
  await logTabBtn.click();
  await page.waitForTimeout(600);
  const pathLog = resolve(outDir, "etax-logs.png");
  await page.screenshot({ path: pathLog, fullPage: false });
  console.log(`Saved screenshot: ${pathLog}`);

  // 3. Click "ใบรับรองดิจิทัล" Tab
  const certTabBtn = page.locator('button:has-text("ใบรับรองดิจิทัล")').first();
  await certTabBtn.click();
  await page.waitForTimeout(600);
  const pathCert = resolve(outDir, "etax-cert.png");
  await page.screenshot({ path: pathCert, fullPage: false });
  console.log(`Saved screenshot: ${pathCert}`);

  // 4. Click "ตั้งค่าการเชื่อมต่อ" Tab
  const configTabBtn = page.locator('button:has-text("ตั้งค่าการเชื่อมต่อ")').first();
  await configTabBtn.click();
  await page.waitForTimeout(600);
  const pathConfig = resolve(outDir, "etax-config.png");
  await page.screenshot({ path: pathConfig, fullPage: false });
  console.log(`Saved screenshot: ${pathConfig}`);

  await browser.close();
  console.log(`Console errors detected: ${errorCount}`);
  if (errorCount > 0) {
    console.error("❌ Round 14 UAT failed due to console/page errors!");
    process.exit(1);
  }
  console.log("✅ Round 14 UAT automation completed successfully!");
}

runUAT().catch((err) => {
  console.error("FATAL ERROR in UAT:", err);
  process.exit(1);
});
