---
source: backend/internal/product/color/models/color.go
tags: [datamodel, general-type]
---

# Color

โครงสร้าง `Color` จากโมดูล mainapi `color` มีฟิลด์ตามซอร์ส `color.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| PartitionIdentity | [[PartitionIdentity\|models.PartitionIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| Code | string | code | code | รหัสรายการ |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| ColorSelect | string | colorselect | colorselect | ค่าของ ColorSelect ตามฟิลด์ `colorselect` ในซอร์ส |
| ColorSystem | string | colorsystem | colorsystem | ค่าของ ColorSystem ตามฟิลด์ `colorsystem` ในซอร์ส |
| ColorHex | string | colorhex | colorhex | ค่าของ ColorHex ตามฟิลด์ `colorhex` ในซอร์ส |
| ColorSelectHex | string | colorselecthex | colorselecthex | ค่าของ ColorSelectHex ตามฟิลด์ `colorselecthex` ในซอร์ส |
| ColorSystemHex | string | colorsystemhex | colorsystemhex | ค่าของ ColorSystemHex ตามฟิลด์ `colorsystemhex` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[PartitionIdentity]], [[NameX]]
