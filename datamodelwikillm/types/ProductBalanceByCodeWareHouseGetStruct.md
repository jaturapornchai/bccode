---
source: process-model.go
tags: [datamodel, general-type]
---

# ProductBalanceByCodeWareHouseGetStruct

โครงสร้างรับ/อ่านยอดคงเหลือสินค้าต่อคลัง (แบบ flat ต่อ row) ไม่มี bson/json tag — ใช้ภายในกระบวนการ query

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ItemCode | string | - | - | รหัสสินค้า |
| WareHouseCode | string | - | - | รหัสคลังสินค้า |
| BalanceQty | float64 | - | - | จำนวนคงเหลือ |
| AverageCost | float64 | - | - | ต้นทุนเฉลี่ย |
| BalanceAmount | float64 | - | - | มูลค่าคงเหลือ |

## ความสัมพันธ์
- `ItemCode` อ้างอิงสินค้า, `WareHouseCode` อ้างอิงคลังสินค้า
- เป็นข้อมูลระดับเดียวกับ [[ProductBalanceByCodeWareHouseStruct]] แต่แบบ flat สำหรับอ่านจากฐานข้อมูล
