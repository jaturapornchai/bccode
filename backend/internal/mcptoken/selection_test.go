package mcptoken

import "testing"

func TestCompanySelectionByMode(t *testing.T) {
	for _, mode := range []string{"readonly", "readwrite"} {
		if codes, err := CompanySelection(mode, "C01", nil); err != nil || len(codes) != 1 {
			t.Fatalf("single %s failed", mode)
		}
	}
	if codes, err := CompanySelection("readonly", "", []string{"C01", "C02"}); err != nil || len(codes) != 2 {
		t.Fatal("readonly batch denied")
	}
	for _, tc := range []struct {
		mode, single string
		many         []string
	}{
		{"readwrite", "", []string{"C01", "C02"}},
		{"readonly", "C01", []string{"C02"}},
		{"readonly", "", []string{"C01", "C01"}},
		{"readonly", "", []string{""}},
		{"readonly", "", make([]string, 51)},
		{"unknown", "C01", nil},
	} {
		if _, err := CompanySelection(tc.mode, tc.single, tc.many); err == nil {
			t.Fatal("invalid selection accepted", tc.mode)
		}
	}
}
