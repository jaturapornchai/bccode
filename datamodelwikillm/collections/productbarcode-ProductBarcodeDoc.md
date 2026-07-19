---
source: backend/internal/product/productbarcode/models/product_barcode.go
collection: productbarcodes
tags: [datamodel, mongodb, mongo-root]
---

# ProductBarcodeDoc

โครงสร้าง `ProductBarcodeDoc` เป็น root document ที่ repository โมดูล `productbarcode` ใช้กับ MongoDB collection `productbarcodes`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| ProductBarcodeData | [[productbarcode-ProductBarcodeData\|ProductBarcodeData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|models.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductBarcodeData|ProductBarcodeData]], [[ActivityDoc]]
- MongoDB collection: `productbarcodes`
