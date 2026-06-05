package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ใบขอซื้อ (Purchase Requisition)
type PurchaseRequisitionTransactionPG struct {
	TransactionPG    `gorm:"embedded;"`
	RequesterCode    string                                    `json:"requestercode" gorm:"column:requestercode"`
	RequesterName    string                                    `json:"requestername" gorm:"column:requestername"`
	DepartmentCode   string                                    `json:"departmentcode" gorm:"column:departmentcode"`
	DepartmentNames  pkgModels.JSONB                           `json:"departmentnames" gorm:"column:departmentnames;type:jsonb"`
	Purpose          string                                    `json:"purpose" gorm:"column:purpose"`
	BudgetCode       string                                    `json:"budgetcode" gorm:"column:budgetcode"`
	BudgetAmount     float64                                   `json:"budgetamount" gorm:"column:budgetamount"`
	Urgency          int8                                      `json:"urgency" gorm:"column:urgency"`
	ConversionStatus string                                    `json:"conversionstatus" gorm:"column:conversionstatus"`
	Items            *[]PurchaseRequisitionDetailTransactionPG `json:"items" gorm:"items;foreignKey:holdingcode,docno"`
}

// รายละเอียด ใบขอซื้อ
type PurchaseRequisitionDetailTransactionPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

func (PurchaseRequisitionTransactionPG) TableName() string {
	return "purchase_requisition_transaction"
}

func (PurchaseRequisitionDetailTransactionPG) TableName() string {
	return "purchase_requisition_transaction_detail"
}

func (s *PurchaseRequisitionTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {
	var details *[]PurchaseRequisitionDetailTransactionPG
	tx.Model(&PurchaseRequisitionDetailTransactionPG{}).Where(" holdingcode=? AND docno=?", s.HoldingCode, s.DocNo).Find(&details)

	for _, tmp := range *details {
		var foundUpdate bool = false
		for _, data := range *s.Items {
			if data.ID == tmp.ID {
				foundUpdate = true
			}
		}
		if !foundUpdate {
			tx.Delete(&PurchaseRequisitionDetailTransactionPG{}, tmp.ID)
		}
	}
	return nil
}

func (s *PurchaseRequisitionDetailTransactionPG) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *PurchaseRequisitionTransactionPG) CompareTo(other *PurchaseRequisitionTransactionPG) bool {
	diff := cmp.Diff(s, other,
		cmpopts.IgnoreFields(PurchaseRequisitionDetailTransactionPG{}, "ID"),
	)
	return diff == ""
}
