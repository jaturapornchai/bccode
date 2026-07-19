---
source: backend/internal/goapi/models/global-model.go
tags: [datamodel, general-type]
---

# productOrderUnitUseModel

โครงสร้างหน่วยนับที่ใช้ได้ของสินค้า พร้อมอัตราแปลงหน่วยเทียบหน่วยมาตรฐาน และ flag ว่าเป็นหน่วยต้นทุนหรือไม่

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Unitcode | string | - | unitcode | รหัสหน่วยนับ |
| Itemunitstd | float64 | - | itemunitstd | ตัวคูณแปลงเป็นหน่วยมาตรฐาน |
| Itemunitdiv | float64 | - | itemunitdiv | ตัวหารแปลงหน่วย |
| Isunitcost | bool | - | isunitcost | เป็นหน่วยต้นทุนหรือไม่ |

## ความสัมพันธ์
- ถูกฝังใน [[productOrderModel]] (Unituses)
- `Unitcode` — รหัสหน่วยนับ ใช้จับคู่กับ [[productOrderUnitModel]]
