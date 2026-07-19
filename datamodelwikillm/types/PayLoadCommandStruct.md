---
source: process-model.go
tags: [datamodel, general-type]
---

# PayLoadCommandStruct

Payload คำสั่งสำหรับสั่ง process สต๊อก/สร้างฐานข้อมูล — ระบุกิจการ ช่วงวันที่ เงื่อนไข รายการสินค้า/คลัง และตัวเลือกการคำนวณ (มีเฉพาะ json tag)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกิจการ (tenant) |
| CommandID | string | - | commandid | รหัสคำสั่ง |
| DocNumberList | []string | - | docnumberlist | รายการเลขที่เอกสาร |
| Condition | string | - | condition | เงื่อนไขการประมวลผล |
| BalanceOnly | string | - | balanceonly | ประมวลผลเฉพาะยอดคงเหลือ |
| MovementOnly | string | - | movementonly | ประมวลผลเฉพาะความเคลื่อนไหว |
| ItemCodeList | json.RawMessage | - | itemcodelist | รายการรหัสสินค้า (raw JSON) |
| BarcodeList | string | - | barcodelist | รายการบาร์โค้ด |
| WarehouseList | [][[WarehouseListItemStruct]] | - | warehouselist | รายการคลังสินค้าที่เลือก |
| FromDate | string | - | fromdate | วันที่เริ่มต้น |
| FinalDate | string | - | finaldate | วันที่สิ้นสุด |
| TimezoneCode | string | - | timezonecode | รหัส timezone |
| LanguageCode | string | - | languagecode | รหัสภาษา |
| Guid | string | - | guid,omitempty | GUID ของคำสั่ง |
| PointQty | *int | - | pointqty | จำนวนทศนิยมของจำนวน (point) |
| PointAmount | *int | - | pointamount | จำนวนทศนิยมของมูลค่า (point) |
| PointCost | *int | - | pointcost | จำนวนทศนิยมของต้นทุน (point) |
| DeleteFirst | *bool | - | deletefirst | ลบข้อมูลก่อนประมวลผลหรือไม่ |
| CreateDatabase | *bool | - | createdatabase | true = drop+create ใหม่ (default), false = calc only (ตาม comment ในโค้ด) |

## ความสัมพันธ์
- Embedded: [[WarehouseListItemStruct]] (slice ใน `WarehouseList`) ซึ่งมี [[LocationItemStruct]] ซ้อนอยู่
- `HoldingCode` อ้างอิงกิจการ (tenant), `DocNumberList` อ้างอิงเลขที่เอกสาร, `ItemCodeList`/`BarcodeList` อ้างอิงสินค้า/บาร์โค้ด
