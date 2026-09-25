//go:build integration

package generalledger

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func budgetMonths(amount Amount) []Amount {
	out := make([]Amount, BudgetPeriods)
	for i := range out {
		out[i] = amount
	}
	return out
}

func (f *pgIntegrityFixture) budget(code string) budgetRecord {
	f.t.Helper()
	raw, err := f.store.Get(f.ctx, f.scope, "budgets", code)
	if err != nil {
		f.t.Fatal(err)
	}
	var r budgetRecord
	if err = json.Unmarshal(raw, &r); err != nil {
		f.t.Fatal(err)
	}
	return r
}

// sameAmount compares decimal strings by value ("1803" == "1803.00" == "1803.00000000").
func sameAmount(got, want string) bool {
	g, err1 := decimal.NewFromString(got)
	w, err2 := decimal.NewFromString(want)
	return err1 == nil && err2 == nil && g.Equal(w)
}

// Budget header + monthly lines (Champ 5500 BCGLBudget, one amount per account per month):
// create → read → update (full line replace + Champ open/closed status) → branch isolation →
// budget-vs-actual (Champ 5530) → delete, each step checked in PostgreSQL.
func TestBudgetLifecycleAndComparisonIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	b := Budget{Code: "BG-2569", Name: "งบประมาณรายได้และค่าใช้จ่าย ปี 2569", FiscalYear: "2026", BranchCode: "B1",
		Lines: []BudgetLine{{AccountCode: "4000", Periods: budgetMonths("100")}, {AccountCode: "5000", Periods: budgetMonths("50.25")}}}

	bad := func(mut func(*Budget)) *Budget { x := b; x.Code = "X"; mut(&x); return &x }
	f.failCode(Command{Resource: "budgets", Action: "create", Budget: bad(func(x *Budget) { x.Name = "" })}, "budget_name_required", "name")
	f.failCode(Command{Resource: "budgets", Action: "create", Budget: bad(func(x *Budget) { x.Status = "approved" })}, "budget_status_invalid", "status")
	f.failCode(Command{Resource: "budgets", Action: "create", Budget: bad(func(x *Budget) { x.FiscalYear = "2099" })}, "budget_fiscal_year_not_found", "fiscalyear")
	f.failCode(Command{Resource: "budgets", Action: "create", Budget: bad(func(x *Budget) { x.Lines = []BudgetLine{b.Lines[0], b.Lines[0]} })}, "budget_account_duplicate", "lines")
	f.failCode(Command{Resource: "budgets", Action: "create", Budget: bad(func(x *Budget) { x.Lines = []BudgetLine{{AccountCode: "9999", Periods: budgetMonths("1")}} })}, "budget_account_not_found", "lines")
	f.failCode(Command{Resource: "budgets", Action: "create", Budget: bad(func(x *Budget) { x.Lines = []BudgetLine{{AccountCode: "4000", Periods: budgetMonths("1.001")}} })}, "budget_amount_scale", "lines")
	f.failCode(Command{Resource: "budgets", Action: "create", Budget: bad(func(x *Budget) { x.Lines = []BudgetLine{{AccountCode: "4000", Periods: budgetMonths("-1")}} })}, "budget_amount_negative", "lines")
	f.failCode(Command{Resource: "budgets", Action: "create", Budget: bad(func(x *Budget) { x.Lines = []BudgetLine{{AccountCode: "4000", Periods: []Amount{"1"}}} })}, "budget_periods_invalid", "lines")
	branchB2 := f.scope
	branchB2.Branch = "B2"
	if _, err := f.execute(branchB2, Command{Resource: "budgets", Action: "create", Budget: &b}); err == nil {
		t.Fatal("branch B2 session created a budget for branch B1")
	} else if u, ok := AsUserError(err); !ok || u.Code != "budget_branch_outside_session" {
		t.Fatalf("branch session error = %v", err)
	}

	// 1. create: header status defaults to open, 24 line rows (2 accounts × 12 months)
	created := f.run(Command{Resource: "budgets", Action: "create", Budget: &b})
	if created.ID != "BG-2569" || created.Version != 1 {
		t.Fatalf("create result %+v", created)
	}
	var lines int
	var total, status string
	if err := f.db.QueryRowContext(f.ctx, `SELECT count(*), sum(l.amount)::text, min(h.status) FROM gl_budget_lines l JOIN gl_budgets h ON h.company=l.company AND h.code=l.budget_code WHERE l.company='C' AND l.budget_code='BG-2569'`).Scan(&lines, &total, &status); err != nil || lines != 24 || total != "1803.00" || status != "open" {
		t.Fatalf("PG after create lines=%d total=%s status=%s err=%v", lines, total, status, err)
	}
	f.deny(f.scope, Command{Resource: "budgets", Action: "create", Budget: &b})

	// 2. read
	got := f.budget("BG-2569")
	if got.Status != "open" || !sameAmount(string(got.Total), "1803") || len(got.Lines) != 2 || !sameAmount(string(got.Lines[1].Total), "603") || !sameAmount(string(got.Lines[1].Periods[11]), "50.25") {
		t.Fatalf("get = %+v", got)
	}

	// 3. update: stale version refused; lines replaced in place, status closed (Champ flag, no lock)
	b.Lines = []BudgetLine{{AccountCode: "4000", Periods: budgetMonths("100")}, {AccountCode: "5000", Periods: budgetMonths("50")}}
	b.Status = "closed"
	f.failCode(Command{Resource: "budgets", Action: "update", ID: "BG-2569", Version: 99, Budget: &b}, CodeStaleVersion, "")
	f.deny(branchB2, Command{Resource: "budgets", Action: "update", ID: "BG-2569", Version: 1, Budget: &b})
	updated := f.run(Command{Resource: "budgets", Action: "update", ID: "BG-2569", Version: 1, Budget: &b})
	if err := f.db.QueryRowContext(f.ctx, `SELECT count(*), sum(l.amount)::text, min(h.status) FROM gl_budget_lines l JOIN gl_budgets h ON h.company=l.company AND h.code=l.budget_code WHERE l.company='C' AND l.budget_code='BG-2569'`).Scan(&lines, &total, &status); err != nil || lines != 24 || total != "1800.00" || status != "closed" || updated.Version != 2 {
		t.Fatalf("PG after update lines=%d total=%s status=%s version=%d err=%v", lines, total, status, updated.Version, err)
	}

	// 4. actuals: posted income 100 in branch B1 counts, the same in B2 must not; expense 1250.50 in B1
	f.post(f.draft("JV-B1", "B1"))
	f.post(f.draft("JV-B2", "B2"))
	expense := f.bookJournal("JV-EXP", "JV")
	er := f.run(Command{Resource: "journals", Action: "create", Journal: &expense})
	f.post(f.journal(er.ID))
	draftOnly := f.bookJournal("JV-DRAFT", "JV")
	f.run(Command{Resource: "journals", Action: "create", Journal: &draftOnly}) // drafts are not actuals

	report, err := f.store.Report(f.ctx, f.scope, "budgetcomparison", ReportQuery{FiscalYear: "2026", From: "2026-01-01", To: "2026-01-31"})
	if err != nil || len(report.Rows) != 2 {
		t.Fatalf("report %+v %v", report, err)
	}
	income, cost := report.Rows[0], report.Rows[1] // Champ order: account, then budget
	if income["accountcode"] != "4000" || !sameAmount(income["budgetamount"], "100") || !sameAmount(income["actualamount"], "100") || !sameAmount(income["variance"], "0") || income["percentused"] != "100.00" || income["budgetstatus"] != "closed" || income["branchcode"] != "B1" {
		t.Fatalf("income row = %v", income)
	}
	if cost["accountcode"] != "5000" || !sameAmount(cost["budgetamount"], "50") || !sameAmount(cost["actualamount"], "1250.50") || !sameAmount(cost["variance"], "-1200.50") {
		t.Fatalf("expense row = %v", cost)
	}
	if !sameAmount(report.Totals["budgetamount"], "150") || !sameAmount(report.Totals["actualamount"], "1350.50") || !sameAmount(report.Totals["variance"], "-1200.50") {
		t.Fatalf("totals = %v", report.Totals)
	}
	q1, err := f.store.Report(f.ctx, f.scope, "budgetcomparison", ReportQuery{FiscalYear: "2026", From: "2026-01-01", To: "2026-03-31", BudgetCode: "BG-2569", AccountCode: "4000"})
	if err != nil || len(q1.Rows) != 1 || !sameAmount(q1.Rows[0]["budgetamount"], "300") || !sameAmount(q1.Rows[0]["actualamount"], "100") {
		t.Fatalf("Q1 = %+v %v", q1.Rows, err)
	}
	if other, err := f.store.Report(f.ctx, f.scope, "budgetcomparison", ReportQuery{FiscalYear: "2026", BudgetCode: "BG-OTHER"}); err != nil || len(other.Rows) != 0 {
		t.Fatalf("unknown budget code = %+v %v", other.Rows, err)
	}
	if b2, err := f.store.Report(f.ctx, branchB2, "budgetcomparison", ReportQuery{FiscalYear: "2026"}); err != nil || len(b2.Rows) != 0 {
		t.Fatalf("branch B2 sees branch B1 budget = %+v %v", b2.Rows, err)
	}

	// 5. list is branch scoped like journals
	if page, err := f.store.List(f.ctx, f.scope, "budgets", "", 1, 10, ListFilter{}); err != nil || page.Total != 1 {
		t.Fatalf("list %+v %v", page, err)
	}
	if page, err := f.store.List(f.ctx, branchB2, "budgets", "", 1, 10, ListFilter{}); err != nil || page.Total != 0 {
		t.Fatalf("branch B2 list %+v %v", page, err)
	}

	// 6. delete removes header and lines; the audit events still rebuild the projection
	f.run(Command{Resource: "budgets", Action: "delete", ID: "BG-2569", Version: updated.Version})
	if err := f.db.QueryRowContext(f.ctx, `SELECT (SELECT count(*) FROM gl_budgets WHERE company='C') + (SELECT count(*) FROM gl_budget_lines WHERE company='C')`).Scan(&lines); err != nil || lines != 0 {
		t.Fatalf("delete left %d rows (%v)", lines, err)
	}
	var events int
	if err := f.db.QueryRowContext(f.ctx, `SELECT count(*) FROM gl_events WHERE company='C' AND EXISTS (SELECT 1 FROM jsonb_array_elements(payload->'changes') c WHERE c->>'kind'='budgets')`).Scan(&events); err != nil || events != 3 {
		t.Fatalf("budget audit events = %d (%v)", events, err)
	}
	if err := f.store.Projection().Rebuild(f.ctx, f.scope); err != nil {
		t.Fatalf("rebuild with budget events: %v", err)
	}

	// 7. a repeated requestid (double click / retry) returns the first result and writes once
	rid := uuid.NewString()
	first := f.run(Command{Resource: "budgets", Action: "create", RequestID: rid, Budget: &b})
	again := f.run(Command{Resource: "budgets", Action: "create", RequestID: rid, Budget: &b})
	if first.ID != again.ID || first.Version != again.Version || first.Sequence != again.Sequence {
		t.Fatalf("replay = %+v, first = %+v", again, first)
	}
	if err := f.db.QueryRowContext(f.ctx, `SELECT count(*) FROM gl_budgets WHERE company='C'`).Scan(&lines); err != nil || lines != 1 {
		t.Fatalf("replay wrote %d headers (%v)", lines, err)
	}
}

