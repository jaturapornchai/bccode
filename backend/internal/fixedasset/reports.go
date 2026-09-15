package fixedasset

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

type Reporter struct {
	db *mongo.Database
}

func NewReporter(db *mongo.Database) *Reporter {
	return &Reporter{db: db}
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

	f := scopeFilter(scope)
	f["isdeleted"] = false
	if typeCode != "" {
		f["assettypecode"] = typeCode
	}

	opts := options.Find().SetSort(bson.D{{Key: "assetcode", Value: 1}})
	cur, err := r.db.Collection("fixed_assets").Find(ctx, f, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var assets []Asset
	if err = cur.All(ctx, &assets); err != nil {
		return nil, err
	}

	// Fetch all depreciation items for these assets in the given fiscal year
	fDep := scopeFilter(scope)
	fDep["fiscalyear"] = fiscalYear
	if period > 0 {
		fDep["period"] = bson.M{"$lte": period}
	}
	fDep["isdeleted"] = false

	curDep, err := r.db.Collection("asset_depreciations").Find(ctx, fDep)
	var allDeprec []DepreciationScheduleItem
	if err == nil {
		_ = curDep.All(ctx, &allDeprec)
	}

	deprecByAsset := make(map[string][]DepreciationScheduleItem)
	for _, d := range allDeprec {
		deprecByAsset[d.AssetCode] = append(deprecByAsset[d.AssetCode], d)
	}

	// Fetch disposals
	fDisp := scopeFilter(scope)
	fDisp["isdeleted"] = false
	curDisp, err := r.db.Collection("asset_disposals").Find(ctx, fDisp)
	var allDisposals []AssetDisposal
	if err == nil {
		_ = curDisp.All(ctx, &allDisposals)
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

	f := scopeFilter(scope)
	f["isdeleted"] = false
	cur, err := r.db.Collection("fixed_assets").Find(ctx, f)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var assets []Asset
	_ = cur.All(ctx, &assets)

	// Fetch year depreciations
	fDep := scopeFilter(scope)
	fDep["fiscalyear"] = fiscalYear
	fDep["isdeleted"] = false
	curDep, _ := r.db.Collection("asset_depreciations").Find(ctx, fDep)
	var allDep []DepreciationScheduleItem
	_ = curDep.All(ctx, &allDep)

	depMap := make(map[string]decimal.Decimal)
	for _, d := range allDep {
		depMap[d.AssetCode] = depMap[d.AssetCode].Add(d.PeriodDeprec.Decimal())
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

		// Thai Tax Law Rules:
		// 1. Passenger cars capped at 1,000,000 THB depreciable base (กม. พรฎ. 315)
		cost := ast.Cost.Decimal()
		if ast.AssetTypeCode == "PASSENGER_CAR" && cost.GreaterThan(decimal.NewFromInt(1000000)) {
			// Tax deprec capped at 20% of 1,000,000 = 200,000/yr
			taxRate := decimal.NewFromFloat(0.20)
			taxDep = decimal.NewFromInt(1000000).Mul(taxRate)
			if taxDep.GreaterThan(acctDep) {
				taxDep = acctDep
			}
			remark = "ยานพาหนะนั่งไม่เกิน 10 ที่นั่ง จำกัดมูลค่าทางภาษี 1,000,000 บาท"
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
