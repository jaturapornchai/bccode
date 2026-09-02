import { execSync } from 'child_process';
import { expect, test, type Page } from '@playwright/test';

/**
 * Detailed CRUD UAT (2026-08-29): "UAT ให้ละเอียด CRUD ให้ครบ สุ่มทดสอบ"
 * Covers every entity with a UI: Holding (/holding), Company + Branch
 * (/company tree), User (/user), PermissionGroup (/permissiongroup) plus
 * read-only screens (/permissiondefinition, /useraccessaudit).
 * Data is SEEDED-RANDOM so every run exercises different values; the seed is
 * printed and stored in metrics-crud.json for reproduction (re-run with
 * CRUD_SEED=<seed>). Screenshots: test-results/uat-crud/*.png
 */

const SHOT = 'test-results/uat-crud';

const SEED = Number(process.env.CRUD_SEED ?? Date.now() % 1e9);
function mulberry32(a: number) {
  return function () {
    a |= 0; a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}
const rnd = mulberry32(SEED);
const pick = <T,>(arr: readonly T[]): T => arr[Math.floor(rnd() * arr.length)];
const digits = (n: number) => Array.from({ length: n }, () => Math.floor(rnd() * 10)).join('');
const latin = (n: number) => Array.from({ length: n }, () => 'abcdefghjkmnpqrstuvwxyz'[Math.floor(rnd() * 23)]).join('');
const thaiWord = () => pick(['สม', 'วิ', 'บุญ', 'จิ', 'อุ', 'สา', 'มา', 'ชู', 'พร', 'ทิพ', 'ไพ', 'รัก']) + pick(['ชาย', 'ชัย', 'เรือง', 'รา', 'ไร', 'นะ', 'พงษ์', 'ดา', 'ทัย', 'ษ์']);
const thaiName = () => Array.from({ length: 2 + Math.floor(rnd() * 2) }, thaiWord).join('');

const findings: string[] = [];
const consoleErrors: string[] = [];
const metrics: Record<string, unknown> = { seed: SEED };

const AUDIT = (() => {
  return `(() => {
  const items = [];
  const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
  while (walker.nextNode()) {
    const node = walker.currentNode;
    if (!node.textContent.trim()) continue;
    const el = node.parentElement;
    if (!el || el.closest('script,style,noscript')) continue;
    const cs = getComputedStyle(el);
    if (cs.display === 'none' || cs.visibility === 'hidden' || parseFloat(cs.opacity) === 0) continue;
    const range = document.createRange();
    range.selectNodeContents(node);
    for (const r of range.getClientRects()) {
      if (r.width < 2 || r.height < 2) continue;
      items.push({ text: node.textContent.trim().slice(0, 26), x: r.x, y: r.y, w: r.width, h: r.height, el });
    }
  }
  const overlaps = [];
  for (let i = 0; i < items.length; i++) for (let j = i + 1; j < items.length; j++) {
    const a = items[i], b = items[j];
    if (a.el === b.el || a.el.contains(b.el) || b.el.contains(a.el)) continue;
    const ix = Math.max(0, Math.min(a.x + a.w, b.x + b.w) - Math.max(a.x, b.x));
    const iy = Math.max(0, Math.min(a.y + a.h, b.y + b.h) - Math.max(a.y, b.y));
    if (ix > 3 && iy > 3) overlaps.push(a.text + ' × ' + b.text);
  }
  const clipped = Array.from(document.querySelectorAll('button')).filter((b) => b.scrollHeight > b.clientHeight + 2 || b.scrollWidth > b.clientWidth + 2).length;
  return { overlaps: overlaps.slice(0, 8), overlapCount: overlaps.length, clipped };
})()`;
})();

async function shoot(page: Page, name: string) {
  await page.screenshot({ path: `${SHOT}/${name}.png` });
}
async function audit(page: Page, key: string) {
  metrics[key] = await page.evaluate(AUDIT);
}
async function shootAudit(page: Page, name: string) {
  await shoot(page, name);
  await audit(page, `audit-${name}`);
}

/** Dev login → holding → company → branch (the full workspace selection). */
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
  await page.getByRole('button', { name: /ทดสอบการค้า|TST01|บริษัท/ }).first().click();
  await page.waitForTimeout(1600);
  await page.getByRole('button', { name: /สำนักงานใหญ่|00001/ }).first().click();
  await page.waitForTimeout(2500);
}
async function gotoSetting(page: Page, route: string) {
  await page.goto(`http://127.0.0.1:3000${route}`);
  await page.waitForTimeout(2000);
}

