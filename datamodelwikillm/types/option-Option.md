---
source: backend/internal/product/option/models/option.go
tags: [datamodel, general-type]
---

# Option

โครงสร้าง `Option` จากโมดูล mainapi `option` มีฟิลด์ตามซอร์ส `option.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Code | string | code | code | รหัสรายการ |
| XOrder | int8 | xorder | xorder | ลำดับการแสดงผล |
| Required | bool | required | required | ค่าของ Required ตามฟิลด์ `required` ในซอร์ส |
| ChoiceType | int8 | choicetype,omitempty | choicetype | ค่าของ ChoiceType ตามฟิลด์ `choicetype` ในซอร์ส |
| MaxSelect | int8 | maxselect,omitempty | maxselect | ค่าของ MaxSelect ตามฟิลด์ `maxselect` ในซอร์ส |
| Name | [[Name\|models.Name]] | inline | - | โครงสร้างฝังแบบ inline |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| Choices | *[][[option-Choice\|Choice]] | choices | choices | ค่าของ Choices ตามฟิลด์ `choices` ในซอร์ส |
| IsStockControl | bool | isstockcontrol | isstockcontrol | ค่าของ IsStockControl ตามฟิลด์ `isstockcontrol` ในซอร์ส |
| OptionDetails | [][[option-OptionDetail\|OptionDetail]] | optiondetails | optiondetails | ค่าของ OptionDetails ตามฟิลด์ `optiondetails` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[Name]], [[NameX]], [[option-Choice|Choice]], [[option-OptionDetail|OptionDetail]]
