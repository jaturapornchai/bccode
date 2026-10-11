//go:build integration

package generalledger

import (
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/google/uuid"

	"smlcloudplatform/internal/cheque"
)

func bankBookFixture(t *testing.T) *pgIntegrityFixture {
	t.Helper()
	f := newPGIntegrityFixture(t)
	if err := cheque.EnsureSchema(f.ctx, f.db); err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct{ kind, code, name, bank string }{
		{"bank", "KBANK", "ธนาคารกสิกรไทย", ""},
		{"bank", "SCB", "ธนาคารไทยพาณิชย์", ""},
		{"bank_branch", "0123", "สาขาสีลม", "KBANK"},
		{"bank_branch", "0456", "สาขาสีลม", "SCB"},
		{"bank_account_type", "CURRENT", "กระแสรายวัน", ""},
	} {
		var bank any
		if row.bank != "" {
			bank = row.bank
		}
		if _, err := f.db.Exec(`INSERT INTO bank_codes(company,kind,code,name,bank_code,created_by,updated_by) VALUES($1,$2,$3,$4,$5,'tester','tester')`, f.scope.Company, row.kind, row.code, row.name, bank); err != nil {
			t.Fatal(err)
		}
	}
	for _, code := range []string{"1100", "1110", "1120", "1130"} {
		account := Account{AccountCode: code, AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "เงินฝากธนาคาร " + code}}}
		f.run(Command{Resource: "accounts", Action: "create", Account: &account})
	}
	return f
}

func bankBookInput(code, number, glAccount string) SubledgerBankAccount {
	return SubledgerBankAccount{Code: code, AccountNumber: number, AccountName: "บริษัท รุ่งเรืองการค้า จำกัด", BankCode: "KBANK", BankBranchCode: "0123", AccountTypeCode: "CURRENT", GLAccountCode: glAccount, Currency: "THB", IsActive: true}
}

func (f *pgIntegrityFixture) bankBookError(cmd Command, code string) *UserError {
	f.t.Helper()
	_, err := f.execute(f.scope, cmd)
	if err == nil {
		f.t.Fatalf("expected %s, command succeeded", code)
	}
	userErr := asUserError(f.t, err)
	if userErr.Code != code {
		f.t.Fatalf("expected %s, got %s: %s", code, userErr.Code, userErr.Message)
	}
	return userErr
}

func (f *pgIntegrityFixture) bankBooks(q string) []map[string]any {
	f.t.Helper()
	page, err := f.store.pg.SubledgerList(f.ctx, f.scope, BankAccountKind, q, 1, 50, "")
	if err != nil {
		f.t.Fatal(err)
	}
	books := []map[string]any{}
	for _, item := range page.Items {
		var book map[string]any
		if err = json.Unmarshal(item, &book); err != nil {
			f.t.Fatal(err)
		}
		books = append(books, book)
	}
	return books
}

// bankBookDecimalEquals เทียบจำนวนเงินแบบข้อความทศนิยมกับค่าที่คาด (NUMERIC ของ PostgreSQL เติมศูนย์ท้ายได้) โดยไม่ผ่าน float
func bankBookDecimalEquals(got any, want string) bool {
	text, ok := got.(string)
	if !ok {
		return false
	}
	a, okA := new(big.Rat).SetString(text)
	b, okB := new(big.Rat).SetString(want)
	return okA && okB && a.Cmp(b) == 0
}

func (f *pgIntegrityFixture) auditActions(code string) []string {
	f.t.Helper()
	rows, err := f.db.Query(`SELECT action FROM gl_subledger_audit WHERE company=$1 AND journal_id=$2 ORDER BY event_no`, f.scope.Company, bankBookAuditPrefix+code)
	if err != nil {
		f.t.Fatal(err)
	}
	defer rows.Close()
	var actions []string
	for rows.Next() {
		var action string
		if err = rows.Scan(&action); err != nil {
			f.t.Fatal(err)
		}
		actions = append(actions, action)
	}
	return actions
}

