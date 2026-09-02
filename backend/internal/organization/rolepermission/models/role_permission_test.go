package models

import (
	"reflect"
	"testing"
)

func TestNormalizeRequestUsesFixedMembershipRolesAndCanonicalValues(t *testing.T) {
	req := RolePermissionRequest{
		RoleCode: " user ",
		Names: []LocalizedName{
			{Code: " TH ", Name: " ผู้ใช้งาน "},
			{Code: "", Name: "ignored"},
		},
		Permissions: []string{"sale-order", " purchase ", "sale-order", ""},
	}

	if err := NormalizeRequest(&req); err != nil {
		t.Fatalf("NormalizeRequest() error = %v", err)
	}
	if req.RoleCode != "USER" {
		t.Fatalf("RoleCode = %q, want USER", req.RoleCode)
	}
	if !reflect.DeepEqual(req.Names, []LocalizedName{{Code: "th", Name: "ผู้ใช้งาน"}}) {
		t.Fatalf("Names = %#v", req.Names)
	}
	if !reflect.DeepEqual(req.Permissions, []string{"purchase", "sale-order"}) {
		t.Fatalf("Permissions = %#v", req.Permissions)
	}
}

func TestNormalizeRequestAcceptsCustomPermissionSetCode(t *testing.T) {
	req := RolePermissionRequest{
		RoleCode: " accounting-th ",
		Names:    []LocalizedName{{Code: "th", Name: "บัญชี"}},
	}
	if err := NormalizeRequest(&req); err != nil {
		t.Fatalf("NormalizeRequest() error = %v", err)
	}
	if req.RoleCode != "ACCOUNTING-TH" {
		t.Fatalf("RoleCode = %q, want ACCOUNTING-TH", req.RoleCode)
	}
}

func TestNormalizeRequestRejectsMalformedSetCode(t *testing.T) {
	for _, code := range []string{"", "A", "has space", "ไทย", "x123456789012345678901234567890"} {
		req := RolePermissionRequest{RoleCode: code, Names: []LocalizedName{{Code: "th", Name: "x"}}}
		if err := NormalizeRequest(&req); err == nil {
			t.Fatalf("NormalizeRequest(%q) error = nil, want invalid code error", code)
		}
	}
}
