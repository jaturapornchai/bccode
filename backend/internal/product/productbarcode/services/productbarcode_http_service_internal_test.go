package services

import (
	"testing"

	"smlcloudplatform/internal/product/productbarcode/models"
)

func TestProductBarcodeImportKeyUsesItemCodeAndBarcode(t *testing.T) {
	svc := ProductBarcodeHttpService{}
	first := models.ProductBarcode{ProductBarcodeBase: models.ProductBarcodeBase{ItemCode: "ITEM-A", Barcode: "SHARED"}}
	second := models.ProductBarcode{ProductBarcodeBase: models.ProductBarcodeBase{ItemCode: "ITEM-B", Barcode: "SHARED"}}

	if svc.getDocIDKey(first) == svc.getDocIDKey(second) {
		t.Fatal("different itemcode values must not collapse to one import key")
	}
}

func TestNormalizeProductBarcodeImportData(t *testing.T) {
	input := []models.ProductBarcode{{
		ProductBarcodeBase: models.ProductBarcodeBase{ItemCode: " item a ", Barcode: " shared 01 "},
	}}

	result := normalizeProductBarcodeImportData(input)
	if result[0].ItemCode != "ITEMA" || result[0].Barcode != "SHARED01" {
		t.Fatalf("unexpected normalized key: %q/%q", result[0].ItemCode, result[0].Barcode)
	}
}

func TestUpdateBOMBarcodeMetadataMigratesCompositeKey(t *testing.T) {
	items := []models.BOMProductBarcode{{
		ItemCode:    "OLD",
		Barcode:     "SHARED",
		Qty:         3,
		DivideValue: 2,
		StandValue:  5,
	}}
	doc := models.ProductBarcodeDoc{}
	doc.ItemCode = "NEW"
	doc.Barcode = "SHARED"
	doc.ItemUnitCode = "PCS"

	if !updateBOMBarcodeMetadata(&items, "OLD", doc) {
		t.Fatal("expected matching BOM item to be updated")
	}
	if items[0].ItemCode != "NEW" || items[0].ItemUnitCode != "PCS" {
		t.Fatalf("metadata was not migrated: %+v", items[0])
	}
	if items[0].Qty != 3 || items[0].DivideValue != 2 || items[0].StandValue != 5 {
		t.Fatalf("calculation values changed during metadata update: %+v", items[0])
	}
}
