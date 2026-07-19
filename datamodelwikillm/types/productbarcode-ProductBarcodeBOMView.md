---
source: backend/internal/product/productbarcode/models/product_bom.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMView

โครงสร้าง `ProductBarcodeBOMView` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| BOMProductBarcode | [[productbarcode-BOMProductBarcode\|BOMProductBarcode]] | inline | - | โครงสร้างฝังแบบ inline |
| ImageURI | string | imageuri | imageuri | URI ของรูปภาพ |
| BOM | [][[productbarcode-ProductBarcodeBOMView\|ProductBarcodeBOMView]] | bom | bom | ค่าของ BOM ตามฟิลด์ `bom` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-BOMProductBarcode|BOMProductBarcode]], [[productbarcode-ProductBarcodeBOMView|ProductBarcodeBOMView]]
