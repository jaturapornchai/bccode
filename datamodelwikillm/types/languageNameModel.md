---
source: backend/internal/goapi/models/global-model.go
tags: [datamodel, general-type]
---

# languageNameModel

โครงสร้างชื่อแบบหลายภาษา เก็บรหัสภาษาคู่กับข้อความชื่อ ใช้ฝังใน model อื่นเพื่อรองรับชื่อ/คำอธิบายหลายภาษา (เช่น th, en)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Code | string | - | code | รหัสภาษา (เช่น th, en) |
| Name | string | - | name | ข้อความชื่อในภาษานั้น |

## ความสัมพันธ์
- ถูกฝังใช้ใน [[productOrderModel]], [[productOrderUnitModel]], [[productOrderOptionModel]], [[productOrderOptionDetailModel]], [[MongoWarehouseModel]], [[MongoWarehouseLocationModel]]
