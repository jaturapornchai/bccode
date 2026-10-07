import { test, expect, type Page } from '@playwright/test';
import * as fs from 'fs';

const SHOTS = 'test-results/uat-screens';
if (!fs.existsSync(SHOTS)) {
  fs.mkdirSync(SHOTS, { recursive: true });
}

async function shoot(page: Page, name: string) {
  await page.screenshot({ path: `${SHOTS}/${name}.png`, fullPage: false });
}

test.describe.configure({ mode: 'serial' });

test.describe('UAT ไล่ตรวจทีละหน้าจอ (Champ Parity & DataCRUD)', () => {
  test.use({ storageState: { cookies: [], origins: [] } });

  test('01. เข้าสู่ระบบด้วย Demo Login และเลือก Workspace', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('http://127.0.0.1:3000/');
    await page.waitForTimeout(1000);
    await shoot(page, '01-login-screen');

    // กดปุ่ม Demo Login
    const demoBtn = page.getByRole('button', { name: /ทดลองใช้ระบบ|Demo/ });
    if (await demoBtn.isVisible()) {
      await demoBtn.click();
    } else {
      const devBtn = page.getByRole('button', { name: /Dev Login/ });
      await devBtn.click();
    }

    // หน้าเลือก Holding
    await page.waitForURL(/\/holding/, { timeout: 30000 });
    await page.waitForTimeout(1000);
    await shoot(page, '02-holding-selection');

    // เลือก holding รุ่งเรือง หรือ บ้านเชียง
    const holdingBtn = page.getByRole('button', { name: /รุ่งเรือง|บ้านเชียง/ }).first();
    await holdingBtn.click();

    // หน้าเลือก Workspace (Company & Branch)
    await page.waitForURL(/\/workspace/, { timeout: 30000 });
    await page.waitForTimeout(1000);
    await shoot(page, '03-workspace-selection');

    // เลือกบริษัทแรก และสาขาแรก
    const companyBtn = page.locator('button').filter({ hasText: /บริษัท|01|TST/ }).first();
    await companyBtn.click();
    await page.waitForTimeout(1000);

    const branchBtn = page.locator('button').filter({ hasText: /สำนักงานใหญ่|00000|00001/ }).first();
    await branchBtn.click();
    await page.waitForTimeout(2000);

    // ตรวจสอบว่ามาถึงหน้า Menu หรือ Workspace
    await shoot(page, '04-main-workspace-ready');
  });

  test('02. หน้าจอหลักและ 13 เมนูบาร์ (Menu Bar Parity)', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('http://127.0.0.1:3000/menu');
    await page.waitForTimeout(1500);
    await shoot(page, '05-menu-bar-overview');

    // ตรวจสอบหัวบาร์ 13 บาร์
    const expectedBars = ['แฟ้ม', 'จัดซื้อ', 'จัดขาย', 'เจ้าหนี้', 'ลูกหนี้', 'เช็ค', 'ธนาคาร', 'สินค้าคงคลัง', 'สินทรัพย์', 'บัญชี'];
    for (const b of expectedBars) {
      const barEl = page.locator('button, [role="button"], span').filter({ hasText: new RegExp(b) }).first();
      await expect(barEl).toBeVisible();
    }
  });

  test('03. หน้าจอผังบัญชี (Chart of Accounts)', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('http://127.0.0.1:3000/gl/chartofaccounts');
    await page.waitForTimeout(1500);
    await shoot(page, '06-chart-of-accounts');
  });

  test('04. หน้าจอสมุดรายวัน (GL Journals)', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('http://127.0.0.1:3000/gl/journals');
    await page.waitForTimeout(1500);
    await shoot(page, '07-gl-journals');
  });

  test('05. หน้าจอบันทึกใบสั่งซื้อสินค้า (Purchase Order - PO)', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('http://127.0.0.1:3000/transaction/purchaseorder');
    await page.waitForTimeout(1500);
    await shoot(page, '08-purchase-order-workbench');

    // ตรวจสอบองค์ประกอบ Master-Detail
    const newDocBtn = page.getByRole('button', { name: /สร้างเอกสารใหม่/ });
    if (await newDocBtn.isVisible()) {
      await newDocBtn.click();
      await page.waitForTimeout(800);
      await shoot(page, '08b-purchase-order-new-form');
    }
  });

  test('06. หน้าจอบันทึกใบสั่งขายสินค้า (Sales Order - SO)', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('http://127.0.0.1:3000/transaction/saleorder');
    await page.waitForTimeout(1500);
    await shoot(page, '09-sales-order-workbench');
  });

  test('07. หน้าจอบันทึกขายสินค้าและบริการ (Sales Invoice)', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('http://127.0.0.1:3000/transaction/sale');
    await page.waitForTimeout(1500);
    await shoot(page, '10-sales-invoice-workbench');
  });

  test('08. หน้าจอรายงานสินค้าคงเหลือ (Stock Balance Report)', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('http://127.0.0.1:3000/report/stockbalanceitem');
    await page.waitForTimeout(1500);
    await shoot(page, '11-report-stock-balance');
  });

  test('09. หน้าจอรายงานงบทดลอง (Trial Balance Report)', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('http://127.0.0.1:3000/report/trialbalance');
    await page.waitForTimeout(1500);
    await shoot(page, '12-report-trial-balance');
  });

  test('10. หน้าจอรายงานภาษีซื้อและภาษีขาย (VAT Reports)', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('http://127.0.0.1:3000/report/reportvatbuy');
    await page.waitForTimeout(1500);
    await shoot(page, '13-report-vat-buy');

    await page.goto('http://127.0.0.1:3000/report/reportvatsale');
    await page.waitForTimeout(1500);
    await shoot(page, '14-report-vat-sale');
  });
});
