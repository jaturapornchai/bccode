---
source: backend/internal/product/productbarcode/models/product_barcode_request.go
tags: [datamodel, general-type]
---

# RefBarcodeImportRequest

โครงสร้าง `RefBarcodeImportRequest` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode_request.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Barcode | string | - | barcode | บาร์โค้ดสินค้า |
| StandValue | float64 | - | standvalue | ค่าของ StandValue ตามฟิลด์ `standvalue` ในซอร์ส |
| DivideValue | float64 | - | dividevalue | ค่าของ DivideValue ตามฟิลด์ `dividevalue` ในซอร์ส |
| BarcodeRef | string | - | barcoderef | ค่าของ BarcodeRef ตามฟิลด์ `barcoderef` ในซอร์ส |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
