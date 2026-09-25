package generalledger

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// Budgets (กำหนดงบประมาณ) live in their own tables (budget.sql): a header per budget and
// one amount per account per fiscal period. They are not gl_records masters; commands still
// run through Execute so they share idempotency, the company lock and the gl_events audit.
// Champ baseline: menu 5500 กำหนดงบประมาณ (table BCGLBudget), reports 5525/5530.
// ADR docs/kms/decisions/2026-09-25-gl-monthly-budget.md.

//go:embed budget.sql
var budgetSchema string

const (
	BudgetPeriods       = 12
	budgetNameMaxRunes  = 200
	// Champ BCGLBudget.Status: a plain open/closed flag (no approval workflow, no lock).
	budgetStatusOpen   = "open"
	budgetStatusClosed = "closed"
)

// budgetAmountLimit is the first value numeric(18,2) cannot store.
var budgetAmountLimit = decimal.New(1, 16)

type BudgetLine struct {
	AccountCode string   `json:"accountcode"`
	AccountName string   `json:"accountname,omitempty"`
	Periods     []Amount `json:"periods"`
	// Total is output on read; on action "spread" it is the annual amount to split.
	Total Amount `json:"total,omitempty"`
}

// Budget is the command payload (Command.Budget) and, with Identity, the read shape.
type Budget struct {
	Code           string       `json:"code"`
	Name           string       `json:"name"`
	FiscalYear     string       `json:"fiscalyear"`
	BranchCode     string       `json:"branchcode"`
	DepartmentCode string       `json:"departmentcode"`
	ProjectCode    string       `json:"projectcode"`
	Status         string       `json:"status,omitempty"`
	Remark         string       `json:"remark"`
	Lines          []BudgetLine `json:"lines,omitempty"`
}

type budgetRecord struct {
	Identity
	Budget
	Total Amount `json:"total"`
}

var (
	errBudgetPayloadMissing    = fieldError("budget_payload_required", "budget", "ไม่พบข้อมูลงบประมาณ กรุณากรอกข้อมูลแล้วบันทึกอีกครั้ง")
	errBudgetNameRequired      = fieldError("budget_name_required", "name", "กรุณาระบุชื่องบประมาณ")
	errBudgetNameTooLong       = fieldError("budget_name_too_long", "name", fmt.Sprintf("ชื่องบประมาณยาวได้ไม่เกิน %d ตัวอักษร", budgetNameMaxRunes))
	errBudgetFiscalYear        = fieldError("budget_fiscal_year_not_found", "fiscalyear", "ไม่พบปีบัญชีที่เลือก กรุณาเลือกปีบัญชีจากรายการ")
	errBudgetBranchOutside     = fieldError("budget_branch_outside_session", "branchcode", "คุณเข้าระบบระดับสาขา กำหนดงบประมาณได้เฉพาะสาขาของคุณ")
	errBudgetLinesRequired     = fieldError("budget_lines_required", "lines", "กรุณาเพิ่มบัญชีในงบประมาณอย่างน้อย 1 บัญชี")
	errBudgetAccountRequired   = fieldError("budget_account_required", "lines", "กรุณาเลือกรหัสบัญชีในทุกบรรทัดของงบประมาณ")
	errBudgetPeriodsInvalid    = fieldError("budget_periods_invalid", "lines", "งบประมาณแต่ละบัญชีต้องมียอด 12 งวด")
	errBudgetAmountNegative    = fieldError("budget_amount_negative", "lines", "จำนวนเงินงบประมาณต้องไม่ติดลบ")
	errBudgetAmountScale       = fieldError("budget_amount_scale", "lines", "จำนวนเงินงบประมาณมีทศนิยมได้ไม่เกิน 2 ตำแหน่ง")
	errBudgetAmountTooLarge    = fieldError("budget_amount_too_large", "lines", "จำนวนเงินงบประมาณสูงเกินกว่าที่ระบบรองรับ")
	errBudgetPeriodOutsideYear = fieldError("budget_period_outside_year", "lines", "ปีบัญชีนี้มีงวดไม่ครบ 12 งวด กรุณาใส่ยอดเฉพาะงวดที่อยู่ในปีบัญชี")
	errBudgetCodeImmutable     = fieldError("budget_code_immutable", "code", "เปลี่ยนรหัสงบประมาณที่บันทึกแล้วไม่ได้")
	errBudgetStatusInvalid     = fieldError("budget_status_invalid", "status", "สถานะงบประมาณต้องเป็น เปิดใช้งาน หรือ ปิด")
)

