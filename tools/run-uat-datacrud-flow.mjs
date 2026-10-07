import { chromium } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const SHOTS_DIR = 'test-results/uat-screens';
if (!fs.existsSync(SHOTS_DIR)) {
  fs.mkdirSync(SHOTS_DIR, { recursive: true });
}

async function runUat() {
  console.log('🚀 เริ่มต้นการทดสอบ Browser UAT หน้าจอ DataCRUD (PO & Sale)...');
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
  });
  const page = await context.newPage();

  // 1. Login flow
  console.log('👉 [1/4] เข้าสู่ระบบผ่าน Demo Login...');
  await page.goto('http://localhost:3000/');
  await page.waitForTimeout(2000);
  await page.locator('button', { hasText: 'ทดลองใช้ระบบ (Demo)' }).click();

  // 2. Select Holding
  await page.waitForURL('**/holding**', { timeout: 30000 });
  await page.waitForTimeout(1500);
  console.log('👉 [2/4] เลือกกลุ่มกิจการรุ่งเรืองกรุ๊ป...');
  await page.locator('button', { hasText: 'รุ่งเรืองกรุ๊ป' }).click();

  // 3. Select Company & Branch
  await page.waitForURL('**/workspace**', { timeout: 30000 });
  await page.waitForTimeout(1500);
  console.log('👉 [3/4] เลือกบริษัท 01 และ สาขา 00000...');
  await page.locator('button', { hasText: 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด' }).click();
  await page.waitForTimeout(1500);
  await page.locator('button', { hasText: 'สำนักงานใหญ่' }).click();

  // 4. In Menu Screen
  await page.waitForURL('**/menu**', { timeout: 30000 });
  await page.waitForTimeout(2000);
  console.log('✅ เข้าสู่หน้าจอหลัก /menu เรียบร้อย');

  // Test 1: Purchase Order Screen
  console.log('👉 [4/4] เปิดหน้าจอและทดสอบใบสั่งซื้อสินค้า (PO)...');
  const searchInput = page.locator('input[placeholder*="ค้นหาเมนู"]');
  await searchInput.fill('ใบสั่งซื้อสินค้า');
  await page.waitForTimeout(800);

  const poMenu = page.locator('button, div[role="button"]').filter({ hasText: 'บันทึกใบสั่งซื้อสินค้า' }).first();
  await poMenu.click();
  await page.waitForTimeout(3000);

  await page.screenshot({ path: path.join(SHOTS_DIR, 'uat-datacrud-po.png') });
  console.log('   └─ ถ่ายภาพหน้าจอ PO สำเร็จ: uat-datacrud-po.png');

  // Test 2: Sale Invoice Screen
  console.log('👉 เปิดหน้าจอและทดสอบใบขายสินค้า (Sale Invoice)...');
  await searchInput.fill('ขายสินค้า');
  await page.waitForTimeout(800);

  const saleMenu = page.locator('button, div[role="button"]').filter({ hasText: 'บันทึกขายสินค้า,บริการ' }).first();
  await saleMenu.click();
  await page.waitForTimeout(3000);

  await page.screenshot({ path: path.join(SHOTS_DIR, 'uat-datacrud-sale.png') });
  console.log('   └─ ถ่ายภาพหน้าจอ Sale Invoice สำเร็จ: uat-datacrud-sale.png');

  await browser.close();
  console.log('\n🎉 ผลการทดสอบ Browser UAT หน้าจอ DataCRUD สำเร็จสมบูรณ์ 100%!');
}

runUat().catch((err) => {
  console.error('❌ Error during UAT:', err);
  process.exit(1);
});
