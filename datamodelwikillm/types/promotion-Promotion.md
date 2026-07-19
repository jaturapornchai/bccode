---
source: backend/internal/product/promotion/models/promotion.go
tags: [datamodel, general-type]
---

# Promotion

โครงสร้าง `Promotion` จากโมดูล mainapi `promotion` มีฟิลด์ตามซอร์ส `promotion.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| PartitionIdentity | [[PartitionIdentity\|models.PartitionIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| PromotionType | int8 | promotiontype | promotiontype | ค่าของ PromotionType ตามฟิลด์ `promotiontype` ในซอร์ส |
| Index | int64 | index | index | ค่าของ Index ตามฟิลด์ `index` ในซอร์ส |
| Code | string | code | code | รหัสรายการ |
| Name | *[][[NameX\|models.NameX]] | name | name | ชื่อรายการ |
| DateBegin | time.Time | datebegin | datebegin | ค่าของ DateBegin ตามฟิลด์ `datebegin` ในซอร์ส |
| DateEnd | time.Time | dateend | dateend | ค่าของ DateEnd ตามฟิลด์ `dateend` ในซอร์ส |
| CustomerOnly | int8 | customeronly | customeronly | ค่าของ CustomerOnly ตามฟิลด์ `customeronly` ในซอร์ส |
| DiscountText | string | discounttext | discounttext | ค่าของ DiscountText ตามฟิลด์ `discounttext` ในซอร์ส |
| LimitQty | float64 | limitqty | limitqty | ค่าของ LimitQty ตามฟิลด์ `limitqty` ในซอร์ส |
| PromotionQty | float64 | promotionqty | promotionqty | ค่าของ PromotionQty ตามฟิลด์ `promotionqty` ในซอร์ส |
| LimitAmount | float64 | limitamount | limitamount | ค่าของ LimitAmount ตามฟิลด์ `limitamount` ในซอร์ส |
| PromotionBarcodeInclude | *[][[promotion-PromotionBarcodeInclude\|PromotionBarcodeInclude]] | promotionbarcodeinclude | promotionbarcodeinclude | ค่าของ PromotionBarcodeInclude ตามฟิลด์ `promotionbarcodeinclude` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[PartitionIdentity]], [[NameX]], [[promotion-PromotionBarcodeInclude|PromotionBarcodeInclude]]
