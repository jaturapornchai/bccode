package purchase_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/purchase"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPurchaseCreditorPhaser(t *testing.T) {

	give := PurchaseTransactionStruct()
	want := models.CreditorTransactionPG{
		GuidFixed: "2RYA2Yri2HRKDF5JFnKpwuGmydO",
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: "2PrIIqTWxoBXv16K310sNwfHmfY",
		},
		TransFlag:      12,
		InquiryType:    1,
		DocNo:          "PU2023062200001",
		DocDate:        time.Date(2023, 6, 22, 6, 46, 25, 0, time.UTC),
		CreditorCode:   "AP001",
		TotalValue:     50,
		TotalBeforeVat: 46.728971962616825,
		TotalVatValue:  3.2710280373831777,
		TotalExceptVat: 0,
		TotalAfterVat:  50,
		TotalAmount:    50,
		PaidAmount:     0,
		BalanceAmount:  50,
		Status:         0,
		IsCancel:       false,
	}

	creditorPhaser := purchase.PurchaseCreditorTransactionPhaser{}

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
