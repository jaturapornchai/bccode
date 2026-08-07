package repositories

import "testing"

func TestProductBarcodeCompanyFiltersOwnTenantScope(t *testing.T) {
	input := map[string]interface{}{
		"holdingcode":  "UNTRUSTED-HOLDING",
		"businesscode": "UNTRUSTED-COMPANY",
		"barcode":      "885000000001",
	}

	filters := productBarcodeCompanyFilters("COMPANY-A", input)

	if _, exists := filters["holdingcode"]; exists {
		t.Fatal("request holdingcode must not be forwarded into the scoped query")
	}
	if filters["businesscode"] != "COMPANY-A" {
		t.Fatalf("expected authenticated company scope, got %#v", filters["businesscode"])
	}
	if filters["barcode"] != input["barcode"] {
		t.Fatal("non-tenant filter was not preserved")
	}
	if input["businesscode"] != "UNTRUSTED-COMPANY" || input["holdingcode"] != "UNTRUSTED-HOLDING" {
		t.Fatal("caller filters were mutated")
	}
}
