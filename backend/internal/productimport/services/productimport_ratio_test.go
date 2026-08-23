package services

import "testing"

func TestParseImportedUnitRatio(t *testing.T) {
	if got, err := parseImportedUnitRatio(12, "Stand Value"); err != nil || got != 12 {
		t.Fatalf("parseImportedUnitRatio(12) = %d, %v", got, err)
	}
	for _, value := range []float64{0, -1, 1.5} {
		if _, err := parseImportedUnitRatio(value, "Stand Value"); err == nil {
			t.Fatalf("parseImportedUnitRatio(%v) accepted invalid ratio", value)
		}
	}
}
