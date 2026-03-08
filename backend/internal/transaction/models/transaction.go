package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"smlcloudplatform/internal/models"
	"time"
)

type TransactionHeader struct {
	DocNo       string    `json:"docno" bson:"docno"`
	DocDatetime time.Time `json:"docdatetime" bson:"docdatetime"`
	// Timezone fields - ใช้สำหรับสร้าง docno ตาม local timezone ของ frontend
	DocDateLocal                    string                  `json:"docdatelocal,omitempty" bson:"docdatelocal,omitempty"` // วันที่ local เช่น "20260107"
	DocTimeLocal                    string                  `json:"doctimelocal,omitempty" bson:"doctimelocal,omitempty"` // เวลา local เช่น "08:30:00"
	Timezone                        string                  `json:"timezone,omitempty" bson:"timezone,omitempty"`         // timezone name เช่น "Asia/Bangkok"
	GuidRef                         string                  `json:"guidref" bson:"guidref"`
	ShiftDocNo                      string                  `json:"shiftdocno" bson:"shiftdocno"`
	DeviceName                      string                  `json:"devicename" bson:"devicename"`
	GuidPos                         string                  `json:"guidpos" bson:"guidpos"`
	TransFlag                       int                     `json:"transflag" bson:"transflag"`
	DocRefType                      int8                    `json:"docreftype" bson:"docreftype"`
	DocReferences                   []TransactionDocRef     `json:"docreferences,omitempty" bson:"docreferences,omitempty"`
	DocRefNo                        string                  `json:"docrefno" bson:"docrefno"`
	DocRefDate                      time.Time               `json:"docrefdate" bson:"docrefdate"`
	DepositDocs                     []TransactionDepositDoc `json:"depositdocs,omitempty" bson:"depositdocs,omitempty"`
	AdvancePaymentDocs              []TransactionDepositDoc `json:"advancepaymentdocs,omitempty" bson:"advancepaymentdocs,omitempty"`
	TaxDocDate                      time.Time               `json:"taxdocdate" bson:"taxdocdate"`
	TaxDocNo                        string                  `json:"taxdocno" bson:"taxdocno"`
	DocType                         int8                    `json:"doctype" bson:"doctype"`
	ImageUri                        string                  `json:"imageuri" bson:"imageurl"`
	InquiryType                     int                     `json:"inquirytype" bson:"inquirytype"`
	VatType                         int8                    `json:"vattype" bson:"vattype"`
	VatRate                         float64                 `json:"vatrate" bson:"vatrate"`
	CustCode                        string                  `json:"custcode" bson:"custcode"`
	CustNames                       *[]models.NameX         `json:"custnames" bson:"custnames"`
	GetPoint                        float64                 `json:"getpoint" bson:"getpoint"`
	UsePoint                        float64                 `json:"usepoint" bson:"usepoint"`
	PointDiscountAmount             float64                 `json:"pointdiscountamount" bson:"pointdiscountamount"`
	Description                     string                  `json:"description" bson:"description"`
	DiscountWord                    string                  `json:"discountword" bson:"discountword"`
	TotalDiscount                   float64                 `json:"totaldiscount" bson:"totaldiscount"`
	TotalValue                      float64                 `json:"totalvalue" bson:"totalvalue"`
	TotalExceptVat                  float64                 `json:"totalexceptvat" bson:"totalexceptvat"`
	TotalAfterVat                   float64                 `json:"totalaftervat" bson:"totalaftervat"`
	TotalBeforeVat                  float64                 `json:"totalbeforevat" bson:"totalbeforevat"`
	TotalVatValue                   float64                 `json:"totalvatvalue" bson:"totalvatvalue"`
	TotalAmount                     float64                 `json:"totalamount" bson:"totalamount"`
	TotalCost                       float64                 `json:"totalcost" bson:"totalcost"`
	PosID                           string                  `json:"posid" bson:"posid"`
	CashierCode                     string                  `json:"cashiercode" bson:"cashiercode"`
	SaleCode                        string                  `json:"salecode" bson:"salecode"`
	SaleName                        string                  `json:"salename" bson:"salename"`
	MemberCode                      string                  `json:"membercode" bson:"membercode"`
	IsCancel                        bool                    `json:"iscancel" bson:"iscancel"`
	IsManualAmount                  bool                    `json:"ismanualamount" bson:"ismanualamount"`
	Status                          int8                    `json:"status" bson:"status"`
	PaymentDetail                   PaymentDetail           `json:"paymentdetail" bson:"paymentdetail"`
	PaymentDetailRaw                string                  `json:"paymentdetailraw" bson:"paymentdetailraw"`
	PayCashAmount                   float64                 `json:"paycashamount" bson:"paycashamount"`
	PayPointAmount                  float64                 `json:"paypointamount" bson:"paypointamount"`
	Branch                          TransactionBranch       `json:"branch" bson:"branch"`
	BillTaxType                     int8                    `json:"billtaxtype" bson:"billtaxtype"`
	CancelDateTime                  string                  `json:"canceldatetime" bson:"canceldatetime"`
	CancelUserCode                  string                  `json:"cancelusercode" bson:"cancelusercode"`
	CancelUserName                  string                  `json:"cancelusername" bson:"cancelusername"`
	CancelDescription               string                  `json:"canceldescription" bson:"canceldescription"`
	CancelReason                    string                  `json:"cancelreason" bson:"cancelreason"`
	FullVatAddress                  string                  `json:"fullvataddress" bson:"fullvataddress"`
	FullVatBranchNumber             string                  `json:"fullvatbranchnumber" bson:"fullvatbranchnumber"`
	FullVatName                     string                  `json:"fullvatname" bson:"fullvatname"`
	FullVatDocNumber                string                  `json:"fullvatdocnumber" bson:"fullvatdocnumber"`
	FullVatTaxID                    string                  `json:"fullvattaxid" bson:"fullvattaxid"`
	FullVatPrint                    bool                    `json:"fullvatprint" bson:"fullvatprint"`
	IsVatRegister                   bool                    `json:"isvatregister" bson:"isvatregister"`
	IsClose                         bool                    `json:"isclosed" bson:"isclose"`
	PrintCopyBillDateTime           []string                `json:"printcopybilldatetime" bson:"printcopybilldatetime"`
	TableNumber                     string                  `json:"tablenumber" bson:"tablenumber"`
	TableOpenDateTime               string                  `json:"tableopendatetime" bson:"tableopendatetime"`
	TableCloseDateTime              string                  `json:"tableclosedatetime" bson:"tableclosedatetime"`
	ManCount                        int                     `json:"mancount" bson:"mancount"`
	WomanCount                      int                     `json:"womancount" bson:"womancount"`
	ChildCount                      int                     `json:"childcount" bson:"childcount"`
	IsTableAllacrateMode            bool                    `json:"istableallacratemode" bson:"istableallacratemode"`
	BuffetCode                      string                  `json:"buffetcode" bson:"buffetcode"`
	CustomerTelephone               string                  `json:"customertelephone" bson:"customertelephone"`
	TotalQty                        float64                 `json:"totalqty" bson:"totalqty"`
	TotalDiscountVatAmount          float64                 `json:"totaldiscountvatamount" bson:"totaldiscountvatamount"`
	TotalDiscountExceptVatAmount    float64                 `json:"totaldiscountexceptvatamount" bson:"totaldiscountexceptvatamount"`
	CashierName                     string                  `json:"cashiername" bson:"cashiername"`
	PayCashChange                   float64                 `json:"paycashchange" bson:"paycashchange"`
	SumQRCode                       float64                 `json:"sumqrcode" bson:"sumqrcode"`
	SumCreditCard                   float64                 `json:"sumcreditcard" bson:"sumcreditcard"`
	SumMoneyTransfer                float64                 `json:"summoneytransfer" bson:"summoneytransfer"`
	SumCheque                       float64                 `json:"sumcheque" bson:"sumcheque"`
	SumCoupon                       float64                 `json:"sumcoupon" bson:"sumcoupon"`
	SumDeposit                      float64                 `json:"sumdeposit" bson:"sumdeposit"`
	SumAdvancePayment               float64                 `json:"sumadvancepayment" bson:"sumadvancepayment"`
	Coupons                         []SaleInvoiceCoupon     `json:"coupons" bson:"coupons"`
	TotalCouponAmount               float64                 `json:"totalcouponamount" bson:"totalcouponamount"`
	CouponDiscountAmount            float64                 `json:"coupondiscountamount" bson:"coupondiscountamount"`
	CouponCashAmount                float64                 `json:"couponcashamount" bson:"couponcashamount"`
	DetailDiscountFormula           string                  `json:"detaildiscountformula" bson:"detaildiscountformula"`
	DetailTotalAmount               float64                 `json:"detailtotalamount" bson:"detailtotalamount"`
	DetailTotalDiscount             float64                 `json:"detailtotaldiscount" bson:"detailtotaldiscount"`
	RoundAmount                     float64                 `json:"roundamount" bson:"roundamount"`
	TotalAmountAfterDiscount        float64                 `json:"totalamountafterdiscount" bson:"totalamountafterdiscount"`
	DetailTotalAmountBeforeDiscount float64                 `json:"detailtotalamountbeforediscount" bson:"detailtotalamountbeforediscount"`
	SumCredit                       float64                 `json:"sumcredit" bson:"sumcredit"`
	// IsCalcSuccess                   bool              `json:"iscalcsuccess" bson:"iscalcsuccess"`
	// IsCalcBOM                       bool              `json:"iscalcbom" bson:"iscalcbom"`
	// BOMCost                         float64           `json:"bomcost" bson:"bomcost"`
	// BOMGUID                         string            `json:"bomguid" bson:"bomguid"`

	// ประเภทการจัดซื้อ (สำหรับ PO Approval)
	PurchaseTypeCode  string          `json:"purchasetypecode" bson:"purchasetypecode"`
	PurchaseTypeNames *[]models.NameX `json:"purchasetypenames" bson:"purchasetypenames"`

	// ข้อมูลผู้สร้างเอกสาร - ส่งไป backend ผ่าน Kafka
	// NOTE: JSON tags ใช้ snake_case ให้ตรงกับ Flutter (@JsonKey) และ PostgreSQL column names
	// BSON tags ยังคงเดิม (backward compatible กับ MongoDB data)
	CreatorCode string    `json:"creator_code" bson:"creatorcode"` // รหัสผู้สร้าง
	CreatorName string    `json:"creator_name" bson:"creatorname"` // ชื่อผู้สร้าง
	CreatedAt   time.Time `json:"created_at" bson:"createdat"`     // วันเวลาที่สร้างเอกสาร

	// ข้อมูลผู้แก้ไขเอกสารล่าสุด - ส่งไป backend ผ่าน Kafka
	// NOTE: ใช้ "modifier" แทน "updater" ให้ตรงกับ Flutter/PostgreSQL naming
	UpdaterCode string    `json:"modifier_code" bson:"updatercode"` // รหัสผู้แก้ไข
	UpdaterName string    `json:"modifier_name" bson:"updatername"` // ชื่อผู้แก้ไข
	UpdatedAt   time.Time `json:"modified_at" bson:"updatedat"`     // วันเวลาที่แก้ไข

	// ============ Multi-Currency Fields ============
	// ข้อมูลสกุลเงินและอัตราแลกเปลี่ยน (สำหรับ PO)
	// NOTE: ลบ omitempty จาก json tag เพื่อส่งค่า 0/"" ไป Kafka ได้ (แต่เก็บ omitempty ใน bson tag สำหรับ MongoDB)
	
	// Currency (field เดิม) = Base Currency (สกุลเงินหลักสำหรับลงบัญชี) - ค่าเริ่มต้น THB
	// ทุกเอกสารจะต้องมีสกุลเงินหลักสำหรับการลงบัญชี
	Currency       string  `json:"currency" bson:"currency,omitempty"`             // สกุลเงินหลัก (THB, etc.)
	CurrencySymbol string  `json:"currencysymbol" bson:"currencysymbol,omitempty"` // สัญลักษณ์สกุลเงินหลัก
	
	// DocCurrency (field ใหม่) = Document Currency (สกุลเงินของเอกสาร)
	// สกุลเงินที่ใช้บันทึกเอกสาร (อาจต่างจาก Base Currency ในกรณีเอกสารต่างประเทศ)
	DocCurrency       string  `json:"doc_currency" bson:"doc_currency,omitempty"`             // สกุลเงินเอกสาร (THB, USD, JPY, EUR, ...)
	DocCurrencySymbol string  `json:"doc_currencysymbol" bson:"doc_currencysymbol,omitempty"` // สัญลักษณ์สกุลเงินเอกสาร ($, €, ¥, ฿, ...)
	ExchangeRate   float64 `json:"exchangerate" bson:"exchangerate"`               // อัตราแลกเปลี่ยน (1 Document Currency = ? Base Currency)

	// ============ Document Currency Fields (ยอดรวมในสกุลเงินเอกสาร) ============
	// NOTE: Fields เดิม (TotalValue, TotalAmount, etc.) เก็บยอดในสกุลเงินหลัก (Base Currency = THB)
	// Fields ใหม่นี้เก็บยอดในสกุลเงินเอกสาร (Document Currency = USD, JPY, etc.)
	// ลบ omitempty จาก json tag เพื่อส่งค่า 0 ไป Kafka ได้ (แต่เก็บ omitempty ใน bson tag สำหรับ MongoDB)
	TotalValueDoc      float64 `json:"totalvalue_doc" bson:"totalvalue_doc,omitempty"`           // มูลค่ารวมสินค้า (Document Currency)
	TotalDiscountDoc   float64 `json:"totaldiscount_doc" bson:"totaldiscount_doc,omitempty"`     // ส่วนลดท้ายบิล (Document Currency)
	TotalVatValueDoc   float64 `json:"totalvatvalue_doc" bson:"totalvatvalue_doc,omitempty"`     // มูลค่าภาษี (Document Currency)
	TotalBeforeVatDoc  float64 `json:"totalbeforevat_doc" bson:"totalbeforevat_doc,omitempty"`   // มูลค่าก่อนหักภาษี (Document Currency)
	TotalAfterVatDoc   float64 `json:"totalaftervat_doc" bson:"totalaftervat_doc,omitempty"`     // มูลค่าหลังหักภาษี (Document Currency)
	TotalAmountDoc     float64 `json:"totalamount_doc" bson:"totalamount_doc,omitempty"`         // มูลค่ารวมทั้งหมด (Document Currency)

	// ============ WHT (Withholding Tax) Fields ============
	// ภาษีหัก ณ ที่จ่าย (สำหรับ PO) - array เพราะอาจมีหลายอัตราในเอกสารเดียวกัน
	// NOTE: ไม่ใช้ omitempty ใน bson tag เพื่อให้ empty array ถูก update ด้วย (เพื่อเคลียร์ค่าเดิม)
	WHTEntries []WHTEntry `json:"wht_entries,omitempty" bson:"wht_entries"`

	// ============ Credit Terms Fields ============
	// เงื่อนไขการชำระเงิน (สำหรับ PO)
	// NOTE: ลบ omitempty จาก bson tag เพื่อให้ update ค่า 0 หรือ "" ได้ (เพื่อเคลียร์ค่าเดิม)
	CreditDays int    `json:"creditdays,omitempty" bson:"creditdays"` // เครดิตเทอม (จำนวนวัน)
	DueDate    string `json:"duedate,omitempty" bson:"duedate"`       // วันครบกำหนดชำระ (YYYY-MM-DD)
}

