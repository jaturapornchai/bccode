-- ============================================================================
-- Realistic Thai SME Accounting Data Clean-up for 'demo' DB (PostgreSQL)
-- Wrapped in a strict transaction (BEGIN ... COMMIT) for ACID data integrity.
-- ============================================================================

BEGIN;

-- 1. Holdings
UPDATE holdings 
SET name = 'กลุ่มกิจการรุ่งเรืองกรุ๊ป' 
WHERE code = 'demo';

UPDATE holdings 
SET name = 'บริษัท สยามพาณิชย์ กรุ๊ป จำกัด (มหาชน)' 
WHERE code = 'THAI_HOLDING';

-- 2. Companies
UPDATE companies 
SET name = 'สำนักงานใหญ่' 
WHERE code = '01';

UPDATE companies 
SET name = 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด' 
WHERE code = 'C01';

UPDATE companies 
SET name = 'บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด' 
WHERE code = 'C02';

UPDATE companies 
SET name = 'ห้างหุ้นส่วนจำกัด รุ่งเรืองการค้าไทย' 
WHERE code = 'C03';

-- 3. Users
UPDATE users 
SET full_name = 'ผู้ดูแลระบบ (Admin)' 
WHERE username = 'admin';

UPDATE users 
SET full_name = 'ผู้ใช้งานระบบบัญชี' 
WHERE username = 'demo';

-- 4. gl_records: Accounts
UPDATE gl_records 
SET payload = jsonb_set(
  payload, 
  '{names,0,name}', 
  to_jsonb(replace(payload->'names'->0->>'name', 'ข้อมูลตัวอย่าง - ', ''))
) 
WHERE payload->'names'->0->>'name' LIKE 'ข้อมูลตัวอย่าง - %';

UPDATE gl_records 
SET payload = jsonb_set(
  payload, 
  '{name}', 
  to_jsonb(replace(payload->>'name', 'ข้อมูลตัวอย่าง - ', ''))
) 
WHERE payload->>'name' LIKE 'ข้อมูลตัวอย่าง - %';

UPDATE gl_records 
SET payload = jsonb_set(
  jsonb_set(payload, '{name}', '"เงินฝากกระแสรายวัน - ธ.กสิกรไทย"'::jsonb), 
  '{names,0,name}', 
  '"เงินฝากกระแสรายวัน - ธ.กสิกรไทย"'::jsonb
) 
WHERE code = 'BM69-110102';

-- 5. gl_records: Account Groups
UPDATE gl_records 
SET payload = jsonb_set(
  payload, 
  '{name}', 
  to_jsonb(replace(payload->>'name', 'ข้อมูลตัวอย่าง - ', ''))
) 
WHERE kind = 'account-groups' AND payload->>'name' LIKE 'ข้อมูลตัวอย่าง - %';

-- 6. gl_records: Periods
UPDATE gl_records 
SET payload = jsonb_set(
  payload, 
  '{name}', 
  to_jsonb(replace(payload->>'name', 'ข้อมูลตัวอย่าง - งวด', 'งวดบัญชี'))
) 
WHERE kind = 'periods' AND payload->>'name' LIKE 'ข้อมูลตัวอย่าง - งวด%';

-- 7. gl_records: Journals
UPDATE gl_records 
SET payload = jsonb_set(
  payload, 
  '{description}', 
  to_jsonb(replace(replace(payload->>'description', 'ข้อมูลตัวอย่าง - ', ''), 'สมมติ', ''))
) 
WHERE kind = 'journals' AND (payload->>'description' LIKE '%ข้อมูลตัวอย่าง%' OR payload->>'description' LIKE '%สมมติ%');

-- 8. gl_lines
UPDATE gl_lines 
SET account_name = replace(account_name, 'ข้อมูลตัวอย่าง - ', '') 
WHERE account_name LIKE 'ข้อมูลตัวอย่าง - %';

UPDATE gl_lines 
SET account_name = 'เงินฝากกระแสรายวัน - ธ.กสิกรไทย' 
WHERE account_code = 'BM69-110102';

UPDATE gl_lines 
SET description = replace(replace(description, 'ข้อมูลตัวอย่าง - ', ''), 'สมมติ', '') 
WHERE description LIKE '%ข้อมูลตัวอย่าง%' OR description LIKE '%สมมติ%';

UPDATE gl_lines 
SET description = 'บันทึกรายการบัญชีประจำวัน' 
WHERE description = 'ข้อมูลตัวอย่าง' OR description = '' OR description IS NULL;

COMMIT;
