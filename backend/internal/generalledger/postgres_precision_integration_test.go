//go:build integration

package generalledger

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestPostgresDecimal128BoundaryAndBSONEventReplay(t *testing.T) {
	p, db := glpgTestDB(t)
	ctx := context.Background()
	glpgSeed(t, p, "A")
	scope := Scope{Holding: "H", Company: "A"}
	data, err := p.Get(ctx, scope, "fiscal-years", "year-2026")
	if err != nil {
		t.Fatal(err)
	}
	var year FiscalYear
	if err = json.Unmarshal(data, &year); err != nil {
		t.Fatal(err)
	}
	year.Version = 2
	year.Scale = 8
	if err = p.Project(ctx, glpgEvent("A", 2, glpgChange(t, "fiscal-years", year.ID, year.Code, year))); err != nil {
		t.Fatal(err)
	}
	amount := "99999999999999999999999999.12345678"
	j := glpgJournal("A", "LARGE", "2026-01-02", "manual", "posted", "1000", "4000", amount)
	event := glpgEvent("A", 3, glpgChange(t, "journals", j.ID, j.DocNo, j))
	event.OccurredAt = time.Date(2026, 1, 2, 0, 0, 0, 123456789, time.UTC)
	if err = p.Project(ctx, event); err != nil {
		t.Fatal(err)
	}
	// BSON Date has millisecond precision even though Go time can have nanos.
	event.Delivered = true
	event.OccurredAt = event.OccurredAt.Truncate(time.Millisecond)
	if err = p.Project(ctx, event); err != nil {
		t.Fatalf("duplicate projection: %v", err)
	}
	var saved string
	if err = db.QueryRow(`SELECT debit::text FROM gl_lines WHERE company='A' AND account_code='1000'`).Scan(&saved); err != nil {
		t.Fatal(err)
	}
	if saved != amount {
		t.Fatalf("rounded boundary: %s", saved)
	}
	report, err := p.Report(ctx, scope, "pnl", ReportQuery{FiscalYear: "2026"})
	if err != nil {
		t.Fatal(err)
	}
	glpgDecimalEqual(t, report.Totals["profit"], amount)
	for _, statement := range []string{`UPDATE gl_events SET event_hash='changed' WHERE company='A'`, `DELETE FROM gl_events WHERE company='A'`, `TRUNCATE gl_events`} {
		if _, err = db.Exec(statement); err == nil {
			t.Fatalf("audit mutation accepted: %s", statement)
		}
	}
	if err = p.Rebuild(ctx, scope); err != nil {
		t.Fatalf("rebuild after nanos event: %v", err)
	}
}

