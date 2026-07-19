---
source: backend/internal/product/ordertype/models/ordertype.go
tags: [datamodel, general-type]
---

# OrderTypeData

โครงสร้าง `OrderTypeData` จากโมดูล mainapi `ordertype` มีฟิลด์ตามซอร์ส `ordertype.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| OrderTypeInfo | [[ordertype-OrderTypeInfo\|OrderTypeInfo]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[ordertype-OrderTypeInfo|OrderTypeInfo]]
