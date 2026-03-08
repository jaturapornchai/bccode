package saleinvoicereturn_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/saleinvoicereturn"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSaleInvoiceReturnDebtorPhaser(t *testing.T) {

	give := SaleInvoiceReturnTransactionStruct()
	want := models.DebtorTransactionPG{
		GuidFixed: "2RFXUaW570MAWkgYgDduGM9WYIk",
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: "2PrIIqTWxoBXv16K310sNwfHmfY",
		},
		TransFlag:      48,
		InquiryType:    22,
		DocNo:          "ST2023061500001",
		DocDate:        time.Date(2023, 6, 15, 16, 33, 32, 0, time.UTC),
		DebtorCode:     "AR002",
		TotalValue:     280,
		TotalBeforeVat: 261.68224299065423,
		TotalVatValue:  18.317757009345794,
		TotalExceptVat: 2,
		TotalAfterVat:  280,
		TotalAmount:    280,
		PaidAmount:     0,
		BalanceAmount:  280,
		Status:         0,
		IsCancel:       false,
	}

	debtorPhaser := saleinvoicereturn.SaleInvoiceReturnDebtorTransactionPhaser{}
	get, err := debtorPhaser.PhaseSingleDoc(give)

	assert.Nil(t, err)

	assert.Equal(t, want.GuidFixed, get.GuidFixed, "GuidFixed")
	assert.Equal(t, want.ShopID, get.ShopID, "ShopID")
	assert.Equal(t, want.TransFlag, get.TransFlag, "TransFlag")
	assert.Equal(t, want.InquiryType, get.InquiryType, "InquiryType")
	assert.Equal(t, want.DocNo, get.DocNo, "DocNo")
	assert.Equal(t, want.DocDate, get.DocDate, "DocDate")
	assert.Equal(t, want.DebtorCode, get.DebtorCode, "CreditorCode")
	assert.Equal(t, *get.DebtorNames[0].Name, "นาง เมย์ ไฟแรง", "DebtorNames")
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
