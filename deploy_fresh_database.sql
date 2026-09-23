-- ============================================================================
-- Fresh Database Reset & Single SME Company Accounting Seed (Pure PostgreSQL)
-- Single Holding: 'rungrueng' (กลุ่มกิจการรุ่งเรืองกรุ๊ป)
-- Single Company: '01' (บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด)
-- Single Branch: '00000' (สำนักงานใหญ่)
-- Multi-level Chart of Accounts (TFRS for NPAEs / DBD Standards: Level 1-4)
-- Balanced Sample Journals (Opening Balance, Purchase, Sale)
-- Strictly lowercase identifiers [a-z0-9_] throughout.
-- ============================================================================

-- 1. Safely disable triggers for append-only tables and clean up
ALTER TABLE gl_events DISABLE TRIGGER ALL;
TRUNCATE TABLE gl_lines CASCADE;
TRUNCATE TABLE gl_records CASCADE;
TRUNCATE TABLE gl_events CASCADE;
TRUNCATE TABLE gl_projection_state CASCADE;
ALTER TABLE gl_events ENABLE TRIGGER ALL;

TRUNCATE TABLE role_permissions CASCADE;
TRUNCATE TABLE holding_members CASCADE;
TRUNCATE TABLE user_sessions CASCADE;
TRUNCATE TABLE user_identities CASCADE;
TRUNCATE TABLE branches CASCADE;
TRUNCATE TABLE companies CASCADE;
TRUNCATE TABLE holdings CASCADE;
TRUNCATE TABLE users CASCADE;
TRUNCATE TABLE chartofaccounts CASCADE;
TRUNCATE TABLE business_types CASCADE;
TRUNCATE TABLE employees CASCADE;

BEGIN;

-- 2. Users (Real display names, Password is "Password123456789", Including jaturapornchai)
INSERT INTO users (id, username, password_hash, email, phone, full_name, is_active, created_at, updated_at)
VALUES 
  ('dd13b9bc-f65f-4fd3-b1b1-f21f3266b51c', 'jaturapornchai', '$2a$10$sfqaFRidw68/wt5jxKeWoe3lArcQ8mVHbVTyXR36MohgSL.9J0OJ6', 'jaturapornchai@gmail.com', '0818889999', 'jaturapornchai ratanapanya', true, now(), now()),
  ('11111111-1111-1111-1111-111111111111', 'admin', '$2a$10$sfqaFRidw68/wt5jxKeWoe3lArcQ8mVHbVTyXR36MohgSL.9J0OJ6', 'admin@rungrueng.co.th', '0812345678', 'ผู้ดูแลระบบ', true, now(), now()),
  ('22222222-2222-2222-2222-222222222222', 'somsak', '$2a$10$sfqaFRidw68/wt5jxKeWoe3lArcQ8mVHbVTyXR36MohgSL.9J0OJ6', 'somsak@rungrueng.co.th', '0898765432', 'สมศักดิ์ บัญชีการค้า', true, now(), now()),
  ('33333333-3333-3333-3333-333333333333', 'demo', '$2a$10$sfqaFRidw68/wt5jxKeWoe3lArcQ8mVHbVTyXR36MohgSL.9J0OJ6', 'somkid@rungrueng.co.th', '0865551234', 'สมคิด พาณิชย์เจริญ', true, now(), now()),
  ('44444444-4444-4444-4444-444444444444', 'wipawan', '$2a$10$sfqaFRidw68/wt5jxKeWoe3lArcQ8mVHbVTyXR36MohgSL.9J0OJ6', 'wipawan@rungrueng.co.th', '0823456789', 'วิภาวรรณ ใจดี', true, now(), now()),
  ('55555555-5555-5555-5555-555555555555', 'anucha', '$2a$10$sfqaFRidw68/wt5jxKeWoe3lArcQ8mVHbVTyXR36MohgSL.9J0OJ6', 'anucha@rungrueng.co.th', '0834567890', 'อนุชา ขยันงาน', true, now(), now());

-- 3. Holdings (Single Holding: 'rungrueng')
INSERT INTO holdings (code, name, tax_id, is_active)
VALUES 
  ('rungrueng', 'กลุ่มกิจการรุ่งเรืองกรุ๊ป', '0105560001235', true);

-- 4. Companies (Single Company: '01')
INSERT INTO companies (holding_code, code, name, tax_id, is_active)
VALUES 
  ('rungrueng', '01', 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด', '0105566123456', true);

-- 5. Branches (Single Branch: '00000' สำนักงานใหญ่)
INSERT INTO branches (holding_code, company_code, code, name, is_headquarters, is_active)
VALUES 
  ('rungrueng', '01', '00000', 'สำนักงานใหญ่', true, true);

-- 6. Holding Memberships (All assigned to 'rungrueng' with clear scopes)
INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes, is_active, created_at)
VALUES 
  ('rungrueng', 'dd13b9bc-f65f-4fd3-b1b1-f21f3266b51c', 'owner', '["*"]'::jsonb, '{}'::jsonb, true, now()),
  ('rungrueng', '11111111-1111-1111-1111-111111111111', 'owner', '["*"]'::jsonb, '{}'::jsonb, true, now()),
  ('rungrueng', '33333333-3333-3333-3333-333333333333', 'owner', '["*"]'::jsonb, '{}'::jsonb, true, now()),
  ('rungrueng', '22222222-2222-2222-2222-222222222222', 'user', '["ACCOUNTANT"]'::jsonb, '[{"scopetype": "company", "allbranches": true, "businesscode": "01"}]'::jsonb, true, now()),
  ('rungrueng', '44444444-4444-4444-4444-444444444444', 'user', '["SALES"]'::jsonb, '[{"scopetype": "branch", "branchcode": "00000", "businesscode": "01"}]'::jsonb, true, now()),
  ('rungrueng', '55555555-5555-5555-5555-555555555555', 'user', '["WAREHOUSE"]'::jsonb, '[{"scopetype": "company", "allbranches": true, "businesscode": "01"}]'::jsonb, true, now());

-- 7. Role Permissions (Full Access for all standard roles in 'rungrueng')
INSERT INTO role_permissions (holding_code, role_code, permissions)
VALUES 
  ('rungrueng', 'owner', '["*"]'::jsonb),
  ('rungrueng', 'admin', '["*"]'::jsonb),
  ('rungrueng', 'accountant', '["*"]'::jsonb),
  ('rungrueng', 'staff', '["*"]'::jsonb),
  ('rungrueng', 'OWNER', '["*"]'::jsonb),
  ('rungrueng', 'ADMIN', '["*"]'::jsonb);

