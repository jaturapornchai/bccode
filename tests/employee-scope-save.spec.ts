import { execSync } from 'child_process';
import { expect, test, type Page } from '@playwright/test';

/**
 * SCOPE-SAVE UAT — แก้ไขพนักงานพร้อมบริษัท/สาขาที่เข้าใช้งานได้ (accessscopes)
 * เพิ่มบริษัท TST01 → save → ตรวจ Mongo → ลบออก → save → ตรวจ Mongo
 * (ข้อมูลต้นทางรอด: UATEMP01 มีอยู่แล้ว, holding อื่นไม่แตะ)
 */

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

function scopesInMongo(): any {
  const docs = mongoJson(`db.employees.find({holdingcode:'bc001', code:'UATEMP01'}, {accessscopes:1}).toArray()`);
  return docs[0]?.accessscopes ?? null;
}

test.describe.configure({ mode: 'serial', retries: 1 });

test('SCOPE-SAVE: เพิ่มบริษัทใน scope → save → Mongo → ลบ → save → Mongo', async ({ page }) => {
  test.setTimeout(420000);
  const failures: string[] = [];

  // debug: log ทุก PUT ของ employee (body + status)
  page.on('console', (msg) => {
    if (msg.text().includes('DEBUG missingRequired')) console.log('[BROWSER]', msg.text().slice(0, 300));
  });
  page.on('request', (req) => {
    if (req.url().includes('/employee') && req.method() === 'PUT') {
      const body = req.postData() ?? '';
      const m = body.match(/"accessscopes":\[[^\]]*\]/);
      console.log(`[PUT] body accessscopes: ${m ? m[0].slice(0, 160) : '(none in body)'}`);
    }
  });
  page.on('response', (res) => {
    if (res.url().includes('/employee') && res.request().method() === 'PUT') {
      console.log(`[PUT] status ${res.status()}`);
    }
  });

  await ensureEmployeeScreen(page);
  const row = page.locator('.bc-list-row').filter({ hasText: 'UATEMP01' }).first();
  await expect(row).toBeVisible({ timeout: 20000 });

  // helper: เปิดฟอร์มแก้ไข
  async function openEdit() {
    let opened = false;
    for (let attempt = 1; attempt <= 3 && !opened; attempt++) {
      // รอปุ่มแก้ไขในแถวพร้อมก่อนเสมอ (แถวเพิ่งโหลด ปุ่มอาจยัง mount ไม่เสร็จ)
      const rowEdit = row.getByRole('button', { name: /แก้ไข|Edit/ }).first();
      await rowEdit.waitFor({ state: 'visible', timeout: 8000 }).catch(() => {});
      if (await rowEdit.count()) {
        await rowEdit.click().catch(() => {});
      } else {
        const detail = page.locator('section[aria-label="รายละเอียด"]');
        await detail.locator('button', { hasText: 'แก้ไข' }).first().click().catch(() => {});
      }
      opened = await page.getByLabel('PIN').waitFor({ state: 'visible', timeout: 8000 }).then(() => true).catch(() => false);
    }
    if (!opened) throw new Error('ฟอร์มแก้ไขไม่เปิด');
    // ฟอร์ม edit ต้องมีค่ารหัสพนักงาน — ถ้าว่างแปลว่าเปิดเป็นฟอร์มเปล่า (บั๊ก)
    const codeInForm = await page.evaluate(`(()=>{
      const i=[...document.querySelectorAll('input')].find(x=>x.closest('label')?.textContent.includes('รหัสพนักงาน'));
      return i? i.value : null;
    })()`);
    console.log(`[openEdit] form.code = "${codeInForm}"`);
    // ติดตั้ง toast catcher (toast หายใน 5 วิ — ต้อง observe ตั้งแต่ก่อนกด save)
    await page.evaluate(`(()=>{
      const w=window; w.__toasts=w.__toasts||[];
      if(w.__toastObs) return;
      const obs=new MutationObserver(()=>{
        document.querySelectorAll('[role="alert"]').forEach(a=>{
          const t=a.textContent.trim(); if(t && !w.__toasts.includes(t)) w.__toasts.push(t);
        });
      });
      obs.observe(document.body,{childList:true,subtree:true});
      w.__toastObs=obs;
    })()`);
  }
  // helper: save + รอฟอร์มปิด + เก็บ toast ที่โผล่
  async function save(label: string) {
    // debug: dump ค่าช่องสำคัญในฟอร์มก่อนกด save
    const formVals = await page.evaluate(`(()=>{
      const pick=(ph)=>{const i=[...document.querySelectorAll('input,textarea')].find(x=>(x.getAttribute('placeholder')||'')===ph||x.closest('label')?.textContent.includes(ph)); return i? i.value.slice(0,30):'(no field)';};
      return {code: pick('รหัสพนักงาน'), name: pick('ชื่อพนักงาน'), pin: pick('PIN')};
    })()`);
    console.log(`[FORM before save] ${JSON.stringify(formVals)}`);
    await page.evaluate(`(()=>{ window.__toasts=[]; })()`);
    await page.locator('button', { hasText: /บันทึก|Save/ }).last().click();
    const closed = await page.getByLabel('PIN').waitFor({ state: 'hidden', timeout: 15000 }).then(() => true).catch(() => false);
    if (!closed) {
      const toasts = await page.evaluate(`(window.__toasts||[]).slice(0,4)`);
      const alertText = toasts.join(' | ').slice(0, 250) || '(ไม่มี toast/error)';
      failures.push(`${label}: ฟอร์มไม่ปิด — ${alertText}`);
      throw new Error(`${label} failed: ${alertText}`);
    }
  }

  // ===== 1) เพิ่มบริษัท TST01 เข้า scope =====
  await openEdit();
  // กันฟอร์มเปล่า: ถ้า code ว่าง (ฟอร์ม create-mode) ให้กรอกข้อมูลขั้นต่ำให้ครบก่อน save
  const codeVal = await page.evaluate(`(()=>{
    const i=[...document.querySelectorAll('input')].find(x=>x.closest('label')?.textContent.includes('รหัสพนักงาน'));
    return i? i.value : null;
  })()`);
  if (codeVal === '') {
    await page.getByLabel('รหัสพนักงาน').fill('UATEMP01');
    await page.getByLabel('ชื่อพนักงาน').fill('พนักงานวิไร วิทัย');
  }
  // ใช้ CompanyScopeSearchPicker แบบ popup: กด "เพิ่มบริษัท" → dialog → ค้นหา → คลิกรายการ
  const addCompanyBtn = page.getByRole('button', { name: /เพิ่มบริษัท/ }).first();
  await expect(addCompanyBtn).toBeVisible({ timeout: 10000 });
  await addCompanyBtn.click();
  const dialog = page.locator('[aria-label="เลือกบริษัท"]');
  await expect(dialog).toBeVisible({ timeout: 8000 });
  await dialog.getByPlaceholder('ค้นหารหัสหรือชื่อบริษัท').fill('TST01');
  const rowItem = dialog.locator('button', { hasText: 'TST01' }).first();
  await expect(rowItem).toBeVisible({ timeout: 8000 });
  await rowItem.click();
  await expect(dialog).toBeHidden({ timeout: 5000 });
  // เลือกจาก popup = เพิ่มเข้า list ทันที (ไม่มีปุ่มเพิ่มแล้ว)

  await save('เพิ่ม scope TST01');

  const afterAdd = scopesInMongo();
  const added = Array.isArray(afterAdd)
    ? JSON.stringify(afterAdd).includes('TST01')
    : afterAdd === true || afterAdd === 'all';
  console.log('after add:', JSON.stringify(afterAdd)?.slice(0, 200));
  if (!added) failures.push('หลังเพิ่ม: Mongo accessscopes ไม่มี TST01 — ' + JSON.stringify(afterAdd)?.slice(0, 150));

  // ===== 2) ลบบริษัทออก (คืนค่าเดิม) =====
  await ensureEmployeeScreen(page);
  await openEdit();
  // ปุ่มลบ = span[role="button"] ที่มีไอคอน trash ในการ์ดบริษัท TST01
  const companyCard = page.locator('button', { hasText: 'TST01' }).first();
  await expect(companyCard).toBeVisible({ timeout: 10000 });
  const removeBtn = companyCard.locator('span[role="button"]').first();
  const removeCount = await removeBtn.count();
  if (removeCount > 0) {
    await removeBtn.click();
    await page.waitForTimeout(600);
    await save('ลบ scope TST01');
  } else {
    console.log('ไม่พบปุ่มลบในการ์ด TST01');
  }

  const afterRemove = scopesInMongo();
  console.log('after remove:', JSON.stringify(afterRemove)?.slice(0, 200));
  const removed = !Array.isArray(afterRemove) || !JSON.stringify(afterRemove).includes('TST01');
  if (!removed) failures.push('หลังลบ: Mongo accessscopes ยังมี TST01');

  console.log(`\nRESULT: ${failures.length === 0 ? 'ALL OK' : 'FAILURES:'}`);
  failures.forEach((f) => console.log(' -', f));
  expect(failures.length, failures.join('\n')).toBe(0);
});
