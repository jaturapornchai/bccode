---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# MarketplaceAttributeValue

โครงสร้าง `MarketplaceAttributeValue` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ValueID | string | valueid | valueid | ค่าของ ValueID ตามฟิลด์ `valueid` ในซอร์ส |
| ValueCode | string | valuecode | valuecode | ค่าของ ValueCode ตามฟิลด์ `valuecode` ในซอร์ส |
| ValueText | string | valuetext | valuetext | ค่าของ ValueText ตามฟิลด์ `valuetext` ในซอร์ส |
| DisplayText | string | displaytext | displaytext | ค่าของ DisplayText ตามฟิลด์ `displaytext` ในซอร์ส |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| SortOrder | int | sortorder | sortorder | ค่าของ SortOrder ตามฟิลด์ `sortorder` ในซอร์ส |
| IsCustomValue | bool | iscustomvalue | iscustomvalue | ค่าของ IsCustomValue ตามฟิลด์ `iscustomvalue` ในซอร์ส |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
