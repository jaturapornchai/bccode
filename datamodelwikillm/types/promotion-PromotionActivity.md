---
source: backend/internal/product/promotion/models/promotion.go
tags: [datamodel, general-type]
---

# PromotionActivity

โครงสร้าง `PromotionActivity` จากโมดูล mainapi `promotion` มีฟิลด์ตามซอร์ส `promotion.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| PromotionData | [[promotion-PromotionData\|PromotionData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityTime | [[ActivityTime\|models.ActivityTime]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[promotion-PromotionData|PromotionData]], [[ActivityTime]]