func TestPostgresBankBookLifecycle(t *testing.T) {
	f := bankBookFixture(t)

	// สร้าง: บันทึกครบ 3 ฟิลด์ใหม่ + ชื่อธนาคารที่ระบบตั้งให้ + เริ่ม version 1
	book := bankBookInput("KBANK-01", "123-4-56789-0", "1100")
	created := f.run(Command{Resource: BankAccountKind, Action: "create", BankAccount: &book})
	if created.ID != "KBANK-01" || created.Version != 1 {
		t.Fatalf("created %+v", created)
	}
	rows := f.bankBooks("")
	if len(rows) != 1 || rows[0]["bank_name"] != "ธนาคารกสิกรไทย" || rows[0]["bank_code"] != "KBANK" || rows[0]["bank_branch_code"] != "0123" || rows[0]["account_type_code"] != "CURRENT" || !bankBookDecimalEquals(rows[0]["gl_balance"], "0") {
		t.Fatalf("listed %v", rows)
	}

	// ตรวจซ้ำ/อ้างอิง/บัญชีผังบัญชีที่ใช้ไม่ได้ — ทุกกรณีไม่สร้างแถวเพิ่ม
	dup := bankBookInput("kbank-01", "555", "1110")
	f.bankBookError(Command{Resource: BankAccountKind, Action: "create", BankAccount: &dup}, "bank_book_duplicate")
	sameNumber := bankBookInput("KBANK-02", "123-4-56789-0", "1110")
	if userErr := f.bankBookError(Command{Resource: BankAccountKind, Action: "create", BankAccount: &sameNumber}, "bank_book_account_number_duplicate"); !strings.Contains(userErr.Message, "KBANK-01") {
		t.Fatalf("duplicate number must name the owner: %s", userErr.Message)
	}
	income := bankBookInput("KBANK-02", "777", "4000")
	f.bankBookError(Command{Resource: BankAccountKind, Action: "create", BankAccount: &income}, "bank_book_gl_account_invalid")
	wrongBranch := bankBookInput("KBANK-02", "777", "1110")
	wrongBranch.BankBranchCode = "0456"
	f.bankBookError(Command{Resource: BankAccountKind, Action: "create", BankAccount: &wrongBranch}, "bank_book_branch_bank_mismatch")
	unknownBank := bankBookInput("KBANK-02", "777", "1110")
	unknownBank.BankCode = "NOPE"
	f.bankBookError(Command{Resource: BankAccountKind, Action: "create", BankAccount: &unknownBank}, "bank_book_ref_not_found")
	if got := len(f.bankBooks("")); got != 1 {
		t.Fatalf("a rejected command left %d books", got)
	}

	// ส่งคำขอเดิมซ้ำ (requestid เดียวกัน ข้อมูลเดียวกัน) ต้องไม่สร้างซ้ำและได้ผลเดิม
	requestID := uuid.NewString()
	second := bankBookInput("KBANK-02", "999-9-99999-9", "1110")
	first := f.run(Command{Resource: BankAccountKind, Action: "create", RequestID: requestID, BankAccount: &second})
	again := f.run(Command{Resource: BankAccountKind, Action: "create", RequestID: requestID, BankAccount: &second})
	if first.Version != 1 || again.Version != 1 || len(f.bankBooks("")) != 2 || len(f.auditActions("KBANK-02")) != 1 {
		t.Fatalf("replay first=%+v again=%+v books=%d audit=%v", first, again, len(f.bankBooks("")), f.auditActions("KBANK-02"))
	}
	changed := second
	changed.AccountName = "ชื่ออื่น"
	f.bankBookError(Command{Resource: BankAccountKind, Action: "create", RequestID: requestID, BankAccount: &changed}, CodeDuplicateRequest)

	// ยอดคงเหลือมาจากรายการที่ผ่านแล้วของบัญชีที่ผูก (NUMERIC เป็นข้อความ ไม่ผ่าน float)
	f.ensureDimension(DepartmentKind, "D")
	f.ensureDimension(ProjectKind, "P")
	// RC1 อ้างสมุด KBANK-01 ผ่านรายการเดินบัญชี (bank line) จึงถือว่าสมุดนี้ถูกใช้แล้ว; RC2 ลงบัญชีของ KBANK-02 ตรงๆ โดยไม่อ้างสมุด
	receipt := func(docNo, account, amount string, details *JournalDetails) {
		t.Helper()
		journal := Journal{DocNo: docNo, Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "รับเงินโอนจากลูกค้า", Kind: "manual", Details: details, Lines: []Line{
			{AccountCode: account, Debit: Amount(amount), Credit: Amount("0")},
			{AccountCode: "4000", Debit: Amount("0"), Credit: Amount(amount), DepartmentCode: "D", ProjectCode: "P"},
		}}
		r := f.run(Command{Resource: "journals", Action: "create", Journal: &journal})
		f.post(f.journal(r.ID))
	}
	receipt("RC1", "1100", "250.75", &JournalDetails{BankLines: []SubledgerBankLine{{LineNumber: 1, BankAccountCode: "KBANK-01", Direction: 1}}})
	receipt("RC2", "1110", "75.25", nil)
	if rows = f.bankBooks("KBANK-01"); len(rows) != 1 || !bankBookDecimalEquals(rows[0]["gl_balance"], "250.75") {
		t.Fatalf("balance after posting: %v", rows)
	}
	if rows = f.bankBooks("KBANK-02"); len(rows) != 1 || !bankBookDecimalEquals(rows[0]["gl_balance"], "75.25") {
		t.Fatalf("second book balance: %v", rows)
	}

	// ค้นหาด้วยเลขที่บัญชี/ชื่อบัญชี
	if len(f.bankBooks("999-9-99999")) != 1 || len(f.bankBooks("รุ่งเรือง")) != 2 || len(f.bankBooks("ไม่มีแน่นอน")) != 0 {
		t.Fatal("search by account number / name failed")
	}

	// เปลี่ยนเลขที่บัญชี: เปลี่ยนเฉพาะเลขที่ รหัสสมุดและบัญชีผังบัญชีเดิม ยอดและใบสำคัญไม่ถูกแตะ
	stale := f.run(Command{Resource: BankAccountKind, Action: ChangeAccountNumberAction, ID: "KBANK-01", Version: 1, NewAccountNumber: "888-8-88888-8"})
	if stale.Version != 2 {
		t.Fatalf("change number result %+v", stale)
	}
	rows = f.bankBooks("KBANK-01")
	if rows[0]["account_number"] != "888-8-88888-8" || rows[0]["gl_account_code"] != "1100" || !bankBookDecimalEquals(rows[0]["gl_balance"], "250.75") || rows[0]["bank_account_code"] != "KBANK-01" {
		t.Fatalf("after change number: %v", rows)
	}
	f.bankBookError(Command{Resource: BankAccountKind, Action: ChangeAccountNumberAction, ID: "KBANK-01", Version: 1, NewAccountNumber: "000"}, CodeStaleVersion)
	f.bankBookError(Command{Resource: BankAccountKind, Action: ChangeAccountNumberAction, ID: "KBANK-01", Version: 2, NewAccountNumber: "999-9-99999-9"}, "bank_book_account_number_duplicate")
	f.bankBookError(Command{Resource: BankAccountKind, Action: ChangeAccountNumberAction, ID: "KBANK-01", Version: 2, NewAccountNumber: "888-8-88888-8"}, "bank_book_number_unchanged")

	// แก้ไขตาม Champ CASHFrmBankBookNew.cpp: ชื่อบัญชี/ปิดใช้งาน/บัญชี GL แก้ได้แม้สมุดมีรายการแล้ว (:164-184,238,247); เลขที่บัญชีของสมุดที่ถูกอ้างแก้ตรงไม่ได้ (:276-281)
	edit := bankBookInput("KBANK-01", "888-8-88888-8", "1100")
	edit.AccountName = "บริษัท รุ่งเรืองการค้า จำกัด (สำนักงานใหญ่)"
	if updated := f.run(Command{Resource: BankAccountKind, Action: "update", ID: "KBANK-01", Version: 2, BankAccount: &edit}); updated.Version != 3 {
		t.Fatalf("update %+v", updated)
	}
	renumber := edit
	renumber.AccountNumber = "111"
	f.bankBookError(Command{Resource: BankAccountKind, Action: "update", ID: "KBANK-01", Version: 3, BankAccount: &renumber}, "bank_book_number_use_change")
	move := edit
	move.GLAccountCode = "1130"
	if moved := f.run(Command{Resource: BankAccountKind, Action: "update", ID: "KBANK-01", Version: 3, BankAccount: &move}); moved.Version != 4 {
		t.Fatalf("move used book to another GL account %+v", moved)
	}
	f.bankBookError(Command{Resource: BankAccountKind, Action: "update", ID: "KBANK-01", Version: 2, BankAccount: &edit}, CodeStaleVersion)
	f.deny(f.scope, Command{Resource: BankAccountKind, Action: "update", ID: "NOPE", Version: 1, BankAccount: ptr(bankBookInput("NOPE", "1", "1100"))})

	// ลบ: สมุดที่ถูกใช้แล้ว (KBANK-01) หรือยังมียอด (KBANK-02, Champ MSG_891) ลบไม่ได้; สมุดว่างลบได้และแถวหายจริง
	f.bankBookError(Command{Resource: BankAccountKind, Action: "delete", ID: "KBANK-01", Version: 4}, "bank_book_in_use")
	f.bankBookError(Command{Resource: BankAccountKind, Action: "delete", ID: "KBANK-02", Version: 1}, "bank_book_balance_not_zero")
	if got := len(f.bankBooks("")); got != 2 {
		t.Fatalf("refused delete changed rows: %d", got)
	}
	// สมุดที่ยังไม่ถูกใช้แก้เลขที่บัญชีตรงในหน้าแก้ได้ (เลขซ้ำสมุดอื่นยังห้าม) ย้ายรหัสผังบัญชีได้ แล้วลบได้
	empty := bankBookInput("KBANK-03", "333-3-33333-3", "1120")
	f.run(Command{Resource: BankAccountKind, Action: "create", BankAccount: &empty})
	taken := empty
	taken.AccountNumber = "999-9-99999-9"
	f.bankBookError(Command{Resource: BankAccountKind, Action: "update", ID: "KBANK-03", Version: 1, BankAccount: &taken}, "bank_book_account_number_duplicate")
	empty.GLAccountCode = "1130"
	empty.AccountNumber = "333-3-33333-4"
	if moved := f.run(Command{Resource: BankAccountKind, Action: "update", ID: "KBANK-03", Version: 1, BankAccount: &empty}); moved.Version != 2 {
		t.Fatalf("move unused book %+v", moved)
	}
	f.run(Command{Resource: BankAccountKind, Action: "delete", ID: "KBANK-03", Version: 2})
	if rows = f.bankBooks(""); len(rows) != 2 {
		t.Fatalf("after delete: %v", rows)
	}
	var left int
	if err := f.db.QueryRow(`SELECT count(*) FROM gl_subledger_bank_accounts WHERE company=$1 AND code='KBANK-03'`, f.scope.Company).Scan(&left); err != nil || left != 0 {
		t.Fatalf("deleted book still in table: %d (%v)", left, err)
	}
	if actions := strings.Join(f.auditActions("KBANK-03"), ","); actions != "bank_account.create,bank_account.update,bank_account.delete" {
		t.Fatalf("audit of deleted book: %s", actions)
	}
	if actions := strings.Join(f.auditActions("KBANK-01"), ","); actions != "bank_account.create,bank_account.change_account_number,bank_account.update,bank_account.update" {
		t.Fatalf("audit of kept book: %s", actions)
	}
}

