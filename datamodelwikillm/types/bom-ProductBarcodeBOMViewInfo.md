---
source: backend/internal/product/bom/models/bom.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMViewInfo

โครงสร้าง `ProductBarcodeBOMViewInfo` จากโมดูล mainapi `bom` มีฟิลด์ตามซอร์ส `bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductBarcodeBOMView | [[bom-ProductBarcodeBOMView\|ProductBarcodeBOMView]] | inline | - | โครงสร้างฝังแบบ inline |
| CheckSum | string | checksum | checksum | ค่าของ CheckSum ตามฟิลด์ `checksum` ในซอร์ส |
| IsCurrentUse | bool | iscurrentuse | iscurrentuse | ค่าของ IsCurrentUse ตามฟิลด์ `iscurrentuse` ในซอร์ส |
| UseInDate | time.Time | useindate | useindate | ค่าของ UseInDate ตามฟิลด์ `useindate` ในซอร์ส |
| StartDate | time.Time | startdate | startdate | วันที่เริ่มต้น |
| EndDate | *time.Time | enddate | enddate | วันที่สิ้นสุด |
| BOMs | *[][[bom-ProductBarcodeBOMVersion\|ProductBarcodeBOMVersion]] | boms,omitempty | boms | ค่าของ BOMs ตามฟิลด์ `boms` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[bom-ProductBarcodeBOMView|ProductBarcodeBOMView]], [[bom-ProductBarcodeBOMVersion|ProductBarcodeBOMVersion]]
