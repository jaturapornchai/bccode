//go:build integration

package generalledger

import (
	"context"
	"testing"
)

type periodProfitWant struct{ code, net, summ string }

func periodProfitFixture(t *testing.T) (*Postgres, Scope) {
	t.Helper()
	p, _ := glpgTestDB(t)
	glpgSeed(t, p, "A") // fiscal year 2026, retained earnings account 3000, accounts 1000..5000
	glpgSendJournal(t, p, 2, glpgJournal("A", "OPEN", "2026-01-01", "opening", "posted", "1000", "3000", "5000"))
	glpgSendJournal(t, p, 3, glpgJournal("A", "SALE-JAN", "2026-01-10", "manual", "posted", "1000", "4000", "1000"))
	glpgSendJournal(t, p, 4, glpgJournal("A", "COST-JAN", "2026-01-20", "manual", "posted", "5000", "1000", "300"))
	branch2 := glpgJournal("A", "SALE-JAN-B2", "2026-01-25", "manual", "posted", "1000", "4000", "50")
	branch2.BranchCode = "B2"
	glpgSendJournal(t, p, 5, branch2)
	glpgSendJournal(t, p, 6, glpgJournal("A", "DRAFT-JAN", "2026-01-26", "manual", "draft", "1000", "4000", "9999"))
	glpgSendJournal(t, p, 7, glpgJournal("A", "SALE-FEB", "2026-02-05", "manual", "posted", "1000", "4000", "100"))
	glpgSendJournal(t, p, 8, glpgJournal("A", "COST-FEB", "2026-02-12", "manual", "posted", "5000", "1000", "400"))
	glpgSendJournal(t, p, 9, glpgJournal("A", "DIVIDEND", "2026-02-15", "manual", "posted", "3000", "1000", "200"))
	// Year closing: revenue 1150 and expense 700 are cleared into retained earnings (profit 450).
	closing := glpgJournal("A", "CLOSE", "2026-03-31", "closing", "posted", "4000", "5000", "0")
	closing.Lines = []Line{
		{AccountCode: "4000", Debit: "1150", Credit: "0", Description: "close revenue", DepartmentCode: "D1", ProjectCode: "P1", CashFlow: "operating"},
		{AccountCode: "5000", Debit: "0", Credit: "700", Description: "close expense", DepartmentCode: "D1", ProjectCode: "P1", CashFlow: "operating"},
		{AccountCode: "3000", Debit: "0", Credit: "450", Description: "profit to retained earnings", DepartmentCode: "D1", ProjectCode: "P1", CashFlow: "operating"},
	}
	glpgSendJournal(t, p, 10, closing)
	periods := []Change{}
	for _, row := range []struct{ code, start, end string }{{"2026-01", "2026-01-01", "2026-01-31"}, {"2026-02", "2026-02-01", "2026-02-28"}, {"2026-03", "2026-03-01", "2026-03-31"}, {"2026-04", "2026-04-01", "2026-04-30"}} {
		period := Master{Identity: glpgIdentity("period-"+row.code, "A", 1), Kind: "periods", Code: row.code, FiscalYear: "2026", StartDate: row.start, EndDate: row.end, IsActive: true}
		periods = append(periods, glpgChange(t, "periods", period.ID, period.Code, period))
	}
	// a period of another fiscal year must never show up
	other := Master{Identity: glpgIdentity("period-2027-01", "A", 1), Kind: "periods", Code: "2027-01", FiscalYear: "2027", StartDate: "2027-01-01", EndDate: "2027-01-31", IsActive: true}
	periods = append(periods, glpgChange(t, "periods", other.ID, other.Code, other))
	if err := p.Project(context.Background(), glpgEvent("A", 11, periods...)); err != nil {
		t.Fatal(err)
	}
	return p, Scope{Holding: "H", Company: "A"}
}

