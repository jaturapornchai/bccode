package fixedasset

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type ReportColumn struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Amount bool   `json:"amount,omitempty"`
}

type ReportResult struct {
	Columns   []ReportColumn      `json:"columns"`
	Rows      []map[string]string `json:"rows"`
	Totals    map[string]string   `json:"totals"`
	TotalRows int                 `json:"totalrows"`
	AsOf      string              `json:"asof"`
}

// passengerCarTaxCostCap is the per-vehicle cost that may be depreciated for tax on a
// passenger car / bus with ≤10 seats (Royal Decree 145 s.5; the excess is disallowed by
// Royal Decree 315 s.4(1)) — docs/kms/21-thai-tax-form-references.md §13.
var passengerCarTaxCostCap = decimal.NewFromInt(1000000)

// passengerCarTaxRatePercent is the highest yearly tax rate for "ทรัพย์สินอย่างอื่น", the class
// vehicles fall in (Royal Decree 145 s.4(5)) — docs/kms/21-thai-tax-form-references.md §13.
var passengerCarTaxRatePercent = decimal.NewFromInt(20)

// passengerCarTaxDeprec is the ภ.ง.ด.50 depreciation of one flagged passenger car for one year
// (docs/kms/21-thai-tax-form-references.md §13): the book amount on the cost up to 1,000,000
// only (RD145 s.5; same rate and holding period as the books, RD ruling 0702/5605), but never
// above 20% a year of that capped cost for the days held (RD145 s.4 opening + s.4(5)).
// heldDays is the days of the calendar year the car was owned (taxHeldDays), yearDays the days
// in that year — not the days of this year's book items, which stop once the book value reaches
// scrap even though the car is still held (3-year life: the last book year).
// The ceiling applies to every flagged car, also at cost ≤ 1,000,000: the flag is the owner
// asking for the passenger-car tax rules. The calculator writes equal yearly rates only (it
// ignores Asset.Method), so RD145 s.4 para 2 — unequal-rate methods may exceed the rate in some
// years when the life is ≥ 100/20 years — never applies to these amounts; revisit it (and s.4
// para 3, whose double-declining option excludes passenger cars) if Method is ever honoured.
func passengerCarTaxDeprec(cost, acctDep decimal.Decimal, heldDays, yearDays int) decimal.Decimal {
	taxDep := acctDep
	if cost.GreaterThan(passengerCarTaxCostCap) {
		taxDep = acctDep.Mul(passengerCarTaxCostCap).Div(cost).Round(2)
	}
	if yearDays > 0 {
		ceiling := decimal.Min(cost, passengerCarTaxCostCap).Mul(passengerCarTaxRatePercent).
			Mul(decimal.NewFromInt(int64(heldDays))).Div(decimal.NewFromInt(int64(100 * yearDays))).Round(2)
		taxDep = decimal.Min(taxDep, ceiling)
	}
	return decimal.Max(taxDep, decimal.Zero)
}

// taxHeldDays is how many days of the calendar year the company owned the asset, the period RD145
// s.4 opening prorates the ceiling by ("ให้คำนวณหักตามระยะเวลาที่ได้ทรัพย์สินนั้นมาในแต่ละรอบระยะเวลาบัญชี",
// docs/kms/21-thai-tax-form-references.md §13): from the acquisition date (purchasedate, else
// startcalcdate) or 1 Jan, to the disposal date or 31 Dec, both days counted. RD145 says nothing
// about the disposal day; it counts as held, like the first day.
func taxHeldDays(asset Asset, disposalDate string, year int) int {
	first := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	last := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	acquired, err := time.Parse("2006-01-02", asset.PurchaseDate)
	if err != nil {
		acquired, err = time.Parse("2006-01-02", asset.StartCalcDate)
	}
	if err == nil && acquired.After(first) {
		first = acquired
	}
	if disposed, err := time.Parse("2006-01-02", disposalDate); err == nil && disposed.Before(last) {
		last = disposed
	}
	if last.Before(first) {
		return 0
	}
	return int(last.Sub(first).Hours()/24) + 1
}

