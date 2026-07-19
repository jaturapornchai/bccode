---
source: backend/internal/models/name.go
tags: [datamodel, general-type, mongo-inline]
---

# Name

โครงสร้าง `Name` ที่ฝังแบบ inline ในโมเดล MongoDB ตามซอร์ส `name.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Name1 | string | name1 | name1 | ชื่อหลัก |
| Name2 | *string | name2 | name2,omitempty | ชื่อภาษา/รูปแบบที่ 2 |
| Name3 | *string | name3 | name3,omitempty | ชื่อภาษา/รูปแบบที่ 3 |
| Name4 | *string | name4 | name4,omitempty | ชื่อภาษา/รูปแบบที่ 4 |
| Name5 | *string | name5 | name5,omitempty | ชื่อภาษา/รูปแบบที่ 5 |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
