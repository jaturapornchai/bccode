package models

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ใบสืบราคา (Request for Quotation)
type RFQTransactionPG struct {
	TransactionPG   `gorm:"embedded;"`
	RefPRDocNo      string                    `json:"refprdocno" gorm:"column:refprdocno"`
	RefPRGuidFixed  string                    `json:"refprguidfixed" gorm:"column:refprguidfixed"`
	SelectedVendor  string                    `json:"selectedvendor" gorm:"column:selectedvendor"`
	SelectionReason string                    `json:"selectionreason" gorm:"column:selectionreason"`
	RefPODocNo      string                    `json:"refpodocno" gorm:"column:refpodocno"`
	RefPOGuidFixed  string                    `json:"refpoguidfixed" gorm:"column:refpoguidfixed"`
	MinVendors      int8                      `json:"minvendors" gorm:"column:minvendors"`
	ConversionStatus string                   `json:"conversionstatus" gorm:"column:conversionstatus"`
	Items           *[]RFQDetailTransactionPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

// รายละเอียด ใบสืบราคา
type RFQDetailTransactionPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

func (RFQTransactionPG) TableName() string {
	return "rfq_transaction"
}

func (RFQDetailTransactionPG) TableName() string {
	return "rfq_transaction_detail"
}

func (s *RFQTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {
	var details *[]RFQDetailTransactionPG
	tx.Model(&RFQDetailTransactionPG{}).Where(" shopid=? AND docno=?", s.ShopID, s.DocNo).Find(&details)

	for _, tmp := range *details {
		var foundUpdate bool = false
		for _, data := range *s.Items {
			if data.ID == tmp.ID {
				foundUpdate = true
			}
		}
		if !foundUpdate {
			tx.Delete(&RFQDetailTransactionPG{}, tmp.ID)
		}
	}
	return nil
}

func (s *RFQDetailTransactionPG) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *RFQTransactionPG) CompareTo(other *RFQTransactionPG) bool {
	diff := cmp.Diff(s, other,
		cmpopts.IgnoreFields(RFQDetailTransactionPG{}, "ID"),
	)
	return diff == ""
}
