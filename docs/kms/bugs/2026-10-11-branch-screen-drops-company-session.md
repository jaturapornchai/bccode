---
date: 2026-10-11
severity: high
component: [frontend]
tags: [bc-account, workspace, session, uat]
fixed: true
---

# Symptom

เปิดจอ "กำหนดค่าเริ่มต้นบริษัท" (`/branch`) แล้วไปเปิดจอบัญชี (เช่น `/gl/dimensions`) → `GET /api/gl/accounts`, `/api/gl/departments` ได้ 409 "กรุณาเลือกบริษัทก่อนใช้งานบัญชี" ทั้งที่ผู้ใช้เลือกบริษัท/สาขาไว้แล้ว — พบจาก `npm run uat` (เปิดจอเดี่ยว ๆ ผ่าน แต่เปิดต่อจาก `/branch` ไม่ผ่าน)

## Root Cause
`frontend/src/app/system-settings/company-branch-tree-view.tsx` `ensureActiveWorkspaceHolding` (เรียกตอนโหลดและตอนบันทึก) ส่ง `select-holding` แค่ `holdingcode` → session ใหม่เป็นระดับ holding ไม่มี `businesscode`
backend GL (`backend/internal/generalledger/httpapi/scope.go:26`) ปฏิเสธเมื่อ `BusinessCode` ว่าง — ทุกจอบัญชีที่เปิดต่อจากนั้นใช้ไม่ได้จนกว่าจะเลือกบริษัทใหม่
(คอมเมนต์ใน `frontend/src/lib/general-ledger.ts` `workspaceBranchCode` เคยเลี่ยงอาการเดียวกันฝั่งสาขาไว้แล้ว)

## Fix
`ensureActiveWorkspaceHolding` ส่ง `businesscode` (`workspace.company.code`) และ `branchuid` (`workspace.branch.guidfixed`) ที่ผู้ใช้เลือกไว้ไปด้วย → session ยังอยู่ระดับบริษัท/สาขาเดิม

## Regression Test
- `npm run uat -- --flows=none --screens=all --routes=/branch,/gl/dimensions` (Git Bash: นำหน้า `MSYS_NO_PATHCONV=1`) — fail ก่อนแก้, pass หลังแก้
- `npm run uat` เต็มชุดเปิดจอตามลำดับเมนูซึ่งครอบลำดับนี้อยู่แล้ว
