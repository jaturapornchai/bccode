---
source: backend/internal/product/product/models/product.go
tags: [datamodel, general-type]
---

# ProductOrderType

โครงสร้าง `ProductOrderType` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Code | string | code | code | รหัสรายการ |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| Price | float64 | price | price | ราคา |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[NameX]]
