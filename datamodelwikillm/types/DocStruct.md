---
source: process-doc-model.go
tags: [datamodel, general-type]
---

# DocStruct

โครงสร้างหัวเอกสาร (document header) ใช้ map ข้อมูลจาก MongoDB ไปยัง struct เพื่อประมวลผลต่อ ครอบคลุมข้อมูลเอกสารขาย/ธุรกรรม เช่น เลขที่เอกสาร ยอดเงิน การชำระเงิน สกุลเงิน และสถานะยกเลิก/อนุมัติ

หมายเหตุ: struct นี้ใช้ tag `json` + `db` (ไม่มี tag `bson`)

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | - | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| CustCode | string | - | custcode | รหัสลูกค้า |
| TransFlag | int | - | transflag | ประเภทธุรกรรม (trans flag) |
| DocNo | string | - | docno | เลขที่เอกสาร |
| DocDateTime | time.Time | - | docdatetime | วันเวลาเอกสาร |
| PeriodDateTime | time.Time | - | perioddatetime | วันเวลางวด |
| TaxDocNo | string | - | taxdocno | เลขที่เอกสารภาษี |
| TotalAmount | float64 | - | totalamount | ยอดรวมเอกสาร |
| RoundAmount | float64 | - | roundamount | จำนวนเงินปัดเศษ |
| PayType | int | - | paytype | ประเภทการชำระเงิน |
| PayCashAmount | float64 | - | paycashamount | จำนวนเงินสดที่รับ |
| PayCashChange | float64 | - | paycashchange | เงินทอน |
| PayCashBalance | float64 | - | paycashbalance | ยอดเงินสดคงเหลือ |
| DeliveryCode | string | - | deliverycode | รหัสการจัดส่ง |
| Checksum | string | - | checksum | ค่า checksum ของเอกสาร |
| BranchID | string | - | branchid | รหัสสาขา |
| SlipURL | string | - | slipurl | URL สลิป |
| SaleChannelCode | string | - | salechannelcode | รหัสช่องทางการขาย |
| DeliveryAmount | float64 | - | deliveryamount | ค่าจัดส่ง |
| IsCancel | bool | - | iscancel | สถานะยกเลิกเอกสาร |
| CancelReason | string | - | cancelreason | เหตุผลการยกเลิก |
| GuidPOS | string | - | guidpos | GUID เครื่อง POS |
| GuidBranch | string | - | guidbranch | GUID สาขา |
| GuidFixed | string | - | guidfixed | GUID ประจำเอกสาร |
| CreatorCode | string | - | creatorcode | รหัสผู้สร้างเอกสาร |
| CreatorName | string | - | creatorname | ชื่อผู้สร้างเอกสาร |
| CreatedAt | time.Time | - | createdat | วันเวลาที่สร้างเอกสาร |
| Currency | string | - | currency | สกุลเงินหลัก (base currency) สำหรับลงบัญชี |
| CurrencySymbol | string | - | currencysymbol | สัญลักษณ์สกุลเงินหลัก |
| DocCurrency | string | - | doccurrency | สกุลเงินเอกสาร (document currency) |
| DocCurrencySymbol | string | - | doccurrencysymbol | สัญลักษณ์สกุลเงินเอกสาร |
| ExchangeRate | float64 | - | exchangerate | อัตราแลกเปลี่ยน |
| TotalAmountDoc | float64 | - | totalamountdoc | ยอดรวมในสกุลเงินเอกสาร |
| IsDelete | bool | - | isdelete | สถานะลบแบบ soft delete |
| ApprovalStatus | string | - | approvalstatus | สถานะอนุมัติ (จาก collection poapprovalstatus) |

## ความสัมพันธ์

- `HoldingCode` — อ้างถึงกลุ่มกิจการ (tenant boundary)
- `CustCode` — อ้างถึงรหัสลูกค้า
- `BranchID` / `GuidBranch` — อ้างถึงสาขา
- `DeliveryCode` — อ้างถึงรหัสวิธี/รายการจัดส่ง
- `SaleChannelCode` — อ้างถึงช่องทางการขาย
- `CreatorCode` — อ้างถึงรหัสผู้ใช้ที่สร้างเอกสาร
- `ApprovalStatus` — มาจาก collection `poapprovalstatus` (ตาม comment ในโค้ด)
- รายละเอียดรายการสินค้าของเอกสารดูที่ [[DocDetailStruct]] และการอ้างอิงเอกสารดูที่ [[DocRefStruct]] (แยก struct ในไฟล์เดียวกัน เชื่อมกันด้วย `DocNo`)
