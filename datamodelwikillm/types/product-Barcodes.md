---
source: backend/internal/product/product/models/product.go
tags: [datamodel, general-type]
---

# Barcodes

โครงสร้าง `Barcodes` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GuidFixed | string | - | guidfixed | GUID ถาวรของรายการ |
| ItemUnitCode | string | - | itemunitcode | รหัสหน่วยนับสินค้า |
| ItemUnitNames | *[][[NameX\|models.NameX]] | - | itemunitnames | ค่าของ ItemUnitNames ตามฟิลด์ `itemunitnames` ในซอร์ส |
| Barcode | string | - | barcode | บาร์โค้ดสินค้า |
| Prices | *[][[product-ProductPrice\|ProductPrice]] | - | prices | ค่าของ Prices ตามฟิลด์ `prices` ในซอร์ส |
| Condition | bool | - | condition | ค่าของ Condition ตามฟิลด์ `condition` ในซอร์ส |
| DivideValue | float64 | - | dividevalue | ค่าของ DivideValue ตามฟิลด์ `dividevalue` ในซอร์ส |
| StandValue | float64 | - | standvalue | ค่าของ StandValue ตามฟิลด์ `standvalue` ในซอร์ส |
| Qty | float64 | - | qty | จำนวน |
| IsMainBarcode | bool | - | ismainbarcode | ค่าของ IsMainBarcode ตามฟิลด์ `ismainbarcode` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[NameX]], [[product-ProductPrice|ProductPrice]]
