package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SaleInvoiceReturnTransactionPG struct {
	TransactionPG  `gorm:"embedded;"`
	DebtorCode string          `json:"creditorcode" gorm:"column:creditorcode"`
	DebtorNames pkgModels.JSONB `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	TotalPayCash float64         `json:"totalpaycash" gorm:"column:totalpaycash;default:0"`
	TotalPayTransfer float64         `json:"totalpaytransfer" gorm:"column:totalpaytransfer;default:0"`
	TotalPayCredit float64         `json:"totalpaycredit" gorm:"column:totalpaycredit;default:0"`
	PointDiscountAmount float64         `json:"pointdiscountamount" gorm:"column:pointdiscountamount;default:0"`
	PayPointAmount float64         `json:"paypointamount" gorm:"column:paypointamount;default:0"`
	IsPOS bool            `json:"ispos" gorm:"column:ispos;default:0"`

	SaleCode string  `json:"salecode" gorm:"column:salecode"`
	SaleName string  `json:"salename" gorm:"column:salename"`
	DetailDiscountFormula string  `json:"detaildiscountformula" gorm:"column:detaildiscountformula"`
	DetailTotalAmount float64 `json:"detailtotalamount" gorm:"column:detailtotalamount;default:0"`
	TotalDiscountVatAmount float64 `json:"totaldiscountvatamount" gorm:"column:totaldiscountvatamount;default:0"`
	TotalDiscountExceptVatAmount float64 `json:"totaldiscountexceptvatamount" gorm:"column:totaldiscountexceptvatamount;default:0"`
	DetailTotalDiscount float64 `json:"detailtotaldiscount" gorm:"column:detailtotaldiscount;default:0"`
	PointsCode string  `json:"points_code" gorm:"column:points_code"`

	Items *[]SaleInvoiceReturnTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

type SaleInvoiceReturnTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
	ManufacturerGUID string          `json:"manufacturerguid" gorm:"column:manufacturerguid"`
	ManufacturerCode string          `json:"manufacturercode" gorm:"column:manufacturercode"`
	ManufacturerNames pkgModels.JSONB `json:"manufacturernames" gorm:"column:manufacturernames;type:jsonb"`
}

// table name

func (t *SaleInvoiceReturnTransactionPG) TableName() string {
	return "saleinvoice_return_transaction"
}

func (t *SaleInvoiceReturnTransactionDetailPG) TableName() string {
	return "saleinvoice_return_transaction_detail"
}

func (t SaleInvoiceReturnTransactionPG) HasDebtorEffectDoc() bool {

	hasCreditorEffectDoc := t.InquiryType == 0 || t.InquiryType == 1
	return hasCreditorEffectDoc
}

func (t SaleInvoiceReturnTransactionPG) HasStockEffectDoc() bool {
	hasStockEffectDoc := t.InquiryType == 0 || t.InquiryType == 2
	return hasStockEffectDoc
}

func (j *SaleInvoiceReturnTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]SaleInvoiceReturnTransactionDetailPG
	tx.Model(&SaleInvoiceReturnTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

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
			tx.Delete(&SaleInvoiceReturnTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *SaleInvoiceReturnTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *SaleInvoiceReturnTransactionPG) CompareTo(other *SaleInvoiceReturnTransactionPG) bool {

	diff := cmp.Diff(s, other,
		cmpopts.IgnoreFields(SaleInvoiceReturnTransactionDetailPG{}, "ID"),
	)

	if diff == "" {
		return true
	}

	return false
}
