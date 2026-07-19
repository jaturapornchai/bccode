---
source: backend/internal/product/optionpattern/models/optionpattern.go
collection: optionpattern
tags: [datamodel, mongodb, mongo-root]
---

# OptionPatternDoc

โครงสร้าง `OptionPatternDoc` เป็น root document ที่ repository โมดูล `optionpattern` ใช้กับ MongoDB collection `optionpattern`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| OptionPatternData | [[optionpattern-OptionPatternData\|OptionPatternData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|models.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[optionpattern-OptionPatternData|OptionPatternData]], [[ActivityDoc]]
- MongoDB collection: `optionpattern`
