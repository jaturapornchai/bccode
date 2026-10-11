package generalledger

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"smlcloudplatform/internal/goapi/language"
)

const bankTestCompany = "01"

var bankTestNow = time.Date(2026, 10, 11, 3, 0, 0, 0, time.UTC)

func bankTestScope() Scope {
	return Scope{Holding: "rungrueng", Company: bankTestCompany, Actor: "uid-jead"}
}

func asUserError(t *testing.T, err error) *UserError {
	t.Helper()
	var userErr *UserError
	if !errors.As(err, &userErr) {
		t.Fatalf("expected *UserError, got %T: %v", err, err)
	}
	return userErr
}

func expectBankBookError(t *testing.T, err error, code, field string, status int) *UserError {
	t.Helper()
	userErr := asUserError(t, err)
	if userErr.Code != code || userErr.Field != field || userErr.Status != status {
		t.Fatalf("got code=%q field=%q status=%d, want code=%q field=%q status=%d (%s)", userErr.Code, userErr.Field, userErr.Status, code, field, status, userErr.Message)
	}
	if strings.TrimSpace(userErr.Message) == "" || userErr.Message == "gl_err_"+code {
		t.Fatalf("code %s has no Thai message: %q", code, userErr.Message)
	}
	return userErr
}

func validBankBook() SubledgerBankAccount {
	return SubledgerBankAccount{
		Code:            "KBANK-01",
		AccountNumber:   "123-4-56789-0",
		AccountName:     "บริษัท รุ่งเรืองการค้า จำกัด",
		BankCode:        "KBANK",
		BankBranchCode:  "0123",
		AccountTypeCode: "CURRENT",
		GLAccountCode:   "11120",
		Currency:        "THB",
		IsActive:        true,
	}
}

// storedBankBook คือสมุดที่บันทึกแล้ว (มีชื่อธนาคารที่ระบบตั้งให้ตอนบันทึก)
func storedBankBook(version int64) SubledgerBankAccount {
	book := validBankBook()
	book.BankName = "ธนาคารกสิกรไทย"
	book.Version = version
	return book
}

func TestBankBookErrorRowsAreCompleteAndConsistent(t *testing.T) {
	// จำนวนค่าที่แทน {0},{1} ของแต่ละรหัส — ต้องตรงกับที่ bank_accounts.go ส่งเป็น Args
	wantArgs := map[string]int{
		"bank_book_payload_required": 0, "bank_book_name_required": 0, "bank_book_name_too_long": 1, "bank_book_number_required": 0,
		"bank_book_number_too_long": 1, "bank_book_bank_required": 0, "bank_book_currency_not_supported": 0, "bank_book_gl_account_invalid": 1,
		"bank_book_ref_not_found": 1, "bank_book_ref_inactive": 1, "bank_book_branch_bank_mismatch": 1, "bank_book_duplicate": 1,
		"bank_book_account_number_duplicate": 2, "bank_book_number_use_change": 0, "bank_book_number_unchanged": 0,
		"bank_book_gl_account_locked": 1, "bank_book_in_use": 1, "bank_book_balance_not_zero": 1,
	}
	if len(bankBookErrorCodes) != len(wantArgs) {
		t.Fatalf("bankBookErrorCodes has %d codes, expected %d", len(bankBookErrorCodes), len(wantArgs))
	}
	placeholder := regexp.MustCompile(`\{\d+\}`)
	for _, code := range bankBookErrorCodes {
		want, known := wantArgs[code]
		if !known {
			t.Fatalf("unexpected error code %s", code)
		}
		key := "gl_err_" + code
		th, en := language.Text(key, "th"), language.Text(key, "en")
		if th == key || th == "" || en == key || en == "" || th == en {
			t.Fatalf("%s needs a real Thai and English row (th=%q en=%q)", key, th, en)
		}
		if cn := language.Text(key, "cn"); cn != en {
			t.Fatalf("%s: other languages must carry the English text, got %q", key, cn)
		}
		thSlots, enSlots := placeholder.FindAllString(th, -1), placeholder.FindAllString(en, -1)
		if len(thSlots) != want || len(enSlots) != want {
			t.Fatalf("%s: placeholders th=%v en=%v, want %d each", key, thSlots, enSlots, want)
		}
	}
}

