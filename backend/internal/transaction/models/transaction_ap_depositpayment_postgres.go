package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// เงินมัดจำเจ้าหนี้
type APDepositPaymentTransactionPG struct {
	GeneralTransactionPG `gorm:"embedded;"`
	CreditorCode string                                 `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames pkgModels.JSONB                        `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items *[]APDepositPaymentTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

// รายละเอียด เงินมัดจำเจ้าหนี้
type APDepositPaymentTransactionDetailPG struct {
	GeneralTransactionDetailPG `gorm:"embedded;"`
}

func (APDepositPaymentTransactionPG) TableName() string {
	return "ap_depositpayment_transaction"
}

func (APDepositPaymentTransactionDetailPG) TableName() string {
	return "ap_depositpayment_transaction_detail"
}

func (m *APDepositPaymentTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]APDepositPaymentTransactionDetailPG
	tx.Model(&APDepositPaymentTransactionDetailPG{}).Where(" shopid=? AND docno=?", m.ShopID, m.DocNo).Find(&details)

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
			tx.Delete(&APDepositPaymentTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (m APDepositPaymentTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *APDepositPaymentTransactionPG) CompareTo(other *APDepositPaymentTransactionPG) bool {
	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(APDepositPaymentTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
