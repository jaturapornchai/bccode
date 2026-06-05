package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// คืนเงินล่วงหน้าลูกหนี้
type ARAdvancePaymentRefundTransactionPG struct {
	GeneralTransactionPG `gorm:"embedded;"`
	CreditorCode         string                                       `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames        pkgModels.JSONB                              `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items                *[]ARAdvancePaymentRefundTransactionDetailPG `json:"items" gorm:"items;foreignKey:holdingcode,docno"`
}

// รายละเอียด คืนเงินล่วงหน้าลูกหนี้
type ARAdvancePaymentRefundTransactionDetailPG struct {
	GeneralTransactionDetailPG `gorm:"embedded;"`
}

func (ARAdvancePaymentRefundTransactionPG) TableName() string {
	return "ar_advancepayment_refund_transaction"
}

func (ARAdvancePaymentRefundTransactionDetailPG) TableName() string {
	return "ar_advancepayment_refund_transaction_detail"
}

func (m *ARAdvancePaymentRefundTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]ARAdvancePaymentRefundTransactionDetailPG
	tx.Model(&ARAdvancePaymentRefundTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", m.HoldingCode, m.DocNo).Find(&details)

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
			tx.Delete(&ARAdvancePaymentRefundTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (m ARAdvancePaymentRefundTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *ARAdvancePaymentRefundTransactionPG) CompareTo(other *ARAdvancePaymentRefundTransactionPG) bool {
	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(ARAdvancePaymentRefundTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
