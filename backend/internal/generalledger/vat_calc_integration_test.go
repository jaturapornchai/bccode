//go:build integration

package generalledger

import "testing"

// GET /gl/v2/vat-calc rounds to the voucher's fiscal-year scale; a blank or unknown year falls back to 2 decimals.
func TestPostgresFiscalYearScale(t *testing.T) {
	f := newPGIntegrityFixture(t)
	y := FiscalYear{Code: "2028", StartDate: "2028-01-01", EndDate: "2028-12-31", IsActive: true, Scale: 0, ProfitLossAccount: "3200", RetainedEarningsAccount: "3100"}
	f.run(Command{Resource: "fiscal-years", Action: "create", FiscalYear: &y})
	for code, want := range map[string]int{"2026": 2, "2028": 0, " 2028 ": 0, "": 2, "1999": 2} {
		if got, err := f.store.FiscalYearScale(f.ctx, f.scope, code); err != nil || got != want {
			t.Fatalf("FiscalYearScale(%q) = %d, %v; want %d", code, got, err, want)
		}
	}
}
