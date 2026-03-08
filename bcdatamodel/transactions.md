# Transaction Models

เอกสารธุรกรรมทั้งหมด — ซื้อ/ขาย/โอน/ปรับ/รับ/เบิก

## โครงสร้างหลัก

ทุก transaction ใช้โครงสร้างเดียวกัน:
```
TransactionMessageQueue
├── ShopIdentity        (shopid)
├── DocIdentity         (guidfixed)
└── Transaction
    ├── TransactionHeader   (header ทุก field)
    └── Details []Detail    (รายการสินค้า)
```

## TransactionHeader

เอกสาร header — ใช้ร่วมกันทุกประเภท (transflag ระบุว่าเป็นประเภทไหน)

**Backend:** `internal/transaction/models/transaction.go`
**Frontend:** `lib/model/transaction_model.dart` → `TransactionModel`

### ข้อมูลพื้นฐาน

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| DocNo | string | `docno` | `docno` | เลขที่เอกสาร |
| DocDatetime | time.Time | `docdatetime` | `docdatetime` | วันเวลาเอกสาร (UTC) |
| DocDateLocal | string | `docdatelocal` | `docdatelocal` | วันที่ local (YYYYMMDD) |
| DocTimeLocal | string | `doctimelocal` | `doctimelocal` | เวลา local (HH:MM:SS) |
| Timezone | string | `timezone` | `timezone` | Timezone (Asia/Bangkok) |
| GuidRef | string | `guidref` | `guidref` | GUID อ้างอิง |
| ShiftDocNo | string | `shiftdocno` | `shiftdocno` | เลขที่กะ |
| DeviceName | string | `devicename` | `devicename` | ชื่ออุปกรณ์ |
| GuidPos | string | `guidpos` | `guidpos` | GUID เครื่อง POS |
| TransFlag | int | `transflag` | `transflag` | **ประเภทเอกสาร** (ดู [enums.md](enums.md)) |
| DocType | int8 | `doctype` | `doctype` | ประเภทย่อย (0=sale, 1=purchase) |
| InquiryType | int | `inquirytype` | `inquirytype` | ประเภทการขาย (0=credit, 1=cash) |
| Status | int8 | `status` | `status` | สถานะเอกสาร |
| IsCancel | bool | `iscancel` | `iscancel` | ยกเลิกแล้วหรือไม่ |
| IsClose | bool | `isclosed` | `isclose` | ปิดเอกสารแล้วหรือไม่ |
| IsManualAmount | bool | `ismanualamount` | `ismanualamount` | กรอกจำนวนเงินเอง |
| Description | string | `description` | `description` | หมายเหตุ |

### ข้อมูลลูกค้า / พนักงาน

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| CustCode | string | `custcode` | `custcode` | รหัสลูกค้า/คู่ค้า |
| CustNames | *[]NameX | `custnames` | `custnames` | ชื่อลูกค้า (หลายภาษา) |
| SaleCode | string | `salecode` | `salecode` | รหัสพนักงานขาย |
| SaleName | string | `salename` | `salename` | ชื่อพนักงานขาย |
| MemberCode | string | `membercode` | `membercode` | รหัสสมาชิก |
| CashierCode | string | `cashiercode` | `cashiercode` | รหัสแคชเชียร์ |
| CashierName | string | `cashiername` | `cashiername` | ชื่อแคชเชียร์ |
| PosID | string | `posid` | `posid` | รหัส POS |
| CustomerTelephone | string | `customertelephone` | `customertelephone` | เบอร์โทรลูกค้า |

### ยอดเงิน (Base Currency)

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| TotalValue | float64 | `totalvalue` | `totalvalue` | มูลค่ารวมสินค้า |
| TotalDiscount | float64 | `totaldiscount` | `totaldiscount` | ส่วนลดท้ายบิล |
| TotalAmountAfterDiscount | float64 | `totalamountafterdiscount` | `totalamountafterdiscount` | ยอดหลังส่วนลด |
| TotalBeforeVat | float64 | `totalbeforevat` | `totalbeforevat` | ยอดก่อน VAT |
| TotalVatValue | float64 | `totalvatvalue` | `totalvatvalue` | มูลค่า VAT |
| TotalAfterVat | float64 | `totalaftervat` | `totalaftervat` | ยอดหลัง VAT |
| TotalExceptVat | float64 | `totalexceptvat` | `totalexceptvat` | ยอดยกเว้น VAT |
| TotalAmount | float64 | `totalamount` | `totalamount` | **ยอดรวมทั้งหมด** |
| TotalCost | float64 | `totalcost` | `totalcost` | ต้นทุนรวม |
| RoundAmount | float64 | `roundamount` | `roundamount` | ปัดเศษ |
| TotalQty | float64 | `totalqty` | `totalqty` | จำนวนรวม |

