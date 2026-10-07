---
date: 2026-10-07
severity: high
component: [backend, frontend]
tags: [bc-account, gl, tax, security, audit]
fixed: true
---

# Symptom

ตรวจ GL 2026-10-05 → แก้รอบ 2026-10-07 (กลุ่ม A ที่ลุงจืดสั่ง "แก้ bug ก่อน") 4 จุด:

1. client สร้างใบสำคัญ `kind` = `closing` / `reversal` / `mapping` ได้เอง → ยอดหายจาก P&L / VAT / WHT
2. API รายงานภาษี (`vat-register`, `vat-summary`, `wht`) และแบบยื่น (`/api/report/tax/form/*`) ตรวจแค่ว่าบริษัทใน token ตรงกับที่ขอ — ไม่ตรวจสมาชิกภาพ/วันหมดอายุ/`access_scopes`
3. ลบแบบยื่นภาษีที่บันทึกแล้วไม่มีร่องรอยว่าใครลบ เมื่อไหร่
4. แถวใหม่ของ "ภาษีหัก ณ ที่จ่าย" ในรายละเอียดใบสำคัญเติมอัตรา 3% ให้เอง ทั้งที่อัตราจริงขึ้นกับประเภทเงินได้ (1 / 1.5 / 3 / 5) — ผู้ใช้เปลี่ยนประเภทแล้วอัตราค้างผิดโดยไม่รู้

## Root Cause

- 1: `mutateJournal` action create ไม่จำกัด `kind`; `Journal.Validate` รับครบ 5 kind (ต้องรับเพื่อตรวจใบเดิมตอน post) จึงไม่มีด่านฝั่งสร้าง
- 2: handler เรียกแค่ `authenticatedCompanyContext` (อ่าน holding/company จาก token); ชุดตรวจสิทธิ์ที่ใช้อยู่แล้วกับหนังสือรับรอง 50 ทวิ (`whtCallerAccess`) ไม่ได้ใช้กับรายงานภาษีอื่น
- 3: `deleteTaxFiling` ลบแถวใน `tax_filings` อย่างเดียว; `tax_filing_history` เขียนเฉพาะตอน save
- 4: ค่า default ของแถวใหม่ (`newRow`, `gl-journal-details.tsx`) ใส่ `wht_rate:"3"`; ทั้งที่ validation (`gl-journal-details.ts`) ออกแบบให้อัตราว่าง = ต้องกรอก

## Fix

- `backend/internal/generalledger/journal_books.go` + `postgres_store.go`: `errJournalKindNotAllowed` (`journal_kind_not_allowed`) ปฏิเสธการ create kind closing/reversal/mapping; `opening`/`manual` ยังสร้างได้; ใบปิดบัญชี/กลับรายการที่ระบบสร้างเองไม่ผ่านเส้นทางนี้
- `backend/internal/goapi/handlers/tax_report.go`: `taxCallerAllowed` ใช้ `whtCallerAccess` ชุดเดียวกับ 50 ทวิ (สมาชิกภาพ active, ไม่หมดอายุ, บริษัท/สาขาใน token อยู่ใน `access_scopes`) → 403 `FORBIDDEN` / `ACCESS_EXPIRED`; เรียกใน `TaxVatRegisterHandler`, `TaxWithholdingHandler` (`tax_report.go`), `TaxVatSummaryHandler` (`tax_vat_summary.go`) หลังตรวจ input และใน `beginTaxForm` (`tax_form.go`) ซึ่งครอบทุก endpoint แบบยื่น
- `tax_form.go`: `deleteTaxFiling(ctx, db, company, user, id, version)` ลบ + เขียนแถว `tax_filing_history` (`action='delete'`, `saved_by` = ผู้ลบ, เอกสารรุ่นสุดท้าย) ใน transaction เดียว (ไม่ใช้ trigger); เพิ่มคอลัมน์ `action TEXT NOT NULL DEFAULT 'save'` แบบ forward-only (`ADD COLUMN IF NOT EXISTS`)
- `frontend/src/app/gl/gl-journal-details.tsx`: `wht_rate` ของแถวใหม่ = `""` (ผู้ใช้ต้องกรอก; จอแสดงข้อความเตือนและบันทึกไม่ผ่านจนกว่าจะกรอก)
- `backend/assets/language/languages.tsv`: เพิ่ม `gl_err_journal_kind_not_allowed`

**ยังไม่แก้ (รอลุงจืดตัดสิน):** A2 trigger append-only 3 ตัว (`gl_events`, `gl_journal_review_events`, `gl_subledger_audit`) ขัด `mydocs/specs/rules.md` แต่เป็นตัวกันแก้ประวัติ; A5 ผลตรวจรีเซ็ตเมื่อ post/reverse (เอกสาร `architecture/gl-journal-review.md` ระบุไว้เป็นข้อจำกัดโดยตั้งใจ); A6 ใบ opening กับ P&L (ต้องเทียบ Champ/ตัดสินกรณีเริ่มใช้ระบบกลางปี); การกรองรายงานภาษีตามสาขาที่มีสิทธิ์ (กระทบตัวเลข ภ.พ.30); ปุ่ม "+ ภาษี 7%" ใน `gl-smart-guard.ts` ที่ฝัง 7% (ไม่มีใน Champ)

## Regression Test

- `generalledger/journal_kind_create_integration_test.go` `TestPostgresCreateRejectsSystemOnlyJournalKinds` (พังก่อนแก้: create kind=closing ได้)
- `goapi/handlers/tax_report_test.go` `TestTaxEndpointsRejectCallerWithoutAccess` (5 endpoint × 3 กรณี: หมดอายุ / บริษัทนอกสิทธิ์ / ไม่มีสมาชิกภาพ)
- `goapi/handlers/tax_form_integration_test.go` `TestTaxFilingStorage` (ลบที่ล้มเหลวไม่ทิ้ง audit; ลบสำเร็จมีแถว `action='delete'` พร้อมผู้ลบ/รุ่น/เอกสาร)
- verify: `go vet`, `go test ./internal/goapi/...`, `go test -tags integration` (generalledger + goapi, Postgres ชั่วคราว), vitest `gl-journal-details` 44 ผ่าน, `tsc --noEmit` สะอาด
