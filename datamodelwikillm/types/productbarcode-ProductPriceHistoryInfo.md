---
source: backend/internal/product/productbarcode/models/product_price_history.go
tags: [datamodel, general-type]
---

# ProductPriceHistoryInfo

โครงสร้าง `ProductPriceHistoryInfo` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_price_history.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductPriceHistory | [[productbarcode-ProductPriceHistory\|ProductPriceHistory]] | inline | - | โครงสร้างฝังแบบ inline |
| CreatedByAvatarThumb | string | - | createdbyavatarthumb | ค่าของ CreatedByAvatarThumb ตามฟิลด์ `createdbyavatarthumb` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[HoldingCodeentity]], [[productbarcode-ProductPriceHistory|ProductPriceHistory]]
