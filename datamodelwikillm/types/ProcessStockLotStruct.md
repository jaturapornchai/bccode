---
source: process-model.go
tags: [datamodel, general-type]
---

# ProcessStockLotStruct

ข้อมูลสต๊อกระดับ lot (LotNumber) ต่อเอกสาร/คลัง/ที่เก็บ พร้อมจำนวน ต้นทุน และยอดคงเหลือ ไม่มี bson/json tag — ใช้ภายในกระบวนการ process

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocDateTime | time.Time | - | - | วันที่/เวลาเอกสาร |
| LotNumber | string | - | - | เลขที่ lot |
| DocNo | string | - | - | เลขที่เอกสาร |
| TransFlag | int | - | - | ประเภทรายการ (trans flag) |
| ItemCode | string | - | - | รหัสสินค้า |
| UnitCode | string | - | - | รหัสหน่วยนับ |
| WhCode | string | - | - | รหัสคลังสินค้า |
| LocationCode | string | - | - | รหัสที่เก็บ |
| Qty | float64 | - | - | จำนวน |
| Price | float64 | - | - | ราคา |
| UnitStand | float64 | - | - | ตัวคูณหน่วยมาตรฐาน |
| UnitDivide | float64 | - | - | ตัวหารหน่วย |
| Cost | float64 | - | - | ต้นทุน |
| BalanceAmount | float64 | - | - | มูลค่าคงเหลือ |
| BalanceQty | float64 | - | - | จำนวนคงเหลือ |
| GuidRef | string | - | - | GUID อ้างอิง |

## ความสัมพันธ์
- `ItemCode` อ้างอิงสินค้า, `UnitCode` อ้างอิงหน่วยนับ, `WhCode` อ้างอิงคลัง, `LocationCode` อ้างอิงที่เก็บ, `DocNo` อ้างอิงเอกสาร
