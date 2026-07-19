---
source: attachment-model.go
tags: [datamodel, general-type]
---

# AttachmentListDataResponse

Response สำหรับ list attachments — มี status/code, จำนวนรายการ (count/total) และรายการไฟล์แนบใน `Data`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Status | string | - | `status` | สถานะผลลัพธ์ |
| Code | int | - | `code` | รหัสผลลัพธ์ |
| Count | int | - | `count` | จำนวนรายการในหน้านี้ |
| Total | int64 | - | `total` | จำนวนรายการทั้งหมด |
| Data | [][[AttachmentListDataItem]] | - | `data` | รายการไฟล์แนบ |

## ความสัมพันธ์

- ฝัง slice ของ [[AttachmentListDataItem]] เป็น payload — เป็นผลลัพธ์ของ [[AttachmentListRequest]]
