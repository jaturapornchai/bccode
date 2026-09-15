package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

// -------------------- isValidReportPeriod --------------------

func TestIsValidReportPeriod(t *testing.T) {
	cases := []struct {
		name  string
		year  int
		month int
		want  bool
	}{
		{"valid", 2026, 9, true},
		{"month zero", 2026, 0, false},
		{"month 13", 2026, 13, false},
		{"month negative", 2026, -1, false},
		{"year too small", 1999, 1, false},
		{"year too large", 2101, 1, false},
		{"boundary month 1", 2026, 1, true},
		{"boundary month 12", 2026, 12, true},
		{"zero year and month", 0, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isValidReportPeriod(tc.year, tc.month); got != tc.want {
				t.Errorf("isValidReportPeriod(%d,%d) = %v, want %v", tc.year, tc.month, got, tc.want)
			}
		})
	}
}

// -------------------- normalizeVatRegisterPaging --------------------

func TestNormalizeVatRegisterPaging(t *testing.T) {
	cases := []struct {
		name       string
		limit      int
		offset     int
		wantLimit  int
		wantOffset int
	}{
		{"defaults when zero", 0, 0, taxRegisterDefaultLimit, 0},
		{"defaults when negative limit", -5, 0, taxRegisterDefaultLimit, 0},
		{"defaults when over max", 9999, 0, taxRegisterDefaultLimit, 0},
		{"within bounds kept as-is", 50, 100, 50, 100},
		{"exactly max limit kept", taxRegisterMaxLimit, 0, taxRegisterMaxLimit, 0},
		{"negative offset clamped to zero", 50, -10, 50, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotLimit, gotOffset := normalizeVatRegisterPaging(tc.limit, tc.offset)
			if gotLimit != tc.wantLimit || gotOffset != tc.wantOffset {
				t.Errorf("normalizeVatRegisterPaging(%d,%d) = (%d,%d), want (%d,%d)",
					tc.limit, tc.offset, gotLimit, gotOffset, tc.wantLimit, tc.wantOffset)
			}
		})
	}
}

// -------------------- computeVatSettlement --------------------

func TestComputeVatSettlement(t *testing.T) {
	cases := []struct {
		name           string
		outputVat      float64
		inputVat       float64
		wantNet        float64
		wantPayable    float64
		wantCreditable float64
	}{
		{"output greater than input => payable", 1000, 400, 600, 600, 0},
		{"input greater than output => creditable", 400, 1000, -600, 0, 600},
		{"equal => zero both", 500, 500, 0, 0, 0},
		{"zero both", 0, 0, 0, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			net, payable, creditable := computeVatSettlement(tc.outputVat, tc.inputVat)
			if net != tc.wantNet || payable != tc.wantPayable || creditable != tc.wantCreditable {
				t.Errorf("computeVatSettlement(%v,%v) = (%v,%v,%v), want (%v,%v,%v)",
					tc.outputVat, tc.inputVat, net, payable, creditable, tc.wantNet, tc.wantPayable, tc.wantCreditable)
			}
		})
	}
}

// -------------------- buildVatRegisterQuery --------------------

func TestBuildVatRegisterQuery_Sale(t *testing.T) {
	query, args := buildVatRegisterQuery("sale", 2026, 9, 200, 0)

	if !strings.Contains(query, "public.saleinvoicetransaction") {
		t.Errorf("expected sale query to reference saleinvoicetransaction, got: %s", query)
	}
	if !strings.Contains(query, "public.debtor") {
		t.Errorf("expected sale query to join debtor, got: %s", query)
	}
	if strings.Contains(query, "purchasetransaction") || strings.Contains(query, "public.creditor") {
		t.Errorf("sale query must not reference purchase tables, got: %s", query)
	}
	wantArgs := []any{2026, 9, 200, 0}
	if len(args) != len(wantArgs) {
		t.Fatalf("expected %d args, got %d (%v)", len(wantArgs), len(args), args)
	}
	for i := range wantArgs {
		if args[i] != wantArgs[i] {
			t.Errorf("arg[%d] = %v, want %v", i, args[i], wantArgs[i])
		}
	}
}

func TestBuildVatRegisterQuery_Purchase(t *testing.T) {
	query, args := buildVatRegisterQuery("purchase", 2026, 9, 50, 100)

	if !strings.Contains(query, "public.purchasetransaction") {
		t.Errorf("expected purchase query to reference purchasetransaction, got: %s", query)
	}
	if !strings.Contains(query, "public.creditor") {
		t.Errorf("expected purchase query to join creditor, got: %s", query)
	}
	if strings.Contains(query, "saleinvoicetransaction") || strings.Contains(query, "public.debtor") {
		t.Errorf("purchase query must not reference sale tables, got: %s", query)
	}
	wantArgs := []any{2026, 9, 50, 100}
	for i := range wantArgs {
		if args[i] != wantArgs[i] {
			t.Errorf("arg[%d] = %v, want %v", i, args[i], wantArgs[i])
		}
	}
}

