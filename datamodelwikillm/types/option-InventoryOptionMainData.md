---
source: backend/internal/product/option/models/option.go
tags: [datamodel, general-type]
---

# InventoryOptionMainData

โครงสร้าง `InventoryOptionMainData` จากโมดูล mainapi `option` มีฟิลด์ตามซอร์ส `option.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCodeentity | [[HoldingCodeentity\|common.HoldingCodeentity]] | inline | - | โครงสร้างฝังแบบ inline |
| InventoryOptionMainInfo | [[option-InventoryOptionMainInfo\|InventoryOptionMainInfo]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[HoldingCodeentity]], [[option-InventoryOptionMainInfo|InventoryOptionMainInfo]]
