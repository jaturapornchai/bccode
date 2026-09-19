-- Create business_types table
CREATE TABLE IF NOT EXISTS business_types (
  id TEXT PRIMARY KEY,
  holding_code TEXT NOT NULL,
  code TEXT NOT NULL,
  names JSONB NOT NULL DEFAULT '[]'::jsonb,
  is_default BOOLEAN NOT NULL DEFAULT false,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(holding_code, code)
);

-- Create employees table
CREATE TABLE IF NOT EXISTS employees (
  id TEXT PRIMARY KEY,
  holding_code TEXT NOT NULL,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  email TEXT NOT NULL DEFAULT '',
  roles JSONB NOT NULL DEFAULT '[]'::jsonb,
  is_enabled BOOLEAN NOT NULL DEFAULT true,
  is_use_pos BOOLEAN NOT NULL DEFAULT true,
  pin_code TEXT NOT NULL DEFAULT '',
  access_scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
  contact JSONB NOT NULL DEFAULT '{}'::jsonb,
  profile_picture TEXT NOT NULL DEFAULT '',
  profile_picture_thumb TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(holding_code, code)
);

-- Alter role_permissions table
ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS id TEXT;
ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS names JSONB DEFAULT '[]'::jsonb;
ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT now();
ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT now();

-- Seed Business Types for rungrueng & demo
INSERT INTO business_types (id, holding_code, code, names, is_default, is_active)
VALUES
  ('bt-01', 'rungrueng', 'BT01', '[{"code":"th","name":"ค้าปลีก-ค้าส่งวัสดุก่อสร้าง"},{"code":"en","name":"Construction Materials Retail & Wholesale"}]'::jsonb, true, true),
  ('bt-02', 'rungrueng', 'BT02', '[{"code":"th","name":"ร้านอาหารและเครื่องดื่ม"},{"code":"en","name":"Food & Beverage / Cafe"}]'::jsonb, false, true),
  ('bt-03', 'rungrueng', 'BT03', '[{"code":"th","name":"ธุรกิจบริการและรับเหมา"},{"code":"en","name":"Service & Contractor"}]'::jsonb, false, true),
  ('bt-04', 'rungrueng', 'BT04', '[{"code":"th","name":"สำนักงานบัญชีและที่ปรึกษา"},{"code":"en","name":"Accounting Firm & Advisory"}]'::jsonb, false, true),
  ('bt-01-demo', 'demo', 'BT01', '[{"code":"th","name":"ค้าปลีก-ค้าส่งวัสดุก่อสร้าง"},{"code":"en","name":"Construction Materials Retail & Wholesale"}]'::jsonb, true, true),
  ('bt-02-demo', 'demo', 'BT02', '[{"code":"th","name":"ร้านอาหารและเครื่องดื่ม"},{"code":"en","name":"Food & Beverage / Cafe"}]'::jsonb, false, true),
  ('bt-03-demo', 'demo', 'BT03', '[{"code":"th","name":"ธุรกิจบริการและรับเหมา"},{"code":"en","name":"Service & Contractor"}]'::jsonb, false, true)
ON CONFLICT (holding_code, code) DO UPDATE SET
  names = EXCLUDED.names,
  is_default = EXCLUDED.is_default,
  is_active = EXCLUDED.is_active;

-- Seed Role Permissions for rungrueng & demo
INSERT INTO role_permissions (holding_code, role_code, id, names, permissions, is_active)
VALUES
  ('rungrueng', 'OWNER', 'rp-owner', '[{"code":"th","name":"เจ้าของกิจการ"},{"code":"en","name":"Owner"}]'::jsonb, '["*"]'::jsonb, true),
  ('rungrueng', 'ADMIN', 'rp-admin', '[{"code":"th","name":"ผู้ดูแลระบบ"},{"code":"en","name":"Administrator"}]'::jsonb, '["*"]'::jsonb, true),
  ('rungrueng', 'ACCOUNTANT', 'rp-acc', '[{"code":"th","name":"ผู้จัดการฝ่ายบัญชีและการเงิน"},{"code":"en","name":"Accounting Manager"}]'::jsonb, '["/gl","/gl/journal","/gl/chart-of-accounts","/gl/reports","/company","/branch","/activelanguages","/businesstypescreen"]'::jsonb, true),
  ('rungrueng', 'SALES', 'rp-sales', '[{"code":"th","name":"พนักงานขายหน้าร้าน (POS & Sales)"},{"code":"en","name":"Sales & Cashier"}]'::jsonb, '["/menu/sale","/pos","/product_barcode_shelf","/price_history"]'::jsonb, true),
  ('rungrueng', 'WAREHOUSE', 'rp-wh', '[{"code":"th","name":"พนักงานคลังสินค้า (Stock & Warehouse)"},{"code":"en","name":"Warehouse Specialist"}]'::jsonb, '["/menu/inventory","/product_barcode_shelf","/product-barcode"]'::jsonb, true),
  ('rungrueng', 'PURCHASING', 'rp-purch', '[{"code":"th","name":"เจ้าหน้าที่จัดซื้อ (Purchasing)"},{"code":"en","name":"Purchasing Officer"}]'::jsonb, '["/menu/purchase","/creditor"]'::jsonb, true),
  ('demo', 'OWNER', 'rp-owner-demo', '[{"code":"th","name":"เจ้าของกิจการ"},{"code":"en","name":"Owner"}]'::jsonb, '["*"]'::jsonb, true),
  ('demo', 'ADMIN', 'rp-admin-demo', '[{"code":"th","name":"ผู้ดูแลระบบ"},{"code":"en","name":"Administrator"}]'::jsonb, '["*"]'::jsonb, true),
  ('demo', 'ACCOUNTANT', 'rp-acc-demo', '[{"code":"th","name":"ผู้จัดการฝ่ายบัญชีและการเงิน"},{"code":"en","name":"Accounting Manager"}]'::jsonb, '["/gl","/gl/journal","/gl/chart-of-accounts","/gl/reports"]'::jsonb, true),
  ('demo', 'SALES', 'rp-sales-demo', '[{"code":"th","name":"พนักงานขายหน้าร้าน (POS & Sales)"},{"code":"en","name":"Sales & Cashier"}]'::jsonb, '["/menu/sale","/pos"]'::jsonb, true)
