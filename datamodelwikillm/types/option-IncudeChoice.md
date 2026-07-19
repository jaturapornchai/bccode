---
source: backend/internal/product/option/models/option.go
tags: [datamodel, general-type]
---

# IncudeChoice

โครงสร้าง `IncudeChoice` จากโมดูล mainapi `option` มีฟิลด์ตามซอร์ส `option.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| ChoiceCode | string | choicecode | choicecode | ค่าของ ChoiceCode ตามฟิลด์ `choicecode` ในซอร์ส |
| Details | [][[option-IncudeChoice\|IncudeChoice]] | choicedetails | choicedetails | ค่าของ Details ตามฟิลด์ `choicedetails` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[option-IncudeChoice|IncudeChoice]]
