---
source: backend/internal/goapi/models/global-model.go
tags: [datamodel, general-type]
---

# productOrderUnitModel

โครงสร้างหน่วยนับของสินค้าในการสั่งซื้อ เก็บรหัสหน่วยนับพร้อมชื่อหน่วยหลายภาษา

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Unitcode | string | - | unitcode | รหัสหน่วยนับ |
| Unitnames | [][[languageNameModel]] | - | unitnames | ชื่อหน่วยนับหลายภาษา |

## ความสัมพันธ์
- ฝัง [[languageNameModel]] (Unitnames)
- ถูกฝังใน [[productOrderModel]] (Units)
- `Unitcode` — รหัสหน่วยนับ ใช้จับคู่กับ Unitcode ใน [[productOrderUnitUseModel]]
