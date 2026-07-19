---
source: backend/internal/product/productbarcode/models/product_ordertype.go
tags: [datamodel, general-type]
---

# ProductOrderTypeMessageQueueRequest

โครงสร้าง `ProductOrderTypeMessageQueueRequest` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_ordertype.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductOrderType | [[productbarcode-ProductOrderType\|ProductOrderType]] | - | - | โครงสร้างฝัง |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductOrderType|ProductOrderType]]
