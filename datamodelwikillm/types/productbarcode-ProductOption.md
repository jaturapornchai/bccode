---
source: backend/internal/product/productbarcode/models/product_option.go
tags: [datamodel, general-type]
---

# ProductOption

โครงสร้าง `ProductOption` จากโมดูล mainapi `productbarcode` มีฟิลด์ตามซอร์ส `product_option.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GUID | string | guid | guid | GUID ของรายการ |
| ChoiceType | int8 | choicetype | choicetype | ค่าของ ChoiceType ตามฟิลด์ `choicetype` ในซอร์ส |
| MaxSelect | uint16 | maxselect | maxselect | ค่าของ MaxSelect ตามฟิลด์ `maxselect` ในซอร์ส |
| MinSelect | uint16 | minselect | minselect | ค่าของ MinSelect ตามฟิลด์ `minselect` ในซอร์ส |
| Names | *[]models.NameX | names | names | รายการชื่อหลายภาษา |
| Choices | *[][[productbarcode-ProductChoice\|ProductChoice]] | choices | choices | ค่าของ Choices ตามฟิลด์ `choices` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[productbarcode-ProductChoice|ProductChoice]]
