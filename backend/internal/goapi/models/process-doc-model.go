package models

import "time"

// เอาไว้ใช้ map ข้อมูลจาก MongoDB ไปยัง Struct เพื่อใช้ต่อไป

type DocStruct struct {
	ShopID          string    `json:"shop_id" db:"shopid"`
	CustCode        string    `json:"cust_code" db:"custcode"`
	TransFlag       int       `json:"trans_flag" db:"transflag"`
	DocNo           string    `json:"doc_no" db:"docno"`
	DocDateTime     time.Time `json:"doc_date_time" db:"docdatetime"`
	PeriodDateTime  time.Time `json:"period_date_time" db:"perioddatetime"`
	TaxDocNo        string    `json:"tax_doc_no" db:"taxdocno"`
	TotalAmount     float64   `json:"total_amount" db:"totalamount"`
	RoundAmount     float64   `json:"round_amount" db:"roundamount"`
	PayType         int       `json:"pay_type" db:"paytype"`
	PayCashAmount   float64   `json:"pay_cash_amount" db:"paycashamount"`
	PayCashChange   float64   `json:"pay_cash_change" db:"paycashchange"`
	PayCashBalance  float64   `json:"pay_cash_balance" db:"paycashbalance"`
	DeliveryCode    string    `json:"delivery_code" db:"deliverycode"`
	Checksum        string    `json:"checksum" db:"checksum"`
	BranchID        string    `json:"branch_id" db:"branchid"`
	SlipURL         string    `json:"slip_url" db:"slipurl"`
	SaleChannelCode string    `json:"sale_channel_code" db:"salechannelcode"`
	DeliveryAmount  float64   `json:"delivery_amount" db:"deliveryamount"`
	IsCancel        bool      `json:"is_cancel" db:"iscancel"`
	CancelReason    string    `json:"cancel_reason" db:"cancelreason"`
	GuidPOS         string    `json:"guid_pos" db:"guidpos"`
	GuidBranch      string    `json:"guid_branch" db:"guidbranch"`
	GuidFixed       string    `json:"guid_fixed" db:"guidfixed"`
	// ข้อมูลผู้สร้างเอกสาร
	CreatorCode string    `json:"creator_code" db:"creator_code"` // รหัสผู้สร้างเอกสาร
	CreatorName string    `json:"creator_name" db:"creator_name"` // ชื่อผู้สร้างเอกสาร
	CreatedAt   time.Time `json:"created_at" db:"created_at"`     // วันเวลาที่สร้างเอกสาร
	
	// ============ Multi-Currency Fields ============
	// สกุลเงินหลัก (Base Currency) - สำหรับลงบัญชี (field เดิม)
	Currency       string  `json:"currency" db:"currency"`
	CurrencySymbol string  `json:"currencysymbol" db:"currencysymbol"`
	
	// สกุลเงินเอกสาร (Document Currency) - field ใหม่
	DocCurrency       string  `json:"doc_currency" db:"doccurrency"`
	DocCurrencySymbol string  `json:"doc_currencysymbol" db:"doccurrencysymbol"`
	ExchangeRate      float64 `json:"exchangerate" db:"exchangerate"`
	
	// ยอดรวมในสกุลเงินเอกสาร
	TotalAmountDoc float64 `json:"totalamount_doc" db:"totalamountdoc"`

	// Soft Delete
	IsDelete bool `json:"isdelete" db:"isdelete"`

	// สถานะการอนุมัติ (จาก po_approval_status collection)
	ApprovalStatus string `json:"approval_status" db:"approval_status"`
}

type DocRefStruct struct {
	DocNo             string `json:"doc_no" db:"docno"`
	DocNoTransFlag    int    `json:"doc_no_trans_flag" db:"docnotransflag"`
	DocRefNo          string `json:"ref_doc_no" db:"refdocno"`
	DocRefNoTransFlag int    `json:"ref_doc_no_trans_flag" db:"refdocnotransflag"`
}

type DocDetailStruct struct {
	DocDateTime          time.Time `json:"doc_date_time" bson:"doc_date_time"`
	DocNo                string    `json:"doc_no" bson:"doc_no"`
	TransFlag            int       `json:"trans_flag" bson:"trans_flag"`
	CalcFlag             float64   `json:"calc_flag" bson:"calc_flag"`
	CalcSeq              int       `json:"calc_seq" bson:"calc_seq"`
	LineNumber           int       `json:"line_number" bson:"line_number"`
	Barcode              string    `json:"barcode" bson:"barcode"`
	UnitCode             string    `json:"unit_code" bson:"unit_code"`
	WhCode               string    `json:"wh_code" bson:"wh_code"`
	LocationCode         string    `json:"location_code" bson:"location_code"`
	ToWhCode             string    `json:"to_wh_code" bson:"to_wh_code"`
	ToLocationCode       string    `json:"to_location_code" bson:"to_location_code"`
	TotalQty             float64   `json:"total_qty" bson:"total_qty"`
	Price                float64   `json:"price" bson:"price"`
	PriceExcludeVat      float64   `json:"price_exclude_vat" bson:"price_exclude_vat"`
	ItemCode             string    `json:"item_code" bson:"item_code"`
	BarcodeMain          string    `json:"barcode_main" bson:"barcode_main"`
	BarcodeRefUnitStand  float64   `json:"barcode_ref_unit_stand" bson:"barcode_ref_unit_stand"`
	BarcodeRefUnitDivide float64   `json:"barcode_ref_unit_divide" bson:"barcode_ref_unit_divide"`
	UnitStand            float64   `json:"unit_stand" bson:"unit_stand"`
	UnitDivide           float64   `json:"unit_divide" bson:"unit_divide"`
	DocRef               string    `json:"doc_ref" bson:"doc_ref"`
	Description          string    `json:"description" bson:"description"`
	SumAmount            float64   `json:"sumamount" bson:"sumamount"`
	IsCalcStock          int8      `json:"is_calc_stock" bson:"is_calc_stock"` // 1 = มีการคำนวณสต็อก (ซื้อ/ขาย), 0 = ไม่คำนวณ

	// ราคาและยอดรวมในสกุลเงินเอกสาร (Document Currency)
	PriceDoc               float64 `json:"price_doc" bson:"price_doc"`
	SumAmountDoc           float64 `json:"sumamount_doc" bson:"sumamount_doc"`
	DiscountAmountDoc      float64 `json:"discountamount_doc" bson:"discountamount_doc"`
	PriceExcludeVatDoc     float64 `json:"priceexcludevat_doc" bson:"priceexcludevat_doc"`
	SumAmountExcludeVatDoc float64 `json:"sumamountexcludevat_doc" bson:"sumamountexcludevat_doc"`
	TotalValueVatDoc       float64 `json:"totalvaluevat_doc" bson:"totalvaluevat_doc"`
}

type DocPaymentStruct struct {
	ShopID         string    `json:"shop_id" db:"shopid"`
	BranchID       string    `json:"branch_id" db:"branchid"`
	DocDateTime    time.Time `json:"doc_date_time" db:"docdatetime"`
	PeriodDateTime time.Time `json:"period_date_time" db:"perioddatetime"`
	ProviderName   string    `json:"provider_name" db:"providername"`
	Amount         float64   `json:"amount" db:"amount"`
	Description    string    `json:"description" db:"description"`
	DocNo          string    `json:"doc_no" db:"docno"`
	TransFlag      int32     `json:"trans_flag" db:"trans_flag"`
	GuidFixed      string    `json:"guid_fixed" db:"guidfixed"`
	GuidBranch     string    `json:"guid_branch" db:"guidbranch"`
}
