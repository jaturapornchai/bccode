package handlers

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
)

func TestValidateAtlasTenantRejectsMismatchedShop(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/atlas/get", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("UserInfo", msmodels.UserInfo{Username: "user@example.com", HoldingCode: "SHOP001"})

	if err := validateAtlasTenant(c, "SHOP002"); err != nil {
		t.Fatalf("validateAtlasTenant returned error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestValidateAtlasTenantAllowsMatchingShop(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/atlas/get", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("UserInfo", msmodels.UserInfo{Username: "user@example.com", HoldingCode: "SHOP001"})

	if err := validateAtlasTenant(c, "SHOP001"); err != nil {
		t.Fatalf("validateAtlasTenant returned error: %v", err)
	}
}

func TestNormalizedAtlasTenantIDPrefersHoldingCode(t *testing.T) {
	got := normalizedAtlasTenantID("HOLDING001")
	if got != "HOLDING001" {
		t.Fatalf("tenant id = %q, want HOLDING001", got)
	}
}

func TestAtlasIdentityFilterUsesGuidFixedWithLegacyReadFallback(t *testing.T) {
	got := atlasIdentityFilter("GUID001", "user@example.com", "CART001", "UID001")
	want := bson.M{
		"$or": []bson.M{
			{"guidfixed": "GUID001"},
			{"email": "GUID001", "cartid": "GUID001"},
			{"email": "user@example.com", "cartid": "CART001"},
			{"user_uid": "UID001"},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("identity filter = %#v, want %#v", got, want)
	}
}
