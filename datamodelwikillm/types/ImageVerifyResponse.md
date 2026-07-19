---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageVerifyResponse

Response สำหรับผลการตรวจสอบสลิปโอนเงิน คืน [[ImageMetadata]] ที่อัปเดตผลตรวจแล้ว

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | status | สถานะผลลัพธ์ |
| Code | int | - | code | รหัสผลลัพธ์ |
| Message | string | - | message,omitempty | ข้อความประกอบ |
| Data | *[[ImageMetadata]] | - | data,omitempty | metadata ของรูปพร้อมผลตรวจสลิป |

## ความสัมพันธ์

- `Data` — embed *[[ImageMetadata]]
- คู่กับ request [[ImageVerifyRequest]]
