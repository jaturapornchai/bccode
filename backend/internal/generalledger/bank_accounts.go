package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"smlcloudplatform/internal/cheque"
	"smlcloudplatform/internal/goapi/language"
)

// สมุดเงินฝากธนาคาร (จอ Champ 5169 "บันทึกสมุดเงินฝากธนาคาร") — resource "bank-accounts" ของ POST /gl/v2/command.
// ตารางเดิม gl_subledger_bank_accounts (payload jsonb) ยังเป็นที่เก็บเดียว ใบสำคัญอ้างสมุดด้วยรหัสสมุด (bank_account_code)
// จึงไม่เปลี่ยนแม้เลขที่บัญชีธนาคารเปลี่ยน; รหัสธนาคาร/สาขา/ประเภทบัญชีตรวจกับทะเบียน bank_codes (internal/cheque)
// docs/kms/champ/design-2026-10-11/cheque-bank.md §3 "ส่วนขยายสมุดเงินฝาก" / §4.7.
const BankAccountKind = "bank-accounts"

// ChangeAccountNumberAction เปลี่ยนเฉพาะเลขที่บัญชีธนาคารของสมุดเดิม — ปุ่ม "เปลี่ยนเลขที่บัญชี" ของ Champ (CASHDlgChangeBookNo.cpp:59-130)
// แต่ไม่ต้องไล่แก้ทุกตารางเพราะเอกสารอ้างรหัสสมุดที่ไม่เปลี่ยน
const ChangeAccountNumberAction = "change-account-number"

// ความยาวตามร่างตารางเป้าหมาย mydocs/datamodels/gl/bank.sql (gl_bank_accounts); รหัสสมุด = AccountCodeMaxRunes (20)
const (
	bankBookNumberMaxRunes = 50
	bankBookNameMaxRunes   = 255
	bankBookAuditPrefix    = "bank-account:"
)

// bankBookError สร้างข้อผิดพลาดของสมุดเงินฝาก: ข้อความไทยมาจากแถว gl_err_<code> ใน languages.tsv (ภาษาอื่นแปลที่ httpapi ด้วยแถวเดียวกัน)
func bankBookError(status int, code, field string, args ...string) *UserError {
	message := language.Text("gl_err_"+code, "th")
	for i, arg := range args {
		message = strings.ReplaceAll(message, "{"+strconv.Itoa(i)+"}", arg)
	}
	return &UserError{Code: code, Field: field, Message: message, Status: status, Args: args}
}

func bankBookInvalid(code, field string, args ...string) *UserError {
	return bankBookError(http.StatusBadRequest, code, field, args...)
}

func bankBookConflict(code, field string, args ...string) *UserError {
	return bankBookError(http.StatusConflict, code, field, args...)
}

// bankBookErrorCodes คือรหัสข้อผิดพลาดทั้งหมดของสมุดเงินฝาก — ทุกตัวต้องมีแถว gl_err_<code> ครบ (bank_accounts_test.go ตรวจ)
var bankBookErrorCodes = []string{
	"bank_book_payload_required", "bank_book_name_required", "bank_book_name_too_long", "bank_book_number_required", "bank_book_number_too_long",
	"bank_book_bank_required", "bank_book_currency_not_supported", "bank_book_gl_account_invalid", "bank_book_ref_not_found", "bank_book_ref_inactive",
	"bank_book_branch_bank_mismatch", "bank_book_duplicate", "bank_book_account_number_duplicate", "bank_book_number_use_change",
	"bank_book_number_unchanged", "bank_book_in_use", "bank_book_balance_not_zero",
}

type bankAccountPlan struct {
	action    string
	code      string
	version   int64
	book      SubledgerBankAccount
	newNumber string
}

// planBankAccount ตรวจรูปแบบคำสั่งและค่าที่ตรวจได้โดยไม่ต้องแตะฐานข้อมูล
func planBankAccount(cmd Command) (bankAccountPlan, error) {
	plan := bankAccountPlan{action: cmd.Action, code: NormalizeCode(cmd.ID), version: cmd.Version}
	switch cmd.Action {
	case "create", "update":
		if cmd.BankAccount == nil {
			return plan, bankBookInvalid("bank_book_payload_required", "")
		}
		plan.book = *cmd.BankAccount
		if cmd.Action == "create" {
			plan.code = plan.book.Code
		} else if plan.code == "" || plan.book.Code != plan.code {
			return plan, bankBookInvalid("bank_book_payload_required", "bank_account_code")
		}
		return plan, validateBankBookInput(&plan.book)
	case "delete", ChangeAccountNumberAction:
		if plan.code == "" {
			return plan, bankBookInvalid("bank_book_payload_required", "id")
		}
		if cmd.Action == ChangeAccountNumberAction {
			plan.newNumber = strings.TrimSpace(cmd.NewAccountNumber)
			return plan, validateBankBookNumber(plan.newNumber, "new_account_number")
		}
		return plan, nil
	}
	return plan, userError(CodeUnsupported, language.Text("gl_err_unsupported_command", "th"))
}

