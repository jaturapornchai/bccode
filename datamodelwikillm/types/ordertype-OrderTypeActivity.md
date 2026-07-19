---
source: backend/internal/product/ordertype/models/ordertype.go
tags: [datamodel, general-type]
---

# OrderTypeActivity

โครงสร้าง `OrderTypeActivity` จากโมดูล mainapi `ordertype` มีฟิลด์ตามซอร์ส `ordertype.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| OrderTypeData | [[ordertype-OrderTypeData\|OrderTypeData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityTime | [[ActivityTime\|models.ActivityTime]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[ordertype-OrderTypeData|OrderTypeData]], [[ActivityTime]]
