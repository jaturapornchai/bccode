---
date: 2026-09-25
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, general-ledger, wht, tax, postgres]
---

# รายงานภาษีหัก ณ ที่จ่าย / ภ.ง.ด.2-3-53 / 50 ทวิ อ่านเฉพาะรายการที่บันทึกในใบสำคัญ — เลิกเดาจากชื่อบัญชี

## Context

- ตั้งแต่ ADR `2026-09-23-wht-tax-base-editable.md` รายงานใช้รายการที่บันทึก (`details.withholdings`) แต่ใบที่ไม่ได้บันทึกยัง **ประมาณ** แถวจากบรรทัดบัญชี: หาบัญชีภาษีหักจากคำในชื่อบัญชี ("หัก ณ ที่จ่าย", "ภ.ง.ด.53" ฯลฯ) แล้วเดาฐาน อัตรา และแบบยื่น
- การเดานี้ผิดซ้ำหลายรอบ (UAT/review 2026-09-24: S14 ฐาน 0 สุทธิติดลบ, S25/S26 ใบโอนยอดระหว่างบัญชีภาษีนับซ้ำ, C1 ใบโอนยอดเป็นแถวอัตรา 100%, S24 ใบยกมาเป็นแถวใน ม.ค., บัญชีที่ชื่ออ้างหลายแบบ) — ต้องมีกติกาพิเศษเพิ่มเรื่อย ๆ (~600 บรรทัดใน `tax_report.go`) และยังขัดกฎ "ห้ามใช้ชื่อ/รหัสบัญชีเป็นเงื่อนไข" ([[no-account-code-conditions]])
- Champ อ่านรายงานและใบแนบภาษีหักจากตารางรายการภาษีหัก `BCAPWTaxList` (ฐาน `BaseOfTax`, อัตรา `WTaxRate` ที่ผู้ใช้กรอก — `D:/project-champ/champ/champ/Script/SQLSERVER_Script.sql:8908`) ไม่ได้อ่านจากบัญชี
- แบบยื่นต้องมีข้อมูลรายผู้รับเงิน (เลขผู้เสียภาษี ประเภทเงินได้ ฐาน อัตรา วันที่จ่าย — RD-MOF-WHT-EXT, RD-50TAWI-FORM ใน `docs/kms/21-thai-tax-form-references.md` §1) ซึ่งบรรทัดบัญชีไม่มี

## Decision

1. `buildWithholdingReport` (`backend/internal/goapi/handlers/tax_report.go`) คืนเฉพาะแถวจาก `details.withholdings` — ลบ `inferredWithholdingRows`, `withholdingRemainders`, `findWithholdingAccounts`, `nonTaxCounterTotals` และคำค้นชื่อบัญชีทั้งหมด; ฟิลด์ `taxbasesource` / `unknownform` ถูกลบจาก API
2. แบบที่รับ = 2/3/53 (ตรงกับแบบที่บันทึกได้ `generalledger.withholdingForms`) — ภ.ง.ด.54 ไม่มีแถวจนกว่าจะตรวจมาตรา 70 และเปิดให้บันทึก
3. งวดที่ไม่มีรายการที่บันทึก → หมายเหตุ `tax_wht_note_no_records` บอกให้บันทึกที่รายละเอียดใบสำคัญ หมวดภาษีหัก ณ ที่จ่าย
4. จอ 50 ทวิ: ทุกแถวเป็นรายการที่บันทึก (เลิกสถานะ/ข้อความ "ประมาณ"); กติกากลับรายการหลังงวด (`generalledger.ReversalMonthSQL`), เครดิต ภ.ง.ด.50/51 และ 50 ทวิ ใช้รายการที่บันทึกชุดเดียวกัน
5. ตัด language key ที่ตาย 6 ตัว (`tax_wht_note_no_accounts`, `tax_wht_note_form_unknown`, `tax_form_note_wht_form_unknown`, `tax_form_note_inferred_base`, `wht_cert_ui_base_inferred`, `wht_cert_ui_record_inferred`)

## Alternatives

- **คงการเดาไว้แต่ติดป้าย** — ยอดที่เดาผิดยังเข้าแบบยื่นได้ถ้าผู้ใช้ไม่สังเกต และต้องเพิ่มกติกาทุกครั้งที่ผังบัญชีต่างไป (ผังของแต่ละกิจการตั้งชื่อไม่เหมือนกัน) — ไม่เลือก
- **เตือนเมื่อยอดบัญชีภาษีหักไม่ตรงกับรายการที่บันทึก** — ต้องรู้ว่าบัญชีไหนคือบัญชีภาษีหักโดยไม่อ่านชื่อ ซึ่งผังบัญชียังไม่มีคุณสมบัตินี้ → เสนอเป็นงานแยก (เพิ่ม attribute ในผังบัญชี) รอลุงจืดตัดสิน

## Consequences

- ✅ แบบยื่นมีแต่ตัวเลขที่ผู้ใช้บันทึก ตรวจย้อนได้ (audit `withholding_replace`), ไม่มีแถวเดาผิด, โค้ดสั้นลง ~580 บรรทัด
- ⚠️ ใบที่ลงบัญชีภาษีหักแต่ลืมบันทึกรายการ **หายจากรายงานเงียบ ๆ** (เหลือแค่หมายเหตุเมื่อทั้งงวดว่าง) — ต้องกระทบยอดบัญชีภาษีหักกับรายงานเอง (งบทดลอง/บัญชีแยกประเภท) จนกว่าจะมีจอกระทบยอดภาษีกับ GL
- ⚠️ ข้อมูล dev บน prod: ใบ `PV6901-S001` (ม.ค. 3,210) และ `PV6901-S002` (ก.พ. 1,000) ของ rungrueng/01 เคยขึ้นเป็นแถวประมาณใน ภ.ง.ด.53 — หลัง deploy จะหายจนกว่าจะบันทึกรายการภาษีหักในใบ

## Evidence

- integration (PG 18): `TestWithholdingReportRecordedOnly`, `TestWithholdingReportReversalsAndForms2026_09_24`, `TestWithholdingReportUsesRecordedBase`, `TestWithholdingReceivedReport`, `TestWithholdingReportFilingMonth`, `TestWithholdingReportUATRegressions2026_09_24` — ใบที่ไม่ได้บันทึก/ใบโอนยอด/ใบยกมาไม่เป็นแถว, รายการที่บันทึกคงเดิม
- unit `TestWithholdingReportNotes`, frontend `wht-certificate-panel.test.ts`, `thai-tax.test.ts`
