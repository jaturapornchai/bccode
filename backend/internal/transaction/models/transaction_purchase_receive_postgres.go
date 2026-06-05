package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseReceiveTransactionPG struct {
	TransactionPG `gorm:"embedded;"`
	CreditorCode  string                                `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames pkgModels.JSONB                       `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items         *[]PurchaseReceiveTransactionDetailPG `json:"items" gorm:"items;foreignKey:holdingcode,docno"`
}

type PurchaseReceiveTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

func (PurchaseReceiveTransactionPG) TableName() string {
	return "purchasereceivetransaction"
}

func (PurchaseReceiveTransactionDetailPG) TableName() string {
	return "purchasereceivetransactiondetail"
}

func (j *PurchaseReceiveTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]PurchaseReceiveTransactionDetailPG
	tx.Model(&PurchaseReceiveTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", j.HoldingCode, j.DocNo).Find(&details)

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
			tx.Delete(&PurchaseReceiveTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *PurchaseReceiveTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *PurchaseReceiveTransactionPG) CompareTo(other *PurchaseReceiveTransactionPG) bool {
	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(PurchaseReceiveTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
