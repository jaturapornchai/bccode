import { expect, test, type Page } from '@playwright/test';

// this suite exercises the login page itself — start LOGGED OUT
// (overrides the project-level storageState)
test.use({ storageState: { cookies: [], origins: [] } });

/**
 * Logout with an EXPIRED token (2026-08-29): "ปุ่มนี้ ตอน token หมดอายุ
 * กดแล้วไม่กลับจอแรก" — logoutAuthSession used to throw when the server
 * rejected the revoke (401 after refresh also failed), so the screens'
 * logout() aborted BEFORE clearing localStorage and redirecting.
 * Fix: best-effort revoke, always clear, never throw.
 * Simulation: intercept POST /api/auth/logout → 401.
 */

async function loginViaDev(page: Page) {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto('/');
  await page.waitForTimeout(900);
  await page.getByRole('button', { name: /Dev Login/ }).click();
  await page.waitForURL(/\/holding/, { timeout: 30000 });
  await page.waitForTimeout(1200);
}

test.describe.configure({ mode: 'serial' });

test('LO-01 workspace logout redirects to / even when revoke returns 401', async ({ page }) => {
  await loginViaDev(page);
  await page.getByRole('button', { name: /บ้านเชียง/ }).first().click();
  await page.waitForURL(/\/workspace/, { timeout: 30000 });
  await page.waitForTimeout(1200);

  // simulate fully-expired tokens: every revoke attempt answers 401
  await page.route('**/api/auth/logout', (route) =>
    route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ success: false }) }),
  );

  await page.locator('button[aria-label="ออกจากระบบ"]').first().click();
  await page.waitForURL(/\/$/, { timeout: 15000 });
  expect(page.url()).toMatch(/\/$/);

  const auth = await page.evaluate(() => localStorage.getItem('bc_auth'));
  const workspace = await page.evaluate(() => localStorage.getItem('bc_workspace'));
  const branch = await page.evaluate(() => localStorage.getItem('bc_branch'));
  expect(auth, 'auth session must be cleared').toBeNull();
  expect(workspace, 'workspace selection must be cleared').toBeNull();
  expect(branch, 'branch selection must be cleared').toBeNull();
  await page.screenshot({ path: 'test-results/logout-expired/01-workspace-logged-out.png' });
});

test('LO-02 holding logout redirects to / even when revoke returns 401', async ({ page }) => {
  await loginViaDev(page);
  await page.route('**/api/auth/logout', (route) =>
    route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ success: false }) }),
  );

  await page.locator('button[aria-label="ออกจากระบบ"], button[title="ออกจากระบบ"]').first().click();
  await page.waitForURL(/\/$/, { timeout: 15000 });
  expect(page.url()).toMatch(/\/$/);
  const auth = await page.evaluate(() => localStorage.getItem('bc_auth'));
  expect(auth, 'auth session must be cleared').toBeNull();
  await page.screenshot({ path: 'test-results/logout-expired/02-holding-logged-out.png' });
});

test('LO-03 logout still works normally (server 200)', async ({ page }) => {
  await loginViaDev(page);
  await page.locator('button[aria-label="ออกจากระบบ"], button[title="ออกจากระบบ"]').first().click();
  await page.waitForURL(/\/$/, { timeout: 15000 });
  expect(page.url()).toMatch(/\/$/);
  const auth = await page.evaluate(() => localStorage.getItem('bc_auth'));
  expect(auth).toBeNull();
});