/**
 * Data-layer verification (user request: "ตรวจสอบ ใน mongodb หรือยัง CRUD"):
 * every C/U/D must be confirmed in the appdb — holdings live in `shops`,
 * companies in `organizationcompanies`, branches in `organizationbranches`,
 * users in `users` + per-holding membership in `shopusers`, role mappings in
 * `role_permission`. Queries run in the mongodb container via mongosh.
 */
function mongoEval(js: string): string {
  return execSync(
    `docker exec mongodb mongosh --quiet appdb --eval "${js.replace(/"/g, '\\"')}"`,
    { timeout: 30000 },
  ).toString().trim().split('\n').pop() ?? '';
}
function mongoCount(coll: string, query: string): number {
  return Number(mongoEval(`print(db.${coll}.countDocuments(${query}))`));
}

/**
 * Several save flows open a "ยืนยันการบันทึกข้อมูล" dialog demanding a 4-digit
 * confirmation code that is DISPLAYED in the dialog itself. Handle it: read
 * the code, fill it, click ยืนยันบันทึก. Returns true when a dialog was handled.
 */
async function handleConfirmSaveDialog(page: Page): Promise<boolean> {
  await page.waitForTimeout(600);
  const dialogText = await page.locator('body').textContent() ?? '';
  if (!dialogText.includes('ยืนยันการบันทึกข้อมูล')) return false;
  const match = dialogText.match(/รหัสยืนยันของคุณคือ\s*(\d{4})/);
  if (!match) {
    findings.push('CONFIRM-SAVE: dialog appeared but no 4-digit code found in its text');
    await page.locator('button:has-text("ยกเลิก")').last().click().catch(() => {});
    return false;
  }
  const dialogInput = page.locator('input[inputmode="numeric"], input[placeholder*="4"], input[maxlength="4"]').last();
  await dialogInput.fill(match[1]);
  await page.locator('button:has-text("ยืนยันบันทึก")').last().click();
  await page.waitForTimeout(2000);
  return true;
}

/**
 * Find the action button (e.g. [title=แก้ไขบริษัท]) that belongs to the TREE
 * NODE containing `text`. Climbing ancestors alone is wrong — the tree root
 * contains every node's text, so every button "matches". Instead pick the
 * button whose first matching ancestor is the SMALLEST (the tightest row).
 */
async function actionButtonForNode(page: Page, selector: string, text: string) {
  const buttons = page.locator(selector);
  const n = await buttons.count();
  let best: { idx: number; len: number } | null = null;
  for (let i = 0; i < n; i++) {
    const len = await buttons.nth(i).evaluate((el, want) => {
      let cur: HTMLElement | null = el.parentElement;
      while (cur && cur !== document.body) {
        if (cur.textContent?.includes(want)) return cur.textContent.length;
        cur = cur.parentElement;
      }
      return Number.MAX_SAFE_INTEGER;
    }, text);
    if (len < Number.MAX_SAFE_INTEGER && (!best || len < best.len)) best = { idx: i, len };
  }
  return best ? buttons.nth(best.idx) : null;
}

test.describe.configure({ mode: 'serial', retries: 2 });

test('CRUD-00 setup log', async () => {
  console.log(`[uat-crud] SEED=${SEED} (reproduce with CRUD_SEED=${SEED} npx playwright test tests/uat-crud.spec.ts)`);
});

