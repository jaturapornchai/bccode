package kafka

import (
	"testing"

	"smlcloudplatform/internal/goapi/models"
)

func TestNormalizeProductBarcodeBatchRequiresOneCompany(t *testing.T) {
	barcodes := []models.MongoProductBarcodeModel{
		{HoldingCode: " holding-a ", BusinessCode: " company-a ", ItemCode: " item-a ", Barcode: " 885 001 "},
		{HoldingCode: "holding-a", BusinessCode: "COMPANY-A", ItemCode: "ITEM-B", Barcode: "885002"},
	}

	holdingCode, businessCode, err := normalizeProductBarcodeBatch(barcodes, true)
	if err != nil {
		t.Fatalf("normalize company batch: %v", err)
	}
	if holdingCode != "holding-a" || businessCode != "COMPANY-A" {
		t.Fatalf("unexpected scope: %q/%q", holdingCode, businessCode)
	}
	if barcodes[0].ItemCode != "ITEM-A" || barcodes[0].Barcode != "885001" {
		t.Fatalf("unexpected normalized identity: %#v", barcodes[0])
	}

	barcodes[1].BusinessCode = "COMPANY-B"
	if _, _, err := normalizeProductBarcodeBatch(barcodes, true); err == nil {
		t.Fatal("mixed-company batch must be rejected")
	}
}

func TestNormalizeProductBarcodeIdentityRequiresCompanyAndItemForUpsert(t *testing.T) {
	barcode := models.MongoProductBarcodeModel{HoldingCode: "HOLDING-A", Barcode: "885001"}
	if err := normalizeProductBarcodeIdentity(&barcode, false); err == nil {
		t.Fatal("delete identity without businesscode must be rejected")
	}

	barcode.BusinessCode = "COMPANY-A"
	if err := normalizeProductBarcodeIdentity(&barcode, true); err == nil {
		t.Fatal("upsert identity without itemcode must be rejected")
	}
}
