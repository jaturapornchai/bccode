import { expect, test } from '@playwright/test';

/**
 * Menu permission failure feedback (2026-08-31) — "เอาตรงนี้ให้จบ" follow-up.
 * ก่อนหน้านี้: เมื่อ /permissiongroup/me ล้ม (เซสชันตาย/เน็ตตาย) เมนูถูกล็อก
 * ทั้งหมด "เงียบ ๆ" ผู้ใช้ 40+ เห็นแต่ 🔒 ไม่มีสิทธิ์ โดยไม่รู้สาเหตุ
 * (เกิดจริงกับแท็บของลุงจืด หลัง backend restart)
 *
 * หลังแก้: ต้องมี toast ภาษาไทยอธิบายทุกกรณี และกรณีเซสชันหมดอายุต้อง
 * พากลับไปหน้า login ทันที (กฎ UX 40+ ข้อ 8: feedback ทุก action)
 *
 * ทุก test จำลอง "ตายกลางทาง" แบบเคสจริง: โหลด /menu สำเร็จก่อน (เซสชันยังดี)
 * แล้วค่อย intercept /me (และ /refresh ในเคสเซสชันตาย) แล้วกระตุ้น permission
 * refetch ด้วย window focus — trigger เดียวกับที่แอปใช้จริงเวลาผู้ใช้กลับมาที่แท็บ
 *
 * NOTE refresh-rotation: ทุก page load แอปจะยิง /api/auth/refresh จริงเสมอ
 * (token เก็บใน memory เท่านั้น) — ทุก test ต้อง afterEach บันทึก storageState
 * ต่อ cookie chain ไม่งั้น test ถัดไป reuse cookie เก่า → revoke ทั้ง session
 * เคสทำลายเซสชัน (MP-01) รันเป็นตัวสุดท้ายเพื่อไม่ทำร้ายเพื่อน
 */

const PERM_ROUTE = '**/api/system-settings/permissiongroup/me**';
const REFRESH_ROUTE = '**/api/auth/refresh';

// serial เท่านั้น: ทุก test แบ่ง refresh cookie กัน (single-use rotation) —
// ถ้ารันขนาน ทุก context เริ่มจาก cookie เดียวกัน → reuse → revoke ทั้ง session
test.describe.configure({ mode: 'serial' });

