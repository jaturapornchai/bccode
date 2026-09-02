import { test, expect, type Page } from '@playwright/test';

// this suite exercises the login page itself — start LOGGED OUT
// (overrides the project-level storageState)
test.use({ storageState: { cookies: [], origins: [] } });

/**
 * UAT + UX/UI audit for BC Ai Account (Next.js frontend at baseURL,
 * default http://127.0.0.1:3000 — the local frontend container).
 *
 * Produces:
 *   test-results/uat/*.png        — screenshots at every meaningful step
 *   test-results/uat/metrics.json — DOM layout metrics (overlap scan,
 *                                   px-locked font audit, contrast signals)
 *
 * Flow under test (loopback → Dev Login is enabled by backend config):
 *   / → /holding → /workspace (เลือกบริษัท → เลือกสาขา) → settings wizard
 */

const shots = 'test-results/uat';
const metrics: Record<string, unknown> = {};

async function shoot(page: Page, name: string) {
  await page.screenshot({ path: `${shots}/${name}.png`, fullPage: false });
}

/** In-page audit: text-overlap scan + px-locked font detection. */
const AUDIT = `(() => {
  const range = document.createRange();
  const items = [];
  const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
  while (walker.nextNode()) {
    const node = walker.currentNode;
    if (!node.textContent.trim()) continue;
    const el = node.parentElement;
    if (!el || el.closest('script,style,noscript')) continue;
    const cs = getComputedStyle(el);
    if (cs.display === 'none' || cs.visibility === 'hidden' || parseFloat(cs.opacity) === 0) continue;
    range.selectNodeContents(node);
    for (const r of range.getClientRects()) {
      if (r.width < 2 || r.height < 2) continue;
      items.push({ text: node.textContent.trim().slice(0, 30), x: r.x, y: r.y, w: r.width, h: r.height, el });
    }
  }
  const overlaps = [];
  for (let i = 0; i < items.length; i++) {
    for (let j = i + 1; j < items.length; j++) {
      const a = items[i], b = items[j];
      if (a.el === b.el || a.el.contains(b.el) || b.el.contains(a.el)) continue;
      const ix = Math.max(0, Math.min(a.x + a.w, b.x + b.w) - Math.max(a.x, b.x));
      const iy = Math.max(0, Math.min(a.y + a.h, b.y + b.h) - Math.max(a.y, b.y));
      if (ix > 3 && iy > 3) {
        overlaps.push(a.text + ' × ' + b.text + ' (' + Math.round(ix) + 'x' + Math.round(iy) + 'px)');
      }
    }
  }
  const clippedButtons = Array.from(document.querySelectorAll('button')).filter(
    (b) => b.scrollHeight > b.clientHeight + 2 || b.scrollWidth > b.clientWidth + 2
  ).length;
  return {
    overlaps: overlaps.slice(0, 12),
    overlapCount: overlaps.length,
    clippedButtons,
    scrollHeight: document.documentElement.scrollHeight,
    viewportHeight: innerHeight,
  };
})()`;

async function audit(page: Page, key: string) {
  metrics[key] = await page.evaluate(AUDIT);
}

test.describe.configure({ mode: 'serial' });

test('UAT-01 login page renders + responsive screenshots', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveTitle(/เข้าสู่ระบบ/);

  for (const [w, h] of [[1280, 800], [1920, 1080], [700, 800]] as const) {
    await page.setViewportSize({ width: w, height: h });
    await page.waitForTimeout(600);
    await shoot(page, `01-login-${w}x${h}`);
    await audit(page, `login-${w}`);
  }
  await page.setViewportSize({ width: 1280, height: 800 });
});

test('UAT-02 login interactions: hover/focus states', async ({ page }) => {
  await page.goto('/');
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.waitForTimeout(800);

  // Fill valid-format (but wrong) creds first so the CTA is enabled — the
  // disabled state has pointer-events:none by design, so hover must target
  // the enabled button.
  await page.locator('.field-group .input-with-icon input').first().fill('bcdemo01');
  await page.locator('#login-username').fill('uat_probe');
  await page.locator('#login-password').fill('uat-password-123456789');

  const cta = page.locator('.login-card .primary-button');
  await expect(cta).toBeEnabled();
  await cta.hover();
  await shoot(page, '02-login-cta-hover');
  await page.locator('#login-username').click();
  await page.locator('#login-password').click();
  await shoot(page, '03-login-input-focus');
  const showPw = page.locator('.input-trailing-icon');
  await showPw.click();
  await shoot(page, '04-login-show-password');
});

test('UAT-03 empty submit is guarded', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(600);
  const cta = page.locator('.login-card .primary-button');
  // canSubmit requires username+password — with empty fields the CTA must be disabled
  await expect(cta).toBeDisabled();
  await shoot(page, '05-login-empty-disabled');
});

