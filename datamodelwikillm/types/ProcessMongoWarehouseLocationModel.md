---
source: mongo-warehouse-model.go
tags: [datamodel, general-type]
---

# ProcessMongoWarehouseLocationModel

โมเดลข้อมูลที่เก็บ (location) ภายในคลังสินค้า ใช้เป็น embedded struct ใน [[ProcessMongoWarehouseModel]] ประกอบด้วยรหัส location และชื่อหลายภาษา

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Code | string | code | code | รหัส location ภายในคลัง |
| Names | []​[[LanguageModel]] | names | names | ชื่อ location หลายภาษา |

## ความสัมพันธ์

- Embed [[LanguageModel]] — ชื่อหลายภาษาใน `Names`
- ถูก embed อยู่ใน [[ProcessMongoWarehouseModel]] (field `Location`)
