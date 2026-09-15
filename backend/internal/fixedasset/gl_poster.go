package fixedasset

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	gl "smlcloudplatform/internal/generalledger"
)

// LedgerPoster is the subset of *generalledger.Store used to post fixed-asset
// journals. Every write goes through the shared GL engine so validation, the
// closed-period guard, the Mongo outbox and PostgreSQL projection all run
// exactly once, in the one place that already implements them correctly.
type LedgerPoster interface {
	Execute(ctx context.Context, scope gl.Scope, cmd gl.Command) (gl.Result, error)
}

type GLPoster struct {
	db     *mongo.Database
	ledger LedgerPoster
}

func NewGLPoster(db *mongo.Database, ledger LedgerPoster) *GLPoster {
	return &GLPoster{db: db, ledger: ledger}
}

// GLLineDoc / GLJournalDoc are the response shape returned to the HTTP and MCP
// callers. They are no longer persisted directly: the actual journal is
// written by generalledger.Store.Execute (Mongo gl_journals + PostgreSQL
// projection); these structs just describe what was posted.
type GLLineDoc struct {
	AccountCode string          `json:"accountcode"`
	Description string          `json:"description"`
	Debit       decimal.Decimal `json:"debit"`
	Credit      decimal.Decimal `json:"credit"`
	BranchCode  string          `json:"branchcode"`
}

type GLJournalDoc struct {
	ID           string      `json:"id"`
	HoldingCode  string      `json:"holdingcode"`
	BusinessCode string      `json:"businesscode"`
	Version      int64       `json:"version"`
	DocNo        string      `json:"docno"`
	Date         string      `json:"date"`
	BookCode     string      `json:"bookcode"`
	FiscalYear   string      `json:"fiscalyear"`
	Description  string      `json:"description"`
	Reference    string      `json:"reference"`
	BranchCode   string      `json:"branchcode"`
	Kind         string      `json:"kind"`
	Status       string      `json:"status"`
	Lines        []GLLineDoc `json:"lines"`
	CreatedAt    time.Time   `json:"createdat"`
	CreatedBy    string      `json:"createdby"`
	UpdatedAt    time.Time   `json:"updatedat"`
	UpdatedBy    string      `json:"updatedby"`
	PostedAt     *time.Time  `json:"postedat,omitempty"`
	PostedBy     string      `json:"postedby,omitempty"`
	IsDeleted    bool        `json:"isdeleted"`
}

func glScope(s Scope) gl.Scope {
	return gl.Scope{Holding: s.Holding, Company: s.Company, Branch: s.Branch, Actor: s.Actor}
}

func glAmount(d decimal.Decimal) gl.Amount { return gl.Amount(d.String()) }

// postJournal creates a draft journal through the GL engine, then posts it, in
// two idempotent Execute calls keyed off the (deterministic) docNo. A retry
// with the same docNo and the same lines replays the cached GL event instead
// of creating a duplicate journal; a retry with different lines under the same
// docNo is rejected by generalledger.Store.Execute, never silently applied.
func (p *GLPoster) postJournal(ctx context.Context, scope Scope, j *gl.Journal) (gl.Result, gl.Result, error) {
	sc := glScope(scope)
	createReq := digest([]byte("fa-gl-create:" + scope.Holding + ":" + scope.Company + ":" + j.DocNo))
	created, err := p.ledger.Execute(ctx, sc, gl.Command{
		Resource:  "journals",
		Action:    "create",
		RequestID: createReq,
		Journal:   j,
	})
	if err != nil {
		return gl.Result{}, gl.Result{}, fmt.Errorf("ไม่สามารถบันทึกสมุดรายวัน GL: %w", err)
	}
	postReq := digest([]byte("fa-gl-post:" + scope.Holding + ":" + scope.Company + ":" + j.DocNo))
	posted, err := p.ledger.Execute(ctx, sc, gl.Command{
		Resource:  "journals",
		Action:    "post",
		ID:        created.ID,
		Version:   created.Version,
		RequestID: postReq,
	})
	if err != nil {
		return gl.Result{}, gl.Result{}, fmt.Errorf("ไม่สามารถผ่านรายการสมุดรายวัน GL: %w", err)
	}
	return created, posted, nil
}