func TestPostgresBankBookKeepsAnInactiveBankItAlreadyUses(t *testing.T) {
	f := bankBookFixture(t)
	book := bankBookInput("KBANK-01", "123", "1100")
	f.run(Command{Resource: BankAccountKind, Action: "create", BankAccount: &book})
	if _, err := f.db.Exec(`UPDATE bank_codes SET is_active=false WHERE company=$1 AND kind='bank' AND code='KBANK'`, f.scope.Company); err != nil {
		t.Fatal(err)
	}
	// สมุดเดิมยังแก้ชื่อต่อได้ แต่สมุดใหม่เลือกธนาคารที่ปิดใช้งานไม่ได้
	book.AccountName = "ชื่อใหม่"
	f.run(Command{Resource: BankAccountKind, Action: "update", ID: "KBANK-01", Version: 1, BankAccount: &book})
	fresh := bankBookInput("KBANK-02", "456", "1110")
	f.bankBookError(Command{Resource: BankAccountKind, Action: "create", BankAccount: &fresh}, "bank_book_ref_inactive")
}

// รายละเอียดบัญชีธนาคารในใบสำคัญต้องนับสมุดที่ทะเบียนเช็คอ้างอยู่ว่า "มีรายการแล้ว" เหมือนจอสมุดเงินฝาก (bankBookUsed)
// — เปลี่ยนเลขที่บัญชีผ่านใบสำคัญไม่ได้ ไม่เช่นนั้นเลขบนเช็คที่ออกไปแล้วจะไม่ตรงกับสมุด
func TestPostgresJournalBankDetailsKeepNumberOfBookUsedByCheques(t *testing.T) {
	f := bankBookFixture(t)
	book := bankBookInput("KBANK-01", "123-4-56789-0", "1100")
	f.run(Command{Resource: BankAccountKind, Action: "create", BankAccount: &book})
	if _, err := f.db.Exec(`INSERT INTO cheque_registers (company, id, cheque_type, doc_no, doc_date, due_date, partner_code, cheque_no, bank_code, bank_branch_code,
		book_code, amount, home_amount, created_by, updated_by) VALUES ($1, $2, 1, 'CHQOUT6901-0001', '2026-01-05', '2026-01-20', 'AP-0007', '50001234', 'KBANK', '0123',
		'KBANK-01', 12500.00, 12500.00, 'tester', 'tester')`, f.scope.Company, uuid.NewString()); err != nil {
		t.Fatal(err)
	}
	details := book
	details.BankName, details.AccountNumber, details.Version = "ธนาคารกสิกรไทย", "999-9-99999-9", 1
	journal := Journal{DocNo: "JV-CQ1", Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "โอนเงินระหว่างบัญชีธนาคาร", Kind: "manual",
		Details: &JournalDetails{BankAccounts: []SubledgerBankAccount{details}}, Lines: []Line{
			{AccountCode: "1100", Debit: Amount("500.00"), Credit: Amount("0")},
			{AccountCode: "1110", Debit: Amount("0"), Credit: Amount("500.00")},
		}}
	_, err := f.execute(f.scope, Command{Resource: "journals", Action: "create", Journal: &journal})
	if err == nil || !strings.Contains(err.Error(), "ห้ามเปลี่ยนเลขบัญชีหรือสกุลเงิน") {
		t.Fatalf("journal changing the number of a book used by cheques: %v", err)
	}
	if rows := f.bankBooks("KBANK-01"); len(rows) != 1 || rows[0]["account_number"] != "123-4-56789-0" {
		t.Fatalf("book after the refused journal: %v", rows)
	}
}

