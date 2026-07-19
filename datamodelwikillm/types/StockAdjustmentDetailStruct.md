---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# StockAdjustmentDetailStruct

รายการรายละเอียด (detail line) ของเอกสารปรับปรุงสต๊อก [[StockAdjustmentStruct]] ระบุสินค้า จำนวน ประเภทการปรับ (เพิ่ม/ลด) และเหตุผล

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| LineNumber | int | linenumber | linenumber | ลำดับบรรทัดรายการ |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| ItemNames | [][[LanguageModel]] | itemnames | itemnames | ชื่อสินค้าหลายภาษา |
| Description | string | description | description | คำอธิบายรายการ |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| WhCode | string | whcode | whcode | รหัสคลังสินค้า |
| LocationCode | string | locationcode | locationcode | รหัสที่เก็บ |
| Qty | float64 | qty | qty | จำนวน |
| Price | float64 | price | price | ราคาต่อหน่วย |
| PriceExcludeVat | float64 | priceexcludevat | priceexcludevat | ราคาไม่รวม VAT |
| DocRef | string | docref | docref | เลขที่เอกสารอ้างอิง |
| SumAmount | float64 | sumamount | sumamount | ยอดรวมบรรทัด |
| AdjustmentType | string | adjustmenttype | adjustmenttype | ประเภทการปรับ: "INCREASE" หรือ "DECREASE" |
| Reason | string | reason | reason | เหตุผลในการปรับ |

## ความสัมพันธ์
- ฝัง [[LanguageModel]] (ItemNames)
- ถูกฝังเป็น `Details` ใน [[StockAdjustmentStruct]]
- ItemCode / Barcode — อ้างอิงสินค้า/บาร์โค้ดสินค้า
- UnitCode — อ้างอิงหน่วยนับ
- WhCode / LocationCode — อ้างอิงคลังสินค้าและที่เก็บ
- DocRef — อ้างอิงเอกสารอื่น
