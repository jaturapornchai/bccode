package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SaleInvoiceTransactionPG struct {
	TransactionPG  `gorm:"embedded;"`
	DebtorCode string          `json:"creditorcode" gorm:"column:creditorcode"`
	DebtorNames pkgModels.JSONB `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	DeliveryAmount float64         `json:"deliveryamount" gorm:"column:deliveryamount"`
	TotalPayCash float64         `json:"totalpaycash" gorm:"column:totalpaycash;default:0"`
	TotalPayTransfer float64         `json:"totalpaytransfer" gorm:"column:totalpaytransfer;default:0"`
	TotalPayCredit float64         `json:"totalpaycredit" gorm:"column:totalpaycredit;default:0"`
	PointDiscountAmount float64         `json:"pointdiscountamount" gorm:"column:pointdiscountamount;default:0"`
	PayPointAmount float64         `json:"paypointamount" gorm:"column:paypointamount;default:0"`
	IsPOS bool            `json:"ispos" gorm:"column:ispos;default:0"`
	IsBom bool            `json:"isbom" gorm:"column:isbom;default:0"`
	IsDelivery bool            `json:"isdelivery" gorm:"column:isdelivery;default:0"`
	IsTransport bool            `json:"istransport" gorm:"column:istransport;default:0"`
	TransportCode string          `json:"transportcode" gorm:"column:transportcode"`
	TransportAmount float64         `json:"transportamount" gorm:"column:transportamount"`
	AlcoholAmount float64         `json:"alcoholamount" gorm:"column:alcoholamount"`
	OtherAmount float64         `json:"otheramount" gorm:"column:otheramount"`
	DrinkAmount float64         `json:"drinkamount" gorm:"column:drinkamount"`
	FoodAmount float64         `json:"foodamount" gorm:"column:foodamount"`

	OrderNumber string  `json:"ordernumber" gorm:"column:ordernumber"`
	SaleCode string  `json:"salecode" gorm:"column:salecode"`
	SaleName string  `json:"salename" gorm:"column:salename" `
	DetailDiscountFormula string  `json:"detaildiscountformula" gorm:"column:detaildiscountformula"`
	DetailTotalAmount float64 `json:"detailtotalamount" gorm:"column:detailtotalamount;default:0"`
	TotalDiscountVatAmount float64 `json:"totaldiscountvatamount" gorm:"column:totaldiscountvatamount;default:0"`
	TotalDiscountExceptVatAmount float64 `json:"totaldiscountexceptvatamount" gorm:"column:totaldiscountexceptvatamount;default:0"`
	DetailTotalDiscount float64 `json:"detailtotaldiscount" gorm:"column:detailtotaldiscount;default:0"`

	SaleChannelCode string  `json:"salechannelcode" gorm:"column:salechannelcode"`
	SaleChannelGP float64 `json:"salechannelgp" gorm:"column:salechannelgp"`
	SaleChannelGPType int8    `json:"salechannelgptype" gorm:"column:salechannelgptype"`
	TakeAway int8    `json:"takeaway" gorm:"column:takeaway"`
	PointsCode string  `json:"points_code" gorm:"column:points_code"`

	Items *[]SaleInvoiceTransactionDetailPG `json:"items" gorm:"items;foreignKey:shopid,docno"`
}

type SaleInvoiceTransactionDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
	ManufacturerGUID string          `json:"manufacturerguid" gorm:"column:manufacturerguid"`
	ManufacturerCode string          `json:"manufacturercode" gorm:"column:manufacturercode"`
	ManufacturerNames pkgModels.JSONB `json:"manufacturernames" gorm:"column:manufacturernames;type:jsonb"`
}

// tableName
func (SaleInvoiceTransactionPG) TableName() string {
	return "saleinvoice_transaction"
}

func (SaleInvoiceTransactionDetailPG) TableName() string {
	return "saleinvoice_transaction_detail"
}

func (t SaleInvoiceTransactionPG) HasCreditorEffectDoc() bool {

	hasCreditorEffectDoc := t.InquiryType == 0
	return hasCreditorEffectDoc
}

func (j *SaleInvoiceTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]SaleInvoiceTransactionDetailPG
	tx.Model(&SaleInvoiceTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

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
			tx.Delete(&SaleInvoiceTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *SaleInvoiceTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *SaleInvoiceTransactionPG) CompareTo(other *SaleInvoiceTransactionPG) bool {

	diff := cmp.Diff(s, other,
		cmpopts.IgnoreFields(SaleInvoiceTransactionDetailPG{}, "ID"),
	)

	if diff == "" {
		return true
	}

	return false
}
