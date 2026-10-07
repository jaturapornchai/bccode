//go:build integration

package generalledger

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// Champ report dialogs pick several departments/projects (GLRepGLDlg.cpp: DepartCode IN (...)); the query keeps the
// single param and takes a comma-separated list.
func TestReportDepartmentProjectMultiSelectIntegration(t *testing.T) {
	p, scope := reportOptionsFixture(t) // every fixture journal is on department D1 / project P1
	ctx := context.Background()
	dimJournal := func(id, date, amount, department, project string) Journal {
		j := glpgJournal("A", id, date, "manual", "posted", "1000", "4000", amount)
		for i := range j.Lines {
			j.Lines[i].DepartmentCode, j.Lines[i].ProjectCode = department, project
		}
		return j
	}
	for i, j := range []Journal{
		dimJournal("DP-22", "2026-04-01", "20", "D2", "P2"),
		dimJournal("DP-32", "2026-04-02", "40", "D3", "P2"),
		dimJournal("DP-N", "2026-04-03", "80", "", ""),
	} {
		glpgSendJournal(t, p, int64(7+i), j)
	}
	docNos := func(name string, q ReportQuery) []string {
		t.Helper()
		q.FiscalYear, q.AccountCode, q.Limit = "2026", "4000", 1000
		report, err := p.Report(ctx, scope, name, q)
		if err != nil {
			t.Fatalf("%s %+v: %v", name, q, err)
		}
		var out []string
		for _, row := range report.Rows {
			if row["docno"] != "" {
				out = append(out, row["docno"])
			}
		}
		sort.Strings(out)
		return out
	}
	all := []string{"DP-22", "DP-32", "DP-N", "SALE"}
	for _, tc := range []struct {
		name        string
		query       ReportQuery
		wantLedger  []string
		wantJournal []string
	}{
		{"no filter", ReportQuery{}, all, all},
		{"one department", ReportQuery{DepartmentCode: "D2"}, []string{"DP-22"}, []string{"DP-22"}},
		{"two departments", ReportQuery{DepartmentCode: "D2,D3"}, []string{"DP-22", "DP-32"}, []string{"DP-22", "DP-32"}},
		{"messy list is trimmed and deduplicated", ReportQuery{DepartmentCode: " D3 , D2,D2,, "}, []string{"DP-22", "DP-32"}, []string{"DP-22", "DP-32"}},
		{"department without entries", ReportQuery{DepartmentCode: "D9"}, nil, nil},
		{"one project", ReportQuery{ProjectCode: "P1"}, []string{"SALE"}, []string{"SALE"}},
		{"two projects", ReportQuery{ProjectCode: "P1,P2"}, []string{"DP-22", "DP-32", "SALE"}, []string{"DP-22", "DP-32", "SALE"}},
		{"departments and projects combine with AND", ReportQuery{DepartmentCode: "D1,D2", ProjectCode: "P2,P9"}, []string{"DP-22"}, []string{"DP-22"}},
	} {
		if got := docNos("ledger", tc.query); !reflect.DeepEqual(got, tc.wantLedger) {
			t.Errorf("ledger %s: %v, want %v", tc.name, got, tc.wantLedger)
		}
		if got := docNos("daily-check", tc.query); !reflect.DeepEqual(got, tc.wantJournal) {
			t.Errorf("daily-check %s: %v, want %v", tc.name, got, tc.wantJournal)
		}
	}

	// the totals of the other filtered reports move with the multi-select too: trial balance of account 4000, period credit
	trial := func(q ReportQuery) string {
		t.Helper()
		q.FiscalYear, q.AccountCode, q.Limit = "2026", "4000", 1000
		report, err := p.Report(ctx, scope, "trialbalance", q)
		if err != nil {
			t.Fatal(err)
		}
		return report.Totals["credit"]
	}
	if got := trial(ReportQuery{DepartmentCode: "D2,D3"}); !sameAmount(got, "60") {
		t.Errorf("trial balance D2,D3 period credit = %q, want 60", got)
	}
	if got := trial(ReportQuery{ProjectCode: "P2", DepartmentCode: "D3"}); !sameAmount(got, "40") {
		t.Errorf("trial balance P2+D3 period credit = %q, want 40", got)
	}
}

func TestBudgetReportsDepartmentProjectMultiSelectIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	create := func(b Budget) { f.run(Command{Resource: "budgets", Action: "create", Budget: &b}) }
	for _, b := range []struct{ code, department, project string }{{"BG-1", "D1", "P1"}, {"BG-2", "D2", ""}, {"BG-3", "D3", "P2"}} {
		create(Budget{Code: b.code, Name: "งบ " + b.code, FiscalYear: "2026", DepartmentCode: b.department, ProjectCode: b.project, Lines: []BudgetLine{{AccountCode: "4000", Periods: budgetMonths("10")}}})
	}
	budgetCodes := func(name string, q ReportQuery) string {
		t.Helper()
		q.FiscalYear, q.From, q.To = "2026", "2026-01-01", "2026-12-31"
		report, err := f.store.Report(f.ctx, f.scope, name, q)
		if err != nil {
			t.Fatalf("%s %+v: %v", name, q, err)
		}
		var out []string
		for _, row := range report.Rows {
			if row["rowtype"] == "budgetline" {
				out = append(out, row["budgetcode"])
			}
		}
		return strings.Join(out, ",")
	}
	for _, name := range []string{"budgetdetail", "budgetcomparison"} {
		for _, tc := range []struct {
			query ReportQuery
			want  string
		}{
			{ReportQuery{}, "BG-1,BG-2,BG-3"},
			{ReportQuery{DepartmentCode: "D2"}, "BG-2"},
			{ReportQuery{DepartmentCode: "D1,D3"}, "BG-1,BG-3"},
			{ReportQuery{ProjectCode: "P1,P2"}, "BG-1,BG-3"},
			{ReportQuery{DepartmentCode: "D1,D2,D3", ProjectCode: "P2"}, "BG-3"},
			{ReportQuery{DepartmentCode: "D9"}, ""},
		} {
			if got := budgetCodes(name, tc.query); got != tc.want {
				t.Errorf("%s %+v = %q, want %q", name, tc.query, got, tc.want)
			}
		}
	}
}

func TestReportDimensionFilterTooManyCodesIntegration(t *testing.T) {
	p, scope := reportOptionsFixture(t)
	codes := make([]string, maxReportDimensionCodes+1)
	for i := range codes {
		codes[i] = fmt.Sprintf("D%03d", i)
	}
	_, err := p.Report(context.Background(), scope, "ledger", ReportQuery{FiscalYear: "2026", DepartmentCode: strings.Join(codes, ",")})
	user, ok := err.(*UserError)
	if !ok || user.Code != "report_dimension_too_many" || user.Field != "departmentcode" || len(user.Args) != 2 || user.Args[0] != "200" || user.Args[1] != "201" {
		t.Fatalf("201 departments = %#v", err)
	}
	if _, err = p.Report(context.Background(), scope, "ledger", ReportQuery{FiscalYear: "2026", ProjectCode: strings.Join(codes[:200], ",")}); err != nil {
		t.Fatalf("200 projects rejected: %v", err)
	}
}
