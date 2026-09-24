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
	for _, a := range []Asset{
		uiCar,
		car("CAR-FULLYEAR", "ANY_CODE", true, "1500000.00"),
		car("CAR-RENTAL", "PASSENGER_CAR", false, "1500000.00"),  // old magic code, rental-business car: no cap
		car("CAR-SMALL", "VEHICLE_PASSENGER", true, "900000.00"), // within the cap: nothing disallowed
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
}
