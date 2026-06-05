package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseTransactionPG struct {
	TransactionPG    `gorm:"embedded;"`
	CreditorCode     string                         `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames    pkgModels.JSONB                `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	TotalPayCash     float64                        `json:"totalpaycash" gorm:"column:totalpaycash"`
	TotalPayTransfer float64                        `json:"totalpaytransfer" gorm:"column:totalpaytransfer"`
	TotalPayCredit   float64                        `json:"totalpaycredit" gorm:"column:totalpaycredit"`
	Items            *[]PurchaseTransactionDetailPG `json:"items" gorm:"items;foreignKey:holdingcode,docno"`
}

type PurchaseTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
	ManufacturerGUID    string          `json:"manufacturerguid" gorm:"column:manufacturerguid"`
	ManufacturerCode    string          `json:"manufacturercode" gorm:"column:manufacturercode"`
	ManufacturerNames   pkgModels.JSONB `json:"manufacturernames" gorm:"column:manufacturernames;type:jsonb"`
}

func (PurchaseTransactionPG) TableName() string {
	return "purchase_transaction"
}

func (PurchaseTransactionDetailPG) TableName() string {
	return "purchase_transaction_detail"
}

func (t PurchaseTransactionPG) HasCreditorEffectDoc() bool {

	hasCreditorEffectDoc := t.InquiryType == 0
	return hasCreditorEffectDoc
}

func (t PurchaseTransactionPG) HasStockEffectDoc() bool {
	return true
}

func (j *PurchaseTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]PurchaseTransactionDetailPG
	tx.Model(&PurchaseTransactionDetailPG{}).Where(" holdingcode=? AND docno=?", j.HoldingCode, j.DocNo).Find(&details)

	// delete un use data
	for _, tmp := range *details {
		var foundUpdate bool = false
		for _, data := range *j.Items {
			if data.ID == tmp.ID {
				foundUpdate = true
			}
		}
		if !foundUpdate {
			// mark delete
			tx.Delete(&PurchaseTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *PurchaseTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *PurchaseTransactionPG) CompareTo(other *PurchaseTransactionPG) bool {

	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(PurchaseTransactionDetailPG{}, "ID"),
	)

	return diff == ""
}
