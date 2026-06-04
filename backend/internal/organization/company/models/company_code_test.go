package models

import "testing"

func TestNormalizeCompanyCodeUppercase(t *testing.T) {
	got := NormalizeCompanyCode(" a-01 ")
	if got != "A-01" {
		t.Fatalf("NormalizeCompanyCode() = %q, want %q", got, "A-01")
	}
}
