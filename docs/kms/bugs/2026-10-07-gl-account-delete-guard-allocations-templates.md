---
date: 2026-10-07
severity: medium
component: [backend]
tags: [bc-account, gl, postgres, chart-of-accounts]
fixed: true
---

# Symptom

- ลบบัญชีที่ถูกใช้อยู่แค่ในปลายทางของการปันส่วน (`allocations`) หรือในบรรทัดรูปแบบรายการรายวัน (`journal-templates`) ได้สำเร็จ ทั้งที่ควรถูกกัน ทำให้การปันส่วน/รูปแบบรายการรายวันชี้ไปบัญชีที่ไม่มีอยู่
- พบระหว่างทำแผง "รายการที่ใช้บัญชีนี้" (Champ 5502): แผงที่ใช้ predicate เดียวกับตัวกันลบไม่เห็นรายการเหล่านี้ด้วย

## Root Cause

- `accountReferencedByRecord` ใน `backend/internal/generalledger/account_references.go` (ใช้ร่วมกันระหว่างตัวกันลบบัญชีใน `postgres_guards.go` กับแผง references) ไม่ตรวจ `allocations.payload.allocaterules[].accountcode` และ `journal-templates.payload.template.lines[].accountcode`
- ตัวกันลบเดิมตรวจเฉพาะบรรทัดใบสำคัญ กฎ `rules` บัญชีแม่ ปีบัญชี ฯลฯ และบัญชีต้นทางของการปันส่วน (`payload.accountcode`) แต่ไม่ลงไปในลิสต์ปลายทางของการปันส่วนและบรรทัดของแม่แบบ จึงบัญชีที่ใช้แค่ที่นั่นถูกลบได้

## Fix

- commit e4f0d959: เพิ่มเงื่อนไข `kind='allocations'` ตรวจ `allocaterules[].accountcode` และ `kind='journal-templates'` ตรวจ `template.lines[].accountcode` ใน `accountReferencedByRecord`; ตัวกันลบและแผง references อ่าน predicate เดียวกัน จึงไม่เบี่ยงกันอีก

## Regression Test

- `TestAccountReferencesIntegration` ใน `backend/internal/generalledger/account_references_integration_test.go` (build tag `integration`): บัญชี 5100 (ปลายทางการปันส่วน) และ 5200 + 1000 (บรรทัดรูปแบบรายการรายวัน) ต้องมี reference ชนิด `allocations` / `journal-templates` และลบไม่ได้ ส่วนบัญชีที่ไม่มี reference ลบได้
- รัน: `go test -tags=integration ./internal/generalledger/ -run TestAccountReferencesIntegration`