### ยอดเงิน (Document Currency)

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| Currency | string | `currency` | `currency` | สกุลเงินหลัก (THB) |
| CurrencySymbol | string | `currencysymbol` | `currencysymbol` | สัญลักษณ์ (฿) |
| DocCurrency | string | `doc_currency` | `doc_currency` | สกุลเงินเอกสาร (USD, JPY, ...) |
| DocCurrencySymbol | string | `doc_currencysymbol` | `doc_currencysymbol` | สัญลักษณ์ ($, ¥, ...) |
| ExchangeRate | float64 | `exchangerate` | `exchangerate` | อัตราแลกเปลี่ยน |
| TotalValueDoc | float64 | `totalvalue_doc` | `totalvalue_doc` | มูลค่ารวม (Doc Currency) |
| TotalDiscountDoc | float64 | `totaldiscount_doc` | `totaldiscount_doc` | ส่วนลด (Doc Currency) |
| TotalVatValueDoc | float64 | `totalvatvalue_doc` | `totalvatvalue_doc` | VAT (Doc Currency) |
| TotalBeforeVatDoc | float64 | `totalbeforevat_doc` | `totalbeforevat_doc` | ก่อน VAT (Doc Currency) |
| TotalAfterVatDoc | float64 | `totalaftervat_doc` | `totalaftervat_doc` | หลัง VAT (Doc Currency) |
| TotalAmountDoc | float64 | `totalamount_doc` | `totalamount_doc` | ยอดรวม (Doc Currency) |

### VAT / ภาษี

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| VatType | int8 | `vattype` | `vattype` | 0=ไม่มี, 1=รวมVAT, 2=อัตรา0, 3=ยกเว้น |
| VatRate | float64 | `vatrate` | `vatrate` | อัตรา VAT (เช่น 7) |
| BillTaxType | int8 | `billtaxtype` | `billtaxtype` | ประเภทภาษีของบิล |
| IsVatRegister | bool | `isvatregister` | `isvatregister` | จดทะเบียน VAT หรือไม่ |
| TotalDiscountVatAmount | float64 | `totaldiscountvatamount` | `totaldiscountvatamount` | VAT ของส่วนลด |
| TotalDiscountExceptVatAmount | float64 | `totaldiscountexceptvatamount` | `totaldiscountexceptvatamount` | ส่วนลดยกเว้น VAT |

### WHT (ภาษีหัก ณ ที่จ่าย)

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| WHTEntries | []WHTEntry | `wht_entries` | `wht_entries` | รายการหัก ณ ที่จ่าย (อาจมีหลายอัตรา) |

**WHTEntry struct:**

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| Description | string | `description` | ประเภทเงินได้ (ค่าขนส่ง, ค่าบริการ, etc.) |
| Rate | float64 | `rate` | อัตราภาษี % (0.5, 1, 2, 3, 5, 10, 15) |
| TaxBase | float64 | `taxbase` | ฐานภาษี |
| Amount | float64 | `amount` | จำนวนเงินหัก (TaxBase × Rate%) |
| Note | string | `note` | หมายเหตุ |

### เครดิตเทอม

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| CreditDays | int | `creditdays` | `creditdays` | จำนวนวันเครดิต (0=เงินสด) |
| DueDate | string | `duedate` | `duedate` | วันครบกำหนด (YYYY-MM-DD) |

### การชำระเงิน

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| PaymentDetail | PaymentDetail | `paymentdetail` | `paymentdetail` | สรุปการชำระ |
| PayCashAmount | float64 | `paycashamount` | `paycashamount` | จ่ายเงินสด |
| PayCashChange | float64 | `paycashchange` | `paycashchange` | เงินทอน |
| PayPointAmount | float64 | `paypointamount` | `paypointamount` | ใช้แต้มจ่าย |
| SumCreditCard | float64 | `sumcreditcard` | `sumcreditcard` | ยอดบัตรเครดิต |
| SumMoneyTransfer | float64 | `summoneytransfer` | `summoneytransfer` | ยอดโอน |
| SumQRCode | float64 | `sumqrcode` | `sumqrcode` | ยอด QR Code |
| SumCheque | float64 | `sumcheque` | `sumcheque` | ยอดเช็ค |
| SumCoupon | float64 | `sumcoupon` | `sumcoupon` | ยอดคูปอง |
| SumDeposit | float64 | `sumdeposit` | `sumdeposit` | ยอดมัดจำ |
| SumAdvancePayment | float64 | `sumadvancepayment` | `sumadvancepayment` | ยอดจ่ายล่วงหน้า |
| SumCredit | float64 | `sumcredit` | `sumcredit` | ยอดเครดิต |

