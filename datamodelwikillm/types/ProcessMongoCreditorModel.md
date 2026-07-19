---
source: mongo-creditor-model.go
tags: [datamodel, general-type]
---

# ProcessMongoCreditorModel

โมเดลข้อมูลเจ้าหนี้ (creditor) ที่ใช้ในขั้นตอน process ข้อมูลจาก MongoDB ประกอบด้วยรหัส ชื่อหลายภาษา ข้อมูลภาษี และเงื่อนไขเครดิต

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | holdingcode | holdingcode | รหัสกิจการ (tenant) |
| GuidFixed | string | guidfixed | guidfixed | GUID ประจำ record |
| Code | string | code | code | รหัสเจ้าหนี้ |
| Names | `[]`[[ProcessMongoCreditorNameModel]] | names | names | รายการชื่อเจ้าหนี้ (หลายรายการ) |
| TaxID | string | taxid | taxid | เลขประจำตัวผู้เสียภาษี |
| PersonalType | int8 | personaltype | personaltype | ประเภทบุคคล (บุคคลธรรมดา/นิติบุคคล) |
| CustomerType | int32 | customertype | customertype | ประเภทลูกค้า |
| BranchNumber | string | branchnumber | branchnumber | เลขที่สาขา (สำหรับภาษี) |
| FundCode | string | fundcode | fundcode | รหัสกองทุน/แหล่งเงิน |
| CreditDay | int32 | creditday | creditday | จำนวนวันเครดิต |
| Email | string | email | email | อีเมล |
| AddressBilling | map[string]any | addressforbilling | addressforbilling | ที่อยู่สำหรับออกบิล (โครงสร้างอิสระ) |

## ความสัมพันธ์

- Embed: `Names` เป็น slice ของ [[ProcessMongoCreditorNameModel]] (ชื่อเจ้าหนี้หลายรายการ)
- `HoldingCode` — อ้างอิงรหัสกิจการ (tenant boundary)
- `Code` — รหัสเจ้าหนี้ ใช้อ้างอิงเจ้าหนี้ในเอกสารอื่น
- `FundCode` — รหัสอ้างอิงกองทุน/แหล่งเงิน (ไม่พบ struct ปลายทางในไฟล์นี้)
