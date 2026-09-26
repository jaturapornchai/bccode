---
date: 2026-09-26
severity: medium
component: [frontend]
tags: [bc-account, gl, financial-statements, ux]
fixed: true
---

# Symptom

1. จอรูปแบบงบการเงิน (`/gl/statement-designer`): กด "ใช้แม่แบบมาตรฐาน..." แล้วเลือกแม่แบบ → ประเภทงบ ชื่อ รหัส (ถ้ายังไม่บันทึก) บรรทัด คอลัมน์ และรูปแบบของแม่แบบที่เปิดอยู่ถูกแทนที่ทันที **ไม่ถามก่อน** แม้มีงานออกแบบที่ยังไม่บันทึก — งานหายเงียบ ๆ
2. ปุ่ม "ใช้แม่แบบมาตรฐาน" ในจอว่างเปิดแม่แบบใหม่ที่ถูกนับว่า "มีการแก้ไข" ทันที ทั้งที่ผู้ใช้ยังไม่ได้แตะอะไร (dirty guard เตือนเกิน)
3. (พบตอนเพิ่ม dialog ยืนยัน) ดับเบิลคลิก "ใช้แม่แบบนี้" → คลิกที่สองตกบนปุ่ม "ใช้แม่แบบนี้แทน" ของ dialog ที่เพิ่งโผล่ใต้เมาส์ → ยืนยันเองโดยผู้ใช้ไม่เห็น dialog; ถ้าตกบนพื้นหลัง dialog ก็ปิดเอง (ปุ่มดูเหมือนไม่ทำงาน)

## Root Cause

- `applyStarterTemplate` (`frontend/src/app/gl/gl-statement-designer.tsx`) เรียก `setTemplate` ตรง ๆ ไม่ผ่าน `confirm` — ต่างจาก `open()` ที่ถามเมื่อ `dirty`
- ปุ่มในจอว่าง `setTemplate(fresh)` โดยไม่ `setOriginal` → `dirty = JSON.stringify(template) !== original` เป็นจริงทันที
- `useConfirmDialog` (`frontend/src/components/ui/confirm-dialog.tsx`) รับคลิกแรกที่ถึงปุ่ม/พื้นหลังทันที ไม่มีตัวกันคลิกซ้อน — ผู้ใช้ 40+ กดปุ่มเบิ้ลบ่อย (กฎโปรเจกต์)

## Fix

- `statementStarterReplaceNeedsConfirm(template, dirty)` (มีแม่แบบเปิดอยู่ และ ยังไม่บันทึก/มีบรรทัด/มีคอลัมน์) → `confirm` tone warning: ชื่อแม่แบบทั้งสอง, รหัสที่จะเปลี่ยน (`statementStarterReplacedCode`), "การแก้ไขที่ยังไม่ได้บันทึกจะหายไป" (เมื่อ dirty), "แม่แบบที่บันทึกไว้แล้วจะยังไม่เปลี่ยน จนกว่าจะกด บันทึกแม่แบบ" (เมื่อมี id); ยกเลิก = หน้าต่างเลือกแม่แบบยังเปิด; สร้างแม่แบบใหม่ด้วย `statementTemplateFromStarter` ตัวเดียวทั้งใน dialog และตอนแทนที่ (`frontend/src/lib/general-ledger.ts`)
- ปุ่มในจอว่างตั้ง `original` คู่กับแม่แบบใหม่
- dialog ยืนยันกลาง: `isEarlyConfirmClick` ไม่รับคลิกที่ `event.detail > 1` หรือภายใน `CONFIRM_EARLY_CLICK_MS` (400ms) หลังเปิด ทั้งปุ่มยืนยัน ยกเลิก X และพื้นหลัง; Escape ยกเลิกได้ทันที — ทุกจอที่ใช้ `confirm()` ได้ผลด้วย (Playwright ต้องรอ ≥ 450ms ก่อนกดยืนยัน — `tests/employee-uat.spec.ts`)
- commit `d29cba39`, deploy r20260926-7; SKILL `ui-scale-polish` §8.57–8.58

## Regression Test

- `frontend/src/lib/general-ledger.test.ts` "standard statement template replaces the open template only after confirmation" — helper ทุกกรณี + ตรวจ source ว่า `applyStarterTemplate` รอ `confirm` ก่อนแทนที่ และปุ่มจอว่างตั้ง `original`
- `frontend/src/components/ui/confirm-dialog.test.ts` — `isEarlyConfirmClick` + ทั้ง 4 ทางปิดผ่านตัวกัน
- UAT production (Demo, rungrueng/01): จอว่างไม่ถาม; ทับด้วย PNL-DBD แล้วยกเลิก → BS-DBD อยู่ครบ; ดับเบิลคลิกเมาส์จริง → dialog ยังเปิด; คลิก detail 2 บนปุ่มยืนยัน → ไม่ยืนยัน; คลิกปกติ → แทนที่; แม่แบบที่บันทึกแล้วขึ้นประโยค "ยังไม่เปลี่ยนจนกว่าจะกดบันทึก" และ PG ยัง version 1 หลังยกเลิก; ลบแม่แบบทดสอบแล้ว (isdeleted true)
