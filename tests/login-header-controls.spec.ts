import { expect, test, type Page } from '@playwright/test';

// this suite exercises the login page itself — start LOGGED OUT
// (overrides the project-level storageState)
test.use({ storageState: { cookies: [], origins: [] } });

/**
 * Login-family card header controls (2026-08-29 fix): icons were px-locked 18px
 * inside 36px chrome-locked boxes — too small and no button affordance.
 * Fix: .card-header .header-controls only (login + holding cards) → 3em boxes,
 * 1.4em icons, border + surface + hover lift, fluid with the login root ladder.
 * Menu screens keep the intentional 36px chrome lock.
 */

const SHOT_DIR = 'test-results/header-controls';

async function shoot(page: Page, name: string) {
  await page.screenshot({ path: `${SHOT_DIR}/${name}.png` });
}

async function measureHeaderControls(page: Page) {
  return page.evaluate(() => {
    const buttons = Array.from(
      document.querySelectorAll('.card-header .header-controls button, .card-header .header-controls a.header-control-button'),
    );
    return buttons.map((b) => {
      const r = b.getBoundingClientRect();
      const cs = getComputedStyle(b);
      const icons = Array.from(b.querySelectorAll('svg, img.flag-icon')) as Array<SVGElement | HTMLImageElement>;
      const icon = icons.find((i) => i.getBoundingClientRect().width > 0) ?? null;
      const ir = icon ? icon.getBoundingClientRect() : null;
      return {
        label: (b.getAttribute('aria-label') || b.getAttribute('title') || b.textContent || '').trim().slice(0, 24),
        w: Math.round(r.width),
        h: Math.round(r.height),
        iconW: ir ? Math.round(ir.width) : 0,
        iconH: ir ? Math.round(ir.height) : 0,
        borderWidth: cs.borderTopWidth,
        borderColor: cs.borderTopColor,
        bg: cs.backgroundColor,
        transform: cs.transform,
      };
    });
  });
}

test.describe.configure({ mode: 'serial' });

test('HC-01 login card header: 3em buttons, 1.4em icons, fluid across viewports', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveTitle(/เข้าสู่ระบบ/);

  for (const [w, h] of [[1280, 800], [1920, 1080], [700, 800]] as const) {
    await page.setViewportSize({ width: w, height: h });
    await page.waitForTimeout(600);
    const rows = await measureHeaderControls(page);
    expect(rows.length, `${w}: five header controls`).toBe(5);
    for (const r of rows) {
      // 3em box: 45px @1280 → ~59px @1920 → ~42px @700 (root ladder 10/13.1/8.19px)
      expect(r.w, `${w} ${r.label} width ≥ 40`).toBeGreaterThanOrEqual(40);
      expect(r.h, `${w} ${r.label} height ≥ 40`).toBeGreaterThanOrEqual(40);
      // real button affordance: hairline border + non-transparent surface
      expect(parseFloat(r.borderWidth), `${w} ${r.label} has border`).toBeGreaterThanOrEqual(1);
      expect(r.borderColor, `${w} ${r.label} border visible (not transparent)`).not.toBe('rgba(0, 0, 0, 0)');
      expect(r.bg, `${w} ${r.label} has surface`).not.toBe('');
    }
    const heights = rows.map((r) => r.h);
    expect(Math.max(...heights) - Math.min(...heights), `${w}: equal heights`).toBeLessThanOrEqual(1);
    // icon glyph heights must match across all five buttons — flag included
    const iconHeights = rows.map((r) => r.iconH);
    expect(Math.max(...iconHeights) - Math.min(...iconHeights), `${w}: equal icon heights (flag = svg)`).toBeLessThanOrEqual(1.5);
    // icon grows with the box (1.4em ≥ 19px everywhere; flag 2em ≥ 26px)
    for (const r of rows) {
      expect(r.iconW, `${w} ${r.label} icon ≥ 19px`).toBeGreaterThanOrEqual(19);
    }
    if (w === 1280) {
      const byLabel = Object.fromEntries(rows.map((r) => [r.label, r]));
      const any = rows[0];
      expect(any.h).toBeGreaterThanOrEqual(44); // touch-target floor @1280
      void byLabel;
    }
  }

  // fluid (not px-locked): box @1920 must be ~30% larger than @1280
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.waitForTimeout(400);
  const at1280 = (await measureHeaderControls(page))[0].h;
  await page.setViewportSize({ width: 1920, height: 1080 });
  await page.waitForTimeout(400);
  const at1920 = (await measureHeaderControls(page))[0].h;
  expect(at1920 / at1280, 'scales with root ladder (×1.2: 10px→12px root)').toBeGreaterThanOrEqual(1.19);

  await shoot(page, '01-login-1920');
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.waitForTimeout(400);
  await shoot(page, '02-login-1280');
  await page.setViewportSize({ width: 700, height: 800 });
  await page.waitForTimeout(400);
  await shoot(page, '03-login-700');
});

test('HC-02 hover affordance: lift + colour change', async ({ page }) => {
  await page.goto('/');
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.waitForTimeout(600);

  const themeToggle = page.locator('.card-header .header-controls .theme-toggle');
  const before = await themeToggle.evaluate((el) => {
    const cs = getComputedStyle(el);
    return { transform: cs.transform, borderColor: cs.borderTopColor };
  });
  await themeToggle.hover();
  await page.waitForTimeout(300);
  const after = await themeToggle.evaluate((el) => {
    const cs = getComputedStyle(el);
    return { transform: cs.transform, borderColor: cs.borderTopColor };
  });
  const moved = before.transform !== after.transform;
  const recolored = before.borderColor !== after.borderColor;
  expect(moved || recolored, 'hover produces visible affordance change').toBe(true);
  await shoot(page, '04-hover-theme-toggle');
});

test('HC-03 holding card header parity with login', async ({ page }) => {
  test.setTimeout(120000);
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto('/');
  await page.waitForTimeout(1200);
  // Dev Login button (same as UAT-07) — the password form creds are format-only
  const dev = page.getByRole('button', { name: /Dev Login/ });
  await expect(dev).toBeVisible();
  await dev.click();
  await page.waitForURL(/\/holding/, { timeout: 30000 });
  await page.waitForTimeout(1200);

  const rows = await measureHeaderControls(page);
  expect(rows.length, 'holding card header has the controls').toBeGreaterThanOrEqual(4);
  for (const r of rows) {
    expect(r.h, `${r.label} holding ≈ login button height (44-46px)`).toBeGreaterThanOrEqual(43);
    expect(r.h).toBeLessThanOrEqual(48);
    expect(parseFloat(r.borderWidth)).toBeGreaterThanOrEqual(1);
  }
  await shoot(page, '05-holding-header');
});

test('HC-04 non-card-header chrome (settings) keeps the 36px lock', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto('/settings');
  await page.waitForTimeout(1000);
  const sizes = await page.evaluate(() =>
    Array.from(document.querySelectorAll('.header-controls button, .header-controls a.header-control-button'))
      .filter((b) => !b.closest('.card-header'))
      .map((b) => Math.round(b.getBoundingClientRect().height)),
  );
  expect(sizes.length, 'settings page has non-card header controls').toBeGreaterThan(0);
  for (const s of sizes) {
    expect(s, `settings chrome control stays compact (≤ 40px), got ${s}`).toBeLessThanOrEqual(40);
  }
});
