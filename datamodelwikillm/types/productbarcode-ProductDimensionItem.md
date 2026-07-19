---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# ProductDimensionItem

โครงสร้าง `ProductDimensionItem` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| IsDisabled | bool | isdisabled | isdisabled | ค่าของ IsDisabled ตามฟิลด์ `isdisabled` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[NameX]]
