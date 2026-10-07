//go:build integration

package generalledger

import (
	"strings"
	"testing"
)

// Champ 5525 budget detail report (GLRepBudgetView): rows per budget × account grouped by account, aggregated in PostgreSQL —
// heading → rows (account's 12 period amounts summed) → account total with count; report totals count the rows only.
func TestBudgetDetailReportIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	create := func(b Budget) { f.run(Command{Resource: "budgets", Action: "create", Budget: &b}) }
	create(Budget{Code: "BG-B", Name: "งบขายและบริหาร", FiscalYear: "2026", BranchCode: "B1", DepartmentCode: "D2",
		Lines: []BudgetLine{{AccountCode: "4000", Periods: budgetMonths("0.10")}}})
	create(Budget{Code: "BG-A", Name: "งบสำนักงาน", FiscalYear: "2026", BranchCode: "B1", DepartmentCode: "D1", ProjectCode: "P1",
		Lines: []BudgetLine{{AccountCode: "5000", Periods: budgetMonths("50.25")}, {AccountCode: "4000", Periods: budgetMonths("100")}}})

	report, err := f.store.Report(f.ctx, f.scope, "budgetdetail", ReportQuery{})
	if err != nil {
		t.Fatal(err)
	}
	kinds := make([]string, len(report.Rows))
	for i, row := range report.Rows {
		kinds[i] = row["accountcode"] + ":" + row["rowtype"] + ":" + row["budgetcode"]
	}
	// account, then budget code (GLRepBudgetDlg.cpp:429); header and total rows sandwich each account's rows
	want := "4000:budgetaccount:,4000:budgetline:BG-A,4000:budgetline:BG-B,4000:budgettotal:,5000:budgetaccount:,5000:budgetline:BG-A,5000:budgettotal:"
	if got := strings.Join(kinds, ","); got != want || report.TotalRows != 7 {
		t.Fatalf("rows = %s (total %d)", got, report.TotalRows)
	}
	head, a4000, b4000, total4000, total5000 := report.Rows[0], report.Rows[1], report.Rows[2], report.Rows[3], report.Rows[6]
	if head["accountname"] != "4000" || head["amount"] != "" {
		t.Fatalf("account heading = %v", head)
	}
	if a4000["budgetname"] != "งบสำนักงาน" || !sameAmount(a4000["amount"], "1200") || a4000["branchcode"] != "B1" || a4000["departmentcode"] != "D1" || a4000["projectcode"] != "P1" || a4000["startdate"] != "2026-01-01" || a4000["enddate"] != "2026-12-31" {
		t.Fatalf("BG-A row = %v", a4000)
	}
	// 12 × 0.10 stays exact in NUMERIC: 1.20, and the account total is 1201.20 for 2 rows
	if !sameAmount(b4000["amount"], "1.20") || b4000["departmentcode"] != "D2" || b4000["projectcode"] != "" {
		t.Fatalf("BG-B row = %v", b4000)
	}
	if !sameAmount(total4000["amount"], "1201.20") || total4000["linecount"] != "2" || !sameAmount(total5000["amount"], "603") || total5000["linecount"] != "1" {
		t.Fatalf("account totals = %v / %v", total4000, total5000)
	}
	if !sameAmount(report.Totals["amount"], "1804.20") || report.Totals["linecount"] != "3" {
		t.Fatalf("grand totals = %v", report.Totals)
	}

	// paging keeps the whole-report totals; filters narrow budgets (budgetcode, account range, fiscal year)
	page2, err := f.store.Report(f.ctx, f.scope, "budgetdetail", ReportQuery{Page: 2, Limit: 4})
	if err != nil || len(page2.Rows) != 3 || page2.TotalRows != 7 || page2.Rows[0]["accountcode"] != "5000" || !sameAmount(page2.Totals["amount"], "1804.20") {
		t.Fatalf("page 2 = %+v %v", page2, err)
	}
	if one, err := f.store.Report(f.ctx, f.scope, "budgetdetail", ReportQuery{BudgetCode: "BG-B"}); err != nil || len(one.Rows) != 3 || one.Totals["linecount"] != "1" {
		t.Fatalf("budgetcode filter = %+v %v", one.Rows, err)
	}
	// budget code range (Champ 5525 GLRepBudgetDlg.cpp:244-251): "to" without "from" is ignored, a reversed range is refused
	if span, err := f.store.Report(f.ctx, f.scope, "budgetdetail", ReportQuery{BudgetCode: "BG-A", BudgetCodeTo: "BG-B"}); err != nil || span.Totals["linecount"] != "3" {
		t.Fatalf("budget code range BG-A..BG-B = %+v %v", span.Totals, err)
	}
	if span, err := f.store.Report(f.ctx, f.scope, "budgetdetail", ReportQuery{BudgetCode: "BG-A", BudgetCodeTo: "BG-A"}); err != nil || span.Totals["linecount"] != "2" {
		t.Fatalf("budget code range BG-A..BG-A = %+v %v", span.Totals, err)
	}
	if span, err := f.store.Report(f.ctx, f.scope, "budgetdetail", ReportQuery{BudgetCodeTo: "BG-A"}); err != nil || span.Totals["linecount"] != "3" {
		t.Fatalf("budget code to without from = %+v %v", span.Totals, err)
	}
	if _, err := f.store.Report(f.ctx, f.scope, "budgetdetail", ReportQuery{BudgetCode: "BG-B", BudgetCodeTo: "BG-A"}); err == nil {
		t.Fatal("reversed budget code range accepted")
	}
	if only, err := f.store.Report(f.ctx, f.scope, "budgetdetail", ReportQuery{AccountFrom: "5000", AccountTo: "5000"}); err != nil || len(only.Rows) != 3 || only.Rows[0]["accountcode"] != "5000" {
		t.Fatalf("account range filter = %+v %v", only.Rows, err)
	}
	if _, err := f.store.Report(f.ctx, f.scope, "budgetdetail", ReportQuery{AccountFrom: "5000", AccountTo: "4000"}); err == nil {
		t.Fatal("reversed account range accepted")
	}
	if none, err := f.store.Report(f.ctx, f.scope, "budgetdetail", ReportQuery{FiscalYear: "2025"}); err != nil || len(none.Rows) != 0 || none.Totals["linecount"] != "0" {
		t.Fatalf("other fiscal year = %+v %v", none, err)
	}

	// a branch session sees only its own branch's budgets and cannot ask for another branch
	branchB2 := f.scope
	branchB2.Branch = "B2"
	if b2, err := f.store.Report(f.ctx, branchB2, "budgetdetail", ReportQuery{}); err != nil || len(b2.Rows) != 0 {
		t.Fatalf("branch B2 sees branch B1 budgets = %+v %v", b2.Rows, err)
	}
	if _, err := f.store.Report(f.ctx, branchB2, "budgetdetail", ReportQuery{BranchCode: "B1"}); err == nil {
		t.Fatal("branch B2 session read branch B1 budgets")
	}
	if b1, err := f.store.Report(f.ctx, f.scope, "budgetdetail", ReportQuery{BranchCode: "B1"}); err != nil || len(b1.Rows) != 7 {
		t.Fatalf("branch B1 filter = %+v %v", b1.Rows, err)
	}
}
