package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type TransactionPayment struct {
	ID            int64        `json:"id" gorm:"column:id;primary"`
	HoldingCode   string       `json:"holding_code" gorm:"column:holding_code"`
	DocNo         string       `json:"docno" gorm:"column:docno"`
	DocDate       time.Time    `json:"docdatetime" gorm:"column:docdate"`
	GuidRef       string       `json:"guid_ref" gorm:"column:guid_ref"`
	TransFlag     int16        `json:"transflag" gorm:"column:transflag"` //  44 ขาย 16 ส่งคืน 239 รับชำระ, 12 ซื้อ 48 รับคืน 19 จ่าย
	DocType       int8         `json:"doc_type" gorm:"column:doc_type"`
	InquiryType   int          `json:"inquirytype" gon:"column:inquirytype"`
	IsCancel      bool         `json:"iscancel" gorm:"column:iscancel"`
	PayCashAmount float64      `json:"paycashamount" gorm:"column:paycashamount"`
	CalcFlag      int8         `json:"calcflag" gorm:"column:calcflag"` // รับเงิน 1 , จ่ายเงิน -1
	BranchCode    string       `json:"branchcode" gorm:"column:branchcode"`
	BranchNames   models.JSONB `json:"branchnames" gorm:"column:branchnames;type:jsonb"`
	CustCode      string       `json:"custcode" gorm:"column:custcode"`
	CustNames     models.JSONB `json:"custnames" gorm:"column:custnames;type:jsonb"`

	PayCashChange    float64 `json:"paycashchange" gorm:"column:paycashchange"`
	SumQRCode        float64 `json:"sumqrcode" gorm:"column:sumqrcode"`
	SumCreditCard    float64 `json:"sumcreditcard" gorm:"column:sumcreditcard"`
	SumMoneyTransfer float64 `json:"summoneytransfer" gorm:"column:summoneytransfer"`
	SumCheque        float64 `json:"sumcheque" gorm:"column:sumcheque"`
	SumCoupon        float64 `json:"sumcoupon" gorm:"column:sumcoupon"`
	TotalAmount      float64 `json:"total_amount" gorm:"column:total_amount"`
	RoundAmount      float64 `json:"roundamount" gorm:"column:roundamount"`
	SumCredit        float64 `json:"sumcredit" gorm:"column:sumcredit"`
}

func (TransactionPayment) TableName() string {
	return "payment_transaction"
}

func (m *TransactionPayment) CompareTo(other *TransactionPayment) bool {

	diff := cmp.Diff(m, other,
		cmpopts.IgnoreFields(TransactionPayment{}, "ID"),
	)

	return diff == ""
}
