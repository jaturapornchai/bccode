package handlers

import (
	"errors"
	"net/http"
	"testing"
)

func TestHasProductV2Permission(t *testing.T) {
	key := productV2PermissionKey
	cases := []struct {
		name    string
		records []productV2PermissionRecord
		role    string
		want    bool
	}{
		{"wildcard", []productV2PermissionRecord{{RoleCode: "USER", Permissions: []string{"*"}}}, "USER", true},
		{"exact key", []productV2PermissionRecord{{RoleCode: "USER", Permissions: []string{"product", key}}}, "USER", true},
		{"user without record", nil, "USER", false},
		{"user record without key", []productV2PermissionRecord{{RoleCode: "USER", Permissions: []string{"product"}}}, "USER", false},
		{"admin without own record", nil, "ADMIN", true},
		{"owner without own record", []productV2PermissionRecord{{RoleCode: "SET-A", Permissions: []string{"product"}}}, "OWNER", true},
		{"admin own record lacking key", []productV2PermissionRecord{{RoleCode: "ADMIN", Permissions: []string{"product"}}}, "ADMIN", false},
		{"key via permission set", []productV2PermissionRecord{{RoleCode: "USER", Permissions: []string{"product"}}, {RoleCode: "SET-A", Permissions: []string{key}}}, "USER", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasProductV2Permission(tc.records, tc.role, key); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestUpdateListing_Forbidden403(t *testing.T) {
	stubProductV2Permission(t, false, nil)
	rec, body := callProductV2(t, ProductV2ItemUpdateListingHandler, `{"itemcode":"X","__v":0,"listing":{"title":"เสื้อยืดคอกลม"}}`, &productV2TestUser)
	if rec.Code != http.StatusForbidden || body.Code != "FORBIDDEN" || body.Message != productV2MsgForbidden {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
}

func TestItemGet_Forbidden403(t *testing.T) {
	stubProductV2Permission(t, false, nil)
	rec, body := callProductV2(t, ProductV2ItemGetHandler, `{"itemcode":"X"}`, &productV2TestUser)
	if rec.Code != http.StatusForbidden || body.Code != "FORBIDDEN" || body.Message != productV2MsgForbidden {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
}

func TestItemGet_ItemCodeRequired(t *testing.T) {
	stubProductV2Permission(t, true, nil)
	rec, body := callProductV2(t, ProductV2ItemGetHandler, `{}`, &productV2TestUser)
	if rec.Code != http.StatusUnprocessableEntity || body.Fields[0].Field != "itemcode" {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
}

func TestPermission_LoaderError500(t *testing.T) {
	stubProductV2Permission(t, false, errors.New("boom"))
	rec, body := callProductV2(t, ProductV2ItemGetHandler, `{"itemcode":"X"}`, &productV2TestUser)
	if rec.Code != http.StatusInternalServerError || body.Code != "INTERNAL_ERROR" {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
}
