---
source: backend/internal/product/product/models/product.go
tags: [datamodel, general-type]
---

# ProductActivity

โครงสร้าง `ProductActivity` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ProductData | [[product-ProductData\|ProductData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityTime | [[ActivityTime\|models.ActivityTime]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[product-ProductData|ProductData]], [[ActivityTime]]
