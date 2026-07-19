---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# MongoDocDetailModel

รายการรายละเอียด (detail line) ของเอกสารธุรกรรมใน [[MongoDocModel]] เก็บข้อมูลสินค้า จำนวน ราคา ส่วนลด ภาษี คลัง/ที่เก็บ รวมถึงราคาในสกุลเงินเอกสาร (Document Currency) มี type alias `ProcessMongoTransDetailModel` ชี้มาที่ struct นี้

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| LineNumber | int | linenumber | linenumber | ลำดับบรรทัดรายการ |
| DocDateTime | time.Time | docdatetime | docdatetime | วันเวลาเอกสาร |
| DocRef | string | docref | docref | เลขที่เอกสารอ้างอิง |
| DocRefDateTime | time.Time | docrefdatetime | docrefdatetime | วันเวลาเอกสารอ้างอิง |
| Barcode | string | barcode | barcode | บาร์โค้ดสินค้า |
| ItemCode | string | itemcode | itemcode | รหัสสินค้า |
| ItemNames | [][[LanguageModel]] | itemnames | itemnames | ชื่อสินค้าหลายภาษา |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| UnitNames | [][[LanguageModel]] | unitnames | unitnames | ชื่อหน่วยนับหลายภาษา |
| ItemType | int | itemtype | itemtype | ประเภทสินค้า |
| ItemGuid | string | itemguid | itemguid | GUID ของสินค้า |
| Description | string | description | description | คำอธิบายรายการ |
| Qty | float64 | qty | qty | จำนวน |
| EventQty | float64 | eventqty | eventqty | จำนวนตามเหตุการณ์ |
| TotalQty | float64 | totalqty | totalqty | จำนวนรวม |
| Price | float64 | price | price | ราคาต่อหน่วย |
| PriceExcludeVat | float64 | priceexcludevat | priceexcludevat | ราคาไม่รวม VAT |
| Discount | string | discount | discount | ส่วนลด (รูปแบบข้อความ) |
| DiscountAmount | float64 | discountamount | discountamount | มูลค่าส่วนลด |
| TotalValueVat | float64 | totalvaluevat | totalvaluevat | มูลค่าภาษี |
| SumAmount | float64 | sumamount | sumamount | ยอดรวมบรรทัด |
| SumAmountExcludeVat | float64 | sumamountexcludevat | sumamountexcludevat | ยอดรวมไม่รวม VAT |
| SumAmountChoice | float64 | sumamountchoice | sumamountchoice | ยอดรวมแบบ choice |
| RefGuid | string | refguid | refguid | GUID อ้างอิง |
| DivideValue | float64 | dividevalue | dividevalue | ตัวหารหน่วยนับ |
| StandValue | float64 | standvalue | standvalue | ตัวคูณหน่วยนับมาตรฐาน |
| VatType | int | - | vattype | ประเภท VAT |
| Remark | string | - | remark | หมายเหตุ |
| MultiUnit | bool | - | multiunit | ใช้หลายหน่วยนับหรือไม่ |
| IsSumPoint | bool | - | issumpoint | นับรวมแต้มหรือไม่ |
| SumOfCost | float64 | - | sumofcost | ต้นทุนรวม |
| AverageCost | float64 | - | averagecost | ต้นทุนเฉลี่ย |
| FoodType | int | - | foodtype | ประเภทอาหาร |
| LastStatus | int | - | laststatus | สถานะล่าสุด |
| IsChoice | int | - | ischoice | เป็นตัวเลือก (choice) หรือไม่ |
| IsPos | int | - | ispos | มาจาก POS หรือไม่ |
| TaxType | int | - | taxtype | ประเภทภาษี |
| VatCal | int | - | vatcal | วิธีคำนวณ VAT |
| WhCode | string | - | whcode | รหัสคลังสินค้า |
| WhNames | [][[LanguageModel]] | - | whnames | ชื่อคลังหลายภาษา |
| ShelfCode | string | - | shelfcode | รหัสชั้นวาง |
| LocationCode | string | - | locationcode | รหัสที่เก็บ |
| LocationNames | [][[LanguageModel]] | - | locationnames | ชื่อที่เก็บหลายภาษา |
| ToWhCode | string | - | towhcode | รหัสคลังปลายทาง |
| ToWhNames | [][[LanguageModel]] | - | towhnames | ชื่อคลังปลายทางหลายภาษา |
| ToLocationCode | string | - | tolocationcode | รหัสที่เก็บปลายทาง |
| ToLocationNames | [][[LanguageModel]] | - | tolocationnames | ชื่อที่เก็บปลายทางหลายภาษา |
| Sku | string | - | sku | SKU สินค้า |
| ExtraJson | string | - | extrajson | ข้อมูลเพิ่มเติมแบบ JSON |
| GroupCode | string | - | groupcode | รหัสกลุ่มสินค้า |
| GroupNames | [][[LanguageModel]] | - | groupnames | ชื่อกลุ่มสินค้าหลายภาษา |
| ManufacturerGuid | string | - | manufacturerguid | GUID ผู้ผลิต |
| ManufacturerCode | string | - | manufacturercode | รหัสผู้ผลิต |
| ManufacturerNames | [][[LanguageModel]] | - | manufacturernames | ชื่อผู้ผลิตหลายภาษา |
| CalcFlag | int | calcflag | calcflag | flag การคำนวณสต๊อก |
| CalcSeq | int | calcseq | calcseq | ลำดับการคำนวณ |
| PriceDoc | float64 | pricedoc | pricedoc | ราคาต่อหน่วยในสกุลเงินเอกสาร (mapstructure: price_doc) |
| SumAmountDoc | float64 | sumamountdoc | sumamountdoc | ยอดรวมในสกุลเงินเอกสาร (mapstructure: sumamount_doc) |
| DiscountAmountDoc | float64 | discountamountdoc | discountamountdoc | ส่วนลดในสกุลเงินเอกสาร (mapstructure: discountamount_doc) |
| PriceExcludeVatDoc | float64 | priceexcludevatdoc | priceexcludevatdoc | ราคาไม่รวม VAT ในสกุลเงินเอกสาร (mapstructure: priceexcludevat_doc) |
| SumAmountExcludeVatDoc | float64 | sumamountexcludevatdoc | sumamountexcludevatdoc | ยอดรวมไม่รวม VAT ในสกุลเงินเอกสาร (mapstructure: sumamountexcludevat_doc) |
| TotalValueVatDoc | float64 | totalvaluevatdoc | totalvaluevatdoc | มูลค่า VAT ในสกุลเงินเอกสาร (mapstructure: totalvaluevat_doc) |

## ความสัมพันธ์
- ฝัง [[LanguageModel]] หลายจุด (ItemNames, UnitNames, WhNames, LocationNames, ToWhNames, ToLocationNames, GroupNames, ManufacturerNames)
- ถูกฝังเป็น `Details` ใน [[MongoDocModel]]
- ItemCode / Barcode — อ้างอิงสินค้า/บาร์โค้ดสินค้า
- UnitCode — อ้างอิงหน่วยนับ
- WhCode / ToWhCode, LocationCode / ToLocationCode — อ้างอิงคลังสินค้าและที่เก็บ (ต้นทาง/ปลายทาง)
- GroupCode — อ้างอิงกลุ่มสินค้า
- ManufacturerCode / ManufacturerGuid — อ้างอิงผู้ผลิต
- DocRef / RefGuid — อ้างอิงเอกสารอื่น
