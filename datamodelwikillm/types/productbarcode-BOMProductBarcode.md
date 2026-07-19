---
source: backend/internal/product/productbarcode/models/product_bom.go
tags: [datamodel, general-type]
---

# BOMProductBarcode

โครงสร้าง `BOMProductBarcode` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| BarcodeGuidFixed | string | guidfixed | guidfixed | ค่าของ BarcodeGuidFixed ตามฟิลด์ `guidfixed` ในซอร์ส |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| Level | int | level | level | ค่าของ Level ตามฟิลด์ `level` ในซอร์ส |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| ItemUnitCode | string | itemunitcode | itemunitcode | รหัสหน่วยนับสินค้า |
| ItemUnitNames | *[][[NameX\|models.NameX]] | itemunitnames | itemunitnames | ค่าของ ItemUnitNames ตามฟิลด์ `itemunitnames` ในซอร์ส |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| Condition | bool | condition | condition | ค่าของ Condition ตามฟิลด์ `condition` ในซอร์ส |
| DivideValue | float64 | dividevalue | dividevalue | ค่าของ DivideValue ตามฟิลด์ `dividevalue` ในซอร์ส |
| StandValue | float64 | standvalue | standvalue | ค่าของ StandValue ตามฟิลด์ `standvalue` ในซอร์ส |
| Qty | float64 | qty | qty | จำนวน |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[NameX]]
