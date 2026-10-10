import { test, expect, type Page } from '@playwright/test';
import * as fs from 'fs';

const SHOTS_DIR = 'test-results/uat-phase5';
if (!fs.existsSync(SHOTS_DIR)) {
  fs.mkdirSync(SHOTS_DIR, { recursive: true });
}

async function capture(page: Page, filename: string) {
  await page.screenshot({ path: `${SHOTS_DIR}/${filename}.png`, fullPage: false });
}

test.describe.configure({ mode: 'serial' });

test.describe('UAT Phase 5: Font Picker, Banking Contacts & Petty Cash Limit', () => {
  test('01. ตรวจสอบการเลือกฟอนต์หน้าจอทั่วไปและฟอนต์รายงาน 6 รายการ', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('/menu');
    await page.waitForTimeout(1500);

    // เปิดหน้าจอเลือกตัวอักษรสำหรับหน้าจอทั่วไป
    await page.evaluate(() => {
      window.dispatchEvent(
        new CustomEvent('bc_open_menu_tab', {
          detail: { id: 'window-font-screen', title: 'เลือกตัวอักษรสำหรับหน้าจอทั่วไป', route: '/settings/font' },
        })
      );
    });
    await page.waitForTimeout(1500);
    await capture(page, '01-font-screen-general');

    // ตรวจสอบแท็บการตั้งค่าฟอนต์ 6 เป้าหมาย
    const tabs = ['หน้าจอทั่วไป', 'รายงานทั่วไป', 'ชื่อของรายงาน', 'คอลัมน์ของรายงาน', 'รายการย่อยของรายงาน', 'ชื่อบริษัทในรายงาน'];
    for (const t of tabs) {
      const tabBtn = page.locator('button').filter({ hasText: t }).first();
      await expect(tabBtn).toBeVisible();
    }

    // สลับไปยังแท็บรายงานทั่วไป
    const reportTab = page.locator('button').filter({ hasText: 'รายงานทั่วไป' }).first();
    await reportTab.click();
    await page.waitForTimeout(800);
    await capture(page, '02-font-screen-report-tab');

    // สลับไปยังแท็บชื่อของรายงาน
    const titleTab = page.locator('button').filter({ hasText: 'ชื่อของรายงาน' }).first();
    await titleTab.click();
    await page.waitForTimeout(800);
    await capture(page, '03-font-screen-title-tab');
  });

  test('02. ตรวจสอบหน้าจอทะเบียนผู้ติดต่อธนาคาร (Banking Contacts CRUD & Features)', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('/menu');
    await page.waitForTimeout(1500);

    // เปิดหน้าจอผู้ติดต่อธนาคาร
    await page.evaluate(() => {
      window.dispatchEvent(
        new CustomEvent('bc_open_menu_tab', {
          detail: { id: 'bank-contact', title: 'บันทึกรายละเอียดผู้ติดต่อธนาคาร', route: '/banking/contacts' },
        })
      );
    });
    await page.waitForTimeout(1500);
    await capture(page, '04-banking-contacts-overview');

    // ตรวจสอบส่วนประกอบ Master-Detail
    await expect(page.getByText('แยกตามสถาบันการเงิน')).toBeVisible();
    await expect(page.getByText('รายการผู้ติดต่อ')).toBeVisible();

    // ทดสอบการค้นหา
    const searchInput = page.locator('input[placeholder*="ค้นหาชื่อ, เบอร์, ธนาคาร"]').first();
    await searchInput.fill('กสิกร');
    await page.waitForTimeout(600);
    await capture(page, '05-banking-contacts-search');
    await searchInput.fill('');

    // ทดสอบการกดปุ่ม เพิ่มผู้ติดต่อ
    const addBtn = page.getByRole('button', { name: /เพิ่มผู้ติดต่อ/ }).first();
    await addBtn.click();
    await page.waitForTimeout(800);
    await capture(page, '06-banking-contacts-new-form');

    // ยกเลิกฟอร์ม
    const cancelBtn = page.getByRole('button', { name: 'ยกเลิก' }).first();
    if (await cancelBtn.isVisible()) {
      await cancelBtn.click();
      await page.waitForTimeout(500);
    }
  });

  test('03. ตรวจสอบหน้าจอกำหนดวงเงินสดย่อย (Petty Cash Limit CRUD & TFRS Money)', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('/menu');
    await page.waitForTimeout(1500);

    // เปิดหน้าจอกำหนดวงเงินสดย่อย
    await page.evaluate(() => {
      window.dispatchEvent(
        new CustomEvent('bc_open_menu_tab', {
          detail: { id: 'petty-cash-limit', title: 'กำหนดวงเงินสดย่อย', route: '/banking/petty-cash-limit' },
        })
      );
    });
    await page.waitForTimeout(1500);
    await capture(page, '07-petty-cash-limit-overview');

    // ตรวจสอบการแสดงผล Master-Detail
    await expect(page.getByText('รายการวงเงินสดย่อย')).toBeVisible();
    await expect(page.getByText('ตารางแสดงสถานะวงเงินสดย่อย')).toBeVisible();

    // ทดสอบเปิดฟอร์มเพิ่มวงเงินสดย่อย
    const addBtn = page.getByRole('button', { name: /เพิ่มวงเงินสดย่อย/ }).first();
    await addBtn.click();
    await page.waitForTimeout(800);
    await capture(page, '08-petty-cash-limit-form');

    // ยกเลิกฟอร์ม
    const cancelBtn = page.getByRole('button', { name: 'ยกเลิก' }).first();
    if (await cancelBtn.isVisible()) {
      await cancelBtn.click();
      await page.waitForTimeout(500);
    }

    await capture(page, '09-petty-cash-limit-ready');
  });
});
