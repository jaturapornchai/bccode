package report

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/lib/pq"

	gl "smlcloudplatform/internal/generalledger"
)

type allowList map[string]bool

func (a allowList) Allows(screen, action string) bool { return a["*"] || action == "" && a[screen] }

// fakeEngine returns one canned GL report and records the query it received
type fakeEngine struct {
	report gl.Report
	err    error
	name   string
	query  gl.ReportQuery
	scope  gl.Scope
}

func (f *fakeEngine) Report(_ context.Context, scope gl.Scope, name string, q gl.ReportQuery) (gl.Report, error) {
	f.scope, f.name, f.query = scope, name, q
	return f.report, f.err
}

func engineRegistry(engine GLEngine) *Registry { return NewRegistry(GLDefinitions(engine, nil)...) }

func TestPrepareChecksInOrder(t *testing.T) {
	r := testRegistry()
	scope := gl.Scope{Holding: "H", Company: "C"}
	run := func(req RunRequest, access Access, export bool) *Failure {
		_, failure := r.Prepare(req, access, scope, export, "th")
		return failure
	}
	ledger := RunRequest{MenuID: "general-ledger", Params: journalInput(nil)}
	wantFailure(t, run(RunRequest{MenuID: "no-such-report"}, allowList{"*": true}, false), "REPORT_UNKNOWN", "menu_id")
	wantFailure(t, run(ledger, allowList{"trial-balance": true}, false), "REPORT_FORBIDDEN", "")
	wantFailure(t, run(RunRequest{MenuID: "financial-statement-designer"}, allowList{"*": true}, false), "REPORT_NOT_READY", "")
	printing := ledger
	printing.Print = true
	wantFailure(t, run(printing, allowList{"general-ledger": true}, false), "REPORT_PRINT_FORBIDDEN", "")
	export := ledger
	export.Format = "csv"
	wantFailure(t, run(export, allowList{"general-ledger": true, PermissionPrint: true}, true), "REPORT_EXPORT_FORBIDDEN", "")
	export.Format = "xlsx"
	wantFailure(t, run(export, allowList{"general-ledger": true, PermissionExport: true}, true), "REPORT_EXPORT_FORMAT_UNAVAILABLE", "format")
	paged := ledger
	paged.Limit = maxLimit + 1
	wantFailure(t, run(paged, allowList{"general-ledger": true}, false), "REPORT_LIMIT_INVALID", "limit")
	paged.Limit, paged.Page = 10, -1
	wantFailure(t, run(paged, allowList{"general-ledger": true}, false), "REPORT_LIMIT_INVALID", "page")
	printing.RowsPerPage = minRowsPerPage - 1
	wantFailure(t, run(printing, allowList{"general-ledger": true, PermissionPrint: true}, false), "REPORT_LIMIT_INVALID", "rows_per_page")
	branch := ledger
	branch.Params = journalInput(map[string]string{"branch": "B2"})
	if _, failure := r.Prepare(branch, allowList{"*": true}, gl.Scope{Holding: "H", Company: "C", Branch: "B1"}, false, "th"); failure == nil || failure.Code != "REPORT_FORBIDDEN" || failure.Field != "branch" {
		t.Fatalf("a branch session must not read another branch: %v", failure)
	}
}

func TestPreparePagingAndPrintModes(t *testing.T) {
	r := testRegistry()
	scope := gl.Scope{Holding: "H", Company: "C"}
	plan, failure := r.Prepare(RunRequest{MenuID: "general-ledger", Params: journalInput(nil)}, allowList{"general-ledger": true}, scope, false, "en")
	if failure != nil {
		t.Fatal(failure)
	}
	if plan.Source.Page != 1 || plan.Source.Limit != defaultLimit || plan.Source.All || plan.Source.Timeout != screenTimeout || plan.Source.Scope != scope || plan.Source.Lang != "en" {
		t.Fatalf("screen plan = %+v", plan.Source)
	}
	plan, failure = r.Prepare(RunRequest{MenuID: "general-ledger", Params: journalInput(nil), Print: true}, allowList{"*": true}, scope, false, "th")
	if failure != nil {
		t.Fatal(failure)
	}
	if !plan.Source.All || plan.RowsPerPage != defaultRowsPerPage || plan.Source.Timeout != printTimeout {
		t.Fatalf("print plan = %+v rows/page %d", plan.Source, plan.RowsPerPage)
	}
}

