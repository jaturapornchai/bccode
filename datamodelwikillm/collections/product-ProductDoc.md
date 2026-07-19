---
source: backend/internal/product/product/models/product.go
collection: products
tags: [datamodel, mongodb, mongo-root]
---

# ProductDoc

โครงสร้าง `ProductDoc` เป็น root document ที่ repository โมดูล `product` ใช้กับ MongoDB collection `products`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| ProductData | [[product-ProductData\|ProductData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|models.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[product-ProductData|ProductData]], [[ActivityDoc]]
- MongoDB collection: `products`
