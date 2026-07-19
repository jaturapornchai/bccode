---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# StockBalanceStruct

เอกสารยอดยกมา (stock balance document, TransFlag 54 — ยอดยกมา ตาม comment ในซอร์ส) เก็บ header พร้อมรายการรายละเอียดยอดยกมา

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | holdingcode | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| BranchId | string | branchid | branchid | รหัสสาขา |
| GuidFixed | string | guidfixed | guidfixed | GUID ถาวรของเอกสาร |
| DocNo | string | docno | docno | เลขที่เอกสาร |
| Description | string | description | description | คำอธิบายเอกสาร |
| DocDateTime | time.Time | docdatetime | docdatetime | วันเวลาเอกสาร |
| TotalAmount | float64 | totalamount | totalamount | ยอดรวมเอกสาร |
| RoundAmount | float64 | roundamount | roundamount | ยอดปัดเศษ |
| PayCashAmount | float64 | paycashamount | paycashamount | ยอดเงินสดที่รับ |
| PayCashChange | float64 | paycashchange | paycashchange | เงินทอน |
| PaymentDetailRaw | string | paymentdetailraw | paymentdetailraw | รายละเอียดการชำระเงิน (raw) |
| SlipUrl | string | slipurl | slipurl | URL สลิปการชำระเงิน |
| SaleChannelCode | string | salechannelcode | salechannelcode | รหัสช่องทางการขาย |
| DeliveryAmount | float64 | deliveryamount | deliveryamount | ค่าจัดส่ง |
| IsCancel | bool | iscancel | iscancel | เอกสารถูกยกเลิกหรือไม่ |
| CancelReason | string | cancelreason | cancelreason | เหตุผลการยกเลิก |
| GuidPos | string | guidpos | guidpos | GUID จากระบบ POS |
| Branch | [[MongoBranchModel]] | branch | branch | ข้อมูลสาขา |
| Details | [][[StockBalanceDetailStruct]] | details | details | รายการรายละเอียดยอดยกมา |
| DocCreatedDateTime | time.Time | doccreateddatetime | doccreateddatetime | วันเวลาที่สร้างเอกสาร |
| DocUpdatedDateTime | time.Time | docupdateddatetime | docupdateddatetime | วันเวลาที่แก้ไขเอกสาร |
| IsDelete | bool | isdelete | isdelete | ถูกลบ (soft delete) หรือไม่ |
| TransFlag | int | transflag | transflag | ประเภทธุรกรรม: 54 (ยอดยกมา) |

## ความสัมพันธ์
- ฝัง [[MongoBranchModel]] (Branch), [[StockBalanceDetailStruct]] (Details)
- เกี่ยวข้องกับ [[ProcessMongoTransDetailTransFlag54Model]] ซึ่งใช้ประมวลผล detail ของ TransFlag 54 เช่นกัน
- HoldingCode — อ้างอิงกลุ่มกิจการ (tenant boundary)
- BranchId — อ้างอิงสาขา
- SaleChannelCode — อ้างอิงช่องทางการขาย
