---
source: process-doc-purchase-model.go
tags: [datamodel, general-type]
---

# PurchaseStatusDocDetailStruct

รายละเอียดรายบรรทัดของสถานะเอกสารซื้อ เก็บข้อมูลสินค้า จำนวนสั่ง/รับแล้ว/ค้างรับ พร้อมรายการเอกสารรับสินค้าที่อ้างอิง

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocNo | string | docno | docno | เลขที่เอกสาร |
| TransFlag | int | transflag | transflag | ประเภทเอกสาร (trans flag) |
| CalcFlag | float64 | calcflag | calcflag | ทิศทางการคำนวณ (calc flag) |
| CalcSeq | int | calcseq | calcseq | ลำดับการคำนวณ |
| LineNumber | int | linenumber | linenumber | ลำดับบรรทัดในเอกสาร |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| WhCode | string | whcode | whcode | รหัสคลังสินค้า |
| LocationCode | string | locationcode | locationcode | รหัสที่เก็บ |
| ToWhCode | string | towhcode | towhcode | รหัสคลังสินค้าปลายทาง |
| ToLocationCode | string | tolocationcode | tolocationcode | รหัสที่เก็บปลายทาง |
| TotalQty | float64 | totalqty | totalqty | จำนวนรวมตามเอกสาร |
| Price | float64 | price | price | ราคา |
| PriceExcludeVat | float64 | priceexcludevat | priceexcludevat | ราคาไม่รวม VAT |
| ReceivedQty | float64 | - | receivedqty | จำนวนที่รับมาแล้ว |
| PendingQty | float64 | - | pendingqty | จำนวนที่รอดำเนินการ |
| ProductName | string | - | productname | ชื่อสินค้า |
| UnitName | string | - | unitname | ชื่อหน่วยนับ |
| DocRefer | [][[PurchaseStatusDocDetailReferStruct]] | - | docrefer | รายการอ้างอิงการรับสินค้า |

## ความสัมพันธ์

- `DocRefer` → embed [[PurchaseStatusDocDetailReferStruct]] (เอกสารรับสินค้าที่อ้างถึงบรรทัดนี้)
- `ItemCode` → อ้างอิงรหัสสินค้า
- `UnitCode` → อ้างอิงรหัสหน่วยนับ
- `WhCode` / `ToWhCode` → อ้างอิงรหัสคลังสินค้า (ต้นทาง/ปลายทาง)
- `LocationCode` / `ToLocationCode` → อ้างอิงรหัสที่เก็บ (ต้นทาง/ปลายทาง)
