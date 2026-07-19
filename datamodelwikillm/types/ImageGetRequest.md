---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageGetRequest

Request body สำหรับดึงรูปภาพรายตัวด้วยชื่อไฟล์

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| FileName | string | - | filename | ชื่อไฟล์ของรูปที่ต้องการ |

## ความสัมพันธ์

- `HoldingCode` — อ้างอิงกลุ่มกิจการ (tenant) เจ้าของรูป
- `FileName` — อ้างอิง [[ImageMetadata]] ที่ต้องการดึง
