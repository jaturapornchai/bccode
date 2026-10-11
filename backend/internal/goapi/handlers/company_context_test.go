package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

func TestAuthenticatedCompanyContext(t *testing.T) {
	tests := []struct {
		name              string
		user              msmodels.UserInfo
		requested         string
		requestedBusiness string
		holding           string
		business          string
		status            int
	}{
		{name: "requires authentication", status: http.StatusUnauthorized},
		{name: "rejects cross holding", user: msmodels.UserInfo{HoldingCode: "H1", BusinessCode: "C1"}, requested: "H2", status: http.StatusForbidden},
		{name: "requires company", user: msmodels.UserInfo{HoldingCode: "H1"}, requested: "H1", status: http.StatusConflict},
		{name: "rejects cross company", user: msmodels.UserInfo{HoldingCode: "H1", BusinessCode: "C1"}, requested: "H1", requestedBusiness: "C2", status: http.StatusForbidden},
		{name: "normalizes server identity", user: msmodels.UserInfo{HoldingCode: " h1 ", BusinessCode: " c1 "}, requested: "H1", holding: "h1", business: "C1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			ctx := e.NewContext(nil, nil)
			if tt.user.HoldingCode != "" {
				ctx.Set("UserInfo", tt.user)
			}
			holding, business, scopeErr := authenticatedCompanyContext(ctx, tt.requested, tt.requestedBusiness)
			if tt.status == 0 {
				if scopeErr != nil || holding != tt.holding || business != tt.business {
					t.Fatalf("got holding=%q business=%q err=%v", holding, business, scopeErr)
				}
				return
			}
			if scopeErr == nil || scopeErr.Status != tt.status {
				t.Fatalf("status = %v, want %d", scopeErr, tt.status)
			}
		})
	}
}

// ResolveReportCompany (ใช้โดย goapi/inventory): บริษัทมาจาก session เท่านั้น ไม่ผ่าน = เขียนข้อผิดพลาดแล้วคืน ok=false
func TestResolveReportCompany(t *testing.T) {
	e := echo.New()
	newCtx := func(user *msmodels.UserInfo) (echo.Context, *httptest.ResponseRecorder) {
		rec := httptest.NewRecorder()
		ctx := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/reports/stock-movement?holdingcode=H1", nil), rec)
		if user != nil {
			ctx.Set("UserInfo", *user)
		}
		return ctx, rec
	}

	ctx, rec := newCtx(&msmodels.UserInfo{HoldingCode: "H1", BusinessCode: " 02 "})
	company, ok, resp := ResolveReportCompany(ctx, "H1")
	if !ok || resp != nil || company != "02" || rec.Body.Len() != 0 {
		t.Fatalf("company=%q ok=%v resp=%v body=%s", company, ok, resp, rec.Body.String())
	}

	failures := []struct {
		name      string
		user      *msmodels.UserInfo
		requested string
		status    int
		code      string
	}{
		{name: "no session", requested: "H1", status: http.StatusUnauthorized, code: "UNAUTHORIZED"},
		{name: "no company selected", user: &msmodels.UserInfo{HoldingCode: "H1"}, requested: "H1", status: http.StatusConflict, code: "COMPANY_REQUIRED"},
		{name: "other holding", user: &msmodels.UserInfo{HoldingCode: "H1", BusinessCode: "01"}, requested: "H2", status: http.StatusForbidden, code: "FORBIDDEN"},
	}
	for _, tt := range failures {
		t.Run(tt.name, func(t *testing.T) {
			ctx, rec := newCtx(tt.user)
			company, ok, _ := ResolveReportCompany(ctx, tt.requested)
			if ok || company != "" || rec.Code != tt.status || !strings.Contains(rec.Body.String(), tt.code) {
				t.Fatalf("company=%q ok=%v status=%d body=%s", company, ok, rec.Code, rec.Body.String())
			}
		})
	}
}
