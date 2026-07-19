---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# ProductBarcodeInfo

โครงสร้าง `ProductBarcodeInfo` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductBarcode | [[productbarcode-ProductBarcode\|ProductBarcode]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[HoldingCodeentity]], [[productbarcode-ProductBarcode|ProductBarcode]]
