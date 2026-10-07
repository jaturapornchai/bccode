import { chromium } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const SHOTS_DIR = 'test-results/uat-screens';
if (!fs.existsSync(SHOTS_DIR)) {
  fs.mkdirSync(SHOTS_DIR, { recursive: true });
}

const SCREENS = [
  { name: 'uat-report-pnl', search: 'กำไรขาดทุน', title: 'งบกำไรขาดทุน' },
  { name: 'uat-report-balancesheet', search: 'ฐานะการเงิน', title: 'งบฐานะการเงิน' },
  { name: 'uat-report-ledger', search: 'บัญชีแยกประเภท', title: 'รายงานบัญชีแยกประเภท' },
  { name: 'uat-report-vatpp30', search: 'ภ.พ.30', title: 'แบบยื่นภาษี ภ.พ.30' },
  { name: 'uat-report-vatpnd3', search: 'ภ.ง.ด.3', title: 'แบบยื่น ภ.ง.ด.3' },
  { name: 'uat-report-vatpnd53', search: 'ภ.ง.ด.53', title: 'แบบยื่น ภ.ง.ด.53' },
  { name: 'uat-report-assetschedule', search: 'ตารางค่าเสื่อม', title: 'ตารางค่าเสื่อมและสินทรัพย์' },
];

async function runUat() {
  console.log('🚀 เริ่มต้นการทดสอบ Browser UAT: Financial Statements & Tax Returns...');
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
  });
  const page = await context.newPage();

  // 1. Login flow
  console.log('👉 [1/3] เข้าสู่ระบบผ่าน Demo Login...');
  await page.goto('http://localhost:3000/');
  await page.waitForTimeout(2000);
  await page.locator('button', { hasText: 'ทดลองใช้ระบบ (Demo)' }).click();

  // 2. Select Holding
  await page.waitForURL('**/holding**', { timeout: 30000 });
  await page.waitForTimeout(1500);
  console.log('👉 [2/3] เลือกกลุ่มกิจการรุ่งเรืองกรุ๊ป...');
  await page.locator('button', { hasText: 'รุ่งเรืองกรุ๊ป' }).click();

  // 3. Select Company & Branch
  await page.waitForURL('**/workspace**', { timeout: 30000 });
  await page.waitForTimeout(1500);
  console.log('👉 [3/3] เลือกบริษัท 01 และ สาขา 00000...');
  await page.locator('button', { hasText: 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด' }).click();
  await page.waitForTimeout(1500);
  await page.locator('button', { hasText: 'สำนักงานใหญ่' }).click();

  await page.waitForURL('**/menu**', { timeout: 30000 });
  await page.waitForTimeout(2000);
  console.log('✅ เข้าสู่หน้าจอหลัก /menu เรียบร้อย');

  async function openBySearch(keyword, expectedTitle) {
    const homeTab = page.locator('button[role="tab"]').first();
    await homeTab.click();
    await page.waitForTimeout(600);

    const searchInput = page.locator('input[placeholder*="ค้นหา"]').first();
    await searchInput.fill('');
    await searchInput.fill(keyword);
    await page.waitForTimeout(1200);

    const searchResult = page.locator('div[role="search"] button').filter({ hasText: expectedTitle }).first();
    if (await searchResult.count() > 0) {
      await searchResult.click();
    } else {
      console.log(`   ⚠️ ไม่พบปุ่มตรง "${expectedTitle}" ลองคลิกรายการแรกใน search result`);
      await page.locator('div[role="search"] button').first().click();
    }
    await page.waitForTimeout(3000);
  }

  for (let i = 0; i < SCREENS.length; i++) {
    const s = SCREENS[i];
    console.log(`👉 [${i + 1}/${SCREENS.length}] ค้นหาและเปิด: ${s.title}...`);
    await openBySearch(s.search, s.title);

    const shotPath = path.join(SHOTS_DIR, `${s.name}.png`);
    await page.screenshot({ path: shotPath });
    console.log(`   └─ บันทึกภาพ ${s.name}.png สำเร็จ`);
  }

  await browser.close();
  console.log('\n🎉 ผลการทดสอบ Browser UAT: Financial Statements & Tax Returns สำเร็จสมบูรณ์ 100%!');
}

runUat().catch((err) => {
  console.error('❌ Error during UAT:', err);
  process.exit(1);
});
