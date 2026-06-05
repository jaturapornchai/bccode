package models

import (
	"time"
)

type LanguageModel struct {
	Code     string `json:"code" bson:"code"`
	Name     string `json:"name" bson:"name"`
	IsAuto   bool   `json:"isauto" bson:"isauto"`
	IsDelete bool   `json:"isdelete" bson:"isdelete"`
}

type MongoDocDetailModel struct {
	LineNumber          int             `json:"linenumber" bson:"linenumber"`
	DocDateTime         time.Time       `json:"docdatetime" bson:"docdatetime"`
	DocRef              string          `json:"docref" bson:"docref"`
	DocRefDateTime      time.Time       `json:"docrefdatetime" bson:"docrefdatetime"`
	Barcode             string          `json:"barcode" bson:"barcode"`
	ItemCode            string          `json:"itemcode" bson:"itemcode"`
	ItemNames           []LanguageModel `json:"itemnames" bson:"itemnames"`
	UnitCode            string          `json:"unitcode" bson:"unitcode"`
	UnitNames           []LanguageModel `json:"unitnames" bson:"unitnames"`
	ItemType            int             `json:"itemtype" bson:"itemtype"`
	ItemGuid            string          `json:"itemguid" bson:"itemguid"`
	Description         string          `json:"description" bson:"description"`
	Qty                 float64         `json:"qty" bson:"qty"`
	EventQty            float64         `json:"eventqty" bson:"eventqty"`
	TotalQty            float64         `json:"totalqty" bson:"totalqty"`
	Price               float64         `json:"price" bson:"price"`
	PriceExcludeVat     float64         `json:"priceexcludevat" bson:"priceexcludevat"`
	Discount            string          `json:"discount" bson:"discount"`
	DiscountAmount      float64         `json:"discountamount" bson:"discountamount"`
	TotalValueVat       float64         `json:"totalvaluevat" bson:"totalvaluevat"`
	SumAmount           float64         `json:"sumamount" bson:"sumamount"`
	SumAmountExcludeVat float64         `json:"sumamountexcludevat" bson:"sumamountexcludevat"`
	SumAmountChoice     float64         `json:"sumamountchoice" bson:"sumamountchoice"`
	RefGuid             string          `json:"refguid" bson:"refguid"`
	DivideValue         float64         `json:"dividevalue" bson:"dividevalue"`
	StandValue          float64         `json:"standvalue" bson:"standvalue"`
	VatType             int             `json:"vattype"`
	Remark              string          `json:"remark"`
	MultiUnit           bool            `json:"multiunit"`
	IsSumPoint          bool            `json:"issumpoint"`
	SumOfCost           float64         `json:"sumofcost"`
	AverageCost         float64         `json:"averagecost"`
	FoodType            int             `json:"foodtype"`
	LastStatus          int             `json:"laststatus"`
	IsChoice            int             `json:"ischoice"`
	IsPos               int             `json:"ispos"`
	TaxType             int             `json:"taxtype"`
	VatCal              int             `json:"vatcal"`
	WhCode              string          `json:"whcode"`
	WhNames             []LanguageModel `json:"whnames"`
	ShelfCode           string          `json:"shelfcode"`
	LocationCode        string          `json:"locationcode"`
	LocationNames       []LanguageModel `json:"locationnames"`
	ToWhCode            string          `json:"towhcode"`
	ToWhNames           []LanguageModel `json:"towhnames"`
	ToLocationCode      string          `json:"tolocationcode"`
	ToLocationNames     []LanguageModel `json:"tolocationnames"`
	Sku                 string          `json:"sku"`
	ExtraJson           string          `json:"extrajson"`
	GroupCode           string          `json:"groupcode"`
	GroupNames          []LanguageModel `json:"groupnames"`
	ManufacturerGuid    string          `json:"manufacturerguid"`
	ManufacturerCode    string          `json:"manufacturercode"`
	ManufacturerNames   []LanguageModel `json:"manufacturernames"`
	CalcFlag            int             `json:"calcflag" bson:"calcflag"`
	CalcSeq             int             `json:"calcseq" bson:"calcseq"`

	// ราคาและยอดรวมในสกุลเงินเอกสาร (Document Currency)
	// mapstructure tags จำเป็นเพราะ keys มี underscore (เช่น "price_doc" ≠ "PriceDoc" ใน mapstructure)
	PriceDoc               float64 `json:"pricedoc" bson:"pricedoc" mapstructure:"price_doc"`
	SumAmountDoc           float64 `json:"sumamountdoc" bson:"sumamountdoc" mapstructure:"sumamount_doc"`
	DiscountAmountDoc      float64 `json:"discountamountdoc" bson:"discountamountdoc" mapstructure:"discountamount_doc"`
	PriceExcludeVatDoc     float64 `json:"priceexcludevatdoc" bson:"priceexcludevatdoc" mapstructure:"priceexcludevat_doc"`
	SumAmountExcludeVatDoc float64 `json:"sumamountexcludevatdoc" bson:"sumamountexcludevatdoc" mapstructure:"sumamountexcludevat_doc"`
	TotalValueVatDoc       float64 `json:"totalvaluevatdoc" bson:"totalvaluevatdoc" mapstructure:"totalvaluevat_doc"`
}

