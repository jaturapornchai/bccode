---
source: backend/internal/product/color/models/color.go
tags: [datamodel, general-type]
---

# ColorActivity

โครงสร้าง `ColorActivity` จากโมดูล mainapi `color` มีฟิลด์ตามซอร์ส `color.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ColorData | [[color-ColorData\|ColorData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityTime | [[ActivityTime\|models.ActivityTime]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[color-ColorData|ColorData]], [[ActivityTime]]
