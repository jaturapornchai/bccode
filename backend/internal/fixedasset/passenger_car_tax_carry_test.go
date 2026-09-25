package fixedasset

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	gl "smlcloudplatform/internal/generalledger"

	"github.com/shopspring/decimal"
)

// passengerCarTaxInYear from the stored schedule (CalculateSchedule) without a database. What the
// yearly ceiling holds back is deducted in later years, also after the book items end, until the
// capped basis is used up; nothing after a disposal — docs/kms/21-thai-tax-form-references.md §13.
func TestPassengerCarTaxCarryForward(t *testing.T) {
	for _, tc := range []struct {
		name               string
		cost, scrap, rate  string
		life               int
		purchase, disposal string
		want               map[int]string // years not listed must be 0
		wantTotal          string
	}{
		{
			// UAT 2026-09-25: before the carry-forward the tax stopped with the book items
			// (666,202.11 in total). Allowable total = min(capped share of all book depreciation,
			// 1,000,000) = 1,499,999.00 × 1M/1.5M = 999,999.33 (the book keeps its 1-baht scrap).
			name: "UAT 1.5M, 3-year, rate 33.33, scrap 1, bought 1 Jul 2026", cost: "1500000.00", scrap: "1.00", rate: "33.33",
			life: 3, purchase: "2026-07-01",
			want: map[int]string{2026: "100821.92", 2027: "200000.00", 2028: "200000.00", 2029: "200000.00",
				2030: "200000.00", 2031: "99177.41"},
			wantTotal: "999999.33",
		},
		{
			// 2028 has 366 days: 200,000 × 306/366; the whole 1,000,000 is used by 2033.
			name: "leap year, bought 1 Mar 2028", cost: "1500000.00", scrap: "0.00", rate: "0",
			life: 3, purchase: "2028-03-01",
			want: map[int]string{2028: "167213.11", 2029: "200000.00", 2030: "200000.00", 2031: "200000.00",
				2032: "200000.00", 2033: "32786.89"},
			wantTotal: "1000000.00",
		},
		{
			// Sold 31 Mar 2027: 2027 stops at the disposal date (200,000 × 90/365) and nothing is
			// carried past the sale, although the stored book items run on to 2029.
			name: "sold mid-life", cost: "1500000.00", scrap: "0.00", rate: "0",
			life: 3, purchase: "2026-01-01", disposal: "2027-03-31",
			want:      map[int]string{2026: "200000.00", 2027: "49315.07"},
			wantTotal: "249315.07",
		},
		{
			// Cost under the cap: 20% of 800,000 a year, and at least 1 baht stays (ป.3/2527 ข้อ 8).
			name: "cost 800,000", cost: "800000.00", scrap: "0.00", rate: "0",
			life: 3, purchase: "2026-01-01",
			want: map[int]string{2026: "160000.00", 2027: "160000.00", 2028: "160000.00", 2029: "160000.00",
				2030: "159999.00"},
			wantTotal: "799999.00",
		},
		{
			// Book rate 10% < 20%: every year is just the capped share of that year's book amount
			// (≈ book × 1M/1.5M); nothing is held back, so nothing is carried after 2035.
			name: "10-year book, rate under the ceiling", cost: "1500000.00", scrap: "0.00", rate: "0",
			life: 10, purchase: "2026-01-01",
			want: map[int]string{2026: "100000.03", 2027: "100000.02", 2028: "100000.01", 2029: "100000.03",
				2030: "100000.02", 2031: "100000.03", 2032: "100000.01", 2033: "100000.02", 2034: "100000.03",
				2035: "99999.80"},
			wantTotal: "1000000.00",
		},
		{
			// Same 10-year book sold 30 Jun 2027: the items after the sale stay stored but are not
			// the car's (gl_poster.go books only stopdate ≤ disposaldate), so the ป.3/2527 ข้อ 2 cap
			// is the capped share of the book to the sale: 224,383.62 × 1M/1.5M − 100,000.03 =
			// 49,589.05, not the day-prorated ceiling 99,178.08 (review 2026-09-25).
			name: "10-year book sold 30 Jun 2027", cost: "1500000.00", scrap: "0.00", rate: "0",
			life: 10, purchase: "2026-01-01", disposal: "2027-06-30",
			want:      map[int]string{2026: "100000.03", 2027: "49589.05"},
			wantTotal: "149589.08",
		},
	} {
		asset := Asset{AssetCode: "CAR", PassengerCarTaxCap: true, Cost: Amount(tc.cost), ScrapValue: Amount(tc.scrap),
			DeprecPercent: Amount(tc.rate), UsefulLifeYears: tc.life, PurchaseDate: tc.purchase, StartCalcDate: tc.purchase}
		items, err := NewCalculator().CalculateSchedule(asset, "")
		if err != nil {
			t.Fatal(err)
		}
		book := make(map[int]decimal.Decimal)
		for _, it := range items {
			y, err := strconv.Atoi(it.FiscalYear)
			if err != nil {
				t.Fatal(err)
			}
			book[y] = book[y].Add(it.PeriodDeprec.Decimal())
		}
		total := decimal.Zero
		for year := 2020; year <= 2040; year++ {
			want := decimal.Zero
			if w, ok := tc.want[year]; ok {
				want = decimal.RequireFromString(w)
			}
			got := passengerCarTaxInYear(asset, items, tc.disposal, year)
			if !got.Equal(want) {
				t.Errorf("%s: %d tax %s (book %s), want %s", tc.name, year, got, book[year], want)
			}
			ceiling := passengerCarTaxCeiling(asset.Cost.Decimal(), taxHeldDays(asset, tc.disposal, year), daysInYear(year))
			if got.GreaterThan(ceiling) || got.IsNegative() {
				t.Errorf("%s: %d tax %s outside 0..ceiling %s", tc.name, year, got, ceiling)
			}
			total = total.Add(got)
		}
		if !total.Equal(decimal.RequireFromString(tc.wantTotal)) {
			t.Errorf("%s: total tax %s, want %s", tc.name, total, tc.wantTotal)
		}
		if total.GreaterThan(passengerCarTaxBasis(asset.Cost.Decimal())) {
			t.Errorf("%s: total tax %s above the basis", tc.name, total)
		}
	}
}

