//go:build integration

package generalledger

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestPostgresProcessBalancesKeepDimensionsAndRecalculateAudit(t *testing.T) {
	p, db := glpgTestDB(t)
	ctx := context.Background()
	glpgSeed(t, p, "A")
	scope := Scope{Holding: "H", Company: "A"}
	glpgSendJournal(t, p, 2, glpgJournal("A", "OPEN", "2026-01-01", "opening", "posted", "1000", "3000", "1000"))
	glpgSendJournal(t, p, 3, glpgJournal("A", "SALE-P1", "2026-01-02", "manual", "posted", "1000", "4000", "100"))
	reverse := glpgJournal("A", "RETURN-P2", "2026-01-03", "manual", "posted", "4000", "1000", "40")
	for i := range reverse.Lines {
		reverse.Lines[i].DepartmentCode = "D2"
		reverse.Lines[i].ProjectCode = "P2"
	}
	glpgSendJournal(t, p, 4, reverse)
	branch2 := glpgJournal("A", "SALE-B2", "2026-01-04", "manual", "posted", "1000", "4000", "200")
	branch2.BranchCode = "B2"
	for i := range branch2.Lines {
		branch2.Lines[i].DepartmentCode = "D2"
		branch2.Lines[i].ProjectCode = "P3"
	}
	glpgSendJournal(t, p, 5, branch2)
	glpgSendJournal(t, p, 6, glpgJournal("A", "DRAFT", "2026-01-05", "manual", "draft", "1000", "4000", "9999"))
	snapshot, err := p.ProcessBalances(ctx, scope, "2026", "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Sequence != 6 || snapshot.FiscalYear != "2026" || snapshot.Currency != "THB" || snapshot.Scale != 2 {
		t.Fatalf("snapshot metadata %#v", snapshot)
	}
	balances := map[string]string{}
	for _, row := range snapshot.Rows {
		balances[row.BranchCode+"/"+row.DepartmentCode+"/"+row.ProjectCode+"/"+row.AccountCode] = row.Balance
	}
	for key, want := range map[string]string{"B1/D1/P1/4000": "-100", "B1/D2/P2/4000": "40", "B2/D2/P3/4000": "-200", "B1/D1/P1/1000": "1100"} {
		value, ok := balances[key]
		if !ok {
			t.Fatalf("lost dimension %s: %#v", key, balances)
		}
		glpgDecimalEqual(t, value, want)
	}
	if _, err = p.ProcessBalances(ctx, Scope{Holding: "H", Company: "A", Branch: "B1"}, "2026", "2026-12-31"); err == nil {
		t.Fatal("accepted partial-company process balances")
	}
	if _, err = p.ProcessBalances(ctx, scope, "2026", "2027-01-01"); err == nil {
		t.Fatal("accepted balances beyond fiscal year")
	}
	bad := glpgEvent("A", 7)
	bad.Action = "journals:post"
	if err = p.Project(ctx, bad); err == nil {
		t.Fatal("accepted empty changes outside recalculate")
	}
	recalculate := glpgEvent("A", 7)
	recalculate.Action = "processes:recalculate"
	recalculate.Changes = []Change{}
	if err = p.Project(ctx, recalculate); err != nil {
		t.Fatal(err)
	}
	recalculate.Delivered = true
	recalculate.Changes = nil
	if err = p.Project(ctx, recalculate); err != nil {
		t.Fatalf("empty replay: %v", err)
	}
	if err = p.Rebuild(ctx, scope); err != nil {
		t.Fatalf("rebuild empty audit event: %v", err)
	}
	after, err := p.ProcessBalances(ctx, scope, "2026", "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if after.Sequence != 7 || !reflect.DeepEqual(after.Rows, snapshot.Rows) {
		t.Fatalf("recalculate changed accounting balances: %#v", after)
	}
	var audits int
	if err = db.QueryRow(`SELECT count(*) FROM gl_events WHERE company='A'`).Scan(&audits); err != nil || audits != 7 {
		t.Fatalf("recalculate audit missing: %d %v", audits, err)
	}
	// Close each income residual in its original dimension. A company/account
	// aggregate would lose the opposing P1 and P2 balances here.
	closing1 := glpgJournal("A", "CLOSE-B1", "2026-12-31", "closing", "posted", "4000", "3000", "100")
	closing1.Lines = append(closing1.Lines, Line{AccountCode: "3000", Debit: "40", Credit: "0", DepartmentCode: "D2", ProjectCode: "P2"}, Line{AccountCode: "4000", Debit: "0", Credit: "40", DepartmentCode: "D2", ProjectCode: "P2"})
	glpgSendJournal(t, p, 8, closing1)
	closing2 := glpgJournal("A", "CLOSE-B2", "2026-12-31", "closing", "posted", "4000", "3000", "200")
	closing2.BranchCode = "B2"
	for i := range closing2.Lines {
		closing2.Lines[i].DepartmentCode = "D2"
		closing2.Lines[i].ProjectCode = "P3"
	}
	glpgSendJournal(t, p, 9, closing2)
	closed, err := p.ProcessBalances(ctx, scope, "2026", "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	branchTotals := map[string]Amount{"B1": "0", "B2": "0"}
	for _, row := range closed.Rows {
		if row.AccountType == "income" || row.AccountType == "expense" {
			t.Fatalf("uncleared dimension %#v", row)
		}
		value, err := ParseAmount(row.Balance)
		if err != nil {
			t.Fatal(err)
		}
		branchTotals[row.BranchCode] = amountFromDecimal(branchTotals[row.BranchCode].Decimal().Add(value.Decimal()))
	}
	for branch, total := range branchTotals {
		if !total.Decimal().IsZero() {
			t.Fatalf("unbalanced carry-forward %s=%s", branch, total)
		}
	}
	data, err := p.Get(ctx, scope, "fiscal-years", "year-2026")
	if err != nil {
		t.Fatal(err)
	}
	var year FiscalYear
	if err = json.Unmarshal(data, &year); err != nil {
		t.Fatal(err)
	}
	year.Identity = glpgIdentity("year-2027", "A", 1)
	year.Code = "2027"
	year.StartDate = "2027-01-01"
	year.EndDate = "2027-12-31"
	if err = p.Project(ctx, glpgEvent("A", 10, glpgChange(t, "fiscal-years", year.ID, year.Code, year))); err != nil {
		t.Fatal(err)
	}
	for index, branch := range []string{"B1", "B2"} {
		opening := glpgJournal("A", "OPEN-2027-"+branch, "2027-01-01", "opening", "posted", "1000", "3000", "1")
		opening.FiscalYear = "2027"
		opening.BranchCode = branch
		opening.Lines = []Line{}
		for _, row := range closed.Rows {
			if row.BranchCode != branch {
				continue
			}
			value, err := ParseAmount(row.Balance)
			if err != nil {
				t.Fatal(err)
			}
			line := Line{AccountCode: row.AccountCode, DepartmentCode: row.DepartmentCode, ProjectCode: row.ProjectCode}
			if value.Decimal().IsNegative() {
				line.Credit = amountFromDecimal(value.Decimal().Abs())
			} else {
				line.Debit = value
			}
			opening.Lines = append(opening.Lines, line)
		}
		glpgSendJournal(t, p, int64(11+index), opening)
	}
	carried, err := p.ProcessBalances(ctx, scope, "2027", "2027-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(carried.Rows, closed.Rows) {
		t.Fatalf("carry-forward lost dimensions: before=%#v after=%#v", closed.Rows, carried.Rows)
	}
	if err = p.Rebuild(ctx, scope); err != nil {
		t.Fatal(err)
	}
	replayed, err := p.ProcessBalances(ctx, scope, "2027", "2027-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Sequence != 12 || !reflect.DeepEqual(replayed.Rows, carried.Rows) {
		t.Fatal("replay changed dimension balances")
	}
}
