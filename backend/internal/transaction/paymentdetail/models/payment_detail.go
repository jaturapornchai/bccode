package models

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type TransactionPaymentDetail struct {
	ID            int64   `json:"id" gorm:"column:id;primary"`
	HoldingCode   string  `json:"holdingcode" gorm:"column:holdingcode"`
	DocNo         string  `json:"docno" gorm:"column:docno"`
	TransFlag     int     `json:"transflag" gorm:"column:transflag"`
	PaymentType   int     `json:"paymenttype" gorm:"column:paymenttype"`
	Amount        float64 `json:"amount" gorm:"column:amount"`
	DocMode       int     `json:"docmode" gorm:"column:docmode"`
	BankCode      string  `json:"bankcode" gorm:"column:bankcode"`
	BankName      string  `json:"bankname" gorm:"column:bankname"`
	BookBankCode  string  `json:"bookbankcode" gorm:"column:bookbankcode"`
	CardNumber    string  `json:"cardnumber" gorm:"column:cardnumber"`
	ApprovedCode  string  `json:"approvedcode" gorm:"column:approvedcode"`
	DocDateTime   string  `json:"docdatetime" gorm:"column:docdatetime"`
	BranchNumber  string  `json:"branchnumber" gorm:"column:branchnumber"`
	BankReference string  `json:"bankreference" gorm:"column:bankreference"`
	DueDate       string  `json:"duedate" gorm:"column:duedate"`
	ChequeNumber  string  `json:"chequenumber" gorm:"column:chequenumber"`
	Code          string  `json:"code" gorm:"column:code"`
	Description   string  `json:"description" gorm:"column:description"`
	Number        string  `json:"number" gorm:"column:number"`
	ReferenceOne  string  `json:"referenceone" gorm:"column:referenceone"`
	ReferenceTwo  string  `json:"referencetwo" gorm:"column:referencetwo"`
	ProviderCode  string  `json:"providercode" gorm:"column:providercode"`
	ProviderName  string  `json:"providername" gorm:"column:providername"`
}

func (TransactionPaymentDetail) TableName() string {
	return "payment_transaction_detail"
}

func (m *TransactionPaymentDetail) CompareTo(other *TransactionPaymentDetail) bool {

	diff := cmp.Diff(m, other,
		cmpopts.IgnoreFields(TransactionPaymentDetail{}, "ID"),
	)

	return diff == ""
}