// A huge year must not walk year by year to it: after the last book year the replay stops once the
// allowance is used up or the car is sold (review 2026-09-25: MaxInt never ended).
func TestPassengerCarTaxInYearStopsEarly(t *testing.T) {
	for _, tc := range []struct{ name, purchase, disposal string }{
		{"allowance used up by 2031", "2026-07-01", ""},
		{"sold with allowance left", "2026-01-01", "2027-03-31"},
	} {
		asset := Asset{AssetCode: "CAR", PassengerCarTaxCap: true, Cost: Amount("1500000.00"), ScrapValue: Amount("0.00"),
			DeprecPercent: Amount("0"), UsefulLifeYears: 3, PurchaseDate: tc.purchase, StartCalcDate: tc.purchase}
		items, err := NewCalculator().CalculateSchedule(asset, "")
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan decimal.Decimal, 1)
		go func() { done <- passengerCarTaxInYear(asset, items, tc.disposal, math.MaxInt) }()
		select {
		case got := <-done:
			if !got.IsZero() {
				t.Errorf("%s: tax in year MaxInt = %s, want 0", tc.name, got)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: passengerCarTaxInYear(MaxInt) did not stop", tc.name)
		}
	}
}

// The report accepts only Christian-era years taxReportMinYear..taxReportMaxYear and refuses the
// rest with a 400 field error before it opens the database.
func TestTaxReportYear(t *testing.T) {
	for in, want := range map[string]int{"1900": 1900, "2026": 2026, " 2026 ": 2026, "02026": 2026, "2400": 2400} {
		got, err := taxReportYear(in)
		if err != nil || got != want {
			t.Errorf("taxReportYear(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	noDB := Connector(func(string) (*sql.DB, error) { return nil, errors.New("database opened") })
	for _, in := range []string{"", "abc", "1899", "2401", "2569", "-1", "100000000", "9223372036854775807", "2026.5"} {
		_, err := NewReporter(noDB).GetTaxReconciliationReport(context.Background(), Scope{}, in)
		user, ok := gl.AsUserError(err)
		if !ok {
			t.Fatalf("fiscalyear %q: error %v, want the fiscal-year field error", in, err)
		}
		if user.Code != codeTaxReportYearInvalid || user.Field != "fiscalyear" || user.HTTPStatus() != http.StatusBadRequest {
			t.Errorf("fiscalyear %q: code %q field %q status %d", in, user.Code, user.Field, user.HTTPStatus())
		}
		// The Thai text comes from languages.tsv (not the raw key) and names both bounds.
		if user.Message == codeTaxReportYearInvalid || !strings.Contains(user.Message, strconv.Itoa(taxReportMinYear)) ||
			!strings.Contains(user.Message, strconv.Itoa(taxReportMaxYear)) {
			t.Errorf("fiscalyear %q: message %q", in, user.Message)
		}
	}
}
