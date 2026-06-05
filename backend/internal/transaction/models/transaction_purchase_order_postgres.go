package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ใบสั่งซื้อ
type PurchaseOrderTransactionPG struct {
	TransactionPG `gorm:"embedded;"`
	CreditorCode  string                              `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames pkgModels.JSONB                     `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items         *[]PurchaseOrderDetailTransactionPG `json:"items" gorm:"items;foreignKey:holdingcode,docno"`
}

// รายละเอียด ใบสั่งซื้อ
type PurchaseOrderDetailTransactionPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

func (PurchaseOrderTransactionPG) TableName() string {
	return "purchaseordertransaction"
}

func (PurchaseOrderDetailTransactionPG) TableName() string {
	return "purchaseordertransactiondetail"
}

func (s *PurchaseOrderTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]PurchaseOrderDetailTransactionPG
	tx.Model(&PurchaseOrderDetailTransactionPG{}).Where(" holdingcode=? AND docno=?", s.HoldingCode, s.DocNo).Find(&details)

	// delete un use data
	for _, tmp := range *details {
		var foundUpdate bool = false
		for _, data := range *s.Items {
			if data.ID == tmp.ID {
				foundUpdate = true
			}
		}
		if !foundUpdate {
			// mark delete
			tx.Delete(&PurchaseOrderDetailTransactionPG{}, tmp.ID)
		}
	}

	return nil
}

func (s *PurchaseOrderDetailTransactionPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *PurchaseOrderTransactionPG) CompareTo(other *PurchaseOrderTransactionPG) bool {
	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(PurchaseOrderDetailTransactionPG{}, "ID"),
	)

	return diff == ""
}
