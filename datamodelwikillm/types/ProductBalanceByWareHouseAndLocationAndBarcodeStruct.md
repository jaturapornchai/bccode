---
source: process-model.go
tags: [datamodel, general-type]
---

# ProductBalanceByWareHouseAndLocationAndBarcodeStruct

ยอดคงเหลือสินค้าแยกตามคลังสินค้า ที่เก็บ และบาร์โค้ด (flat ต่อ row)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| WareHouseCode | string | whcode | whcode | รหัสคลังสินค้า |
| LocationCode | string | locationcode | locationcode | รหัสที่เก็บ |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| ItemName | string | itemname | itemname | ชื่อสินค้า |
| BarcodeList | string | barcodelist | barcodelist | รายการบาร์โค้ดของสินค้า |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| UnitName | string | unitname | unitname | ชื่อหน่วยนับ |
| BalanceQty | float64 | balanceqty | balanceqty | จำนวนคงเหลือ |
| BalanceWord | string | balanceword | balanceword | ยอดคงเหลือแบบข้อความ |
| IsAutoPacking | uint64 | isautopacking | isautopacking | ธง auto packing |

## ความสัมพันธ์
- `WareHouseCode` อ้างอิงคลังสินค้า, `LocationCode` อ้างอิงที่เก็บ, `ItemCode` อ้างอิงสินค้า, `BarcodeList` อ้างอิงบาร์โค้ด, `UnitCode` อ้างอิงหน่วยนับ
- โครงสร้างใกล้เคียง [[ProductBalanceByWareHouseAndBarcodeStruct]] (เพิ่ม `LocationCode`)
