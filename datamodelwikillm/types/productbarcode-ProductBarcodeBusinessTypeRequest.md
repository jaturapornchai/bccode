---
source: backend/internal/product/productbarcode/models/product_barcode_request.go
tags: [datamodel, general-type]
---

# ProductBarcodeBusinessTypeRequest

โครงสร้าง `ProductBarcodeBusinessTypeRequest` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode_request.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| BusinessType | [[productbarcode-ProductBarcodeBusinessType\|ProductBarcodeBusinessType]] | - | businesstype | ค่าของ BusinessType ตามฟิลด์ `businesstype` ในซอร์ส |
| Products | []string | - | products | ค่าของ Products ตามฟิลด์ `products` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductBarcodeBusinessType|ProductBarcodeBusinessType]]
