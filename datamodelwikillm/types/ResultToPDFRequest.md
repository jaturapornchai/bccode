---
source: query-result-model.go
tags: [datamodel, general-type]
---

# ResultToPDFRequest

Request model สำหรับ endpoint `/resulttopdf` — สั่งสร้าง PDF จากผลลัพธ์ query ที่เก็บไว้ตาม GUID พร้อมกำหนดลำดับคอลัมน์และชื่อคอลัมน์ภาษาไทย

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัส holding (บังคับ) |
| GUID | string | - | guid | GUID ของผลลัพธ์ query (บังคับ) |
| PDFConfig | [[PDFConfig]] | - | pdfconfig | การตั้งค่า PDF |
| ColumnOrder | []string | - | columnorder | ลำดับคอลัมน์ใน PDF |
| ColumnNames | map[string]string | - | columnnames | ชื่อคอลัมน์ภาษาไทย |

## ความสัมพันธ์

- ฝัง [[PDFConfig]] เป็นการตั้งค่าเอกสาร
- `HoldingCode` — อ้างอิงกลุ่มกิจการ (holding) เป็นขอบเขต tenant
- `GUID` — อ้างอิงผลลัพธ์ query ที่สร้างจาก [[ResultFromQueryRequest]]
