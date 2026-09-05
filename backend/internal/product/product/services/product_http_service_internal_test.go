package services

import (
	"encoding/json"
	"testing"
	"time"

	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/product/models"
	barcodeModel "smlcloudplatform/internal/product/productbarcode/models"
)

func TestEnforceProductBaseUnit(t *testing.T) {
	doc := models.ProductDoc{}
	if err := enforceProductBaseUnit(&doc); err == nil {
		t.Fatal("missing base unit must be rejected")
	}

	doc.UnitCode = " pcs "
	doc.Condition = true
	doc.DivideValue = 24
	doc.StandValue = 2
	if err := enforceProductBaseUnit(&doc); err != nil {
		t.Fatal(err)
	}
	if doc.UnitCode != "PCS" || doc.Condition || doc.DivideValue != 1 || doc.StandValue != 1 {
		t.Fatalf("base unit was not normalized to 1:1: %#v", doc.ProductData)
	}
	if doc.UnitConversions == nil {
		t.Fatal("unit conversions must default to an empty array")
	}
}

func TestSyncLinkedBarcodeUnitSnapshots(t *testing.T) {
	boxNames := []common.NameX{*common.NewNameXWithCodeName("th", "กล่อง")}
	product := models.ProductDoc{}
	product.UnitCode = "PCS"
	product.UnitConversions = []models.ProductUnitConversion{{UnitCode: "BOX", UnitNames: &boxNames, DivideValue: 1, StandValue: 12}}
	barcodes := []barcodeModel.ProductBarcodeDoc{{}}
	barcodes[0].Barcode = "8850000000001"
	barcodes[0].ItemUnitCode = "BOX"
	barcodes[0].ItemUnitGuid = "UNTRUSTED"
	updatedAt := time.Date(2026, time.August, 7, 10, 0, 0, 0, time.UTC)

	if err := syncLinkedBarcodeUnitSnapshots(product, barcodes, "admin", updatedAt); err != nil {
		t.Fatal(err)
	}
	if barcodes[0].ItemUnitGuid != "" || barcodes[0].ItemUnitNames != &boxNames || barcodes[0].DivideValue != 1 || barcodes[0].StandValue != 12 {
		t.Fatalf("barcode unit snapshot did not follow Product: %#v", barcodes[0].ProductBarcodeBase)
	}
	if barcodes[0].UpdatedBy != "admin" || !barcodes[0].UpdatedAt.Equal(updatedAt) {
		t.Fatalf("barcode audit snapshot was not updated: %#v", barcodes[0].ActivityDoc)
	}

	barcodes[0].ItemUnitCode = "CASE"
	if err := syncLinkedBarcodeUnitSnapshots(product, barcodes, "admin", updatedAt); err == nil {
		t.Fatal("removing a unit used by a Barcode must fail")
	}
}

func TestEnforceProductUnitConversions(t *testing.T) {
	tests := []struct {
		name  string
		units []models.ProductUnitConversion
	}{
		{name: "blank code", units: []models.ProductUnitConversion{{DivideValue: 1, StandValue: 12}}},
		{name: "same as base", units: []models.ProductUnitConversion{{UnitCode: " pcs ", DivideValue: 1, StandValue: 12}}},
		{name: "duplicate", units: []models.ProductUnitConversion{{UnitCode: "BOX", DivideValue: 1, StandValue: 12}, {UnitCode: " box ", DivideValue: 1, StandValue: 24}}},
		{name: "zero ratio", units: []models.ProductUnitConversion{{UnitCode: "BOX", DivideValue: 0, StandValue: 12}}},
		{name: "negative ratio", units: []models.ProductUnitConversion{{UnitCode: "BOX", DivideValue: 1, StandValue: -12}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doc := models.ProductDoc{}
			doc.UnitCode = "PCS"
			doc.UnitConversions = test.units
			if err := enforceProductBaseUnit(&doc); err == nil {
				t.Fatal("invalid unit conversion must be rejected")
			}
		})
	}

	doc := models.ProductDoc{}
	doc.UnitCode = "PCS"
	doc.UnitConversions = []models.ProductUnitConversion{{UnitCode: " box ", DivideValue: 1, StandValue: 12}}
	if err := enforceProductBaseUnit(&doc); err != nil {
		t.Fatal(err)
	}
	if doc.UnitConversions[0].UnitCode != "BOX" {
		t.Fatalf("unit code was not normalized: %#v", doc.UnitConversions[0])
	}
	divide, stand, ok := doc.UnitRatio("box")
	if !ok || divide != 1 || stand != 12 {
		t.Fatalf("unexpected exact unit ratio: %d/%d, %v", stand, divide, ok)
	}
}

func TestProductProjectionMessagesPreserveContract(t *testing.T) {
	doc := models.ProductDoc{}
	doc.HoldingCode = "H"
	doc.BusinessCode = "A"
	doc.GuidFixed = "P"
	doc.Barcodes = []models.Barcodes{{Barcode: "read-only"}}
	barcode := barcodeModel.ProductBarcodeDoc{}
	barcode.HoldingCode = "H"
	barcode.BusinessCode = "A"
	barcode.Barcode = "SKU"
	messages, err := productProjectionMessages("when-product-updated", doc, []barcodeModel.ProductBarcodeDoc{barcode})
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[0].Topic != "when-product-updated" || messages[1].Topic != "when-product-barcode-bulk-updated" || messages[0].Key != messages[1].Key {
		t.Fatal("product/barcode event routing changed")
	}
	var restored models.ProductDoc
	if err := json.Unmarshal(messages[0].Payload, &restored); err != nil {
		t.Fatal(err)
	}
	if len(restored.Barcodes) != 0 || len(doc.Barcodes) != 1 {
		t.Fatal("read-only barcode data leaked or source mutated")
	}
	var restoredBarcodes []barcodeModel.ProductBarcodeDoc
	if err := json.Unmarshal(messages[1].Payload, &restoredBarcodes); err != nil {
		t.Fatal(err)
	}
	if len(restoredBarcodes) != 1 || restoredBarcodes[0].BusinessCode != "A" || restoredBarcodes[0].Barcode != "SKU" {
		t.Fatal("barcode snapshot changed")
	}
}
