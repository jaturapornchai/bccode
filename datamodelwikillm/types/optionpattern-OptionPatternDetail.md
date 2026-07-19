---
source: backend/internal/product/optionpattern/models/optionpattern.go
tags: [datamodel, general-type]
---

# OptionPatternDetail

โครงสร้าง `OptionPatternDetail` จากโมดูล mainapi `optionpattern` มีฟิลด์ตามซอร์ส `optionpattern.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| XOrder | int8 | xorder | xorder | ลำดับการแสดงผล |
| OptionCode | string | optioncode | optioncode | ค่าของ OptionCode ตามฟิลด์ `optioncode` ในซอร์ส |
| Option | *[[option-Option\|optionModel.Option]] | inline | - | โครงสร้างฝังแบบ inline |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[option-Option|Option]]
