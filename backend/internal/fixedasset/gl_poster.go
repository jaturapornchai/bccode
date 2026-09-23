package fixedasset

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	gl "smlcloudplatform/internal/generalledger"
)

// LedgerPoster is the subset of *generalledger.Store used to post fixed-asset
// journals. Every write goes through the shared GL engine so validation, the
// closed-period guard and the PostgreSQL transaction all run
// exactly once, in the one place that already implements them correctly.
type LedgerPoster interface {
	Execute(ctx context.Context, scope gl.Scope, cmd gl.Command) (gl.Result, error)
	List(ctx context.Context, scope gl.Scope, resource, query string, page, limit int, filter gl.ListFilter) (gl.Page, error)
}

type GLPoster struct {
	records *records
	ledger  LedgerPoster
}

func NewGLPoster(connect Connector, ledger LedgerPoster) *GLPoster {
	return &GLPoster{records: newRecords(connect), ledger: ledger}
}

// GLLineDoc / GLJournalDoc are the response shape returned to the HTTP and MCP
// callers. They are no longer persisted directly: the actual journal is
// written by generalledger.PostgresStore.Execute; these structs just describe what was posted.
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
	db, err := p.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	items, err := queryRecords[DepreciationScheduleItem](ctx, db, scope.Company, kindDepreciation, ` AND payload->>'fiscalyear' = $3 AND (payload->>'period')::int = $4 AND NOT COALESCE((payload->>'isposted')::boolean, false)`, ` ORDER BY code`, fiscalYear, period)
	if err != nil {
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
	assets, err := queryRecords[Asset](ctx, db, scope.Company, kindAsset, ` AND code = ANY($3)`, "", pq.Array(assetCodes))
	if err != nil {
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

	// 5. Post through the general ledger engine (validation + closed-period
	// guard + PostgreSQL sql.Tx all happen inside Execute).
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
	for _, it := range items {
		it.IsPosted = true
		it.JournalDocNo = docNo
		it.PostedAt = &now
		it.UpdatedAt = now
		it.UpdatedBy = scope.Actor
		if err := putRecord(ctx, db, scope.Company, kindDepreciation, it.ID, depreciationKey(it), it); err != nil {
			return nil, fmt.Errorf("ผ่านรายการสำเร็จแต่ไม่สามารถอัปเดตสถานะค่าเสื่อมราคา: %w", err)
		}
	}

	return journal, nil
}

// ReverseDepreciation cancels a posted depreciation journal and resets asset depreciation flags.
func (p *GLPoster) ReverseDepreciation(ctx context.Context, scope Scope, docNo, reason string, now time.Time) error {
	journal, err := p.findJournal(ctx, scope, docNo)
	if err != nil {
		return err
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
	db, err := p.records.db(ctx, scope.Holding)
	if err != nil {
		return err
	}
	items, err := queryRecords[DepreciationScheduleItem](ctx, db, scope.Company, kindDepreciation, ` AND payload->>'journaldocno' = $3`, "", docNo)
	if err != nil {
		return err
	}
	for _, it := range items {
		it.IsPosted = false
		it.JournalDocNo = ""
		it.PostedAt = nil
		it.UpdatedAt = now
		it.UpdatedBy = scope.Actor
		if err := putRecord(ctx, db, scope.Company, kindDepreciation, it.ID, depreciationKey(it), it); err != nil {
			return err
		}
	}
	return nil
}

// findJournal looks up a GL journal by its exact document number.
func (p *GLPoster) findJournal(ctx context.Context, scope Scope, docNo string) (*gl.Journal, error) {
	page, err := p.ledger.List(ctx, glScope(scope), "journals", docNo, 1, 100, gl.ListFilter{})
	if err != nil {
		return nil, fmt.Errorf("ไม่พบใบสำคัญสมุดรายวัน %s: %w", docNo, err)
	}
	for _, raw := range page.Items {
		var journal gl.Journal
		if err := json.Unmarshal(raw, &journal); err != nil {
			return nil, err
		}
		if journal.DocNo == docNo && !journal.IsDeleted {
			return &journal, nil
		}
	}
	return nil, fmt.Errorf("ไม่พบใบสำคัญสมุดรายวัน %s", docNo)
}

// DisposeAsset processes asset sale/write-off and generates the GL journal entry.
func (p *GLPoster) DisposeAsset(ctx context.Context, scope Scope, disposal AssetDisposal, now time.Time) (*AssetDisposal, *GLJournalDoc, error) {
	if CleanString(disposal.AssetCode) == "" {
		return nil, nil, fmt.Errorf("กรุณาระบุรหัสสินทรัพย์ที่ต้องการจำหน่าย")
	}
	if disposal.DisposalDate == "" {
		disposal.DisposalDate = now.Format("2006-01-02")
	}
	// ต้องตรวจรูปแบบวันที่ก่อนใช้งาน เพราะปีงบประมาณถูกตัดมาจากสี่ตัวอักษรแรกของสตริงนี้
	// วันที่สั้นกว่าที่ควรจะทำให้โปรแกรมหยุดทำงานทั้งตัว
	if _, err := time.Parse("2006-01-02", disposal.DisposalDate); err != nil {
		return nil, nil, fmt.Errorf("รูปแบบวันที่จำหน่ายไม่ถูกต้อง ต้องเป็น ปปปป-ดด-วว: %s", disposal.DisposalDate)
	}
	if disposal.DocNo == "" {
		disposal.DocNo = fmt.Sprintf("DISP-%s", disposal.AssetCode)
	}

	// 1. Load Asset
	db, err := p.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, nil, err
	}
	loaded, err := firstRecord[Asset](ctx, db, scope.Company, kindAsset, ` AND code = $3`, disposal.AssetCode)
	if err != nil {
		return nil, nil, fmt.Errorf("ไม่พบสินทรัพย์รหัส %s: %w", disposal.AssetCode, err)
	}
	asset := *loaded
	if asset.Status == "disposed" {
		return nil, nil, fmt.Errorf("สินทรัพย์รหัส %s ถูกจำหน่ายไปแล้ว", disposal.AssetCode)
	}

	// 2. Calculate Total Accumulated Depreciation up to disposal date
	latestDeprec, err := queryRecords[DepreciationScheduleItem](ctx, db, scope.Company, kindDepreciation, ` AND payload->>'assetcode' = $3 AND payload->>'stopdate' <= $4`, ` ORDER BY payload->>'stopdate' DESC LIMIT 1`, disposal.AssetCode, disposal.DisposalDate)
	if err != nil {
		return nil, nil, err
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
	disposal.Identity = identityFor(scope, kindDisposal, disposal.DocNo, Identity{}, now)
	if err := putRecord(ctx, db, scope.Company, kindDisposal, disposal.ID, disposal.DocNo, disposal); err != nil {
		return nil, nil, fmt.Errorf("ไม่สามารถบันทึกประวัติการจำหน่าย: %w", err)
	}

	// 6. Update Asset Status to Disposed
	// ขายสินทรัพย์แล้วสถานะต้องเปลี่ยนจริง ถ้าอัปเดตไม่สำเร็จแล้วเงียบไว้
	// สินทรัพย์ที่ขายไปแล้วจะยังคิดค่าเสื่อมราคาต่อและถูกขายซ้ำได้
	asset.Status = "disposed"
	asset.UpdatedAt = now
	asset.UpdatedBy = scope.Actor
	if err := putRecord(ctx, db, scope.Company, kindAsset, asset.ID, asset.AssetCode, asset); err != nil {
		return nil, nil, fmt.Errorf("บันทึกการจำหน่ายสำเร็จแต่ไม่สามารถเปลี่ยนสถานะสินทรัพย์เป็นจำหน่ายแล้ว: %w", err)
	}

	return &disposal, journal, nil
}
