---
source: barcode-model.go
tags: [datamodel, general-type]
---

# BarcodeModel

โมเดลข้อมูลบาร์โค้ดสินค้าแบบ flat สำหรับใช้ภายใน goapi ประกอบด้วยรหัสสินค้า ชื่อหลายภาษา หน่วยนับ กลุ่มสินค้า และราคา ไม่มี bson/json tag (ไม่ได้ map ตรงกับ collection ใด) มี constructor `NewBarcodeModel()` คืนค่า default ว่าง/ศูนย์

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | - | รหัสกิจการ (tenant) |
| ItemCode | string | - | - | รหัสสินค้า |
| Barcode | string | - | - | บาร์โค้ด |
| BarcodeRef | string | - | - | บาร์โค้ดอ้างอิง (บาร์โค้ดหลักที่อ้างถึง) |
| Name0 | string | - | - | ชื่อสินค้า ภาษาที่ 0 |
| Name1 | string | - | - | ชื่อสินค้า ภาษาที่ 1 |
| Name2 | string | - | - | ชื่อสินค้า ภาษาที่ 2 |
| Name3 | string | - | - | ชื่อสินค้า ภาษาที่ 3 |
| Name4 | string | - | - | ชื่อสินค้า ภาษาที่ 4 |
| Name5 | string | - | - | ชื่อสินค้า ภาษาที่ 5 |
| UnitCode | string | - | - | รหัสหน่วยนับ |
| UnitName | string | - | - | ชื่อหน่วยนับ |
| GroupCode | string | - | - | รหัสกลุ่มสินค้า |
| GroupNames | string | - | - | ชื่อกลุ่มสินค้า |
| Price1 | float64 | - | - | ราคาระดับ 1 |
| PriceRetail | float64 | - | - | ราคาขายปลีก |
| UnitStand | float64 | - | - | ค่าตัวคูณหน่วย (standard value) |
| UnitDivide | float64 | - | - | ค่าตัวหารหน่วย (divide value) |
| BarcodeRefUnitStand | float64 | - | - | ค่าตัวคูณหน่วยของบาร์โค้ดอ้างอิง |
| BarcodeRefUnitDivide | float64 | - | - | ค่าตัวหารหน่วยของบาร์โค้ดอ้างอิง |
| IsStock | int | - | - | flag ตัดสต็อกหรือไม่ |
| ItemType | int | - | - | ประเภทสินค้า |
| MaterialType | int | - | - | ประเภทวัตถุดิบ |
| Checksum | string | - | - | ค่า checksum ของข้อมูล |
| ImageUri | string | - | - | URI รูปภาพสินค้า |

## ความสัมพันธ์

- ไม่มี struct ซ้อนภายใน (flat struct ล้วน)
- `HoldingCode` — อ้างอิงกิจการ (tenant)
- `ItemCode` — อ้างอิงรหัสสินค้า
- `Barcode` / `BarcodeRef` — อ้างอิงบาร์โค้ดสินค้า (BarcodeRef ชี้ไปบาร์โค้ดอื่นของสินค้าเดียวกัน — ดู [[BarcodeRefModel]])
- `UnitCode` — อ้างอิงหน่วยนับ
- `GroupCode` — อ้างอิงกลุ่มสินค้า
