package creditorpayment_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"testing"
	"time"

	"smlcloudplatform/internal/transaction/transactionconsumer/creditorpayment"

	"github.com/stretchr/testify/assert"
)

func wantDataCreditPayment() *models.CreditorPaymentTransactionPG {

	brnachNames := pkgModels.JSONB{
		*pkgModels.NewNameXWithCodeName("th", "สาขาที่ 1"),
	}

	want := models.CreditorPaymentTransactionPG{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: "2IZS0jFeRXWPidSupyXN7zQIlaS",
		},
		GuidFixed:        "2UIT0vYecL1mMA8NvvAvjnVqqwR",
		DocNo:            "DE2023080200001",
		DocDate:          time.Date(2023, 8, 2, 8, 38, 59, 0, time.UTC),
		BranchCode:       "branch01",
		BranchNames:      pkgModels.JSONB(brnachNames),
		CreditorCode:     "AP002",
		TotalAmount:      14400,
		TotalPayCash:     14400,
		TotalPayTransfer: 0,
		TotalPayCredit:   0,
		Details: &[]models.CreditorPaymentTransactionDetailPG{
			{
				DocNo:         "DE2023080200001",
				ShopID:        "2IZS0jFeRXWPidSupyXN7zQIlaS",
				LineNumber:    0,
				BillingNo:     "PU2023080200001",
				BillType:      12,
				BillAmount:    14400,
				BalanceAmount: 14400,
				PayAmount:     14400,
			},
		},
	}

	return &want
}

func TestCreditPaymentTransactionPhaser(t *testing.T) {

	giveInput := `{
		"id": "000000000000000000000000",
		"shopid": "2IZS0jFeRXWPidSupyXN7zQIlaS",
		"guidfixed": "2UIT0vYecL1mMA8NvvAvjnVqqwR",
		"docno": "DE2023080200001",
		"docdatetime": "2023-08-02T08:38:59.000Z",
		"doctype": 0,
		"branch": {
			"code": "branch01",
			"names": [
				{
					"code": "th",
					"name": "สาขาที่ 1",
					"isauto": false,
					"isdelete": false
				}
			]
		},
		"transflag": 51,
		"custcode": "AP002",
		"custnames": [
			{
				"code": "th",
				"name": "เจ้าหนี้ทั่วไป 2",
				"isauto": false,
				"isdelete": false
			},
			{
				"code": "en",
				"name": "",
				"isauto": false,
				"isdelete": false
			},
			{
				"code": "ja",
				"name": "",
				"isauto": false,
				"isdelete": false
			}
		],
		"salecode": "",
		"salename": "",
		"totalpaymentamount": 14400,
		"totalamount": 14400,
		"totalbalance": 14400,
		"totalvalue": 14400,
		"details": [
			{
				"selected": true,
				"docno": "PU2023080200001",
				"docdatetime": "2023-08-02T08:37:02.518Z",
				"transflag": 12,
				"value": 14400,
				"balance": 14400,
				"paymentamount": 14400
			}
		],
		"paymentdetail": {
			"cashamounttext": "",
			"cashamount": 14400,
			"paymentcreditcards": [],
			"paymenttransfers": []
		}
	}`

	want := wantDataCreditPayment()

	phaser := creditorpayment.CreditorPaymentTransactionPhaser{}
	got, err := phaser.PhaseSingleDoc(giveInput)

	assert.Nil(t, err)
	assert.Equal(t, want.GuidFixed, got.GuidFixed, "guidfixed")
	assert.Equal(t, want.ShopID, got.ShopID, "shopid")
	assert.Equal(t, want.DocNo, got.DocNo, "docno")
	assert.Equal(t, want.DocDate, got.DocDate, "docdate")
	assert.Equal(t, want.CreditorCode, got.CreditorCode, "creditorcode")
	assert.Equal(t, want.TotalAmount, got.TotalAmount, "totalamount")
	assert.Equal(t, want.TotalPayCash, got.TotalPayCash, "totalpaycash")
	assert.Equal(t, want.TotalPayTransfer, got.TotalPayTransfer, "totalpaytransfer")
	assert.Equal(t, want.TotalPayCredit, got.TotalPayCredit, "totalpaycredit")
	assert.Equal(t, want.BranchCode, got.BranchCode, "branchcode")
	assert.Equal(t, want.BranchNames, got.BranchNames, "branchnames")

	assert.Equal(t, (*got.Details)[0].DocNo, (*want.Details)[0].DocNo, "item.docno")
	assert.Equal(t, (*got.Details)[0].ShopID, (*want.Details)[0].ShopID, "item.shopid")
	assert.Equal(t, (*got.Details)[0].LineNumber, (*want.Details)[0].LineNumber, "item.linenumber")
	assert.Equal(t, (*got.Details)[0].BillingNo, (*want.Details)[0].BillingNo, "item.billingno")
	assert.Equal(t, (*got.Details)[0].BillType, (*want.Details)[0].BillType, "item.billtype")
	assert.Equal(t, (*got.Details)[0].BillAmount, (*want.Details)[0].BillAmount, "item.billamount")
	assert.Equal(t, (*got.Details)[0].BalanceAmount, (*want.Details)[0].BalanceAmount, "item.balanceamount")
	assert.Equal(t, (*got.Details)[0].PayAmount, (*want.Details)[0].PayAmount, "item.payamount")

}
