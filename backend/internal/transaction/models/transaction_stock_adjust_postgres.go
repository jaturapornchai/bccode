package models

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockAdjustmentTransactionPG struct {
	TransactionPG `gorm:"embedded;"`
	Items         *[]StockAdjustmentTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

type StockAdjustmentTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

// table name
func (StockAdjustmentTransactionPG) TableName() string {
	return "stock_adjust_transaction"
}

func (StockAdjustmentTransactionDetailPG) TableName() string {
	return "stock_adjust_transaction_detail"
}

func (j *StockAdjustmentTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]StockAdjustmentTransactionDetailPG
	tx.Model(&StockAdjustmentTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

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
			tx.Delete(&StockAdjustmentTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *StockAdjustmentTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *StockAdjustmentTransactionPG) CompareTo(other *StockAdjustmentTransactionPG) bool {

	diff := cmp.Diff(s, other,
		//cmpopts.IgnoreFields(StockAdjustmentTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(StockAdjustmentTransactionDetailPG{}, "ID"),
	)

	if diff == "" {
		return true
	}

	return false
}
