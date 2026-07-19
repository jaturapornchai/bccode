---
source: backend/internal/models/name.go
tags: [datamodel, general-type, mongo-inline]
---

# UnitName

โครงสร้าง `UnitName` ที่ฝังแบบ inline ในโมเดล MongoDB ตามซอร์ส `name.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| UnitName1 | string | unitname1 | unitname1 | ชื่อหน่วยหลัก |
| UnitName2 | *string | unitname2,omitempty | unitname2,omitempty | ชื่อหน่วยภาษา/รูปแบบที่ 2 |
| UnitName3 | *string | unitname3,omitempty | unitname3,omitempty | ชื่อหน่วยภาษา/รูปแบบที่ 3 |
| UnitName4 | *string | unitname4,omitempty | unitname4,omitempty | ชื่อหน่วยภาษา/รูปแบบที่ 4 |
| UnitName5 | *string | unitname5,omitempty | unitname5,omitempty | ชื่อหน่วยภาษา/รูปแบบที่ 5 |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
