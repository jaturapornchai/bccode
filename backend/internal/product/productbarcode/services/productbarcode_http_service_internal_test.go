package services

import (
	"testing"
	"time"

	common "smlcloudplatform/internal/models"
	productmodels "smlcloudplatform/internal/product/product/models"
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

func TestNormalizeCompanyBarcodeRequestKeepsCoreAndRequiresIdentity(t *testing.T) {
	request := models.ProductBarcodeRequest{ProductBarcodeBase: models.ProductBarcodeBase{
		ItemCode:     " item a ",
		Barcode:      " 885 001 ",
		ItemUnitCode: " pcs ",
		GroupCode:    "MUST-DROP",
		Qty:          99,
		ImageURI:     "barcode-main.png",
		Images:       &[]models.ProductImage{{XOrder: 1, URI: "barcode-gallery.png"}},
		Videos:       &[]models.ProductVideo{{XOrder: 1, URI: "barcode-video.mp4", PosterURI: "barcode-video.jpg"}},
		Description:  "barcode detail",
	}}

	core, err := normalizeCompanyBarcodeRequest(request)
	if err != nil {
		t.Fatalf("normalize company barcode request: %v", err)
	}
	if core.ItemCode != "ITEMA" || core.Barcode != "885001" || core.ItemUnitCode != "PCS" {
		t.Fatalf("unexpected normalized key: %q/%q/%q", core.ItemCode, core.Barcode, core.ItemUnitCode)
	}
	if core.GroupCode != "" || core.Qty != 0 {
		t.Fatalf("product or stock fields leaked into core mutation: %#v", core.ProductBarcodeBase)
	}
	if core.ImageURI != "barcode-main.png" || core.Description != "barcode detail" || core.Images == nil || len(*core.Images) != 1 || core.Videos == nil || len(*core.Videos) != 1 || (*core.Videos)[0].PosterURI != "barcode-video.jpg" {
		t.Fatalf("barcode media or description was dropped: %#v", core.ProductBarcodeBase)
	}

	request.ItemUnitCode = ""
	if _, err := normalizeCompanyBarcodeRequest(request); err == nil {
		t.Fatal("missing itemunitcode must be rejected")
	}
}

func TestNormalizeCompanyBarcodeRequestValidatesEAN13CheckDigit(t *testing.T) {
	request := models.ProductBarcodeRequest{ProductBarcodeBase: models.ProductBarcodeBase{
		ItemCode:     "ITEM-A",
		Barcode:      "4006381333931",
		ItemUnitCode: "PCS",
	}}
	if _, err := normalizeCompanyBarcodeRequest(request); err != nil {
		t.Fatalf("valid EAN-13 rejected: %v", err)
	}

	request.Barcode = "4006381333932"
	if _, err := normalizeCompanyBarcodeRequest(request); err == nil {
		t.Fatal("invalid EAN-13 check digit must be rejected")
	}

	request.Barcode = "CODE128-ABC"
	if _, err := normalizeCompanyBarcodeRequest(request); err != nil {
		t.Fatalf("non-EAN internal barcode rejected: %v", err)
	}
}

func TestValidateImmutableBarcodeIdentity(t *testing.T) {
	stored := models.ProductBarcodeDoc{}
	stored.ItemCode = "ITEM-A"
	stored.Barcode = "885001"

	same := models.ProductBarcodeRequest{ProductBarcodeBase: models.ProductBarcodeBase{ItemCode: stored.ItemCode, Barcode: stored.Barcode}}
	if err := validateImmutableBarcodeIdentity(same, stored); err != nil {
		t.Fatalf("unchanged identity rejected: %v", err)
	}

	changed := same
	changed.ItemCode = "ITEM-B"
	if err := validateImmutableBarcodeIdentity(changed, stored); err == nil {
		t.Fatal("changing itemcode must be rejected")
	}
}

func TestMinimumProductForBarcodeUsesCompanyAndFallbackName(t *testing.T) {
	now := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	barcode := models.ProductBarcodeDoc{}
	barcode.ItemCode = "ITEM-A"
	barcode.ItemUnitGuid = "UNIT-GUID"
	barcode.ItemUnitCode = "PCS"
	barcode.ItemUnitNames = &[]common.NameX{*common.NewNameXWithCodeName("th", "ชิ้น")}

	product := minimumProductForBarcode("HOLDING-A", "COMPANY-A", "admin", barcode, now)
	if product.HoldingCode != "HOLDING-A" || product.BusinessCode != "COMPANY-A" || product.Code != "ITEM-A" {
		t.Fatalf("unexpected product scope or code: %#v", product.ProductData)
	}
	if product.GuidFixed == "" || product.UnitGuid != "UNIT-GUID" || product.CreatedBy != "admin" {
		t.Fatalf("minimum product identity/audit incomplete: %#v", product)
	}
	if product.UnitCode != "PCS" || product.UnitNames == nil || product.DivideValue != 1 || product.StandValue != 1 {
		t.Fatalf("minimum product must persist one 1:1 base unit: %#v", product.ProductData)
	}
	if product.UnitConversions == nil || len(product.UnitConversions) != 0 {
		t.Fatalf("minimum product must start without additional units: %#v", product.UnitConversions)
	}
	if product.Names == nil || len(*product.Names) != 1 || (*product.Names)[0].Name == nil || *(*product.Names)[0].Name != "ITEM-A" {
		t.Fatalf("expected fallback product name from itemcode: %#v", product.Names)
	}
	if !product.CreatedAt.Equal(now) || !product.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected product timestamps: %v/%v", product.CreatedAt, product.UpdatedAt)
	}
}

func TestApplyProductUnitToBarcode(t *testing.T) {
	product := productmodels.ProductDoc{}
	product.Code = "ITEM-A"
	product.UnitCode = "PCS"
	boxNames := []common.NameX{*common.NewNameXWithCodeName("th", "กล่อง")}
	product.UnitConversions = []productmodels.ProductUnitConversion{{
		UnitCode: "BOX", UnitNames: &boxNames, DivideValue: 1, StandValue: 12,
	}}

	barcode := models.ProductBarcodeDoc{}
	barcode.ItemUnitCode = "BOX"
	barcode.ItemUnitGuid = "UNTRUSTED"
	barcode.ItemUnitNames = &[]common.NameX{*common.NewNameXWithCodeName("th", "ผิด")}
	barcode.Condition = true
	if err := applyProductUnitToBarcode(&barcode, product); err != nil {
		t.Fatal(err)
	}
	if barcode.Condition || barcode.DivideValue != 1 || barcode.StandValue != 12 || barcode.ItemUnitGuid != "" || barcode.ItemUnitNames != &boxNames {
		t.Fatalf("barcode did not derive the product unit ratio: %#v", barcode.ProductBarcodeBase)
	}

	barcode.ItemUnitCode = "CASE"
	if err := applyProductUnitToBarcode(&barcode, product); err == nil {
		t.Fatal("barcode unit outside the product must be rejected")
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

func TestCompanyBarcodeSkipsLegacyPriceHistory(t *testing.T) {
	if shouldRecordLegacyPriceHistory("COMPANY-A") {
		t.Fatal("company-owned barcode must not write to holding-only price history")
	}
	if !shouldRecordLegacyPriceHistory("") {
		t.Fatal("legacy holding-only path should keep its existing price history behavior")
	}
}
