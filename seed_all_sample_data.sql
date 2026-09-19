-- ============================================================================
-- Complete Sample Data & Alias Migration for BCAI Account (PostgreSQL)
-- Wrapped in a strict transaction (BEGIN ... COMMIT) for ACID data integrity.
-- Supports both '01' (default company code) and 'C01', 'C02', 'C03' (demo companies).
-- ============================================================================

BEGIN;

-- 1. Holdings
INSERT INTO holdings (code, name, tax_id, is_active)
VALUES 
  ('demo', 'บริษัท บีซีเอไอ สาธิต จำกัด (กลุ่มกิจการรุ่งเรือง)', '0105560001234', true),
  ('THAI_HOLDING', 'บริษัท สยามพาณิชย์ กรุ๊ป จำกัด (มหาชน)', '0107565000123', true)
ON CONFLICT (code) DO UPDATE SET is_active = true, name = EXCLUDED.name;

-- 2. Companies
INSERT INTO companies (holding_code, code, name, tax_id, is_active)
VALUES 
  ('demo', '01', 'สำนักงานใหญ่ (01 - Demo Headquarters)', '0105560001234', true),
  ('demo', 'C01', 'บริษัท รุ่งเรืองวัสดุก่อสร้าง จำกัด (C01)', '0105566123450', true),
  ('demo', 'C02', 'บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด (C02)', '0125567001236', true),
  ('demo', 'C03', 'ห้างหุ้นส่วนจำกัด รุ่งเรืองการค้า (C03)', '0115565012348', true),
  ('THAI_HOLDING', '01', 'สำนักงานใหญ่ (01 - Headquarters)', '0107565000123', true),
  ('THAI_HOLDING', 'C01', 'สำนักงานใหญ่ (C01 - Headquarters)', '0107565000123', true)
ON CONFLICT (holding_code, code) DO UPDATE SET is_active = true, name = EXCLUDED.name;

-- 3. Branches
INSERT INTO branches (holding_code, company_code, code, name, is_headquarters, is_active)
VALUES 
  ('demo', '01', '00000', 'สำนักงานใหญ่', true, true),
  ('demo', 'C01', '00000', 'สำนักงานใหญ่', true, true),
  ('demo', 'C01', '00001', 'สาขาลาดหลุมแก้ว', false, true),
  ('demo', 'C02', '00000', 'สำนักงานใหญ่', true, true),
  ('demo', 'C03', '00000', 'สำนักงานใหญ่', true, true),
  ('THAI_HOLDING', '01', '00000', 'สำนักงานใหญ่', true, true),
  ('THAI_HOLDING', 'C01', '00000', 'สำนักงานใหญ่', true, true)
ON CONFLICT (holding_code, company_code, code) DO NOTHING;

-- 4. Alias GL Records: clone from C01 to 01 (for both demo and any 01 queries)
INSERT INTO gl_records(company, kind, id, code, version, payload)
SELECT '01', kind, id, code, version, jsonb_set(payload, '{businesscode}', '"01"'::jsonb)
FROM gl_records WHERE company = 'C01'
ON CONFLICT (company, kind, id) DO UPDATE SET 
  code = EXCLUDED.code, 
  version = EXCLUDED.version, 
  payload = EXCLUDED.payload;

-- 5. Alias GL Lines: clone from C01 to 01
INSERT INTO gl_lines(
  company, journal_id, line_no, doc_no, entry_date, fiscal_year, 
  book_code, branch_code, department_code, project_code, kind, 
  currency, scale, account_code, account_name, account_type, 
  normal_balance, is_cash, description, cash_flow, debit, credit
)
SELECT 
  '01', journal_id, line_no, doc_no, entry_date, fiscal_year, 
  book_code, branch_code, department_code, project_code, kind, 
  currency, scale, account_code, account_name, account_type, 
  normal_balance, is_cash, description, cash_flow, debit, credit
FROM gl_lines WHERE company = 'C01'
ON CONFLICT (company, journal_id, line_no) DO NOTHING;

-- 6. Alias GL Events: clone from C01 to 01
INSERT INTO gl_events(company, id, sequence, event_hash, occurred_at, payload)
SELECT '01', id, sequence, event_hash, occurred_at, jsonb_set(payload, '{businesscode}', '"01"'::jsonb)
FROM gl_events WHERE company = 'C01'
ON CONFLICT (company, id) DO NOTHING;

-- 7. Alias GL Projection State: clone from C01 to 01
INSERT INTO gl_projection_state(company, sequence)
SELECT '01', sequence
FROM gl_projection_state WHERE company = 'C01'
ON CONFLICT (company) DO UPDATE SET sequence = EXCLUDED.sequence;

-- 8. Alias Product Barcodes: clone from C01 to 01
INSERT INTO productbarcode (
  holding_code, businesscode, parid, barcode, names, unitcode, unitnames,
  balanceqty, standvalue, dividevalue, balanceamount, averagecost,
  itemcode, itemtype, materialtype, brandcode, brandnames,
  categorycode, categorynames, classcode, classnames, designcode, designnames,
  gradecode, gradenames, groupcode, groupnames, groupsubonecode, groupsubonenames,
  groupsubtwocode, groupsubtwonames, modelcode, modelnames, patterncode, patternnames, bom
)
SELECT 
  holding_code, '01', parid, barcode, names, unitcode, unitnames,
  balanceqty, standvalue, dividevalue, balanceamount, averagecost,
  itemcode, itemtype, materialtype, brandcode, brandnames,
  categorycode, categorynames, classcode, classnames, designcode, designnames,
  gradecode, gradenames, groupcode, groupnames, groupsubonecode, groupsubonenames,
  groupsubtwocode, groupsubtwonames, modelcode, modelnames, patterncode, patternnames, bom
FROM productbarcode WHERE businesscode = 'C01'
ON CONFLICT (holding_code, businesscode, barcode) DO NOTHING;

-- 9. Legacy Chart of Accounts Table (if present)
INSERT INTO chartofaccounts(
  holdingcode, parid, accountcode, accountname, financial_statements, 
  accountcategory, accountbalancetype, accountgroup, accountlevel, consolidateaccountcode
)
SELECT 
  COALESCE(payload->>'holdingcode', 'demo'),
  COALESCE(payload->>'parentaccountcode', ''),
  code,
  COALESCE(payload->'names'->0->>'name', payload->>'name', ''),
  0,
  CASE 
    WHEN payload->>'accounttype' = 'asset' THEN 1
    WHEN payload->>'accounttype' = 'liability' THEN 2
    WHEN payload->>'accounttype' = 'equity' THEN 3
    WHEN payload->>'accounttype' = 'income' THEN 4
    WHEN payload->>'accounttype' = 'expense' THEN 5
    ELSE 1
  END,
  CASE WHEN payload->>'normalbalance' = 'credit' THEN 2 ELSE 1 END,
  COALESCE(payload->>'accountgroup', ''),
  COALESCE((payload->>'level')::smallint, 1),
  ''
FROM gl_records 
WHERE kind = 'accounts' AND company = 'C01'
ON CONFLICT (holdingcode, accountcode) DO UPDATE SET accountname = EXCLUDED.accountname;

COMMIT;
