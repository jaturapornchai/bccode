---
source: backend/internal/product/bom/models/bom_postgres.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMViewPG

โครงสร้าง `ProductBarcodeBOMViewPG` จากโมดูล mainapi `bom` มีฟิลด์ตามซอร์ส `bom_postgres.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| GuidFixed | string | guidfixed | guidfixed | GUID ถาวรของรายการ |
| BOMProductBarcode | [[bom-BomProductBarcodePg\|BomProductBarcodePg]] | - | - | ค่าของ BOMProductBarcode ตามฟิลด์ `BOMProductBarcode` ในซอร์ส |
| ImageURI | string | - | imageuri | URI ของรูปภาพ |
| BOM | BOMViewPg | - | bom | ค่าของ BOM ตามฟิลด์ `bom` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[bom-BomProductBarcodePg|BomProductBarcodePg]]
