package handlers

import (
	"net/http"
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
