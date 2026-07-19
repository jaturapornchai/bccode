---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# MongoBranchModel

ข้อมูลสาขาแบบย่อ (รหัส, GUID, ชื่อหลายภาษา) ใช้ฝังในเอกสารธุรกรรมต่างๆ มี type alias `BranchModel` ชี้มาที่ struct นี้

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| Code | string | code | code | รหัสสาขา |
| GuidFixed | string | guidfixed | guidfixed | GUID ถาวรของสาขา |
| Names | [][[LanguageModel]] | names | names | ชื่อสาขาหลายภาษา |

## ความสัมพันธ์
- ฝัง [[LanguageModel]] (Names)
- ถูกฝังเป็น `Branch` ใน [[MongoDocModel]], [[StockTransferStruct]], [[StockReceiveProductStruct]], [[StockPickupProductStruct]], [[StockReturnProductStruct]], [[StockAdjustmentStruct]], [[StockBalanceStruct]]
