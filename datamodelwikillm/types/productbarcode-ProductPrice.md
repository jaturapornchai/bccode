---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# ProductPrice

โครงสร้าง `ProductPrice` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| KeyNumber | int | keynumber | keynumber | ค่าของ KeyNumber ตามฟิลด์ `keynumber` ในซอร์ส |
| Price | float64 | price | price | ราคา |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