test.describe('menu permission failure feedback', () => {
  test.afterEach(async ({ page }) => {
    // cookie chain — เหตุผลดูหัวไฟล์ (refresh rotation single-use)
    await page.context().storageState({ path: '.auth/user.json' }).catch(() => {});
  });

  /** เปิด /menu พร้อม self-heal: ถ้า cookie โดน rotate จนตาย (เด้ง login) ให้
      Dev Login ใหม่แล้วกลับ /menu — ชุดนี้ไม่มี openEmployeeScreen ให้พึ่ง */
  async function openMenu(page: import('@playwright/test').Page) {
    await page.goto('/menu');
    const devBtn = page.getByRole('button', { name: /Dev Login/ });
    try {
      await expect(page.getByText('บ้านเชียง').or(devBtn).first()).toBeVisible({ timeout: 20000 });
    } catch {
      // ยังโหลดไม่ตั้ง — ลอง goto ซ้ำอีกครั้งเดียว
      await page.goto('/menu');
    }
    if (await devBtn.count()) {
      await devBtn.click();
      await page.waitForURL('**/holding*', { timeout: 30000 });
      await page.getByRole('button', { name: /บ้านเชียง/ }).first().click();
      await page.waitForURL('**/workspace*', { timeout: 30000 });
      await page.getByRole('button', { name: /สาขาทดสอบไทย|TST03/ }).first().click();
      await page.getByRole('button', { name: /สำนักงานใหญ่|00001/ }).first().click();
      await page.waitForTimeout(1500);
      await page.goto('/menu');
      await expect(page.getByText('บ้านเชียง').first()).toBeVisible({ timeout: 20000 });
    }
  }

  test('MP-02 network error → Thai toast, stays on /menu (no redirect)', async ({ page }) => {
    await openMenu(page);

    // เน็ตตายกลางทาง: /me ไปไม่ถึงเซิร์ฟเวอร์ แล้วผู้ใช้กลับมาที่แท็บ (focus)
    await page.route(PERM_ROUTE, (route) => route.abort());
    await page.evaluate(() => window.dispatchEvent(new Event('focus')));

    await expect(page.getByText(/เชื่อมต่อเซิร์ฟเวอร์ไม่ได้/)).toBeVisible({ timeout: 15000 });
    // not kicked out — the user stays; refetch happens on window focus
    await expect(page).toHaveURL(/\/menu/);
  });

  test('MP-03 signed in but no permission record → Thai toast, stays on /menu', async ({ page }) => {
    await openMenu(page);

    // มีเซสชันแต่ backend ไม่มี record สิทธิ์ที่ active สำหรับ holding นี้
    await page.route(PERM_ROUTE, (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ success: true, data: [] }) }),
    );
    await page.evaluate(() => window.dispatchEvent(new Event('focus')));

    await expect(page.getByText(/ไม่พบสิทธิ์การใช้งานเมนู/)).toBeVisible({ timeout: 15000 });
    await expect(page).toHaveURL(/\/menu/);
  });

  test('MP-04 breadcrumb แสดงจำนวนเซสชันออนไลน์จาก Redis จริง', async ({ page }) => {
    await openMenu(page);

    // chip บน breadcrumb: "ผู้ใช้งานออนไลน์: N เซสชัน" (N ≥ 1 — มีเซสชันของเทสนี้เอง)
    const chip = page.getByText(/ผู้ใช้งานออนไลน์: \d+ เซสชัน/);
    await expect(chip).toBeVisible({ timeout: 15000 });
    const text = await chip.textContent();
    const shown = Number((text ?? '').match(/(\d+)/)?.[1] ?? 0);

    // ตรวจสอบข้ามฝั่งกับ Redis จริง: activesessions ของ backend ต้องใกล้เคียงค่าที่แสดง
    // (ใช้ Lua EVAL คำสั่งเดียว — execSync บน Windows ผ่าน cmd.exe ใช้ pipe/while ไม่ได้)
    const { execSync } = await import('child_process');
    const lua = "local ks=redis.call('keys','session-*') local now=tonumber(redis.call('time')[1])*1000 local n=0 for _,k in ipairs(ks) do if string.sub(k,1,15)~='session-revoked' then local ls=tonumber(redis.call('hget',k,'lastseenat') or '0') if now-ls<=1800000 then n=n+1 end end end return n";
    const raw = execSync(`docker exec redis redis-cli EVAL "${lua}" 0`, { timeout: 45000 })
      .toString().trim();
    const redisActive = Number(raw);
    expect(shown, `UI shows ${shown}, Redis active(30m)=${redisActive}`).toBeGreaterThanOrEqual(1);
    expect(Math.abs(shown - redisActive), 'UI count matches Redis active window').toBeLessThanOrEqual(3);
  });

  test('MP-01 session dies mid-use → Thai toast + redirected to login', async ({ page }) => {
    // 1. เข้า /menu ตอนเซสชันยังดี (ยังไม่ intercept อะไร)
    await openMenu(page);
    await expect(page).toHaveURL(/\/menu/);

    // 2. เซสชันตาย: /me และ /refresh ตอบ 401 ทั้งคู่ (authFetch ลอง refresh แล้ว fail)
    await page.route(PERM_ROUTE, (route) => route.fulfill({ status: 401, body: 'unauthorized' }));
    await page.route(REFRESH_ROUTE, (route) => route.fulfill({ status: 401, body: 'unauthorized' }));

    // 3. กระตุ้น permission refetch เหมือนผู้ใช้กลับมาที่แท็บ (focus event)
    await page.evaluate(() => window.dispatchEvent(new Event('focus')));

    // 4. toast ไทยอธิบายเหตุผล (ไม่ใช่ล็อกเงียบ ๆ) + พาออกจากจอ protected
    await expect(page.getByText('เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่')).toBeVisible({ timeout: 15000 });
    await expect(page.getByRole('heading', { name: 'ลงชื่อเข้าใช้' })).toBeVisible({ timeout: 15000 });
    await expect(page).toHaveURL(/\/$|\/login/);
  });
});