-- 8. Business Types for 'rungrueng'
INSERT INTO business_types (id, holding_code, code, names, is_default, is_active)
VALUES
  ('bt-01', 'rungrueng', 'BT01', '[{"code":"th","name":"ค้าปลีก-ค้าส่งวัสดุก่อสร้าง"},{"code":"en","name":"Construction Materials Retail & Wholesale"}]'::jsonb, true, true),
  ('bt-02', 'rungrueng', 'BT02', '[{"code":"th","name":"ร้านอาหารและเครื่องดื่ม"},{"code":"en","name":"Food & Beverage / Cafe"}]'::jsonb, false, true),
  ('bt-03', 'rungrueng', 'BT03', '[{"code":"th","name":"ธุรกิจบริการและรับเหมา"},{"code":"en","name":"Service & Contractor"}]'::jsonb, false, true),
  ('bt-04', 'rungrueng', 'BT04', '[{"code":"th","name":"สำนักงานบัญชีและที่ปรึกษา"},{"code":"en","name":"Accounting Firm & Advisory"}]'::jsonb, false, true)
ON CONFLICT (holding_code, code) DO NOTHING;

-- 9. Fiscal Year (ปีบัญชี 2569)
INSERT INTO gl_records (company, kind, id, code, version, payload)
VALUES 
  ('01', 'fiscal-years', 'year-2569', '2569', 1, '{"id": "year-2569", "code": "2569", "startdate": "2026-01-01", "enddate": "2026-12-31", "isactive": true, "scale": 2, "profitlossaccount": "32102", "retainedearningsaccount": "32101", "closed": false}');

-- 10. Journal Books (สมุดรายวันมาตรฐาน 6 เล่ม)
INSERT INTO gl_records (company, kind, id, code, version, payload)
VALUES
  ('01', 'journal-books', 'book-gj', 'GJ', 1, '{"code": "GJ", "name": "สมุดรายวันทั่วไป", "names": [{"code": "th", "name": "สมุดรายวันทั่วไป"}], "isactive": true}'),
  ('01', 'journal-books', 'book-pv', 'PV', 1, '{"code": "PV", "name": "สมุดรายวันจ่ายเงิน", "names": [{"code": "th", "name": "สมุดรายวันจ่ายเงิน"}], "isactive": true}'),
  ('01', 'journal-books', 'book-rv', 'RV', 1, '{"code": "RV", "name": "สมุดรายวันรับเงิน", "names": [{"code": "th", "name": "สมุดรายวันรับเงิน"}], "isactive": true}'),
  ('01', 'journal-books', 'book-ap', 'AP', 1, '{"code": "AP", "name": "สมุดรายวันซื้อสินค้า", "names": [{"code": "th", "name": "สมุดรายวันซื้อสินค้า"}], "isactive": true}'),
  ('01', 'journal-books', 'book-ar', 'AR', 1, '{"code": "AR", "name": "สมุดรายวันขายสินค้า", "names": [{"code": "th", "name": "สมุดรายวันขายสินค้า"}], "isactive": true}'),
  ('01', 'journal-books', 'book-jv', 'JV', 1, '{"code": "JV", "name": "สมุดรายวันทั่วไป (JV)", "names": [{"code": "th", "name": "สมุดรายวันทั่วไป (JV)"}], "isactive": true}');