func TestGLQueryMapsParams(t *testing.T) {
	q := glQuery(SourceRequest{Params: map[string]string{
		"date_from": "2026-01-01", "date_to": "2026-01-31", "account_from": "1000", "account_to": "2000", "branch": "B1",
		"departments": "D1,D2", "projects": "P1", "books": "JV", "doc_no_from": "JV001", "doc_no_to": "JV009", "post_status": "posted",
		"budget_code": "BG1", "budget_from": "BG2", "budget_to": "BG9", "summary": "true", "gross": "false", "day_totals": "true",
	}, Sort: "date", Page: 2, Limit: 50, All: true})
	if q.From != "2026-01-01" || q.To != "2026-01-31" || q.AccountFrom != "1000" || q.AccountTo != "2000" || q.BranchCode != "B1" ||
		q.DepartmentCode != "D1,D2" || q.ProjectCode != "P1" || q.BookCode != "JV" || q.DocNoFrom != "JV001" || q.DocNoTo != "JV009" ||
		q.PostStatus != "posted" || q.BudgetCode != "BG2" || q.BudgetCodeTo != "BG9" || !q.Summary || q.Gross || !q.DayTotals ||
		q.OrderBy != "date" || q.Page != 2 || q.Limit != 50 || q.RowCap != MaxPrintRows+1 {
		t.Fatalf("query = %+v", q)
	}
}

func TestExecuteRefusesPrintAboveCap(t *testing.T) {
	engine := &fakeEngine{report: gl.Report{TotalRows: MaxPrintRows + 1}}
	plan, failure := engineRegistry(engine).Prepare(RunRequest{MenuID: "general-ledger", Params: journalInput(nil), Print: true}, allowList{"*": true}, gl.Scope{Holding: "H", Company: "C"}, false, "th")
	if failure != nil {
		t.Fatal(failure)
	}
	_, failure = plan.Execute(context.Background())
	wantFailure(t, failure, "REPORT_EXPORT_TOO_LARGE", "")
	if _, body := failure.Response("th"); !strings.Contains(body.Message, "100,000") {
		t.Fatalf("message = %s", body.Message)
	}
	if engine.query.RowCap != MaxPrintRows+1 {
		t.Fatalf("row cap = %d", engine.query.RowCap)
	}
}

func TestFromEngineMapsErrors(t *testing.T) {
	user := &gl.UserError{Code: "report_range_outside_fiscal_year", Message: "ช่วงไม่ถูก", Status: 422, Field: "from"}
	f := fromEngine(fmt.Errorf("wrapped: %w", user))
	if f.Code != "REPORT_DATE_RANGE_INVALID" || f.Field != "date_from" || f.Key != "gl_err_report_range_outside_fiscal_year" || f.Status != 422 {
		t.Fatalf("user error = %+v", f)
	}
	for err, want := range map[error]string{
		gl.ErrProjectionPending:                    "REPORT_PROJECTION_PENDING",
		context.DeadlineExceeded:                   "REPORT_TIMEOUT",
		&pq.Error{Code: "57014"}:                   "REPORT_TIMEOUT",
		errors.New("disk is full"):                 "REPORT_ENGINE_FAILED",
		fail("REPORT_PARAM_INVALID", "group_from"): "REPORT_PARAM_INVALID",
	} {
		if got := fromEngine(err); got.Code != want {
			t.Errorf("%v -> %s, want %s", err, got.Code, want)
		}
	}
	if status, body := fromEngine(errors.New("connection refused to 10.0.0.1")).Response("en"); status != 500 || strings.Contains(body.Message, "10.0.0.1") {
		t.Fatalf("engine failure must not leak its cause: %d %s", status, body.Message)
	}
}

func TestFailureResponseSpeaksRequestLanguage(t *testing.T) {
	status, body := fail("REPORT_VARIANT_FORBIDDEN", "gross").Response("en")
	if status != 403 || body.ErrorCode != "REPORT_VARIANT_FORBIDDEN" || body.Field != "gross" || body.StatusCode != 403 {
		t.Fatalf("response = %d %+v", status, body)
	}
	if !strings.HasPrefix(body.Message, "This menu fixes") || !strings.HasPrefix(body.MessageTH, "เมนูนี้กำหนด") {
		t.Fatalf("messages = %q / %q", body.Message, body.MessageTH)
	}
}

