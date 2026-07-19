---
source: pdf-history-model.go
collection: pdfhistory
tags: [datamodel, mongodb, mongo-root]
---

# PdfHistory

โมเดลเก็บประวัติการพิมพ์เอกสาร PDF (ตาม comment ในโค้ด) บันทึกข้อมูลเอกสารที่พิมพ์ ไฟล์ PDF ที่สร้าง และรายละเอียดการพิมพ์ เช่น theme, template, ผู้พิมพ์, จำนวนครั้งที่ reprint

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id,omitempty | รหัสเอกสารใน MongoDB |
| HoldingCode | string | holdingcode | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| Collection | string | collection | collection | collection ที่ดึงข้อมูลมาสร้าง PDF |
| DocNo | string | docno | docno | เลขที่เอกสาร |
| DocDate | time.Time | docdate | docdate | วันที่เอกสาร |
| Title | string | title | title | ชื่อเอกสาร |
| FileName | string | filename | filename | ชื่อไฟล์ PDF |
| R2Key | string | r2key | - | path ของไฟล์ใน R2 (ไม่ส่งออก frontend) |
| FileSize | int64 | filesize | filesize | ขนาดไฟล์ (bytes) |
| PageSize | string | pagesize | pagesize | ขนาดกระดาษ เช่น A4, A3 |
| Orientation | string | orientation | orientation | แนวกระดาษ P (แนวตั้ง) / L (แนวนอน) |
| Language | string | language | language | ภาษาของเอกสาร เช่น th, en |
| ThemeName | string | themename | themename | theme ที่ใช้พิมพ์ |
| TemplateID | string | templateid | templateid | template ที่ใช้พิมพ์ |
| PrintedBy | string | printedby | printedby | ผู้พิมพ์ (optional) |
| PrintedAt | time.Time | printedat | printedat | วันเวลาที่พิมพ์ |
| ReprintCount | int | reprintcount | reprintcount | จำนวนครั้งที่ reprint |
| CreatedAt | time.Time | createdat | createdat | วันเวลาที่สร้าง record |
| CustomerName | string | customername | customername | ชื่อลูกหนี้/ลูกค้า (AR) |
| VendorName | string | vendorname | vendorname | ชื่อเจ้าหนี้/ผู้ขาย (AP) |
| TotalAmount | float64 | totalamount | totalamount | ยอดรวมของเอกสาร |

## ความสัมพันธ์

- `HoldingCode` — อ้างอิงรหัสกลุ่มกิจการ (tenant boundary)
- `Collection` + `DocNo` — อ้างอิงเอกสารต้นทางใน collection ที่ระบุ ด้วยเลขที่เอกสาร
- ใช้เป็นข้อมูลต้นทางของ [[PdfHistoryListItem]] ในการตอบรายการประวัติ PDF
