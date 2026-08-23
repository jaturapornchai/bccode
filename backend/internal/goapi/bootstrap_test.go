package goapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestGoAPIRequestHoldingCodeReadsJSONAndPreservesBody(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/get", strings.NewReader(`{"holdingcode":"SHOP001"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	holdingCode, err := goAPIRequestHoldingCode(c)
	if err != nil {
		t.Fatalf("goAPIRequestHoldingCode returned error: %v", err)
	}
	if holdingCode != "SHOP001" {
		t.Fatalf("holdingCode = %q, want SHOP001", holdingCode)
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		t.Fatalf("read preserved body: %v", err)
	}
	if string(body) != `{"holdingcode":"SHOP001"}` {
		t.Fatalf("body = %q, want original body", string(body))
	}
}

func TestHoldingCodeFromPayloadReadsNestedJSONBody(t *testing.T) {
	holdingCode, err := holdingCodeFromPayload(map[string]interface{}{
		"body": `{"holdingcode":"SHOP002"}`,
	})
	if err != nil {
		t.Fatalf("holdingCodeFromPayload returned error: %v", err)
	}
	if holdingCode != "SHOP002" {
		t.Fatalf("holdingCode = %q, want SHOP002", holdingCode)
	}
}

func TestGoAPIRouteSurfaceExcludesOperationalEndpoints(t *testing.T) {
	e := echo.New()
	New().RegisterRoutes(e.Group("/goapi"), "/goapi")

	forbidden := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/goapi/reportget"},
		{http.MethodPost, "/goapi/reportpost"},
		{http.MethodGet, "/goapi/rebuild/progress/:jobId"},
		{http.MethodPost, "/goapi/resultget"},
		{http.MethodPost, "/goapi/resulttopdf"},
		{http.MethodPost, "/goapi/genpdf"},
		{http.MethodGet, "/goapi/genpdf/history"},
		{http.MethodPost, "/goapi/genpdf/history"},
		{http.MethodGet, "/goapi/genpdf/reprint/:id"},
		{http.MethodPost, "/goapi/get"},
		{http.MethodPost, "/goapi/exec"},
		{http.MethodPost, "/goapi/getdoc"},
		{http.MethodPost, "/goapi/mongogetdata"},
		{http.MethodPost, "/goapi/resultfromquery"},
		{http.MethodPost, "/goapi/atlas/get"},
		{http.MethodPost, "/goapi/atlas/update"},
		{http.MethodPost, "/goapi/atlas/delete"},
		{http.MethodPost, "/goapi/clickhouse/query"},
		{http.MethodPost, "/goapi/clickhouse/querys"},
		{http.MethodPost, "/goapi/clickhouse/select"},
		{http.MethodPost, "/goapi/copymongouattodev"},
		{http.MethodPost, "/goapi/previewcopymongo"},
		{http.MethodGet, "/goapi/listsourceshops"},
		{http.MethodGet, "/goapi/api/migrate/currency"},
		{http.MethodGet, "/goapi/api/migrate/currency-backfill"},
		{http.MethodGet, "/goapi/api/migrate/clickhouse-softdelete"},
		{http.MethodPost, "/goapi/api/inventory/create-tables"},
		{http.MethodPost, "/goapi/test/sale-order"},
		{http.MethodPost, "/goapi/test/purchase"},
		{http.MethodPost, "/goapi/test/purchase-order"},
		{http.MethodPost, "/goapi/test/purchase-partial"},
		{http.MethodPost, "/goapi/api/lineoa/test"},
		{http.MethodPost, "/goapi/api/approval/test-email"},
		{http.MethodPost, "/goapi/api/approval/test-line-push"},
		{http.MethodPost, "/goapi/api/v1/ai-provider/test"},
		{http.MethodPost, "/goapi/api/deploy/backend"},
		{http.MethodPost, "/goapi/api/deploy/frontend"},
		{http.MethodGet, "/goapi/api/deploy/status"},
	}

	routes := make(map[string]struct{}, len(e.Routes()))
	for _, route := range e.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}
	for _, tc := range forbidden {
		if _, registered := routes[tc.method+" "+tc.path]; registered {
			t.Errorf("forbidden route is still registered: %s %s", tc.method, tc.path)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/goapi/api/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status = %d, want 200", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/goapi/api/transaction/calculate", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("domain route status = %d, want 401 without a token", rec.Code)
	}
}
