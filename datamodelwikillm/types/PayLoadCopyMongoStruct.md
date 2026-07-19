---
source: process-model.go
tags: [datamodel, general-type]
---

# PayLoadCopyMongoStruct

Payload สำหรับคำสั่ง copy ข้อมูล Mongo จากกิจการ/สภาพแวดล้อมต้นทางไปปลายทาง (มีเฉพาะ json tag)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| SourceHoldingCode | string | - | sourceholdingcode | รหัสกิจการต้นทาง |
| TargetHoldingCode | string | - | targetholdingcode | รหัสกิจการปลายทาง |
| SourceEnvironment | string | - | sourceenvironment | สภาพแวดล้อมต้นทาง (environment) |
| TargetEnvironment | string | - | targetenvironment | สภาพแวดล้อมปลายทาง (environment) |

## ความสัมพันธ์
- `SourceHoldingCode` / `TargetHoldingCode` อ้างอิงรหัสกิจการ (tenant) ต้นทางและปลายทาง
