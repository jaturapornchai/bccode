---
source: attachment-model.go
tags: [datamodel, general-type]
---

# AttachmentListRequest

Request สำหรับดึงรายการไฟล์แนบ กรองตาม holding/screen/เอกสาร พร้อม pagination (limit/skip)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | `holdingcode` | รหัสกลุ่มกิจการ (tenant) |
| ScreenType | string | - | `screentype,omitempty` | กรองตามประเภทหน้าจอ/เอกสาร |
| DocNo | string | - | `docno,omitempty` | กรองตามเลขที่เอกสาร |
| GuidFixed | string | - | `guidfixed,omitempty` | กรองตาม GUID ของเอกสาร |
| Limit | int64 | - | `limit,omitempty` | จำนวนรายการต่อหน้า |
| Skip | int64 | - | `skip,omitempty` | จำนวนรายการที่ข้าม (pagination) |

## ความสัมพันธ์

- ใช้ query [[AttachmentMetadata]] ใน collection `attachments` — ผลลัพธ์ตอบกลับเป็น [[AttachmentListDataResponse]]
- `HoldingCode` — อ้างถึงกลุ่มกิจการ · `DocNo`/`GuidFixed` — อ้างถึงเอกสารต้นทาง
