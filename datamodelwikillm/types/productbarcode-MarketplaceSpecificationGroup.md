---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# MarketplaceSpecificationGroup

โครงสร้าง `MarketplaceSpecificationGroup` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GroupCode | string | groupcode | groupcode | ค่าของ GroupCode ตามฟิลด์ `groupcode` ในซอร์ส |
| GroupName | string | groupname | groupname | ค่าของ GroupName ตามฟิลด์ `groupname` ในซอร์ส |
| Attributes | *[][[productbarcode-MarketplaceAttribute\|MarketplaceAttribute]] | attributes | attributes | ค่าของ Attributes ตามฟิลด์ `attributes` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-MarketplaceAttribute|MarketplaceAttribute]]
