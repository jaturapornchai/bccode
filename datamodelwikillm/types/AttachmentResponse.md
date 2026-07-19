---
source: attachment-model.go
tags: [datamodel, general-type]
---

# AttachmentResponse

Response มาตรฐานของ attachment API หนึ่งรายการ — มี status/code/message และข้อมูลไฟล์แนบใน `Data`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | `status` | สถานะผลลัพธ์ |
| Code | int | - | `code` | รหัสผลลัพธ์ |
| Message | string | - | `message,omitempty` | ข้อความประกอบ (ถ้ามี) |
| Data | *[[AttachmentMetadata]] | - | `data,omitempty` | ข้อมูลไฟล์แนบ |

## ความสัมพันธ์

- ฝัง [[AttachmentMetadata]] เป็น payload (`Data`)
