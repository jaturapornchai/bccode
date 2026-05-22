package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// คืนเงินมัดจำลูกหนี้
type ARDepositPaymentRefundTransactionPG struct {
	GeneralTransactionPG `gorm:"embedded;"`
	CreditorCode string                                       `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames pkgModels.JSONB                              `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items *[]ARDepositPaymentRefundTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

// รายละเอียด คืนเงินมัดจำลูกหนี้
type ARDepositPaymentRefundTransactionDetailPG struct {
	GeneralTransactionDetailPG `gorm:"embedded;"`
}

func (ARDepositPaymentRefundTransactionPG) TableName() string {
	return "ar_depositpayment_refund_transaction"
}

func (ARDepositPaymentRefundTransactionDetailPG) TableName() string {
	return "ar_depositpayment_refund_transaction_detail"
}

func (m *ARDepositPaymentRefundTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]ARDepositPaymentRefundTransactionDetailPG
	tx.Model(&ARDepositPaymentRefundTransactionDetailPG{}).Where(" shopid=? AND docno=?", m.ShopID, m.DocNo).Find(&details)

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
			tx.Delete(&ARDepositPaymentRefundTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (m ARDepositPaymentRefundTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *ARDepositPaymentRefundTransactionPG) CompareTo(other *ARDepositPaymentRefundTransactionPG) bool {
	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(ARDepositPaymentRefundTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
