---
source: backend/internal/product/productgroup/models/productgroup.go
tags: [datamodel, general-type]
---

# ProductGroupInfo

โครงสร้าง `ProductGroupInfo` จากโมดูล mainapi `productgroup` มีฟิลด์ตามซอร์ส `productgroup.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductGroup | [[productgroup-ProductGroup\|ProductGroup]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[productgroup-ProductGroup|ProductGroup]]
