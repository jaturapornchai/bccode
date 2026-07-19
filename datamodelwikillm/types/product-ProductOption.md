---
source: backend/internal/product/product/models/product.go
tags: [datamodel, general-type]
---

# ProductOption

โครงสร้าง `ProductOption` จากโมดูล mainapi `product` มีฟิลด์ตามซอร์ส `product.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| GUID | string | guid | guid | GUID ของรายการ |
| ChoiceType | int8 | choicetype | choicetype | ค่าของ ChoiceType ตามฟิลด์ `choicetype` ในซอร์ส |
| MaxSelect | uint16 | maxselect | maxselect | ค่าของ MaxSelect ตามฟิลด์ `maxselect` ในซอร์ส |
| MinSelect | uint16 | minselect | minselect | ค่าของ MinSelect ตามฟิลด์ `minselect` ในซอร์ส |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| Choices | *[][[product-ProductChoice\|ProductChoice]] | choices | choices | ค่าของ Choices ตามฟิลด์ `choices` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[NameX]], [[product-ProductChoice|ProductChoice]]
