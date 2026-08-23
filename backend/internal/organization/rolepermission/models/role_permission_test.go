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

func TestNormalizeRequestRejectsUnmappedCustomRole(t *testing.T) {
	req := RolePermissionRequest{
		RoleCode: "ACCOUNTANT",
		Names:    []LocalizedName{{Code: "th", Name: "บัญชี"}},
	}
	if err := NormalizeRequest(&req); err == nil {
		t.Fatal("NormalizeRequest() error = nil, want unsupported role error")
	}
}
