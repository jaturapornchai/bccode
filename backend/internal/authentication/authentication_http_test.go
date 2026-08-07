package authentication

import "testing"

func TestSelectableCompanyFilterRequiresActiveCompany(t *testing.T) {
	filter := selectableCompanyFilter("holdingtest", "COMP-A")

	if filter["holdingcode"] != "holdingtest" {
		t.Fatalf("holdingcode filter mismatch: %v", filter["holdingcode"])
	}
	if filter["code"] != "COMP-A" {
		t.Fatalf("company code filter mismatch: %v", filter["code"])
	}
	if filter["isactive"] != true {
		t.Fatalf("inactive companies must not be selectable: %v", filter["isactive"])
	}
	if _, ok := filter["deletedat"]; !ok {
		t.Fatal("soft-deleted companies must not be selectable")
	}
}
