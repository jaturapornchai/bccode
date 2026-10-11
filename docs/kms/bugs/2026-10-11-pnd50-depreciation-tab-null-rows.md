---
date: 2026-10-11
severity: medium
component: [backend, frontend]
tags: [bc-account, go, fixed-asset, uat]
fixed: true
---

# Symptom

จอทะเบียนสินทรัพย์ แท็บค่าเสื่อมราคาทางภาษี (ภ.ง.ด.50) ของบริษัทที่ยังไม่มีสินทรัพย์ จอพังทั้งแท็บ (React error อ่าน `.length`/`.map` ของ null) — พบจาก `npm run uat` ขั้น UI smoke

## Root Cause
`backend/internal/fixedasset/reports.go` (`GetAssetScheduleReport`, `GetTaxReconciliationReport`) ประกาศ `var rows []map[string]string` ไม่มีสินทรัพย์ → slice เป็น nil → JSON `"rows":null`
ฝั่ง `frontend/src/app/asset/fixed-assets-screen.tsx` แท็บภาษีอ่าน `taxReportData.rows.length` / `.columns.map` ตรง ๆ ไม่มี fallback

## Fix
- backend: `rows := []map[string]string{}` ทั้ง 2 รายงาน → คืน `[]` เสมอ
- frontend: `(taxReportData?.rows ?? [])`, `(taxReportData?.columns ?? [])`, เงื่อนไขว่าง `!taxReportData?.rows?.length` (defensive ตาม AGENTS.md)

## Regression Test
- `backend/internal/fixedasset/reports_empty_test.go` (`TestReportsEmptyCompanyRowsAreJSONArray`) — ยืนยันแล้วว่า fail เมื่อถอดการแก้ และ pass หลังแก้
- `npm run uat -- --flows=none --screens=asset` เปิดแท็บจริง
