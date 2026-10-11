---
date: 2026-10-11
severity: high
component: [backend]
tags: [bc-account, gl, erp, cheque, champ-parity]
fixed: true
---

# Symptom

จอบัญชีแยกประเภท (`POST /gl/v2/command`) แก้ไข ลบ หรือกลับรายการใบสำคัญที่เอกสาร ERP หรือเอกสารสถานะเช็คสร้างไว้ได้ตรงๆ เช่น กลับรายการใบสำคัญของใบกำกับขายจากจอ GL แล้วใบกำกับขายยังแสดงว่าลงบัญชีอยู่ หรือใบสำคัญของเอกสารเช็คถูกกลับรายการแต่ทะเบียนเช็คยังเป็นสถานะเดิม ทำให้เอกสารต้นทางกับ GL ไม่ตรงกัน

## Root Cause

`mutateJournal` (`backend/internal/generalledger/postgres_store.go`) ตรวจแค่สถานะ/งวด/เวอร์ชัน ไม่ดูว่าใบสำคัญเป็นของโมดูลไหน และไม่แยกว่าคำสั่งมาจากโมดูลเจ้าของเอกสาร (`ExecuteSourceTx`) หรือจากจอ GL ทั้งที่ Champ ห้าม: `BCTrans.Source != 0` บันทึกไม่ได้ (`champ/champ/glfrmjournal.cpp:524-529`), ลบไม่ได้ (`:721-724`), ยกเลิก Ctrl+F5 ผ่านการตรวจเดียวกัน (`:1554-1557`); การกลับรายการทำตอนยกเลิกเอกสารต้นทาง (`myespeciallyformview.cpp:2714`)

## Fix

- `generalledger.SourceSystemERP` / `SourceSystemCheque` + `moduleSourceSystems` (`source_posting.go`) เป็นรายชื่อระบบต้นทางของโมดูลในโปรแกรม; ERP (`transaction/httpapi/posting.go`) และเช็ค (`cheque/glposting/glposting.go`) ใช้ค่าคงที่นี้
- `ExecuteSourceTx` ตั้ง `Command.fromModule` และบังคับว่าใบสำคัญที่สร้างต้องมีระบบต้นทางที่อยู่ในรายชื่อ
- `mutateJournal`: คำสั่งที่ไม่ได้มาจากโมดูล → แก้ไข/ลบ/กลับรายการใบสำคัญของโมดูลไม่ได้ (`journal_from_module`, 409) แต่ยังผ่านบัญชีได้ (Champ 5512 ผ่านบัญชีได้ทุกแหล่ง); สร้างใบสำคัญด้วยชื่อระบบต้นทางของโมดูลไม่ได้ (`journal_source_system_reserved`)
- ใบสำคัญที่นำเข้าจากระบบภายนอก (`source_system` อื่น) ยังแก้จาก GL ได้เหมือนเดิม เพราะไม่มีเอกสารต้นทางในโปรแกรมให้แก้แทน

## Regression Test

`TestGLCommandCannotChangeModuleVoucherIntegration` (`backend/internal/generalledger/source_posting_integration_test.go`) — พังก่อนแก้ ("GL update was accepted"): GL แก้/ลบร่าง, กลับรายการใบที่ผ่านแล้ว, สร้างด้วย `source_system=erp` ถูกปฏิเสธด้วย code ที่ถูกต้อง; GL ผ่านบัญชีได้; โมดูลกลับรายการใบของตัวเองได้