func budgetAccountDuplicate(code string) error {
	return fieldError("budget_account_duplicate", "lines", "บัญชี "+code+" อยู่ในงบประมาณนี้แล้ว กรุณารวมยอดไว้บรรทัดเดียว")
}
func budgetAccountNotFound(code string) error {
	return fieldError("budget_account_not_found", "lines", "ไม่พบรหัสบัญชี "+code+" ในผังบัญชี")
}
func budgetAccountNotPosting(code string) error {
	return fieldError("budget_account_not_posting", "lines", "บัญชี "+code+" เป็นบัญชีคุมหรือปิดใช้งาน กรุณาเลือกบัญชีย่อยที่ลงรายการได้")
}

// SpreadAnnual splits an annual amount into 12 monthly amounts: each month is the annual
// divided by 12 truncated to 2 decimals, and the last month takes the remainder so the
// months always add back to the annual exactly (100000 → 8333.33 × 11 + 8333.37).
func SpreadAnnual(annual decimal.Decimal) []decimal.Decimal {
	month := annual.Div(decimal.NewFromInt(BudgetPeriods)).Truncate(2)
	out := make([]decimal.Decimal, BudgetPeriods)
	for i := 0; i < BudgetPeriods-1; i++ {
		out[i] = month
	}
	out[BudgetPeriods-1] = annual.Sub(month.Mul(decimal.NewFromInt(BudgetPeriods - 1)))
	return out
}

func checkBudgetAmount(a Amount) error {
	d := a.Decimal()
	if d.IsNegative() {
		return errBudgetAmountNegative
	}
	if !d.Equal(d.Truncate(2)) {
		return errBudgetAmountScale
	}
	if d.GreaterThanOrEqual(budgetAmountLimit) {
		return errBudgetAmountTooLarge
	}
	return nil
}

// spreadBudget answers action "spread" without touching the database.
func spreadBudget(cmd Command) (Result, error) {
	if cmd.Budget == nil || len(cmd.Budget.Lines) == 0 {
		return Result{}, errBudgetLinesRequired
	}
	lines := make([]BudgetLine, len(cmd.Budget.Lines))
	for i, line := range cmd.Budget.Lines {
		if err := checkBudgetAmount(line.Total); err != nil {
			return Result{}, err
		}
		periods := make([]Amount, BudgetPeriods)
		for p, d := range SpreadAnnual(line.Total.Decimal()) {
			periods[p] = Amount(d.StringFixed(2))
		}
		lines[i] = BudgetLine{AccountCode: line.AccountCode, Periods: periods, Total: Amount(line.Total.Decimal().StringFixed(2))}
	}
	return Result{Lines: lines}, nil
}

func normalizeBudget(b *Budget) {
	for _, code := range []*string{&b.Code, &b.FiscalYear, &b.BranchCode, &b.DepartmentCode, &b.ProjectCode} {
		*code = NormalizeCode(*code)
	}
	b.Name = strings.TrimSpace(b.Name)
	b.Status = strings.ToLower(strings.TrimSpace(b.Status))
	if b.Status == "" {
		b.Status = budgetStatusOpen
	}
	b.Remark = strings.TrimSpace(b.Remark)
	b.Lines = append([]BudgetLine(nil), b.Lines...)
	for i := range b.Lines {
		b.Lines[i].AccountCode = NormalizeCode(b.Lines[i].AccountCode)
	}
}

// fiscalPeriodStarts returns the start date of every period of the fiscal year: period 1
// starts on the fiscal start date, later periods on the 1st of each following month, and
// periods that would start after the fiscal end date do not exist.
func fiscalPeriodStarts(y FiscalYear) []string {
	start, err1 := time.Parse("2006-01-02", y.StartDate)
	end, err2 := time.Parse("2006-01-02", y.EndDate)
	if err1 != nil || err2 != nil {
		return nil
	}
	first := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	out := []string{y.StartDate}
	for n := 1; n < BudgetPeriods; n++ {
		p := first.AddDate(0, n, 0)
		if p.After(end) {
			break
		}
		out = append(out, p.Format("2006-01-02"))
	}
	return out
}

