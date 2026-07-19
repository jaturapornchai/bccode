---
source: mongo-product-model.go
collection: productbarcodes
tags: [datamodel, mongodb, mongo-root]
---

# MongoProductBarcodeModel

โมเดลข้อมูลบาร์โค้ดสินค้าใน MongoDB collection `productbarcodes` เก็บข้อมูลสินค้าต่อบาร์โค้ด ทั้งชื่อหลายภาษา หน่วยนับ ราคา และรหัสจัดกลุ่ม/หมวดหมู่ต่างๆ ของสินค้า

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | holdingcode | holdingcode | รหัสกิจการ (tenant) |
| GuidFixed | string | guidfixed | guidfixed | GUID ประจำเอกสาร |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| Names | `[]`[[LanguageModel]] | names | names | ชื่อสินค้าหลายภาษา |
| GroupCode | string | groupcode | groupcode | รหัสกลุ่มสินค้า |
| GroupNames | `[]`[[LanguageModel]] | groupnames | groupnames | ชื่อกลุ่มสินค้าหลายภาษา |
| ItemUnitCode | string | itemunitcode | itemunitcode | รหัสหน่วยนับ |
| ItemUnitNames | `[]`[[LanguageModel]] | itemunitnames | itemunitnames | ชื่อหน่วยนับหลายภาษา |
| Prices | `[]`[[PriceModel]] | prices | prices | รายการราคาขาย |
| DivideValue | float64 | dividevalue | dividevalue | ตัวหารสำหรับแปลงหน่วย |
| StandValue | float64 | standvalue | standvalue | ค่ามาตรฐานสำหรับแปลงหน่วย |
| RefBarCodes | `[]`[[ProcessMongoBarcodeRefBarcodeModel]] | refbarcodes | refbarcodes | บาร์โค้ดอ้างอิง/บาร์โค้ดย่อย |
| ItemType | int | itemtype | itemtype | ประเภทสินค้า |
| MaterialType | int | materialtype | materialtype | ประเภทวัตถุดิบ |
| IsUseSubBarcodes | bool | isusesubbarcodes | isusesubbarcodes | ใช้บาร์โค้ดย่อยหรือไม่ |
| ImageUri | string | imageuri | imageuri | URI รูปภาพสินค้า |
| BrandCode | string | brandcode | brandcode | รหัสแบรนด์ |
| BrandNames | `[]`[[LanguageModel]] | brandnames | brandnames | ชื่อแบรนด์หลายภาษา |
| CategoryCode | string | categorycode | categorycode | รหัสหมวดหมู่สินค้า |
| CategoryNames | `[]`[[LanguageModel]] | categorynames | categorynames | ชื่อหมวดหมู่หลายภาษา |
| ClassCode | string | classcode | classcode | รหัส class สินค้า |
| ClassNames | `[]`[[LanguageModel]] | classnames | classnames | ชื่อ class หลายภาษา |
| DesignCode | string | designcode | designcode | รหัส design สินค้า |
| DesignNames | `[]`[[LanguageModel]] | designnames | designnames | ชื่อ design หลายภาษา |
| GradeCode | string | gradecode | gradecode | รหัสเกรดสินค้า |
| GradeNames | `[]`[[LanguageModel]] | gradenames | gradenames | ชื่อเกรดหลายภาษา |
| ModelCode | string | modelcode | modelcode | รหัสรุ่นสินค้า |
| ModelNames | `[]`[[LanguageModel]] | modelnames | modelnames | ชื่อรุ่นหลายภาษา |
| PatternCode | string | patterncode | patterncode | รหัสลาย/แพทเทิร์นสินค้า |
| PatternNames | `[]`[[LanguageModel]] | patternnames | patternnames | ชื่อลาย/แพทเทิร์นหลายภาษา |
| GroupSubOneCode | string | groupsubonecode | groupsubonecode | รหัสกลุ่มย่อยที่ 1 |
| GroupSubOneNames | `[]`[[LanguageModel]] | groupsubonenames | groupsubonenames | ชื่อกลุ่มย่อยที่ 1 หลายภาษา |
| GroupSubTwoCode | string | groupsubtwocode | groupsubtwocode | รหัสกลุ่มย่อยที่ 2 |
| GroupSubTwoNames | `[]`[[LanguageModel]] | groupsubtwonames | groupsubtwonames | ชื่อกลุ่มย่อยที่ 2 หลายภาษา |

## ความสัมพันธ์

- Embedded: [[LanguageModel]] (ชื่อหลายภาษาของสินค้า/กลุ่ม/หน่วย/แบรนด์/หมวดหมู่ ฯลฯ), [[PriceModel]] (รายการราคา), [[ProcessMongoBarcodeRefBarcodeModel]] (บาร์โค้ดอ้างอิง)
- HoldingCode → อ้างถึงกิจการ (tenant boundary)
- ItemCode → รหัสสินค้าหลักที่บาร์โค้ดนี้สังกัด
- Barcode → บาร์โค้ดประจำรายการนี้
- GroupCode / CategoryCode / BrandCode / ClassCode / DesignCode / GradeCode / ModelCode / PatternCode / GroupSubOneCode / GroupSubTwoCode → รหัสอ้างอิงกลุ่ม/หมวดหมู่/คุณลักษณะของสินค้า (เก็บชื่อคู่กันในฟิลด์ *Names)
- ItemUnitCode → รหัสหน่วยนับของสินค้า
