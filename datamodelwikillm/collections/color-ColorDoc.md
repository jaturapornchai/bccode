---
source: backend/internal/product/color/models/color.go
collection: colors
tags: [datamodel, mongodb, mongo-root]
---

# ColorDoc

โครงสร้าง `ColorDoc` เป็น root document ที่ repository โมดูล `color` ใช้กับ MongoDB collection `colors`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| ColorData | [[color-ColorData\|ColorData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|models.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[color-ColorData|ColorData]], [[ActivityDoc]]
- MongoDB collection: `colors`
