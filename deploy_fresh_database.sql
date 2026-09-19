-- ============================================================================
-- Fresh Database Reset & Realistic SME Accounting Seed (Pure PostgreSQL)
-- Strictly lowercase identifiers [a-z0-9_] throughout.
-- Authentic Thai accounting chart, fiscal year, and balanced journals.
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

BEGIN;

-- 2. Users (Lower case, Real display names, Password is "Password123456789")
INSERT INTO users (id, username, password_hash, email, phone, full_name, is_active, created_at, updated_at)
VALUES 
  ('11111111-1111-1111-1111-111111111111', 'admin', '$2a$10$sfqaFRidw68/wt5jxKeWoe3lArcQ8mVHbVTyXR36MohgSL.9J0OJ6', 'admin@rungrueng.co.th', '0812345678', 'ผู้ดูแลระบบ', true, now(), now()),
  ('22222222-2222-2222-2222-222222222222', 'somsak', '$2a$10$sfqaFRidw68/wt5jxKeWoe3lArcQ8mVHbVTyXR36MohgSL.9J0OJ6', 'somsak@rungrueng.co.th', '0898765432', 'สมศักดิ์ บัญชีการค้า', true, now(), now()),
  ('33333333-3333-3333-3333-333333333333', 'demo', '$2a$10$sfqaFRidw68/wt5jxKeWoe3lArcQ8mVHbVTyXR36MohgSL.9J0OJ6', 'somkid@rungrueng.co.th', '0865551234', 'สมคิด พาณิชย์เจริญ', true, now(), now());

-- 3. Holdings
INSERT INTO holdings (code, name, tax_id, is_active)
VALUES 
  ('rungrueng', 'กลุ่มกิจการรุ่งเรืองกรุ๊ป', '0105560001234', true),
  ('demo', 'กลุ่มกิจการรุ่งเรืองกรุ๊ป', '0105560001234', true),
  ('bcai_projection', 'กลุ่มกิจการรุ่งเรืองกรุ๊ป', '0105560001234', true);

