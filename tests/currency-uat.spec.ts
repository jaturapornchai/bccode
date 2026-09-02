import { execSync } from 'child_process';
import { expect, test, type Page } from '@playwright/test';

/**
 * Currency UAT (2026-08-30) — ตามกฎ "UAT ต้องตรวจ CRUD + MongoDB เสมอ":
 * every C/U/D is verified in the `currency` collection (and branch basecurrency
 * in `organizationbranches`) via mongosh. Guards under test come from the
 * product rule "อย่างน้อยต้องมี 1 สกุลเงิน": last-active delete/disable block
 * (UI + backend), base-currency delete/disable block (UI), auto-seed THB,
 * set-base from the currency screen, rate-vs-base history, quick-add presets.
 */

const SHOT = 'test-results/currency-uat';
const consoleErrors: string[] = [];

function mongoEval(js: string): string {
  return execSync(`docker exec mongodb mongosh --quiet appdb --eval "${js.replace(/"/g, '\\"')}"`, { timeout: 30000 })
    .toString().trim().split('\n').pop() ?? '';
}
function mongoCount(coll: string, query: string): number {
  return Number(mongoEval(`print(db.${coll}.countDocuments(${query}))`));
}

async function shoot(page: Page, name: string) {
  await page.screenshot({ path: `${SHOT}/${name}.png` });
}

async function fullLogin(page: Page) {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto('/');
  await page.waitForTimeout(900);
  await page.getByRole('button', { name: /Dev Login/ }).click();
  await page.waitForURL(/\/holding/, { timeout: 30000 });
  await page.waitForTimeout(1500);
  await page.getByRole('button', { name: /บ้านเชียง/ }).first().click();
  await page.waitForURL(/\/workspace/, { timeout: 30000 });
  await page.waitForTimeout(1200);
  // pick TST03 so the base-currency branch is deterministic
  await page.getByRole('button', { name: /สาขาทดสอบไทย/ }).first().click();
  await page.waitForTimeout(1800);
  await page.getByRole('button', { name: /สำนักงานใหญ่/ }).first().click();
  await page.waitForTimeout(3000);
}

async function openCurrencyScreen(page: Page) {
  await page.goto('http://127.0.0.1:3000/menu');
  await page.waitForTimeout(2200);
  await page.locator('aside input').first().fill('สกุลเงิน');
  await page.waitForTimeout(700);
  await page.locator('aside [role="treeitem"] button', { hasText: 'สกุลเงิน' }).first().click();
  await page.waitForTimeout(3000);
}

/** the currency CARD containing the h2 with exactly this code — NOT the
    search/quick-add card which also contains the code text, and robust to the
    card's radius class (rounded-lg, not rounded-xl). */
function currencyCard(page: Page, code: string) {
  return page
    .locator('[role="tabpanel"]:not([hidden])')
    .locator(`xpath=.//h2[normalize-space()=${JSON.stringify(code)}]/ancestor::div[contains(@class,"rounded")][last()]`);
}

/** blocked delete buttons carry the GUARD message as their aria-label, not "ลบ" */
const DELETE_BUTTON = /ลบ|Delete|สกุลเงินหลักของสาขา|อย่างน้อย/;

test.describe.configure({ mode: 'serial', retries: 1 });

// refresh-token rotation เป็น single-use: ทุก test ต้องบันทึก cookie ที่ rotate
// แล้วทิ้งไว้ให้ test ถัดไป ไม่งั้นตัวหลัง reuse cookie เก่า → /refresh 401 →
// เด้งหน้า login กลางชุด (เคสจริง battery 2026-08-31)
test.afterEach(async ({ page }) => {
  await page.context().storageState({ path: '.auth/user.json' }).catch(() => {});
});

test('CU-01 auto-seed: empty currency list → THB created + base set (verified in Mongo)', async ({ page }) => {
  test.setTimeout(180000);
  // deterministic start: clear this UAT holding's currency docs (exact holding
  // scope, never a broad pattern) — desired end state is exactly THB anyway
  execSync("docker exec mongodb mongosh --quiet appdb --eval \"db.currency.deleteMany({holdingcode: 'bc001'})\"", { timeout: 30000 });
  await fullLogin(page);
  page.on('pageerror', (e) => consoleErrors.push(`currency: ${e.message}`));
  await openCurrencyScreen(page);
  await page.waitForTimeout(3500); // give auto-seed a moment
  await shoot(page, '01-after-autoseed');

  const thbActive = mongoCount('currency', '{"code": "THB", "isdisabled": false, "deletedat": {"$exists": false}}');
  expect(thbActive, 'THB must exist and be active in mongodb currency').toBeGreaterThanOrEqual(1);
  const baseSet = mongoCount('organizationbranches', '{"holdingcode": "bc001", "businesscode": "TST03", "basecurrency": "THB"}');
  expect(baseSet, 'current branch basecurrency must be THB in mongodb').toBeGreaterThanOrEqual(1);
});