// PostDepreciation posts monthly depreciation for a specific year and period into GL Journal.
func (p *GLPoster) PostDepreciation(ctx context.Context, scope Scope, fiscalYear string, period int, date, docNo string, now time.Time) (*GLJournalDoc, error) {
	if fiscalYear == "" || period < 1 || period > 12 {
		return nil, fmt.Errorf("กรุณาระบุปีบัญชีและงวดที่ต้องการผ่านรายการ (1-12)")
	}
	if docNo == "" {
		docNo = fmt.Sprintf("JV-FA-%s-%02d", fiscalYear, period)
	}
	if date == "" {
		date = now.Format("2006-01-02")
	}

	// 1. Find all unposted depreciation schedule items for this year & period
	f := scopeFilter(scope)
	f["fiscalyear"] = fiscalYear
	f["period"] = period
	f["isposted"] = false
	f["isdeleted"] = false

	cur, err := p.db.Collection("asset_depreciations").Find(ctx, f)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var items []DepreciationScheduleItem
	if err = cur.All(ctx, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("ไม่พบรายการค่าเสื่อมราคาที่ยังไม่ได้ผ่านรายการสำหรับปี %s งวดที่ %d", fiscalYear, period)
	}

	// 2. Fetch asset details to get account codes
	assetCodes := make([]string, len(items))
	for i, it := range items {
		assetCodes[i] = it.AssetCode
	}

	fAssets := scopeFilter(scope)
	fAssets["assetcode"] = bson.M{"$in": assetCodes}
	fAssets["isdeleted"] = false

	curAssets, err := p.db.Collection("fixed_assets").Find(ctx, fAssets)
	if err != nil {
		return nil, err
	}
	defer curAssets.Close(ctx)

	var assets []Asset
	if err = curAssets.All(ctx, &assets); err != nil {
		return nil, err
	}

	assetMap := make(map[string]Asset)
	for _, a := range assets {
		assetMap[a.AssetCode] = a
	}

	// 3. Group depreciation amounts by Expense Account and Accumulated Depreciation Account.
	// Branch is not a GL journal-line dimension, so every line here shares the
	// journal's own branch (the acting user's scope); see gl_poster.go P0-1/P0-2 fix notes.
	type accGroup struct {
		expenseAcc string
		accumAcc   string
	}
	grouped := make(map[accGroup]decimal.Decimal)

	for _, it := range items {
		ast, ok := assetMap[it.AssetCode]
		if !ok {
			continue
		}
		expAcc := ast.DeprecExpenseAccountCode
		if expAcc == "" {
			expAcc = "520103" // Default Depreciation Expense
		}
		accAcc := ast.AccumDeprecAccountCode
		if accAcc == "" {
			accAcc = "129101" // Default Accum Depreciation
		}

		key := accGroup{expenseAcc: expAcc, accumAcc: accAcc}
		grouped[key] = grouped[key].Add(it.PeriodDeprec.Decimal())
	}

	// 4. Build GL Lines
	var lines []GLLineDoc
	totalDebit := decimal.Zero
	totalCredit := decimal.Zero

	for k, amount := range grouped {
		if amount.LessThanOrEqual(decimal.Zero) {
			continue
		}
		// Debit Depreciation Expense
		lines = append(lines, GLLineDoc{
			AccountCode: k.expenseAcc,
			Description: fmt.Sprintf("ค่าเสื่อมราคาประจำงวด %d/%s", period, fiscalYear),
			Debit:       amount,
			Credit:      decimal.Zero,
			BranchCode:  scope.Branch,
		})
		totalDebit = totalDebit.Add(amount)

		// Credit Accumulated Depreciation
		lines = append(lines, GLLineDoc{
			AccountCode: k.accumAcc,
			Description: fmt.Sprintf("ค่าเสื่อมราคาสะสมประจำงวด %d/%s", period, fiscalYear),
			Debit:       decimal.Zero,
			Credit:      amount,
			BranchCode:  scope.Branch,
		})
		totalCredit = totalCredit.Add(amount)
	}

	if !totalDebit.Equal(totalCredit) {
		return nil, fmt.Errorf("ยอดเดบิต (%.2f) และเครดิต (%.2f) ไม่สมดุลกัน", totalDebit.InexactFloat64(), totalCredit.InexactFloat64())
	}
	if totalDebit.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("ยอดรวมค่าเสื่อมราคาเป็น 0")
	}

	// 5. Post through the general ledger engine (Mongo write + closed-period
	// guard + outbox + PostgreSQL projection all happen inside Execute).
	glLines := make([]gl.Line, 0, len(lines))
	for _, l := range lines {
		glLines = append(glLines, gl.Line{
			AccountCode: l.AccountCode,
			Description: l.Description,
			Debit:       glAmount(l.Debit),
			Credit:      glAmount(l.Credit),
		})
	}
	description := fmt.Sprintf("บันทึกค่าเสื่อมราคาสินทรัพย์ประจำงวด %d/%s", period, fiscalYear)
	reference := fmt.Sprintf("FA-%s-%02d", fiscalYear, period)
	journalInput := &gl.Journal{
		DocNo:       docNo,
		Date:        date,
		BookCode:    "JV",
		FiscalYear:  fiscalYear,
		Description: description,
		Reference:   reference,
		BranchCode:  scope.Branch,
		Kind:        "manual",
		Lines:       glLines,
	}
	created, posted, err := p.postJournal(ctx, scope, journalInput)
	if err != nil {
		return nil, err
	}

	journal := &GLJournalDoc{
		ID:           created.ID,
		HoldingCode:  scope.Holding,
		BusinessCode: scope.Company,
		Version:      posted.Version,
		DocNo:        docNo,
		Date:         date,
		BookCode:     "JV",
		FiscalYear:   fiscalYear,
		Description:  description,
		Reference:    reference,
		BranchCode:   scope.Branch,
		Kind:         "manual",
		Status:       "posted",
		Lines:        lines,
		CreatedAt:    now,
		CreatedBy:    scope.Actor,
		UpdatedAt:    now,
		UpdatedBy:    scope.Actor,
		PostedAt:     &now,
		PostedBy:     scope.Actor,
		IsDeleted:    false,
	}

	// 6. Update depreciation items to mark as posted
	itemIDs := make([]string, len(items))
	for i, it := range items {
		itemIDs[i] = it.ID
	}
	fUpdate := scopeFilter(scope)
	fUpdate["_id"] = bson.M{"$in": itemIDs}

	u := bson.M{
		"$set": bson.M{
			"isposted":     true,
			"journaldocno": docNo,
			"postedat":     now,
			"updatedat":    now,
			"updatedby":    scope.Actor,
		},
	}
	_, err = p.db.Collection("asset_depreciations").UpdateMany(ctx, fUpdate, u)
	if err != nil {
		return nil, fmt.Errorf("ผ่านรายการสำเร็จแต่ไม่สามารถอัปเดตสถานะค่าเสื่อมราคา: %w", err)
	}

	return journal, nil
}