func validateBankBookNumber(number, field string) error {
	if number == "" {
		return bankBookInvalid("bank_book_number_required", field)
	}
	if utf8.RuneCountInString(number) > bankBookNumberMaxRunes {
		return bankBookInvalid("bank_book_number_too_long", field, strconv.Itoa(bankBookNumberMaxRunes))
	}
	return nil
}

func validateBankBookInput(b *SubledgerBankAccount) error {
	b.AccountNumber = strings.TrimSpace(b.AccountNumber)
	b.AccountName = strings.TrimSpace(b.AccountName)
	b.Currency = strings.ToUpper(strings.TrimSpace(b.Currency))
	// ชื่อฟิลด์ในข้อความมาจากแถวที่มีอยู่แล้ว (bank_book_code = รหัสสมุดบัญชี, account_code = รหัสบัญชี); เหตุผล/วิธีแก้ใช้แถว gl_err_code_* เดิม
	if err := checkCode(b.Code, "bank_account_code", language.Text("bank_book_code", "th"), AccountCodeMaxRunes); err != nil {
		return err
	}
	if err := checkCode(b.GLAccountCode, "gl_account_code", language.Text("account_code", "th"), AccountCodeMaxRunes); err != nil {
		return err
	}
	switch {
	case b.AccountName == "":
		return bankBookInvalid("bank_book_name_required", "account_name")
	case utf8.RuneCountInString(b.AccountName) > bankBookNameMaxRunes:
		return bankBookInvalid("bank_book_name_too_long", "account_name", strconv.Itoa(bankBookNameMaxRunes))
	case b.BankCode == "":
		return bankBookInvalid("bank_book_bank_required", "bank_code")
	// จนกว่า D-FX จะตัดสิน สมุดเงินฝากรับเฉพาะบาท (ตรงกับ CHECK ของ bank.sql)
	case b.Currency != "" && b.Currency != "THB":
		return bankBookInvalid("bank_book_currency_not_supported", "currency_code")
	}
	b.Currency = "THB"
	return validateBankBookNumber(b.AccountNumber, "account_number")
}