type TransactionMoneyHeader struct {
	DocNo       string    `json:"docno" bson:"docno"`
	DocDatetime time.Time `json:"docdatetime" bson:"docdatetime"`
	GuidRef     string    `json:"guidref" bson:"guidref"`
	ShiftDocNo  string    `json:"shiftdocno" bson:"shiftdocno"`

	TransFlag             int                 `json:"transflag" bson:"transflag"`
	DocRefType            int8                `json:"docreftype" bson:"docreftype"`
	DocReferences         []TransactionDocRef `json:"docreferences,omitempty" bson:"docreferences,omitempty"`
	DocRefNo              string              `json:"docrefno" bson:"docrefno"`
	DocRefDate            time.Time           `json:"docrefdate" bson:"docrefdate"`
	TaxDocDate            time.Time           `json:"taxdocdate" bson:"taxdocdate"`
	TaxDocNo              string              `json:"taxdocno" bson:"taxdocno"`
	DocType               int8                `json:"doctype" bson:"doctype"`
	ImageUri              string              `json:"imageuri" bson:"imageurl"`
	InquiryType           int                 `json:"inquirytype" bson:"inquirytype"`
	CustCode              string              `json:"custcode" bson:"custcode"`
	CustNames             *[]models.NameX     `json:"custnames" bson:"custnames"`
	Description           string              `json:"description" bson:"description"`
	TotalAmount           float64             `json:"totalamount" bson:"totalamount"`
	SaleCode              string              `json:"salecode" bson:"salecode"`
	SaleName              string              `json:"salename" bson:"salename"`
	MemberCode            string              `json:"membercode" bson:"membercode"`
	IsCancel              bool                `json:"iscancel" bson:"iscancel"`
	Status                int8                `json:"status" bson:"status"`
	Branch                TransactionBranch   `json:"branch" bson:"branch"`
	CancelDateTime        string              `json:"canceldatetime" bson:"canceldatetime"`
	CancelUserCode        string              `json:"cancelusercode" bson:"cancelusercode"`
	CancelUserName        string              `json:"cancelusername" bson:"cancelusername"`
	CancelDescription     string              `json:"canceldescription" bson:"canceldescription"`
	CancelReason          string              `json:"cancelreason" bson:"cancelreason"`
	PrintCopyBillDateTime []string            `json:"printcopybilldatetime" bson:"printcopybilldatetime"`
	CustomerTelephone     string              `json:"customertelephone" bson:"customertelephone"`
	TotalQty              float64             `json:"totalqty" bson:"totalqty"`
	RoundAmount           float64             `json:"roundamount" bson:"roundamount"`
	AccountNumber         string              `json:"accountnumber" bson:"accountnumber"`
	BankCode              string              `json:"bankcode" bson:"bankcode"`
	BankNames             []models.NameX      `json:"banknames" bson:"banknames"`
	BankBranch            string              `json:"bankbranch" bson:"bankbranch"`
	Remark                string              `json:"remark" bson:"remark"`
}

