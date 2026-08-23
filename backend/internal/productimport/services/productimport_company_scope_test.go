package services

import (
	"testing"

	productmodels "smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/productimport/models"
)

func TestNormalizeProductImportIdentity(t *testing.T) {
	doc := models.ProductImportRaw{
		Code:     " item-a ",
		Barcode:  " barcode-a ",
		UnitCode: " ea ",
	}

	if err := normalizeProductImportIdentity(&doc); err != nil {
		t.Fatalf("normalize identity: %v", err)
	}
	if doc.Code != "ITEM-A" || doc.Barcode != "BARCODE-A" || doc.UnitCode != "EA" {
		t.Fatalf("unexpected normalized identity: %#v", doc)
	}
}

func TestNormalizeProductImportIdentityRequiresEveryKey(t *testing.T) {
	for name, doc := range map[string]models.ProductImportRaw{
		"code":    {Barcode: "BARCODE-A", UnitCode: "EA"},
		"barcode": {Code: "ITEM-A", UnitCode: "EA"},
		"unit":    {Code: "ITEM-A", Barcode: "BARCODE-A"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := normalizeProductImportIdentity(&doc); err == nil {
				t.Fatal("expected required identity error")
			}
		})
	}
}

func TestValidateProductImportUpdateIdentityRejectsRelationshipChanges(t *testing.T) {
	existing := productmodels.ProductBarcodeDoc{}
	existing.ItemCode = "ITEM-A"
	existing.ItemUnitCode = "EA"

	if err := validateProductImportUpdateIdentity(existing, models.ProductImportRaw{Code: "ITEM-B", UnitCode: "EA"}); err == nil {
		t.Fatal("expected product code change to be rejected")
	}
	if err := validateProductImportUpdateIdentity(existing, models.ProductImportRaw{Code: "ITEM-A", UnitCode: "BOX"}); err == nil {
		t.Fatal("expected unit change to be rejected")
	}
	if err := validateProductImportUpdateIdentity(existing, models.ProductImportRaw{Code: "ITEM-A", UnitCode: "EA"}); err != nil {
		t.Fatalf("expected unchanged identity to pass: %v", err)
	}
}
