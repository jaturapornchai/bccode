---
source: backend/internal/product/productbarcode/models/product_price_history.go
tags: [datamodel, general-type]
---

# PriceChangeRequest

โครงสร้าง `PriceChangeRequest` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_price_history.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ProductBarcodeGUID | string | - | productbarcodeguid | ค่าของ ProductBarcodeGUID ตามฟิลด์ `productbarcodeguid` ในซอร์ส |
| ItemCode | string | - | itemcode | รหัสสินค้า |
| Barcode | string | - | barcode | บาร์โค้ดสินค้า |
| ProductName | string | - | productname | ค่าของ ProductName ตามฟิลด์ `productname` ในซอร์ส |
| OldPrices | [][[productbarcode-ProductPrice\|ProductPrice]] | - | oldprices | ค่าของ OldPrices ตามฟิลด์ `oldprices` ในซอร์ส |
| NewPrices | [][[productbarcode-ProductPrice\|ProductPrice]] | - | newprices | ค่าของ NewPrices ตามฟิลด์ `newprices` ในซอร์ส |
| Remark | string | - | remark | ค่าของ Remark ตามฟิลด์ `remark` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductPrice|ProductPrice]]
