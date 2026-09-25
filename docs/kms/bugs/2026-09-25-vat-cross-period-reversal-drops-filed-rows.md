---
date: 2026-09-25
severity: high
component: [backend, frontend, db]
tags: [bc-account, postgres, vat, withholding-tax]
fixed: true
---

# Symptom

1. ใบสำคัญขาย ม.ค. ที่ยื่น ภ.พ.30 ม.ค. ไปแล้ว ถูกกลับรายการในเดือน ก.พ. → เปิดรายงานภาษีขาย/ภ.พ.30 ของเดือน ม.ค. ใหม่ **ยอดหายเงียบ ๆ** ไม่ตรงกับแบบที่ยื่นไปแล้ว และไม่มีที่ไหนบอกว่ามีรายการที่ยื่นแล้วถูกยกเลิก (Champ มีรายงานภาษีซื้อ/ขายที่ยกเลิกข้ามงวด 5522/5523)
2. เครดิตภาษีถูกหัก ณ ที่จ่ายใน ภ.ง.ด.50/51 ลดลงย้อนหลังแบบเดียวกัน และนับใบยอดยกมา/ปิดบัญชีที่มีรายการภาษีหักเป็นเครดิตซ้ำ
3. รายงาน ภ.ง.ด.3/53 ยังแสดงแถวที่กลับรายการเดือนหลัง (ถูกต้อง) แต่กดพิมพ์ 50 ทวิ ของแถวนั้นได้ "ไม่พบ"

## Root Cause

- `VatRecordsForPeriod` (`backend/internal/generalledger/subledger_vat.go`) อ่านเฉพาะ `status = 'posted'` — ต้นฉบับที่กลับรายการกลายเป็น `reversed` จึงหลุดจากทุกงวด ไม่ว่ากลับรายการเมื่อไร
- `WithheldFromCompanyTotal` / `RecordedWithholding` (`subledger_withholding.go`) ก็อ่านเฉพาะ `posted` และไม่ตัด `kind` = `reversal`/`opening`/`closing` — ขณะที่รายงานภาษีหัก (`tax_report.go` `whtReversalMonthSQL`) ใช้กติกา "กลับเดือนหลัง = ยังอยู่ในเดือนเดิม" มาตั้งแต่ 2026-09-24 → สามที่ใช้กติกาไม่ตรงกัน

## Fix

- VAT: CTE `live` — ต้นฉบับ `reversed` ที่เดือนของใบกลับรายการ > งวดภาษีของรายการ ยังอยู่ในงวดเดิม พร้อม `reversaldocno`/`reversaldate`; กลับในงวดเดียวกัน = ไม่แสดง (แบบ Champ `CancelOutPeriod`, `GLRepInputTaxCHMView.cpp`)
- มุมมองใหม่ `view: "reversed_later"` (`VatCrossPeriodCancellations`) = รายการที่ยื่นในงวดก่อนแล้วกลับรายการในเดือนที่เลือก — แท็บ "ยกเลิกข้ามงวด" ในจอรายงานภาษีซื้อ/ขาย (ยุบ Champ 5522/5523 ไม่เพิ่มเมนู) แสดงเพื่อตรวจเท่านั้น ไม่หักจาก ภ.พ.30
- หมายเหตุตามภาษา: แถว (`tax_wht_row_reversed_later`), รายงาน (`tax_vat_note_reversed_later`, `tax_vat_note_cross_period_view`), ภ.พ.30 (`tax_form_note_vat_reversed_later`, `tax_form_note_vat_cross_period_cancel`) บอกทางแก้ตาม ม.83/4 (ยื่นเพิ่มเติม) หรือใบลดหนี้ ม.82/10 — ตรวจถ้อยคำจาก rd.go.th 5206/5207 เอง (ทะเบียน `docs/kms/21` §1 RD-CODE-83-4, §10)
- WHT: `generalledger.ReversalMonthSQL` ตัวเดียวใช้ทั้งรายงาน เครดิต ภ.ง.ด.50/51 (นับแยก `ReversedLater` + หมายเหตุ `tax_form_note_cit_wht_reversed_later`) และ 50 ทวิ; เครดิตตัดใบกลับรายการ/ยอดยกมา/ปิดบัญชี
- แก้อ้างอิงผิด "ม.82/3 วรรคสอง" → **วรรคสี่** (วรรคหนึ่ง–สามคือวิธีคำนวณ/ชำระ/เครดิต) ใน code, test, ทะเบียน และ skill

## Regression Test

- `TestPostgresVatReversedAfterTaxPeriod` — กลับหลังงวด / ในงวด / ข้ามปี ธ.ค.→ม.ค. / ภาษีซื้อเฉพาะใช้สิทธิ / ใบกำกับที่บันทึกใหม่เตือนซ้ำกับต้นฉบับ (mutation `>`→`>=`, `<`→`<=`, ตัดกรอง `claim_status` ถูกจับครบ)
- `TestPostgresWithholdingReversedAfterPeriod` — เครดิตครึ่งปี/ทั้งปี, ใบยอดยกมา, 50 ทวิ กลับเดือนหลัง/เดือนเดียวกัน (mutation 3 จุดถูกจับ)
- `TestVatReversedAfterPeriodFormsAndView` (handlers) — ภ.พ.30 ส.ค. คงยอด + หมายเหตุ, ภ.พ.30 ก.ย. ไม่หักแต่เตือน, หมายเหตุแถวยกเลิกข้ามงวด
- `TestVatRegisterReversedLaterNotes` + กรณี `INVALID_VIEW` ใน `TestTaxReportErrorsFollowLanguage`
