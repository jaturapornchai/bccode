---
source: mongo-customer-model.go
tags: [datamodel, general-type]
---

# ProcessMongoCustomerModel

โมเดลข้อมูลลูกค้าสำหรับขั้นตอน process ฝั่ง goapi ประกอบด้วยรหัสลูกค้า ประเภทบุคคล ชื่อ (หลายรายการ) เลขผู้เสียภาษี และประเภทลูกค้า

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | holdingcode | holdingcode | รหัสกิจการ (tenant) |
| Code | string | code | code | รหัสลูกค้า |
| PersonalType | int | personaltype | personaltype | ประเภทบุคคล (เช่น บุคคลธรรมดา/นิติบุคคล) |
| Names | [][[ProcessMongoCustomerNameModel]] | names | names | รายการชื่อลูกค้า |
| TaxID | string | taxid | taxid | เลขประจำตัวผู้เสียภาษี |
| CustomerType | int | customertype | customertype | ประเภทลูกค้า |

## ความสัมพันธ์

- ฝัง (embed) `[]` [[ProcessMongoCustomerNameModel]] ในฟิลด์ `Names`
- `HoldingCode` — อ้างอิงรหัสกิจการ/tenant ที่ข้อมูลลูกค้าสังกัด
