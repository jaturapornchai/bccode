package models

import (
	"smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type DebtorPG struct {
	ShopID string `json:"shopid" gorm:"column:shopid"`
	GuidFixed string `json:"guid_fixed" bson:"guid_fixed" gorm:"column:guid_fixed;primaryKey"`
	models.PartitionIdentity `gorm:"embedded;"`
	Code string       `json:"code" gorm:"column:code;primaryKey"`
	Names models.JSONB `json:"names"  gorm:"column:names;type:jsonb" `
	TaxId string       `json:"tax_id" gorm:"column:tax_id"`
	PersonalType int8         `json:"personal_type" gorm:"column:personal_type"`
	CustomerType int          `json:"customer_type" gorm:"column:customer_type"`
	BranchNumber string       `json:"branch_number" gorm:"column:branch_number"`
	FundCode string       `json:"fund_code" gorm:"column:fund_code"`
	CreditDay int          `json:"creditday" gorm:"column:creditday"`
	PhonePrimary string       `json:"phone_primary" gorm:"column:phone_primary"`
	PhoneSecondary string       `json:"phone_secondary" gorm:"column:phone_secondary"`
	DebtorBalanceAmount float64      `json:"debtor_balance_amount" gorm:"column:debtor_balance_amount"`
}

func (DebtorPG) TableName() string {
	return "debtor"
}

func (s *DebtorPG) CompareTo(other *DebtorPG) bool {

	diff := cmp.Diff(s, other,
		cmpopts.IgnoreFields(DebtorPG{}, "ShopID", "GuidFixed"),
	)

	return diff == ""
}
