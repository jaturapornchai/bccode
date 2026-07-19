---
source: backend/internal/product/ordertype/models/ordertype.go
collection: productordertypes
tags: [datamodel, mongodb, mongo-root]
---

# OrderTypeDoc

โครงสร้าง `OrderTypeDoc` เป็น root document ที่ repository โมดูล `ordertype` ใช้กับ MongoDB collection `productordertypes`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| OrderTypeData | [[ordertype-OrderTypeData\|OrderTypeData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|models.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[ordertype-OrderTypeData|OrderTypeData]], [[ActivityDoc]]
- MongoDB collection: `productordertypes`
