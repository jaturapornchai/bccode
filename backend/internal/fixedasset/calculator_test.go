package fixedasset

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculator_StraightLineStandard(t *testing.T) {
	calc := NewCalculator()

	asset := Asset{
		AssetCode:       "TEST-EQ-001",
		Cost:            Amount("120000.00"),
		ScrapValue:      Amount("1.00"),
		UsefulLifeYears: 5,
		DeprecPercent:   Amount("20.00"),
		PurchaseDate:    "2026-01-01",
		StartCalcDate:   "2026-01-01",
	}

	schedule, err := calc.CalculateSchedule(asset, "2031-12-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schedule) == 0 {
		t.Fatalf("expected schedule items, got none")
	}

	// First month (Jan 2026 = 31 days)
	first := schedule[0]
	if first.Period != 1 || first.FiscalYear != "2026" {
		t.Errorf("expected Jan 2026, got %s period %d", first.FiscalYear, first.Period)
	}
	if first.PeriodDeprec.Decimal().LessThanOrEqual(decimal.Zero) {
		t.Errorf("expected positive deprec, got %s", first.PeriodDeprec)
	}

	// Last item should have NetBookValue == ScrapValue (1.00)
	last := schedule[len(schedule)-1]
	if !last.NetBookValue.Decimal().Equal(decimal.NewFromFloat(1)) {
		t.Errorf("expected final NetBookValue = 1.00, got %s", last.NetBookValue)
	}

	// Total accumulated deprecation should be Cost - Scrap = 119,999.00
	expectedAccum := asset.Cost.Decimal().Sub(asset.ScrapValue.Decimal())
	if !last.AccumDeprec.Decimal().Equal(expectedAccum) {
		t.Errorf("expected total accum %s, got %s", expectedAccum, last.AccumDeprec)
	}
}

func TestCalculator_FirstYearSpecialAllowance(t *testing.T) {
	calc := NewCalculator()

	// Computer equipment: Cost 50,000 THB, 40% initial allowance in first month, 3 years life (20% p.a.)
	asset := Asset{
		AssetCode:        "TEST-COM-001",
		Cost:             Amount("50000.00"),
		ScrapValue:       Amount("1.00"),
		UsefulLifeYears:  3,
		DeprecPercent:    Amount("20.00"),
		FirstYearPercent: Amount("40.00"), // 40% = 20,000 THB in first month
		PurchaseDate:     "2026-03-15",
		StartCalcDate:    "2026-03-15",
	}

	schedule, err := calc.CalculateSchedule(asset, "2030-12-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	first := schedule[0]
	// First period should include 40% allowance (20,000) + daily deprec for 17 days
	if first.PeriodDeprec.Decimal().LessThan(decimal.NewFromFloat(20000)) {
		t.Errorf("expected first month deprec >= 20,000, got %s", first.PeriodDeprec)
	}

	last := schedule[len(schedule)-1]
	if !last.NetBookValue.Decimal().Equal(decimal.NewFromFloat(1)) {
		t.Errorf("expected final NetBookValue = 1.00, got %s", last.NetBookValue)
	}
}
