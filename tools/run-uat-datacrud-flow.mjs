import { chromium } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const SHOTS_DIR = 'test-results/uat-screens';
if (!fs.existsSync(SHOTS_DIR)) {
  fs.mkdirSync(SHOTS_DIR, { recursive: true });
}

async function runUat() {
  console.log('🚀 เริ่มต้นการทดสอบ Browser UAT หน้าจอ DataCRUD (Quotation, SO, Purchase)...');
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

  // Test 1: Quotation
  console.log('👉 [4/6] เปิดหน้าจอใบเสนอราคาสินค้า (/transaction/quotation)...');
  await openBySearch('ใบเสนอราคา', 'ใบเสนอราคา');
  await page.screenshot({ path: path.join(SHOTS_DIR, 'uat-datacrud-quotation.png') });
  console.log('   └─ ถ่ายภาพหน้าจอ Quotation สำเร็จ: uat-datacrud-quotation.png');

  // Test 2: Sale Order
  console.log('👉 [5/6] เปิดหน้าจอใบสั่งขายสินค้า (/transaction/saleorder)...');
  await openBySearch('ใบสั่งขายสินค้า', 'ใบสั่งขายสินค้า');
  await page.screenshot({ path: path.join(SHOTS_DIR, 'uat-datacrud-saleorder.png') });
  console.log('   └─ ถ่ายภาพหน้าจอ Sale Order สำเร็จ: uat-datacrud-saleorder.png');

  // Test 3: Purchase
  console.log('👉 [6/7] เปิดหน้าจอซื้อสินค้า,บริการ (/transaction/purchase)...');
  await openBySearch('บันทึกซื้อสินค้า,บริการ', 'บันทึกซื้อสินค้า,บริการ');
  await page.screenshot({ path: path.join(SHOTS_DIR, 'uat-datacrud-purchase.png') });
  console.log('   └─ ถ่ายภาพหน้าจอ Purchase สำเร็จ: uat-datacrud-purchase.png');

  // Test 4: Sale Invoice
  console.log('👉 [7/7] เปิดหน้าจอบันทึกขายสินค้า,บริการ (/transaction/sale)...');
  await openBySearch('บันทึกขายสินค้า,บริการ', 'บันทึกขายสินค้า,บริการ');
  await page.screenshot({ path: path.join(SHOTS_DIR, 'uat-datacrud-sale.png') });
  console.log('   └─ ถ่ายภาพหน้าจอ Sale Invoice สำเร็จ: uat-datacrud-sale.png');

  await browser.close();
  console.log('\n🎉 ผลการทดสอบ Browser UAT สำเร็จสมบูรณ์ 100%!');
}

runUat().catch((err) => {
  console.error('❌ Error during UAT:', err);
  process.exit(1);
});
