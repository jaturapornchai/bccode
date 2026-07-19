---
source: attachment-model.go
collection: attachments
tags: [datamodel, mongodb, mongo-root]
---

# AttachmentMetadata

ข้อมูล metadata ของไฟล์แนบเอกสาร (เช่น PO, Sale, Purchase) เก็บใน collection `attachments` โดยตัวไฟล์จริงเก็บใน R2 bucket และอ้างถึงผ่าน `R2Key`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | `_id,omitempty` | `id` | ObjectID ของ document |
| HoldingCode | string | `holdingcode` | `holdingcode` | รหัสกลุ่มกิจการ (tenant) |
| ScreenType | string | `screentype` | `screentype` | ประเภทหน้าจอ/เอกสาร เช่น purchaseorder, sale, purchase |
| DocNo | string | `docno` | `docno` | เลขที่เอกสาร |
| GuidFixed | string | `guidfixed` | `guidfixed` | GUID ของเอกสาร |
| FileName | string | `filename` | `filename` | ชื่อไฟล์ใน R2 (timestamp_hash + extension) |
| OriginalName | string | `originalname` | `originalname` | ชื่อไฟล์เดิมตอน upload |
| ContentType | string | `contenttype` | `contenttype` | MIME type ของไฟล์ |
| FileType | string | `filetype` | `filetype` | ชนิดไฟล์ เช่น pdf, xlsx, jpg, png |
| Size | int64 | `size` | `size` | ขนาดไฟล์ (bytes) |
| R2Key | string | `r2key` | `-` | key ใน R2 bucket (ไม่ส่งออกทาง JSON) |
| Description | string | `description,omitempty` | `description,omitempty` | คำอธิบายไฟล์แนบ |
| UploadedBy | string | `uploadedby` | `uploadedby` | usercode ผู้ upload |
| UploadedName | string | `uploadedname` | `uploadedname` | ชื่อผู้ upload |
| CreatedAt | time.Time | `createdat` | `createdat` | เวลาสร้าง |
| UpdatedAt | time.Time | `updatedat` | `updatedat` | เวลาแก้ไขล่าสุด |

## ความสัมพันธ์

- `HoldingCode` — อ้างถึงกลุ่มกิจการ (tenant boundary)
- `DocNo` + `GuidFixed` + `ScreenType` — อ้างถึงเอกสารต้นทางที่ไฟล์แนบสังกัด (เช่นเอกสาร PO/Sale/Purchase)
- ถูกใช้เป็น payload ใน [[AttachmentResponse]] (field `Data`)