type Transaction struct {
	TransactionHeader `bson:"inline"`
	Details           *[]Detail `json:"details" bson:"details"`
}

type TransactionMoney struct {
	TransactionMoneyHeader `bson:"inline"`
	Details                *[]DetailMoney `json:"details" bson:"details"`
}

type DetailMoney struct {
	DocDatetime      time.Time       `json:"docdatetime" bson:"docdatetime"`
	BankCode         string          `json:"bankcode" bson:"bankcode"`
	BankNames        *[]models.NameX `json:"banknames" bson:"banknames"`
	BankBranch       string          `json:"bankbranch" bson:"bankbranch"`
	AccountNumber    string          `json:"accountnumber" bson:"accountnumber"`
	Description      string          `json:"description" bson:"description"`
	Amount           float64         `json:"amount" bson:"amount"`
	ToBankCode       string          `json:"tobankcode" bson:"tobankcode"`
	ToBankNames      *[]models.NameX `json:"tobanknames" bson:"tobanknames"`
	ToBankBranch     string          `json:"tobankbranch" bson:"tobankbranch"`
	ToAccountNo      string          `json:"toaccountno" bson:"toaccountno"`
	Fee              float64         `json:"fee" bson:"fee"`
	TotalValue       float64         `json:"totalvalue" bson:"totalvalue"`
	TotalExpense     float64         `json:"totalexpense" bson:"totalexpense"`
	TotalAmount      float64         `json:"totalamount" bson:"totalamount"`
	Remark           string          `json:"remark" bson:"remark"`
	CreditCardNumber string          `json:"creditcardnumber" bson:"creditcardnumber"`
	CreditCardType   string          `json:"creditcardtype" bson:"creditcardtype"`
	ExpiredDate      string          `json:"expireddate" bson:"expireddate"`
	Docref           string          `json:"docref" bson:"docref"`
	ChqueNumber      string          `json:"chequenumber" bson:"chequenumber"`
	ChqueBankCode    string          `json:"chequebankcode" bson:"chequebankcode"`
	ChqueBankNames   *[]models.NameX `json:"chequebanknames" bson:"chequebanknames"`
	ChqueDate        string          `json:"chequedate" bson:"chequedate"`
}