func TestBuildVatRegisterQuery_ExcludesCancelledAndZeroVat(t *testing.T) {
	query, _ := buildVatRegisterQuery("sale", 2026, 9, 200, 0)
	if !strings.Contains(query, "t.iscancel = false") {
		t.Errorf("expected query to exclude cancelled documents, got: %s", query)
	}
	if !strings.Contains(query, "t.totalvatvalue <> 0") {
		t.Errorf("expected query to only include documents with VAT, got: %s", query)
	}
}

// -------------------- buildPP30SalesQuery / buildPP30PurchaseQuery --------------------

func TestBuildPP30SalesQuery(t *testing.T) {
	query, args := buildPP30SalesQuery(2026, 9)
	if !strings.Contains(query, "public.saleinvoicetransaction") {
		t.Errorf("expected sales query to reference saleinvoicetransaction, got: %s", query)
	}
	if !strings.Contains(query, "SUM(s.totalvatvalue)") {
		t.Errorf("expected outputvat to be SUM(totalvatvalue) not a computed rate, got: %s", query)
	}
	if strings.Contains(query, "* 0.07") || strings.Contains(query, "*0.07") {
		t.Errorf("must not compute VAT from base * 7%%, got: %s", query)
	}
	wantArgs := []any{2026, 9}
	for i := range wantArgs {
		if args[i] != wantArgs[i] {
			t.Errorf("arg[%d] = %v, want %v", i, args[i], wantArgs[i])
		}
	}
}

func TestBuildPP30PurchaseQuery(t *testing.T) {
	query, args := buildPP30PurchaseQuery(2026, 9)
	if !strings.Contains(query, "public.purchasetransaction") {
		t.Errorf("expected purchase query to reference purchasetransaction, got: %s", query)
	}
	if !strings.Contains(query, "SUM(p.totalvatvalue)") {
		t.Errorf("expected inputvat to be SUM(totalvatvalue) not a computed rate, got: %s", query)
	}
	wantArgs := []any{2026, 9}
	for i := range wantArgs {
		if args[i] != wantArgs[i] {
			t.Errorf("arg[%d] = %v, want %v", i, args[i], wantArgs[i])
		}
	}
}

// -------------------- HTTP-level validation (pre-DB) --------------------

var taxReportTestUser = msmodels.UserInfo{Username: "tester", HoldingCode: "h1", BusinessCode: "b1", UID: "u1"}

func callTaxReportHandler(t *testing.T, handler echo.HandlerFunc, body string, user *msmodels.UserInfo) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	if user != nil {
		ctx.Set("UserInfo", *user)
	}
	if err := handler(ctx); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	return rec
}

func TestTaxVatRegisterHandler_Unauthorized(t *testing.T) {
	rec := callTaxReportHandler(t, TaxVatRegisterHandler, `{"year":2026,"month":9,"type":"sale"}`, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTaxVatRegisterHandler_InvalidType(t *testing.T) {
	rec := callTaxReportHandler(t, TaxVatRegisterHandler, `{"year":2026,"month":9,"type":"refund"}`, &taxReportTestUser)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INVALID_TYPE") {
		t.Errorf("expected INVALID_TYPE code, got body=%s", rec.Body.String())
	}
}

func TestTaxVatRegisterHandler_InvalidPeriod(t *testing.T) {
	rec := callTaxReportHandler(t, TaxVatRegisterHandler, `{"year":2026,"month":13,"type":"sale"}`, &taxReportTestUser)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INVALID_PERIOD") {
		t.Errorf("expected INVALID_PERIOD code, got body=%s", rec.Body.String())
	}
}

func TestTaxVatRegisterHandler_InvalidPayload(t *testing.T) {
	rec := callTaxReportHandler(t, TaxVatRegisterHandler, `{invalid-json`, &taxReportTestUser)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INVALID_PAYLOAD") {
		t.Errorf("expected INVALID_PAYLOAD code, got body=%s", rec.Body.String())
	}
}

func TestPP30SummaryHandler_Unauthorized(t *testing.T) {
	rec := callTaxReportHandler(t, PP30SummaryHandler, `{"year":2026,"month":9}`, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPP30SummaryHandler_InvalidPeriod(t *testing.T) {
	rec := callTaxReportHandler(t, PP30SummaryHandler, `{"year":2026,"month":0}`, &taxReportTestUser)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INVALID_PERIOD") {
		t.Errorf("expected INVALID_PERIOD code, got body=%s", rec.Body.String())
	}
}

func TestPP30SummaryHandler_InvalidPayload(t *testing.T) {
	rec := callTaxReportHandler(t, PP30SummaryHandler, `{invalid-json`, &taxReportTestUser)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "INVALID_PAYLOAD") {
		t.Errorf("expected INVALID_PAYLOAD code, got body=%s", rec.Body.String())
	}
}
