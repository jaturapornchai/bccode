package saleinvoice_test

import (
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/saleinvoice"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSaleInvoiceStockPhaser(t *testing.T) {
	giveSaleInvoice := SaleInvoiceTransactionStruct()
	want := models.StockTransaction{
		GuidFixed: "2TKOzSqEElEKNuIacaMHxbc4GgU",
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: "2Eh6e3pfWvXTp0yV3CyFEhKPjdI",
		},
		TransFlag:      44,
		InquiryType:    1,
		DocNo:          "a91d29f5-67af-4334-8999-8bc49ed73b4a",
		DocDate:        time.Date(2023, 7, 31, 7, 29, 28, 0, time.UTC),
		GuidRef:        "zzzzz",
		DocRefType:     4,
		DocRefNo:       "REFNO",
		DocRefDate:     time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		VatType:        1,
		VatRate:        7,
		DiscountWord:   "100",
		TotalDiscount:  100,
		TotalValue:     2000,
		TotalBeforeVat: 2,
		TotalExceptVat: 1000,
		TotalVatValue:  51.02678028444716,
		TotalAfterVat:  2,
		TotalAmount:    2000,
		TotalCost:      0,
		Status:         0,
		IsCancel:       false,
		Details: &[]models.StockTransactionDetail{
			{
				ShopID:              "2Eh6e3pfWvXTp0yV3CyFEhKPjdI",
				DocNo:               "a91d29f5-67af-4334-8999-8bc49ed73b4a",
				Barcode:             "8850086130359",
				UnitCode:            "ซอง",
				Qty:                 5,
				Price:               6,
				PriceExcludeVat:     99,
				Discount:            "2",
				DiscountAmount:      2,
				SumAmount:           1250,
				SumAmountExcludeVat: 1245,
				StandValue:          1,
				DivideValue:         1,
				WhCode:              "POSWH000",
				LocationCode:        "POSLC000",
				CalcFlag:            -1,
				VatType:             1,
				DocRef:              "--",
			},
		},
	}

	stockPhaser := saleinvoice.SaleInvoiceTransactionStockPhaser{}
	get, err := stockPhaser.PhaseSingleDoc(giveSaleInvoice)

	assert.Nil(t, err)

	assert.Equal(t, want.GuidFixed, get.GuidFixed, "GuidFixed")
	assert.Equal(t, want.ShopID, get.ShopID, "ShopID")
	assert.Equal(t, want.TransFlag, get.TransFlag, "TransFlag")
	assert.Equal(t, want.InquiryType, get.InquiryType, "InquiryType")
	assert.Equal(t, want.DocNo, get.DocNo, "DocNo")
	assert.Equal(t, want.DocDate, get.DocDate, "DocDate")
	assert.Equal(t, want.GuidRef, get.GuidRef, "GuidRef")
	assert.Equal(t, want.DocRefType, get.DocRefType, "DocRefType")
	assert.Equal(t, want.DocRefNo, get.DocRefNo, "DocRefNo")
	assert.Equal(t, want.DocRefDate, get.DocRefDate, "DocRefDate")
	assert.Equal(t, want.VatType, get.VatType, "VatType")
	assert.Equal(t, want.DiscountWord, get.DiscountWord, "DiscountWord")
	assert.Equal(t, want.TotalDiscount, get.TotalDiscount, "TotalDiscount")
	assert.Equal(t, want.TotalValue, get.TotalValue, "TotalValue")
	assert.Equal(t, want.TotalBeforeVat, get.TotalBeforeVat, "TotalBeforeVat")
	assert.Equal(t, want.TotalExceptVat, get.TotalExceptVat, "TotalExceptVat")
	assert.Equal(t, want.TotalVatValue, get.TotalVatValue, "TotalVatValue")
	assert.Equal(t, want.TotalAfterVat, get.TotalAfterVat, "TotalAfterVat")
	assert.Equal(t, want.TotalAmount, get.TotalAmount, "TotalAmount")
	assert.Equal(t, want.TotalCost, get.TotalCost, "TotalCost")
	assert.Equal(t, want.Status, get.Status, "Status")
	assert.Equal(t, want.IsCancel, get.IsCancel, "IsCancel")

	assert.Equal(t, (*want.Details)[0].ShopID, (*get.Details)[0].ShopID, "Details[0].ShopID")
	assert.Equal(t, (*want.Details)[0].DocNo, (*get.Details)[0].DocNo, "Details[0].DocNo")
	assert.Equal(t, (*want.Details)[0].Barcode, (*get.Details)[0].Barcode, "Details[0].Barcode")
	assert.Equal(t, (*want.Details)[0].UnitCode, (*get.Details)[0].UnitCode, "Details[0].UnitCode")
	assert.Equal(t, (*want.Details)[0].Qty, (*get.Details)[0].Qty, "Details[0].Qty")
	assert.Equal(t, (*want.Details)[0].Price, (*get.Details)[0].Price, "Details[0].Price")
	assert.Equal(t, (*want.Details)[0].PriceExcludeVat, (*get.Details)[0].PriceExcludeVat, "Details[0].PriceExcludeVat")
	assert.Equal(t, (*want.Details)[0].Discount, (*get.Details)[0].Discount, "Details[0].Discount")
	assert.Equal(t, (*want.Details)[0].DiscountAmount, (*get.Details)[0].DiscountAmount, "Details[0].DiscountAmount")
	assert.Equal(t, (*want.Details)[0].SumAmount, (*get.Details)[0].SumAmount, "Details[0].SumAmount")
	assert.Equal(t, (*want.Details)[0].SumAmountExcludeVat, (*get.Details)[0].SumAmountExcludeVat, "Details[0].SumAmountExcludeVat")
	assert.Equal(t, (*want.Details)[0].StandValue, (*get.Details)[0].StandValue, "Details[0].StandValue")
	assert.Equal(t, (*want.Details)[0].DivideValue, (*get.Details)[0].DivideValue, "Details[0].DivideValue")
	assert.Equal(t, (*want.Details)[0].WhCode, (*get.Details)[0].WhCode, "Details[0].WhCode")
	assert.Equal(t, (*want.Details)[0].LocationCode, (*get.Details)[0].LocationCode, "Details[0].LocationCode")
	assert.Equal(t, (*want.Details)[0].CalcFlag, (*get.Details)[0].CalcFlag, "Details[0].CalcFlag")
	assert.Equal(t, (*want.Details)[0].VatType, (*get.Details)[0].VatType, "Details[0].VatType")
	assert.Equal(t, (*want.Details)[0].DocRef, (*get.Details)[0].DocRef, "Details[0].DocRef")
}
