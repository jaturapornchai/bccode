---
source: mongo-barcode-model.go
tags: [datamodel, general-type]
---

# ProcessMongoBarcodePriceModel

โมเดลราคาที่ฝังใน [[ProcessMongoBarcodeModel]] (field `prices`) เก็บราคาต่อระดับราคา (key number)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| KeyNumber | int | keynumber | keynumber | หมายเลขระดับราคา |
| Price | float64 | price | price | ราคา |

## ความสัมพันธ์
- ฝังอยู่ใน [[ProcessMongoBarcodeModel]] (field `prices`)
