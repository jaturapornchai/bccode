import { test, expect, type BrowserContext, type Page } from '@playwright/test';
import { execFileSync } from 'child_process';

/**
 * UAT Phase 5 (rewritten 2026-10-11) — facade screens → "รอพัฒนา" (docs/kms/champ/design-2026-10-11/facade-screens.md)
 *
 *  A. 30 เมนูที่เคยเป็นจอ UI-only (ข้อมูลจำลอง/localStorage) ต้องเปิดเป็นการ์ด "รอพัฒนา" จริงผ่านแถบเมนู Radix
 *     — มีชิป data-menu-pending, ไม่มีช่องกรอก/ปุ่มบันทึกปลอม
 *     รวมเมนู จัดขาย > บันทึกใบรับเงินมัดจำ: ถอดจอเงาแล้ว แต่จอ ERP จริง (ErpCrudWorkbench, hasLineItems:false) ไม่มีช่องกรอกยอดเงิน
 *     → บันทึกแล้ว totalamount = 0 เสมอ จึงคงเป็นการ์ด "รอพัฒนา" (RETIRED_BACKEND_ROUTES ใน menu-screen-status.ts)
 *  B. (ปิดไว้ด้วย test.describe.fixme) CRUD ทีละ step ของ ใบรับเงินมัดจำ บนจอ ERP จริง พร้อมตรวจ PostgreSQL
 *     (ฐาน holding rungrueng ตาราง doc, transflag 32) ทันทีหลังแต่ละ operation — เปิดคืนพร้อมกันในงานเดียวกับที่ถอด
 *     "/transaction/receivedeposit" ออกจาก RETIRED_BACKEND_ROUTES (เมื่อจอมีช่องยอดเงินระดับหัวเอกสาร แบบ decimal string)
 *     และเปลี่ยน test.fixme 08 เป็น test ปกติ; 09 (backend รับหัวเอกสารว่าง) ต้องยืนยันกฎกับ Champ ก่อน
 *  C. เมนู เลือกตัวอักษรสำหรับหน้าจอทั่วไป (/settings/font) ยังเป็นจอจริง (ฟอนต์หน้าจอมีผลทันที)
 *
 * ใช้ session จาก auth.setup.ts (Demo → rungrueng → บริษัท 01 → สาขา 00000) ผ่าน storageState
 * ลบข้อมูลทดสอบด้วย custname เฉพาะของรอบนี้เท่านั้น (ไม่ใช้ regex กว้าง)
 */

const TENANT_CONTAINER = process.env.PG_CONTAINER ?? 'postgres';
const TENANT_USER = process.env.PG_USER ?? 'postgres';
const TENANT_DB = process.env.PG_TENANT_DB ?? 'rungrueng';
const BUSINESS_CODE = '01';
const RECEIVE_DEPOSIT_FLAG = 32;

/** One statement against the holding database (execFileSync without a shell, so SQL quotes survive on Windows). */
function tenantQuery(sql: string): string {
  const args = ['exec', TENANT_CONTAINER, 'psql', '-U', TENANT_USER, '-d', TENANT_DB, '-v', 'ON_ERROR_STOP=1', '-At', '-c', sql];
  return execFileSync('docker', args, { timeout: 45000 }).toString().trim().split('\n').pop() ?? '';
}

function sqlText(value: string): string {
  return `'${value.replace(/'/g, "''")}'`;
}