type TransactionMessageQueue struct {
	models.ShopIdentity `bson:"inline"`
	models.DocIdentity  `bson:"inline"`
	Transaction         `bson:"inline"`
}

type TransactionBranch struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
}

type TransactionDepositDoc struct {
	models.DocIdentity `bson:"inline"`
	DocNo              string    `json:"docno" bson:"docno"`
	DocDatetime        time.Time `json:"docdatetime" bson:"docdatetime"`
	TotalAmount        float64   `json:"totalamount" bson:"totalamount"`
	Balance            float64   `json:"balance" bson:"balance"`
	UseAmount          float64   `json:"useamount" bson:"useamount"`
	Remark             string    `json:"remark" bson:"remark"`
}

type TransactionDocRef struct {
	models.DocIdentity `bson:"inline"`
	DocNo              string    `json:"docno" bson:"docno"`
	DocDatetime        time.Time `json:"docdatetime" bson:"docdatetime"`
}

type JSONBTransactionBranch TransactionBranch

// Value Marshal
func (a JSONBTransactionBranch) Value() (driver.Value, error) {

	j, err := json.Marshal(a)
	return j, err
}

// Scan Unmarshal
func (a *JSONBTransactionBranch) Scan(value interface{}) error {

	dataBytes, ok := value.([]byte)

	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(dataBytes, &a)
}

