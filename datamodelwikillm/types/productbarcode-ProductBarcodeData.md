---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# ProductBarcodeData

โครงสร้าง `ProductBarcodeData` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductBarcodeInfo | [[productbarcode-ProductBarcodeInfo\|ProductBarcodeInfo]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[productbarcode-ProductBarcodeInfo|ProductBarcodeInfo]]
