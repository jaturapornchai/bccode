package models

import "time"

type StockData struct {
	ID                  int64     `json:"id" gorm:"column:id;primary"`
	HoldingCode         string    `json:"holdingcode" gorm:"column:holdingcode"`
	DocNo               string    `json:"docno" gorm:"column:docno"`
	DocDate             time.Time `json:"docdate" gorm:"column:docdate"`
	VatType             int8      `json:"vattype" gorm:"column:vattype"`
	TaxType             int8      `json:"taxtype" gorm:"column:taxtype"`
	CalcFlag            int8      `json:"calcflag" gorm:"column:calcflag"`
	TransFlag           int16     `json:"transflag" gorm:"column:transflag" `
	InquiryType         int       `json:"inquirytype" gorm:"column:inquirytype"`
	Barcode             string    `json:"barcode" gorm:"column:barcode"`
	MainBarcodeRef      string    `json:"mainbarcoderef" gorm:"column:mainbarcoderef"`
	ItemType            int8      `json:"itemtype" gorm:"column:itemtype"`
	WhCode              string    `json:"whcode" gorm:"whcode"`
	LocationCode        string    `json:"locationcode" gorm:"locationcode"`
	UnitCode            string    `json:"unitcode" gorm:"column:unitcode"`
	StandValue          float64   `json:"standvalue" gorm:"column:standvalue"`
	DivideValue         float64   `json:"dividevalue" gorm:"column:dividevalue"`
	Qty                 float64   `json:"qty" gorm:"column:qty"`
	CalcQty             float64   `json:"calcqty" gorm:"column:calcqty"`
	Price               float64   `json:"price" gorm:"column:price"`
	PriceExcludeVat     float64   `json:"priceexcludevat" gorm:"column:priceexcludevat"`
	SumAmount           float64   `json:"sumamount" gorm:"column:sumamount"`
	SumAmountExcludeVat float64   `json:"sumamountexcludevat" gorm:"column:sumamountexcludevat"`
	LineNumber          int16     `json:"linenumber" gorm:"column:linenumber"`
	DocRef              string    `json:"docref" gorm:"column:docref"`
	CostPerUnit         float64   `json:"costperunit" gorm:"column:costperunit"`       // ทุนต่อหน่วย
	TotalCost           float64   `json:"totalcost" gorm:"column:totalcost"`           // ต้นทุนรวม
	BalanceQty          float64   `json:"balanceqty" gorm:"column:balanceqty"`         // ยอดคงเหลือ
	BalanceAmount       float64   `json:"balanceamount" gorm:"column:balanceamount"`   // มูลค่าคงเหลือ
	BalanceAverage      float64   `json:"balanceaverage" gorm:"column:balanceaverage"` // ต้นทุนเฉลี่ยคงเหลือ
}

type StockProcessRequest struct {
	HoldingCode string `json:"holdingcode"`
	Barcode     string `json:"barcode"`
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
