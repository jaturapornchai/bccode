---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageDeleteRequest

Request body สำหรับลบรูปภาพ ระบุได้ทั้ง ImageID หรือ FileName

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| ImageID | string | - | imageid | id ของรูปที่จะลบ |
| FileName | string | - | filename | ใช้ filename แทน image_id ได้ |

## ความสัมพันธ์

- `HoldingCode` — อ้างอิงกลุ่มกิจการ (tenant) เจ้าของรูป
- `ImageID` / `FileName` — อ้างอิง [[ImageMetadata]] ที่จะลบ
