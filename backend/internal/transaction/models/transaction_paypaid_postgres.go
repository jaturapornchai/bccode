package models

import (
	"smlcloudplatform/internal/models"
	"time"
)

type TransactionPayPaidPG struct {
	GuidFixed string `json:"guid_fixed" gorm:"column:guid_fixed"`
	models.ShopIdentity      `bson:"inline"`
	models.PartitionIdentity `gorm:"embedded;"`
	TransFlag int8         `json:"transflag" gorm:"column:transflag" `
	DocNo string       `json:"docno" gorm:"column:docno;primaryKey"`
	DocDate time.Time    `json:"docdate" gorm:"column:docdate"`
	DocType int8         `json:"doc_type" gorm:"column:doc_type"`
	BranchCode string       `json:"branchcode" gorm:"column:branchcode"`
	BranchNames models.JSONB `json:"branchnames" gorm:"column:branchnames;type:jsonb"`

	TotalPaymentAmount float64 `json:"totalpaymentamount" gorm:"column:totalpaymentamount"`
	TotalAmount float64 `json:"total_amount" gorm:"column:total_amount"`
	TotalBalance float64 `json:"totalbalance" gorm:"column:totalbalance"`
	TotalValue float64 `json:"totalvalue" gorm:"column:totalvalue"`
	IsCancel bool    `json:"iscancel" gorm:"column:iscancel"`
}
