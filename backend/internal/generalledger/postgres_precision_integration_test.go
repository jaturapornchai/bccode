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
