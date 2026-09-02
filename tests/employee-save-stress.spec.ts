import { execSync } from 'child_process';
import { expect, test, type Page } from '@playwright/test';

/**
 * Employee save STRESS (2026-09-01) — ลุงจืดรายงาน "บางครั้ง บันทึกไม่ได้"
 * วนแก้ไข+บันทึก UATEMP01 หลายรอบ จับทุกความล้มเหลว:
 *  - toast/error บนจอ, ฟอร์มไม่ปิด, HTTP status (log), Mongo ค่าเปลี่ยนจริงทีละ step
 * รูปแบบ: PIN เปลี่ยนทุกรอบ (สุ่ม seeded) — ตรวจ Mongo หลัง save ทุกรอบ
 */

const ROUNDS = Number(process.env.ROUNDS ?? 8);

function mongoJson(js: string): any {
  const out = execSync(
    `docker exec mongodb mongosh --quiet appdb --eval "JSON.stringify(${js})"`,
    { timeout: 45000 },
  ).toString().trim();
  const start = out.indexOf('[') >= 0 ? out.indexOf('[') : out.indexOf('{');
  return JSON.parse(out.slice(start));
}

async function uiLogin(page: Page) {
  await page.goto('/');
  await page.getByRole('button', { name: /Dev Login/ }).click();
  await page.waitForURL('**/holding*', { timeout: 30000 });
  await page.getByRole('button', { name: /บ้านเชียง/ }).first().click();
  await page.waitForURL('**/workspace*', { timeout: 30000 });
  await page.getByRole('button', { name: /สาขาทดสอบไทย|TST03/ }).first().click();
  await page.getByRole('button', { name: /สำนักงานใหญ่|00001/ }).first().click();
  await page.waitForTimeout(1500);
}

async function ensureEmployeeScreen(page: Page) {
  await page.goto('/employee');
  const search = page.locator('input[placeholder*="ค้นหา พนักงาน"]');
  const devBtn = page.getByRole('button', { name: /Dev Login/ });
  const settled = await expect(search.or(devBtn).first()).toBeVisible({ timeout: 20000 }).then(() => true).catch(() => false);
  if (!settled || (await devBtn.count()) > 0) {
    await uiLogin(page);
    await page.goto('/employee');
    await page.context().storageState({ path: '.auth/user.json' }).catch(() => {});
  }
  await expect(search).toBeVisible({ timeout: 20000 });
}

test.describe.configure({ mode: 'serial', retries: 0 });

test('SAVE-STRESS: แก้ไข+บันทึก UATEMP01 ซ้ำหลายรอบ หาอาการบันทึกไม่ได้', async ({ page }) => {
  test.setTimeout(600000);
  const failures: string[] = [];

  for (let round = 1; round <= ROUNDS; round++) {
    const newPin = String(200000 + Math.floor(Math.random() * 799999));
    let roundError = '';

    try {
      await ensureEmployeeScreen(page);
      const row = page.locator('.bc-list-row').filter({ hasText: 'UATEMP01' }).first();
      let rowVisible = false;
      for (let attempt = 1; attempt <= 3 && !rowVisible; attempt++) {
        await ensureEmployeeScreen(page);
        rowVisible = await row.waitFor({ state: 'visible', timeout: 12000 }).then(() => true).catch(() => false);
      }
      if (!rowVisible) throw new Error('แถว UATEMP01 ไม่แสดง (self-heal 3 ครั้งแล้ว)');

      // เปิดฟอร์มแก้ไข
      let editOpened = false;
      for (let attempt = 1; attempt <= 3 && !editOpened; attempt++) {
        const rowEdit = row.getByRole('button', { name: /แก้ไข|Edit/ }).first();
        if (await rowEdit.count()) await rowEdit.click().catch(() => {});
        else {
          const detail = page.locator('section[aria-label="รายละเอียด"]');
          await detail.locator('button', { hasText: 'แก้ไข' }).first().click().catch(() => {});
        }
        editOpened = await page.getByLabel('PIN').waitFor({ state: 'visible', timeout: 8000 }).then(() => true).catch(() => false);
      }
      if (!editOpened) throw new Error('ฟอร์มแก้ไขไม่เปิด');

      // เปลี่ยน PIN
      const pinInput = page.getByLabel('PIN');
      await pinInput.fill(newPin);

      // บันทึก — สำเร็จ = ฟอร์มปิด (PIN field หาย)
      await page.locator('button', { hasText: /บันทึก|Save/ }).last().click();
      const closed = await pinInput.waitFor({ state: 'hidden', timeout: 15000 }).then(() => true).catch(() => false);

      if (!closed) {
        // เก็บ error บนจอ (toast/ข้อความแดง)
        const alertText = await page.evaluate(`(()=>{
          const alerts=[...document.querySelectorAll('[role="alert"], .text-destructive, [class*="destructive"]')];
          return alerts.map(a=>a.textContent.trim()).filter(Boolean).slice(0,3).join(' | ').slice(0,200);
        })()`);
        throw new Error(`ฟอร์มไม่ปิดหลังบันทึก — error บนจอ: ${alertText || '(ไม่มีข้อความ)'}`);
      }

      // Mongo ทีละ step: PIN เปลี่ยนจริง
      const docs = mongoJson(`db.employees.find({holdingcode:'bc001', code:'UATEMP01'}, {pincode:1}).toArray()`);
      const dbPin = docs[0]?.pincode;
      if (dbPin !== newPin) throw new Error(`Mongo pincode=${dbPin} แต่ตั้งใจ ${newPin}`);
    } catch (err) {
      roundError = String(err).slice(0, 250);
      failures.push(`รอบ ${round}: ${roundError}`);
      console.log(`FAIL round ${round}: ${roundError}`);
    }

    process.stdout.write(`round ${round}/${ROUNDS}: ${roundError ? 'FAIL — ' + roundError.slice(0, 80) : 'ok (pin saved)'}\n`);
  }

  console.log(`\nRESULT: ${ROUNDS - failures.length}/${ROUNDS} save ok`);
  if (failures.length) {
    console.log('FAILURES:\n' + failures.join('\n'));
  }
  expect(failures.length, `บันทึกล้มเหลว ${failures.length}/${ROUNDS} รอบ:\n${failures.join('\n')}`).toBe(0);
});
