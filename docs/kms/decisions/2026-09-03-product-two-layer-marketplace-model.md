---
date: 2026-09-03
status: accepted
tags: [bc-account, architecture, product, marketplace, mongodb]
---

# สินค้า 2 ชั้นใน collection เดียว + channel_* แยกสำหรับลงขายออนไลน์

> แทนที่ [2026-06-30-marketplace-channelmappings-field.md](2026-06-30-marketplace-channelmappings-field.md) (ADR เดิม status: superseded แล้ว)

## Context
ผู้ใช้ไทย 2 กลุ่มใช้ระบบเดียวกัน: นักบัญชี (ต้องการรหัส/ชื่อ/หน่วย/หมวดครบเร็ว ๆ เพื่อออกเอกสารภาษี) และเจ้าของกิจการ (ต้องการ สี/ขนาด/น้ำหนัก/รูป/ราคาขายส่ง เพื่อลง marketplace) ลุงจืดไม่อยากแยกระบบเพราะข้อมูลเชื่อมกัน (ขายออนไลน์ → สต๊อก/บัญชีตัวเดียวกัน)

## Decision
- **ไม่แยก collection** — `product` และ `productbarcode` แบ่งเป็น 2 ชั้นในเอกสารเดียว: ชั้นบัญชี (code/names/unitcode/หมวด/bom/จุดสั่งซื้อ) กับชั้นลงขาย (`listing`, `package*`, รูป/วิดีโอ+thumb) ชั้นลงขายเว้นว่างได้ทั้งหมด ไม่กระทบเอกสารบัญชี
- **ตัวเลือก (variant) = productbarcode** — 1 บาร์โค้ด = 1 ตัวเลือกที่ขายได้ (สี×ขนาด หรือหน่วยแพ็ก) มีราคา/สต๊อกของตัวเอง; `product.listing.tiers[].options[]` (≤2 ชั้น, ≤50 ชุดผสม) + `productbarcode.listing.tierindex` ชี้ตำแหน่ง
- **สิ่งที่ช่องทางเป็นเจ้าของแยกออก** เป็น 4 collection ใหม่: `channel_shop` (ร้าน+คลัง+ค่าจำกัด/ช่องขนส่ง cache, ไม่เก็บ token), `channel_listing` (สินค้า×ร้าน: externalitemid/externalmodelid, หมวด/คุณลักษณะ/แบรนด์ของช่องทาง, ราคา/สต๊อกที่ส่ง+อ่านกลับ, สถานะ, sync), `channel_category_map`, `channel_brand_map` (จับคู่ครั้งเดียวต่อบริษัท)
- ชื่อฟิลด์/คำอธิบายเป็นกลาง ไม่อ้างชื่อ vendor ใด (ค่า enum `channel` = shopee|lazada|tiktok|line เป็นข้อมูล ไม่ใช่ที่มาของ design)
- MongoModel project "BC Ai Account" diagram "ข้อมูลหลัก" (efaf857d) — lint 0 issue, descriptions ไทยครบ

## Alternatives
- แยก `product_extra` collection ต่อสินค้า → ปฏิเสธ: join ทุกจอ, sync 2 doc, สิทธิ์แยกได้อยู่แล้วด้วย field-level permission
- เก็บ marketplace ทั้งหมดใน product (แบบ Go `marketplaceproducts[]` เดิม) → ปฏิเสธ: sync เขียนบ่อยชนกับงานบัญชี, ไม่รองรับหลายร้าน/หลายคลัง
- variant เป็น collection ใหม่ → ปฏิเสธ: ธุรกรรม (transaction.go) ใช้ barcode+itemcode+unitcode อยู่แล้ว บาร์โค้ดคือ variant โดยธรรมชาติ

## Consequences
- (+) นักบัญชีกรอก 4 ช่องแล้วออกเอกสารได้; เจ้าของเติมทีหลังโดยไม่แตะข้อมูลบัญชี
- (+) หลายร้าน/หลายช่องทาง/หลายคลัง รองรับใน channel_* โดยไม่แก้ product
- (−) ต้อง deprecate Go `marketplaceproducts[]`, `marketplaceskumappings`, `sellersku`, `skupackage*` และ orphan `tab-product-marketplace.tsx` (รอลุงจืดยืนยัน)
- (−) validation ข้ามเอกสาร (tierindex ↔ tiers, GTIN, wholesale ≤ price) ต้องทำที่ backend
- ถัดไป: API contract `/goapi/product/v2/*` (neutral naming) → Go implementation เป็นเฟส (item+model+price/stock ก่อน) โดยชั้นบัญชีห้ามถูก API นี้เขียนทับ
