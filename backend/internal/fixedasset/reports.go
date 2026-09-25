package fixedasset

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/language"

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

// passengerCarTaxCeiling is the most one flagged car may deduct for one calendar year: 20% of the
// cost up to 1,000,000 for the days held (RD145 s.4 opening + s.4(5) + s.5; order ป.3/2527 item 3
// prorates by days) — docs/kms/21-thai-tax-form-references.md §13. heldDays comes from
// taxHeldDays, not from the book items, which stop once the book value reaches scrap even though
// the car is still held. The ceiling applies to every flagged car, also at cost ≤ 1,000,000: the
// flag is the owner asking for the passenger-car tax rules. The calculator writes equal yearly
// rates only (it ignores Asset.Method), so RD145 s.4 para 2 — unequal-rate methods may exceed the
// rate in some years when the life is ≥ 100/20 years — never applies here; revisit it (and s.4
// para 3, whose double-declining option excludes passenger cars) if Method is ever honoured.
func passengerCarTaxCeiling(cost decimal.Decimal, heldDays, yearDays int) decimal.Decimal {
	if heldDays <= 0 || yearDays <= 0 {
		return decimal.Zero
	}
	return decimal.Min(cost, passengerCarTaxCostCap).Mul(passengerCarTaxRatePercent).
		Mul(decimal.NewFromInt(int64(heldDays))).Div(decimal.NewFromInt(int64(100 * yearDays))).Round(2)
}

// passengerCarTaxShare is the part of a book depreciation total that falls on the cost up to
// 1,000,000 (RD145 s.5; same rate and holding period as the books, RD ruling 0702/5605):
// book × 1,000,000 ÷ cost above the cap, all of it otherwise.
func passengerCarTaxShare(cost, book decimal.Decimal) decimal.Decimal {
	if cost.GreaterThan(passengerCarTaxCostCap) {
		return book.Mul(passengerCarTaxCostCap).Div(cost).Round(2)
	}
	return book
}

// passengerCarTaxBasis is the most a flagged car's tax depreciation may add up to over its life:
// RD145 s.8 never lets it reach the whole cost and order ป.3/2527 item 8 keeps at least 1 baht,
// but for a car above the cap exactly the cost above the cap — so the whole capped cost is
// deductible (item 8 still prints the old 500,000 cap; RD145 s.5 now says 1,000,000).
func passengerCarTaxBasis(cost decimal.Decimal) decimal.Decimal {
	if cost.GreaterThan(passengerCarTaxCostCap) {
		return passengerCarTaxCostCap
	}
	return decimal.Max(cost.Sub(decimal.NewFromInt(1)), decimal.Zero)
}

// passengerCarTaxInYear replays one flagged car's book schedule (its items of every year up to
// year) and returns the ภ.ง.ด.50 depreciation of the calendar year year —
// docs/kms/21-thai-tax-form-references.md §13:
//   - a year never deducts more than passengerCarTaxCeiling, so nothing after the disposal date;
//   - the running total never passes the book depreciation on the capped cost (order ป.3/2527
//     item 2: a lower book rate is the tax rate) nor passengerCarTaxBasis;
//   - what the ceiling held back (added back on the ภ.ง.ด.50) is deducted in the next years, also
//     after the book schedule has ended, until the running total catches up — RD ruling
//     กค 0702/570 (the books wrote the cost off early: "ต้องหักค่าสึกหรอและค่าเสื่อมราคาต่อไปจนหมด")
//     and กค 0811/09658 (an added-back amount is depreciated over the remaining life).
//
// Items that stop after the disposal date are left out: disposal (gl_poster.go DisposeAsset)
// keeps the rest of the schedule but books only the items with stopdate ≤ disposaldate, so the
// item 2 cap of the disposal year must not count the months after the sale either.
// Opening accumulated depreciation (beginaccumdeprec) counts as already deducted at its capped
// share: the years before it are not in the system, so nothing is carried from them.
// Rounding: 2 places, half away from zero (decimal.Round), taken on the running book total so
// yearly roundings do not drift.
func passengerCarTaxInYear(asset Asset, items []DepreciationScheduleItem, disposalDate string, year int) decimal.Decimal {
	bookByYear := make(map[int]decimal.Decimal)
	firstYear, lastBookYear, found := 0, 0, false
	for _, item := range items {
		y, err := strconv.Atoi(item.FiscalYear)
		if err != nil || y > year || (disposalDate != "" && item.StopDate > disposalDate) {
			continue
		}
		bookByYear[y] = bookByYear[y].Add(item.PeriodDeprec.Decimal())
		if !found || y < firstYear {
			firstYear = y
		}
		if !found || y > lastBookYear {
			lastBookYear = y
		}
		found = true
	}
	cost := asset.Cost.Decimal()
	if !found || !cost.IsPositive() || year < firstYear {
		return decimal.Zero
	}
	disposalYear := 0
	if disposed, err := time.Parse("2006-01-02", disposalDate); err == nil {
		disposalYear = disposed.Year()
	}
	basis := passengerCarTaxBasis(cost)
	bookTotal := asset.BeginAccumDeprec.Decimal()
	deducted := passengerCarTaxShare(cost, bookTotal)
	for y := firstYear; y <= year; y++ {
		bookTotal = bookTotal.Add(bookByYear[y])
		allowed := decimal.Min(passengerCarTaxShare(cost, bookTotal), basis).Sub(deducted)
		ceiling := passengerCarTaxCeiling(cost, taxHeldDays(asset, disposalDate, y), daysInYear(y))
		tax := decimal.Max(decimal.Min(allowed, ceiling), decimal.Zero)
		if y == year {
			return tax
		}
		deducted = deducted.Add(tax)
		// After the last book year the allowance stops growing: once it is used up, or the car is
		// sold, every later year is 0 — stop here instead of walking on to year.
		if y >= lastBookYear && (!allowed.Sub(tax).IsPositive() || (disposalYear > 0 && y >= disposalYear)) {
			return decimal.Zero
		}
	}
	return decimal.Zero
}

