---
date: 2026-10-11
severity: high
component: [backend, db]
tags: [bc-account, postgres, erp, audit-trail]
fixed: true
---

# Symptom

ปรับสถานะใบรับสินค้า (receipt adjustment) แล้วงานถูกยกเลิกทั้ง transaction ทั้งที่ไม่มีข้อความผิดพลาดที่บอกสาเหตุ (พบจากการอ่านโค้ดในการตรวจ handoff Champ parity 2026-10-11)

## Root Cause
`backend/internal/transaction/httpapi/receipt_adjustments.go` บันทึก `erp_doc_events` ด้วย action `adjust_status` แต่ CHECK ของตาราง (`schema.go`) ยอมแค่ `approve/unapprove/cancel/update` และโค้ดกลืน error ด้วย `_, _ = tx.ExecContext(...)`
ใน PostgreSQL คำสั่งที่ล้มจะทำให้ transaction ทั้งก้อนเป็น aborted → คำสั่งถัดไปและ `COMMIT` ล้ม → การปรับสถานะหายทั้งหมด

## Fix
- `schema.go`: CHECK ของ `erp_doc_events.action` เพิ่ม `adjust_status` (ทั้ง CREATE และ DROP/ADD CONSTRAINT ใน ensureSchema)
- `receipt_adjustments.go`: ไม่กลืน error อีกแล้ว คืน error ขึ้นไป + บันทึก `to_status` และเหตุผลเป็น key (แปลตอนอ่าน)

## Regression Test
- `backend/internal/transaction/httpapi/receipt_adjustments_test.go` (`TestProcessReceiptAdjustments`)
- กฎทั่วไป: ห้าม `_, _ =` กับคำสั่งเขียนใน `sql.Tx` — error ใด ๆ ใน tx ต้อง return