// กลุ่มกิจการที่ยังไม่เคยเปิดโมดูลเช็ค: สมุดที่เกิดจากรายละเอียดในใบสำคัญต้องแก้ผ่านใบสำคัญได้ (ตรวจการใช้ที่ทะเบียนเช็คต้องไม่ error เพราะไม่มีตาราง)
func TestPostgresJournalBankDetailsBeforeChequeTablesExist(t *testing.T) {
	f := newPGIntegrityFixture(t)
	bank := SubledgerBankAccount{Code: "BANK", BankName: "ธนาคารกสิกรไทย", AccountNumber: "123-4-56789-0", AccountName: "บริษัท รุ่งเรืองการค้า จำกัด", GLAccountCode: "1000", IsActive: true}
	journal := func(docNo string, b SubledgerBankAccount) error {
		j := Journal{DocNo: docNo, Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "รับเงินโอนจากลูกค้า", Kind: "manual",
			Details: &JournalDetails{BankAccounts: []SubledgerBankAccount{b}}, Lines: []Line{
				{AccountCode: "1000", Debit: Amount("500.00"), Credit: Amount("0")},
				{AccountCode: "3000", Debit: Amount("0"), Credit: Amount("500.00")},
			}}
		_, err := f.execute(f.scope, Command{Resource: "journals", Action: "create", Journal: &j})
		return err
	}
	if err := journal("JV-B1", bank); err != nil {
		t.Fatal(err)
	}
	bank.AccountNumber, bank.Version = "999-9-99999-9", 1
	if err := journal("JV-B2", bank); err != nil {
		t.Fatalf("change details of an unused book: %v", err)
	}
}

func TestPostgresBankBookScopedPerCompany(t *testing.T) {
	f := bankBookFixture(t)
	book := bankBookInput("KBANK-01", "123", "1100")
	f.run(Command{Resource: BankAccountKind, Action: "create", BankAccount: &book})
	other := f.scope
	other.Company = "C2"
	if page, err := f.store.pg.SubledgerList(f.ctx, other, BankAccountKind, "", 1, 50, ""); err != nil || len(page.Items) != 0 {
		t.Fatalf("another company sees %d books (err %v)", len(page.Items), err)
	}
	f.deny(other, Command{Resource: BankAccountKind, Action: "delete", ID: "KBANK-01", Version: 1})
}