// ProcessMongoTransDetailTransFlag54Model is used for TransFlag 54 (ยอดยกมา)
type ProcessMongoTransDetailTransFlag54Model struct {
	ItemCode        string          `json:"itemcode" bson:"itemcode"`
	ItemNames       []LanguageModel `json:"itemnames" bson:"itemnames"`
	Barcode         string          `json:"barcode" bson:"barcode"`
	UnitCode        string          `json:"unitcode" bson:"unitcode"`
	LineNumber      int             `json:"linenumber" bson:"linenumber"`
	WhCode          string          `json:"whcode" bson:"whcode"`
	LocationCode    string          `json:"locationcode" bson:"locationcode"`
	ToWhCode        string          `json:"towhcode" bson:"towhcode"`
	ToLocationCode  string          `json:"tolocationcode" bson:"tolocationcode"`
	Qty             float64         `json:"qty" bson:"qty"`
	CalcFlag        int             `json:"calcflag" bson:"calcflag"`
	CalcSeq         int             `json:"calcseq" bson:"calcseq"`
	Price           float64         `json:"price" bson:"price"`
	PriceExcludeVat float64         `json:"priceexcludevat" bson:"priceexcludevat"`
	DocRef          string          `json:"docref" bson:"docref"`
	SumAmount       float64         `json:"sumamount" bson:"sumamount"`
	DocNo           string          `json:"docno" bson:"docno"`
	DocDateTime     time.Time       `json:"docdatetime" bson:"docdatetime"`
}

type MongoBranchModel struct {
	Code      string          `json:"code" bson:"code"`
	GuidFixed string          `json:"guidfixed" bson:"guidfixed"`
	Names     []LanguageModel `json:"names" bson:"names"`
}

type MongoDocReferenceModel struct {
	DocNo       string    `json:"docno" bson:"docno"`
	DocDateTime time.Time `json:"docdatetime" bson:"docdatetime"`
}

