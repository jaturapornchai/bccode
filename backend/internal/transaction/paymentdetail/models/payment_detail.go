package models

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type TransactionPaymentDetail struct {
	ID            int64   `json:"id" gorm:"column:id;primary"`
	ShopID        string  `json:"shopid" gorm:"column:shopid"`
	DocNo         string  `json:"docno" gorm:"column:docno"`
	TransFlag     int     `json:"trans_flag" gorm:"column:trans_flag"`
	PaymentType   int     `json:"payment_type" gorm:"column:payment_type"`
	Amount        float64 `json:"amount" gorm:"column:amount"`
	DocMode       int     `json:"doc_mode" gorm:"column:doc_mode"`
	BankCode      string  `json:"bank_code" gorm:"column:bank_code"`
	BankName      string  `json:"bank_name" gorm:"column:bank_name"`
	BookBankCode  string  `json:"book_bank_code" gorm:"column:book_bank_code"`
	CardNumber    string  `json:"card_number" gorm:"column:card_number"`
	ApprovedCode  string  `json:"approved_code" gorm:"column:approved_code"`
	DocDateTime   string  `json:"doc_date_time" gorm:"column:doc_date_time"`
	BranchNumber  string  `json:"branch_number" gorm:"column:branch_number"`
	BankReference string  `json:"bank_reference" gorm:"column:bank_reference"`
	DueDate       string  `json:"due_date" gorm:"column:due_date"`
	ChequeNumber  string  `json:"cheque_number" gorm:"column:cheque_number"`
	Code          string  `json:"code" gorm:"column:code"`
	Description   string  `json:"description" gorm:"column:description"`
	Number        string  `json:"number" gorm:"column:number"`
	ReferenceOne  string  `json:"reference_one" gorm:"column:reference_one"`
	ReferenceTwo  string  `json:"reference_two" gorm:"column:reference_two"`
	ProviderCode  string  `json:"provider_code" gorm:"column:provider_code"`
	ProviderName  string  `json:"provider_name" gorm:"column:provider_name"`
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
