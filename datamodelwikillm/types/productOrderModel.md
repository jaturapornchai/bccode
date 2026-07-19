---
source: backend/internal/goapi/models/global-model.go
tags: [datamodel, general-type]
---

# productOrderModel

โครงสร้างข้อมูลสินค้าสำหรับการสั่งซื้อ/แสดงขาย ประกอบด้วยชื่อหลายภาษา หน่วยนับ ราคา รูปภาพ และตัวเลือกสินค้า (options) มีฟังก์ชัน `DataProductForTest()` สร้างข้อมูลตัวอย่างสำหรับทดสอบ

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| Barcode | string | - | barcode | บาร์โค้ดสินค้า |
| Itemcode | string | - | itemcode | รหัสสินค้า |
| Names | [][[languageNameModel]] | - | names | ชื่อสินค้าหลายภาษา |
| Unituses | [][[productOrderUnitUseModel]] | - | unituses | หน่วยนับที่ใช้ได้ พร้อมอัตราแปลงหน่วย |
| Unitcode | string | - | unitcode | รหัสหน่วยนับหลัก |
| Units | [][[productOrderUnitModel]] | - | units | รายการหน่วยนับพร้อมชื่อหลายภาษา |
| Unitcost | string | - | unitcost | รหัสหน่วยนับต้นทุน |
| Unitstandard | string | - | unitstandard | รหัสหน่วยนับมาตรฐาน |
| Multiunit | bool | - | multiunit | ใช้หลายหน่วยนับหรือไม่ |
| Itemtype | int | - | itemtype | ประเภทสินค้า |
| Itemvat | int | - | itemvat | ประเภท VAT ของสินค้า |
| Normalprice | float64 | - | normalprice | ราคาปกติ |
| Price | float64 | - | price | ราคาขาย |
| Memberprice | float64 | - | memberprice | ราคาสมาชิก |
| Pricerangemin | float64 | - | pricerangemin | ราคาต่ำสุดของช่วงราคา |
| Pricerangemax | float64 | - | pricerangemax | ราคาสูงสุดของช่วงราคา |
| Images | [][[productOrderImageModel]] | - | images | รูปภาพสินค้า (URI) |
| Recommended | bool | - | recommended | สินค้าแนะนำ |
| Shoprecommended | bool | - | shoprecommended | สินค้าแนะนำโดยร้านค้า |
| Havepoint | bool | - | havepoint | มีคะแนนสะสมหรือไม่ |
| Starpersent | float64 | - | starpersent | เปอร์เซ็นต์คะแนนดาว |
| Ordercount | int | - | ordercount | จำนวนครั้งที่ถูกสั่งซื้อ |
| Descriptions | [][[languageNameModel]] | - | descriptions | คำอธิบายสินค้าหลายภาษา |
| Options | [][[productOrderOptionModel]] | - | options | ตัวเลือกสินค้า (เช่น สี ขนาด) |
| Orderminimum | float64 | - | orderminimum | จำนวนสั่งซื้อขั้นต่ำ |

## ความสัมพันธ์
- ฝัง [[languageNameModel]] (Names, Descriptions), [[productOrderUnitUseModel]] (Unituses), [[productOrderUnitModel]] (Units), [[productOrderImageModel]] (Images), [[productOrderOptionModel]] (Options)
- `HoldingCode` — อ้างอิงกลุ่มกิจการ (tenant)
- `Barcode` — อ้างอิงบาร์โค้ดสินค้า
- `Itemcode` — อ้างอิงรหัสสินค้า
- `Unitcode` / `Unitcost` / `Unitstandard` — อ้างอิงรหัสหน่วยนับ