-- 4. Companies
INSERT INTO companies (holding_code, code, name, tax_id, is_active)
VALUES 
  ('rungrueng', '01', 'สำนักงานใหญ่ รุ่งเรืองค้าวัสดุ', '0105560001234', true),
  ('rungrueng', 'c01', 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด', '0105566123450', true),
  ('rungrueng', 'c02', 'บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด', '0125567001236', true),
  ('rungrueng', 'c03', 'ห้างหุ้นส่วนจำกัด รุ่งเรืองการค้าไทย', '0115565012348', true),
  ('demo', '01', 'สำนักงานใหญ่ รุ่งเรืองค้าวัสดุ', '0105560001234', true),
  ('demo', 'c01', 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด', '0105566123450', true),
  ('demo', 'c02', 'บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด', '0125567001236', true),
  ('demo', 'c03', 'ห้างหุ้นส่วนจำกัด รุ่งเรืองการค้าไทย', '0115565012348', true),
  ('bcai_projection', '01', 'สำนักงานใหญ่ รุ่งเรืองค้าวัสดุ', '0105560001234', true),
  ('bcai_projection', 'c01', 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด', '0105566123450', true),
  ('bcai_projection', 'c02', 'บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด', '0125567001236', true),
  ('bcai_projection', 'c03', 'ห้างหุ้นส่วนจำกัด รุ่งเรืองการค้าไทย', '0115565012348', true);

-- 5. Branches
INSERT INTO branches (holding_code, company_code, code, name, is_headquarters, is_active)
VALUES 
  ('rungrueng', '01', '00000', 'สำนักงานใหญ่', true, true),
  ('rungrueng', 'c01', '00000', 'สำนักงานใหญ่', true, true),
  ('rungrueng', 'c01', '00001', 'สาขาลาดหลุมแก้ว', false, true),
  ('rungrueng', 'c02', '00000', 'สำนักงานใหญ่', true, true),
  ('rungrueng', 'c03', '00000', 'สำนักงานใหญ่', true, true),
  ('demo', '01', '00000', 'สำนักงานใหญ่', true, true),
  ('demo', 'c01', '00000', 'สำนักงานใหญ่', true, true),
  ('demo', 'c01', '00001', 'สาขาลาดหลุมแก้ว', false, true),
  ('demo', 'c02', '00000', 'สำนักงานใหญ่', true, true),
  ('demo', 'c03', '00000', 'สำนักงานใหญ่', true, true),
  ('bcai_projection', '01', '00000', 'สำนักงานใหญ่', true, true),
  ('bcai_projection', 'c01', '00000', 'สำนักงานใหญ่', true, true),
  ('bcai_projection', 'c01', '00001', 'สาขาลาดหลุมแก้ว', false, true),
  ('bcai_projection', 'c02', '00000', 'สำนักงานใหญ่', true, true),
  ('bcai_projection', 'c03', '00000', 'สำนักงานใหญ่', true, true);

-- 6. Holding Memberships
INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes, is_active, created_at)
VALUES 
  ('rungrueng', '11111111-1111-1111-1111-111111111111', 'owner', '["*"]'::jsonb, '{}'::jsonb, true, now()),
  ('rungrueng', '22222222-2222-2222-2222-222222222222', 'admin', '["*"]'::jsonb, '{}'::jsonb, true, now()),
  ('rungrueng', '33333333-3333-3333-3333-333333333333', 'owner', '["*"]'::jsonb, '{}'::jsonb, true, now()),
  ('demo', '11111111-1111-1111-1111-111111111111', 'owner', '["*"]'::jsonb, '{}'::jsonb, true, now()),
  ('demo', '22222222-2222-2222-2222-222222222222', 'admin', '["*"]'::jsonb, '{}'::jsonb, true, now()),
  ('demo', '33333333-3333-3333-3333-333333333333', 'owner', '["*"]'::jsonb, '{}'::jsonb, true, now()),
  ('bcai_projection', '11111111-1111-1111-1111-111111111111', 'owner', '["*"]'::jsonb, '{}'::jsonb, true, now()),
  ('bcai_projection', '22222222-2222-2222-2222-222222222222', 'admin', '["*"]'::jsonb, '{}'::jsonb, true, now()),
  ('bcai_projection', '33333333-3333-3333-3333-333333333333', 'owner', '["*"]'::jsonb, '{}'::jsonb, true, now());

-- 7. Role Permissions (Full Access for all standard roles)
INSERT INTO role_permissions (holding_code, role_code, permissions)
VALUES 
  ('rungrueng', 'owner', '["*"]'::jsonb),
  ('rungrueng', 'admin', '["*"]'::jsonb),
  ('rungrueng', 'accountant', '["*"]'::jsonb),
  ('rungrueng', 'staff', '["*"]'::jsonb),
  ('rungrueng', 'OWNER', '["*"]'::jsonb),
  ('rungrueng', 'ADMIN', '["*"]'::jsonb),
  ('demo', 'owner', '["*"]'::jsonb),
  ('demo', 'admin', '["*"]'::jsonb),
  ('demo', 'accountant', '["*"]'::jsonb),
  ('demo', 'staff', '["*"]'::jsonb),
  ('demo', 'OWNER', '["*"]'::jsonb),
  ('demo', 'ADMIN', '["*"]'::jsonb),
  ('bcai_projection', 'owner', '["*"]'::jsonb),
  ('bcai_projection', 'admin', '["*"]'::jsonb),
  ('bcai_projection', 'accountant', '["*"]'::jsonb),
  ('bcai_projection', 'staff', '["*"]'::jsonb),
  ('bcai_projection', 'OWNER', '["*"]'::jsonb),
  ('bcai_projection', 'ADMIN', '["*"]'::jsonb);

-- 8. Fiscal Year (ปีบัญชี 2569)
INSERT INTO gl_records (company, kind, id, code, version, payload)
VALUES 
  ('01', 'fiscal-years', 'year-2569', '2569', 1, '{"id": "year-2569", "code": "2569", "startdate": "2026-01-01", "enddate": "2026-12-31", "isactive": true, "scale": 2, "profitlossaccount": "33000", "retainedearningsaccount": "32000", "closed": false}');

-- 9. Journal Books (สมุดรายวันมาตรฐาน 6 เล่ม)
INSERT INTO gl_records (company, kind, id, code, version, payload)
VALUES
  ('01', 'journal-books', 'book-gj', 'GJ', 1, '{"code": "GJ", "name": "สมุดรายวันทั่วไป", "names": [{"code": "th", "name": "สมุดรายวันทั่วไป"}], "isactive": true}'),
  ('01', 'journal-books', 'book-pv', 'PV', 1, '{"code": "PV", "name": "สมุดรายวันจ่ายเงิน", "names": [{"code": "th", "name": "สมุดรายวันจ่ายเงิน"}], "isactive": true}'),
  ('01', 'journal-books', 'book-rv', 'RV', 1, '{"code": "RV", "name": "สมุดรายวันรับเงิน", "names": [{"code": "th", "name": "สมุดรายวันรับเงิน"}], "isactive": true}'),
  ('01', 'journal-books', 'book-ap', 'AP', 1, '{"code": "AP", "name": "สมุดรายวันซื้อสินค้า", "names": [{"code": "th", "name": "สมุดรายวันซื้อสินค้า"}], "isactive": true}'),
  ('01', 'journal-books', 'book-ar', 'AR', 1, '{"code": "AR", "name": "สมุดรายวันขายสินค้า", "names": [{"code": "th", "name": "สมุดรายวันขายสินค้า"}], "isactive": true}'),
  ('01', 'journal-books', 'book-jv', 'JV', 1, '{"code": "JV", "name": "สมุดรายวันทั่วไป (JV)", "names": [{"code": "th", "name": "สมุดรายวันทั่วไป (JV)"}], "isactive": true}');

-- 10. Chart of Accounts (ผังบัญชีหลายระดับตามมาตรฐานวิชาชีพบัญชีไทย Level 1-4)
INSERT INTO gl_records (company, kind, id, code, version, payload)
VALUES
  -- หมวด 1 สินทรัพย์ (Assets)
  ('01', 'accounts', 'acc-10000', '10000', 1, '{"accountcode": "10000", "name": "สินทรัพย์", "names": [{"code": "th", "name": "สินทรัพย์"}], "accounttype": "asset", "level": 1, "parentaccountcode": "", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11000-grp', '11000-GRP', 1, '{"accountcode": "11000-GRP", "name": "สินทรัพย์หมุนเวียน", "names": [{"code": "th", "name": "สินทรัพย์หมุนเวียน"}], "accounttype": "asset", "level": 2, "parentaccountcode": "10000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11010', '11010', 1, '{"accountcode": "11010", "name": "เงินสดและรายการเทียบเท่าเงินสด", "names": [{"code": "th", "name": "เงินสดและรายการเทียบเท่าเงินสด"}], "accounttype": "asset", "level": 3, "parentaccountcode": "11000-GRP", "normalbalance": "debit", "allowposting": false, "iscash": true, "isactive": true}'),
  ('01', 'accounts', 'acc-11000', '11000', 1, '{"accountcode": "11000", "name": "เงินสด", "names": [{"code": "th", "name": "เงินสด"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11010", "normalbalance": "debit", "allowposting": true, "iscash": true, "isactive": true}'),
  ('01', 'accounts', 'acc-11100', '11100', 1, '{"accountcode": "11100", "name": "เงินฝากกระแสรายวัน", "names": [{"code": "th", "name": "เงินฝากกระแสรายวัน"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11010", "normalbalance": "debit", "allowposting": true, "iscash": true, "isactive": true}'),
  ('01', 'accounts', 'acc-11200', '11200', 1, '{"accountcode": "11200", "name": "เงินฝากออมทรัพย์", "names": [{"code": "th", "name": "เงินฝากออมทรัพย์"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11010", "normalbalance": "debit", "allowposting": true, "iscash": true, "isactive": true}'),
  ('01', 'accounts', 'acc-11020', '11020', 1, '{"accountcode": "11020", "name": "ลูกหนี้การค้าและลูกหนี้อื่น", "names": [{"code": "th", "name": "ลูกหนี้การค้าและลูกหนี้อื่น"}], "accounttype": "asset", "level": 3, "parentaccountcode": "11000-GRP", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11300', '11300', 1, '{"accountcode": "11300", "name": "ลูกหนี้การค้า", "names": [{"code": "th", "name": "ลูกหนี้การค้า"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11020", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11030', '11030', 1, '{"accountcode": "11030", "name": "สินค้าคงเหลือ", "names": [{"code": "th", "name": "สินค้าคงเหลือ"}], "accounttype": "asset", "level": 3, "parentaccountcode": "11000-GRP", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11400', '11400', 1, '{"accountcode": "11400", "name": "สินค้าคงเหลือสำเร็จรูป", "names": [{"code": "th", "name": "สินค้าคงเหลือสำเร็จรูป"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11030", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11040', '11040', 1, '{"accountcode": "11040", "name": "สินทรัพย์หมุนเวียนอื่น", "names": [{"code": "th", "name": "สินทรัพย์หมุนเวียนอื่น"}], "accounttype": "asset", "level": 3, "parentaccountcode": "11000-GRP", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-11500', '11500', 1, '{"accountcode": "11500", "name": "ภาษีซื้อ", "names": [{"code": "th", "name": "ภาษีซื้อ"}], "accounttype": "asset", "level": 4, "parentaccountcode": "11040", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12000-grp', '12000-GRP', 1, '{"accountcode": "12000-GRP", "name": "สินทรัพย์ไม่หมุนเวียน", "names": [{"code": "th", "name": "สินทรัพย์ไม่หมุนเวียน"}], "accounttype": "asset", "level": 2, "parentaccountcode": "10000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12010', '12010', 1, '{"accountcode": "12010", "name": "ที่ดิน อาคารและอุปกรณ์", "names": [{"code": "th", "name": "ที่ดิน อาคารและอุปกรณ์"}], "accounttype": "asset", "level": 3, "parentaccountcode": "12000-GRP", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12000', '12000', 1, '{"accountcode": "12000", "name": "อาคารและอุปกรณ์", "names": [{"code": "th", "name": "อาคารและอุปกรณ์"}], "accounttype": "asset", "level": 4, "parentaccountcode": "12010", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-12100', '12100', 1, '{"accountcode": "12100", "name": "ค่าเสื่อมราคาสะสม-อาคารและอุปกรณ์", "names": [{"code": "th", "name": "ค่าเสื่อมราคาสะสม-อาคารและอุปกรณ์"}], "accounttype": "asset", "level": 4, "parentaccountcode": "12010", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  
  -- หมวด 2 หนี้สิน (Liabilities)
  ('01', 'accounts', 'acc-20000', '20000', 1, '{"accountcode": "20000", "name": "หนี้สิน", "names": [{"code": "th", "name": "หนี้สิน"}], "accounttype": "liability", "level": 1, "parentaccountcode": "", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21000-grp', '21000-GRP', 1, '{"accountcode": "21000-GRP", "name": "หนี้สินหมุนเวียน", "names": [{"code": "th", "name": "หนี้สินหมุนเวียน"}], "accounttype": "liability", "level": 2, "parentaccountcode": "20000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21000', '21000', 1, '{"accountcode": "21000", "name": "เจ้าหนี้การค้า", "names": [{"code": "th", "name": "เจ้าหนี้การค้า"}], "accounttype": "liability", "level": 3, "parentaccountcode": "21000-GRP", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21100', '21100', 1, '{"accountcode": "21100", "name": "ค่าใช้จ่ายค้างจ่าย", "names": [{"code": "th", "name": "ค่าใช้จ่ายค้างจ่าย"}], "accounttype": "liability", "level": 3, "parentaccountcode": "21000-GRP", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21200', '21200', 1, '{"accountcode": "21200", "name": "ภาษีขาย", "names": [{"code": "th", "name": "ภาษีขาย"}], "accounttype": "liability", "level": 3, "parentaccountcode": "21000-GRP", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-21300', '21300', 1, '{"accountcode": "21300", "name": "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย", "names": [{"code": "th", "name": "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย"}], "accounttype": "liability", "level": 3, "parentaccountcode": "21000-GRP", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-22000-grp', '22000-GRP', 1, '{"accountcode": "22000-GRP", "name": "หนี้สินไม่หมุนเวียน", "names": [{"code": "th", "name": "หนี้สินไม่หมุนเวียน"}], "accounttype": "liability", "level": 2, "parentaccountcode": "20000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-22000', '22000', 1, '{"accountcode": "22000", "name": "เงินกู้ยืมระยะยาวจากสถาบันการเงิน", "names": [{"code": "th", "name": "เงินกู้ยืมระยะยาวจากสถาบันการเงิน"}], "accounttype": "liability", "level": 3, "parentaccountcode": "22000-GRP", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- หมวด 3 ส่วนของเจ้าของ (Equity)
  ('01', 'accounts', 'acc-30000', '30000', 1, '{"accountcode": "30000", "name": "ส่วนของเจ้าของ", "names": [{"code": "th", "name": "ส่วนของเจ้าของ"}], "accounttype": "equity", "level": 1, "parentaccountcode": "", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-31000', '31000', 1, '{"accountcode": "31000", "name": "ทุนจดทะเบียน", "names": [{"code": "th", "name": "ทุนจดทะเบียน"}], "accounttype": "equity", "level": 2, "parentaccountcode": "30000", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-32000', '32000', 1, '{"accountcode": "32000", "name": "กำไรสะสม", "names": [{"code": "th", "name": "กำไรสะสม"}], "accounttype": "equity", "level": 2, "parentaccountcode": "30000", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-33000', '33000', 1, '{"accountcode": "33000", "name": "กำไร(ขาดทุน)สุทธิประจำงวด", "names": [{"code": "th", "name": "กำไร(ขาดทุน)สุทธิประจำงวด"}], "accounttype": "equity", "level": 2, "parentaccountcode": "30000", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- หมวด 4 รายได้ (Income)
  ('01', 'accounts', 'acc-40000', '40000', 1, '{"accountcode": "40000", "name": "รายได้", "names": [{"code": "th", "name": "รายได้"}], "accounttype": "income", "level": 1, "parentaccountcode": "", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-41000-grp', '41000-GRP', 1, '{"accountcode": "41000-GRP", "name": "รายได้จากการขายและบริการ", "names": [{"code": "th", "name": "รายได้จากการขายและบริการ"}], "accounttype": "income", "level": 2, "parentaccountcode": "40000", "normalbalance": "credit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-41000', '41000', 1, '{"accountcode": "41000", "name": "รายได้จากการขายสินค้า", "names": [{"code": "th", "name": "รายได้จากการขายสินค้า"}], "accounttype": "income", "level": 3, "parentaccountcode": "41000-GRP", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-41100', '41100', 1, '{"accountcode": "41100", "name": "รายได้จากการให้บริการ", "names": [{"code": "th", "name": "รายได้จากการให้บริการ"}], "accounttype": "income", "level": 3, "parentaccountcode": "41000-GRP", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-42000', '42000', 1, '{"accountcode": "42000", "name": "รายได้อื่น", "names": [{"code": "th", "name": "รายได้อื่น"}], "accounttype": "income", "level": 2, "parentaccountcode": "40000", "normalbalance": "credit", "allowposting": true, "iscash": false, "isactive": true}'),

  -- หมวด 5 ค่าใช้จ่าย (Expenses)
  ('01', 'accounts', 'acc-50000', '50000', 1, '{"accountcode": "50000", "name": "ค่าใช้จ่าย", "names": [{"code": "th", "name": "ค่าใช้จ่าย"}], "accounttype": "expense", "level": 1, "parentaccountcode": "", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-51000', '51000', 1, '{"accountcode": "51000", "name": "ต้นทุนขายสินค้า", "names": [{"code": "th", "name": "ต้นทุนขายสินค้า"}], "accounttype": "expense", "level": 2, "parentaccountcode": "50000", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-52000-grp', '52000-GRP', 1, '{"accountcode": "52000-GRP", "name": "ค่าใช้จ่ายในการดำเนินงาน", "names": [{"code": "th", "name": "ค่าใช้จ่ายในการดำเนินงาน"}], "accounttype": "expense", "level": 2, "parentaccountcode": "50000", "normalbalance": "debit", "allowposting": false, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-52000', '52000', 1, '{"accountcode": "52000", "name": "เงินเดือนและค่าแรงพนักงาน", "names": [{"code": "th", "name": "เงินเดือนและค่าแรงพนักงาน"}], "accounttype": "expense", "level": 3, "parentaccountcode": "52000-GRP", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-52100', '52100', 1, '{"accountcode": "52100", "name": "ค่าเช่าสำนักงานและคลังสินค้า", "names": [{"code": "th", "name": "ค่าเช่าสำนักงานและคลังสินค้า"}], "accounttype": "expense", "level": 3, "parentaccountcode": "52000-GRP", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-52200', '52200', 1, '{"accountcode": "52200", "name": "ค่าไฟฟ้าและน้ำประปา", "names": [{"code": "th", "name": "ค่าไฟฟ้าและน้ำประปา"}], "accounttype": "expense", "level": 3, "parentaccountcode": "52000-GRP", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-52300', '52300', 1, '{"accountcode": "52300", "name": "ค่าโทรศัพท์และอินเทอร์เน็ต", "names": [{"code": "th", "name": "ค่าโทรศัพท์และอินเทอร์เน็ต"}], "accounttype": "expense", "level": 3, "parentaccountcode": "52000-GRP", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-52400', '52400', 1, '{"accountcode": "52400", "name": "ค่าเสื่อมราคา-อาคารและอุปกรณ์", "names": [{"code": "th", "name": "ค่าเสื่อมราคา-อาคารและอุปกรณ์"}], "accounttype": "expense", "level": 3, "parentaccountcode": "52000-GRP", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}'),
  ('01', 'accounts', 'acc-52500', '52500', 1, '{"accountcode": "52500", "name": "ดอกเบี้ยจ่าย", "names": [{"code": "th", "name": "ดอกเบี้ยจ่าย"}], "accounttype": "expense", "level": 3, "parentaccountcode": "52000-GRP", "normalbalance": "debit", "allowposting": true, "iscash": false, "isactive": true}');

-- 11. Journals in gl_records (Posted Documents)
INSERT INTO gl_records (company, kind, id, code, version, payload)
VALUES
  ('01', 'journals', 'jnl-2569-0001', 'GJ2569/001', 1, '{"id": "jnl-2569-0001", "docno": "GJ2569/001", "date": "2026-01-01", "bookcode": "GJ", "branchcode": "00000", "fiscalyear": "2569", "description": "บันทึกยอดยกมาต้นงวด บัญชีเงินฝาก สินค้าคงเหลือ อาคารและอุปกรณ์", "reference": "", "kind": "general", "status": "posted", "lines": [{"accountcode": "11100", "accountname": "เงินฝากกระแสรายวัน", "description": "ยอดยกมาต้นงวด บัญชีเงินฝากกระแสรายวัน", "debit": "500000.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "11400", "accountname": "สินค้าคงเหลือ", "description": "ยอดยกมาต้นงวด สินค้าคงเหลือยกมา", "debit": "350000.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "12000", "accountname": "อาคารและอุปกรณ์", "description": "ยอดยกมาต้นงวด อาคารและอุปกรณ์", "debit": "800000.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "22000", "accountname": "เงินกู้ยืมระยะยาวจากสถาบันการเงิน", "description": "ยอดยกมาต้นงวด เงินกู้ยืมระยะยาว", "debit": "0", "credit": "450000.00", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "31000", "accountname": "ทุนจดทะเบียน", "description": "ยอดยกมาต้นงวด ทุนจดทะเบียนชำระแล้ว", "debit": "0", "credit": "1000000.00", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "32000", "accountname": "กำไรสะสม", "description": "ยอดยกมาต้นงวด กำไรสะสมต้นงวด", "debit": "0", "credit": "200000.00", "departmentcode": "", "projectcode": "", "cashflow": ""}]}'),
  ('01', 'journals', 'jnl-2569-0002', 'AP2569/001', 1, '{"id": "jnl-2569-0002", "docno": "AP2569/001", "date": "2026-01-15", "bookcode": "AP", "branchcode": "00000", "fiscalyear": "2569", "description": "ซื้อปูนซีเมนต์ปอร์ตแลนด์และเหล็กเส้นเป็นเงินเชื่อ", "reference": "INV-1054", "kind": "general", "status": "posted", "lines": [{"accountcode": "11400", "accountname": "สินค้าคงเหลือ", "description": "ซื้อปูนซีเมนต์ปอร์ตแลนด์และเหล็กเส้นเป็นเงินเชื่อ", "debit": "100000.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "11500", "accountname": "ภาษีซื้อ", "description": "ภาษีซื้อ 7% ใบกำกับภาษีเล่มที่ 1 เลขที่ 1054", "debit": "7000.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "21000", "accountname": "เจ้าหนี้การค้า", "description": "เจ้าหนี้ บริษัท สยามวัสดุก่อสร้าง จำกัด", "debit": "0", "credit": "107000.00", "departmentcode": "", "projectcode": "", "cashflow": ""}]}'),
  ('01', 'journals', 'jnl-2569-0003', 'AR2569/001', 1, '{"id": "jnl-2569-0003", "docno": "AR2569/001", "date": "2026-01-20", "bookcode": "AR", "branchcode": "00000", "fiscalyear": "2569", "description": "ขายวัสดุก่อสร้างตามใบสั่งซื้อ PO-9801", "reference": "IV2569-001", "kind": "general", "status": "posted", "lines": [{"accountcode": "11300", "accountname": "ลูกหนี้การค้า", "description": "ลูกหนี้ ห้างหุ้นส่วนจำกัด บางกอกคอนสตรัคชั่น", "debit": "160500.00", "credit": "0", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "41000", "accountname": "รายได้จากการขายสินค้า", "description": "ขายวัสดุก่อสร้างตามใบสั่งซื้อ PO-9801", "debit": "0", "credit": "150000.00", "departmentcode": "", "projectcode": "", "cashflow": ""}, {"accountcode": "21200", "accountname": "ภาษีขาย", "description": "ภาษีขาย 7% ใบกำกับภาษีเลขที่ IV2569-001", "debit": "0", "credit": "10500.00", "departmentcode": "", "projectcode": "", "cashflow": ""}]}');

-- Clone gl_records to other companies (c01, c02, c03)
INSERT INTO gl_records (company, kind, id, code, version, payload)
SELECT 'c01', kind, id, code, version, payload FROM gl_records WHERE company = '01';

INSERT INTO gl_records (company, kind, id, code, version, payload)
SELECT 'c02', kind, id, code, version, payload FROM gl_records WHERE company = '01';

INSERT INTO gl_records (company, kind, id, code, version, payload)
SELECT 'c03', kind, id, code, version, payload FROM gl_records WHERE company = '01';

-- 12. Balanced Journal Lines in gl_lines (Strictly Lowercase & Valid Constraints)
-- รายการที่ 1: บันทึกยอดยกมาต้นงวด
INSERT INTO gl_lines (
  company, journal_id, line_no, doc_no, entry_date, fiscal_year, 
  book_code, branch_code, department_code, project_code, kind, 
  currency, scale, account_code, account_name, account_type, 
  normal_balance, is_cash, description, cash_flow, debit, credit
) VALUES 
  ('01', 'jnl-2569-0001', 1, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '11100', 'เงินฝากกระแสรายวัน', 'asset', 'debit', true, 'ยอดยกมาต้นงวด บัญชีเงินฝากกระแสรายวัน', '', 500000.00, 0),
  ('01', 'jnl-2569-0001', 2, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '11400', 'สินค้าคงเหลือ', 'asset', 'debit', false, 'ยอดยกมาต้นงวด สินค้าคงเหลือยกมา', '', 350000.00, 0),
  ('01', 'jnl-2569-0001', 3, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '12000', 'อาคารและอุปกรณ์', 'asset', 'debit', false, 'ยอดยกมาต้นงวด อาคารและอุปกรณ์', '', 800000.00, 0),
  ('01', 'jnl-2569-0001', 4, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '22000', 'เงินกู้ยืมระยะยาวจากสถาบันการเงิน', 'liability', 'credit', false, 'ยอดยกมาต้นงวด เงินกู้ยืมระยะยาว', '', 0, 450000.00),
  ('01', 'jnl-2569-0001', 5, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '31000', 'ทุนจดทะเบียน', 'equity', 'credit', false, 'ยอดยกมาต้นงวด ทุนจดทะเบียนชำระแล้ว', '', 0, 1000000.00),
  ('01', 'jnl-2569-0001', 6, 'GJ2569/001', '2026-01-01', '2569', 'GJ', '00000', '', '', 'general', 'THB', 2, '32000', 'กำไรสะสม', 'equity', 'credit', false, 'ยอดยกมาต้นงวด กำไรสะสมต้นงวด', '', 0, 200000.00);

-- รายการที่ 2: ซื้อสินค้าเป็นเงินเชื่อ
INSERT INTO gl_lines (
  company, journal_id, line_no, doc_no, entry_date, fiscal_year, 
  book_code, branch_code, department_code, project_code, kind, 
  currency, scale, account_code, account_name, account_type, 
  normal_balance, is_cash, description, cash_flow, debit, credit
) VALUES 
  ('01', 'jnl-2569-0002', 1, 'AP2569/001', '2026-01-15', '2569', 'AP', '00000', '', '', 'general', 'THB', 2, '11400', 'สินค้าคงเหลือ', 'asset', 'debit', false, 'ซื้อปูนซีเมนต์ปอร์ตแลนด์และเหล็กเส้นเป็นเงินเชื่อ', '', 100000.00, 0),
  ('01', 'jnl-2569-0002', 2, 'AP2569/001', '2026-01-15', '2569', 'AP', '00000', '', '', 'general', 'THB', 2, '11500', 'ภาษีซื้อ', 'asset', 'debit', false, 'ภาษีซื้อ 7% ใบกำกับภาษีเล่มที่ 1 เลขที่ 1054', '', 7000.00, 0),
  ('01', 'jnl-2569-0002', 3, 'AP2569/001', '2026-01-15', '2569', 'AP', '00000', '', '', 'general', 'THB', 2, '21000', 'เจ้าหนี้การค้า', 'liability', 'credit', false, 'เจ้าหนี้ บริษัท สยามวัสดุก่อสร้าง จำกัด', '', 0, 107000.00);

-- รายการที่ 3: ขายสินค้าเป็นเงินเชื่อ
INSERT INTO gl_lines (
  company, journal_id, line_no, doc_no, entry_date, fiscal_year, 
  book_code, branch_code, department_code, project_code, kind, 
  currency, scale, account_code, account_name, account_type, 
  normal_balance, is_cash, description, cash_flow, debit, credit
) VALUES 
  ('01', 'jnl-2569-0003', 1, 'AR2569/001', '2026-01-20', '2569', 'AR', '00000', '', '', 'general', 'THB', 2, '11300', 'ลูกหนี้การค้า', 'asset', 'debit', false, 'ลูกหนี้ ห้างหุ้นส่วนจำกัด บางกอกคอนสตรัคชั่น', '', 160500.00, 0),
  ('01', 'jnl-2569-0003', 2, 'AR2569/001', '2026-01-20', '2569', 'AR', '00000', '', '', 'general', 'THB', 2, '41000', 'รายได้จากการขายสินค้า', 'income', 'credit', false, 'ขายวัสดุก่อสร้างตามใบสั่งซื้อ PO-9801', '', 0, 150000.00),
  ('01', 'jnl-2569-0003', 3, 'AR2569/001', '2026-01-20', '2569', 'AR', '00000', '', '', 'general', 'THB', 2, '21200', 'ภาษีขาย', 'liability', 'credit', false, 'ภาษีขาย 7% ใบกำกับภาษีเลขที่ IV2569-001', '', 0, 10500.00);

-- Clone lines to c01
INSERT INTO gl_lines (
  company, journal_id, line_no, doc_no, entry_date, fiscal_year, 
  book_code, branch_code, department_code, project_code, kind, 
  currency, scale, account_code, account_name, account_type, 
  normal_balance, is_cash, description, cash_flow, debit, credit
)
SELECT 
  'c01', journal_id, line_no, doc_no, entry_date, fiscal_year, 
  book_code, branch_code, department_code, project_code, kind, 
  currency, scale, account_code, account_name, account_type, 
  normal_balance, is_cash, description, cash_flow, debit, credit
FROM gl_lines WHERE company = '01';

-- 13. Projection State
INSERT INTO gl_projection_state (company, sequence)
VALUES 
  ('01', 3),
  ('c01', 3),
  ('c02', 0),
  ('c03', 0)
ON CONFLICT (company) DO UPDATE SET sequence = EXCLUDED.sequence;

COMMIT;
