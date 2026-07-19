---
source: process-model.go
tags: [datamodel, general-type]
---

# ProductBalanceByCodeLocationStruct

ยอดคงเหลือสินค้าระดับที่เก็บ (location) ใช้ซ้อนอยู่ใน ProductBalanceByCodeWareHouseStruct

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| LocationCode | string | locationcode | locationcode | รหัสที่เก็บ |
| BalanceQty | float64 | balanceqty | balanceqty | จำนวนคงเหลือ |
| BalanceWord | string | balanceword | balanceword | ยอดคงเหลือแบบข้อความ |

## ความสัมพันธ์
- ถูก embed ใน [[ProductBalanceByCodeWareHouseStruct]] (field `Locations`)
- `LocationCode` อ้างอิงที่เก็บในคลังสินค้า
