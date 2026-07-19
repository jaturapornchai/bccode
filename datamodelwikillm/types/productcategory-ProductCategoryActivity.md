---
source: backend/internal/product/productcategory/models/productcategory.go
tags: [datamodel, general-type]
---

# ProductCategoryActivity

โครงสร้าง `ProductCategoryActivity` จากโมดูล mainapi `productcategory` มีฟิลด์ตามซอร์ส `productcategory.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ProductCategoryData | [[productcategory-ProductCategoryData\|ProductCategoryData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityTime | [[ActivityTime\|models.ActivityTime]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productcategory-ProductCategoryData|ProductCategoryData]], [[ActivityTime]]