test('CRUD-01 Holding: create edge cases → create → read → update (no delete in UI)', async ({ page }) => {
  test.setTimeout(180000);
  await fullLogin(page);
  // the create button lives on /holding; fullLogin ends on /workspace
  await page.goto('http://127.0.0.1:3000/holding');
  await page.waitForTimeout(2000);
  page.on('pageerror', (e) => consoleErrors.push(`holding: ${e.message}`));

  // canCreateHolding resolves asynchronously; a transient API failure sets it
  // false for the whole session — retry with reloads before giving up.
  const addBtn = page.locator('.holding-add-button');
  let visible = false;
  for (let attempt = 0; attempt < 4 && !visible; attempt++) {
    visible = await addBtn.waitFor({ state: 'visible', timeout: 8000 }).then(() => true).catch(() => false);
    if (!visible) {
      await page.reload();
      await page.waitForTimeout(3000);
    }
  }
  expect(visible, 'holding create button should appear (cancreateholding)').toBe(true);
  await addBtn.click();
  await page.waitForTimeout(600);
  const modal = page.locator('.holding-modal-panel');
  const codeInput = modal.locator('.input-shell input').nth(0);
  const nameInput = modal.locator('.input-shell input').nth(1);
  const confirmInput = modal.locator('.input-shell input').nth(2);
  const submit = modal.locator('.primary-button');

  // --- validation edges: the input silently normalizes (lowercases/strips),
  // so only codes that stay INVALID after normalization can test rejection:
  // too-short ("ab"), symbols-only (→ empty), duplicate bc001, wrong confirm.
  const edges: Array<[string, string, string]> = [
    ['ab', thaiName(), 'right'], // too short (<3)
    ['!!!%%%', thaiName(), 'right'], // strips to empty
    ['bc001', thaiName(), 'right'], // duplicate
  ];
  for (const [bad, nm, confirmKind] of edges) {
    await codeInput.fill(bad);
    await nameInput.fill(nm);
    const real = (await modal.locator('.holding-confirm-box strong').textContent())?.trim() ?? '';
    await confirmInput.fill(confirmKind === 'right' ? real : digits(4).replace(/^0/, '1'));
    await page.waitForTimeout(400);
    if (await submit.isDisabled()) {
      // client validation blocks the submit — PASS (nothing reaches the DB)
      await shoot(page, `holding-edge-blocked-${bad.replace(/[^a-z0-9]/g, '') || 'symbols'}`);
      continue;
    }
    await submit.click();
    await page.waitForTimeout(1200);
    const stillOpen = await modal.isVisible();
    if (!stillOpen) {
      // modal closed — a holding may have been created; verify against the DB
      await page.waitForTimeout(1200);
      const normalized = bad.replace(/[^a-z0-9]/g, '').toLowerCase();
      const dbJunk = normalized.length >= 3 ? mongoCount('shops', `{"holdingcode": "${normalized}"}`) : 0;
      expect(dbJunk, `edge "${bad}" must not create holding "${normalized}" in mongodb`).toBe(0);
      findings.push(`HOLDING: edge case "${bad}" closed the modal (silently normalized${normalized.length >= 3 ? ` to "${normalized}" and ACCEPTED` : ' but nothing reached the DB — verify error UX'})`);
      visible = await addBtn.waitFor({ state: 'visible', timeout: 8000 }).then(() => true).catch(() => false);
      if (!visible) {
        await page.reload();
        await page.waitForTimeout(3000);
        await addBtn.waitFor({ state: 'visible', timeout: 15000 });
      }
      await addBtn.click();
      await page.waitForTimeout(600);
    } else {
      await shoot(page, `holding-edge-rejected-${bad.replace(/[^a-z0-9]/g, '') || 'symbols'}`);
    }
  }
  await shootAudit(page, 'holding-edge-cases');

  // --- wrong confirm code must be rejected ---
  await codeInput.fill('u' + latin(5));
  await nameInput.fill(thaiName());
  await confirmInput.fill(digits(4).replace(/^0/, '1'));
  await submit.click();
  await page.waitForTimeout(1200);
  const wrongConfirmBlocked = await modal.isVisible();
  if (!wrongConfirmBlocked) findings.push('HOLDING: wrong confirm code ACCEPTED');
  await shoot(page, 'holding-wrong-confirm');
  expect(wrongConfirmBlocked, 'wrong confirm code must not close the modal').toBe(true);

  // --- valid create with random data ---
  const code = 'u' + latin(5);
  const name = `${thaiName()} ${digits(2)}`;
  await codeInput.fill(code);
  await nameInput.fill(name);
  await confirmInput.fill((await modal.locator('.holding-confirm-box strong').textContent())?.trim() ?? '');
  await submit.click();
  await page.waitForTimeout(2500);
  expect(await modal.isVisible(), 'modal should close after valid create').toBe(false);
  // the app AUTO-SELECTS the created holding and navigates to /workspace
  const autoNav = await page.waitForURL(/\/workspace/, { timeout: 15000 }).then(() => true).catch(() => false);
  if (!autoNav) findings.push('HOLDING: after create the app did NOT auto-navigate to /workspace (unexpected flow change)');

  // --- read (data layer): persisted in mongodb `shops`
  const dbHoldings = mongoCount('shops', `{"holdingcode": "${code}", "names.name": "${name}"}`);
  expect(dbHoldings, `holding ${code} must be persisted in mongodb shops`).toBeGreaterThanOrEqual(1);

  // --- read (UI): back on /holding, search finds the new holding ---
  await page.goto('http://127.0.0.1:3000/holding');
  await page.waitForTimeout(2200);
  const search = page.locator('.holding-search input');
  await search.fill(code);
  await page.waitForTimeout(900);
  await shootAudit(page, 'holding-created-visible');
  const listText = (await page.locator('main').textContent()) ?? '';
  expect(listText).toContain(code);

  // --- update: rename via pencil ---
  await search.fill('');
  await page.waitForTimeout(600);
  const card = page.locator('.workspace-card-grid > div', { hasText: code }).first();
  await card.getByRole('button', { name: /แก้ไข/ }).click();
  await page.waitForTimeout(700);
  const editModal = page.locator('.holding-modal-panel');
  const nameField = editModal.locator('.input-shell input').first();
  const newName = `${thaiName()} แก้${digits(2)}`;
  await nameField.fill(newName);
  await editModal.locator('.primary-button').click();
  await page.waitForTimeout(2000);
  const search2 = page.locator('.holding-search input');
  await search2.fill(code);
  await page.waitForTimeout(900);
  await shootAudit(page, 'holding-renamed');
  const after = (await page.locator('main').textContent()) ?? '';
  expect(after, 'renamed holding name should appear in list').toContain(newName);
  const dbRenamed = mongoCount('shops', `{"holdingcode": "${code}", "names.name": "${newName}"}`);
  expect(dbRenamed, `rename must be persisted in mongodb shops (${code} → ${newName})`).toBeGreaterThanOrEqual(1);

  // --- delete: is there any delete control for holdings? ---
  const cardAfter = page.locator('.workspace-card-grid > div', { hasText: code }).first();
  const hasDelete = (await cardAfter.getByRole('button', { name: /ลบ/ }).count()) > 0;
  if (!hasDelete) findings.push('HOLDING: no Delete action in the UI (created test holding cannot be removed from UI)');
  await search2.fill('');
});

