---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageVerifyRequest

Request body สำหรับตรวจสอบสลิปโอนเงิน (verify slip) ระบุรูปด้วย ImageID หรือ FileName

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| ImageID | string | - | imageid | id ของรูปสลิป (หรือใช้ filename) |
| FileName | string | - | filename | ใช้แทน image_id ได้ |
| CheckDuplicate | bool | - | checkduplicate | ตรวจสอบสลิปซ้ำด้วยหรือไม่ |
| Type | string | - | type | "bank" (default) หรือ "truewallet" |

## ความสัมพันธ์

- `HoldingCode` — อ้างอิงกลุ่มกิจการ (tenant) เจ้าของรูป
- `ImageID` / `FileName` — อ้างอิง [[ImageMetadata]] ที่จะตรวจสอบ
- ผลลัพธ์คืนเป็น [[ImageVerifyResponse]]
