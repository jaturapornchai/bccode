---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# StockTransferDetailStruct

รายการรายละเอียด (detail line) ของเอกสารโอนย้ายสินค้า [[StockTransferStruct]] ระบุสินค้า จำนวน ราคา และคลัง/ที่เก็บต้นทาง-ปลายทาง

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| LineNumber | int | linenumber | linenumber | ลำดับบรรทัดรายการ |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| Description | string | description | description | คำอธิบายรายการ |
| BarcodeMain | string | barcodemain | barcodemain | บาร์โค้ดหลักของสินค้า |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| WhCode | string | whcode | whcode | รหัสคลังต้นทาง |
| LocationCode | string | locationcode | locationcode | รหัสที่เก็บต้นทาง |
| ToWhCode | string | towhcode | towhcode | รหัสคลังปลายทาง |
| ToLocationCode | string | tolocationcode | tolocationcode | รหัสที่เก็บปลายทาง |
| Qty | float64 | qty | qty | จำนวน |
| Price | float64 | price | price | ราคาต่อหน่วย |
| PriceExcludeVat | float64 | priceexcludevat | priceexcludevat | ราคาไม่รวม VAT |
| UnitStand | float64 | unitstand | unitstand | ตัวคูณหน่วยนับมาตรฐาน |
| UnitDivide | float64 | unitdivide | unitdivide | ตัวหารหน่วยนับ |
| DocRef | string | docref | docref | เลขที่เอกสารอ้างอิง |
| SumAmount | float64 | sumamount | sumamount | ยอดรวมบรรทัด |

## ความสัมพันธ์
- ถูกฝังเป็น `Details` ใน [[StockTransferStruct]]
- ItemCode / Barcode / BarcodeMain — อ้างอิงสินค้า/บาร์โค้ดสินค้า
- UnitCode — อ้างอิงหน่วยนับ
- WhCode / LocationCode — คลัง/ที่เก็บต้นทาง; ToWhCode / ToLocationCode — คลัง/ที่เก็บปลายทาง
- DocRef — อ้างอิงเอกสารอื่น
