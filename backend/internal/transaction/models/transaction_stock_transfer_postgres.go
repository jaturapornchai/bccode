package models

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockTransferTransactionPG struct {
	TransactionPG `gorm:"embedded;"`
	Items         *[]StockTransferTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

type StockTransferTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
	ToWhCode            string `json:"towhcode" bson:"towhcode"`
	ToLocationCode      string `json:"tolocationcode" bson:"tolocationcode"`
	CalcFlag            int8   `json:"calcflag" gorm:"column:calcflag"`
}

func (StockTransferTransactionPG) TableName() string {
	return "stock_transfer_transaction"
}

func (StockTransferTransactionDetailPG) TableName() string {
	return "stock_transfer_transaction_detail"
}

func (j *StockTransferTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]StockTransferTransactionDetailPG
	tx.Model(&StockTransferTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

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
			tx.Delete(&StockTransferTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *StockTransferTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *StockTransferTransactionPG) CompareTo(other *StockTransferTransactionPG) bool {

	diff := cmp.Diff(s, other,
		//cmpopts.IgnoreFields(StockAdjustmentTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(StockTransferTransactionDetailPG{}, "ID"),
	)

	if diff == "" {
		return true
	}

	return false
}
