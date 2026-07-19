---
source: backend/internal/product/unit/models/unit.go
collection: units
tags: [datamodel, mongodb, mongo-root]
---

# UnitDoc

โครงสร้าง `UnitDoc` เป็น root document ที่ repository โมดูล `unit` ใช้กับ MongoDB collection `units`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| UnitData | [[unit-UnitData\|UnitData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|models.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[unit-UnitData|UnitData]], [[ActivityDoc]]
- MongoDB collection: `units`
