---
source: backend/internal/product/productcategory/models/productcategory.go
tags: [datamodel, general-type]
---

# ProductCategoryInfo

โครงสร้าง `ProductCategoryInfo` จากโมดูล mainapi `productcategory` มีฟิลด์ตามซอร์ส `productcategory.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductCategory | [[productcategory-ProductCategory\|ProductCategory]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[productcategory-ProductCategory|ProductCategory]]
