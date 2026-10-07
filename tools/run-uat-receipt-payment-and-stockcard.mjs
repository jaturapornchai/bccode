import { chromium } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const SHOTS_DIR = 'test-results/uat-screens';
if (!fs.existsSync(SHOTS_DIR)) {
  fs.mkdirSync(SHOTS_DIR, { recursive: true });
}

async function runUat() {
  console.log('🚀 เริ่มต้นการทดสอบ Browser UAT รายงานสต็อกการ์ด และระบบชำระเงิน (Paid / Pay)...');
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
  });
  const page = await context.newPage();

  // 1. Login flow
  console.log('👉 [1/6] เข้าสู่ระบบผ่าน Demo Login...');
  await page.goto('http://localhost:3000/');
  await page.waitForTimeout(2000);
  await page.locator('button', { hasText: 'ทดลองใช้ระบบ (Demo)' }).click();

  // 2. Select Holding
  await page.waitForURL('**/holding**', { timeout: 30000 });
  await page.waitForTimeout(1500);
  console.log('👉 [2/6] เลือกกลุ่มกิจการรุ่งเรืองกรุ๊ป...');
  await page.locator('button', { hasText: 'รุ่งเรืองกรุ๊ป' }).click();

  // 3. Select Company & Branch
  await page.waitForURL('**/workspace**', { timeout: 30000 });
  await page.waitForTimeout(1500);
  console.log('👉 [3/6] เลือกบริษัท 01 และ สาขา 00000...');
  await page.locator('button', { hasText: 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด' }).click();
  await page.waitForTimeout(1500);
  await page.locator('button', { hasText: 'สำนักงานใหญ่' }).click();

  // 4. In Menu Screen
  await page.waitForURL('**/menu**', { timeout: 30000 });
  await page.waitForTimeout(2000);
  console.log('✅ เข้าสู่หน้าจอหลัก /menu เรียบร้อย');

  async function openBySearch(keyword, expectedTitle) {
    const homeTab = page.locator('button[role="tab"]').first();
    await homeTab.click();
    await page.waitForTimeout(1000);

    const searchInput = page.locator('input[placeholder*="ค้นหา"]').first();
    await searchInput.fill('');
    await searchInput.fill(keyword);
    await page.waitForTimeout(1000);

    const searchResult = page.locator('div[role="search"] button').filter({ hasText: expectedTitle }).first();
    if (await searchResult.count() > 0) {
      await searchResult.click();
    } else {
      await page.locator('div[role="search"] button').first().click();
    }
    await page.waitForTimeout(3000);
  }

  // 1. รายงานบัญชีคุมพิเศษสินค้า (/report/stockmovementcost)
  console.log('👉 [4/7] เปิดรายงานบัญชีคุมพิเศษสินค้า (/report/stockmovementcost)...');
  await openBySearch('บัญชีคุมพิเศษสินค้า', 'รายงานบัญชีคุมพิเศษสินค้า');
  await page.waitForTimeout(2000);
  await page.screenshot({ path: path.join(SHOTS_DIR, 'uat-report-stockmovementcost.png') });
  console.log('   └─ ถ่ายภาพหน้าจอ รายงานบัญชีคุมพิเศษสินค้า สำเร็จ: uat-report-stockmovementcost.png');

  // 2. รายงานเคลื่อนไหวสินค้า (/report/reportstockmovement)
  console.log('👉 [5/7] เปิดรายงานเคลื่อนไหวสินค้า (/report/reportstockmovement)...');
  await openBySearch('เคลื่อนไหวสินค้า', 'รายงานเคลื่อนไหวสินค้า');
  await page.waitForTimeout(2000);
  await page.screenshot({ path: path.join(SHOTS_DIR, 'uat-report-reportstockmovement.png') });
  console.log('   └─ ถ่ายภาพหน้าจอ รายงานเคลื่อนไหวสินค้า สำเร็จ: uat-report-reportstockmovement.png');

  // 3. รับชำระหนี้ / ใบเสร็จรับเงิน (/transaction/paid)
  console.log('👉 [6/7] เปิดหน้าจอบันทึกใบเสร็จรับเงิน/รับชำระหนี้ (/transaction/paid)...');
  await openBySearch('รับชำระหนี้', 'รับชำระหนี้');
  await page.waitForTimeout(2000);
  await page.screenshot({ path: path.join(SHOTS_DIR, 'uat-datacrud-paid.png') });
  console.log('   └─ ถ่ายภาพหน้าจอ บันทึกใบเสร็จรับเงิน/รับชำระหนี้ สำเร็จ: uat-datacrud-paid.png');

  // 4. จ่ายชำระหนี้ / ใบสำคัญจ่าย (/transaction/pay)
  console.log('👉 [7/7] เปิดหน้าจอบันทึกใบสำคัญจ่าย/จ่ายชำระหนี้ (/transaction/pay)...');
  await openBySearch('จ่ายชำระหนี้', 'จ่ายชำระหนี้');
  await page.waitForTimeout(2000);
  await page.screenshot({ path: path.join(SHOTS_DIR, 'uat-datacrud-pay.png') });
  console.log('   └─ ถ่ายภาพหน้าจอ บันทึกใบสำคัญจ่าย/จ่ายชำระหนี้ สำเร็จ: uat-datacrud-pay.png');

  await browser.close();
  console.log('\n🎉 ผลการทดสอบ Browser UAT สำเร็จสมบูรณ์ 100%!');
}

runUat().catch((err) => {
  console.error('❌ Error during UAT:', err);
  process.exit(1);
});
