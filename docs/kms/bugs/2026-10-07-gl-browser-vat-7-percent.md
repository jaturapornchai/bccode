---
date: 2026-10-07
severity: medium
component: [frontend]
tags: [bc-account, gl, tax, vat, accounting-number]
fixed: true
---

# Symptom

- จอบันทึกรายวันมีการ์ด "ตรวจสอบภาษีมูลค่าเพิ่ม 7%" และปุ่ม "ปรับเป็น 7% พอดี" ที่คำนวณ VAT 7% ของฐานภาษีใน browser แล้วแก้ยอดบรรทัดภาษีให้เอง
- จอขายสินทรัพย์ (`frontend/src/app/asset/fixed-assets-screen.tsx`) เติม VAT อัตโนมัติด้วย `Math.round(price * 0.07 * 100) / 100` ทุกครั้งที่พิมพ์ราคาขาย

## Root Cause

- ตัวตรวจ Smart Tax & Balance Guard (commit 068ac7c6, 2026-09-18) ฝังอัตรา 7% ใน `calculateVat7`/`setExactVatLine` ของ `frontend/src/lib/gl-smart-guard.ts` ก่อนมีกรอบภาษีบนหัวใบสำคัญ
- ขัด `mydocs/specs/gl/tax-design-notes.md:19-20` (ห้ามอัตราตั้งต้น 7%/3%, ตรวจเงื่อนไขภาษีที่ Backend) และกฎห้ามคำนวณเงินด้วย float ในจอขายสินทรัพย์
- เมื่อย้ายกรอบภาษีขึ้นหัวเอกสาร (backend `vat_calc.go` คำนวณจากอัตราที่ผู้ใช้กรอก) ตัวตรวจ 7% ไม่ถูกลบตาม

## Fix

- commit 4fcca458: ลบส่วน VAT ของ `gl-smart-guard.ts` เหลือ `analyzeGLBalance` (ดุลเดบิต/เครดิต) + `autoBalanceJournalLines`; ลบการ์ด VAT ใน `gl-journals.tsx`; ลบ 4 key ภาษาที่ไม่มีผู้ใช้
- จอขายสินทรัพย์ไม่เติม VAT เอง ผู้ใช้กรอกจำนวนภาษีตามใบกำกับ; ป้ายไม่ระบุ 7%

## Regression Test

- `frontend/src/lib/gl-smart-guard.test.ts` ทดสอบเฉพาะดุลเดบิต/เครดิตและ auto-balance (ไม่มีฟังก์ชันอัตราภาษีให้เรียกอีก)
- ตรวจซ้ำ: `grep -rn "0\.07\|calculateVat7\|setExactVatLine" frontend/src` ต้องไม่พบในโค้ดคำนวณ
