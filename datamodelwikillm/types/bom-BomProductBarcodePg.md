---
source: backend/internal/product/bom/models/bom_postgres.go
tags: [datamodel, general-type]
---

# BomProductBarcodePg

โครงสร้าง `BomProductBarcodePg` จากโมดูล mainapi `bom` มีฟิลด์ตามซอร์ส `bom_postgres.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| BarcodeGuidFixed | string | - | guidfixed | ค่าของ BarcodeGuidFixed ตามฟิลด์ `guidfixed` ในซอร์ส |
| Level | int | - | level | ค่าของ Level ตามฟิลด์ `level` ในซอร์ส |
| Names | [[bom-JSONB\|JSONB]] | - | names | รายการชื่อหลายภาษา |
| ItemUnitCode | string | - | itemunitcode | รหัสหน่วยนับสินค้า |
| ItemUnitNames | [[bom-JSONB\|JSONB]] | - | itemunitnames | ค่าของ ItemUnitNames ตามฟิลด์ `itemunitnames` ในซอร์ส |
| Barcode | string | - | barcode | บาร์โค้ดสินค้า |
| Condition | bool | - | condition | ค่าของ Condition ตามฟิลด์ `condition` ในซอร์ส |
| DivideValue | float64 | - | dividevalue | ค่าของ DivideValue ตามฟิลด์ `dividevalue` ในซอร์ส |
| StandValue | float64 | - | standvalue | ค่าของ StandValue ตามฟิลด์ `standvalue` ในซอร์ส |
| Qty | float64 | - | qty | จำนวน |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[bom-JSONB|JSONB]]
