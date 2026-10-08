---
date: 2026-10-08
status: accepted
tags: [bc-account, erp, transaction, go, champ-parity]
---

# ทะเบียนประเภทเอกสาร ERP เดียว (`backend/internal/doctype`)

## Context
หลัง commit `6c4447f8` (ปลดล็อกจอเอกสาร ERP) พบว่า `/transaction/:module`, เครื่องคำนวณสต็อก (`stockengine`) และรายงานแต่ละตัวถือพจนานุกรม transflag คนละชุดบนตาราง `docdetail` เดียวกัน (เช่น 56/60/62/72/310 ความหมายขัดกัน), key ซ้ำ flag (34, 62, 28), `apiPath` ฝั่ง frontend 20 เส้นทางไม่ตรง route ของ backend, และเอกสารที่ใช้ `apiPath` ร่วมกันถูกเก็บเป็นใบซื้อ/ใบจ่ายแล้วลง GL/สต็อกผิดประเภท รายละเอียด: [`bugs/2026-10-08-erp-unlock-6c4447f8-audit.md`](../bugs/2026-10-08-erp-unlock-6c4447f8-audit.md)

## Decision
- ทะเบียนเดียว `backend/internal/doctype/doctype.go` — แต่ละประเภทมี `{Key, Flag, Prefix, Stock}` (63 ประเภท; `Stock` = None/In/Out/ByLine) ทุกส่วนอ่านจากที่นี่: `/transaction/:module`, `myglobal.TransFlagsToProcess = doctype.StockFlags()`, `stockengine`, รายงานขาย/หนี้/สต็อก
- เก็บเลข transflag ของ BC ไว้ (ไม่ใช้ `GLSource` ของ Champ) — Champ ไม่มี transflag ต่อเอกสาร; ทดสอบกันซ้ำและกันชนกันที่ `doctype_test.go`
- ฝั่ง frontend `apiPath` = key ชั้นเดียวของทะเบียน; `erp-transaction-modules.ts` เท่าทะเบียนเป๊ะและเป็น allowlist ของ BFF — `erp-transaction.test.ts` อ่าน `doctype.go` เทียบ
- ลงบัญชี/ตัดสต็อกอัตโนมัติเฉพาะ 8 ประเภทที่ `SaveDoc` ทำอยู่ ที่เหลือ store-only และจอแจ้งผู้ใช้ (ไม่แกล้งทำเป็นลงบัญชี)
- เลขที่เอกสาร: backend ออกให้เมื่อ `docno` ว่าง `PREFIX+YY(พ.ศ.)+MM-NNNN` คำนำหน้าคงที่ตามทะเบียน
- ตารางเอกสาร/สต็อกสร้างเองครั้งแรกที่ใช้ (`httpapi/schema.go`) เพราะ DDL ไม่เคยอยู่ในโค้ด

## Alternatives
- **ใช้ `GLSource` ของ Champ เป็นฐาน** — ไม่เลือก: ครอบคลุมเฉพาะเอกสารที่ลง GL ไม่ครบ 63 ประเภท และเปลี่ยน flag ที่ข้อมูลเดิมใช้อยู่
- **แยกตารางต่อเอกสารแบบ Champ** — ไม่เลือก: เกินขอบเขต (Strict Parity ห้ามคิดโครงใหม่) และต้อง migrate ทุกรายงาน
- **คืน guard "รอพัฒนา" ให้ทุกจอที่ไม่มี backend** — ไม่เลือกสำหรับเอกสาร: ขัดคำสั่ง "ไม่มีคำว่ารอพัฒนา"; ใช้ guard เฉพาะจอสินค้า 4 จอที่ไม่มี `/api/product` จริง

## Consequences
- เพิ่ม/ลบประเภทเอกสาร = แก้ `doctype.go` ที่เดียว แล้วเทสต์บังคับให้แก้ frontend ตาม; **deploy frontend อย่างเดียวไม่ได้** ต้อง `fast-deploy.py --all`
- คำนำหน้าเลขที่ผู้ใช้ตั้งเองต่อบริษัทตาม Champ (`BCSysTRFormat.DocFormat`) ยังไม่ได้ทำ
- ยังค้าง (รอลุงจืดอนุมัติ): ตรวจสิทธิ์รายเมนู, กลับรายการ GL/สต็อกตอนแก้/ลบเอกสารที่ลงบัญชีแล้ว, ผลต่อสต็อกของ ByLine และ debit/credit note
