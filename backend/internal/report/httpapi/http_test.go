package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/internal/report"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/microservice/models"
)

// reportContext answers only the IContext methods the report handlers use (anything else panics)
type reportContext struct {
	microservice.IContext
	req      *http.Request
	user     models.UserInfo
	recorder *httptest.ResponseRecorder
	status   int
	body     any
}

func newContext(method, target, body, lang string, user models.UserInfo) *reportContext {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if lang != "" {
		req.Header.Set("Accept-Language", lang)
	}
	return &reportContext{req: req, user: user, recorder: httptest.NewRecorder()}
}

func (c *reportContext) Request() *http.Request              { return c.req }
func (c *reportContext) QueryParam(name string) string       { return c.req.URL.Query().Get(name) }
func (c *reportContext) Header(name string) string           { return c.req.Header.Get(name) }
func (c *reportContext) UserInfo() models.UserInfo           { return c.user }
func (c *reportContext) ResponseWriter() http.ResponseWriter { return c.recorder }
func (c *reportContext) Response(code int, data interface{}) { c.status, c.body = code, data }
func (c *reportContext) errorBody(t *testing.T) apperr.Response {
	t.Helper()
	body, ok := c.body.(apperr.Response)
	if !ok {
		t.Fatalf("status %d body %#v is not an error envelope", c.status, c.body)
	}
	return body
}

type fakeEngine struct{ report gl.Report }

func (f fakeEngine) Report(context.Context, gl.Scope, string, gl.ReportQuery) (gl.Report, error) {
	return f.report, nil
}

var member = models.UserInfo{Username: "demo", Name: "สมชาย ใจดี", HoldingCode: "rungrueng", BusinessCode: "01"}

func handler(permissions ...string) *Http {
	engine := fakeEngine{report: gl.Report{
		Columns:   []gl.ReportColumn{{Key: "accountcode"}, {Key: "accountname"}, {Key: "debit", Amount: true, Group: "ยอดประจำงวด"}, {Key: "credit", Amount: true, Group: "ยอดประจำงวด"}},
		Rows:      []map[string]string{{"accountcode": "11110", "accountname": "เงินสด", "debit": "=1+1", "credit": "0.00"}},
		Totals:    map[string]string{"debit": "100.005", "credit": "0"},
		TotalRows: 1,
	}}
	h := New(nil, report.NewRegistry(report.GLDefinitions(engine, nil)...),
		func(context.Context, microservice.IContext) (glhttp.SessionAccess, error) {
			return glhttp.NewSessionAccess(gl.Scope{Holding: "rungrueng", Company: "01", Actor: "demo"}, permissions...), nil
		},
		func(context.Context, string, string, string, string) (string, string, error) {
			return "บริษัท รุ่งเรืองการค้า จำกัด", "สำนักงานใหญ่", nil
		})
	h.now = func() time.Time { return time.Date(2026, 10, 11, 3, 0, 0, 0, time.UTC) }
	return h
}

const tbBody = `{"menu_id":"trial-balance","params":{"date_from":"2026-01-01","date_to":"2026-01-31"}}`

func TestRegistryListsPermittedMenusOnly(t *testing.T) {
	c := newContext(http.MethodGet, "/report/v1/registry?module=gl", "", "en", member)
	if err := handler("trial-balance", "general-ledger").Registry(c); err != nil {
		t.Fatal(err)
	}
	data := c.body.(map[string]interface{})["data"].(registryResponse)
	if c.status != 200 || data.CanPrint || data.CanExport || len(data.Reports) != 2 || data.Reports[0].MenuID != "general-ledger" {
		t.Fatalf("status %d data %+v", c.status, data)
	}
	c = newContext(http.MethodGet, "/report/v1/registry?module=gl", "", "", member)
	_ = handler("*").Registry(c)
	data = c.body.(map[string]interface{})["data"].(registryResponse)
	if !data.CanPrint || !data.CanExport || len(data.Reports) != 14 {
		t.Fatalf("owner registry = %d reports print %v export %v", len(data.Reports), data.CanPrint, data.CanExport)
	}
}

func TestCompanyIsRequired(t *testing.T) {
	c := newContext(http.MethodPost, "/report/v1/run", tbBody, "", models.UserInfo{Username: "demo", HoldingCode: "rungrueng"})
	_ = handler("*").Run(c)
	if body := c.errorBody(t); c.status != 400 || body.ErrorCode != "REPORT_COMPANY_REQUIRED" {
		t.Fatalf("status %d body %+v", c.status, body)
	}
}

func TestRunRefusals(t *testing.T) {
	for _, tc := range []struct {
		name, body, code string
		permissions      []string
		status           int
	}{
		{"malformed body", `{"menu_id":`, "REPORT_BAD_REQUEST", []string{"*"}, 400},
		{"unknown menu", `{"menu_id":"payroll"}`, "REPORT_UNKNOWN", []string{"*"}, 404},
		{"no menu permission", tbBody, "REPORT_FORBIDDEN", []string{"general-ledger"}, 403},
		{"planned report", `{"menu_id":"financial-statement-designer"}`, "REPORT_NOT_READY", []string{"*"}, 409},
		{"print without report:print", strings.Replace(tbBody, `{"menu_id"`, `{"print":true,"menu_id"`, 1), "REPORT_PRINT_FORBIDDEN", []string{"trial-balance"}, 403},
		{"variant preset override", strings.Replace(tbBody, `"date_to"`, `"post_status":"all","date_to"`, 1), "REPORT_VARIANT_FORBIDDEN", []string{"trial-balance"}, 403},
		{"company in params", strings.Replace(tbBody, `"date_to"`, `"company":"02","date_to"`, 1), "REPORT_PARAM_INVALID", []string{"trial-balance"}, 422},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newContext(http.MethodPost, "/report/v1/run", tc.body, "en", member)
			_ = handler(tc.permissions...).Run(c)
			body := c.errorBody(t)
			if c.status != tc.status || body.ErrorCode != tc.code || body.Message == "" || body.MessageTH == "" || body.Message == body.MessageTH {
				t.Fatalf("status %d body %+v", c.status, body)
			}
		})
	}
}

