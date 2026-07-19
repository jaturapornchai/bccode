---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# MongoDocReferenceModel

ข้อมูลอ้างอิงเอกสารอื่น (เลขที่ + วันเวลา) ใช้ฝังเป็น array ใน [[MongoDocModel]] มี type alias `ProcessMongoDocReferenceModel` ชี้มาที่ struct นี้

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocNo | string | docno | docno | เลขที่เอกสารที่อ้างอิง |
| DocDateTime | time.Time | docdatetime | docdatetime | วันเวลาเอกสารที่อ้างอิง |

## ความสัมพันธ์
- ถูกฝังเป็น `DocReferences` ใน [[MongoDocModel]]
- DocNo — อ้างอิงเลขที่เอกสารอื่น