func requirePeriodProfit(t *testing.T, got PeriodProfitResult, want []periodProfitWant) {
	t.Helper()
	if len(got.Periods) != len(want) {
		t.Fatalf("periods = %+v, want %d rows", got.Periods, len(want))
	}
	for i, row := range want {
		if got.Periods[i].Code != row.code {
			t.Fatalf("row %d code = %s, want %s", i, got.Periods[i].Code, row.code)
		}
		glpgDecimalEqual(t, string(got.Periods[i].NetProfitLoss), row.net)
		glpgDecimalEqual(t, string(got.Periods[i].SummProfitLoss), row.summ)
	}
}

// Champ 5511 columns: profit is positive and loss negative (revenue - expense); the accumulated column
// carries the retained earnings forward, ignores drafts and opening/closing journals as profit, and a
// branch session only sees its own branch.
func TestPostgresPeriodProfitNetAndAccumulated(t *testing.T) {
	p, scope := periodProfitFixture(t)
	ctx := context.Background()

	all, err := p.PeriodProfit(ctx, scope, "2026")
	if err != nil {
		t.Fatal(err)
	}
	if all.FiscalYear != "2026" {
		t.Fatalf("fiscal year = %q", all.FiscalYear)
	}
	// Jan: profit 1000+50-300; the 9999 draft is ignored. Feb: loss 100-400. Mar/Apr: no activity, and the
	// March closing journal moves nothing (counting it would add 450 to March and April).
	requirePeriodProfit(t, all, []periodProfitWant{
		{"2026-01", "750", "5750"},
		{"2026-02", "-300", "5250"},
		{"2026-03", "0", "5250"},
		{"2026-04", "0", "5250"},
	})

	b1, err := p.PeriodProfit(ctx, Scope{Holding: "H", Company: "A", Branch: "B1"}, "2026")
	if err != nil {
		t.Fatal(err)
	}
	requirePeriodProfit(t, b1, []periodProfitWant{
		{"2026-01", "700", "5700"},
		{"2026-02", "-300", "5200"},
		{"2026-03", "0", "5200"},
		{"2026-04", "0", "5200"},
	})
	b2, err := p.PeriodProfit(ctx, Scope{Holding: "H", Company: "A", Branch: "B2"}, "2026")
	if err != nil {
		t.Fatal(err)
	}
	requirePeriodProfit(t, b2, []periodProfitWant{
		{"2026-01", "50", "50"},
		{"2026-02", "0", "50"},
		{"2026-03", "0", "50"},
		{"2026-04", "0", "50"},
	})

	for _, year := range []string{"", "2099"} {
		empty, err := p.PeriodProfit(ctx, scope, year)
		if err != nil || len(empty.Periods) != 0 {
			t.Fatalf("year %q: %+v, %v; want no periods", year, empty, err)
		}
	}
	other, err := p.PeriodProfit(ctx, Scope{Holding: "H", Company: "OTHER"}, "2026")
	if err != nil || len(other.Periods) != 0 {
		t.Fatalf("other company: %+v, %v; want no periods", other, err)
	}
}

// Without a retained earnings account on the fiscal year the accumulated column is the year-to-date profit.
func TestPostgresPeriodProfitWithoutRetainedAccount(t *testing.T) {
	p, scope := periodProfitFixture(t)
	year := FiscalYear{Identity: glpgIdentity("year-2026", "A", 2), Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true, Scale: 2}
	if err := p.Project(context.Background(), glpgEvent("A", 12, glpgChange(t, "fiscal-years", year.ID, year.Code, year))); err != nil {
		t.Fatal(err)
	}
	got, err := p.PeriodProfit(context.Background(), scope, "2026")
	if err != nil {
		t.Fatal(err)
	}
	requirePeriodProfit(t, got, []periodProfitWant{
		{"2026-01", "750", "750"},
		{"2026-02", "-300", "450"},
		{"2026-03", "0", "450"},
		{"2026-04", "0", "450"},
	})
}
