---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# ProductBarcode

โครงสร้าง `ProductBarcode` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| PartitionIdentity | [[PartitionIdentity\|models.PartitionIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductBarcodeBase | [[productbarcode-ProductBarcodeBase\|ProductBarcodeBase]] | inline | - | โครงสร้างฝังแบบ inline |
| RefBarcodes | *[][[productbarcode-RefProductBarcode\|RefProductBarcode]] | refbarcodes | refbarcodes | ค่าของ RefBarcodes ตามฟิลด์ `refbarcodes` ในซอร์ส |
| BOM | *[][[productbarcode-BOMProductBarcode\|BOMProductBarcode]] | bom | bom | ค่าของ BOM ตามฟิลด์ `bom` ในซอร์ส |
| BOMs | *[][[productbarcode-ProductBarcodeBOMVersion\|ProductBarcodeBOMVersion]] | boms | boms | ค่าของ BOMs ตามฟิลด์ `boms` ในซอร์ส |
| BusinessTypes | *[][[productbarcode-ProductBarcodeBusinessType\|ProductBarcodeBusinessType]] | businesstypes | businesstypes | ค่าของ BusinessTypes ตามฟิลด์ `businesstypes` ในซอร์ส |
| IgnoreBranches | *[][[productbarcode-ProductBarcodeBranch\|ProductBarcodeBranch]] | ignorebranches | ignorebranches | ค่าของ IgnoreBranches ตามฟิลด์ `ignorebranches` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[PartitionIdentity]], [[productbarcode-ProductBarcodeBase|ProductBarcodeBase]], [[productbarcode-RefProductBarcode|RefProductBarcode]], [[productbarcode-BOMProductBarcode|BOMProductBarcode]], [[productbarcode-ProductBarcodeBOMVersion|ProductBarcodeBOMVersion]], [[productbarcode-ProductBarcodeBusinessType|ProductBarcodeBusinessType]], [[productbarcode-ProductBarcodeBranch|ProductBarcodeBranch]]
