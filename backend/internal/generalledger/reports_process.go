package generalledger

import (
	"context"
	"database/sql"
	"fmt"
)

// ProcessBalance is an exact residual, retaining dimensions required when a
// closing or opening journal is generated. Balance is signed debit - credit;
// sums may exceed one line's Amount range, so the caller validates each output.
type ProcessBalance struct {
	AccountCode    string
	AccountName    string
	AccountType    string
	BranchCode     string
	DepartmentCode string
	ProjectCode    string
	Balance        string
}

type ProcessBalanceSnapshot struct {
	Sequence   int64
	FiscalYear string
	AsOf       string
	Scale      int
	Rows       []ProcessBalance
}

const ProcessBalanceRowLimit = 10000

// ProcessBalances reads its watermark and balances in one repeatable-read
// transaction. A partial result can never be used to close an accounting year.
func (p *Postgres) ProcessBalances(ctx context.Context, scope Scope, fiscalYear, to string) (ProcessBalanceSnapshot, error) {
	result := ProcessBalanceSnapshot{Rows: []ProcessBalance{}}
	if scope.Branch != "" {
		return result, fmt.Errorf("ยอดสำหรับปิดบัญชีต้องอ่านครบทุกสาขาในบริษัท")
	}
	if fiscalYear == "" || !validDate(to) {
		return result, fmt.Errorf("กรุณาระบุปีบัญชีและวันที่ประมวลผล")
	}
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return result, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	// This first SELECT fixes the snapshot used by all subsequent reads.
	err = tx.QueryRowContext(ctx, `SELECT sequence FROM gl_projection_state WHERE company=$1`, scope.Company).Scan(&result.Sequence)
	if err != nil && err != sql.ErrNoRows {
		return result, err
	}
	rc, err := newReportContext(ctx, tx, scope, ReportQuery{FiscalYear: fiscalYear, To: to})
	if err != nil {
		return result, err
	}
	result.FiscalYear = rc.fiscal.Code
	result.Scale = rc.fiscal.Scale
	result.AsOf = rc.query.To
	rows, err := tx.QueryContext(ctx, `SELECT account_code,MAX(account_name),MAX(account_type),branch_code,department_code,project_code,SUM(debit-credit)::text
      FROM gl_lines WHERE company=$1 AND fiscal_year=$2 AND entry_date<=$3::date
      GROUP BY account_code,branch_code,department_code,project_code HAVING SUM(debit-credit)<>0
      ORDER BY branch_code,department_code,project_code,account_code LIMIT $4`, scope.Company, fiscalYear, to, ProcessBalanceRowLimit+1)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var balance ProcessBalance
		if err = rows.Scan(&balance.AccountCode, &balance.AccountName, &balance.AccountType, &balance.BranchCode, &balance.DepartmentCode, &balance.ProjectCode, &balance.Balance); err != nil {
			rows.Close()
			return result, err
		}
		result.Rows = append(result.Rows, balance)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	if len(result.Rows) > ProcessBalanceRowLimit {
		return ProcessBalanceSnapshot{}, fmt.Errorf("ยอดแยกตามมิติมากกว่า %d รายการ กรุณาติดต่อผู้ดูแลก่อนประมวลผล", ProcessBalanceRowLimit)
	}
	if err = tx.Commit(); err != nil {
		return ProcessBalanceSnapshot{}, err
	}
	return result, nil
}

// HasDraftJournals checks whether unposted draft journals exist up to the processing date.
func (p *Postgres) HasDraftJournals(ctx context.Context, scope Scope, fiscalYear, to string) (bool, error) {
	if fiscalYear == "" || !validDate(to) {
		return false, fmt.Errorf("กรุณาระบุปีบัญชีและวันที่ประมวลผล")
	}
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return false, err
	}
	var count int
	query := `SELECT count(*) FROM gl_records
		WHERE company = $1 AND kind = 'journals'
		AND NOT COALESCE((payload->>'isdeleted')::boolean, false)
		AND payload->>'fiscalyear' = $2
		AND payload->>'status' = 'draft'
		AND (payload->>'date')::date <= $3::date`
	err = db.QueryRowContext(ctx, query, scope.Company, fiscalYear, to).Scan(&count)
	return count > 0, err
}

// HasOpeningJournal checks whether an active opening journal already exists for the fiscal year.
func (p *Postgres) HasOpeningJournal(ctx context.Context, scope Scope, fiscalYear string) (bool, error) {
	if fiscalYear == "" {
		return false, fmt.Errorf("กรุณาระบุปีบัญชี")
	}
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return false, err
	}
	var count int
	query := `SELECT count(*) FROM gl_records
		WHERE company = $1 AND kind = 'journals'
		AND NOT COALESCE((payload->>'isdeleted')::boolean, false)
		AND payload->>'fiscalyear' = $2
		AND payload->>'kind' = 'opening'`
	err = db.QueryRowContext(ctx, query, scope.Company, fiscalYear).Scan(&count)
	return count > 0, err
}
