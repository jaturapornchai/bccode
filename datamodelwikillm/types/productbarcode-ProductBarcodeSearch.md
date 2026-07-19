---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# ProductBarcodeSearch

โครงสร้าง `ProductBarcodeSearch` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ICCode | string | - | iccode | ค่าของ ICCode ตามฟิลด์ `iccode` ในซอร์ส |
| Barcode | string | - | barcode | บาร์โค้ดสินค้า |
| UnitCode | string | - | unitcode | รหัสหน่วยนับ |
| Price | string | - | price | ราคา |
| Names | []string | - | names | รายการชื่อหลายภาษา |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
