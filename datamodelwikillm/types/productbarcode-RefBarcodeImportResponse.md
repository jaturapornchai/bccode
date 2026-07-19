---
source: backend/internal/product/productbarcode/models/product_barcode_request.go
tags: [datamodel, general-type]
---

# RefBarcodeImportResponse

โครงสร้าง `RefBarcodeImportResponse` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode_request.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Success | bool | - | success | สถานะความสำเร็จ |
| TotalItems | int | - | totalitems | ค่าของ TotalItems ตามฟิลด์ `totalitems` ในซอร์ส |
| Updated | int | - | updated | ค่าของ Updated ตามฟิลด์ `updated` ในซอร์ส |
| Failed | int | - | failed | ค่าของ Failed ตามฟิลด์ `failed` ในซอร์ส |
| Errors | [][[productbarcode-RefBarcodeImportError\|RefBarcodeImportError]] | - | errors,omitempty | รายการข้อผิดพลาด |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-RefBarcodeImportError|RefBarcodeImportError]]
