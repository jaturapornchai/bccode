package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ใบสั่งขาย
type SaleOrderPG struct {
	TransactionPG `gorm:"embedded;"`
	DebtorCode    string               `json:"creditorcode" gorm:"column:creditorcode"`
	DebtorNames   pkgModels.JSONB      `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items         *[]SaleOrderDetailPG `json:"items" gorm:"items;foreignKey:holdingcode,docno"`
}

// รายละเอียด ใบสั่งขาย
type SaleOrderDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

func (SaleOrderPG) TableName() string {
	return "saleorder_transaction"
}

func (SaleOrderDetailPG) TableName() string {
	return "saleorder_transaction_detail"
}

func (m *SaleOrderPG) BeforeUpdate(tx *gorm.DB) error {

	// find old data
	var details *[]SaleOrderDetailPG
	tx.Model(&SaleOrderDetailPG{}).Where(" holdingcode=? AND docno=?", m.HoldingCode, m.DocNo).Find(&details)

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
			tx.Delete(&SaleOrderDetailPG{}, tmp.ID)
		}
	}
	return nil
}

func (sod *SaleOrderDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})

	return nil
}

func (s *SaleOrderPG) CompareTo(other *SaleOrderPG) bool {

	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(SaleOrderDetailPG{}, "ID"),
	)

	return diff == ""
}
