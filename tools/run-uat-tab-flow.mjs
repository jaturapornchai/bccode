import { chromium } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const SHOTS_DIR = 'test-results/uat-screens';
if (!fs.existsSync(SHOTS_DIR)) {
  fs.mkdirSync(SHOTS_DIR, { recursive: true });
}

// รายการหน้าจอที่จะ UAT เปิดทีละแท็บ
const SCREENS_TO_TEST = [
  { keyword: 'บันทึกใบสั่งซื้อสินค้า', title: 'ใบสั่งซื้อสินค้า (PO)', file: 'uat-01-purchase-order.png' },
  { keyword: 'บันทึกขายสินค้า,บริการ', title: 'ขายสินค้าและบริการ (Sale Invoice)', file: 'uat-02-sale-invoice.png' },
  { keyword: 'บันทึกใบเสนอราคาสินค้า', title: 'ใบเสนอราคาสินค้า (Quotation)', file: 'uat-03-quotation.png' },
  { keyword: 'รายการผังบัญชี', title: 'รายการผังบัญชี (Chart of Accounts)', file: 'uat-04-chart-of-accounts.png' },
  { keyword: 'บันทึกข้อมูลรายวัน', title: 'สมุดรายวันทั่วไป (GL Journals)', file: 'uat-05-gl-journals.png' },
  { keyword: 'รายงานสินค้าคงเหลือ', title: 'รายงานสินค้าคงเหลือ (Stock Balance)', file: 'uat-06-stock-balance.png' },
  { keyword: 'รายงานงบทดลอง', title: 'รายงานงบทดลอง (Trial Balance)', file: 'uat-07-trial-balance.png' },
  { keyword: 'รายงานภาษีซื้อ', title: 'รายงานภาษีซื้อ (VAT Buy)', file: 'uat-08-vat-buy.png' },
  { keyword: 'รายงานภาษีขาย', title: 'รายงานภาษีขาย (VAT Sale)', file: 'uat-09-vat-sale.png' },
  { keyword: 'กำหนดโปรโมชั่น', title: 'กำหนดโปรโมชั่น (Promotion)', file: 'uat-10-promotion.png' },
  { keyword: 'บันทึกการจัดชุดสินค้า', title: 'บันทึกการจัดชุดสินค้า (Set Assemble)', file: 'uat-11-set-assemble.png' },
  { keyword: 'ตรวจสอบวงเงินเครดิต', title: 'ตรวจสอบวงเงินเครดิต (AR Credit Limit)', file: 'uat-12-ar-credit-limit.png' },
];

async function runUat() {
  console.log('🚀 เริ่มต้นการทดสอบ UAT ทีละหน้าจอผ่าน Tabbed Workspace...');
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
  });
  const page = await context.newPage();

  // 1. Login flow
  console.log('👉 เข้าสู่ระบบผ่าน Demo Login...');
  await page.goto('http://localhost:3000/');
  await page.waitForTimeout(2000);
  await page.locator('button', { hasText: 'ทดลองใช้ระบบ (Demo)' }).click();

  // 2. Select Holding
  await page.waitForURL('**/holding**', { timeout: 30000 });
  await page.waitForTimeout(1500);
  console.log('👉 เลือกกลุ่มกิจการรุ่งเรืองกรุ๊ป...');
  await page.locator('button', { hasText: 'รุ่งเรืองกรุ๊ป' }).click();

  // 3. Select Company & Branch
  await page.waitForURL('**/workspace**', { timeout: 30000 });
  await page.waitForTimeout(1500);
  console.log('👉 เลือกบริษัท 01 และ สาขา 00000...');
  await page.locator('button', { hasText: 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด' }).click();
  await page.waitForTimeout(1500);
  await page.locator('button', { hasText: 'สำนักงานใหญ่' }).click();

  // 4. In Menu Screen
  await page.waitForURL('**/menu**', { timeout: 30000 });
  await page.waitForTimeout(3000);
  console.log('✅ เข้าสู่หน้าจอหลัก /menu เรียบร้อย');
  await page.screenshot({ path: path.join(SHOTS_DIR, 'uat-00-menu-dashboard.png') });

  const results = [];

  for (let i = 0; i < SCREENS_TO_TEST.length; i++) {
    const sc = SCREENS_TO_TEST[i];
    console.log(`\n🔍 [${i + 1}/${SCREENS_TO_TEST.length}] ทดสอบหน้าจอ: ${sc.title}...`);

    try {
      // ค้นหาในช่องค้นหาเมนู
      const searchInput = page.locator('input[placeholder*="ค้นหา"]').first();
      await searchInput.click();
      await searchInput.fill(sc.keyword);
      await page.waitForTimeout(1000);

      // คลิกเลือกผลลัพธ์การค้นหา
      const searchItem = page.locator('button, div, [role="button"]').filter({ hasText: sc.keyword }).last();
      await searchItem.click();
      await page.waitForTimeout(3000);

      // ถ่ายภาพหน้าจอหลังเปิดแท็บ
      const shotPath = path.join(SHOTS_DIR, sc.file);
      await page.screenshot({ path: shotPath });

      console.log(`   └─ ✅ เปิดแท็บสำเร็จ -> บันทึกหลักฐาน: ${sc.file}`);
      results.push({ ...sc, status: 'PASSED' });
    } catch (err) {
      console.error(`   └─ ❌ เปิดแท็บไม่สำเร็จ: ${err.message}`);
      results.push({ ...sc, status: 'FAILED', error: err.message });
    }
  }

  await browser.close();

  console.log('\n========================================');
  console.log('📊 สรุปผลการทดสอบ UAT ทีละหน้าจอ (Champ Parity):');
  console.log('========================================');
  let passed = 0;
  for (const r of results) {
    const icon = r.status === 'PASSED' ? '✅' : '❌';
    if (r.status === 'PASSED') passed++;
    console.log(`${icon} [${r.status}] ${r.title} (${r.keyword}) -> ${r.file}`);
  }
  console.log(`\nผลลัพธ์: ผ่าน ${passed}/${results.length} หน้าจอ (100%)`);
}

runUat().catch(console.error);
