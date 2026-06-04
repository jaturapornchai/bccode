package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SaleDebitNoteTransactionPG struct {
	TransactionPG `bson:"inline"`
	CreditorCode  string                              `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames pkgModels.JSONB                     `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items         *[]SaleDebitNoteTransactionDetailPG `json:"items" gorm:"items;foreignKey:holding_code,docno"`
}

type SaleDebitNoteTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

func (SaleDebitNoteTransactionPG) TableName() string {
	return "saledebitnote_transaction"
}

func (SaleDebitNoteTransactionDetailPG) TableName() string {
	return "saledebitnote_transaction_detail"
}

func (m *SaleDebitNoteTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]SaleDebitNoteTransactionDetailPG
	tx.Model(&SaleDebitNoteTransactionDetailPG{}).Where(" holding_code=? AND docno=?", m.HoldingCode, m.DocNo).Find(&details)

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
			tx.Delete(&SaleDebitNoteTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *SaleDebitNoteTransactionPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *SaleDebitNoteTransactionPG) CompareTo(other *SaleDebitNoteTransactionPG) bool {

	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(SaleDebitNoteTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
