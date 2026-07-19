---
source: backend/internal/product/productbarcode/models/product_bom.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMViewInfo

โครงสร้าง `ProductBarcodeBOMViewInfo` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductBarcodeBOMView | [[productbarcode-ProductBarcodeBOMView\|ProductBarcodeBOMView]] | inline | - | โครงสร้างฝังแบบ inline |
| CheckSum | string | checksum | checksum | ค่าของ CheckSum ตามฟิลด์ `checksum` ในซอร์ส |
| IsCurrentUse | bool | iscurrentuse | iscurrentuse | ค่าของ IsCurrentUse ตามฟิลด์ `iscurrentuse` ในซอร์ส |
| UseInDate | time.Time | useindate | useindate | ค่าของ UseInDate ตามฟิลด์ `useindate` ในซอร์ส |
| StartDate | time.Time | startdate | startdate | วันที่เริ่มต้น |
| EndDate | *time.Time | enddate | enddate | วันที่สิ้นสุด |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[productbarcode-ProductBarcodeBOMView|ProductBarcodeBOMView]]
