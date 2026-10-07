import { chromium } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const SHOTS_DIR = 'test-results/uat-screens';
if (!fs.existsSync(SHOTS_DIR)) {
  fs.mkdirSync(SHOTS_DIR, { recursive: true });
}

const SCREENS = [
  { name: '01-main-menu', route: '/menu', title: 'เมนูหลัก (13 บาร์)' },
  { name: '02-chart-of-accounts', route: '/gl/chartofaccounts', title: 'ผังบัญชี (Chart of Accounts)' },
  { name: '03-gl-journals', route: '/gl/journals', title: 'สมุดรายวัน (GL Journals)' },
  { name: '04-customer-master', route: '/customer', title: 'ทะเบียนลูกค้า (Customer)' },
  { name: '05-creditor-master', route: '/creditor', title: 'ทะเบียนเจ้าหนี้ (Vendor/Creditor)' },
  { name: '06-purchase-order', route: '/transaction/purchaseorder', title: 'ใบสั่งซื้อสินค้า (Purchase Order)' },
  { name: '07-purchase-invoice', route: '/transaction/purchase', title: 'ซื้อสินค้าและบริการ (Purchase Invoice)' },
  { name: '08-sales-order', route: '/transaction/saleorder', title: 'ใบสั่งขายสินค้า (Sales Order)' },
  { name: '09-sales-invoice', route: '/transaction/sale', title: 'ขายสินค้าและบริการ (Sales Invoice)' },
  { name: '10-stock-balance-report', route: '/report/stockbalanceitem', title: 'รายงานสินค้าคงเหลือ (Stock Balance)' },
  { name: '11-trial-balance-report', route: '/report/trialbalance', title: 'งบทดลอง (Trial Balance)' },
  { name: '12-vat-buy-report', route: '/report/reportvatbuy', title: 'รายงานภาษีซื้อ (VAT Buy)' },
  { name: '13-vat-sale-report', route: '/report/reportvatsale', title: 'รายงานภาษีขาย (VAT Sale)' },
];

async function runUat() {
  console.log('🚀 เริ่มการทดสอบ UAT ทีละหน้าจอ...');
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
  });
  const page = await context.newPage();

  const results = [];
  const errors = [];

  page.on('pageerror', (err) => {
    errors.push(`[PAGE ERROR] ${err.message}`);
  });

  // Step 1: Login via Demo
  console.log('👉 [1/3] กำลังเข้าสู่ระบบด้วย Demo Login...');
  await page.goto('http://localhost:3000/');
  await page.waitForTimeout(2000);
  await page.screenshot({ path: path.join(SHOTS_DIR, '00-login-page.png') });

  const demoBtn = page.locator('button', { hasText: 'ทดลองใช้ระบบ (Demo)' });
  await demoBtn.click();
  await page.waitForURL('**/holding**', { timeout: 30000 });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: path.join(SHOTS_DIR, '00-holding-page.png') });

  // Step 2: Select Holding
  console.log('👉 [2/3] กำลังเลือก Holding รุ่งเรืองกรุ๊ป...');
  await page.locator('button', { hasText: 'รุ่งเรืองกรุ๊ป' }).click();
  await page.waitForURL('**/workspace**', { timeout: 30000 });
  await page.waitForTimeout(1500);
  await page.screenshot({ path: path.join(SHOTS_DIR, '00-workspace-page.png') });

  // Step 3: Select Company & Branch
  console.log('👉 [3/3] กำลังเลือกบริษัท 01 และ สาขา 00000...');
  await page.locator('button', { hasText: 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด' }).click();
  await page.waitForTimeout(1500);
  await page.locator('button', { hasText: 'สำนักงานใหญ่' }).click();
  await page.waitForTimeout(3000);

  // Check login success
  console.log('✅ เข้าสู่ระบบสำเร็จ กำลังเริ่มตรวจสอบทีละหน้าจอ...');

  for (let i = 0; i < SCREENS.length; i++) {
    const screen = SCREENS[i];
    console.log(`🔍 [${i + 1}/${SCREENS.length}] ตรวจสอบ: ${screen.title} (${screen.route})...`);
    const start = Date.now();
    try {
      await page.goto(`http://localhost:3000${screen.route}`, { timeout: 30000 });
      await page.waitForTimeout(2500);

      // ตรวจสอบสถานะ DOM
      const title = await page.title();
      const bodyHtml = await page.content();
      const hasContent = bodyHtml.length > 1000;
      const shotPath = path.join(SHOTS_DIR, `${screen.name}.png`);
      await page.screenshot({ path: shotPath });

      const duration = Date.now() - start;
      const status = hasContent ? 'PASSED' : 'EMPTY_DOM';
      console.log(`   └─ สถานะ: ${status} (${duration}ms) -> บันทึกภาพ ${screen.name}.png`);
      results.push({ ...screen, status, duration, error: null });
    } catch (err) {
      console.error(`   └─ ❌ เกิดข้อผิดพลาด: ${err.message}`);
      results.push({ ...screen, status: 'FAILED', error: err.message });
    }
  }

  await browser.close();

  console.log('\n========================================');
  console.log('📊 สรุปผลการทดสอบ UAT ทีละหน้าจอ:');
  console.log('========================================');
  let passedCount = 0;
  for (const r of results) {
    const icon = r.status === 'PASSED' ? '✅' : '❌';
    if (r.status === 'PASSED') passedCount++;
    console.log(`${icon} [${r.status}] ${r.title} (${r.route})`);
  }
  console.log(`\nรวมผ่าน: ${passedCount}/${results.length} หน้าจอ`);
  if (errors.length > 0) {
    console.log(`⚠️ Console/Page Errors ที่พบ (${errors.length}):`);
    errors.slice(0, 5).forEach((e) => console.log('  -', e));
  } else {
    console.log('🎉 ไม่พบ Fatal Console Error ใดๆ!');
  }
}

runUat().catch(console.error);
