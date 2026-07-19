---
source: mongo-creditor-model.go
tags: [datamodel, general-type]
---

# ProcessMongoCreditorNameModel

โมเดลย่อยเก็บชื่อเจ้าหนี้หนึ่งรายการ ใช้เป็นสมาชิกใน field `Names` ของ [[ProcessMongoCreditorModel]]

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Name | string | name | name | ชื่อเจ้าหนี้ |

## ความสัมพันธ์

- ถูก embed เป็น `[]` ใน [[ProcessMongoCreditorModel]] (field `Names`)
