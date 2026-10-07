//go:build integration

package generalledger

import (
	"context"
	"strings"
	"testing"
)

// reportOptionsFixture: ผังบัญชีหลายระดับ (บัญชีคุม 1 > 11 > 1000/1100, 2 > 2000, 4 > 4000) + ใบสำคัญที่ผ่านรายการแล้ว
//
//	OPEN  2026-01-01 opening  Dr 1000 / Cr 3000  1,000
//	Z1    2026-01-10          Dr 2000 / Cr 1000     50  (2000 ไปกลับในเดือนมกราคม → ยอดเป็นศูนย์)
//	Z2    2026-01-11          Dr 1000 / Cr 2000     50
//	SALE  2026-02-01          Dr 1100 / Cr 4000    300  (อ้างอิง INV-1)
//	COST  2026-03-01          Dr 5000 / Cr 1000    100
func reportOptionsFixture(t *testing.T) (*Postgres, Scope) {
	t.Helper()
	p, _ := glpgTestDB(t)
	company := "A"
	changes := []Change{}
	for _, row := range []struct {
		code, parent, kind, normal string
		level                      int
		posting                    bool
	}{
		{"1", "", "asset", "debit", 1, false}, {"11", "1", "asset", "debit", 2, false}, {"1000", "11", "asset", "debit", 3, true}, {"1100", "11", "asset", "debit", 3, true},
		{"2", "", "liability", "credit", 1, false}, {"2000", "2", "liability", "credit", 2, true},
		{"3000", "", "equity", "credit", 1, true},
		{"4", "", "income", "credit", 1, false}, {"4000", "4", "income", "credit", 2, true},
		{"5000", "", "expense", "debit", 1, true},
	} {
		a := Account{Identity: glpgIdentity("account-"+row.code, company, 1), AccountCode: row.code, Names: []Name{{Code: "th", Name: "บัญชี " + row.code}}, AccountType: row.kind, ParentAccountCode: row.parent, NormalBalance: row.normal, AllowPosting: row.posting, IsActive: true, Level: row.level}
		changes = append(changes, glpgChange(t, "accounts", a.ID, a.AccountCode, a))
	}
	f := FiscalYear{Identity: glpgIdentity("year-2026", company, 1), Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true, Scale: 2, RetainedEarningsAccount: "3000"}
	book := Master{Identity: glpgIdentity("book-JV", company, 1), Kind: "journal-books", Code: "JV", Name: "สมุดรายวันทั่วไป", BookType: BookTypeGeneral, IsActive: true}
	changes = append(changes, glpgChange(t, "fiscal-years", f.ID, f.Code, f), glpgChange(t, "journal-books", book.ID, book.Code, book))
	if err := p.Project(context.Background(), glpgEvent(company, 1, changes...)); err != nil {
		t.Fatal(err)
	}
	sale := glpgJournal(company, "SALE", "2026-02-01", "manual", "posted", "1100", "4000", "300")
	sale.Reference = "INV-1"
	for i, j := range []Journal{
		glpgJournal(company, "OPEN", "2026-01-01", "opening", "posted", "1000", "3000", "1000"),
		glpgJournal(company, "Z1", "2026-01-10", "manual", "posted", "2000", "1000", "50"),
		glpgJournal(company, "Z2", "2026-01-11", "manual", "posted", "1000", "2000", "50"),
		sale,
		glpgJournal(company, "COST", "2026-03-01", "manual", "posted", "5000", "1000", "100"),
	} {
		glpgSendJournal(t, p, int64(i+2), j)
	}
	return p, Scope{Holding: "H", Company: company}
}

func reportCodes(report Report) []string {
	out := make([]string, len(report.Rows))
	for i, row := range report.Rows {
		out[i] = row["accountcode"]
		if row["rowtype"] == "control" {
			out[i] += "*"
		}
	}
	return out
}

