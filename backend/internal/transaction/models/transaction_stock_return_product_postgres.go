package models

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockReturnProductTransactionPG struct {
	TransactionPG `gorm:"embedded;"`
	Items         *[]StockReturnProductTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

type StockReturnProductTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

func (t *StockReturnProductTransactionPG) TableName() string {
	return "stock_return_product_transaction"
}

func (t *StockReturnProductTransactionDetailPG) TableName() string {
	return "stock_return_product_transaction_detail"
}

func (j *StockReturnProductTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]StockReturnProductTransactionDetailPG
	tx.Model(&StockReturnProductTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

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
			tx.Delete(&StockReturnProductTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *StockReturnProductTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *StockReturnProductTransactionPG) CompareTo(other *StockReturnProductTransactionPG) bool {

	diff := cmp.Diff(s, other,
		//cmpopts.IgnoreFields(StockReturnProductTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(StockReturnProductTransactionDetailPG{}, "ID"),
	)

	if diff == "" {
		return true
	}

	return false
}
