---
source: backend/internal/product/productbarcode/models/product_bom.go
tags: [datamodel, general-type]
---

# ProductBarcodeBOMHistoryInfo

โครงสร้าง `ProductBarcodeBOMHistoryInfo` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_bom.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|models.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ProductBarcodeBOMHistory | [[DocIdentity\|models.DocIdentity]] | inline | - | ค่าของ ProductBarcodeBOMHistory ตามฟิลด์ `inline` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]]
