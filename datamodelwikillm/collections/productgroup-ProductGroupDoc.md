---
source: backend/internal/product/productgroup/models/productgroup.go
collection: productgroups
tags: [datamodel, mongodb, mongo-root]
---

# ProductGroupDoc

โครงสร้าง `ProductGroupDoc` เป็น root document ที่ repository โมดูล `productgroup` ใช้กับ MongoDB collection `productgroups`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| ProductGroupData | [[productgroup-ProductGroupData\|ProductGroupData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|models.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productgroup-ProductGroupData|ProductGroupData]], [[ActivityDoc]]
- MongoDB collection: `productgroups`
