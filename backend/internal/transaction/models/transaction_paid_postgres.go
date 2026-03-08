package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// การรับเงิน
type PaidTransactionPG struct {
	TransactionPayPaidPG `gorm:"embedded;"`
	CreditorCode         string                    `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames        pkgModels.JSONB           `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	TotalPayCash         float64                   `json:"totalpaycash" gorm:"column:totalpaycash"`
	TotalPayTransfer     float64                   `json:"totalpaytransfer" gorm:"column:totalpaytransfer"`
	TotalPayCredit       float64                   `json:"totalpaycredit" gorm:"column:totalpaycredit"`
	Items                *[]PayTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

// รายละเอียดการรับเงิน
type PaidTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

func (PaidTransactionPG) TableName() string {
	return "paid_transaction"
}

func (PaidTransactionDetailPG) TableName() string {
	return "paid_transaction_detail"
}

func (t PaidTransactionPG) HasStockEffectDoc() bool {
	return true
}

func (j *PaidTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]PayTransactionDetailPG
	tx.Model(&PayTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

	// delete un use data
	for _, tmp := range *details {
		var foundUpdate bool = false
		for _, data := range *j.Items {
			if data.ID == tmp.ID {
				foundUpdate = true
			}
		}
		if !foundUpdate {
			// mark delete
			tx.Delete(&PayTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *PaidTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *PaidTransactionPG) CompareTo(other *PayTransactionPG) bool {

	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PayTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(PayTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