test('CU-02 quick-add USD (one click) → persisted in Mongo', async ({ page }) => {
  test.setTimeout(180000);
  // pre-clean my own test datum (exact code — never a broad pattern)
  execSync(`docker exec mongodb mongosh --quiet appdb --eval "db.currency.deleteMany({code: 'USD'})"`, { timeout: 30000 });
  await fullLogin(page);
  await openCurrencyScreen(page);
  await page.waitForTimeout(2500);

  const quickBtn = page.locator('button', { hasText: /^USD$/ }).first();
  await expect(quickBtn, 'USD quick-add button visible').toBeVisible();
  await quickBtn.click();
  await page.waitForTimeout(2500);
  await shoot(page, '02-usd-added');

  const usdActive = mongoCount('currency', '{"code": "USD", "isdisabled": false, "deletedat": {"$exists": false}}');
  expect(usdActive, 'USD must be persisted active in mongodb currency').toBeGreaterThanOrEqual(1);
});

test('CU-03 set base currency from the currency screen (THB → USD → THB)', async ({ page }) => {
  test.setTimeout(180000);
  await fullLogin(page);
  await openCurrencyScreen(page);
  await page.waitForTimeout(2500);

  const usdCard = currencyCard(page, 'USD');
  const setBaseBtn = usdCard.getByRole('button', { name: /ตั้งเป็นสกุลเงินหลัก|Set as base/ });
  await expect(setBaseBtn, 'USD card has set-base button').toBeVisible();
  await setBaseBtn.click();
  await page.waitForTimeout(2500);
  await shoot(page, '03-base-usd');
  let base = mongoCount('organizationbranches', '{"holdingcode": "bc001", "businesscode": "TST03", "basecurrency": "USD"}');
  expect(base, 'branch basecurrency must be USD in mongodb').toBeGreaterThanOrEqual(1);

  const thbCard = currencyCard(page, 'THB');
  await thbCard.getByRole('button', { name: /ตั้งเป็นสกุลเงินหลัก|Set as base/ }).click();
  await page.waitForTimeout(2500);
  base = mongoCount('organizationbranches', '{"holdingcode": "bc001", "businesscode": "TST03", "basecurrency": "THB"}');
  expect(base, 'branch basecurrency back to THB in mongodb').toBeGreaterThanOrEqual(1);
});

test('CU-04 update: rate vs base persisted into exchangerates history', async ({ page }) => {
  test.setTimeout(180000);
  await fullLogin(page);
  await openCurrencyScreen(page);
  await page.waitForTimeout(2500);

  const usdCard = currencyCard(page, 'USD');
  await usdCard.getByRole('button', { name: /แก้ไข|Edit/ }).click();
  await page.waitForTimeout(1000);
  const rateField = page.locator('input[type="number"]');
  await expect(rateField, 'rate field visible in edit form').toBeVisible();
  await rateField.fill('35.5');
  await page.locator('button', { hasText: /บันทึก|Save/ }).last().click();
  await page.waitForTimeout(3000);
  await shoot(page, '04-rate-saved');

  const rateSaved = Number(mongoEval(
    'print((db.currency.findOne({"code": "USD", "deletedat": {"$exists": false}})?.exchangerates ?? []).filter(e => e.rate === 35.5).length)',
  ));
  expect(rateSaved, 'rate 35.5 must be in mongodb exchangerates history').toBeGreaterThanOrEqual(1);
  const cardText = await page.locator('[role="tabpanel"]:not([hidden])').textContent();
  expect(cardText, 'card must display the rate vs base').toContain('35.5');
});

