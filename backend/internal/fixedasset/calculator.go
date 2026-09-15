package fixedasset

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

// daysInMonth returns the number of days in the given year and month.
func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// isLeapYear checks if the year is a leap year.
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func daysInYear(year int) int {
	if isLeapYear(year) {
		return 366
	}
	return 365
}

// CalculateSchedule generates monthly depreciation schedule matching Champ's CDepreciation.
func (c *Calculator) CalculateSchedule(asset Asset, upToDate string) ([]DepreciationScheduleItem, error) {
	cost := asset.Cost.Decimal()
	if cost.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("ราคาทุนสินทรัพย์ต้องมากกว่า 0")
	}

	startDate, err := time.Parse("2006-01-02", asset.StartCalcDate)
	if err != nil {
		return nil, fmt.Errorf("วันที่เริ่มคิดค่าเสื่อมไม่ถูกต้อง (ต้องเป็น YYYY-MM-DD): %w", err)
	}

	// Determine Useful Life and Percent
	percent := asset.DeprecPercent.Decimal()
	if percent.LessThanOrEqual(decimal.Zero) {
		if asset.UsefulLifeYears <= 0 {
			return nil, fmt.Errorf("ต้องระบุอายุการใช้งาน (ปี) หรืออัตราค่าเสื่อมราคา (%%)")
		}
		percent = decimal.NewFromInt(100).Div(decimal.NewFromInt(int64(asset.UsefulLifeYears)))
	}

	var maxDate time.Time
	if upToDate != "" {
		parsedMax, err := time.Parse("2006-01-02", upToDate)
		if err == nil {
			maxDate = parsedMax
		}
	}
	if maxDate.IsZero() {
		// Default calculate for UsefulLifeYears + 2 years cushion
		yearsCushion := asset.UsefulLifeYears + 2
		if yearsCushion < 5 {
			yearsCushion = 5
		}
		maxDate = startDate.AddDate(yearsCushion, 0, 0)
	}

	scrap := asset.ScrapValue.Decimal()
	if scrap.LessThan(decimal.Zero) {
		scrap = decimal.Zero
	}

	currentAccum := asset.BeginAccumDeprec.Decimal()
	currentNetBook := cost.Sub(currentAccum)

	// Depreciable base = Cost - Scrap
	depreciableTotal := cost.Sub(scrap)
	if depreciableTotal.LessThanOrEqual(decimal.Zero) {
		// No depreciation if Cost <= Scrap
		return nil, nil
	}

	var schedule []DepreciationScheduleItem

	currentYear := startDate.Year()
	currentMonth := startDate.Month()
	firstPeriod := true

	for {
		if currentNetBook.LessThanOrEqual(scrap) {
			break
		}

		monthDays := daysInMonth(currentYear, currentMonth)
		periodStartDay := 1
		if firstPeriod {
			periodStartDay = startDate.Day()
		}

		pStartDate := time.Date(currentYear, currentMonth, periodStartDay, 0, 0, 0, 0, time.UTC)
		pStopDate := time.Date(currentYear, currentMonth, monthDays, 0, 0, 0, 0, time.UTC)

		if pStartDate.After(maxDate) {
			break
		}

		daysInPeriod := monthDays - periodStartDay + 1
		yearDays := daysInYear(currentYear)

		// Daily depreciation rate for the year = (Depreciable Total * (Percent / 100)) / yearDays
		annualDeprec := depreciableTotal.Mul(percent).Div(decimal.NewFromInt(100))
		dailyDeprec := annualDeprec.Div(decimal.NewFromInt(int64(yearDays)))

		periodAmount := dailyDeprec.Mul(decimal.NewFromInt(int64(daysInPeriod))).Round(2)

		// Check first year special allowance (e.g. 40% initial allowance for computers under Thai tax code)
		firstYearPercent := asset.FirstYearPercent.Decimal()
		if firstPeriod && firstYearPercent.GreaterThan(decimal.Zero) {
			firstYearAllowance := cost.Mul(firstYearPercent).Div(decimal.NewFromInt(100)).Round(2)
			periodAmount = periodAmount.Add(firstYearAllowance)
		}

		// Ensure we don't depreciate below scrap value
		maxPossibleDeprec := currentNetBook.Sub(scrap)
		if periodAmount.GreaterThan(maxPossibleDeprec) {
			periodAmount = maxPossibleDeprec
		}

		if periodAmount.LessThanOrEqual(decimal.Zero) {
			break
		}

		currentAccum = currentAccum.Add(periodAmount)
		currentNetBook = currentNetBook.Sub(periodAmount)

		item := DepreciationScheduleItem{
			AssetCode:    asset.AssetCode,
			FiscalYear:   fmt.Sprintf("%d", currentYear),
			Period:       int(currentMonth),
			StartDate:    pStartDate.Format("2006-01-02"),
			StopDate:     pStopDate.Format("2006-01-02"),
			Days:         daysInPeriod,
			PeriodDeprec: AmountFromDecimal(periodAmount),
			AccumDeprec:  AmountFromDecimal(currentAccum),
			NetBookValue: AmountFromDecimal(currentNetBook),
			IsPosted:     false,
		}

		schedule = append(schedule, item)

		firstPeriod = false
		currentMonth++
		if currentMonth > 12 {
			currentMonth = 1
			currentYear++
		}
	}

	return schedule, nil
}
