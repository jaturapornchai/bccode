---
source: image-model.go
collection: images
tags: [datamodel, mongodb, mongo-root]
---

# ImageMetadata

ข้อมูล metadata ของรูปภาพที่เก็บใน MongoDB (collection `images`) รวมข้อมูลไฟล์ใน R2 storage และผลการตรวจสอบสลิปโอนเงินผ่าน Thunder API — `R2Key` เก็บใน DB แต่ไม่ส่งออกไป frontend (`json:"-"`)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | MongoDB ObjectID |
| HoldingCode | string | holdingcode | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| FileName | string | filename | filename | ชื่อไฟล์ใน R2 (hash + extension) |
| OriginalName | string | originalname | originalname | ชื่อไฟล์เดิมตอน upload |
| ContentType | string | contenttype | contenttype | MIME type |
| Size | int64 | size | size | ขนาดไฟล์ (bytes) |
| R2Key | string | r2key | - | key ใน R2 bucket (ไม่ส่งออก frontend) |
| Category | string | category,omitempty | category,omitempty | หมวดหมู่รูปภาพ |
| Description | string | description,omitempty | description,omitempty | คำอธิบายรูปภาพ |
| Tags | []string | tags,omitempty | tags,omitempty | tag ของรูปภาพ |
| UploadedBy | string | uploadedby,omitempty | uploadedby,omitempty | ผู้อัปโหลด |
| CreatedAt | time.Time | createdat | createdat | เวลาสร้าง |
| UpdatedAt | time.Time | updatedat | updatedat | เวลาแก้ไขล่าสุด |
| Verified | bool | verified | verified | ผ่านการตรวจสอบสลิปหรือไม่ |
| VerifiedAt | *time.Time | verifiedat,omitempty | verifiedat,omitempty | เวลาที่ตรวจสอบ |
| VerifyStatus | string | verifystatus,omitempty | verifystatus,omitempty | สถานะตรวจสอบ: success, error, not_found |
| VerifyError | string | verifyerror,omitempty | verifyerror,omitempty | error message ถ้ามี |
| SlipData | map[string]interface{} | slipdata,omitempty | slipdata,omitempty | ข้อมูล slip ดิบจาก Thunder API |
| TransRef | string | transref,omitempty | transref,omitempty | หมายเลขอ้างอิงการโอน (Transaction Reference) |
| TransAmount | float64 | transamount,omitempty | transamount,omitempty | จำนวนเงินที่โอน |
| TransDate | string | transdate,omitempty | transdate,omitempty | วันที่ทำรายการ |
| SenderName | string | sendername,omitempty | sendername,omitempty | ชื่อผู้โอน |
| SenderBank | string | senderbank,omitempty | senderbank,omitempty | ธนาคารผู้โอน |
| ReceiverName | string | receivername,omitempty | receivername,omitempty | ชื่อผู้รับ |
| ReceiverBank | string | receiverbank,omitempty | receiverbank,omitempty | ธนาคารผู้รับ |
| IsDuplicate | bool | isduplicate,omitempty | isduplicate,omitempty | เป็นสลิปซ้ำหรือไม่ |
| SlipType | string | sliptype,omitempty | sliptype,omitempty | ประเภทสลิป: bank หรือ truewallet |

## ความสัมพันธ์

- `HoldingCode` — อ้างอิงกลุ่มกิจการ (tenant) เจ้าของรูปภาพ
- ถูกใช้เป็น payload ใน [[ImageResponse]], [[ImageDataResponse]], [[ImageListResponse]], [[ImageListResponseWithTotal]], [[ImageVerifyResponse]]
- [[ImageListDataItem]] เป็น projection ของ struct นี้สำหรับ list response (fields แยก ไม่ embed)
