---
source: backend/internal/product/product/models/product.go
tags: [datamodel, general-type]
---

# ProductDimension

โครงสร้าง `ProductDimension` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| IsDisabled | bool | isdisabled | isdisabled | ค่าของ IsDisabled ตามฟิลด์ `isdisabled` ในซอร์ส |
| Item | [[product-ProductDimensionItem\|ProductDimensionItem]] | item | item | ค่าของ Item ตามฟิลด์ `item` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[NameX]], [[product-ProductDimensionItem|ProductDimensionItem]]
