---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# ProductBarcodeActivity

โครงสร้าง `ProductBarcodeActivity` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ProductBarcodeData | [[productbarcode-ProductBarcodeData\|ProductBarcodeData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityTime | [[ActivityTime\|models.ActivityTime]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductBarcodeData|ProductBarcodeData]], [[ActivityTime]]
