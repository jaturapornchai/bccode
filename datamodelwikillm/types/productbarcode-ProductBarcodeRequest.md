---
source: backend/internal/product/productbarcode/models/product_barcode_request.go
tags: [datamodel, general-type]
---

# ProductBarcodeRequest

โครงสร้าง `ProductBarcodeRequest` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode_request.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ProductBarcodeBase | [[productbarcode-ProductBarcodeBase\|ProductBarcodeBase]] | - | - | โครงสร้างฝัง |
| RefBarcodes | [][[productbarcode-BarcodeRequest\|BarcodeRequest]] | - | refbarcodes | ค่าของ RefBarcodes ตามฟิลด์ `refbarcodes` ในซอร์ส |
| BOM | [][[productbarcode-BOMRequest\|BOMRequest]] | - | bom | ค่าของ BOM ตามฟิลด์ `bom` ในซอร์ส |
| BOMs | [][[productbarcode-BOMVersionRequest\|BOMVersionRequest]] | - | boms | ค่าของ BOMs ตามฟิลด์ `boms` ในซอร์ส |
| IgnoreBranches | [][[productbarcode-ProductBarcodeBranch\|ProductBarcodeBranch]] | - | ignorebranches | ค่าของ IgnoreBranches ตามฟิลด์ `ignorebranches` ในซอร์ส |
| BusinessTypes | [][[productbarcode-ProductBarcodeBusinessType\|ProductBarcodeBusinessType]] | - | businesstypes | ค่าของ BusinessTypes ตามฟิลด์ `businesstypes` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductBarcodeBase|ProductBarcodeBase]], [[productbarcode-BarcodeRequest|BarcodeRequest]], [[productbarcode-BOMRequest|BOMRequest]], [[productbarcode-BOMVersionRequest|BOMVersionRequest]], [[productbarcode-ProductBarcodeBranch|ProductBarcodeBranch]], [[productbarcode-ProductBarcodeBusinessType|ProductBarcodeBusinessType]]
