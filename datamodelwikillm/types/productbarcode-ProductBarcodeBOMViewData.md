---
source: backend/internal/product/productbarcode/models/product_bom.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMViewData

โครงสร้าง `ProductBarcodeBOMViewData` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductBarcodeBOMViewInfo | [[productbarcode-ProductBarcodeBOMViewInfo\|ProductBarcodeBOMViewInfo]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[productbarcode-ProductBarcodeBOMViewInfo|ProductBarcodeBOMViewInfo]]