func TestReportAccountRangeFiltersEveryAccountReport(t *testing.T) {
	p, scope := reportOptionsFixture(t)
	ctx := context.Background()
	q := ReportQuery{FiscalYear: "2026", AccountFrom: "1000", AccountTo: "1100", Limit: 1000}
	ledger, err := p.Report(ctx, scope, "ledger", q)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range ledger.Rows {
		if row["accountcode"] != "1000" && row["accountcode"] != "1100" {
			t.Fatalf("ledger row outside range: %v", row)
		}
	}
	lines := 0
	for _, row := range ledger.Rows {
		if row["rowtype"] == "line" {
			lines++
		}
	}
	// 1000: Z1, Z2, COST · 1100: SALE — OPEN อยู่ในแถว "ยอดยกมา" ไม่ใช่รายการเคลื่อนไหว
	if lines != 4 || ledger.TotalRows != 10 { // + ยอดยกมา/ยอดยกไป/ยอดรวม ต่อบัญชี (Champ 5534)
		t.Fatalf("ledger rows in range = %v", ledger.Rows)
	}
	trial, err := p.Report(ctx, scope, "trialbalance", ReportQuery{FiscalYear: "2026", AccountFrom: "2000"})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(reportCodes(trial), ","); got != "2000,3000,4000,5000" {
		t.Fatalf("trial balance from 2000 = %s", got)
	}
	journal, err := p.Report(ctx, scope, "gljournal", ReportQuery{FiscalYear: "2026", AccountTo: "1100"})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range journal.Rows {
		if row["accountcode"] > "1100" {
			t.Fatalf("journal row above range end: %v", row)
		}
	}
	if _, err = p.Report(ctx, scope, "trialbalance", ReportQuery{FiscalYear: "2026", AccountFrom: "5", AccountTo: "1"}); err == nil {
		t.Fatal("reversed account range accepted")
	} else {
		setUserError(t, err, "report_account_range_invalid")
	}
	// daily-check and budget comparison share r.args ($10/$11) — they must still run with a range
	for _, name := range []string{"daily-check", "budgetcomparison", "annual-balances", "cashflow"} {
		if _, err = p.Report(ctx, scope, name, ReportQuery{FiscalYear: "2026", AccountFrom: "1000", AccountTo: "1100"}); err != nil {
			t.Fatalf("%s with account range: %v", name, err)
		}
	}
}

func TestTrialBalanceControlTotalsAndHideZero(t *testing.T) {
	p, scope := reportOptionsFixture(t)
	ctx := context.Background()
	plain, err := p.Report(ctx, scope, "trialbalance", ReportQuery{FiscalYear: "2026"})
	if err != nil {
		t.Fatal(err)
	}
	tree, err := p.Report(ctx, scope, "trialbalance", ReportQuery{FiscalYear: "2026", ControlTotals: true})
	if err != nil {
		t.Fatal(err)
	}
	// control total rows follow their descendants (Champ GLRepTrialBalanceView) and nest by level
	if got := strings.Join(reportCodes(tree), ","); got != "1000,1100,11*,1*,2000,2*,3000,4000,4*,5000" {
		t.Fatalf("control order = %s", got)
	}
	for key, want := range plain.Totals {
		glpgDecimalEqual(t, tree.Totals[key], want) // control rows never count twice
	}
	var control11, control1 map[string]string
	for _, row := range tree.Rows {
		switch {
		case row["rowtype"] == "control" && row["accountcode"] == "11":
			control11 = row
		case row["rowtype"] == "control" && row["accountcode"] == "1":
			control1 = row
		}
	}
	if control11 == nil || control1 == nil || control11["accountname"] != "บัญชี 11" || control11["accountlevel"] != "2" || control1["accountlevel"] != "1" {
		t.Fatalf("control rows = %v / %v", control11, control1)
	}
	glpgDecimalEqual(t, control11["endingdebit"], "1200") // 1000: 900 + 1100: 300
	glpgDecimalEqual(t, control1["endingdebit"], "1200")
	glpgDecimalEqual(t, control11["openingdebit"], "1000")

	// From 2026-02-01: 2000 nets to zero in January and has no movement after → hidden; others stay
	query := ReportQuery{FiscalYear: "2026", From: "2026-02-01", HideZero: true, ControlTotals: true}
	hidden, err := p.Report(ctx, scope, "workingpaper", query)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(reportCodes(hidden), ","); got != "1000,1100,11*,1*,3000,4000,4*,5000" {
		t.Fatalf("hide zero = %s", got)
	}
	query.HideZero = false
	shown, err := p.Report(ctx, scope, "workingpaper", query)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(reportCodes(shown), ","), "2000,2*") {
		t.Fatalf("zero account missing without the option: %v", reportCodes(shown))
	}
	glpgDecimalEqual(t, hidden.Totals["endingdebit"], shown.Totals["endingdebit"])
}

