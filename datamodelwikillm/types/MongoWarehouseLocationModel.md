---
source: backend/internal/goapi/models/global-model.go
tags: [datamodel, general-type]
---

# MongoWarehouseLocationModel

โครงสร้าง location (ที่เก็บ) ภายในคลังสินค้า สำหรับ Kafka message เก็บรหัส location พร้อมชื่อหลายภาษา

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Code | string | - | code | รหัส location ภายในคลัง |
| Names | [][[languageNameModel]] | - | names | ชื่อ location หลายภาษา |

## ความสัมพันธ์
- ฝัง [[languageNameModel]] (Names)
- ถูกฝังใน [[MongoWarehouseModel]] (Location)
