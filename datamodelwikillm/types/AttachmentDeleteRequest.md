---
source: attachment-model.go
tags: [datamodel, general-type]
---

# AttachmentDeleteRequest

Request สำหรับลบไฟล์แนบ ระบุด้วย AttachmentID หรือ FileName ภายใต้ holding เดียวกัน

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | `holdingcode` | รหัสกลุ่มกิจการ (tenant) |
| AttachmentID | string | - | `attachmentid,omitempty` | ObjectID ของไฟล์แนบที่จะลบ |
| FileName | string | - | `filename,omitempty` | ชื่อไฟล์ใน R2 ที่จะลบ (ทางเลือกแทน AttachmentID) |

## ความสัมพันธ์

- อ้างถึง [[AttachmentMetadata]] ใน collection `attachments` ผ่าน `AttachmentID` (`_id`) หรือ `FileName`
- `HoldingCode` — อ้างถึงกลุ่มกิจการ
