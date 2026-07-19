---
source: barcode-model.go
tags: [datamodel, general-type]
---

# BarcodeRefModel

โมเดลความสัมพันธ์ระหว่างบาร์โค้ดกับบาร์โค้ดอ้างอิง พร้อมค่าตัวคูณ/ตัวหารหน่วยสำหรับแปลงหน่วยระหว่างกัน ไม่มี bson/json tag (ใช้ภายใน goapi)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | - | รหัสกิจการ (tenant) |
| Barcode | string | - | - | บาร์โค้ด |
| BarcodeRef | string | - | - | บาร์โค้ดอ้างอิง |
| ItemCode | string | - | - | รหัสสินค้า |
| UnitCode | string | - | - | รหัสหน่วยนับ |
| StandValue | float64 | - | - | ค่าตัวคูณหน่วย (standard value) |
| DivideValue | float64 | - | - | ค่าตัวหารหน่วย (divide value) |

## ความสัมพันธ์

- ไม่มี struct ซ้อนภายใน (flat struct ล้วน)
- `HoldingCode` — อ้างอิงกิจการ (tenant)
- `Barcode` / `BarcodeRef` — จับคู่บาร์โค้ดกับบาร์โค้ดอ้างอิง (ใช้คู่กับ [[BarcodeModel]])
- `ItemCode` — อ้างอิงรหัสสินค้า
- `UnitCode` — อ้างอิงหน่วยนับ