type MongoDocModel struct {
	HoldingCode    string                   `json:"holdingcode" bson:"holdingcode"`
	BranchId       string                   `json:"branchid" bson:"branchid"`
	GuidFixed      string                   `json:"guidfixed" bson:"guidfixed"`
	DocNo          string                   `json:"docno" bson:"docno"`
	Description    string                   `json:"description" bson:"description"`
	CreatorCode    string                   `json:"creatorcode" bson:"creatorcode" mapstructure:"creator_code"`   // รหัสผู้สร้าง (JSON snake_case ตรงกับ mainapi)
	CreatorName    string                   `json:"creatorname" bson:"creatorname" mapstructure:"creator_name"`   // ชื่อผู้สร้าง
	CreatedAt      time.Time                `json:"createdat" bson:"createdat" mapstructure:"createdat"`          // วันเวลาที่สร้างเอกสาร
	ModifierCode   string                   `json:"modifiercode" bson:"updatercode" mapstructure:"modifier_code"` // รหัสผู้แก้ไข
	ModifierName   string                   `json:"modifiername" bson:"updatername" mapstructure:"modifier_name"` // ชื่อผู้แก้ไข
	ModifiedAt     time.Time                `json:"modifiedat" bson:"updatedat" mapstructure:"modified_at"`       // วันเวลาที่แก้ไข
	DocDateTime    time.Time                `json:"docdatetime" bson:"docdatetime"`
	DocRefDate     time.Time                `json:"docrefdate" bson:"docrefdate"`
	DocReferences  []MongoDocReferenceModel `json:"docreferences" bson:"docreferences"`
	TaxDocDate     time.Time                `json:"taxdocdate" bson:"taxdocdate"`
	TaxDocNo       string                   `json:"taxdocno" bson:"taxdocno"`
	TransFlag      int                      `json:"transflag" bson:"transflag"`
	VatType        int                      `json:"vattype" bson:"vattype"`
	VatRate        float64                  `json:"vatrate" bson:"vatrate"`
	CustCode       string                   `json:"custcode" bson:"custcode"`
	ManCount       int                      `json:"mancount" bson:"mancount"`
	WomanCount     int                      `json:"womancount" bson:"womancount"`
	ChildCount     int                      `json:"childcount" bson:"childcount"`
	TotalAmount    float64                  `json:"totalamount" bson:"totalamount"`
	TotalValue     float64                  `json:"totalvalue" bson:"totalvalue"`
	TotalBeforeVat float64                  `json:"totalbeforevat" bson:"totalbeforevat"`
	TotalVatValue  float64                  `json:"totalvatvalue" bson:"totalvatvalue"`

	// ============ Multi-Currency Fields ============
	// สกุลเงินหลัก (Base Currency) - สำหรับลงบัญชี (field เดิม)
	Currency       string `json:"currency" bson:"currency,omitempty"`
	CurrencySymbol string `json:"currencysymbol" bson:"currencysymbol,omitempty"`

	// สกุลเงินเอกสาร (Document Currency) - field ใหม่
	// mapstructure tags จำเป็นเพราะ DecodeKafkaMessage ใช้ mapstructure (match field name case-insensitive)
	// keys ที่มี underscore เช่น "doc_currency" จะไม่ match กับ field name "DocCurrency" (= "doccurrency")
	DocCurrency       string  `json:"doccurrency" bson:"doccurrency,omitempty" mapstructure:"doc_currency"`
	DocCurrencySymbol string  `json:"doccurrencysymbol" bson:"doccurrencysymbol,omitempty" mapstructure:"doc_currencysymbol"`
	ExchangeRate      float64 `json:"exchangerate" bson:"exchangerate,omitempty"`

	// ยอดรวมในสกุลเงินเอกสาร
	TotalAmountDoc float64 `json:"totalamountdoc" bson:"totalamountdoc,omitempty" mapstructure:"totalamount_doc"`

	PayCashAmount    float64               `json:"paycashamount" bson:"paycashamount"`
	PayCashChange    float64               `json:"paycashchange" bson:"paycashchange"`
	RoundAmount      float64               `json:"roundamount" bson:"roundamount"`
	PaymentDetailRaw string                `json:"paymentdetailraw" bson:"paymentdetailraw"`
	SlipUrl          string                `json:"slipurl" bson:"slipurl"`
	SaleChannelCode  string                `json:"salechannelcode" bson:"salechannelcode"`
	DeliveryAmount   float64               `json:"deliveryamount" bson:"deliveryamount"`
	Details          []MongoDocDetailModel `json:"details" bson:"details"`
	IsCancel         bool                  `json:"iscancel" bson:"iscancel"`
	CancelReason     string                `json:"cancelreason" bson:"cancelreason"`
	GuidPos          string                `json:"guidpos" bson:"guidpos"`
	Branch           MongoBranchModel      `json:"branch" bson:"branch"`
	DeletedAt        *time.Time            `json:"deletedat" bson:"deletedat"`
	IsDelete         bool                  `json:"isdelete" bson:"isdelete"`
}

// Type aliases for backward compatibility
type ProcessMongoTransModel = MongoDocModel
type ProcessMongoTransDetailModel = MongoDocDetailModel
type ProcessMongoDocReferenceModel = MongoDocReferenceModel
type BranchModel = MongoBranchModel

// StockTransferStruct represents a stock transfer document
type StockTransferStruct struct {
	HoldingCode        string                      `json:"holdingcode" bson:"holdingcode"`
	BranchCode         string                      `json:"branchcode" bson:"branchcode"`
	DocNo              string                      `json:"docno" bson:"docno"`
	RefNo              string                      `json:"refno" bson:"refno"`
	Description        string                      `json:"description" bson:"description"`
	DocDateTime        time.Time                   `json:"docdatetime" bson:"docdatetime"`
	TotalAmount        float64                     `json:"totalamount" bson:"totalamount"`
	CustCode           string                      `json:"custcode" bson:"custcode"`
	IsCancel           bool                        `json:"iscancel" bson:"iscancel"`
	Guid               string                      `json:"guid" bson:"guid"`
	DocCreatedDateTime time.Time                   `json:"doccreateddatetime" bson:"doccreateddatetime"`
	DocUpdatedDateTime time.Time                   `json:"docupdateddatetime" bson:"docupdateddatetime"`
	IsDelete           bool                        `json:"isdelete" bson:"isdelete"`
	Branch             MongoBranchModel            `json:"branch" bson:"branch"`
	Details            []StockTransferDetailStruct `json:"details" bson:"details"`
}

