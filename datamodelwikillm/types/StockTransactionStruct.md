---
source: process-model.go
tags: [datamodel, general-type]
---

# StockTransactionStruct

โครงสร้างจับคู่ชื่อประเภทรายการสต๊อกกับกลุ่มค่า trans flag ที่เกี่ยวข้อง ไม่มี bson/json tag — ใช้ภายใน

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Name | string | - | - | ชื่อประเภทรายการ |
| Flags | []int | - | - | กลุ่มค่า trans flag ของประเภทนั้น |

## ความสัมพันธ์
- `Flags` สัมพันธ์กับ field `TransFlag` ในโมเดล process อื่นๆ เช่น [[ProcessStockMovementDetailStruct]], [[ProcessStockCostDetailStruct]], [[ProcessStockLotStruct]]
