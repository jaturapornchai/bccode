---
source: process-model.go
tags: [datamodel, general-type]
---

# ProductBalanceStruct

โครงสร้างเก็บข้อมูลยอดคงเหลือสินค้าแต่ละ row (ตาม comment ในโค้ด: "กำหนด struct เพื่อเก็บข้อมูลแต่ละ row") แยกตามคลังและที่เก็บ

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ItemCode | string | - | - | รหัสสินค้า |
| ItemName | string | - | - | ชื่อสินค้า |
| WhCode | string | - | - | รหัสคลังสินค้า |
| LocationCode | string | - | - | รหัสที่เก็บ |
| UnitCode | string | - | - | รหัสหน่วยนับ |
| UnitName | string | - | - | ชื่อหน่วยนับ |
| BalanceQty | float64 | - | - | จำนวนคงเหลือ |
| BalanceAmount | float64 | - | - | มูลค่าคงเหลือ |
| BalanceWord | string | - | - | ยอดคงเหลือแบบข้อความ |
| IsAutoPacking | uint64 | - | - | ธง auto packing |

## ความสัมพันธ์
- `ItemCode` อ้างอิงสินค้า, `WhCode` อ้างอิงคลัง, `LocationCode` อ้างอิงที่เก็บ, `UnitCode` อ้างอิงหน่วยนับ
