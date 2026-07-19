---
source: backend/internal/product/promotion/models/promotion.go
tags: [datamodel, general-type]
---

# PromotionData

โครงสร้าง `PromotionData` จากโมดูล mainapi `promotion` มีฟิลด์ตามซอร์ส `promotion.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| PromotionInfo | [[promotion-PromotionInfo\|PromotionInfo]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[promotion-PromotionInfo|PromotionInfo]]