// executeBankAccount รันคำสั่งสมุดเงินฝากใน transaction เดียว: ล็อกบริษัท → ตรวจซ้ำคำขอ → ตรวจกับข้อมูลจริง → เขียน → บันทึกประวัติ
func (s *PostgresStore) executeBankAccount(ctx context.Context, scope Scope, cmd Command) (Result, error) {
	plan, err := planBankAccount(cmd)
	if err != nil {
		return Result{}, err
	}
	raw, err := json.Marshal(cmd)
	if err != nil {
		return Result{}, err
	}
	db, err := s.pg.database(ctx, scope.Holding)
	if err != nil {
		return Result{}, err
	}
	// bank_codes เป็นของ internal/cheque — สร้างก่อนเปิด tx เพื่อให้บริษัทที่ยังไม่เคยเปิดจอกำหนดรหัสก็บันทึกสมุดได้
	if err = cheque.EnsureSchema(ctx, db); err != nil {
		return Result{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()
	if err = lockCompany(ctx, tx, scope.Company); err != nil {
		return Result{}, err
	}
	hash := digest(raw)
	if result, found, replayErr := replayBankAccount(ctx, tx, scope.Company, plan.code, cmd.RequestID, hash); replayErr != nil || found {
		return result, replayErr
	}
	result, err := applyBankAccount(ctx, tx, scope, cmd.RequestID, hash, plan, time.Now().UTC())
	if err != nil {
		return Result{}, err
	}
	if err = tx.Commit(); err != nil {
		return Result{}, err
	}
	return result, nil
}

// replayBankAccount คืนผลเดิมเมื่อ requestid นี้เคยสำเร็จแล้ว (ส่งซ้ำจากการกดสองครั้ง/เน็ตหลุด); ข้อมูลต่างจากเดิม = ปฏิเสธ
func replayBankAccount(ctx context.Context, tx *sql.Tx, company, code, requestID, hash string) (Result, bool, error) {
	var raw []byte
	err := tx.QueryRowContext(ctx, `SELECT payload FROM gl_subledger_audit WHERE company=$1 AND journal_id=$2 AND payload->>'request_id'=$3 ORDER BY event_no LIMIT 1`,
		company, bankBookAuditPrefix+code, requestID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{}, false, nil
	}
	if err != nil {
		return Result{}, false, err
	}
	var saved struct {
		RequestHash string `json:"request_hash"`
		Version     int64  `json:"version"`
	}
	if err = json.Unmarshal(raw, &saved); err != nil {
		return Result{}, false, err
	}
	if saved.RequestHash != hash {
		return Result{}, false, conflictError(CodeDuplicateRequest, "", language.Text("gl_err_duplicate_request", "th"))
	}
	return Result{ID: code, Version: saved.Version}, true, nil
}

func applyBankAccount(ctx context.Context, tx *sql.Tx, scope Scope, requestID, hash string, plan bankAccountPlan, now time.Time) (Result, error) {
	var old *SubledgerBankAccount
	if plan.action != "create" {
		found, err := loadBankBook(ctx, tx, scope.Company, plan.code)
		if err != nil {
			return Result{}, err
		}
		if plan.version < 1 || found.Version != plan.version {
			return Result{}, userError(CodeStaleVersion, language.Text("gl_err_stale_version", "th"))
		}
		old = &found
	}
	var next SubledgerBankAccount
	var err error
	switch plan.action {
	case "create":
		next, err = createBankBook(ctx, tx, scope.Company, plan.book)
	case "update":
		next, err = updateBankBook(ctx, tx, scope.Company, *old, plan.book)
	case ChangeAccountNumberAction:
		next, err = changeBankBookNumber(ctx, tx, scope.Company, *old, plan.newNumber)
	case "delete":
		err = checkBankBookDeletable(ctx, tx, scope.Company, *old)
		next = *old
	}
	if err != nil {
		return Result{}, err
	}
	if old != nil && plan.action == "update" && jsonEqual(*old, next) {
		return Result{ID: plan.code, Version: old.Version}, nil // ไม่มีอะไรเปลี่ยน: ไม่เพิ่ม version ไม่ลงประวัติ
	}
	if plan.action == "delete" {
		_, err = tx.ExecContext(ctx, `DELETE FROM gl_subledger_bank_accounts WHERE company=$1 AND code=$2`, scope.Company, plan.code)
	} else {
		err = saveBankBook(ctx, tx, scope.Company, next)
	}
	if err != nil {
		return Result{}, err
	}
	if err = auditBankBook(ctx, tx, scope, plan, requestID, hash, old, next, now); err != nil {
		return Result{}, err
	}
	return Result{ID: plan.code, Version: next.Version}, nil
}

func createBankBook(ctx context.Context, tx *sql.Tx, company string, book SubledgerBankAccount) (SubledgerBankAccount, error) {
	var existing string
	err := tx.QueryRowContext(ctx, `SELECT code FROM gl_subledger_bank_accounts WHERE company=$1 AND upper(code)=upper($2)`, company, book.Code).Scan(&existing)
	if err == nil {
		return book, bankBookConflict("bank_book_duplicate", "bank_account_code", existing)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return book, err
	}
	if err = checkBankBookReferences(ctx, tx, company, &book, nil); err != nil {
		return book, err
	}
	if err = checkBankBookGLAccount(ctx, tx, company, book.GLAccountCode); err != nil {
		return book, err
	}
	if err = checkBankBookNumberFree(ctx, tx, company, book.Code, book.AccountNumber, "account_number"); err != nil {
		return book, err
	}
	book.Version = 1
	return book, nil
}

// updateBankBook ตาม Champ CASHFrmBankBookNew.cpp: แก้บัญชี GL ได้เสมอแม้สมุดถูกอ้างแล้ว (:164-184,238,247);
// เลขที่บัญชีแก้ตรงได้เฉพาะสมุดที่ยังไม่ถูกอ้าง (:276-281 ปิดช่องเมื่อ CheckBookNoReference) — ถูกอ้างแล้วต้องใช้ปุ่มเปลี่ยนเลขที่บัญชี
func updateBankBook(ctx context.Context, tx *sql.Tx, company string, old, book SubledgerBankAccount) (SubledgerBankAccount, error) {
	if book.AccountNumber != old.AccountNumber {
		used, err := bankBookUsed(ctx, tx, company, old.Code)
		if err != nil {
			return book, err
		}
		if used {
			return book, bankBookInvalid("bank_book_number_use_change", "account_number")
		}
		if err := checkBankBookNumberFree(ctx, tx, company, old.Code, book.AccountNumber, "account_number"); err != nil {
			return book, err
		}
	}
	if err := checkBankBookReferences(ctx, tx, company, &book, &old); err != nil {
		return book, err
	}
	if book.GLAccountCode != old.GLAccountCode {
		if err := checkBankBookGLAccount(ctx, tx, company, book.GLAccountCode); err != nil {
			return book, err
		}
	}
	book.Version = old.Version
	if jsonEqual(old, book) {
		return book, nil
	}
	book.Version = old.Version + 1
	return book, nil
}

func changeBankBookNumber(ctx context.Context, tx *sql.Tx, company string, old SubledgerBankAccount, number string) (SubledgerBankAccount, error) {
	if number == old.AccountNumber {
		return old, bankBookInvalid("bank_book_number_unchanged", "new_account_number")
	}
	if err := checkBankBookNumberFree(ctx, tx, company, old.Code, number, "new_account_number"); err != nil {
		return old, err
	}
	next := old
	next.AccountNumber = number
	next.Version = old.Version + 1
	return next, nil
}

// checkBankBookDeletable ตาม Champ MSG_891: ลบไม่ได้เมื่อมียอดคงเหลือ; BC เพิ่มกันสมุดที่มีรายการเดินบัญชี/ใบรายการธนาคารอ้างอยู่ (FK ก็กันอยู่แล้ว)
func checkBankBookDeletable(ctx context.Context, tx *sql.Tx, company string, book SubledgerBankAccount) error {
	used, err := bankBookUsed(ctx, tx, company, book.Code)
	if err != nil {
		return err
	}
	if used {
		return bankBookConflict("bank_book_in_use", "id", book.Code)
	}
	var nonZero bool
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(debit-credit),0)<>0 FROM gl_lines WHERE company=$1 AND account_code=$2`, company, book.GLAccountCode).Scan(&nonZero); err != nil {
		return err
	}
	if nonZero {
		return bankBookConflict("bank_book_balance_not_zero", "id", book.Code)
	}
	return nil
}

// checkBankBookReferences ตรวจธนาคาร/สาขา/ประเภทบัญชีกับทะเบียน bank_codes และตั้ง bank_name จากชื่อธนาคาร;
// รหัสที่ปิดการใช้งานแล้วเลือกใหม่ไม่ได้ แต่สมุดเดิมที่ใช้รหัสนั้นอยู่แก้ฟิลด์อื่นต่อได้ (old != nil และรหัสไม่เปลี่ยน)
func checkBankBookReferences(ctx context.Context, tx *sql.Tx, company string, book, old *SubledgerBankAccount) error {
	pick := func(kind, code, oldCode, field string) (cheque.Code, error) {
		found, ok, err := cheque.LookupCode(ctx, tx, company, kind, code)
		if err != nil {
			return found, err
		}
		if !ok {
			return found, bankBookInvalid("bank_book_ref_not_found", field, code)
		}
		if !found.IsActive && code != oldCode {
			return found, bankBookInvalid("bank_book_ref_inactive", field, code)
		}
		return found, nil
	}
	var oldBank, oldBranch, oldType string
	if old != nil {
		oldBank, oldBranch, oldType = old.BankCode, old.BankBranchCode, old.AccountTypeCode
	}
	bank, err := pick(cheque.KindBank, book.BankCode, oldBank, "bank_code")
	if err != nil {
		return err
	}
	book.BankName = bank.Name
	if book.BankBranchCode != "" {
		branch, err := pick(cheque.KindBankBranch, book.BankBranchCode, oldBranch, "bank_branch_code")
		if err != nil {
			return err
		}
		// สาขาที่ไม่ระบุธนาคารใช้ได้กับทุกธนาคาร
		if branch.BankCode != "" && branch.BankCode != book.BankCode {
			return bankBookInvalid("bank_book_branch_bank_mismatch", "bank_branch_code", book.BankBranchCode)
		}
	}
	if book.AccountTypeCode != "" {
		if _, err = pick(cheque.KindBankAccountType, book.AccountTypeCode, oldType, "account_type_code"); err != nil {
			return err
		}
	}
	return nil
}

// checkBankBookGLAccount: บัญชีของสมุดต้องเป็นบัญชีสินทรัพย์ที่เปิดใช้งานและลงรายการได้ (ใช้คุณสมบัติของบัญชี ไม่ดูเลขรหัส)
func checkBankBookGLAccount(ctx context.Context, tx *sql.Tx, company, code string) error {
	var payload []byte
	err := tx.QueryRowContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='accounts' AND code=$2`, company, code).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return bankBookInvalid("bank_book_gl_account_invalid", "gl_account_code", code)
	}
	if err != nil {
		return err
	}
	var account Account
	if err = json.Unmarshal(payload, &account); err != nil {
		return err
	}
	if account.IsDeleted || !account.IsActive || !account.AllowPosting || account.AccountType != "asset" {
		return bankBookInvalid("bank_book_gl_account_invalid", "gl_account_code", code)
	}
	return nil
}

