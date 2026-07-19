---
source: backend/internal/product/product/models/product.go
tags: [datamodel, general-type]
---

# MarketplaceAttribute

โครงสร้าง `MarketplaceAttribute` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| AttributeID | string | attributeid | attributeid | ค่าของ AttributeID ตามฟิลด์ `attributeid` ในซอร์ส |
| AttributeCode | string | attributecode | attributecode | ค่าของ AttributeCode ตามฟิลด์ `attributecode` ในซอร์ส |
| AttributeName | string | attributename | attributename | ค่าของ AttributeName ตามฟิลด์ `attributename` ในซอร์ส |
| InputType | string | inputtype | inputtype | ค่าของ InputType ตามฟิลด์ `inputtype` ในซอร์ส |
| Scope | string | scope | scope | ค่าของ Scope ตามฟิลด์ `scope` ในซอร์ส |
| IsRequired | bool | isrequired | isrequired | ค่าของ IsRequired ตามฟิลด์ `isrequired` ในซอร์ส |
| IsSaleProp | bool | issaleprop | issaleprop | ค่าของ IsSaleProp ตามฟิลด์ `issaleprop` ในซอร์ส |
| IsCustom | bool | iscustom | iscustom | ค่าของ IsCustom ตามฟิลด์ `iscustom` ในซอร์ส |
| Values | *[][[product-MarketplaceAttributeValue\|MarketplaceAttributeValue]] | values | values | ค่าของ Values ตามฟิลด์ `values` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[product-MarketplaceAttributeValue|MarketplaceAttributeValue]]
