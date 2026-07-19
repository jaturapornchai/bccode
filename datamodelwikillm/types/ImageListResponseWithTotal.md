---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageListResponseWithTotal

Response สำหรับรายการรูปภาพ (metadata) พร้อมจำนวนทั้งหมด (total) สำหรับ pagination

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | status | สถานะผลลัพธ์ |
| Code | int | - | code | รหัสผลลัพธ์ |
| Count | int | - | count | จำนวนรายการที่คืนมา |
| Total | int64 | - | total | จำนวนทั้งหมด (ไม่รวม limit) |
| Data | [][[ImageMetadata]] | - | data | รายการ metadata รูปภาพ |

## ความสัมพันธ์

- `Data` — embed [][[ImageMetadata]]
- คู่กับ request [[ImageListRequest]]