type Detail struct {
	InquiryType         int8            `json:"inquirytype" bson:"inquirytype"`
	LineNumber          int             `json:"linenumber" bson:"linenumber"`
	DocDatetime         time.Time       `json:"docdatetime" bson:"docdatetime"`
	DocRef              string          `json:"docref" bson:"docref"`
	DocRefDatetime      time.Time       `json:"docrefdatetime" bson:"docrefdatetime"`
	CalcFlag            int8            `json:"calcflag" bson:"calcflag"`
	Barcode             string          `json:"barcode" bson:"barcode"`
	ItemCode            string          `json:"itemcode" bson:"itemcode"`
	ItemNames           *[]models.NameX `json:"itemnames" bson:"itemnames"`
	UnitCode            string          `json:"unitcode" bson:"unitcode"`
	UnitNames           *[]models.NameX `json:"unitnames" bson:"unitnames" `
	ItemType            int8            `json:"itemtype" bson:"itemtype"`
	ItemGuid            string          `json:"itemguid" bson:"itemguid"`
	ImageUri            string          `json:"imageuri" bson:"imageurl"`
	Description         string          `json:"description" bson:"description"`
	Qty                 float64         `json:"qty" bson:"qty"`
	EventQty            float64         `json:"eventqty" bson:"eventqty"`
	Reason              string          `json:"reason" bson:"reason"`
	TotalQty            float64         `json:"totalqty" bson:"totalqty"`
	Price               float64         `json:"price" bson:"price"`
	Discount            string          `json:"discount" bson:"discount"`
	DiscountAmount      float64         `json:"discountamount" bson:"discountamount"`
	TotalValueVat       float64         `json:"totalvaluevat" bson:"totalvaluevat"`
	PriceExcludeVat     float64         `json:"priceexcludevat" bson:"priceexcludevat"`
	SumAmount           float64         `json:"sumamount" bson:"sumamount"`
	SumAmountExcludeVat float64         `json:"sumamountexcludevat" bson:"sumamountexcludevat"`
	RefGuid             string          `json:"refguid" bson:"refguid"`
	DivideValue         float64         `json:"dividevalue" bson:"dividevalue"`
	StandValue          float64         `json:"standvalue" bson:"standvalue"`
	VatType             int8            `json:"vattype" bson:"vattype"`
	Remark              string          `json:"remark" bson:"remark"`
	MultiUnit           bool            `json:"multiunit" bson:"multiunit"`
	IssumPoint          bool            `json:"issumpoint" bson:"issumpoint"`
	SumOfCost           float64         `json:"sumofcost" bson:"sumofcost"`
	AverageCost         float64         `json:"averagecost" bson:"averagecost"`
	FoodType            int8            `json:"foodtype" bson:"column:foodtype"`
	LastStatus          int8            `json:"laststatus" bson:"laststatus"`
	IsChoice            int8            `json:"ischoice" bson:"ischoice"`
	IsPos               int8            `json:"ispos" bson:"ispos"`
	TaxType             int8            `json:"taxtype" bson:"taxtype"`
	VatCal              int             `json:"vatcal" bson:"vatcal"`
	WhCode              string          `json:"whcode" bson:"whcode"`
	WhNames             *[]models.NameX `json:"whnames" bson:"whnames"`
	ShelfCode           string          `json:"shelfcode" bson:"shelfcode"`
	LocationCode        string          `json:"locationcode" bson:"locationcode"`
	LocationNames       *[]models.NameX `json:"locationnames" bson:"locationnames"`
	ToWhCode            string          `json:"towhcode" bson:"towhcode"`
	ToWhNames           *[]models.NameX `json:"towhnames" bson:"towhnames"`
	ToLocationCode      string          `json:"tolocationcode" bson:"tolocationcode"`
	ToLocationNames     *[]models.NameX `json:"tolocationnames" bson:"tolocationnames" `
	SKU                 string          `json:"sku" bson:"sku"`
	ExtraJson           string          `json:"extrajson" bson:"extrajson"`
	GroupCode           string          `json:"groupcode" bson:"groupcode"`
	GroupNames          *[]models.NameX `json:"groupnames" bson:"groupnames"`
	ManufacturerGUID    string          `json:"manufacturerguid" bson:"manufacturerguid"`
	ManufacturerCode    string          `json:"manufacturercode" bson:"manufacturercode"`
	ManufacturerNames   *[]models.NameX `json:"manufacturernames" bson:"manufacturernames"`
	SumAmountChoice     float64         `json:"sumamountchoice" bson:"sumamountchoice"`

	// ราคาและยอดรวมในสกุลเงินเอกสาร (Document Currency)
	PriceDoc              float64 `json:"price_doc" bson:"price_doc,omitempty"`
	SumAmountDoc          float64 `json:"sumamount_doc" bson:"sumamount_doc,omitempty"`
	DiscountAmountDoc     float64 `json:"discountamount_doc" bson:"discountamount_doc,omitempty"`
	PriceExcludeVatDoc    float64 `json:"priceexcludevat_doc" bson:"priceexcludevat_doc,omitempty"`
	SumAmountExcludeVatDoc float64 `json:"sumamountexcludevat_doc" bson:"sumamountexcludevat_doc,omitempty"`
	TotalValueVatDoc      float64 `json:"totalvaluevat_doc" bson:"totalvaluevat_doc,omitempty"`
}

