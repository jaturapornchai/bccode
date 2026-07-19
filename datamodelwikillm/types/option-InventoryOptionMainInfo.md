---
source: backend/internal/product/option/models/option.go
tags: [datamodel, general-type]
---

# InventoryOptionMainInfo

โครงสร้าง `InventoryOptionMainInfo` จากโมดูล mainapi `option` มีฟิลด์ตามซอร์ส `option.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocIdentity | [[DocIdentity\|common.DocIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| InventoryOptionMain | [[option-InventoryOptionMain\|InventoryOptionMain]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[DocIdentity]], [[option-InventoryOptionMain|InventoryOptionMain]]
