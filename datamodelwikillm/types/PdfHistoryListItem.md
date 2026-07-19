---
source: pdf-history-model.go
tags: [datamodel, general-type]
---

# PdfHistoryListItem

Item ในรายการประวัติ PDF พร้อม URL (ตาม comment ในโค้ด) — รูปแบบส่งออกของ [[PdfHistory]] ที่ตัด R2Key ออกและเพิ่ม presigned URL สำหรับดาวน์โหลดไฟล์

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | - | id | รหัสเอกสารใน MongoDB |
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| Collection | string | - | collection | collection ที่ดึงข้อมูลมาสร้าง PDF |
| DocNo | string | - | docno | เลขที่เอกสาร |
| Title | string | - | title | ชื่อเอกสาร |
| FileName | string | - | filename | ชื่อไฟล์ PDF |
| FileSize | int64 | - | filesize | ขนาดไฟล์ (bytes) |
| PageSize | string | - | pagesize | ขนาดกระดาษ เช่น A4, A3 |
| Orientation | string | - | orientation | แนวกระดาษ P / L |
| Language | string | - | language | ภาษาของเอกสาร เช่น th, en |
| ThemeName | string | - | themename | theme ที่ใช้พิมพ์ |
| TemplateID | string | - | templateid | template ที่ใช้พิมพ์ |
| PrintedBy | string | - | printedby | ผู้พิมพ์ |
| PrintedAt | time.Time | - | printedat | วันเวลาที่พิมพ์ |
| ReprintCount | int | - | reprintcount | จำนวนครั้งที่ reprint |
| URL | string | - | url,omitempty | Presigned URL สำหรับดาวน์โหลด (หมดอายุได้) |
| CreatedAt | time.Time | - | createdat | วันเวลาที่สร้าง record |

## ความสัมพันธ์

- map มาจาก [[PdfHistory]] (ตัด `R2Key`, `DocDate`, `CustomerName`, `VendorName`, `TotalAmount` ออก และเพิ่ม `URL`)
- `HoldingCode` — อ้างอิงรหัสกลุ่มกิจการ (tenant boundary)
- `Collection` + `DocNo` — อ้างอิงเอกสารต้นทาง
- ถูก embed เป็น slice ใน [[PdfHistoryListResponse]]
