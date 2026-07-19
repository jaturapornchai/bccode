---
source: backend/internal/product/promotion/models/promotion.go
tags: [datamodel, general-type]
---

# ProductBarcode

โครงสร้าง `ProductBarcode` จากโมดูล mainapi `promotion` มีฟิลด์ตามซอร์ส `promotion.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GuidFixed | string | guidfixed | guidfixed | GUID ถาวรของรายการ |
| DiscountText | string | discounttext | discounttext | ค่าของ DiscountText ตามฟิลด์ `discounttext` ในซอร์ส |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| Name | *[][[NameX\|models.NameX]] | name | name | ชื่อรายการ |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| UnitName | *[][[NameX\|models.NameX]] | unitname | unitname | ค่าของ UnitName ตามฟิลด์ `unitname` ในซอร์ส |
| Price | float64 | price | price | ราคา |
| Qty | float64 | qty | qty | จำนวน |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[NameX]]
