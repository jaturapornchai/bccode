---
source: backend/internal/product/product/models/product.go
tags: [datamodel, general-type]
---

# ProductSupplier

โครงสร้าง `ProductSupplier` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GuidFixed | string | guidfixed | guidfixed | GUID ถาวรของรายการ |
| Code | string | code | code | รหัสรายการ |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[NameX]]