// StockTransferDetailStruct represents a stock transfer detail line
type StockTransferDetailStruct struct {
	LineNumber      int     `json:"linenumber" bson:"linenumber"`
	ItemCode        string  `json:"itemcode" bson:"itemcode"`
	Description     string  `json:"description" bson:"description"`
	BarcodeMain     string  `json:"barcodemain" bson:"barcodemain"`
	Barcode         string  `json:"barcode" bson:"barcode"`
	UnitCode        string  `json:"unitcode" bson:"unitcode"`
	WhCode          string  `json:"whcode" bson:"whcode"`
	LocationCode    string  `json:"locationcode" bson:"locationcode"`
	ToWhCode        string  `json:"towhcode" bson:"towhcode"`
	ToLocationCode  string  `json:"tolocationcode" bson:"tolocationcode"`
	Qty             float64 `json:"qty" bson:"qty"`
	Price           float64 `json:"price" bson:"price"`
	PriceExcludeVat float64 `json:"priceexcludevat" bson:"priceexcludevat"`
	UnitStand       float64 `json:"unitstand" bson:"unitstand"`
	UnitDivide      float64 `json:"unitdivide" bson:"unitdivide"`
	DocRef          string  `json:"docref" bson:"docref"`
	SumAmount       float64 `json:"sumamount" bson:"sumamount"`
}

// StockReceiveProductStruct represents a stock receive product document (TransFlag 60)
type StockReceiveProductStruct struct {
	HoldingCode        string                            `json:"holdingcode" bson:"holdingcode"`
	BranchId           string                            `json:"branchid" bson:"branchid"`
	GuidFixed          string                            `json:"guidfixed" bson:"guidfixed"`
	DocNo              string                            `json:"docno" bson:"docno"`
	Description        string                            `json:"description" bson:"description"`
	DocDateTime        time.Time                         `json:"docdatetime" bson:"docdatetime"`
	TotalAmount        float64                           `json:"totalamount" bson:"totalamount"`
	RoundAmount        float64                           `json:"roundamount" bson:"roundamount"`
	PayCashAmount      float64                           `json:"paycashamount" bson:"paycashamount"`
	PayCashChange      float64                           `json:"paycashchange" bson:"paycashchange"`
	PaymentDetailRaw   string                            `json:"paymentdetailraw" bson:"paymentdetailraw"`
	SlipUrl            string                            `json:"slipurl" bson:"slipurl"`
	SaleChannelCode    string                            `json:"salechannelcode" bson:"salechannelcode"`
	DeliveryAmount     float64                           `json:"deliveryamount" bson:"deliveryamount"`
	IsCancel           bool                              `json:"iscancel" bson:"iscancel"`
	CancelReason       string                            `json:"cancelreason" bson:"cancelreason"`
	GuidPos            string                            `json:"guidpos" bson:"guidpos"`
	Branch             MongoBranchModel                  `json:"branch" bson:"branch"`
	Details            []StockReceiveProductDetailStruct `json:"details" bson:"details"`
	DocCreatedDateTime time.Time                         `json:"doccreateddatetime" bson:"doccreateddatetime"`
	DocUpdatedDateTime time.Time                         `json:"docupdateddatetime" bson:"docupdateddatetime"`
	IsDelete           bool                              `json:"isdelete" bson:"isdelete"`
}