// Budgets live outside gl_records, yet they are references like journals: while a budget exists
// its fiscal year cannot be moved, shortened or deleted (period_no counts from the fiscal start)
// and its accounts cannot be deleted. Budgets are usually entered before the year has any
// journal, so FY 2027 of the fixture has none; the budget is its only reference.
func TestBudgetBlocksFiscalYearAndAccountChangesIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	acc := Account{AccountCode: "5100", AccountType: "expense", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "ค่าเช่าสำนักงานและคลังสินค้า"}}}
	ar := f.run(Command{Resource: "accounts", Action: "create", Account: &acc})
	b := Budget{Code: "BG-2570", Name: "งบค่าเช่าสำนักงานและคลังสินค้า ปี 2570", FiscalYear: "2027", Lines: []BudgetLine{{AccountCode: "5100", Periods: budgetMonths("20000")}}}
	created := f.run(Command{Resource: "budgets", Action: "create", Budget: &b})

	y := f.years["2027"]
	for name, mut := range map[string]func(*FiscalYear){
		"move start": func(x *FiscalYear) { x.StartDate, x.EndDate = "2027-04-01", "2028-03-31" },
		"shorten":    func(x *FiscalYear) { x.EndDate = "2027-09-30" },
		"deactivate": func(x *FiscalYear) { x.IsActive = false },
	} {
		next := y
		mut(&next)
		if _, err := f.execute(f.scope, Command{Resource: "fiscal-years", Action: "update", ID: y.ID, Version: y.Version, FiscalYear: &next}); err == nil || !strings.Contains(err.Error(), "ปีบัญชีมีข้อมูลอ้างอิง") {
			t.Fatalf("%s budgeted fiscal year: %v", name, err)
		}
	}
	if _, err := f.execute(f.scope, Command{Resource: "fiscal-years", Action: "delete", ID: y.ID, Version: y.Version}); err == nil || !strings.Contains(err.Error(), "ปีบัญชีมีข้อมูลอ้างอิง") {
		t.Fatalf("delete budgeted fiscal year: %v", err)
	}
	f.failCode(Command{Resource: "accounts", Action: "delete", ID: ar.ID, Version: ar.Version}, CodeReferenced, "")

	// PostgreSQL: the year keeps its dates and the account is still live
	var start, end string
	var liveAccounts int
	if err := f.db.QueryRowContext(f.ctx, `SELECT payload->>'startdate', payload->>'enddate' FROM gl_records WHERE company='C' AND kind='fiscal-years' AND code='2027' AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`).Scan(&start, &end); err != nil || start != "2027-01-01" || end != "2027-12-31" {
		t.Fatalf("FY 2027 after refused changes = %s..%s (%v)", start, end, err)
	}
	if err := f.db.QueryRowContext(f.ctx, `SELECT count(*) FROM gl_records WHERE company='C' AND kind='accounts' AND code='5100' AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`).Scan(&liveAccounts); err != nil || liveAccounts != 1 {
		t.Fatalf("account 5100 live rows = %d (%v)", liveAccounts, err)
	}

	// the budget was the only reference: once it is deleted the account and the year can go
	f.run(Command{Resource: "budgets", Action: "delete", ID: "BG-2570", Version: created.Version})
	f.run(Command{Resource: "accounts", Action: "delete", ID: ar.ID, Version: ar.Version})
	f.run(Command{Resource: "fiscal-years", Action: "delete", ID: y.ID, Version: y.Version})
}

