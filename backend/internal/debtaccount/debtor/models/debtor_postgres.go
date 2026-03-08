package models

import (
	"smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type DebtorPG struct {
	ShopID                   string `json:"shopid" gorm:"column:shopid"`
	GuidFixed                string `json:"guidfixed" bson:"guidfixed" gorm:"column:guidfixed;primaryKey"`
	models.PartitionIdentity `gorm:"embedded;"`
	Code                     string       `json:"code" gorm:"column:code;primaryKey"`
	Names                    models.JSONB `json:"names"  gorm:"column:names;type:jsonb" `
	TaxId                    string       `json:"taxid" gorm:"column:taxid"`
	PersonalType             int8         `json:"personaltype" gorm:"column:personaltype"`
	CustomerType             int          `json:"customertype" gorm:"column:customertype"`
	BranchNumber             string       `json:"branchnumber" gorm:"column:branchnumber"`
	FundCode                 string       `json:"fundcode" gorm:"column:fundcode"`
	CreditDay                int          `json:"creditday" gorm:"column:creditday"`
	PhonePrimary             string       `json:"phoneprimary" gorm:"column:phoneprimary"`
	PhoneSecondary           string       `json:"phonesecondary" gorm:"column:phonesecondary"`
	DebtorBalanceAmount      float64      `json:"debtorbalanceamount" gorm:"column:debtorbalanceamount"`
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