// StockReceiveProductDetailStruct represents a stock receive product detail line
type StockReceiveProductDetailStruct struct {
	LineNumber      int             `json:"linenumber" bson:"linenumber"`
	ItemCode        string          `json:"itemcode" bson:"itemcode"`
	ItemNames       []LanguageModel `json:"itemnames" bson:"itemnames"`
	Description     string          `json:"description" bson:"description"`
	Barcode         string          `json:"barcode" bson:"barcode"`
	UnitCode        string          `json:"unitcode" bson:"unitcode"`
	WhCode          string          `json:"whcode" bson:"whcode"`
	LocationCode    string          `json:"locationcode" bson:"locationcode"`
	Qty             float64         `json:"qty" bson:"qty"`
	Price           float64         `json:"price" bson:"price"`
	PriceExcludeVat float64         `json:"priceexcludevat" bson:"priceexcludevat"`
	DocRef          string          `json:"docref" bson:"docref"`
	SumAmount       float64         `json:"sumamount" bson:"sumamount"`
}

// StockPickupProductStruct represents a stock pickup product document (TransFlag 56)
type StockPickupProductStruct struct {
	HoldingCode        string                           `json:"holdingcode" bson:"holdingcode"`
	BranchId           string                           `json:"branchid" bson:"branchid"`
	GuidFixed          string                           `json:"guidfixed" bson:"guidfixed"`
	DocNo              string                           `json:"docno" bson:"docno"`
	Description        string                           `json:"description" bson:"description"`
	DocDateTime        time.Time                        `json:"docdatetime" bson:"docdatetime"`
	TotalAmount        float64                          `json:"totalamount" bson:"totalamount"`
	RoundAmount        float64                          `json:"roundamount"`
	PayCashAmount      float64                          `json:"paycashamount" bson:"paycashamount"`
	PayCashChange      float64                          `json:"paycashchange" bson:"paycashchange"`
	PaymentDetailRaw   string                           `json:"paymentdetailraw" bson:"paymentdetailraw"`
	SlipUrl            string                           `json:"slipurl" bson:"slipurl"`
	SaleChannelCode    string                           `json:"salechannelcode" bson:"salechannelcode"`
	DeliveryAmount     float64                          `json:"deliveryamount" bson:"deliveryamount"`
	IsCancel           bool                             `json:"iscancel" bson:"iscancel"`
	CancelReason       string                           `json:"cancelreason" bson:"cancelreason"`
	GuidPos            string                           `json:"guidpos" bson:"guidpos"`
	Branch             MongoBranchModel                 `json:"branch" bson:"branch"`
	Details            []StockPickupProductDetailStruct `json:"details" bson:"details"`
	DocCreatedDateTime time.Time                        `json:"doccreateddatetime" bson:"doccreateddatetime"`
	DocUpdatedDateTime time.Time                        `json:"docupdateddatetime" bson:"docupdateddatetime"`
	IsDelete           bool                             `json:"isdelete" bson:"isdelete"`
}

// StockPickupProductDetailStruct represents a stock pickup product detail line
type StockPickupProductDetailStruct struct {
	LineNumber      int             `json:"linenumber" bson:"linenumber"`
	ItemCode        string          `json:"itemcode" bson:"itemcode"`
	ItemNames       []LanguageModel `json:"itemnames" bson:"itemnames"`
	Description     string          `json:"description" bson:"description"`
	Barcode         string          `json:"barcode" bson:"barcode"`
	UnitCode        string          `json:"unitcode" bson:"unitcode"`
	WhCode          string          `json:"whcode" bson:"whcode"`
	LocationCode    string          `json:"locationcode" bson:"locationcode"`
	Qty             float64         `json:"qty" bson:"qty"`
	Price           float64         `json:"price" bson:"price"`
	PriceExcludeVat float64         `json:"priceexcludevat" bson:"priceexcludevat"`
	DocRef          string          `json:"docref" bson:"docref"`
	SumAmount       float64         `json:"sumamount" bson:"sumamount"`
}

