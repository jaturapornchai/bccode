---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageUploadRequest

Request body สำหรับอัปโหลดรูปภาพ (รับได้ทั้ง form และ json)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| Category | string | - | category | หมวดหมู่รูปภาพ |
| Description | string | - | description | คำอธิบายรูปภาพ |
| Tags | []string | - | tags | tag ของรูปภาพ |
| UploadedBy | string | - | uploadedby | ผู้อัปโหลด |

## ความสัมพันธ์

- `HoldingCode` — อ้างอิงกลุ่มกิจการ (tenant) ปลายทางของรูปที่อัปโหลด
- ผลลัพธ์การอัปโหลดถูกเก็บเป็น [[ImageMetadata]]
