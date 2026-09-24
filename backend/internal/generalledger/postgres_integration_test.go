//go:build integration

package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func glpgTestDB(t *testing.T) (*Postgres, *sql.DB) {
	t.Helper()
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN to isolated PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "bc_gl_test_" + fmt.Sprintf("%x", uuid.New())
	if _, err = admin.Exec(`CREATE SCHEMA ` + pq.QuoteIdentifier(schema)); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, err := admin.Exec(`DROP SCHEMA ` + pq.QuoteIdentifier(schema) + ` CASCADE`)
		if err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	projection := NewPostgres(func(holding string) (*sql.DB, error) {
		if holding != "H" {
			return nil, fmt.Errorf("holding denied")
		}
		return db, nil
	})
	return projection, db
}

func glpgIdentity(id, company string, version int64) Identity {
	return Identity{ID: id, HoldingCode: "H", BusinessCode: company, Version: version}
}
func glpgChange(t *testing.T, kind, id, code string, value any) Change {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return Change{Kind: kind, ID: id, Code: code, Payload: string(payload)}
}
func glpgEvent(company string, sequence int64, changes ...Change) Event {
	return Event{ID: fmt.Sprintf("%s-%d", company, sequence), HoldingCode: "H", BusinessCode: company, Sequence: sequence, RequestID: fmt.Sprintf("request-%s-%d", company, sequence), Actor: "integration-test", Action: "test", OccurredAt: time.Date(2026, 1, 1, 0, 0, int(sequence), 0, time.UTC), Changes: changes}
}

func glpgSeed(t *testing.T, p *Postgres, company string) {
	t.Helper()
	changes := []Change{}
	for _, row := range []struct {
		code, kind, normal string
		cash               bool
	}{{"1000", "asset", "debit", true}, {"1100", "asset", "debit", false}, {"2000", "liability", "credit", false}, {"3000", "equity", "credit", false}, {"4000", "income", "credit", false}, {"5000", "expense", "debit", false}} {
		a := Account{Identity: glpgIdentity("account-"+row.code, company, 1), AccountCode: row.code, Names: []Name{{Code: "th", Name: "บัญชี " + row.code}}, AccountType: row.kind, NormalBalance: row.normal, AllowPosting: true, IsActive: true, IsCash: row.cash}
		changes = append(changes, glpgChange(t, "accounts", a.ID, a.AccountCode, a))
	}
	f := FiscalYear{Identity: glpgIdentity("year-2026", company, 1), Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true, Scale: 2, RetainedEarningsAccount: "3000"}
	changes = append(changes, glpgChange(t, "fiscal-years", f.ID, f.Code, f))
	// the store checks journal.bookcode against the journal-books master (glpgJournal uses JV)
	book := Master{Identity: glpgIdentity("book-JV", company, 1), Kind: "journal-books", Code: "JV", Name: "สมุดรายวันทั่วไป", BookType: BookTypeGeneral, IsActive: true}
	changes = append(changes, glpgChange(t, "journal-books", book.ID, book.Code, book))
	if err := p.Project(context.Background(), glpgEvent(company, 1, changes...)); err != nil {
		t.Fatal(err)
	}
}

func glpgJournal(company, id, date, kind, status, debitAccount, creditAccount, amount string) Journal {
	return Journal{Identity: glpgIdentity(id, company, 1), DocNo: id, Date: date, BookCode: "JV", FiscalYear: "2026", Description: "รายการทดสอบ", BranchCode: "B1", Kind: kind, Status: status, Lines: []Line{{AccountCode: debitAccount, Debit: Amount(amount), Credit: "0", Description: "เดบิต", DepartmentCode: "D1", ProjectCode: "P1", CashFlow: "operating"}, {AccountCode: creditAccount, Debit: "0", Credit: Amount(amount), Description: "เครดิต", DepartmentCode: "D1", ProjectCode: "P1", CashFlow: "operating"}}}
}
func glpgSendJournal(t *testing.T, p *Postgres, sequence int64, j Journal) {
	t.Helper()
	if err := p.Project(context.Background(), glpgEvent(j.BusinessCode, sequence, glpgChange(t, "journals", j.ID, j.DocNo, j))); err != nil {
		t.Fatal(err)
	}
}
func glpgDecimalEqual(t *testing.T, value, want string) {
	t.Helper()
	v, err := ParseAmount(value)
	if err != nil {
		t.Fatal(err)
	}
	w, err := ParseAmount(want)
	if err != nil {
		t.Fatal(err)
	}
	if !v.Decimal().Equal(w.Decimal()) {
		t.Fatalf("amount=%s want=%s", value, want)
	}
}