/** Facade menus (route → menu path as shown in the Champ menu bar). */
const FACADE_MENUS: Array<[string, string[]]> = [
  ['/sale/reservation', ['จัดขาย', 'ใบสั่งจองสินค้า']],
  ['/sale/dispatch', ['จัดขาย', 'ใบสั่งจ่ายสินค้า']],
  ['/transaction/etax', ['จัดขาย', 'ใบกำกับภาษีอิเล็กทรอนิกส์']],
  ['/ap/payment-prepare', ['เจ้าหนี้', 'บันทึกเตรียมจ่ายชำระหนี้']],
  ['/ar/billing-note-auto', ['ลูกหนี้', 'บันทึกใบวางบิลอัตโนมัติ']],
  ['/banking/contacts', ['ธนาคาร/เงินสด', 'บันทึกรายละเอียดผู้ติดต่อธนาคาร']],
  ['/banking/petty-cash-limit', ['ธนาคาร/เงินสด', 'กำหนดวงเงินสดย่อย']],
  ['/inventory/product-supplier-update', ['สินค้าคงคลัง', 'ปรับปรุง ข้อมูล Supplier ตามสินค้า']],
  ['/inventory/stock-count-handheld', ['สินค้าคงคลัง', 'การตรวจนับสินค้า', 'ตรวจนับสต็อกโดย Hand Held']],
  ['/tools/transaction-lock', ['สินค้าคงคลัง', 'Transaction Lock']],
  ['/settings/journal-format', ['จัดการระบบ', 'กำหนดรูปแบบรายวัน']],
  ['/settings/work-date', ['จัดการระบบ', 'กำหนดวันที่ทำงานเริ่มต้น']],
  ['/settings/product-by-warehouse', ['จัดการระบบ', 'กำหนดสินค้าตามคลัง']],
  ['/settings/default-warehouse', ['จัดการระบบ', 'กำหนดคลังที่เก็บสินค้าเริ่มต้น']],
  ['/settings/currency', ['จัดการระบบ', 'กำหนดสกุลเงิน']],
  ['/settings/main-map', ['จัดการระบบ', 'กำหนดแผนที่หลัก']],
  ['/settings/credit-debit-reasons', ['จัดการระบบ', 'กำหนดสาเหตุการลดหนี้/เพิ่มหนี้']],
  ['/settings/payment-terms', ['จัดการระบบ', 'เงื่อนไขการชำระเงิน']],
  ['/settings/customer-inquiry', ['จัดการระบบ', 'แสดงข้อมูลลูกค้า']],
  ['/settings/product-inquiry', ['จัดการระบบ', 'แสดงข้อมูลสินค้า']],
  ['/settings/font-report', ['หน้าต่าง', 'เลือกตัวอักษรในการแสดงผลต่าง', 'เลือกตัวอักษรสำหรับรายงานทั่วไป']],
  ['/settings/font-report-title', ['หน้าต่าง', 'เลือกตัวอักษรในการแสดงผลต่าง', 'เลือกตัวอักษรสำหรับชื่อของรายงาน']],
  ['/settings/font-report-column', ['หน้าต่าง', 'เลือกตัวอักษรในการแสดงผลต่าง', 'เลือกตัวอักษรสำหรับคอลัมน์ของรายงาน']],
  ['/settings/font-report-detail', ['หน้าต่าง', 'เลือกตัวอักษรในการแสดงผลต่าง', 'เลือกตัวอักษรสำหรับรายการย่อยของรายงาน']],
  ['/settings/font-report-company', ['หน้าต่าง', 'เลือกตัวอักษรในการแสดงผลต่าง', 'เลือกตัวอักษรสำหรับชื่อบริษัทในรายงาน']],
  ['/settings/about', ['ช่วยเหลือ', 'About']],
  ['/settings/standard-software', ['ช่วยเหลือ', 'มาตรฐานซอฟท์แวร์ของกรมสรรพากร (Standard software)']],
];

/** Real ERP screens kept as รอพัฒนา because the workbench cannot hold the document's money yet. */
const INCOMPLETE_ERP_MENUS: Array<[string, string[]]> = [
  ['/transaction/receivedeposit', ['จัดขาย', 'บันทึกใบรับเงินมัดจำ']],
];

/**
 * Open /menu and walk the Radix menu bar: head button → submenu triggers → final menuitem.
 * Submenu panels open far from the pointer, so Playwright's instant mouse jump would leave Radix's pointer-grace
 * area and close them — the walk uses focus + keyboard (ArrowRight opens a submenu, Enter selects) instead.
 */
async function openMenuPath(page: Page, path: string[], reload = true) {
  if (reload) {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto('/menu');
  }
  const [head, ...rest] = path;
  const headButton = page.locator('nav').getByRole('button', { name: head, exact: true }).first();
  await headButton.waitFor({ state: 'visible', timeout: 30000 });
  // the dev server re-renders the bar while hydrating — retry the whole walk until it sticks
  await expect(async () => {
    await page.keyboard.press('Escape');
    await headButton.click({ timeout: 5000 });
    for (const [index, label] of rest.entries()) {
      const entry = page.getByRole('menuitem', { name: label, exact: true }).first();
      await entry.waitFor({ state: 'visible', timeout: 5000 });
      await entry.focus();
      await page.keyboard.press(index === rest.length - 1 ? 'Enter' : 'ArrowRight');
    }
    // the selected item closes the menu — if it is still open the Enter landed on a re-rendered node, so retry
    await expect(page.getByRole('menu')).toHaveCount(0, { timeout: 3000 });
  }).toPass({ timeout: 45000 });
}
// ทั้งไฟล์ใช้ session เดียว (browser context เดียว) เรียงลำดับ — refresh token ของ Demo หมุนเวียนทุกครั้งที่ใช้
// จึงเปิด context ใหม่จาก storageState เดียวกันซ้ำหลายครั้งไม่ได้ (ครั้งที่ 2 เป็นต้นไปจะถูกเด้งไปหน้า login)
test.describe.configure({ mode: 'serial' });