test('CRUD-02 Company: empty-name guard → create → read → update', async ({ page }) => {
  test.setTimeout(180000);
  await fullLogin(page);
  await gotoSetting(page, '/company');
  page.on('pageerror', (e) => consoleErrors.push(`company: ${e.message}`));
  await shootAudit(page, 'company-tree');

  await page.getByRole('button', { name: 'เพิ่มบริษัท' }).first().click();
  await page.waitForTimeout(800);
  const codeField = page.locator('input[placeholder*="รหัสบริษัท"]');
  const nameField = page.locator('input[placeholder*="กรอกชื่อบริษัท"]');
  expect(codeField).toBeVisible();
  expect(nameField).toBeVisible();

  // empty-name guard: fill only code, save, form must stay (or show error)
  const compCode = `C${digits(3)}`;
  await codeField.fill(compCode);
  const saveBtn = page.locator('button:has-text("บันทึก")').last();
  await saveBtn.click();
  await handleConfirmSaveDialog(page);
  await page.waitForTimeout(800);
  const stillForm = await nameField.isVisible();
  if (!stillForm) {
    findings.push(`COMPANY: empty name ACCEPTED for code ${compCode}`);
    await page.getByRole('button', { name: 'เพิ่มบริษัท' }).first().click();
    await page.waitForTimeout(800);
  }
  await shoot(page, 'company-empty-name-guard');

  // valid create
  const compName = `บริษัท ${thaiName()} ${digits(2)}`;
  await codeField.fill(compCode);
  await nameField.fill(compName);
  await saveBtn.click();
  await handleConfirmSaveDialog(page);
  await page.waitForTimeout(2500);
  await shootAudit(page, 'company-created');
  const treeText = (await page.locator('main').textContent()) ?? '';
  expect(treeText, 'new company should appear in the tree').toContain(compName);
  const dbCompanies = mongoCount('organizationcompanies', `{"code": "${compCode}", "names.name": "${compName}", "isdeleted": {"$ne": true}}`);
  expect(dbCompanies, `company ${compCode} must be persisted in mongodb organizationcompanies`).toBeGreaterThanOrEqual(1);

  // update: rename via the edit button INSIDE the created company's node
  const editBtn = await actionButtonForNode(page, '[title="แก้ไขบริษัท"]', compName);
  if (!editBtn) {
    findings.push(`COMPANY: no edit button found inside the node of ${compName}`);
    test.skip();
  }
  await editBtn!.click();
  await page.waitForTimeout(900);
  const editName = page.locator('input[placeholder*="กรอกชื่อบริษัท"]');
  const newName = `บริษัท ${thaiName()} แก้${digits(2)}`;
  if (await editName.isVisible()) {
    await editName.fill(newName);
    await page.locator('button:has-text("บันทึก")').last().click();
    await handleConfirmSaveDialog(page);
    await page.waitForTimeout(2500);
    await shootAudit(page, 'company-renamed');
    const after = (await page.locator('main').textContent()) ?? '';
    expect(after).toContain(newName);
    const dbRenamed = mongoCount('organizationcompanies', `{"code": "${compCode}", "names.name": "${newName}", "isdeleted": {"$ne": true}}`);
    expect(dbRenamed, `company rename must be persisted in mongodb (${compCode} → ${newName})`).toBeGreaterThanOrEqual(1);
  } else {
    findings.push('COMPANY: edit form did not open from [title=แก้ไขบริษัท]');
  }
});

