---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# StockAdjustmentStruct

เอกสารปรับปรุงสต๊อก (stock adjustment document, TransFlag 66 = เพิ่ม หรือ 68 = ลด ตาม comment ในซอร์ส) เก็บ header พร้อมรายการรายละเอียดการปรับ

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | holdingcode | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| BranchId | string | branchid | branchid | รหัสสาขา |
| GuidFixed | string | guidfixed | guidfixed | GUID ถาวรของเอกสาร |
| DocNo | string | docno | docno | เลขที่เอกสาร |
| Description | string | description | description | คำอธิบายเอกสาร |
| DocDateTime | time.Time | docdatetime | docdatetime | วันเวลาเอกสาร |
| TotalAmount | float64 | totalamount | totalamount | ยอดรวมเอกสาร |
| RoundAmount | float64 | - | roundamount | ยอดปัดเศษ (ไม่มี bson tag) |
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
| Details | [][[StockAdjustmentDetailStruct]] | details | details | รายการรายละเอียดการปรับสต๊อก |
| DocCreatedDateTime | time.Time | doccreateddatetime | doccreateddatetime | วันเวลาที่สร้างเอกสาร |
| DocUpdatedDateTime | time.Time | docupdateddatetime | docupdateddatetime | วันเวลาที่แก้ไขเอกสาร |
| IsDelete | bool | isdelete | isdelete | ถูกลบ (soft delete) หรือไม่ |
| TransFlag | int | transflag | transflag | ประเภทธุรกรรม: 66 (เพิ่ม) หรือ 68 (ลด) |

## ความสัมพันธ์
- ฝัง [[MongoBranchModel]] (Branch), [[StockAdjustmentDetailStruct]] (Details)
- HoldingCode — อ้างอิงกลุ่มกิจการ (tenant boundary)
- BranchId — อ้างอิงสาขา
- SaleChannelCode — อ้างอิงช่องทางการขาย
