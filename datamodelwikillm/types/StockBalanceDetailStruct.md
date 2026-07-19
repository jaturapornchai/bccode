---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# StockBalanceDetailStruct

รายการรายละเอียด (detail line) ของเอกสารยอดยกมา [[StockBalanceStruct]] ระบุสินค้า จำนวน ราคา คลัง/ที่เก็บ พร้อมวันที่และประเภทของยอดยกมา

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
| BalanceDate | time.Time | balancedate | balancedate | วันที่สำหรับยอดยกมา |
| BalanceType | string | balancetype | balancetype | ประเภทของยอดยกมา |

## ความสัมพันธ์
- ฝัง [[LanguageModel]] (ItemNames)
- ถูกฝังเป็น `Details` ใน [[StockBalanceStruct]]
- ItemCode / Barcode — อ้างอิงสินค้า/บาร์โค้ดสินค้า
- UnitCode — อ้างอิงหน่วยนับ
- WhCode / LocationCode — อ้างอิงคลังสินค้าและที่เก็บ
- DocRef — อ้างอิงเอกสารอื่น
