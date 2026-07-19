---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageListResponse

Response สำหรับรายการรูปภาพ (metadata อย่างเดียว ไม่มี total)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | status | สถานะผลลัพธ์ |
| Code | int | - | code | รหัสผลลัพธ์ |
| Count | int | - | count | จำนวนรายการที่คืนมา |
| Data | [][[ImageMetadata]] | - | data | รายการ metadata รูปภาพ |

## ความสัมพันธ์

- `Data` — embed [][[ImageMetadata]]
- คู่กับ request [[ImageListRequest]]