func TestRunReturnsLabelledReport(t *testing.T) {
	c := newContext(http.MethodPost, "/report/v1/run", tbBody, "", member)
	if err := handler("trial-balance").Run(c); err != nil {
		t.Fatal(err)
	}
	if c.status != 200 {
		t.Fatalf("status %d body %#v", c.status, c.body)
	}
	data := c.body.(map[string]interface{})["data"].(report.RunResponse)
	if data.Columns[2].Label != "เดบิต" || data.Columns[2].Group != "ยอดประจำงวด" || data.Totals["debit"] != "100.01" {
		t.Fatalf("columns %+v totals %v", data.Columns, data.Totals)
	}
	h := data.Header
	if h.CompanyName != "บริษัท รุ่งเรืองการค้า จำกัด" || h.BranchName != "สำนักงานใหญ่" || h.Title != "รายงานงบทดลอง" || h.PrintedBy != "สมชาย ใจดี" || h.PrintedAt != "2026-10-11T10:00:00+07:00" {
		t.Fatalf("header = %+v", h)
	}
}

func TestExportStreamsCSV(t *testing.T) {
	body := strings.Replace(tbBody, `{"menu_id"`, `{"format":"csv","menu_id"`, 1)
	c := newContext(http.MethodPost, "/report/v1/export", body, "", member)
	_ = handler("trial-balance").Export(c)
	if b := c.errorBody(t); c.status != 403 || b.ErrorCode != "REPORT_EXPORT_FORBIDDEN" {
		t.Fatalf("export without report:export: %d %+v", c.status, b)
	}
	c = newContext(http.MethodPost, "/report/v1/export", body, "", member)
	if err := handler("trial-balance", report.PermissionExport).Export(c); err != nil {
		t.Fatal(err)
	}
	res := c.recorder.Result()
	text := c.recorder.Body.String()
	if res.StatusCode != 200 || res.Header.Get("Content-Type") != "text/csv; charset=utf-8" || res.Header.Get("Cache-Control") != "no-store" ||
		!strings.Contains(res.Header.Get("Content-Disposition"), `filename="report-trial-balance.csv"`) {
		t.Fatalf("headers = %v", res.Header)
	}
	if !strings.HasPrefix(text, "\uFEFF") || !strings.Contains(text, `"11110","เงินสด","'=1+1","0.00"`) {
		t.Fatalf("csv = %q", text)
	}
}

// Champ "สิทธิ์เฉพาะ" ticked on a USER's permission group (screen /permissiongroup): the role's entries are what scope.go
// unions for a USER, so a role with report:print / report:export prints and exports, the same role without them is refused
func TestUserRoleSpecialRights(t *testing.T) {
	printBody := strings.Replace(tbBody, `{"menu_id"`, `{"print":true,"menu_id"`, 1)
	exportBody := strings.Replace(tbBody, `{"menu_id"`, `{"format":"csv","menu_id"`, 1)
	for _, tc := range []struct {
		name    string
		role    []string
		allowed bool
	}{
		{"role with special rights", []string{"general-ledger", "trial-balance", report.PermissionPrint, report.PermissionExport}, true},
		{"role with screens only", []string{"general-ledger", "trial-balance"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newContext(http.MethodGet, "/report/v1/registry?module=gl", "", "", member)
			_ = handler(tc.role...).Registry(c)
			if data := c.body.(map[string]interface{})["data"].(registryResponse); data.CanPrint != tc.allowed || data.CanExport != tc.allowed {
				t.Fatalf("registry print %v export %v", data.CanPrint, data.CanExport)
			}
			c = newContext(http.MethodPost, "/report/v1/run", printBody, "", member)
			_ = handler(tc.role...).Run(c)
			if tc.allowed && c.status != 200 {
				t.Fatalf("print refused: %d %#v", c.status, c.body)
			}
			if !tc.allowed && (c.status != 403 || c.errorBody(t).ErrorCode != "REPORT_PRINT_FORBIDDEN") {
				t.Fatalf("print without report:print: %d %#v", c.status, c.body)
			}
			c = newContext(http.MethodPost, "/report/v1/export", exportBody, "", member)
			_ = handler(tc.role...).Export(c)
			if tc.allowed && c.recorder.Result().StatusCode != 200 {
				t.Fatalf("export refused: %d %#v", c.status, c.body)
			}
			if !tc.allowed && (c.status != 403 || c.errorBody(t).ErrorCode != "REPORT_EXPORT_FORBIDDEN") {
				t.Fatalf("export without report:export: %d %#v", c.status, c.body)
			}
		})
	}
}

func TestSessionErrorsUseGLContract(t *testing.T) {
	h := handler("*")
	h.session = func(context.Context, microservice.IContext) (glhttp.SessionAccess, error) {
		return glhttp.SessionAccess{}, errors.New("membership lookup failed")
	}
	c := newContext(http.MethodPost, "/report/v1/run", tbBody, "", member)
	_ = h.Run(c)
	if c.status < 400 {
		t.Fatalf("a session failure must not run the report: %d", c.status)
	}
}
