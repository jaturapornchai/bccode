// UI smoke: Demo login (the real login page button, never a password) → open every implemented menu screen
// through the menu search → no "รอพัฒนา" chip, no console/page error, no API ≥ 400 → screenshot → close the tab.
// Prod: holding RUNGRUENG company 01 only (never TOFFT1), open screens only — no clicks inside a screen.
import path from 'node:path';
import fs from 'node:fs';
import { chromium } from 'playwright';
import { implementedScreens } from './screens.mjs';

// Dev-server noise that is not an app failure.
const IGNORED_CONSOLE = [/Download the React DevTools/, /\[HMR\]/, /\[Fast Refresh\]/, /favicon\.ico/];

async function demoLogin(page, fe, check) {
  await page.goto(fe + '/', { waitUntil: 'domcontentloaded' });
  await page.locator('.dev-login-button, button:has-text("ทดลองใช้ระบบ")').first().click({ timeout: 30000 });
  await page.waitForURL(/\/holding/, { timeout: 30000 });
  await page.locator('button').filter({ hasText: /รุ่งเรือง|rungrueng/i }).first().click({ timeout: 15000 });
  await page.waitForURL(/\/workspace/, { timeout: 30000 });
  await page.locator('button').filter({ hasText: /บริษัท|01/ }).first().click({ timeout: 15000 });
  await page.locator('button').filter({ hasText: /สำนักงานใหญ่|00000/ }).first().click({ timeout: 15000 });
  await page.waitForTimeout(1500);
  const ws = await page.evaluate(() => JSON.parse(localStorage.getItem('bc_workspace') ?? 'null'));
  return check('login.workspace', /rungrueng/i.test(ws?.shop?.holdingcode ?? ''), `holding=${ws?.shop?.holdingcode}`);
}

export async function runUiSmoke({ fe, out, heads, routes, dark, recorder }) {
  const all = implementedScreens();
  const byHead = heads.includes('all') ? all : all.filter((s) => heads.includes(s.head));
  const screens = routes ? routes.map((r) => byHead.find((s) => s.route === r)).filter(Boolean) : byHead;
  const { check } = recorder('ui');
  check('screens.selected', screens.length > 0, `${screens.length}/${all.length} screens for ${heads.join(',')}`);
  const shotDir = path.join(out, 'shots');
  fs.rmSync(shotDir, { recursive: true, force: true });
  fs.mkdirSync(shotDir, { recursive: true });

  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1440, height: 900 }, locale: 'th-TH', timezoneId: 'Asia/Bangkok' });
  let issues = [];
  let apiErrors = []; // response bodies arrive async; awaited before each verdict
  page.on('console', (m) => {
    if (m.type() === 'error' && !IGNORED_CONSOLE.some((re) => re.test(m.text()))) issues.push(`console: ${m.text()}`);
  });
  page.on('pageerror', (e) => issues.push(`pageerror: ${e.message}`));
  page.on('response', (r) => {
    const u = new URL(r.url());
    if (r.status() >= 400 && u.pathname.startsWith('/api/')) apiErrors.push(r.text().catch(() => '').then((t) => `api ${r.status()} ${r.request().method()} ${u.pathname} ${t.slice(0, 120)}`));
  });

  let shots = 0;
  try {
    if (!(await demoLogin(page, fe, check))) return 0;
    await page.goto(fe + '/menu', { waitUntil: 'domcontentloaded' });
    const search = page.locator('input.\\!pl-10').first();
    await search.waitFor({ timeout: 30000 });
    issues = [];

    for (const s of screens) {
      const { check: screenCheck } = recorder(`ui:${s.head}`);
      issues = [];
      apiErrors = [];
      try {
        if (s.searchable) {
          await search.fill(s.route);
          await page.locator(`button[data-menu-route="${s.route}"]`).first().click({ timeout: 8000 });
        } else {
          await page.locator(`[data-menu-head="${s.head}"]`).first().click({ timeout: 8000 });
          await page.locator(`[role="menuitem"][data-menu-route="${s.route}"]`).first().click({ timeout: 8000 });
        }
        const panel = page.locator('section[role="tabpanel"]:not([hidden])').last();
        await panel.waitFor({ timeout: 15000 });
        await page.waitForLoadState('networkidle', { timeout: 12000 }).catch(() => issues.push('slow: network busy > 12s'));
        if (!s.dataPending && (await panel.locator('[data-menu-pending]').count())) issues.push('รอพัฒนา chip on an implemented screen');
        if (!((await panel.innerText()).trim())) issues.push('blank screen');
        const file = `${s.head}-${s.id}`.replace(/[^a-z0-9-]/gi, '_');
        await page.screenshot({ path: path.join(shotDir, `${file}.png`) });
        shots++;
        if (dark) {
          await page.locator('button.theme-toggle').first().click();
          await page.waitForTimeout(300);
          await page.screenshot({ path: path.join(shotDir, `${file}-dark.png`) });
          await page.locator('button.theme-toggle').first().click();
          shots++;
        }
        issues.push(...(await Promise.all(apiErrors)));
        screenCheck(s.id, issues.length === 0, `${s.route} ${[...new Set(issues)].join(' | ')}`);
      } catch (e) {
        screenCheck(s.id, false, `${s.route} ${e.message.split('\n')[0]} ${[...new Set(issues)].join(' | ')}`);
      }
      // Close the tab; a dirty guard or a stuck screen falls back to a fresh /menu.
      const closeButton = page.locator('[role="tab"][aria-selected="true"] ~ button, [role="tab"][aria-selected="true"] button').first();
      const tabsBefore = await page.locator('[role="tab"]').count();
      await closeButton.click({ timeout: 3000 }).catch(() => {});
      if ((await page.locator('[role="tab"]').count()) >= tabsBefore || !(await search.isVisible())) {
        await page.goto(fe + '/menu', { waitUntil: 'domcontentloaded' });
        await search.waitFor({ timeout: 30000 });
      }
    }
  } catch (e) {
    check('EXCEPTION', false, e.stack ?? e.message);
  } finally {
    await browser.close();
  }
  return shots;
}
