---
source: backend/internal/product/promotion/models/promotion.go
tags: [datamodel, general-type]
---

# PromotionBarcodeInclude

โครงสร้าง `PromotionBarcodeInclude` จากโมดูล mainapi `promotion` มีฟิลด์ตามซอร์ส `promotion.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| PromotionProduct | *[][[promotion-ProductBarcode\|ProductBarcode]] | promotionproduct | promotionproduct | ค่าของ PromotionProduct ตามฟิลด์ `promotionproduct` ในซอร์ส |
| IncludeProduct | *[][[promotion-ProductBarcode\|ProductBarcode]] | includeproduct | includeproduct | ค่าของ IncludeProduct ตามฟิลด์ `includeproduct` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[promotion-ProductBarcode|ProductBarcode]]
