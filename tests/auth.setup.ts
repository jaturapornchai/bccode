import { expect, test } from '@playwright/test';

test('login once and save workspace state', async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.goto('/');

  // Click the Demo Login button directly via class or role
  const demoBtn = page.locator('.dev-login-button, button:has-text("ทดลองใช้ระบบ")').first();
  await demoBtn.waitFor({ state: 'visible', timeout: 20000 });
  await demoBtn.click();

  // Wait for /holding redirect
  await page.waitForURL(/\/holding/, { timeout: 30000 });
  await page.waitForTimeout(1000);

  // Select holding "กลุ่มกิจการรุ่งเรืองกรุ๊ป" (or first holding card)
  const holdingBtn = page.locator('button').filter({ hasText: /รุ่งเรือง|rungrueng|บ้านเชียง/i }).first();
  await holdingBtn.waitFor({ state: 'visible', timeout: 10000 });
  await holdingBtn.click();

  // Wait for /workspace redirect
  await page.waitForURL(/\/workspace/, { timeout: 30000 });
  await page.waitForTimeout(1000);

  // Select company (01 / บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด)
  const companyBtn = page.locator('button').filter({ hasText: /บริษัท|01|TST/i }).first();
  await companyBtn.waitFor({ state: 'visible', timeout: 10000 });
  await companyBtn.click();
  await page.waitForTimeout(1000);

  // Select branch (00000 / สำนักงานใหญ่)
  const branchBtn = page.locator('button').filter({ hasText: /สำนักงานใหญ่|00000|00001/i }).first();
  await branchBtn.waitFor({ state: 'visible', timeout: 10000 });
  await branchBtn.click();
  await page.waitForTimeout(2000);

  // Verify auth session in localStorage
  const auth = await page.evaluate(() => JSON.parse(localStorage.getItem('bc_auth') ?? 'null'));
  expect(auth?.username).toBeTruthy();
  const ws = await page.evaluate(() => JSON.parse(localStorage.getItem('bc_workspace') ?? 'null'));
  expect(ws?.shop?.holdingcode).toBeTruthy();

  await page.context().storageState({ path: '.auth/user.json' });
});
