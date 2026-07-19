---
source: mongo-warehouse-model.go
tags: [datamodel, general-type]
---

# ProcessMongoWarehouseModel

โมเดลข้อมูลคลังสินค้า (warehouse) ที่ใช้ในขั้นตอน process ฝั่ง goapi ประกอบด้วยรหัสคลัง ชื่อหลายภาษา และรายการ location ภายในคลัง

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | holdingcode | holdingcode | รหัสกิจการ (tenant) เจ้าของข้อมูล |
| Code | string | code | code | รหัสคลังสินค้า |
| Names | []​[[LanguageModel]] | names | names | ชื่อคลังสินค้าหลายภาษา |
| Location | []​[[ProcessMongoWarehouseLocationModel]] | location | location | รายการที่เก็บ (location) ภายในคลัง |

## ความสัมพันธ์

- Embed [[LanguageModel]] — ชื่อหลายภาษาใน `Names`
- Embed [[ProcessMongoWarehouseLocationModel]] — รายการ location ภายในคลังใน `Location`
- `HoldingCode` — อ้างอิงกิจการ (tenant/holding) เจ้าของคลังสินค้า
