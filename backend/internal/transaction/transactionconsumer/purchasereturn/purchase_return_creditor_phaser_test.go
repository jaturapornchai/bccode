package purchasereturn_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/purchasereturn"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPurchaseReturnCreditorPhaser(t *testing.T) {

	give := PurchaseReturnTransactionStruct()
	want := models.CreditorTransactionPG{
		GuidFixed: "2PxduUIwAoptr2OTwROegQ98Uvq",
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: "2PrIIqTWxoBXv16K310sNwfHmfY",
		},
		TransFlag:      16,
		InquiryType:    1,
		DocNo:          "PO23050616392C90",
		DocDate:        time.Date(2023, 5, 6, 9, 41, 21, 0, time.UTC),
		CreditorCode:   "AP001",
		TotalValue:     20,
		TotalBeforeVat: 20,
		TotalVatValue:  0,
		TotalExceptVat: 0,
		TotalAfterVat:  20,
		TotalAmount:    20,
		PaidAmount:     0,
		BalanceAmount:  20,
		Status:         0,
		IsCancel:       false,
	}

	creditorPhaser := purchasereturn.PurchaseReturnTransactionCreditorPhaser{}
	get, err := creditorPhaser.PhaseSingleDoc(give)

	assert.Nil(t, err)

	assert.Equal(t, want.GuidFixed, get.GuidFixed, "GuidFixed")
	assert.Equal(t, want.ShopID, get.ShopID, "ShopID")
	assert.Equal(t, want.TransFlag, get.TransFlag, "TransFlag")
	assert.Equal(t, want.InquiryType, get.InquiryType, "InquiryType")
	assert.Equal(t, want.DocNo, get.DocNo, "DocNo")
	assert.Equal(t, want.DocDate, get.DocDate, "DocDate")
	assert.Equal(t, want.CreditorCode, get.CreditorCode, "CreditorCode")
	assert.Equal(t, *get.CreditorNames[0].Name, "เจ้าหนี้ทั่วไป", "CreditorNames")
	assert.Equal(t, want.TotalValue, get.TotalValue, "TotalValue")
	assert.Equal(t, want.TotalBeforeVat, get.TotalBeforeVat, "TotalBeforeVat")
	assert.Equal(t, want.TotalVatValue, get.TotalVatValue, "TotalVatValue")
	assert.Equal(t, want.TotalExceptVat, get.TotalExceptVat, "TotalExceptVat")
	assert.Equal(t, want.TotalAfterVat, get.TotalAfterVat, "TotalAfterVat")
	assert.Equal(t, want.TotalAmount, get.TotalAmount, "TotalAmount")
	assert.Equal(t, want.PaidAmount, get.PaidAmount, "PaidAmount")
	assert.Equal(t, want.BalanceAmount, get.BalanceAmount, "BalanceAmount")
	assert.Equal(t, want.Status, get.Status, "Status")
	assert.Equal(t, want.IsCancel, get.IsCancel, "IsCancel")

}