// the response labels columns from the catalog (variant overrides first), rounds totals once and builds page totals
func TestResponseFromGLReport(t *testing.T) {
	engine := &fakeEngine{report: gl.Report{
		Columns: []gl.ReportColumn{{Key: "accountcode", Label: "x"}, {Key: "accountname", Label: "x"},
			{Key: "openingdebit", Label: "เดบิต", Amount: true, Group: "ยอดยกมา"}, {Key: "openingcredit", Label: "เครดิต", Amount: true, Group: "ยอดยกมา"},
			{Key: "debit", Label: "เดบิต", Amount: true, Group: "ยอดประจำงวด"}, {Key: "credit", Label: "เครดิต", Amount: true, Group: "ยอดประจำงวด"}},
		Rows: []map[string]string{
			{"rowtype": "control", "accountcode": "1000", "debit": "0.30"},
			{"accountcode": "1100", "debit": "0.1", "credit": "0"},
			{"accountcode": "1200", "debit": "0.2", "credit": "0"},
		},
		Totals:    map[string]string{"debit": "0.3000", "credit": "0", "difference": "0.3", "internal": "1"},
		TotalRows: 3, AsOf: "2026-10-11T00:00:00Z",
	}}
	r := engineRegistry(engine)
	plan, failure := r.Prepare(RunRequest{MenuID: "trial-balance", Params: journalInput(map[string]string{"hide_zero": "true", "control_totals": "false"}), Print: true, RowsPerPage: 5}, allowList{"*": true}, gl.Scope{Holding: "H", Company: "C"}, false, "en")
	if failure != nil {
		t.Fatal(failure)
	}
	result, failure := plan.Execute(context.Background())
	if failure != nil {
		t.Fatal(failure)
	}
	if engine.name != "trialbalance" || engine.query.PostStatus != "posted" || engine.query.Gross {
		t.Fatalf("engine got %s %+v", engine.name, engine.query)
	}
	resp := plan.Response(result, Header{CompanyName: "บริษัท รุ่งเรืองการค้า จำกัด"}, "en")
	if resp.Columns[0].Label != "Account Code" || resp.Columns[1].Label != "Item" || resp.Columns[2].Group != "Opening Balance" || resp.Columns[2].Scale != 2 {
		t.Fatalf("columns = %+v", resp.Columns)
	}
	if resp.Rows[0][KindKey] != KindControlTotal || resp.Rows[1][KindKey] != KindDetail {
		t.Fatalf("row kinds = %v", resp.Rows)
	}
	if resp.Totals["debit"] != "0.30" || resp.Totals["difference"] != "0.30" || resp.Totals["internal"] != "" {
		t.Fatalf("totals = %v", resp.Totals)
	}
	if resp.TotalsLabels["difference"] != "Difference" || resp.TotalsLabels["debit"] != "" {
		t.Fatalf("totals labels = %v", resp.TotalsLabels)
	}
	if len(resp.PageTotals) != 1 || resp.PageTotals[0].Sums["debit"] != "0.30" {
		t.Fatalf("page totals = %+v", resp.PageTotals)
	}
	if resp.Header.Title != "Trial Balance" || resp.Header.CompanyName == "" || resp.Page != 1 || resp.Limit != 3 || resp.Warnings == nil {
		t.Fatalf("header/paging = %+v page %d limit %d", resp.Header, resp.Page, resp.Limit)
	}
	conditions := map[string]Condition{}
	for _, c := range resp.Header.Conditions {
		conditions[c.Key] = c
	}
	if _, ok := conditions["post_status"]; ok {
		t.Fatal("a preset must not print as a chosen condition")
	}
	if c, ok := conditions["hide_zero"]; !ok || c.Value != "" || c.Label != "Hide Zero Balances" {
		t.Fatalf("bool condition = %+v", c)
	}
	if _, ok := conditions["control_totals"]; ok {
		t.Fatal("a switched-off bool must not print")
	}
}

func TestSummaryVariantRelabelsColumns(t *testing.T) {
	d, v := lookup(t, "general-ledger-summary")
	views := d.columnViews(v, []SourceColumn{{Key: "accountcode"}, {Key: "accountname"}, {Key: "unknownengine", Amount: true}}, "en")
	if views[0].Label != "Account Code" || views[1].Label != "Item" {
		t.Fatalf("labels = %+v", views)
	}
	if views[2].Label != "report_col_unknownengine" || views[2].Type != ColAmount || views[2].Scale != 2 {
		t.Fatalf("an uncatalogued column falls back to a language key, never engine text: %+v", views[2])
	}
}
