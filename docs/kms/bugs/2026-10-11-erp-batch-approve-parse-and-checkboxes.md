---
date: 2026-10-11
severity: medium
component: [frontend]
tags: [bc-account, erp, approval]
fixed: true
---

# Symptom

อนุมัติเอกสาร ERP หลายใบพร้อมกันแล้วจอบอกว่าอนุมัติได้ 0 ใบเสมอ และใบที่ล้มไม่แสดงเหตุผล; บนจอยกเลิกอนุมัติ/ยกเลิกเอกสาร ไม่มีช่องติ๊กเลือกหลายใบ ทำให้ปุ่มทำทีละหลายใบใช้ไม่ได้

## Root Cause
- backend `/transaction/:module/approve-batch` คืน `data` เป็น **array** ของ `{id, success, code, message}` แต่ `frontend/src/lib/erp-transaction.ts` `approveBatchErpDocs` อ่าน `data.data.items/.total/.approved` (รูปแบบเก่า) และ type ฝั่ง TS ใช้ `docno` แทน `id`
- `frontend/src/app/operations/operations-workbench.tsx` แสดง checkbox เฉพาะ `actionType === "approve"`

## Fix
- parse array ตาม contract จริงเหมือน `unapproveBatchErpDocs`/`cancelBatchErpDocs`, ใช้ field `id`
- แสดง checkbox ทุก action ที่มีปุ่มทำหลายใบ (approve/unapprove/cancel)

## Regression Test
- `frontend/src/lib/erp-transaction.test.ts` (กรณี approve-batch สำเร็จบางใบ/ล้มบางใบ)
- `frontend/src/app/operations/operations-workbench.test.ts`
