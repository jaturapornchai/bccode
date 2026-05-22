package models

import (
	"smlcloudplatform/internal/models"
	"time"
)

type TransactionPG struct {
	GuidFixed string `json:"guid_fixed" gorm:"column:guid_fixed"`
	models.ShopIdentity      `bson:"inline"`
	models.PartitionIdentity `gorm:"embedded;"`
	InquiryType int          `json:"inquirytype" gorm:"column:inquirytype"`
	TransFlag int16        `json:"transflag" gorm:"column:transflag" `
	DocNo string       `json:"docno" gorm:"column:docno;primaryKey"`
	DocDate time.Time    `json:"docdate" gorm:"column:docdate"`
	DocRefType int8         `json:"docreftype" gorm:"column:docreftype"`
	DocRefNo string       `json:"docrefno" gorm:"column:docrefno"`
	DeviceName string       `json:"devicename" gorm:"column:devicename"`
	GuidPos string       `json:"guidpos" gorm:"column:guidpos"`
	DocRefDate time.Time    `json:"docrefdate" gorm:"column:docrefdate"`
	BranchCode string       `json:"branchcode" gorm:"column:branchcode"`
	BranchNames models.JSONB `json:"branchnames" gorm:"column:branchnames;type:jsonb"`
	Description string       `json:"description" gorm:"column:description"`
	TaxDocNo string       `json:"taxdocno"  gorm:"column:taxdocno"`
	TaxDocDate time.Time    `json:"taxdocdate" gorm:"column:taxdocdate"`
	IsCancel bool         `json:"iscancel" gorm:"column:iscancel"`
	IsBom bool         `json:"isbom" gorm:"column:isbom"`
	Status int8         `json:"status" gorm:"column:status"`
	VatType int8         `json:"vat_type" gorm:"column:vat_type" `
	VatRate float64      `json:"vatrate" gorm:"column:vatrate"`
	TotalValue float64      `json:"totalvalue" gorm:"column:totalvalue"`
	DiscountWord string       `json:"discountword" gorm:"column:discountword"`
	DeliveryAmount float64      `json:"deliveryamount" gorm:"column:deliveryamount"`
	TotalDiscount float64      `json:"totaldiscount" gorm:"column:totaldiscount"`
	TotalBeforeVat float64      `json:"totalbeforevat" gorm:"column:totalbeforevat"`
	TotalVatValue float64      `json:"totalvatvalue" gorm:"column:totalvatvalue"`
	TotalExceptVat float64      `json:"totalexceptvat" gorm:"column:totalexceptvat"`
	TotalAfterVat float64      `json:"totalaftervat" gorm:"column:totalaftervat"`
	TotalAmount float64      `json:"total_amount" gorm:"column:total_amount"`
	GuidRef string       `json:"guid_ref" gorm:"column:guid_ref"`
	PointDiscountAmount float64      `json:"pointdiscountamount" gorm:"column:pointdiscountamount"`
	PayPointAmount float64      `json:"paypointamount" gorm:"column:paypointamount"`
	IsManualAmount bool         `json:"ismanualamount" gorm:"column:ismanualamount"`
	AlcoholAmount float64      `json:"alcoholamount" gorm:"column:alcoholamount"`
	OtherAmount float64      `json:"otheramount" gorm:"column:otheramount"`
	DrinkAmount float64      `json:"drinkamount" gorm:"column:drinkamount"`
	FoodAmount float64      `json:"foodamount" gorm:"column:foodamount"`
}

