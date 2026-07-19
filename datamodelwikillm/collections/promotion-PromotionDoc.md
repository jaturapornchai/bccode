---
source: backend/internal/product/promotion/models/promotion.go
collection: productpromotions
tags: [datamodel, mongodb, mongo-root]
---

# PromotionDoc

โครงสร้าง `PromotionDoc` เป็น root document ที่ repository โมดูล `promotion` ใช้กับ MongoDB collection `productpromotions`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| PromotionData | [[promotion-PromotionData\|PromotionData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|models.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[promotion-PromotionData|PromotionData]], [[ActivityDoc]]
- MongoDB collection: `productpromotions`
