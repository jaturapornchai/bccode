import { execSync } from 'child_process';
import { expect, test, type Page } from '@playwright/test';

/**
 * Employee UAT (2026-08-30) — "เพิ่มพนักงาน 10 คน UAT CRUD ด้วย"
 * Screen: /employee (generic main-crud, API /api/system-settings/employee,
 * Mongo collection `employees` under holdingcode bc001).
 * Rules applied: seeded-random data, MongoDB verified AFTER EVERY step
 * (create→check, update→check, delete→check), side-effect check (other rows
 * survive), cleanup only by exact ids. The 10 seeded employees STAY in the
 * system (user asked to add them); one sacrificial temp row proves Delete.
 *
 * SPEED (2026-08-30): uses the project storageState (login once in
 * auth.setup.ts) — no per-test UI login — and event-based waits
 * (expect(...).toBeVisible auto-retries) instead of fixed sleeps.
 */

const SHOT = 'test-results/employee-uat';
const consoleErrors: string[] = [];

function mongoEval(js: string): string {
  const cmd = `docker exec mongodb mongosh --quiet appdb --eval "${js.replace(/"/g, '\\"')}"`;
  // docker exec can blip transiently (seen once: E-01 verify step) — retry once
  let lastErr: unknown;
  for (let attempt = 1; attempt <= 2; attempt++) {
    try {
      return execSync(cmd, { timeout: 45000 }).toString().trim().split('\n').pop() ?? '';
    } catch (err) {
      lastErr = err;
      if (attempt < 2) Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 1500); // sync sleep
    }
  }
  throw lastErr;
}
function mongoCount(coll: string, query: string): number {
  return Number(mongoEval(`print(db.${coll}.countDocuments(${query}))`));
}

async function shoot(page: Page, name: string) {
  await page.screenshot({ path: `${SHOT}/${name}.png` });
}

/** UI login fallback — session can die mid-run (refresh rotation); self-heal */
async function uiLogin(page: Page) {
  await page.goto('/');
  await page.getByRole('button', { name: /Dev Login/ }).click();
  await page.waitForURL(/\/holding/, { timeout: 30000 });
  await page.getByRole('button', { name: /บ้านเชียง/ }).first().click();
  await page.waitForURL(/\/workspace/, { timeout: 30000 });
  await page.getByRole('button', { name: /สาขาทดสอบไทย|TST03/ }).first().click();
  await page.getByRole('button', { name: /สำนักงานใหญ่|00001/ }).first().click();
  await page.waitForTimeout(1500);
}

/** storageState holds the session (fast path). If the session died mid-run
    (login page detected) re-login via UI, then land back on /employee. */
async function openEmployeeScreen(page: Page) {
  await page.goto('/employee');
  const search = page.locator('input[placeholder*="ค้นหา พนักงาน"]');
  const devBtn = page.getByRole('button', { name: /Dev Login/ });
  // wait for EITHER state to settle (redirect may lag), then branch
  await expect(search.or(devBtn).first()).toBeVisible({ timeout: 15000 });
  if (await devBtn.count()) {
    await uiLogin(page);
    await page.goto('/employee');
  }
  await expect(search).toBeVisible({ timeout: 15000 });
  return search;
}

/** guard for mid-loop use: re-auth + return to /employee if the session died */
async function ensureEmployeeScreen(page: Page) {
  const search = page.locator('input[placeholder*="ค้นหา พนักงาน"]');
  const devBtn = page.getByRole('button', { name: /Dev Login/ });
  const settled = await expect(search.or(devBtn).first()).toBeVisible({ timeout: 15000 }).then(() => true).catch(() => false);
  const needLogin = !settled || (await devBtn.count()) > 0;
  if (needLogin) {
    await openEmployeeScreen(page);
    // Persist the freshly rotated cookie so the next test/context starts here,
    // not from a revoked snapshot.
    await page.context().storageState({ path: '.auth/user.json' }).catch(() => {});
  }
}

