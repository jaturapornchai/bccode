package services_test

import (
	pkg_models "smlcloudplatform/internal/models"
	productbarcode_models "smlcloudplatform/internal/product/productbarcode/models"
	trans_models "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/stockreturnproduct/services"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseProductBarcode(t *testing.T) {

	transDetail := trans_models.Detail{}
	transDetail.ItemGuid = "GUID001"
	transDetail.ItemCode = "I0001"
	transDetail.ItemNames = &[]pkg_models.NameX{
		*pkg_models.NewNameXWithCodeName("en", "Item 1"),
	}
	transDetail.ItemType = 1
	transDetail.UnitCode = "U0001"
	transDetail.UnitNames = &[]pkg_models.NameX{
		*pkg_models.NewNameXWithCodeName("en", "Unit 1"),
	}

	transDetail.ManufacturerGUID = "MGUID01"
	transDetail.ManufacturerCode = "M0001"
	transDetail.ManufacturerNames = &[]pkg_models.NameX{
		*pkg_models.NewNameXWithCodeName("en", "Manufacturer 1"),
	}
	transDetail.GroupCode = "G0001"
	transDetail.GroupNames = &[]pkg_models.NameX{
		*pkg_models.NewNameXWithCodeName("en", "Group 1"),
	}
	transDetail.TaxType = 1
	transDetail.VatType = 2
	transDetail.Discount = "DISCOUNT1"
	transDetail.DivideValue = 1
	transDetail.StandValue = 1
	transDetail.VatCal = 1

	transDetail.Qty = 10
	transDetail.Price = 100
	transDetail.DiscountAmount = 10
	transDetail.CalcFlag = 1
	transDetail.PriceExcludeVat = 90
	transDetail.InquiryType = 1

	productBarcode := productbarcode_models.ProductBarcodeInfo{}

	productBarcode.GuidFixed = "GUID001"
	productBarcode.ItemCode = "I0001"
	productBarcode.Names = &[]pkg_models.NameX{
		*pkg_models.NewNameXWithCodeName("en", "Item 1"),
	}
	productBarcode.ItemType = 1
	productBarcode.ItemUnitCode = "U0001"
	productBarcode.ItemUnitNames = &[]pkg_models.NameX{
		*pkg_models.NewNameXWithCodeName("en", "Unit 1"),
	}
	productBarcode.ManufacturerGUID = "MGUID01"
	productBarcode.ManufacturerCode = "M0001"
	productBarcode.ManufacturerNames = &[]pkg_models.NameX{
		*pkg_models.NewNameXWithCodeName("en", "Manufacturer 1"),
	}
	productBarcode.GroupCode = "G0001"
	productBarcode.GroupNames = &[]pkg_models.NameX{
		*pkg_models.NewNameXWithCodeName("en", "Group 1"),
	}
	productBarcode.TaxType = 1
	productBarcode.VatType = 2
	productBarcode.Discount = "DISCOUNT1"
	productBarcode.DivideValue = 1
	productBarcode.StandValue = 1
	productBarcode.VatCal = 1

	saleInvoiceParser := services.StockReturnProductParser{}
	transDetail = saleInvoiceParser.ParseProductBarcode(transDetail, productBarcode)

	require.Equal(t, "GUID001", transDetail.ItemGuid)
	require.Equal(t, "I0001", transDetail.ItemCode)
	require.Equal(t, "Item 1", *(*transDetail.ItemNames)[0].Name)
	require.Equal(t, int8(1), transDetail.ItemType)
	require.Equal(t, "U0001", transDetail.UnitCode)
	require.Equal(t, "Unit 1", *(*transDetail.UnitNames)[0].Name)
	require.Equal(t, "MGUID01", transDetail.ManufacturerGUID)
	require.Equal(t, "M0001", transDetail.ManufacturerCode)
	require.Equal(t, "Manufacturer 1", *(*transDetail.ManufacturerNames)[0].Name)
	require.Equal(t, "G0001", transDetail.GroupCode)
	require.Equal(t, "Group 1", *(*transDetail.GroupNames)[0].Name)
	require.Equal(t, int8(1), transDetail.TaxType)
	require.Equal(t, int8(2), transDetail.VatType)
	require.Equal(t, "DISCOUNT1", transDetail.Discount)
	require.Equal(t, float64(1), transDetail.DivideValue)
	require.Equal(t, float64(1), transDetail.StandValue)
	require.Equal(t, 1, transDetail.VatCal)
	require.Equal(t, 10.0, transDetail.Qty)
	require.Equal(t, 100.0, transDetail.Price)
	require.Equal(t, 10.0, transDetail.DiscountAmount)
	require.Equal(t, int8(1), transDetail.CalcFlag)
	require.Equal(t, 90.0, transDetail.PriceExcludeVat)
	require.Equal(t, int8(1), transDetail.InquiryType)
}
