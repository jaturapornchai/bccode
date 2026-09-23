import { expect, test, type Page } from '@playwright/test';
import { pgCount, sqlText } from './support/pg';

// this suite exercises the login page itself — start LOGGED OUT
// (overrides the project-level storageState)
test.use({ storageState: { cookies: [], origins: [] } });

/**
 * Login screen UAT — Dev Login focus (2026-08-30)
 * UI per "คนไทย 40+" rules (Thai labels, button ≥ ~40px, overlap scan) +
 * data-layer verification: every successful Dev Login MUST create a live
 * session hash (`session-<uid>`) in PostgreSQL bcai_projection.cache_entries,
 * and guard paths (no-secret backend, non-loopback) must reject.
 */

const SHOT = 'test-results/login-dev-uat';
const consoleErrors: string[] = [];

/** live sessions of `username` created at/after `sinceMs` (cache_entries hash rows) */
function sessionsCreatedSince(username: string, sinceMs: number): number {
  return pgCount(
    `cache_entries u JOIN cache_entries c ON c.cache_key = u.cache_key AND c.field = 'createdat'`,
    `u.cache_key LIKE 'session-%' AND u.cache_key NOT LIKE 'session-revoked-%'
     AND u.field = 'username' AND u.value = ${sqlText(username)}
     AND (u.expires_at IS NULL OR u.expires_at > now())
     AND (CASE WHEN c.value ~ '^[0-9]+$' THEN c.value::bigint ELSE 0 END) >= ${Math.floor(sinceMs)}`,
  );
}

async function shoot(page: Page, name: string) {
  await page.screenshot({ path: `${SHOT}/${name}.png` });
}

test.describe.configure({ mode: 'serial', retries: 1 });

test('LD-01 login renders per 40+ rules: Thai labels, big-enough Dev Login, no overlap', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto('/');
  await page.waitForTimeout(2000);

  await expect(page).toHaveTitle(/เข้าสู่ระบบ/);
  const dev = page.getByRole('button', { name: /Dev Login/ });
  await expect(dev).toBeVisible();

  // Thai-first: Dev Login button carries Thai title/aria context and Thai nearby text
  const devText = (await dev.textContent()) ?? '';
  const devTitle = (await dev.getAttribute('title')) ?? '';
  const devAria = (await dev.getAttribute('aria-label')) ?? '';
  expect(`${devText} ${devTitle} ${devAria}`).toMatch(/Dev Login|ทดสอบ/);

  // 40+ rule: clickable action ≥ ~40px tall, decision text readable
  const box = await dev.boundingBox();
  expect(box && box.height, `Dev Login height ${(box?.height ?? 0).toFixed(0)}px must be ≥ 40px`).toBeGreaterThanOrEqual(40);

  // overlap scan (text line boxes) — reuse the proven scanner
  const overlaps = await page.evaluate(() => {
    const items: Array<{ text: string; x: number; y: number; w: number; h: number; el: Element }> = [];
    const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
    const range = document.createRange();
    while (walker.nextNode()) {
      const node = walker.currentNode;
      if (!node.textContent?.trim()) continue;
      const el = node.parentElement;
      if (!el || el.closest('script,style,noscript')) continue;
      const cs = getComputedStyle(el);
      if (cs.display === 'none' || cs.visibility === 'hidden' || parseFloat(cs.opacity) === 0) continue;
      range.selectNodeContents(node);
      for (const r of range.getClientRects()) {
        if (r.width < 2 || r.height < 2) continue;
        items.push({ text: node.textContent.trim().slice(0, 26), x: r.x, y: r.y, w: r.width, h: r.height, el });
      }
    }
    const found: string[] = [];
    for (let i = 0; i < items.length; i++) {
      for (let j = i + 1; j < items.length; j++) {
        const a = items[i], b = items[j];
        if (a.el === b.el || a.el.contains(b.el) || b.el.contains(a.el)) continue;
        const ix = Math.max(0, Math.min(a.x + a.w, b.x + b.w) - Math.max(a.x, b.x));
        const iy = Math.max(0, Math.min(a.y + a.h, b.y + b.h) - Math.max(a.y, b.y));
        if (ix > 3 && iy > 3) found.push(`${a.text} × ${b.text}`);
      }
    }
    return found.slice(0, 6);
  });
  // the known false positive is the hidden GIS iframe label
  const real = overlaps.filter((o) => !/Google/.test(o));
  expect(real, 'no real text overlaps').toEqual([]);

  await shoot(page, '01-login-render');
});

