package fixedasset

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	gl "smlcloudplatform/internal/generalledger"
)

func TestPickFiscalYearUsesTheYearOfTheVoucherDate(t *testing.T) {
	years := []gl.FiscalYear{
		{Code: "2568", StartDate: "2025-04-01", EndDate: "2026-03-31", IsActive: true, Closed: true},
		{Code: "2569", StartDate: "2026-04-01", EndDate: "2027-03-31", IsActive: true},
		{Code: "2570", StartDate: "2027-04-01", EndDate: "2028-03-31", IsActive: false},
		{Code: "2571", StartDate: "2028-04-01", EndDate: "2029-03-31", IsActive: true, Identity: gl.Identity{IsDeleted: true}},
	}
	for _, tc := range []struct {
		date, want, code string
	}{
		{"2026-04-01", "2569", ""}, // first day of the year
		{"2026-12-31", "2569", ""}, // ค.ศ. 2026 December still belongs to the April year
		{"2027-01-31", "2569", ""},
		{"2027-03-31", "2569", ""}, // last day of the year
		{"2026-03-31", "", codeFiscalYearClosed},
		{"2027-04-30", "", codeFiscalYearClosed}, // inactive
		{"2028-06-30", "", codeFiscalYearNotFound},
		{"2031-01-31", "", codeFiscalYearNotFound},
	} {
		got, err := pickFiscalYear(years, tc.date, "date")
		if tc.code == "" {
			if err != nil || got != tc.want {
				t.Fatalf("%s: got %q, %v; want %q", tc.date, got, err, tc.want)
			}
			continue
		}
		user, ok := gl.AsUserError(err)
		if !ok || user.Code != tc.code || user.Field != "date" || got != "" {
			t.Fatalf("%s: got %q, %#v; want %s on date", tc.date, got, user, tc.code)
		}
	}

	// Overlapping years are refused, never resolved to the first row.
	overlap := []gl.FiscalYear{
		{Code: "2569", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true},
		{Code: "2569A", StartDate: "2026-06-01", EndDate: "2027-05-31", IsActive: true},
	}
	_, err := pickFiscalYear(overlap, "2026-07-31", "disposaldate")
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeFiscalYearAmbiguous || user.Field != "disposaldate" {
		t.Fatalf("overlap: got %#v", err)
	}
	if got, err := pickFiscalYear(overlap, "2026-05-31", "date"); err != nil || got != "2569" {
		t.Fatalf("outside the overlap: got %q, %v", got, err)
	}
}

func TestDepreciationVoucherDateDefaultsToPeriodEnd(t *testing.T) {
	for _, tc := range []struct {
		year      string
		period    int
		date      string
		wantYear  string
		wantDate  string
		periodEnd string
	}{
		{"2026", 12, "", "2026", "2026-12-31", "2026-12-31"},
		{" 2026 ", 2, "", "2026", "2026-02-28", "2026-02-28"},
		{"2028", 2, "", "2028", "2028-02-29", "2028-02-29"}, // leap year
		{"2026", 7, "2026-07-15", "2026", "2026-07-15", "2026-07-31"},
	} {
		year, date, periodEnd, err := depreciationVoucherDate(tc.year, tc.period, tc.date)
		if err != nil || year != tc.wantYear || date != tc.wantDate || periodEnd != tc.periodEnd {
			t.Fatalf("%q/%d/%q: got %q %q %q %v", tc.year, tc.period, tc.date, year, date, periodEnd, err)
		}
	}
	// The schedule is keyed by ค.ศ.: a พ.ศ. or malformed year is refused on the year field.
	for _, bad := range []string{"2569", "abc", "1899", "26"} {
		_, _, _, err := depreciationVoucherDate(bad, 7, "")
		if user, ok := gl.AsUserError(err); !ok || user.Code != codePostYearInvalid || user.Field != "fiscalyear" {
			t.Fatalf("%q: got %#v", bad, err)
		}
	}
	_, _, _, err := depreciationVoucherDate("2026", 7, "31/07/2026")
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeDateInvalid || user.Field != "date" {
		t.Fatalf("bad date: %#v", err)
	}
}

func TestDepreciationFiscalYearKeepsThePeriodInItsOwnYear(t *testing.T) {
	years := []gl.FiscalYear{
		{Code: "2569", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true, Closed: true},
		{Code: "2570", StartDate: "2027-01-01", EndDate: "2027-12-31", IsActive: true},
	}
	if got, err := depreciationFiscalYear(years, "2027", 1, "2027-01-31", "2027-01-31"); err != nil || got != "2570" {
		t.Fatalf("January 2027: %q %v", got, err)
	}
	// December 2026 posted on 10 January 2027: never filed in 2570, and 2569 is closed.
	_, err := depreciationFiscalYear(years, "2026", 12, "2027-01-10", "2026-12-31")
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeFiscalYearClosed || user.Field != "fiscalyear" {
		t.Fatalf("closed period year: %#v", err)
	}
	years[0].Closed = false
	_, err = depreciationFiscalYear(years, "2026", 12, "2027-01-10", "2026-12-31")
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeDateOutsidePeriodYear || user.Field != "date" {
		t.Fatalf("date in another year: %#v", err)
	}
	if user, _ := gl.AsUserError(err); !strings.Contains(user.Message, "2569") || !strings.Contains(user.Message, "2570") || !strings.Contains(user.Message, "2026-12-31") {
		t.Fatalf("message must name both years and the period end: %s", user.Message)
	}
	if got, err := depreciationFiscalYear(years, "2026", 12, "2026-12-15", "2026-12-31"); err != nil || got != "2569" {
		t.Fatalf("date inside the period year: %q %v", got, err)
	}
	_, err = depreciationFiscalYear(years, "2027", 12, "2028-01-05", "2027-12-31")
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeFiscalYearNotFound || user.Field != "date" {
		t.Fatalf("date without a year: %#v", err)
	}
}

