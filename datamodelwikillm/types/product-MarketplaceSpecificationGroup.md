---
source: backend/internal/product/product/models/product.go
tags: [datamodel, general-type]
---

# MarketplaceSpecificationGroup

โครงสร้าง `MarketplaceSpecificationGroup` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GroupCode | string | groupcode | groupcode | ค่าของ GroupCode ตามฟิลด์ `groupcode` ในซอร์ส |
| GroupName | string | groupname | groupname | ค่าของ GroupName ตามฟิลด์ `groupname` ในซอร์ส |
| Attributes | *[][[product-MarketplaceAttribute\|MarketplaceAttribute]] | attributes | attributes | ค่าของ Attributes ตามฟิลด์ `attributes` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[product-MarketplaceAttribute|MarketplaceAttribute]]
