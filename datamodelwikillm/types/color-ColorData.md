---
source: backend/internal/product/color/models/color.go
tags: [datamodel, general-type]
---

# ColorData

โครงสร้าง `ColorData` จากโมดูล mainapi `color` มีฟิลด์ตามซอร์ส `color.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|models.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ColorInfo | [[color-ColorInfo\|ColorInfo]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[color-ColorInfo|ColorInfo]]
