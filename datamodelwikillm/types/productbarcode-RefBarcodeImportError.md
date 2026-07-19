---
source: backend/internal/product/productbarcode/models/product_barcode_request.go
tags: [datamodel, general-type]
---

# RefBarcodeImportError

โครงสร้าง `RefBarcodeImportError` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode_request.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Row | int | - | row | ค่าของ Row ตามฟิลด์ `row` ในซอร์ส |
| Barcode | string | - | barcode | บาร์โค้ดสินค้า |
| Error | string | - | error | ข้อความข้อผิดพลาด |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
