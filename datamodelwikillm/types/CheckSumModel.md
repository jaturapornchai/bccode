---
source: backend/internal/goapi/models/global-model.go
tags: [datamodel, general-type]
---

# CheckSumModel

โครงสร้างเก็บค่า checksum ของข้อมูล อ้างอิงด้วย RefCode พร้อมระบุชื่อฐานข้อมูล MongoDB ที่เกี่ยวข้อง

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| RefCode | string | - | refcode | รหัสอ้างอิงของข้อมูลที่ทำ checksum |
| Checksum | string | - | checksum | ค่า checksum ของข้อมูล |
| MongoDbName | string | - | mongodbname | ชื่อฐานข้อมูล MongoDB ที่ข้อมูลอยู่ |

## ความสัมพันธ์
- `RefCode` — รหัสอ้างอิงไปยังข้อมูลต้นทางที่ถูกคำนวณ checksum (ไม่ระบุ collection ชัดเจนในโค้ด)
