import { chromium } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const SHOTS_DIR = 'test-results/uat-phase7';
if (!fs.existsSync(SHOTS_DIR)) {
  fs.mkdirSync(SHOTS_DIR, { recursive: true });
}

const targetUrl = process.env.TARGET_URL || 'https://account.bcaicloud.com';

async function runPhase7Uat() {
  console.log(`🚀 เริ่มต้นการทดสอบ Browser UAT Phase 7 บนเป้าหมาย: ${targetUrl}`);
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    ignoreHTTPSErrors: true,
  });
  const page = await context.newPage();

  try {
    // 1. เข้าหน้า Login
    console.log('👉 [1/5] เข้าสู่ระบบผ่าน Demo Login...');
    await page.goto(`${targetUrl}/`, { waitUntil: 'domcontentloaded', timeout: 30000 });
    await page.waitForTimeout(2000);

    const demoBtn = page.locator('button:has-text("ทดลองใช้ระบบ"), .dev-login-button').first();
    await demoBtn.waitFor({ state: 'visible', timeout: 15000 });
    await demoBtn.click();
    console.log('   ✅ คลิกปุ่มทดลองใช้ระบบ (Demo) สำเร็จ');

    // 2. เลือก Holding
    await page.waitForURL('**/holding**', { timeout: 30000 });
    await page.waitForTimeout(1500);
    console.log('👉 [2/5] เลือกกลุ่มกิจการ...');
    const holdingBtn = page.locator('button').filter({ hasText: /รุ่งเรือง|rungrueng|บ้านเชียง/i }).first();
    await holdingBtn.click();
    console.log('   ✅ เลือกกลุ่มกิจการสำเร็จ');

    // 3. เลือก Company & Branch
    await page.waitForURL('**/workspace**', { timeout: 30000 });
    await page.waitForTimeout(1500);
    console.log('👉 [3/5] เลือกบริษัท 01 และ สาขา 00000...');
    const companyBtn = page.locator('button').filter({ hasText: /บริษัท|01|TST/i }).first();
    await companyBtn.click();
    await page.waitForTimeout(1200);

    const branchBtn = page.locator('button').filter({ hasText: /สำนักงานใหญ่|00000|00001/i }).first();
    await branchBtn.click();
    await page.waitForTimeout(2000);

    // 4. เข้าสู่หน้าจอ Menu
    await page.waitForURL('**/menu**', { timeout: 30000 });
    await page.waitForTimeout(2000);
    console.log('✅ เข้าสู่หน้าจอหลัก /menu เรียบร้อย');
    await page.screenshot({ path: path.join(SHOTS_DIR, '00-menu-main.png') });

    async function searchAndOpenMenu(searchText, targetItemTitle) {
      const searchInput = page.locator('input[placeholder*="ค้นหาเมนู"]').first();
      await searchInput.fill('');
      await searchInput.fill(searchText);
      await page.waitForTimeout(800);
      const targetBtn = page.locator('button').filter({ hasText: targetItemTitle }).first();
      await targetBtn.waitFor({ state: 'visible', timeout: 8000 });
      await targetBtn.click();
      await page.waitForTimeout(2500);
    }

    // 5. ทดสอบเปิด Automatic Billing Note
    console.log('👉 [4/5] UAT ทดสอบหน้าจอบันทึกใบวางบิลอัตโนมัติ (Automatic Billing Note)...');
    await searchAndOpenMenu('วางบิลอัตโนมัติ', 'บันทึกใบวางบิลอัตโนมัติ');
    await page.screenshot({ path: path.join(SHOTS_DIR, '01-billing-note-auto.png') });
    console.log('   ✅ เปิดหน้าจอ Billing Note Auto สำเร็จ บันทึกภาพ 01-billing-note-auto.png');

    // 6. ทดสอบเปิด Transaction Lock
    console.log('👉 [5/5] UAT ทดสอบหน้าจอ Transaction Lock (กำหนดสิทธิ์และล็อกเอกสาร)...');
    await searchAndOpenMenu('Transaction Lock', 'Transaction Lock');
    await page.screenshot({ path: path.join(SHOTS_DIR, '02-transaction-lock.png') });
    console.log('   ✅ เปิดหน้าจอ Transaction Lock สำเร็จ บันทึกภาพ 02-transaction-lock.png');

    // 7. ทดสอบเปิด AP Payment Prepare
    console.log('👉 [6/5] UAT ทดสอบหน้าจอบันทึกเตรียมจ่ายชำระหนี้ (AP Payment Prepare)...');
    await searchAndOpenMenu('เตรียมจ่าย', 'บันทึกเตรียมจ่ายชำระหนี้');
    await page.screenshot({ path: path.join(SHOTS_DIR, '03-ap-payment-prepare.png') });
    console.log('   ✅ เปิดหน้าจอ AP Payment Prepare สำเร็จ บันทึกภาพ 03-ap-payment-prepare.png');

    console.log('🎉 การทดสอบ Browser UAT Phase 7 สำเร็จครบถ้วน 100%!');
  } catch (err) {
    console.error('❌ UAT Error:', err);
    await page.screenshot({ path: path.join(SHOTS_DIR, 'error.png') });
    process.exit(1);
  } finally {
    await browser.close();
  }
}

runPhase7Uat();
