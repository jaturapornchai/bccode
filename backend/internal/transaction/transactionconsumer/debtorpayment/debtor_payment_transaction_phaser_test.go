package debtorpayment_test

import (
	pkgModels "smlcloudplatform/internal/models"
	models "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/debtorpayment"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func wantDataDebtorPayment() *models.DebtorPaymentTransactionPG {

	branchNames := pkgModels.JSONB{
		*pkgModels.NewNameXWithCodeName("th", "สาขาที่ 1"),
	}

	want := models.DebtorPaymentTransactionPG{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: "2IZS0jFeRXWPidSupyXN7zQIlaS",
		},
		GuidFixed:        "2UIExOige65Ekkq6O2nj7F6BEez",
		DocNo:            "EE2023062200001",
		DocDate:          time.Date(2023, 6, 22, 6, 48, 15, 0, time.UTC),
		BranchCode:       "branch01",
		BranchNames:      pkgModels.JSONB(branchNames),
		DebtorCode:       "AR001",
		TotalAmount:      70,
		TotalPayCash:     15,
		TotalPayTransfer: 25,
		TotalPayCredit:   30,
		Details: &[]models.DebtorPaymentTransactionDetailPG{
			{
				DocNo:         "EE2023062200001",
				ShopID:        "2IZS0jFeRXWPidSupyXN7zQIlaS",
				LineNumber:    0,
				BillingNo:     "PO2305041636C27E",
				BillType:      44,
				BillAmount:    80,
				BalanceAmount: 80,
				PayAmount:     50,
			},
		},
	}

	return &want
}

func TestDebtorPaymentDocumentPhaser(t *testing.T) {

	giveStr := `{
		"id": "000000000000000000000000",
		"shopid": "2IZS0jFeRXWPidSupyXN7zQIlaS",
		"guidfixed": "2UIExOige65Ekkq6O2nj7F6BEez",
		"docno": "EE2023062200001",
		"docdatetime": "2023-06-22T06:48:15.000Z",
		"doctype": 1,
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
		"transflag": 50,
		"custcode": "AR001",
		"custnames": [
			{
				"code": "th",
				"name": "ลูกค้าทั่วไป",
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
		"totalpaymentamount": 70,
		"totalamount": 70,
		"totalbalance": 348,
		"totalvalue": 348,
		"details": [
			{
				"selected": false,
				"docno": "PO2305041636C27E",
				"docdatetime": "2023-05-04T09:36:14.014Z",
				"transflag": 44,
				"value": 80,
				"balance": 80,
				"paymentamount": 50
			},
			{
				"selected": false,
				"docno": "PO2305051638BC5A",
				"docdatetime": "2023-05-05T09:38:36.036Z",
				"transflag": 44,
				"value": 60,
				"balance": 60,
				"paymentamount": 60
			},
			{
				"selected": false,
				"docno": "PO23051816522287",
				"docdatetime": "2023-05-18T09:52:02.115Z",
				"transflag": 44,
				"value": 60,
				"balance": 60,
				"paymentamount": 60
			},
			{
				"selected": false,
				"docno": "PO23052116545F75",
				"docdatetime": "1480-05-21T10:11:56Z",
				"transflag": 44,
				"value": 60,
				"balance": 60,
				"paymentamount": 60
			},
			{
				"selected": false,
				"docno": "SI2023061500002",
				"docdatetime": "2023-06-15T13:09:54.324Z",
				"transflag": 44,
				"value": 6,
				"balance": 6,
				"paymentamount": 6
			},
			{
				"selected": true,
				"docno": "SI2023061900001",
				"docdatetime": "2023-06-19T04:26:37.061Z",
				"transflag": 44,
				"value": 70,
				"balance": 70,
				"paymentamount": 70
			},
			{
				"selected": false,
				"docno": "SI2023062200001",
				"docdatetime": "2023-06-22T06:47:52.77Z",
				"transflag": 44,
				"value": 12,
				"balance": 12,
				"paymentamount": 12
			}
		],
		"paymentdetail": {
			"cashamounttext": "",
			"cashamount": 15,
			"paymentcreditcards": [
				{
					"docdatetime": "2023-06-22T09:37:42.468Z",
					"cardnumber": "456456",
					"amount": 30,
					"chargeword": "",
					"chargevalue": 0,
					"totalnetworth": 30
				}
			],
			"paymenttransfers": [
				{
					"docdatetime": "2023-06-22T09:37:47.774Z",
					"bankcode": "SCB",
					"banknames": [
						{
							"code": "th",
							"name": "ธ.ไทยพาณิชย์",
							"isauto": false,
							"isdelete": false
						}
					],
					"accountnumber": "987654321",
					"amount": 25
				}
			]
		}
	}`

	want := wantDataDebtorPayment()

	got, err := debtorpayment.DebtorPaymentTransactionPhaser{}.PhaseSingleDoc(giveStr)

	assert.Nil(t, err)

	assert.Equal(t, want.GuidFixed, got.GuidFixed, "guidfixed")
	assert.Equal(t, want.ShopID, got.ShopID, "shopid")
	assert.Equal(t, want.DocNo, got.DocNo, "docno")
	assert.Equal(t, want.DocDate, got.DocDate, "docdate")
	assert.Equal(t, want.DebtorCode, got.DebtorCode, "debtorcode")
	assert.Equal(t, want.TotalAmount, got.TotalAmount, "totalamount")
	assert.Equal(t, want.TotalPayCash, got.TotalPayCash, "totalpaycash")
	assert.Equal(t, want.TotalPayTransfer, got.TotalPayTransfer, "totalpaytransfer")
	assert.Equal(t, want.TotalPayCredit, got.TotalPayCredit, "totalpaycredit")

	assert.Equal(t, (*got.Details)[0].DocNo, (*want.Details)[0].DocNo, "item.docno")
	assert.Equal(t, (*got.Details)[0].ShopID, (*want.Details)[0].ShopID, "item.shopid")
	assert.Equal(t, (*got.Details)[0].LineNumber, (*want.Details)[0].LineNumber, "item.linenumber")
	assert.Equal(t, (*got.Details)[0].BillingNo, (*want.Details)[0].BillingNo, "item.billingno")
	assert.Equal(t, (*got.Details)[0].BillType, (*want.Details)[0].BillType, "item.billtype")
	assert.Equal(t, (*got.Details)[0].BillAmount, (*want.Details)[0].BillAmount, "item.billamount")
	assert.Equal(t, (*got.Details)[0].BalanceAmount, (*want.Details)[0].BalanceAmount, "item.balanceamount")
	assert.Equal(t, (*got.Details)[0].PayAmount, (*want.Details)[0].PayAmount, "item.payamount")

}
