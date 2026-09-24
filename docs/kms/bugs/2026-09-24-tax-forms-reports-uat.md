---
date: 2026-09-24
severity: high
component: [backend, frontend]
tags: [bc-account, postgres, gl, tax, vat, wht]
fixed: true
---

# Symptom

พบระหว่าง UAT รายงานภาษีและแบบยื่นในโมดูล GL (2026-09-24, Demo › rungrueng/01/00000 + integration test บน PostgreSQL จริง) และจากคำถามลุงจืด "ใบกำกับภาษีซื้อ เจ้าหนี้หลายคน แต่เจอเลขที่เดียวกัน ทำยังไง":

1. **รายงานภาษีหัก ณ ที่จ่าย / ใบแนบ ภ.ง.ด. มีแถวผี**
   - ใบกลับรายการ (`kind` reversal) และต้นฉบับที่กลับแล้ว ยังเข้ารายงาน
   - ใบยอดยกมา 1 ม.ค. (`kind` opening) ที่ยกหนี้ภาษีหักค้างจ่ายของ ธ.ค. ปีก่อน กลายเป็นแถวอัตรา 100% ในงวด ม.ค. (UAT S24) ใบปิดบัญชี (`kind` closing) ก็เป็นแบบเดียวกัน
   - ใบโอนยอดระหว่างบัญชีภาษีหัก (Dr ภ.ง.ด.3 ค้างจ่าย / Cr ภ.ง.ด.53 ค้างจ่าย) กลายเป็นแถวฐาน 0 ภาษีใน ภ.ง.ด.53 จึงนับซ้ำ (UAT S25/S26, รีวิว C1)
2. **งวดยื่นผิดเดือน:** รายการภาษีหักที่บันทึกไว้ ลงบัญชี 31 ต.ค. แต่จ่ายจริง 2 พ.ย. เข้างวด ต.ค. ตามวันที่ใบ และแถว/ใบแนบเรียงตามวันที่ใบ (UAT S10/S12)
3. **บันทึกรายการภาษีหักไว้บางราย ยอดในแบบไม่ตรงบัญชี:** ภาษีในบัญชีส่วนที่ยังไม่ได้บันทึกรายละเอียดหายไปเฉย ๆ (รีวิว C2) และแถวประมาณไม่มีแบบยื่น 50 ทวิ จึงเริ่มที่ ภ.ง.ด.53 เสมอ (รีวิว C6)
4. **VAT ไม่ระบุอัตรากลายเป็น 0%:** คำสั่งที่ไม่ส่ง `vat_rate` บันทึกผ่าน และยอดขายเข้าช่องฐานภาษีโดยไม่มีภาษีขาย (UAT V13)
5. **ภาษีซื้อใช้สิทธิผิดงวดได้:** เลือกงวดใช้สิทธิก่อนเดือนที่ออกใบกำกับ หรือหลังเกิน 6 เดือน (หรือพิมพ์ปี พ.ศ. เป็นปีงวด) ก็บันทึกผ่าน error VAT ทุกสาเหตุรวมเป็น `vat_invalid` ตัวเดียว จอจึงชี้ช่องที่ผิดไม่ได้ (รีวิว C7)
6. **ไม่มีตัวจับใบกำกับซ้ำ:** ใบกำกับฉบับเดียวกันบันทึกในหลายใบสำคัญได้โดยไม่มีคำเตือน ถ้าตรวจด้วย "เลขที่ใบกำกับ" อย่างเดียว เจ้าหนี้คนละรายที่บังเอิญใช้เลขเดียวกันจะถูกเตือนผิด
7. **ยอดรวมในแบบยื่นเชื่อค่าจาก browser:** บรรทัดรวมที่ส่งมาตอนบันทึก/พิมพ์ถูกเก็บตามที่ส่ง ถ้าค่าในจอค้างหรือถูกแก้ ยอดรวมในฉบับที่บันทึกและ PDF จะไม่ตรงกับยอดรายการ

## Root Cause

