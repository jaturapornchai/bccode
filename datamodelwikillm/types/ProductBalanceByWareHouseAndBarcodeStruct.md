---
source: process-model.go
tags: [datamodel, general-type]
---

# ProductBalanceByWareHouseAndBarcodeStruct

ยอดคงเหลือสินค้าแยกตามคลังสินค้าและบาร์โค้ด (flat ต่อ row)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| WareHouseCode | string | whcode | whcode | รหัสคลังสินค้า |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| ItemName | string | itemname | itemname | ชื่อสินค้า |
| BarcodeList | string | barcodelist | barcodelist | รายการบาร์โค้ดของสินค้า |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| UnitName | string | unitname | unitname | ชื่อหน่วยนับ |
| BalanceQty | float64 | balanceqty | balanceqty | จำนวนคงเหลือ |
| BalanceWord | string | balanceword | balanceword | ยอดคงเหลือแบบข้อความ |
| IsAutoPacking | uint64 | isautopacking | isautopacking | ธง auto packing |

## ความสัมพันธ์
- `WareHouseCode` อ้างอิงคลังสินค้า, `ItemCode` อ้างอิงสินค้า, `BarcodeList` อ้างอิงบาร์โค้ด, `UnitCode` อ้างอิงหน่วยนับ
- โครงสร้างใกล้เคียง [[ProductBalanceByWareHouseAndLocationAndBarcodeStruct]] (เพิ่มระดับ location)
