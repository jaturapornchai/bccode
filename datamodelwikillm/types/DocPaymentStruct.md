---
source: process-doc-model.go
tags: [datamodel, general-type]
---

# DocPaymentStruct

โครงสร้างรายการชำระเงินของเอกสาร (document payment) เก็บผู้ให้บริการชำระเงิน จำนวนเงิน และข้อมูลอ้างอิงเอกสาร/สาขา

หมายเหตุ: struct นี้ใช้ tag `json` + `db` (ไม่มี tag `bson`)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| BranchID | string | - | branchid | รหัสสาขา |
| DocDateTime | time.Time | - | docdatetime | วันเวลาเอกสาร |
| PeriodDateTime | time.Time | - | perioddatetime | วันเวลางวด |
| ProviderName | string | - | providername | ชื่อผู้ให้บริการชำระเงิน |
| Amount | float64 | - | amount | จำนวนเงิน |
| Description | string | - | description | คำอธิบายรายการ |
| DocNo | string | - | docno | เลขที่เอกสาร |
| TransFlag | int32 | - | transflag | ประเภทธุรกรรม (trans flag) |
| GuidFixed | string | - | guidfixed | GUID ประจำรายการ |
| GuidBranch | string | - | guidbranch | GUID สาขา |

## ความสัมพันธ์

- `HoldingCode` — อ้างถึงกลุ่มกิจการ (tenant boundary)
- `DocNo` — เชื่อมกับหัวเอกสาร (ดู [[DocStruct]])
- `BranchID` / `GuidBranch` — อ้างถึงสาขา
