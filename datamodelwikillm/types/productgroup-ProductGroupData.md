---
source: backend/internal/product/productgroup/models/productgroup.go
tags: [datamodel, general-type]
---

# ProductGroupData

โครงสร้าง `ProductGroupData` จากโมดูล mainapi `productgroup` มีฟิลด์ตามซอร์ส `productgroup.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductGroupInfo | [[productgroup-ProductGroupInfo\|ProductGroupInfo]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[productgroup-ProductGroupInfo|ProductGroupInfo]]