type PaymentDetail struct {
	CashAmountText     string               `json:"cashamounttext" bson:"cashamounttext"`
	CashAmount         float64              `json:"cashamount" bson:"cashamount"`
	PaymentCreditCards *[]PaymentCreditCard `json:"paymentcreditcards" bson:"paymentcreditcards"`
	PaymentTransfers   *[]PaymentTransfer   `json:"paymenttransfers" bson:"paymenttransfers"`
}

type PaymentCreditCard struct {
	DocDatetime   time.Time `json:"docdatetime" bson:"docdatetime"`
	CardNumber    string    `json:"cardnumber" bson:"cardnumber"`
	Amount        float64   `json:"amount" bson:"amount"`
	ChargeWord    string    `json:"chargeword" bson:"chargeword"`
	ChargeValue   float64   `json:"chargevalue" bson:"chargevalue"`
	TotalNetWorth float64   `json:"totalnetworth" bson:"totalnetworth"`
}

type PaymentTransfer struct {
	DocDatetime   time.Time       `json:"docdatetime" bson:"docdatetime"`
	BankCode      string          `json:"bankcode" bson:"bankcode"`
	BankNames     *[]models.NameX `json:"banknames" bson:"banknames"`
	AccountNumber string          `json:"accountnumber" bson:"accountnumber"`
	Amount        float64         `json:"amount" bson:"amount"`
}

