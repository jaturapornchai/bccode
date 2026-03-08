package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type ShiftPG struct {
	ShopID                   string `json:"shopid" gorm:"column:shopid"`
	GuidFixed                string `json:"guidfixed" bson:"guidfixed" gorm:"column:guidfixed;primaryKey"`
	models.PartitionIdentity `gorm:"embedded;"`
	PosId                    string       `json:"posid" gorm:"column:posid"`
	DocNo                    string       `json:"docno" gorm:"column:docno"`
	UserCode                 string       `json:"usercode" gorm:"column:usercode"`
	Username                 models.JSONB `json:"names"  gorm:"column:names;type:jsonb" `
	DocType                  int          `json:"doctype" gorm:"column:doctype"`
	DocDate                  time.Time    `json:"docdate" gorm:"column:docdate"`
	Remark                   string       `json:"remark" gorm:"column:remark"`
	Amount                   float64      `json:"amount" gorm:"column:amount"`
	CreditCard               float64      `json:"creditcard" gorm:"column:creditcard"`
	CreditDay                float64      `json:"creditday" gorm:"column:creditday"`
	PromptPay                float64      `json:"promptpay" gorm:"column:promptpay"`
	Transfer                 float64      `json:"transfer" gorm:"column:transfer"`
	Cheque                   float64      `json:"cheque" gorm:"column:cheque"`
	Coupon                   float64      `json:"coupon" gorm:"column:coupon"`
}

func (ShiftPG) TableName() string {
	return "shift"
}

func (s *ShiftPG) CompareTo(other *ShiftPG) bool {

	diff := cmp.Diff(s, other,
		cmpopts.IgnoreFields(ShiftPG{}, "ShopID", "GuidFixed"),
	)

	return diff == ""
}
