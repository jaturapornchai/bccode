---
source: backend/internal/product/option/models/option.go
tags: [datamodel, general-type]
---

# OptionDetail

โครงสร้าง `OptionDetail` จากโมดูล mainapi `option` มีฟิลด์ตามซอร์ส `option.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| OptionDetailCode | string | optiondetailcode | optiondetailcode | ค่าของ OptionDetailCode ตามฟิลด์ `optiondetailcode` ในซอร์ส |
| Name | [[Name\|models.Name]] | inline | - | โครงสร้างฝังแบบ inline |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| Image | string | image | image | ค่าของ Image ตามฟิลด์ `image` ในซอร์ส |
| ChoiceDetails | [][[option-IncudeChoice\|IncudeChoice]] | choicedetails | choicedetails | ค่าของ ChoiceDetails ตามฟิลด์ `choicedetails` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[Name]], [[NameX]], [[option-IncudeChoice|IncudeChoice]]