func TestCheckVoucherDate(t *testing.T) {
	if err := checkVoucherDate("2026-07-31", "date"); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"31/07/2026", "2026-02-30", "2026-7-31", "2569-07-31x", ""} {
		user, ok := gl.AsUserError(checkVoucherDate(bad, "date"))
		if !ok || user.Code != codeDateInvalid || user.Field != "date" {
			t.Fatalf("%q: got %#v", bad, user)
		}
	}
}

// yearOnlyLedger serves the fiscal-year list and fails the test on any write.
type yearOnlyLedger struct {
	t      *testing.T
	years  []gl.FiscalYear
	listed []string
}

func (l *yearOnlyLedger) Execute(context.Context, gl.Scope, gl.Command) (gl.Result, error) {
	l.t.Fatal("a refused voucher must not reach the GL engine")
	return gl.Result{}, nil
}

func (l *yearOnlyLedger) List(_ context.Context, _ gl.Scope, resource, _ string, _, _ int, _ gl.ListFilter) (gl.Page, error) {
	l.listed = append(l.listed, resource)
	page := gl.Page{}
	for _, y := range l.years {
		raw, err := json.Marshal(y)
		if err != nil {
			l.t.Fatal(err)
		}
		page.Items = append(page.Items, raw)
	}
	return page, nil
}

// The year check runs before the schedule, the branch or the ledger is touched (no database here),
// so a voucher dated outside every open fiscal year is refused with a field the screen can point at.
func TestPostingRefusesVoucherDateOutsideOpenFiscalYears(t *testing.T) {
	ctx := context.Background()
	scope := Scope{Holding: "RUNGRUENG", Company: "01", Branch: "00000", Actor: "fa-unit"}
	now := time.Date(2026, 9, 25, 3, 0, 0, 0, time.UTC)
	ledger := &yearOnlyLedger{t: t, years: []gl.FiscalYear{
		{Code: "2569", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true, Closed: true},
	}}
	poster := NewGLPoster(nil, ledger, nil)

	_, err := poster.PostDepreciation(ctx, scope, "2026", 7, "31/07/2026", "", "", now)
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeDateInvalid || user.Field != "date" {
		t.Fatalf("bad date: %#v", err)
	}
	if len(ledger.listed) != 0 {
		t.Fatalf("a malformed date must be refused before the fiscal years are read, listed %v", ledger.listed)
	}
	_, err = poster.PostDepreciation(ctx, scope, "2569", 7, "", "", "", now)
	if user, ok := gl.AsUserError(err); !ok || user.Code != codePostYearInvalid || user.Field != "fiscalyear" {
		t.Fatalf("พ.ศ. year: %#v", err)
	}
	if len(ledger.listed) != 0 {
		t.Fatalf("a bad year must be refused before the fiscal years are read, listed %v", ledger.listed)
	}
	_, err = poster.PostDepreciation(ctx, scope, "2026", 7, "", "", "", now)
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeFiscalYearClosed || user.Field != "fiscalyear" {
		t.Fatalf("closed year: %#v", err)
	}
	_, err = poster.PostDepreciation(ctx, scope, "2027", 1, "2027-01-31", "", "", now)
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeFiscalYearNotFound || user.Field != "fiscalyear" {
		t.Fatalf("no year: %#v", err)
	}
	_, _, err = poster.DisposeAsset(ctx, scope, AssetDisposal{AssetCode: "FA-001", DisposalDate: "2027-02-15", DisposalType: "write_off"}, now)
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeFiscalYearNotFound || user.Field != "disposaldate" {
		t.Fatalf("disposal without a year: %#v", err)
	}
	_, _, err = poster.DisposeAsset(ctx, scope, AssetDisposal{AssetCode: "FA-001", DisposalDate: "2026/02/15", DisposalType: "write_off"}, now)
	if user, ok := gl.AsUserError(err); !ok || user.Code != codeDateInvalid || user.Field != "disposaldate" {
		t.Fatalf("disposal bad date: %#v", err)
	}
	for _, resource := range ledger.listed {
		if resource != "fiscal-years" {
			t.Fatalf("only the fiscal years may be read before refusing, read %q", resource)
		}
	}
}
