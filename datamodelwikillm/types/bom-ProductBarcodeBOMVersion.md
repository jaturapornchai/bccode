---
source: backend/internal/product/bom/models/bom.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMVersion

โครงสร้าง `ProductBarcodeBOMVersion` จากโมดูล mainapi `bom` มีฟิลด์ตามซอร์ส `bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GuidFixed | string | guidfixed | guidfixed | GUID ถาวรของรายการ |
| StartDate | time.Time | startdate | startdate | วันที่เริ่มต้น |
| EndDate | *time.Time | enddate | enddate | วันที่สิ้นสุด |
| BOM | *[][[bom-ProductBarcodeBOMView\|ProductBarcodeBOMView]] | bom | bom | ค่าของ BOM ตามฟิลด์ `bom` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[bom-ProductBarcodeBOMView|ProductBarcodeBOMView]]
