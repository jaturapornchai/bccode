---
source: backend/internal/product/productbarcode/models/product_ordertype.go
tags: [datamodel, general-type]
---

# ProductOrderType

โครงสร้าง `ProductOrderType` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_ordertype.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Code | string | code | code | รหัสรายการ |
| Names | *[]models.NameX | names | names | รายการชื่อหลายภาษา |
| Price | float64 | price | price | ราคา |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
