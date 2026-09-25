package fixedasset

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// The passenger-car cap (docs/kms/21-thai-tax-form-references.md §13) follows the asset's own
// passengercartaxcap flag, never its type code, and scales the book depreciation to the
// capped cost so a part year stays part year.
func TestTaxReconciliationPassengerCarCap(t *testing.T) {
	connect, company := testConnector(t)
	ctx := context.Background()
	scope := Scope{Holding: "TEST_HOLDING", Company: company, Branch: "HQ", Actor: "CPA_AUDITOR"}
	now := time.Now().UTC()
	store := NewStore(connect)

	// The JSON the asset form sends after the "รถยนต์นั่งไม่เกิน 10 ที่นั่ง" preset.
	var uiCar Asset
	if err := json.Unmarshal([]byte(`{"assetcode":"CAR-UI","names":[{"code":"th","name":"รถยนต์นั่งผู้บริหาร"}],
		"assettypecode":"VEHICLE_PASSENGER","passengercartaxcap":true,"cost":"1500000.00","scrapvalue":"0.00",
		"usefullifeyears":5,"deprecpercent":"20.00","method":"straight_line",
		"purchasedate":"2026-07-01","startcalcdate":"2026-07-01","status":"active"}`), &uiCar); err != nil {
		t.Fatal(err)
	}
	car := func(code, typeCode string, capped bool, cost string) Asset {
		a := uiCar
		a.AssetCode, a.AssetTypeCode, a.PassengerCarTaxCap, a.Cost = code, typeCode, capped, Amount(cost)
		return a
	}
	// 3-year life: book rate 33.33% is above the 20% tax ceiling (RD145 s.4(5)).
	threeYear := func(code, cost, method string) Asset {
		a := car(code, "VEHICLE_PASSENGER", true, cost)
		a.UsefulLifeYears, a.DeprecPercent, a.Method = 3, Amount("0.00"), method
		a.PurchaseDate, a.StartCalcDate = "2026-01-01", "2026-01-01"
		return a
	}
	car3YJul := threeYear("CAR-3Y-JUL", "1500000.00", "straight_line")
	car3YJul.PurchaseDate, car3YJul.StartCalcDate = "2026-07-01", "2026-07-01"
	for _, a := range []Asset{
		uiCar,
		car("CAR-FULLYEAR", "ANY_CODE", true, "1500000.00"),
		car("CAR-RENTAL", "PASSENGER_CAR", false, "1500000.00"),  // old magic code, rental-business car: no cap
		car("CAR-SMALL", "VEHICLE_PASSENGER", true, "900000.00"), // within the cap: nothing disallowed
		threeYear("CAR-3Y", "1500000.00", "straight_line"),
		threeYear("CAR-3Y-SYD", "1500000.00", "sum_of_years"), // Method is ignored by the calculator
		threeYear("CAR-3Y-800K", "800000.00", "straight_line"),
		car3YJul,
	} {
		if a.AssetCode == "CAR-FULLYEAR" {
			a.PurchaseDate, a.StartCalcDate = "2026-01-01", "2026-01-01"
		}
		if _, err := store.CreateAsset(ctx, scope, a, now); err != nil {
			t.Fatalf("CreateAsset %s: %v", a.AssetCode, err)
		}
	}

	rep, err := NewReporter(connect).GetTaxReconciliationReport(ctx, scope, "2026")
	if err != nil {
		t.Fatal(err)
	}
	rows := map[string]map[string]string{}
	for _, row := range rep.Rows {
		rows[row["assetcode"]] = row
	}
	amount := func(code, key string) decimal.Decimal {
		t.Helper()
		row, ok := rows[code]
		if !ok {
			t.Fatalf("no report row for %s", code)
		}
		return decimal.RequireFromString(row[key])
	}

	// Full year at 20%: tax is 20% of 1M = 200,000 (RD ruling 0702/5605) — within the satang the
	// monthly schedule rounding leaves in the book amount.
	fullBook, fullTax := amount("CAR-FULLYEAR", "accounting_deprec"), amount("CAR-FULLYEAR", "tax_deprec")
	if !fullTax.Equal(fullBook.Mul(passengerCarTaxCostCap).Div(decimal.RequireFromString("1500000")).Round(2)) ||
		fullTax.Sub(decimal.RequireFromString("200000")).Abs().GreaterThan(decimal.RequireFromString("0.05")) {
		t.Fatalf("full-year tax depreciation = %s (book %s), want ≈200000", fullTax, fullBook)
	}

	// Bought 1 Jul: tax is 1M/1.5M of the part-year book amount, not a flat 200,000.
	acct := amount("CAR-UI", "accounting_deprec")
	wantTax := acct.Mul(passengerCarTaxCostCap).Div(decimal.RequireFromString("1500000")).Round(2)
	if got := amount("CAR-UI", "tax_deprec"); !got.Equal(wantTax) || !got.LessThan(acct) {
		t.Fatalf("part-year capped tax depreciation = %s, want %s (book %s)", got, wantTax, acct)
	}
	if got := amount("CAR-UI", "tax_difference"); !got.Equal(acct.Sub(wantTax)) {
		t.Fatalf("tax difference = %s, want %s", got, acct.Sub(wantTax))
	}

	for _, code := range []string{"CAR-RENTAL", "CAR-SMALL"} {
		if tax, book := amount(code, "tax_deprec"), amount(code, "accounting_deprec"); !tax.Equal(book) || !amount(code, "tax_difference").IsZero() {
			t.Fatalf("%s must not be capped: tax %s book %s", code, tax, book)
		}
	}

	// Book rate above 20%: the yearly ceiling (20% of the capped cost, prorated by days held) wins.
	for code, want := range map[string]string{
		"CAR-3Y":      "200000.00", // not 333,333.33 (book ≈500,000 × 1M/1.5M)
		"CAR-3Y-SYD":  "200000.00",
		"CAR-3Y-800K": "160000.00", // 20% of 800,000, not the book ≈266,666.67
		"CAR-3Y-JUL":  "100821.92", // 200,000 × 184/365 days held
	} {
		tax, book := amount(code, "tax_deprec"), amount(code, "accounting_deprec")
		if !tax.Equal(decimal.RequireFromString(want)) {
			t.Fatalf("%s tax depreciation = %s (book %s), want %s", code, tax, book, want)
		}
		if diff := amount(code, "tax_difference"); !diff.Equal(book.Sub(tax)) || diff.IsNegative() {
			t.Fatalf("%s tax difference = %s, want book %s − tax %s", code, diff, book, tax)
		}
	}

	// The ceiling follows the days the car is owned, not the days of the book items (RD145 s.4
	// opening), and what the ceiling held back is deducted in later years until the capped basis is
	// used up (ป.3/2527 ข้อ 8, กค 0702/570). Bought 1 Jul 2026, 3-year life, scrap 0: tax 100,821.92
	// + 200,000 × 4 + 99,178.08 = 1,000,000.00, the last two years with no book items (a negative
	// difference = a deduction on the ภ.ง.ด.50). Sold 31 Mar 2027: the 2027 ceiling stops at the
	// disposal date (200,000 × 90/365).
	db, err := connect(scope.Holding)
	if err != nil {
		t.Fatal(err)
	}
	sold := AssetDisposal{DocNo: "DISP-CAR-3Y", AssetCode: "CAR-3Y", DisposalDate: "2027-03-31", DisposalType: "sale"}
	sold.Identity = identityFor(scope, kindDisposal, sold.DocNo, Identity{}, now)
	if err := putRecord(ctx, db, scope.Company, kindDisposal, sold.ID, sold.DocNo, sold); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ year, code, want, wantBook string }{
		{"2029", "CAR-3Y-JUL", "200000.00", "247945.21"},
		{"2030", "CAR-3Y-JUL", "200000.00", "0.00"},
		{"2031", "CAR-3Y-JUL", "99178.08", "0.00"},
		{"2027", "CAR-3Y", "49315.07", ""},
		{"2028", "CAR-3Y", "0.00", ""},
	} {
		rep, err := NewReporter(connect).GetTaxReconciliationReport(ctx, scope, tc.year)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]string
		for _, row := range rep.Rows {
			if row["assetcode"] == tc.code {
				got = row
			}
		}
		if got["tax_deprec"] != tc.want {
			t.Fatalf("%s %s tax depreciation = %q, want %s", tc.year, tc.code, got["tax_deprec"], tc.want)
		}
		if tc.wantBook != "" && got["accounting_deprec"] != tc.wantBook {
			t.Fatalf("%s %s book depreciation = %q, want %s", tc.year, tc.code, got["accounting_deprec"], tc.wantBook)
		}
		book, tax := decimal.RequireFromString(got["accounting_deprec"]), decimal.RequireFromString(got["tax_deprec"])
		if got["tax_difference"] != book.Sub(tax).StringFixed(2) {
			t.Fatalf("%s %s tax difference = %q, want book − tax", tc.year, tc.code, got["tax_difference"])
		}
	}
}

