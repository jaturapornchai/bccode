---
source: backend/internal/product/ordertype/models/ordertype.go
tags: [datamodel, general-type]
---

# OrderTypeInfo

โครงสร้าง `OrderTypeInfo` จากโมดูล mainapi `ordertype` มีฟิลด์ตามซอร์ส `ordertype.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| OrderType | [[ordertype-OrderType\|OrderType]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[ordertype-OrderType|OrderType]]
