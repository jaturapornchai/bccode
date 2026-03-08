package saleinvoice_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/saleinvoice"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSaleInvoiceDebtorPhaser(t *testing.T) {
	give := SaleInvoiceTransactionStruct()
	want := models.DebtorTransactionPG{
		GuidFixed: "2TKOzSqEElEKNuIacaMHxbc4GgU",
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: "2Eh6e3pfWvXTp0yV3CyFEhKPjdI",
		},
		TransFlag:      44,
		InquiryType:    1,
		DocNo:          "a91d29f5-67af-4334-8999-8bc49ed73b4a",
		DocDate:        time.Date(2023, 7, 31, 7, 29, 28, 0, time.UTC),
		DebtorCode:     "POS001",
		TotalValue:     2000,
		TotalBeforeVat: 2,
		TotalVatValue:  51.02678028444716,
		TotalExceptVat: 1000,
		TotalAfterVat:  2,
		TotalAmount:    2000,
		PaidAmount:     0,
		BalanceAmount:  2000,
		Status:         0,
		IsCancel:       false,
	}

	debtorPhaser := saleinvoice.SaleInvoiceDebtorTransactionPhaser{}
	get, err := debtorPhaser.PhaseSingleDoc(give)

	assert.Nil(t, err)

	assert.Equal(t, want.GuidFixed, get.GuidFixed, "GuidFixed")
	assert.Equal(t, want.ShopID, get.ShopID, "ShopID")
	assert.Equal(t, want.TransFlag, get.TransFlag, "TransFlag")
	assert.Equal(t, want.InquiryType, get.InquiryType, "InquiryType")
	assert.Equal(t, want.DocNo, get.DocNo, "DocNo")
	assert.Equal(t, want.DocDate, get.DocDate, "DocDate")
	assert.Equal(t, want.DebtorCode, get.DebtorCode, "CreditorCode")
	assert.Equal(t, *get.DebtorNames[0].Name, "ลูกค้าทั่วไป", "DebtorNames")
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
