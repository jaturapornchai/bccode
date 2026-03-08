package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseReturnTransactionPG struct {
	TransactionPG    `gorm:"embedded;"`
	CreditorCode     string                               `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames    pkgModels.JSONB                      `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	TotalPayCash     float64                              `json:"totalpaycash" gorm:"column:totalpaycash"`
	TotalPayTransfer float64                              `json:"totalpaytransfer" gorm:"column:totalpaytransfer"`
	TotalPayCredit   float64                              `json:"totalpaycredit" gorm:"column:totalpaycredit"`
	Items            *[]PurchaseReturnTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

type PurchaseReturnTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
	ManufacturerGUID    string          `json:"manufacturerguid" gorm:"column:manufacturerguid"`
	ManufacturerCode    string          `json:"manufacturercode" gorm:"column:manufacturercode"`
	ManufacturerNames   pkgModels.JSONB `json:"manufacturernames" gorm:"column:manufacturernames;type:jsonb"`
}

func (PurchaseReturnTransactionPG) TableName() string {
	return "purchase_return_transaction"
}

func (PurchaseReturnTransactionDetailPG) TableName() string {
	return "purchase_return_transaction_detail"
}

func (t PurchaseReturnTransactionPG) HasCreditorEffectDoc() bool {

	hasCreditorEffectDoc := t.InquiryType == 0 || t.InquiryType == 1
	return hasCreditorEffectDoc
}

func (t PurchaseReturnTransactionPG) HasStockEffectDoc() bool {
	hasStockEffectDoc := t.InquiryType == 0 || t.InquiryType == 2
	return hasStockEffectDoc
}

func (j *PurchaseReturnTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]PurchaseTransactionDetailPG
	tx.Model(&PurchaseTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

	// delete un use data
	for _, tmp := range *details {
		var foundUpdate bool = false
		for _, data := range *j.Items {
			if data.ID == tmp.ID {
				foundUpdate = true
			}
		}
		if foundUpdate == false {
			// mark delete
			tx.Delete(&PurchaseTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *PurchaseReturnTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *PurchaseReturnTransactionPG) CompareTo(other *PurchaseReturnTransactionPG) bool {

	diff := cmp.Diff(s, other,
		cmpopts.IgnoreFields(PurchaseReturnTransactionDetailPG{}, "ID"),
	)

	if diff == "" {
		return true
	}

	return false
}