func TestBankBookErrorSubstitutesArgsInThai(t *testing.T) {
	err := bankBookConflict("bank_book_account_number_duplicate", "account_number", "123-4-56789-0", "KBANK-02")
	if !strings.Contains(err.Message, "123-4-56789-0") || !strings.Contains(err.Message, "KBANK-02") || strings.Contains(err.Message, "{") {
		t.Fatalf("placeholders not substituted: %q", err.Message)
	}
	if len(err.Args) != 2 || err.Status != http.StatusConflict {
		t.Fatalf("args/status wrong: %+v", err)
	}
}

func TestPlanBankAccountRejectsMalformedCommands(t *testing.T) {
	good := validBankBook()
	with := func(edit func(*SubledgerBankAccount)) *SubledgerBankAccount {
		book := good
		edit(&book)
		return &book
	}
	long := func(n int) string { return strings.Repeat("ก", n) }
	cases := []struct {
		name   string
		cmd    Command
		code   string
		field  string
		status int
	}{
		{"create without payload", Command{Action: "create"}, "bank_book_payload_required", "", 400},
		{"update with another book's payload", Command{Action: "update", ID: "KBANK-02", BankAccount: &good}, "bank_book_payload_required", "bank_account_code", 400},
		{"update without id", Command{Action: "update", BankAccount: &good}, "bank_book_payload_required", "bank_account_code", 400},
		{"delete without id", Command{Action: "delete"}, "bank_book_payload_required", "id", 400},
		{"change number without id", Command{Action: ChangeAccountNumberAction, NewAccountNumber: "999"}, "bank_book_payload_required", "id", 400},
		{"change number blank", Command{Action: ChangeAccountNumberAction, ID: "KBANK-01", NewAccountNumber: "   "}, "bank_book_number_required", "new_account_number", 400},
		{"change number too long", Command{Action: ChangeAccountNumberAction, ID: "KBANK-01", NewAccountNumber: strings.Repeat("9", 51)}, "bank_book_number_too_long", "new_account_number", 400},
		{"create without name", Command{Action: "create", BankAccount: with(func(b *SubledgerBankAccount) { b.AccountName = "  " })}, "bank_book_name_required", "account_name", 400},
		{"create name over 255 characters", Command{Action: "create", BankAccount: with(func(b *SubledgerBankAccount) { b.AccountName = long(256) })}, "bank_book_name_too_long", "account_name", 400},
		{"create without bank", Command{Action: "create", BankAccount: with(func(b *SubledgerBankAccount) { b.BankCode = "" })}, "bank_book_bank_required", "bank_code", 400},
		{"create in a foreign currency", Command{Action: "create", BankAccount: with(func(b *SubledgerBankAccount) { b.Currency = "USD" })}, "bank_book_currency_not_supported", "currency_code", 400},
		{"create without account number", Command{Action: "create", BankAccount: with(func(b *SubledgerBankAccount) { b.AccountNumber = " " })}, "bank_book_number_required", "account_number", 400},
		{"create number over 50 characters", Command{Action: "create", BankAccount: with(func(b *SubledgerBankAccount) { b.AccountNumber = strings.Repeat("1", 51) })}, "bank_book_number_too_long", "account_number", 400},
		{"create with a space in the code", Command{Action: "create", BankAccount: with(func(b *SubledgerBankAccount) { b.Code = "KB 01" })}, "code_has_space", "bank_account_code", 400},
		{"create without GL account", Command{Action: "create", BankAccount: with(func(b *SubledgerBankAccount) { b.GLAccountCode = "" })}, "code_required", "gl_account_code", 400},
		{"create with a 21-character code", Command{Action: "create", BankAccount: with(func(b *SubledgerBankAccount) { b.Code = strings.Repeat("A", 21) })}, "code_too_long", "bank_account_code", 400},
		{"unsupported action", Command{Action: "post", ID: "KBANK-01"}, CodeUnsupported, "", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := planBankAccount(tc.cmd)
			userErr := asUserError(t, err)
			if userErr.Code != tc.code || userErr.Field != tc.field || userErr.Status != tc.status {
				t.Fatalf("got code=%q field=%q status=%d, want code=%q field=%q status=%d", userErr.Code, userErr.Field, userErr.Status, tc.code, tc.field, tc.status)
			}
		})
	}
}

