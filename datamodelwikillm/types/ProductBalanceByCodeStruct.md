---
source: process-model.go
tags: [datamodel, general-type]
---

# ProductBalanceByCodeStruct

ยอดคงเหลือสินค้าตามรหัสสินค้า (by code) พร้อมรายละเอียดแยกตามคลังใน `WareHouses`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| ItemName | string | itemname | itemname | ชื่อสินค้า |
| BarcodeList | string | barcodelist | barcodelist | รายการบาร์โค้ดของสินค้า |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| UnitName | string | unitname | unitname | ชื่อหน่วยนับ |
| BalanceQty | float64 | balanceqty | balanceqty | จำนวนคงเหลือ |
| AverageCost | float64 | averagecost | averagecost | ต้นทุนเฉลี่ย |
| BalanceAmount | float64 | balanceamount | balanceamount | มูลค่าคงเหลือ |
| BalanceWord | string | balanceword | balanceword | ยอดคงเหลือแบบข้อความ |
| IsAutoPacking | uint64 | isautopacking | isautopacking | ธง auto packing |
| WareHouses | [][[ProductBalanceByCodeWareHouseStruct]] | warehouses | warehouses | ยอดคงเหลือแยกตามคลัง |

## ความสัมพันธ์
- Embedded: [[ProductBalanceByCodeWareHouseStruct]] (slice ใน `WareHouses`) ซึ่งมี [[ProductBalanceByCodeLocationStruct]] ซ้อนอยู่อีกชั้น
- `ItemCode` อ้างอิงสินค้า, `BarcodeList` อ้างอิงบาร์โค้ด, `UnitCode` อ้างอิงหน่วยนับ