type TransactionDetailPG struct {
	ID uint   `gorm:"primarykey"`
	ShopID string `json:"shopid" gorm:"column:shopid"`
	GuidFixed string `json:"guid_fixed" gorm:"column:guid_fixed"`
	models.PartitionIdentity `gorm:"embedded;"`
	DocNo string       `json:"docno" gorm:"column:docno"`
	LineNumber int8         `json:"line_number" gorm:"column:line_number"`
	Barcode string       `json:"barcode" gorm:"column:barcode"`
	ItemNames models.JSONB `json:"itemnames" gorm:"column:itemnames;type:jsonb"`
	UnitCode string       `json:"unitcode" gorm:"column:unitcode"`
	Qty float64      `json:"qty" gorm:"column:qty"`
	Price float64      `json:"price" gorm:"column:price"`
	PriceExcludeVat float64      `json:"priceexcludevat" gorm:"column:priceexcludevat"`
	Discount string       `json:"discount" gorm:"column:discount"`
	DiscountAmount float64      `json:"discountamount" gorm:"column:discountamount"`
	SumAmount float64      `json:"sum_amount" gorm:"column:sum_amount"`
	SumAmountExcludeVat float64      `json:"sumamountexcludevat" gorm:"column:sumamountexcludevat"`
	SumAmountChoice float64      `json:"sumamountchoice" gorm:"column:sumamountchoice"`
	RefGuid string       `json:"refguid" gorm:"column:refguid"`
	WhCode string       `json:"whcode" gorm:"column:whcode"`
	WhNames models.JSONB `json:"whnames" gorm:"column:whnames;type:jsonb"`
	LocationCode string       `json:"locationcode" gorm:"column:locationcode"`
	LocationNames models.JSONB `json:"locationnames" gorm:"column:locationnames;type:jsonb"`
	VatCal int8         `json:"vatcal" gorm:"column:vatcal"`
	FoodType int8         `json:"foodtype" gorm:"column:foodtype"`
	VatType int8         `json:"vat_type" gorm:"column:vat_type"`
	TaxType int8         `json:"tax_type" gorm:"column:tax_type"`
	IsChoice int8         `json:"ischoice" gorm:"column:ischoice"`
	StandValue float64      `json:"standvalue" gorm:"column:standvalue"`
	DivideValue float64      `json:"dividevalue" gorm:"column:dividevalue"`
	ItemType int8         `json:"item_type" gorm:"column:item_type"`
	ItemGuid string       `json:"item_guid" gorm:"column:item_guid"`
	TotalValueVat float64      `json:"totalvaluevat" gorm:"column:totalvaluevat"`
	DocRef string       `json:"docref" gorm:"column:docref"`
	DocRefDateTime time.Time    `json:"docrefdatetime" gorm:"column:docrefdatetime"`
	Remark string       `json:"remark" gorm:"column:remark"`
	WhCodeDestination string       `json:"whcodedestination" gorm:"column:whcodedestination"`
	WhDestinationNames models.JSONB `json:"whcodedestinationnames" gorm:"column:whcodedestinationnames;type:jsonb"`
	LocationCodeDestination string       `json:"locationcodedestination" gorm:"column:locationcodedestination"`
	LocationDestination models.JSONB `json:"locationdestination" gorm:"column:locationdestination;type:jsonb"`
	UnitNames models.JSONB `json:"unitnames" gorm:"column:unitnames;type:jsonb"`
	GroupCode string       `json:"group_code" gorm:"column:group_code"`
	GroupNames models.JSONB `json:"group_names" gorm:"column:group_names;type:jsonb"`
	DocDate time.Time    `json:"docdate" gorm:"column:docdate"`
}

type GeneralTransactionPG struct {
	GuidFixed string `json:"guid_fixed" gorm:"column:guid_fixed"`
	models.ShopIdentity      `bson:"inline"`
	models.PartitionIdentity `gorm:"embedded;"`
	InquiryType int          `json:"inquirytype" gorm:"column:inquirytype"`
	TransFlag int16        `json:"transflag" gorm:"column:transflag" `
	DocNo string       `json:"docno" gorm:"column:docno;primaryKey"`
	DocDate time.Time    `json:"docdate" gorm:"column:docdate"`
	Description string       `json:"description" gorm:"column:description"`
	TaxDocNo string       `json:"taxdocno"  gorm:"column:taxdocno"`
	TaxDocDate time.Time    `json:"taxdocdate" gorm:"column:taxdocdate"`
	BranchCode string       `json:"branchcode" gorm:"column:branchcode"`
	BranchNames models.JSONB `json:"branchnames" gorm:"column:branchnames;type:jsonb"`
	VatType int8         `json:"vat_type" gorm:"column:vat_type" `
	VatRate float64      `json:"vatrate" gorm:"column:vatrate"`
	TotalValue float64      `json:"totalvalue" gorm:"column:totalvalue"`
	TotalBeforeVat float64      `json:"totalbeforevat" gorm:"column:totalbeforevat"`
	TotalVatValue float64      `json:"totalvatvalue" gorm:"column:totalvatvalue"`
	TotalAfterVat float64      `json:"totalaftervat" gorm:"column:totalaftervat"`
	TotalAmount float64      `json:"total_amount" gorm:"column:total_amount"`
	IsManualAmount bool         `json:"ismanualamount" gorm:"column:ismanualamount"`
}

type GeneralTransactionDetailPG struct {
	ID uint   `gorm:"primarykey"`
	ShopID string `json:"shopid" gorm:"column:shopid"`
	GuidFixed string `json:"guid_fixed" gorm:"column:guid_fixed"`
	models.PartitionIdentity `gorm:"embedded;"`
	DocNo string    `json:"docno" gorm:"column:docno"`
	DocDate time.Time `json:"docdate" gorm:"column:docdate"`
	LineNumber int8      `json:"line_number" gorm:"column:line_number"`
	Description string    `json:"description" gorm:"column:description"`
	Amount float64   `json:"amount" gorm:"column:amount"`
}
