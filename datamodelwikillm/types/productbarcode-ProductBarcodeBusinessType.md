---
source: backend/internal/product/productbarcode/models/product_barcode.go
tags: [datamodel, general-type]
---

# ProductBarcodeBusinessType

โครงสร้าง `ProductBarcodeBusinessType` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Code | string | code | code | รหัสรายการ |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| IsIgnore | bool | isignore | isignore | ค่าของ IsIgnore ตามฟิลด์ `isignore` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[NameX]]
