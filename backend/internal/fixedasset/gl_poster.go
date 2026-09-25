package fixedasset

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

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

// BranchChecker validates a voucher header branch for the session. HTTP wires it to
// generalledger/httpapi.CheckJournalBranch so fixed-asset journals follow the GL screen rule:
// a branch session may leave it blank (= its own branch) but not pick another branch; a
// company-wide session must name an active branch of the company.
type BranchChecker func(ctx context.Context, scope Scope, branch string) error

type GLPoster struct {
	records     *records
	ledger      LedgerPoster
	checkBranch BranchChecker
}

func NewGLPoster(connect Connector, ledger LedgerPoster, checkBranch BranchChecker) *GLPoster {
	return &GLPoster{records: newRecords(connect), ledger: ledger, checkBranch: checkBranch}
}

var errBranchCheckUnavailable = errors.New("ระบบตรวจสาขาของใบสำคัญยังไม่พร้อม กรุณาติดต่อผู้ดูแลระบบ")

// Labels of the account settings a user must fill; the asset's own account wins over its type's.
const (
	labelCostAccount    = "รหัสบัญชีสินทรัพย์ (ราคาทุน)"
	labelAccumAccount   = "รหัสบัญชีค่าเสื่อมราคาสะสม"
	labelExpenseAccount = "รหัสบัญชีค่าใช้จ่ายค่าเสื่อมราคา"
)

// faFieldError is a user-input failure the screen can point at (code + JSON field + Thai text).
func faFieldError(code, field, message string) error {
	return &gl.UserError{Code: code, Field: field, Status: http.StatusBadRequest, Message: message}
}

// journalBranch validates the voucher header branch and returns the branch to store; blank
// means the session branch, which the check only allows for a branch session.
func (p *GLPoster) journalBranch(ctx context.Context, scope Scope, branch string) (string, error) {
	branch = gl.NormalizeCode(branch)
	if p.checkBranch == nil {
		return "", errBranchCheckUnavailable
	}
	if err := p.checkBranch(ctx, scope, branch); err != nil {
		return "", err
	}
	if branch == "" {
		branch = scope.Branch
	}
	return branch, nil
}

// disposalBranch posts a disposal in the asset's branch under the same rule, with messages that
// point at the asset: the disposal dialog has no branch field of its own.
func (p *GLPoster) disposalBranch(ctx context.Context, scope Scope, asset Asset) (string, error) {
	branch, err := p.journalBranch(ctx, scope, asset.BranchCode)
	user, ok := gl.AsUserError(err)
	if !ok {
		return branch, err
	}
	switch user.Code {
	case "journal_branch_required":
		return "", faFieldError(user.Code, "branchcode", fmt.Sprintf("สินทรัพย์ %s ยังไม่ได้กำหนดสาขา และคุณเข้าระบบระดับบริษัท ระบบจึงเลือกสาขาของใบสำคัญให้ไม่ได้ กรุณากำหนดสาขาในข้อมูลสินทรัพย์ หรือเข้าระบบในสาขาที่ต้องการ", asset.AssetCode))
	case "journal_branch_outside_session":
		return "", faFieldError(user.Code, "branchcode", fmt.Sprintf("สินทรัพย์ %s อยู่สาขา “%s” ไม่ตรงกับสาขาที่เข้าระบบ “%s” กรุณาเข้าระบบในสาขาของสินทรัพย์หรือระดับบริษัท แล้วจำหน่ายอีกครั้ง", asset.AssetCode, gl.NormalizeCode(asset.BranchCode), scope.Branch))
	}
	return "", err
}

// assetAccounts are the GL accounts of one asset, resolved from the asset, then its asset type.
type assetAccounts struct{ cost, accum, expense string }

