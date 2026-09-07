# Product read model parity + vattype→taxtype rename

**Date:** 2026-06-10
**Status:** Accepted (ลุงจืด confirm)
**Tags:** #bc-account #postgres #mongodb #clickhouse #schema

## Decision
1. **`vattype` → `taxtype` rename ใน model ใหม่** — pgsql/clickhouse `taxtype` (0=ไม่มี/1=รวม/2=แยก/3=ยกเว้น) map จาก mongo `vattype`; `taxtype` เดิมใน source ("ภาษีอื่น") ไม่ถูกใช้ใน model ใหม่
2. **MongoDB `products` มี field ครบ 1:1 ตาม pgsql read model** — รวม `balanceqty`/`pendingrecvqty`/`pendingsendqty` (Decimal128) เป็น **write-back cache ที่ stock engine เขียนเท่านั้น** user/API Create/Edit ห้ามแตะ; ค่าแสดงผลอ่านจาก read model
3. **pgsql เพิ่ม** `guidfixed` (UNIQUE, CRUD identity เท่านั้น ไม่ใช่ relation key), `updatedat timestamptz` (UTC+0 freshness), `pendingrecvqty`/`pendingsendqty` `numeric(38,8) NOT NULL DEFAULT 0`; `balanceqty` เปลี่ยนเป็น NOT NULL
4. **ตัด index low-cardinality 5 ตัว** (`taxtype`/`taxrate`/`costtype`/`unittype`/`allcompanies`) — planner ไม่ใช้ + ลด write amplification ตอน Kafka sync
5. SELECT ตัวอย่างใช้ `jsonb_array_elements ... WITH ORDINALITY` + `ord` tie-breaker การันตี fallback "ชื่อแรกที่มีค่า"

## Risk noted (dissent บันทึกไว้)
stock fields ใน mongo เป็น duplicate ของ pgsql อาจไม่ตรงกันชั่วขณะระหว่าง sync — mitigate ด้วย single-writer (stock engine) + doc ระบุให้อ่านจาก read model

## Files
- `D:\bccode-model\pgsql\product-schema.html` (+ mirrors `public/pgsql`, `public/scripts/pgsql`)
- `D:\bccode-model\mongodb\product-schema.html`, `clickhouse\product-schema.html` (+ mirrors)
- `styles/schema-doc.css` (เพิ่ม `.null-no`/`.null-yes`/`.src-path`)
- Regression: `tests/schema-language-doc.test.mjs` (test ใหม่ parity), `tests/ui-modernization.test.mjs` (อัปเดต assertion เก่า)
- Tests: 51/51 pass; DDL ยังไม่ได้รันบน PostgreSQL จริง (ไม่มี psql ในเครื่อง)
