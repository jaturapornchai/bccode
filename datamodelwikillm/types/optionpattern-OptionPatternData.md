---
source: backend/internal/product/optionpattern/models/optionpattern.go
tags: [datamodel, general-type]
---

# OptionPatternData

โครงสร้าง `OptionPatternData` จากโมดูล mainapi `optionpattern` มีฟิลด์ตามซอร์ส `optionpattern.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| OptionPatternInfo | [[optionpattern-OptionPatternInfo\|OptionPatternInfo]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[optionpattern-OptionPatternInfo|OptionPatternInfo]]