// resolveAssetAccounts never guesses a code: charts of accounts differ per company, so an
// account missing on both the asset and its type stays blank and the caller asks the user.
func resolveAssetAccounts(asset Asset, assetType *AssetType) assetAccounts {
	var t AssetType
	if assetType != nil {
		t = *assetType
	}
	pick := func(own, typed string) string {
		if code := gl.NormalizeCode(own); code != "" {
			return code
		}
		return gl.NormalizeCode(typed)
	}
	return assetAccounts{
		cost:    pick(asset.AssetAccountCode, t.AssetAccountCode),
		accum:   pick(asset.AccumDeprecAccountCode, t.AccumDeprecAccountCode),
		expense: pick(asset.DeprecExpenseAccountCode, t.DeprecExpenseAccountCode),
	}
}

// assetAccountMissing names the account setting and the assets that lack it.
func assetAccountMissing(field, label string, assetCodes []string) error {
	shown, more := assetCodes, ""
	if len(shown) > 10 {
		shown, more = shown[:10], fmt.Sprintf(" และอีก %d รายการ", len(assetCodes)-10)
	}
	return faFieldError("fa_account_required", field, fmt.Sprintf("ยังไม่ได้กำหนด%sของสินทรัพย์ %s%s กรุณากำหนดที่ข้อมูลสินทรัพย์หรือประเภทสินทรัพย์ แล้วทำรายการอีกครั้ง", label, strings.Join(shown, ", "), more))
}

// assetTypesByCode loads the asset types the assets refer to, keyed by type code.
func assetTypesByCode(ctx context.Context, q queryer, company string, assets []Asset) (map[string]*AssetType, error) {
	byCode := map[string]*AssetType{}
	seen := map[string]bool{}
	codes := []string{}
	for _, a := range assets {
		if a.AssetTypeCode != "" && !seen[a.AssetTypeCode] {
			seen[a.AssetTypeCode] = true
			codes = append(codes, a.AssetTypeCode)
		}
	}
	if len(codes) == 0 {
		return byCode, nil
	}
	types, err := queryRecords[AssetType](ctx, q, company, kindType, ` AND code = ANY($3)`, "", pq.Array(codes))
	if err != nil {
		return nil, err
	}
	for i := range types {
		byCode[types[i].TypeCode] = &types[i]
	}
	return byCode, nil
}

