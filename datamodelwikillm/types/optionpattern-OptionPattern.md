---
source: backend/internal/product/optionpattern/models/optionpattern.go
tags: [datamodel, general-type]
---

# OptionPattern

โครงสร้าง `OptionPattern` จากโมดูล mainapi `optionpattern` มีฟิลด์ตามซอร์ส `optionpattern.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| PartitionIdentity | [[PartitionIdentity\|models.PartitionIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| PatternCode | string | patterncode | patterncode | ค่าของ PatternCode ตามฟิลด์ `patterncode` ในซอร์ส |
| Name | [[Name\|models.Name]] | inline | - | โครงสร้างฝังแบบ inline |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| OptionPatternDetails | *[][[optionpattern-OptionPatternDetail\|OptionPatternDetail]] | optionpatterndetails | optionpatterndetails | ค่าของ OptionPatternDetails ตามฟิลด์ `optionpatterndetails` ในซอร์ส |
| ColorCode | string | colorcode | colorcode | ค่าของ ColorCode ตามฟิลด์ `colorcode` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[PartitionIdentity]], [[Name]], [[NameX]], [[optionpattern-OptionPatternDetail|OptionPatternDetail]]
