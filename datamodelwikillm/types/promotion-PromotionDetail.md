---
source: backend/internal/product/promotion/models/promotion.go
tags: [datamodel, general-type]
---

# PromotionDetail

โครงสร้าง `PromotionDetail` จากโมดูล mainapi `promotion` มีฟิลด์ตามซอร์ส `promotion.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DetailType | int8 | detailtype | detailtype | ค่าของ DetailType ตามฟิลด์ `detailtype` ในซอร์ส |
| Minimum | float64 | minimum | minimum | ค่าของ Minimum ตามฟิลด์ `minimum` ในซอร์ส |
| Discount | float64 | discount | discount | ค่าของ Discount ตามฟิลด์ `discount` ในซอร์ส |
| ProductBarcode | [[promotion-ProductBarcode\|ProductBarcode]] | productbarcode | productbarcode | ค่าของ ProductBarcode ตามฟิลด์ `productbarcode` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[promotion-ProductBarcode|ProductBarcode]]
