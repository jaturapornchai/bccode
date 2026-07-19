---
source: process-model.go
tags: [datamodel, general-type]
---

# ProductBalanceByCodeWareHouseStruct

ยอดคงเหลือสินค้าระดับคลังสินค้า (warehouse) พร้อมรายละเอียดแยกตามที่เก็บใน `Locations` ใช้ซ้อนอยู่ใน ProductBalanceByCodeStruct

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| WareHouseCode | string | warehousecode | warehousecode | รหัสคลังสินค้า |
| BalanceQty | float64 | balanceqty | balanceqty | จำนวนคงเหลือ |
| AverageCost | float64 | averagecost | averagecost | ต้นทุนเฉลี่ย |
| BalanceWord | string | balanceword | balanceword | ยอดคงเหลือแบบข้อความ |
| BalanceAmount | float64 | balanceamount | balanceamount | มูลค่าคงเหลือ |
| Locations | [][[ProductBalanceByCodeLocationStruct]] | locations | locations | ยอดคงเหลือแยกตามที่เก็บ |

## ความสัมพันธ์
- ถูก embed ใน [[ProductBalanceByCodeStruct]] (field `WareHouses`)
- Embedded: [[ProductBalanceByCodeLocationStruct]] (slice ใน `Locations`)
- `WareHouseCode` อ้างอิงคลังสินค้า