// ReverseDepreciation cancels a posted depreciation journal and resets asset depreciation flags.
func (p *GLPoster) ReverseDepreciation(ctx context.Context, scope Scope, docNo, reason string, now time.Time) error {
	fJ := scopeFilter(scope)
	fJ["docno"] = docNo
	fJ["isdeleted"] = false

	var journal gl.Journal
	err := p.db.Collection("gl_journals").FindOne(ctx, fJ).Decode(&journal)
	if err != nil {
		return fmt.Errorf("ไม่พบใบสำคัญสมุดรายวัน %s: %w", docNo, err)
	}
	if journal.Status != "posted" {
		return fmt.Errorf("ใบสำคัญ %s ไม่ได้อยู่ในสถานะผ่านรายการ", docNo)
	}

	// 1. Reverse through the general ledger engine: it flips debit/credit,
	// keeps the original journal for audit, and re-projects PostgreSQL.
	reversalDocNo := "REV-" + docNo
	if len(reversalDocNo) > 60 {
		reversalDocNo = reversalDocNo[:60]
	}
	reverseDate := now.Format("2006-01-02")
	if reverseDate < journal.Date {
		reverseDate = journal.Date
	}
	requestID := digest([]byte("fa-gl-reverse:" + scope.Holding + ":" + scope.Company + ":" + docNo))
	_, err = p.ledger.Execute(ctx, glScope(scope), gl.Command{
		Resource:  "journals",
		Action:    "reverse",
		ID:        journal.ID,
		Version:   journal.Version,
		DocNo:     reversalDocNo,
		Date:      reverseDate,
		Reason:    reason,
		RequestID: requestID,
	})
	if err != nil {
		return fmt.Errorf("ไม่สามารถกลับรายการสมุดรายวัน GL: %w", err)
	}

	// 2. Reset asset depreciation items
	fDep := scopeFilter(scope)
	fDep["journaldocno"] = docNo

	uDep := bson.M{
		"$set": bson.M{
			"isposted":     false,
			"journaldocno": "",
			"postedat":     nil,
			"updatedat":    now,
			"updatedby":    scope.Actor,
		},
	}
	_, err = p.db.Collection("asset_depreciations").UpdateMany(ctx, fDep, uDep)
	return err
}

