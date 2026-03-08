package services

import (
	productbarcode_models "smlcloudplatform/internal/product/productbarcode/models"
	trans_models "smlcloudplatform/internal/transaction/models"
)

type SaleInvocieReturnParser struct{}

func (SaleInvocieReturnParser) ParseProductBarcode(detail trans_models.Detail, productBarcode productbarcode_models.ProductBarcodeInfo) trans_models.Detail {

	detail.ItemGuid = productBarcode.GuidFixed
	detail.ItemCode = productBarcode.ItemCode
	detail.ItemNames = productBarcode.Names
	detail.ItemType = productBarcode.ItemType
	detail.UnitCode = productBarcode.ItemUnitCode
	detail.UnitNames = productBarcode.ItemUnitNames
	detail.ManufacturerGUID = productBarcode.ManufacturerGUID
	detail.ManufacturerCode = productBarcode.ManufacturerCode
	detail.ManufacturerNames = productBarcode.ManufacturerNames
	detail.GroupCode = productBarcode.GroupCode
	detail.GroupNames = productBarcode.GroupNames

	detail.TaxType = productBarcode.TaxType
	detail.VatType = productBarcode.VatType
	detail.Discount = productBarcode.Discount

	detail.DivideValue = productBarcode.DivideValue
	detail.StandValue = productBarcode.StandValue
	detail.VatCal = productBarcode.VatCal

	return detail
}
