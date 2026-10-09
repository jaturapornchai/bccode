package generalledger

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ExecuteSourceTx runs a journal command (create, post or reverse) for a module that posts on behalf of its own
// document — ERP sales, purchases, receipts — inside that module's transaction. The document and its voucher then
// commit or roll back together, and the voucher still goes through every rule of the ledger: open fiscal year,
// unlocked period, active journal book, postable accounts, debit = credit, idempotent request, audit event.
// The caller owns Begin/Commit and must have ensured the ledger schema (EnsureSchema) on the same database.
func ExecuteSourceTx(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command) (Result, error) {
	if scope.Holding == "" || scope.Company == "" || scope.Actor == "" {
		return Result{}, fmt.Errorf("กรุณาเลือกบริษัทและเข้าสู่ระบบก่อนใช้งานบัญชี")
	}
	normalizeCommandCodes(&cmd)
	if err := rejectNUL(cmd); err != nil {
		return Result{}, err
	}
	if !validRequestID(cmd.RequestID) {
		return Result{}, fmt.Errorf("รหัสคำขอไม่ถูกต้อง กรุณาลองบันทึกอีกครั้ง")
	}
	if cmd.Resource != "journals" || !contains([]string{"create", "post", "reverse"}, cmd.Action) {
		return Result{}, fmt.Errorf("ไม่รองรับคำสั่งบัญชีนี้")
	}
	if cmd.Action != "create" && cmd.ID == "" {
		return Result{}, fmt.Errorf("กรุณาเลือกรายการบัญชี")
	}
	return (&PostgresStore{}).executeInTx(ctx, tx, scope, cmd)
}

// LockCompanyTx takes the same company lock every ledger command takes. A module that touches its own rows
// before posting calls it first, so concurrent documents of one company queue in one order and cannot deadlock
// against a ledger command that holds the lock and waits for the module's rows.
func LockCompanyTx(ctx context.Context, tx *sql.Tx, company string) error {
	return lockCompany(ctx, tx, company)
}

// SourceFiscalYear returns the code of the open fiscal year that covers date, with a Thai user error when there
// is none or it is closed — the voucher of a source document never falls back to a guessed year.
func SourceFiscalYear(ctx context.Context, tx *sql.Tx, scope Scope, date string) (string, error) {
	if !validDate(date) {
		return "", fieldError("source_date_invalid", "date", "วันที่ของเอกสารไม่ถูกต้อง กรุณาเลือกวันที่ให้ครบ วัน เดือน ปี")
	}
	var code string
	err := tx.QueryRowContext(ctx, `SELECT code FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND payload->>'startdate'<=$2 AND payload->>'enddate'>=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, scope.Company, date).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		e := fieldError(codeSourceFiscalYearNotFound, "date", fmt.Sprintf("ไม่พบปีบัญชีที่ครอบคลุมวันที่ %s กรุณาเพิ่มปีบัญชีที่เมนู ปีบัญชีและบัญชีปิดปี หรือเลือกวันที่อื่น", date))
		e.Args = []string{date}
		return "", e
	}
	if err != nil {
		return "", err
	}
	year, err := (&PostgresStore{}).pgYear(ctx, tx, scope, code)
	if err != nil {
		return "", err
	}
	if year.Closed || !year.IsActive {
		e := fieldError(codeSourceFiscalYearClosed, "date", fmt.Sprintf("ปีบัญชี %s ของวันที่ %s ปิดแล้วหรือไม่ได้เปิดใช้งาน กรุณาเลือกวันที่ในปีบัญชีที่เปิดอยู่", code, date))
		e.Args = []string{code, date}
		return "", e
	}
	return code, nil
}

// SourceJournalInfo is what a module needs to know about the voucher of its document before reversing it.
type SourceJournalInfo struct {
	Status  string // draft | posted | reversed
	Date    string
	DocNo   string
	Version int64
}

// SourceJournal reads the live state of one voucher by id; found is false when it does not exist.
func SourceJournal(ctx context.Context, tx *sql.Tx, scope Scope, id string) (SourceJournalInfo, bool, error) {
	var info SourceJournalInfo
	err := tx.QueryRowContext(ctx, `SELECT version, COALESCE(payload->>'status',''), COALESCE(payload->>'date',''), COALESCE(payload->>'docno','') FROM gl_records WHERE company=$1 AND kind='journals' AND id=$2`, scope.Company, id).Scan(&info.Version, &info.Status, &info.Date, &info.DocNo)
	if errors.Is(err, sql.ErrNoRows) {
		return SourceJournalInfo{}, false, nil
	}
	return info, err == nil, err
}

// SourceAccountIsCash reports the chart's own "cash/bank" flag of an account — a module decides from this master
// attribute, never from the shape of the account code, whether a posting line is a receivable/payable or cash.
func SourceAccountIsCash(ctx context.Context, tx *sql.Tx, scope Scope, code string) (bool, error) {
	accounts, err := loadLineAccounts(ctx, tx, scope.Company, []Line{{AccountCode: code}})
	if err != nil {
		return false, err
	}
	return accounts[code].IsCash, nil
}

const (
	codeSourceFiscalYearNotFound = "source_fiscal_year_not_found"
	codeSourceFiscalYearClosed   = "source_fiscal_year_closed"
)

// SourceReversalDate dates the reversal of a posted source voucher on the voucher's own date while the ledger
// still takes vouchers there (the period nets to zero and a corrected document posts back into it). A date in a
// closed year or locked period cannot change, so the reversal goes to today (by the holding's calendar, passed in
// as today) and never before the original.
func SourceReversalDate(ctx context.Context, tx *sql.Tx, scope Scope, original, today string) (string, error) {
	open, err := sourceDateOpen(ctx, tx, scope, original)
	if err != nil {
		return "", err
	}
	if open {
		return original, nil
	}
	date := today
	if date < original {
		date = original
	}
	if _, err := SourceFiscalYear(ctx, tx, scope, date); err != nil {
		return "", err
	}
	return date, nil
}

func sourceDateOpen(ctx context.Context, tx *sql.Tx, scope Scope, date string) (bool, error) {
	if _, err := SourceFiscalYear(ctx, tx, scope, date); err != nil {
		if user, ok := AsUserError(err); ok && (user.Code == codeSourceFiscalYearNotFound || user.Code == codeSourceFiscalYearClosed) {
			return false, nil
		}
		return false, err
	}
	locked, err := periodLocked(ctx, tx, scope.Company, date)
	return !locked, err
}