ON CONFLICT (holding_code, role_code) DO UPDATE SET
  names = EXCLUDED.names,
  permissions = EXCLUDED.permissions,
  is_active = EXCLUDED.is_active;

-- Seed Employees for rungrueng & demo
INSERT INTO employees (id, holding_code, code, name, email, roles, is_enabled, is_use_pos, pin_code, access_scopes)
VALUES
  ('emp-001', 'rungrueng', 'EMP001', 'สมคิด พาณิชย์เจริญ', 'somkid@rungrueng.co.th', '["OWNER"]'::jsonb, true, true, '1234', '[{"scopetype":"holding","holdingcode":"rungrueng","allbranches":true}]'::jsonb),
  ('emp-002', 'rungrueng', 'EMP002', 'สมศักดิ์ บัญชีการค้า', 'somsak@rungrueng.co.th', '["ACCOUNTANT"]'::jsonb, true, false, '2345', '[{"scopetype":"company","businesscode":"01","allbranches":true},{"scopetype":"company","businesscode":"c01","allbranches":true}]'::jsonb),
  ('emp-003', 'rungrueng', 'EMP003', 'วิภาวรรณ ใจดี', 'wipawan@rungrueng.co.th', '["SALES"]'::jsonb, true, true, '3456', '[{"scopetype":"branch","businesscode":"c01","branchcode":"00000"},{"scopetype":"branch","businesscode":"c01","branchcode":"00001"}]'::jsonb),
  ('emp-004', 'rungrueng', 'EMP004', 'อนุชา ขยันงาน', 'anucha@rungrueng.co.th', '["WAREHOUSE"]'::jsonb, true, true, '4567', '[{"scopetype":"company","businesscode":"c01","allbranches":true}]'::jsonb),
  ('emp-005', 'rungrueng', 'EMP005', 'จตุรพรชัย รัตนปัญญา', 'jaturapornchai@gmail.com', '["OWNER"]'::jsonb, true, true, '9999', '[{"scopetype":"holding","holdingcode":"rungrueng","allbranches":true}]'::jsonb),
  ('emp-001-demo', 'demo', 'EMP001', 'สมคิด พาณิชย์เจริญ', 'somkid@rungrueng.co.th', '["OWNER"]'::jsonb, true, true, '1234', '[{"scopetype":"holding","holdingcode":"demo","allbranches":true}]'::jsonb),
  ('emp-002-demo', 'demo', 'EMP002', 'สมศักดิ์ บัญชีการค้า', 'somsak@rungrueng.co.th', '["ACCOUNTANT"]'::jsonb, true, false, '2345', '[{"scopetype":"holding","holdingcode":"demo","allbranches":true}]'::jsonb)
ON CONFLICT (holding_code, code) DO UPDATE SET
  name = EXCLUDED.name,
  email = EXCLUDED.email,
  roles = EXCLUDED.roles,
  is_enabled = EXCLUDED.is_enabled,
  is_use_pos = EXCLUDED.is_use_pos,
  pin_code = EXCLUDED.pin_code,
  access_scopes = EXCLUDED.access_scopes;

-- Seed Users in users & holding_members
INSERT INTO users (id, username, password_hash, email, full_name, is_active)
VALUES
  ('44444444-4444-4444-4444-444444444444', 'wipawan', '$2a$10$7EqJtq98hPqEX7fNZaFWoOZhB5P1M5t.7.e2p7.G11gZ6.QZ16C2u', 'wipawan@rungrueng.co.th', 'วิภาวรรณ ใจดี', true),
  ('55555555-5555-5555-5555-555555555555', 'anucha', '$2a$10$7EqJtq98hPqEX7fNZaFWoOZhB5P1M5t.7.e2p7.G11gZ6.QZ16C2u', 'anucha@rungrueng.co.th', 'อนุชา ขยันงาน', true)
ON CONFLICT (username) DO NOTHING;

INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes, is_active)
VALUES
  ('rungrueng', '44444444-4444-4444-4444-444444444444', 'user', '["SALES"]'::jsonb, '[{"scopetype":"branch","businesscode":"c01","branchcode":"00000"},{"scopetype":"branch","businesscode":"c01","branchcode":"00001"}]'::jsonb, true),
  ('rungrueng', '55555555-5555-5555-5555-555555555555', 'user', '["WAREHOUSE"]'::jsonb, '[{"scopetype":"company","businesscode":"c01","allbranches":true}]'::jsonb, true),
  ('demo', '44444444-4444-4444-4444-444444444444', 'user', '["SALES"]'::jsonb, '{}'::jsonb, true),
  ('demo', '55555555-5555-5555-5555-555555555555', 'user', '["WAREHOUSE"]'::jsonb, '{}'::jsonb, true)
ON CONFLICT DO NOTHING;

-- Update somsak role and permissions in holding_members
UPDATE holding_members
SET role = 'user', permission_sets = '["ACCOUNTANT"]'::jsonb,
    access_scopes = '[{"scopetype":"company","businesscode":"01","allbranches":true},{"scopetype":"company","businesscode":"c01","allbranches":true}]'::jsonb
WHERE user_id = '22222222-2222-2222-2222-222222222222';
