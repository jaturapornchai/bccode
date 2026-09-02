import { execSync } from 'child_process';
import { expect, test, type Page } from '@playwright/test';

/**
 * Employee photo UAT (2026-08-31) — ตามกฎใหม่ "รูปภาพห้ามเก็บใน MongoDB":
 * อัปโหลดรูปพนักงานผ่าน UI → ตรวจว่า
 *  1) Mongo เก็บเป็น URI (/goapi/s3/file/...) เท่านั้น ไม่มี binary/base64
 *  2) ไฟล์จริง (ต้นฉบับ + thumbnail) อยู่ใน S3/MinIO — GET ผ่าน /goapi/s3/file/ ได้ 200
 *  3) จอ list + detail แสดงรูป (img ใช้ URI)
 * ใช้พนักงาน UATEMP01 (ข้อมูล UAT ถาวร) — helpers ลอกจาก employee-uat.spec.ts
 * (chain cookie .auth/user.json + self-heal ที่ผ่าน 7/7 หลายรอบ)
 */

const SHOT = 'test-results/employee-photo';

function mongoJson(js: string): any {
  const out = execSync(
    `docker exec mongodb mongosh --quiet appdb --eval "JSON.stringify(${js})"`,
    { timeout: 45000, shell: 'cmd.exe' },
  ).toString().trim();
  const start = out.indexOf('[') >= 0 ? out.indexOf('[') : out.indexOf('{');
  return JSON.parse(out.slice(start));
}

async function uiLogin(page: Page) {
  await page.goto('/');
  await page.getByRole('button', { name: /Dev Login/ }).click();
  await page.waitForURL('**/holding*', { timeout: 30000 });
  await page.getByRole('button', { name: /บ้านเชียง/ }).first().click();
  await page.waitForURL('**/workspace*', { timeout: 30000 });
  await page.getByRole('button', { name: /สาขาทดสอบไทย|TST03/ }).first().click();
  await page.getByRole('button', { name: /สำนักงานใหญ่|00001/ }).first().click();
  await page.waitForTimeout(1500);
}

async function ensureEmployeeScreen(page: Page) {
  const search = page.locator('input[placeholder*="ค้นหา พนักงาน"]');
  const devBtn = page.getByRole('button', { name: /Dev Login/ });
  const settled = await expect(search.or(devBtn).first()).toBeVisible({ timeout: 15000 }).then(() => true).catch(() => false);
  const needLogin = !settled || (await devBtn.count()) > 0;
  if (needLogin) {
    await uiLogin(page);
    await page.goto('/employee');
    await page.context().storageState({ path: '.auth/user.json' }).catch(() => {});
  }
  await expect(search).toBeVisible({ timeout: 15000 });
}

test.describe.configure({ mode: 'serial', retries: 1 });

test.afterEach(async ({ page }) => {
  await page.context().storageState({ path: '.auth/user.json' }).catch(() => {});
});

