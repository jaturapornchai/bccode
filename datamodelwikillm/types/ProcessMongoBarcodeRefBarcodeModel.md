---
source: mongo-barcode-model.go
tags: [datamodel, general-type]
---

# ProcessMongoBarcodeRefBarcodeModel

โมเดลบาร์โค้ดอ้างอิงที่ฝังใน [[ProcessMongoBarcodeModel]] (field `refbarcodes`) เก็บบาร์โค้ดอื่นของสินค้าเดียวกันพร้อมหน่วยนับและค่าแปลงหน่วย (ตัวคูณ/ตัวหาร)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Barcode | string | barcode | barcode | บาร์โค้ดอ้างอิง |
| ItemUnitCode | string | itemunitcode | itemunitcode | รหัสหน่วยนับของบาร์โค้ดนี้ |
| UnitStand | float64 | standvalue | standvalue | ค่ามาตรฐานสำหรับแปลงหน่วย (ตัวคูณ) |
| UnitDivide | float64 | dividevalue | dividevalue | ค่าหารสำหรับแปลงหน่วย |

## ความสัมพันธ์
- ฝังอยู่ใน [[ProcessMongoBarcodeModel]] (field `refbarcodes`)
- `Barcode` — อ้างอิงบาร์โค้ดสินค้า
- `ItemUnitCode` — อ้างอิงรหัสหน่วยนับ
