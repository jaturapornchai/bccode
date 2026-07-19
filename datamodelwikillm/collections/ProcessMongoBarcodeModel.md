---
source: mongo-barcode-model.go
collection: productbarcodes
tags: [datamodel, mongodb, mongo-root]
---

# ProcessMongoBarcodeModel

โมเดลข้อมูลบาร์โค้ดสินค้าใน collection `productbarcodes` ใช้ในขั้นตอน process ของ goapi ประกอบด้วยรหัสสินค้า บาร์โค้ด กลุ่มสินค้า หน่วยนับ ราคา และบาร์โค้ดอ้างอิง (หน่วยแปลง)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | holdingcode | holdingcode | รหัสกิจการ (tenant) |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| Barcode | string | barcode | barcode | บาร์โค้ดหลักของสินค้า |
| ItemType | int | itemtype | itemtype | ประเภทสินค้า |
| MaterialType | int | materialtype | materialtype | ประเภทวัสดุ |
| Names | [][[LanguageModel]] | names | names | ชื่อสินค้าหลายภาษา |
| GroupCode | string | groupcode | groupcode | รหัสกลุ่มสินค้า |
| GroupNames | [][[LanguageModel]] | groupnames | groupnames | ชื่อกลุ่มสินค้าหลายภาษา |
| ItemUnitCode | string | itemunitcode | itemunitcode | รหัสหน่วยนับ |
| ItemUnitNames | [][[LanguageModel]] | itemunitnames | itemunitnames | ชื่อหน่วยนับหลายภาษา |
| Prices | [][[ProcessMongoBarcodePriceModel]] | prices | prices | รายการราคาขาย |
| RefBarCodes | [][[ProcessMongoBarcodeRefBarcodeModel]] | refbarcodes | refbarcodes | บาร์โค้ดอ้างอิง/หน่วยแปลง |
| ImageUri | string | imageuri | imageuri | URI รูปภาพสินค้า |

## ความสัมพันธ์
- Embedded: [[LanguageModel]] (Names, GroupNames, ItemUnitNames), [[ProcessMongoBarcodePriceModel]] (Prices), [[ProcessMongoBarcodeRefBarcodeModel]] (RefBarCodes)
- `HoldingCode` — อ้างอิงรหัสกิจการ (tenant boundary)
- `ItemCode` — อ้างอิงรหัสสินค้า
- `GroupCode` — อ้างอิงรหัสกลุ่มสินค้า
- `ItemUnitCode` — อ้างอิงรหัสหน่วยนับ