test('EMP-PHOTO: อัปโหลดรูปพนักงาน → S3 + thumb + Mongo เก็บ URI เท่านั้น', async ({ page }) => {
  test.setTimeout(300000);
  await ensureEmployeeScreen(page);

  // เปิดฟอร์มแก้ไข UATEMP01 — self-heal ทุกครั้งที่แถวยังไม่มา
  const row = page.locator('.bc-list-row').filter({ hasText: 'UATEMP01' }).first();
  let rowVisible = false;
  for (let attempt = 1; attempt <= 3 && !rowVisible; attempt++) {
    await ensureEmployeeScreen(page);
    rowVisible = await row.waitFor({ state: 'visible', timeout: 12000 }).then(() => true).catch(() => false);
  }
  expect(rowVisible, 'เห็นแถว UATEMP01 หลัง self-heal').toBe(true);

  // เปิดฟอร์มแก้ไข: กดปุ่มแก้ไขของแถว (เหมือน employee-uat E-03) — รอฟอร์มจริง
  let editOpened = false;
  for (let attempt = 1; attempt <= 3 && !editOpened; attempt++) {
    const rowEdit = row.getByRole('button', { name: /แก้ไข|Edit/ }).first();
    if (await rowEdit.count()) {
      await rowEdit.click().catch(() => {});
    } else {
      const detail = page.locator('section[aria-label="รายละเอียด"]');
      await detail.locator('button', { hasText: 'แก้ไข' }).first().click().catch(() => {});
    }
    editOpened = await page.locator('button', { hasText: /บันทึก|Save/ }).waitFor({ state: 'visible', timeout: 8000 }).then(() => true).catch(() => false);
  }
  expect(editOpened, 'ฟอร์มแก้ไขเปิด').toBe(true);
  // field รูปพนักงาน — input ต้องอยู่ "ใน" section รูปพนักงานเท่านั้น
  // (จอมี input[type=file] อื่นด้วย เช่น ช่อง CSV นำเข้าพนักงาน — ห้ามชนกัน)
  const photoSection = page.locator('section').filter({ has: page.getByText('รูปพนักงาน', { exact: true }) }).first();
  await expect(photoSection).toBeVisible({ timeout: 15000 });
  const fileInput = photoSection.locator('input[type="file"]');
  expect(await fileInput.count(), 'มี file input ใน section รูปพนักงาน').toBeGreaterThanOrEqual(1);

  // PNG 1x1 ที่ valid (base64 มาตรฐาน) — PNG เสียจะโดน handler ปฏิเสธตอนสร้าง thumbnail
  const png = Buffer.from(
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADElEQVR4nGP4z8AAAAMBAQDJ/pLvAAAAAElFTkSuQmCC',
    'base64',
  );
  const uploadOk = page.waitForResponse(
    (r) => r.url().includes('/api/upload/image') && r.status() === 200,
    { timeout: 30000 },
  );
  const uploadOk2 = page.waitForResponse(
    (r) => r.url().includes('/api/upload/image') && r.status() === 200,
    { timeout: 30000 },
  );
  await fileInput.setInputFiles({ name: 'uat-photo.png', mimeType: 'image/png', buffer: png });
  // อัปโหลด 2 ชิ้นตามกฎ: ต้นฉบับ + thumbnail
  await uploadOk;
  await uploadOk2;

  // บันทึกฟอร์ม — สำเร็จ = ฟอร์มปิด (file input หายไป)
  await page.locator('button', { hasText: /บันทึก|Save/ }).last().click();
  await expect(fileInput).toBeHidden({ timeout: 20000 });

  // 1) Mongo: URI เท่านั้น (ไม่มี base64/binary) — single-quote กัน cmd.exe กิน
  const docs = mongoJson(`db.employees.find({holdingcode:'bc001', code:'UATEMP01'}, {profilepicture:1, profilepicturethumb:1}).toArray()`);
  expect(docs.length).toBe(1);
  const uri = String(docs[0].profilepicture ?? '');
  const thumbUri = String(docs[0].profilepicturethumb ?? '');
  expect(uri, 'profilepicture เป็น URI S3').toMatch(/^\/goapi\/s3\/file\//);
  expect(thumbUri, 'profilepicturethumb มีคู่เสมอ (กฎ thumbnail)').toMatch(/^\/goapi\/s3\/file\//);
  expect(uri.length, 'ไม่มี binary ใน Mongo').toBeLessThan(300);
  expect(thumbUri.length).toBeLessThan(300);
  expect(uri).not.toContain('base64');

  // 2) S3/MinIO: ไฟล์จริง (ต้นฉบับ + thumb) ต้องอยู่ใน bucket — เช็คไฟล์ใน volume ตรง ๆ
  // (GET /goapi/s3/file/ ต้องใช้ auth — จึงตรวจไฟล์ฝั่ง MinIO โดยตรง)
  for (const [label, u] of [['original', uri], ['thumb', thumbUri]] as const) {
    const keyPath = u.replace('/goapi/s3/file/', '');
    const ls = execSync(
      `docker exec minio ls -la /data/bcai-account/${keyPath}`,
      { timeout: 30000, shell: 'cmd.exe' },
    ).toString().trim();
    expect(ls, `${label} object อยู่ใน MinIO bucket (${keyPath})`).toMatch(/\s\d+\s/); // มีขนาดไฟล์
  }

  // 3) UI: list แสดงรูป (thumb-first)
  await page.reload();
  await ensureEmployeeScreen(page);
  const rowImg = page.locator('.bc-list-row').filter({ hasText: 'UATEMP01' }).locator('img').first();
  await expect(rowImg).toBeVisible({ timeout: 15000 });
  // src เป็น blob: (authenticated-image ดึงด้วย token แล้ว render) หรือ URI ตรง ๆ ก็ได้
  await expect(rowImg).toHaveAttribute('src', /^(blob:|.*goapi\/s3\/file\/)/);
});
