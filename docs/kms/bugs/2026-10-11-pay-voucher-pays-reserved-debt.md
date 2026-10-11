---
date: 2026-10-11
severity: high
component: [backend]
tags: [bc-account, ap, erp, champ-parity, money]
fixed: true
---

# Symptom

1. ใบจ่ายชำระหนี้ (PV) ที่พิมพ์บรรทัดหนี้เอง (ไม่ได้ดึงจากใบเตรียมจ่าย) จ่ายหนี้ที่ใบเตรียมจ่ายอื่นจองไว้ครบแล้วได้ — ยอดที่ใบเตรียมจ่ายจองไว้จึงจ่ายซ้ำไม่ได้ภายหลัง
2. ใบจ่ายที่ลงวันที่ล่วงหน้าตัดหนี้ไปแล้ว แต่จอเตรียมจ่ายยังแสดงยอดที่เตรียมจ่ายได้เต็ม และบันทึกใบเตรียมจ่ายเต็มยอดได้
3. PV ที่มีบรรทัดจากใบเตรียมจ่ายยอดติดลบ/ศูนย์ หักยอดรวมของคู่ (ใบเตรียมจ่าย, หนี้) ทำให้บรรทัดอื่นของคู่เดียวกันเกินยอดที่จองไว้ได้

## Root Cause

- `settleLines` (`backend/internal/transaction/httpapi/pay_settlement.go`) ตรวจแค่ยอดรวมต่อหนี้ ≤ ยอดคงเหลือ ไม่แยกบรรทัดจ่ายตรงออกมาเทียบกับยอดที่ไม่ถูกจอง ส่วน Champ ซ่อนหนี้ที่ถูกจองจากตัวเลือกหนี้ของใบจ่าย (`champ/champ/apfrmappayment1view.cpp:1612-1617`)
- `gl.APOpenItems` (`backend/internal/generalledger/ap_open_items.go`) คำนวณ `Available` ที่ as-of อย่างเดียว ใบจ่ายที่ลงวันที่หลัง as-of จึงมองไม่เห็น ทั้งที่ Champ ลดยอดค้างของหนี้ทันทีที่บันทึกใบจ่ายโดยไม่สนวันที่ (`champ-lib/BC5PROCESS/bcprocess.cpp:932-935`)
- `checkPayPrepareLines` รวมยอดต่อคู่โดยไม่ตรวจว่ายอดแต่ละบรรทัดเป็นบวก

## Fix

- `APOpenItems` อ่านซ้ำที่ `reservationAsOf` (ณ ปัจจุบัน) สำหรับหนี้ที่คืน แล้วจำกัด `Available + PrepareUsed` ไม่ให้เกินค่าเดียวกัน ณ ปัจจุบัน — แก้จอเตรียมจ่าย, `checkPrepareLines` และ `settleLines` ที่จุดเดียว
- `settleLines` รวมบรรทัดที่ `paybillno` ว่างต่อหนี้ แล้วปฏิเสธเมื่อเกิน `min(Available, ยอดค้างปัจจุบัน)` ด้วย `erp_settle_exceeds_remaining` ที่บรรทัดแรกของหนี้นั้น
- `checkPayPrepareLines` ปฏิเสธบรรทัดยอด ≤ 0 ด้วย `erp_amount_invalid` ก่อนรวม

## Regression Test

- `TestIntegrationPrepareAmountsScaleTwo` (`ap_payment_prepare_integration_test.go`) — พังก่อนแก้: PV จ่ายตรง 600 ทั้งที่ใบเตรียมจ่ายจองไว้ครบ ผ่านได้
- `TestIntegrationPrepareSeesFutureDatedPayment` — พังก่อนแก้: หลัง PV ลงวันที่ล่วงหน้าจ่าย 1,070 ตัวเลือกหนี้ยังแสดงจ่ายได้ 1,070.00
- `TestCheckPayPrepareLinesRejectsNonPositiveLine` (`ap_payment_prepare_test.go`) — พังก่อนแก้: บรรทัด -500 ไม่ถูกปฏิเสธ
