---
source: query-result-model.go
tags: [datamodel, general-type]
---

# PaginationInfo

ข้อมูลการแบ่งหน้า (pagination) ของผลลัพธ์ query ใช้ฝังใน [[ResultGetResponse]]

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Total | int | - | total | จำนวนแถวทั้งหมด |
| Limit | int | - | limit | จำนวนแถวต่อหน้า |
| Offset | int | - | offset | offset ปัจจุบัน |
| Count | int | - | count | จำนวนแถวในหน้าปัจจุบัน |

## ความสัมพันธ์

- ถูกฝังเป็น field `Pagination` ใน [[ResultGetResponse]]
