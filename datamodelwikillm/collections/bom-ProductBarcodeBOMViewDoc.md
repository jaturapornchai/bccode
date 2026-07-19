---
source: backend/internal/product/bom/models/bom.go
collection: productbarcodeboms
tags: [datamodel, mongodb, mongo-root]
---

# ProductBarcodeBOMViewDoc

โครงสร้าง `ProductBarcodeBOMViewDoc` เป็น root document ที่ repository โมดูล `bom` ใช้กับ MongoDB collection `productbarcodeboms`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| ProductBarcodeBOMViewData | [[bom-ProductBarcodeBOMViewData\|ProductBarcodeBOMViewData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|models.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[bom-ProductBarcodeBOMViewData|ProductBarcodeBOMViewData]], [[ActivityDoc]]
- MongoDB collection: `productbarcodeboms`
