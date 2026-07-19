---
source: process-model.go
tags: [datamodel, general-type]
---

# ProductBalanceByCodeLocationGetStruct

โครงสร้างรับ/อ่านยอดคงเหลือสินค้าต่อคลังและที่เก็บ (แบบ flat ต่อ row) ไม่มี bson/json tag — ใช้ภายในกระบวนการ query

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ItemCode | string | - | - | รหัสสินค้า |
| WareHouseCode | string | - | - | รหัสคลังสินค้า |
| LocationCode | string | - | - | รหัสที่เก็บ |
| BalanceQty | float64 | - | - | จำนวนคงเหลือ |

## ความสัมพันธ์
- `ItemCode` อ้างอิงสินค้า, `WareHouseCode` อ้างอิงคลัง, `LocationCode` อ้างอิงที่เก็บ
- เป็นข้อมูลระดับเดียวกับ [[ProductBalanceByCodeLocationStruct]] แต่แบบ flat สำหรับอ่านจากฐานข้อมูล
