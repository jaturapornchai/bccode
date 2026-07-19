---
source: backend/internal/product/unit/models/unit.go
tags: [datamodel, general-type]
---

# UnitData

โครงสร้าง `UnitData` จากโมดูล mainapi `unit` มีฟิลด์ตามซอร์ส `unit.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| UnitInfo | [[unit-UnitInfo\|UnitInfo]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[unit-UnitInfo|UnitInfo]]
