---
source: backend/internal/product/productcategory/models/productcategory.go
tags: [datamodel, general-type]
---

# ProductCategory

โครงสร้าง `ProductCategory` จากโมดูล mainapi `productcategory` มีฟิลด์ตามซอร์ส `productcategory.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| PartitionIdentity | [[PartitionIdentity\|models.PartitionIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| ChildCount | int | childcount | childcount | จำนวนรายการลูก |
| ParentGUID | string | parentguid | parentguid | GUID ของรายการแม่ |
| ParentGUIDAll | string | parentguidall | parentguidall | ลำดับ GUID ของรายการแม่ทั้งหมด |
| ImageUri | string | imageuri | imageuri | URI ของรูปภาพ |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| XSorts | *[][[XSort\|models.XSort]] | xsorts | xsorts | ค่าของ XSorts ตามฟิลด์ `xsorts` ในซอร์ส |
| CodeList | *[][[productcategory-CodeXSort\|CodeXSort]] | codelist | codelist | ค่าของ CodeList ตามฟิลด์ `codelist` ในซอร์ส |
| UseImageOrColor | bool | useimageorcolor | useimageorcolor | ค่าของ UseImageOrColor ตามฟิลด์ `useimageorcolor` ในซอร์ส |
| ColorSelect | string | colorselect | colorselect | ค่าของ ColorSelect ตามฟิลด์ `colorselect` ในซอร์ส |
| ColorSelectHex | string | colorselecthex | colorselecthex | ค่าของ ColorSelectHex ตามฟิลด์ `colorselecthex` ในซอร์ส |
| IsDisabled | bool | isdisabled | isdisabled | ค่าของ IsDisabled ตามฟิลด์ `isdisabled` ในซอร์ส |
| CoverURI | string | coveruri | coveruri | ค่าของ CoverURI ตามฟิลด์ `coveruri` ในซอร์ส |
| GroupNumber | int | groupnumber | groupnumber | ค่าของ GroupNumber ตามฟิลด์ `groupnumber` ในซอร์ส |
| TimeForSales | *[][[productcategory-ProductCategoryTimeForSale\|ProductCategoryTimeForSale]] | timeforsales | timeforsales | ค่าของ TimeForSales ตามฟิลด์ `timeforsales` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[PartitionIdentity]], [[NameX]], [[XSort]], [[productcategory-CodeXSort|CodeXSort]], [[productcategory-ProductCategoryTimeForSale|ProductCategoryTimeForSale]]
