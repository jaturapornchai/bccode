---
source: mongo-product-model.go
tags: [datamodel, general-type]
---

# PriceModel

โมเดลราคาแบบฝัง (embedded) ใช้เก็บราคาหนึ่งรายการพร้อมหมายเลขระดับราคา (key number) ใช้ในฟิลด์ `Prices` ของ [[MongoProductBarcodeModel]]

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| KeyNumber | int | keynumber | keynumber | หมายเลขระดับ/ลำดับราคา |
| Price | float64 | price | price | มูลค่าราคา |

## ความสัมพันธ์

- ถูกฝังใน [[MongoProductBarcodeModel]] (ฟิลด์ Prices)
