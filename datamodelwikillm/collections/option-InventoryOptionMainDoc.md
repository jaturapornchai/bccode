---
source: backend/internal/product/option/models/option.go
collection: inventoryOptions
tags: [datamodel, mongodb, mongo-root]
---

# InventoryOptionMainDoc

โครงสร้าง `InventoryOptionMainDoc` เป็น root document ที่ repository โมดูล `option` ใช้กับ MongoDB collection `inventoryOptions`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| InventoryOptionMainData | [[option-InventoryOptionMainData\|InventoryOptionMainData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|common.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[option-InventoryOptionMainData|InventoryOptionMainData]], [[ActivityDoc]]
- MongoDB collection: `inventoryOptions`
