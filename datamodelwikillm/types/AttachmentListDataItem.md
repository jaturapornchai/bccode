---
source: attachment-model.go
tags: [datamodel, general-type]
---

# AttachmentListDataItem

รายการไฟล์แนบหนึ่งรายการสำหรับ list response — โครงเหมือน AttachmentMetadata แต่ตัด R2Key ออกและเพิ่ม `URL` (private backend URL) สำหรับให้ client เข้าถึงไฟล์

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | - | `id` | ObjectID ของไฟล์แนบ |
| HoldingCode | string | - | `holdingcode` | รหัสกลุ่มกิจการ (tenant) |
| ScreenType | string | - | `screentype` | ประเภทหน้าจอ/เอกสาร |
| DocNo | string | - | `docno` | เลขที่เอกสาร |
| GuidFixed | string | - | `guidfixed` | GUID ของเอกสาร |
| FileName | string | - | `filename` | ชื่อไฟล์ใน R2 |
| OriginalName | string | - | `originalname` | ชื่อไฟล์เดิม |
| ContentType | string | - | `contenttype` | MIME type |
| FileType | string | - | `filetype` | ชนิดไฟล์ เช่น pdf, xlsx, jpg, png |
| Size | int64 | - | `size` | ขนาดไฟล์ (bytes) |
| Description | string | - | `description,omitempty` | คำอธิบายไฟล์แนบ |
| UploadedBy | string | - | `uploadedby` | usercode ผู้ upload |
| UploadedName | string | - | `uploadedname` | ชื่อผู้ upload |
| CreatedAt | time.Time | - | `createdat` | เวลาสร้าง |
| UpdatedAt | time.Time | - | `updatedat` | เวลาแก้ไขล่าสุด |
| URL | string | - | `url,omitempty` | Private backend URL สำหรับเข้าถึงไฟล์ |

## ความสัมพันธ์

- แปลงมาจาก [[AttachmentMetadata]] (collection `attachments`) — ใช้เป็น item ใน [[AttachmentListDataResponse]]
- `HoldingCode` — อ้างถึงกลุ่มกิจการ · `DocNo`/`GuidFixed` — อ้างถึงเอกสารต้นทาง
