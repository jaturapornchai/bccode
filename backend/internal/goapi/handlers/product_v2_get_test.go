package handlers

import (
	"encoding/json"
	"testing"

	barcodemodels "smlcloudplatform/internal/product/productbarcode/models"
)

func TestBuildProductV2GetResponse(t *testing.T) {
	doc := productV2ProductDoc{}
	doc.Code = "SHIRT-001"
	doc.UnitCode = "PCS"
	models := []productV2ModelDoc{
		{Barcode: "8850000000011", ItemUnitCode: "PCS", Prices: []productV2ModelPrice{{KeyNumber: 2, Price: 150}, {KeyNumber: 1, Price: 199}},
			Listing: &barcodemodels.ProductBarcodeListing{TierIndex: []int{0, 1}, GTIN: "8850000000011", IsForSale: true}},
		{Barcode: "SHIRT-001-02", Prices: []productV2ModelPrice{{KeyNumber: 3, Price: 90}, {KeyNumber: 2, Price: 80}}, PackageWeight: 0.3},
		{Barcode: "SHIRT-001-03"},
	}
	raw, err := json.Marshal(buildProductV2GetResponse(doc, models))
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		Accounting struct {
			Code     string `json:"code"`
			Readonly bool   `json:"readonly"`
			IsActive bool   `json:"isactive"`
		} `json:"accounting"`
		Listing struct {
			Condition     string          `json:"condition"`
			Wholesale     []any           `json:"wholesale"`
			Tiers         []any           `json:"tiers"`
			Preorder      *map[string]any `json:"preorder"`
			PurchaseLimit *map[string]any `json:"purchaselimit"`
		} `json:"listing"`
		Media struct {
			Images []any `json:"images"`
		} `json:"media"`
		Models []struct {
			Barcode   string              `json:"barcode"`
			TierIndex []int               `json:"tierindex"`
			Price     string              `json:"price"`
			GTIN      string              `json:"gtin"`
			IsForSale bool                `json:"isforsale"`
			Package   *map[string]float64 `json:"package"`
		} `json:"models"`
		Version int64 `json:"__v"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%v\n%s", err, raw)
	}
	if out.Version != 0 || !out.Accounting.Readonly || !out.Accounting.IsActive || out.Accounting.Code != "SHIRT-001" {
		t.Fatalf("accounting/__v: %s", raw)
	}
	if out.Listing.Condition != "NEW" || out.Listing.Wholesale == nil || out.Listing.Tiers == nil || out.Listing.Preorder == nil || out.Listing.PurchaseLimit == nil || out.Media.Images == nil {
		t.Fatalf("nil listing must serialize with condition NEW, preorder/purchaselimit objects and empty arrays: %s", raw)
	}
	inactive := false
	doc.IsActive = &inactive
	if productV2IsActive(doc) {
		t.Fatal("isactive=false in document must be returned as-is")
	}
	if len(out.Models) != 3 {
		t.Fatalf("models: %s", raw)
	}
	m0, m1, m2 := out.Models[0], out.Models[1], out.Models[2]
	if m0.Price != "199.00" || m0.GTIN != "8850000000011" || !m0.IsForSale || len(m0.TierIndex) != 2 || m0.Package != nil {
		t.Fatalf("model0: %+v", m0)
	}
	if m1.Price != "80.00" || m1.Package == nil || (*m1.Package)["weight"] != 0.3 {
		t.Fatalf("model1: %+v", m1)
	}
	if m2.Price != "" || m2.TierIndex == nil || m2.Package != nil {
		t.Fatalf("model2: %+v", m2)
	}
}