// checkJournalDocNo validates the voucher number before anything is posted. A number the
// system builds from the book and asset codes can outgrow doc_no VARCHAR(30) (counted in runes),
// so that failure says what was built and how to fix it.
func checkJournalDocNo(docNo, field string, generated bool) error {
	err := gl.CheckDocNo(docNo, field)
	user, ok := gl.AsUserError(err)
	if !ok || !generated {
		return err
	}
	if user.Code == "code_too_long" {
		return faFieldError(user.Code, field, fmt.Sprintf("เลขที่ใบสำคัญที่ระบบสร้างให้ “%s” ยาว %d ตัวอักษร เกินที่รับได้ %d ตัว กรุณาระบุเลขที่ใบสำคัญเองให้ไม่เกิน %d ตัวอักษร", docNo, utf8.RuneCountInString(docNo), gl.DocNoMaxRunes, gl.DocNoMaxRunes))
	}
	return faFieldError(user.Code, field, fmt.Sprintf("เลขที่ใบสำคัญที่ระบบสร้างให้ “%s” ใช้ไม่ได้: %s กรุณาระบุเลขที่ใบสำคัญเอง", docNo, user.Message))
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
// branch is the voucher header branch (blank = the session branch, branch sessions only).
// fiscalYear/period pick the schedule rows (ค.ศ. calendar year + month); a blank date is the
// period's last day, and the journal goes into the GL fiscal year of that period.
func (p *GLPoster) PostDepreciation(ctx context.Context, scope Scope, fiscalYear string, period int, date, docNo, branch string, now time.Time) (*GLJournalDoc, error) {
	if strings.TrimSpace(fiscalYear) == "" || period < 1 || period > 12 {
		return nil, fmt.Errorf("กรุณาระบุปีบัญชีและงวดที่ต้องการผ่านรายการ (1-12)")
	}
	fiscalYear, date, periodEnd, err := depreciationVoucherDate(fiscalYear, period, date)
	if err != nil {
		return nil, err
	}
	years, err := p.fiscalYears(ctx, scope)
	if err != nil {
		return nil, err
	}
	glYear, err := depreciationFiscalYear(years, fiscalYear, period, date, periodEnd)
	if err != nil {
		return nil, err
	}
	branch, err = p.journalBranch(ctx, scope, branch)
	if err != nil {
		return nil, err
	}
	book, err := p.generalBookCode(ctx, scope)
	if err != nil {
		return nil, err
	}
	// Book codes are user-defined: the generated number starts with the chosen book, never "JV".
	docNo = gl.NormalizeCode(docNo)
	generated := docNo == ""
	if generated {
		docNo = fmt.Sprintf("%s-FA-%s-%02d", book, fiscalYear, period)
	}
	if err := checkJournalDocNo(docNo, "docno", generated); err != nil {
		return nil, err
	}

	// 1. Find all unposted depreciation schedule items for this year & period
	db, err := p.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	unlock, err := lockDepreciationPeriod(ctx, db, scope.Company, fiscalYear, period)
	if err != nil {
		return nil, err
	}
	defer unlock()
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
	types, err := assetTypesByCode(ctx, db, scope.Company, assets)
	if err != nil {
		return nil, err
	}

	// 3. Group depreciation amounts by Expense Account and Accumulated Depreciation Account.
	// Branch is not a GL journal-line dimension, so every line here shares the journal's
	// header branch (checked above).
	type accGroup struct {
		expenseAcc string
		accumAcc   string
	}
	grouped := make(map[accGroup]decimal.Decimal)
	var missingExpense, missingAccum []string

	for _, it := range items {
		amount := it.PeriodDeprec.Decimal()
		if amount.IsZero() {
			continue
		}
		ast, ok := assetMap[it.AssetCode]
		if !ok {
			// Skipping would still mark the row posted below with no journal line behind it.
			return nil, fmt.Errorf("ไม่พบข้อมูลสินทรัพย์ %s ของรายการค่าเสื่อมราคางวดนี้ กรุณาตรวจสอบทะเบียนสินทรัพย์ก่อนผ่านรายการ", it.AssetCode)
		}
		acc := resolveAssetAccounts(ast, types[ast.AssetTypeCode])
		if acc.expense == "" {
			missingExpense = append(missingExpense, ast.AssetCode)
		}
		if acc.accum == "" {
			missingAccum = append(missingAccum, ast.AssetCode)
		}
		key := accGroup{expenseAcc: acc.expense, accumAcc: acc.accum}
		grouped[key] = grouped[key].Add(amount)
	}
	if len(missingExpense) > 0 {
		return nil, assetAccountMissing("deprecexpenseaccountcode", labelExpenseAccount, missingExpense)
	}
	if len(missingAccum) > 0 {
		return nil, assetAccountMissing("accumdeprecaccountcode", labelAccumAccount, missingAccum)
	}

	// 4. Build GL Lines in a stable order: map order is random, and a retry must send the same
	// lines under the same docno or the GL engine rejects it as a different journal.
	keys := make([]accGroup, 0, len(grouped))
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].expenseAcc != keys[j].expenseAcc {
			return keys[i].expenseAcc < keys[j].expenseAcc
		}
		return keys[i].accumAcc < keys[j].accumAcc
	})
	var lines []GLLineDoc
	totalDebit := decimal.Zero
	totalCredit := decimal.Zero

	for _, k := range keys {
		amount := grouped[k]
		if amount.LessThanOrEqual(decimal.Zero) {
			continue
		}
		// Debit Depreciation Expense
		lines = append(lines, GLLineDoc{
			AccountCode: k.expenseAcc,
			Description: fmt.Sprintf("ค่าเสื่อมราคาประจำงวด %d/%s", period, fiscalYear),
			Debit:       amount,
			Credit:      decimal.Zero,
			BranchCode:  branch,
		})
		totalDebit = totalDebit.Add(amount)

		// Credit Accumulated Depreciation
		lines = append(lines, GLLineDoc{
			AccountCode: k.accumAcc,
			Description: fmt.Sprintf("ค่าเสื่อมราคาสะสมประจำงวด %d/%s", period, fiscalYear),
			Debit:       decimal.Zero,
			Credit:      amount,
			BranchCode:  branch,
		})
		totalCredit = totalCredit.Add(amount)
	}

	if !totalDebit.Equal(totalCredit) {
		return nil, fmt.Errorf("ยอดเดบิต (%s) และเครดิต (%s) ไม่สมดุลกัน", totalDebit.String(), totalCredit.String())
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
	docNo, err = p.depreciationDocNo(ctx, scope, docNo, reference, generated, func(candidate string) (bool, error) {
		return depreciationRowsMarked(ctx, db, scope.Company, candidate)
	})
	if err != nil {
		return nil, err
	}
	journalInput := &gl.Journal{
		DocNo:       docNo,
		Date:        date,
		BookCode:    book,
		FiscalYear:  glYear,
		Description: description,
		Reference:   reference,
		BranchCode:  branch,
		Kind:        "manual",
		Lines:       glLines,
	}
	created, posted, err := p.postJournal(ctx, scope, journalInput)
	if err != nil {
		return nil, err
	}
	if err := p.checkJournalPosted(ctx, scope, docNo); err != nil {
		return nil, err
	}

	journal := &GLJournalDoc{
		ID:           created.ID,
		HoldingCode:  scope.Holding,
		BusinessCode: scope.Company,
		Version:      posted.Version,
		DocNo:        docNo,
		Date:         date,
		BookCode:     book,
		FiscalYear:   glYear,
		Description:  description,
		Reference:    reference,
		BranchCode:   branch,
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

	// 6. Mark the period's schedule rows posted (all or none)
	if err := setDepreciationPosted(ctx, db, scope, items, docNo, now); err != nil {
		return nil, fmt.Errorf("ผ่านรายการสำเร็จแต่ไม่สามารถอัปเดตสถานะค่าเสื่อมราคา: %w", err)
	}

	return journal, nil
}

// ReverseDepreciation cancels a posted depreciation journal and resets asset depreciation flags.
func (p *GLPoster) ReverseDepreciation(ctx context.Context, scope Scope, docNo, reason string, now time.Time) error {
	journal, err := p.findJournal(ctx, scope, docNo)
	if err != nil {
		return err
	}
	// Already reversed (an earlier attempt whose row reset failed, or a reversal made in GL):
	// only the schedule rows are still to be freed.
	alreadyReversed := journal.Status == journalStatusReversed
	if !alreadyReversed && journal.Status != journalStatusPosted {
		return fmt.Errorf("ใบสำคัญ %s ไม่ได้อยู่ในสถานะผ่านรายการ", docNo)
	}
	if !alreadyReversed {
		if err := p.reverseJournal(ctx, scope, journal, reason, now); err != nil {
			return err
		}
	}

	db, err := p.records.db(ctx, scope.Holding)
	if err != nil {
		return err
	}
	items, err := queryRecords[DepreciationScheduleItem](ctx, db, scope.Company, kindDepreciation, ` AND payload->>'journaldocno' = $3`, "", docNo)
	if err != nil {
		return err
	}
	return setDepreciationPosted(ctx, db, scope, items, "", now)
}

// reverseJournal reverses a posted depreciation journal through the general ledger engine: it
// flips debit/credit, keeps the original journal for audit, and re-projects PostgreSQL.
func (p *GLPoster) reverseJournal(ctx context.Context, scope Scope, journal *gl.Journal, reason string, now time.Time) error {
	docNo := journal.DocNo
	// doc_no is VARCHAR(30) characters: cut by runes so a Thai document number is never split mid-character.
	reversalDocNo := reversalDocNoPrefix + docNo
	if runes := []rune(reversalDocNo); len(runes) > gl.DocNoMaxRunes {
		reversalDocNo = string(runes[:gl.DocNoMaxRunes])
	}
	reverseDate, err := p.reversalDate(ctx, scope, journal.Date, now)
	if err != nil {
		return err
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
	return nil
}

// generalBookCode picks the company's active general journal book (booktype 1) by type —
// book codes are user-defined, so the poster never assumes "JV".
func (p *GLPoster) generalBookCode(ctx context.Context, scope Scope) (string, error) {
	page, err := p.ledger.List(ctx, glScope(scope), "journal-books", "", 1, 1000, gl.ListFilter{})
	if err != nil {
		return "", err
	}
	books := make([]gl.Master, 0, len(page.Items))
	for _, raw := range page.Items {
		var book gl.Master
		if err := json.Unmarshal(raw, &book); err != nil {
			return "", err
		}
		books = append(books, book)
	}
	return gl.ChooseBookCode(books, gl.BookTypeGeneral)
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
		disposal.DisposalDate = businessDate(now)
	}
	if err := checkVoucherDate(disposal.DisposalDate, "disposaldate"); err != nil {
		return nil, nil, err
	}
	glYear, err := p.fiscalYearCodeAt(ctx, scope, disposal.DisposalDate, "disposaldate")
	if err != nil {
		return nil, nil, err
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
	branch, err := p.disposalBranch(ctx, scope, asset)
	if err != nil {
		return nil, nil, err
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

	// 3. Build Disposal GL Journal. Every account comes from the asset, its asset type or the
	// disposal input — never a guessed code (charts of accounts differ per company).
	types, err := assetTypesByCode(ctx, db, scope.Company, []Asset{asset})
	if err != nil {
		return nil, nil, err
	}
	acc := resolveAssetAccounts(asset, types[asset.AssetTypeCode])
	settlementAcc := gl.NormalizeCode(disposal.SettlementAccountCode)
	gainLossAcc := gl.NormalizeCode(disposal.GainLossAccountCode)
	vatAcc := gl.NormalizeCode(disposal.VatAccountCode)
	totalReceipt := salePrice.Add(vatAmount)
	switch {
	case totalReceipt.GreaterThan(decimal.Zero) && settlementAcc == "":
		return nil, nil, faFieldError("fa_account_required", "settlementaccountcode", fmt.Sprintf("กรุณาระบุรหัสบัญชีรับชำระ (เงินสด/เงินฝาก/ลูกหนี้) สำหรับยอดรับ %s บาท", totalReceipt.StringFixed(2)))
	case accumDeprec.GreaterThan(decimal.Zero) && acc.accum == "":
		return nil, nil, assetAccountMissing("accumdeprecaccountcode", labelAccumAccount, []string{asset.AssetCode})
	case !gainLoss.IsZero() && gainLossAcc == "":
		return nil, nil, faFieldError("fa_account_required", "gainlossaccountcode", fmt.Sprintf("กรุณาระบุรหัสบัญชีกำไร/ขาดทุนจากการจำหน่ายสินทรัพย์ สำหรับยอด %s บาท", gainLoss.Abs().StringFixed(2)))
	case acc.cost == "":
		return nil, nil, assetAccountMissing("assetaccountcode", labelCostAccount, []string{asset.AssetCode})
	case vatAmount.GreaterThan(decimal.Zero) && vatAcc == "":
		return nil, nil, faFieldError("fa_account_required", "vataccountcode", fmt.Sprintf("กรุณาระบุรหัสบัญชีภาษีขาย สำหรับภาษีมูลค่าเพิ่ม %s บาท", vatAmount.StringFixed(2)))
	}
	var lines []GLLineDoc

	// Dr. Cash / AR (Sale Price + VAT)
	if totalReceipt.GreaterThan(decimal.Zero) {
		lines = append(lines, GLLineDoc{
			AccountCode: settlementAcc,
			Description: fmt.Sprintf("รับชำระเงินค่าจำหน่ายสินทรัพย์ %s", asset.AssetCode),
			Debit:       totalReceipt,
			Credit:      decimal.Zero,
			BranchCode:  branch,
		})
	}

	// Dr. Accumulated Depreciation
	if accumDeprec.GreaterThan(decimal.Zero) {
		lines = append(lines, GLLineDoc{
			AccountCode: acc.accum,
			Description: fmt.Sprintf("ตัดค่าเสื่อมราคาสะสม สินทรัพย์ %s", asset.AssetCode),
			Debit:       accumDeprec,
			Credit:      decimal.Zero,
			BranchCode:  branch,
		})
	}

	// Dr. Loss on Disposal (if GainLoss < 0)
	if gainLoss.LessThan(decimal.Zero) {
		lines = append(lines, GLLineDoc{
			AccountCode: gainLossAcc,
			Description: fmt.Sprintf("ขาดทุนจากการจำหน่ายสินทรัพย์ %s", asset.AssetCode),
			Debit:       gainLoss.Abs(),
			Credit:      decimal.Zero,
			BranchCode:  branch,
		})
	}

	// Cr. Asset Cost
	lines = append(lines, GLLineDoc{
		AccountCode: acc.cost,
		Description: fmt.Sprintf("ตัดราคาทุนสินทรัพย์ %s", asset.AssetCode),
		Debit:       decimal.Zero,
		Credit:      cost,
		BranchCode:  branch,
	})

	// Cr. Gain on Disposal (if GainLoss > 0)
	if gainLoss.GreaterThan(decimal.Zero) {
		lines = append(lines, GLLineDoc{
			AccountCode: gainLossAcc,
			Description: fmt.Sprintf("กำไรจากการจำหน่ายสินทรัพย์ %s", asset.AssetCode),
			Debit:       decimal.Zero,
			Credit:      gainLoss,
			BranchCode:  branch,
		})
	}

	// Cr. Output VAT (if any)
	if vatAmount.GreaterThan(decimal.Zero) {
		lines = append(lines, GLLineDoc{
			AccountCode: vatAcc,
			Description: fmt.Sprintf("ภาษีขายจากการจำหน่ายสินทรัพย์ %s", asset.AssetCode),
			Debit:       decimal.Zero,
			Credit:      vatAmount,
			BranchCode:  branch,
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
		return nil, nil, fmt.Errorf("รายการจำหน่ายสินทรัพย์ไม่สมดุล: เดบิต (%s) != เครดิต (%s)", totDr.String(), totCr.String())
	}

	book, err := p.generalBookCode(ctx, scope)
	if err != nil {
		return nil, nil, err
	}
	// The voucher number is the user's, else the chosen general book + asset code — which can
	// outgrow doc_no VARCHAR(30), so it is checked before anything is posted.
	journalDocNo := gl.NormalizeCode(disposal.JournalDocNo)
	generated := journalDocNo == ""
	if generated {
		journalDocNo = book + "-DISP-" + asset.AssetCode
	}
	if err := checkJournalDocNo(journalDocNo, "journaldocno", generated); err != nil {
		return nil, nil, err
	}
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
		BookCode:    book,
		FiscalYear:  glYear,
		Description: description,
		Reference:   disposal.DocNo,
		BranchCode:  branch,
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
		BookCode:     book,
		FiscalYear:   glYear,
		Description:  description,
		Reference:    disposal.DocNo,
		BranchCode:   branch,
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
