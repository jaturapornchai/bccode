---
source: backend/internal/product/promotion/models/promotion.go
tags: [datamodel, general-type]
---

# PromotionInfo

โครงสร้าง `PromotionInfo` จากโมดูล mainapi `promotion` มีฟิลด์ตามซอร์ส `promotion.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Promotion | [[promotion-Promotion\|Promotion]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[promotion-Promotion|Promotion]]
