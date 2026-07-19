---
source: process-model.go
tags: [datamodel, general-type]
---

# ProcessStockMovementStruct

โครงสร้างข้อมูลความเคลื่อนไหวสต๊อกของสินค้าหนึ่งรายการ (item) พร้อมรายการเคลื่อนไหวย่อยใน `Details` ใช้ในกระบวนการ process สต๊อก

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| Name | string | name | name | ชื่อสินค้า |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| UnitName | string | unitname | unitname | ชื่อหน่วยนับ |
| Details | [][[ProcessStockMovementDetailStruct]] | details | details | รายการเคลื่อนไหวสต๊อกแต่ละบรรทัด |

## ความสัมพันธ์
- Embedded: [[ProcessStockMovementDetailStruct]] (slice ใน `Details`)
- `ItemCode` อ้างอิงรหัสสินค้า, `UnitCode` อ้างอิงรหัสหน่วยนับ
