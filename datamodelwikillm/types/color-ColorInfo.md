---
source: backend/internal/product/color/models/color.go
tags: [datamodel, general-type]
---

# ColorInfo

โครงสร้าง `ColorInfo` จากโมดูล mainapi `color` มีฟิลด์ตามซอร์ส `color.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Color | [[color-Color\|Color]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[color-Color|Color]]
