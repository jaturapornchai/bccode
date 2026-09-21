-- ============================================================================
-- Thai Standard Chart of Accounts for Holding: rungrueng (3 Levels, 187 Accounts)
-- Compatible with Thai Accounting Standards (TFRS for NPAEs / DBD / RD)
-- Wrapped in transaction for ACID safety
-- ============================================================================

BEGIN;


-- ============================================================================
-- Holding: rungrueng
-- ============================================================================
INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'fiscal-years', '62d7e44f-4ff3-4e81-a02f-c9c726f7022a', '2569', 1, '{"id":"62d7e44f-4ff3-4e81-a02f-c9c726f7022a","code":"2569","kind":"fiscal-years","startdate":"2026-01-01","enddate":"2026-12-31","scale":2,"profitlossaccount":"3222","retainedearningsaccount":"3221","closed":false,"isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', '0bfbf871-5c3b-4aee-ac4f-250dbca88725', 'JV', 1, '{"id":"0bfbf871-5c3b-4aee-ac4f-250dbca88725","code":"JV","kind":"journal-books","name":"สมุดรายวันทั่วไป","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', '2c9931b4-6b78-41e0-aa3a-909248e809af', 'PV', 1, '{"id":"2c9931b4-6b78-41e0-aa3a-909248e809af","code":"PV","kind":"journal-books","name":"สมุดรายวันจ่าย","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', 'c36b6200-9681-489d-a74b-bc731858de04', 'RV', 1, '{"id":"c36b6200-9681-489d-a74b-bc731858de04","code":"RV","kind":"journal-books","name":"สมุดรายวันรับ","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', 'af12bafe-c5e6-4b06-a7fd-c421a8082ef2', 'SV', 1, '{"id":"af12bafe-c5e6-4b06-a7fd-c421a8082ef2","code":"SV","kind":"journal-books","name":"สมุดรายวันซื้อ","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', '1b2f8cf5-63bc-48f6-acb7-2bde29af3c94', 'UV', 1, '{"id":"1b2f8cf5-63bc-48f6-acb7-2bde29af3c94","code":"UV","kind":"journal-books","name":"สมุดรายวันขาย","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', 'ffe79b19-a7a9-4e0f-a4e3-56b0f3df05ed', 'GJ', 1, '{"id":"ffe79b19-a7a9-4e0f-a4e3-56b0f3df05ed","code":"GJ","kind":"journal-books","name":"สมุดรายวันทั่วไป (GJ)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', '5971c812-11dd-47ac-a81a-39f11f014a57', 'AP', 1, '{"id":"5971c812-11dd-47ac-a81a-39f11f014a57","code":"AP","kind":"journal-books","name":"สมุดรายวันซื้อเชื่อ (AP)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', '3842f561-aaa6-47b6-a64a-314e023a112d', 'AR', 1, '{"id":"3842f561-aaa6-47b6-a64a-314e023a112d","code":"AR","kind":"journal-books","name":"สมุดรายวันขายเชื่อ (AR)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'account-groups', '8391e933-d5ab-48c2-a2ef-fc48e4f9ce06', 'BM69-A', 1, '{"id":"8391e933-d5ab-48c2-a2ef-fc48e4f9ce06","code":"BM69-A","kind":"account-groups","name":"สินทรัพย์","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'account-groups', '25034456-6364-4eb8-a3bb-8012980fd401', 'BM69-L', 1, '{"id":"25034456-6364-4eb8-a3bb-8012980fd401","code":"BM69-L","kind":"account-groups","name":"หนี้สิน","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'account-groups', '99df7064-3ca5-4cac-a48f-c0ecb575a60a', 'BM69-E', 1, '{"id":"99df7064-3ca5-4cac-a48f-c0ecb575a60a","code":"BM69-E","kind":"account-groups","name":"ส่วนของเจ้าของ","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'account-groups', '4b1c6c1c-60d2-4f1e-a502-1dfbb33006f3', 'BM69-R', 1, '{"id":"4b1c6c1c-60d2-4f1e-a502-1dfbb33006f3","code":"BM69-R","kind":"account-groups","name":"รายได้","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'account-groups', 'bcac0fb6-664d-4b13-a9b3-466a3a5b9442', 'BM69-X', 1, '{"id":"bcac0fb6-664d-4b13-a9b3-466a3a5b9442","code":"BM69-X","kind":"account-groups","name":"ค่าใช้จ่าย","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"rungrueng","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;


-- Accounts for Holding: rungrueng, Company: 01
INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '6282ad7f-3fa5-4ca9-aae8-2a29cf446e93', '1000', 1, '{"id":"6282ad7f-3fa5-4ca9-aae8-2a29cf446e93","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1000","names":[{"code":"th","name":"สินทรัพย์"},{"code":"en","name":"Assets"}],"accounttype":"asset","parentaccountcode":null,"normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd28b5e28-c151-4b8c-ad59-cb73601155fd', '1100', 1, '{"id":"d28b5e28-c151-4b8c-ad59-cb73601155fd","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1100","names":[{"code":"th","name":"สินทรัพย์หมุนเวียน"},{"code":"en","name":"Current Assets"}],"accounttype":"asset","parentaccountcode":"1000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'bb124db0-9485-458f-ad5e-9957a9c612c6', '1111', 1, '{"id":"bb124db0-9485-458f-ad5e-9957a9c612c6","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1111","names":[{"code":"th","name":"เงินสดในมือ"},{"code":"en","name":"Cash on Hand"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1a714d8c-de56-4f1e-a054-bb694eace06a', '1112', 1, '{"id":"1a714d8c-de56-4f1e-a054-bb694eace06a","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1112","names":[{"code":"th","name":"เงินสดย่อย - สำนักงานใหญ่"},{"code":"en","name":"Petty Cash - Head Office"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5a58cea3-7325-468a-ac0a-ea4e56087db1', '1113', 1, '{"id":"5a58cea3-7325-468a-ac0a-ea4e56087db1","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1113","names":[{"code":"th","name":"เงินสดย่อย - ฝ่ายปฏิบัติการและสาขา"},{"code":"en","name":"Petty Cash - Operations & Branch"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8b016836-12fe-4041-ad9d-4ad17cb446d3', '1121', 1, '{"id":"8b016836-12fe-4041-ad9d-4ad17cb446d3","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1121","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย"},{"code":"en","name":"Cash at Bank - Current KBANK"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '60db0dfb-ea3b-4993-ae87-8949bfd2e8f1', '1122', 1, '{"id":"60db0dfb-ea3b-4993-ae87-8949bfd2e8f1","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1122","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารกรุงเทพ"},{"code":"en","name":"Cash at Bank - Current BBL"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '6468362e-4c28-4c18-ada7-236861c1e96f', '1123', 1, '{"id":"6468362e-4c28-4c18-ada7-236861c1e96f","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1123","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารไทยพาณิชย์"},{"code":"en","name":"Cash at Bank - Current SCB"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '91295325-c53e-460e-abbe-67ba89366451', '1124', 1, '{"id":"91295325-c53e-460e-abbe-67ba89366451","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1124","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารกสิกรไทย"},{"code":"en","name":"Cash at Bank - Savings KBANK"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0b97827b-77cf-4a8c-ae95-0530de77c143', '1125', 1, '{"id":"0b97827b-77cf-4a8c-ae95-0530de77c143","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1125","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารกรุงเทพ"},{"code":"en","name":"Cash at Bank - Savings BBL"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7c07a3a6-4d18-4f49-ac35-8a9c6dbeccd4', '1126', 1, '{"id":"7c07a3a6-4d18-4f49-ac35-8a9c6dbeccd4","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1126","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารไทยพาณิชย์"},{"code":"en","name":"Cash at Bank - Savings SCB"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '4675d874-dc17-4096-ad66-a5dcd2d63ba2', '1127', 1, '{"id":"4675d874-dc17-4096-ad66-a5dcd2d63ba2","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1127","names":[{"code":"th","name":"เงินฝากประจำระยะสั้น (ไม่เกิน 3 เดือน)"},{"code":"en","name":"Short-term Fixed Deposit"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '53c95fae-069f-4df9-a14b-184bbd2916fd', '1131', 1, '{"id":"53c95fae-069f-4df9-a14b-184bbd2916fd","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1131","names":[{"code":"th","name":"ลูกหนี้การค้า - ในประเทศ"},{"code":"en","name":"Trade Accounts Receivable - Domestic"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'dd633f4f-5e59-4a3f-aa78-a9a1da31e173', '1132', 1, '{"id":"dd633f4f-5e59-4a3f-aa78-a9a1da31e173","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1132","names":[{"code":"th","name":"ลูกหนี้การค้า - ต่างประเทศ"},{"code":"en","name":"Trade Accounts Receivable - Overseas"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b5e64b92-4cdd-4234-a7d0-f2f250d54923', '1133', 1, '{"id":"b5e64b92-4cdd-4234-a7d0-f2f250d54923","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1133","names":[{"code":"th","name":"ตั๋วเงินรับการค้า"},{"code":"en","name":"Notes Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c2745539-907b-402d-af3a-60b8407637b8', '1134', 1, '{"id":"c2745539-907b-402d-af3a-60b8407637b8","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1134","names":[{"code":"th","name":"เช็ครับลงวันที่ล่วงหน้า"},{"code":"en","name":"Post-dated Cheques Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a891f54a-9f11-449c-abb4-b7f02db4c4ad', '1138', 1, '{"id":"a891f54a-9f11-449c-abb4-b7f02db4c4ad","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1138","names":[{"code":"th","name":"ลูกหนี้อื่นและเงินยืมทดรอง"},{"code":"en","name":"Other Receivables & Advances"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '01bd4f07-f251-440b-a4f1-8ce5f0f838f1', '1139', 1, '{"id":"01bd4f07-f251-440b-a4f1-8ce5f0f838f1","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1139","names":[{"code":"th","name":"ค่าเผื่อผลขาดทุนด้านเครดิตที่คาดว่าจะเกิดขึ้น"},{"code":"en","name":"Allowance for Expected Credit Losses"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'fefa092f-ad1f-407a-a196-0031c528baa6', '1141', 1, '{"id":"fefa092f-ad1f-407a-a196-0031c528baa6","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1141","names":[{"code":"th","name":"สินค้าสำเร็จรูป"},{"code":"en","name":"Finished Goods"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1d8acf9f-7324-4561-a12b-fbd18527521b', '1142', 1, '{"id":"1d8acf9f-7324-4561-a12b-fbd18527521b","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1142","names":[{"code":"th","name":"สินค้าระหว่างทำ / งานระหว่างทำ"},{"code":"en","name":"Work in Process"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '270f8d8e-b8d0-4a1f-a19f-012a74434b93', '1143', 1, '{"id":"270f8d8e-b8d0-4a1f-a19f-012a74434b93","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1143","names":[{"code":"th","name":"วัตถุดิบและส่วนประกอบ"},{"code":"en","name":"Raw Materials & Components"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '83854a26-98dd-4c50-ad87-d454930cb068', '1144', 1, '{"id":"83854a26-98dd-4c50-ad87-d454930cb068","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1144","names":[{"code":"th","name":"วัสดุสิ้นเปลืองและบรรจุภัณฑ์"},{"code":"en","name":"Factory Supplies & Packaging"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '793065e7-4013-4704-a81e-37670633a20d', '1145', 1, '{"id":"793065e7-4013-4704-a81e-37670633a20d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1145","names":[{"code":"th","name":"สินค้าระหว่างทาง"},{"code":"en","name":"Goods in Transit"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5fce1f96-81fd-498c-a90f-6241adffc09e', '1149', 1, '{"id":"5fce1f96-81fd-498c-a90f-6241adffc09e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1149","names":[{"code":"th","name":"ค่าเผื่อการลดมูลค่าสินค้าคงเหลือ"},{"code":"en","name":"Allowance for Inventory Devaluation"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a385ddb3-41ce-4348-ab9d-7bdbaa02d18f', '1151', 1, '{"id":"a385ddb3-41ce-4348-ab9d-7bdbaa02d18f","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1151","names":[{"code":"th","name":"ภาษีซื้อ"},{"code":"en","name":"Input VAT"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'dea57fd5-b027-4d60-a13a-0e5e82b859a7', '1152', 1, '{"id":"dea57fd5-b027-4d60-a13a-0e5e82b859a7","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1152","names":[{"code":"th","name":"ภาษีซื้อยังไม่ถึงกำหนดชำระ"},{"code":"en","name":"Undue Input VAT"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a17ab6ca-075e-4c98-a430-b1187019b8f4', '1153', 1, '{"id":"a17ab6ca-075e-4c98-a430-b1187019b8f4","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1153","names":[{"code":"th","name":"ภาษีเงินได้ถูกหัก ณ ที่จ่าย (WHT ถูกหัก)"},{"code":"en","name":"Withholding Tax Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e7c013ff-fd3d-4d87-a77a-5df63b5deb40', '1154', 1, '{"id":"e7c013ff-fd3d-4d87-a77a-5df63b5deb40","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1154","names":[{"code":"th","name":"ภาษีเงินได้นิติบุคคลจ่ายล่วงหน้า (ภ.ง.ด.51)"},{"code":"en","name":"Prepaid Corporate Income Tax"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '3aaa0f00-2b32-42c8-a8a4-70be5a332666', '1161', 1, '{"id":"3aaa0f00-2b32-42c8-a8a4-70be5a332666","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1161","names":[{"code":"th","name":"ค่าเบี้ยประกันภัยจ่ายล่วงหน้า"},{"code":"en","name":"Prepaid Insurance"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1f94895c-952f-4539-a500-7839c20613f6', '1162', 1, '{"id":"1f94895c-952f-4539-a500-7839c20613f6","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1162","names":[{"code":"th","name":"ค่าเช่าจ่ายล่วงหน้า"},{"code":"en","name":"Prepaid Rent"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8ba15e69-a183-45a3-ad3a-af8d6dc5d45c', '1163', 1, '{"id":"8ba15e69-a183-45a3-ad3a-af8d6dc5d45c","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1163","names":[{"code":"th","name":"เงินมัดจำค่าสินค้าและบริการล่วงหน้า"},{"code":"en","name":"Advance Payments for Goods & Services"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b333479b-4209-417b-a0ad-a6e5b0532795', '1169', 1, '{"id":"b333479b-4209-417b-a0ad-a6e5b0532795","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1169","names":[{"code":"th","name":"สินทรัพย์หมุนเวียนอื่น"},{"code":"en","name":"Other Current Assets"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'cc4c0a1c-11e3-4205-ad11-782a45f15e36', '1200', 1, '{"id":"cc4c0a1c-11e3-4205-ad11-782a45f15e36","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1200","names":[{"code":"th","name":"สินทรัพย์ไม่หมุนเวียน"},{"code":"en","name":"Non-Current Assets"}],"accounttype":"asset","parentaccountcode":"1000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5ae289a8-57f4-450b-a775-4ffc8da35ba8', '1211', 1, '{"id":"5ae289a8-57f4-450b-a775-4ffc8da35ba8","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1211","names":[{"code":"th","name":"ที่ดิน"},{"code":"en","name":"Land"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f9234105-6058-4b33-a524-6980dc3c961b', '1212', 1, '{"id":"f9234105-6058-4b33-a524-6980dc3c961b","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1212","names":[{"code":"th","name":"ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Land Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e446052e-0b32-4d30-a110-23e751c76b9d', '1213', 1, '{"id":"e446052e-0b32-4d30-a110-23e751c76b9d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1213","names":[{"code":"th","name":"อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Buildings & Constructions"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd1711531-c7f4-4c7f-aa0e-be3f874d591d', '1214', 1, '{"id":"d1711531-c7f4-4c7f-aa0e-be3f874d591d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1214","names":[{"code":"th","name":"ส่วนปรับปรุงอาคารและระบบสาธารณูปโภค"},{"code":"en","name":"Building Improvements & Systems"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'ee459eda-1b59-4e51-a91d-5ca58539e44b', '1221', 1, '{"id":"ee459eda-1b59-4e51-a91d-5ca58539e44b","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1221","names":[{"code":"th","name":"เครื่องจักรและอุปกรณ์โรงงาน"},{"code":"en","name":"Machinery & Factory Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '88e7e55c-e960-4520-ada5-4a0317161644', '1222', 1, '{"id":"88e7e55c-e960-4520-ada5-4a0317161644","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1222","names":[{"code":"th","name":"เครื่องตกแต่ง ติดตั้ง และเฟอร์นิเจอร์"},{"code":"en","name":"Furniture & Fixtures"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0a43d968-7ded-4b47-a8bf-48bcb2706762', '1223', 1, '{"id":"0a43d968-7ded-4b47-a8bf-48bcb2706762","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1223","names":[{"code":"th","name":"เครื่องใช้และอุปกรณ์สำนักงาน"},{"code":"en","name":"Office Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1c78d609-64dd-4008-a114-c05ba727ceee', '1224', 1, '{"id":"1c78d609-64dd-4008-a114-c05ba727ceee","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1224","names":[{"code":"th","name":"คอมพิวเตอร์และอุปกรณ์ประมวลผล"},{"code":"en","name":"Computer & Hardware Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f964a333-f248-486b-ae10-521c3cc4fd34', '1231', 1, '{"id":"f964a333-f248-486b-ae10-521c3cc4fd34","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1231","names":[{"code":"th","name":"ยานพาหนะและรถบรรทุกขนส่ง"},{"code":"en","name":"Vehicles & Transport Trucks"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '19f1799c-1a4a-446e-abe5-6eb5700aeee8', '1241', 1, '{"id":"19f1799c-1a4a-446e-abe5-6eb5700aeee8","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1241","names":[{"code":"th","name":"งานระหว่างก่อสร้างและติดตั้งเครื่องจักร"},{"code":"en","name":"Construction & Installation in Progress"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'dd66ae3f-e9e8-4cd1-ad9d-3482fe204e7c', '1281', 1, '{"id":"dd66ae3f-e9e8-4cd1-ad9d-3482fe204e7c","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1281","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Acc. Dep. - Land Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '53c17a81-9dd7-40b2-a28c-2b5c58945894', '1282', 1, '{"id":"53c17a81-9dd7-40b2-a28c-2b5c58945894","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1282","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Acc. Dep. - Buildings"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5818c543-7896-486e-adab-7d80a1b7020b', '1283', 1, '{"id":"5818c543-7896-486e-adab-7d80a1b7020b","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1283","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ส่วนปรับปรุงอาคาร"},{"code":"en","name":"Acc. Dep. - Building Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '60460deb-0187-45e1-ac82-894a6f8b3dc4', '1284', 1, '{"id":"60460deb-0187-45e1-ac82-894a6f8b3dc4","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1284","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องจักรและอุปกรณ์"},{"code":"en","name":"Acc. Dep. - Machinery & Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e1fb3891-81e3-4fe0-ae03-22e7b5f499cc', '1285', 1, '{"id":"e1fb3891-81e3-4fe0-ae03-22e7b5f499cc","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1285","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องตกแต่งและติดตั้ง"},{"code":"en","name":"Acc. Dep. - Furniture & Fixtures"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9806d8f7-144f-4784-a0a2-d3881035273e', '1286', 1, '{"id":"9806d8f7-144f-4784-a0a2-d3881035273e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1286","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องใช้สำนักงาน"},{"code":"en","name":"Acc. Dep. - Office Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b39a3ecb-540c-4613-a570-7ee98def1e57', '1287', 1, '{"id":"b39a3ecb-540c-4613-a570-7ee98def1e57","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1287","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - คอมพิวเตอร์และอุปกรณ์"},{"code":"en","name":"Acc. Dep. - Computer & Hardware"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f6902371-ea52-421c-ac1e-0f8d08885077', '1288', 1, '{"id":"f6902371-ea52-421c-ac1e-0f8d08885077","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1288","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ยานพาหนะ"},{"code":"en","name":"Acc. Dep. - Vehicles"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9e3c988d-b357-4896-a22e-4dc881fa22d0', '1291', 1, '{"id":"9e3c988d-b357-4896-a22e-4dc881fa22d0","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1291","names":[{"code":"th","name":"โปรแกรมคอมพิวเตอร์และสิทธิการใช้งาน"},{"code":"en","name":"Software & Licenses"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '6887213a-6bd2-4e98-a10a-4cb61d1a9b5d', '1292', 1, '{"id":"6887213a-6bd2-4e98-a10a-4cb61d1a9b5d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1292","names":[{"code":"th","name":"ค่าตัดจำหน่ายสะสม - โปรแกรมคอมพิวเตอร์"},{"code":"en","name":"Acc. Amortization - Software"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '387cfb8d-f358-4bd6-afbf-dd491c46ae26', '1295', 1, '{"id":"387cfb8d-f358-4bd6-afbf-dd491c46ae26","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1295","names":[{"code":"th","name":"เงินประกันและเงินมัดจำระยะยาว"},{"code":"en","name":"Long-term Deposits & Guarantees"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '3ff4ce00-d07d-4c88-ad0f-88718b145c16', '2000', 1, '{"id":"3ff4ce00-d07d-4c88-ad0f-88718b145c16","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2000","names":[{"code":"th","name":"หนี้สิน"},{"code":"en","name":"Liabilities"}],"accounttype":"liability","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'fe79f190-fd87-4ff7-a00a-63dfdef293dd', '2100', 1, '{"id":"fe79f190-fd87-4ff7-a00a-63dfdef293dd","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2100","names":[{"code":"th","name":"หนี้สินหมุนเวียน"},{"code":"en","name":"Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0b060df0-c203-4416-aa1e-d4a1eb8bac7a', '2111', 1, '{"id":"0b060df0-c203-4416-aa1e-d4a1eb8bac7a","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2111","names":[{"code":"th","name":"เจ้าหนี้การค้า - ในประเทศ"},{"code":"en","name":"Trade Accounts Payable - Domestic"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e60d022f-e710-4891-abfd-fa6e7f5b8ba8', '2112', 1, '{"id":"e60d022f-e710-4891-abfd-fa6e7f5b8ba8","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2112","names":[{"code":"th","name":"เจ้าหนี้การค้า - ต่างประเทศ"},{"code":"en","name":"Trade Accounts Payable - Overseas"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0d83b683-820c-43dd-acd6-9c4c4074798e', '2113', 1, '{"id":"0d83b683-820c-43dd-acd6-9c4c4074798e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2113","names":[{"code":"th","name":"ตั๋วเงินจ่ายการค้า"},{"code":"en","name":"Notes Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '053a0cfd-8cd1-48f0-a0e9-0851f0cbcf4d', '2114', 1, '{"id":"053a0cfd-8cd1-48f0-a0e9-0851f0cbcf4d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2114","names":[{"code":"th","name":"เช็คจ่ายลงวันที่ล่วงหน้า"},{"code":"en","name":"Post-dated Cheques Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '4f109e32-e4fd-44ca-a346-c5b7078da2b2', '2121', 1, '{"id":"4f109e32-e4fd-44ca-a346-c5b7078da2b2","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2121","names":[{"code":"th","name":"เจ้าหนี้อื่นและเงินทดรองรับ"},{"code":"en","name":"Other Payables & Advances"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0342455c-7b8c-4458-a035-d59ff37a93c4', '2122', 1, '{"id":"0342455c-7b8c-4458-a035-d59ff37a93c4","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2122","names":[{"code":"th","name":"เจ้าหนี้กรมสรรพากร"},{"code":"en","name":"Revenue Department Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7ae5333f-5d4b-4e41-a229-ab13407b1d40', '2131', 1, '{"id":"7ae5333f-5d4b-4e41-a229-ab13407b1d40","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2131","names":[{"code":"th","name":"ภาษีขาย"},{"code":"en","name":"Output VAT"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '80e8fce8-aeeb-4c30-af86-f8a4e0a85e2d', '2132', 1, '{"id":"80e8fce8-aeeb-4c30-af86-f8a4e0a85e2d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2132","names":[{"code":"th","name":"ภาษีขายยังไม่ถึงกำหนดชำระ"},{"code":"en","name":"Undue Output VAT"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5fa97848-7120-4628-a71f-40355758d396', '2141', 1, '{"id":"5fa97848-7120-4628-a71f-40355758d396","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2141","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.1 (เงินเดือน)"},{"code":"en","name":"WHT Payable - P.N.D.1 (Salaries)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '26bf6747-0e63-495d-ad7c-1d2372dd1298', '2142', 1, '{"id":"26bf6747-0e63-495d-ad7c-1d2372dd1298","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2142","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.3 (บุคคลธรรมดา)"},{"code":"en","name":"WHT Payable - P.N.D.3 (Individuals)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '4a14c505-f93b-4bd7-a082-cfa1cf40c7b2', '2143', 1, '{"id":"4a14c505-f93b-4bd7-a082-cfa1cf40c7b2","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2143","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.53 (นิติบุคคล)"},{"code":"en","name":"WHT Payable - P.N.D.53 (Corporations)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c453db76-6e2d-431a-a789-92bcc9f5371d', '2144', 1, '{"id":"c453db76-6e2d-431a-a789-92bcc9f5371d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2144","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.2 (ดอกเบี้ย/ปันผล)"},{"code":"en","name":"WHT Payable - P.N.D.2"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd900d527-45e3-49a1-a677-fb62a3dc3337', '2145', 1, '{"id":"d900d527-45e3-49a1-a677-fb62a3dc3337","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2145","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.54 (ส่งไปต่างประเทศ)"},{"code":"en","name":"WHT Payable - P.N.D.54 (Overseas)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a3a418f7-2b9c-4b8e-a61c-52e737d6434d', '2151', 1, '{"id":"a3a418f7-2b9c-4b8e-a61c-52e737d6434d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2151","names":[{"code":"th","name":"เงินสมทบกองทุนประกันสังคมค้างจ่าย (ลูกจ้าง+นายจ้าง)"},{"code":"en","name":"Social Security Fund Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '94ae90fc-43c8-4c3a-aac2-6d03405bd849', '2152', 1, '{"id":"94ae90fc-43c8-4c3a-aac2-6d03405bd849","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2152","names":[{"code":"th","name":"เงินเดือนและค่าจ้างค้างจ่าย"},{"code":"en","name":"Accrued Salaries & Wages"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1127e76a-4377-4423-a024-3a41ba2f9501', '2153', 1, '{"id":"1127e76a-4377-4423-a024-3a41ba2f9501","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2153","names":[{"code":"th","name":"ค่าเช่าค้างจ่าย"},{"code":"en","name":"Accrued Rent Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8122092b-8f7a-4c68-ad68-b803a27c32bb', '2154', 1, '{"id":"8122092b-8f7a-4c68-ad68-b803a27c32bb","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2154","names":[{"code":"th","name":"ค่าน้ำประปาและค่าไฟฟ้าค้างจ่าย"},{"code":"en","name":"Accrued Utilities Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0c042312-f52d-4358-a0d4-748c13c28225', '2155', 1, '{"id":"0c042312-f52d-4358-a0d4-748c13c28225","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2155","names":[{"code":"th","name":"ค่าโทรศัพท์และอินเทอร์เน็ตค้างจ่าย"},{"code":"en","name":"Accrued Telephone & Internet"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7b64250f-e15e-49d9-a36e-9bc93bbb236e', '2156', 1, '{"id":"7b64250f-e15e-49d9-a36e-9bc93bbb236e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2156","names":[{"code":"th","name":"ค่าสอบบัญชีและทำบัญชีค้างจ่าย"},{"code":"en","name":"Accrued Audit & Accounting Fees"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '3768b3dc-e7f8-409b-a686-3369ce67d396', '2157', 1, '{"id":"3768b3dc-e7f8-409b-a686-3369ce67d396","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2157","names":[{"code":"th","name":"ดอกเบี้ยค้างจ่าย"},{"code":"en","name":"Accrued Interest Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'df5f73be-1348-40f4-a8ce-f7fa4bed6c96', '2158', 1, '{"id":"df5f73be-1348-40f4-a8ce-f7fa4bed6c96","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2158","names":[{"code":"th","name":"โบนัสพนักงานค้างจ่าย"},{"code":"en","name":"Accrued Staff Bonuses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '4a2859e6-eb5e-4192-a26e-dbb46e3d0512', '2159', 1, '{"id":"4a2859e6-eb5e-4192-a26e-dbb46e3d0512","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2159","names":[{"code":"th","name":"ค่าใช้จ่ายค้างจ่ายอื่น"},{"code":"en","name":"Other Accrued Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7e54fe23-2056-431a-a490-7baf31d320db', '2161', 1, '{"id":"7e54fe23-2056-431a-a490-7baf31d320db","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2161","names":[{"code":"th","name":"เงินรับล่วงหน้าค่าสินค้าและบริการจากลูกค้า"},{"code":"en","name":"Advances Received from Customers"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '06fe8650-01cb-4810-a819-e64908cb22a8', '2171', 1, '{"id":"06fe8650-01cb-4810-a819-e64908cb22a8","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2171","names":[{"code":"th","name":"เงินเบิกเกินบัญชีธนาคาร (O/D)"},{"code":"en","name":"Bank Overdrafts (O/D)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5105fef7-3f40-49c8-ab34-5fa594db7934', '2172', 1, '{"id":"5105fef7-3f40-49c8-ab34-5fa594db7934","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2172","names":[{"code":"th","name":"เงินกู้ยืมระยะสั้นจากสถาบันการเงิน"},{"code":"en","name":"Short-term Borrowings from Banks"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '468e1b92-26ca-49e7-aa45-989b208778f9', '2173', 1, '{"id":"468e1b92-26ca-49e7-aa45-989b208778f9","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2173","names":[{"code":"th","name":"เงินกู้ยืมระยะสั้นจากกรรมการหรือบุคคลที่เกี่ยวข้องกัน"},{"code":"en","name":"Short-term Loans from Directors"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '550dd3c5-58b7-4040-a80a-6f182313ba4b', '2181', 1, '{"id":"550dd3c5-58b7-4040-a80a-6f182313ba4b","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2181","names":[{"code":"th","name":"ส่วนของหนี้สินระยะยาวที่ถึงกำหนดชำระภายในหนึ่งปี"},{"code":"en","name":"Current Portion of Long-term Debt"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '60fa30c6-325b-4c9d-a772-5d88dee87fbc', '2200', 1, '{"id":"60fa30c6-325b-4c9d-a772-5d88dee87fbc","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2200","names":[{"code":"th","name":"หนี้สินไม่หมุนเวียน"},{"code":"en","name":"Non-Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f1f67d12-4643-49cc-a578-85c96042fd64', '2211', 1, '{"id":"f1f67d12-4643-49cc-a578-85c96042fd64","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2211","names":[{"code":"th","name":"เงินกู้ยืมระยะยาวจากสถาบันการเงิน"},{"code":"en","name":"Long-term Bank Loans"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '11bcf208-cf49-4f88-afbd-feb678fae77b', '2212', 1, '{"id":"11bcf208-cf49-4f88-afbd-feb678fae77b","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2212","names":[{"code":"th","name":"หนี้สินตามสัญญาเช่าทางการเงินระยะยาว"},{"code":"en","name":"Long-term Financial Lease Liabilities"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '734a38b3-bd43-45e8-a27b-b33e77da8223', '2221', 1, '{"id":"734a38b3-bd43-45e8-a27b-b33e77da8223","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2221","names":[{"code":"th","name":"เงินกู้ยืมระยะยาวจากกรรมการหรือผู้ถือหุ้น"},{"code":"en","name":"Long-term Loans from Directors/Shareholders"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '99cede98-1533-419c-a127-a86001fbf18a', '2231', 1, '{"id":"99cede98-1533-419c-a127-a86001fbf18a","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2231","names":[{"code":"th","name":"ประมาณการหนี้สินผลประโยชน์พนักงาน"},{"code":"en","name":"Provision for Employee Benefits"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd298c666-aaa8-4dab-af37-3738cc50d5c5', '2291', 1, '{"id":"d298c666-aaa8-4dab-af37-3738cc50d5c5","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2291","names":[{"code":"th","name":"หนี้สินไม่หมุนเวียนอื่น"},{"code":"en","name":"Other Non-Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '71c2ab25-7336-4254-aaca-608cefa13f4f', '3000', 1, '{"id":"71c2ab25-7336-4254-aaca-608cefa13f4f","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3000","names":[{"code":"th","name":"ส่วนของเจ้าของ"},{"code":"en","name":"Equity"}],"accounttype":"equity","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'da442e12-8212-484d-a24f-4406ce953fc2', '3100', 1, '{"id":"da442e12-8212-484d-a24f-4406ce953fc2","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3100","names":[{"code":"th","name":"ทุนจดทะเบียนและส่วนเกินทุน"},{"code":"en","name":"Share Capital & Premium"}],"accounttype":"equity","parentaccountcode":"3000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8fd54f83-878e-482b-a681-e2fd60aee91f', '3111', 1, '{"id":"8fd54f83-878e-482b-a681-e2fd60aee91f","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3111","names":[{"code":"th","name":"ทุนเรือนหุ้น - หุ้นสามัญ"},{"code":"en","name":"Authorized Share Capital - Common Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a7c55f68-5d3c-4e30-a5b0-ef76767d2c7e', '3112', 1, '{"id":"a7c55f68-5d3c-4e30-a5b0-ef76767d2c7e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3112","names":[{"code":"th","name":"ทุนเรือนหุ้น - หุ้นบุริมสิทธิ"},{"code":"en","name":"Authorized Share Capital - Preferred Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0b8f8d28-43a4-40e7-a409-86cfd2637a0e', '3121', 1, '{"id":"0b8f8d28-43a4-40e7-a409-86cfd2637a0e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3121","names":[{"code":"th","name":"ส่วนเกินมูลค่าหุ้นสามัญ"},{"code":"en","name":"Share Premium - Common Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b469b8b2-ac1a-4632-a635-96b0cf63f98f', '3131', 1, '{"id":"b469b8b2-ac1a-4632-a635-96b0cf63f98f","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3131","names":[{"code":"th","name":"ทุนส่วนของเจ้าของ (ห้างหุ้นส่วน/บุคคลธรรมดา)"},{"code":"en","name":"Owner''s / Partner''s Capital"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f29e99ad-5122-4f88-afc7-2876535a9fe0', '3132', 1, '{"id":"f29e99ad-5122-4f88-afc7-2876535a9fe0","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3132","names":[{"code":"th","name":"เงินถอนใช้ส่วนตัวของเจ้าของกิจการ"},{"code":"en","name":"Owner''s Drawings"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8a51af4f-1ffe-4125-a03c-62589db8587a', '3200', 1, '{"id":"8a51af4f-1ffe-4125-a03c-62589db8587a","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3200","names":[{"code":"th","name":"กำไร (ขาดทุน) สะสม"},{"code":"en","name":"Retained Earnings"}],"accounttype":"equity","parentaccountcode":"3000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0bb4271a-a3f6-49df-a18e-8a07f4f57b9c', '3211', 1, '{"id":"0bb4271a-a3f6-49df-a18e-8a07f4f57b9c","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3211","names":[{"code":"th","name":"สำรองตามกฎหมาย"},{"code":"en","name":"Legal Reserve"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '87e95a09-1f58-436a-ab26-d2ae4ef14685', '3212', 1, '{"id":"87e95a09-1f58-436a-ab26-d2ae4ef14685","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3212","names":[{"code":"th","name":"สำรองอื่นเพื่อวัตถุประสงค์เฉพาะ"},{"code":"en","name":"Other Appropriated Reserves"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '35642de5-9491-4abe-a6c4-7c32fb5a8689', '3221', 1, '{"id":"35642de5-9491-4abe-a6c4-7c32fb5a8689","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3221","names":[{"code":"th","name":"กำไร (ขาดทุน) สะสมยังไม่ได้จัดสรร"},{"code":"en","name":"Unappropriated Retained Earnings"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '63bc3aa9-b176-4802-a3ea-ee1072dd8d3f', '3222', 1, '{"id":"63bc3aa9-b176-4802-a3ea-ee1072dd8d3f","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3222","names":[{"code":"th","name":"กำไร (ขาดทุน) สุทธิประจำปี"},{"code":"en","name":"Net Profit/Loss for Current Year"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9f75471e-489f-4db8-acb6-2946924187c4', '3223', 1, '{"id":"9f75471e-489f-4db8-acb6-2946924187c4","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3223","names":[{"code":"th","name":"เงินปันผลจ่าย"},{"code":"en","name":"Dividends Declared & Paid"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8b6a3ba9-5a14-4992-a0c9-a5a191bc42b5', '4000', 1, '{"id":"8b6a3ba9-5a14-4992-a0c9-a5a191bc42b5","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4000","names":[{"code":"th","name":"รายได้"},{"code":"en","name":"Income"}],"accounttype":"income","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c1ef1922-86d5-42ee-a750-041d05a3ce8e', '4100', 1, '{"id":"c1ef1922-86d5-42ee-a750-041d05a3ce8e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4100","names":[{"code":"th","name":"รายได้จากการขายและการให้บริการ"},{"code":"en","name":"Revenue from Sales & Services"}],"accounttype":"income","parentaccountcode":"4000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9ce79250-eb77-4e12-a329-451f96571925', '4111', 1, '{"id":"9ce79250-eb77-4e12-a329-451f96571925","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4111","names":[{"code":"th","name":"รายได้จากการขายสินค้า - มีภาษีมูลค่าเพิ่ม 7%"},{"code":"en","name":"Sales Revenue - VAT 7%"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0f4e6902-d141-4e06-a507-6f816f81d483', '4112', 1, '{"id":"0f4e6902-d141-4e06-a507-6f816f81d483","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4112","names":[{"code":"th","name":"รายได้จากการขายสินค้า - อัตราภาษี 0% / ส่งออก"},{"code":"en","name":"Sales Revenue - Zero Rated / Export"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '4991c90b-2dbc-4548-adb1-01a5d0d4807f', '4113', 1, '{"id":"4991c90b-2dbc-4548-adb1-01a5d0d4807f","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4113","names":[{"code":"th","name":"รายได้จากการขายสินค้า - ได้รับยกเว้นภาษีมูลค่าเพิ่ม"},{"code":"en","name":"Sales Revenue - VAT Exempted"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9c4f8c3a-e35e-43c1-a4ab-53bcea569f97', '4121', 1, '{"id":"9c4f8c3a-e35e-43c1-a4ab-53bcea569f97","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4121","names":[{"code":"th","name":"รายได้จากการให้บริการ - มีภาษีมูลค่าเพิ่ม 7%"},{"code":"en","name":"Service Income - VAT 7%"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e648e0c1-3af0-4b2e-a7c6-13998ebb6a8e', '4122', 1, '{"id":"e648e0c1-3af0-4b2e-a7c6-13998ebb6a8e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4122","names":[{"code":"th","name":"รายได้จากการให้บริการ - ได้รับยกเว้นภาษีมูลค่าเพิ่ม"},{"code":"en","name":"Service Income - VAT Exempted"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '405cf11a-4614-4bec-a836-d1f85365e9f1', '4131', 1, '{"id":"405cf11a-4614-4bec-a836-d1f85365e9f1","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4131","names":[{"code":"th","name":"รับคืนสินค้าและลดหนี้ขาย"},{"code":"en","name":"Sales Returns and Allowances"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5db29e6d-a58c-4ecd-a93e-d81987ad35dd', '4132', 1, '{"id":"5db29e6d-a58c-4ecd-a93e-d81987ad35dd","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4132","names":[{"code":"th","name":"ส่วนลดจ่าย"},{"code":"en","name":"Sales Cash Discounts"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1a3882e7-917d-4caa-a8d3-988531a3f90a', '4141', 1, '{"id":"1a3882e7-917d-4caa-a8d3-988531a3f90a","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4141","names":[{"code":"th","name":"รายได้ค่าบริการขนส่งและจัดส่งสินค้า"},{"code":"en","name":"Freight & Delivery Income"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '72e8a892-01cc-40d6-afe3-ae7bf2f53fc8', '4200', 1, '{"id":"72e8a892-01cc-40d6-afe3-ae7bf2f53fc8","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4200","names":[{"code":"th","name":"รายได้อื่น"},{"code":"en","name":"Other Income"}],"accounttype":"income","parentaccountcode":"4000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '10346b3c-08e2-4a1b-adb7-9fe2007b6108', '4211', 1, '{"id":"10346b3c-08e2-4a1b-adb7-9fe2007b6108","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4211","names":[{"code":"th","name":"ดอกเบี้ยรับจากสถาบันการเงิน"},{"code":"en","name":"Interest Income from Banks"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'eec54a88-d244-4168-a8e8-396a8c5e4588', '4212', 1, '{"id":"eec54a88-d244-4168-a8e8-396a8c5e4588","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4212","names":[{"code":"th","name":"เงินปันผลรับจากเงินลงทุน"},{"code":"en","name":"Dividend Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f885baba-8276-4865-a673-6157022b926a', '4221', 1, '{"id":"f885baba-8276-4865-a673-6157022b926a","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4221","names":[{"code":"th","name":"กำไรจากการจำหน่ายทรัพย์สินถาวร"},{"code":"en","name":"Gain on Disposal of Fixed Assets"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c02a0123-85e4-40b1-aa12-0de0ca0d7db4', '4231', 1, '{"id":"c02a0123-85e4-40b1-aa12-0de0ca0d7db4","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4231","names":[{"code":"th","name":"กำไรจากอัตราแลกเปลี่ยนเงินตราต่างประเทศ"},{"code":"en","name":"Gain on Foreign Exchange"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c3857fc1-3d8a-4843-ae82-a14ed0a684f2', '4241', 1, '{"id":"c3857fc1-3d8a-4843-ae82-a14ed0a684f2","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4241","names":[{"code":"th","name":"รายได้ค่าเช่าอาคารและอุปกรณ์"},{"code":"en","name":"Rental Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9d85cd83-7903-423b-a964-6de8ead5b163', '4251', 1, '{"id":"9d85cd83-7903-423b-a964-6de8ead5b163","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4251","names":[{"code":"th","name":"หนี้สูญได้รับคืน"},{"code":"en","name":"Bad Debts Recovered"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '60585866-3da1-4c42-a1e2-66e517dc5e94', '4291', 1, '{"id":"60585866-3da1-4c42-a1e2-66e517dc5e94","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4291","names":[{"code":"th","name":"รายได้เบ็ดเตล็ดอื่น"},{"code":"en","name":"Miscellaneous Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e3efbe23-0511-4a40-a150-57af2513f3a3', '5000', 1, '{"id":"e3efbe23-0511-4a40-a150-57af2513f3a3","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5000","names":[{"code":"th","name":"ค่าใช้จ่าย"},{"code":"en","name":"Expenses"}],"accounttype":"expense","parentaccountcode":null,"normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '53f7310d-e774-46da-a546-bf566599b780', '5100', 1, '{"id":"53f7310d-e774-46da-a546-bf566599b780","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5100","names":[{"code":"th","name":"ต้นทุนขายและบริการ"},{"code":"en","name":"Cost of Goods Sold & Services"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '584ee0aa-3432-40e8-a286-0b4e6b43b770', '5111', 1, '{"id":"584ee0aa-3432-40e8-a286-0b4e6b43b770","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5111","names":[{"code":"th","name":"ซื้อสินค้าสำเร็จรูป"},{"code":"en","name":"Purchases of Merchandise"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5b0ea3cf-90e2-4977-a417-d9295f72139c', '5112', 1, '{"id":"5b0ea3cf-90e2-4977-a417-d9295f72139c","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5112","names":[{"code":"th","name":"ค่าขนส่งเข้า"},{"code":"en","name":"Freight-In"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'cabc7118-43b0-4d18-a1d8-5abab7e6f471', '5113', 1, '{"id":"cabc7118-43b0-4d18-a1d8-5abab7e6f471","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5113","names":[{"code":"th","name":"ส่งคืนสินค้าและส่วนลดที่ได้รับ"},{"code":"en","name":"Purchase Returns and Allowances"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a9140902-d58b-44e9-a7d3-99c86fa27202', '5114', 1, '{"id":"a9140902-d58b-44e9-a7d3-99c86fa27202","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5114","names":[{"code":"th","name":"ส่วนลดรับ"},{"code":"en","name":"Purchase Cash Discounts"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f869e862-ac06-419f-a793-876917e8e119', '5121', 1, '{"id":"f869e862-ac06-419f-a793-876917e8e119","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5121","names":[{"code":"th","name":"ต้นทุนสินค้าสำเร็จรูปที่ขาย"},{"code":"en","name":"Cost of Goods Sold"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'ce32f866-7b37-4274-a58d-34bd9462893f', '5131', 1, '{"id":"ce32f866-7b37-4274-a58d-34bd9462893f","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5131","names":[{"code":"th","name":"ต้นทุนค่าแรงและบริการ"},{"code":"en","name":"Direct Labor & Service Costs"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'ddfb9036-48c7-4793-ad57-63bbf071d0c0', '5132', 1, '{"id":"ddfb9036-48c7-4793-ad57-63bbf071d0c0","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5132","names":[{"code":"th","name":"ค่าจ้างเหมาบริการช่วงภายนอก (Subcontractor)"},{"code":"en","name":"Subcontractor Costs"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd6694490-20aa-494f-aa56-80f2367e65e2', '5141', 1, '{"id":"d6694490-20aa-494f-aa56-80f2367e65e2","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5141","names":[{"code":"th","name":"สินค้าสูญหาย เสียหาย และสินค้าชำรุด"},{"code":"en","name":"Inventory Shrinkage & Damage"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9782dc7f-bfdb-4ae1-a873-ad690d33341a', '5200', 1, '{"id":"9782dc7f-bfdb-4ae1-a873-ad690d33341a","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5200","names":[{"code":"th","name":"ค่าใช้จ่ายในการขาย"},{"code":"en","name":"Selling Expenses"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'edda99d8-1101-4f2b-a045-5891f394dfab', '5211', 1, '{"id":"edda99d8-1101-4f2b-a045-5891f394dfab","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5211","names":[{"code":"th","name":"เงินเดือน ค่าล่วงเวลา และเบี้ยเลี้ยงฝ่ายขาย"},{"code":"en","name":"Sales Salaries, Overtime & Allowances"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7b8ff237-c613-45df-ad86-6586e9e862f3', '5212', 1, '{"id":"7b8ff237-c613-45df-ad86-6586e9e862f3","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5212","names":[{"code":"th","name":"ค่านายหน้าและค่าคอมมิชชั่นฝ่ายขาย"},{"code":"en","name":"Sales Commissions & Incentives"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8fd05718-949f-4439-acd3-948ef476a8d1', '5221', 1, '{"id":"8fd05718-949f-4439-acd3-948ef476a8d1","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5221","names":[{"code":"th","name":"ค่าโฆษณา ประชาสัมพันธ์ และการตลาดออนไลน์"},{"code":"en","name":"Advertising, PR & Online Marketing"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '88763fae-901b-4ca9-aa19-e59bec5624b4', '5222', 1, '{"id":"88763fae-901b-4ca9-aa19-e59bec5624b4","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5222","names":[{"code":"th","name":"ค่าส่งเสริมการขาย ของแถม และตัวอย่างสินค้า"},{"code":"en","name":"Sales Promotion, Gifts & Samples"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7f623c2d-abfe-46cf-afe6-5b9235208b69', '5231', 1, '{"id":"7f623c2d-abfe-46cf-afe6-5b9235208b69","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5231","names":[{"code":"th","name":"ค่าขนส่งสินค้าออกและบริการจัดส่งให้ลูกค้า"},{"code":"en","name":"Outward Freight & Shipping to Customers"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '85aa472b-282b-4e03-a7c6-c5ead0e6ed2e', '5232', 1, '{"id":"85aa472b-282b-4e03-a7c6-c5ead0e6ed2e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5232","names":[{"code":"th","name":"ค่าวัสดุหีบห่อและบรรจุภัณฑ์สำหรับการขาย"},{"code":"en","name":"Packaging & Packing Materials"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '80933f13-68bc-43a9-af7d-1df976176d29', '5241', 1, '{"id":"80933f13-68bc-43a9-af7d-1df976176d29","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5241","names":[{"code":"th","name":"ค่าจัดงานแสดงสินค้าและการออกบูธ"},{"code":"en","name":"Exhibition & Trade Fair Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '44e0b46d-bd0a-40b5-a9ad-bd75fff549c5', '5251', 1, '{"id":"44e0b46d-bd0a-40b5-a9ad-bd75fff549c5","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5251","names":[{"code":"th","name":"ค่าน้ำมันและค่าเดินทางพบลูกค้าฝ่ายขาย"},{"code":"en","name":"Sales Fuel & Travel Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0f36a2e5-1c4e-4aae-a34e-f501a91f84bd', '5291', 1, '{"id":"0f36a2e5-1c4e-4aae-a34e-f501a91f84bd","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5291","names":[{"code":"th","name":"ค่าใช้จ่ายในการขายอื่น"},{"code":"en","name":"Other Selling Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8cf721e1-a992-4ddf-a9e4-c5648d33aa2d', '5300', 1, '{"id":"8cf721e1-a992-4ddf-a9e4-c5648d33aa2d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5300","names":[{"code":"th","name":"ค่าใช้จ่ายในการบริหาร"},{"code":"en","name":"Administrative Expenses"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '33920f8c-500f-410f-a4b5-46b436a41d9a', '5311', 1, '{"id":"33920f8c-500f-410f-a4b5-46b436a41d9a","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5311","names":[{"code":"th","name":"เงินเดือน ค่าจ้าง และค่าล่วงเวลาพนักงานสำนักงาน"},{"code":"en","name":"Office Salaries, Wages & Overtime"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0f8a0b42-46c9-47d1-ae41-c1197293b106', '5312', 1, '{"id":"0f8a0b42-46c9-47d1-ae41-c1197293b106","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5312","names":[{"code":"th","name":"ค่าตอบแทนและเบี้ยประชุมกรรมการ"},{"code":"en","name":"Directors'' Remuneration & Meeting Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'ceff0af8-4055-4d73-aba3-25ed46f891ff', '5313', 1, '{"id":"ceff0af8-4055-4d73-aba3-25ed46f891ff","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5313","names":[{"code":"th","name":"โบนัสพนักงานประจำปี"},{"code":"en","name":"Annual Staff Bonuses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1c88fbff-2bfe-4d5f-a720-f83ccb5ff7ff', '5314', 1, '{"id":"1c88fbff-2bfe-4d5f-a720-f83ccb5ff7ff","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5314","names":[{"code":"th","name":"เงินสมทบกองทุนประกันสังคม (ส่วนของนายจ้าง)"},{"code":"en","name":"Social Security Fund - Employer''s Contribution"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '431396f7-de86-48ad-afdd-ca9f78a2eba8', '5315', 1, '{"id":"431396f7-de86-48ad-afdd-ca9f78a2eba8","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5315","names":[{"code":"th","name":"เงินสมทบกองทุนเงินทดแทนและกองทุนสำรองเลี้ยงชีพ"},{"code":"en","name":"Workmen Compensation & Provident Fund"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '2d6e36c4-87cf-43c3-a280-3b3f669f656b', '5316', 1, '{"id":"2d6e36c4-87cf-43c3-a280-3b3f669f656b","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5316","names":[{"code":"th","name":"ค่าสวัสดิการพนักงานและชุดยูนิฟอร์ม"},{"code":"en","name":"Staff Welfare & Uniforms"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '4f8ee65e-2855-44c3-abce-176163c4fbbf', '5317', 1, '{"id":"4f8ee65e-2855-44c3-abce-176163c4fbbf","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5317","names":[{"code":"th","name":"ค่าฝึกอบรมและสัมมนาบุคลากร"},{"code":"en","name":"Training & Seminar Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c05c4587-3041-4490-a04d-2d97759fbece', '5321', 1, '{"id":"c05c4587-3041-4490-a04d-2d97759fbece","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5321","names":[{"code":"th","name":"ค่าเช่าอาคารสำนักงานและพื้นที่ประกอบการ"},{"code":"en","name":"Office Building Rent"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7d2b8e28-65af-4163-ab80-743372fee45d', '5322', 1, '{"id":"7d2b8e28-65af-4163-ab80-743372fee45d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5322","names":[{"code":"th","name":"ค่าน้ำประปา"},{"code":"en","name":"Water Utility Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8df584ca-cc54-462e-aa58-1d0be3b22d2f', '5323', 1, '{"id":"8df584ca-cc54-462e-aa58-1d0be3b22d2f","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5323","names":[{"code":"th","name":"ค่าไฟฟ้า"},{"code":"en","name":"Electricity Utility Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '432c8f45-9e5f-4458-a781-460f3194f606', '5324', 1, '{"id":"432c8f45-9e5f-4458-a781-460f3194f606","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5324","names":[{"code":"th","name":"ค่าโทรศัพท์และค่าบริการโทรคมนาคม"},{"code":"en","name":"Telephone & Telecom Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8feeacc6-e7c4-48c7-a790-b450f271b7de', '5325', 1, '{"id":"8feeacc6-e7c4-48c7-a790-b450f271b7de","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5325","names":[{"code":"th","name":"ค่าบริการอินเทอร์เน็ต ระบบเซิร์ฟเวอร์ และคลาวด์"},{"code":"en","name":"Internet, Server & Cloud Hosting Services"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '01938294-e896-48da-a7d6-9507db2b5c0d', '5331', 1, '{"id":"01938294-e896-48da-a7d6-9507db2b5c0d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5331","names":[{"code":"th","name":"ค่าเครื่องเขียน แบบพิมพ์ และวัสดุสำนักงาน"},{"code":"en","name":"Stationery, Printing & Office Supplies"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '69b60698-ff49-4f2a-a1aa-701a75fb7263', '5332', 1, '{"id":"69b60698-ff49-4f2a-a1aa-701a75fb7263","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5332","names":[{"code":"th","name":"ค่าไปรษณีย์และค่าส่งเอกสารพัสดุ"},{"code":"en","name":"Postal & Courier Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '93d59cda-8355-4456-a3e3-893f6c744b8b', '5341', 1, '{"id":"93d59cda-8355-4456-a3e3-893f6c744b8b","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5341","names":[{"code":"th","name":"ค่าตรวจสอบบัญชี (Audit Fees)"},{"code":"en","name":"Audit Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '505648de-0f00-4127-ae6f-332f394d94eb', '5342', 1, '{"id":"505648de-0f00-4127-ae6f-332f394d94eb","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5342","names":[{"code":"th","name":"ค่าจัดทำบัญชีและที่ปรึกษาภาษีอากร"},{"code":"en","name":"Accounting & Tax Consultation Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '2c18f4db-097c-46a8-a51f-fb40aa82fa2e', '5343', 1, '{"id":"2c18f4db-097c-46a8-a51f-fb40aa82fa2e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5343","names":[{"code":"th","name":"ค่าธรรมเนียมวิชาชีพกฎหมายและที่ปรึกษาธุรกิจ"},{"code":"en","name":"Legal & Business Advisory Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7158d2fd-9152-408e-a9e6-8a2a785eb3db', '5351', 1, '{"id":"7158d2fd-9152-408e-a9e6-8a2a785eb3db","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5351","names":[{"code":"th","name":"ค่าเบี้ยประกันภัยทรัพย์สิน อาคาร และยานพาหนะ"},{"code":"en","name":"Property, Building & Vehicle Insurance"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '92c1fc15-20d5-4def-af39-bd657e7c1a6b', '5352', 1, '{"id":"92c1fc15-20d5-4def-af39-bd657e7c1a6b","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5352","names":[{"code":"th","name":"ค่าซ่อมแซมและบำรุงรักษาอาคารและอุปกรณ์"},{"code":"en","name":"Repairs & Maintenance Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'fca610b9-8677-4c3d-abd1-703a0cf433c9', '5353', 1, '{"id":"fca610b9-8677-4c3d-abd1-703a0cf433c9","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5353","names":[{"code":"th","name":"ค่าบริการทำความสะอาดและรักษาความปลอดภัย"},{"code":"en","name":"Cleaning & Security Services"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'fba9a343-33bc-495c-ae9a-a46d941337b6', '5361', 1, '{"id":"fba9a343-33bc-495c-ae9a-a46d941337b6","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5361","names":[{"code":"th","name":"ค่าเสื่อมราคา - ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Depreciation - Land Improvements"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7476ee84-2d0c-43c8-a505-7d58deef6ec4', '5362', 1, '{"id":"7476ee84-2d0c-43c8-a505-7d58deef6ec4","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5362","names":[{"code":"th","name":"ค่าเสื่อมราคา - อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Depreciation - Buildings"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'ce40746d-c63f-4fca-a947-e97af563b977', '5363', 1, '{"id":"ce40746d-c63f-4fca-a947-e97af563b977","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5363","names":[{"code":"th","name":"ค่าเสื่อมราคา - ส่วนปรับปรุงอาคาร"},{"code":"en","name":"Depreciation - Building Improvements"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9237a25d-5595-40d0-abe3-b904ec07383e', '5364', 1, '{"id":"9237a25d-5595-40d0-abe3-b904ec07383e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5364","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องจักรและอุปกรณ์"},{"code":"en","name":"Depreciation - Machinery & Equipment"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'db9b303c-856d-4bba-af41-286840b6cc6e', '5365', 1, '{"id":"db9b303c-856d-4bba-af41-286840b6cc6e","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5365","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องตกแต่งและติดตั้ง"},{"code":"en","name":"Depreciation - Furniture & Fixtures"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c8450289-1069-4861-a411-e8ee63df6fb5', '5366', 1, '{"id":"c8450289-1069-4861-a411-e8ee63df6fb5","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5366","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องใช้สำนักงาน"},{"code":"en","name":"Depreciation - Office Equipment"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '95bccc79-08a2-4def-a24b-d199c0a726c9', '5367', 1, '{"id":"95bccc79-08a2-4def-a24b-d199c0a726c9","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5367","names":[{"code":"th","name":"ค่าเสื่อมราคา - คอมพิวเตอร์และอุปกรณ์"},{"code":"en","name":"Depreciation - Computer & Hardware"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e20cc96b-a775-48b6-a967-90b090bf1f41', '5368', 1, '{"id":"e20cc96b-a775-48b6-a967-90b090bf1f41","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5368","names":[{"code":"th","name":"ค่าเสื่อมราคา - ยานพาหนะ"},{"code":"en","name":"Depreciation - Vehicles"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '30104b9e-95d7-4e82-a312-0a61e5cbb2cc', '5371', 1, '{"id":"30104b9e-95d7-4e82-a312-0a61e5cbb2cc","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5371","names":[{"code":"th","name":"ค่าตัดจำหน่ายโปรแกรมคอมพิวเตอร์และซอฟต์แวร์"},{"code":"en","name":"Amortization - Software & Applications"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7a16c661-d422-4873-a27a-6dac32fef70a', '5381', 1, '{"id":"7a16c661-d422-4873-a27a-6dac32fef70a","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5381","names":[{"code":"th","name":"ค่าธรรมเนียมราชการและใบอนุญาตประกอบกิจการ"},{"code":"en","name":"Government Licenses & Registration Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'fbf957dd-a960-4ceb-a112-97f6f1eafe05', '5382', 1, '{"id":"fbf957dd-a960-4ceb-a112-97f6f1eafe05","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5382","names":[{"code":"th","name":"ภาษีที่ดินและสิ่งปลูกสร้าง"},{"code":"en","name":"Land and Building Taxes"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'aa92fde3-5195-49e6-a62c-5bae0f5e84ce', '5383', 1, '{"id":"aa92fde3-5195-49e6-a62c-5bae0f5e84ce","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5383","names":[{"code":"th","name":"ภาษีป้าย"},{"code":"en","name":"Signboard Tax"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b90a0430-117d-44a9-a552-7812df6816f8', '5384', 1, '{"id":"b90a0430-117d-44a9-a552-7812df6816f8","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5384","names":[{"code":"th","name":"ภาษีซื้อต้องห้าม / ภาษีซื้อที่ไม่สามารถขอคืนได้"},{"code":"en","name":"Non-refundable Input Tax (Disallowed Input VAT)"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c36c58bd-77e8-4961-a148-20172900afd3', '5385', 1, '{"id":"c36c58bd-77e8-4961-a148-20172900afd3","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5385","names":[{"code":"th","name":"เบี้ยปรับ เงินเพิ่ม และค่าปรับทางภาษีอากร"},{"code":"en","name":"Tax Penalties, Fines & Surcharges"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'fb1419fd-8d91-4a7d-a5f5-fe1ed018bea8', '5391', 1, '{"id":"fb1419fd-8d91-4a7d-a5f5-fe1ed018bea8","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5391","names":[{"code":"th","name":"ค่ารับรองและบริการลูกค้า"},{"code":"en","name":"Entertainment & Hospitality Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '92edab5d-c911-4e02-a2c3-b47d2178e818', '5392', 1, '{"id":"92edab5d-c911-4e02-a2c3-b47d2178e818","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5392","names":[{"code":"th","name":"เงินบริจาคเพื่อการกุศลและการศึกษา"},{"code":"en","name":"Charitable & Educational Donations"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd82f760b-3019-491f-a6cf-39e67ae4b61f', '5393', 1, '{"id":"d82f760b-3019-491f-a6cf-39e67ae4b61f","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5393","names":[{"code":"th","name":"หนี้สูญและหนี้สงสัยจะสูญ"},{"code":"en","name":"Bad Debts & Doubtful Accounts Expense"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f42233e7-4da3-46a8-a05f-cafea7bd2c28', '5394', 1, '{"id":"f42233e7-4da3-46a8-a05f-cafea7bd2c28","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5394","names":[{"code":"th","name":"ขาดทุนจากอัตราแลกเปลี่ยนเงินตราต่างประเทศ"},{"code":"en","name":"Loss on Foreign Exchange"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'dac08b43-fc2f-4333-a949-9ad4d32a0100', '5395', 1, '{"id":"dac08b43-fc2f-4333-a949-9ad4d32a0100","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5395","names":[{"code":"th","name":"ขาดทุนจากการจำหน่ายและตัดจำหน่ายทรัพย์สิน"},{"code":"en","name":"Loss on Disposal and Write-off of Assets"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'bdfe439f-bcaf-4c9c-ae6d-9f341e7a15e5', '5399', 1, '{"id":"bdfe439f-bcaf-4c9c-ae6d-9f341e7a15e5","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5399","names":[{"code":"th","name":"ค่าใช้จ่ายในการบริหารอื่น"},{"code":"en","name":"Other Administrative Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c939fc67-1e62-441a-a2fe-496aed78e44d', '5400', 1, '{"id":"c939fc67-1e62-441a-a2fe-496aed78e44d","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5400","names":[{"code":"th","name":"ต้นทุนทางการเงินและภาษีเงินได้"},{"code":"en","name":"Financial Costs & Income Tax"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0b20291d-aec7-41c8-a92c-c19b00579429', '5411', 1, '{"id":"0b20291d-aec7-41c8-a92c-c19b00579429","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5411","names":[{"code":"th","name":"ดอกเบี้ยจ่ายเงินกู้ยืมสถาบันการเงิน"},{"code":"en","name":"Interest Expense - Bank Loans"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e145b89e-0228-4543-aa02-2ec31d5168a4', '5412', 1, '{"id":"e145b89e-0228-4543-aa02-2ec31d5168a4","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5412","names":[{"code":"th","name":"ดอกเบี้ยจ่ายเงินเบิกเกินบัญชี (O/D)"},{"code":"en","name":"Interest Expense - Bank Overdraft (O/D)"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '3e796e57-afe0-41a9-a922-fa5254d9416b', '5413', 1, '{"id":"3e796e57-afe0-41a9-a922-fa5254d9416b","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5413","names":[{"code":"th","name":"ดอกเบี้ยจ่ายตามสัญญาเช่าทางการเงิน"},{"code":"en","name":"Interest Expense - Financial Leases"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '66d45e07-9259-42e3-a726-ddf55c60fa73', '5421', 1, '{"id":"66d45e07-9259-42e3-a726-ddf55c60fa73","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5421","names":[{"code":"th","name":"ค่าธรรมเนียมธนาคารและธุรกรรมทางการเงิน"},{"code":"en","name":"Bank Charges & Transaction Fees"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8c6f6599-4adf-4cd3-a2e6-c0a7b2e775f4', '5431', 1, '{"id":"8c6f6599-4adf-4cd3-a2e6-c0a7b2e775f4","holdingcode":"rungrueng","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5431","names":[{"code":"th","name":"ภาษีเงินได้นิติบุคคลประจำงวด"},{"code":"en","name":"Corporate Income Tax Expense"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'fiscal-years', '329212a4-38f3-4cad-ad1e-533b77218353', '2569', 1, '{"id":"329212a4-38f3-4cad-ad1e-533b77218353","code":"2569","kind":"fiscal-years","startdate":"2026-01-01","enddate":"2026-12-31","scale":2,"profitlossaccount":"3222","retainedearningsaccount":"3221","closed":false,"isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', 'bb651e3b-2b32-4605-a5bf-dcb3889cf3fa', 'JV', 1, '{"id":"bb651e3b-2b32-4605-a5bf-dcb3889cf3fa","code":"JV","kind":"journal-books","name":"สมุดรายวันทั่วไป","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', 'e89b2e9e-1bc2-4c99-a14c-70b4e04d55db', 'PV', 1, '{"id":"e89b2e9e-1bc2-4c99-a14c-70b4e04d55db","code":"PV","kind":"journal-books","name":"สมุดรายวันจ่าย","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', '03671108-35ff-42e7-a19e-8fc33c4fb40b', 'RV', 1, '{"id":"03671108-35ff-42e7-a19e-8fc33c4fb40b","code":"RV","kind":"journal-books","name":"สมุดรายวันรับ","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', 'a41967be-5c83-4bed-a9f0-c1507b7bbcf5', 'SV', 1, '{"id":"a41967be-5c83-4bed-a9f0-c1507b7bbcf5","code":"SV","kind":"journal-books","name":"สมุดรายวันซื้อ","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', '79ae6928-9f7a-4c10-a018-5e8b4ea7cb36', 'UV', 1, '{"id":"79ae6928-9f7a-4c10-a018-5e8b4ea7cb36","code":"UV","kind":"journal-books","name":"สมุดรายวันขาย","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', 'c0b8e6bb-d819-43a3-ae69-0cf2e01ca5c6', 'GJ', 1, '{"id":"c0b8e6bb-d819-43a3-ae69-0cf2e01ca5c6","code":"GJ","kind":"journal-books","name":"สมุดรายวันทั่วไป (GJ)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', '160258cf-2bf3-4893-a175-f2d6c57039c6', 'AP', 1, '{"id":"160258cf-2bf3-4893-a175-f2d6c57039c6","code":"AP","kind":"journal-books","name":"สมุดรายวันซื้อเชื่อ (AP)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', '3d371016-74bd-488c-aa69-5ca4c70d4886', 'AR', 1, '{"id":"3d371016-74bd-488c-aa69-5ca4c70d4886","code":"AR","kind":"journal-books","name":"สมุดรายวันขายเชื่อ (AR)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'account-groups', 'f2067420-eb97-4cd7-a11e-fdea90ae3622', 'BM69-A', 1, '{"id":"f2067420-eb97-4cd7-a11e-fdea90ae3622","code":"BM69-A","kind":"account-groups","name":"สินทรัพย์","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'account-groups', '02949af5-6d30-46a2-a5f2-992162f32ac3', 'BM69-L', 1, '{"id":"02949af5-6d30-46a2-a5f2-992162f32ac3","code":"BM69-L","kind":"account-groups","name":"หนี้สิน","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'account-groups', '0fea3c61-6a85-4457-a187-5790fcf9c894', 'BM69-E', 1, '{"id":"0fea3c61-6a85-4457-a187-5790fcf9c894","code":"BM69-E","kind":"account-groups","name":"ส่วนของเจ้าของ","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'account-groups', '83c2c5d5-e7c1-4438-a70e-330e118e715a', 'BM69-R', 1, '{"id":"83c2c5d5-e7c1-4438-a70e-330e118e715a","code":"BM69-R","kind":"account-groups","name":"รายได้","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'account-groups', 'ee58ca08-d205-4f5a-a0fb-6d6bc67125cf', 'BM69-X', 1, '{"id":"ee58ca08-d205-4f5a-a0fb-6d6bc67125cf","code":"BM69-X","kind":"account-groups","name":"ค่าใช้จ่าย","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"rungrueng","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;


-- Accounts for Holding: rungrueng, Company: c01
INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ef1f8e4d-a24f-4942-aeb2-ba47ca51e6b4', '1000', 1, '{"id":"ef1f8e4d-a24f-4942-aeb2-ba47ca51e6b4","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1000","names":[{"code":"th","name":"สินทรัพย์"},{"code":"en","name":"Assets"}],"accounttype":"asset","parentaccountcode":null,"normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '111a7d29-6105-4993-a49d-487627922011', '1100', 1, '{"id":"111a7d29-6105-4993-a49d-487627922011","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1100","names":[{"code":"th","name":"สินทรัพย์หมุนเวียน"},{"code":"en","name":"Current Assets"}],"accounttype":"asset","parentaccountcode":"1000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '4a77387b-293c-4f3e-a1aa-72825a0b4f62', '1111', 1, '{"id":"4a77387b-293c-4f3e-a1aa-72825a0b4f62","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1111","names":[{"code":"th","name":"เงินสดในมือ"},{"code":"en","name":"Cash on Hand"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c6c69e99-7f1b-4ee6-a93e-606400cba5b1', '1112', 1, '{"id":"c6c69e99-7f1b-4ee6-a93e-606400cba5b1","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1112","names":[{"code":"th","name":"เงินสดย่อย - สำนักงานใหญ่"},{"code":"en","name":"Petty Cash - Head Office"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6f5cadb9-076e-49f3-a837-c48c10bf1a18', '1113', 1, '{"id":"6f5cadb9-076e-49f3-a837-c48c10bf1a18","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1113","names":[{"code":"th","name":"เงินสดย่อย - ฝ่ายปฏิบัติการและสาขา"},{"code":"en","name":"Petty Cash - Operations & Branch"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '065be7c8-9e2e-4406-ad50-d8141c3b5a93', '1121', 1, '{"id":"065be7c8-9e2e-4406-ad50-d8141c3b5a93","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1121","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย"},{"code":"en","name":"Cash at Bank - Current KBANK"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'bcd4824b-1db3-4b3b-a104-5db8ee8e0777', '1122', 1, '{"id":"bcd4824b-1db3-4b3b-a104-5db8ee8e0777","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1122","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารกรุงเทพ"},{"code":"en","name":"Cash at Bank - Current BBL"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8111d4fb-e291-4ed9-ae68-b62307204861', '1123', 1, '{"id":"8111d4fb-e291-4ed9-ae68-b62307204861","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1123","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารไทยพาณิชย์"},{"code":"en","name":"Cash at Bank - Current SCB"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'de28e0dd-2979-467b-a219-d95b44b5561a', '1124', 1, '{"id":"de28e0dd-2979-467b-a219-d95b44b5561a","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1124","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารกสิกรไทย"},{"code":"en","name":"Cash at Bank - Savings KBANK"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '58b381b9-7953-4409-a73f-7afd828272d7', '1125', 1, '{"id":"58b381b9-7953-4409-a73f-7afd828272d7","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1125","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารกรุงเทพ"},{"code":"en","name":"Cash at Bank - Savings BBL"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '28c68c75-7530-45cb-a409-614a59ceea77', '1126', 1, '{"id":"28c68c75-7530-45cb-a409-614a59ceea77","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1126","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารไทยพาณิชย์"},{"code":"en","name":"Cash at Bank - Savings SCB"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '52babf34-827f-4fe0-af64-47af45bc2d2e', '1127', 1, '{"id":"52babf34-827f-4fe0-af64-47af45bc2d2e","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1127","names":[{"code":"th","name":"เงินฝากประจำระยะสั้น (ไม่เกิน 3 เดือน)"},{"code":"en","name":"Short-term Fixed Deposit"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6b423cac-28e3-45c0-a09f-61f66ee6121a', '1131', 1, '{"id":"6b423cac-28e3-45c0-a09f-61f66ee6121a","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1131","names":[{"code":"th","name":"ลูกหนี้การค้า - ในประเทศ"},{"code":"en","name":"Trade Accounts Receivable - Domestic"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '4da06c6f-f27e-49cc-a0d7-b0ac94e6dfe3', '1132', 1, '{"id":"4da06c6f-f27e-49cc-a0d7-b0ac94e6dfe3","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1132","names":[{"code":"th","name":"ลูกหนี้การค้า - ต่างประเทศ"},{"code":"en","name":"Trade Accounts Receivable - Overseas"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3b35af8f-e97e-4973-ad70-f2972d28e0ff', '1133', 1, '{"id":"3b35af8f-e97e-4973-ad70-f2972d28e0ff","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1133","names":[{"code":"th","name":"ตั๋วเงินรับการค้า"},{"code":"en","name":"Notes Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ea7f8156-7e76-4ccf-acd4-6c7d3fd8fab8', '1134', 1, '{"id":"ea7f8156-7e76-4ccf-acd4-6c7d3fd8fab8","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1134","names":[{"code":"th","name":"เช็ครับลงวันที่ล่วงหน้า"},{"code":"en","name":"Post-dated Cheques Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'cf77a9fa-23e8-4f28-a447-6c918f438f4a', '1138', 1, '{"id":"cf77a9fa-23e8-4f28-a447-6c918f438f4a","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1138","names":[{"code":"th","name":"ลูกหนี้อื่นและเงินยืมทดรอง"},{"code":"en","name":"Other Receivables & Advances"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a0792493-711b-4cfc-abc5-812e6fc187f9', '1139', 1, '{"id":"a0792493-711b-4cfc-abc5-812e6fc187f9","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1139","names":[{"code":"th","name":"ค่าเผื่อผลขาดทุนด้านเครดิตที่คาดว่าจะเกิดขึ้น"},{"code":"en","name":"Allowance for Expected Credit Losses"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '24bd72f4-93e0-46b4-a31b-14a5982855ed', '1141', 1, '{"id":"24bd72f4-93e0-46b4-a31b-14a5982855ed","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1141","names":[{"code":"th","name":"สินค้าสำเร็จรูป"},{"code":"en","name":"Finished Goods"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3faa8c06-3047-4d9f-a183-1ef5a85aa33f', '1142', 1, '{"id":"3faa8c06-3047-4d9f-a183-1ef5a85aa33f","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1142","names":[{"code":"th","name":"สินค้าระหว่างทำ / งานระหว่างทำ"},{"code":"en","name":"Work in Process"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2f785fed-d181-4c1c-ae51-f6243d12271f', '1143', 1, '{"id":"2f785fed-d181-4c1c-ae51-f6243d12271f","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1143","names":[{"code":"th","name":"วัตถุดิบและส่วนประกอบ"},{"code":"en","name":"Raw Materials & Components"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '4c40a4a2-bd51-427d-a712-7e35d7ac30fc', '1144', 1, '{"id":"4c40a4a2-bd51-427d-a712-7e35d7ac30fc","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1144","names":[{"code":"th","name":"วัสดุสิ้นเปลืองและบรรจุภัณฑ์"},{"code":"en","name":"Factory Supplies & Packaging"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '02859563-eddd-4d6c-ad1d-ad107202f7a8', '1145', 1, '{"id":"02859563-eddd-4d6c-ad1d-ad107202f7a8","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1145","names":[{"code":"th","name":"สินค้าระหว่างทาง"},{"code":"en","name":"Goods in Transit"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8a8ec90a-56f0-4156-a3a8-197f033bf034', '1149', 1, '{"id":"8a8ec90a-56f0-4156-a3a8-197f033bf034","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1149","names":[{"code":"th","name":"ค่าเผื่อการลดมูลค่าสินค้าคงเหลือ"},{"code":"en","name":"Allowance for Inventory Devaluation"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'd366a801-baac-4ad6-aba1-0325a4f26c32', '1151', 1, '{"id":"d366a801-baac-4ad6-aba1-0325a4f26c32","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1151","names":[{"code":"th","name":"ภาษีซื้อ"},{"code":"en","name":"Input VAT"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c3a4f1f0-0120-49b1-a9af-82fc0734d0bd', '1152', 1, '{"id":"c3a4f1f0-0120-49b1-a9af-82fc0734d0bd","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1152","names":[{"code":"th","name":"ภาษีซื้อยังไม่ถึงกำหนดชำระ"},{"code":"en","name":"Undue Input VAT"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6f9fa0c8-7a94-443b-ab90-2aa4c4b93285', '1153', 1, '{"id":"6f9fa0c8-7a94-443b-ab90-2aa4c4b93285","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1153","names":[{"code":"th","name":"ภาษีเงินได้ถูกหัก ณ ที่จ่าย (WHT ถูกหัก)"},{"code":"en","name":"Withholding Tax Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'cc7924fa-67c5-4ba8-a691-de81a438b4ad', '1154', 1, '{"id":"cc7924fa-67c5-4ba8-a691-de81a438b4ad","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1154","names":[{"code":"th","name":"ภาษีเงินได้นิติบุคคลจ่ายล่วงหน้า (ภ.ง.ด.51)"},{"code":"en","name":"Prepaid Corporate Income Tax"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1664d4f1-4654-4a1d-a64b-f40960cf91f9', '1161', 1, '{"id":"1664d4f1-4654-4a1d-a64b-f40960cf91f9","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1161","names":[{"code":"th","name":"ค่าเบี้ยประกันภัยจ่ายล่วงหน้า"},{"code":"en","name":"Prepaid Insurance"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '461efc19-1d13-4934-a52e-c31adb430684', '1162', 1, '{"id":"461efc19-1d13-4934-a52e-c31adb430684","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1162","names":[{"code":"th","name":"ค่าเช่าจ่ายล่วงหน้า"},{"code":"en","name":"Prepaid Rent"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b5356055-f7d4-44c1-a734-8f93a3843310', '1163', 1, '{"id":"b5356055-f7d4-44c1-a734-8f93a3843310","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1163","names":[{"code":"th","name":"เงินมัดจำค่าสินค้าและบริการล่วงหน้า"},{"code":"en","name":"Advance Payments for Goods & Services"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e95649bc-495c-479f-a8cc-5cee56f2214b', '1169', 1, '{"id":"e95649bc-495c-479f-a8cc-5cee56f2214b","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1169","names":[{"code":"th","name":"สินทรัพย์หมุนเวียนอื่น"},{"code":"en","name":"Other Current Assets"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c72b3bf4-a3b8-44e6-a675-4d82aad94194', '1200', 1, '{"id":"c72b3bf4-a3b8-44e6-a675-4d82aad94194","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1200","names":[{"code":"th","name":"สินทรัพย์ไม่หมุนเวียน"},{"code":"en","name":"Non-Current Assets"}],"accounttype":"asset","parentaccountcode":"1000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '72e9fa31-1a7b-49e4-a100-190c32e4f629', '1211', 1, '{"id":"72e9fa31-1a7b-49e4-a100-190c32e4f629","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1211","names":[{"code":"th","name":"ที่ดิน"},{"code":"en","name":"Land"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e7b58576-f515-47c9-a369-9d1a208a0d98', '1212', 1, '{"id":"e7b58576-f515-47c9-a369-9d1a208a0d98","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1212","names":[{"code":"th","name":"ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Land Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '565f9663-ae3c-40b8-ae89-c15f22c3624e', '1213', 1, '{"id":"565f9663-ae3c-40b8-ae89-c15f22c3624e","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1213","names":[{"code":"th","name":"อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Buildings & Constructions"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ea369a86-7ced-4735-a4bf-a620f4bb10c2', '1214', 1, '{"id":"ea369a86-7ced-4735-a4bf-a620f4bb10c2","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1214","names":[{"code":"th","name":"ส่วนปรับปรุงอาคารและระบบสาธารณูปโภค"},{"code":"en","name":"Building Improvements & Systems"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '938cccc1-4052-4889-afcc-6b1c9e96855d', '1221', 1, '{"id":"938cccc1-4052-4889-afcc-6b1c9e96855d","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1221","names":[{"code":"th","name":"เครื่องจักรและอุปกรณ์โรงงาน"},{"code":"en","name":"Machinery & Factory Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f8da9ecf-8212-491a-ada7-a729db8d3183', '1222', 1, '{"id":"f8da9ecf-8212-491a-ada7-a729db8d3183","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1222","names":[{"code":"th","name":"เครื่องตกแต่ง ติดตั้ง และเฟอร์นิเจอร์"},{"code":"en","name":"Furniture & Fixtures"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b01f2260-5371-4a08-a9bb-9fc6ef6379b4', '1223', 1, '{"id":"b01f2260-5371-4a08-a9bb-9fc6ef6379b4","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1223","names":[{"code":"th","name":"เครื่องใช้และอุปกรณ์สำนักงาน"},{"code":"en","name":"Office Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b63e04c9-1b0e-4c2c-abef-535a76fdd1ba', '1224', 1, '{"id":"b63e04c9-1b0e-4c2c-abef-535a76fdd1ba","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1224","names":[{"code":"th","name":"คอมพิวเตอร์และอุปกรณ์ประมวลผล"},{"code":"en","name":"Computer & Hardware Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1ea16cfb-acfc-4d10-a1e9-65e420bb066f', '1231', 1, '{"id":"1ea16cfb-acfc-4d10-a1e9-65e420bb066f","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1231","names":[{"code":"th","name":"ยานพาหนะและรถบรรทุกขนส่ง"},{"code":"en","name":"Vehicles & Transport Trucks"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '641c03a9-d718-417f-a98a-f8856ed8c08a', '1241', 1, '{"id":"641c03a9-d718-417f-a98a-f8856ed8c08a","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1241","names":[{"code":"th","name":"งานระหว่างก่อสร้างและติดตั้งเครื่องจักร"},{"code":"en","name":"Construction & Installation in Progress"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '094ebdd2-00eb-4134-a15d-178b70f336f4', '1281', 1, '{"id":"094ebdd2-00eb-4134-a15d-178b70f336f4","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1281","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Acc. Dep. - Land Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '949da850-4468-4447-a430-3d052bb8608a', '1282', 1, '{"id":"949da850-4468-4447-a430-3d052bb8608a","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1282","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Acc. Dep. - Buildings"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '4a093630-ad45-4197-aed9-2c8192ff0e7b', '1283', 1, '{"id":"4a093630-ad45-4197-aed9-2c8192ff0e7b","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1283","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ส่วนปรับปรุงอาคาร"},{"code":"en","name":"Acc. Dep. - Building Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '44df1b6c-e36a-4c1d-a542-34606f348999', '1284', 1, '{"id":"44df1b6c-e36a-4c1d-a542-34606f348999","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1284","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องจักรและอุปกรณ์"},{"code":"en","name":"Acc. Dep. - Machinery & Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '5793a75b-4892-4f90-a281-3103a47c8bcb', '1285', 1, '{"id":"5793a75b-4892-4f90-a281-3103a47c8bcb","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1285","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องตกแต่งและติดตั้ง"},{"code":"en","name":"Acc. Dep. - Furniture & Fixtures"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '4e37297b-be88-4e31-a9b9-4b49108d885b', '1286', 1, '{"id":"4e37297b-be88-4e31-a9b9-4b49108d885b","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1286","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องใช้สำนักงาน"},{"code":"en","name":"Acc. Dep. - Office Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b6830ccb-8a63-44d1-a690-a749c792f77b', '1287', 1, '{"id":"b6830ccb-8a63-44d1-a690-a749c792f77b","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1287","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - คอมพิวเตอร์และอุปกรณ์"},{"code":"en","name":"Acc. Dep. - Computer & Hardware"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a8b67fbc-f809-4f89-a3a4-1e01a2fb8c27', '1288', 1, '{"id":"a8b67fbc-f809-4f89-a3a4-1e01a2fb8c27","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1288","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ยานพาหนะ"},{"code":"en","name":"Acc. Dep. - Vehicles"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '7b715063-122f-44b1-a58d-a8504a4bccbf', '1291', 1, '{"id":"7b715063-122f-44b1-a58d-a8504a4bccbf","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1291","names":[{"code":"th","name":"โปรแกรมคอมพิวเตอร์และสิทธิการใช้งาน"},{"code":"en","name":"Software & Licenses"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'bcd6fd9a-5820-40db-aab8-6d31932cf9d2', '1292', 1, '{"id":"bcd6fd9a-5820-40db-aab8-6d31932cf9d2","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1292","names":[{"code":"th","name":"ค่าตัดจำหน่ายสะสม - โปรแกรมคอมพิวเตอร์"},{"code":"en","name":"Acc. Amortization - Software"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '214b7d42-1b6b-422d-ad68-f3f0ac7a9ad1', '1295', 1, '{"id":"214b7d42-1b6b-422d-ad68-f3f0ac7a9ad1","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1295","names":[{"code":"th","name":"เงินประกันและเงินมัดจำระยะยาว"},{"code":"en","name":"Long-term Deposits & Guarantees"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e692cff9-2693-4ca8-ae6b-436e631a9928', '2000', 1, '{"id":"e692cff9-2693-4ca8-ae6b-436e631a9928","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2000","names":[{"code":"th","name":"หนี้สิน"},{"code":"en","name":"Liabilities"}],"accounttype":"liability","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '9fba6aa5-f65e-49b5-abb6-1e182886eca9', '2100', 1, '{"id":"9fba6aa5-f65e-49b5-abb6-1e182886eca9","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2100","names":[{"code":"th","name":"หนี้สินหมุนเวียน"},{"code":"en","name":"Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'd92a2d8a-4788-4b43-a9d4-b555ba3d1ad4', '2111', 1, '{"id":"d92a2d8a-4788-4b43-a9d4-b555ba3d1ad4","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2111","names":[{"code":"th","name":"เจ้าหนี้การค้า - ในประเทศ"},{"code":"en","name":"Trade Accounts Payable - Domestic"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2b7bcf1f-0f1a-4c18-a170-7be37833cfed', '2112', 1, '{"id":"2b7bcf1f-0f1a-4c18-a170-7be37833cfed","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2112","names":[{"code":"th","name":"เจ้าหนี้การค้า - ต่างประเทศ"},{"code":"en","name":"Trade Accounts Payable - Overseas"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1d1c0740-48c4-4e77-ad8d-b33624c2a1e4', '2113', 1, '{"id":"1d1c0740-48c4-4e77-ad8d-b33624c2a1e4","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2113","names":[{"code":"th","name":"ตั๋วเงินจ่ายการค้า"},{"code":"en","name":"Notes Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'd3d9e4ca-c45f-492f-aaaa-c1ea96b9e88f', '2114', 1, '{"id":"d3d9e4ca-c45f-492f-aaaa-c1ea96b9e88f","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2114","names":[{"code":"th","name":"เช็คจ่ายลงวันที่ล่วงหน้า"},{"code":"en","name":"Post-dated Cheques Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '14c51fdd-ea4e-41a5-a0d6-dca4259049a3', '2121', 1, '{"id":"14c51fdd-ea4e-41a5-a0d6-dca4259049a3","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2121","names":[{"code":"th","name":"เจ้าหนี้อื่นและเงินทดรองรับ"},{"code":"en","name":"Other Payables & Advances"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'fb10683d-0ce6-47b9-a078-1b0908619861', '2122', 1, '{"id":"fb10683d-0ce6-47b9-a078-1b0908619861","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2122","names":[{"code":"th","name":"เจ้าหนี้กรมสรรพากร"},{"code":"en","name":"Revenue Department Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '51f775f8-f9d6-4c25-ab3e-c5c185dd044c', '2131', 1, '{"id":"51f775f8-f9d6-4c25-ab3e-c5c185dd044c","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2131","names":[{"code":"th","name":"ภาษีขาย"},{"code":"en","name":"Output VAT"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '24eddbe3-febb-4216-a582-19d40c1db5ce', '2132', 1, '{"id":"24eddbe3-febb-4216-a582-19d40c1db5ce","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2132","names":[{"code":"th","name":"ภาษีขายยังไม่ถึงกำหนดชำระ"},{"code":"en","name":"Undue Output VAT"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'cf0fbc88-8a30-4566-a4d1-1409c163f753', '2141', 1, '{"id":"cf0fbc88-8a30-4566-a4d1-1409c163f753","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2141","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.1 (เงินเดือน)"},{"code":"en","name":"WHT Payable - P.N.D.1 (Salaries)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '5b871125-e0a4-46c8-accc-542ddb0a8bfc', '2142', 1, '{"id":"5b871125-e0a4-46c8-accc-542ddb0a8bfc","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2142","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.3 (บุคคลธรรมดา)"},{"code":"en","name":"WHT Payable - P.N.D.3 (Individuals)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e769c35c-baeb-4e06-ad81-bb5f07799775', '2143', 1, '{"id":"e769c35c-baeb-4e06-ad81-bb5f07799775","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2143","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.53 (นิติบุคคล)"},{"code":"en","name":"WHT Payable - P.N.D.53 (Corporations)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6293d384-298b-412b-a8fc-4a295b205c26', '2144', 1, '{"id":"6293d384-298b-412b-a8fc-4a295b205c26","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2144","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.2 (ดอกเบี้ย/ปันผล)"},{"code":"en","name":"WHT Payable - P.N.D.2"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '0eefc9da-abd1-46e2-ad36-50368b37c201', '2145', 1, '{"id":"0eefc9da-abd1-46e2-ad36-50368b37c201","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2145","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.54 (ส่งไปต่างประเทศ)"},{"code":"en","name":"WHT Payable - P.N.D.54 (Overseas)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2f45069c-5210-4d66-a145-ba42ea5d80ac', '2151', 1, '{"id":"2f45069c-5210-4d66-a145-ba42ea5d80ac","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2151","names":[{"code":"th","name":"เงินสมทบกองทุนประกันสังคมค้างจ่าย (ลูกจ้าง+นายจ้าง)"},{"code":"en","name":"Social Security Fund Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '5613ee21-f7c6-4a06-a266-72af23b41d89', '2152', 1, '{"id":"5613ee21-f7c6-4a06-a266-72af23b41d89","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2152","names":[{"code":"th","name":"เงินเดือนและค่าจ้างค้างจ่าย"},{"code":"en","name":"Accrued Salaries & Wages"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '44918c4c-1655-4078-af3d-ed080c359bfd', '2153', 1, '{"id":"44918c4c-1655-4078-af3d-ed080c359bfd","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2153","names":[{"code":"th","name":"ค่าเช่าค้างจ่าย"},{"code":"en","name":"Accrued Rent Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f0fb2e46-9e4e-4b51-a18b-16f1a797d31d', '2154', 1, '{"id":"f0fb2e46-9e4e-4b51-a18b-16f1a797d31d","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2154","names":[{"code":"th","name":"ค่าน้ำประปาและค่าไฟฟ้าค้างจ่าย"},{"code":"en","name":"Accrued Utilities Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '543672aa-048a-4fcd-a7c9-267737cd6776', '2155', 1, '{"id":"543672aa-048a-4fcd-a7c9-267737cd6776","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2155","names":[{"code":"th","name":"ค่าโทรศัพท์และอินเทอร์เน็ตค้างจ่าย"},{"code":"en","name":"Accrued Telephone & Internet"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '01295cff-bb60-4c7d-a554-c144738a3e0c', '2156', 1, '{"id":"01295cff-bb60-4c7d-a554-c144738a3e0c","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2156","names":[{"code":"th","name":"ค่าสอบบัญชีและทำบัญชีค้างจ่าย"},{"code":"en","name":"Accrued Audit & Accounting Fees"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '581d0d8d-d427-473b-abfb-5bac2b1b239f', '2157', 1, '{"id":"581d0d8d-d427-473b-abfb-5bac2b1b239f","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2157","names":[{"code":"th","name":"ดอกเบี้ยค้างจ่าย"},{"code":"en","name":"Accrued Interest Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'cb2e2c7c-e201-49cf-a78f-78aabc18da53', '2158', 1, '{"id":"cb2e2c7c-e201-49cf-a78f-78aabc18da53","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2158","names":[{"code":"th","name":"โบนัสพนักงานค้างจ่าย"},{"code":"en","name":"Accrued Staff Bonuses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'd699c9db-5bb4-4d52-a0de-9ef9ecdd43f0', '2159', 1, '{"id":"d699c9db-5bb4-4d52-a0de-9ef9ecdd43f0","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2159","names":[{"code":"th","name":"ค่าใช้จ่ายค้างจ่ายอื่น"},{"code":"en","name":"Other Accrued Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '208e7981-cb70-4b1d-ad13-6c4bd27033f8', '2161', 1, '{"id":"208e7981-cb70-4b1d-ad13-6c4bd27033f8","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2161","names":[{"code":"th","name":"เงินรับล่วงหน้าค่าสินค้าและบริการจากลูกค้า"},{"code":"en","name":"Advances Received from Customers"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '133faaca-369f-406a-a5d6-4de35f526099', '2171', 1, '{"id":"133faaca-369f-406a-a5d6-4de35f526099","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2171","names":[{"code":"th","name":"เงินเบิกเกินบัญชีธนาคาร (O/D)"},{"code":"en","name":"Bank Overdrafts (O/D)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '0bab5278-367f-4b52-a991-fe8c98810ea0', '2172', 1, '{"id":"0bab5278-367f-4b52-a991-fe8c98810ea0","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2172","names":[{"code":"th","name":"เงินกู้ยืมระยะสั้นจากสถาบันการเงิน"},{"code":"en","name":"Short-term Borrowings from Banks"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c4ac960f-9caf-4782-a9b7-7ef4bb44bb81', '2173', 1, '{"id":"c4ac960f-9caf-4782-a9b7-7ef4bb44bb81","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2173","names":[{"code":"th","name":"เงินกู้ยืมระยะสั้นจากกรรมการหรือบุคคลที่เกี่ยวข้องกัน"},{"code":"en","name":"Short-term Loans from Directors"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '04aa9fc9-c9a4-4731-a671-521bb7307ac3', '2181', 1, '{"id":"04aa9fc9-c9a4-4731-a671-521bb7307ac3","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2181","names":[{"code":"th","name":"ส่วนของหนี้สินระยะยาวที่ถึงกำหนดชำระภายในหนึ่งปี"},{"code":"en","name":"Current Portion of Long-term Debt"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6bfbeacb-dfb0-497b-a42e-3c3a748934da', '2200', 1, '{"id":"6bfbeacb-dfb0-497b-a42e-3c3a748934da","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2200","names":[{"code":"th","name":"หนี้สินไม่หมุนเวียน"},{"code":"en","name":"Non-Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f3ae0856-a045-4d68-aaba-cd3a6cc3ff44', '2211', 1, '{"id":"f3ae0856-a045-4d68-aaba-cd3a6cc3ff44","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2211","names":[{"code":"th","name":"เงินกู้ยืมระยะยาวจากสถาบันการเงิน"},{"code":"en","name":"Long-term Bank Loans"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '9365ae3e-9009-4b9b-ab5f-493c8f82cda1', '2212', 1, '{"id":"9365ae3e-9009-4b9b-ab5f-493c8f82cda1","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2212","names":[{"code":"th","name":"หนี้สินตามสัญญาเช่าทางการเงินระยะยาว"},{"code":"en","name":"Long-term Financial Lease Liabilities"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '872c8bdc-2a76-443d-a1e2-942125b35855', '2221', 1, '{"id":"872c8bdc-2a76-443d-a1e2-942125b35855","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2221","names":[{"code":"th","name":"เงินกู้ยืมระยะยาวจากกรรมการหรือผู้ถือหุ้น"},{"code":"en","name":"Long-term Loans from Directors/Shareholders"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '28222afc-26ad-4b0a-a660-e701768d1474', '2231', 1, '{"id":"28222afc-26ad-4b0a-a660-e701768d1474","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2231","names":[{"code":"th","name":"ประมาณการหนี้สินผลประโยชน์พนักงาน"},{"code":"en","name":"Provision for Employee Benefits"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '62e4e49f-8dce-4ce3-afdb-3382ae05817f', '2291', 1, '{"id":"62e4e49f-8dce-4ce3-afdb-3382ae05817f","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2291","names":[{"code":"th","name":"หนี้สินไม่หมุนเวียนอื่น"},{"code":"en","name":"Other Non-Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b7179dd5-85f8-4064-af7c-9c2295942396', '3000', 1, '{"id":"b7179dd5-85f8-4064-af7c-9c2295942396","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3000","names":[{"code":"th","name":"ส่วนของเจ้าของ"},{"code":"en","name":"Equity"}],"accounttype":"equity","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '35a98504-1d50-4b2c-aaa8-ecf945c924fd', '3100', 1, '{"id":"35a98504-1d50-4b2c-aaa8-ecf945c924fd","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3100","names":[{"code":"th","name":"ทุนจดทะเบียนและส่วนเกินทุน"},{"code":"en","name":"Share Capital & Premium"}],"accounttype":"equity","parentaccountcode":"3000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '56286a72-4220-4d1c-a064-f4c51a353ade', '3111', 1, '{"id":"56286a72-4220-4d1c-a064-f4c51a353ade","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3111","names":[{"code":"th","name":"ทุนเรือนหุ้น - หุ้นสามัญ"},{"code":"en","name":"Authorized Share Capital - Common Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '02c9b1ba-d40d-445a-a043-f5d570383b4a', '3112', 1, '{"id":"02c9b1ba-d40d-445a-a043-f5d570383b4a","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3112","names":[{"code":"th","name":"ทุนเรือนหุ้น - หุ้นบุริมสิทธิ"},{"code":"en","name":"Authorized Share Capital - Preferred Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6df35ac1-afa9-44b0-ad14-7e9fc4726ef3', '3121', 1, '{"id":"6df35ac1-afa9-44b0-ad14-7e9fc4726ef3","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3121","names":[{"code":"th","name":"ส่วนเกินมูลค่าหุ้นสามัญ"},{"code":"en","name":"Share Premium - Common Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c5a64df9-dfcc-4a21-a60d-fedd8a431010', '3131', 1, '{"id":"c5a64df9-dfcc-4a21-a60d-fedd8a431010","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3131","names":[{"code":"th","name":"ทุนส่วนของเจ้าของ (ห้างหุ้นส่วน/บุคคลธรรมดา)"},{"code":"en","name":"Owner''s / Partner''s Capital"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '051e973d-8b40-4a17-aa4e-937209145651', '3132', 1, '{"id":"051e973d-8b40-4a17-aa4e-937209145651","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3132","names":[{"code":"th","name":"เงินถอนใช้ส่วนตัวของเจ้าของกิจการ"},{"code":"en","name":"Owner''s Drawings"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'd65ecefa-0429-451b-a6a9-ef35e74d43f6', '3200', 1, '{"id":"d65ecefa-0429-451b-a6a9-ef35e74d43f6","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3200","names":[{"code":"th","name":"กำไร (ขาดทุน) สะสม"},{"code":"en","name":"Retained Earnings"}],"accounttype":"equity","parentaccountcode":"3000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '57a872ca-ad67-425b-a3fb-c250ddaecf40', '3211', 1, '{"id":"57a872ca-ad67-425b-a3fb-c250ddaecf40","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3211","names":[{"code":"th","name":"สำรองตามกฎหมาย"},{"code":"en","name":"Legal Reserve"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '76fc9c33-48af-45c7-a635-c401e350efea', '3212', 1, '{"id":"76fc9c33-48af-45c7-a635-c401e350efea","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3212","names":[{"code":"th","name":"สำรองอื่นเพื่อวัตถุประสงค์เฉพาะ"},{"code":"en","name":"Other Appropriated Reserves"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'd85df776-be9d-46b9-a388-6df291981d0f', '3221', 1, '{"id":"d85df776-be9d-46b9-a388-6df291981d0f","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3221","names":[{"code":"th","name":"กำไร (ขาดทุน) สะสมยังไม่ได้จัดสรร"},{"code":"en","name":"Unappropriated Retained Earnings"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '965b745c-1b82-423a-a2ff-d88cce8e0ac2', '3222', 1, '{"id":"965b745c-1b82-423a-a2ff-d88cce8e0ac2","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3222","names":[{"code":"th","name":"กำไร (ขาดทุน) สุทธิประจำปี"},{"code":"en","name":"Net Profit/Loss for Current Year"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8e366d56-b562-4f67-ad7b-94cc1b026bc0', '3223', 1, '{"id":"8e366d56-b562-4f67-ad7b-94cc1b026bc0","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3223","names":[{"code":"th","name":"เงินปันผลจ่าย"},{"code":"en","name":"Dividends Declared & Paid"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ddd0dd00-896d-497b-afd8-3844cf6d4969', '4000', 1, '{"id":"ddd0dd00-896d-497b-afd8-3844cf6d4969","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4000","names":[{"code":"th","name":"รายได้"},{"code":"en","name":"Income"}],"accounttype":"income","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6303cd78-9bed-499f-aa80-370e306ef789', '4100', 1, '{"id":"6303cd78-9bed-499f-aa80-370e306ef789","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4100","names":[{"code":"th","name":"รายได้จากการขายและการให้บริการ"},{"code":"en","name":"Revenue from Sales & Services"}],"accounttype":"income","parentaccountcode":"4000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3cc38115-d8af-4712-a9cc-1c08934dc8f8', '4111', 1, '{"id":"3cc38115-d8af-4712-a9cc-1c08934dc8f8","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4111","names":[{"code":"th","name":"รายได้จากการขายสินค้า - มีภาษีมูลค่าเพิ่ม 7%"},{"code":"en","name":"Sales Revenue - VAT 7%"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'fddae0cf-8935-477d-a9d2-8e66cc0334f9', '4112', 1, '{"id":"fddae0cf-8935-477d-a9d2-8e66cc0334f9","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4112","names":[{"code":"th","name":"รายได้จากการขายสินค้า - อัตราภาษี 0% / ส่งออก"},{"code":"en","name":"Sales Revenue - Zero Rated / Export"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '73e9eea6-f616-41f9-a2fc-91c183c7d92a', '4113', 1, '{"id":"73e9eea6-f616-41f9-a2fc-91c183c7d92a","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4113","names":[{"code":"th","name":"รายได้จากการขายสินค้า - ได้รับยกเว้นภาษีมูลค่าเพิ่ม"},{"code":"en","name":"Sales Revenue - VAT Exempted"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3bb4295d-e986-48d9-aae9-2e3ab81c7f54', '4121', 1, '{"id":"3bb4295d-e986-48d9-aae9-2e3ab81c7f54","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4121","names":[{"code":"th","name":"รายได้จากการให้บริการ - มีภาษีมูลค่าเพิ่ม 7%"},{"code":"en","name":"Service Income - VAT 7%"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c32adc12-8d12-495f-a2ee-1e7048c2ee5c', '4122', 1, '{"id":"c32adc12-8d12-495f-a2ee-1e7048c2ee5c","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4122","names":[{"code":"th","name":"รายได้จากการให้บริการ - ได้รับยกเว้นภาษีมูลค่าเพิ่ม"},{"code":"en","name":"Service Income - VAT Exempted"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '85b311f1-9e4e-4baa-a776-b637c85ecf8a', '4131', 1, '{"id":"85b311f1-9e4e-4baa-a776-b637c85ecf8a","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4131","names":[{"code":"th","name":"รับคืนสินค้าและลดหนี้ขาย"},{"code":"en","name":"Sales Returns and Allowances"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a06d80cb-b397-4d99-a0d9-a00b2f25e27b', '4132', 1, '{"id":"a06d80cb-b397-4d99-a0d9-a00b2f25e27b","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4132","names":[{"code":"th","name":"ส่วนลดจ่าย"},{"code":"en","name":"Sales Cash Discounts"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '9ca50ed0-38b0-476e-adac-ef34017292ef', '4141', 1, '{"id":"9ca50ed0-38b0-476e-adac-ef34017292ef","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4141","names":[{"code":"th","name":"รายได้ค่าบริการขนส่งและจัดส่งสินค้า"},{"code":"en","name":"Freight & Delivery Income"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '28a5d96b-ba5b-40c7-a421-63aa5e10e0cc', '4200', 1, '{"id":"28a5d96b-ba5b-40c7-a421-63aa5e10e0cc","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4200","names":[{"code":"th","name":"รายได้อื่น"},{"code":"en","name":"Other Income"}],"accounttype":"income","parentaccountcode":"4000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c6a4c498-17f1-4fa0-ad60-ec0aa88974cb', '4211', 1, '{"id":"c6a4c498-17f1-4fa0-ad60-ec0aa88974cb","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4211","names":[{"code":"th","name":"ดอกเบี้ยรับจากสถาบันการเงิน"},{"code":"en","name":"Interest Income from Banks"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '174f02a9-3ce8-4da5-a545-17a39aea7e5d', '4212', 1, '{"id":"174f02a9-3ce8-4da5-a545-17a39aea7e5d","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4212","names":[{"code":"th","name":"เงินปันผลรับจากเงินลงทุน"},{"code":"en","name":"Dividend Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1b3c5976-ae36-461c-a837-3e46c32f883b', '4221', 1, '{"id":"1b3c5976-ae36-461c-a837-3e46c32f883b","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4221","names":[{"code":"th","name":"กำไรจากการจำหน่ายทรัพย์สินถาวร"},{"code":"en","name":"Gain on Disposal of Fixed Assets"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'bd4e9b44-93a3-4df8-a7c4-9b9c0ccc0679', '4231', 1, '{"id":"bd4e9b44-93a3-4df8-a7c4-9b9c0ccc0679","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4231","names":[{"code":"th","name":"กำไรจากอัตราแลกเปลี่ยนเงินตราต่างประเทศ"},{"code":"en","name":"Gain on Foreign Exchange"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3aa38ca1-3d89-4aee-aab0-1f8346701064', '4241', 1, '{"id":"3aa38ca1-3d89-4aee-aab0-1f8346701064","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4241","names":[{"code":"th","name":"รายได้ค่าเช่าอาคารและอุปกรณ์"},{"code":"en","name":"Rental Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ae112a20-ee53-4ebd-a142-4df2e4fef3b2', '4251', 1, '{"id":"ae112a20-ee53-4ebd-a142-4df2e4fef3b2","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4251","names":[{"code":"th","name":"หนี้สูญได้รับคืน"},{"code":"en","name":"Bad Debts Recovered"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '9b9583ab-0d80-470e-a9f8-c0a7813e9259', '4291', 1, '{"id":"9b9583ab-0d80-470e-a9f8-c0a7813e9259","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4291","names":[{"code":"th","name":"รายได้เบ็ดเตล็ดอื่น"},{"code":"en","name":"Miscellaneous Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '309271a2-25b1-445c-a8ca-3eae8987bfa1', '5000', 1, '{"id":"309271a2-25b1-445c-a8ca-3eae8987bfa1","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5000","names":[{"code":"th","name":"ค่าใช้จ่าย"},{"code":"en","name":"Expenses"}],"accounttype":"expense","parentaccountcode":null,"normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a7fc49bf-ac47-4020-a3f2-6b16e0053250', '5100', 1, '{"id":"a7fc49bf-ac47-4020-a3f2-6b16e0053250","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5100","names":[{"code":"th","name":"ต้นทุนขายและบริการ"},{"code":"en","name":"Cost of Goods Sold & Services"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a92f693e-8c58-4a8e-aa2c-e80dd5df688a', '5111', 1, '{"id":"a92f693e-8c58-4a8e-aa2c-e80dd5df688a","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5111","names":[{"code":"th","name":"ซื้อสินค้าสำเร็จรูป"},{"code":"en","name":"Purchases of Merchandise"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a92fddc3-0c69-450c-a45e-29ae47b39a90', '5112', 1, '{"id":"a92fddc3-0c69-450c-a45e-29ae47b39a90","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5112","names":[{"code":"th","name":"ค่าขนส่งเข้า"},{"code":"en","name":"Freight-In"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '679176c5-3e76-471e-a1c4-3a1594db0dcb', '5113', 1, '{"id":"679176c5-3e76-471e-a1c4-3a1594db0dcb","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5113","names":[{"code":"th","name":"ส่งคืนสินค้าและส่วนลดที่ได้รับ"},{"code":"en","name":"Purchase Returns and Allowances"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f718fe64-ff8c-48c9-ae09-eb2cf1a0cf12', '5114', 1, '{"id":"f718fe64-ff8c-48c9-ae09-eb2cf1a0cf12","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5114","names":[{"code":"th","name":"ส่วนลดรับ"},{"code":"en","name":"Purchase Cash Discounts"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6752d344-5d2b-4bd1-a7e0-201574785e63', '5121', 1, '{"id":"6752d344-5d2b-4bd1-a7e0-201574785e63","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5121","names":[{"code":"th","name":"ต้นทุนสินค้าสำเร็จรูปที่ขาย"},{"code":"en","name":"Cost of Goods Sold"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '154350b8-e6f1-49aa-abb0-89d4c7503f48', '5131', 1, '{"id":"154350b8-e6f1-49aa-abb0-89d4c7503f48","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5131","names":[{"code":"th","name":"ต้นทุนค่าแรงและบริการ"},{"code":"en","name":"Direct Labor & Service Costs"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a7a93005-e1a4-4294-a540-33ce1e0d760f', '5132', 1, '{"id":"a7a93005-e1a4-4294-a540-33ce1e0d760f","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5132","names":[{"code":"th","name":"ค่าจ้างเหมาบริการช่วงภายนอก (Subcontractor)"},{"code":"en","name":"Subcontractor Costs"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '884a4111-2020-4ef5-a9be-38fe65e99926', '5141', 1, '{"id":"884a4111-2020-4ef5-a9be-38fe65e99926","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5141","names":[{"code":"th","name":"สินค้าสูญหาย เสียหาย และสินค้าชำรุด"},{"code":"en","name":"Inventory Shrinkage & Damage"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '95d388ab-5631-4cc1-ad0f-6b76b95f3249', '5200', 1, '{"id":"95d388ab-5631-4cc1-ad0f-6b76b95f3249","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5200","names":[{"code":"th","name":"ค่าใช้จ่ายในการขาย"},{"code":"en","name":"Selling Expenses"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '01c63a68-c69d-4fb8-a1f9-88e26864bd54', '5211', 1, '{"id":"01c63a68-c69d-4fb8-a1f9-88e26864bd54","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5211","names":[{"code":"th","name":"เงินเดือน ค่าล่วงเวลา และเบี้ยเลี้ยงฝ่ายขาย"},{"code":"en","name":"Sales Salaries, Overtime & Allowances"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a0910589-0ab8-4c43-aac7-8a38cd0c89f3', '5212', 1, '{"id":"a0910589-0ab8-4c43-aac7-8a38cd0c89f3","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5212","names":[{"code":"th","name":"ค่านายหน้าและค่าคอมมิชชั่นฝ่ายขาย"},{"code":"en","name":"Sales Commissions & Incentives"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3d478b6f-5e84-41de-a487-b47917b2d6a8', '5221', 1, '{"id":"3d478b6f-5e84-41de-a487-b47917b2d6a8","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5221","names":[{"code":"th","name":"ค่าโฆษณา ประชาสัมพันธ์ และการตลาดออนไลน์"},{"code":"en","name":"Advertising, PR & Online Marketing"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '104c90f2-a956-4eba-aaaf-72c2fd0abfb0', '5222', 1, '{"id":"104c90f2-a956-4eba-aaaf-72c2fd0abfb0","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5222","names":[{"code":"th","name":"ค่าส่งเสริมการขาย ของแถม และตัวอย่างสินค้า"},{"code":"en","name":"Sales Promotion, Gifts & Samples"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '917bd7f9-81cf-4b55-a9c3-b53f8737108e', '5231', 1, '{"id":"917bd7f9-81cf-4b55-a9c3-b53f8737108e","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5231","names":[{"code":"th","name":"ค่าขนส่งสินค้าออกและบริการจัดส่งให้ลูกค้า"},{"code":"en","name":"Outward Freight & Shipping to Customers"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'bfa6d2d3-a7ec-4a6b-a71e-1b8a304e7137', '5232', 1, '{"id":"bfa6d2d3-a7ec-4a6b-a71e-1b8a304e7137","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5232","names":[{"code":"th","name":"ค่าวัสดุหีบห่อและบรรจุภัณฑ์สำหรับการขาย"},{"code":"en","name":"Packaging & Packing Materials"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f0068e5b-71ba-40bf-a8a7-68e4894b24fe', '5241', 1, '{"id":"f0068e5b-71ba-40bf-a8a7-68e4894b24fe","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5241","names":[{"code":"th","name":"ค่าจัดงานแสดงสินค้าและการออกบูธ"},{"code":"en","name":"Exhibition & Trade Fair Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'fd9e0c57-9713-46ff-ad8c-5edb35b51534', '5251', 1, '{"id":"fd9e0c57-9713-46ff-ad8c-5edb35b51534","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5251","names":[{"code":"th","name":"ค่าน้ำมันและค่าเดินทางพบลูกค้าฝ่ายขาย"},{"code":"en","name":"Sales Fuel & Travel Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '9209f5f0-b240-481e-a451-d535b446e51a', '5291', 1, '{"id":"9209f5f0-b240-481e-a451-d535b446e51a","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5291","names":[{"code":"th","name":"ค่าใช้จ่ายในการขายอื่น"},{"code":"en","name":"Other Selling Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '923b5e17-0cfd-42ad-aeb8-9c006fa4d39f', '5300', 1, '{"id":"923b5e17-0cfd-42ad-aeb8-9c006fa4d39f","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5300","names":[{"code":"th","name":"ค่าใช้จ่ายในการบริหาร"},{"code":"en","name":"Administrative Expenses"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '753437b9-7c22-430d-a847-baf176372f8d', '5311', 1, '{"id":"753437b9-7c22-430d-a847-baf176372f8d","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5311","names":[{"code":"th","name":"เงินเดือน ค่าจ้าง และค่าล่วงเวลาพนักงานสำนักงาน"},{"code":"en","name":"Office Salaries, Wages & Overtime"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '53007615-b1ac-4f23-a597-9f7ef712db7b', '5312', 1, '{"id":"53007615-b1ac-4f23-a597-9f7ef712db7b","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5312","names":[{"code":"th","name":"ค่าตอบแทนและเบี้ยประชุมกรรมการ"},{"code":"en","name":"Directors'' Remuneration & Meeting Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '7078a484-8c0a-41e6-ac2c-620e01266ad2', '5313', 1, '{"id":"7078a484-8c0a-41e6-ac2c-620e01266ad2","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5313","names":[{"code":"th","name":"โบนัสพนักงานประจำปี"},{"code":"en","name":"Annual Staff Bonuses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6e64c30b-e2b9-42bf-a398-95ade465308b', '5314', 1, '{"id":"6e64c30b-e2b9-42bf-a398-95ade465308b","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5314","names":[{"code":"th","name":"เงินสมทบกองทุนประกันสังคม (ส่วนของนายจ้าง)"},{"code":"en","name":"Social Security Fund - Employer''s Contribution"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2bf07f82-d30a-4345-a255-f7a68d979d82', '5315', 1, '{"id":"2bf07f82-d30a-4345-a255-f7a68d979d82","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5315","names":[{"code":"th","name":"เงินสมทบกองทุนเงินทดแทนและกองทุนสำรองเลี้ยงชีพ"},{"code":"en","name":"Workmen Compensation & Provident Fund"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '28155529-dcf9-4d7b-a4e7-75a157729aa6', '5316', 1, '{"id":"28155529-dcf9-4d7b-a4e7-75a157729aa6","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5316","names":[{"code":"th","name":"ค่าสวัสดิการพนักงานและชุดยูนิฟอร์ม"},{"code":"en","name":"Staff Welfare & Uniforms"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c8155249-fb33-4ce3-a644-b971afec3949', '5317', 1, '{"id":"c8155249-fb33-4ce3-a644-b971afec3949","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5317","names":[{"code":"th","name":"ค่าฝึกอบรมและสัมมนาบุคลากร"},{"code":"en","name":"Training & Seminar Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c3d32ae3-b1f5-43fa-a509-4a0fa2aed339', '5321', 1, '{"id":"c3d32ae3-b1f5-43fa-a509-4a0fa2aed339","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5321","names":[{"code":"th","name":"ค่าเช่าอาคารสำนักงานและพื้นที่ประกอบการ"},{"code":"en","name":"Office Building Rent"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '68d3e2f3-817c-4160-a874-735a813765d9', '5322', 1, '{"id":"68d3e2f3-817c-4160-a874-735a813765d9","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5322","names":[{"code":"th","name":"ค่าน้ำประปา"},{"code":"en","name":"Water Utility Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '91ecd071-4b2d-4acb-aa79-b7f18afa0a61', '5323', 1, '{"id":"91ecd071-4b2d-4acb-aa79-b7f18afa0a61","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5323","names":[{"code":"th","name":"ค่าไฟฟ้า"},{"code":"en","name":"Electricity Utility Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3b52e34d-3445-498c-a0dc-b8b63b02c169', '5324', 1, '{"id":"3b52e34d-3445-498c-a0dc-b8b63b02c169","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5324","names":[{"code":"th","name":"ค่าโทรศัพท์และค่าบริการโทรคมนาคม"},{"code":"en","name":"Telephone & Telecom Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2abf8c77-8537-453f-aa77-d2f82c15ff7c', '5325', 1, '{"id":"2abf8c77-8537-453f-aa77-d2f82c15ff7c","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5325","names":[{"code":"th","name":"ค่าบริการอินเทอร์เน็ต ระบบเซิร์ฟเวอร์ และคลาวด์"},{"code":"en","name":"Internet, Server & Cloud Hosting Services"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '10c3acbc-687c-4de7-a505-a949b42a4865', '5331', 1, '{"id":"10c3acbc-687c-4de7-a505-a949b42a4865","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5331","names":[{"code":"th","name":"ค่าเครื่องเขียน แบบพิมพ์ และวัสดุสำนักงาน"},{"code":"en","name":"Stationery, Printing & Office Supplies"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a17249f9-620c-4705-ac65-ad5ddb9b3bb9', '5332', 1, '{"id":"a17249f9-620c-4705-ac65-ad5ddb9b3bb9","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5332","names":[{"code":"th","name":"ค่าไปรษณีย์และค่าส่งเอกสารพัสดุ"},{"code":"en","name":"Postal & Courier Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '19e5a64a-6dc6-4ff9-ad4b-37c670b056cc', '5341', 1, '{"id":"19e5a64a-6dc6-4ff9-ad4b-37c670b056cc","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5341","names":[{"code":"th","name":"ค่าตรวจสอบบัญชี (Audit Fees)"},{"code":"en","name":"Audit Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '12507632-5205-43d4-a985-316656d207c4', '5342', 1, '{"id":"12507632-5205-43d4-a985-316656d207c4","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5342","names":[{"code":"th","name":"ค่าจัดทำบัญชีและที่ปรึกษาภาษีอากร"},{"code":"en","name":"Accounting & Tax Consultation Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '9838362c-ee45-4197-a85b-352748569d0d', '5343', 1, '{"id":"9838362c-ee45-4197-a85b-352748569d0d","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5343","names":[{"code":"th","name":"ค่าธรรมเนียมวิชาชีพกฎหมายและที่ปรึกษาธุรกิจ"},{"code":"en","name":"Legal & Business Advisory Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '63bffa24-9a65-45b0-aeb7-baec15b51dad', '5351', 1, '{"id":"63bffa24-9a65-45b0-aeb7-baec15b51dad","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5351","names":[{"code":"th","name":"ค่าเบี้ยประกันภัยทรัพย์สิน อาคาร และยานพาหนะ"},{"code":"en","name":"Property, Building & Vehicle Insurance"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '87fef709-7a53-471d-a4c5-6551224b1660', '5352', 1, '{"id":"87fef709-7a53-471d-a4c5-6551224b1660","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5352","names":[{"code":"th","name":"ค่าซ่อมแซมและบำรุงรักษาอาคารและอุปกรณ์"},{"code":"en","name":"Repairs & Maintenance Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '5a3511eb-074a-440a-ab25-f2f02f0361e2', '5353', 1, '{"id":"5a3511eb-074a-440a-ab25-f2f02f0361e2","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5353","names":[{"code":"th","name":"ค่าบริการทำความสะอาดและรักษาความปลอดภัย"},{"code":"en","name":"Cleaning & Security Services"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '78a481c7-3c55-4121-a71f-f9478988a4da', '5361', 1, '{"id":"78a481c7-3c55-4121-a71f-f9478988a4da","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5361","names":[{"code":"th","name":"ค่าเสื่อมราคา - ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Depreciation - Land Improvements"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '369c00b1-0e42-4930-a185-144021a10ac0', '5362', 1, '{"id":"369c00b1-0e42-4930-a185-144021a10ac0","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5362","names":[{"code":"th","name":"ค่าเสื่อมราคา - อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Depreciation - Buildings"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1f5a7f46-8734-4486-a0ee-d7a23d98bc02', '5363', 1, '{"id":"1f5a7f46-8734-4486-a0ee-d7a23d98bc02","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5363","names":[{"code":"th","name":"ค่าเสื่อมราคา - ส่วนปรับปรุงอาคาร"},{"code":"en","name":"Depreciation - Building Improvements"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '948daa3b-cb61-4e79-a9e4-22adb26dc2ed', '5364', 1, '{"id":"948daa3b-cb61-4e79-a9e4-22adb26dc2ed","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5364","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องจักรและอุปกรณ์"},{"code":"en","name":"Depreciation - Machinery & Equipment"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a37d47c6-c789-4176-a2ff-1c2ad8b07a26', '5365', 1, '{"id":"a37d47c6-c789-4176-a2ff-1c2ad8b07a26","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5365","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องตกแต่งและติดตั้ง"},{"code":"en","name":"Depreciation - Furniture & Fixtures"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '285455c7-0876-4aa4-a555-f66d5189e416', '5366', 1, '{"id":"285455c7-0876-4aa4-a555-f66d5189e416","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5366","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องใช้สำนักงาน"},{"code":"en","name":"Depreciation - Office Equipment"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3521afee-624a-41d3-a08a-54d6bc6bd362', '5367', 1, '{"id":"3521afee-624a-41d3-a08a-54d6bc6bd362","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5367","names":[{"code":"th","name":"ค่าเสื่อมราคา - คอมพิวเตอร์และอุปกรณ์"},{"code":"en","name":"Depreciation - Computer & Hardware"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e3addc0d-796e-4127-aefa-5e49935f5d20', '5368', 1, '{"id":"e3addc0d-796e-4127-aefa-5e49935f5d20","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5368","names":[{"code":"th","name":"ค่าเสื่อมราคา - ยานพาหนะ"},{"code":"en","name":"Depreciation - Vehicles"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '5d2da401-65ee-4c05-ab93-b2c20098eb08', '5371', 1, '{"id":"5d2da401-65ee-4c05-ab93-b2c20098eb08","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5371","names":[{"code":"th","name":"ค่าตัดจำหน่ายโปรแกรมคอมพิวเตอร์และซอฟต์แวร์"},{"code":"en","name":"Amortization - Software & Applications"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '33545e99-c27b-48ae-ad33-2b3a643c89bd', '5381', 1, '{"id":"33545e99-c27b-48ae-ad33-2b3a643c89bd","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5381","names":[{"code":"th","name":"ค่าธรรมเนียมราชการและใบอนุญาตประกอบกิจการ"},{"code":"en","name":"Government Licenses & Registration Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ee59a6dc-7529-4d6c-afe2-a0ffee8f8d91', '5382', 1, '{"id":"ee59a6dc-7529-4d6c-afe2-a0ffee8f8d91","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5382","names":[{"code":"th","name":"ภาษีที่ดินและสิ่งปลูกสร้าง"},{"code":"en","name":"Land and Building Taxes"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f0e32323-b626-4dac-a274-2f88d2ad2177', '5383', 1, '{"id":"f0e32323-b626-4dac-a274-2f88d2ad2177","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5383","names":[{"code":"th","name":"ภาษีป้าย"},{"code":"en","name":"Signboard Tax"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8363765f-c62a-4ab0-aec8-2ad8a14ba0db', '5384', 1, '{"id":"8363765f-c62a-4ab0-aec8-2ad8a14ba0db","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5384","names":[{"code":"th","name":"ภาษีซื้อต้องห้าม / ภาษีซื้อที่ไม่สามารถขอคืนได้"},{"code":"en","name":"Non-refundable Input Tax (Disallowed Input VAT)"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c3f140d0-7bcc-4341-a8f3-6075f4b7490c', '5385', 1, '{"id":"c3f140d0-7bcc-4341-a8f3-6075f4b7490c","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5385","names":[{"code":"th","name":"เบี้ยปรับ เงินเพิ่ม และค่าปรับทางภาษีอากร"},{"code":"en","name":"Tax Penalties, Fines & Surcharges"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a66229ac-64a9-4728-a293-4df73c3b35c8', '5391', 1, '{"id":"a66229ac-64a9-4728-a293-4df73c3b35c8","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5391","names":[{"code":"th","name":"ค่ารับรองและบริการลูกค้า"},{"code":"en","name":"Entertainment & Hospitality Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '4906966b-5c59-4c7f-a089-35f4971c5fb5', '5392', 1, '{"id":"4906966b-5c59-4c7f-a089-35f4971c5fb5","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5392","names":[{"code":"th","name":"เงินบริจาคเพื่อการกุศลและการศึกษา"},{"code":"en","name":"Charitable & Educational Donations"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a7e4432f-df5c-429f-a111-548c179b577d', '5393', 1, '{"id":"a7e4432f-df5c-429f-a111-548c179b577d","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5393","names":[{"code":"th","name":"หนี้สูญและหนี้สงสัยจะสูญ"},{"code":"en","name":"Bad Debts & Doubtful Accounts Expense"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '92559c8c-839a-486e-a896-f3268a50e040', '5394', 1, '{"id":"92559c8c-839a-486e-a896-f3268a50e040","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5394","names":[{"code":"th","name":"ขาดทุนจากอัตราแลกเปลี่ยนเงินตราต่างประเทศ"},{"code":"en","name":"Loss on Foreign Exchange"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'bb20210e-4856-4ee5-a182-770c86290f7e', '5395', 1, '{"id":"bb20210e-4856-4ee5-a182-770c86290f7e","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5395","names":[{"code":"th","name":"ขาดทุนจากการจำหน่ายและตัดจำหน่ายทรัพย์สิน"},{"code":"en","name":"Loss on Disposal and Write-off of Assets"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '71c18e1d-b88a-43f4-a2ac-cfb1d8fdef2c', '5399', 1, '{"id":"71c18e1d-b88a-43f4-a2ac-cfb1d8fdef2c","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5399","names":[{"code":"th","name":"ค่าใช้จ่ายในการบริหารอื่น"},{"code":"en","name":"Other Administrative Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '69b96417-e8a4-4465-af72-9f9da432b176', '5400', 1, '{"id":"69b96417-e8a4-4465-af72-9f9da432b176","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5400","names":[{"code":"th","name":"ต้นทุนทางการเงินและภาษีเงินได้"},{"code":"en","name":"Financial Costs & Income Tax"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '87525568-cf4a-41ec-a355-366aeaeb1039', '5411', 1, '{"id":"87525568-cf4a-41ec-a355-366aeaeb1039","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5411","names":[{"code":"th","name":"ดอกเบี้ยจ่ายเงินกู้ยืมสถาบันการเงิน"},{"code":"en","name":"Interest Expense - Bank Loans"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2181aeab-0d1b-40d4-a8bc-30eb51ac8c53', '5412', 1, '{"id":"2181aeab-0d1b-40d4-a8bc-30eb51ac8c53","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5412","names":[{"code":"th","name":"ดอกเบี้ยจ่ายเงินเบิกเกินบัญชี (O/D)"},{"code":"en","name":"Interest Expense - Bank Overdraft (O/D)"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'cf0a65e1-21d8-46c7-a8a5-0c6c57f70d71', '5413', 1, '{"id":"cf0a65e1-21d8-46c7-a8a5-0c6c57f70d71","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5413","names":[{"code":"th","name":"ดอกเบี้ยจ่ายตามสัญญาเช่าทางการเงิน"},{"code":"en","name":"Interest Expense - Financial Leases"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '18562edc-e5f4-40dc-add0-651dcfbc1d47', '5421', 1, '{"id":"18562edc-e5f4-40dc-add0-651dcfbc1d47","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5421","names":[{"code":"th","name":"ค่าธรรมเนียมธนาคารและธุรกรรมทางการเงิน"},{"code":"en","name":"Bank Charges & Transaction Fees"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ae68c1c9-783e-48ad-a39c-4eb5426df4cf', '5431', 1, '{"id":"ae68c1c9-783e-48ad-a39c-4eb5426df4cf","holdingcode":"rungrueng","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5431","names":[{"code":"th","name":"ภาษีเงินได้นิติบุคคลประจำงวด"},{"code":"en","name":"Corporate Income Tax Expense"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

COMMIT;