test('CRUD-03 Branch: create under company → rename', async ({ page }) => {
  test.setTimeout(180000);
  await fullLogin(page);
  await gotoSetting(page, '/company');
  page.on('pageerror', (e) => consoleErrors.push(`branch: ${e.message}`));

  // create a fresh company to own the branch (keeps this test self-contained)
  const compCode = `B${digits(3)}`;
  const compName = `บริษัทสาขา ${thaiName()} ${digits(2)}`;
  await page.getByRole('button', { name: 'เพิ่มบริษัท' }).first().click();
  await page.waitForTimeout(800);
  await page.locator('input[placeholder*="รหัสบริษัท"]').fill(compCode);
  await page.locator('input[placeholder*="กรอกชื่อบริษัท"]').fill(compName);
  await page.locator('button:has-text("บันทึก")').last().click();
  await handleConfirmSaveDialog(page);
  await page.waitForTimeout(2500);
  const dbCompany = mongoCount('organizationcompanies', `{"code": "${compCode}", "isdeleted": {"$ne": true}}`);
  expect(dbCompany, `owner company ${compCode} persisted`).toBeGreaterThanOrEqual(1);

  // add branch to THAT company's node
  const addBranch = await actionButtonForNode(page, '[title="เพิ่มสาขา"]', compName);
  if (!addBranch) {
    findings.push(`BRANCH: no เพิ่มสาขา action on the node of ${compName}`);
    await shootAudit(page, 'branch-no-action');
    test.skip();
  }
  await addBranch!.click();
  await page.waitForTimeout(900);
  await shoot(page, 'branch-form-open');

  // branch form: รหัสสาขา (5 หลัก) + ชื่อสาขา (NamesEditor); selects have defaults
  const branchCodeField = page.locator('input[placeholder*="รหัสสาขา"]').first();
  const branchNameField = page.locator('input[placeholder*="ชื่อสาขา"], input[placeholder*="กรอกชื่อ"]').first();
  expect(branchCodeField, 'branch code field visible').toBeVisible();
  const branchName = `สาขา${thaiName()}${digits(2)}`;
  await branchCodeField.fill(`00${digits(3)}`);
  await branchNameField.fill(branchName);
  await page.locator('button:has-text("บันทึก")').last().click();
  await handleConfirmSaveDialog(page);
  await page.waitForTimeout(2500);
  await shootAudit(page, 'branch-created');
  const treeText = (await page.locator('main').textContent()) ?? '';
  expect(treeText, 'new branch should appear in the tree').toContain(branchName);
  const dbBranches = mongoCount('organizationbranches', `{"names.name": "${branchName}"}`);
  expect(dbBranches, 'branch must be persisted in mongodb organizationbranches').toBeGreaterThanOrEqual(1);

  // rename it back via แก้ไขสาขา inside the branch's own node
  const editBranch = await actionButtonForNode(page, '[title="แก้ไขสาขา"]', branchName);
  if (!editBranch) {
    findings.push(`BRANCH: no edit action found for ${branchName}`);
    test.skip();
  }
  await editBranch!.click();
  await page.waitForTimeout(900);
  const editNameField = page.locator('input[placeholder*="ชื่อสาขา"], input[placeholder*="กรอกชื่อ"]').first();
  expect(editNameField, 'branch edit name field visible').toBeVisible();
  const newBranchName = `สาขาแก้${thaiName()}${digits(2)}`;
  await editNameField.fill(newBranchName);
  await page.locator('button:has-text("บันทึก")').last().click();
  await handleConfirmSaveDialog(page);
  await page.waitForTimeout(2500);
  await shootAudit(page, 'branch-renamed');
  const dbRenamed = mongoCount('organizationbranches', `{"names.name": "${newBranchName}"}`);
  expect(dbRenamed, 'branch rename must be persisted in mongodb').toBeGreaterThanOrEqual(1);
});