### แต้ม / คูปอง

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| GetPoint | float64 | `getpoint` | `getpoint` | แต้มที่ได้รับ |
| UsePoint | float64 | `usepoint` | `usepoint` | แต้มที่ใช้ |
| PointDiscountAmount | float64 | `pointdiscountamount` | `pointdiscountamount` | ส่วนลดจากแต้ม |
| Coupons | []SaleInvoiceCoupon | `coupons` | `coupons` | คูปองที่ใช้ |
| TotalCouponAmount | float64 | `totalcouponamount` | `totalcouponamount` | ยอดคูปองรวม |
| CouponDiscountAmount | float64 | `coupondiscountamount` | `coupondiscountamount` | ส่วนลดคูปอง |
| CouponCashAmount | float64 | `couponcashamount` | `couponcashamount` | คูปองเงินสด |

### ประเภทการจัดซื้อ (PO Approval)

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| PurchaseTypeCode | string | `purchasetypecode` | `purchasetypecode` | รหัสประเภทจัดซื้อ |
| PurchaseTypeNames | *[]NameX | `purchasetypenames` | `purchasetypenames` | ชื่อประเภท (หลายภาษา) |

### สาขา

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| Branch | TransactionBranch | `branch` | `branch` | ข้อมูลสาขา |

**TransactionBranch:**

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| GuidFixed | string | `guidfixed` | GUID สาขา |
| Code | string | `code` | รหัสสาขา |
| Names | *[]NameX | `names` | ชื่อสาขา (หลายภาษา) |

### ข้อมูลผู้สร้าง/แก้ไข (Audit)

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| CreatorCode | string | `creator_code` | `creatorcode` | รหัสผู้สร้าง |
| CreatorName | string | `creator_name` | `creatorname` | ชื่อผู้สร้าง |
| CreatedAt | time.Time | `created_at` | `createdat` | วันเวลาสร้าง |
| UpdaterCode | string | `modifier_code` | `updatercode` | รหัสผู้แก้ไข |
| UpdaterName | string | `modifier_name` | `updatername` | ชื่อผู้แก้ไข |
| UpdatedAt | time.Time | `modified_at` | `updatedat` | วันเวลาแก้ไข |

> **สำคัญ:** JSON ใช้ snake_case (`creator_code`, `modifier_code`) ตรงกับ Flutter
> BSON ใช้ camelCase (`creatorcode`, `updatercode`) ตรงกับ MongoDB เดิม

### ยกเลิก

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| CancelDateTime | string | `canceldatetime` | วันเวลายกเลิก |
| CancelUserCode | string | `cancelusercode` | รหัสผู้ยกเลิก |
| CancelUserName | string | `cancelusername` | ชื่อผู้ยกเลิก |
| CancelDescription | string | `canceldescription` | คำอธิบาย |
| CancelReason | string | `cancelreason` | เหตุผล |

### ใบกำกับภาษีเต็มรูป

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| FullVatName | string | `fullvatname` | ชื่อ |
| FullVatAddress | string | `fullvataddress` | ที่อยู่ |
| FullVatTaxID | string | `fullvattaxid` | เลขประจำตัวผู้เสียภาษี |
| FullVatBranchNumber | string | `fullvatbranchnumber` | เลขสาขา |
| FullVatDocNumber | string | `fullvatdocnumber` | เลขที่ใบกำกับ |
| FullVatPrint | bool | `fullvatprint` | พิมพ์แล้วหรือไม่ |

### ร้านอาหาร (Restaurant mode)

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| TableNumber | string | `tablenumber` | หมายเลขโต๊ะ |
| TableOpenDateTime | string | `tableopendatetime` | เปิดโต๊ะ |
| TableCloseDateTime | string | `tableclosedatetime` | ปิดโต๊ะ |
| ManCount | int | `mancount` | ผู้ชาย |
| WomanCount | int | `womancount` | ผู้หญิง |
| ChildCount | int | `childcount` | เด็ก |
| IsTableAllacrateMode | bool | `istableallacratemode` | สั่ง a la carte |
| BuffetCode | string | `buffetcode` | รหัสบุฟเฟ่ต์ |

