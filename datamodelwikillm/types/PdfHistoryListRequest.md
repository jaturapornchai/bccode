---
source: pdf-history-model.go
tags: [datamodel, general-type]
---

# PdfHistoryListRequest

Request สำหรับดึงรายการประวัติ PDF (ตาม comment ในโค้ด) รองรับ filter ตาม collection/เลขที่เอกสาร และ pagination ด้วย limit/skip

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| Collection | string | - | collection | filter ตาม collection (optional) |
| DocNo | string | - | docno | filter ตามเลขที่เอกสาร (optional) |
| Limit | int64 | - | limit | จำนวนรายการต่อหน้า |
| Skip | int64 | - | skip | จำนวนรายการที่ข้าม (pagination) |

## ความสัมพันธ์

- `HoldingCode` — อ้างอิงรหัสกลุ่มกิจการ (tenant boundary)
- `Collection` + `DocNo` — เงื่อนไขค้นหา record ใน [[PdfHistory]]
- ใช้คู่กับ [[PdfHistoryListResponse]] เป็น request/response ของ API รายการประวัติ PDF