test('UAT-04 wrong credentials show friendly error', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(600);
  await page.locator('.field-group .input-with-icon input').first().fill('bcdemo01');
  await page.locator('#login-username').fill('uat_wrong_user');
  await page.locator('#login-password').fill('definitely-wrong-password');
  const cta = page.locator('.login-card .primary-button');
  await cta.click();
  // Friendly, non-revealing message per docs/login.md:83 + ERROR styling/icon
  // (defect D1 fixed 2026-08-29: connectionState used to flip it to success)
  const msg = page.locator('.login-card .message').first();
  await expect(msg).toBeVisible({ timeout: 15000 });
  await expect(msg).toHaveClass(/(^|\s)error($|\s)/);
  await expect(msg).not.toHaveClass(/success/);
  // D1: error icon must be circle-alert, never the success check
  await expect(msg.locator('svg.lucide-circle-alert')).toHaveCount(1);
  await expect(msg.locator('svg.lucide-check-circle, svg.lucide-circle-check')).toHaveCount(0);
  // D2: message text per docs/login.md:83
  await expect(msg).toContainText('ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง');
  await shoot(page, '06-login-wrong-creds-error');
  await audit(page, 'login-error');
});

test('UAT-05 double-click submit does not double-fire', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(600);
  await page.locator('.field-group .input-with-icon input').first().fill('bcdemo01');
  await page.locator('#login-username').fill('uat_wrong_user');
  await page.locator('#login-password').fill('definitely-wrong-password');
  const cta = page.locator('.login-card .primary-button');
  await cta.dblclick();
  await page.waitForTimeout(1200);
  // Either disabled-while-loading or single error state — assert exactly one message node
  const messages = await page.locator('.login-card .message').count();
  expect(messages).toBeLessThanOrEqual(1);
  await shoot(page, '07-login-double-click');
});

test('UAT-06 long/special input does not break layout', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(600);
  const long = 'ก'.repeat(120) + '!@#$%^&*()<>{}';
  await page.locator('#login-username').fill(long);
  await page.locator('#login-password').fill('x'.repeat(100));
  await shoot(page, '08-login-long-input');
  await audit(page, 'login-long-input');
});

test('UAT-07 full happy path: dev login → holding → workspace → settings', async ({ page }) => {
  test.setTimeout(120000);
  await page.goto('/');
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.waitForTimeout(1200);

  const dev = page.getByRole('button', { name: /Dev Login/ });
  await expect(dev).toBeVisible();
  await shoot(page, '09-login-dev-visible');
  await dev.click();
  await page.waitForURL(/\/holding/, { timeout: 30000 });
  await page.waitForTimeout(1200);
  await shoot(page, '10-holding-list');
  await audit(page, 'holding');

  // search filter
  const search = page.locator('.holding-search input');
  if (await search.count()) {
    await search.fill('บ้านเชียง');
    await page.waitForTimeout(500);
    await shoot(page, '11-holding-search');
    await search.fill('');
    await page.waitForTimeout(300);
  }

  // select the holding card
  const card = page.getByRole('button', { name: /บ้านเชียง/ }).first();
  await card.click();
  await page.waitForURL(/\/workspace/, { timeout: 30000 });
  await page.waitForTimeout(1200);
  await shoot(page, '12-workspace-companies');
  await audit(page, 'workspace-companies');

  // pick the first company
  const company = page.getByRole('button', { name: /ทดสอบการค้า|TST01|บริษัท/ }).first();
  await company.click();
  await page.waitForTimeout(1500);
  await shoot(page, '13-workspace-branches');
  await audit(page, 'workspace-branches');

  // pick the first branch (สำนักงานใหญ่)
  const branch = page.getByRole('button', { name: /สำนักงานใหญ่|00001/ }).first();
  await branch.click();
  await page.waitForTimeout(2500);
  await shoot(page, '14-after-branch-select');
  await audit(page, 'after-branch');

  // Settings screen (main functions): open via /settings
  await page.goto('/settings');
  await page.waitForTimeout(1500);
  await shoot(page, '15-settings-open');
  await audit(page, 'settings-open');

  const step3 = page.getByRole('button', { name: /รายการสิทธิ์หน้าจอ/ });
  if (await step3.count()) {
    await step3.first().click();
    await page.waitForTimeout(900);
    await shoot(page, '16-settings-step3');
    await audit(page, 'settings-step3');
  }

  const collapse = page.locator('button[title="ย่อเมนู"]');
  if (await collapse.count()) {
    await collapse.click();
    await page.waitForTimeout(600);
    await shoot(page, '17-settings-collapsed');
    await audit(page, 'settings-collapsed');
  }

  // Small-screen settings (700px pane — the earlier overlap trap zone)
  await page.setViewportSize({ width: 700, height: 800 });
  await page.waitForTimeout(700);
  await shoot(page, '18-settings-700px');
  await audit(page, 'settings-700');
});

test.afterAll(async () => {
  // Persist DOM audit metrics next to the screenshots for the report step.
  const fs = await import('node:fs');
  fs.mkdirSync(shots, { recursive: true });
  fs.writeFileSync(`${shots}/metrics.json`, JSON.stringify(metrics, null, 2));
});