// StockReturnProductStruct represents a stock return product document (TransFlag 58)
type StockReturnProductStruct struct {
	HoldingCode        string                           `json:"holdingcode" bson:"holdingcode"`
	BranchId           string                           `json:"branchid" bson:"branchid"`
	GuidFixed          string                           `json:"guidfixed" bson:"guidfixed"`
	DocNo              string                           `json:"docno" bson:"docno"`
	Description        string                           `json:"description" bson:"description"`
	DocDateTime        time.Time                        `json:"docdatetime" bson:"docdatetime"`
	TotalAmount        float64                          `json:"totalamount" bson:"totalamount"`
	RoundAmount        float64                          `json:"roundamount"`
	PayCashAmount      float64                          `json:"paycashamount" bson:"paycashamount"`
	PayCashChange      float64                          `json:"paycashchange" bson:"paycashchange"`
	PaymentDetailRaw   string                           `json:"paymentdetailraw" bson:"paymentdetailraw"`
	SlipUrl            string                           `json:"slipurl" bson:"slipurl"`
	SaleChannelCode    string                           `json:"salechannelcode" bson:"salechannelcode"`
	DeliveryAmount     float64                          `json:"deliveryamount" bson:"deliveryamount"`
	IsCancel           bool                             `json:"iscancel" bson:"iscancel"`
	CancelReason       string                           `json:"cancelreason" bson:"cancelreason"`
	GuidPos            string                           `json:"guidpos" bson:"guidpos"`
	Branch             MongoBranchModel                 `json:"branch" bson:"branch"`
	Details            []StockReturnProductDetailStruct `json:"details" bson:"details"`
	DocCreatedDateTime time.Time                        `json:"doccreateddatetime" bson:"doccreateddatetime"`
	DocUpdatedDateTime time.Time                        `json:"docupdateddatetime" bson:"docupdateddatetime"`
	IsDelete           bool                             `json:"isdelete" bson:"isdelete"`
}

// StockReturnProductDetailStruct represents a stock return product detail line
type StockReturnProductDetailStruct struct {
	LineNumber      int             `json:"linenumber" bson:"linenumber"`
	ItemCode        string          `json:"itemcode" bson:"itemcode"`
	ItemNames       []LanguageModel `json:"itemnames" bson:"itemnames"`
	Description     string          `json:"description" bson:"description"`
	Barcode         string          `json:"barcode" bson:"barcode"`
	UnitCode        string          `json:"unitcode" bson:"unitcode"`
	WhCode          string          `json:"whcode" bson:"whcode"`
	LocationCode    string          `json:"locationcode" bson:"locationcode"`
	Qty             float64         `json:"qty" bson:"qty"`
	Price           float64         `json:"price" bson:"price"`
	PriceExcludeVat float64         `json:"priceexcludevat" bson:"priceexcludevat"`
	DocRef          string          `json:"docref" bson:"docref"`
	SumAmount       float64         `json:"sumamount" bson:"sumamount"`
}

// StockAdjustmentStruct represents a stock adjustment document (TransFlag 66, 68)
type StockAdjustmentStruct struct {
	HoldingCode        string                        `json:"holdingcode" bson:"holdingcode"`
	BranchId           string                        `json:"branchid" bson:"branchid"`
	GuidFixed          string                        `json:"guidfixed" bson:"guidfixed"`
	DocNo              string                        `json:"docno" bson:"docno"`
	Description        string                        `json:"description" bson:"description"`
	DocDateTime        time.Time                     `json:"docdatetime" bson:"docdatetime"`
	TotalAmount        float64                       `json:"totalamount" bson:"totalamount"`
	RoundAmount        float64                       `json:"roundamount"`
	PayCashAmount      float64                       `json:"paycashamount" bson:"paycashamount"`
	PayCashChange      float64                       `json:"paycashchange" bson:"paycashchange"`
	PaymentDetailRaw   string                        `json:"paymentdetailraw" bson:"paymentdetailraw"`
	SlipUrl            string                        `json:"slipurl" bson:"slipurl"`
	SaleChannelCode    string                        `json:"salechannelcode" bson:"salechannelcode"`
	DeliveryAmount     float64                       `json:"deliveryamount" bson:"deliveryamount"`
	IsCancel           bool                          `json:"iscancel" bson:"iscancel"`
	CancelReason       string                        `json:"cancelreason" bson:"cancelreason"`
	GuidPos            string                        `json:"guidpos" bson:"guidpos"`
	Branch             MongoBranchModel              `json:"branch" bson:"branch"`
	Details            []StockAdjustmentDetailStruct `json:"details" bson:"details"`
	DocCreatedDateTime time.Time                     `json:"doccreateddatetime" bson:"doccreateddatetime"`
	DocUpdatedDateTime time.Time                     `json:"docupdateddatetime" bson:"docupdateddatetime"`
	IsDelete           bool                          `json:"isdelete" bson:"isdelete"`
	TransFlag          int                           `json:"transflag" bson:"transflag"` // 66 (เพิ่ม) หรือ 68 (ลด)
}

