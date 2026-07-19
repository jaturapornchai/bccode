---
source: process-model.go
tags: [datamodel, general-type]
---

# ProductBarcodePackingStruct

ข้อมูลหน่วยบรรจุของบาร์โค้ดสำหรับ auto packing (ตาม comment ในโค้ด) เก็บชื่อหน่วยและอัตราส่วนหน่วยของบาร์โค้ดอ้างอิง

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| UnitName | string | - | - | ชื่อหน่วยนับ |
| BarcodeRefUnitStand | float64 | - | - | ตัวคูณหน่วยมาตรฐานของบาร์โค้ดอ้างอิง |
| BarcodeRefUnitDivide | float64 | - | - | ตัวหารหน่วยของบาร์โค้ดอ้างอิง |

## ความสัมพันธ์
- ไม่มี embedded struct; ใช้ข้อมูลอัตราส่วนหน่วยจากบาร์โค้ดอ้างอิง (barcode ref)