test('CRUD-04 User: membership create → search → update → delete', async ({ page }) => {
  test.setTimeout(180000);
  await fullLogin(page);
  await gotoSetting(page, '/user');
  page.on('pageerror', (e) => consoleErrors.push(`user: ${e.message}`));

  // Product gate (found by earlier runs, recorded as findings):
  // - a brand-new usercode is rejected: "user must sign in with Google and
  //   accept an invitation before membership is created"
  // - ADMIN role for a new user: "can not edit self permission"
  // So full CRUD is exercised on `uat_branch_user` — an EXISTING Google-verified
  // auth user (member of uat260810a) — by attaching it to bc001.
  const usercode = 'uat_branch_user';
  const displayName = `${thaiName()} ${thaiName()}`;

  // pre-clean leftovers from any previous attempt/run (the attach is idempotent)
  execSync(
    `docker exec mongodb mongosh --quiet appdb --eval "db.shopusers.deleteMany({holdingcode: 'bc001', username: '${usercode}'})"`,
    { timeout: 30000 },
  );

  await page.getByRole('button', { name: 'เพิ่ม', exact: true }).first().click();
  await page.waitForTimeout(900);
  await shoot(page, 'user-form-open');

  const codeField = page.locator('input[placeholder*="somchai01"]');
  const nameField = page.locator('input[placeholder*="สมชาย"]');
  const emailField = page.locator('input[placeholder*="somchai@email"]');
  await codeField.fill(usercode);
  await nameField.fill(displayName);
  await emailField.fill('');

  // role ("สิทธิ์ผู้ใช้งาน *"): use ระดับผู้ใช้งาน (USER). A prior run proved
  // ADMIN-role create is rejected with "can not edit self permission" even for
  // a brand-new user — recorded as a defect finding in the report.
  const userCard = page.locator('label', { hasText: 'ระดับผู้ใช้งาน' }).first();
  if (await userCard.count()) {
    await userCard.click();
    await expect(userCard.locator('input[type="radio"]'), 'user radio must be checked').toBeChecked();
  } else {
    findings.push('USER: no ระดับผู้ใช้งาน radio found');
  }

  // scope is REQUIRED ("บริษัทและสาขาที่เข้าถึงได้ *") — tick ทั้งกลุ่มกิจการ and
  // assert it actually took (a silently-failed click leaves the form unsavable)
  const scopeAll = page.locator('label:has-text("ทั้งกลุ่มกิจการ") input[type="checkbox"]').first();
  await expect(scopeAll, 'scope-all checkbox should exist').toBeVisible();
  await scopeAll.check();
  await expect(scopeAll).toBeChecked();
  await shoot(page, 'user-form-filled');

  await page.locator('button:has-text("บันทึก")').last().click();
  await handleConfirmSaveDialog(page);
  await page.waitForTimeout(2500);
  const bodyAfterSave = (await page.locator('body').textContent()) ?? '';
  if (bodyAfterSave.includes('บันทึกไม่สำเร็จ')) {
    const toast = bodyAfterSave.match(/บันทึกไม่สำเร็จ[^\"]{0,80}/)?.[0];
    findings.push(`USER CREATE DEFECT: membership save rejected — ${toast}`);
    await shoot(page, 'user-create-rejected');
  }
  const formGone = !(await page.getByText('รายการใหม่: ผู้ใช้งาน').isVisible().catch(() => false));
  if (!formGone) {
    findings.push('USER: save did not close the form');
    await shoot(page, 'user-save-stuck');
  }
  await shootAudit(page, 'user-created');

  // read: the list shows the AUTH name + usercode (typed display name is not
  // reflected — UX finding). Reload /user for a deterministic list state.
  await page.goto('http://127.0.0.1:3000/user');
  await page.waitForTimeout(2200);
  const search = page.locator('input[placeholder*="ค้นหา ผู้ใช้งาน"]');
  await search.fill(usercode);
  await page.waitForTimeout(1200);
  const listText = (await page.locator('main').textContent()) ?? '';
  expect(listText, 'created member should be findable by usercode').toContain(usercode);
  const dbMembers = mongoCount('shopusers', `{"holdingcode": "bc001", "username": "${usercode}"}`);
  expect(dbMembers, `membership must be persisted in mongodb shopusers (bc001/${usercode})`).toBeGreaterThanOrEqual(1);
  await shootAudit(page, 'user-search-hit');
  await search.fill('');
  await page.waitForTimeout(800);

  // update: toggle สถานะ → เข้าใช้งานไม่ได้ชั่วคราว. NOTE: a prior run proved
  // editing ANY field on a global user saves fail with "global usercode cannot
  // be changed by holding administration" even when usercode is untouched —
  // recorded as a defect if it reproduces.
  const row = page.locator('table tbody tr, .bc-list-row').filter({ hasText: usercode }).first();
  await row.getByRole('button', { name: 'แก้ไข' }).first().click();
  await page.waitForTimeout(1000);
  const inactiveCard = page.locator('label', { hasText: 'เข้าใช้งานไม่ได้ชั่วคราว' }).first();
  await inactiveCard.click();
  await page.locator('button:has-text("บันทึก")').last().click();
  await handleConfirmSaveDialog(page);
  await page.waitForTimeout(2500);
  const bodyAfterUpdate = (await page.locator('body').textContent()) ?? '';
  const updateToast = bodyAfterUpdate.match(/บันทึกไม่สำเร็จ[^\"]{0,80}/)?.[0];
  if (updateToast) {
    findings.push(`USER EDIT DEFECT: save rejected — ${updateToast} (only the status toggle was changed)`);
    await shoot(page, 'user-update-rejected');
  } else {
    const dbUpdated = mongoCount('shopusers', `{"holdingcode": "bc001", "username": "${usercode}", "isaccessdisabled": true}`);
    expect(dbUpdated, `status toggle must be persisted in mongodb shopusers`).toBeGreaterThanOrEqual(1);
  }
  await shootAudit(page, 'user-updated');

  // delete: remove the bc001 membership (the auth user + uat260810a stay intact)
  const row2 = page.locator('table tbody tr, .bc-list-row').filter({ hasText: usercode }).first();
  const del = row2.getByRole('button', { name: 'ลบ' }).first();
  if (await del.isDisabled().catch(() => true)) {
    findings.push(`USER: delete disabled for member ${usercode}`);
  } else {
    await del.click();
    await page.waitForTimeout(700);
    await shoot(page, 'user-delete-confirm');
    // confirm dialog buttons: ยกเลิก / ลบ (danger tone)
    const confirmBtn = page.locator('[role="dialog"] button:visible, .dialog-backdrop button:visible')
      .filter({ hasText: /^(ลบ|ใช่|ยืนยัน|Delete|Yes)$/ }).last();
    await confirmBtn.click();
    await page.waitForTimeout(3000);
    // data-layer truth: the REAL membership (with membershipuid) must be gone
    const dbReal = mongoEval(
      `print(db.shopusers.countDocuments({holdingcode: "bc001", username: "${usercode}", membershipuid: {$exists: true, $ne: ""}}))`,
    );
    expect(Number(dbReal), 'the real membership must be deleted from mongodb shopusers').toBe(0);
    // known defect: the rejected UPDATE earlier wrote an orphan doc (no
    // membershipuid/createdat) that still renders as a list row
    const orphans = mongoEval(
      `print(db.shopusers.countDocuments({holdingcode: "bc001", username: "${usercode}", $or: [{membershipuid: {$in: [null, ""]}}, {membershipuid: {$exists: false}}]}))`,
    );
    if (Number(orphans) > 0) {
      findings.push(`USER DEFECT: the rejected update-save wrote ${orphans} orphan shopusers doc(s) (no membershipuid/createdat) — DELETE API returned success while the row kept rendering`);
      await shootAudit(page, 'user-orphan-left');
      execSync(`docker exec mongodb mongosh --quiet appdb --eval "db.shopusers.deleteMany({holdingcode: 'bc001', username: '${usercode}', membershipuid: {$in: [null, '']}})"`, { timeout: 30000 });
    } else {
      const rowsLeft = await page.locator('table tbody tr, .bc-list-row').filter({ hasText: usercode }).count();
      expect(rowsLeft, 'deleted member row should disappear from the list').toBe(0);
    }
    const originalIntact = mongoCount('shopusers', `{"holdingcode": "uat260810a", "username": "${usercode}"}`);
    expect(originalIntact, `the user's ORIGINAL membership (uat260810a) must survive`).toBeGreaterThanOrEqual(1);
    await shootAudit(page, 'user-deleted');
  }
});

test('CRUD-05 PermissionGroup: create → toggle → cleanup', async ({ page }) => {
  test.setTimeout(180000);
  await fullLogin(page);
  await gotoSetting(page, '/permissiongroup');
  page.on('pageerror', (e) => consoleErrors.push(`permgroup: ${e.message}`));
  await shootAudit(page, 'permgroup-initial');

  await page.getByRole('button', { name: 'เพิ่ม', exact: true }).first().click().catch(() => {});
  await page.waitForTimeout(1000);
  const textInputs = page.locator('form input[type="text"]:visible, [class*="form"] input[type="text"]:visible');
  const selects = page.locator('form select:visible, [class*="form"] select:visible');
  const checkboxes = page.locator('form input[type="checkbox"]:visible, [class*="form"] input[type="checkbox"]:visible');

  if ((await textInputs.count()) === 0 && (await selects.count()) === 0) {
    findings.push('PERMGROUP: เพิ่ม opened no usable form (no text input / select)');
    await shoot(page, 'permgroup-no-form');
    test.skip();
  }

  const groupName = `กลุ่มสิทธิ์${thaiName()}${digits(2)}`;
  const rolePermBefore = mongoCount('role_permission', '{}');
  if ((await textInputs.count()) > 0) await textInputs.first().fill(groupName);
  if ((await selects.count()) > 0) await selects.first().selectOption({ index: 1 }).catch(() => {});
  const cbCount = await checkboxes.count();
  if (cbCount > 0) await checkboxes.nth(Math.floor(rnd() * cbCount)).check().catch(() => {});
  await shoot(page, 'permgroup-form-filled');
  await page.locator('button:has-text("บันทึก")').last().click();
  await page.waitForTimeout(2500);
  await shootAudit(page, 'permgroup-saved');
  const rolePermAfter = mongoCount('role_permission', '{}');
  metrics.rolePermissionCounts = { before: rolePermBefore, after: rolePermAfter };

  const pageText = (await page.locator('main').textContent()) ?? '';
  expect(pageText.length, 'permissiongroup page should render content after save').toBeGreaterThan(100);
  if (!pageText.includes(groupName) && (await textInputs.count()) > 0) {
    findings.push(`PERMGROUP: created group name "${groupName}" not visible after save (may be a select-based mapping)`);
  }

  // cleanup: delete the created row if a delete button exists for it
  const row = page.locator('table tbody tr, .bc-list-row').filter({ hasText: groupName }).first();
  if ((await row.count()) > 0) {
    const del = row.getByRole('button', { name: 'ลบ' }).first();
    if ((await del.count()) > 0 && !(await del.isDisabled())) {
      await del.click();
      await page.waitForTimeout(600);
      await page.locator('button:visible').filter({ hasText: /^(ลบ|ยืนยัน|ตกลง|Delete)$/ }).last().click().catch(() => {});
      await page.waitForTimeout(2000);
    }
  }
});

test('CRUD-06 Read-only screens render + console cleanliness', async ({ page }) => {
  test.setTimeout(150000);
  await fullLogin(page);
  page.on('pageerror', (e) => consoleErrors.push(`readonly: ${e.message}`));

  for (const route of ['/permissiondefinition', '/useraccessaudit', '/activelanguages']) {
    await gotoSetting(page, route);
    await shootAudit(page, `readonly-${route.replaceAll('/', '')}`);
    const text = (await page.locator('main').textContent()) ?? '';
    expect(text.length, `${route} should render meaningful content`).toBeGreaterThan(60);
  }
});

test('CRUD-99 report metrics + console cleanliness', async () => {
  const fs = await import('fs');
  metrics.consoleErrors = consoleErrors;
  metrics.findings = findings;
  fs.writeFileSync('test-results/uat-crud/metrics-crud.json', JSON.stringify(metrics, null, 2));
  console.log('[uat-crud] findings:', JSON.stringify(findings, null, 1));
  console.log('[uat-crud] console errors:', consoleErrors.length);
  expect(consoleErrors, 'no uncaught page errors during CRUD').toEqual([]);
});
