---
source: attachment-model.go
tags: [datamodel, general-type]
---

# AttachmentUploadRequest

Request สำหรับ upload ไฟล์แนบผ่าน multipart form data (file, holdingcode, screentype, docno, guidfixed, description, uploadedby, uploadedname)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| File | string | - | - (form: `file`) | multipart file ที่ upload |
| HoldingCode | string | - | - (form: `holdingcode`) | รหัสกลุ่มกิจการ (tenant) |
| ScreenType | string | - | - (form: `screentype`) | ประเภทหน้าจอ/เอกสาร |
| DocNo | string | - | - (form: `docno`) | เลขที่เอกสาร |
| GuidFixed | string | - | - (form: `guidfixed`) | GUID ของเอกสาร |
| Description | string | - | - (form: `description`) | คำอธิบายไฟล์แนบ |
| UploadedBy | string | - | - (form: `uploadedby`) | usercode ผู้ upload |
| UploadedName | string | - | - (form: `uploadedname`) | ชื่อผู้ upload |

## ความสัมพันธ์

- ข้อมูลจาก request นี้ถูกใช้สร้าง [[AttachmentMetadata]] ใน collection `attachments`
- `HoldingCode` — อ้างถึงกลุ่มกิจการ · `DocNo`/`GuidFixed` — อ้างถึงเอกสารต้นทาง
