---
date: 2026-10-11
severity: medium
component: [backend]
tags: [bc-account, ap, erp, champ-parity]
fixed: true
---

# Symptom

ใบจ่ายชำระหนี้ (PV) ที่ดึงบรรทัดจากใบเตรียมจ่ายไว้แล้ว แก้ไขไม่ได้อีกเลยหลังจากใบเตรียมจ่ายใบนั้นถูกยกเลิก — บันทึกแล้วขึ้น "ใบเตรียมจ่ายนี้ถูกยกเลิกแล้ว" แม้แค่แก้ยอดหรือหมายเหตุของบรรทัดเดิม ทั้งที่ Champ แก้ได้

## Root Cause

`checkPayPrepareLines` (`backend/internal/transaction/httpapi/pay_settlement.go`) ปฏิเสธทุกบรรทัดที่อ้างใบเตรียมจ่ายที่ยกเลิกแล้ว โดยไม่แยกบรรทัดที่ PV ดึงไว้ก่อนการยกเลิกออกจากบรรทัดที่เพิ่งดึงใหม่ ส่วน Champ `PaymentProcessStatement` (`champ-lib/BC5PROCESS/bcprocess.cpp:490-515`) แค่ UPDATE ยอดค้างของ `BCStatementSub` โดยไม่ตรวจการยกเลิก และจอใบจ่ายบันทึก `PayBillNo` เดิมของบรรทัดตามที่มี (`champ/champ/apfrmappayment1view.cpp:971`)

## Fix

- `SaveDoc` (`save.go`) อ่านคู่ `(paybillno, debt_document_id)` ของ PV เดิมก่อนลบบรรทัด (`payPrepareLinesOf`) แล้วส่งให้ `checkPayPrepareLines`
- ใบเตรียมจ่ายที่ยกเลิกแล้ว: คู่เดิมแก้ต่อได้ ไม่ตรวจยอดคงเหลือของใบเตรียมจ่าย (ยอดจ่ายยังคุมด้วยยอดคงเหลือของหนี้ใน `settleLines`); คู่ใหม่ยังถูกปฏิเสธด้วย `ap_prepare_cancelled_locked`
- ข้อเทียบ Champ อีกข้อ "ตัวเลือกหนี้ซ่อนใบที่ถูกจองบางส่วน" (`apfrmappayment1view.cpp:1612-1617`) ไม่ต้องแก้: PV ของเราไม่มีตัวเลือกหนี้ตรง ดึงได้ทางใบเตรียมจ่าย (Ctrl+F3) เท่านั้น

## Regression Test

`TestIntegrationPaymentPrepareLifecycle` (`backend/internal/transaction/httpapi/ap_payment_prepare_integration_test.go`) — พังก่อนแก้ ("edit of a payment whose preparation was cancelled: ใบเตรียมจ่ายนี้ถูกยกเลิกแล้ว"): แก้ PV เดิมจาก 100 เป็น 90 ได้ ยอดค้างหนี้เป็น 124; เพิ่มบรรทัดใหม่จากใบเตรียมจ่ายที่ยกเลิกแล้วถูกปฏิเสธ