func TestPlanBankAccountAcceptsAndCleansValidCommands(t *testing.T) {
	book := validBankBook()
	book.AccountName = "  " + book.AccountName + " "
	book.AccountNumber = " 123-4-56789-0 "
	book.Currency = "thb"
	plan, err := planBankAccount(Command{Action: "create", BankAccount: &book})
	if err != nil {
		t.Fatal(err)
	}
	if plan.code != "KBANK-01" || plan.book.AccountName != "บริษัท รุ่งเรืองการค้า จำกัด" || plan.book.AccountNumber != "123-4-56789-0" || plan.book.Currency != "THB" {
		t.Fatalf("book not cleaned: %+v", plan)
	}
	// เงินบาทเป็นค่าเริ่มต้นเมื่อไม่ได้ส่งสกุลเงิน
	noCurrency := validBankBook()
	noCurrency.Currency = ""
	if plan, err = planBankAccount(Command{Action: "update", ID: "KBANK-01", Version: 3, BankAccount: &noCurrency}); err != nil || plan.book.Currency != "THB" || plan.version != 3 {
		t.Fatalf("update plan: %+v %v", plan, err)
	}
	// ตัดช่องว่างหัวท้ายของเลขที่ใหม่ แต่ตัวอักษรอื่นคงเดิม (เลขที่บัญชีเป็นข้อความ ไม่ใช่ตัวเลข)
	plan, err = planBankAccount(Command{Action: ChangeAccountNumberAction, ID: " KBANK-01 ", Version: 5, NewAccountNumber: " 007-1 "})
	if err != nil || plan.newNumber != "007-1" || plan.code != "KBANK-01" {
		t.Fatalf("change plan: %+v %v", plan, err)
	}
}

func TestExecuteBankAccountRejectsBadCommandBeforeTouchingTheDatabase(t *testing.T) {
	// ไม่มีฐานข้อมูล (pg ว่าง): ถ้าโค้ดเลยจุดตรวจรูปแบบไปจะ panic
	_, err := (&PostgresStore{}).executeBankAccount(context.Background(), bankTestScope(), Command{Action: "create", RequestID: "r1"})
	expectBankBookError(t, err, "bank_book_payload_required", "", http.StatusBadRequest)
}

// ---- sqlmock helpers ----

func bankTx(t *testing.T) (*sql.Tx, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	return tx, mock
}

func expectBookRow(mock sqlmock.Sqlmock, book SubledgerBankAccount) {
	raw, _ := json.Marshal(book)
	mock.ExpectQuery(`SELECT payload,version FROM gl_subledger_bank_accounts WHERE company=\$1 AND code=\$2`).WithArgs(bankTestCompany, book.Code).
		WillReturnRows(sqlmock.NewRows([]string{"payload", "version"}).AddRow(raw, book.Version))
}

func expectNoBookRow(mock sqlmock.Sqlmock, code string) {
	mock.ExpectQuery(`SELECT payload,version FROM gl_subledger_bank_accounts WHERE company=\$1 AND code=\$2`).WithArgs(bankTestCompany, code).WillReturnError(sql.ErrNoRows)
}

// expectCode คาดการค้นรหัสในทะเบียน bank_codes (FOR SHARE) — found=false คือไม่มีรหัสนี้
func expectCode(mock sqlmock.Sqlmock, kind, code, name, bankCode string, active, found bool) {
	query := mock.ExpectQuery(`FROM bank_codes WHERE company = \$1 AND kind = \$2 AND code = \$3 FOR SHARE`).WithArgs(bankTestCompany, kind, code)
	if !found {
		query.WillReturnError(sql.ErrNoRows)
		return
	}
	query.WillReturnRows(sqlmock.NewRows([]string{"kind", "code", "name", "bank_code", "user_group", "is_active", "version"}).AddRow(kind, code, name, bankCode, "", active, 1))
}

func expectValidReferences(mock sqlmock.Sqlmock) {
	expectCode(mock, "bank", "KBANK", "ธนาคารกสิกรไทย", "", true, true)
	expectCode(mock, "bank_branch", "0123", "สาขาสีลม", "KBANK", true, true)
	expectCode(mock, "bank_account_type", "CURRENT", "กระแสรายวัน", "", true, true)
}

func expectGLAccount(mock sqlmock.Sqlmock, code string, account *Account) {
	query := mock.ExpectQuery(`SELECT payload FROM gl_records WHERE company=\$1 AND kind='accounts' AND code=\$2`).WithArgs(bankTestCompany, code)
	if account == nil {
		query.WillReturnError(sql.ErrNoRows)
		return
	}
	raw, _ := json.Marshal(account)
	query.WillReturnRows(sqlmock.NewRows([]string{"payload"}).AddRow(raw))
}

func postableAsset() *Account {
	return &Account{AccountCode: "11120", AccountType: "asset", AllowPosting: true, IsActive: true}
}