// seeded PRNG + Thai name generator (same pattern as uat-crud)
const SEED = Number(process.env.EMP_SEED ?? Date.now() % 1e9);
function mulberry32(a: number) {
  return function () {
    a |= 0; a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}
const rnd = mulberry32(SEED);
const pick = (arr: readonly string[]): string => arr[Math.floor(rnd() * arr.length)];
const thaiWord = () => pick(['สม', 'วิ', 'บุญ', 'จิ', 'อุ', 'สา', 'มา', 'ชู', 'พร', 'ทิพ', 'ไพ', 'รัก']) + pick(['ชาย', 'ชัย', 'เรือง', 'รา', 'ไร', 'นะ', 'พงษ์', 'ดา', 'ทัย', 'ษ์']);
const thaiName = () => Array.from({ length: 2 + Math.floor(rnd() * 2) }, thaiWord).join('');

/** 10 employees that STAY in the system (user request: เพิ่มพนักงาน 10 คน) */
const EMPLOYEES = Array.from({ length: 10 }, (_, i) => ({
  code: `UATEMP${String(i + 1).padStart(2, '0')}`,
  name: `พนักงาน${thaiWord()} ${thaiWord()}`,
  email: i % 3 === 0 ? `uatemp${i + 1}@uat.test` : '',
  pin: String(100000 + Math.floor(rnd() * 899999)),
}));

/** row for one employee — self-heals the screen first, then waits patiently */
async function listRow(page: Page, code: string) {
  await ensureEmployeeScreen(page);
  const row = page.locator('table tbody tr, .bc-list-row').filter({ hasText: code }).first();
  await expect(row).toBeVisible({ timeout: 15000 });
  return row;
}

/** create one employee via the form; form closes itself on success */
async function createEmployee(page: Page, code: string, name: string, email: string, pin: string, wholeHolding: boolean) {
  const codeInput = page.getByLabel('รหัสพนักงาน');
  if (await codeInput.isVisible().catch(() => false)) {
    await page.locator('button', { hasText: /ยกเลิก|Cancel/ }).first().click();
    await expect(codeInput).toBeHidden();
  }
  await page.getByRole('button', { name: 'เพิ่ม', exact: true }).first().click();
  await expect(codeInput).toBeVisible();
  await codeInput.fill(code);
  await page.getByLabel('ชื่อพนักงาน').fill(name);
  if (email) await page.getByLabel('อีเมล').fill(email);
  if (pin) await page.getByLabel('PIN').fill(pin);
  if (wholeHolding) {
    const scopeChk = page.locator('label', { hasText: 'ทั้งกลุ่มกิจการ' }).locator('input[type="checkbox"]').first();
    if (await scopeChk.count()) await scopeChk.check().catch(() => {});
  }
  await page.locator('button', { hasText: /บันทึก|Save/ }).last().click();
  // event wait: success = form closes (auto-retries up to 5s)
  await expect(codeInput).toBeHidden({ timeout: 10000 });
}

test.describe.configure({ mode: 'serial', retries: 1 });

// keep the refresh-token chain continuous: each test rotates the cookie
// (single-use rotation) — persist the rotated cookie for the NEXT test,
// otherwise the next context starts with a revoked snapshot and logs out
test.afterEach(async ({ page }) => {
  await page.context().storageState({ path: '.auth/user.json' }).catch(() => {});
});

test('E-01 CREATE ×10 employees — Mongo verified after EACH create', async ({ page }) => {
  test.setTimeout(420000);
  await openEmployeeScreen(page);
  page.on('pageerror', (e) => consoleErrors.push(`employee: ${e.message}`));

  // pre-clean my own exact codes from any previous attempt (scoped: exact prefix + holding)
  execSync('docker exec mongodb mongosh --quiet appdb --eval "db.employees.deleteMany({holdingcode: \'bc001\', code: {$regex: \'^UATEMP[0-9]{2}$\'}})"', { timeout: 30000 });

  for (const emp of EMPLOYEES) {
    await ensureEmployeeScreen(page);
    await createEmployee(page, emp.code, emp.name, emp.email, emp.pin, true);
    // step-by-step Mongo verification — immediately after this create
    const n = mongoCount(
      'employees',
      `{"holdingcode": "bc001", "code": "${emp.code}", "name": "${emp.name}", "isenabled": true}`,
    );
    expect(n, `${emp.code} must be persisted with its name right after creation`).toBeGreaterThanOrEqual(1);
  }
  await shoot(page, '01-created-10');

  const total = mongoCount('employees', '{"holdingcode": "bc001", "code": {"$regex": "^UATEMP[0-9]{2}$"}}');
  expect(total, 'exactly 10 UAT employees persisted').toBe(10);
});

test('E-02 READ: search filters and finds every created employee', async ({ page }) => {
  test.setTimeout(300000);
  await openEmployeeScreen(page);
  await ensureEmployeeScreen(page);

  const searchInput = () => page.locator('input[placeholder*="ค้นหา พนักงาน"]');
  // the app can kill the session mid-action (refresh rotation race, see report
  // defect R1) — retry the fill once through the self-healing screen.
  // Use a fresh locator after recovery because the old node was detached.
  await searchInput().fill('UATEMP', { timeout: 10000 }).catch(async () => {
    await ensureEmployeeScreen(page);
    await searchInput().fill('UATEMP');
  });
  const rows = page.locator('table tbody tr, .bc-list-row').filter({ hasText: 'UATEMP' });
  await expect(rows.first(), 'search results load (event wait)').toBeVisible({ timeout: 15000 });
  const rowTexts = (await rows.allTextContents()).join(' ');
  for (const emp of EMPLOYEES) {
    expect(rowTexts, "list should contain " + emp.code).toContain(emp.code);
  }
  await shoot(page, '02-search-all');

  // narrow search → exactly one row (event wait on the narrowed result)
  await searchInput().fill(EMPLOYEES[4].code);
  const one = page.locator('table tbody tr, .bc-list-row').filter({ hasText: EMPLOYEES[4].code });
  await expect(one.first()).toBeVisible({ timeout: 15000 });
  const count = await one.count();
  expect(count, 'search narrows to 1 row').toBe(1);
  await shoot(page, '02-search-one');
  // clearing the query restores the FULL list (all 10 UAT rows visible again)
  await searchInput().fill('');
  await expect(rows.first()).toBeVisible({ timeout: 15000 });
  await expect(rows, 'clearing search restores all 10 rows').toHaveCount(10, { timeout: 15000 });
});

test('E-03 UPDATE: rename employee #05 → verified in Mongo (same doc, not orphan)', async ({ page }) => {
  test.setTimeout(300000);
  await openEmployeeScreen(page);

  const target = EMPLOYEES[4];
  const row = await listRow(page, target.code);
  await row.getByRole('button', { name: /แก้ไข|Edit/ }).first().click();
  const nameInput = page.getByLabel('ชื่อพนักงาน');
  await expect(nameInput).toBeVisible();
  const newName = `${target.name} แก้ไข`;
  await nameInput.fill(newName);
  await page.locator('button', { hasText: /บันทึก|Save/ }).last().click();
  await expect(nameInput).toBeHidden({ timeout: 10000 });
  await shoot(page, '03-updated');

  // step-by-step Mongo verification: the SAME doc (same guid) carries the new name
  const guid = mongoEval(`print(db.employees.findOne({holdingcode: "bc001", code: "${target.code}"}).guidfixed)`);
  const updated = mongoCount('employees', `{"holdingcode": "bc001", "guidfixed": "${guid}", "name": "${newName}"}`);
  expect(updated, `rename must land on the SAME doc (${target.code}) in mongodb`).toBeGreaterThanOrEqual(1);
  const oldNameLeft = mongoCount('employees', `{"holdingcode": "bc001", "code": "${target.code}", "name": "${target.name}"}`);
  expect(oldNameLeft, 'old name must not remain on any doc').toBe(0);

  // side-effect check: the other 9 survive untouched
  for (const emp of EMPLOYEES.filter((_, i) => i !== 4)) {
    const n = mongoCount('employees', `{"holdingcode": "bc001", "code": "${emp.code}", "name": "${emp.name}"}`);
    expect(n, `${emp.code} must be untouched by the rename`).toBeGreaterThanOrEqual(1);
  }
});

test('E-04 DELETE: temp employee removed via UI → gone in Mongo; the 10 remain', async ({ page }) => {
  test.setTimeout(300000);
  await openEmployeeScreen(page);

  // sacrificial row (exact code, cleaned by id)
  const tempCode = 'UATETMP99';
  execSync(`docker exec mongodb mongosh --quiet appdb --eval "db.employees.deleteMany({code: '${tempCode}'})"`, { timeout: 30000 });
  await createEmployee(page, tempCode, 'พนักงานชั่วคราว ลบทดสอบ', '', '', true);
  let n = mongoCount('employees', `{"holdingcode": "bc001", "code": "${tempCode}"}`);
  expect(n, 'temp employee created').toBeGreaterThanOrEqual(1);

  const row = await listRow(page, tempCode);
  await row.getByRole('button', { name: /ลบ|Delete/ }).first().click();
  const confirmBtn = page.locator('button:visible').filter({ hasText: /^(ลบ|ใช่|ยืนยัน|ตกลง|Delete|Yes)$/ }).last();
  await expect(confirmBtn).toBeVisible();
  await confirmBtn.click();
  await expect(row).toBeHidden({ timeout: 10000 });
  await shoot(page, '04-deleted');

  // The app uses soft-delete (deletedat); verify the temp row is no longer active.
  n = mongoCount('employees', `{"holdingcode": "bc001", "code": "${tempCode}", "deletedat": {"$exists": false}}`);
  expect(n, 'temp employee must be gone from active list in mongodb').toBe(0);

  // side-effect: the 10 stay
  const remain = mongoCount('employees', '{"holdingcode": "bc001", "code": {"$regex": "^UATEMP[0-9]{2}$"}, "deletedat": {"$exists": false}}');
  expect(remain, 'the 10 UAT employees must remain').toBe(10);
});

test('E-05 EDGE: duplicate code rejected + empty required fields blocked', async ({ page }) => {
  test.setTimeout(300000);
  await openEmployeeScreen(page);

  // duplicate code
  const dup = EMPLOYEES[0];
  await createEmployee(page, dup.code, 'ชื่อซ้ำ ทดสอบ', '', '', false).catch(() => {});
  // duplicate → the form STAYS open with an error toast; verify then cancel
  const bodyText = (await page.locator('body').textContent()) ?? '';
  expect(bodyText, 'duplicate code must be rejected with an error').toMatch(/ซ้ำ|already|duplicate/i);
  const dupCount = mongoCount('employees', `{"holdingcode": "bc001", "code": "${dup.code}"}`);
  expect(dupCount, 'no second doc for the duplicate code').toBe(1);
  await shoot(page, '05-duplicate-rejected');
  await page.locator('button', { hasText: /ยกเลิก|Cancel/ }).first().click();
  await expect(page.getByLabel('รหัสพนักงาน')).toBeHidden({ timeout: 10000 });

  // empty required fields
  await page.getByRole('button', { name: 'เพิ่ม', exact: true }).first().click();
  const codeInput = page.getByLabel('รหัสพนักงาน');
  await expect(codeInput).toBeVisible();
  await page.locator('button', { hasText: /บันทึก|Save/ }).last().click();
  const errText = (await page.locator('body').textContent()) ?? '';
  expect(errText, 'empty form must be blocked with a Thai error').toMatch(/กรุณา|required/i);
  await shoot(page, '06-empty-blocked');
  await page.locator('button', { hasText: /ยกเลิก|Cancel/ }).first().click();
  await expect(codeInput).toBeHidden({ timeout: 10000 });
});

test('E-99 final state: exactly 10 UAT employees in Mongo + console clean', async () => {
  const total = mongoCount('employees', '{"holdingcode": "bc001", "code": {"$regex": "^UATEMP[0-9]{2}$"}, "deletedat": {"$exists": false}}');
  expect(total, '10 UAT employees remain (user request: เพิ่มพนักงาน 10 คน)').toBe(10);
  const temp = mongoCount('employees', '{"code": "UATETMP99", "deletedat": {"$exists": false}}');
  expect(temp, 'sacrificial temp row cleaned from active list').toBe(0);
  console.log('[employee-uat] console errors:', consoleErrors.length);
  expect(consoleErrors, 'no uncaught page errors').toEqual([]);
});