// codeTaxReportYearInvalid is the error code and language key (languages.tsv) of a fiscalyear the
// tax reconciliation report does not accept.
const codeTaxReportYearInvalid = "fa_err_fiscal_year_invalid"

// taxReportMinYear..taxReportMaxYear are the fiscal years the tax reconciliation report accepts:
// Christian-era years, the calendar year calculator.go writes into fiscalyear (a Buddhist-era
// year, 2443 and up, matches no item). The bound also caps the yearly replay in
// passengerCarTaxInYear — review 2026-09-25: fiscalyear=100000000 took about 150 s of CPU and
// 16 GB of heap per flagged car, and MaxInt never ended. The error text in languages.tsv names
// both bounds; change it with them.
const (
	taxReportMinYear = 1900
	taxReportMaxYear = 2400
)

// taxReportYear parses the fiscalyear of the tax reconciliation report, or returns a 400 field
// error (code and language key codeTaxReportYearInvalid, field fiscalyear).
func taxReportYear(fiscalYear string) (int, error) {
	year, err := strconv.Atoi(strings.TrimSpace(fiscalYear))
	if err != nil || year < taxReportMinYear || year > taxReportMaxYear {
		return 0, faFieldError(codeTaxReportYearInvalid, "fiscalyear", language.Text(codeTaxReportYearInvalid, "th"))
	}
	return year, nil
}

// passengerCarTaxForYear is the ภ.ง.ด.50 depreciation for year of every flagged car in assets. It
// reads each car's whole book schedule, because what the ceiling held back in earlier years is
// deducted later (passengerCarTaxInYear) — also in a year with no book items.
func passengerCarTaxForYear(ctx context.Context, q queryer, company string, assets []Asset, disposedOn map[string]string, year int) (map[string]decimal.Decimal, error) {
	var codes []string
	for _, a := range assets {
		if a.PassengerCarTaxCap {
			codes = append(codes, a.AssetCode)
		}
	}
	taxByCar := make(map[string]decimal.Decimal)
	if len(codes) == 0 {
		return taxByCar, nil
	}
	codesJSON, err := json.Marshal(codes)
	if err != nil {
		return nil, err
	}
	history, err := queryRecords[DepreciationScheduleItem](ctx, q, company, kindDepreciation,
		` AND payload->>'assetcode' IN (SELECT jsonb_array_elements_text($3::jsonb))`, "", string(codesJSON))
	if err != nil {
		return nil, err
	}
	itemsByCar := make(map[string][]DepreciationScheduleItem)
	for _, item := range history {
		itemsByCar[item.AssetCode] = append(itemsByCar[item.AssetCode], item)
	}
	for _, a := range assets {
		if a.PassengerCarTaxCap {
			taxByCar[a.AssetCode] = passengerCarTaxInYear(a, itemsByCar[a.AssetCode], disposedOn[a.AssetCode], year)
		}
	}
	return taxByCar, nil
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
	year, err := taxReportYear(fiscalYear)
	if err != nil {
		return nil, err
	}
	fiscalYear = strconv.Itoa(year) // "02026" / " 2026" must still match the stored "2026"
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
	// year). A disposal on or before 31 Dec ends the days held (a car sold in an earlier year: 0 days).
	disposedOn := make(map[string]string)
	disposals, err := queryRecords[AssetDisposal](ctx, db, scope.Company, kindDisposal, ` AND payload->>'disposaldate' <= $3`, "", fmt.Sprintf("%04d-12-31", year))
	if err != nil {
		return nil, err
	}
	for _, d := range disposals {
		disposedOn[d.AssetCode] = d.DisposalDate
	}
	carTax, err := passengerCarTaxForYear(ctx, db, scope.Company, assets, disposedOn, year)
	if err != nil {
		return nil, err
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
			taxDep = carTax[ast.AssetCode]
			remark = "รถยนต์นั่ง/รถยนต์โดยสารไม่เกิน 10 ที่นั่ง: หักจากต้นทุนไม่เกิน 1,000,000 บาท อัตราไม่เกินร้อยละ 20 ต่อปี ส่วนที่เกินเพดานยกไปหักปีถัดไปจนครบ"
		} else if ast.FirstYearPercent.Decimal().GreaterThan(decimal.Zero) {
			remark = fmt.Sprintf("สิทธิประโยชน์หักค่าสึกหรอปีแรกพิเศษ %s%%", ast.FirstYearPercent.String())
		}

		// Negative = a deduction on the ภ.ง.ด.50: a flagged car catching up what the ceiling held back.
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
