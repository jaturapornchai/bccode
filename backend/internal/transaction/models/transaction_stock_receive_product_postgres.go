package models

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockReceiveProductTransactionPG struct {
	TransactionPG `gorm:"embedded;"`
	Items         *[]StockReceiveProductTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

type StockReceiveProductTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

// table name
func (t *StockReceiveProductTransactionPG) TableName() string {
	return "stock_receive_product_transaction"
}

func (t *StockReceiveProductTransactionDetailPG) TableName() string {
	return "stock_receive_product_transaction_detail"
}

func (j *StockReceiveProductTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]StockReceiveProductTransactionDetailPG
	tx.Model(&StockReceiveProductTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

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
			tx.Delete(&StockReceiveProductTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *StockReceiveProductTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *StockReceiveProductTransactionPG) CompareTo(other *StockReceiveProductTransactionPG) bool {

	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(StockReceiveProductTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(StockReceiveProductTransactionDetailPG{}, "ID"),
	)

	if diff == "" {
		return true
	}

	return false
}