// taxHeldDays counts the days of the year the asset is owned: acquisition (purchasedate, else
// startcalcdate) or 1 Jan to disposal or 31 Dec, both days included.
func TestTaxHeldDays(t *testing.T) {
	for _, tc := range []struct {
		name, purchase, start, disposal string
		year, want                      int
	}{
		{"bought 1 Jul", "2026-07-01", "2026-07-01", "", 2026, 184},
		{"bought earlier, still held", "2026-07-01", "2026-07-01", "", 2029, 365},
		{"leap year", "2026-07-01", "2026-07-01", "", 2028, 366},
		{"bought 31 Dec", "2026-12-31", "2026-12-31", "", 2026, 1},
		{"sold 31 Mar", "2026-01-01", "2026-01-01", "2027-03-31", 2027, 90},
		{"bought and sold in the year", "2026-07-01", "2026-07-01", "2026-07-31", 2026, 31},
		{"sold in an earlier year", "2026-01-01", "2026-01-01", "2027-03-31", 2028, 0},
		{"bought in a later year", "2027-01-01", "2027-01-01", "", 2026, 0},
		{"no purchase date: start date", "", "2026-07-01", "", 2026, 184},
		{"no dates: whole year", "", "", "", 2026, 365},
	} {
		asset := Asset{PurchaseDate: tc.purchase, StartCalcDate: tc.start}
		if got := taxHeldDays(asset, tc.disposal, tc.year); got != tc.want {
			t.Errorf("%s: days held = %d, want %d", tc.name, got, tc.want)
		}
	}
}