// StockAdjustmentDetailStruct represents a stock adjustment detail line
type StockAdjustmentDetailStruct struct {
	LineNumber      int             `json:"linenumber" bson:"linenumber"`
	ItemCode        string          `json:"itemcode" bson:"itemcode"`
	ItemNames       []LanguageModel `json:"itemnames" bson:"itemnames"`
	Description     string          `json:"description" bson:"description"`
	Barcode         string          `json:"barcode" bson:"barcode"`
	UnitCode        string          `json:"unitcode" bson:"unitcode"`
	WhCode          string          `json:"whcode" bson:"whcode"`
	LocationCode    string          `json:"locationcode" bson:"locationcode"`
	Qty             float64         `json:"qty" bson:"qty"`
	Price           float64         `json:"price" bson:"price"`
	PriceExcludeVat float64         `json:"priceexcludevat" bson:"priceexcludevat"`
	DocRef          string          `json:"docref" bson:"docref"`
	SumAmount       float64         `json:"sumamount" bson:"sumamount"`
	AdjustmentType  string          `json:"adjustmenttype" bson:"adjustmenttype"` // "INCREASE" หรือ "DECREASE"
	Reason          string          `json:"reason" bson:"reason"`                 // เหตุผลในการปรับ
}

// StockBalanceStruct represents a stock balance document (TransFlag 54) - ยอดยกมา
type StockBalanceStruct struct {
	HoldingCode        string                     `json:"holdingcode" bson:"holdingcode"`
	BranchId           string                     `json:"branchid" bson:"branchid"`
	GuidFixed          string                     `json:"guidfixed" bson:"guidfixed"`
	DocNo              string                     `json:"docno" bson:"docno"`
	Description        string                     `json:"description" bson:"description"`
	DocDateTime        time.Time                  `json:"docdatetime" bson:"docdatetime"`
	TotalAmount        float64                    `json:"totalamount" bson:"totalamount"`
	RoundAmount        float64                    `json:"roundamount" bson:"roundamount"`
	PayCashAmount      float64                    `json:"paycashamount" bson:"paycashamount"`
	PayCashChange      float64                    `json:"paycashchange" bson:"paycashchange"`
	PaymentDetailRaw   string                     `json:"paymentdetailraw" bson:"paymentdetailraw"`
	SlipUrl            string                     `json:"slipurl" bson:"slipurl"`
	SaleChannelCode    string                     `json:"salechannelcode" bson:"salechannelcode"`
	DeliveryAmount     float64                    `json:"deliveryamount" bson:"deliveryamount"`
	IsCancel           bool                       `json:"iscancel" bson:"iscancel"`
	CancelReason       string                     `json:"cancelreason" bson:"cancelreason"`
	GuidPos            string                     `json:"guidpos" bson:"guidpos"`
	Branch             MongoBranchModel           `json:"branch" bson:"branch"`
	Details            []StockBalanceDetailStruct `json:"details" bson:"details"`
	DocCreatedDateTime time.Time                  `json:"doccreateddatetime" bson:"doccreateddatetime"`
	DocUpdatedDateTime time.Time                  `json:"docupdateddatetime" bson:"docupdateddatetime"`
	IsDelete           bool                       `json:"isdelete" bson:"isdelete"`
	TransFlag          int                        `json:"transflag" bson:"transflag"` // 54 (ยอดยกมา)
}

// StockBalanceDetailStruct represents a stock balance detail line
type StockBalanceDetailStruct struct {
	LineNumber      int             `json:"linenumber" bson:"linenumber"`
	ItemCode        string          `json:"itemcode" bson:"itemcode"`
	ItemNames       []LanguageModel `json:"itemnames" bson:"itemnames"`
	Description     string          `json:"description" bson:"description"`
	Barcode         string          `json:"barcode" bson:"barcode"`
	UnitCode        string          `json:"unitcode" bson:"unitcode"`
	WhCode          string          `json:"whcode" bson:"whcode"`
	LocationCode    string          `json:"locationcode" bson:"locationcode"`
	Qty             float64         `json:"qty" bson:"qty"`
	Price           float64         `json:"price" bson:"price"`
	PriceExcludeVat float64         `json:"priceexcludevat" bson:"priceexcludevat"`
	DocRef          string          `json:"docref" bson:"docref"`
	SumAmount       float64         `json:"sumamount" bson:"sumamount"`
	BalanceDate     time.Time       `json:"balancedate" bson:"balancedate"` // วันที่สำหรับยอดยกมา
	BalanceType     string          `json:"balancetype" bson:"balancetype"` // ประเภทของยอดยกมา
}
