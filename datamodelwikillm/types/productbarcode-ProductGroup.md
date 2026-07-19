---
source: backend/internal/product/productbarcode/models/product_group.go
tags: [datamodel, general-type]
---

# ProductGroup

โครงสร้าง `ProductGroup` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_group.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GuidFixed | string | guidfixed | guidfixed | GUID ถาวรของรายการ |
| Code | string | code | code | รหัสรายการ |
| Names | *[]models.NameX | names | names | รายการชื่อหลายภาษา |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