// Action "spread" follows the fiscal year's real period count, so a short first year gets its
// whole annual amount in its own periods and the result can be saved unchanged.
func TestBudgetSpreadShortFiscalYearIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	y := FiscalYear{Code: "2028", StartDate: "2028-04-01", EndDate: "2028-12-31", IsActive: true, Scale: 2, ProfitLossAccount: "3200", RetainedEarningsAccount: "3100"}
	f.run(Command{Resource: "fiscal-years", Action: "create", FiscalYear: &y})
	spread := func(year string) (Result, error) {
		return f.execute(f.scope, Command{Resource: "budgets", Action: "spread", Budget: &Budget{FiscalYear: year, Lines: []BudgetLine{{AccountCode: "5000", Total: "90000"}}}})
	}

	r, err := spread("2028")
	if err != nil || len(r.Lines) != 1 {
		t.Fatalf("spread 9-period year = %+v, %v", r, err)
	}
	p := r.Lines[0].Periods
	if p[0] != "10000.00" || p[8] != "10000.00" || p[9] != "0.00" || p[11] != "0.00" {
		t.Fatalf("9-period spread = %v", p)
	}
	b := Budget{Code: "BG-2571", Name: "งบค่าใช้จ่ายในการขายและบริหาร ปีบัญชีแรก", FiscalYear: "2028", Lines: []BudgetLine{{AccountCode: "5000", Periods: p}}}
	f.run(Command{Resource: "budgets", Action: "create", Budget: &b})
	var total string
	if err := f.db.QueryRowContext(f.ctx, `SELECT sum(amount)::text FROM gl_budget_lines WHERE company='C' AND budget_code='BG-2571'`).Scan(&total); err != nil || total != "90000.00" {
		t.Fatalf("PG saved spread total = %s (%v)", total, err)
	}

	if r, err := spread(""); err != nil || r.Lines[0].Periods[11] != "7500.00" {
		t.Fatalf("spread without a fiscal year = %+v, %v", r, err)
	}
	if _, err := spread("2099"); err == nil {
		t.Fatal("spread accepted an unknown fiscal year")
	} else if u, ok := AsUserError(err); !ok || u.Code != "budget_fiscal_year_not_found" {
		t.Fatalf("unknown fiscal year error = %v", err)
	}
}