type Reporter struct {
	records *records
}

func NewReporter(connect Connector) *Reporter {
	return &Reporter{records: newRecords(connect)}
}

// GetAssetScheduleReport builds the full Fixed Asset Schedule report.
func (r *Reporter) GetAssetScheduleReport(ctx context.Context, scope Scope, fiscalYear string, period int, typeCode string) (*ReportResult, error) {
	if fiscalYear == "" {
		fiscalYear = fmt.Sprintf("%d", 2026)
	}

	cols := []ReportColumn{
		{Key: "assetcode", Label: "รหัสสินทรัพย์"},
		{Key: "assetname", Label: "ชื่อสินทรัพย์"},
		{Key: "purchasedate", Label: "วันที่ได้มา"},
		{Key: "deprecpercent", Label: "อัตรา (%)", Amount: true},
		{Key: "begincost", Label: "ราคาทุนยกมา", Amount: true},
		{Key: "costaddition", Label: "ทุนเพิ่มขึ้น", Amount: true},
		{Key: "costdisposal", Label: "ทุนลดลง (จำหน่าย)", Amount: true},
		{Key: "endingcost", Label: "ราคาทุนยกไป", Amount: true},
		{Key: "beginaccum", Label: "ค่าเสื่อมยกมา", Amount: true},
		{Key: "perioddeprec", Label: "ค่าเสื่อมงวดนี้", Amount: true},
		{Key: "disposalaccum", Label: "ค่าเสื่อมลดลง", Amount: true},
		{Key: "endingaccum", Label: "ค่าเสื่อมสะสมยกไป", Amount: true},
		{Key: "netbookvalue", Label: "มูลค่าตามบัญชียกไป", Amount: true},
	}

	db, err := r.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	assets, err := queryRecords[Asset](ctx, db, scope.Company, kindAsset, ` AND ($3 = '' OR payload->>'assettypecode' = $3)`, ` ORDER BY code`, typeCode)
	if err != nil {
		return nil, err
	}

	// Fetch all depreciation items for these assets in the given fiscal year
	allDeprec, err := queryRecords[DepreciationScheduleItem](ctx, db, scope.Company, kindDepreciation, ` AND payload->>'fiscalyear' = $3 AND ($4 <= 0 OR (payload->>'period')::int <= $4)`, "", fiscalYear, period)
	if err != nil {
		return nil, err
	}

	deprecByAsset := make(map[string][]DepreciationScheduleItem)
	for _, d := range allDeprec {
		deprecByAsset[d.AssetCode] = append(deprecByAsset[d.AssetCode], d)
	}

	allDisposals, err := queryRecords[AssetDisposal](ctx, db, scope.Company, kindDisposal, "", "")
	if err != nil {
		return nil, err
	}
	dispMap := make(map[string]AssetDisposal)
	for _, d := range allDisposals {
		dispMap[d.AssetCode] = d
	}

	var rows []map[string]string
	totBeginCost := decimal.Zero
	totAddition := decimal.Zero
	totDisposalCost := decimal.Zero
	totEndingCost := decimal.Zero
	totBeginAccum := decimal.Zero
	totPeriodDep := decimal.Zero
	totDisposalAccum := decimal.Zero
	totEndingAccum := decimal.Zero
	totNetBook := decimal.Zero

	for _, ast := range assets {
		depItems := deprecByAsset[ast.AssetCode]
		disp, hasDisposal := dispMap[ast.AssetCode]

		beginCost := ast.Cost.Decimal()
		addition := decimal.Zero
		disposalCost := decimal.Zero
		disposalAccum := decimal.Zero

		if hasDisposal {
			disposalCost = ast.Cost.Decimal()
			disposalAccum = disp.AccumDeprecAtDisposal.Decimal()
		}

		endingCost := beginCost.Add(addition).Sub(disposalCost)

		beginAccum := ast.BeginAccumDeprec.Decimal()
		periodDep := decimal.Zero

		for _, item := range depItems {
			if period > 0 && item.Period == period {
				periodDep = periodDep.Add(item.PeriodDeprec.Decimal())
			} else if period <= 0 {
				periodDep = periodDep.Add(item.PeriodDeprec.Decimal())
			}
		}

		endingAccum := beginAccum.Add(periodDep).Sub(disposalAccum)
		netBook := endingCost.Sub(endingAccum)
		if netBook.LessThan(decimal.Zero) {
			netBook = decimal.Zero
		}

		row := map[string]string{
			"assetcode":     ast.AssetCode,
			"assetname":     ast.ThaiName(),
			"purchasedate":  ast.PurchaseDate,
			"deprecpercent": ast.DeprecPercent.Decimal().StringFixed(2) + "%",
			"begincost":     beginCost.StringFixed(2),
			"costaddition":  addition.StringFixed(2),
			"costdisposal":  disposalCost.StringFixed(2),
			"endingcost":    endingCost.StringFixed(2),
			"beginaccum":    beginAccum.StringFixed(2),
			"perioddeprec":  periodDep.StringFixed(2),
			"disposalaccum": disposalAccum.StringFixed(2),
			"endingaccum":   endingAccum.StringFixed(2),
			"netbookvalue":  netBook.StringFixed(2),
		}
		rows = append(rows, row)

		totBeginCost = totBeginCost.Add(beginCost)
		totAddition = totAddition.Add(addition)
		totDisposalCost = totDisposalCost.Add(disposalCost)
		totEndingCost = totEndingCost.Add(endingCost)
		totBeginAccum = totBeginAccum.Add(beginAccum)
		totPeriodDep = totPeriodDep.Add(periodDep)
		totDisposalAccum = totDisposalAccum.Add(disposalAccum)
		totEndingAccum = totEndingAccum.Add(endingAccum)
		totNetBook = totNetBook.Add(netBook)
	}

	totals := map[string]string{
		"assetcode":     "รวมทั้งหมด",
		"assetname":     fmt.Sprintf("%d รายการ", len(rows)),
		"begincost":     totBeginCost.StringFixed(2),
		"costaddition":  totAddition.StringFixed(2),
		"costdisposal":  totDisposalCost.StringFixed(2),
		"endingcost":    totEndingCost.StringFixed(2),
		"beginaccum":    totBeginAccum.StringFixed(2),
		"perioddeprec":  totPeriodDep.StringFixed(2),
		"disposalaccum": totDisposalAccum.StringFixed(2),
		"endingaccum":   totEndingAccum.StringFixed(2),
		"netbookvalue":  totNetBook.StringFixed(2),
	}

	return &ReportResult{
		Columns:   cols,
		Rows:      rows,
		Totals:    totals,
		TotalRows: len(rows),
		AsOf:      fiscalYear,
	}, nil
}