// DisposeAsset processes asset sale/write-off and generates the GL journal entry.
func (p *GLPoster) DisposeAsset(ctx context.Context, scope Scope, disposal AssetDisposal, now time.Time) (*AssetDisposal, *GLJournalDoc, error) {
	if CleanString(disposal.AssetCode) == "" {
		return nil, nil, fmt.Errorf("กรุณาระบุรหัสสินทรัพย์ที่ต้องการจำหน่าย")
	}
	if disposal.DisposalDate == "" {
		disposal.DisposalDate = now.Format("2006-01-02")
	}
	if disposal.DocNo == "" {
		disposal.DocNo = fmt.Sprintf("DISP-%s", disposal.AssetCode)
	}

	// 1. Load Asset
	var asset Asset
	f := scopeFilter(scope)
	f["assetcode"] = disposal.AssetCode
	f["isdeleted"] = false
	err := p.db.Collection("fixed_assets").FindOne(ctx, f).Decode(&asset)
	if err != nil {
		return nil, nil, fmt.Errorf("ไม่พบสินทรัพย์รหัส %s: %w", disposal.AssetCode, err)
	}
	if asset.Status == "disposed" {
		return nil, nil, fmt.Errorf("สินทรัพย์รหัส %s ถูกจำหน่ายไปแล้ว", disposal.AssetCode)
	}

	// 2. Calculate Total Accumulated Depreciation up to disposal date
	fDep := scopeFilter(scope)
	fDep["assetcode"] = disposal.AssetCode
	fDep["stopdate"] = bson.M{"$lte": disposal.DisposalDate}
	fDep["isdeleted"] = false

	opts := options.Find().SetSort(bson.D{{Key: "stopdate", Value: -1}}).SetLimit(1)
	cur, err := p.db.Collection("asset_depreciations").Find(ctx, fDep, opts)
	var latestDeprec []DepreciationScheduleItem
	if err == nil {
		_ = cur.All(ctx, &latestDeprec)
	}

	accumDeprec := asset.BeginAccumDeprec.Decimal()
	if len(latestDeprec) > 0 {
		accumDeprec = latestDeprec[0].AccumDeprec.Decimal()
	}

	cost := asset.Cost.Decimal()
	netBookValue := cost.Sub(accumDeprec)
	if netBookValue.LessThan(decimal.Zero) {
		netBookValue = decimal.Zero
	}

	// Gain/Loss = SalePrice - NetBookValue
	salePrice := disposal.SalePrice.Decimal()
	vatAmount := disposal.VatAmount.Decimal()
	gainLoss := salePrice.Sub(netBookValue)

	disposal.AccumDeprecAtDisposal = AmountFromDecimal(accumDeprec)
	disposal.NetBookValueAtDisposal = AmountFromDecimal(netBookValue)
	disposal.GainLoss = AmountFromDecimal(gainLoss)

	// 3. Build Disposal GL Journal
	var lines []GLLineDoc
	settlementAcc := disposal.SettlementAccountCode
	if settlementAcc == "" {
		settlementAcc = "110101" // Default Cash
	}
	costAcc := asset.AssetAccountCode
	if costAcc == "" {
		costAcc = "120101" // Default Asset Cost
	}
	accumAcc := asset.AccumDeprecAccountCode
	if accumAcc == "" {
		accumAcc = "129101" // Default Accum Deprec
	}

	// Dr. Cash / AR (Sale Price + VAT)
	totalReceipt := salePrice.Add(vatAmount)
	if totalReceipt.GreaterThan(decimal.Zero) {
		lines = append(lines, GLLineDoc{
			AccountCode: settlementAcc,
			Description: fmt.Sprintf("รับชำระเงินค่าจำหน่ายสินทรัพย์ %s", asset.AssetCode),
			Debit:       totalReceipt,
			Credit:      decimal.Zero,
			BranchCode:  asset.BranchCode,
		})
	}

	// Dr. Accumulated Depreciation
	if accumDeprec.GreaterThan(decimal.Zero) {
		lines = append(lines, GLLineDoc{
			AccountCode: accumAcc,
			Description: fmt.Sprintf("ตัดค่าเสื่อมราคาสะสม สินทรัพย์ %s", asset.AssetCode),
			Debit:       accumDeprec,
			Credit:      decimal.Zero,
			BranchCode:  asset.BranchCode,
		})
	}

	// Dr. Loss on Disposal (if GainLoss < 0)
	if gainLoss.LessThan(decimal.Zero) {
		lossAcc := disposal.GainLossAccountCode
		if lossAcc == "" {
			lossAcc = "530101" // Default Loss on Asset Disposal
		}
		lines = append(lines, GLLineDoc{
			AccountCode: lossAcc,
			Description: fmt.Sprintf("ขาดทุนจากการจำหน่ายสินทรัพย์ %s", asset.AssetCode),
			Debit:       gainLoss.Abs(),
			Credit:      decimal.Zero,
			BranchCode:  asset.BranchCode,
		})
	}

	// Cr. Asset Cost
	lines = append(lines, GLLineDoc{
		AccountCode: costAcc,
		Description: fmt.Sprintf("ตัดราคาทุนสินทรัพย์ %s", asset.AssetCode),
		Debit:       decimal.Zero,
		Credit:      cost,
		BranchCode:  asset.BranchCode,
	})

	// Cr. Gain on Disposal (if GainLoss > 0)
	if gainLoss.GreaterThan(decimal.Zero) {
		gainAcc := disposal.GainLossAccountCode
		if gainAcc == "" {
			gainAcc = "420101" // Default Gain on Asset Disposal
		}
		lines = append(lines, GLLineDoc{
			AccountCode: gainAcc,
			Description: fmt.Sprintf("กำไรจากการจำหน่ายสินทรัพย์ %s", asset.AssetCode),
			Debit:       decimal.Zero,
			Credit:      gainLoss,
			BranchCode:  asset.BranchCode,
		})
	}

	// Cr. Output VAT (if any)
	if vatAmount.GreaterThan(decimal.Zero) {
		lines = append(lines, GLLineDoc{
			AccountCode: "210301", // Output VAT
			Description: fmt.Sprintf("ภาษีขายจากการจำหน่ายสินทรัพย์ %s", asset.AssetCode),
			Debit:       decimal.Zero,
			Credit:      vatAmount,
			BranchCode:  asset.BranchCode,
		})
	}

	// Verify balance
	totDr := decimal.Zero
	totCr := decimal.Zero
	for _, l := range lines {
		totDr = totDr.Add(l.Debit)
		totCr = totCr.Add(l.Credit)
	}
	if !totDr.Equal(totCr) {
		return nil, nil, fmt.Errorf("รายการจำหน่ายสินทรัพย์ไม่สมดุล: เดบิต (%.2f) != เครดิต (%.2f)", totDr.InexactFloat64(), totCr.InexactFloat64())
	}

	journalDocNo := fmt.Sprintf("JV-DISP-%s", asset.AssetCode)
	disposal.JournalDocNo = journalDocNo

	// 4. Post through the general ledger engine
	glLines := make([]gl.Line, 0, len(lines))
	for _, l := range lines {
		glLines = append(glLines, gl.Line{
			AccountCode: l.AccountCode,
			Description: l.Description,
			Debit:       glAmount(l.Debit),
			Credit:      glAmount(l.Credit),
		})
	}
	description := fmt.Sprintf("บันทึกจำหน่ายสินทรัพย์ %s (%s)", asset.AssetCode, asset.ThaiName())
	journalInput := &gl.Journal{
		DocNo:       journalDocNo,
		Date:        disposal.DisposalDate,
		BookCode:    "JV",
		FiscalYear:  disposal.DisposalDate[:4],
		Description: description,
		Reference:   disposal.DocNo,
		BranchCode:  asset.BranchCode,
		Kind:        "manual",
		Lines:       glLines,
	}
	created, posted, err := p.postJournal(ctx, scope, journalInput)
	if err != nil {
		return nil, nil, err
	}

	journal := &GLJournalDoc{
		ID:           created.ID,
		HoldingCode:  scope.Holding,
		BusinessCode: scope.Company,
		Version:      posted.Version,
		DocNo:        journalDocNo,
		Date:         disposal.DisposalDate,
		BookCode:     "JV",
		FiscalYear:   disposal.DisposalDate[:4],
		Description:  description,
		Reference:    disposal.DocNo,
		BranchCode:   asset.BranchCode,
		Kind:         "manual",
		Status:       "posted",
		Lines:        lines,
		CreatedAt:    now,
		CreatedBy:    scope.Actor,
		UpdatedAt:    now,
		UpdatedBy:    scope.Actor,
		PostedAt:     &now,
		PostedBy:     scope.Actor,
		IsDeleted:    false,
	}

	// 5. Save Disposal Record
	disposal.Identity = identityFor(scope, "disposals", disposal.DocNo, Identity{}, now)
	_, err = p.db.Collection("asset_disposals").InsertOne(ctx, disposal)
	if err != nil {
		return nil, nil, fmt.Errorf("ไม่สามารถบันทึกประวัติการจำหน่าย: %w", err)
	}

	// 6. Update Asset Status to Disposed
	uAsset := bson.M{
		"$set": bson.M{
			"status":    "disposed",
			"updatedat": now,
			"updatedby": scope.Actor,
		},
	}
	_, _ = p.db.Collection("fixed_assets").UpdateOne(ctx, f, uAsset)

	return &disposal, journal, nil
}
