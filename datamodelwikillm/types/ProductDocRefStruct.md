---
source: process-model.go
tags: [datamodel, general-type]
---

# ProductDocRefStruct

ข้อมูลอ้างอิงสินค้าต่อบาร์โค้ด — รหัสสินค้า บาร์โค้ด หน่วยนับ และอัตราส่วนหน่วย

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| Barcode | string | barcode | barcode | บาร์โค้ด |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| UnitStand | float64 | unitstand | unitstand | ตัวคูณหน่วยมาตรฐาน |
| UnitDivide | float64 | unitdivide | unitdivide | ตัวหารหน่วย |

## ความสัมพันธ์
- `ItemCode` อ้างอิงสินค้า, `Barcode` อ้างอิงบาร์โค้ดสินค้า, `UnitCode` อ้างอิงหน่วยนับ
