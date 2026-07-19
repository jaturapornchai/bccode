---
source: backend/internal/goapi/models/global-model.go
tags: [datamodel, general-type]
---

# MongoWarehouseModel

โครงสร้างข้อมูลคลังสินค้าสำหรับ Kafka message (ตาม comment ในโค้ด "Warehouse models for Kafka messages") เก็บรหัสคลัง ชื่อหลายภาษา และรายการ location ภายในคลัง

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| Code | string | - | code | รหัสคลังสินค้า |
| Names | [][[languageNameModel]] | - | names | ชื่อคลังสินค้าหลายภาษา |
| Location | [][[MongoWarehouseLocationModel]] | - | location | รายการ location (ที่เก็บ) ภายในคลัง |

## ความสัมพันธ์
- ฝัง [[languageNameModel]] (Names), [[MongoWarehouseLocationModel]] (Location)
- `HoldingCode` — อ้างอิงกลุ่มกิจการ (tenant)
- `Code` — รหัสคลังสินค้า ใช้อ้างอิงคลังในระบบสต๊อก
