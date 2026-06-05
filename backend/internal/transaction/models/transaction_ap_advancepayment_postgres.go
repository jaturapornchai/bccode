package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// เงินมัดจำเจ้าหนี้
type APAdvancePaymentTransactionPG struct {
	GeneralTransactionPG `gorm:"embedded;"`
	CreditorCode         string                                 `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames        pkgModels.JSONB                        `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items                *[]APAdvancePaymentTransactionDetailPG `json:"items" gorm:"items;foreignKey:holdingcode,docno"`
}

// รายละเอียด เงินมัดจำเจ้าหนี้
type APAdvancePaymentTransactionDetailPG struct {
	GeneralTransactionDetailPG `gorm:"embedded;"`
}

func (APAdvancePaymentTransactionPG) TableName() string {
	return "ap_advancepayment_transaction"
}

func (APAdvancePaymentTransactionDetailPG) TableName() string {
	return "ap_advancepayment_transaction_detail"
}

func (m *APAdvancePaymentTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]APAdvancePaymentTransactionDetailPG
	tx.Model(&APAdvancePaymentTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", m.HoldingCode, m.DocNo).Find(&details)

	// delete un use data
	for _, tmp := range *details {
		var foundUpdate bool = false
		for _, data := range *m.Items {
			if data.ID == tmp.ID {
				foundUpdate = true
			}
		}
		if !foundUpdate {
			// mark delete
			tx.Delete(&APAdvancePaymentTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (m APAdvancePaymentTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *APAdvancePaymentTransactionPG) CompareTo(other *APAdvancePaymentTransactionPG) bool {
	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(APAdvancePaymentTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
