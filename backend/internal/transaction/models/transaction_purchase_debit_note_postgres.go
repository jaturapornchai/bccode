package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// เพิ่มหนี้ซื้อสินค้า
type PurchaseDebitNoteTransactionPG struct {
	TransactionPG `bson:"inline"`
	CreditorCode  string                                  `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames pkgModels.JSONB                         `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items         *[]PurchaseDebitNoteTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

// รายละเอียด เพิ่มหนี้ซื้อสินค้า
type PurchaseDebitNoteTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

func (PurchaseDebitNoteTransactionPG) TableName() string {
	return "purchase_debit_note_transaction"
}

func (PurchaseDebitNoteTransactionDetailPG) TableName() string {
	return "purchase_debit_note_transaction_detail"
}

func (m *PurchaseDebitNoteTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]PurchaseDebitNoteTransactionDetailPG
	tx.Model(&PurchaseDebitNoteTransactionDetailPG{}).Where(" shopid=? AND docno=?", m.ShopID, m.DocNo).Find(&details)

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
			tx.Delete(&PurchaseDebitNoteTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *PurchaseDebitNoteTransactionPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *PurchaseDebitNoteTransactionPG) CompareTo(other *PurchaseDebitNoteTransactionPG) bool {

	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(PurchaseDebitNoteTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