// The three pieces of the passenger-car rule without a database — docs/kms/21-thai-tax-form-references.md §13:
// the yearly ceiling (20% of the cost up to 1M, by days held), the capped share of a book amount,
// and the most the tax can ever deduct (1M above the cap, else cost − 1 baht).
func TestPassengerCarTaxPieces(t *testing.T) {
	d := decimal.RequireFromString
	for _, tc := range []struct {
		name, cost         string
		heldDays, yearDays int
		want               string
	}{
		{"full year, cost over the cap", "1500000", 365, 365, "200000.00"},
		{"bought 1 Jul (184 days)", "1500000", 184, 365, "100821.92"},
		{"leap year, bought 1 Mar (306 of 366 days)", "1500000", 306, 366, "167213.11"},
		{"sold 31 Mar (90 days)", "1500000", 90, 365, "49315.07"},
		{"cost 800,000", "800000", 365, 365, "160000.00"},
		{"not held this year", "1500000", 0, 365, "0"},
	} {
		if got := passengerCarTaxCeiling(d(tc.cost), tc.heldDays, tc.yearDays); !got.Equal(d(tc.want)) {
			t.Errorf("ceiling %s: %s, want %s", tc.name, got, tc.want)
		}
	}
	for _, tc := range []struct{ cost, book, want string }{
		{"1500000", "500000.00", "333333.33"},
		{"1500000", "252054.78", "168036.52"},
		{"1000000", "500000.00", "500000.00"},
		{"800000", "266666.67", "266666.67"},
	} {
		if got := passengerCarTaxShare(d(tc.cost), d(tc.book)); !got.Equal(d(tc.want)) {
			t.Errorf("share of %s at cost %s: %s, want %s", tc.book, tc.cost, got, tc.want)
		}
	}
	for _, tc := range []struct{ cost, want string }{
		{"1500000", "1000000"},
		{"1000000", "999999"},
		{"800000", "799999"},
		{"0.50", "0"},
	} {
		if got := passengerCarTaxBasis(d(tc.cost)); !got.Equal(d(tc.want)) {
			t.Errorf("basis at cost %s: %s, want %s", tc.cost, got, tc.want)
		}
	}
}