func TestLedgerHideOpeningAndSummary(t *testing.T) {
	p, scope := reportOptionsFixture(t)
	ctx := context.Background()
	base := ReportQuery{FiscalYear: "2026", From: "2026-02-01", AccountFrom: "1000", AccountTo: "1000"}
	normal, err := p.Report(ctx, scope, "ledger", base)
	if err != nil {
		t.Fatal(err)
	}
	// Champ 5534: ยอดยกมา → รายการ → ยอดยกไป (ลงฝั่งตรงข้ามให้ดุล) → ยอดรวม
	if len(normal.Rows) != 4 || normal.Rows[0]["rowtype"] != "opening" || normal.Rows[1]["rowtype"] != "line" || normal.Rows[2]["rowtype"] != "carry" || normal.Rows[3]["rowtype"] != "total" {
		t.Fatalf("ledger 1000 from February = %v", normal.Rows)
	}
	glpgDecimalEqual(t, normal.Rows[0]["debit"], "1000")
	glpgDecimalEqual(t, normal.Rows[0]["balance"], "1000")
	glpgDecimalEqual(t, normal.Rows[1]["balance"], "900")
	glpgDecimalEqual(t, normal.Rows[2]["credit"], "900")
	glpgDecimalEqual(t, normal.Rows[3]["debit"], "1000")
	glpgDecimalEqual(t, normal.Rows[3]["credit"], "1000")
	glpgDecimalEqual(t, normal.Totals["credit"], "100") // แถวพิเศษไม่นับในยอดรวมรายงาน
	glpgDecimalEqual(t, normal.Totals["opening"], "1000")

	base.HideOpening = true
	hidden, err := p.Report(ctx, scope, "ledger", base)
	if err != nil {
		t.Fatal(err)
	}
	for _, column := range hidden.Columns {
		if column.Key == "opening" {
			t.Fatal("opening column shown with ไม่แสดงยอดยกมา")
		}
	}
	if _, ok := hidden.Totals["opening"]; ok {
		t.Fatalf("opening total shown: %v", hidden.Totals)
	}
	glpgDecimalEqual(t, hidden.Rows[0]["balance"], "-100") // running balance starts at zero (Champ 5534)
	if len(hidden.Rows) != 3 || hidden.Rows[0]["rowtype"] != "line" {
		t.Fatalf("hide opening rows = %v", hidden.Rows)
	}
	glpgDecimalEqual(t, hidden.Rows[1]["debit"], "100") // ยอดยกไป
	glpgDecimalEqual(t, hidden.Totals["balance"], "-100")

	// 5535: balance as of the end date per account, one side only, control totals by level
	summary, err := p.Report(ctx, scope, "ledger", ReportQuery{FiscalYear: "2026", To: "2026-12-31", Summary: true, ControlTotals: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(reportCodes(summary), ","); got != "1000,1100,11*,1*,2000,2*,3000,4000,4*,5000" {
		t.Fatalf("summary order = %s", got)
	}
	want := map[string][2]string{"1000": {"900", "0"}, "1100": {"300", "0"}, "11": {"1200", "0"}, "2000": {"0", "0"}, "3000": {"0", "1000"}, "4000": {"0", "300"}, "5000": {"100", "0"}}
	for _, row := range summary.Rows {
		if amounts, ok := want[row["accountcode"]]; ok {
			glpgDecimalEqual(t, row["debit"], amounts[0])
			glpgDecimalEqual(t, row["credit"], amounts[1])
		}
	}
	glpgDecimalEqual(t, summary.Totals["debit"], "1300")
	glpgDecimalEqual(t, summary.Totals["credit"], "1300")
}

func TestJournalSummaryOneRowPerDocument(t *testing.T) {
	p, scope := reportOptionsFixture(t)
	report, err := p.Report(context.Background(), scope, "gljournal", ReportQuery{FiscalYear: "2026", Summary: true})
	if err != nil {
		t.Fatal(err)
	}
	docs := make([]string, len(report.Rows))
	for i, row := range report.Rows {
		docs[i] = row["docno"]
		if row["journalid"] == "" {
			t.Fatalf("summary row without journal id for drill-down: %v", row)
		}
		if row["docno"] == "SALE" {
			if row["referenceno"] != "INV-1" || row["description"] != "รายการทดสอบ" || row["bookcode"] != "JV" {
				t.Fatalf("SALE summary row = %v", row)
			}
			glpgDecimalEqual(t, row["amount"], "300")
		}
	}
	if got := strings.Join(docs, ","); got != "Z1,Z2,SALE,COST" { // opening entries stay out like the detail report
		t.Fatalf("summary documents = %s", got)
	}
	glpgDecimalEqual(t, report.Totals["amount"], "500")
}
