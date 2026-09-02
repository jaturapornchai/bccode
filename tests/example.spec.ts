import { test, expect } from '@playwright/test';

// Example smoke test against the local BC Ai Account frontend (baseURL from
// playwright.config.ts — default http://127.0.0.1:3000). Requires the local
// frontend container to be running.

test('login page has title', async ({ page }) => {
  await page.goto('/');

  // Expect the login page title to contain the Thai app name.
  await expect(page).toHaveTitle(/เข้าสู่ระบบ \| BC Ai Account/);
});

test('login card renders auth controls', async ({ page }) => {
  await page.goto('/');

  // The hero heading and the password field should be present.
  await expect(page.getByRole('heading', { name: 'เข้าสู่ระบบ' }).first()).toBeVisible();
  await expect(page.locator('#login-password')).toBeVisible();
});
