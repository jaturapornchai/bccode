package generalledger

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestSpreadAnnualRemainderToLastMonth(t *testing.T) {
	for _, tc := range []struct{ annual, month, last string }{
		{"100000", "8333.33", "8333.37"},
		{"1200", "100", "100"},
		{"0.05", "0", "0.05"},
		{"0", "0", "0"},
	} {
		months := SpreadAnnual(decimal.RequireFromString(tc.annual))
		sum := decimal.Zero
		for i, m := range months {
			sum = sum.Add(m)
			want := tc.month
			if i == BudgetPeriods-1 {
				want = tc.last
			}
			if !m.Equal(decimal.RequireFromString(want)) {
				t.Fatalf("%s month %d = %s, want %s", tc.annual, i+1, m, want)
			}
		}
		if !sum.Equal(decimal.RequireFromString(tc.annual)) {
			t.Fatalf("%s months add to %s", tc.annual, sum)
		}
	}
}

func TestSpreadCommandValidatesAmounts(t *testing.T) {
	r, err := spreadBudget(Command{Budget: &Budget{Lines: []BudgetLine{{AccountCode: "5000", Total: "100000"}}}})
	if err != nil || len(r.Lines) != 1 || r.Lines[0].Periods[0] != "8333.33" || r.Lines[0].Periods[11] != "8333.37" || r.Lines[0].Total != "100000.00" {
		t.Fatalf("spread = %+v, %v", r, err)
	}
	for amount, code := range map[Amount]string{"-1": "budget_amount_negative", "1.005": "budget_amount_scale", "10000000000000000": "budget_amount_too_large"} {
		_, err := spreadBudget(Command{Budget: &Budget{Lines: []BudgetLine{{Total: amount}}}})
		if u, ok := AsUserError(err); !ok || u.Code != code {
			t.Fatalf("%s: %v, want %s", amount, err, code)
		}
	}
}

func TestFiscalPeriodStarts(t *testing.T) {
	got := fiscalPeriodStarts(FiscalYear{StartDate: "2026-04-15", EndDate: "2026-12-31"})
	if len(got) != 9 || got[0] != "2026-04-15" || got[1] != "2026-05-01" || got[8] != "2026-12-01" {
		t.Fatalf("short year periods = %v", got)
	}
	if p := budgetPeriodsInRange(FiscalYear{StartDate: "2026-01-01", EndDate: "2026-12-31"}, "2026-01-01", "2026-03-31"); len(p) != 3 || p[2] != 3 {
		t.Fatalf("Q1 periods = %v", p)
	}
}

// Champ BCGLBudget.Status is an open/closed flag: blank means open, case and spaces ignored.
func TestNormalizeBudgetStatus(t *testing.T) {
	for in, want := range map[string]string{"": "open", " Closed ": "closed", "OPEN": "open", "approved": "approved"} {
		b := Budget{Status: in}
		normalizeBudget(&b)
		if b.Status != want {
			t.Fatalf("status %q -> %q, want %q", in, b.Status, want)
		}
	}
}
