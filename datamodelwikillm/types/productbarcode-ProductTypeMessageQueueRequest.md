---
source: backend/internal/product/productbarcode/models/product_type.go
tags: [datamodel, general-type]
---

# ProductTypeMessageQueueRequest

โครงสร้าง `ProductTypeMessageQueueRequest` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_type.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductType | [[productbarcode-ProductType\|ProductType]] | - | - | โครงสร้างฝัง |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductType|ProductType]]
