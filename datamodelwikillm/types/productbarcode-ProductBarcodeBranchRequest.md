---
source: backend/internal/product/productbarcode/models/product_barcode_request.go
tags: [datamodel, general-type]
---

# ProductBarcodeBranchRequest

โครงสร้าง `ProductBarcodeBranchRequest` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_barcode_request.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Branch | [[productbarcode-ProductBarcodeBranch\|ProductBarcodeBranch]] | - | branch | ค่าของ Branch ตามฟิลด์ `branch` ในซอร์ส |
| Products | []string | - | products | ค่าของ Products ตามฟิลด์ `products` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductBarcodeBranch|ProductBarcodeBranch]]
