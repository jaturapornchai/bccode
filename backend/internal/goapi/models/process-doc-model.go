package models

import "time"

// เอาไว้ใช้ map ข้อมูลจาก MongoDB ไปยัง Struct เพื่อใช้ต่อไป

type DocStruct struct {
	HoldingCode     string    `json:"holdingcode" db:"holdingcode"`
	CustCode        string    `json:"custcode" db:"custcode"`
	TransFlag       int       `json:"transflag" db:"transflag"`
	DocNo           string    `json:"docno" db:"docno"`
	DocDateTime     time.Time `json:"docdatetime" db:"docdatetime"`
	PeriodDateTime  time.Time `json:"perioddatetime" db:"perioddatetime"`
	TaxDocNo        string    `json:"taxdocno" db:"taxdocno"`
	TotalAmount     float64   `json:"totalamount" db:"totalamount"`
	RoundAmount     float64   `json:"roundamount" db:"roundamount"`
	PayType         int       `json:"paytype" db:"paytype"`
	PayCashAmount   float64   `json:"paycashamount" db:"paycashamount"`
	PayCashChange   float64   `json:"paycashchange" db:"paycashchange"`
	PayCashBalance  float64   `json:"paycashbalance" db:"paycashbalance"`
	DeliveryCode    string    `json:"deliverycode" db:"deliverycode"`
	Checksum        string    `json:"checksum" db:"checksum"`
	BranchID        string    `json:"branchid" db:"branchid"`
	SlipURL         string    `json:"slipurl" db:"slipurl"`
	SaleChannelCode string    `json:"salechannelcode" db:"salechannelcode"`
	DeliveryAmount  float64   `json:"deliveryamount" db:"deliveryamount"`
	IsCancel        bool      `json:"iscancel" db:"iscancel"`
	CancelReason    string    `json:"cancelreason" db:"cancelreason"`
	GuidPOS         string    `json:"guidpos" db:"guidpos"`
	GuidBranch      string    `json:"guidbranch" db:"guidbranch"`
	GuidFixed       string    `json:"guidfixed" db:"guidfixed"`
	// ข้อมูลผู้สร้างเอกสาร
	CreatorCode string    `json:"creatorcode" db:"creator_code"` // รหัสผู้สร้างเอกสาร
	CreatorName string    `json:"creatorname" db:"creator_name"` // ชื่อผู้สร้างเอกสาร
	CreatedAt   time.Time `json:"createdat" db:"created_at"`     // วันเวลาที่สร้างเอกสาร

	// ============ Multi-Currency Fields ============
	// สกุลเงินหลัก (Base Currency) - สำหรับลงบัญชี (field เดิม)
	Currency       string `json:"currency" db:"currency"`
	CurrencySymbol string `json:"currencysymbol" db:"currency_symbol"`

	// สกุลเงินเอกสาร (Document Currency) - field ใหม่
	DocCurrency       string  `json:"doccurrency" db:"doc_currency"`
	DocCurrencySymbol string  `json:"doccurrencysymbol" db:"doc_currency_symbol"`
	ExchangeRate      float64 `json:"exchangerate" db:"exchange_rate"`

	// ยอดรวมในสกุลเงินเอกสาร
	TotalAmountDoc float64 `json:"totalamountdoc" db:"totalamount_doc"`

	// Soft Delete
	IsDelete bool `json:"isdelete" db:"isdelete"`

	// Approval status from poapprovalstatus collection.
	ApprovalStatus string `json:"approvalstatus" db:"approval_status"`
}

type DocRefStruct struct {
	DocNo             string `json:"docno" db:"docno"`
	DocNoTransFlag    int    `json:"docnotransflag" db:"docnotransflag"`
	DocRefNo          string `json:"refdocno" db:"docnoref"`
	DocRefNoTransFlag int    `json:"refdocnotransflag" db:"docnoreftransflag"`
}

type DocDetailStruct struct {
	DocDateTime          time.Time `json:"docdatetime" bson:"docdatetime"`
	DocNo                string    `json:"docno" bson:"docno"`
	TransFlag            int       `json:"transflag" bson:"transflag"`
	CalcFlag             float64   `json:"calcflag" bson:"calcflag"`
	CalcSeq              int       `json:"calcseq" bson:"calcseq"`
	LineNumber           int       `json:"linenumber" bson:"linenumber"`
	Barcode              string    `json:"barcode" bson:"barcode"`
	UnitCode             string    `json:"unitcode" bson:"unitcode"`
	WhCode               string    `json:"whcode" bson:"whcode"`
	LocationCode         string    `json:"locationcode" bson:"locationcode"`
	ToWhCode             string    `json:"towhcode" bson:"towhcode"`
	ToLocationCode       string    `json:"tolocationcode" bson:"tolocationcode"`
	TotalQty             float64   `json:"totalqty" bson:"totalqty"`
	Price                float64   `json:"price" bson:"price"`
	PriceExcludeVat      float64   `json:"priceexcludevat" bson:"priceexcludevat"`
	ItemCode             string    `json:"itemcode" bson:"itemcode"`
	BarcodeMain          string    `json:"barcodemain" bson:"barcodemain"`
	BarcodeRefUnitStand  float64   `json:"barcoderefunitstand" bson:"barcoderefunitstand"`
	BarcodeRefUnitDivide float64   `json:"barcoderefunitdivide" bson:"barcoderefunitdivide"`
	UnitStand            float64   `json:"unitstand" bson:"unitstand"`
	UnitDivide           float64   `json:"unitdivide" bson:"unitdivide"`
	DocRef               string    `json:"docref" bson:"docref"`
	Description          string    `json:"description" bson:"description"`
	SumAmount            float64   `json:"sumamount" bson:"sumamount"`
	IsCalcStock          int8      `json:"iscalcstock" bson:"iscalcstock"` // 1 = มีการคำนวณสต็อก (ซื้อ/ขาย), 0 = ไม่คำนวณ

	// ราคาและยอดรวมในสกุลเงินเอกสาร (Document Currency)
	PriceDoc               float64 `json:"pricedoc" bson:"pricedoc"`
	SumAmountDoc           float64 `json:"sumamountdoc" bson:"sumamountdoc"`
	DiscountAmountDoc      float64 `json:"discountamountdoc" bson:"discountamountdoc"`
	PriceExcludeVatDoc     float64 `json:"priceexcludevatdoc" bson:"priceexcludevatdoc"`
	SumAmountExcludeVatDoc float64 `json:"sumamountexcludevatdoc" bson:"sumamountexcludevatdoc"`
	TotalValueVatDoc       float64 `json:"totalvaluevatdoc" bson:"totalvaluevatdoc"`
}

type DocPaymentStruct struct {
	HoldingCode    string    `json:"holdingcode" db:"holdingcode"`
	BranchID       string    `json:"branchid" db:"branchid"`
	DocDateTime    time.Time `json:"docdatetime" db:"docdatetime"`
	PeriodDateTime time.Time `json:"perioddatetime" db:"perioddatetime"`
	ProviderName   string    `json:"providername" db:"providername"`
	Amount         float64   `json:"amount" db:"amount"`
	Description    string    `json:"description" db:"description"`
	DocNo          string    `json:"docno" db:"docno"`
	TransFlag      int32     `json:"transflag" db:"transflag"`
	GuidFixed      string    `json:"guidfixed" db:"guidfixed"`
	GuidBranch     string    `json:"guidbranch" db:"guidbranch"`
}