1. `buildWithholdingReport` อ่าน `gl_lines` + `details.withholdings` ของทุกใบโดยไม่แยกชนิดใบ ใบยกมา/ปิดบัญชี/กลับรายการเป็นแค่การยกยอดหรือหักล้าง ไม่ใช่การจ่ายเงินได้ในงวด ส่วนแถวประมาณ (inferred) ใช้ "ยอดฝั่งตรงข้าม" เป็นฐาน ใบโอนยอดที่ฝั่งตรงข้ามมีแต่บัญชีภาษีจึงได้ฐาน 0 แต่ยังนับยอดภาษี
2. รายการที่บันทึกใช้วันที่ใบเป็นงวด ทั้งที่แหล่งทางการกำหนดให้ยื่นตามเดือนที่จ่ายเงินได้:
   - RD-MOF-WHT-EXT ข้อ 2 ยื่น "ภายในเจ็ดวัน นับแต่วันสิ้นเดือนของเดือนที่จ่ายเงินได้พึงประเมิน"
   - ประกาศอธิบดีกรมสรรพากร เกี่ยวกับภาษีเงินได้ (ฉบับที่ 111) ข้อ 3 (RD-P111-2545, https://www.rd.go.th/9041.html) ใช้ถ้อยคำเดียวกัน
   - รายละเอียดอยู่ใน `docs/kms/21-thai-tax-form-references.md` §9
3. โค้ดเดิมตัดสินรายใบว่า "บันทึกแล้วหรือยัง" ถ้าบันทึกไว้บางส่วน แถวประมาณจะถูกข้ามทั้งใบ ส่วนแถวประมาณเองก็ไม่ได้อ่านแบบยื่นจากชื่อบัญชี
4. `cloneJournalDetails` คัดลอกรายละเอียดผ่าน JSON และ `Amount.MarshalJSON` แปลงอัตราว่างเป็น `"0"` การตรวจอัตราที่ทำหลังการคัดลอกจึงไม่เคยเห็นค่าว่าง
5. ไม่มีการเทียบงวดใช้สิทธิกับเดือนที่ออกใบกำกับ ทั้งตอนบันทึกและตอนกระทบยอด (reconcile)
   - RD-VAT-P4: ประกาศอธิบดีฯ VAT ฉบับที่ 4 ข้อ 2 แก้โดยฉบับที่ 76 "ต้องไม่เกินหกเดือนนับแต่เดือนถัดจากเดือนที่ออกใบกำกับภาษี"
   - RD-CODE-82: ม.82/3
6. ไม่มีแนวคิด "ตัวตนของใบกำกับ" ตาม RD-CODE-86-4 (ม.86/4 (2)(4)) ใบกำกับระบุด้วยผู้ออกคู่กับเลขที่ เลขที่ไม่ซ้ำเฉพาะภายในผู้ออกรายเดียว
7. handler บันทึก/พิมพ์แบบยื่นใช้ `Document` จาก request ตรง ๆ โดยไม่คำนวณบรรทัดรวมซ้ำ

## Fix

1. `backend/internal/goapi/handlers/tax_report.go`
   - `whtNonTaxJournalKinds` = reversal/opening/closing (`:453`) อ่านเฉพาะ `status = 'posted'`
   - `inferredWithholdingRows` (`:654`) ไม่สร้างแถวให้ใบที่ฝั่งตรงข้ามมีแต่บัญชีภาษี (ใบโอนยอด)
   - บัญชีภาษีหักหาจาก **ชื่อ** และ `accounttype` เท่านั้น ห้ามใช้รหัสบัญชี
2. `recordedWithholdingRows` (`:589`)
   - งวด = เดือนของ `payment_date` ถ้าวันที่จ่ายว่างหรือผิดรูปแบบ ใช้วันที่ใบแทน
   - กรองใน SQL คำสั่งเดียวทั้งงวด
   - เรียงแถวตามวันที่จ่าย → วันที่ใบ → ใบ ใบแนบ ภ.ง.ด.53 พิมพ์ตามลำดับเดียวกัน
3. `withholdingRemainders` (`:825`)
   - ส่วนที่บัญชีภาษีหักเกินยอดที่บันทึกเป็นแถวประมาณ ฐานไม่ทราบ (ฐาน 0.00 อัตรา/สุทธิว่าง) ยอดรวมจึงตรงบัญชี
   - บันทึกครบแล้วไม่มีแถวซ้ำ แม้แบบที่บันทึกต่างจากแบบของบัญชี
   - `whtAccount.formType` (`:462`) ตั้ง `formtype` จากชื่อบัญชีที่อ้างแบบเดียว ถ้าอ้างหลายแบบหรือไม่อ้างเลย ให้ว่าง
   - frontend ใช้ `formtype` เป็นค่าเริ่มต้นของแบบ 50 ทวิ
4. `requireVatRates` (`backend/internal/generalledger/subledger_vat.go:55`, เรียกจาก `subledger.go:38`) ปฏิเสธรายการที่ไม่มี `vat_rate` **ก่อน** คัดลอก ใช้ทั้งตอนสร้างใบและตอนกระทบยอด (`vat_rate_invalid`, field `vat_rate`)
5. การตรวจงวดใช้สิทธิภาษีซื้อ
   - `checkPurchaseClaimWindow` (`subledger_vat.go:169`) ตรวจเฉพาะ `tax_type` 1 + `claim_status` 1 นับเดือนด้วยเลขจำนวนเต็ม
     - งวดก่อนเดือนที่ออกใบ → `vat_claim_before_invoice_month`
     - เกิน 6 เดือน → `vat_claim_window_exceeded`
     - error ทั้งสองชี้ field `tax_period_month`
   - error VAT อื่นแยก code/field ทีละสาเหตุ (`fieldError`)
   - frontend ใช้ `vatClaimTiming` (`frontend/src/lib/gl-journal-details.ts`) เตือนก่อนบันทึก
     - ข้อความใช้ key เดียวกับ backend (`gl_err_vat_claim_*`)
     - ใช้สิทธิช้า 1–6 เดือน จะแนะนำข้อความ "ถือเป็นภาษีซื้อในเดือนภาษี..." ที่ต้องมีบนใบกำกับ
6. ตรวจใบกำกับซ้ำ `vatInvoiceKeySQL` (`subledger_vat.go:210`) + `VatRecordsForPeriod`
   - key = `tax_type` + ผู้ออก + เลขที่ใบกำกับ + วันที่ใบกำกับ
   - ผู้ออกฝั่งซื้อ = เลขผู้เสียภาษี + สาขา ถ้าไม่มีใช้รหัสคู่ค้า ถ้าไม่มีอีกใช้ชื่อ ฝั่งขาย = สาขาของเราเอง
   - นับเฉพาะใบที่ยังมีผล (ร่าง/ผ่านบัญชี ไม่ลบ ไม่ใช่ใบกลับรายการ) ข้ามทุกงวดใน SQL คำสั่งเดียว
   - ส่งผลเป็น `duplicatedocnos` รายแถว, `duplicatecount` ในสรุป และหมายเหตุ ภ.พ.30 `tax_form_note_duplicate_invoice` (`tax_form_vat.go:40`)
   - **เตือนเท่านั้น ไม่บล็อก** เจ้าหนี้คนละรายใช้เลขเดียวกันจะไม่ถูกเตือน
7. `prepareTaxDocument` (`tax_form.go:275`) ทำตามลำดับ: ตัดช่องว่าง → ตรวจค่า → คำนวณบรรทัดรวมใหม่ด้วยสูตรเดียวกับ `/compute` ใช้ทั้งตอนบันทึกและพิมพ์ PDF ยอดรวมจาก browser ถูกแทนด้วยผลคำนวณเสมอ

## Regression Test

- `backend/internal/goapi/handlers/tax_withholding_uat_integration_test.go` `TestWithholdingReportUATRegressions2026_09_24` ครอบคลุม:
  - S24 ใบยกมา
  - S25/S26 ใบโอนยอด
  - S14 บัญชีรายได้ชื่อมีคำว่า VAT
  - S10/S12 เดือนตามวันที่จ่าย + ลำดับใบแนบ
- `backend/internal/goapi/handlers/tax_report_review_integration_test.go`:
  - `TestWithholdingReportReviewFindings2026_09_24` (C1/C2/C6)
  - `TestVatDuplicateInvoiceReachesRegisterAndPP30`: ผู้ขายอีกรายใช้เลขเดียวกันต้องไม่ถูกเตือน
- `backend/internal/goapi/handlers/tax_withholding_recorded_integration_test.go`: ต้นฉบับที่กลับรายการแล้วไม่เข้ารายงาน
- `backend/internal/goapi/handlers/tax_report_test.go`: `TestWithholdingRemainders`, `TestBuildVatRegisterDuplicateWarnings`, `TestWithholdingNetBlankWhenBaseUnknown`, `TestTaxFormComputeReturnsRecomputedTotals`
- `backend/internal/goapi/handlers/tax_form_test.go`: `TestPrepareTaxDocumentRecomputesTotals`
- `backend/internal/generalledger/tax_uat_integration_test.go`: `TestPostgresVatRateRequiredOnCreateAndReconcile` (V13)
- `backend/internal/generalledger/subledger_vat_duplicate_integration_test.go`: `TestPostgresVatDuplicateInvoiceWarnings`, `TestPostgresPurchaseClaimWindowOnSaveAndReconcile`
- `backend/internal/generalledger/subledger_tax_errors_test.go`: `TestPurchaseVatClaimWindow` (รวมกรณีพิมพ์ปี พ.ศ.), `TestCloneJournalDetailsRejectsMissingVatRate`, `TestSubledgerVatReportsFieldOfEachError`
- frontend:
  - `frontend/src/lib/gl-journal-details.test.ts` (`vatClaimTiming` รวมข้ามปี)
  - `frontend/src/lib/thai-tax.test.ts` (`duplicatedocnos`/`duplicatecount`, `formtype` → แบบ 50 ทวิ)
  - `frontend/src/lib/tax-forms.test.ts` (ยอดจากบัญชีเปลี่ยน / ยืนยันก่อนดึงยอดทับ)
  - `frontend/src/app/gl/gl-language-keys.test.ts` (key ที่จอ GL ใช้ต้องมีใน `languages.tsv`)

## ยังเปิดอยู่

- ภาษีขาย (`tax_type` 2) ยังไม่ตรวจกรอบเวลา เพราะยังไม่ได้ตรวจแหล่งทางการ
- ใบเพิ่มหนี้/ใบลดหนี้ฝั่งซื้อใช้กรอบ 6 เดือนเดียวกับใบกำกับ รอลุงจืดยืนยัน
- ระบบไม่มีช่องเลขเล่มใบกำกับ ผู้ออกรายเดียวใช้เลขเดียวกันต่างเล่มในวันเดียวกันจึงถูกเตือนเกิน (เตือนเท่านั้น)
- ใบกำกับฉบับใหม่ที่ออกแทนฉบับเดิม (ป.86/2542 ข้อ 25) จับไม่ได้ ถ้าบันทึกทั้งสองฉบับ
- แหล่งทางการของทุกข้อ: `docs/kms/21-thai-tax-form-references.md` §1, §9, §10
