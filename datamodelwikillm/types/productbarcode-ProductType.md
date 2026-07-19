---
source: backend/internal/product/productbarcode/models/product_type.go
tags: [datamodel, general-type]
---

# ProductType

โครงสร้าง `ProductType` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_type.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Code | string | code | code | รหัสรายการ |
| Names | *[]models.NameX | names | names | รายการชื่อหลายภาษา |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
