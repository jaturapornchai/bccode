package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

func TestValidateAtlasTenantRejectsMismatchedShop(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/atlas/get", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("UserInfo", msmodels.UserInfo{Username: "user@example.com", ShopID: "SHOP001"})

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
	c.Set("UserInfo", msmodels.UserInfo{Username: "user@example.com", ShopID: "SHOP001"})

	if err := validateAtlasTenant(c, "SHOP001"); err != nil {
		t.Fatalf("validateAtlasTenant returned error: %v", err)
	}
}