-- 11. Chart of Accounts (ผังบัญชีหลายระดับตามมาตรฐานวิชาชีพบัญชีไทย TFRS for NPAEs / DBD Level 1-4)
INSERT INTO gl_records (company, kind, id, code, version, payload)
VALUES
  -- ==========================================
  -- หมวด 1 สินทรัพย์ (Assets) — Normal: Debit
  -- ==========================================
  ('01', 'accounts', 'acc-10000', '10000', 1, '{"accountcode": "10000", "name": "สินทรัพย์", "names": [{"code": "th", "name": "สินทรัพย์"}], "accounttype": "asset", "level": 1, "parentaccountcode": "", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11000', '11000', 1, '{"accountcode": "11000", "name": "สินทรัพย์หมุนเวียน", "names": [{"code": "th", "name": "สินทรัพย์หมุนเวียน"}], "accounttype": "asset", "level": 2, "parentaccountcode": "10000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  
  -- 1.1 เงินสดและรายการเทียบเท่าเงินสด
  ('01', 'accounts', 'acc-11100', '11100', 1, '{"accountcode": "11100", "name": "เงินสดและรายการเทียบเท่าเงินสด", "names": [{"code": "th", "name": "เงินสดและรายการเทียบเท่าเงินสด"}], "accounttype": "asset", "level": 3, "parentaccountcode": "11000", "normalbalance": "debit", "allowposting": false, "iscash": true, "isactive": true}'),
  ('01', 'accounts', 'acc-11101', '11101', 1, '{"accountcode": "11101", "name": "เงินสดในมือ", "names": [{"code": "th", "name": "เงินสดในมือ"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11100", "normalbalance": "debit", "allowposting": true, "iscash": true, "isactive": true}'),
  ('01', 'accounts', 'acc-11102', '11102', 1, '{"accountcode": "11102", "name": "เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย", "names": [{"code": "th", "name": "เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11100", "normalbalance": "debit", "allowposting": true, "iscash": true, "isactive": true}'),
  ('01', 'accounts', 'acc-11103', '11103', 1, '{"accountcode": "11103", "name": "เงินฝากออมทรัพย์ - ธนาคารไทยพาณิชย์", "names": [{"code": "th", "name": "เงินฝากออมทรัพย์ - ธนาคารไทยพาณิชย์"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11100", "normalbalance": "debit", "allowposting": true, "iscash": true, "isactive": true}'),
  ('01', 'accounts', 'acc-11104', '11104', 1, '{"accountcode": "11104", "name": "เงินสดย่อย", "names": [{"code": "th", "name": "เงินสดย่อย"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11100", "normalbalance": "debit", "allowposting": true, "iscash": true, "isactive": true}'),

  -- 1.2 ลูกหนี้การค้าและลูกหนี้อื่น
  ('01', 'accounts', 'acc-11200', '11200', 1, '{"accountcode": "11200", "name": "ลูกหนี้การค้าและลูกหนี้อื่น", "names": [{"code": "th", "name": "ลูกหนี้การค้าและลูกหนี้อื่น"}], "accounttype": "asset", "level": 3, "parentaccountcode": "11000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11201', '11201', 1, '{"accountcode": "11201", "name": "ลูกหนี้การค้า", "names": [{"code": "th", "name": "ลูกหนี้การค้า"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11200", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11202', '11202', 1, '{"accountcode": "11202", "name": "เช็ครับล่วงหน้า", "names": [{"code": "th", "name": "เช็ครับล่วงหน้า"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11200", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11203', '11203', 1, '{"accountcode": "11203", "name": "ลูกหนี้อื่น", "names": [{"code": "th", "name": "ลูกหนี้อื่น"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11200", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- 1.3 สินค้าคงเหลือ
  ('01', 'accounts', 'acc-11300', '11300', 1, '{"accountcode": "11300", "name": "สินค้าคงเหลือ", "names": [{"code": "th", "name": "สินค้าคงเหลือ"}], "accounttype": "asset", "level": 3, "parentaccountcode": "11000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11301', '11301', 1, '{"accountcode": "11301", "name": "สินค้าสำเร็จรูป", "names": [{"code": "th", "name": "สินค้าสำเร็จรูป"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11300", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11302', '11302', 1, '{"accountcode": "11302", "name": "วัตถุดิบ", "names": [{"code": "th", "name": "วัตถุดิบ"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11300", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- 1.4 สินทรัพย์หมุนเวียนอื่น
  ('01', 'accounts', 'acc-11400', '11400', 1, '{"accountcode": "11400", "name": "สินทรัพย์หมุนเวียนอื่น", "names": [{"code": "th", "name": "สินทรัพย์หมุนเวียนอื่น"}], "accounttype": "asset", "level": 3, "parentaccountcode": "11000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11401', '11401', 1, '{"accountcode": "11401", "name": "ภาษีซื้อ", "names": [{"code": "th", "name": "ภาษีซื้อ"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11400", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11402', '11402', 1, '{"accountcode": "11402", "name": "ภาษีซื้อยังไม่ถึงกำหนด", "names": [{"code": "th", "name": "ภาษีซื้อยังไม่ถึงกำหนด"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11400", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11403', '11403', 1, '{"accountcode": "11403", "name": "ภาษีถูกหัก ณ ที่จ่าย", "names": [{"code": "th", "name": "ภาษีถูกหัก ณ ที่จ่าย"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11400", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11404', '11404', 1, '{"accountcode": "11404", "name": "ค่าใช้จ่ายจ่ายล่วงหน้า", "names": [{"code": "th", "name": "ค่าใช้จ่ายจ่ายล่วงหน้า"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11400", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- 1.5 สินทรัพย์ไม่หมุนเวียน
  ('01', 'accounts', 'acc-12000', '12000', 1, '{"accountcode": "12000", "name": "สินทรัพย์ไม่หมุนเวียน", "names": [{"code": "th", "name": "สินทรัพย์ไม่หมุนเวียน"}], "accounttype": "asset", "level": 2, "parentaccountcode": "10000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12100', '12100', 1, '{"accountcode": "12100", "name": "ที่ดิน อาคารและอุปกรณ์", "names": [{"code": "th", "name": "ที่ดิน อาคารและอุปกรณ์"}], "accounttype": "asset", "level": 3, "parentaccountcode": "12000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12101', '12101', 1, '{"accountcode": "12101", "name": "ที่ดิน", "names": [{"code": "th", "name": "ที่ดิน"}], "accounttype": "asset", "level": 4, "parentaccountcode": "12100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12102', '12102', 1, '{"accountcode": "12102", "name": "อาคารและสิ่งปลูกสร้าง", "names": [{"code": "th", "name": "อาคารและสิ่งปลูกสร้าง"}], "accounttype": "asset", "level": 4, "parentaccountcode": "12100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12103', '12103', 1, '{"accountcode": "12103", "name": "ค่าเสื่อมราคาสะสม-อาคาร", "names": [{"code": "th", "name": "ค่าเสื่อมราคาสะสม-อาคาร"}], "accounttype": "asset", "level": 4, "parentaccountcode": "12100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12104', '12104', 1, '{"accountcode": "12104", "name": "เครื่องตกแต่งและอุปกรณ์สำนักงาน", "names": [{"code": "th", "name": "เครื่องตกแต่งและอุปกรณ์สำนักงาน"}], "accounttype": "asset", "level": 4, "parentaccountcode": "12100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12105', '12105', 1, '{"accountcode": "12105", "name": "ค่าเสื่อมราคาสะสม-เครื่องตกแต่งและอุปกรณ์", "names": [{"code": "th", "name": "ค่าเสื่อมราคาสะสม-เครื่องตกแต่งและอุปกรณ์"}], "accounttype": "asset", "level": 4, "parentaccountcode": "12100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12106', '12106', 1, '{"accountcode": "12106", "name": "ยานพาหนะ", "names": [{"code": "th", "name": "ยานพาหนะ"}], "accounttype": "asset", "level": 4, "parentaccountcode": "12100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12107', '12107', 1, '{"accountcode": "12107", "name": "ค่าเสื่อมราคาสะสม-ยานพาหนะ", "names": [{"code": "th", "name": "ค่าเสื่อมราคาสะสม-ยานพาหนะ"}], "accounttype": "asset", "level": 4, "parentaccountcode": "12100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- ==========================================
  -- หมวด 2 หนี้สิน (Liabilities) — Normal: Credit
  -- ==========================================
  ('01', 'accounts', 'acc-20000', '20000', 1, '{"accountcode": "20000", "name": "หนี้สิน", "names": [{"code": "th", "name": "หนี้สิน"}], "accounttype": "liability", "level": 1, "parentaccountcode": "", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21000', '21000', 1, '{"accountcode": "21000", "name": "หนี้สินหมุนเวียน", "names": [{"code": "th", "name": "หนี้สินหมุนเวียน"}], "accounttype": "liability", "level": 2, "parentaccountcode": "20000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  
  -- 2.1 เจ้าหนี้การค้าและเจ้าหนี้อื่น
  ('01', 'accounts', 'acc-21100', '21100', 1, '{"accountcode": "21100", "name": "เจ้าหนี้การค้าและเจ้าหนี้หมุนเวียนอื่น", "names": [{"code": "th", "name": "เจ้าหนี้การค้าและเจ้าหนี้หมุนเวียนอื่น"}], "accounttype": "liability", "level": 3, "parentaccountcode": "21000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21101', '21101', 1, '{"accountcode": "21101", "name": "เจ้าหนี้การค้า", "names": [{"code": "th", "name": "เจ้าหนี้การค้า"}], "accounttype": "liability", "level": 4, "parentaccountcode": "21100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21102', '21102', 1, '{"accountcode": "21102", "name": "เช็คจ่ายล่วงหน้า", "names": [{"code": "th", "name": "เช็คจ่ายล่วงหน้า"}], "accounttype": "liability", "level": 4, "parentaccountcode": "21100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21103', '21103', 1, '{"accountcode": "21103", "name": "เจ้าหนี้อื่น", "names": [{"code": "th", "name": "เจ้าหนี้อื่น"}], "accounttype": "liability", "level": 4, "parentaccountcode": "21100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- 2.2 เงินกู้ยืมระยะสั้น
  ('01', 'accounts', 'acc-21200', '21200', 1, '{"accountcode": "21200", "name": "เงินกู้ยืมระยะสั้น", "names": [{"code": "th", "name": "เงินกู้ยืมระยะสั้น"}], "accounttype": "liability", "level": 3, "parentaccountcode": "21000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21201', '21201', 1, '{"accountcode": "21201", "name": "เงินเบิกเกินบัญชีธนาคาร (O/D)", "names": [{"code": "th", "name": "เงินเบิกเกินบัญชีธนาคาร (O/D)"}], "accounttype": "liability", "level": 4, "parentaccountcode": "21200", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- 2.3 หนี้สินหมุนเวียนอื่น
  ('01', 'accounts', 'acc-21300', '21300', 1, '{"accountcode": "21300", "name": "หนี้สินหมุนเวียนอื่น", "names": [{"code": "th", "name": "หนี้สินหมุนเวียนอื่น"}], "accounttype": "liability", "level": 3, "parentaccountcode": "21000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21301', '21301', 1, '{"accountcode": "21301", "name": "ภาษีขาย", "names": [{"code": "th", "name": "ภาษีขาย"}], "accounttype": "liability", "level": 4, "parentaccountcode": "21300", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21302', '21302', 1, '{"accountcode": "21302", "name": "ภาษีขายยังไม่ถึงกำหนด", "names": [{"code": "th", "name": "ภาษีขายยังไม่ถึงกำหนด"}], "accounttype": "liability", "level": 4, "parentaccountcode": "21300", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21303', '21303', 1, '{"accountcode": "21303", "name": "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย", "names": [{"code": "th", "name": "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย"}], "accounttype": "liability", "level": 4, "parentaccountcode": "21300", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21304', '21304', 1, '{"accountcode": "21304", "name": "ค่าใช้จ่ายค้างจ่าย", "names": [{"code": "th", "name": "ค่าใช้จ่ายค้างจ่าย"}], "accounttype": "liability", "level": 4, "parentaccountcode": "21300", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21305', '21305', 1, '{"accountcode": "21305", "name": "ประกันสังคมค้างจ่าย", "names": [{"code": "th", "name": "ประกันสังคมค้างจ่าย"}], "accounttype": "liability", "level": 4, "parentaccountcode": "21300", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- 2.4 หนี้สินไม่หมุนเวียน
  ('01', 'accounts', 'acc-22000', '22000', 1, '{"accountcode": "22000", "name": "หนี้สินไม่หมุนเวียน", "names": [{"code": "th", "name": "หนี้สินไม่หมุนเวียน"}], "accounttype": "liability", "level": 2, "parentaccountcode": "20000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-22100', '22100', 1, '{"accountcode": "22100", "name": "เงินกู้ยืมระยะยาว", "names": [{"code": "th", "name": "เงินกู้ยืมระยะยาว"}], "accounttype": "liability", "level": 3, "parentaccountcode": "22000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-22101', '22101', 1, '{"accountcode": "22101", "name": "เงินกู้ยืมระยะยาวจากสถาบันการเงิน", "names": [{"code": "th", "name": "เงินกู้ยืมระยะยาวจากสถาบันการเงิน"}], "accounttype": "liability", "level": 4, "parentaccountcode": "22100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- ==========================================
  -- หมวด 3 ส่วนของเจ้าของ (Equity) — Normal: Credit
  -- ==========================================
  ('01', 'accounts', 'acc-30000', '30000', 1, '{"accountcode": "30000", "name": "ส่วนของเจ้าของ", "names": [{"code": "th", "name": "ส่วนของเจ้าของ"}], "accounttype": "equity", "level": 1, "parentaccountcode": "", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-31000', '31000', 1, '{"accountcode": "31000", "name": "ทุนเรือนหุ้น", "names": [{"code": "th", "name": "ทุนเรือนหุ้น"}], "accounttype": "equity", "level": 2, "parentaccountcode": "30000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-31100', '31100', 1, '{"accountcode": "31100", "name": "ทุนจดทะเบียน", "names": [{"code": "th", "name": "ทุนจดทะเบียน"}], "accounttype": "equity", "level": 3, "parentaccountcode": "31000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-31101', '31101', 1, '{"accountcode": "31101", "name": "ทุนหุ้นสามัญ", "names": [{"code": "th", "name": "ทุนหุ้นสามัญ"}], "accounttype": "equity", "level": 4, "parentaccountcode": "31100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-32000', '32000', 1, '{"accountcode": "32000", "name": "กำไร(ขาดทุน)สะสม", "names": [{"code": "th", "name": "กำไร(ขาดทุน)สะสม"}], "accounttype": "equity", "level": 2, "parentaccountcode": "30000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-32100', '32100', 1, '{"accountcode": "32100", "name": "กำไร(ขาดทุน)สะสม", "names": [{"code": "th", "name": "กำไร(ขาดทุน)สะสม"}], "accounttype": "equity", "level": 3, "parentaccountcode": "32000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-32101', '32101', 1, '{"accountcode": "32101", "name": "กำไรสะสมที่ยังไม่ได้จัดสรร", "names": [{"code": "th", "name": "กำไรสะสมที่ยังไม่ได้จัดสรร"}], "accounttype": "equity", "level": 4, "parentaccountcode": "32100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-32102', '32102', 1, '{"accountcode": "32102", "name": "กำไร(ขาดทุน)สุทธิประจำปี", "names": [{"code": "th", "name": "กำไร(ขาดทุน)สุทธิประจำปี"}], "accounttype": "equity", "level": 4, "parentaccountcode": "32100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- ==========================================
  -- หมวด 4 รายได้ (Income) — Normal: Credit
  -- ==========================================
  ('01', 'accounts', 'acc-40000', '40000', 1, '{"accountcode": "40000", "name": "รายได้", "names": [{"code": "th", "name": "รายได้"}], "accounttype": "income", "level": 1, "parentaccountcode": "", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-41000', '41000', 1, '{"accountcode": "41000", "name": "รายได้จากการดำเนินงาน", "names": [{"code": "th", "name": "รายได้จากการดำเนินงาน"}], "accounttype": "income", "level": 2, "parentaccountcode": "40000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-41100', '41100', 1, '{"accountcode": "41100", "name": "รายได้จากการขาย", "names": [{"code": "th", "name": "รายได้จากการขาย"}], "accounttype": "income", "level": 3, "parentaccountcode": "41000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-41101', '41101', 1, '{"accountcode": "41101", "name": "รายได้จากการขายสินค้า", "names": [{"code": "th", "name": "รายได้จากการขายสินค้า"}], "accounttype": "income", "level": 4, "parentaccountcode": "41100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-41102', '41102', 1, '{"accountcode": "41102", "name": "รับคืนและส่วนลดจ่าย", "names": [{"code": "th", "name": "รับคืนและส่วนลดจ่าย"}], "accounttype": "income", "level": 4, "parentaccountcode": "41100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-41200', '41200', 1, '{"accountcode": "41200", "name": "รายได้จากการให้บริการ", "names": [{"code": "th", "name": "รายได้จากการให้บริการ"}], "accounttype": "income", "level": 3, "parentaccountcode": "41000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-41201', '41201', 1, '{"accountcode": "41201", "name": "รายได้จากการให้บริการ", "names": [{"code": "th", "name": "รายได้จากการให้บริการ"}], "accounttype": "income", "level": 4, "parentaccountcode": "41200", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-42000', '42000', 1, '{"accountcode": "42000", "name": "รายได้อื่น", "names": [{"code": "th", "name": "รายได้อื่น"}], "accounttype": "income", "level": 2, "parentaccountcode": "40000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-42100', '42100', 1, '{"accountcode": "42100", "name": "รายได้อื่น", "names": [{"code": "th", "name": "รายได้อื่น"}], "accounttype": "income", "level": 3, "parentaccountcode": "42000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-42101', '42101', 1, '{"accountcode": "42101", "name": "ดอกเบี้ยรับ", "names": [{"code": "th", "name": "ดอกเบี้ยรับ"}], "accounttype": "income", "level": 4, "parentaccountcode": "42100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-42102', '42102', 1, '{"accountcode": "42102", "name": "รายได้เบ็ดเตล็ด", "names": [{"code": "th", "name": "รายได้เบ็ดเตล็ด"}], "accounttype": "income", "level": 4, "parentaccountcode": "42100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- ==========================================
  -- หมวด 5 ค่าใช้จ่าย (Expenses) — Normal: Debit
  -- ==========================================
  ('01', 'accounts', 'acc-50000', '50000', 1, '{"accountcode": "50000", "name": "ค่าใช้จ่าย", "names": [{"code": "th", "name": "ค่าใช้จ่าย"}], "accounttype": "expense", "level": 1, "parentaccountcode": "", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-51000', '51000', 1, '{"accountcode": "51000", "name": "ต้นทุนขายและบริการ", "names": [{"code": "th", "name": "ต้นทุนขายและบริการ"}], "accounttype": "expense", "level": 2, "parentaccountcode": "50000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-51100', '51100', 1, '{"accountcode": "51100", "name": "ต้นทุนขาย", "names": [{"code": "th", "name": "ต้นทุนขาย"}], "accounttype": "expense", "level": 3, "parentaccountcode": "51000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-51101', '51101', 1, '{"accountcode": "51101", "name": "ต้นทุนขาย", "names": [{"code": "th", "name": "ต้นทุนขาย"}], "accounttype": "expense", "level": 4, "parentaccountcode": "51100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-51102', '51102', 1, '{"accountcode": "51102", "name": "ซื้อสินค้า", "names": [{"code": "th", "name": "ซื้อสินค้า"}], "accounttype": "expense", "level": 4, "parentaccountcode": "51100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-51103', '51103', 1, '{"accountcode": "51103", "name": "ค่าขนส่งเข้า", "names": [{"code": "th", "name": "ค่าขนส่งเข้า"}], "accounttype": "expense", "level": 4, "parentaccountcode": "51100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-51104', '51104', 1, '{"accountcode": "51104", "name": "ส่งคืนและส่วนลดรับ", "names": [{"code": "th", "name": "ส่งคืนและส่วนลดรับ"}], "accounttype": "expense", "level": 4, "parentaccountcode": "51100", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- 5.2 ค่าใช้จ่ายในการขาย
  ('01', 'accounts', 'acc-52000', '52000', 1, '{"accountcode": "52000", "name": "ค่าใช้จ่ายในการขาย", "names": [{"code": "th", "name": "ค่าใช้จ่ายในการขาย"}], "accounttype": "expense", "level": 2, "parentaccountcode": "50000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-52100', '52100', 1, '{"accountcode": "52100", "name": "ค่าใช้จ่ายในการขาย", "names": [{"code": "th", "name": "ค่าใช้จ่ายในการขาย"}], "accounttype": "expense", "level": 3, "parentaccountcode": "52000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-52101', '52101', 1, '{"accountcode": "52101", "name": "เงินเดือนและคอมมิชชั่นฝ่ายขาย", "names": [{"code": "th", "name": "เงินเดือนและคอมมิชชั่นฝ่ายขาย"}], "accounttype": "expense", "level": 4, "parentaccountcode": "52100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-52102', '52102', 1, '{"accountcode": "52102", "name": "ค่าโฆษณาและส่งเสริมการขาย", "names": [{"code": "th", "name": "ค่าโฆษณาและส่งเสริมการขาย"}], "accounttype": "expense", "level": 4, "parentaccountcode": "52100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-52103', '52103', 1, '{"accountcode": "52103", "name": "ค่าขนส่งออก", "names": [{"code": "th", "name": "ค่าขนส่งออก"}], "accounttype": "expense", "level": 4, "parentaccountcode": "52100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- 5.3 ค่าใช้จ่ายในการบริหาร
  ('01', 'accounts', 'acc-53000', '53000', 1, '{"accountcode": "53000", "name": "ค่าใช้จ่ายในการบริหาร", "names": [{"code": "th", "name": "ค่าใช้จ่ายในการบริหาร"}], "accounttype": "expense", "level": 2, "parentaccountcode": "50000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-53100', '53100', 1, '{"accountcode": "53100", "name": "ค่าใช้จ่ายในการบริหาร", "names": [{"code": "th", "name": "ค่าใช้จ่ายในการบริหาร"}], "accounttype": "expense", "level": 3, "parentaccountcode": "53000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-53101', '53101', 1, '{"accountcode": "53101", "name": "เงินเดือนและค่าจ้างฝ่ายบริหาร", "names": [{"code": "th", "name": "เงินเดือนและค่าจ้างฝ่ายบริหาร"}], "accounttype": "expense", "level": 4, "parentaccountcode": "53100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-53102', '53102', 1, '{"accountcode": "53102", "name": "เงินสมทบกองทุนประกันสังคม", "names": [{"code": "th", "name": "เงินสมทบกองทุนประกันสังคม"}], "accounttype": "expense", "level": 4, "parentaccountcode": "53100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-53103', '53103', 1, '{"accountcode": "53103", "name": "ค่าเช่าสำนักงาน", "names": [{"code": "th", "name": "ค่าเช่าสำนักงาน"}], "accounttype": "expense", "level": 4, "parentaccountcode": "53100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-53104', '53104', 1, '{"accountcode": "53104", "name": "ค่าไฟฟ้าและน้ำประปา", "names": [{"code": "th", "name": "ค่าไฟฟ้าและน้ำประปา"}], "accounttype": "expense", "level": 4, "parentaccountcode": "53100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-53105', '53105', 1, '{"accountcode": "53105", "name": "ค่าโทรศัพท์และอินเทอร์เน็ต", "names": [{"code": "th", "name": "ค่าโทรศัพท์และอินเทอร์เน็ต"}], "accounttype": "expense", "level": 4, "parentaccountcode": "53100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-53106', '53106', 1, '{"accountcode": "53106", "name": "ค่าเครื่องเขียนและแบบพิมพ์", "names": [{"code": "th", "name": "ค่าเครื่องเขียนและแบบพิมพ์"}], "accounttype": "expense", "level": 4, "parentaccountcode": "53100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-53107', '53107', 1, '{"accountcode": "53107", "name": "ค่าเสื่อมราคา-อาคาร", "names": [{"code": "th", "name": "ค่าเสื่อมราคา-อาคาร"}], "accounttype": "expense", "level": 4, "parentaccountcode": "53100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-53108', '53108', 1, '{"accountcode": "53108", "name": "ค่าเสื่อมราคา-เครื่องตกแต่งและอุปกรณ์", "names": [{"code": "th", "name": "ค่าเสื่อมราคา-เครื่องตกแต่งและอุปกรณ์"}], "accounttype": "expense", "level": 4, "parentaccountcode": "53100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-53109', '53109', 1, '{"accountcode": "53109", "name": "ค่าเสื่อมราคา-ยานพาหนะ", "names": [{"code": "th", "name": "ค่าเสื่อมราคา-ยานพาหนะ"}], "accounttype": "expense", "level": 4, "parentaccountcode": "53100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- 5.4 ต้นทุนทางการเงิน
  ('01', 'accounts', 'acc-54000', '54000', 1, '{"accountcode": "54000", "name": "ต้นทุนทางการเงินและภาษี", "names": [{"code": "th", "name": "ต้นทุนทางการเงินและภาษี"}], "accounttype": "expense", "level": 2, "parentaccountcode": "50000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-54100', '54100', 1, '{"accountcode": "54100", "name": "ต้นทุนทางการเงิน", "names": [{"code": "th", "name": "ต้นทุนทางการเงิน"}], "accounttype": "expense", "level": 3, "parentaccountcode": "54000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-54101', '54101', 1, '{"accountcode": "54101", "name": "ดอกเบี้ยจ่ายและค่าธรรมเนียมธนาคาร", "names": [{"code": "th", "name": "ดอกเบี้ยจ่ายและค่าธรรมเนียมธนาคาร"}], "accounttype": "expense", "level": 4, "parentaccountcode": "54100", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}');

-- 12. Legacy chartofaccounts table sync for 'rungrueng'
INSERT INTO chartofaccounts (holdingcode, accountcode, accountname, accountcategory, accountlevel, consolidateaccountcode)
VALUES
  ('rungrueng', '10000', 'สินทรัพย์', 1, 1, ''),
  ('rungrueng', '11000', 'สินทรัพย์หมุนเวียน', 1, 2, '10000'),
  ('rungrueng', '11100', 'เงินสดและรายการเทียบเท่าเงินสด', 1, 3, '11000'),
  ('rungrueng', '11101', 'เงินสดในมือ', 1, 4, '11100'),
  ('rungrueng', '11102', 'เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย', 1, 4, '11100'),
  ('rungrueng', '11103', 'เงินฝากออมทรัพย์ - ธนาคารไทยพาณิชย์', 1, 4, '11100'),
  ('rungrueng', '11104', 'เงินสดย่อย', 1, 4, '11100'),
  ('rungrueng', '11200', 'ลูกหนี้การค้าและลูกหนี้อื่น', 1, 3, '11000'),
  ('rungrueng', '11201', 'ลูกหนี้การค้า', 1, 4, '11200'),
  ('rungrueng', '11300', 'สินค้าคงเหลือ', 1, 3, '11000'),
  ('rungrueng', '11301', 'สินค้าสำเร็จรูป', 1, 4, '11300'),
  ('rungrueng', '11400', 'สินทรัพย์หมุนเวียนอื่น', 1, 3, '11000'),
  ('rungrueng', '11401', 'ภาษีซื้อ', 1, 4, '11400'),
  ('rungrueng', '12000', 'สินทรัพย์ไม่หมุนเวียน', 1, 2, '10000'),
  ('rungrueng', '12100', 'ที่ดิน อาคารและอุปกรณ์', 1, 3, '12000'),
  ('rungrueng', '12101', 'ที่ดิน', 1, 4, '12100'),
  ('rungrueng', '12102', 'อาคารและสิ่งปลูกสร้าง', 1, 4, '12100'),
  ('rungrueng', '12103', 'ค่าเสื่อมราคาสะสม-อาคาร', 1, 4, '12100'),
  ('rungrueng', '20000', 'หนี้สิน', 2, 1, ''),
  ('rungrueng', '21000', 'หนี้สินหมุนเวียน', 2, 2, '20000'),
  ('rungrueng', '21100', 'เจ้าหนี้การค้าและเจ้าหนี้หมุนเวียนอื่น', 2, 3, '21000'),
  ('rungrueng', '21101', 'เจ้าหนี้การค้า', 2, 4, '21100'),
  ('rungrueng', '21300', 'หนี้สินหมุนเวียนอื่น', 2, 3, '21000'),
  ('rungrueng', '21301', 'ภาษีขาย', 2, 4, '21300'),
  ('rungrueng', '22000', 'หนี้สินไม่หมุนเวียน', 2, 2, '20000'),
  ('rungrueng', '22100', 'เงินกู้ยืมระยะยาว', 2, 3, '22000'),
  ('rungrueng', '22101', 'เงินกู้ยืมระยะยาวจากสถาบันการเงิน', 2, 4, '22100'),
  ('rungrueng', '30000', 'ส่วนของเจ้าของ', 3, 1, ''),
  ('rungrueng', '31000', 'ทุนเรือนหุ้น', 3, 2, '30000'),
  ('rungrueng', '31100', 'ทุนจดทะเบียน', 3, 3, '31000'),
  ('rungrueng', '31101', 'ทุนหุ้นสามัญ', 3, 4, '31100'),
  ('rungrueng', '32000', 'กำไร(ขาดทุน)สะสม', 3, 2, '30000'),
  ('rungrueng', '32100', 'กำไร(ขาดทุน)สะสม', 3, 3, '32000'),
  ('rungrueng', '32101', 'กำไรสะสมที่ยังไม่ได้จัดสรร', 3, 4, '32100'),
  ('rungrueng', '32102', 'กำไร(ขาดทุน)สุทธิประจำปี', 3, 4, '32100'),
  ('rungrueng', '40000', 'รายได้', 4, 1, ''),
  ('rungrueng', '41000', 'รายได้จากการดำเนินงาน', 4, 2, '40000'),
  ('rungrueng', '41100', 'รายได้จากการขาย', 4, 3, '41000'),
  ('rungrueng', '41101', 'รายได้จากการขายสินค้า', 4, 4, '41100'),
  ('rungrueng', '50000', 'ค่าใช้จ่าย', 5, 1, ''),
  ('rungrueng', '51000', 'ต้นทุนขายและบริการ', 5, 2, '50000'),
  ('rungrueng', '51100', 'ต้นทุนขาย', 5, 3, '51000'),
  ('rungrueng', '51101', 'ต้นทุนขาย', 5, 4, '51100'),
  ('rungrueng', '51102', 'ซื้อสินค้า', 5, 4, '51100'),
  ('rungrueng', '52000', 'ค่าใช้จ่ายในการขาย', 5, 2, '50000'),
  ('rungrueng', '53000', 'ค่าใช้จ่ายในการบริหาร', 5, 2, '50000'),
  ('rungrueng', '53100', 'ค่าใช้จ่ายในการบริหาร', 5, 3, '53000'),
  ('rungrueng', '53101', 'เงินเดือนและค่าจ้างฝ่ายบริหาร', 5, 4, '53100');

-- 13. Journals in gl_records (Posted Documents using new Level 4 accounts)
INSERT INTO gl_records (company, kind, id, code, version, payload)
VALUES
  ('01', 'journals', 'jnl-2569-0001', 'GJ2569/001', 1, '{"id": "jnl-2569-0001", "docno": "GJ2569/001", "date": "2026-01-01", "bookcode": "GJ", "branchcode": "00000", "fiscalyear": "2569", "description": "บันทึกยอดยกมาต้นงวด บัญชีเงินฝาก สินค้าคงเหลือ อาคารและอุปกรณ์", "reference": "", "kind": "general", "status": "posted", "lines": [{"accountcode": "11102", "accountname": "เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย", "description": "ยอดยกมาต้นงวด บัญชีเงินฝากกระแสรายวัน", "debit": "500000.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "11301", "accountname": "สินค้าสำเร็จรูป", "description": "ยอดยกมาต้นงวด สินค้าสำเร็จรูปยกมา", "debit": "350000.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "12102", "accountname": "อาคารและสิ่งปลูกสร้าง", "description": "ยอดยกมาต้นงวด อาคารและสิ่งปลูกสร้าง", "debit": "800000.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "22101", "accountname": "เงินกู้ยืมระยะยาวจากสถาบันการเงิน", "description": "ยอดยกมาต้นงวด เงินกู้ยืมระยะยาว", "debit": "0", "credit": "450000.00", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "31101", "accountname": "ทุนหุ้นสามัญ", "description": "ยอดยกมาต้นงวด ทุนจดทะเบียนชำระแล้ว", "debit": "0", "credit": "1000000.00", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "32101", "accountname": "กำไรสะสมที่ยังไม่ได้จัดสรร", "description": "ยอดยกมาต้นงวด กำไรสะสมต้นงวด", "debit": "0", "credit": "200000.00", "departmentcode": "", "projectcode": "", "cashflow": ""}]}'),
  ('01', 'journals', 'jnl-2569-0002', 'AP2569/001', 1, '{"id": "jnl-2569-0002", "docno": "AP2569/001", "date": "2026-01-15", "bookcode": "AP", "branchcode": "00000", "fiscalyear": "2569", "description": "ซื้อปูนซีเมนต์ปอร์ตแลนด์และเหล็กเส้นเป็นเงินเชื่อ", "reference": "INV-1054", "kind": "general", "status": "posted", "lines": [{"accountcode": "51102", "accountname": "ซื้อสินค้า", "description": "ซื้อปูนซีเมนต์ปอร์ตแลนด์และเหล็กเส้นเป็นเงินเชื่อ", "debit": "100000.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "11401", "accountname": "ภาษีซื้อ", "description": "ภาษีซื้อ 7% ใบกำกับภาษีเล่มที่ 1 เลขที่ 1054", "debit": "7000.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "21101", "accountname": "เจ้าหนี้การค้า", "description": "เจ้าหนี้ บริษัท สยามวัสดุก่อสร้าง จำกัด", "debit": "0", "credit": "107000.00", "departmentcode": "", "projectcode": "", "cashflow": ""}]}'),
  ('01', 'journals', 'jnl-2569-0003', 'AR2569/001', 1, '{"id": "jnl-2569-0003", "docno": "AR2569/001", "date": "2026-01-20", "bookcode": "AR", "branchcode": "00000", "fiscalyear": "2569", "description": "ขายวัสดุก่อสร้างตามใบสั่งซื้อ PO-9801", "reference": "IV2569-001", "kind": "general", "status": "posted", "lines": [{"accountcode": "11201", "accountname": "ลูกหนี้การค้า", "description": "ลูกหนี้ ห้างหุ้นส่วนจำกัด บางกอกคอนสตรัคชั่น", "debit": "160500.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "41101", "accountname": "รายได้จากการขายสินค้า", "description": "ขายวัสดุก่อสร้างตามใบสั่งซื้อ PO-9801", "debit": "0", "credit": "150000.00", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "21301", "accountname": "ภาษีขาย", "description": "ภาษีขาย 7% ใบกำกับภาษีเลขที่ IV2569-001", "debit": "0", "credit": "10500.00", "departmentcode": "", "projectcode": "", "cashflow": ""}]}');

-- 14. Balanced Journal Lines in gl_lines (Strictly Lowercase & Valid Constraints)
-- รายการที่ 1: บันทึกยอดยกมาต้นงวด
INSERT INTO gl_lines (
  company, journal_id, line_no, doc_no, entry_date, fiscal_year, 
  book_code, branch_code, department_code, project_code, kind, 
  currency, scale, account_code, account_name, account_type, 
  normal_balance, is_cash, description, cash_flow, debit, credit
) VALUES 
  ('01', 'jnl-2569-0001', 1, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '11102', 'เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย', 'asset', 'debit', true, 'ยอดยกมาต้นงวด บัญชีเงินฝากกระแสรายวัน', '', 500000.00, 0),
  ('01', 'jnl-2569-0001', 2, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '11301', 'สินค้าสำเร็จรูป', 'asset', 'debit', false, 'ยอดยกมาต้นงวด สินค้าสำเร็จรูปยกมา', '', 350000.00, 0),
  ('01', 'jnl-2569-0001', 3, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '12102', 'อาคารและสิ่งปลูกสร้าง', 'asset', 'debit', false, 'ยอดยกมาต้นงวด อาคารและสิ่งปลูกสร้าง', '', 800000.00, 0),
  ('01', 'jnl-2569-0001', 4, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '22101', 'เงินกู้ยืมระยะยาวจากสถาบันการเงิน', 'liability', 'credit', false, 'ยอดยกมาต้นงวด เงินกู้ยืมระยะยาว', '', 0, 450000.00),
  ('01', 'jnl-2569-0001', 5, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '31101', 'ทุนหุ้นสามัญ', 'equity', 'credit', false, 'ยอดยกมาต้นงวด ทุนจดทะเบียนชำระแล้ว', '', 0, 1000000.00),
  ('01', 'jnl-2569-0001', 6, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '32101', 'กำไรสะสมที่ยังไม่ได้จัดสรร', 'equity', 'credit', false, 'ยอดยกมาต้นงวด กำไรสะสมต้นงวด', '', 0, 200000.00);

-- รายการที่ 2: ซื้อสินค้าเป็นเงินเชื่อ
INSERT INTO gl_lines (
  company, journal_id, line_no, doc_no, entry_date, fiscal_year, 
  book_code, branch_code, department_code, project_code, kind, 
  currency, scale, account_code, account_name, account_type, 
  normal_balance, is_cash, description, cash_flow, debit, credit
) VALUES 
  ('01', 'jnl-2569-0002', 1, 'AP2569/001', '2026-01-15', '2569', 'AP', '00000', '', '', 'general', 'THB', 2, '51102', 'ซื้อสินค้า', 'expense', 'debit', false, 'ซื้อปูนซีเมนต์ปอร์ตแลนด์และเหล็กเส้นเป็นเงินเชื่อ', '', 100000.00, 0),
  ('01', 'jnl-2569-0002', 2, 'AP2569/001', '2026-01-15', '2569', 'AP', '00000', '', '', 'general', 'THB', 2, '11401', 'ภาษีซื้อ', 'asset', 'debit', false, 'ภาษีซื้อ 7% ใบกำกับภาษีเล่มที่ 1 เลขที่ 1054', '', 7000.00, 0),
  ('01', 'jnl-2569-0002', 3, 'AP2569/001', '2026-01-15', '2569', 'AP', '00000', '', '', 'general', 'THB', 2, '21101', 'เจ้าหนี้การค้า', 'liability', 'credit', false, 'เจ้าหนี้ บริษัท สยามวัสดุก่อสร้าง จำกัด', '', 0, 107000.00);

-- รายการที่ 3: ขายสินค้าเป็นเงินเชื่อ
INSERT INTO gl_lines (
  company, journal_id, line_no, doc_no, entry_date, fiscal_year, 
  book_code, branch_code, department_code, project_code, kind, 
  currency, scale, account_code, account_name, account_type, 
  normal_balance, is_cash, description, cash_flow, debit, credit
) VALUES 
  ('01', 'jnl-2569-0003', 1, 'AR2569/001', '2026-01-20', '2569', 'AR', '00000', '', '', 'general', 'THB', 2, '11201', 'ลูกหนี้การค้า', 'asset', 'debit', false, 'ลูกหนี้ ห้างหุ้นส่วนจำกัด บางกอกคอนสตรัคชั่น', '', 160500.00, 0),
  ('01', 'jnl-2569-0003', 2, 'AR2569/001', '2026-01-20', '2569', 'AR', '00000', '', '', 'general', 'THB', 2, '41101', 'รายได้จากการขายสินค้า', 'income', 'credit', false, 'ขายวัสดุก่อสร้างตามใบสั่งซื้อ PO-9801', '', 0, 150000.00),
  ('01', 'jnl-2569-0003', 3, 'AR2569/001', '2026-01-20', '2569', 'AR', '00000', '', '', 'general', 'THB', 2, '21301', 'ภาษีขาย', 'liability', 'credit', false, 'ภาษีขาย 7% ใบกำกับภาษีเลขที่ IV2569-001', '', 0, 10500.00);

-- 15. Projection State
INSERT INTO gl_projection_state (company, sequence)
VALUES 
  ('01', 3)
ON CONFLICT (company) DO UPDATE SET sequence = EXCLUDED.sequence;

-- 16. Embed id and version in all gl_records payloads for API parity
UPDATE gl_records 
SET payload = jsonb_set(
  jsonb_set(payload, '{id}', to_jsonb(id)),
  '{version}', to_jsonb(version)
)
WHERE payload->>'id' IS NULL OR payload->>'version' IS NULL;

COMMIT;
