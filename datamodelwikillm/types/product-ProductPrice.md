---
source: backend/internal/product/product/models/product.go
tags: [datamodel, general-type]
---

# ProductPrice

โครงสร้าง `ProductPrice` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| KeyNumber | int | - | keynumber | ค่าของ KeyNumber ตามฟิลด์ `keynumber` ในซอร์ส |
| Price | float64 | - | price | ราคา |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