let context: BrowserContext;
let page: Page;

test.beforeAll(async ({ browser }) => {
  context = await browser.newContext({
    baseURL: process.env.PW_BASE_URL ?? 'http://127.0.0.1:3000',
    storageState: '.auth/user.json',
    viewport: { width: 1440, height: 900 },
  });
  page = await context.newPage();
});

test.afterAll(async () => {
  await context.close();
});

test.describe('A. เมนูที่เคยเป็นจอจำลอง (UI-only) ต้องเป็นการ์ด "รอพัฒนา"', () => {
  test.describe.configure({ timeout: 90000 });

  const PLANNED_MENUS = [...FACADE_MENUS, ...INCOMPLETE_ERP_MENUS];

  test('รายการเมนูครบ 30 เมนู + ใบรับเงินมัดจำ ไม่ซ้ำ', () => {
    expect(FACADE_MENUS).toHaveLength(30);
    expect(PLANNED_MENUS).toHaveLength(31);
    expect(new Set(PLANNED_MENUS.map(([route]) => route)).size).toBe(31);
  });

  test('เปิดครบ 31 เมนูผ่านแถบเมนู → การ์ด รอพัฒนา ไม่มีช่องกรอก/ปุ่มบันทึกปลอม', async () => {
    test.setTimeout(900000);
    const failures: string[] = [];
    for (const [index, [route, path]] of PLANNED_MENUS.entries()) {
      try {
        await test.step(`${path.join(' › ')} (${route})`, async () => {
          await openMenuPath(page, path, index === 0);
          const chip = page.locator(`[data-menu-pending="${route}"]`);
          await expect(chip).toBeVisible({ timeout: 15000 });
          await expect(chip).toHaveText('รอพัฒนา');
          const card = chip.locator('xpath=ancestor::*[contains(@class,"min-h-[420px]")][1]');
          await expect(card).toContainText('เมนูในแผนพัฒนา');
          await expect(card).toContainText('ยังบันทึกหรือประมวลผลข้อมูลไม่ได้');
          // การ์ดต้องไม่มีช่องกรอกหรือปุ่มบันทึกปลอมของจอเดิม
          await expect(card.locator('input, textarea, select')).toHaveCount(0);
          await expect(card.getByRole('button', { name: /บันทึก|เพิ่ม|ลบ/ })).toHaveCount(0);
        });
      } catch (error) {
        failures.push(`${route}: ${(error as Error).message.split('\n')[0]}`);
      }
    }
    expect(failures, failures.join('\n')).toEqual([]);
  });
});