func expectNumberFree(mock sqlmock.Sqlmock, exceptCode, number, owner string) {
	query := mock.ExpectQuery(`SELECT code FROM gl_subledger_bank_accounts WHERE company=\$1 AND code<>\$2 AND upper\(btrim\(payload->>'account_number'\)\)=upper\(\$3\)`).WithArgs(bankTestCompany, exceptCode, number)
	if owner == "" {
		query.WillReturnError(sql.ErrNoRows)
		return
	}
	query.WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow(owner))
}

func expectUsed(mock sqlmock.Sqlmock, code string, used bool) {
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM gl_subledger_bank_lines WHERE company=\$1 AND bank_account_code=\$2 UNION ALL SELECT 1 FROM gl_subledger_statements`).WithArgs(bankTestCompany, code).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(used))
}

// savedBook เก็บสมุดที่ถูกเขียนลงตารางเพื่อตรวจ payload หลังเรียกเสร็จ
type savedBook struct{ payload string }

func (s *savedBook) Match(value driver.Value) bool {
	text, ok := value.(string)
	s.payload = text
	return ok
}

func (s *savedBook) decode(t *testing.T) SubledgerBankAccount {
	t.Helper()
	var book SubledgerBankAccount
	if err := json.Unmarshal([]byte(s.payload), &book); err != nil {
		t.Fatalf("saved payload is not a bank book: %q", s.payload)
	}
	return book
}

func expectSave(mock sqlmock.Sqlmock, code string, version int64, saved *savedBook) {
	mock.ExpectExec(`INSERT INTO gl_subledger_bank_accounts\(company,code,version,payload\) VALUES\(\$1,\$2,\$3,\$4\) ON CONFLICT\(company,code\) DO UPDATE`).
		WithArgs(bankTestCompany, code, version, saved).WillReturnResult(sqlmock.NewResult(0, 1))
}

func expectAudit(mock sqlmock.Sqlmock, code, action string, audit *savedBook) {
	mock.ExpectExec(`INSERT INTO gl_subledger_audit\(company,journal_id,action,actor,occurred_at,payload\)`).
		WithArgs(bankTestCompany, "bank-account:"+code, action, "uid-jead", bankTestNow, audit).WillReturnResult(sqlmock.NewResult(0, 1))
}

func runBankAccount(t *testing.T, tx *sql.Tx, cmd Command) (Result, error) {
	t.Helper()
	plan, err := planBankAccount(cmd)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	return applyBankAccount(context.Background(), tx, bankTestScope(), "11111111-2222-4333-8444-555555555555", "hash-1", plan, bankTestNow)
}

func finish(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

// ---- create ----

func TestCreateBankBookChecksEverythingThenSavesWithBankNameAndAudit(t *testing.T) {
	tx, mock := bankTx(t)
	mock.ExpectQuery(`SELECT code FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper\(code\)=upper\(\$2\)`).WithArgs(bankTestCompany, "KBANK-01").WillReturnError(sql.ErrNoRows)
	expectValidReferences(mock)
	expectGLAccount(mock, "11120", postableAsset())
	expectNumberFree(mock, "KBANK-01", "123-4-56789-0", "")
	var saved, audit savedBook
	expectSave(mock, "KBANK-01", 1, &saved)
	expectAudit(mock, "KBANK-01", "bank_account.create", &audit)

	result, err := runBankAccount(t, tx, Command{Action: "create", BankAccount: ptr(validBankBook())})
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "KBANK-01" || result.Version != 1 {
		t.Fatalf("result: %+v", result)
	}
	book := saved.decode(t)
	if book.BankName != "ธนาคารกสิกรไทย" || book.BankCode != "KBANK" || book.BankBranchCode != "0123" || book.AccountTypeCode != "CURRENT" || book.Currency != "THB" || book.Version != 1 {
		t.Fatalf("saved book: %+v", book)
	}
	if !strings.Contains(audit.payload, `"request_hash":"hash-1"`) || !strings.Contains(audit.payload, `"after"`) || strings.Contains(audit.payload, `"before":{`) {
		t.Fatalf("audit payload: %s", audit.payload)
	}
	finish(t, mock)
}

func TestCreateBankBookWithoutBranchAndTypeSkipsThoseLookups(t *testing.T) {
	tx, mock := bankTx(t)
	book := validBankBook()
	book.BankBranchCode, book.AccountTypeCode = "", ""
	mock.ExpectQuery(`SELECT code FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper\(code\)=upper\(\$2\)`).WithArgs(bankTestCompany, "KBANK-01").WillReturnError(sql.ErrNoRows)
	expectCode(mock, "bank", "KBANK", "ธนาคารกสิกรไทย", "", true, true)
	expectGLAccount(mock, "11120", postableAsset())
	expectNumberFree(mock, "KBANK-01", "123-4-56789-0", "")
	var saved, audit savedBook
	expectSave(mock, "KBANK-01", 1, &saved)
	expectAudit(mock, "KBANK-01", "bank_account.create", &audit)
	if _, err := runBankAccount(t, tx, Command{Action: "create", BankAccount: &book}); err != nil {
		t.Fatal(err)
	}
	finish(t, mock)
}

func TestCreateBankBookRejections(t *testing.T) {
	cases := []struct {
		name   string
		expect func(mock sqlmock.Sqlmock)
		code   string
		field  string
		status int
		arg    string
	}{
		{"code already used (any letter case)", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`SELECT code FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper\(code\)=upper\(\$2\)`).WithArgs(bankTestCompany, "KBANK-01").WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow("kbank-01"))
		}, "bank_book_duplicate", "bank_account_code", 409, "kbank-01"},
		{"bank code not in the register", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper`).WithArgs(bankTestCompany, "KBANK-01").WillReturnError(sql.ErrNoRows)
			expectCode(m, "bank", "KBANK", "", "", false, false)
		}, "bank_book_ref_not_found", "bank_code", 400, "KBANK"},
		{"bank code inactive", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper`).WithArgs(bankTestCompany, "KBANK-01").WillReturnError(sql.ErrNoRows)
			expectCode(m, "bank", "KBANK", "ธนาคารกสิกรไทย", "", false, true)
		}, "bank_book_ref_inactive", "bank_code", 400, "KBANK"},
		{"branch of another bank", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper`).WithArgs(bankTestCompany, "KBANK-01").WillReturnError(sql.ErrNoRows)
			expectCode(m, "bank", "KBANK", "ธนาคารกสิกรไทย", "", true, true)
			expectCode(m, "bank_branch", "0123", "สาขาสีลม", "SCB", true, true)
		}, "bank_book_branch_bank_mismatch", "bank_branch_code", 400, "0123"},
		{"account type missing", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper`).WithArgs(bankTestCompany, "KBANK-01").WillReturnError(sql.ErrNoRows)
			expectCode(m, "bank", "KBANK", "ธนาคารกสิกรไทย", "", true, true)
			expectCode(m, "bank_branch", "0123", "สาขาสีลม", "KBANK", true, true)
			expectCode(m, "bank_account_type", "CURRENT", "", "", false, false)
		}, "bank_book_ref_not_found", "account_type_code", 400, "CURRENT"},
		{"GL account missing", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper`).WithArgs(bankTestCompany, "KBANK-01").WillReturnError(sql.ErrNoRows)
			expectValidReferences(m)
			expectGLAccount(m, "11120", nil)
		}, "bank_book_gl_account_invalid", "gl_account_code", 400, "11120"},
		{"GL account is not an asset", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper`).WithArgs(bankTestCompany, "KBANK-01").WillReturnError(sql.ErrNoRows)
			expectValidReferences(m)
			expectGLAccount(m, "11120", &Account{AccountType: "revenue", AllowPosting: true, IsActive: true})
		}, "bank_book_gl_account_invalid", "gl_account_code", 400, "11120"},
		{"GL account is a header (no posting)", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper`).WithArgs(bankTestCompany, "KBANK-01").WillReturnError(sql.ErrNoRows)
			expectValidReferences(m)
			expectGLAccount(m, "11120", &Account{AccountType: "asset", AllowPosting: false, IsActive: true})
		}, "bank_book_gl_account_invalid", "gl_account_code", 400, "11120"},
		{"GL account inactive", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper`).WithArgs(bankTestCompany, "KBANK-01").WillReturnError(sql.ErrNoRows)
			expectValidReferences(m)
			expectGLAccount(m, "11120", &Account{AccountType: "asset", AllowPosting: true, IsActive: false})
		}, "bank_book_gl_account_invalid", "gl_account_code", 400, "11120"},
		{"account number used by another book", func(m sqlmock.Sqlmock) {
			m.ExpectQuery(`FROM gl_subledger_bank_accounts WHERE company=\$1 AND upper`).WithArgs(bankTestCompany, "KBANK-01").WillReturnError(sql.ErrNoRows)
			expectValidReferences(m)
			expectGLAccount(m, "11120", postableAsset())
			expectNumberFree(m, "KBANK-01", "123-4-56789-0", "SCB-01")
		}, "bank_book_account_number_duplicate", "account_number", 409, "SCB-01"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tx, mock := bankTx(t)
			tc.expect(mock)
			_, err := runBankAccount(t, tx, Command{Action: "create", BankAccount: ptr(validBankBook())})
			userErr := expectBankBookError(t, err, tc.code, tc.field, tc.status)
			if !strings.Contains(userErr.Message, tc.arg) {
				t.Fatalf("message should name %q: %q", tc.arg, userErr.Message)
			}
			finish(t, mock) // ไม่มี INSERT ใดๆ หลังตรวจไม่ผ่าน
		})
	}
}

// ---- update ----

func TestUpdateBankBookRejections(t *testing.T) {
	t.Run("book not found", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectNoBookRow(mock, "KBANK-01")
		_, err := runBankAccount(t, tx, Command{Action: "update", ID: "KBANK-01", Version: 2, BankAccount: ptr(validBankBook())})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("want ErrNotFound, got %v", err)
		}
		finish(t, mock)
	})
	t.Run("stale version", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectBookRow(mock, storedBankBook(3))
		_, err := runBankAccount(t, tx, Command{Action: "update", ID: "KBANK-01", Version: 2, BankAccount: ptr(validBankBook())})
		if asUserError(t, err).Code != CodeStaleVersion {
			t.Fatalf("want stale_version, got %v", err)
		}
		finish(t, mock)
	})
	t.Run("version below 1", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectBookRow(mock, storedBankBook(3))
		_, err := runBankAccount(t, tx, Command{Action: "update", ID: "KBANK-01", BankAccount: ptr(validBankBook())})
		if asUserError(t, err).Code != CodeStaleVersion {
			t.Fatalf("want stale_version, got %v", err)
		}
		finish(t, mock)
	})
	t.Run("account number cannot be edited here", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectBookRow(mock, storedBankBook(3))
		book := validBankBook()
		book.AccountNumber = "999-9-99999-9"
		_, err := runBankAccount(t, tx, Command{Action: "update", ID: "KBANK-01", Version: 3, BankAccount: &book})
		expectBankBookError(t, err, "bank_book_number_use_change", "account_number", http.StatusBadRequest)
		finish(t, mock)
	})
	t.Run("GL account locked once the book has movements", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectBookRow(mock, storedBankBook(3))
		expectValidReferences(mock)
		expectGLAccount(mock, "11121", &Account{AccountType: "asset", AllowPosting: true, IsActive: true})
		expectUsed(mock, "KBANK-01", true)
		book := validBankBook()
		book.GLAccountCode = "11121"
		_, err := runBankAccount(t, tx, Command{Action: "update", ID: "KBANK-01", Version: 3, BankAccount: &book})
		expectBankBookError(t, err, "bank_book_gl_account_locked", "gl_account_code", http.StatusConflict)
		finish(t, mock)
	})
	t.Run("a bank that became inactive cannot be picked for another book", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectBookRow(mock, storedBankBook(3))
		expectCode(mock, "bank", "SCB", "ธนาคารไทยพาณิชย์", "", false, true)
		book := validBankBook()
		book.BankCode, book.BankBranchCode = "SCB", ""
		_, err := runBankAccount(t, tx, Command{Action: "update", ID: "KBANK-01", Version: 3, BankAccount: &book})
		expectBankBookError(t, err, "bank_book_ref_inactive", "bank_code", http.StatusBadRequest)
		finish(t, mock)
	})
}

func TestUpdateBankBookWithNoRealChangeWritesNothing(t *testing.T) {
	tx, mock := bankTx(t)
	expectBookRow(mock, storedBankBook(3))
	expectValidReferences(mock)
	result, err := runBankAccount(t, tx, Command{Action: "update", ID: "KBANK-01", Version: 3, BankAccount: ptr(validBankBook())})
	if err != nil || result.Version != 3 {
		t.Fatalf("result %+v err %v", result, err)
	}
	finish(t, mock) // ไม่มี INSERT ลงตารางสมุดหรือประวัติ
}

func TestUpdateBankBookKeepsAnInactiveCodeItAlreadyUses(t *testing.T) {
	tx, mock := bankTx(t)
	expectBookRow(mock, storedBankBook(3))
	// ธนาคารเดิมถูกปิดใช้งานทีหลัง: สมุดเดิมยังแก้ชื่อบัญชีต่อได้
	expectCode(mock, "bank", "KBANK", "ธนาคารกสิกรไทย", "", false, true)
	expectCode(mock, "bank_branch", "0123", "สาขาสีลม", "KBANK", true, true)
	expectCode(mock, "bank_account_type", "CURRENT", "กระแสรายวัน", "", true, true)
	var saved, audit savedBook
	expectSave(mock, "KBANK-01", 4, &saved)
	expectAudit(mock, "KBANK-01", "bank_account.update", &audit)
	book := validBankBook()
	book.AccountName = "บริษัท รุ่งเรืองการค้า จำกัด (สำนักงานใหญ่)"
	book.IsActive = false
	result, err := runBankAccount(t, tx, Command{Action: "update", ID: "KBANK-01", Version: 3, BankAccount: &book})
	if err != nil || result.Version != 4 {
		t.Fatalf("result %+v err %v", result, err)
	}
	got := saved.decode(t)
	if got.AccountName != book.AccountName || got.IsActive || got.AccountNumber != "123-4-56789-0" || got.Version != 4 {
		t.Fatalf("saved book: %+v", got)
	}
	if !strings.Contains(audit.payload, `"before":{`) || !strings.Contains(audit.payload, `"after":{`) {
		t.Fatalf("audit must carry before and after: %s", audit.payload)
	}
	finish(t, mock)
}

// ---- change account number ----

func TestChangeAccountNumberChangesOnlyTheNumber(t *testing.T) {
	tx, mock := bankTx(t)
	expectBookRow(mock, storedBankBook(5))
	expectNumberFree(mock, "KBANK-01", "999-9-99999-9", "")
	var saved, audit savedBook
	// มีแค่ INSERT สมุด + ประวัติ: ไม่มีการแก้ตารางอื่น (ใบสำคัญอ้างรหัสสมุด ไม่ใช่เลขที่บัญชี)
	expectSave(mock, "KBANK-01", 6, &saved)
	expectAudit(mock, "KBANK-01", "bank_account.change_account_number", &audit)
	result, err := runBankAccount(t, tx, Command{Action: ChangeAccountNumberAction, ID: "KBANK-01", Version: 5, NewAccountNumber: "999-9-99999-9"})
	if err != nil || result.ID != "KBANK-01" || result.Version != 6 {
		t.Fatalf("result %+v err %v", result, err)
	}
	got, before := saved.decode(t), storedBankBook(5)
	if got.AccountNumber != "999-9-99999-9" {
		t.Fatalf("number not changed: %+v", got)
	}
	got.AccountNumber, got.Version = before.AccountNumber, before.Version
	if !jsonEqual(got, before) {
		t.Fatalf("something besides the number changed: %+v vs %+v", got, before)
	}
	if !strings.Contains(audit.payload, `"123-4-56789-0"`) || !strings.Contains(audit.payload, `"999-9-99999-9"`) {
		t.Fatalf("audit must show old and new number: %s", audit.payload)
	}
	finish(t, mock)
}

func TestChangeAccountNumberRejections(t *testing.T) {
	t.Run("same number", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectBookRow(mock, storedBankBook(5))
		_, err := runBankAccount(t, tx, Command{Action: ChangeAccountNumberAction, ID: "KBANK-01", Version: 5, NewAccountNumber: "123-4-56789-0"})
		expectBankBookError(t, err, "bank_book_number_unchanged", "new_account_number", http.StatusBadRequest)
		finish(t, mock)
	})
	t.Run("number belongs to another book", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectBookRow(mock, storedBankBook(5))
		expectNumberFree(mock, "KBANK-01", "555-5-55555-5", "SCB-01")
		_, err := runBankAccount(t, tx, Command{Action: ChangeAccountNumberAction, ID: "KBANK-01", Version: 5, NewAccountNumber: "555-5-55555-5"})
		userErr := expectBankBookError(t, err, "bank_book_account_number_duplicate", "new_account_number", http.StatusConflict)
		if !strings.Contains(userErr.Message, "SCB-01") {
			t.Fatalf("message should name the owner: %q", userErr.Message)
		}
		finish(t, mock)
	})
	t.Run("stale version", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectBookRow(mock, storedBankBook(6))
		_, err := runBankAccount(t, tx, Command{Action: ChangeAccountNumberAction, ID: "KBANK-01", Version: 5, NewAccountNumber: "999"})
		if asUserError(t, err).Code != CodeStaleVersion {
			t.Fatalf("want stale_version, got %v", err)
		}
		finish(t, mock)
	})
}

// ---- delete ----

func TestDeleteBankBook(t *testing.T) {
	t.Run("used by movements or statements", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectBookRow(mock, storedBankBook(2))
		expectUsed(mock, "KBANK-01", true)
		_, err := runBankAccount(t, tx, Command{Action: "delete", ID: "KBANK-01", Version: 2})
		expectBankBookError(t, err, "bank_book_in_use", "id", http.StatusConflict)
		finish(t, mock)
	})
	t.Run("GL account still has a balance (Champ MSG_891)", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectBookRow(mock, storedBankBook(2))
		expectUsed(mock, "KBANK-01", false)
		mock.ExpectQuery(`SELECT COALESCE\(SUM\(debit-credit\),0\)<>0 FROM gl_lines WHERE company=\$1 AND account_code=\$2`).WithArgs(bankTestCompany, "11120").WillReturnRows(sqlmock.NewRows([]string{"nonzero"}).AddRow(true))
		_, err := runBankAccount(t, tx, Command{Action: "delete", ID: "KBANK-01", Version: 2})
		expectBankBookError(t, err, "bank_book_balance_not_zero", "id", http.StatusConflict)
		finish(t, mock)
	})
	t.Run("unused book with a zero balance is removed and audited", func(t *testing.T) {
		tx, mock := bankTx(t)
		expectBookRow(mock, storedBankBook(2))
		expectUsed(mock, "KBANK-01", false)
		mock.ExpectQuery(`SELECT COALESCE\(SUM\(debit-credit\),0\)<>0 FROM gl_lines`).WithArgs(bankTestCompany, "11120").WillReturnRows(sqlmock.NewRows([]string{"nonzero"}).AddRow(false))
		mock.ExpectExec(`DELETE FROM gl_subledger_bank_accounts WHERE company=\$1 AND code=\$2`).WithArgs(bankTestCompany, "KBANK-01").WillReturnResult(sqlmock.NewResult(0, 1))
		var audit savedBook
		expectAudit(mock, "KBANK-01", "bank_account.delete", &audit)
		result, err := runBankAccount(t, tx, Command{Action: "delete", ID: "KBANK-01", Version: 2})
		if err != nil || result.ID != "KBANK-01" || result.Version != 2 {
			t.Fatalf("result %+v err %v", result, err)
		}
		if strings.Contains(audit.payload, `"after"`) || !strings.Contains(audit.payload, `"before":{`) {
			t.Fatalf("delete audit keeps only the before image: %s", audit.payload)
		}
		finish(t, mock)
	})
}

// ---- replay ----

func TestReplayBankAccount(t *testing.T) {
	query := `SELECT payload FROM gl_subledger_audit WHERE company=\$1 AND journal_id=\$2 AND payload->>'request_id'=\$3`
	t.Run("first time", func(t *testing.T) {
		tx, mock := bankTx(t)
		mock.ExpectQuery(query).WithArgs(bankTestCompany, "bank-account:KBANK-01", "r1").WillReturnError(sql.ErrNoRows)
		_, found, err := replayBankAccount(context.Background(), tx, bankTestCompany, "KBANK-01", "r1", "h1")
		if err != nil || found {
			t.Fatalf("found=%v err=%v", found, err)
		}
		finish(t, mock)
	})
	t.Run("same request again returns the original result", func(t *testing.T) {
		tx, mock := bankTx(t)
		mock.ExpectQuery(query).WithArgs(bankTestCompany, "bank-account:KBANK-01", "r1").WillReturnRows(sqlmock.NewRows([]string{"payload"}).AddRow([]byte(`{"request_hash":"h1","version":4}`)))
		result, found, err := replayBankAccount(context.Background(), tx, bankTestCompany, "KBANK-01", "r1", "h1")
		if err != nil || !found || result.ID != "KBANK-01" || result.Version != 4 {
			t.Fatalf("result=%+v found=%v err=%v", result, found, err)
		}
		finish(t, mock)
	})
	t.Run("same request id with different data is refused", func(t *testing.T) {
		tx, mock := bankTx(t)
		mock.ExpectQuery(query).WithArgs(bankTestCompany, "bank-account:KBANK-01", "r1").WillReturnRows(sqlmock.NewRows([]string{"payload"}).AddRow([]byte(`{"request_hash":"other","version":4}`)))
		_, _, err := replayBankAccount(context.Background(), tx, bankTestCompany, "KBANK-01", "r1", "h1")
		if asUserError(t, err).Code != CodeDuplicateRequest {
			t.Fatalf("want duplicate_request, got %v", err)
		}
		finish(t, mock)
	})
}

func ptr[T any](value T) *T { return &value }
