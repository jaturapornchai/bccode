---
source: backend/internal/product/productbarcode/models/product_unit.go
tags: [datamodel, general-type]
---

# ProductUnitMessageQueueRequest

โครงสร้าง `ProductUnitMessageQueueRequest` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_unit.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductUnit | [[productbarcode-ProductUnit\|ProductUnit]] | - | - | โครงสร้างฝัง |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductUnit|ProductUnit]]