// ปิดไว้จนกว่าจอ ERP ใบรับเงินมัดจำจะมีช่องกรอกยอดเงิน (ดูหัวไฟล์) — ตอนนี้เมนูเป็นการ์ดรอพัฒนา เปิดจอ ERP ไม่ได้
test.describe.fixme('B. จัดขาย › บันทึกใบรับเงินมัดจำ = จอ ERP จริง (CRUD + PostgreSQL)', () => {
  test.describe.configure({ timeout: 90000 });

  const seed = Date.now().toString(36).slice(-5).toUpperCase();
  const customer = `UAT-มัดจำ-${seed}`;
  const customerEdge = `UAT-มัดจำ-ซ้ำ-${seed}`;
  const where = (name: string) =>
    `transflag = ${RECEIVE_DEPOSIT_FLAG} AND businesscode = ${sqlText(BUSINESS_CODE)} AND custname = ${sqlText(name)}`;
  const countDocs = (name: string) => Number(tenantQuery(`SELECT count(*) FROM doc WHERE ${where(name)}`));
  const countOtherDeposits = () =>
    Number(
      tenantQuery(
        `SELECT count(*) FROM doc WHERE transflag = ${RECEIVE_DEPOSIT_FLAG} AND businesscode = ${sqlText(BUSINESS_CODE)} AND custname NOT LIKE ${sqlText(`UAT-มัดจำ-%-${seed}`)} AND custname <> ${sqlText(customer)}`,
      ),
    );
  let targetId = '';
  let targetDocno = '';
  let otherDepositsBefore = 0;

  const openWorkbench = async (page: Page) => {
    await openMenuPath(page, ['จัดขาย', 'บันทึกใบรับเงินมัดจำ']);
    await expect(page.getByRole('button', { name: /สร้างเอกสารใหม่/ })).toBeVisible({ timeout: 20000 });
  };

  const fillForm = async (page: Page, fields: { custname: string; description: string; docno?: string }) => {
    await page.getByPlaceholder('ระบุชื่อคู่ค้า / ผู้ติดต่อ').fill(fields.custname);
    await page.getByPlaceholder('บันทึกข้อความเพิ่มเติม').fill(fields.description);
    if (fields.docno !== undefined) {
      await page.getByPlaceholder('ระบบออกเลขที่เอกสารให้อัตโนมัติเมื่อบันทึก').fill(fields.docno);
    }
  };

  const saveForm = async (page: Page) => {
    const saved = page.waitForResponse(
      (res) => res.url().includes('/receivedeposit') && ['POST', 'PUT'].includes(res.request().method()),
      { timeout: 20000 },
    );
    await page.getByRole('button', { name: /บันทึกเอกสาร/ }).first().click();
    return saved;
  };

  test.afterAll(() => {
    // ล้างเฉพาะข้อมูลทดสอบของรอบนี้ (custname ที่มี seed) — ห้ามใช้ regex กว้าง
    for (const name of [customer, customerEdge]) {
      tenantQuery(`DELETE FROM doc WHERE ${where(name)}`);
    }
  });

  test('01. เมนูเปิดจอ ERP จริง ไม่ใช่การ์ดรอพัฒนา และเริ่มจากไม่มีข้อมูลทดสอบ', async () => {
    otherDepositsBefore = countOtherDeposits();
    expect(countDocs(customer)).toBe(0);
    await openWorkbench(page);
    await expect(page.locator('[data-menu-pending]')).toHaveCount(0);
    await expect(page.getByText('บันทึกใบรับเงินมัดจำ').first()).toBeVisible();
    await expect(page.getByPlaceholder('ค้นหาเลขที่, คู่ค้า, รายละเอียด...')).toBeVisible();
  });

  test('02. Create → แถวใหม่ใน PostgreSQL ค่าถูกต้อง', async () => {
    await openWorkbench(page);
    await page.getByRole('button', { name: /สร้างเอกสารใหม่/ }).first().click();
    await fillForm(page, { custname: customer, description: 'เงินมัดจำค่าวัสดุก่อสร้างงวดแรก' });
    const res = await saveForm(page);
    expect(res.status(), 'POST /receivedeposit').toBeLessThan(300);

    // ตรวจ PostgreSQL ทันทีหลัง Create
    expect(countDocs(customer)).toBe(1);
    const row = tenantQuery(
      `SELECT id || '|' || docno || '|' || description || '|' || coalesce(iscancel::text, 'null') FROM doc WHERE ${where(customer)}`,
    ).split('|');
    [targetId, targetDocno] = row;
    expect(targetId).toMatch(/^[0-9]+$/);
    expect(targetDocno.length).toBeGreaterThan(0);
    expect(row[2]).toBe('เงินมัดจำค่าวัสดุก่อสร้างงวดแรก');
    expect(row[3]).not.toBe('true');

    // และเห็นในรายการหน้าจอ
    await expect(page.getByRole('row').filter({ hasText: customer })).toHaveCount(1);
  });

  test('03. Update → แถวเดิมเปลี่ยน (id เดิม ไม่เกิดแถวใหม่/orphan)', async () => {
    test.skip(!targetId, 'ต้องผ่านขั้น Create ก่อน');
    await openWorkbench(page);
    // ค้นด้วยเลขที่เอกสาร — backend กรองคำค้น (q) ที่ docno/custcode เท่านั้น (ชื่อลูกค้า/รายละเอียดไม่ถูกค้น ดู open issue)
    await page.getByPlaceholder('ค้นหาเลขที่, คู่ค้า, รายละเอียด...').fill(targetDocno);
    const row = page.getByRole('row').filter({ hasText: customer });
    await expect(row).toHaveCount(1);
    await row.getByRole('button', { name: 'แก้ไขเอกสาร' }).click();
    await page.getByPlaceholder('บันทึกข้อความเพิ่มเติม').fill('เงินมัดจำค่าวัสดุก่อสร้าง (แก้ไขรายละเอียด)');
    const res = await saveForm(page);
    expect(res.status(), 'PUT /receivedeposit').toBeLessThan(300);

    // ตรวจ PostgreSQL ทันทีหลัง Update
    expect(countDocs(customer)).toBe(1);
    const after = tenantQuery(`SELECT id || '|' || docno || '|' || description FROM doc WHERE ${where(customer)}`).split('|');
    expect(after[0]).toBe(targetId);
    expect(after[1]).toBe(targetDocno);
    expect(after[2]).toBe('เงินมัดจำค่าวัสดุก่อสร้าง (แก้ไขรายละเอียด)');
  });

  test('04. Edge: อักขระพิเศษ/quote/HTML ถูกเก็บตามที่พิมพ์ (parameterized)', async () => {
    test.skip(!targetId, 'ต้องผ่านขั้น Create ก่อน');
    const special = `มัดจำ "พิเศษ" 'ทดสอบ' <b>x</b>; DROP TABLE doc; --`;
    await openWorkbench(page);
    // ค้นด้วยเลขที่เอกสาร — backend กรองคำค้น (q) ที่ docno/custcode เท่านั้น (ชื่อลูกค้า/รายละเอียดไม่ถูกค้น ดู open issue)
    await page.getByPlaceholder('ค้นหาเลขที่, คู่ค้า, รายละเอียด...').fill(targetDocno);
    await page.getByRole('row').filter({ hasText: customer }).getByRole('button', { name: 'แก้ไขเอกสาร' }).click();
    await page.getByPlaceholder('บันทึกข้อความเพิ่มเติม').fill(special);
    const res = await saveForm(page);
    expect(res.status()).toBeLessThan(300);

    expect(tenantQuery(`SELECT description FROM doc WHERE ${where(customer)}`)).toBe(special);
    // ตารางยังอยู่ และแถวเดิม id เดิม
    expect(tenantQuery(`SELECT id FROM doc WHERE ${where(customer)}`)).toBe(targetId);
  });

  test('05. Edge: เลขที่เอกสารซ้ำ → backend ปฏิเสธ (success:false) + error ไทย และไม่เกิดสองแถวที่เลขที่เดียวกัน', async () => {
    test.skip(!targetDocno, 'ต้องผ่านขั้น Create ก่อน');
    await openWorkbench(page);
    await page.getByRole('button', { name: /สร้างเอกสารใหม่/ }).first().click();
    await fillForm(page, { custname: customerEdge, description: 'ทดสอบเลขที่ซ้ำ', docno: targetDocno });
    const res = await saveForm(page);
    // สัญญา error ของ backend: ความผิดที่ผู้ใช้แก้ได้ตอบ success:false (ไม่ใช่ 5xx) — ห้ามนับเป็นบันทึกสำเร็จ
    expect(res.status(), 'เลขที่ซ้ำต้องไม่ใช่ 5xx').toBeLessThan(500);
    expect((await res.json()).success, 'เลขที่ซ้ำต้องถูกปฏิเสธ (success:false)').toBe(false);
    // ข้อความไทยจาก backend (gl_err_erp_doc_duplicate): เลขที่เอกสาร … ถูกใช้ไปแล้ว
    // (ตัวประกาศเส้นทางของ Next ก็เป็น role=alert — กรองด้วยข้อความเพื่อไม่ให้ชน strict mode)
    await expect(page.getByRole('alert').filter({ hasText: 'ถูกใช้ไปแล้ว' })).toBeVisible();

    const sameDocno = Number(
      tenantQuery(
        `SELECT count(*) FROM doc WHERE transflag = ${RECEIVE_DEPOSIT_FLAG} AND businesscode = ${sqlText(BUSINESS_CODE)} AND docno = ${sqlText(targetDocno)}`,
      ),
    );
    expect(sameDocno, 'เลขที่เอกสารเดียวกันต้องมีแถวเดียว').toBe(1);
    // เอกสารเดิมไม่ถูกทับ
    expect(tenantQuery(`SELECT custname FROM doc WHERE id = ${targetId}`)).toBe(customer);
  });

  // บั๊กจริงที่พบระหว่าง UAT: handleSaveDoc (src/app/crud/erp-crud-workbench.tsx) ไม่เช็ก loading/ref จึงยิง POST ซ้ำ → 2 เอกสาร
  // นอกขอบเขตงานนี้ (รายงานใน open issue) — ใช้ fixme (ข้าม ไม่นับเป็นผ่าน) ไม่ใช้ test.fail เพราะ test.fail ผ่านได้ด้วยเหตุอื่น
  // (selector เปลี่ยน/timeout) ซ่อน regression; เมื่อแก้ที่ root cause แล้วให้เปลี่ยนเป็น test ปกติใน commit เดียวกัน
  test.fixme('06. Edge: ดับเบิลคลิกบันทึกเอกสารใหม่ → ได้แถวเดียว', async () => {
    // เคลียร์แถวทดสอบ edge ของขั้น 05 (ถ้า backend ยอมรับ) เพื่อให้ขั้นนี้นับจากศูนย์
    tenantQuery(`DELETE FROM doc WHERE ${where(customerEdge)}`);
    await openWorkbench(page);
    await page.getByRole('button', { name: /สร้างเอกสารใหม่/ }).first().click();
    await fillForm(page, { custname: customerEdge, description: 'ทดสอบดับเบิลคลิกบันทึก' });
    await page.getByRole('button', { name: /บันทึกเอกสาร/ }).first().dblclick();
    await expect(page.getByRole('row').filter({ hasText: customerEdge }).first()).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(1500); // ให้คำขอที่สองถ้ามี จบก่อนตรวจฐานข้อมูล

    expect(countDocs(customerEdge), 'ดับเบิลคลิกต้องไม่สร้างเอกสารซ้ำ').toBe(1);
  });

  // ยังไม่มีปุ่มยกเลิกเอกสารทั้งใบบนจอ ERP (backend มี POST /transaction/:module/:id/cancel แล้ว) — เปิดทดสอบเมื่อมีปุ่ม
  test.fixme('07. ยกเลิกเอกสาร (header-level cancel) → iscancel = true + แถว erp_doc_events', async () => {
    // ขั้นที่ต้องตรวจเมื่อมีปุ่ม: iscancel = true ใน doc และมีแถว action = 'cancel' ใน erp_doc_events ของเอกสารนี้
  });

  // ช่องว่างของจอ ERP จริง (ErpCrudWorkbench แสดงยอดรวมเฉพาะโมดูล hasLineItems:true — erp-transaction.ts receivedeposit เป็น false):
  // ไม่มีช่องกรอกยอดเงินมัดจำ → บันทึกแล้ว doc.totalamount = 0 เสมอ — design facade-screens.md §4 UAT ข้อ 1-2 (ยอดเป็น numeric ตรงที่กรอก) ทำไม่ได้
  // เปิดทดสอบเมื่อจอมีช่องยอดเงินระดับหัวเอกสาร (decimal string) แล้วปรับ locator ให้ตรงช่องจริง
  test.fixme('08. ยอดเงินมัดจำ: Create + Update → doc.totalamount เป็น numeric ตรงที่กรอก (แถว id เดิม)', async () => {
    const amountCustomer = `UAT-มัดจำ-ยอด-${seed}`;
    try {
      await openWorkbench(page);
      await page.getByRole('button', { name: /สร้างเอกสารใหม่/ }).first().click();
      await fillForm(page, { custname: amountCustomer, description: 'ทดสอบยอดเงินมัดจำ' });
      await page.getByLabel('จำนวนเงินมัดจำ').fill('50000.00');
      const created = await saveForm(page);
      expect(created.status(), 'POST /receivedeposit').toBeLessThan(300);
      expect(tenantQuery(`SELECT totalamount::text FROM doc WHERE ${where(amountCustomer)}`)).toBe('50000.00000000');
      const amountId = tenantQuery(`SELECT id FROM doc WHERE ${where(amountCustomer)}`);

      await page.getByLabel('จำนวนเงินมัดจำ').fill('35000.50');
      const updated = await saveForm(page);
      expect(updated.status(), 'PUT /receivedeposit').toBeLessThan(300);
      expect(tenantQuery(`SELECT id || '|' || totalamount::text FROM doc WHERE ${where(amountCustomer)}`)).toBe(`${amountId}|35000.50000000`);
    } finally {
      tenantQuery(`DELETE FROM doc WHERE ${where(amountCustomer)}`);
    }
  });

  // backend ปัจจุบันไม่ตรวจลูกค้า/ยอดของใบรับเงินมัดจำ (SaveDoc รับ custname ว่างและยอด 0) — ต้องยืนยันกฎกับ Champ ก่อนกำหนดว่าควรปฏิเสธ
  test.fixme('09. Edge: บันทึกหัวเอกสารว่าง (ไม่มีลูกค้า/ยอด) → ต้องไม่เกิดแถวใหม่', async () => {
    const before = countOtherDeposits();
    await openWorkbench(page);
    await page.getByRole('button', { name: /สร้างเอกสารใหม่/ }).first().click();
    const res = await saveForm(page);
    expect((await res.json()).success, 'หัวเอกสารว่างต้องถูกปฏิเสธ (success:false)').toBe(false);
    expect(countOtherDeposits()).toBe(before);
  });

  test('10. Delete → เอกสารถูกตั้งธง isdelete (soft delete) และเอกสารอื่นไม่ถูกแตะ', async () => {
    test.skip(!targetId, 'ต้องผ่านขั้น Create ก่อน');
    const edgeBefore = countDocs(customerEdge);
    await openWorkbench(page);
    // ค้นด้วยเลขที่เอกสาร — backend กรองคำค้น (q) ที่ docno/custcode เท่านั้น (ชื่อลูกค้า/รายละเอียดไม่ถูกค้น ดู open issue)
    await page.getByPlaceholder('ค้นหาเลขที่, คู่ค้า, รายละเอียด...').fill(targetDocno);
    const row = page.getByRole('row').filter({ hasText: customer });
    await expect(row).toHaveCount(1);
    await row.getByRole('button', { name: 'ลบเอกสาร' }).click();
    const confirmButton = page.getByRole('dialog').getByRole('button', { name: 'ลบเอกสาร', exact: true });
    await expect(confirmButton).toBeVisible();
    await page.waitForTimeout(600); // dialog ยืนยันเมินคลิกที่มาเร็วกว่า 400ms หลังเปิด (CONFIRM_EARLY_CLICK_MS)
    await confirmButton.click();

    // backend DeleteDoc เป็น soft delete (UPDATE doc SET isdelete = true) — แถวยังอยู่แต่ต้องถูกตั้งธง isdelete และไม่เหลือแถวที่ยังใช้งาน
    await expect
      .poll(() => tenantQuery(`SELECT COALESCE(isdelete, false) FROM doc WHERE id = ${targetId}`), { timeout: 15000 })
      .toBe('t');
    expect(tenantQuery(`SELECT count(*) FROM doc WHERE ${where(customer)} AND NOT COALESCE(isdelete, false)`)).toBe('0');
    // เอกสารรับเงินมัดจำอื่น (ไม่ใช่ของทดสอบ) ต้องเท่าเดิม และเอกสาร edge ของขั้น 06 ต้องไม่ถูกแตะ
    expect(countOtherDeposits()).toBe(otherDepositsBefore);
    expect(countDocs(customerEdge)).toBe(edgeBefore);
  });
});

test.describe('C. เลือกตัวอักษรสำหรับหน้าจอทั่วไป ยังเป็นจอจริง', () => {
  test.describe.configure({ timeout: 90000 });

  test('เปิดจอ เลือกฟอนต์แล้วมีผลทันที และไม่ใช่การ์ดรอพัฒนา', async () => {
    await openMenuPath(page, ['หน้าต่าง', 'เลือกตัวอักษรในการแสดงผลต่าง', 'เลือกตัวอักษรสำหรับหน้าจอทั่วไป']);
    await expect(page.getByRole('heading', { name: 'เลือกตัวอักษรสำหรับหน้าจอทั่วไป' })).toBeVisible({ timeout: 15000 });
    await expect(page.locator('[data-menu-pending]')).toHaveCount(0);

    const notoCard = page.locator('h3', { hasText: 'Noto Thai' }).first();
    await notoCard.click();
    await expect(page.locator('html')).toHaveAttribute('data-app-font', 'noto-sans-thai');

    // คืนค่าเริ่มต้น
    await page.getByRole('button', { name: /คืนค่าเริ่มต้น/ }).click();
    await expect(page.locator('html')).toHaveAttribute('data-app-font', 'inter');
  });
});
