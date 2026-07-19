---
source: backend/internal/product/bom/models/bom.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMViewData

โครงสร้าง `ProductBarcodeBOMViewData` จากโมดูล mainapi `bom` มีฟิลด์ตามซอร์ส `bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductBarcodeBOMViewInfo | [[bom-ProductBarcodeBOMViewInfo\|ProductBarcodeBOMViewInfo]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[bom-ProductBarcodeBOMViewInfo|ProductBarcodeBOMViewInfo]]
