package models

import "time"

type StockData struct {
	ID int64     `json:"id" gorm:"column:id;primary"`
	ShopID string    `json:"shop_id" gorm:"column:shopid"`
	DocNo string    `json:"doc_no" gorm:"column:docno"`
	DocDate time.Time `json:"docdate" gorm:"column:docdate"`
	VatType int8      `json:"vat_type" gorm:"column:vat_type"`
	TaxType int8      `json:"tax_type" gorm:"column:tax_type"`
	CalcFlag int8      `json:"calcflag" gorm:"column:calcflag"`
	TransFlag int16     `json:"transflag" gorm:"column:transflag" `
	InquiryType int       `json:"inquirytype" gorm:"column:inquirytype"`
	Barcode string    `json:"barcode" gorm:"column:barcode"`
	MainBarcodeRef string    `json:"main_barcode_ref" gorm:"column:mainbarcoderef"`
	ItemType int8      `json:"item_type" gorm:"column:item_type"`
	WhCode string    `json:"whcode" gorm:"whcode"`
	LocationCode string    `json:"locationcode" gorm:"locationcode"`
	UnitCode string    `json:"unitcode" gorm:"column:unitcode"`
	StandValue float64   `json:"standvalue" gorm:"column:standvalue"`
	DivideValue float64   `json:"dividevalue" gorm:"column:dividevalue"`
	Qty float64   `json:"qty" gorm:"column:qty"`
	CalcQty float64   `json:"calcqty" gorm:"column:calcqty"`
	Price float64   `json:"price" gorm:"column:price"`
	PriceExcludeVat float64   `json:"priceexcludevat" gorm:"column:priceexcludevat"`
	SumAmount float64   `json:"sum_amount" gorm:"column:sum_amount"`
	SumAmountExcludeVat float64   `json:"sumamountexcludevat" gorm:"column:sumamountexcludevat"`
	LineNumber int16     `json:"line_number" gorm:"column:line_number"`
	DocRef string    `json:"docref" gorm:"column:docref"`
	CostPerUnit float64   `json:"costperunit" gorm:"column:costperunit"`       // ทุนต่อหน่วย
	TotalCost float64   `json:"total_cost" gorm:"column:total_cost"`           // ต้นทุนรวม
	BalanceQty float64   `json:"balance_qty" gorm:"column:balance_qty"`         // ยอดคงเหลือ
	BalanceAmount float64   `json:"balanceamount" gorm:"column:balanceamount"`   // มูลค่าคงเหลือ
	BalanceAverage float64   `json:"balanceaverage" gorm:"column:balanceaverage"` // ต้นทุนเฉลี่ยคงเหลือ
}

type StockProcessRequest struct {
	ShopID string `json:"shopid"`
	Barcode string `json:"barcode"`
}

func (s *StockData) HasCostFromOtherDoc() bool {

	if s.TransFlag == 16 && s.DocRef != "" {
		return true
	}

	if s.TransFlag == 48 && s.DocRef != "" {
		return true
	}

	if s.TransFlag == 58 && s.DocRef != "" {
		return true
	}

	if s.TransFlag == 72 && s.CalcFlag == 1 {
		return true
	}

	return false
}
