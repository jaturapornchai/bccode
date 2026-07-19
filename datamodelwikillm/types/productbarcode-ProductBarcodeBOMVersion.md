---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMVersion

โครงสร้าง `ProductBarcodeBOMVersion` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GuidFixed | string | guidfixed | guidfixed | GUID ถาวรของรายการ |
| StartDate | time.Time | startdate | startdate | วันที่เริ่มต้น |
| EndDate | *time.Time | enddate | enddate | วันที่สิ้นสุด |
| BOM | *[][[productbarcode-BOMProductBarcode\|BOMProductBarcode]] | bom | bom | ค่าของ BOM ตามฟิลด์ `bom` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-BOMProductBarcode|BOMProductBarcode]]
