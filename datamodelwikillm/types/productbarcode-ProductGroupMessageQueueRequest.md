---
source: backend/internal/product/productbarcode/models/product_group.go
tags: [datamodel, general-type]
---

# ProductGroupMessageQueueRequest

โครงสร้าง `ProductGroupMessageQueueRequest` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_group.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductGroup | [[productbarcode-ProductGroup\|ProductGroup]] | - | - | โครงสร้างฝัง |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductGroup|ProductGroup]]
