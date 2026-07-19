---
source: mongo-trans-model.go
tags: [datamodel, general-type]
---

# MongoDocModel

เอกสารธุรกรรมหลัก (header) เก็บข้อมูลเลขที่เอกสาร วันที่ ภาษี ยอดรวม การชำระเงิน สาขา และรายการรายละเอียด รองรับหลายสกุลเงิน (Base Currency สำหรับลงบัญชี + Document Currency) มี type alias `ProcessMongoTransModel` ชี้มาที่ struct นี้

| Field | Type | bson | json | คำอธิบาย |
|---|---|---|---|---|
| HoldingCode | string | holdingcode | holdingcode | รหัสกลุ่มกิจการ (tenant) |
| BranchId | string | branchid | branchid | รหัสสาขา |
| GuidFixed | string | guidfixed | guidfixed | GUID ถาวรของเอกสาร |
| DocNo | string | docno | docno | เลขที่เอกสาร |
| Description | string | description | description | คำอธิบายเอกสาร |
| CreatorCode | string | creatorcode | creatorcode | รหัสผู้สร้าง (mapstructure: creator_code) |
| CreatorName | string | creatorname | creatorname | ชื่อผู้สร้าง (mapstructure: creator_name) |
| CreatedAt | time.Time | createdat | createdat | วันเวลาที่สร้างเอกสาร |
| ModifierCode | string | updatercode | modifiercode | รหัสผู้แก้ไข (mapstructure: modifier_code) |
| ModifierName | string | updatername | modifiername | ชื่อผู้แก้ไข (mapstructure: modifier_name) |
| ModifiedAt | time.Time | updatedat | modifiedat | วันเวลาที่แก้ไข (mapstructure: modified_at) |
| DocDateTime | time.Time | docdatetime | docdatetime | วันเวลาเอกสาร |
| DocRefDate | time.Time | docrefdate | docrefdate | วันที่เอกสารอ้างอิง |
| DocReferences | [][[MongoDocReferenceModel]] | docreferences | docreferences | รายการเอกสารอ้างอิง |
| TaxDocDate | time.Time | taxdocdate | taxdocdate | วันที่เอกสารภาษี |
| TaxDocNo | string | taxdocno | taxdocno | เลขที่เอกสารภาษี |
| TransFlag | int | transflag | transflag | ประเภทธุรกรรม (trans flag) |
| VatType | int | vattype | vattype | ประเภท VAT |
| VatRate | float64 | vatrate | vatrate | อัตรา VAT |
| CustCode | string | custcode | custcode | รหัสลูกค้า/คู่ค้า |
| ManCount | int | mancount | mancount | จำนวนลูกค้าชาย |
| WomanCount | int | womancount | womancount | จำนวนลูกค้าหญิง |
| ChildCount | int | childcount | childcount | จำนวนลูกค้าเด็ก |
| TotalAmount | float64 | totalamount | totalamount | ยอดรวมสุทธิ |
| TotalValue | float64 | totalvalue | totalvalue | มูลค่ารวม |
| TotalBeforeVat | float64 | totalbeforevat | totalbeforevat | ยอดรวมก่อน VAT |
| TotalVatValue | float64 | totalvatvalue | totalvatvalue | มูลค่า VAT รวม |
| Currency | string | currency,omitempty | currency | สกุลเงินหลัก (Base Currency) สำหรับลงบัญชี |
| CurrencySymbol | string | currencysymbol,omitempty | currencysymbol | สัญลักษณ์สกุลเงินหลัก |
| DocCurrency | string | doccurrency,omitempty | doccurrency | สกุลเงินเอกสาร (mapstructure: doc_currency) |
| DocCurrencySymbol | string | doccurrencysymbol,omitempty | doccurrencysymbol | สัญลักษณ์สกุลเงินเอกสาร (mapstructure: doc_currencysymbol) |
| ExchangeRate | float64 | exchangerate,omitempty | exchangerate | อัตราแลกเปลี่ยน |
| TotalAmountDoc | float64 | totalamountdoc,omitempty | totalamountdoc | ยอดรวมในสกุลเงินเอกสาร (mapstructure: totalamount_doc) |
| PayCashAmount | float64 | paycashamount | paycashamount | ยอดเงินสดที่รับ |
| PayCashChange | float64 | paycashchange | paycashchange | เงินทอน |
| RoundAmount | float64 | roundamount | roundamount | ยอดปัดเศษ |
| PaymentDetailRaw | string | paymentdetailraw | paymentdetailraw | รายละเอียดการชำระเงิน (raw) |
| SlipUrl | string | slipurl | slipurl | URL สลิปการชำระเงิน |
| SaleChannelCode | string | salechannelcode | salechannelcode | รหัสช่องทางการขาย |
| DeliveryAmount | float64 | deliveryamount | deliveryamount | ค่าจัดส่ง |
| Details | [][[MongoDocDetailModel]] | details | details | รายการรายละเอียดเอกสาร |
| IsCancel | bool | iscancel | iscancel | เอกสารถูกยกเลิกหรือไม่ |
| CancelReason | string | cancelreason | cancelreason | เหตุผลการยกเลิก |
| GuidPos | string | guidpos | guidpos | GUID จากระบบ POS |
| Branch | [[MongoBranchModel]] | branch | branch | ข้อมูลสาขา |
| DeletedAt | *time.Time | deletedat | deletedat | วันเวลาที่ลบ (soft delete) |
| IsDelete | bool | isdelete | isdelete | ถูกลบ (soft delete) หรือไม่ |

## ความสัมพันธ์
- ฝัง [[MongoDocReferenceModel]] (DocReferences), [[MongoDocDetailModel]] (Details), [[MongoBranchModel]] (Branch)
- HoldingCode — อ้างอิงกลุ่มกิจการ (tenant boundary)
- BranchId — อ้างอิงสาขา
- CustCode — อ้างอิงลูกค้า/คู่ค้า
- SaleChannelCode — อ้างอิงช่องทางการขาย
- CreatorCode / ModifierCode — อ้างอิงผู้ใช้ที่สร้าง/แก้ไข
