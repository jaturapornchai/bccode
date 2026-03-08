package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// คืนเงินมัดจำเจ้าหนี้
type APDepositPaymentRefundTransactionPG struct {
	GeneralTransactionPG `gorm:"embedded;"`
	CreditorCode         string                                       `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames        pkgModels.JSONB                              `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items                *[]APDepositPaymentRefundTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

// รายละเอียด คืนเงินมัดจำเจ้าหนี้
type APDepositPaymentRefundTransactionDetailPG struct {
	GeneralTransactionDetailPG `gorm:"embedded;"`
}

func (APDepositPaymentRefundTransactionPG) TableName() string {
	return "ap_depositpayment_refund_transaction"
}

func (APDepositPaymentRefundTransactionDetailPG) TableName() string {
	return "ap_depositpayment_refund_transaction_detail"
}

func (m *APDepositPaymentRefundTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]APDepositPaymentRefundTransactionDetailPG
	tx.Model(&APDepositPaymentRefundTransactionDetailPG{}).Where(" shopid=? AND docno=?", m.ShopID, m.DocNo).Find(&details)

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
			tx.Delete(&APDepositPaymentRefundTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (m APDepositPaymentRefundTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *APDepositPaymentRefundTransactionPG) CompareTo(other *APDepositPaymentRefundTransactionPG) bool {
	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(APDepositPaymentRefundTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
