---
source: backend/internal/product/unit/models/unit.go
tags: [datamodel, general-type]
---

# UnitInfo

โครงสร้าง `UnitInfo` จากโมดูล mainapi `unit` มีฟิลด์ตามซอร์ส `unit.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Unit | [[unit-Unit\|Unit]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[unit-Unit|Unit]]