func (s *PostgresStore) validateBudget(ctx context.Context, tx *sql.Tx, scope Scope, b Budget) error {
	if err := checkCode(b.Code, "code", "รหัสงบประมาณ", codeMaxRunes); err != nil {
		return err
	}
	if b.Name == "" {
		return errBudgetNameRequired
	}
	if utf8.RuneCountInString(b.Name) > budgetNameMaxRunes {
		return errBudgetNameTooLong
	}
	if b.Status != budgetStatusOpen && b.Status != budgetStatusClosed {
		return errBudgetStatusInvalid
	}
	for _, dim := range []struct{ value, field, label string }{{b.BranchCode, "branchcode", "รหัสสาขา"}, {b.DepartmentCode, "departmentcode", "รหัสแผนก"}, {b.ProjectCode, "projectcode", "รหัสโครงการ"}} {
		if dim.value != "" {
			if err := checkCode(dim.value, dim.field, dim.label, codeMaxRunes); err != nil {
				return err
			}
		}
	}
	if scope.Branch != "" && b.BranchCode != scope.Branch {
		return errBudgetBranchOutside
	}
	year, err := s.pgYear(ctx, tx, scope, b.FiscalYear)
	if err != nil || b.FiscalYear == "" {
		return errBudgetFiscalYear
	}
	periodCount := len(fiscalPeriodStarts(year))
	if len(b.Lines) == 0 {
		return errBudgetLinesRequired
	}
	seen := map[string]bool{}
	lookup := make([]Line, 0, len(b.Lines))
	for _, line := range b.Lines {
		if line.AccountCode == "" {
			return errBudgetAccountRequired
		}
		if seen[line.AccountCode] {
			return budgetAccountDuplicate(line.AccountCode)
		}
		seen[line.AccountCode] = true
		if len(line.Periods) != BudgetPeriods {
			return errBudgetPeriodsInvalid
		}
		for p, amount := range line.Periods {
			if err := checkBudgetAmount(amount); err != nil {
				return err
			}
			if p >= periodCount && !amount.Decimal().IsZero() {
				return errBudgetPeriodOutsideYear
			}
		}
		lookup = append(lookup, Line{AccountCode: line.AccountCode})
	}
	accounts, err := loadLineAccounts(ctx, tx, scope.Company, lookup)
	if err != nil {
		return err
	}
	for _, line := range b.Lines {
		a, ok := accounts[line.AccountCode]
		if !ok {
			return budgetAccountNotFound(line.AccountCode)
		}
		if !a.AllowPosting || !a.IsActive {
			return budgetAccountNotPosting(line.AccountCode)
		}
	}
	return nil
}