### เอกสารอ้างอิง

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| DocRefType | int8 | `docreftype` | ประเภทอ้างอิง |
| DocRefNo | string | `docrefno` | เลขที่อ้างอิง |
| DocRefDate | time.Time | `docrefdate` | วันที่อ้างอิง |
| DocReferences | []TransactionDocRef | `docreferences` | เอกสารอ้างอิง (หลายใบ) |
| DepositDocs | []TransactionDepositDoc | `depositdocs` | เอกสารมัดจำ |
| AdvancePaymentDocs | []TransactionDepositDoc | `advancepaymentdocs` | เอกสารจ่ายล่วงหน้า |
| TaxDocNo | string | `taxdocno` | เลขที่ใบกำกับภาษี |
| TaxDocDate | time.Time | `taxdocdate` | วันที่ใบกำกับ |

---

## Detail (รายการสินค้า)

**Backend:** `internal/transaction/models/transaction.go` → `Detail`
**Frontend:** `lib/model/transaction_model.dart` → `TransactionDetailModel`

### ข้อมูลสินค้า

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| LineNumber | int | `linenumber` | `linenumber` | ลำดับบรรทัด |
| DocDatetime | time.Time | `docdatetime` | `docdatetime` | วันเวลา |
| Barcode | string | `barcode` | `barcode` | บาร์โค้ด |
| ItemCode | string | `itemcode` | `itemcode` | รหัสสินค้า |
| ItemGuid | string | `itemguid` | `itemguid` | GUID สินค้า |
| ItemNames | *[]NameX | `itemnames` | `itemnames` | ชื่อสินค้า (หลายภาษา) |
| ItemType | int8 | `itemtype` | `itemtype` | 0=ปกติ, 1=ชุด |
| UnitCode | string | `unitcode` | `unitcode` | รหัสหน่วย |
| UnitNames | *[]NameX | `unitnames` | `unitnames` | ชื่อหน่วย (หลายภาษา) |
| ImageUri | string | `imageuri` | `imageurl` | URL รูปสินค้า |
| SKU | string | `sku` | `sku` | SKU |
| GroupCode | string | `groupcode` | `groupcode` | รหัสกลุ่มสินค้า |
| GroupNames | *[]NameX | `groupnames` | `groupnames` | ชื่อกลุ่ม |
| ManufacturerGUID | string | `manufacturerguid` | `manufacturerguid` | GUID ผู้ผลิต |
| ManufacturerCode | string | `manufacturercode` | `manufacturercode` | รหัสผู้ผลิต |
| ManufacturerNames | *[]NameX | `manufacturernames` | `manufacturernames` | ชื่อผู้ผลิต |

### จำนวน / ราคา (Base Currency)

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| Qty | float64 | `qty` | `qty` | จำนวน |
| TotalQty | float64 | `totalqty` | `totalqty` | จำนวนรวม (Qty × Event) |
| EventQty | float64 | `eventqty` | `eventqty` | จำนวนเหตุการณ์ |
| Price | float64 | `price` | `price` | ราคาต่อหน่วย |
| PriceExcludeVat | float64 | `priceexcludevat` | `priceexcludevat` | ราคาไม่รวม VAT |
| Discount | string | `discount` | `discount` | สูตรส่วนลด (เช่น "10+5") |
| DiscountAmount | float64 | `discountamount` | `discountamount` | จำนวนส่วนลด |
| SumAmount | float64 | `sumamount` | `sumamount` | ยอดรวม (หลังส่วนลด) |
| SumAmountExcludeVat | float64 | `sumamountexcludevat` | `sumamountexcludevat` | ยอดไม่รวม VAT |
| TotalValueVat | float64 | `totalvaluevat` | `totalvaluevat` | มูลค่า VAT |
| SumOfCost | float64 | `sumofcost` | `sumofcost` | ต้นทุนรวม |
| AverageCost | float64 | `averagecost` | `averagecost` | ต้นทุนเฉลี่ย |

### จำนวน / ราคา (Document Currency)

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| PriceDoc | float64 | `price_doc` | `price_doc` | ราคา (Doc Currency) |
| SumAmountDoc | float64 | `sumamount_doc` | `sumamount_doc` | ยอดรวม (Doc Currency) |
| DiscountAmountDoc | float64 | `discountamount_doc` | `discountamount_doc` | ส่วนลด (Doc Currency) |
| PriceExcludeVatDoc | float64 | `priceexcludevat_doc` | `priceexcludevat_doc` | ราคาไม่รวม VAT (Doc Currency) |
| SumAmountExcludeVatDoc | float64 | `sumamountexcludevat_doc` | `sumamountexcludevat_doc` | ยอดไม่รวม VAT (Doc Currency) |
| TotalValueVatDoc | float64 | `totalvaluevat_doc` | `totalvaluevat_doc` | VAT (Doc Currency) |

