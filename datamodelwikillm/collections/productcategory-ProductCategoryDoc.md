---
source: backend/internal/product/productcategory/models/productcategory.go
collection: productcategories
tags: [datamodel, mongodb, mongo-root]
---

# ProductCategoryDoc

โครงสร้าง `ProductCategoryDoc` เป็น root document ที่ repository โมดูล `productcategory` ใช้กับ MongoDB collection `productcategories`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| ProductCategoryData | [[productcategory-ProductCategoryData\|ProductCategoryData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|models.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productcategory-ProductCategoryData|ProductCategoryData]], [[ActivityDoc]]
- MongoDB collection: `productcategories`
