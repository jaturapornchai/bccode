---
source: backend/internal/product/productbarcode/models/product_price_history.go
tags: [datamodel, general-type]
---

# PriceHistoryFilter

โครงสร้าง `PriceHistoryFilter` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_price_history.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Barcode | string | - | barcode | บาร์โค้ดสินค้า |
| ItemCode | string | - | itemcode | รหัสสินค้า |
| ProductBarcodeGUID | string | - | productbarcodeguid | ค่าของ ProductBarcodeGUID ตามฟิลด์ `productbarcodeguid` ในซอร์ส |
| Action | string | - | action | ค่าของ Action ตามฟิลด์ `action` ในซอร์ส |
| CreatedBy | string | - | createdby | ค่าของ CreatedBy ตามฟิลด์ `createdby` ในซอร์ส |
| FromDate | string | - | fromdate | วันที่เริ่มช่วง |
| ToDate | string | - | todate | วันที่สิ้นสุดช่วง |

## ความสัมพันธ์

- ไม่มีฟิลด์ที่ใช้ struct ซ้อนตามประกาศนี้
