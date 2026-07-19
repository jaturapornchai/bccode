---
source: backend/internal/product/bom/models/bom.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMView

โครงสร้าง `ProductBarcodeBOMView` จากโมดูล mainapi `bom` มีฟิลด์ตามซอร์ส `bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| BOMProductBarcode | [[bom-BOMProductBarcode\|BOMProductBarcode]] | inline | - | โครงสร้างฝังแบบ inline |
| ImageURI | string | imageuri | imageuri | URI ของรูปภาพ |
| BOM | *[][[bom-ProductBarcodeBOMView\|ProductBarcodeBOMView]] | bom | bom | ค่าของ BOM ตามฟิลด์ `bom` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[bom-BOMProductBarcode|BOMProductBarcode]], [[bom-ProductBarcodeBOMView|ProductBarcodeBOMView]]
