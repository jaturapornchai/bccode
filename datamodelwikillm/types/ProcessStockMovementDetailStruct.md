---
source: process-model.go
tags: [datamodel, general-type]
---

# ProcessStockMovementDetailStruct

รายละเอียดการเคลื่อนไหวสต๊อกหนึ่งบรรทัด (ต่อเอกสาร/คลัง/ที่เก็บ) พร้อมจำนวน ราคา ต้นทุนเฉลี่ย และยอดคงเหลือ

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| IsExtra | bool | isextra | isextra | เป็นรายการพิเศษ (extra) หรือไม่ |
| DocDateTime | time.Time | docdatetime | docdatetime | วันที่/เวลาเอกสาร |
| DocNo | string | docno | docno | เลขที่เอกสาร |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| TransFlag | int | transflag | transflag | ประเภทรายการ (trans flag) |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| WhCode | string | whcode | whcode | รหัสคลังสินค้า |
| LocationCode | string | locationcode | locationcode | รหัสที่เก็บในคลัง |
| TotalQty | float64 | totalqty | totalqty | จำนวนรวม |
| Price | float64 | price | price | ราคา |
| UnitStand | float64 | unitstand | unitstand | ตัวคูณหน่วยมาตรฐาน |
| UnitDivide | float64 | unitdivide | unitdivide | ตัวหารหน่วย |
| AverageCost | float64 | averagecost | averagecost | ต้นทุนเฉลี่ย |
| CalcAmount | float64 | calcamount | calcamount | มูลค่าที่คำนวณได้ |
| BalanceAmount | float64 | balanceamount | balanceamount | มูลค่าคงเหลือ |
| BalanceQty | float64 | balanceqty | balanceqty | จำนวนคงเหลือ |
| UnitCost | float64 | unitcost | unitcost | ต้นทุนต่อหน่วย |
| DocRef | string | docref | docref | เลขที่เอกสารอ้างอิง |

## ความสัมพันธ์
- ถูก embed ใน [[ProcessStockMovementStruct]] (field `Details`)
- `ItemCode` อ้างอิงรหัสสินค้า, `UnitCode` อ้างอิงหน่วยนับ, `WhCode` อ้างอิงคลังสินค้า, `LocationCode` อ้างอิงที่เก็บ, `DocNo`/`DocRef` อ้างอิงเอกสาร
