---
source: backend/internal/product/unit/models/unit.go
tags: [datamodel, general-type]
---

# UnitActivity

โครงสร้าง `UnitActivity` จากโมดูล mainapi `unit` มีฟิลด์ตามซอร์ส `unit.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| UnitData | [[unit-UnitData\|UnitData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityTime | [[ActivityTime\|models.ActivityTime]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[unit-UnitData|UnitData]], [[ActivityTime]]