test('CU-05 guards: base cannot be deleted; last-active cannot be deleted (UI + backend bypass)', async ({ page }) => {
  test.setTimeout(180000);
  await fullLogin(page);
  await openCurrencyScreen(page);
  await page.waitForTimeout(2500);

  // ensure USD exists (self-healing across attempts/retries)
  const usdCount = mongoCount('currency', '{"code": "USD", "deletedat": {"$exists": false}}');
  if (usdCount === 0) {
    await page.evaluate(async () => {
      const auth = JSON.parse(localStorage.getItem('bc_auth') ?? '{}');
    const refreshed = await fetch('/api/auth/refresh', { method: 'POST', credentials: 'same-origin' }).then((r) => r.json()).catch(() => null);
    const token = refreshed?.token || auth.token;
      await fetch('/api/currency', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'x-bc-backend-url': auth.backendUrl, Authorization: `Bearer ${token}` },
        body: JSON.stringify({ backendUrl: auth.backendUrl, guidfixed: '', code: 'USD', name: 'US Dollar', symbol: '$', isdisabled: false }),
      });
    });
    await page.reload();
    await page.waitForTimeout(3000);
  }

  // THB is base → delete button disabled with the Thai guard reason
  const thbCard = currencyCard(page, 'THB');
  const thbDelete = thbCard.getByRole('button', { name: DELETE_BUTTON });
  await expect(thbDelete).toBeDisabled();
  await expect(thbDelete).toHaveAttribute('title', /สกุลเงินหลัก|base/i);
  await shoot(page, '05-base-delete-blocked');

  // delete USD via UI (allowed) → confirm dialog
  const usdCard = currencyCard(page, 'USD');
  await usdCard.getByRole('button', { name: DELETE_BUTTON }).click();
  await page.waitForTimeout(800);
  const confirmBtn = page.locator('button:visible').filter({ hasText: /^(ลบ|ใช่|ยืนยัน|ตกลง|Delete|Yes)$/ }).last();
  await confirmBtn.click();
  await page.waitForTimeout(3000);
  const usdGone = mongoCount('currency', '{"code": "USD", "deletedat": {"$exists": false}}');
  expect(usdGone, 'USD deleted (soft) in mongodb').toBe(0);
  await shoot(page, '05-usd-deleted');

  // now THB is the ONLY active → UI delete disabled with the last-active reason
  await expect(thbDelete).toBeDisabled();
  await expect(thbDelete).toHaveAttribute('title', /อย่างน้อย|สกุล/i);

  // BACKEND BYPASS: direct API delete of THB must be rejected by the Go guard
  const bypass = await page.evaluate(async () => {
    const auth = JSON.parse(localStorage.getItem('bc_auth') ?? '{}');
    const refreshed = await fetch('/api/auth/refresh', { method: 'POST', credentials: 'same-origin' }).then((r) => r.json()).catch(() => null);
    const token = refreshed?.token || auth.token;
    const list = await fetch('/api/currency?page=1&limit=1000', {
      headers: { 'x-bc-backend-url': auth.backendUrl, Authorization: `Bearer ${token}` },
      cache: 'no-store',
    }).then((r) => r.json());
    const thb = (list.data ?? []).find((c: { code: string }) => c.code === 'THB');
    if (!thb) return { status: 0, body: 'THB not found in list' };
    const res = await fetch(`/api/currency/${encodeURIComponent(thb.guidfixed)}`, {
      method: 'DELETE',
      headers: { 'x-bc-backend-url': auth.backendUrl, Authorization: `Bearer ${token}` },
    });
    return { status: res.status, body: await res.text() };
  });
  expect(bypass.status, 'API bypass must be rejected (not 200)').not.toBe(200);
  expect(String(bypass.body)).toMatch(/อย่างน้อย|1 สกุล/i);
  const thbStillActive = mongoCount('currency', '{"code": "THB", "isdisabled": false, "deletedat": {"$exists": false}}');
  expect(thbStillActive, 'THB must survive the bypass attempt in mongodb').toBeGreaterThanOrEqual(1);
  await shoot(page, '05-bypass-rejected');
});

test('CU-06 guard: cannot disable the last active currency via API', async ({ page }) => {
  test.setTimeout(180000);
  await fullLogin(page);
  await openCurrencyScreen(page);
  await page.waitForTimeout(2500);

  // NOTE: the form has NO enable/disable control by design (status is read-only in UI)
  // — the disable path exists only at the API level, so this test exercises that.
  const bypass = await page.evaluate(async () => {
    const auth = JSON.parse(localStorage.getItem('bc_auth') ?? '{}');
    const refreshed = await fetch('/api/auth/refresh', { method: 'POST', credentials: 'same-origin' }).then((r) => r.json()).catch(() => null);
    const token = refreshed?.token || auth.token;
    const list = await fetch('/api/currency?page=1&limit=1000', {
      headers: { 'x-bc-backend-url': auth.backendUrl, Authorization: `Bearer ${token}` },
      cache: 'no-store',
    }).then((r) => r.json());
    const thb = (list.data ?? []).find((c: { code: string }) => c.code === 'THB');
    if (!thb) return { status: 0, body: 'not found' };
    const res = await fetch(`/api/currency/${encodeURIComponent(thb.guidfixed)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', 'x-bc-backend-url': auth.backendUrl, Authorization: `Bearer ${token}` },
      body: JSON.stringify({ backendUrl: auth.backendUrl, guidfixed: thb.guidfixed, code: 'THB', name: thb.name, symbol: thb.symbol, isdisabled: true }),
    });
    return { status: res.status, body: await res.text() };
  });
  expect(bypass.status, 'API bypass disable must be rejected (not 200)').not.toBe(200);
  expect(String(bypass.body)).toMatch(/อย่างน้อย|1 สกุล/i);
  const thbStillEnabled = mongoCount('currency', '{"code": "THB", "isdisabled": false, "deletedat": {"$exists": false}}');
  expect(thbStillEnabled, 'THB must still be active in mongodb after bypass').toBeGreaterThanOrEqual(1);
  await shoot(page, '06-disable-blocked-api');
});

test('CU-99 final state + console cleanliness', async () => {
  const active = mongoCount('currency', '{"isdisabled": false, "deletedat": {"$exists": false}}');
  expect(active, 'final state: at least one active currency always').toBeGreaterThanOrEqual(1);
  const base = mongoCount('organizationbranches', '{"holdingcode": "bc001", "businesscode": "TST03", "basecurrency": "THB"}');
  expect(base, 'final state: branch base = THB').toBeGreaterThanOrEqual(1);
  console.log('[currency-uat] console errors:', consoleErrors.length);
  expect(consoleErrors, 'no uncaught page errors').toEqual([]);
});
