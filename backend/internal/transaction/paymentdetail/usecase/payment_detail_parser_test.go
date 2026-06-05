package usecase_test

import (
	transmodels "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/paymentdetail/models"
	"smlcloudplatform/internal/transaction/paymentdetail/usecase"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTransactionToPaymentDetail(t *testing.T) {

	// Test case 1
	transaction1 := transmodels.TransactionMessageQueue{}
	transaction1.HoldingCode = "shop01"
	transaction1.DocNo = "DOC1"
	transaction1.TransFlag = 44
	transaction1.PaymentDetailRaw = `[
		{
			"docmode": 0,
			"transflag": 1,
			"bankcode": "ชำระด้วยบัตรเครดิต",
			"bankname": "ชำระด้วยบัตรเครดิต",
			"bookbankcode": "ชำระด้วยบัตรเครดิต",
			"cardnumber": "ชำระด้วยบัตรเครดิต",
			"approvedcode": "x1",
			"docdatetime": "2024-01-11T13:00:51.153809",
			"branchnumber": "b1",
			"bankreference": "brf1",
			"duedate": "2024-01-11T13:00:51.153804",
			"chequenumber": "chq001",
			"code": "01",
			"description": "CreditCard",
			"number": "coupon1",
			"referenceone": "r1",
			"referencetwo": "r2",
			"providercode": "pc1",
			"providername": "pn1",
			"amount": 10.75
		}
	]`

	paymentDetail1 := []models.TransactionPaymentDetail{
		{
			HoldingCode:   "shop01",
			DocNo:         "DOC1",
			DocMode:       0,
			TransFlag:     44,
			PaymentType:   1,
			BankCode:      "ชำระด้วยบัตรเครดิต",
			BankName:      "ชำระด้วยบัตรเครดิต",
			BookBankCode:  "ชำระด้วยบัตรเครดิต",
			CardNumber:    "ชำระด้วยบัตรเครดิต",
			ApprovedCode:  "x1",
			DocDateTime:   "2024-01-11T13:00:51.153809",
			BranchNumber:  "b1",
			BankReference: "brf1",
			DueDate:       "2024-01-11T13:00:51.153804",
			ChequeNumber:  "chq001",
			Code:          "01",
			Description:   "CreditCard",
			Number:        "coupon1",
			ReferenceOne:  "r1",
			ReferenceTwo:  "r2",
			ProviderCode:  "pc1",
			ProviderName:  "pn1",
			Amount:        10.75,
		},
	}

	// Test case 2
	transaction2 := transmodels.TransactionMessageQueue{}
	transaction2.DocNo = "DOC2"
	transaction2.TransFlag = 16
	transaction2.PaymentDetailRaw = ``

	paymentDetail2 := []models.TransactionPaymentDetail{}

	// Test case 3
	transaction3 := transmodels.TransactionMessageQueue{}
	transaction3.DocNo = "DOC3"
	transaction3.TransFlag = 16
	transaction3.PaymentDetailRaw = `[]`

	paymentDetail3 := []models.TransactionPaymentDetail{}

	// Test case 4
	transaction4 := transmodels.TransactionMessageQueue{}
	transaction4.DocNo = "DOC4"
	transaction4.TransFlag = 16
	transaction4.PaymentDetailRaw = "null"

	paymentDetail4 := []models.TransactionPaymentDetail{}

	// Test case 5
	transaction5 := transmodels.TransactionMessageQueue{}
	transaction5.DocNo = "DOC5"
	transaction5.TransFlag = 16
	transaction5.PaymentDetailRaw = `[]]`

	paymentDetail5 := []models.TransactionPaymentDetail{}

	tests := []struct {
		name            string
		giveHoldingCode string
		give            transmodels.TransactionMessageQueue
		expectErr       bool
		expect          []models.TransactionPaymentDetail
	}{
		{
			name:      "pass full body",
			give:      transaction1,
			expectErr: false,
			expect:    paymentDetail1,
		},
		{
			name:      "pass empty body",
			give:      transaction2,
			expectErr: false,
			expect:    paymentDetail2,
		},
		{
			name:      "pass empty array body",
			give:      transaction3,
			expectErr: false,
			expect:    paymentDetail3,
		},
		{
			name:      "pass null body",
			give:      transaction4,
			expectErr: false,
			expect:    paymentDetail4,
		},
		{
			name:      "pass invalid body",
			give:      transaction5,
			expectErr: true,
			expect:    paymentDetail5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get, err := usecase.ParseTransactionToPaymentDetail(tt.give)

			if !tt.expectErr {
				require.NoError(t, err)
				assert.Equal(t, tt.expect, get)
			}
		})
	}

}
