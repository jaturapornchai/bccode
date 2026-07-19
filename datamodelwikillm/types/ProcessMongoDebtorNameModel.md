---
source: mongo-debtor-model.go
tags: [datamodel, general-type]
---

# ProcessMongoDebtorNameModel

โมเดลย่อยเก็บชื่อของลูกหนี้ 1 รายการ ใช้ฝังเป็น list ใน [[ProcessMongoDebtorModel]]

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Name | string | name | name | ชื่อลูกหนี้ |

## ความสัมพันธ์

- ถูกฝังเป็น `[]` ใน field `Names` ของ [[ProcessMongoDebtorModel]]
