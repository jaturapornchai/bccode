---
date: 2026-09-23
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, general-ledger, wht, tax, postgres]
---

# ฐานภาษีหัก ณ ที่จ่ายแก้ได้เสมอ — เก็บเป็นรายละเอียดใบสำคัญ (details.withholdings) ไม่ใช่ค่าประมาณจากบรรทัดบัญชี

## Context

- ลุงจืดสั่ง 2026-09-23: "ตรวจข้อมูลใน pgsql ด้วยทุกครั้ง ต้องระวังเรื่องฐานภาษีด้วย ต้องเปลี่ยนแปลงได้เสมอ"
- เดิมรายงาน ภ.ง.ด.3/53 และ 50 ทวิ **ประมาณ** ฐานภาษีจากบรรทัดบัญชีฝั่งตรงข้ามของบัญชีภาษีหัก ณ ที่จ่าย — ใบที่จ่ายรวม VAT (107,000) ได้ฐานผิด (ควรเป็นก่อน VAT 100,000) และผู้ใช้แก้ไม่ได้
- Champ: `BCAPWTaxList.BaseOfTax` / `WTaxRate` เป็นคอลัมน์ที่ผู้ใช้กรอกเอง (`D:/project-champ/champ/champ/Script/SQLSERVER_Script.sql:8908`) — ภาษี = ฐาน × อัตรา แก้ได้
- `mydocs/datamodels/gl/wht.sql` (สเปกลุงจืด) กำหนด `wht_records` + `wht_record_lines` (base_amount, tax_amount, wht_rate, income_tax_type, condition_type, form_type) แต่โค้ด GL ปัจจุบันเก็บใบสำคัญเป็น JSON ใน `gl_records.payload` (รายละเอียดประกอบอื่นอยู่ใน `details` เหมือนกัน)

## Decision

1. เก็บภาษีหัก ณ ที่จ่ายเป็น `details.withholdings[]` ของใบสำคัญ (ชื่อฟิลด์ตาม wht.sql: `wht_direction`, `form_type` PND2/PND3/PND53 เท่านั้น — ไม่มี PND1 ตามกฎไม่ทำเงินเดือน, `partner_code`, `payment_date`, `income_tax_type`, `condition_type` 1–3, `wht_rate`, `base_amount`, `tax_amount`, `wht_cert_no`) — ตรวจใน `sql.Tx` เดียวกับการบันทึกใบสำคัญ (`backend/internal/generalledger/subledger_withholding.go`)
2. `tax_amount` ว่าง = backend คำนวณ ฐาน × อัตรา ปัดตาม scale ปีบัญชี; ผู้ใช้พิมพ์เองได้ (ตามเอกสารจริง) แต่ต้อง 0 ≤ ภาษี ≤ ฐาน, อัตรา 0–100, คู่ค้าต้องมีและใช้งานอยู่
3. **แก้ได้เสมอ**: ฉบับร่างแก้ผ่าน update ปกติ; ผ่านบัญชีแล้วแก้ผ่าน reconcile (ยอด GL ไม่เปลี่ยน — เป็นหลักฐานประกอบ) โดย **แทนทั้งชุด** + บังคับเหตุผล + เขียน `gl_subledger_audit` action `withholding_replace` (before/after/reason; ตารางนี้ append-only ด้วย trigger)
4. รายงาน ภ.ง.ด. (`buildWithholdingReport`) ใช้รายการที่บันทึกก่อน (`taxbasesource: "recorded"`, 1 แถวต่อรายการ, กรองตามทิศทาง+แบบ); ใบที่ไม่ได้บันทึกยังประมาณจากบรรทัดบัญชีและติดป้าย `"inferred"` ให้ผู้ใช้ตรวจ
5. จอบันทึกรายวันโหลดใบสำคัญจาก backend หลังบันทึก เพื่อให้เห็นภาษีที่ backend คำนวณทันที

## Alternatives

- **ทำตาราง `wht_records`/`wht_record_lines` ตาม wht.sql ทันที** — ตรงสเปกเชิงโครงสร้าง แต่ GL ทั้งชุดยังเก็บใบสำคัญ/รายละเอียดเป็น JSON (`gl_records`) การแยกตารางเฉพาะ WHT ตอนนี้ต้องทำ dual path กับ journal payload — เลื่อนไปทำพร้อมการย้าย GL ไป schema ของ `mydocs/datamodels/gl/` ทั้งชุด (ชื่อฟิลด์ใช้ชุดเดียวกันแล้ว ย้ายได้ตรง)
- **ให้แก้ฐานในจอ 50 ทวิ อย่างเดียว** — ตัวเลขในหนังสือรับรองกับรายงาน ภ.ง.ด. จะไม่ตรงกัน (ไม่มีที่เก็บ) จึงไม่เลือก
- **ห้ามแก้หลังผ่านบัญชี** — ขัดคำสั่ง "เปลี่ยนแปลงได้เสมอ" และ Champ ก็แก้ได้ก่อนปิดงวด; ใช้ audit แทนการล็อก

## Consequences

- ✅ ฐานภาษีเป็นข้อมูลจริงที่ตรวจสอบย้อนหลังได้ (audit + เหตุผล), รายงานกับ 50 ทวิ ใช้ตัวเลขชุดเดียวกัน
- ⚠️ ยังไม่ตรวจว่า `tax_amount` ในรายละเอียดตรงกับยอดบรรทัดบัญชีภาษีหัก ณ ที่จ่ายของใบเดียวกัน (UAT พบ 935 vs 1,000 บันทึกได้)
- ⚠️ ชื่อ/ที่อยู่ผู้ถูกหักอ่านจากทะเบียนคู่ค้าปัจจุบัน ไม่ได้ snapshot ตอนออกหนังสือรับรอง
- ⚠️ หลังผ่านบัญชีลบรายการภาษีจนหมดไม่ได้ (reconcile ต้องมีอย่างน้อย 1 รายการ)
- ⚠️ ภาษีมูลค่าเพิ่ม (vat.sql) ยังไม่ได้ทำแบบเดียวกัน — รายงาน VAT ยังเปิดอยู่ใน `docs/kms/bugs/2026-09-23-vat-report-reads-missing-erp-tables.md`

## Evidence

- `TestPostgresWithholdingBaseEditableAlways` (PG จริง): สร้าง → แก้ร่าง → ค่าผิด 5 แบบถูกปฏิเสธ → ผ่านบัญชี → แก้ฐานไม่มีเหตุผลถูกปฏิเสธ → แก้พร้อมเหตุผล → audit before/after/reason, `gl_lines` ไม่เปลี่ยน
- `TestWithholdingReportUsesRecordedBase`: แถว recorded ใช้ฐานที่บันทึก, แถว inferred ยังประมาณ, ภ.ง.ด.3 ไม่เห็นรายการ ภ.ง.ด.53
- UAT จอจริง (Demo, local 2026-09-23): ฐาน 100,000 × 1% → PG 1000; แก้ 93,457.94 → 934.58; พิมพ์เอง 935; ผ่านบัญชี; แก้ฐาน 100,000 → PG version 5 posted, audit `withholding_replace`; 50 ทวิ prefill recorded + PDF 2 หน้า
