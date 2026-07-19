---
source: backend/internal/product/productgroup/models/productgroup.go
tags: [datamodel, general-type]
---

# ProductGroupActivity

โครงสร้าง `ProductGroupActivity` จากโมดูล mainapi `productgroup` มีฟิลด์ตามซอร์ส `productgroup.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ProductGroupData | [[productgroup-ProductGroupData\|ProductGroupData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityTime | [[ActivityTime\|models.ActivityTime]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productgroup-ProductGroupData|ProductGroupData]], [[ActivityTime]]
