---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# ProcessMongoTransDetailTransFlag54Model

รายการรายละเอียดธุรกรรมสำหรับ TransFlag 54 (ยอดยกมา) ตาม comment ในซอร์ส เป็นเวอร์ชันย่อของ detail line ที่ใช้ในการประมวลผลยอดยกมา

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| ItemNames | [][[LanguageModel]] | itemnames | itemnames | ชื่อสินค้าหลายภาษา |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| LineNumber | int | linenumber | linenumber | ลำดับบรรทัดรายการ |
| WhCode | string | whcode | whcode | รหัสคลังสินค้า |
| LocationCode | string | locationcode | locationcode | รหัสที่เก็บ |
| ToWhCode | string | towhcode | towhcode | รหัสคลังปลายทาง |
| ToLocationCode | string | tolocationcode | tolocationcode | รหัสที่เก็บปลายทาง |
| Qty | float64 | qty | qty | จำนวน |
| CalcFlag | int | calcflag | calcflag | flag การคำนวณสต๊อก |
| CalcSeq | int | calcseq | calcseq | ลำดับการคำนวณ |
| Price | float64 | price | price | ราคาต่อหน่วย |
| PriceExcludeVat | float64 | priceexcludevat | priceexcludevat | ราคาไม่รวม VAT |
| DocRef | string | docref | docref | เลขที่เอกสารอ้างอิง |
| SumAmount | float64 | sumamount | sumamount | ยอดรวมบรรทัด |
| DocNo | string | docno | docno | เลขที่เอกสาร |
| DocDateTime | time.Time | docdatetime | docdatetime | วันเวลาเอกสาร |

## ความสัมพันธ์
- ฝัง [[LanguageModel]] (ItemNames)
- ItemCode / Barcode — อ้างอิงสินค้า/บาร์โค้ดสินค้า
- UnitCode — อ้างอิงหน่วยนับ
- WhCode / LocationCode / ToWhCode / ToLocationCode — อ้างอิงคลังสินค้าและที่เก็บ (ต้นทาง/ปลายทาง)
- DocNo / DocRef — เลขที่เอกสารและเอกสารอ้างอิง
