package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// คืนเงินล่วงหน้าเจ้าหนี้
type APAdvancePaymentRefundTransactionPG struct {
	GeneralTransactionPG `gorm:"embedded;"`
	CreditorCode         string                                       `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames        pkgModels.JSONB                              `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items                *[]APAdvancePaymentRefundTransactionDetailPG `json:"items" gorm:"items;foreignKey:holding_code,docno"`
}

// รายละเอียด คืนเงินล่วงหน้าเจ้าหนี้
type APAdvancePaymentRefundTransactionDetailPG struct {
	GeneralTransactionDetailPG `gorm:"embedded;"`
}

func (APAdvancePaymentRefundTransactionPG) TableName() string {
	return "ap_advancepayment_refund_transaction"
}

func (APAdvancePaymentRefundTransactionDetailPG) TableName() string {
	return "ap_advancepayment_refund_transaction_detail"
}

func (m *APAdvancePaymentRefundTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]APAdvancePaymentRefundTransactionDetailPG
	tx.Model(&APAdvancePaymentRefundTransactionDetailPG{}).Where(" holding_code=? AND docno=?", m.HoldingCode, m.DocNo).Find(&details)

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
			tx.Delete(&APAdvancePaymentRefundTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (m APAdvancePaymentRefundTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *APAdvancePaymentRefundTransactionPG) CompareTo(other *APAdvancePaymentRefundTransactionPG) bool {
	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(APAdvancePaymentRefundTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
