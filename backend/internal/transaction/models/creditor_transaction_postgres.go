package models

import (
	pkgModels "smlcloudplatform/internal/models"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type CreditorTransactionPG struct {
	pkgModels.ShopIdentity      `bson:"inline"`
	pkgModels.PartitionIdentity `gorm:"embedded;"`
	GuidFixed string          `json:"guid_fixed" gorm:"column:guid_fixed"`
	DocNo string          `json:"docno" gorm:"column:docno;primaryKey"`
	DocDate time.Time       `json:"docdate" gorm:"column:docdate"`
	BranchCode string          `json:"branchcode" gorm:"column:branchcode"`
	BranchNames pkgModels.JSONB `json:"branchnames" gorm:"column:branchnames;type:jsonb"`
	CreditorCode string          `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames pkgModels.JSONB `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	InquiryType int             `json:"inquirytype" gorm:"column:inquirytype"`
	TransFlag int             `json:"transflag" gorm:"column:transflag" `
	TotalValue float64         `json:"totalvalue" gorm:"column:totalvalue"`
	TotalBeforeVat float64         `json:"totalbeforevat" gorm:"column:totalbeforevat"`
	TotalVatValue float64         `json:"totalvatvalue" gorm:"column:totalvatvalue"`
	TotalExceptVat float64         `json:"totalexceptvat" gorm:"column:totalexceptvat"`
	TotalAfterVat float64         `json:"totalaftervat" gorm:"column:totalaftervat"`
	TotalAmount float64         `json:"total_amount" gorm:"column:total_amount"`
	PaidAmount float64         `json:"paidamount" gorm:"column:paidamount"`
	BalanceAmount float64         `json:"balanceamount" gorm:"column:balanceamount"`
	Status int8            `json:"status" gorm:"column:status"`
	IsCancel bool            `json:"iscancel" gorm:"column:iscancel"`
}

func (CreditorTransactionPG) TableName() string {
	return "creditor_transaction"
}

func (m *CreditorTransactionPG) CompareTo(other *CreditorTransactionPG) bool {

	diff := cmp.Diff(m, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(CreditorTransactionPG{}, "PaidAmount", "BalanceAmount", "Status"),
	)

	return diff == ""
}
