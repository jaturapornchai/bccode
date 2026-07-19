---
source: backend/internal/product/unit/models/unit.go
tags: [datamodel, general-type]
---

# Unit

โครงสร้าง `Unit` จากโมดูล mainapi `unit` มีฟิลด์ตามซอร์ส `unit.go`

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| PartitionIdentity | [[PartitionIdentity\|models.PartitionIdentity]] | inline | - | โครงสร้างฝังแบบ inline |
| UnitCode | string | unitcode | unitcode | รหัสหน่วยนับ |
| UnitName | [[UnitName\|models.UnitName]] | inline | - | โครงสร้างฝังแบบ inline |
| Names | *[][[NameX\|models.NameX]] | names | names | รายการชื่อหลายภาษา |
| CompanyGuids | []string | companyguids | companyguids | ค่าของ CompanyGuids ตามฟิลด์ `companyguids` ในซอร์ส |

## ความสัมพันธ์

- Struct ที่ฝังหรืออ้างอิง: [[PartitionIdentity]], [[UnitName]], [[NameX]]
