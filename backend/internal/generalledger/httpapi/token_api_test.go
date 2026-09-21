package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"smlcloudplatform/internal/mcptoken"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/microservice/models"
)

func TestAPITokenReadWriteBoundary(t *testing.T) {
	for _, tc := range []struct {
		mode          string
		write, denied bool
		status        int
	}{
		{"readonly", true, false, 403}, {"readonly", false, false, 200}, {"readwrite", true, false, 200}, {"readwrite", true, true, 401}, {"unknown", false, false, 401},
	} {
		runs := 0
		h := &Http{}
		run := func(c microservice.IContext) error {
			runs++
			if c.UserInfo().BusinessCode != "bound-company" {
				t.Fatal("scope changed")
			}
			c.Response(200, map[string]any{"success": true})
			return nil
		}
		auth := func(context.Context, string) (mcptoken.Principal, error) {
			if tc.denied {
				return mcptoken.Principal{}, errors.New("revoked")
			}
			return mcptoken.Principal{Kind: "api", Mode: tc.mode, User: models.UserInfo{BusinessCode: "bound-company"}}, nil
		}
		r := httptest.NewRequest("POST", "/integration/gl/v2/command?company=OTHER", nil)
		r.Header.Set("Authorization", "Bearer test")
		w := httptest.NewRecorder()
		c := echo.New().NewContext(r, w)
		resolve := func(_ context.Context, p mcptoken.Principal, _ string) (mcptoken.Principal, error) { return p, nil }
		if err := h.tokenAPIWithAuth(run, tc.write, auth, resolve)(c); err != nil {
			t.Fatal(err)
		}
		if w.Code != tc.status {
			t.Fatalf("mode %s: got %d want %d", tc.mode, w.Code, tc.status)
		}
		if tc.status != 200 && runs != 0 {
			t.Fatal("denied request reached handler")
		}
	}
}

func TestAPIBatchReadsAreSeparatedAndFullyAuthorized(t *testing.T) {
	for _, tc := range []struct {
		name, mode           string
		denySecond           bool
		wantStatus, wantRuns int
	}{
		{"readonly multiple", "readonly", false, 200, 2},
		{"readwrite multiple reads forbidden", "readwrite", false, 403, 0},
		{"unauthorized company rejects entire selection", "readonly", true, 403, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := &Http{}
			runs := 0
			auth := func(context.Context, string) (mcptoken.Principal, error) {
				return mcptoken.Principal{Kind: "api", Mode: tc.mode}, nil
			}
			resolve := func(_ context.Context, p mcptoken.Principal, code string) (mcptoken.Principal, error) {
				if tc.denySecond && code == "C02" {
					return p, errors.New("denied")
				}
				p.User.BusinessCode = code
				return p, nil
			}
			run := func(c microservice.IContext) error {
				runs++
				c.Response(200, map[string]any{"success": true, "data": map[string]string{"company": c.UserInfo().BusinessCode, "amount": "0.30"}})
				return nil
			}
			r := httptest.NewRequest("GET", "/integration/gl/v2/accounts", nil)
			r.Header.Set("Authorization", "Bearer test")
			r.Header.Set("X-BC-Company-Codes", "C01,C02")
			w := httptest.NewRecorder()
			c := echo.New().NewContext(r, w)
			if err := h.tokenAPIWithAuth(run, false, auth, resolve)(c); err != nil {
				t.Fatal(err)
			}
			if w.Code != tc.wantStatus || runs != tc.wantRuns {
				t.Fatalf("status=%d runs=%d", w.Code, runs)
			}
			if tc.wantRuns == 2 {
				var result struct {
					Data struct {
						Companies []struct {
							CompanyCode string `json:"companyCode"`
						} `json:"companies"`
					} `json:"data"`
				}
				if json.Unmarshal(w.Body.Bytes(), &result) != nil || len(result.Data.Companies) != 2 || result.Data.Companies[0].CompanyCode != "C01" || result.Data.Companies[1].CompanyCode != "C02" {
					t.Fatal(w.Body.String())
				}
			}
		})
	}
}
