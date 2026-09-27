---
date: 2026-09-27
severity: low
component: [frontend]
tags: [bc-account, gl, financial-statements]
fixed: true
---

# Symptom

จอรูปแบบงบการเงิน (`/gl/statement-designer`): แม่แบบที่ผู้ใช้ปิด "เปิดใช้งานแม่แบบนี้" ไว้ เมื่อกด "ใช้แม่แบบมาตรฐาน..." ทับแล้วบันทึก กลายเป็นเปิดใช้งานเงียบ ๆ — แม่แบบที่ตั้งใจปิดจึงกลับมาอยู่ในรายการพิมพ์ชุดงบการเงิน (`statementSetTemplates` / `reports/statement-set` เลือกเฉพาะที่เปิดใช้)

## Root Cause

`statementTemplateFromStarter` (`frontend/src/lib/general-ledger.ts`) สร้างแม่แบบใหม่ด้วย `...starter` แล้วคืนแค่ `id`/`version`/`code` ของแม่แบบที่เปิดอยู่ — `isactive` ซึ่งเป็น **สถานะ** ไม่ใช่รูปแบบ จึงได้ค่าของแม่แบบมาตรฐาน (true) เสมอ

## Fix

`isactive: current?.isactive ?? starter.isactive` — มีแม่แบบเปิดอยู่ = คงค่าเดิม, ไม่มี = ค่าของแม่แบบมาตรฐาน; ฟิลด์อื่นที่ `save()` ส่ง (ประเภทงบ ชื่อ รูปแบบ บรรทัด คอลัมน์) เป็นรูปแบบที่แม่แบบมาตรฐานต้องแทนที่อยู่แล้ว — commit `ef7d1f26`, deploy r20260927-2

## Regression Test

- `frontend/src/lib/general-ledger.test.ts` "keeps the open template's active flag instead of taking the starter's" — บันทึกแล้วปิด/เปิดไว้, ไม่มีแม่แบบเปิด (true/false), ใหม่ที่ผู้ใช้ยกเลิกติ๊ก, ใหม่ว่าง
- UAT production (Demo rungrueng/01): COGS-STMT ปิดใช้ → บันทึก (PG `isactive=false` v2) → ใช้แม่แบบมาตรฐานทับ (ถามยืนยัน) → ช่องยังไม่ติ๊ก → บันทึก → PG `isactive=false` v3; `reports/statement-set?templates=COGS-STMT` = 400 `statement_set_template_inactive`
