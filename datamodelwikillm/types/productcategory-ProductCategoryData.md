---
source: backend/internal/product/productcategory/models/productcategory.go
tags: [datamodel, general-type]
---

# ProductCategoryData

โครงสร้าง `ProductCategoryData` จากโมดูล mainapi `productcategory` มีฟิลด์ตามซอร์ส `productcategory.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductCategoryInfo | [[productcategory-ProductCategoryInfo\|ProductCategoryInfo]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[productcategory-ProductCategoryInfo|ProductCategoryInfo]]
