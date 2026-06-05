package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// เงินมัดจำลูกหนี้
type ARDepositPaymentTransactionPG struct {
	GeneralTransactionPG `gorm:"embedded;"`
	CreditorCode         string                                 `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames        pkgModels.JSONB                        `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items                *[]ARDepositPaymentTransactionDetailPG `json:"items" gorm:"items;foreignKey:holdingcode,docno"`
}

// รายละเอียด เงินมัดจำลูกหนี้
type ARDepositPaymentTransactionDetailPG struct {
	GeneralTransactionDetailPG `gorm:"embedded;"`
}

func (ARDepositPaymentTransactionPG) TableName() string {
	return "ardepositpaymenttransaction"
}

func (ARDepositPaymentTransactionDetailPG) TableName() string {
	return "ardepositpaymenttransactiondetail"
}

func (m *ARDepositPaymentTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]ARDepositPaymentTransactionDetailPG
	tx.Model(&ARDepositPaymentTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", m.HoldingCode, m.DocNo).Find(&details)

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
			tx.Delete(&ARDepositPaymentTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (m ARDepositPaymentTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *ARDepositPaymentTransactionPG) CompareTo(other *ARDepositPaymentTransactionPG) bool {
	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(ARDepositPaymentTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
