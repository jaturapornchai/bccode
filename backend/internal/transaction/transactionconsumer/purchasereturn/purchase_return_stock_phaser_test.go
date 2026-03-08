package purchasereturn_test

import (
	"smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/purchasereturn"
	"testing"
	"time"

	pkgModels "smlcloudplatform/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestPurchaseReturnStockPhaser(t *testing.T) {

	givePurchaseReturn := PurchaseReturnTransactionStruct()
	want := models.StockTransaction{
		GuidFixed: "2PxduUIwAoptr2OTwROegQ98Uvq",
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: "2PrIIqTWxoBXv16K310sNwfHmfY",
		},
		TransFlag:      16,
		InquiryType:    1,
		DocNo:          "PO23050616392C90",
		DocDate:        time.Date(2023, 5, 6, 9, 41, 21, 0, time.UTC),
		GuidRef:        "d4d2eddd-2f36-424f-92d2-3d0cb6c50b3f",
		DocRefType:     3,
		DocRefNo:       "PO2305051637AAD9",
		DocRefDate:     time.Date(2023, 5, 5, 9, 37, 25, 0, time.UTC),
		VatType:        2,
		VatRate:        7,
		DiscountWord:   "3%",
		TotalDiscount:  15,
		TotalValue:     20,
		TotalBeforeVat: 20,
		TotalExceptVat: 0,
		TotalVatValue:  0,
		TotalAfterVat:  20,
		TotalAmount:    20,
		TotalCost:      0,
		Status:         0,
		IsCancel:       false,
		Details: &[]models.StockTransactionDetail{
			{
				ShopID:              "2PrIIqTWxoBXv16K310sNwfHmfY",
				DocNo:               "PO23050616392C90",
				Barcode:             "BARCODE015",
				UnitCode:            "PCE",
				Qty:                 2,
				Price:               10,
				PriceExcludeVat:     10,
				Discount:            "2%",
				DiscountAmount:      5,
				SumAmount:           20,
				SumAmountExcludeVat: 20,
				StandValue:          1,
				DivideValue:         1,
				WhCode:              "00000",
				LocationCode:        "",
				CalcFlag:            -1,
				VatType:             2,
				DocRef:              "PO2305051637AAD9",
			},
		},
	}

	stockPhaser := purchasereturn.PurchaseReturnTransactionStockPhaser{}

	get, err := stockPhaser.PhaseSingleDoc(givePurchaseReturn)

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
