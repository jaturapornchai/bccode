---
source: process-doc-model.go
tags: [datamodel, general-type]
---

# DocRefStruct

โครงสร้างการอ้างอิงระหว่างเอกสาร (document reference) เก็บคู่เลขที่เอกสารต้นทางกับเลขที่เอกสารอ้างอิง พร้อม trans flag ของแต่ละฝั่ง

หมายเหตุ: struct นี้ใช้ tag `json` + `db` (ไม่มี tag `bson`)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocNo | string | - | docno | เลขที่เอกสาร |
| DocNoTransFlag | int | - | docnotransflag | trans flag ของเอกสาร |
| DocRefNo | string | - | refdocno | เลขที่เอกสารอ้างอิง |
| DocRefNoTransFlag | int | - | refdocnotransflag | trans flag ของเอกสารอ้างอิง |

## ความสัมพันธ์

- `DocNo` — อ้างถึงเอกสารหลัก (ดู [[DocStruct]])
- `DocRefNo` — อ้างถึงเลขที่เอกสารอื่นที่ถูกอ้างอิง
