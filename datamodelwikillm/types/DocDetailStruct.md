---
source: process-doc-model.go
tags: [datamodel, general-type]
---

# DocDetailStruct

โครงสร้างรายการรายละเอียดเอกสาร (document detail line) ใช้ map ข้อมูลรายการสินค้าต่อบรรทัดจาก MongoDB เช่น บาร์โค้ด หน่วยนับ คลัง/ที่เก็บ จำนวน ราคา และยอดเงินทั้งสกุลเงินหลักและสกุลเงินเอกสาร

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| DocDateTime | time.Time | docdatetime | docdatetime | วันเวลาเอกสาร |
| DocNo | string | docno | docno | เลขที่เอกสาร |
| TransFlag | int | transflag | transflag | ประเภทธุรกรรม (trans flag) |
| CalcFlag | float64 | calcflag | calcflag | flag ทิศทางการคำนวณ |
| CalcSeq | int | calcseq | calcseq | ลำดับการคำนวณ |
| LineNumber | int | linenumber | linenumber | ลำดับบรรทัดรายการ |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| WhCode | string | whcode | whcode | รหัสคลังสินค้า |
| LocationCode | string | locationcode | locationcode | รหัสที่เก็บในคลัง |
| ToWhCode | string | towhcode | towhcode | รหัสคลังปลายทาง |
| ToLocationCode | string | tolocationcode | tolocationcode | รหัสที่เก็บปลายทาง |
| TotalQty | float64 | totalqty | totalqty | จำนวนรวม |
| Price | float64 | price | price | ราคาต่อหน่วย |
| PriceExcludeVat | float64 | priceexcludevat | priceexcludevat | ราคาต่อหน่วยไม่รวม VAT |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| BarcodeMain | string | barcodemain | barcodemain | บาร์โค้ดหลักของสินค้า |
| BarcodeRefUnitStand | float64 | barcoderefunitstand | barcoderefunitstand | ตัวตั้งหน่วยของบาร์โค้ดอ้างอิง |
| BarcodeRefUnitDivide | float64 | barcoderefunitdivide | barcoderefunitdivide | ตัวหารหน่วยของบาร์โค้ดอ้างอิง |
| UnitStand | float64 | unitstand | unitstand | ตัวตั้งหน่วยนับ |
| UnitDivide | float64 | unitdivide | unitdivide | ตัวหารหน่วยนับ |
| DocRef | string | docref | docref | เลขที่เอกสารอ้างอิงของรายการ |
| Description | string | description | description | คำอธิบายรายการ |
| SumAmount | float64 | sumamount | sumamount | ยอดรวมของรายการ |
| IsCalcStock | int8 | iscalcstock | iscalcstock | 1 = มีการคำนวณสต็อก (ซื้อ/ขาย), 0 = ไม่คำนวณ |
| PriceDoc | float64 | pricedoc | pricedoc | ราคาต่อหน่วยในสกุลเงินเอกสาร |
| SumAmountDoc | float64 | sumamountdoc | sumamountdoc | ยอดรวมรายการในสกุลเงินเอกสาร |
| DiscountAmountDoc | float64 | discountamountdoc | discountamountdoc | ส่วนลดในสกุลเงินเอกสาร |
| PriceExcludeVatDoc | float64 | priceexcludevatdoc | priceexcludevatdoc | ราคาไม่รวม VAT ในสกุลเงินเอกสาร |
| SumAmountExcludeVatDoc | float64 | sumamountexcludevatdoc | sumamountexcludevatdoc | ยอดรวมไม่รวม VAT ในสกุลเงินเอกสาร |
| TotalValueVatDoc | float64 | totalvaluevatdoc | totalvaluevatdoc | มูลค่า VAT ในสกุลเงินเอกสาร |

## ความสัมพันธ์

- `DocNo` — เชื่อมกับหัวเอกสาร (ดู [[DocStruct]])
- `Barcode` / `BarcodeMain` — อ้างถึงบาร์โค้ดสินค้า
- `ItemCode` — อ้างถึงรหัสสินค้า
- `UnitCode` — อ้างถึงหน่วยนับ
- `WhCode` / `LocationCode` — อ้างถึงคลังและที่เก็บต้นทาง
- `ToWhCode` / `ToLocationCode` — อ้างถึงคลังและที่เก็บปลายทาง (กรณีโอนย้าย)
- `DocRef` — อ้างถึงเลขที่เอกสารอื่น
