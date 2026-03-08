package models

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockBalanceTransactionPG struct {
	TransactionPG `gorm:"embedded;"`
	Items         *[]StockBalanceTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

type StockBalanceTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

// table name
func (t *StockBalanceTransactionPG) TableName() string {
	return "stock_balance_product_transaction"
}

func (t *StockBalanceTransactionDetailPG) TableName() string {
	return "stock_balance_product_transaction_detail"
}

func (j *StockBalanceTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]StockBalanceTransactionDetailPG
	tx.Model(&StockBalanceTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

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
			tx.Delete(&StockBalanceTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *StockBalanceTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *StockBalanceTransactionPG) CompareTo(other *StockBalanceTransactionPG) bool {

	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(StockBalanceTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(StockBalanceTransactionDetailPG{}, "ID"),
	)

	if diff == "" {
		return true
	}

	return false
}

func (s *StockBalanceTransactionDetailPG) CompareTo(other *StockBalanceTransactionDetailPG) bool {

	diff := cmp.Diff(s, other,
		cmpopts.IgnoreFields("ID"),
	)

	return diff == ""
}
