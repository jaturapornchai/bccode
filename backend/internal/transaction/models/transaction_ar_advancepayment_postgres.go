package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// เงินล่วงหน้าลูกหนี้
type ARAdvancePaymentTransactionPG struct {
	GeneralTransactionPG `gorm:"embedded;"`
	CreditorCode string                                 `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames pkgModels.JSONB                        `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items *[]ARAdvancePaymentTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

// รายละเอียด เงินล่วงหน้าลูกหนี้
type ARAdvancePaymentTransactionDetailPG struct {
	GeneralTransactionDetailPG `gorm:"embedded;"`
}

func (ARAdvancePaymentTransactionPG) TableName() string {
	return "ar_advancepayment_transaction"
}

func (ARAdvancePaymentTransactionDetailPG) TableName() string {
	return "ar_advancepayment_transaction_detail"
}

func (m *ARAdvancePaymentTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]ARAdvancePaymentTransactionDetailPG
	tx.Model(&ARAdvancePaymentTransactionDetailPG{}).Where(" shopid=? AND docno=?", m.ShopID, m.DocNo).Find(&details)

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
			tx.Delete(&ARAdvancePaymentTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (m ARAdvancePaymentTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *ARAdvancePaymentTransactionPG) CompareTo(other *ARAdvancePaymentTransactionPG) bool {
	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(ARAdvancePaymentTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
