package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type SaleInvoiceBomPricePg struct {
	HoldingCode string                `json:"holdingcode" gorm:"column:holdingcode"`
	GuidFixed   string                `json:"guidfixed" bson:"guidfixed" gorm:"column:guidfixed;primaryKey"`
	BOMGuid     string                `json:"bomguid" gorm:"column:bomguid"`
	DocNo       string                `json:"docno" gorm:"column:docno"`
	Prices      SaleInvoicePriceJSONB `json:"prices" gorm:"column:prices"`
	Barcode     string                `json:"barcode" bson:"barcode"`
	Qty         float64               `json:"qty" bson:"qty"`
	Price       float64               `json:"price" bson:"price"`
	Ratio       float64               `json:"ratio" bson:"ratio"`
}

type SaleInvoicePricePg struct {
	ItemCode string  `json:"itemcode" bson:"itemcode"`
	Barcode  string  `json:"barcode" bson:"barcode"`
	Qty      float64 `json:"qty" bson:"qty"`
	Price    float64 `json:"price" bson:"price"`
	Ratio    float64 `json:"ratio" bson:"ratio"`
}

func (s *SaleInvoiceBomPricePg) CompareTo(other *SaleInvoiceBomPricePg) bool {

	diff := cmp.Diff(s, other,
		cmpopts.IgnoreFields(SaleInvoiceBomPricePg{}, "HoldingCode", "GuidFixed"),
	)

	return diff == ""
}

type SaleInvoicePriceJSONB []SaleInvoicePricePg

// Value Marshal
func (a SaleInvoicePriceJSONB) Value() (driver.Value, error) {

	j, err := json.Marshal(a)
	return j, err
	//return json.Marshal(a)
}

// Scan Unmarshal
func (a *SaleInvoicePriceJSONB) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}
