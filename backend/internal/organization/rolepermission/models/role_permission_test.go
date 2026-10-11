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

// Champ "สิทธิ์เฉพาะ" (report:print / report:export) are saved in the group as the literal right when ticked and as
// "!"+right when the admin unticks it (report.DeniedPermission); every other entry keeps the screen[:action] format
func TestNormalizeRequestKeepsSpecialRightDecisions(t *testing.T) {
	req := RolePermissionRequest{
		RoleCode:    "ACCOUNTING",
		Names:       []LocalizedName{{Code: "th", Name: "บัญชี"}},
		Permissions: []string{"gl-journals", "report:print", "!report:export"},
	}
	if err := NormalizeRequest(&req); err != nil {
		t.Fatalf("NormalizeRequest() error = %v", err)
	}
	if !reflect.DeepEqual(req.Permissions, []string{"!report:export", "gl-journals", "report:print"}) {
		t.Fatalf("Permissions = %#v", req.Permissions)
	}
	for _, entry := range []string{"!gl-journals", "report:delete-all", "!report:print:create", "!"} {
		bad := RolePermissionRequest{RoleCode: "ACCOUNTING", Names: []LocalizedName{{Code: "th", Name: "บัญชี"}}, Permissions: []string{entry}}
		if err := NormalizeRequest(&bad); err == nil {
			t.Fatalf("NormalizeRequest(%q) error = nil, want invalid entry", entry)
		}
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
