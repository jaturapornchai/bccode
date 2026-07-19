---
source: backend/internal/product/ordertype/models/ordertype.go
tags: [datamodel, general-type]
---

# OrderType

โครงสร้าง `OrderType` จากโมดูล mainapi `ordertype` มีฟิลด์ตามซอร์ส `ordertype.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| PartitionIdentity | [[PartitionIdentity\|models.PartitionIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Code | string | code | code | รหัสรายการ |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| Prices | *[][[ordertype-OrderTypePrice\|OrderTypePrice]] | prices | prices | ค่าของ Prices ตามฟิลด์ `prices` ในซอร์ส |
| Remarks | []*[][[NameX\|models.NameX]] | remarks | remarks | ค่าของ Remarks ตามฟิลด์ `remarks` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[PartitionIdentity]], [[NameX]], [[ordertype-OrderTypePrice|OrderTypePrice]]
