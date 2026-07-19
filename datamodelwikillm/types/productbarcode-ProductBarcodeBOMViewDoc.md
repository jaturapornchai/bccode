---
source: backend/internal/product/productbarcode/models/product_bom.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMViewDoc

โครงสร้าง `ProductBarcodeBOMViewDoc` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ID | primitive.ObjectID | _id,omitempty | id | รหัส ObjectID ของเอกสาร |
| ProductBarcodeBOMViewData | [[productbarcode-ProductBarcodeBOMViewData\|ProductBarcodeBOMViewData]] | inline | - | โครงสร้างฝังแบบ inline |
| ActivityDoc | [[ActivityDoc\|models.ActivityDoc]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductBarcodeBOMViewData|ProductBarcodeBOMViewData]], [[ActivityDoc]]
