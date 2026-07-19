---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageResponse

Response เมื่ออัปโหลดรูปภาพสำเร็จ ห่อ [[ImageMetadata]] พร้อม status/code

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | status | สถานะผลลัพธ์ |
| Code | int | - | code | รหัสผลลัพธ์ |
| Message | string | - | message,omitempty | ข้อความประกอบ |
| Data | *[[ImageMetadata]] | - | data,omitempty | metadata ของรูปภาพ |

## ความสัมพันธ์

- `Data` — embed *[[ImageMetadata]]