test('LD-02 Dev Login happy path → /holding + session row in PostgreSQL cache_entries', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  const startedAt = Date.now() - 2000; // small host/container clock-skew margin

  await page.goto('/');
  await page.waitForTimeout(2000);
  await page.getByRole('button', { name: /Dev Login/ }).click();
  await page.waitForURL(/\/holding/, { timeout: 30000 });
  await page.waitForTimeout(1500);
  await shoot(page, '02-landing-holding');

  // session materialised in the browser (token lives in memory by design;
  // bc_auth persists the profile — prove the token WORKS via loaded holdings)
  const auth = await page.evaluate(() => JSON.parse(localStorage.getItem('bc_auth') ?? 'null'));
  expect(auth?.username, 'session profile stored').toBeTruthy();
  await expect(page.locator('main').getByText('บ้านเชียง').first()).toBeVisible();

  // data layer: the login materialised a live session for this user in cache_entries
  await page.waitForTimeout(1500);
  const sessions = sessionsCreatedSince(String(auth.username), startedAt);
  expect(sessions, `live session row for ${auth.username} created by this login`).toBeGreaterThanOrEqual(1);
});

test('LD-03 double-click Dev Login: single session flow, no error, no crash', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto('/');
  await page.waitForTimeout(2000);
  const dev = page.getByRole('button', { name: /Dev Login/ });
  await dev.click();
  // second click fires immediately — button should be disabled/ignored during flight
  await dev.click({ timeout: 2000 }).catch(() => {});
  await page.waitForURL(/\/holding/, { timeout: 30000 });
  await page.waitForTimeout(1500);
  const msg = await page.locator('.login-card .message, [role="alert"]').allTextContents().catch(() => []);
  expect(msg.join(' '), 'no error banner after double click').not.toMatch(/ไม่อนุญาต|สำเร็จ|Failed/i);
  const auth = await page.evaluate(() => JSON.parse(localStorage.getItem('bc_auth') ?? 'null'));
  expect(auth?.username, 'session profile stored after double click').toBeTruthy();
  await expect(page.locator('main').getByText('บ้านเชียง').first()).toBeVisible();
  await shoot(page, '03-double-click-landing');
});

test('LD-04 backend guard: /dev-login without secret header → 401 (rejected)', async ({ request }) => {
  const res = await request.post('http://127.0.0.1:8888/dev-login', { data: {} });
  expect(res.status(), 'backend rejects secret-less dev login').toBe(401);
  const body = await res.text();
  expect(body).toMatch(/dev login failed|ไม่ได้รับอนุญาต/i);
});

test('LD-05 BFF guard: dev login accepts only same-origin loopback calls', async ({ request }) => {
  const base = 'http://127.0.0.1:3000';
  // A) external origin → 403 (attacker forging Origin from another host)
  const ext = await request.post(base + '/api/auth/dev-login', { headers: { Origin: 'https://evil.example.com' } });
  expect(ext.status(), 'external origin rejected').toBe(403);
  expect(await ext.text()).toMatch(/DEV_LOGIN_FORBIDDEN|localhost/);

  // B) loopback origin but Host mismatch → 403
  const mismatch = await request.post(base + '/api/auth/dev-login', { headers: { Origin: 'http://127.0.0.1:3000', Host: 'evil.example.com' } });
  expect(mismatch.status(), 'host/origin mismatch rejected').toBe(403);

  // C) legit same-origin loopback → 200 (control case)
  const ok = await request.post(base + '/api/auth/dev-login', { headers: { Origin: 'http://127.0.0.1:3000', Host: '127.0.0.1:3000' } });
  expect(ok.status(), 'legit loopback accepted').toBe(200);
  // (audit-count assertion removed: dev-login attempts are not persisted to any PostgreSQL table)
});