// WHTEntry represents a single withholding tax entry
// เอกสาร 1 ใบอาจมีหลายอัตราภาษี ณ ที่จ่าย
type WHTEntry struct {
	Description string  `json:"description" bson:"description"`       // ประเภทเงินได้ (ค่าขนส่ง, ค่าบริการ, ค่าเช่า, etc.)
	Rate        float64 `json:"rate" bson:"rate"`                     // อัตราภาษี % (0.5, 1, 2, 3, 5, 10, 15)
	TaxBase     float64 `json:"taxbase" bson:"taxbase"`               // ยอดที่คำนวณได้ (ฐานภาษี)
	Amount      float64 `json:"amount" bson:"amount"`                 // จำนวนเงินหัก ณ ที่จ่าย (TaxBase × Rate%)
	Note        string  `json:"note,omitempty" bson:"note,omitempty"` // คำอธิบายเพิ่มเติม / หมายเหตุ
}

// SaleInvoiceCoupon represents a coupon used in a sale invoice
type SaleInvoiceCoupon struct {
	CouponNo          string  `json:"couponno" bson:"couponno"`
	CouponAmount      float64 `json:"couponamount" bson:"couponamount"`
	CouponDescription string  `json:"coupondescription" bson:"coupondescription"`
	CouponType        string  `json:"coupontype" bson:"coupontype"`
}
