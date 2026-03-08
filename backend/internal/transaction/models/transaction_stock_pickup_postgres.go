package models

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockPickUpTransactionPG struct {
	TransactionPG `gorm:"embedded;"`
	// CreditorCode     string                            `json:"creditorcode" gorm:"column:creditorcode"`
	// VatType          int8                              `json:"vattype" gorm:"column:vattype" `
	// VatRate          float64                           `json:"vatrate" gorm:"column:vatrate"`
	// DocRefNo         string                            `json:"docrefno" gorm:"column:docrefno"`
	// DocRefDate       time.Time                         `json:"docrefdate" gorm:"column:docrefdate"`
	// TaxDocNo         string                            `json:"taxdocno"  gorm:"column:taxdocno"`
	// TaxDocDate       time.Time                         `json:"taxdocdate" gorm:"column:taxdocdate"`
	// TotalValue       float64                           `json:"totalvalue" gorm:"column:totalvalue"`
	// DiscountWord     string                            `json:"discountword" gorm:"column:discountword"`
	// TotalDiscount    float64                           `json:"totaldiscount" gorm:"column:totaldiscount"`
	// TotalBeforeVat   float64                           `json:"totalbeforevat" gorm:"column:totalbeforevat"`
	// TotalVatValue    float64                           `json:"totalvatvalue" gorm:"column:totalvatvalue"`
	// TotalAfterVat    float64                           `json:"totalaftervat" gorm:"column:totalaftervat"`
	// TotalExceptVat   float64                           `json:"totalexceptvat" gorm:"column:totalexceptvat"`
	// TotalAmount      float64                           `json:"totalamount" gorm:"column:totalamount"`
	// TotalPayCash     float64                           `json:"totalpaycash" gorm:"column:totalpaycash"`
	// TotalPayTransfer float64                           `json:"totalpaytransfer" gorm:"column:totalpaytransfer"`
	// TotalPayCredit   float64                           `json:"totalpaycredit" gorm:"column:totalpaycredit"`
	Items *[]StockPickUpTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

type StockPickUpTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
	// Barcode             string  `json:"barcode" gorm:"column:barcode"`
	// UnitCode            string  `json:"unitcode" gorm:"column:unitcode"`
	// Qty                 float64 `json:"qty" gorm:"column:qty"`
	// Price               float64 `json:"price" gorm:"column:price"`
	// Discount            string  `json:"discount" gorm:"column:discount"`
	// DiscountAmount      float64 `json:"discountamount" gorm:"column:discountamount"`
	// SumAmount           float64 `json:"sumamount" gorm:"column:sumamount"`
}

func (StockPickUpTransactionPG) TableName() string {
	return "stock_pickup_transaction"
}

func (StockPickUpTransactionDetailPG) TableName() string {
	return "stock_pickup_transaction_detail"
}

func (j *StockPickUpTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]StockPickUpTransactionDetailPG
	tx.Model(&StockPickUpTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

	// delete un use data
	for _, tmp := range *details {
		var foundUpdate bool = false
		for _, data := range *j.Items {
			if data.ID == tmp.ID {
				foundUpdate = true
			}
		}
		if foundUpdate == false {
			// mark delete
			tx.Delete(&StockPickUpTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *StockPickUpTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *StockPickUpTransactionPG) CompareTo(other *StockPickUpTransactionPG) bool {

	diff := cmp.Diff(s, other,
		//cmpopts.IgnoreFields(StockPickUpTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(StockPickUpTransactionDetailPG{}, "ID"),
	)

	if diff == "" {
		return true
	}

	return false
}
