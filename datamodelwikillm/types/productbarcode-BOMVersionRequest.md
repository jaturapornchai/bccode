---
source: backend/internal/product/productbarcode/models/product_barcode_request.go
tags: [datamodel, general-type]
---

# BOMVersionRequest

โครงสร้าง `BOMVersionRequest` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode_request.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GuidFixed | string | - | guidfixed | GUID ถาวรของรายการ |
| StartDate | time.Time | - | startdate | วันที่เริ่มต้น |
| EndDate | *time.Time | - | enddate | วันที่สิ้นสุด |
| BOM | [][[productbarcode-BOMRequest\|BOMRequest]] | - | bom | ค่าของ BOM ตามฟิลด์ `bom` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-BOMRequest|BOMRequest]]
