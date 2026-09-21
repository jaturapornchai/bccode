-- ============================================================================
-- Thai Standard Chart of Accounts for Holding: demo (3 Levels, 187 Accounts)
-- Compatible with Thai Accounting Standards (TFRS for NPAEs / DBD / RD)
-- Wrapped in transaction for ACID safety
-- ============================================================================

BEGIN;


-- ============================================================================
-- Holding: demo
-- ============================================================================
INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'fiscal-years', 'cd52a0fb-ee8c-4694-a9ae-225f87fa8626', '2569', 1, '{"id":"cd52a0fb-ee8c-4694-a9ae-225f87fa8626","code":"2569","kind":"fiscal-years","startdate":"2026-01-01","enddate":"2026-12-31","scale":2,"profitlossaccount":"3222","retainedearningsaccount":"3221","closed":false,"isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', '6fc89197-b70b-40e7-a4d5-be8a914f62e6', 'JV', 1, '{"id":"6fc89197-b70b-40e7-a4d5-be8a914f62e6","code":"JV","kind":"journal-books","name":"สมุดรายวันทั่วไป","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', 'a4645f9b-05b3-4f3e-ab5a-3f7f37a7b692', 'PV', 1, '{"id":"a4645f9b-05b3-4f3e-ab5a-3f7f37a7b692","code":"PV","kind":"journal-books","name":"สมุดรายวันจ่าย","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', '3136f2eb-3b5a-468d-a830-7014550b2e6e', 'RV', 1, '{"id":"3136f2eb-3b5a-468d-a830-7014550b2e6e","code":"RV","kind":"journal-books","name":"สมุดรายวันรับ","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', '0eac5830-2fcc-418a-a5e2-891ed68bcbce', 'SV', 1, '{"id":"0eac5830-2fcc-418a-a5e2-891ed68bcbce","code":"SV","kind":"journal-books","name":"สมุดรายวันซื้อ","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', 'b5cac34c-945f-4cc4-a6f7-968516d0a332', 'UV', 1, '{"id":"b5cac34c-945f-4cc4-a6f7-968516d0a332","code":"UV","kind":"journal-books","name":"สมุดรายวันขาย","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', '4ef2748d-a382-475f-ae59-61e56ee9af9a', 'GJ', 1, '{"id":"4ef2748d-a382-475f-ae59-61e56ee9af9a","code":"GJ","kind":"journal-books","name":"สมุดรายวันทั่วไป (GJ)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', '79827d8c-d971-4354-a926-ed8bc56b8717', 'AP', 1, '{"id":"79827d8c-d971-4354-a926-ed8bc56b8717","code":"AP","kind":"journal-books","name":"สมุดรายวันซื้อเชื่อ (AP)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'journal-books', 'fae573e3-9c35-435d-a16d-4d0f28096f89', 'AR', 1, '{"id":"fae573e3-9c35-435d-a16d-4d0f28096f89","code":"AR","kind":"journal-books","name":"สมุดรายวันขายเชื่อ (AR)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'account-groups', '7cb70afa-f5ae-46f4-a9f2-7b0030b0c905', 'BM69-A', 1, '{"id":"7cb70afa-f5ae-46f4-a9f2-7b0030b0c905","code":"BM69-A","kind":"account-groups","name":"สินทรัพย์","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'account-groups', '026b2f81-e447-4ad1-a18e-c81a4df09f4d', 'BM69-L', 1, '{"id":"026b2f81-e447-4ad1-a18e-c81a4df09f4d","code":"BM69-L","kind":"account-groups","name":"หนี้สิน","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'account-groups', 'ef72c188-d210-46cb-afae-e9db2a21bae4', 'BM69-E', 1, '{"id":"ef72c188-d210-46cb-afae-e9db2a21bae4","code":"BM69-E","kind":"account-groups","name":"ส่วนของเจ้าของ","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'account-groups', '0feb46c5-4e8c-42ed-aded-97995439666c', 'BM69-R', 1, '{"id":"0feb46c5-4e8c-42ed-aded-97995439666c","code":"BM69-R","kind":"account-groups","name":"รายได้","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'account-groups', '7bf6d72a-32eb-45cd-a273-2062b431f5bb', 'BM69-X', 1, '{"id":"7bf6d72a-32eb-45cd-a273-2062b431f5bb","code":"BM69-X","kind":"account-groups","name":"ค่าใช้จ่าย","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"demo","businesscode":"01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;


-- Accounts for Holding: demo, Company: 01
INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '12e1a26e-699a-4880-a221-1412cfe7284d', '1000', 1, '{"id":"12e1a26e-699a-4880-a221-1412cfe7284d","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1000","names":[{"code":"th","name":"สินทรัพย์"},{"code":"en","name":"Assets"}],"accounttype":"asset","parentaccountcode":null,"normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a4be3f6a-1b8a-4ed7-ada5-238e25d1ca5c', '1100', 1, '{"id":"a4be3f6a-1b8a-4ed7-ada5-238e25d1ca5c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1100","names":[{"code":"th","name":"สินทรัพย์หมุนเวียน"},{"code":"en","name":"Current Assets"}],"accounttype":"asset","parentaccountcode":"1000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0396afc4-625d-4a4b-a3c2-ff19c07b69e9', '1111', 1, '{"id":"0396afc4-625d-4a4b-a3c2-ff19c07b69e9","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1111","names":[{"code":"th","name":"เงินสดในมือ"},{"code":"en","name":"Cash on Hand"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '86646975-3b76-4690-adef-a336e2a54a8d', '1112', 1, '{"id":"86646975-3b76-4690-adef-a336e2a54a8d","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1112","names":[{"code":"th","name":"เงินสดย่อย - สำนักงานใหญ่"},{"code":"en","name":"Petty Cash - Head Office"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f0614c06-5c0d-4e7a-a40c-99b79d0aa6c5', '1113', 1, '{"id":"f0614c06-5c0d-4e7a-a40c-99b79d0aa6c5","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1113","names":[{"code":"th","name":"เงินสดย่อย - ฝ่ายปฏิบัติการและสาขา"},{"code":"en","name":"Petty Cash - Operations & Branch"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8130c206-20e6-4ffd-a389-809246aa164e', '1121', 1, '{"id":"8130c206-20e6-4ffd-a389-809246aa164e","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1121","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย"},{"code":"en","name":"Cash at Bank - Current KBANK"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9f686fac-a855-4222-a02e-f505c5b9f23f', '1122', 1, '{"id":"9f686fac-a855-4222-a02e-f505c5b9f23f","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1122","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารกรุงเทพ"},{"code":"en","name":"Cash at Bank - Current BBL"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '31848cad-16d0-4458-a4b0-463bfd11d7d6', '1123', 1, '{"id":"31848cad-16d0-4458-a4b0-463bfd11d7d6","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1123","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารไทยพาณิชย์"},{"code":"en","name":"Cash at Bank - Current SCB"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '76b5a521-6d00-4de7-acca-8eccd83d3285', '1124', 1, '{"id":"76b5a521-6d00-4de7-acca-8eccd83d3285","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1124","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารกสิกรไทย"},{"code":"en","name":"Cash at Bank - Savings KBANK"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a0a9839e-adb4-4248-a02c-a7484178ccea', '1125', 1, '{"id":"a0a9839e-adb4-4248-a02c-a7484178ccea","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1125","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารกรุงเทพ"},{"code":"en","name":"Cash at Bank - Savings BBL"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '2c9d1ae3-45ea-41af-a555-53899ee3d25c', '1126', 1, '{"id":"2c9d1ae3-45ea-41af-a555-53899ee3d25c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1126","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารไทยพาณิชย์"},{"code":"en","name":"Cash at Bank - Savings SCB"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd7a90de3-a69f-4ed1-a81c-526d6c53ffb4', '1127', 1, '{"id":"d7a90de3-a69f-4ed1-a81c-526d6c53ffb4","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1127","names":[{"code":"th","name":"เงินฝากประจำระยะสั้น (ไม่เกิน 3 เดือน)"},{"code":"en","name":"Short-term Fixed Deposit"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '921c08ec-9c22-4905-ae20-52a1229214ee', '1131', 1, '{"id":"921c08ec-9c22-4905-ae20-52a1229214ee","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1131","names":[{"code":"th","name":"ลูกหนี้การค้า - ในประเทศ"},{"code":"en","name":"Trade Accounts Receivable - Domestic"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'eb3a084e-34ae-4a89-ad49-ac713a730e83', '1132', 1, '{"id":"eb3a084e-34ae-4a89-ad49-ac713a730e83","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1132","names":[{"code":"th","name":"ลูกหนี้การค้า - ต่างประเทศ"},{"code":"en","name":"Trade Accounts Receivable - Overseas"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'cb9c689d-a092-4fc7-a42c-d7db5d698210', '1133', 1, '{"id":"cb9c689d-a092-4fc7-a42c-d7db5d698210","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1133","names":[{"code":"th","name":"ตั๋วเงินรับการค้า"},{"code":"en","name":"Notes Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e594ef80-db21-42e1-a15a-938d6ab30d8b', '1134', 1, '{"id":"e594ef80-db21-42e1-a15a-938d6ab30d8b","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1134","names":[{"code":"th","name":"เช็ครับลงวันที่ล่วงหน้า"},{"code":"en","name":"Post-dated Cheques Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '92532d3d-9e78-48b2-a298-8039517b71ff', '1138', 1, '{"id":"92532d3d-9e78-48b2-a298-8039517b71ff","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1138","names":[{"code":"th","name":"ลูกหนี้อื่นและเงินยืมทดรอง"},{"code":"en","name":"Other Receivables & Advances"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b958c3a8-5f93-43a0-a3b6-7d56ab0c97f1', '1139', 1, '{"id":"b958c3a8-5f93-43a0-a3b6-7d56ab0c97f1","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1139","names":[{"code":"th","name":"ค่าเผื่อผลขาดทุนด้านเครดิตที่คาดว่าจะเกิดขึ้น"},{"code":"en","name":"Allowance for Expected Credit Losses"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd095166f-c921-4d92-a8df-221ee600c9da', '1141', 1, '{"id":"d095166f-c921-4d92-a8df-221ee600c9da","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1141","names":[{"code":"th","name":"สินค้าสำเร็จรูป"},{"code":"en","name":"Finished Goods"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e9406262-8c7a-4807-a294-7aa99035d4e1', '1142', 1, '{"id":"e9406262-8c7a-4807-a294-7aa99035d4e1","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1142","names":[{"code":"th","name":"สินค้าระหว่างทำ / งานระหว่างทำ"},{"code":"en","name":"Work in Process"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5fe892d1-0972-4d49-ac69-95abd114e5e0', '1143', 1, '{"id":"5fe892d1-0972-4d49-ac69-95abd114e5e0","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1143","names":[{"code":"th","name":"วัตถุดิบและส่วนประกอบ"},{"code":"en","name":"Raw Materials & Components"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e1b15fbe-b997-4bfa-a13c-f06f5f86bbc8', '1144', 1, '{"id":"e1b15fbe-b997-4bfa-a13c-f06f5f86bbc8","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1144","names":[{"code":"th","name":"วัสดุสิ้นเปลืองและบรรจุภัณฑ์"},{"code":"en","name":"Factory Supplies & Packaging"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b2781d24-bec7-4699-afb7-d6a63bb57278', '1145', 1, '{"id":"b2781d24-bec7-4699-afb7-d6a63bb57278","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1145","names":[{"code":"th","name":"สินค้าระหว่างทาง"},{"code":"en","name":"Goods in Transit"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '743864e0-e1d4-4aaa-ae59-ffc5351aab85', '1149', 1, '{"id":"743864e0-e1d4-4aaa-ae59-ffc5351aab85","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1149","names":[{"code":"th","name":"ค่าเผื่อการลดมูลค่าสินค้าคงเหลือ"},{"code":"en","name":"Allowance for Inventory Devaluation"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5c069b85-8fe0-4695-a438-841553447341', '1151', 1, '{"id":"5c069b85-8fe0-4695-a438-841553447341","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1151","names":[{"code":"th","name":"ภาษีซื้อ"},{"code":"en","name":"Input VAT"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'cb91f59d-7abb-4b9e-a4bc-e06f4637ed9e', '1152', 1, '{"id":"cb91f59d-7abb-4b9e-a4bc-e06f4637ed9e","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1152","names":[{"code":"th","name":"ภาษีซื้อยังไม่ถึงกำหนดชำระ"},{"code":"en","name":"Undue Input VAT"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1d7f5eb6-8fbe-44c8-a9e6-b40fd938e42b', '1153', 1, '{"id":"1d7f5eb6-8fbe-44c8-a9e6-b40fd938e42b","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1153","names":[{"code":"th","name":"ภาษีเงินได้ถูกหัก ณ ที่จ่าย (WHT ถูกหัก)"},{"code":"en","name":"Withholding Tax Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5f6638c0-bb2c-4ab7-a7e2-842b34a4d923', '1154', 1, '{"id":"5f6638c0-bb2c-4ab7-a7e2-842b34a4d923","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1154","names":[{"code":"th","name":"ภาษีเงินได้นิติบุคคลจ่ายล่วงหน้า (ภ.ง.ด.51)"},{"code":"en","name":"Prepaid Corporate Income Tax"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'bcf65c57-4807-49ee-a32f-efa54ac073cc', '1161', 1, '{"id":"bcf65c57-4807-49ee-a32f-efa54ac073cc","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1161","names":[{"code":"th","name":"ค่าเบี้ยประกันภัยจ่ายล่วงหน้า"},{"code":"en","name":"Prepaid Insurance"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c48b27f5-79b5-4b81-ae51-5816ba44e830', '1162', 1, '{"id":"c48b27f5-79b5-4b81-ae51-5816ba44e830","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1162","names":[{"code":"th","name":"ค่าเช่าจ่ายล่วงหน้า"},{"code":"en","name":"Prepaid Rent"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'dfe0ed8a-afb0-4f96-adb7-37f998f7c61e', '1163', 1, '{"id":"dfe0ed8a-afb0-4f96-adb7-37f998f7c61e","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1163","names":[{"code":"th","name":"เงินมัดจำค่าสินค้าและบริการล่วงหน้า"},{"code":"en","name":"Advance Payments for Goods & Services"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '8949ae3a-d8a5-4a6d-a763-8d4d93f78df5', '1169', 1, '{"id":"8949ae3a-d8a5-4a6d-a763-8d4d93f78df5","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1169","names":[{"code":"th","name":"สินทรัพย์หมุนเวียนอื่น"},{"code":"en","name":"Other Current Assets"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5986fcd4-cc94-491b-a0fb-45d5fa87d7d2', '1200', 1, '{"id":"5986fcd4-cc94-491b-a0fb-45d5fa87d7d2","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1200","names":[{"code":"th","name":"สินทรัพย์ไม่หมุนเวียน"},{"code":"en","name":"Non-Current Assets"}],"accounttype":"asset","parentaccountcode":"1000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0cbf6a5f-3195-4c8c-afa2-9df13313a710', '1211', 1, '{"id":"0cbf6a5f-3195-4c8c-afa2-9df13313a710","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1211","names":[{"code":"th","name":"ที่ดิน"},{"code":"en","name":"Land"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '39bd3210-2a19-4fe9-a6e1-781ffc90036c', '1212', 1, '{"id":"39bd3210-2a19-4fe9-a6e1-781ffc90036c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1212","names":[{"code":"th","name":"ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Land Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '58ad645b-71cb-447a-a555-1d35db63735e', '1213', 1, '{"id":"58ad645b-71cb-447a-a555-1d35db63735e","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1213","names":[{"code":"th","name":"อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Buildings & Constructions"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd10c6395-946c-40fa-ad47-651415733d0f', '1214', 1, '{"id":"d10c6395-946c-40fa-ad47-651415733d0f","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1214","names":[{"code":"th","name":"ส่วนปรับปรุงอาคารและระบบสาธารณูปโภค"},{"code":"en","name":"Building Improvements & Systems"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'de28e387-4f46-4ad0-afb9-42f3d1b4bcb9', '1221', 1, '{"id":"de28e387-4f46-4ad0-afb9-42f3d1b4bcb9","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1221","names":[{"code":"th","name":"เครื่องจักรและอุปกรณ์โรงงาน"},{"code":"en","name":"Machinery & Factory Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '98eb4f60-5d5d-42c5-a684-20f5cf4fe88e', '1222', 1, '{"id":"98eb4f60-5d5d-42c5-a684-20f5cf4fe88e","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1222","names":[{"code":"th","name":"เครื่องตกแต่ง ติดตั้ง และเฟอร์นิเจอร์"},{"code":"en","name":"Furniture & Fixtures"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f07cc695-4d75-4290-a3ad-095d3c464b2f', '1223', 1, '{"id":"f07cc695-4d75-4290-a3ad-095d3c464b2f","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1223","names":[{"code":"th","name":"เครื่องใช้และอุปกรณ์สำนักงาน"},{"code":"en","name":"Office Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd3633b60-7a48-46af-acf5-23c9c2db61be', '1224', 1, '{"id":"d3633b60-7a48-46af-acf5-23c9c2db61be","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1224","names":[{"code":"th","name":"คอมพิวเตอร์และอุปกรณ์ประมวลผล"},{"code":"en","name":"Computer & Hardware Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd7287efc-aaf4-48dd-a49c-e8d58fb08b0c', '1231', 1, '{"id":"d7287efc-aaf4-48dd-a49c-e8d58fb08b0c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1231","names":[{"code":"th","name":"ยานพาหนะและรถบรรทุกขนส่ง"},{"code":"en","name":"Vehicles & Transport Trucks"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '826ebb58-91eb-4f94-a28d-d4e719b9f013', '1241', 1, '{"id":"826ebb58-91eb-4f94-a28d-d4e719b9f013","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1241","names":[{"code":"th","name":"งานระหว่างก่อสร้างและติดตั้งเครื่องจักร"},{"code":"en","name":"Construction & Installation in Progress"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '3cba8c27-e311-4a71-a395-a2e657fb99f7', '1281', 1, '{"id":"3cba8c27-e311-4a71-a395-a2e657fb99f7","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1281","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Acc. Dep. - Land Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '755083f7-2aae-4bb1-a9c1-b3ae23c5cc7d', '1282', 1, '{"id":"755083f7-2aae-4bb1-a9c1-b3ae23c5cc7d","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1282","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Acc. Dep. - Buildings"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '27ea0448-ece5-452a-a6f6-5a9959d8a6a4', '1283', 1, '{"id":"27ea0448-ece5-452a-a6f6-5a9959d8a6a4","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1283","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ส่วนปรับปรุงอาคาร"},{"code":"en","name":"Acc. Dep. - Building Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '37dc7a27-a719-4021-ae6b-054206722bda', '1284', 1, '{"id":"37dc7a27-a719-4021-ae6b-054206722bda","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1284","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องจักรและอุปกรณ์"},{"code":"en","name":"Acc. Dep. - Machinery & Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '24ae9673-4474-427c-a69b-19c355bfd2c6', '1285', 1, '{"id":"24ae9673-4474-427c-a69b-19c355bfd2c6","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1285","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องตกแต่งและติดตั้ง"},{"code":"en","name":"Acc. Dep. - Furniture & Fixtures"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '305bc2ba-285c-4da8-a2a4-bc8b9170f56c', '1286', 1, '{"id":"305bc2ba-285c-4da8-a2a4-bc8b9170f56c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1286","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องใช้สำนักงาน"},{"code":"en","name":"Acc. Dep. - Office Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c3aba50c-c2d8-4daa-a3b5-325d339b58ae', '1287', 1, '{"id":"c3aba50c-c2d8-4daa-a3b5-325d339b58ae","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1287","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - คอมพิวเตอร์และอุปกรณ์"},{"code":"en","name":"Acc. Dep. - Computer & Hardware"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '06d35111-4a94-46a7-a44c-11d15a739df0', '1288', 1, '{"id":"06d35111-4a94-46a7-a44c-11d15a739df0","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1288","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ยานพาหนะ"},{"code":"en","name":"Acc. Dep. - Vehicles"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7140617d-651c-4d80-a70d-0900c5124f83', '1291', 1, '{"id":"7140617d-651c-4d80-a70d-0900c5124f83","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1291","names":[{"code":"th","name":"โปรแกรมคอมพิวเตอร์และสิทธิการใช้งาน"},{"code":"en","name":"Software & Licenses"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '42733cdc-416d-4c53-a48f-24a32520e8dc', '1292', 1, '{"id":"42733cdc-416d-4c53-a48f-24a32520e8dc","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1292","names":[{"code":"th","name":"ค่าตัดจำหน่ายสะสม - โปรแกรมคอมพิวเตอร์"},{"code":"en","name":"Acc. Amortization - Software"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c0201e98-f20a-4bb6-a02c-d477863609b1', '1295', 1, '{"id":"c0201e98-f20a-4bb6-a02c-d477863609b1","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1295","names":[{"code":"th","name":"เงินประกันและเงินมัดจำระยะยาว"},{"code":"en","name":"Long-term Deposits & Guarantees"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a701daa9-134c-41f7-a666-4815453abbd0', '2000', 1, '{"id":"a701daa9-134c-41f7-a666-4815453abbd0","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2000","names":[{"code":"th","name":"หนี้สิน"},{"code":"en","name":"Liabilities"}],"accounttype":"liability","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '790a18f9-1324-40fb-a65e-13ed7b77f7ef', '2100', 1, '{"id":"790a18f9-1324-40fb-a65e-13ed7b77f7ef","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2100","names":[{"code":"th","name":"หนี้สินหมุนเวียน"},{"code":"en","name":"Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '4998d997-9b8f-447a-a4fd-bec4cbf23c6a', '2111', 1, '{"id":"4998d997-9b8f-447a-a4fd-bec4cbf23c6a","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2111","names":[{"code":"th","name":"เจ้าหนี้การค้า - ในประเทศ"},{"code":"en","name":"Trade Accounts Payable - Domestic"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a903f789-1066-4e29-a6c6-8ef1f39ad5ae', '2112', 1, '{"id":"a903f789-1066-4e29-a6c6-8ef1f39ad5ae","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2112","names":[{"code":"th","name":"เจ้าหนี้การค้า - ต่างประเทศ"},{"code":"en","name":"Trade Accounts Payable - Overseas"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '45bc9687-eb8e-4b7c-acda-68c5921e0030', '2113', 1, '{"id":"45bc9687-eb8e-4b7c-acda-68c5921e0030","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2113","names":[{"code":"th","name":"ตั๋วเงินจ่ายการค้า"},{"code":"en","name":"Notes Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '13d81fcb-9559-47d3-a1fb-15781ea02d24', '2114', 1, '{"id":"13d81fcb-9559-47d3-a1fb-15781ea02d24","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2114","names":[{"code":"th","name":"เช็คจ่ายลงวันที่ล่วงหน้า"},{"code":"en","name":"Post-dated Cheques Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '901936ec-b3c6-43f4-a836-aa9f3f34c19a', '2121', 1, '{"id":"901936ec-b3c6-43f4-a836-aa9f3f34c19a","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2121","names":[{"code":"th","name":"เจ้าหนี้อื่นและเงินทดรองรับ"},{"code":"en","name":"Other Payables & Advances"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '06a35bc8-118f-451b-aefe-c96054c7a071', '2122', 1, '{"id":"06a35bc8-118f-451b-aefe-c96054c7a071","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2122","names":[{"code":"th","name":"เจ้าหนี้กรมสรรพากร"},{"code":"en","name":"Revenue Department Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd327897b-8c0e-405b-a281-dcafc8f62135', '2131', 1, '{"id":"d327897b-8c0e-405b-a281-dcafc8f62135","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2131","names":[{"code":"th","name":"ภาษีขาย"},{"code":"en","name":"Output VAT"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'ddbde76d-e1ee-48ff-a32f-7999346ac593', '2132', 1, '{"id":"ddbde76d-e1ee-48ff-a32f-7999346ac593","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2132","names":[{"code":"th","name":"ภาษีขายยังไม่ถึงกำหนดชำระ"},{"code":"en","name":"Undue Output VAT"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '2e875057-ff06-4c47-a5b9-aa56aba17738', '2141', 1, '{"id":"2e875057-ff06-4c47-a5b9-aa56aba17738","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2141","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.1 (เงินเดือน)"},{"code":"en","name":"WHT Payable - P.N.D.1 (Salaries)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5d2cadb9-befe-4a3b-a358-509a21e11538', '2142', 1, '{"id":"5d2cadb9-befe-4a3b-a358-509a21e11538","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2142","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.3 (บุคคลธรรมดา)"},{"code":"en","name":"WHT Payable - P.N.D.3 (Individuals)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5f7df044-59b8-4cb5-a0ff-4e6fa2d9c06b', '2143', 1, '{"id":"5f7df044-59b8-4cb5-a0ff-4e6fa2d9c06b","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2143","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.53 (นิติบุคคล)"},{"code":"en","name":"WHT Payable - P.N.D.53 (Corporations)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f1384a2d-98dc-4382-ae78-74d0b0be266b', '2144', 1, '{"id":"f1384a2d-98dc-4382-ae78-74d0b0be266b","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2144","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.2 (ดอกเบี้ย/ปันผล)"},{"code":"en","name":"WHT Payable - P.N.D.2"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd48eed52-bd15-4afe-a448-5162a4a270a6', '2145', 1, '{"id":"d48eed52-bd15-4afe-a448-5162a4a270a6","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2145","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.54 (ส่งไปต่างประเทศ)"},{"code":"en","name":"WHT Payable - P.N.D.54 (Overseas)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1a464896-c5c9-4021-afcd-c09fdb679ff3', '2151', 1, '{"id":"1a464896-c5c9-4021-afcd-c09fdb679ff3","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2151","names":[{"code":"th","name":"เงินสมทบกองทุนประกันสังคมค้างจ่าย (ลูกจ้าง+นายจ้าง)"},{"code":"en","name":"Social Security Fund Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '43a35a71-5367-43fd-a67c-f93725f46751', '2152', 1, '{"id":"43a35a71-5367-43fd-a67c-f93725f46751","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2152","names":[{"code":"th","name":"เงินเดือนและค่าจ้างค้างจ่าย"},{"code":"en","name":"Accrued Salaries & Wages"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '2046907c-b1fa-461c-a6cb-6f33c8973214', '2153', 1, '{"id":"2046907c-b1fa-461c-a6cb-6f33c8973214","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2153","names":[{"code":"th","name":"ค่าเช่าค้างจ่าย"},{"code":"en","name":"Accrued Rent Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '94dd75e7-db38-4abc-a9dd-eb4e2ab230a2', '2154', 1, '{"id":"94dd75e7-db38-4abc-a9dd-eb4e2ab230a2","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2154","names":[{"code":"th","name":"ค่าน้ำประปาและค่าไฟฟ้าค้างจ่าย"},{"code":"en","name":"Accrued Utilities Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '4c3be7b5-2d89-4748-aca9-5dd1cbb7d39d', '2155', 1, '{"id":"4c3be7b5-2d89-4748-aca9-5dd1cbb7d39d","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2155","names":[{"code":"th","name":"ค่าโทรศัพท์และอินเทอร์เน็ตค้างจ่าย"},{"code":"en","name":"Accrued Telephone & Internet"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'dbeb31dd-5315-416b-a5b5-4ba366b5ebc3', '2156', 1, '{"id":"dbeb31dd-5315-416b-a5b5-4ba366b5ebc3","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2156","names":[{"code":"th","name":"ค่าสอบบัญชีและทำบัญชีค้างจ่าย"},{"code":"en","name":"Accrued Audit & Accounting Fees"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a764fbe6-95ee-4c08-a6a6-58f27de7c623', '2157', 1, '{"id":"a764fbe6-95ee-4c08-a6a6-58f27de7c623","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2157","names":[{"code":"th","name":"ดอกเบี้ยค้างจ่าย"},{"code":"en","name":"Accrued Interest Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'fdc9b254-1cdd-4ab4-aab2-6ecfe7267eea', '2158', 1, '{"id":"fdc9b254-1cdd-4ab4-aab2-6ecfe7267eea","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2158","names":[{"code":"th","name":"โบนัสพนักงานค้างจ่าย"},{"code":"en","name":"Accrued Staff Bonuses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1123a3eb-bc18-4450-a5ca-31b3f3864825', '2159', 1, '{"id":"1123a3eb-bc18-4450-a5ca-31b3f3864825","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2159","names":[{"code":"th","name":"ค่าใช้จ่ายค้างจ่ายอื่น"},{"code":"en","name":"Other Accrued Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '496bf099-c241-4cf9-a67d-88b5bcc16bfa', '2161', 1, '{"id":"496bf099-c241-4cf9-a67d-88b5bcc16bfa","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2161","names":[{"code":"th","name":"เงินรับล่วงหน้าค่าสินค้าและบริการจากลูกค้า"},{"code":"en","name":"Advances Received from Customers"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '38199149-a709-4573-a3d8-4d24d4d209d1', '2171', 1, '{"id":"38199149-a709-4573-a3d8-4d24d4d209d1","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2171","names":[{"code":"th","name":"เงินเบิกเกินบัญชีธนาคาร (O/D)"},{"code":"en","name":"Bank Overdrafts (O/D)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9935f8c8-54be-428a-a1f5-1437695fc722', '2172', 1, '{"id":"9935f8c8-54be-428a-a1f5-1437695fc722","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2172","names":[{"code":"th","name":"เงินกู้ยืมระยะสั้นจากสถาบันการเงิน"},{"code":"en","name":"Short-term Borrowings from Banks"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '00c7d6a3-b852-4b66-a3bf-cedb9f1052c4', '2173', 1, '{"id":"00c7d6a3-b852-4b66-a3bf-cedb9f1052c4","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2173","names":[{"code":"th","name":"เงินกู้ยืมระยะสั้นจากกรรมการหรือบุคคลที่เกี่ยวข้องกัน"},{"code":"en","name":"Short-term Loans from Directors"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a3d6fb74-c460-412f-aa35-f565937e7f12', '2181', 1, '{"id":"a3d6fb74-c460-412f-aa35-f565937e7f12","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2181","names":[{"code":"th","name":"ส่วนของหนี้สินระยะยาวที่ถึงกำหนดชำระภายในหนึ่งปี"},{"code":"en","name":"Current Portion of Long-term Debt"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '56c87691-81e5-4e9f-ac83-b8cf9c9ce042', '2200', 1, '{"id":"56c87691-81e5-4e9f-ac83-b8cf9c9ce042","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2200","names":[{"code":"th","name":"หนี้สินไม่หมุนเวียน"},{"code":"en","name":"Non-Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a1bc5dc6-5705-48d5-a1ed-6795929e535c', '2211', 1, '{"id":"a1bc5dc6-5705-48d5-a1ed-6795929e535c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2211","names":[{"code":"th","name":"เงินกู้ยืมระยะยาวจากสถาบันการเงิน"},{"code":"en","name":"Long-term Bank Loans"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c0afedc8-bc69-40aa-afd1-1acba6ae5575', '2212', 1, '{"id":"c0afedc8-bc69-40aa-afd1-1acba6ae5575","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2212","names":[{"code":"th","name":"หนี้สินตามสัญญาเช่าทางการเงินระยะยาว"},{"code":"en","name":"Long-term Financial Lease Liabilities"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7a482b82-67f6-461a-a62a-9857016503d3', '2221', 1, '{"id":"7a482b82-67f6-461a-a62a-9857016503d3","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2221","names":[{"code":"th","name":"เงินกู้ยืมระยะยาวจากกรรมการหรือผู้ถือหุ้น"},{"code":"en","name":"Long-term Loans from Directors/Shareholders"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a3929cb0-31cf-47e7-a80d-bd80ab6b81bc', '2231', 1, '{"id":"a3929cb0-31cf-47e7-a80d-bd80ab6b81bc","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2231","names":[{"code":"th","name":"ประมาณการหนี้สินผลประโยชน์พนักงาน"},{"code":"en","name":"Provision for Employee Benefits"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c36ed6f7-7fdb-42ba-adfc-ee54213554b6', '2291', 1, '{"id":"c36ed6f7-7fdb-42ba-adfc-ee54213554b6","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2291","names":[{"code":"th","name":"หนี้สินไม่หมุนเวียนอื่น"},{"code":"en","name":"Other Non-Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b47d6c42-d7ea-4ad8-ae24-d46f7ca5a798', '3000', 1, '{"id":"b47d6c42-d7ea-4ad8-ae24-d46f7ca5a798","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3000","names":[{"code":"th","name":"ส่วนของเจ้าของ"},{"code":"en","name":"Equity"}],"accounttype":"equity","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '61f72d8d-2623-4b8f-a7fb-f2d9efa2c8a5', '3100', 1, '{"id":"61f72d8d-2623-4b8f-a7fb-f2d9efa2c8a5","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3100","names":[{"code":"th","name":"ทุนจดทะเบียนและส่วนเกินทุน"},{"code":"en","name":"Share Capital & Premium"}],"accounttype":"equity","parentaccountcode":"3000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'bb18a5c1-255a-4081-a187-327c479b31a1', '3111', 1, '{"id":"bb18a5c1-255a-4081-a187-327c479b31a1","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3111","names":[{"code":"th","name":"ทุนเรือนหุ้น - หุ้นสามัญ"},{"code":"en","name":"Authorized Share Capital - Common Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e58b5737-e8d2-4bda-ab6a-f676d78bbd6b', '3112', 1, '{"id":"e58b5737-e8d2-4bda-ab6a-f676d78bbd6b","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3112","names":[{"code":"th","name":"ทุนเรือนหุ้น - หุ้นบุริมสิทธิ"},{"code":"en","name":"Authorized Share Capital - Preferred Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7b7ec529-2177-4dc5-a132-6b531a0db88a', '3121', 1, '{"id":"7b7ec529-2177-4dc5-a132-6b531a0db88a","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3121","names":[{"code":"th","name":"ส่วนเกินมูลค่าหุ้นสามัญ"},{"code":"en","name":"Share Premium - Common Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '00b3dd53-12cf-406f-acec-6d7e8131dd60', '3131', 1, '{"id":"00b3dd53-12cf-406f-acec-6d7e8131dd60","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3131","names":[{"code":"th","name":"ทุนส่วนของเจ้าของ (ห้างหุ้นส่วน/บุคคลธรรมดา)"},{"code":"en","name":"Owner''s / Partner''s Capital"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5a8dfe23-5e07-4e7b-ac50-18456cdf1788', '3132', 1, '{"id":"5a8dfe23-5e07-4e7b-ac50-18456cdf1788","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3132","names":[{"code":"th","name":"เงินถอนใช้ส่วนตัวของเจ้าของกิจการ"},{"code":"en","name":"Owner''s Drawings"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e1971b5e-d800-4944-a440-651c84d19b5a', '3200', 1, '{"id":"e1971b5e-d800-4944-a440-651c84d19b5a","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3200","names":[{"code":"th","name":"กำไร (ขาดทุน) สะสม"},{"code":"en","name":"Retained Earnings"}],"accounttype":"equity","parentaccountcode":"3000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '56b5346a-7595-456a-ab59-45c78cf426ea', '3211', 1, '{"id":"56b5346a-7595-456a-ab59-45c78cf426ea","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3211","names":[{"code":"th","name":"สำรองตามกฎหมาย"},{"code":"en","name":"Legal Reserve"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b5a4a579-a07e-4075-a58f-02307acc0400', '3212', 1, '{"id":"b5a4a579-a07e-4075-a58f-02307acc0400","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3212","names":[{"code":"th","name":"สำรองอื่นเพื่อวัตถุประสงค์เฉพาะ"},{"code":"en","name":"Other Appropriated Reserves"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b5022df0-2ebc-433c-a4b0-5b5fb565f557', '3221', 1, '{"id":"b5022df0-2ebc-433c-a4b0-5b5fb565f557","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3221","names":[{"code":"th","name":"กำไร (ขาดทุน) สะสมยังไม่ได้จัดสรร"},{"code":"en","name":"Unappropriated Retained Earnings"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '022e4656-84bc-4940-a27f-1d9cc591a423', '3222', 1, '{"id":"022e4656-84bc-4940-a27f-1d9cc591a423","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3222","names":[{"code":"th","name":"กำไร (ขาดทุน) สุทธิประจำปี"},{"code":"en","name":"Net Profit/Loss for Current Year"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '91f83ccb-0cc0-4095-a499-ddf329c9c80f', '3223', 1, '{"id":"91f83ccb-0cc0-4095-a499-ddf329c9c80f","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3223","names":[{"code":"th","name":"เงินปันผลจ่าย"},{"code":"en","name":"Dividends Declared & Paid"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '35c77d64-66ad-48ef-a40c-b428e0bc7547', '4000', 1, '{"id":"35c77d64-66ad-48ef-a40c-b428e0bc7547","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4000","names":[{"code":"th","name":"รายได้"},{"code":"en","name":"Income"}],"accounttype":"income","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '94628469-b789-4921-a71f-3ac0481e1202', '4100', 1, '{"id":"94628469-b789-4921-a71f-3ac0481e1202","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4100","names":[{"code":"th","name":"รายได้จากการขายและการให้บริการ"},{"code":"en","name":"Revenue from Sales & Services"}],"accounttype":"income","parentaccountcode":"4000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f7598db6-7d62-4e3b-ad56-fc325e0ceeaa', '4111', 1, '{"id":"f7598db6-7d62-4e3b-ad56-fc325e0ceeaa","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4111","names":[{"code":"th","name":"รายได้จากการขายสินค้า - มีภาษีมูลค่าเพิ่ม 7%"},{"code":"en","name":"Sales Revenue - VAT 7%"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f44ba4ef-b70a-4361-a201-93d7ed5c3aa7', '4112', 1, '{"id":"f44ba4ef-b70a-4361-a201-93d7ed5c3aa7","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4112","names":[{"code":"th","name":"รายได้จากการขายสินค้า - อัตราภาษี 0% / ส่งออก"},{"code":"en","name":"Sales Revenue - Zero Rated / Export"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '00ac4bf8-84b9-42c3-ace1-dc6328579e9f', '4113', 1, '{"id":"00ac4bf8-84b9-42c3-ace1-dc6328579e9f","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4113","names":[{"code":"th","name":"รายได้จากการขายสินค้า - ได้รับยกเว้นภาษีมูลค่าเพิ่ม"},{"code":"en","name":"Sales Revenue - VAT Exempted"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '19172099-b2f5-4865-aced-d767acea4ef6', '4121', 1, '{"id":"19172099-b2f5-4865-aced-d767acea4ef6","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4121","names":[{"code":"th","name":"รายได้จากการให้บริการ - มีภาษีมูลค่าเพิ่ม 7%"},{"code":"en","name":"Service Income - VAT 7%"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'cbed5dc8-06d7-4735-af15-e9072999f82e', '4122', 1, '{"id":"cbed5dc8-06d7-4735-af15-e9072999f82e","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4122","names":[{"code":"th","name":"รายได้จากการให้บริการ - ได้รับยกเว้นภาษีมูลค่าเพิ่ม"},{"code":"en","name":"Service Income - VAT Exempted"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '6b858be6-b08c-47d4-a50c-3ce5eec4236c', '4131', 1, '{"id":"6b858be6-b08c-47d4-a50c-3ce5eec4236c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4131","names":[{"code":"th","name":"รับคืนสินค้าและลดหนี้ขาย"},{"code":"en","name":"Sales Returns and Allowances"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b2912b16-55ff-4229-a1b2-b362326c6e5e', '4132', 1, '{"id":"b2912b16-55ff-4229-a1b2-b362326c6e5e","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4132","names":[{"code":"th","name":"ส่วนลดจ่าย"},{"code":"en","name":"Sales Cash Discounts"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '41b37ec1-298b-4da0-a8d8-d6326603db4f', '4141', 1, '{"id":"41b37ec1-298b-4da0-a8d8-d6326603db4f","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4141","names":[{"code":"th","name":"รายได้ค่าบริการขนส่งและจัดส่งสินค้า"},{"code":"en","name":"Freight & Delivery Income"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '7408d26d-8499-44a3-abf7-dca0e29507df', '4200', 1, '{"id":"7408d26d-8499-44a3-abf7-dca0e29507df","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4200","names":[{"code":"th","name":"รายได้อื่น"},{"code":"en","name":"Other Income"}],"accounttype":"income","parentaccountcode":"4000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'ff896a56-63f5-4159-ab05-937cf419949c', '4211', 1, '{"id":"ff896a56-63f5-4159-ab05-937cf419949c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4211","names":[{"code":"th","name":"ดอกเบี้ยรับจากสถาบันการเงิน"},{"code":"en","name":"Interest Income from Banks"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '2b792994-a037-450a-a9bf-b503f3ee3ccd', '4212', 1, '{"id":"2b792994-a037-450a-a9bf-b503f3ee3ccd","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4212","names":[{"code":"th","name":"เงินปันผลรับจากเงินลงทุน"},{"code":"en","name":"Dividend Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b4229b7d-823f-4a42-ac5e-8ba4daa13679', '4221', 1, '{"id":"b4229b7d-823f-4a42-ac5e-8ba4daa13679","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4221","names":[{"code":"th","name":"กำไรจากการจำหน่ายทรัพย์สินถาวร"},{"code":"en","name":"Gain on Disposal of Fixed Assets"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1d7c25a4-ce69-435e-aabb-ae2bc97e1a78', '4231', 1, '{"id":"1d7c25a4-ce69-435e-aabb-ae2bc97e1a78","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4231","names":[{"code":"th","name":"กำไรจากอัตราแลกเปลี่ยนเงินตราต่างประเทศ"},{"code":"en","name":"Gain on Foreign Exchange"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e1aab20f-0a5d-4d13-a83e-4045d0f671be', '4241', 1, '{"id":"e1aab20f-0a5d-4d13-a83e-4045d0f671be","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4241","names":[{"code":"th","name":"รายได้ค่าเช่าอาคารและอุปกรณ์"},{"code":"en","name":"Rental Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b9f53ac4-9ab0-4d22-af48-aacaca420681', '4251', 1, '{"id":"b9f53ac4-9ab0-4d22-af48-aacaca420681","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4251","names":[{"code":"th","name":"หนี้สูญได้รับคืน"},{"code":"en","name":"Bad Debts Recovered"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b8f46195-eb96-44e6-ae8f-bcda0e876966', '4291', 1, '{"id":"b8f46195-eb96-44e6-ae8f-bcda0e876966","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4291","names":[{"code":"th","name":"รายได้เบ็ดเตล็ดอื่น"},{"code":"en","name":"Miscellaneous Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c6d8630f-990d-4e0d-a0ce-c09b9bb4ea85', '5000', 1, '{"id":"c6d8630f-990d-4e0d-a0ce-c09b9bb4ea85","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5000","names":[{"code":"th","name":"ค่าใช้จ่าย"},{"code":"en","name":"Expenses"}],"accounttype":"expense","parentaccountcode":null,"normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '2f6e4add-3482-49d8-ac11-504e4192bb17', '5100', 1, '{"id":"2f6e4add-3482-49d8-ac11-504e4192bb17","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5100","names":[{"code":"th","name":"ต้นทุนขายและบริการ"},{"code":"en","name":"Cost of Goods Sold & Services"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '09d1982b-2e44-40e3-af0a-3858d8e1c404', '5111', 1, '{"id":"09d1982b-2e44-40e3-af0a-3858d8e1c404","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5111","names":[{"code":"th","name":"ซื้อสินค้าสำเร็จรูป"},{"code":"en","name":"Purchases of Merchandise"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'ff187f38-ca98-44c9-a564-aac7ff730c5a', '5112', 1, '{"id":"ff187f38-ca98-44c9-a564-aac7ff730c5a","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5112","names":[{"code":"th","name":"ค่าขนส่งเข้า"},{"code":"en","name":"Freight-In"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '51198fb7-0384-4017-a459-92853d4bdc9f', '5113', 1, '{"id":"51198fb7-0384-4017-a459-92853d4bdc9f","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5113","names":[{"code":"th","name":"ส่งคืนสินค้าและส่วนลดที่ได้รับ"},{"code":"en","name":"Purchase Returns and Allowances"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1312fb13-4d13-41f8-aa86-7d1b92a26281', '5114', 1, '{"id":"1312fb13-4d13-41f8-aa86-7d1b92a26281","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5114","names":[{"code":"th","name":"ส่วนลดรับ"},{"code":"en","name":"Purchase Cash Discounts"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'fcf3efce-49b8-45e5-a734-0812101fa7c2', '5121', 1, '{"id":"fcf3efce-49b8-45e5-a734-0812101fa7c2","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5121","names":[{"code":"th","name":"ต้นทุนสินค้าสำเร็จรูปที่ขาย"},{"code":"en","name":"Cost of Goods Sold"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f3dbdd43-fe11-4fe9-a953-07fcdb007531', '5131', 1, '{"id":"f3dbdd43-fe11-4fe9-a953-07fcdb007531","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5131","names":[{"code":"th","name":"ต้นทุนค่าแรงและบริการ"},{"code":"en","name":"Direct Labor & Service Costs"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f29c8c21-36eb-4dc7-a368-99e512bf1d6d', '5132', 1, '{"id":"f29c8c21-36eb-4dc7-a368-99e512bf1d6d","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5132","names":[{"code":"th","name":"ค่าจ้างเหมาบริการช่วงภายนอก (Subcontractor)"},{"code":"en","name":"Subcontractor Costs"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '30aa61f9-c294-48de-adfd-1ece9e6f264c', '5141', 1, '{"id":"30aa61f9-c294-48de-adfd-1ece9e6f264c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5141","names":[{"code":"th","name":"สินค้าสูญหาย เสียหาย และสินค้าชำรุด"},{"code":"en","name":"Inventory Shrinkage & Damage"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '72bf05f6-6644-473e-aa18-48933a6ac587', '5200', 1, '{"id":"72bf05f6-6644-473e-aa18-48933a6ac587","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5200","names":[{"code":"th","name":"ค่าใช้จ่ายในการขาย"},{"code":"en","name":"Selling Expenses"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '2a485721-3fa9-4451-ac93-58cfa5bf7680', '5211', 1, '{"id":"2a485721-3fa9-4451-ac93-58cfa5bf7680","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5211","names":[{"code":"th","name":"เงินเดือน ค่าล่วงเวลา และเบี้ยเลี้ยงฝ่ายขาย"},{"code":"en","name":"Sales Salaries, Overtime & Allowances"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b4fbbfc4-e06e-4c41-a920-4affc7dec7fe', '5212', 1, '{"id":"b4fbbfc4-e06e-4c41-a920-4affc7dec7fe","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5212","names":[{"code":"th","name":"ค่านายหน้าและค่าคอมมิชชั่นฝ่ายขาย"},{"code":"en","name":"Sales Commissions & Incentives"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '6caf479e-8120-43a5-af5f-62fa7cdddf24', '5221', 1, '{"id":"6caf479e-8120-43a5-af5f-62fa7cdddf24","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5221","names":[{"code":"th","name":"ค่าโฆษณา ประชาสัมพันธ์ และการตลาดออนไลน์"},{"code":"en","name":"Advertising, PR & Online Marketing"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '818b59cc-7fc4-4535-ab78-9b8f3c059c7b', '5222', 1, '{"id":"818b59cc-7fc4-4535-ab78-9b8f3c059c7b","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5222","names":[{"code":"th","name":"ค่าส่งเสริมการขาย ของแถม และตัวอย่างสินค้า"},{"code":"en","name":"Sales Promotion, Gifts & Samples"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '317f5e2b-2394-4e1f-a7b5-46baf0376b61', '5231', 1, '{"id":"317f5e2b-2394-4e1f-a7b5-46baf0376b61","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5231","names":[{"code":"th","name":"ค่าขนส่งสินค้าออกและบริการจัดส่งให้ลูกค้า"},{"code":"en","name":"Outward Freight & Shipping to Customers"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '38924244-8d43-474e-a60a-04e1363d96af', '5232', 1, '{"id":"38924244-8d43-474e-a60a-04e1363d96af","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5232","names":[{"code":"th","name":"ค่าวัสดุหีบห่อและบรรจุภัณฑ์สำหรับการขาย"},{"code":"en","name":"Packaging & Packing Materials"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '4c1a3bb9-b5cd-4cc3-a0eb-2f7b15824470', '5241', 1, '{"id":"4c1a3bb9-b5cd-4cc3-a0eb-2f7b15824470","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5241","names":[{"code":"th","name":"ค่าจัดงานแสดงสินค้าและการออกบูธ"},{"code":"en","name":"Exhibition & Trade Fair Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f24195f8-0110-4655-abba-343b6994fa99', '5251', 1, '{"id":"f24195f8-0110-4655-abba-343b6994fa99","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5251","names":[{"code":"th","name":"ค่าน้ำมันและค่าเดินทางพบลูกค้าฝ่ายขาย"},{"code":"en","name":"Sales Fuel & Travel Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1f577756-05ba-4c76-a6e7-ba61d3a1c22c', '5291', 1, '{"id":"1f577756-05ba-4c76-a6e7-ba61d3a1c22c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5291","names":[{"code":"th","name":"ค่าใช้จ่ายในการขายอื่น"},{"code":"en","name":"Other Selling Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '19b87eb0-ef3b-41d6-aefa-337adc7b0ae2', '5300', 1, '{"id":"19b87eb0-ef3b-41d6-aefa-337adc7b0ae2","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5300","names":[{"code":"th","name":"ค่าใช้จ่ายในการบริหาร"},{"code":"en","name":"Administrative Expenses"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b6510fac-de8f-423d-a72d-a76a3046720c', '5311', 1, '{"id":"b6510fac-de8f-423d-a72d-a76a3046720c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5311","names":[{"code":"th","name":"เงินเดือน ค่าจ้าง และค่าล่วงเวลาพนักงานสำนักงาน"},{"code":"en","name":"Office Salaries, Wages & Overtime"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '20a4afe1-a3d5-40d8-af31-ed29083d072e', '5312', 1, '{"id":"20a4afe1-a3d5-40d8-af31-ed29083d072e","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5312","names":[{"code":"th","name":"ค่าตอบแทนและเบี้ยประชุมกรรมการ"},{"code":"en","name":"Directors'' Remuneration & Meeting Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f5ae843a-6a1d-4a43-a290-92df579fffbb', '5313', 1, '{"id":"f5ae843a-6a1d-4a43-a290-92df579fffbb","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5313","names":[{"code":"th","name":"โบนัสพนักงานประจำปี"},{"code":"en","name":"Annual Staff Bonuses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '2eb3d413-f331-41d2-ac2f-7ad0341369e1', '5314', 1, '{"id":"2eb3d413-f331-41d2-ac2f-7ad0341369e1","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5314","names":[{"code":"th","name":"เงินสมทบกองทุนประกันสังคม (ส่วนของนายจ้าง)"},{"code":"en","name":"Social Security Fund - Employer''s Contribution"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '08994731-a5bb-4ab4-a72a-2508925bd78b', '5315', 1, '{"id":"08994731-a5bb-4ab4-a72a-2508925bd78b","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5315","names":[{"code":"th","name":"เงินสมทบกองทุนเงินทดแทนและกองทุนสำรองเลี้ยงชีพ"},{"code":"en","name":"Workmen Compensation & Provident Fund"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f6571e83-11ad-4a4e-a290-7311ab9924e5', '5316', 1, '{"id":"f6571e83-11ad-4a4e-a290-7311ab9924e5","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5316","names":[{"code":"th","name":"ค่าสวัสดิการพนักงานและชุดยูนิฟอร์ม"},{"code":"en","name":"Staff Welfare & Uniforms"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '17221c0d-e405-4417-a899-bf930dbc3eb6', '5317', 1, '{"id":"17221c0d-e405-4417-a899-bf930dbc3eb6","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5317","names":[{"code":"th","name":"ค่าฝึกอบรมและสัมมนาบุคลากร"},{"code":"en","name":"Training & Seminar Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '3667af23-1d6f-4a07-a099-caf00ba34496', '5321', 1, '{"id":"3667af23-1d6f-4a07-a099-caf00ba34496","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5321","names":[{"code":"th","name":"ค่าเช่าอาคารสำนักงานและพื้นที่ประกอบการ"},{"code":"en","name":"Office Building Rent"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '2c880e7b-b2c7-4e50-a9fd-5596a3df676a', '5322', 1, '{"id":"2c880e7b-b2c7-4e50-a9fd-5596a3df676a","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5322","names":[{"code":"th","name":"ค่าน้ำประปา"},{"code":"en","name":"Water Utility Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '82509b59-8c88-4875-ad29-c35559328e1e', '5323', 1, '{"id":"82509b59-8c88-4875-ad29-c35559328e1e","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5323","names":[{"code":"th","name":"ค่าไฟฟ้า"},{"code":"en","name":"Electricity Utility Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'e5e2182f-e659-44ec-a081-7f003f55a0cf', '5324', 1, '{"id":"e5e2182f-e659-44ec-a081-7f003f55a0cf","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5324","names":[{"code":"th","name":"ค่าโทรศัพท์และค่าบริการโทรคมนาคม"},{"code":"en","name":"Telephone & Telecom Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '4f0cf349-02c4-43fd-a835-f0a476f1b16b', '5325', 1, '{"id":"4f0cf349-02c4-43fd-a835-f0a476f1b16b","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5325","names":[{"code":"th","name":"ค่าบริการอินเทอร์เน็ต ระบบเซิร์ฟเวอร์ และคลาวด์"},{"code":"en","name":"Internet, Server & Cloud Hosting Services"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '51b322af-f268-49f8-a85f-0d92e0a27d31', '5331', 1, '{"id":"51b322af-f268-49f8-a85f-0d92e0a27d31","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5331","names":[{"code":"th","name":"ค่าเครื่องเขียน แบบพิมพ์ และวัสดุสำนักงาน"},{"code":"en","name":"Stationery, Printing & Office Supplies"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f01b8aa1-7584-4614-a179-e824d492e750', '5332', 1, '{"id":"f01b8aa1-7584-4614-a179-e824d492e750","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5332","names":[{"code":"th","name":"ค่าไปรษณีย์และค่าส่งเอกสารพัสดุ"},{"code":"en","name":"Postal & Courier Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'dde619d8-b0bb-4b39-afa3-270c8743baeb', '5341', 1, '{"id":"dde619d8-b0bb-4b39-afa3-270c8743baeb","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5341","names":[{"code":"th","name":"ค่าตรวจสอบบัญชี (Audit Fees)"},{"code":"en","name":"Audit Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '4b21187b-9e3f-4266-ac84-0ac48e1a6760', '5342', 1, '{"id":"4b21187b-9e3f-4266-ac84-0ac48e1a6760","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5342","names":[{"code":"th","name":"ค่าจัดทำบัญชีและที่ปรึกษาภาษีอากร"},{"code":"en","name":"Accounting & Tax Consultation Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'd7f8faec-d358-45e8-a11d-fe08e0ae6c01', '5343', 1, '{"id":"d7f8faec-d358-45e8-a11d-fe08e0ae6c01","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5343","names":[{"code":"th","name":"ค่าธรรมเนียมวิชาชีพกฎหมายและที่ปรึกษาธุรกิจ"},{"code":"en","name":"Legal & Business Advisory Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9fbe6f47-c7af-47fc-a227-d9e60813ac2e', '5351', 1, '{"id":"9fbe6f47-c7af-47fc-a227-d9e60813ac2e","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5351","names":[{"code":"th","name":"ค่าเบี้ยประกันภัยทรัพย์สิน อาคาร และยานพาหนะ"},{"code":"en","name":"Property, Building & Vehicle Insurance"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '883a16d5-3658-4969-ab1f-0b5ecad2037c', '5352', 1, '{"id":"883a16d5-3658-4969-ab1f-0b5ecad2037c","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5352","names":[{"code":"th","name":"ค่าซ่อมแซมและบำรุงรักษาอาคารและอุปกรณ์"},{"code":"en","name":"Repairs & Maintenance Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '40ad79b1-699a-4c70-a907-12b91351f0a6', '5353', 1, '{"id":"40ad79b1-699a-4c70-a907-12b91351f0a6","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5353","names":[{"code":"th","name":"ค่าบริการทำความสะอาดและรักษาความปลอดภัย"},{"code":"en","name":"Cleaning & Security Services"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '0e6df9cd-6ab5-4143-a9c1-5c551f15bb22', '5361', 1, '{"id":"0e6df9cd-6ab5-4143-a9c1-5c551f15bb22","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5361","names":[{"code":"th","name":"ค่าเสื่อมราคา - ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Depreciation - Land Improvements"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'a544ecca-34c1-4b3c-a17c-2ea93780171f', '5362', 1, '{"id":"a544ecca-34c1-4b3c-a17c-2ea93780171f","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5362","names":[{"code":"th","name":"ค่าเสื่อมราคา - อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Depreciation - Buildings"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'b115cf6f-ef6b-4834-ac32-aaefb0eb68f1', '5363', 1, '{"id":"b115cf6f-ef6b-4834-ac32-aaefb0eb68f1","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5363","names":[{"code":"th","name":"ค่าเสื่อมราคา - ส่วนปรับปรุงอาคาร"},{"code":"en","name":"Depreciation - Building Improvements"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '41e31853-9f26-464d-ae0a-d0dbbb136577', '5364', 1, '{"id":"41e31853-9f26-464d-ae0a-d0dbbb136577","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5364","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องจักรและอุปกรณ์"},{"code":"en","name":"Depreciation - Machinery & Equipment"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '348bad5e-ddc9-4dbf-a138-76571d313e68', '5365', 1, '{"id":"348bad5e-ddc9-4dbf-a138-76571d313e68","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5365","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องตกแต่งและติดตั้ง"},{"code":"en","name":"Depreciation - Furniture & Fixtures"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '09d3408c-ba82-4ef6-a8da-1f2da590a80d', '5366', 1, '{"id":"09d3408c-ba82-4ef6-a8da-1f2da590a80d","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5366","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องใช้สำนักงาน"},{"code":"en","name":"Depreciation - Office Equipment"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'fbd2ce3b-67f1-4bb7-a41b-a5da25de0350', '5367', 1, '{"id":"fbd2ce3b-67f1-4bb7-a41b-a5da25de0350","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5367","names":[{"code":"th","name":"ค่าเสื่อมราคา - คอมพิวเตอร์และอุปกรณ์"},{"code":"en","name":"Depreciation - Computer & Hardware"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '1cff4e4f-16ca-426c-a5ce-62f234062266', '5368', 1, '{"id":"1cff4e4f-16ca-426c-a5ce-62f234062266","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5368","names":[{"code":"th","name":"ค่าเสื่อมราคา - ยานพาหนะ"},{"code":"en","name":"Depreciation - Vehicles"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '439be659-6fb3-4ecd-a052-d8583f8b1e55', '5371', 1, '{"id":"439be659-6fb3-4ecd-a052-d8583f8b1e55","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5371","names":[{"code":"th","name":"ค่าตัดจำหน่ายโปรแกรมคอมพิวเตอร์และซอฟต์แวร์"},{"code":"en","name":"Amortization - Software & Applications"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '390d1180-ddf1-44f4-ac96-c3d173a1100a', '5381', 1, '{"id":"390d1180-ddf1-44f4-ac96-c3d173a1100a","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5381","names":[{"code":"th","name":"ค่าธรรมเนียมราชการและใบอนุญาตประกอบกิจการ"},{"code":"en","name":"Government Licenses & Registration Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '156b263a-9f65-48f8-a9cd-06c8bdff9818', '5382', 1, '{"id":"156b263a-9f65-48f8-a9cd-06c8bdff9818","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5382","names":[{"code":"th","name":"ภาษีที่ดินและสิ่งปลูกสร้าง"},{"code":"en","name":"Land and Building Taxes"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '17557de4-ba8e-4ff0-a2f2-c23b9b652918', '5383', 1, '{"id":"17557de4-ba8e-4ff0-a2f2-c23b9b652918","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5383","names":[{"code":"th","name":"ภาษีป้าย"},{"code":"en","name":"Signboard Tax"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'fcf87ee9-360f-461a-a494-dea98d4cc826', '5384', 1, '{"id":"fcf87ee9-360f-461a-a494-dea98d4cc826","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5384","names":[{"code":"th","name":"ภาษีซื้อต้องห้าม / ภาษีซื้อที่ไม่สามารถขอคืนได้"},{"code":"en","name":"Non-refundable Input Tax (Disallowed Input VAT)"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'f342edde-c412-490e-a376-8ff162480d6d', '5385', 1, '{"id":"f342edde-c412-490e-a376-8ff162480d6d","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5385","names":[{"code":"th","name":"เบี้ยปรับ เงินเพิ่ม และค่าปรับทางภาษีอากร"},{"code":"en","name":"Tax Penalties, Fines & Surcharges"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '03ef2ae6-c731-44e3-a5cb-3a5b65708092', '5391', 1, '{"id":"03ef2ae6-c731-44e3-a5cb-3a5b65708092","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5391","names":[{"code":"th","name":"ค่ารับรองและบริการลูกค้า"},{"code":"en","name":"Entertainment & Hospitality Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'ea0ba16b-946d-426d-a528-8e51abc16daf', '5392', 1, '{"id":"ea0ba16b-946d-426d-a528-8e51abc16daf","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5392","names":[{"code":"th","name":"เงินบริจาคเพื่อการกุศลและการศึกษา"},{"code":"en","name":"Charitable & Educational Donations"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '5e5cc698-c400-471f-a653-590e6e32a4d4', '5393', 1, '{"id":"5e5cc698-c400-471f-a653-590e6e32a4d4","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5393","names":[{"code":"th","name":"หนี้สูญและหนี้สงสัยจะสูญ"},{"code":"en","name":"Bad Debts & Doubtful Accounts Expense"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '881b52c7-b50c-4be3-a51b-0f7b75ebfd3b', '5394', 1, '{"id":"881b52c7-b50c-4be3-a51b-0f7b75ebfd3b","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5394","names":[{"code":"th","name":"ขาดทุนจากอัตราแลกเปลี่ยนเงินตราต่างประเทศ"},{"code":"en","name":"Loss on Foreign Exchange"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '94711d11-62d3-4158-ab02-f51fff3d6446', '5395', 1, '{"id":"94711d11-62d3-4158-ab02-f51fff3d6446","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5395","names":[{"code":"th","name":"ขาดทุนจากการจำหน่ายและตัดจำหน่ายทรัพย์สิน"},{"code":"en","name":"Loss on Disposal and Write-off of Assets"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '9b8f7ac6-fc41-4d13-afb1-096b62609888', '5399', 1, '{"id":"9b8f7ac6-fc41-4d13-afb1-096b62609888","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5399","names":[{"code":"th","name":"ค่าใช้จ่ายในการบริหารอื่น"},{"code":"en","name":"Other Administrative Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '88713205-69b7-4de6-af18-710d5a44b8a0', '5400', 1, '{"id":"88713205-69b7-4de6-af18-710d5a44b8a0","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5400","names":[{"code":"th","name":"ต้นทุนทางการเงินและภาษีเงินได้"},{"code":"en","name":"Financial Costs & Income Tax"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c71bfb32-2c8f-4f08-ad5a-9e54586cfc73', '5411', 1, '{"id":"c71bfb32-2c8f-4f08-ad5a-9e54586cfc73","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5411","names":[{"code":"th","name":"ดอกเบี้ยจ่ายเงินกู้ยืมสถาบันการเงิน"},{"code":"en","name":"Interest Expense - Bank Loans"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'c486a79f-c4a3-444f-a997-e2d38395e9ae', '5412', 1, '{"id":"c486a79f-c4a3-444f-a997-e2d38395e9ae","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5412","names":[{"code":"th","name":"ดอกเบี้ยจ่ายเงินเบิกเกินบัญชี (O/D)"},{"code":"en","name":"Interest Expense - Bank Overdraft (O/D)"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '545b4c84-1f38-4f17-a0a2-a577351eeab5', '5413', 1, '{"id":"545b4c84-1f38-4f17-a0a2-a577351eeab5","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5413","names":[{"code":"th","name":"ดอกเบี้ยจ่ายตามสัญญาเช่าทางการเงิน"},{"code":"en","name":"Interest Expense - Financial Leases"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', 'fc12a01f-7e57-4ead-a496-4df6daa0de70', '5421', 1, '{"id":"fc12a01f-7e57-4ead-a496-4df6daa0de70","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5421","names":[{"code":"th","name":"ค่าธรรมเนียมธนาคารและธุรกรรมทางการเงิน"},{"code":"en","name":"Bank Charges & Transaction Fees"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('01', 'accounts', '44049863-be74-44fa-aafd-bcd71e4b2a88', '5431', 1, '{"id":"44049863-be74-44fa-aafd-bcd71e4b2a88","holdingcode":"demo","businesscode":"01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5431","names":[{"code":"th","name":"ภาษีเงินได้นิติบุคคลประจำงวด"},{"code":"en","name":"Corporate Income Tax Expense"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'fiscal-years', '0010f804-dd1b-44a9-aff6-a5b9334a6700', '2569', 1, '{"id":"0010f804-dd1b-44a9-aff6-a5b9334a6700","code":"2569","kind":"fiscal-years","startdate":"2026-01-01","enddate":"2026-12-31","scale":2,"profitlossaccount":"3222","retainedearningsaccount":"3221","closed":false,"isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', '30c89604-7d65-44ac-abef-ec537b7149b1', 'JV', 1, '{"id":"30c89604-7d65-44ac-abef-ec537b7149b1","code":"JV","kind":"journal-books","name":"สมุดรายวันทั่วไป","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', '79e931da-23d0-4812-aa0f-8081ae9073c1', 'PV', 1, '{"id":"79e931da-23d0-4812-aa0f-8081ae9073c1","code":"PV","kind":"journal-books","name":"สมุดรายวันจ่าย","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', '2b0ecf7c-4c64-48b1-ad24-f6108632d440', 'RV', 1, '{"id":"2b0ecf7c-4c64-48b1-ad24-f6108632d440","code":"RV","kind":"journal-books","name":"สมุดรายวันรับ","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', '0a8e1457-3f61-4873-a315-fc13af4d787a', 'SV', 1, '{"id":"0a8e1457-3f61-4873-a315-fc13af4d787a","code":"SV","kind":"journal-books","name":"สมุดรายวันซื้อ","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', '68b35d8e-614e-4aa2-af02-14b7fcfa431a', 'UV', 1, '{"id":"68b35d8e-614e-4aa2-af02-14b7fcfa431a","code":"UV","kind":"journal-books","name":"สมุดรายวันขาย","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', '6ca5538e-1383-405c-a1fa-af236d85c1c1', 'GJ', 1, '{"id":"6ca5538e-1383-405c-a1fa-af236d85c1c1","code":"GJ","kind":"journal-books","name":"สมุดรายวันทั่วไป (GJ)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', '3fc65770-466e-4b6d-a7a0-02b910d655f9', 'AP', 1, '{"id":"3fc65770-466e-4b6d-a7a0-02b910d655f9","code":"AP","kind":"journal-books","name":"สมุดรายวันซื้อเชื่อ (AP)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'journal-books', 'ebb4d35a-361f-4443-ae68-4248cfa392a0', 'AR', 1, '{"id":"ebb4d35a-361f-4443-ae68-4248cfa392a0","code":"AR","kind":"journal-books","name":"สมุดรายวันขายเชื่อ (AR)","isactive":true,"version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO NOTHING;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'account-groups', '2a845c5e-5c53-48c7-af22-a0ec55dbb6e7', 'BM69-A', 1, '{"id":"2a845c5e-5c53-48c7-af22-a0ec55dbb6e7","code":"BM69-A","kind":"account-groups","name":"สินทรัพย์","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'account-groups', 'befbb567-e83c-41fc-a3b3-adfbb60c6d11', 'BM69-L', 1, '{"id":"befbb567-e83c-41fc-a3b3-adfbb60c6d11","code":"BM69-L","kind":"account-groups","name":"หนี้สิน","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'account-groups', '3eee57e0-d48a-4baf-a7e4-1cc1fc492d7e', 'BM69-E', 1, '{"id":"3eee57e0-d48a-4baf-a7e4-1cc1fc492d7e","code":"BM69-E","kind":"account-groups","name":"ส่วนของเจ้าของ","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'account-groups', 'adc8edb7-488e-4e88-a70d-a9edef40cc50', 'BM69-R', 1, '{"id":"adc8edb7-488e-4e88-a70d-a9edef40cc50","code":"BM69-R","kind":"account-groups","name":"รายได้","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'account-groups', '7401420b-67df-424c-aed7-fc74b7aaf148', 'BM69-X', 1, '{"id":"7401420b-67df-424c-aed7-fc74b7aaf148","code":"BM69-X","kind":"account-groups","name":"ค่าใช้จ่าย","amount":"0","locked":false,"version":1,"isactive":true,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","isdeleted":false,"updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","holdingcode":"demo","businesscode":"c01"}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET payload = EXCLUDED.payload;


-- Accounts for Holding: demo, Company: c01
INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f211b163-d2e0-44a2-a399-a8304e28ec39', '1000', 1, '{"id":"f211b163-d2e0-44a2-a399-a8304e28ec39","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1000","names":[{"code":"th","name":"สินทรัพย์"},{"code":"en","name":"Assets"}],"accounttype":"asset","parentaccountcode":null,"normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3f329b56-c66b-4aac-add9-0b3cf3822706', '1100', 1, '{"id":"3f329b56-c66b-4aac-add9-0b3cf3822706","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1100","names":[{"code":"th","name":"สินทรัพย์หมุนเวียน"},{"code":"en","name":"Current Assets"}],"accounttype":"asset","parentaccountcode":"1000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a56c66ac-e6a8-4354-aaf7-f0de33bfefd8', '1111', 1, '{"id":"a56c66ac-e6a8-4354-aaf7-f0de33bfefd8","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1111","names":[{"code":"th","name":"เงินสดในมือ"},{"code":"en","name":"Cash on Hand"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '167b4e78-9d65-4263-a1d2-20343c3633c6', '1112', 1, '{"id":"167b4e78-9d65-4263-a1d2-20343c3633c6","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1112","names":[{"code":"th","name":"เงินสดย่อย - สำนักงานใหญ่"},{"code":"en","name":"Petty Cash - Head Office"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3a6443d6-e3ed-4e09-a2fd-6b0a5ed06c62', '1113', 1, '{"id":"3a6443d6-e3ed-4e09-a2fd-6b0a5ed06c62","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1113","names":[{"code":"th","name":"เงินสดย่อย - ฝ่ายปฏิบัติการและสาขา"},{"code":"en","name":"Petty Cash - Operations & Branch"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'cc977282-0f26-4dd5-a19a-f3aedfa02fe2', '1121', 1, '{"id":"cc977282-0f26-4dd5-a19a-f3aedfa02fe2","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1121","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย"},{"code":"en","name":"Cash at Bank - Current KBANK"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8c74003e-cf0a-4094-ae0b-cc4ca31af3ae', '1122', 1, '{"id":"8c74003e-cf0a-4094-ae0b-cc4ca31af3ae","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1122","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารกรุงเทพ"},{"code":"en","name":"Cash at Bank - Current BBL"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '651be380-0618-48da-a257-64546a630a6c', '1123', 1, '{"id":"651be380-0618-48da-a257-64546a630a6c","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1123","names":[{"code":"th","name":"เงินฝากกระแสรายวัน - ธนาคารไทยพาณิชย์"},{"code":"en","name":"Cash at Bank - Current SCB"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '03c830db-29d9-4951-a119-673fd54a94ef', '1124', 1, '{"id":"03c830db-29d9-4951-a119-673fd54a94ef","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1124","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารกสิกรไทย"},{"code":"en","name":"Cash at Bank - Savings KBANK"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '552032c2-2a2e-4431-a572-e790bcb97267', '1125', 1, '{"id":"552032c2-2a2e-4431-a572-e790bcb97267","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1125","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารกรุงเทพ"},{"code":"en","name":"Cash at Bank - Savings BBL"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ea250ae9-518f-4a7e-a354-dbf7eef41909', '1126', 1, '{"id":"ea250ae9-518f-4a7e-a354-dbf7eef41909","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1126","names":[{"code":"th","name":"เงินฝากออมทรัพย์ - ธนาคารไทยพาณิชย์"},{"code":"en","name":"Cash at Bank - Savings SCB"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '7d626d88-d4d7-4979-a2a1-1563708a0a8e', '1127', 1, '{"id":"7d626d88-d4d7-4979-a2a1-1563708a0a8e","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1127","names":[{"code":"th","name":"เงินฝากประจำระยะสั้น (ไม่เกิน 3 เดือน)"},{"code":"en","name":"Short-term Fixed Deposit"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":true,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '334d34cd-3658-4ac5-a742-30815f3d18ab', '1131', 1, '{"id":"334d34cd-3658-4ac5-a742-30815f3d18ab","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1131","names":[{"code":"th","name":"ลูกหนี้การค้า - ในประเทศ"},{"code":"en","name":"Trade Accounts Receivable - Domestic"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'd0b8b7bc-a4bc-4801-ae10-3d4b1bb6f846', '1132', 1, '{"id":"d0b8b7bc-a4bc-4801-ae10-3d4b1bb6f846","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1132","names":[{"code":"th","name":"ลูกหนี้การค้า - ต่างประเทศ"},{"code":"en","name":"Trade Accounts Receivable - Overseas"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '7dbc5527-bd29-43cc-a82d-c8e6eaec64d7', '1133', 1, '{"id":"7dbc5527-bd29-43cc-a82d-c8e6eaec64d7","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1133","names":[{"code":"th","name":"ตั๋วเงินรับการค้า"},{"code":"en","name":"Notes Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1f3c075b-9385-43be-af7f-de20763c028c', '1134', 1, '{"id":"1f3c075b-9385-43be-af7f-de20763c028c","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1134","names":[{"code":"th","name":"เช็ครับลงวันที่ล่วงหน้า"},{"code":"en","name":"Post-dated Cheques Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e3b003cf-7e12-4e95-a6b4-437f4980baae', '1138', 1, '{"id":"e3b003cf-7e12-4e95-a6b4-437f4980baae","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1138","names":[{"code":"th","name":"ลูกหนี้อื่นและเงินยืมทดรอง"},{"code":"en","name":"Other Receivables & Advances"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '7f6dcfb7-d12c-4a97-a560-d537a28f3532', '1139', 1, '{"id":"7f6dcfb7-d12c-4a97-a560-d537a28f3532","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1139","names":[{"code":"th","name":"ค่าเผื่อผลขาดทุนด้านเครดิตที่คาดว่าจะเกิดขึ้น"},{"code":"en","name":"Allowance for Expected Credit Losses"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2a6bb0eb-9e90-4317-a9a1-c16eb9acf94b', '1141', 1, '{"id":"2a6bb0eb-9e90-4317-a9a1-c16eb9acf94b","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1141","names":[{"code":"th","name":"สินค้าสำเร็จรูป"},{"code":"en","name":"Finished Goods"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f51e8300-5128-4a2d-a3e2-eb060f5bde38', '1142', 1, '{"id":"f51e8300-5128-4a2d-a3e2-eb060f5bde38","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1142","names":[{"code":"th","name":"สินค้าระหว่างทำ / งานระหว่างทำ"},{"code":"en","name":"Work in Process"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '4c42ed1b-d6ed-4d21-a25c-d498da8907b3', '1143', 1, '{"id":"4c42ed1b-d6ed-4d21-a25c-d498da8907b3","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1143","names":[{"code":"th","name":"วัตถุดิบและส่วนประกอบ"},{"code":"en","name":"Raw Materials & Components"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'd1123e16-04dc-4c32-aba1-f3f72451b47d', '1144', 1, '{"id":"d1123e16-04dc-4c32-aba1-f3f72451b47d","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1144","names":[{"code":"th","name":"วัสดุสิ้นเปลืองและบรรจุภัณฑ์"},{"code":"en","name":"Factory Supplies & Packaging"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b5b2de51-241f-4918-a1c5-a4401cdc4202', '1145', 1, '{"id":"b5b2de51-241f-4918-a1c5-a4401cdc4202","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1145","names":[{"code":"th","name":"สินค้าระหว่างทาง"},{"code":"en","name":"Goods in Transit"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'd74d868d-76c1-4b97-a4b1-fd64905029f5', '1149', 1, '{"id":"d74d868d-76c1-4b97-a4b1-fd64905029f5","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1149","names":[{"code":"th","name":"ค่าเผื่อการลดมูลค่าสินค้าคงเหลือ"},{"code":"en","name":"Allowance for Inventory Devaluation"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ee7a2094-8824-49c5-aa0a-302a7c993010', '1151', 1, '{"id":"ee7a2094-8824-49c5-aa0a-302a7c993010","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1151","names":[{"code":"th","name":"ภาษีซื้อ"},{"code":"en","name":"Input VAT"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '86179e0b-9c27-41f2-a7a8-31979051c0af', '1152', 1, '{"id":"86179e0b-9c27-41f2-a7a8-31979051c0af","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1152","names":[{"code":"th","name":"ภาษีซื้อยังไม่ถึงกำหนดชำระ"},{"code":"en","name":"Undue Input VAT"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'aea238c3-3f49-4ea6-a5e0-7af76e99bcae', '1153', 1, '{"id":"aea238c3-3f49-4ea6-a5e0-7af76e99bcae","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1153","names":[{"code":"th","name":"ภาษีเงินได้ถูกหัก ณ ที่จ่าย (WHT ถูกหัก)"},{"code":"en","name":"Withholding Tax Receivable"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'de575542-5962-43bd-ae9e-c86df6bc552f', '1154', 1, '{"id":"de575542-5962-43bd-ae9e-c86df6bc552f","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1154","names":[{"code":"th","name":"ภาษีเงินได้นิติบุคคลจ่ายล่วงหน้า (ภ.ง.ด.51)"},{"code":"en","name":"Prepaid Corporate Income Tax"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'fd626712-448b-4f57-a5f8-594bfd559881', '1161', 1, '{"id":"fd626712-448b-4f57-a5f8-594bfd559881","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1161","names":[{"code":"th","name":"ค่าเบี้ยประกันภัยจ่ายล่วงหน้า"},{"code":"en","name":"Prepaid Insurance"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1d181110-f226-4df4-af07-755bf7f9e7a5', '1162', 1, '{"id":"1d181110-f226-4df4-af07-755bf7f9e7a5","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1162","names":[{"code":"th","name":"ค่าเช่าจ่ายล่วงหน้า"},{"code":"en","name":"Prepaid Rent"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '54e4e1fc-010f-4e23-ab69-054c833d3661', '1163', 1, '{"id":"54e4e1fc-010f-4e23-ab69-054c833d3661","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1163","names":[{"code":"th","name":"เงินมัดจำค่าสินค้าและบริการล่วงหน้า"},{"code":"en","name":"Advance Payments for Goods & Services"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c5b26233-1006-4b23-a60d-e29a71540470', '1169', 1, '{"id":"c5b26233-1006-4b23-a60d-e29a71540470","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1169","names":[{"code":"th","name":"สินทรัพย์หมุนเวียนอื่น"},{"code":"en","name":"Other Current Assets"}],"accounttype":"asset","parentaccountcode":"1100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '17819c66-54a9-46a9-abe3-48c971576a31', '1200', 1, '{"id":"17819c66-54a9-46a9-abe3-48c971576a31","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1200","names":[{"code":"th","name":"สินทรัพย์ไม่หมุนเวียน"},{"code":"en","name":"Non-Current Assets"}],"accounttype":"asset","parentaccountcode":"1000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a8025b0e-33be-469f-ae1f-9b20355bbf66', '1211', 1, '{"id":"a8025b0e-33be-469f-ae1f-9b20355bbf66","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1211","names":[{"code":"th","name":"ที่ดิน"},{"code":"en","name":"Land"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '375d0749-f43c-4245-a8e0-69f16dfbd842', '1212', 1, '{"id":"375d0749-f43c-4245-a8e0-69f16dfbd842","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1212","names":[{"code":"th","name":"ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Land Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2b313829-78f0-4f0b-aae8-1d2bae444b27', '1213', 1, '{"id":"2b313829-78f0-4f0b-aae8-1d2bae444b27","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1213","names":[{"code":"th","name":"อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Buildings & Constructions"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '7a55b7e0-ff8f-4bb5-a7ea-bdb251aacc2b', '1214', 1, '{"id":"7a55b7e0-ff8f-4bb5-a7ea-bdb251aacc2b","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1214","names":[{"code":"th","name":"ส่วนปรับปรุงอาคารและระบบสาธารณูปโภค"},{"code":"en","name":"Building Improvements & Systems"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '30054a5b-b14f-4a57-a908-35ec8ca6dc98', '1221', 1, '{"id":"30054a5b-b14f-4a57-a908-35ec8ca6dc98","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1221","names":[{"code":"th","name":"เครื่องจักรและอุปกรณ์โรงงาน"},{"code":"en","name":"Machinery & Factory Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'eef0697f-952c-4050-a46b-bd9efb4bc908', '1222', 1, '{"id":"eef0697f-952c-4050-a46b-bd9efb4bc908","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1222","names":[{"code":"th","name":"เครื่องตกแต่ง ติดตั้ง และเฟอร์นิเจอร์"},{"code":"en","name":"Furniture & Fixtures"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '383d63e6-d668-47bd-a65e-31183e27a628', '1223', 1, '{"id":"383d63e6-d668-47bd-a65e-31183e27a628","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1223","names":[{"code":"th","name":"เครื่องใช้และอุปกรณ์สำนักงาน"},{"code":"en","name":"Office Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '0cd49133-a614-459f-ac64-03a59ffeaa07', '1224', 1, '{"id":"0cd49133-a614-459f-ac64-03a59ffeaa07","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1224","names":[{"code":"th","name":"คอมพิวเตอร์และอุปกรณ์ประมวลผล"},{"code":"en","name":"Computer & Hardware Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b15667ca-cc95-4346-a562-0444d6adb00a', '1231', 1, '{"id":"b15667ca-cc95-4346-a562-0444d6adb00a","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1231","names":[{"code":"th","name":"ยานพาหนะและรถบรรทุกขนส่ง"},{"code":"en","name":"Vehicles & Transport Trucks"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ede2f729-0a52-497e-a43d-d2d1b14aa0c2', '1241', 1, '{"id":"ede2f729-0a52-497e-a43d-d2d1b14aa0c2","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1241","names":[{"code":"th","name":"งานระหว่างก่อสร้างและติดตั้งเครื่องจักร"},{"code":"en","name":"Construction & Installation in Progress"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '22b9ea22-2aef-4b80-aa2a-7b4855101a36', '1281', 1, '{"id":"22b9ea22-2aef-4b80-aa2a-7b4855101a36","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1281","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Acc. Dep. - Land Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8626383b-abaa-4639-a781-eb7fdb9c5301', '1282', 1, '{"id":"8626383b-abaa-4639-a781-eb7fdb9c5301","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1282","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Acc. Dep. - Buildings"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e9ef3719-704c-454b-a732-b8ac7978b933', '1283', 1, '{"id":"e9ef3719-704c-454b-a732-b8ac7978b933","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1283","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ส่วนปรับปรุงอาคาร"},{"code":"en","name":"Acc. Dep. - Building Improvements"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '375ee45d-370c-4570-a30d-24c2dfb03361', '1284', 1, '{"id":"375ee45d-370c-4570-a30d-24c2dfb03361","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1284","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องจักรและอุปกรณ์"},{"code":"en","name":"Acc. Dep. - Machinery & Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6d89c696-5b96-40f2-a759-22219996c585', '1285', 1, '{"id":"6d89c696-5b96-40f2-a759-22219996c585","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1285","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องตกแต่งและติดตั้ง"},{"code":"en","name":"Acc. Dep. - Furniture & Fixtures"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '097bbc3b-0af7-4367-a1db-dfc237279318', '1286', 1, '{"id":"097bbc3b-0af7-4367-a1db-dfc237279318","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1286","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - เครื่องใช้สำนักงาน"},{"code":"en","name":"Acc. Dep. - Office Equipment"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '40683678-52d0-44e0-a758-68b92789c2a1', '1287', 1, '{"id":"40683678-52d0-44e0-a758-68b92789c2a1","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1287","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - คอมพิวเตอร์และอุปกรณ์"},{"code":"en","name":"Acc. Dep. - Computer & Hardware"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '64003f0f-debc-4a66-adad-a9a8cafc9a4f', '1288', 1, '{"id":"64003f0f-debc-4a66-adad-a9a8cafc9a4f","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1288","names":[{"code":"th","name":"ค่าเสื่อมราคาสะสม - ยานพาหนะ"},{"code":"en","name":"Acc. Dep. - Vehicles"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '15fb4898-58b2-4602-a49a-24b21101d814', '1291', 1, '{"id":"15fb4898-58b2-4602-a49a-24b21101d814","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1291","names":[{"code":"th","name":"โปรแกรมคอมพิวเตอร์และสิทธิการใช้งาน"},{"code":"en","name":"Software & Licenses"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '49a53a33-9200-48f9-ad33-f816cc0adba3', '1292', 1, '{"id":"49a53a33-9200-48f9-ad33-f816cc0adba3","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1292","names":[{"code":"th","name":"ค่าตัดจำหน่ายสะสม - โปรแกรมคอมพิวเตอร์"},{"code":"en","name":"Acc. Amortization - Software"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '52603aa2-9d06-4ec4-a792-6676864d93d8', '1295', 1, '{"id":"52603aa2-9d06-4ec4-a792-6676864d93d8","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"1295","names":[{"code":"th","name":"เงินประกันและเงินมัดจำระยะยาว"},{"code":"en","name":"Long-term Deposits & Guarantees"}],"accounttype":"asset","parentaccountcode":"1200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-A","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '0eab3b2e-ef07-4e0e-a1dd-3fc1dfbb5244', '2000', 1, '{"id":"0eab3b2e-ef07-4e0e-a1dd-3fc1dfbb5244","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2000","names":[{"code":"th","name":"หนี้สิน"},{"code":"en","name":"Liabilities"}],"accounttype":"liability","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '5bb182c6-c98d-4f15-aee6-5b4599abf5e1', '2100', 1, '{"id":"5bb182c6-c98d-4f15-aee6-5b4599abf5e1","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2100","names":[{"code":"th","name":"หนี้สินหมุนเวียน"},{"code":"en","name":"Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ec576ce4-9796-49d7-ac20-feded2a2521d', '2111', 1, '{"id":"ec576ce4-9796-49d7-ac20-feded2a2521d","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2111","names":[{"code":"th","name":"เจ้าหนี้การค้า - ในประเทศ"},{"code":"en","name":"Trade Accounts Payable - Domestic"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '9e998882-caef-4b7f-ab92-59e34f9c2ab3', '2112', 1, '{"id":"9e998882-caef-4b7f-ab92-59e34f9c2ab3","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2112","names":[{"code":"th","name":"เจ้าหนี้การค้า - ต่างประเทศ"},{"code":"en","name":"Trade Accounts Payable - Overseas"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8fbd0fe5-6239-497d-a891-292e4a3bf2d5', '2113', 1, '{"id":"8fbd0fe5-6239-497d-a891-292e4a3bf2d5","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2113","names":[{"code":"th","name":"ตั๋วเงินจ่ายการค้า"},{"code":"en","name":"Notes Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'cb2a65c9-96dd-411d-a788-bce6a345539a', '2114', 1, '{"id":"cb2a65c9-96dd-411d-a788-bce6a345539a","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2114","names":[{"code":"th","name":"เช็คจ่ายลงวันที่ล่วงหน้า"},{"code":"en","name":"Post-dated Cheques Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a3c18a2b-5faf-4b2f-a8b5-09ed9c9511b7', '2121', 1, '{"id":"a3c18a2b-5faf-4b2f-a8b5-09ed9c9511b7","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2121","names":[{"code":"th","name":"เจ้าหนี้อื่นและเงินทดรองรับ"},{"code":"en","name":"Other Payables & Advances"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '65fec728-c928-4c17-ae7b-3938fe5aa0ce', '2122', 1, '{"id":"65fec728-c928-4c17-ae7b-3938fe5aa0ce","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2122","names":[{"code":"th","name":"เจ้าหนี้กรมสรรพากร"},{"code":"en","name":"Revenue Department Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '709279f0-3d65-4c61-aec2-354b1478fb89', '2131', 1, '{"id":"709279f0-3d65-4c61-aec2-354b1478fb89","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2131","names":[{"code":"th","name":"ภาษีขาย"},{"code":"en","name":"Output VAT"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '4e369a1e-ae4e-49e0-aafe-0dc059547941', '2132', 1, '{"id":"4e369a1e-ae4e-49e0-aafe-0dc059547941","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2132","names":[{"code":"th","name":"ภาษีขายยังไม่ถึงกำหนดชำระ"},{"code":"en","name":"Undue Output VAT"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '40b00be7-df80-4a0c-abc7-bf627f48a203', '2141', 1, '{"id":"40b00be7-df80-4a0c-abc7-bf627f48a203","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2141","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.1 (เงินเดือน)"},{"code":"en","name":"WHT Payable - P.N.D.1 (Salaries)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c7c943cb-053e-48bd-ab0d-f4548d046dab', '2142', 1, '{"id":"c7c943cb-053e-48bd-ab0d-f4548d046dab","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2142","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.3 (บุคคลธรรมดา)"},{"code":"en","name":"WHT Payable - P.N.D.3 (Individuals)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '132ec42e-5cd7-4490-aff4-9add963989a9', '2143', 1, '{"id":"132ec42e-5cd7-4490-aff4-9add963989a9","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2143","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.53 (นิติบุคคล)"},{"code":"en","name":"WHT Payable - P.N.D.53 (Corporations)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '0b785eb4-4ef8-4b46-a142-e0483449d316', '2144', 1, '{"id":"0b785eb4-4ef8-4b46-a142-e0483449d316","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2144","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.2 (ดอกเบี้ย/ปันผล)"},{"code":"en","name":"WHT Payable - P.N.D.2"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'eba4f1e6-90b0-43e6-a7b8-441dd3e8fa12', '2145', 1, '{"id":"eba4f1e6-90b0-43e6-a7b8-441dd3e8fa12","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2145","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.54 (ส่งไปต่างประเทศ)"},{"code":"en","name":"WHT Payable - P.N.D.54 (Overseas)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b58ad237-b029-43ec-a9f5-e381bd5cbf1d', '2151', 1, '{"id":"b58ad237-b029-43ec-a9f5-e381bd5cbf1d","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2151","names":[{"code":"th","name":"เงินสมทบกองทุนประกันสังคมค้างจ่าย (ลูกจ้าง+นายจ้าง)"},{"code":"en","name":"Social Security Fund Payable"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e8ed2655-895c-41d3-adf5-50cfd34521d1', '2152', 1, '{"id":"e8ed2655-895c-41d3-adf5-50cfd34521d1","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2152","names":[{"code":"th","name":"เงินเดือนและค่าจ้างค้างจ่าย"},{"code":"en","name":"Accrued Salaries & Wages"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f0b5b0a5-2c45-48d9-a68d-95a80fccda6d', '2153', 1, '{"id":"f0b5b0a5-2c45-48d9-a68d-95a80fccda6d","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2153","names":[{"code":"th","name":"ค่าเช่าค้างจ่าย"},{"code":"en","name":"Accrued Rent Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8e4e885f-add4-4cef-a012-72fbbb53380b', '2154', 1, '{"id":"8e4e885f-add4-4cef-a012-72fbbb53380b","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2154","names":[{"code":"th","name":"ค่าน้ำประปาและค่าไฟฟ้าค้างจ่าย"},{"code":"en","name":"Accrued Utilities Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '830e9a62-1f88-4927-a559-eba15728e465', '2155', 1, '{"id":"830e9a62-1f88-4927-a559-eba15728e465","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2155","names":[{"code":"th","name":"ค่าโทรศัพท์และอินเทอร์เน็ตค้างจ่าย"},{"code":"en","name":"Accrued Telephone & Internet"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e81cd5fb-681a-4491-acbf-4f9c764d86dc', '2156', 1, '{"id":"e81cd5fb-681a-4491-acbf-4f9c764d86dc","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2156","names":[{"code":"th","name":"ค่าสอบบัญชีและทำบัญชีค้างจ่าย"},{"code":"en","name":"Accrued Audit & Accounting Fees"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '33a85415-8628-466d-ab42-2c9bdb2b60af', '2157', 1, '{"id":"33a85415-8628-466d-ab42-2c9bdb2b60af","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2157","names":[{"code":"th","name":"ดอกเบี้ยค้างจ่าย"},{"code":"en","name":"Accrued Interest Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '117fc25a-4ed4-4f03-a70d-e37fee95ffd2', '2158', 1, '{"id":"117fc25a-4ed4-4f03-a70d-e37fee95ffd2","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2158","names":[{"code":"th","name":"โบนัสพนักงานค้างจ่าย"},{"code":"en","name":"Accrued Staff Bonuses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c03cc429-72ae-4cce-a50c-3ab76a3edbf9', '2159', 1, '{"id":"c03cc429-72ae-4cce-a50c-3ab76a3edbf9","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2159","names":[{"code":"th","name":"ค่าใช้จ่ายค้างจ่ายอื่น"},{"code":"en","name":"Other Accrued Expenses"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8a603daf-e0bb-4907-a608-bfc3679bf7a8', '2161', 1, '{"id":"8a603daf-e0bb-4907-a608-bfc3679bf7a8","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2161","names":[{"code":"th","name":"เงินรับล่วงหน้าค่าสินค้าและบริการจากลูกค้า"},{"code":"en","name":"Advances Received from Customers"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2e1db72c-2006-4a9b-a6b4-2faed20b8dcb', '2171', 1, '{"id":"2e1db72c-2006-4a9b-a6b4-2faed20b8dcb","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2171","names":[{"code":"th","name":"เงินเบิกเกินบัญชีธนาคาร (O/D)"},{"code":"en","name":"Bank Overdrafts (O/D)"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '08fb275c-200f-47f8-a1ef-97a67e3f6c83', '2172', 1, '{"id":"08fb275c-200f-47f8-a1ef-97a67e3f6c83","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2172","names":[{"code":"th","name":"เงินกู้ยืมระยะสั้นจากสถาบันการเงิน"},{"code":"en","name":"Short-term Borrowings from Banks"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1009ab86-c220-47f4-a9d0-0b225674fb41', '2173', 1, '{"id":"1009ab86-c220-47f4-a9d0-0b225674fb41","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2173","names":[{"code":"th","name":"เงินกู้ยืมระยะสั้นจากกรรมการหรือบุคคลที่เกี่ยวข้องกัน"},{"code":"en","name":"Short-term Loans from Directors"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'd6679964-03ce-4835-a90e-e62aa074c661', '2181', 1, '{"id":"d6679964-03ce-4835-a90e-e62aa074c661","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2181","names":[{"code":"th","name":"ส่วนของหนี้สินระยะยาวที่ถึงกำหนดชำระภายในหนึ่งปี"},{"code":"en","name":"Current Portion of Long-term Debt"}],"accounttype":"liability","parentaccountcode":"2100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'cb60eaf9-c2ad-4e5d-a4c9-85eed984f693', '2200', 1, '{"id":"cb60eaf9-c2ad-4e5d-a4c9-85eed984f693","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2200","names":[{"code":"th","name":"หนี้สินไม่หมุนเวียน"},{"code":"en","name":"Non-Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2016927b-6fb6-45ee-a8a1-2c8bc81c0c69', '2211', 1, '{"id":"2016927b-6fb6-45ee-a8a1-2c8bc81c0c69","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2211","names":[{"code":"th","name":"เงินกู้ยืมระยะยาวจากสถาบันการเงิน"},{"code":"en","name":"Long-term Bank Loans"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6d530185-27ee-4c47-a736-c74fec70ece8', '2212', 1, '{"id":"6d530185-27ee-4c47-a736-c74fec70ece8","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2212","names":[{"code":"th","name":"หนี้สินตามสัญญาเช่าทางการเงินระยะยาว"},{"code":"en","name":"Long-term Financial Lease Liabilities"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2fddb497-cae7-416a-a1b7-da400f2c1df1', '2221', 1, '{"id":"2fddb497-cae7-416a-a1b7-da400f2c1df1","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2221","names":[{"code":"th","name":"เงินกู้ยืมระยะยาวจากกรรมการหรือผู้ถือหุ้น"},{"code":"en","name":"Long-term Loans from Directors/Shareholders"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'be1082cd-dd8f-4717-af1e-9cd2e83b4780', '2231', 1, '{"id":"be1082cd-dd8f-4717-af1e-9cd2e83b4780","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2231","names":[{"code":"th","name":"ประมาณการหนี้สินผลประโยชน์พนักงาน"},{"code":"en","name":"Provision for Employee Benefits"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'cd370de5-2193-4fea-a217-dd05949f754e', '2291', 1, '{"id":"cd370de5-2193-4fea-a217-dd05949f754e","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"2291","names":[{"code":"th","name":"หนี้สินไม่หมุนเวียนอื่น"},{"code":"en","name":"Other Non-Current Liabilities"}],"accounttype":"liability","parentaccountcode":"2200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-L","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c9dc85fd-cbc7-4b9e-aacb-c46a4b639aaa', '3000', 1, '{"id":"c9dc85fd-cbc7-4b9e-aacb-c46a4b639aaa","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3000","names":[{"code":"th","name":"ส่วนของเจ้าของ"},{"code":"en","name":"Equity"}],"accounttype":"equity","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '74056e95-f117-4cb1-abaa-eb2b24fdad41', '3100', 1, '{"id":"74056e95-f117-4cb1-abaa-eb2b24fdad41","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3100","names":[{"code":"th","name":"ทุนจดทะเบียนและส่วนเกินทุน"},{"code":"en","name":"Share Capital & Premium"}],"accounttype":"equity","parentaccountcode":"3000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '32e84400-43d1-4296-afff-0e0b1e6a725a', '3111', 1, '{"id":"32e84400-43d1-4296-afff-0e0b1e6a725a","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3111","names":[{"code":"th","name":"ทุนเรือนหุ้น - หุ้นสามัญ"},{"code":"en","name":"Authorized Share Capital - Common Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '15b2ee58-b048-49a7-a07a-ab82e4ef5e61', '3112', 1, '{"id":"15b2ee58-b048-49a7-a07a-ab82e4ef5e61","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3112","names":[{"code":"th","name":"ทุนเรือนหุ้น - หุ้นบุริมสิทธิ"},{"code":"en","name":"Authorized Share Capital - Preferred Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '41327aac-6a04-40df-ace8-5bae94b927ef', '3121', 1, '{"id":"41327aac-6a04-40df-ace8-5bae94b927ef","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3121","names":[{"code":"th","name":"ส่วนเกินมูลค่าหุ้นสามัญ"},{"code":"en","name":"Share Premium - Common Shares"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'dc398320-1e8e-4648-a64e-b4fa9afa9ad7', '3131', 1, '{"id":"dc398320-1e8e-4648-a64e-b4fa9afa9ad7","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3131","names":[{"code":"th","name":"ทุนส่วนของเจ้าของ (ห้างหุ้นส่วน/บุคคลธรรมดา)"},{"code":"en","name":"Owner''s / Partner''s Capital"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a9320ca6-ebe9-446b-a826-e852e992fa32', '3132', 1, '{"id":"a9320ca6-ebe9-446b-a826-e852e992fa32","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3132","names":[{"code":"th","name":"เงินถอนใช้ส่วนตัวของเจ้าของกิจการ"},{"code":"en","name":"Owner''s Drawings"}],"accounttype":"equity","parentaccountcode":"3100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b580f02d-aaec-4f0b-a9e6-6a072715c442', '3200', 1, '{"id":"b580f02d-aaec-4f0b-a9e6-6a072715c442","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3200","names":[{"code":"th","name":"กำไร (ขาดทุน) สะสม"},{"code":"en","name":"Retained Earnings"}],"accounttype":"equity","parentaccountcode":"3000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1e6d0249-cc74-4d58-a00b-16de67b0ee53', '3211', 1, '{"id":"1e6d0249-cc74-4d58-a00b-16de67b0ee53","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3211","names":[{"code":"th","name":"สำรองตามกฎหมาย"},{"code":"en","name":"Legal Reserve"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6d86a084-0bd4-4b98-a241-d18fbaf87eb1', '3212', 1, '{"id":"6d86a084-0bd4-4b98-a241-d18fbaf87eb1","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3212","names":[{"code":"th","name":"สำรองอื่นเพื่อวัตถุประสงค์เฉพาะ"},{"code":"en","name":"Other Appropriated Reserves"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3af0418f-a3c3-4dd7-a69c-b7422e78025c', '3221', 1, '{"id":"3af0418f-a3c3-4dd7-a69c-b7422e78025c","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3221","names":[{"code":"th","name":"กำไร (ขาดทุน) สะสมยังไม่ได้จัดสรร"},{"code":"en","name":"Unappropriated Retained Earnings"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'd6d0a347-8c41-44de-a285-d9175c54787a', '3222', 1, '{"id":"d6d0a347-8c41-44de-a285-d9175c54787a","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3222","names":[{"code":"th","name":"กำไร (ขาดทุน) สุทธิประจำปี"},{"code":"en","name":"Net Profit/Loss for Current Year"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6b051383-34fe-4367-ac54-6846acb41f83', '3223', 1, '{"id":"6b051383-34fe-4367-ac54-6846acb41f83","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"3223","names":[{"code":"th","name":"เงินปันผลจ่าย"},{"code":"en","name":"Dividends Declared & Paid"}],"accounttype":"equity","parentaccountcode":"3200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-E","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '671e3f7d-9204-4d05-a09c-7ae315ae0c2e', '4000', 1, '{"id":"671e3f7d-9204-4d05-a09c-7ae315ae0c2e","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4000","names":[{"code":"th","name":"รายได้"},{"code":"en","name":"Income"}],"accounttype":"income","parentaccountcode":null,"normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3fee72b7-3724-45e2-a768-4dcb09df1bdf', '4100', 1, '{"id":"3fee72b7-3724-45e2-a768-4dcb09df1bdf","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4100","names":[{"code":"th","name":"รายได้จากการขายและการให้บริการ"},{"code":"en","name":"Revenue from Sales & Services"}],"accounttype":"income","parentaccountcode":"4000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2c14bcff-cd8a-4839-a0db-a2757e46b252', '4111', 1, '{"id":"2c14bcff-cd8a-4839-a0db-a2757e46b252","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4111","names":[{"code":"th","name":"รายได้จากการขายสินค้า - มีภาษีมูลค่าเพิ่ม 7%"},{"code":"en","name":"Sales Revenue - VAT 7%"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '91c5c4ee-987e-4a56-a88a-5226a0451b8b', '4112', 1, '{"id":"91c5c4ee-987e-4a56-a88a-5226a0451b8b","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4112","names":[{"code":"th","name":"รายได้จากการขายสินค้า - อัตราภาษี 0% / ส่งออก"},{"code":"en","name":"Sales Revenue - Zero Rated / Export"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '26378caa-490d-412e-ae86-6e8b49099329', '4113', 1, '{"id":"26378caa-490d-412e-ae86-6e8b49099329","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4113","names":[{"code":"th","name":"รายได้จากการขายสินค้า - ได้รับยกเว้นภาษีมูลค่าเพิ่ม"},{"code":"en","name":"Sales Revenue - VAT Exempted"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2c1be38f-dd34-4a93-aebe-eafb2b52f6cc', '4121', 1, '{"id":"2c1be38f-dd34-4a93-aebe-eafb2b52f6cc","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4121","names":[{"code":"th","name":"รายได้จากการให้บริการ - มีภาษีมูลค่าเพิ่ม 7%"},{"code":"en","name":"Service Income - VAT 7%"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '19b47134-8b4e-43d4-a7a2-90e6ea6c7635', '4122', 1, '{"id":"19b47134-8b4e-43d4-a7a2-90e6ea6c7635","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4122","names":[{"code":"th","name":"รายได้จากการให้บริการ - ได้รับยกเว้นภาษีมูลค่าเพิ่ม"},{"code":"en","name":"Service Income - VAT Exempted"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8fa87e71-09ff-416e-a1b9-68e0dc993371', '4131', 1, '{"id":"8fa87e71-09ff-416e-a1b9-68e0dc993371","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4131","names":[{"code":"th","name":"รับคืนสินค้าและลดหนี้ขาย"},{"code":"en","name":"Sales Returns and Allowances"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2670cebc-48e0-47fe-a8a0-87afba0d787f', '4132', 1, '{"id":"2670cebc-48e0-47fe-a8a0-87afba0d787f","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4132","names":[{"code":"th","name":"ส่วนลดจ่าย"},{"code":"en","name":"Sales Cash Discounts"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e8c252fa-3304-4782-a354-98f65e8f6fbd', '4141', 1, '{"id":"e8c252fa-3304-4782-a354-98f65e8f6fbd","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4141","names":[{"code":"th","name":"รายได้ค่าบริการขนส่งและจัดส่งสินค้า"},{"code":"en","name":"Freight & Delivery Income"}],"accounttype":"income","parentaccountcode":"4100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f60dee21-fc72-4107-af4e-aac5ea53058e', '4200', 1, '{"id":"f60dee21-fc72-4107-af4e-aac5ea53058e","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4200","names":[{"code":"th","name":"รายได้อื่น"},{"code":"en","name":"Other Income"}],"accounttype":"income","parentaccountcode":"4000","normalbalance":"credit","allowposting":false,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8d593ec6-f47b-47fc-a4ed-8e8733c2bf1f', '4211', 1, '{"id":"8d593ec6-f47b-47fc-a4ed-8e8733c2bf1f","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4211","names":[{"code":"th","name":"ดอกเบี้ยรับจากสถาบันการเงิน"},{"code":"en","name":"Interest Income from Banks"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '14e906a8-196b-4c81-a566-300c507462d1', '4212', 1, '{"id":"14e906a8-196b-4c81-a566-300c507462d1","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4212","names":[{"code":"th","name":"เงินปันผลรับจากเงินลงทุน"},{"code":"en","name":"Dividend Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '7fce582e-7891-4c47-a06d-837d454dd3a5', '4221', 1, '{"id":"7fce582e-7891-4c47-a06d-837d454dd3a5","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4221","names":[{"code":"th","name":"กำไรจากการจำหน่ายทรัพย์สินถาวร"},{"code":"en","name":"Gain on Disposal of Fixed Assets"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a30bc172-0c84-45ab-ab42-0e37e8b4d048', '4231', 1, '{"id":"a30bc172-0c84-45ab-ab42-0e37e8b4d048","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4231","names":[{"code":"th","name":"กำไรจากอัตราแลกเปลี่ยนเงินตราต่างประเทศ"},{"code":"en","name":"Gain on Foreign Exchange"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e4f6ce12-628d-4e05-a2f7-ac84f9bf5957', '4241', 1, '{"id":"e4f6ce12-628d-4e05-a2f7-ac84f9bf5957","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4241","names":[{"code":"th","name":"รายได้ค่าเช่าอาคารและอุปกรณ์"},{"code":"en","name":"Rental Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '53609b32-7053-47c8-a2fc-59c039ec2144', '4251', 1, '{"id":"53609b32-7053-47c8-a2fc-59c039ec2144","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4251","names":[{"code":"th","name":"หนี้สูญได้รับคืน"},{"code":"en","name":"Bad Debts Recovered"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'eafc446e-95e6-4183-ab94-42f87351dc31', '4291', 1, '{"id":"eafc446e-95e6-4183-ab94-42f87351dc31","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"4291","names":[{"code":"th","name":"รายได้เบ็ดเตล็ดอื่น"},{"code":"en","name":"Miscellaneous Income"}],"accounttype":"income","parentaccountcode":"4200","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-R","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '7601ba7c-06eb-4898-acd1-c3b4cfd7285b', '5000', 1, '{"id":"7601ba7c-06eb-4898-acd1-c3b4cfd7285b","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5000","names":[{"code":"th","name":"ค่าใช้จ่าย"},{"code":"en","name":"Expenses"}],"accounttype":"expense","parentaccountcode":null,"normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":1}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'eb4381aa-40cd-431c-a06b-ba141106dfd6', '5100', 1, '{"id":"eb4381aa-40cd-431c-a06b-ba141106dfd6","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5100","names":[{"code":"th","name":"ต้นทุนขายและบริการ"},{"code":"en","name":"Cost of Goods Sold & Services"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '6c5c61af-1824-42e2-aa80-b4cba9b6162b', '5111', 1, '{"id":"6c5c61af-1824-42e2-aa80-b4cba9b6162b","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5111","names":[{"code":"th","name":"ซื้อสินค้าสำเร็จรูป"},{"code":"en","name":"Purchases of Merchandise"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '3945ed15-a8ee-4f7b-ae0e-40e510773251', '5112', 1, '{"id":"3945ed15-a8ee-4f7b-ae0e-40e510773251","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5112","names":[{"code":"th","name":"ค่าขนส่งเข้า"},{"code":"en","name":"Freight-In"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8dc6b12f-f16d-41d7-ab33-a98487251b10', '5113', 1, '{"id":"8dc6b12f-f16d-41d7-ab33-a98487251b10","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5113","names":[{"code":"th","name":"ส่งคืนสินค้าและส่วนลดที่ได้รับ"},{"code":"en","name":"Purchase Returns and Allowances"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'bd346e98-0646-44dc-a71a-72e2f0788120', '5114', 1, '{"id":"bd346e98-0646-44dc-a71a-72e2f0788120","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5114","names":[{"code":"th","name":"ส่วนลดรับ"},{"code":"en","name":"Purchase Cash Discounts"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"credit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f60b3a95-010d-4f14-a507-d4949ef4e0a5', '5121', 1, '{"id":"f60b3a95-010d-4f14-a507-d4949ef4e0a5","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5121","names":[{"code":"th","name":"ต้นทุนสินค้าสำเร็จรูปที่ขาย"},{"code":"en","name":"Cost of Goods Sold"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '727dedae-7a56-4a4b-af52-91e9ff961ea5', '5131', 1, '{"id":"727dedae-7a56-4a4b-af52-91e9ff961ea5","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5131","names":[{"code":"th","name":"ต้นทุนค่าแรงและบริการ"},{"code":"en","name":"Direct Labor & Service Costs"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '535c4c90-4df8-4a3e-a722-5e3c0e8cfad5', '5132', 1, '{"id":"535c4c90-4df8-4a3e-a722-5e3c0e8cfad5","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5132","names":[{"code":"th","name":"ค่าจ้างเหมาบริการช่วงภายนอก (Subcontractor)"},{"code":"en","name":"Subcontractor Costs"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e7e49aff-b46a-408b-a84f-791366d9e7e1', '5141', 1, '{"id":"e7e49aff-b46a-408b-a84f-791366d9e7e1","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5141","names":[{"code":"th","name":"สินค้าสูญหาย เสียหาย และสินค้าชำรุด"},{"code":"en","name":"Inventory Shrinkage & Damage"}],"accounttype":"expense","parentaccountcode":"5100","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ba162ad4-2ad8-44af-a95d-35f9e598cbfd', '5200', 1, '{"id":"ba162ad4-2ad8-44af-a95d-35f9e598cbfd","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5200","names":[{"code":"th","name":"ค่าใช้จ่ายในการขาย"},{"code":"en","name":"Selling Expenses"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2ea0c861-1a55-4746-a4f0-73d7f4f87349', '5211', 1, '{"id":"2ea0c861-1a55-4746-a4f0-73d7f4f87349","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5211","names":[{"code":"th","name":"เงินเดือน ค่าล่วงเวลา และเบี้ยเลี้ยงฝ่ายขาย"},{"code":"en","name":"Sales Salaries, Overtime & Allowances"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1ffd018c-1388-435a-aa8f-72cc32a18779', '5212', 1, '{"id":"1ffd018c-1388-435a-aa8f-72cc32a18779","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5212","names":[{"code":"th","name":"ค่านายหน้าและค่าคอมมิชชั่นฝ่ายขาย"},{"code":"en","name":"Sales Commissions & Incentives"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '0139d802-1678-491c-aa0d-b344956941fe', '5221', 1, '{"id":"0139d802-1678-491c-aa0d-b344956941fe","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5221","names":[{"code":"th","name":"ค่าโฆษณา ประชาสัมพันธ์ และการตลาดออนไลน์"},{"code":"en","name":"Advertising, PR & Online Marketing"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e39d660b-0350-49eb-ae40-5e17a613ad01', '5222', 1, '{"id":"e39d660b-0350-49eb-ae40-5e17a613ad01","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5222","names":[{"code":"th","name":"ค่าส่งเสริมการขาย ของแถม และตัวอย่างสินค้า"},{"code":"en","name":"Sales Promotion, Gifts & Samples"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '239b2ea4-42c4-4b1e-a4a4-684dcd471202', '5231', 1, '{"id":"239b2ea4-42c4-4b1e-a4a4-684dcd471202","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5231","names":[{"code":"th","name":"ค่าขนส่งสินค้าออกและบริการจัดส่งให้ลูกค้า"},{"code":"en","name":"Outward Freight & Shipping to Customers"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '09b42fdd-e679-4008-a677-fce600402952', '5232', 1, '{"id":"09b42fdd-e679-4008-a677-fce600402952","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5232","names":[{"code":"th","name":"ค่าวัสดุหีบห่อและบรรจุภัณฑ์สำหรับการขาย"},{"code":"en","name":"Packaging & Packing Materials"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '0aec3f7e-1f43-412c-a735-f2d30d35a894', '5241', 1, '{"id":"0aec3f7e-1f43-412c-a735-f2d30d35a894","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5241","names":[{"code":"th","name":"ค่าจัดงานแสดงสินค้าและการออกบูธ"},{"code":"en","name":"Exhibition & Trade Fair Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '40233104-f8ad-44a3-ab36-2ef2938d2c90', '5251', 1, '{"id":"40233104-f8ad-44a3-ab36-2ef2938d2c90","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5251","names":[{"code":"th","name":"ค่าน้ำมันและค่าเดินทางพบลูกค้าฝ่ายขาย"},{"code":"en","name":"Sales Fuel & Travel Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'bd54121b-9305-4506-a130-89b0382ac300', '5291', 1, '{"id":"bd54121b-9305-4506-a130-89b0382ac300","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5291","names":[{"code":"th","name":"ค่าใช้จ่ายในการขายอื่น"},{"code":"en","name":"Other Selling Expenses"}],"accounttype":"expense","parentaccountcode":"5200","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'db126e2c-a1b7-4470-a65b-7158cee69b2d', '5300', 1, '{"id":"db126e2c-a1b7-4470-a65b-7158cee69b2d","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5300","names":[{"code":"th","name":"ค่าใช้จ่ายในการบริหาร"},{"code":"en","name":"Administrative Expenses"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '5ea3567f-c18c-4bbf-a15e-73e5c24aaa9b', '5311', 1, '{"id":"5ea3567f-c18c-4bbf-a15e-73e5c24aaa9b","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5311","names":[{"code":"th","name":"เงินเดือน ค่าจ้าง และค่าล่วงเวลาพนักงานสำนักงาน"},{"code":"en","name":"Office Salaries, Wages & Overtime"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a86394f5-7b63-4202-a3f2-90f12f6aba8d', '5312', 1, '{"id":"a86394f5-7b63-4202-a3f2-90f12f6aba8d","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5312","names":[{"code":"th","name":"ค่าตอบแทนและเบี้ยประชุมกรรมการ"},{"code":"en","name":"Directors'' Remuneration & Meeting Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f90119e4-faf5-4141-a9e6-1ce72b01827f', '5313', 1, '{"id":"f90119e4-faf5-4141-a9e6-1ce72b01827f","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5313","names":[{"code":"th","name":"โบนัสพนักงานประจำปี"},{"code":"en","name":"Annual Staff Bonuses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '02629238-bed5-4514-a2fb-662f083cf387', '5314', 1, '{"id":"02629238-bed5-4514-a2fb-662f083cf387","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5314","names":[{"code":"th","name":"เงินสมทบกองทุนประกันสังคม (ส่วนของนายจ้าง)"},{"code":"en","name":"Social Security Fund - Employer''s Contribution"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e579dcc2-41ad-4cf3-ad74-bc7801efa3fb', '5315', 1, '{"id":"e579dcc2-41ad-4cf3-ad74-bc7801efa3fb","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5315","names":[{"code":"th","name":"เงินสมทบกองทุนเงินทดแทนและกองทุนสำรองเลี้ยงชีพ"},{"code":"en","name":"Workmen Compensation & Provident Fund"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c82c105c-35e2-4bea-ac2b-8641ab4e673a', '5316', 1, '{"id":"c82c105c-35e2-4bea-ac2b-8641ab4e673a","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5316","names":[{"code":"th","name":"ค่าสวัสดิการพนักงานและชุดยูนิฟอร์ม"},{"code":"en","name":"Staff Welfare & Uniforms"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '59f173ef-95db-464c-a5e7-1e177107d3f5', '5317', 1, '{"id":"59f173ef-95db-464c-a5e7-1e177107d3f5","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5317","names":[{"code":"th","name":"ค่าฝึกอบรมและสัมมนาบุคลากร"},{"code":"en","name":"Training & Seminar Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '399704df-0929-41f8-a944-94b8319d0600', '5321', 1, '{"id":"399704df-0929-41f8-a944-94b8319d0600","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5321","names":[{"code":"th","name":"ค่าเช่าอาคารสำนักงานและพื้นที่ประกอบการ"},{"code":"en","name":"Office Building Rent"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '72fab943-a552-41a8-a165-0cd5333dd3cd', '5322', 1, '{"id":"72fab943-a552-41a8-a165-0cd5333dd3cd","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5322","names":[{"code":"th","name":"ค่าน้ำประปา"},{"code":"en","name":"Water Utility Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'bf5392ee-33b7-42a5-a37e-983e1cb5e0cf', '5323', 1, '{"id":"bf5392ee-33b7-42a5-a37e-983e1cb5e0cf","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5323","names":[{"code":"th","name":"ค่าไฟฟ้า"},{"code":"en","name":"Electricity Utility Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'cffb4f69-c9c5-4cfc-a417-f9d444cd4a18', '5324', 1, '{"id":"cffb4f69-c9c5-4cfc-a417-f9d444cd4a18","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5324","names":[{"code":"th","name":"ค่าโทรศัพท์และค่าบริการโทรคมนาคม"},{"code":"en","name":"Telephone & Telecom Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'de386e18-3796-47a2-a329-06016caf9bea', '5325', 1, '{"id":"de386e18-3796-47a2-a329-06016caf9bea","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5325","names":[{"code":"th","name":"ค่าบริการอินเทอร์เน็ต ระบบเซิร์ฟเวอร์ และคลาวด์"},{"code":"en","name":"Internet, Server & Cloud Hosting Services"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f11c6459-6676-4f39-a0e6-72d8c64bc1d4', '5331', 1, '{"id":"f11c6459-6676-4f39-a0e6-72d8c64bc1d4","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5331","names":[{"code":"th","name":"ค่าเครื่องเขียน แบบพิมพ์ และวัสดุสำนักงาน"},{"code":"en","name":"Stationery, Printing & Office Supplies"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '0e447024-2faf-47ca-ac00-977aabe4a792', '5332', 1, '{"id":"0e447024-2faf-47ca-ac00-977aabe4a792","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5332","names":[{"code":"th","name":"ค่าไปรษณีย์และค่าส่งเอกสารพัสดุ"},{"code":"en","name":"Postal & Courier Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '46ad8547-7fe4-4c2c-abea-f499c80eb33f', '5341', 1, '{"id":"46ad8547-7fe4-4c2c-abea-f499c80eb33f","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5341","names":[{"code":"th","name":"ค่าตรวจสอบบัญชี (Audit Fees)"},{"code":"en","name":"Audit Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '14a256dd-362f-4cda-a284-64125c441294', '5342', 1, '{"id":"14a256dd-362f-4cda-a284-64125c441294","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5342","names":[{"code":"th","name":"ค่าจัดทำบัญชีและที่ปรึกษาภาษีอากร"},{"code":"en","name":"Accounting & Tax Consultation Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'c14ba1b8-4e2b-402b-a87a-c6f203488437', '5343', 1, '{"id":"c14ba1b8-4e2b-402b-a87a-c6f203488437","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5343","names":[{"code":"th","name":"ค่าธรรมเนียมวิชาชีพกฎหมายและที่ปรึกษาธุรกิจ"},{"code":"en","name":"Legal & Business Advisory Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '474a9562-4684-447a-a9bb-580b5faa7abf', '5351', 1, '{"id":"474a9562-4684-447a-a9bb-580b5faa7abf","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5351","names":[{"code":"th","name":"ค่าเบี้ยประกันภัยทรัพย์สิน อาคาร และยานพาหนะ"},{"code":"en","name":"Property, Building & Vehicle Insurance"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a9ee58fa-ed15-4553-a9dd-a01b6fc27ce8', '5352', 1, '{"id":"a9ee58fa-ed15-4553-a9dd-a01b6fc27ce8","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5352","names":[{"code":"th","name":"ค่าซ่อมแซมและบำรุงรักษาอาคารและอุปกรณ์"},{"code":"en","name":"Repairs & Maintenance Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'babf815c-53c9-415f-a2cf-c27c4f287e92', '5353', 1, '{"id":"babf815c-53c9-415f-a2cf-c27c4f287e92","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5353","names":[{"code":"th","name":"ค่าบริการทำความสะอาดและรักษาความปลอดภัย"},{"code":"en","name":"Cleaning & Security Services"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '2134b4da-7d33-4cba-a5c1-1f964b7725f2', '5361', 1, '{"id":"2134b4da-7d33-4cba-a5c1-1f964b7725f2","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5361","names":[{"code":"th","name":"ค่าเสื่อมราคา - ส่วนปรับปรุงที่ดิน"},{"code":"en","name":"Depreciation - Land Improvements"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '409ba671-7681-4abe-af02-bfa466f1080b', '5362', 1, '{"id":"409ba671-7681-4abe-af02-bfa466f1080b","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5362","names":[{"code":"th","name":"ค่าเสื่อมราคา - อาคารและสิ่งปลูกสร้าง"},{"code":"en","name":"Depreciation - Buildings"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'dc68c239-dc37-4290-a4a7-6216bd959855', '5363', 1, '{"id":"dc68c239-dc37-4290-a4a7-6216bd959855","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5363","names":[{"code":"th","name":"ค่าเสื่อมราคา - ส่วนปรับปรุงอาคาร"},{"code":"en","name":"Depreciation - Building Improvements"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '42c44e89-dd58-483d-a656-c8cf8a771462', '5364', 1, '{"id":"42c44e89-dd58-483d-a656-c8cf8a771462","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5364","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องจักรและอุปกรณ์"},{"code":"en","name":"Depreciation - Machinery & Equipment"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1aba23cd-bfce-45e9-a0db-8f5533ad84c8', '5365', 1, '{"id":"1aba23cd-bfce-45e9-a0db-8f5533ad84c8","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5365","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องตกแต่งและติดตั้ง"},{"code":"en","name":"Depreciation - Furniture & Fixtures"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b7ceab6e-1cd2-4a77-a5be-d2c7eedb67f9', '5366', 1, '{"id":"b7ceab6e-1cd2-4a77-a5be-d2c7eedb67f9","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5366","names":[{"code":"th","name":"ค่าเสื่อมราคา - เครื่องใช้สำนักงาน"},{"code":"en","name":"Depreciation - Office Equipment"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'f8068f7b-1b15-4c82-a674-582a201895c9', '5367', 1, '{"id":"f8068f7b-1b15-4c82-a674-582a201895c9","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5367","names":[{"code":"th","name":"ค่าเสื่อมราคา - คอมพิวเตอร์และอุปกรณ์"},{"code":"en","name":"Depreciation - Computer & Hardware"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'e1fd8e05-53dd-4568-a301-6f162c4b966d', '5368', 1, '{"id":"e1fd8e05-53dd-4568-a301-6f162c4b966d","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5368","names":[{"code":"th","name":"ค่าเสื่อมราคา - ยานพาหนะ"},{"code":"en","name":"Depreciation - Vehicles"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '23a04820-8f2e-45ed-af77-309f1f270453', '5371', 1, '{"id":"23a04820-8f2e-45ed-af77-309f1f270453","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5371","names":[{"code":"th","name":"ค่าตัดจำหน่ายโปรแกรมคอมพิวเตอร์และซอฟต์แวร์"},{"code":"en","name":"Amortization - Software & Applications"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '5d9c8c4e-efda-465d-a714-42cdf553cea8', '5381', 1, '{"id":"5d9c8c4e-efda-465d-a714-42cdf553cea8","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5381","names":[{"code":"th","name":"ค่าธรรมเนียมราชการและใบอนุญาตประกอบกิจการ"},{"code":"en","name":"Government Licenses & Registration Fees"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b6f9571f-e3db-439d-a03d-0d18f1bd1634', '5382', 1, '{"id":"b6f9571f-e3db-439d-a03d-0d18f1bd1634","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5382","names":[{"code":"th","name":"ภาษีที่ดินและสิ่งปลูกสร้าง"},{"code":"en","name":"Land and Building Taxes"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '0ee4b285-dda8-45ec-a3e7-4e5fcab3a39e', '5383', 1, '{"id":"0ee4b285-dda8-45ec-a3e7-4e5fcab3a39e","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5383","names":[{"code":"th","name":"ภาษีป้าย"},{"code":"en","name":"Signboard Tax"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '34c5dc39-faac-4fb6-a574-e297e7d4bc89', '5384', 1, '{"id":"34c5dc39-faac-4fb6-a574-e297e7d4bc89","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5384","names":[{"code":"th","name":"ภาษีซื้อต้องห้าม / ภาษีซื้อที่ไม่สามารถขอคืนได้"},{"code":"en","name":"Non-refundable Input Tax (Disallowed Input VAT)"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '1499dd0d-e87e-4215-a78b-dd08fbd7e797', '5385', 1, '{"id":"1499dd0d-e87e-4215-a78b-dd08fbd7e797","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5385","names":[{"code":"th","name":"เบี้ยปรับ เงินเพิ่ม และค่าปรับทางภาษีอากร"},{"code":"en","name":"Tax Penalties, Fines & Surcharges"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ea110862-bacd-4a0c-a074-d1d318b9c70a', '5391', 1, '{"id":"ea110862-bacd-4a0c-a074-d1d318b9c70a","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5391","names":[{"code":"th","name":"ค่ารับรองและบริการลูกค้า"},{"code":"en","name":"Entertainment & Hospitality Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'a92e3b76-9b62-46c0-a03b-cca9d8f37b99', '5392', 1, '{"id":"a92e3b76-9b62-46c0-a03b-cca9d8f37b99","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5392","names":[{"code":"th","name":"เงินบริจาคเพื่อการกุศลและการศึกษา"},{"code":"en","name":"Charitable & Educational Donations"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '27ade70a-09c9-49e7-a6fb-e764e28a3ef7', '5393', 1, '{"id":"27ade70a-09c9-49e7-a6fb-e764e28a3ef7","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5393","names":[{"code":"th","name":"หนี้สูญและหนี้สงสัยจะสูญ"},{"code":"en","name":"Bad Debts & Doubtful Accounts Expense"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '9a4f34fe-c380-41e2-a696-f3778cf8e8a7', '5394', 1, '{"id":"9a4f34fe-c380-41e2-a696-f3778cf8e8a7","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5394","names":[{"code":"th","name":"ขาดทุนจากอัตราแลกเปลี่ยนเงินตราต่างประเทศ"},{"code":"en","name":"Loss on Foreign Exchange"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'ae0840c6-ef73-4469-a385-dd01952ffb61', '5395', 1, '{"id":"ae0840c6-ef73-4469-a385-dd01952ffb61","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5395","names":[{"code":"th","name":"ขาดทุนจากการจำหน่ายและตัดจำหน่ายทรัพย์สิน"},{"code":"en","name":"Loss on Disposal and Write-off of Assets"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b02114d1-48df-4296-a470-3751bff770e9', '5399', 1, '{"id":"b02114d1-48df-4296-a470-3751bff770e9","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5399","names":[{"code":"th","name":"ค่าใช้จ่ายในการบริหารอื่น"},{"code":"en","name":"Other Administrative Expenses"}],"accounttype":"expense","parentaccountcode":"5300","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', 'b9704591-8608-452f-aed4-0dfc62daa296', '5400', 1, '{"id":"b9704591-8608-452f-aed4-0dfc62daa296","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5400","names":[{"code":"th","name":"ต้นทุนทางการเงินและภาษีเงินได้"},{"code":"en","name":"Financial Costs & Income Tax"}],"accounttype":"expense","parentaccountcode":"5000","normalbalance":"debit","allowposting":false,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":2}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '8a514a4c-a1d3-44f1-a98a-90052551c30f', '5411', 1, '{"id":"8a514a4c-a1d3-44f1-a98a-90052551c30f","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5411","names":[{"code":"th","name":"ดอกเบี้ยจ่ายเงินกู้ยืมสถาบันการเงิน"},{"code":"en","name":"Interest Expense - Bank Loans"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '166dfcda-a9ca-455e-afd6-5652b23a21a3', '5412', 1, '{"id":"166dfcda-a9ca-455e-afd6-5652b23a21a3","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5412","names":[{"code":"th","name":"ดอกเบี้ยจ่ายเงินเบิกเกินบัญชี (O/D)"},{"code":"en","name":"Interest Expense - Bank Overdraft (O/D)"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '34c4ed46-4bdd-428d-af88-f7321e56dc15', '5413', 1, '{"id":"34c4ed46-4bdd-428d-af88-f7321e56dc15","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5413","names":[{"code":"th","name":"ดอกเบี้ยจ่ายตามสัญญาเช่าทางการเงิน"},{"code":"en","name":"Interest Expense - Financial Leases"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '7a74942c-d57e-4372-a7b2-c30c46e13688', '5421', 1, '{"id":"7a74942c-d57e-4372-a7b2-c30c46e13688","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5421","names":[{"code":"th","name":"ค่าธรรมเนียมธนาคารและธุรกรรมทางการเงิน"},{"code":"en","name":"Bank Charges & Transaction Fees"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

INSERT INTO gl_records(company, kind, id, code, version, payload)
VALUES ('c01', 'accounts', '268ab1fa-dcb8-42e9-ad40-be68847c8667', '5431', 1, '{"id":"268ab1fa-dcb8-42e9-ad40-be68847c8667","holdingcode":"demo","businesscode":"c01","version":1,"createdat":"2026-09-20T00:00:00.000Z","createdby":"system","updatedat":"2026-09-20T00:00:00.000Z","updatedby":"system","isdeleted":false,"accountcode":"5431","names":[{"code":"th","name":"ภาษีเงินได้นิติบุคคลประจำงวด"},{"code":"en","name":"Corporate Income Tax Expense"}],"accounttype":"expense","parentaccountcode":"5400","normalbalance":"debit","allowposting":true,"isactive":true,"accountgroup":"BM69-X","iscash":false,"level":3}'::jsonb)
ON CONFLICT (company, kind, id) DO UPDATE SET
  code = EXCLUDED.code,
  version = gl_records.version + 1,
  payload = EXCLUDED.payload;

COMMIT;
