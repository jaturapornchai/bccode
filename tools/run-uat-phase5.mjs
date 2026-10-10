import { chromium } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const SHOTS_DIR = 'test-results/uat-phase5';
if (!fs.existsSync(SHOTS_DIR)) {
  fs.mkdirSync(SHOTS_DIR, { recursive: true });
}

const targetUrl = process.env.TARGET_URL || 'https://account.bcaicloud.com';

async function runPhase5Uat() {
  console.log(`🚀 เริ่มต้นการทดสอบ Browser UAT Phase 5 บนเป้าหมาย: ${targetUrl}`);
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    ignoreHTTPSErrors: true,
  });
  const page = await context.newPage();

  try {
    // 1. เข้าหน้า Login
    console.log('👉 [1/6] เข้าสู่ระบบผ่าน Demo Login...');
    await page.goto(`${targetUrl}/`, { waitUntil: 'domcontentloaded', timeout: 30000 });
    await page.waitForTimeout(2000);

    const demoBtn = page.locator('button:has-text("ทดลองใช้ระบบ"), .dev-login-button').first();
    await demoBtn.waitFor({ state: 'visible', timeout: 15000 });
    await demoBtn.click();
    console.log('   ✅ คลิกปุ่มทดลองใช้ระบบ (Demo) สำเร็จ');

    // 2. เลือก Holding
    await page.waitForURL('**/holding**', { timeout: 30000 });
    await page.waitForTimeout(1500);
    console.log('👉 [2/6] เลือกกลุ่มกิจการ...');
    const holdingBtn = page.locator('button').filter({ hasText: /รุ่งเรือง|rungrueng|บ้านเชียง/i }).first();
    await holdingBtn.click();
    console.log('   ✅ เลือกกลุ่มกิจการสำเร็จ');

    // 3. เลือก Company & Branch
    await page.waitForURL('**/workspace**', { timeout: 30000 });
    await page.waitForTimeout(1500);
    console.log('👉 [3/6] เลือกบริษัท 01 และ สาขา 00000...');
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
    await page.screenshot({ path: path.join(SHOTS_DIR, '01-menu-main.png') });

    async function searchAndOpenMenu(searchText, targetItemTitle) {
      const searchInput = page.locator('input[placeholder*="ค้นหาเมนู"]').first();
      await searchInput.fill('');
      await searchInput.fill(searchText);
      await page.waitForTimeout(800);
      const targetBtn = page.locator('button').filter({ hasText: targetItemTitle }).first();
      await targetBtn.waitFor({ state: 'visible', timeout: 8000 });
      await targetBtn.click();
      await page.waitForTimeout(2000);
    }

    // 5. ทดสอบเปิด Screen Font Settings
    console.log('👉 [4/6] UAT ทดสอบหน้าจอ Font Settings (6 เป้าหมาย)...');
    await searchAndOpenMenu('ตัวอักษร', 'เลือกตัวอักษรสำหรับหน้าจอทั่วไป');
    await page.screenshot({ path: path.join(SHOTS_DIR, '02-font-settings-screen.png') });
    console.log('   ✅ เปิดหน้าจอ Font Settings สำเร็จ บันทึกภาพ 02-font-settings-screen.png');

    // 6. ทดสอบเปิด Banking Contacts
    console.log('👉 [5/6] UAT ทดสอบหน้าจอผู้ติดต่อธนาคาร (Banking Contacts)...');
    await searchAndOpenMenu('ผู้ติดต่อธนาคาร', 'บันทึกรายละเอียดผู้ติดต่อธนาคาร');
    await page.screenshot({ path: path.join(SHOTS_DIR, '03-banking-contacts.png') });
    console.log('   ✅ เปิดหน้าจอ Banking Contacts สำเร็จ บันทึกภาพ 03-banking-contacts.png');

    // 7. ทดสอบเปิด Petty Cash Limit
    console.log('👉 [6/6] UAT ทดสอบหน้าจอกำหนดวงเงินสดย่อย (Petty Cash Limit)...');
    await searchAndOpenMenu('วงเงินสดย่อย', 'กำหนดวงเงินสดย่อย');
    await page.screenshot({ path: path.join(SHOTS_DIR, '04-petty-cash-limit.png') });
    console.log('   ✅ เปิดหน้าจอ Petty Cash Limit สำเร็จ บันทึกภาพ 04-petty-cash-limit.png');

    console.log('🎉 การทดสอบ Browser UAT Phase 5 สำเร็จครบถ้วน 100%!');
  } catch (err) {
    console.error('❌ UAT Error:', err);
    process.exitCode = 1;
  } finally {
    await browser.close();
  }
}

runPhase5Uat();