func (s *PostgresStore) mutateBudget(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if cmd.Action == "create" {
		if cmd.Budget == nil {
			return nil, errBudgetPayloadMissing
		}
		b := *cmd.Budget
		if err := s.validateBudget(ctx, tx, scope, b); err != nil {
			return nil, err
		}
		result, err := tx.ExecContext(ctx, `INSERT INTO gl_budgets(company,code,name,fiscal_year,branch_code,department_code,project_code,status,remark,version,created_by,created_at,updated_by,updated_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,1,$10,$11,$10,$11) ON CONFLICT DO NOTHING`, scope.Company, b.Code, b.Name, b.FiscalYear, b.BranchCode, b.DepartmentCode, b.ProjectCode, b.Status, b.Remark, scope.Actor, now)
		if err != nil {
			return nil, err
		}
		if n, err := result.RowsAffected(); err != nil {
			return nil, err
		} else if n == 0 {
			return nil, userError(CodeDuplicateCode, "รหัสงบประมาณนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น")
		}
		if err = insertBudgetLines(ctx, tx, scope.Company, b); err != nil {
			return nil, err
		}
		return budgetChange(ctx, tx, scope, b.Code, false)
	}

	old, err := loadBudgetHeader(ctx, tx, scope, cmd.ID, true)
	if err != nil {
		return nil, err
	}
	if err := checkVersion(cmd, old.Identity); err != nil {
		return nil, err
	}
	switch cmd.Action {
	case "update":
		if cmd.Budget == nil {
			return nil, errBudgetPayloadMissing
		}
		b := *cmd.Budget
		if b.Code != "" && b.Code != old.Code {
			return nil, errBudgetCodeImmutable
		}
		b.Code = old.Code
		if err := s.validateBudget(ctx, tx, scope, b); err != nil {
			return nil, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE gl_budgets SET name=$3,fiscal_year=$4,branch_code=$5,department_code=$6,project_code=$7,status=$8,remark=$9,version=version+1,updated_by=$10,updated_at=$11 WHERE company=$1 AND code=$2`,
			scope.Company, b.Code, b.Name, b.FiscalYear, b.BranchCode, b.DepartmentCode, b.ProjectCode, b.Status, b.Remark, scope.Actor, now); err != nil {
			return nil, err
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM gl_budget_lines WHERE company=$1 AND budget_code=$2`, scope.Company, b.Code); err != nil {
			return nil, err
		}
		if err = insertBudgetLines(ctx, tx, scope.Company, b); err != nil {
			return nil, err
		}
		return budgetChange(ctx, tx, scope, b.Code, false)
	case "delete":
		changes, err := budgetChange(ctx, tx, scope, old.Code, true)
		if err != nil {
			return nil, err
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM gl_budgets WHERE company=$1 AND code=$2`, scope.Company, old.Code); err != nil {
			return nil, err
		}
		return changes, nil
	}
	return nil, userError(CodeUnsupported, "ไม่รองรับคำสั่งนี้สำหรับงบประมาณ")
}

func insertBudgetLines(ctx context.Context, tx *sql.Tx, company string, b Budget) error {
	accounts, periods, amounts := []string{}, []int64{}, []string{}
	for _, line := range b.Lines {
		for p, amount := range line.Periods {
			if amount.Decimal().IsZero() {
				continue
			}
			accounts = append(accounts, line.AccountCode)
			periods = append(periods, int64(p+1))
			amounts = append(amounts, amount.Decimal().StringFixed(2))
		}
	}
	// Accounts with no amount in any period still belong to the budget (a zero budget).
	for _, line := range b.Lines {
		zero := true
		for _, amount := range line.Periods {
			if !amount.Decimal().IsZero() {
				zero = false
			}
		}
		if zero {
			accounts, periods, amounts = append(accounts, line.AccountCode), append(periods, 1), append(amounts, "0.00")
		}
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO gl_budget_lines(company,budget_code,account_code,period_no,amount)
		SELECT $1,$2,a,p,m::numeric FROM unnest($3::text[],$4::smallint[],$5::text[]) AS t(a,p,m)`, company, b.Code, pq.Array(accounts), pq.Array(periods), pq.Array(amounts))
	return err
}

// budgetChange is the audit snapshot written to gl_events (applyChanges does not project it).
func budgetChange(ctx context.Context, tx *sql.Tx, scope Scope, code string, deleted bool) ([]Change, error) {
	record, err := readBudget(ctx, tx, scope, code)
	if err != nil {
		return nil, err
	}
	if deleted {
		record.IsDeleted = true
		record.Version++
	}
	return pgChange("budgets", record.ID, record.Code, record)
}

type budgetQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

const budgetHeaderColumns = `code,name,fiscal_year,branch_code,department_code,project_code,status,remark,version,created_by,created_at,updated_by,updated_at`

func scanBudgetHeader(row interface{ Scan(...any) error }, scope Scope, extra ...any) (budgetRecord, error) {
	var r budgetRecord
	dest := []any{&r.Code, &r.Name, &r.FiscalYear, &r.BranchCode, &r.DepartmentCode, &r.ProjectCode, &r.Status, &r.Remark, &r.Version, &r.CreatedBy, &r.CreatedAt, &r.UpdatedBy, &r.UpdatedAt}
	err := row.Scan(append(dest, extra...)...)
	r.ID, r.HoldingCode, r.BusinessCode = r.Code, scope.Holding, scope.Company
	r.CreatedAt, r.UpdatedAt = r.CreatedAt.UTC(), r.UpdatedAt.UTC()
	return r, err
}

func loadBudgetHeader(ctx context.Context, q budgetQuerier, scope Scope, code string, forUpdate bool) (budgetRecord, error) {
	lock := ""
	if forUpdate {
		lock = " FOR UPDATE"
	}
	r, err := scanBudgetHeader(q.QueryRowContext(ctx, `SELECT `+budgetHeaderColumns+` FROM gl_budgets WHERE company=$1 AND code=$2 AND ($3='' OR branch_code=$3)`+lock, scope.Company, code, scope.Branch), scope)
	if errors.Is(err, sql.ErrNoRows) {
		return r, ErrNotFound
	}
	return r, err
}

// readBudget returns the header with every line as 12 period amounts and the totals.
func readBudget(ctx context.Context, q budgetQuerier, scope Scope, code string) (budgetRecord, error) {
	r, err := loadBudgetHeader(ctx, q, scope, code, false)
	if err != nil {
		return r, err
	}
	rows, err := q.QueryContext(ctx, `SELECT l.account_code,
			COALESCE((SELECT n->>'name' FROM gl_records acc, jsonb_array_elements(COALESCE(acc.payload->'names','[]'::jsonb)) n WHERE acc.company=l.company AND acc.kind='accounts' AND acc.code=l.account_code AND NOT COALESCE((acc.payload->>'isdeleted')::boolean,false) AND n->>'code'='th' LIMIT 1),''),
			l.period_no,l.amount::text
		FROM gl_budget_lines l WHERE l.company=$1 AND l.budget_code=$2 ORDER BY l.account_code,l.period_no`, scope.Company, code)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	total := decimal.Zero
	r.Lines = []BudgetLine{}
	for rows.Next() {
		var account, name, amount string
		var period int
		if err = rows.Scan(&account, &name, &period, &amount); err != nil {
			return r, err
		}
		if len(r.Lines) == 0 || r.Lines[len(r.Lines)-1].AccountCode != account {
			periods := make([]Amount, BudgetPeriods)
			for i := range periods {
				periods[i] = "0.00"
			}
			r.Lines = append(r.Lines, BudgetLine{AccountCode: account, AccountName: name, Periods: periods, Total: "0.00"})
		}
		line := &r.Lines[len(r.Lines)-1]
		line.Periods[period-1] = Amount(amount)
		sum := line.Total.Decimal().Add(Amount(amount).Decimal())
		line.Total = Amount(sum.StringFixed(2))
		total = total.Add(Amount(amount).Decimal())
	}
	r.Total = Amount(total.StringFixed(2))
	return r, rows.Err()
}

func (p *Postgres) listBudgets(ctx context.Context, scope Scope, search string, page, limit int) (Page, error) {
	page, limit = pageBounds(page, limit)
	result := Page{Items: []json.RawMessage{}, Page: page, Limit: limit}
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return result, err
	}
	where := `b.company=$1 AND ($2='' OR b.branch_code=$2) AND ($3='' OR strpos(lower(b.code||' '||b.name||' '||b.remark),lower($3))>0)`
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM gl_budgets b WHERE `+where, scope.Company, scope.Branch, search).Scan(&result.Total); err != nil {
		return result, err
	}
	rows, err := db.QueryContext(ctx, `SELECT `+prefixColumns("b.", budgetHeaderColumns)+`,
			COALESCE((SELECT SUM(l.amount) FROM gl_budget_lines l WHERE l.company=b.company AND l.budget_code=b.code),0)::numeric(20,2)::text
		FROM gl_budgets b WHERE `+where+` ORDER BY b.fiscal_year DESC,b.code LIMIT $4 OFFSET $5`, scope.Company, scope.Branch, search, limit, (page-1)*limit)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var total string
		r, err := scanBudgetHeader(rows, scope, &total)
		if err != nil {
			return result, err
		}
		r.Total = Amount(total)
		data, err := json.Marshal(r)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, data)
	}
	return result, rows.Err()
}

func (p *Postgres) getBudget(ctx context.Context, scope Scope, code string) (json.RawMessage, error) {
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	r, err := readBudget(ctx, db, scope, NormalizeCode(code))
	if err != nil {
		return nil, err
	}
	return json.Marshal(r)
}

func prefixColumns(prefix, columns string) string {
	parts := strings.Split(columns, ",")
	for i := range parts {
		parts[i] = prefix + parts[i]
	}
	return strings.Join(parts, ",")
}
