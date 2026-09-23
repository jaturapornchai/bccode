---
date: 2026-09-23
severity: high
component: [backend, db]
tags: [bc-account, postgres, tax, vat]
fixed: true
---

# Symptom

รายงานภาษีขาย/ภาษีซื้อ (`POST /api/report/tax/vat-register`) และสรุป ภ.พ.30 (`POST /api/report/tax/pp30-summary`) บน production ตอบ 500 `QUERY_ERROR` ทุก holding — จอภาษีมูลค่าเพิ่มจึงขึ้น "โหลดข้อมูลไม่สำเร็จ" มาตลอด

## Root Cause

1. handler ต่อฐาน holding (`mypg.PgSqlFastConnect(holdingCode)` — `backend/internal/goapi/handlers/tax_report.go`) แต่ query อ่าน `public.saleinvoicetransaction` / `public.purchasetransaction` / `public.debtor` / `public.creditor` ซึ่ง **ไม่มีในฐาน holding ใดเลย** — ตรวจ prod 2026-09-23: มีเฉพาะใน `bcai_projection` (ของค้างยุค Kafka projection) และมี 0 แถว
2. query รายการใช้คอลัมน์ `m.name0` แต่ DDL จริงของ `debtor`/`creditor` มี `names` ไม่มี `name0`; ฝั่งขายยัง join `debtor` ด้วย `t.creditorcode`
3. ตาราง ERP มี `holdingcode` แต่ query ไม่กรองบริษัท (company) — ถ้าวันหนึ่งตารางมีข้อมูล รายงานจะรวมทุกบริษัท

### Fix

แก้แล้ว 2026-09-23 — อ่านภาษีมูลค่าเพิ่มจากรายละเอียดใบสำคัญ GL แทนตารางเอกสาร ERP (ผู้ใช้ GL อย่างเดียวใช้ได้เลย):

1. ใบสำคัญ GL มีรายละเอียด `details.vats[]` (`SubledgerVat` ใน `backend/internal/generalledger/subledger_vat.go`, ชื่อฟิลด์ตาม `mydocs/datamodels/gl/vat.sql`) — ผู้ใช้กรอกฐานภาษี/ภาษี/งวดภาษีเอง แก้ได้เสมอแม้ผ่านบัญชีแล้ว (reconcile พร้อมเหตุผล → audit `vat_replace`); ตรวจตาม CHECK ของ vat.sql ภายใน `sql.Tx` เดียวกับการบันทึก
2. `generalledger.VatRecordsForPeriod(ctx, db, company, year, month, taxType)` อ่านเฉพาะใบสำคัญ **posted** ที่ไม่ถูกลบ ของบริษัทที่ผู้ใช้เลือก (`company = $1`) ตามงวดภาษี; ภาษีซื้อนับเฉพาะ `claim_status = 1` (ใช้สิทธิในงวดนี้) — ใบร่าง/กลับรายการ/บริษัทอื่นไม่นับ
3. `TaxVatRegisterHandler` / `PP30SummaryHandler` (`backend/internal/goapi/handlers/tax_report.go`) เรียก `generalledger.EnsureSchema` แล้วสรุปด้วย `buildVatRegister` / `sumPP30` (decimal ปัด 2 ตำแหน่งต่อแถว; ใบลดหนี้ document_type 3 หักออก, ใบเพิ่มหนี้บวกเพิ่ม); JSON contract เดิมไม่เปลี่ยน และยังใช้ `computeVatSettlement` เดิม — query builder ตาราง ERP (`buildVatRegisterQuery`, `buildPP30SalesQuery` ฯลฯ) และ `moneySQL` ถูกลบ (ต่อมาวันเดียวกัน `PP30SummaryHandler`/`computeVatSettlement` และ route `pp30-summary` ถูกลบ — แบบ ภ.พ.30 ย้ายไป `/api/report/tax/form/*` ที่ใช้ `sumPP30` ตัวเดิม, ADR `decisions/2026-09-23-rd-tax-forms-engine.md`)
4. เมนู ภาษีซื้อ/ภาษีขาย/ภ.พ.30 กลับเป็นใช้งานได้ (`LIVE_TAX_FORMS` ใน `frontend/src/lib/menu-screen-status.ts`); ภ.พ.36 ยังรอเชื่อมข้อมูล
5. จอสมุดรายวันมี section "ภาษีมูลค่าเพิ่ม (ใบกำกับภาษี)" (`frontend/src/app/gl/gl-journal-details.tsx`) ให้กรอกรายการใบกำกับ — แบบแผน UI ใน skill `ui-scale-polish` §8.32

## Regression Test

- `TestVatReportsReadRecordedVat` (`backend/internal/goapi/handlers/tax_vat_integration_test.go`, tag `integration`) — seed ใบสำคัญ posted/draft/reversed + ใบลดหนี้ + ภาษีซื้อต้องห้าม/ต่างงวด แล้วยิง handler จริง: ทะเบียนภาษีขาย/ซื้อ ยอดรวม และ ภ.พ.30 (ยกมา 400) ต้องตรงตัวเลข; บริษัทอื่นต้องได้ว่าง
- `TestPostgresVatRecordsForPeriod` + `TestPostgresVatBaseEditableAlways` (`backend/internal/generalledger/subledger_vat_integration_test.go`) — กรองงวด/ประเภท/สถานะ/บริษัท และฐานภาษีแก้ได้ทั้งก่อน-หลังผ่านบัญชี โดยยอด GL ไม่เปลี่ยน + มี audit ค่าเดิม
- `TestBuildVatRegisterSignsAndTotals`, `TestSumPP30FromRecordedVat` (`backend/internal/goapi/handlers/tax_report_test.go`) — เครื่องหมายใบลดหนี้และยอดรวม decimal
- `TestTaxVatQueriesRunOnRealSchema` ถูกลบ เพราะไม่มี query ตาราง ERP แล้ว
- หลัง deploy: smoke test บน prod ต้องได้ 200 + `summary` จริงจาก holding `rungrueng` ไม่ใช่แค่ดู HTTP status
