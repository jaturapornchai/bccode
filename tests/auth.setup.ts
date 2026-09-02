import { expect, test } from '@playwright/test';

/**
 * Auth setup — runs ONCE per full test run (project "setup").
 * Logs in via the real Dev Login UI, selects the TST03 workspace, then saves
 * storageState (.auth/user.json): refresh cookie + bc_auth/bc_workspace/
 * bc_branch localStorage. Every other project reuses it, so individual tests
 * skip the ~12s login flow entirely (the app bootstraps its in-memory token
 * from the refresh cookie + stored profile on load — see client-auth-session).
 */

test('login once and save workspace state', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto('/');
  await page.waitForTimeout(1500);
  await page.getByRole('button', { name: /Dev Login/ }).click();
  await page.waitForURL(/\/holding/, { timeout: 30000 });
  await page.waitForTimeout(1200);
  await page.getByRole('button', { name: /บ้านเชียง/ }).first().click();
  await page.waitForURL(/\/workspace/, { timeout: 30000 });
  await page.waitForTimeout(1000);
  await page.getByRole('button', { name: /สาขาทดสอบไทย|TST03/ }).first().click();
  await page.waitForTimeout(1200);
  await page.getByRole('button', { name: /สำนักงานใหญ่|00001/ }).first().click();
  await page.waitForTimeout(2500);

  // sanity: session + workspace selection actually materialised
  const auth = await page.evaluate(() => JSON.parse(localStorage.getItem('bc_auth') ?? 'null'));
  expect(auth?.username).toBeTruthy();
  const ws = await page.evaluate(() => JSON.parse(localStorage.getItem('bc_workspace') ?? 'null'));
  expect(ws?.shop?.holdingcode).toBe('bc001');

  await page.context().storageState({ path: '.auth/user.json' });
});
