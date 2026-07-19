---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# StockTransferStruct

เอกสารโอนย้ายสินค้าระหว่างคลัง/ที่เก็บ (stock transfer document ตาม comment ในซอร์ส) เก็บ header ของเอกสารพร้อมรายการรายละเอียด

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | holdingcode | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| BranchCode | string | branchcode | branchcode | รหัสสาขา |
| DocNo | string | docno | docno | เลขที่เอกสาร |
| RefNo | string | refno | refno | เลขที่อ้างอิง |
| Description | string | description | description | คำอธิบายเอกสาร |
| DocDateTime | time.Time | docdatetime | docdatetime | วันเวลาเอกสาร |
| TotalAmount | float64 | totalamount | totalamount | ยอดรวมเอกสาร |
| CustCode | string | custcode | custcode | รหัสลูกค้า/คู่ค้า |
| IsCancel | bool | iscancel | iscancel | เอกสารถูกยกเลิกหรือไม่ |
| Guid | string | guid | guid | GUID ของเอกสาร |
| DocCreatedDateTime | time.Time | doccreateddatetime | doccreateddatetime | วันเวลาที่สร้างเอกสาร |
| DocUpdatedDateTime | time.Time | docupdateddatetime | docupdateddatetime | วันเวลาที่แก้ไขเอกสาร |
| IsDelete | bool | isdelete | isdelete | ถูกลบ (soft delete) หรือไม่ |
| Branch | [[MongoBranchModel]] | branch | branch | ข้อมูลสาขา |
| Details | [][[StockTransferDetailStruct]] | details | details | รายการรายละเอียดการโอนย้าย |

## ความสัมพันธ์
- ฝัง [[MongoBranchModel]] (Branch), [[StockTransferDetailStruct]] (Details)
- HoldingCode — อ้างอิงกลุ่มกิจการ (tenant boundary)
- BranchCode — อ้างอิงสาขา
- CustCode — อ้างอิงลูกค้า/คู่ค้า
- RefNo — เลขที่เอกสารอ้างอิง