// GetTaxReconciliationReport generates comparative tax vs accounting depreciation for Thai CPA / PND 50.
func (r *Reporter) GetTaxReconciliationReport(ctx context.Context, scope Scope, fiscalYear string) (*ReportResult, error) {
	cols := []ReportColumn{
		{Key: "assetcode", Label: "รหัสสินทรัพย์"},
		{Key: "assetname", Label: "ชื่อสินทรัพย์"},
		{Key: "cost", Label: "ราคาทุน", Amount: true},
		{Key: "accounting_deprec", Label: "ค่าเสื่อมราคาทางบัญชี", Amount: true},
		{Key: "tax_deprec", Label: "ค่าเสื่อมราคาทางภาษี (ภ.ง.ด.50)", Amount: true},
		{Key: "tax_difference", Label: "ผลต่าง (ปรับปรุงกำไรสุทธิ)", Amount: true},
		{Key: "remark", Label: "หมายเหตุเกณฑ์ภาษี"},
	}

	db, err := r.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	assets, err := queryRecords[Asset](ctx, db, scope.Company, kindAsset, "", ` ORDER BY code`)
	if err != nil {
		return nil, err
	}
	allDep, err := queryRecords[DepreciationScheduleItem](ctx, db, scope.Company, kindDepreciation, ` AND payload->>'fiscalyear' = $3`, "", fiscalYear)
	if err != nil {
		return nil, err
	}

	depMap := make(map[string]decimal.Decimal)
	for _, d := range allDep {
		depMap[d.AssetCode] = depMap[d.AssetCode].Add(d.PeriodDeprec.Decimal())
	}
	// The items above are this calendar year's (calculator.go sets fiscalyear = the calendar
	// year); an unparsable year matched no items, so the fallback never prorates anything.
	// A disposal on or before 31 Dec ends the days held (a car sold in an earlier year: 0 days).
	year, yearErr := strconv.Atoi(fiscalYear)
	yearDays := 0
	disposedOn := make(map[string]string)
	if yearErr == nil {
		yearDays = daysInYear(year)
		disposals, err := queryRecords[AssetDisposal](ctx, db, scope.Company, kindDisposal, ` AND payload->>'disposaldate' <= $3`, "", fmt.Sprintf("%04d-12-31", year))
		if err != nil {
			return nil, err
		}
		for _, d := range disposals {
			disposedOn[d.AssetCode] = d.DisposalDate
		}
	}

	var rows []map[string]string
	totCost := decimal.Zero
	totAcct := decimal.Zero
	totTax := decimal.Zero
	totDiff := decimal.Zero

	for _, ast := range assets {
		acctDep := depMap[ast.AssetCode]
		taxDep := acctDep
		remark := "หักตามอัตราปกติ"

		cost := ast.Cost.Decimal()
		if ast.PassengerCarTaxCap {
			taxDep = passengerCarTaxDeprec(cost, acctDep, taxHeldDays(ast, disposedOn[ast.AssetCode], year), yearDays)
			remark = "รถยนต์นั่ง/รถยนต์โดยสารไม่เกิน 10 ที่นั่ง: หักจากต้นทุนไม่เกิน 1,000,000 บาท อัตราไม่เกินร้อยละ 20 ต่อปี"
		} else if ast.FirstYearPercent.Decimal().GreaterThan(decimal.Zero) {
			remark = fmt.Sprintf("สิทธิประโยชน์หักค่าสึกหรอปีแรกพิเศษ %s%%", ast.FirstYearPercent.String())
		}

		diff := acctDep.Sub(taxDep)

		rows = append(rows, map[string]string{
			"assetcode":         ast.AssetCode,
			"assetname":         ast.ThaiName(),
			"cost":              cost.StringFixed(2),
			"accounting_deprec": acctDep.StringFixed(2),
			"tax_deprec":        taxDep.StringFixed(2),
			"tax_difference":    diff.StringFixed(2),
			"remark":            remark,
		})

		totCost = totCost.Add(cost)
		totAcct = totAcct.Add(acctDep)
		totTax = totTax.Add(taxDep)
		totDiff = totDiff.Add(diff)
	}

	totals := map[string]string{
		"assetcode":         "รวมทั้งหมด",
		"assetname":         fmt.Sprintf("%d รายการ", len(rows)),
		"cost":              totCost.StringFixed(2),
		"accounting_deprec": totAcct.StringFixed(2),
		"tax_deprec":        totTax.StringFixed(2),
		"tax_difference":    totDiff.StringFixed(2),
	}

	return &ReportResult{
		Columns:   cols,
		Rows:      rows,
		Totals:    totals,
		TotalRows: len(rows),
		AsOf:      fiscalYear,
	}, nil
}