### หน่วยนับ / คำนวณ Stock

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| CalcFlag | int8 | `calcflag` | `calcflag` | ทิศทาง stock: 1=เข้า, -1=ออก |
| DivideValue | float64 | `dividevalue` | `dividevalue` | ตัวหาร (unit conversion) |
| StandValue | float64 | `standvalue` | `standvalue` | ตัวตั้ง (unit conversion) |
| MultiUnit | bool | `multiunit` | `multiunit` | ใช้หลายหน่วยนับ |
| IssumPoint | bool | `issumpoint` | `issumpoint` | สะสมแต้ม |
| IsChoice | int8 | `ischoice` | `ischoice` | เป็นตัวเลือก |
| SumAmountChoice | float64 | `sumamountchoice` | `sumamountchoice` | ยอดตัวเลือก |

### คลังสินค้า

| Field | Go Type | JSON | BSON | หมายเหตุ |
|-------|---------|------|------|----------|
| WhCode | string | `whcode` | `whcode` | รหัสคลังต้นทาง |
| WhNames | *[]NameX | `whnames` | `whnames` | ชื่อคลัง |
| LocationCode | string | `locationcode` | `locationcode` | รหัสตำแหน่ง |
| LocationNames | *[]NameX | `locationnames` | `locationnames` | ชื่อตำแหน่ง |
| ShelfCode | string | `shelfcode` | `shelfcode` | รหัสชั้นวาง |
| ToWhCode | string | `towhcode` | `towhcode` | รหัสคลังปลายทาง (โอน) |
| ToWhNames | *[]NameX | `towhnames` | `towhnames` | ชื่อคลังปลายทาง |
| ToLocationCode | string | `tolocationcode` | `tolocationcode` | รหัสตำแหน่งปลายทาง |
| ToLocationNames | *[]NameX | `tolocationnames` | `tolocationnames` | ชื่อตำแหน่งปลายทาง |

---

## PaymentDetail (สรุปการชำระ)

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| CashAmountText | string | `cashamounttext` | จำนวนเงินเป็นตัวหนังสือ |
| CashAmount | float64 | `cashamount` | จำนวนเงินสด |
| PaymentCreditCards | *[]PaymentCreditCard | `paymentcreditcards` | บัตรเครดิต |
| PaymentTransfers | *[]PaymentTransfer | `paymenttransfers` | โอนเงิน |

### PaymentCreditCard

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| DocDatetime | time.Time | `docdatetime` | วันเวลารายการ |
| CardNumber | string | `cardnumber` | เลขบัตร |
| Amount | float64 | `amount` | จำนวนเงิน |
| ChargeWord | string | `chargeword` | สูตรค่าธรรมเนียม |
| ChargeValue | float64 | `chargevalue` | ค่าธรรมเนียม |
| TotalNetWorth | float64 | `totalnetworth` | ยอดสุทธิ |

### PaymentTransfer

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| DocDatetime | time.Time | `docdatetime` | วันเวลารายการ |
| BankCode | string | `bankcode` | รหัสธนาคาร |
| BankNames | *[]NameX | `banknames` | ชื่อธนาคาร (หลายภาษา) |
| AccountNumber | string | `accountnumber` | เลขบัญชี |
| Amount | float64 | `amount` | จำนวนเงิน |

---

## SaleInvoiceCoupon (คูปอง)

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| CouponNo | string | `couponno` | เลขที่คูปอง |
| CouponAmount | float64 | `couponamount` | จำนวนเงิน |
| CouponDescription | string | `coupondescription` | คำอธิบาย |
| CouponType | string | `coupontype` | ประเภทคูปอง |

---

## TransactionDocRef (เอกสารอ้างอิง)

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| GuidFixed | string | `guidfixed` | GUID เอกสารอ้างอิง |
| DocNo | string | `docno` | เลขที่เอกสารอ้างอิง |
| DocDatetime | time.Time | `docdatetime` | วันเวลาเอกสาร |

## TransactionDepositDoc (เอกสารมัดจำ/จ่ายล่วงหน้า)

| Field | Go Type | JSON | หมายเหตุ |
|-------|---------|------|----------|
| GuidFixed | string | `guidfixed` | GUID |
| DocNo | string | `docno` | เลขที่เอกสาร |
| DocDatetime | time.Time | `docdatetime` | วันเวลา |
| TotalAmount | float64 | `totalamount` | ยอดรวม |
| Balance | float64 | `balance` | ยอดคงเหลือ |
| UseAmount | float64 | `useamount` | ยอดที่ใช้ |
| Remark | string | `remark` | หมายเหตุ |