func TestPostgresProjectionSequenceAndImmutableLines(t *testing.T) {
	p, db := glpgTestDB(t)
	ctx := context.Background()
	glpgSeed(t, p, "A")
	scope := Scope{Holding: "H", Company: "A"}
	draft := glpgJournal("A", "DRAFT", "2026-01-02", "manual", "draft", "1000", "4000", "0.30")
	glpgSendJournal(t, p, 2, draft)
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM gl_lines`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("draft projected lines=%d err=%v", count, err)
	}
	draft.Version = 2
	draft.Status = "posted"
	event := glpgEvent("A", 3, glpgChange(t, "journals", draft.ID, draft.DocNo, draft))
	if err := p.Project(ctx, event); err != nil {
		t.Fatal(err)
	}
	event.Delivered = true
	if err := p.Project(ctx, event); err != nil {
		t.Fatalf("duplicate delivery: %v", err)
	}
	changed := event
	changed.Actor = "other"
	if err := p.Project(ctx, changed); err == nil {
		t.Fatal("accepted mismatched duplicate")
	}
	gap := glpgEvent("A", 5, glpgChange(t, "journals", draft.ID, draft.DocNo, draft))
	if err := p.Project(ctx, gap); err == nil {
		t.Fatal("accepted sequence gap")
	}
	stale := glpgEvent("A", 4, glpgChange(t, "journals", draft.ID, draft.DocNo, draft))
	if err := p.Project(ctx, stale); err == nil {
		t.Fatal("accepted stale record version")
	}
	tampered := draft
	tampered.Version = 3
	tampered.Lines = append([]Line{}, draft.Lines...)
	tampered.Lines[0].Debit = "1"
	tampered.Lines[1].Credit = "1"
	if err := p.Project(ctx, glpgEvent("A", 4, glpgChange(t, "journals", tampered.ID, tampered.DocNo, tampered))); err == nil {
		t.Fatal("accepted amendment of posted values")
	}
	version, err := p.Version(ctx, scope)
	if err != nil || version != 3 {
		t.Fatalf("failed events advanced watermark=%d err=%v", version, err)
	}
	draft.Version = 3
	draft.Status = "reversed"
	reversal := glpgJournal("A", "REV-1", "2026-01-03", "reversal", "posted", "4000", "1000", "0.30")
	reversal.ReversalOf = draft.ID
	if err = p.Project(ctx, glpgEvent("A", 4, glpgChange(t, "journals", draft.ID, draft.DocNo, draft), glpgChange(t, "journals", reversal.ID, reversal.DocNo, reversal))); err != nil {
		t.Fatal(err)
	}
	var amount string
	if err = db.QueryRow(`SELECT count(*),SUM(debit-credit)::text FROM gl_lines`).Scan(&count, &amount); err != nil || count != 4 {
		t.Fatalf("reversal lines=%d err=%v", count, err)
	}
	glpgDecimalEqual(t, amount, "0")
	trial, err := p.Report(ctx, scope, "trialbalance", ReportQuery{FiscalYear: "2026"})
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range trial.Rows {
		glpgDecimalEqual(t, row["balance"], "0")
	}
	glpgSeed(t, p, "B")
	other := glpgJournal("B", "OTHER", "2026-01-02", "manual", "posted", "1000", "4000", "9")
	glpgSendJournal(t, p, 2, other)
	if err = p.Rebuild(ctx, scope); err != nil {
		t.Fatal(err)
	}
	var aCount, bCount, eventCount int
	if err = db.QueryRow(`SELECT count(*) FILTER(WHERE company='A'),count(*) FILTER(WHERE company='B') FROM gl_lines`).Scan(&aCount, &bCount); err != nil || aCount != 4 || bCount != 2 {
		t.Fatalf("rebuild cross tenant: %d %d %v", aCount, bCount, err)
	}
	if err = db.QueryRow(`SELECT count(*) FROM gl_events WHERE company='A'`).Scan(&eventCount); err != nil || eventCount != 4 {
		t.Fatalf("rebuild destroyed audit %d %v", eventCount, err)
	}
	var recordPayload []byte
	if err = db.QueryRow(`SELECT payload FROM gl_records WHERE company='A' AND kind='journals' AND id='DRAFT'`).Scan(&recordPayload); err != nil {
		t.Fatal(err)
	}
	var restored Journal
	if err = json.Unmarshal(recordPayload, &restored); err != nil || restored.Status != "reversed" {
		t.Fatalf("record restore: %v %#v", err, restored)
	}
}

func TestPostgresAccountingReportsExactAndScoped(t *testing.T) {
	p, db := glpgTestDB(t)
	ctx := context.Background()
	glpgSeed(t, p, "A")
	scope := Scope{Holding: "H", Company: "A"}
	glpgSendJournal(t, p, 2, glpgJournal("A", "OPEN", "2026-01-01", "opening", "posted", "1000", "3000", "1000"))
	sale := glpgJournal("A", "SALE", "2026-01-02", "manual", "posted", "1000", "4000", "0.30")
	sale.Lines = []Line{{AccountCode: "1000", Debit: "0.1", Credit: "0", DepartmentCode: "D1", ProjectCode: "P1", CashFlow: "operating"}, {AccountCode: "1000", Debit: "0.2", Credit: "0", DepartmentCode: "D1", ProjectCode: "P1", CashFlow: "operating"}, {AccountCode: "4000", Debit: "0", Credit: "0.3", DepartmentCode: "D1", ProjectCode: "P1"}}
	glpgSendJournal(t, p, 3, sale)
	glpgSendJournal(t, p, 4, glpgJournal("A", "COST", "2026-01-03", "manual", "posted", "5000", "1000", "10"))
	unclassified := glpgJournal("A", "OTHER", "2026-01-04", "manual", "posted", "1000", "4000", "2")
	unclassified.Lines[0].CashFlow = ""
	glpgSendJournal(t, p, 5, unclassified)
	glpgSendJournal(t, p, 6, glpgJournal("A", "DRAFT", "2026-01-05", "manual", "draft", "1000", "4000", "9999"))
	b2 := glpgJournal("A", "BRANCH2", "2026-01-05", "manual", "posted", "1000", "4000", "20")
	b2.BranchCode = "B2"
	glpgSendJournal(t, p, 7, b2)
	forecast := Master{Identity: glpgIdentity("plan-1", "A", 1), Kind: "forecast", Code: "PLAN1", Name: "รับเงินแผน", IsActive: true, FiscalYear: "2026", StartDate: "2026-02-01", Direction: "in", Amount: "100", BranchCode: "B1"}
	if err := p.Project(ctx, glpgEvent("A", 8, glpgChange(t, "forecast", forecast.ID, forecast.Code, forecast))); err != nil {
		t.Fatal(err)
	}
	query := ReportQuery{FiscalYear: "2026", BranchCode: "B1", From: "2026-01-01", To: "2026-12-31"}
	var report Report
	var err error
	for _, name := range []string{"gljournal", "ledger", "trialbalance", "workingpaper", "pnl", "balancesheet", "annual-balances", "daily-check", "cashflow", "cashflowforecast", "project-pnl", "dimensionpnl", "projectsummary", "dashboard", "executivesummary", "financialgraphs"} {
		t.Run(name, func(t *testing.T) {
			report, err = p.Report(ctx, scope, name, query)
			if err != nil {
				t.Fatal(err)
			}
			if report.TotalRows == 0 {
				t.Fatal("empty report")
			}
			if name == "gljournal" || name == "ledger" {
				for _, row := range report.Rows {
					if row["journalid"] == "" {
						t.Fatal("missing source journal ID")
					}
					if _, exists := row["__line_no"]; exists {
						t.Fatal("leaked private sort key")
					}
				}
				for _, column := range report.Columns {
					if column.Key == "journalid" || column.Key == "__line_no" {
						t.Fatal("metadata became visible column")
					}
				}
			}
			if report.AsOf != "2026-12-31" {
				t.Fatal(report.AsOf)
			}
			if len(report.Rows) > query.Limit && query.Limit > 0 {
				t.Fatal("page overflow")
			}
		})
	}
	pnl, err := p.Report(ctx, scope, "pnl", query)
	if err != nil {
		t.Fatal(err)
	}
	glpgDecimalEqual(t, pnl.Totals["revenue"], "2.3")
	glpgDecimalEqual(t, pnl.Totals["expense"], "10")
	glpgDecimalEqual(t, pnl.Totals["profit"], "-7.7")
	trial, err := p.Report(ctx, scope, "trialbalance", query)
	if err != nil {
		t.Fatal(err)
	}
	glpgDecimalEqual(t, trial.Totals["difference"], "0")
	glpgDecimalEqual(t, trial.Totals["openingdebit"], "1000")
	glpgDecimalEqual(t, trial.Totals["debit"], "12.3")
	bs, err := p.Report(ctx, scope, "balancesheet", query)
	if err != nil {
		t.Fatal(err)
	}
	glpgDecimalEqual(t, bs.Totals["assets"], "992.3")
	glpgDecimalEqual(t, bs.Totals["equity"], "992.3")
	glpgDecimalEqual(t, bs.Totals["difference"], "0")
	cash, err := p.Report(ctx, scope, "cashflow", query)
	if err != nil {
		t.Fatal(err)
	}
	glpgDecimalEqual(t, cash.Totals["opening"], "1000")
	glpgDecimalEqual(t, cash.Totals["closing"], "992.3")
	if len(cash.Warnings) != 1 || cash.Totals["unclassifiedlines"] != "1" {
		t.Fatalf("missing unclassified warning %#v", cash)
	}
	ledgerQuery := query
	ledgerQuery.AccountCode = "1000"
	ledgerQuery.Limit = 1
	ledgerQuery.Page = 2
	ledger, err := p.Report(ctx, scope, "ledger", ledgerQuery)
	if err != nil {
		t.Fatal(err)
	}
	if ledger.TotalRows != 4 || len(ledger.Rows) != 1 {
		t.Fatalf("pagination total=%d rows=%d", ledger.TotalRows, len(ledger.Rows))
	}
	glpgDecimalEqual(t, ledger.Rows[0]["balance"], "1000.3")
	feb := query
	feb.From = "2026-02-01"
	feb.To = "2026-02-28"
	future, err := p.Report(ctx, scope, "cashflowforecast", feb)
	if err != nil {
		t.Fatal(err)
	}
	glpgDecimalEqual(t, future.Totals["opening"], "992.3")
	glpgDecimalEqual(t, future.Totals["closing"], "1092.3")
	if _, err = p.Report(ctx, Scope{Holding: "H", Company: "A", Branch: "B2"}, "pnl", query); err == nil {
		t.Fatal("cross branch report accepted")
	}
	if _, err = p.Report(ctx, Scope{Holding: "H", Company: "B"}, "pnl", query); err == nil {
		t.Fatal("cross company year accepted")
	}
	if _, err = p.Report(ctx, scope, "pnl", ReportQuery{FiscalYear: "2026", From: "2025-12-01"}); err == nil {
		t.Fatal("cross fiscal year accepted")
	}
	// Closing removes current earnings from the balance sheet but must not
	// erase the year's revenue/expense from its performance statement.
	close := glpgJournal("A", "CLOSE", "2026-12-31", "closing", "posted", "4000", "3000", "2.3")
	close.Lines = append(close.Lines, Line{AccountCode: "3000", Debit: "10", Credit: "0"}, Line{AccountCode: "5000", Debit: "0", Credit: "10"})
	glpgSendJournal(t, p, 9, close)
	pnl, err = p.Report(ctx, scope, "pnl", query)
	if err != nil {
		t.Fatal(err)
	}
	glpgDecimalEqual(t, pnl.Totals["profit"], "-7.7")
	bs, err = p.Report(ctx, scope, "balancesheet", query)
	if err != nil {
		t.Fatal(err)
	}
	glpgDecimalEqual(t, bs.Totals["currentearnings"], "0")
	glpgDecimalEqual(t, bs.Totals["difference"], "0")
	var typ string
	if err = db.QueryRow(`SELECT pg_typeof(debit)::text FROM gl_lines LIMIT 1`).Scan(&typ); err != nil || typ != "numeric" {
		t.Fatalf("accounting storage=%s err=%v", typ, err)
	}
}