// checkBankBookNumberFree: เลขที่บัญชีธนาคารต้องไม่ซ้ำกับสมุดอื่นของบริษัท (Champ: BookNo เป็น PK; ไม่แยกตัวพิมพ์ใหญ่เล็ก)
func checkBankBookNumberFree(ctx context.Context, tx *sql.Tx, company, exceptCode, number, field string) error {
	var owner string
	err := tx.QueryRowContext(ctx, `SELECT code FROM gl_subledger_bank_accounts WHERE company=$1 AND code<>$2 AND upper(btrim(payload->>'account_number'))=upper($3) ORDER BY code LIMIT 1`,
		company, exceptCode, number).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return bankBookConflict("bank_book_account_number_duplicate", field, number, owner)
}

// bankBookUsed: มีรายการเดินบัญชี ใบรายการธนาคาร ทะเบียนเช็ค หรือเอกสารเปลี่ยนสถานะเช็คอ้างสมุดนี้แล้วหรือยัง
func bankBookUsed(ctx context.Context, tx *sql.Tx, company, code string) (bool, error) {
	var used bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_bank_lines WHERE company=$1 AND bank_account_code=$2 UNION ALL SELECT 1 FROM gl_subledger_statements WHERE company=$1 AND bank_account_code=$2 UNION ALL SELECT 1 FROM cheque_registers WHERE company=$1 AND book_code=$2 AND deleted_at IS NULL UNION ALL SELECT 1 FROM cheque_docs WHERE company=$1 AND book_code=$2)`, company, code).Scan(&used)
	return used, err
}

func loadBankBook(ctx context.Context, tx *sql.Tx, company, code string) (SubledgerBankAccount, error) {
	var book SubledgerBankAccount
	var raw []byte
	var version int64
	err := tx.QueryRowContext(ctx, `SELECT payload,version FROM gl_subledger_bank_accounts WHERE company=$1 AND code=$2`, company, code).Scan(&raw, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return book, ErrNotFound
	}
	if err != nil {
		return book, err
	}
	if err = json.Unmarshal(raw, &book); err != nil {
		return book, err
	}
	book.Version = version
	return book, nil
}

func saveBankBook(ctx context.Context, tx *sql.Tx, company string, book SubledgerBankAccount) error {
	raw, err := json.Marshal(book)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO gl_subledger_bank_accounts(company,code,version,payload) VALUES($1,$2,$3,$4) ON CONFLICT(company,code) DO UPDATE SET version=EXCLUDED.version,payload=EXCLUDED.payload`,
		company, book.Code, book.Version, string(raw))
	return err
}

// auditBankBook ลงประวัติแบบ append-only (gl_subledger_audit) พร้อมภาพก่อน/หลัง; request_id + request_hash ใช้ตรวจคำขอซ้ำด้วย
func auditBankBook(ctx context.Context, tx *sql.Tx, scope Scope, plan bankAccountPlan, requestID, hash string, before *SubledgerBankAccount, after SubledgerBankAccount, now time.Time) error {
	payload := map[string]any{"request_id": requestID, "request_hash": hash, "version": after.Version, "before": before}
	if plan.action != "delete" {
		payload["after"] = after
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	action := "bank_account." + strings.ReplaceAll(plan.action, "-", "_")
	_, err = tx.ExecContext(ctx, `INSERT INTO gl_subledger_audit(company,journal_id,action,actor,occurred_at,payload) VALUES($1,$2,$3,$4,$5,$6)`,
		scope.Company, bankBookAuditPrefix+plan.code, action, scope.Actor, now, string(raw))
	return err
}
