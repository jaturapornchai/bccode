import { expect, test, type Page } from '@playwright/test';

// this suite exercises the login page itself — start LOGGED OUT
// (overrides the project-level storageState)
test.use({ storageState: { cookies: [], origins: [] } });

/**
 * Clear (X) buttons inside the three login inputs (2026-08-29):
 * mount only when the field has text, aria-labelled per field, password X
 * parks left of the eye toggle, clearing re-disables the CTA.
 */

const SHOT = 'test-results/clear-buttons';

async function shoot(page: Page, name: string) {
  await page.screenshot({ path: `${SHOT}/${name}.png` });
}

test.beforeEach(async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.goto('/');
  await expect(page).toHaveTitle(/เข้าสู่ระบบ/);
});

test('CB-01 clear button appears with text, clears, and unmounts', async ({ page }) => {
  // holding code
  const holding = page.locator('.field-group .input-with-icon input').first();
  await expect(page.locator('.holding-code-field .input-clear-icon')).toHaveCount(0);
  await holding.fill('bcdemo01');
  const x1 = page.locator('.holding-code-field .input-clear-icon');
  await expect(x1).toBeVisible();
  await expect(x1).toHaveAttribute('aria-label', 'ล้าง รหัสกลุ่มกิจการ');
  await x1.click();
  await expect(holding).toHaveValue('');
  await expect(x1).toHaveCount(0);
  await shoot(page, '01-holding-cleared');

  // username
  const user = page.locator('#login-username');
  await user.fill('uat_probe');
  const x2 = page.locator('#login-username ~ .input-clear-icon');
  await expect(x2).toBeVisible();
  await expect(x2).toHaveAttribute('aria-label', 'ล้าง ชื่อผู้ใช้');
  await x2.click();
  await expect(user).toHaveValue('');

  // password
  const pw = page.locator('#login-password');
  await pw.fill('uat-password-123456789');
  const x3 = page.locator('#login-password ~ .input-clear-icon');
  await expect(x3).toBeVisible();
  await expect(x3).toHaveAttribute('aria-label', 'ล้าง รหัสผ่าน');
  await shoot(page, '02-password-x-and-eye');
  await x3.click();
  await expect(pw).toHaveValue('');
});

test('CB-02 password: X left of eye, no intersection, eye still toggles', async ({ page }) => {
  const pw = page.locator('#login-password');
  await pw.fill('secret-password');
  // let the framer-motion entrance settle before measuring geometry
  await page.waitForTimeout(600);
  const x = page.locator('#login-password ~ .input-clear-icon');
  const eye = page.locator('#login-password ~ .input-trailing-icon');
  await expect(x).toBeVisible();
  await expect(eye).toBeVisible();

  const xr = await x.boundingBox();
  const er = await eye.boundingBox();
  expect(xr && er).toBeTruthy();
  // no intersection (1px subpixel tolerance), eye is the rightmost control
  expect(xr!.x + xr!.width).toBeLessThanOrEqual(er!.x + 1);
  // both are inside the input row vertically
  expect(Math.abs((xr!.y + xr!.height / 2) - (er!.y + er!.height / 2))).toBeLessThanOrEqual(1);

  await eye.click();
  await expect(pw).toHaveAttribute('type', 'text');
  await eye.click();
  await expect(pw).toHaveAttribute('type', 'password');
  await x.click();
  await expect(pw).toHaveValue('');
});

test('CB-03 clearing fields re-disables the CTA', async ({ page }) => {
  const holding = page.locator('.field-group .input-with-icon input').first();
  const user = page.locator('#login-username');
  const pw = page.locator('#login-password');
  const cta = page.locator('.login-card .primary-button');
  await holding.fill('bcdemo01');
  await user.fill('uat_probe');
  await pw.fill('uat-password-123456789');
  // let the framer-motion entrance settle before the enabled-state assert
  await page.waitForTimeout(600);
  await expect(cta).toBeEnabled();
  await page.locator('#login-password ~ .input-clear-icon').click();
  await expect(cta).toBeDisabled();
  await shoot(page, '03-cta-disabled-after-clear');
});
