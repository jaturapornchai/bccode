---
source: image-model.go
tags: [datamodel, general-type]
---

# ImageListDataItem

Item ใน list รูปภาพพร้อม URL หรือ base64 — เป็น projection ของ [[ImageMetadata]] ที่ใช้ fields แยกแทนการ embed เพื่อให้ json marshal ทำงานถูกต้อง (ฟิลด์ตรวจสลิปเป็น pointer เพื่อแยก "ยังไม่ตรวจ" = null)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | - | id | MongoDB ObjectID |
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| FileName | string | - | filename | ชื่อไฟล์ใน R2 |
| OriginalName | string | - | originalname | ชื่อไฟล์เดิม |
| ContentType | string | - | contenttype | MIME type |
| Size | int64 | - | size | ขนาดไฟล์ (bytes) |
| Category | string | - | category,omitempty | หมวดหมู่รูปภาพ |
| Description | string | - | description,omitempty | คำอธิบายรูปภาพ |
| Tags | []string | - | tags,omitempty | tag ของรูปภาพ |
| UploadedBy | string | - | uploadedby,omitempty | ผู้อัปโหลด |
| CreatedAt | time.Time | - | createdat | เวลาสร้าง |
| UpdatedAt | time.Time | - | updatedat | เวลาแก้ไขล่าสุด |
| URL | string | - | url,omitempty | Private backend URL ของรูป |
| Data | string | - | data,omitempty | base64 data URL (deprecated, ใช้ url แทน) |
| Verified | *bool | - | verified | true = ตรวจสอบแล้ว, null = ยังไม่ตรวจสอบ |
| VerifyStatus | *string | - | verifystatus | สถานะตรวจสอบ: "success", "duplicate", "not_found", "error" |
| IsDuplicate | *bool | - | isduplicate | true = สลิปซ้ำ |
| TransRef | *string | - | transref,omitempty | หมายเลขอ้างอิงการโอน |
| TransAmount | *float64 | - | transamount,omitempty | จำนวนเงินที่โอน |
| SenderName | *string | - | sendername,omitempty | ชื่อผู้โอน |
| ReceiverName | *string | - | receivername,omitempty | ชื่อผู้รับ |
| SlipType | *string | - | sliptype,omitempty | bank หรือ truewallet |

## ความสัมพันธ์

- projection ของ [[ImageMetadata]] (fields แยก ไม่ embed)
- `HoldingCode` — อ้างอิงกลุ่มกิจการ (tenant) เจ้าของรูป
- ถูกใช้ใน [[ImageListDataResponse]]
