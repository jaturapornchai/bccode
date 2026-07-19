---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# LanguageModel

โครงสร้างชื่อแบบหลายภาษา เก็บรหัสภาษาและชื่อในภาษานั้น ใช้ฝัง (embed) เป็น array ในโมเดลอื่นๆ เช่นชื่อสินค้า ชื่อหน่วยนับ ชื่อคลัง

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Code | string | code | code | รหัสภาษา (เช่น th, en) |
| Name | string | name | name | ชื่อ/ข้อความในภาษานั้น |
| IsAuto | bool | isauto | isauto | เป็นค่าที่สร้างอัตโนมัติหรือไม่ |
| IsDelete | bool | isdelete | isdelete | ถูกลบ (soft delete) หรือไม่ |

## ความสัมพันธ์
- ถูกฝังเป็น `[]LanguageModel` ในหลายโมเดล เช่น [[MongoDocDetailModel]], [[MongoBranchModel]], [[StockReceiveProductDetailStruct]] ฯลฯ
