---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageDataResponse

Response ดึงรูปภาพพร้อมข้อมูลรูปแบบ base64

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | status | สถานะผลลัพธ์ |
| Code | int | - | code | รหัสผลลัพธ์ |
| Message | string | - | message,omitempty | ข้อความประกอบ |
| Data | *[[ImageMetadata]] | - | data,omitempty | metadata ของรูปภาพ |
| Base64 | string | - | base64,omitempty | ข้อมูลรูปภาพ encode เป็น base64 |

## ความสัมพันธ์

- `Data` — embed *[[ImageMetadata]]
