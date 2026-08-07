package repositories

import "testing"

func TestProductCompanyFiltersOwnsBusinessScope(t *testing.T) {
	input := map[string]interface{}{
		"businesscode": "UNTRUSTED",
		"code":         "P001",
	}

	got := productCompanyFilters("COMPANY01", input)

	if got["businesscode"] != "COMPANY01" {
		t.Fatalf("businesscode = %v, want COMPANY01", got["businesscode"])
	}
	if input["businesscode"] != "UNTRUSTED" {
		t.Fatalf("input filter was mutated: %v", input["businesscode"])
	}
}