func TestPostgresListFiltersMasterCRUDAndTenantIsolation(t *testing.T) {
	p, _ := glpgTestDB(t)
	ctx := context.Background()
	glpgSeed(t, p, "A")
	scope := Scope{Holding: "H", Company: "A", Branch: "B1"}
	b1 := glpgJournal("A", "B1-DRAFT", "2026-01-02", "manual", "draft", "1000", "4000", "1")
	glpgSendJournal(t, p, 2, b1)
	b2 := glpgJournal("A", "B2-DRAFT", "2026-01-02", "manual", "draft", "1000", "4000", "2")
	b2.BranchCode = "B2"
	glpgSendJournal(t, p, 3, b2)
	uv := glpgJournal("A", "B1-UV", "2026-01-02", "manual", "posted", "1000", "4000", "3")
	uv.BookCode = "UV"
	glpgSendJournal(t, p, 4, uv)
	page, err := p.List(ctx, scope, "journals", "", 1, 1, ListFilter{BookCode: "JV", Status: "draft", Kind: "manual"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("filter after pagination: %#v", page)
	}
	var found Journal
	if err = json.Unmarshal(page.Items[0], &found); err != nil || found.ID != b1.ID {
		t.Fatalf("wrong branch/filter %s %v", found.ID, err)
	}
	if _, err = p.Get(ctx, scope, "journals", b2.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross branch GET: %v", err)
	}
	page, err = p.List(ctx, scope, "journals", "' OR true --", 1, 100, ListFilter{})
	if err != nil || page.Total != 0 {
		t.Fatalf("search SQL injection: %#v %v", page, err)
	}
	if _, err = p.List(ctx, scope, "journals", "", 1, 10, ListFilter{Status: "invalid"}); err == nil {
		t.Fatal("invalid enum accepted")
	}
	plan := Master{Identity: glpgIdentity("plan", "A", 1), Kind: "forecast", Code: "PLAN", Name: "แผนรับเงิน", IsActive: true, Amount: "12.34", BranchCode: "B1", FiscalYear: "2026", StartDate: "2026-02-01", EndDate: "2026-02-01", Direction: "in"}
	if err = p.Project(ctx, glpgEvent("A", 5, glpgChange(t, "forecast", plan.ID, plan.Code, plan))); err != nil {
		t.Fatal(err)
	}
	data, err := p.Get(ctx, scope, "forecast", plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	var saved Master
	if err = json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Amount != "12.34" {
		t.Fatal(saved.Amount)
	}
	plan.Version = 2
	plan.Amount = "98.76"
	if err = p.Project(ctx, glpgEvent("A", 6, glpgChange(t, "forecast", plan.ID, plan.Code, plan))); err != nil {
		t.Fatal(err)
	}
	data, err = p.Get(ctx, scope, "forecast", plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.ID != "plan" || saved.Version != 2 || saved.Amount != "98.76" {
		t.Fatalf("master update %#v", saved)
	}
	if _, err = p.Get(ctx, Scope{Holding: "H", Company: "A", Branch: "B2"}, "forecast", plan.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross branch forecast: %v", err)
	}
	plan.Version = 3
	plan.IsDeleted = true
	if err = p.Project(ctx, glpgEvent("A", 7, glpgChange(t, "forecast", plan.ID, plan.Code, plan))); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Get(ctx, scope, "forecast", plan.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted master returned: %v", err)
	}
	page, err = p.List(ctx, scope, "forecast", "", 1, 100, ListFilter{})
	if err != nil || page.Total != 0 {
		t.Fatalf("deleted master in list: %#v %v", page, err)
	}
	if _, err = p.Get(ctx, scope, "journals", b1.ID); err != nil {
		t.Fatalf("master delete changed journal: %v", err)
	}
}

func TestPostgresListFiltersJournalDocumentDateRange(t *testing.T) {
	p, _ := glpgTestDB(t)
	ctx := context.Background()
	glpgSeed(t, p, "A")
	scope := Scope{Holding: "H", Company: "A", Branch: "B1"}
	for i, day := range []string{"2026-01-31", "2026-02-01", "2026-02-28", "2026-03-01"} {
		j := glpgJournal("A", "DR-"+day, day, "manual", "draft", "1000", "4000", "1")
		glpgSendJournal(t, p, int64(2+i), j)
	}
	posted := glpgJournal("A", "DR-POSTED", "2026-02-15", "manual", "posted", "1000", "4000", "2")
	glpgSendJournal(t, p, 6, posted)

	docnos := func(filter ListFilter, search string) []string {
		t.Helper()
		page, err := p.List(ctx, scope, "journals", search, 1, 100, filter)
		if err != nil {
			t.Fatalf("list %+v: %v", filter, err)
		}
		var out []string
		for _, raw := range page.Items {
			var j Journal
			if err = json.Unmarshal(raw, &j); err != nil {
				t.Fatal(err)
			}
			out = append(out, j.DocNo)
		}
		if int(page.Total) != len(out) {
			t.Fatalf("total %d differs from items %d for %+v", page.Total, len(out), filter)
		}
		return out
	}
	want := func(name string, got []string, expected ...string) {
		t.Helper()
		if len(got) != len(expected) {
			t.Fatalf("%s: got %v want %v", name, got, expected)
		}
		seen := map[string]bool{}
		for _, docno := range got {
			seen[docno] = true
		}
		for _, docno := range expected {
			if !seen[docno] {
				t.Fatalf("%s: got %v want %v", name, got, expected)
			}
		}
	}
	want("no range", docnos(ListFilter{}, ""), "DR-2026-01-31", "DR-2026-02-01", "DR-2026-02-28", "DR-2026-03-01", "DR-POSTED")
	want("from only is inclusive", docnos(ListFilter{DateFrom: "2026-02-01"}, ""), "DR-2026-02-01", "DR-2026-02-28", "DR-2026-03-01", "DR-POSTED")
	want("to only is inclusive", docnos(ListFilter{DateTo: "2026-02-28"}, ""), "DR-2026-01-31", "DR-2026-02-01", "DR-2026-02-28", "DR-POSTED")
	want("whole month", docnos(ListFilter{DateFrom: "2026-02-01", DateTo: "2026-02-28"}, ""), "DR-2026-02-01", "DR-2026-02-28", "DR-POSTED")
	want("one day", docnos(ListFilter{DateFrom: "2026-02-28", DateTo: "2026-02-28"}, ""), "DR-2026-02-28")
	want("no match", docnos(ListFilter{DateFrom: "2027-01-01"}, ""))
	want("with status", docnos(ListFilter{Status: "draft", DateFrom: "2026-02-01", DateTo: "2026-02-28"}, ""), "DR-2026-02-01", "DR-2026-02-28")
	want("with search", docnos(ListFilter{DateFrom: "2026-02-01", DateTo: "2026-02-28"}, "posted"), "DR-POSTED")

	page, err := p.List(ctx, scope, "journals", "", 2, 1, ListFilter{DateFrom: "2026-02-01", DateTo: "2026-02-28"})
	if err != nil || page.Total != 3 || len(page.Items) != 1 {
		t.Fatalf("range filter must apply before pagination: %#v %v", page, err)
	}

	for name, filter := range map[string]ListFilter{
		"from not a date":  {DateFrom: "2026-2-1"},
		"to not a date":    {DateTo: "' OR true --"},
		"impossible day":   {DateFrom: "2026-02-30"},
		"reversed range":   {DateFrom: "2026-03-01", DateTo: "2026-02-01"},
		"timestamp format": {DateTo: "2026-02-01T00:00:00Z"},
	} {
		_, err = p.List(ctx, scope, "journals", "", 1, 10, filter)
		var user *UserError
		if !errors.As(err, &user) || user.Status != 400 || user.Code != "journal_filter_date_invalid" && user.Code != "journal_filter_date_range_invalid" {
			t.Fatalf("%s: want a 400 journal_filter_date_* error, got %v", name, err)
		}
	}

	// The range is only a journal document-date filter; it must not hide other records.
	all, err := p.List(ctx, scope, "accounts", "", 1, 1000, ListFilter{})
	if err != nil || all.Total == 0 {
		t.Fatalf("accounts: %#v %v", all, err)
	}
	filtered, err := p.List(ctx, scope, "accounts", "", 1, 1000, ListFilter{DateFrom: "2099-01-01"})
	if err != nil || filtered.Total != all.Total {
		t.Fatalf("date range changed non-journal list: %d vs %d (%v)", filtered.Total, all.Total, err)
	}
}
