//go:build integration

package glposting_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/cheque"
	"smlcloudplatform/internal/cheque/glposting"
	gl "smlcloudplatform/internal/generalledger"
)

// chequeIT คือฐานข้อมูลทดสอบใน schema สุ่มบน BC_GL_TEST_POSTGRES_DSN (ลบทิ้งเมื่อจบ) + บัญชี/สมุด/ผู้ติดต่อของบริษัทรุ่งเรือง
type chequeIT struct {
	t       *testing.T
	ctx     context.Context
	db      *sql.DB
	store   *gl.PostgresStore
	setup   gl.Scope // คำสั่งข้อมูลหลักระดับบริษัท
	scope   cheque.Scope
	nonce   int64
	counter int
}

func newChequeIT(t *testing.T) *chequeIT {
	t.Helper()
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN to isolated PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("bc_cheque_it_%d", time.Now().UnixNano())
	if _, err = admin.Exec(`CREATE SCHEMA ` + pq.QuoteIdentifier(schema)); err != nil {
		admin.Close()
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(`DROP SCHEMA ` + pq.QuoteIdentifier(schema) + ` CASCADE`); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
		admin.Close()
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	connect := func(string) (*sql.DB, error) { return db, nil }
	f := &chequeIT{
		t: t, ctx: context.Background(), db: db, store: gl.NewPostgresStore(gl.NewPostgres(connect)),
		setup: gl.Scope{Holding: "RUNGRUENG_CQ", Company: "01", Actor: "cq-it"},
		scope: cheque.Scope{Holding: "RUNGRUENG_CQ", Company: "01", Branch: "00000", Actor: "cq-it"},
		nonce: time.Now().UnixNano(),
	}
	if err = gl.EnsureSchema(f.ctx, db); err != nil {
		t.Fatal(err)
	}
	if err = cheque.EnsureSchema(f.ctx, db); err != nil {
		t.Fatal(err)
	}
	f.seedMasters()
	return f
}

func (f *chequeIT) rid() string {
	f.counter++
	return fmt.Sprintf("cq-it-%03d-%d", f.counter, f.nonce)
}

func (f *chequeIT) run(cmd gl.Command) gl.Result {
	f.t.Helper()
	cmd.RequestID = f.rid()
	r, err := f.store.Execute(f.ctx, f.setup, cmd)
	if err != nil {
		f.t.Fatalf("GL %s %s: %v", cmd.Resource, cmd.Action, err)
	}
	return r
}

func (f *chequeIT) glFail(cmd gl.Command, code string) {
	f.t.Helper()
	cmd.RequestID = f.rid()
	_, err := f.store.Execute(f.ctx, f.setup, cmd)
	if user, ok := gl.AsUserError(err); !ok || user.Code != code {
		f.t.Fatalf("GL %s %s: want %s, got %v", cmd.Resource, cmd.Action, code, err)
	}
}

func (f *chequeIT) exec(query string, args ...any) {
	f.t.Helper()
	if _, err := f.db.ExecContext(f.ctx, query, args...); err != nil {
		f.t.Fatalf("%s: %v", query, err)
	}
}

func (f *chequeIT) count(query string, args ...any) int {
	f.t.Helper()
	var n int
	if err := f.db.QueryRowContext(f.ctx, query, args...).Scan(&n); err != nil {
		f.t.Fatalf("%s: %v", query, err)
	}
	return n
}

// seedMasters: ผังบัญชีหลายระดับ, ปีบัญชี 2569, สมุดรายวันเช็ค, ธนาคาร/สาขา, สมุดเงินฝาก 2 เล่ม, ลูกค้า/ผู้จำหน่าย
func (f *chequeIT) seedMasters() {
	for _, a := range []struct {
		code, parent, name, kind, normal string
		level                            int
		posting                          bool
	}{
		{"10000", "", "สินทรัพย์", "asset", "debit", 1, false},
		{"11000", "10000", "สินทรัพย์หมุนเวียน", "asset", "debit", 2, false},
		{"11100", "11000", "เงินสดและรายการเทียบเท่าเงินสด", "asset", "debit", 3, false},
		{"11120", "11100", "เงินฝากกระแสรายวัน-ธนาคารไทยพาณิชย์", "asset", "debit", 4, true},
		{"11130", "11100", "เงินฝากออมทรัพย์-ธนาคารกสิกรไทย", "asset", "debit", 4, true},
		{"11300", "11000", "เช็คและตั๋วเงินรับ", "asset", "debit", 3, false},
		{"11310", "11300", "เช็ครับลงวันที่ล่วงหน้า", "asset", "debit", 4, true},
		{"11320", "11300", "เช็ครับคืน", "asset", "debit", 4, true},
		{"11400", "11000", "ลูกหนี้การค้า", "asset", "debit", 3, false},
		{"11410", "11400", "ลูกหนี้การค้า-ในประเทศ", "asset", "debit", 4, true},
		{"20000", "", "หนี้สิน", "liability", "credit", 1, false},
		{"21000", "20000", "หนี้สินหมุนเวียน", "liability", "credit", 2, false},
		{"21300", "21000", "เช็คจ่ายและตั๋วเงินจ่าย", "liability", "credit", 3, false},
		{"21310", "21300", "เช็คจ่ายลงวันที่ล่วงหน้า", "liability", "credit", 4, true},
		{"30000", "", "ส่วนของเจ้าของ", "equity", "credit", 1, false},
		{"31000", "30000", "กำไรสะสม", "equity", "credit", 2, true},
		{"32000", "30000", "กำไร(ขาดทุน)สุทธิประจำงวด", "equity", "credit", 2, true},
		{"50000", "", "ค่าใช้จ่าย", "expense", "debit", 1, false},
		{"53000", "50000", "ค่าใช้จ่ายในการบริหาร", "expense", "debit", 2, false},
		{"53100", "53000", "ค่าธรรมเนียมธนาคาร", "expense", "debit", 3, false},
		{"53110", "53100", "ค่าธรรมเนียมเช็คและโอนเงิน", "expense", "debit", 4, true},
	} {
		f.run(gl.Command{Resource: "accounts", Action: "create", Account: &gl.Account{
			AccountCode: a.code, ParentAccountCode: a.parent, Names: []gl.Name{{Code: "th", Name: a.name}},
			AccountType: a.kind, NormalBalance: a.normal, Level: a.level, AllowPosting: a.posting, IsActive: true,
		}})
	}
	f.run(gl.Command{Resource: "fiscal-years", Action: "create", FiscalYear: &gl.FiscalYear{
		Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", Scale: 2, IsActive: true,
		ProfitLossAccount: "32000", RetainedEarningsAccount: "31000",
	}})
	f.run(gl.Command{Resource: "journal-books", Action: "create", Master: &gl.Master{
		Kind: "journal-books", Code: "CQ", Name: "สมุดรายวันเช็คและธนาคาร", BookType: gl.BookTypeGeneral, IsActive: true,
	}})
	for _, row := range []struct{ kind, code, name, bank string }{
		{"bank", "KBANK", "ธนาคารกสิกรไทย", ""},
		{"bank", "SCB", "ธนาคารไทยพาณิชย์", ""},
		{"bank_branch", "0123", "สาขาสีลม", "KBANK"},
		{"bank_branch", "0456", "สาขาบางนา", "SCB"},
		{"bank_account_type", "CURRENT", "กระแสรายวัน", ""},
	} {
		var bank any
		if row.bank != "" {
			bank = row.bank
		}
		f.exec(`INSERT INTO bank_codes(company,kind,code,name,bank_code,created_by,updated_by) VALUES('01',$1,$2,$3,$4,'cq-it','cq-it')`, row.kind, row.code, row.name, bank)
	}
	for _, b := range []struct{ code, number, bank, branch, account string }{
		{"SCB-01", "402-3-45678-9", "SCB", "0456", "11120"},
		{"KBANK-02", "123-4-56789-0", "KBANK", "0123", "11130"},
	} {
		book := gl.SubledgerBankAccount{Code: b.code, AccountNumber: b.number, AccountName: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด",
			BankCode: b.bank, BankBranchCode: b.branch, AccountTypeCode: "CURRENT", GLAccountCode: b.account, Currency: "THB", IsActive: true}
		f.run(gl.Command{Resource: gl.BankAccountKind, Action: "create", BankAccount: &book})
	}
	f.exec(`INSERT INTO gl_subledger_partners(company,code,version,payload) VALUES
		('01','AR-0012',1,'{"partner_code":"AR-0012","name_th":"บริษัท ก่อสร้างไทยรุ่งเรือง จำกัด","is_customer":true,"is_active":true}'),
		('01','AP-0007',1,'{"partner_code":"AP-0007","name_th":"บริษัท เหล็กสยามค้าส่ง จำกัด","is_supplier":true,"is_active":true}')`)
}

func (f *chequeIT) mapping(code, name, sourceDoc string, rules ...gl.MappingRule) {
	f.t.Helper()
	f.run(gl.Command{Resource: "mappings", Action: "create", Master: &gl.Master{
		Kind: "mappings", Code: code, Name: name, SourceDoc: sourceDoc, BookCode: "CQ", IsActive: true, Rules: rules,
	}})
}

func rule(account, from, side, source string) gl.MappingRule {
	return gl.MappingRule{AccountCode: account, AccountFrom: from, Side: side, Source: source}
}

func (f *chequeIT) receive(no, amount, docDate, dueDate string) cheque.Register {
	f.t.Helper()
	r, err := cheque.CreateRegister(f.ctx, f.db, f.scope, f.rid(), cheque.RegisterInput{
		ChequeType: cheque.TypeReceived, DocDate: docDate, DueDate: dueDate, PartnerCode: "AR-0012", ChequeNo: no,
		BankCode: "KBANK", BankBranchCode: "0123", Amount: amount, Description: "รับชำระค่าเหล็กเส้นตามใบแจ้งหนี้ IV6909-0042",
	})
	if err != nil {
		f.t.Fatalf("receive cheque %s: %v", no, err)
	}
	return r
}

func (f *chequeIT) doc(flag int, date, book string, cancelType *int, lines ...cheque.DocLineInput) (cheque.Doc, error) {
	return cheque.CreateDoc(f.ctx, f.db, glposting.Ledger{}, f.scope, f.rid(), cheque.DocInput{
		Flag: flag, DocDate: date, BookCode: book, CancelType: cancelType, Lines: lines,
		Description: "ปรับสถานะเช็คตามใบแจ้งยอดธนาคาร",
	})
}

func (f *chequeIT) mustDoc(flag int, date, book string, cancelType *int, lines ...cheque.DocLineInput) cheque.Doc {
	f.t.Helper()
	d, err := f.doc(flag, date, book, cancelType, lines...)
	if err != nil {
		f.t.Fatalf("doc %d on %s: %v", flag, date, err)
	}
	return d
}

func (f *chequeIT) cancel(d cheque.Doc, reason string) (cheque.Doc, error) {
	return cheque.CancelDoc(f.ctx, f.db, glposting.Ledger{}, f.scope, f.rid(), d.Transflag, d.DocNo, d.Version, reason)
}

func (f *chequeIT) state(id string) (int, string) {
	f.t.Helper()
	var (
		status int
		book   string
	)
	if err := f.db.QueryRowContext(f.ctx, `SELECT status, COALESCE(book_code,'') FROM cheque_registers WHERE company='01' AND id=$1`, id).Scan(&status, &book); err != nil {
		f.t.Fatal(err)
	}
	return status, book
}

func (f *chequeIT) wantState(id string, status int, book string) {
	f.t.Helper()
	if s, b := f.state(id); s != status || b != book {
		f.t.Fatalf("cheque %s = status %d book %q, want %d %q", id, s, b, status, book)
	}
}

type glLine struct {
	account       string
	debit, credit decimal.Decimal
}

// activeJournal คืนบรรทัดของใบสำคัญที่ใช้งานอยู่ของเอกสาร (ตรวจจาก gl_lines ที่ผ่านรายการจริง)
func (f *chequeIT) activeJournal(d cheque.Doc) (string, []glLine) {
	f.t.Helper()
	var id string
	if err := f.db.QueryRowContext(f.ctx, `SELECT journal_id FROM cheque_doc_journals WHERE company='01' AND transflag=$1 AND doc_no=$2 AND voided_at IS NULL`,
		d.Transflag, d.DocNo).Scan(&id); err != nil {
		f.t.Fatalf("journal of %s: %v", d.DocNo, err)
	}
	rows, err := f.db.QueryContext(f.ctx, `SELECT account_code, debit::text, credit::text FROM gl_lines WHERE company='01' AND journal_id=$1 ORDER BY line_no`, id)
	if err != nil {
		f.t.Fatal(err)
	}
	defer rows.Close()
	var lines []glLine
	for rows.Next() {
		var (
			l             glLine
			debit, credit string
		)
		if err = rows.Scan(&l.account, &debit, &credit); err != nil {
			f.t.Fatal(err)
		}
		l.debit, l.credit = decimal.RequireFromString(debit), decimal.RequireFromString(credit)
		lines = append(lines, l)
	}
	return id, lines
}

func (f *chequeIT) wantJournal(d cheque.Doc, want ...glLine) string {
	f.t.Helper()
	id, got := f.activeJournal(d)
	if len(got) != len(want) {
		f.t.Fatalf("journal of %s = %+v, want %+v", d.DocNo, got, want)
	}
	for i := range want {
		if got[i].account != want[i].account || !got[i].debit.Equal(want[i].debit) || !got[i].credit.Equal(want[i].credit) {
			f.t.Fatalf("journal of %s line %d = %+v, want %+v", d.DocNo, i+1, got[i], want[i])
		}
	}
	return id
}

func dr(account, amount string) glLine {
	return glLine{account: account, debit: decimal.RequireFromString(amount), credit: decimal.Zero}
}

func cr(account, amount string) glLine {
	return glLine{account: account, debit: decimal.Zero, credit: decimal.RequireFromString(amount)}
}

func wantKey(t *testing.T, err error, key string) *cheque.Error {
	t.Helper()
	e, ok := cheque.AsError(err)
	if !ok || e.Key != key {
		t.Fatalf("want %s, got %v", key, err)
	}
	return e
}

func intRef(v int) *int { return &v }

// TestChequeLifecycleOnRealLedger เดินวงจรเช็ครับ/จ่ายตาม Champ บนเครื่องบัญชีจริง และตรวจ PostgreSQL หลังทุกขั้น:
// ทะเบียน (D8 ยืนยันก่อนแก้) → นำฝาก → ผ่าน+ค่าธรรมเนียม (เดบิต = เครดิต, บัญชีจากสมุด) → คืน → นำเข้าใหม่ → ยกเลิกชนิด 3
// → ยกเลิกย้อนทีละใบ (กันเอกสารที่ใหม่กว่า) → แก้เอกสาร (รอบใหม่ + กลับรายการ) → ไม่มีรูปแบบลงบัญชี → งวดปิด → คำนวณยอดใหม่
func TestChequeLifecycleOnRealLedger(t *testing.T) {
	f := newChequeIT(t)
	const (
		debit  = gl.MappingSideDebit
		credit = gl.MappingSideCredit
		book   = gl.MappingAccountFromBankBook
	)

	// รูปแบบลงบัญชี: แหล่งยอดตระกูลเช็คเท่านั้น และบัญชีจากสมุดเงินฝากใช้ได้เฉพาะเอกสารเช็ค
	f.glFail(gl.Command{Resource: "mappings", Action: "create", Master: &gl.Master{Kind: "mappings", Code: "CQ-BAD1", Name: "ผิดแหล่งยอด",
		SourceDoc: "chequediscount", BookCode: "CQ", IsActive: true,
		Rules: []gl.MappingRule{rule("11120", "", debit, "total"), rule("11310", "", credit, "cheque_total")}}}, "mapping_rule_cheque_source_invalid")
	f.glFail(gl.Command{Resource: "mappings", Action: "create", Master: &gl.Master{Kind: "mappings", Code: "CQ-BAD2", Name: "สมุดในเอกสารขาย",
		SourceDoc: "sale-invoice", BookCode: "CQ", IsActive: true,
		Rules: []gl.MappingRule{rule("11410", book, debit, "total"), rule("11310", "", credit, "total")}}}, "mapping_rule_account_from_invalid")
	f.mapping("CQ-420", "ลงบัญชีเช็ครับผ่าน", "chequereceivedclear",
		rule("", book, debit, "cheque_net"), rule("53110", "", debit, "cheque_fee"), rule("11310", "", credit, "cheque_total"))
	f.mapping("CQ-421", "ลงบัญชีเช็ครับคืน", "chequereceivedreturn",
		rule("11320", "", debit, "cheque_total"), rule("", book, credit, "cheque_total"))
	f.mapping("CQ-422", "ลงบัญชีนำเช็คเข้าใหม่", "chequeredeposit",
		rule("", book, debit, "cheque_net"), rule("53110", "", debit, "cheque_fee"), rule("11320", "", credit, "cheque_total"))
	f.mapping("CQ-423", "ลงบัญชียกเลิกเช็ครับ", "chequereceivedcancel",
		rule("11410", "", debit, "cheque_total"), rule("11310", book, credit, "cheque_total"))
	f.mapping("CQ-425", "ลงบัญชีเช็คจ่ายผ่าน", "chequeissuedclear",
		rule("21310", "", debit, "cheque_total"), rule("", book, credit, "cheque_total"))

	// 1. ทะเบียน: เช็ครับ 2 ใบ (ในมือ ไม่มีสมุด), เช็คจ่าย 1 ใบ (สมุด SCB-01)
	a := f.receive("10234567", "15250.75", "2026-10-01", "2026-10-15")
	b := f.receive("10234568", "8400.00", "2026-10-02", "2026-10-05")
	c, err := cheque.CreateRegister(f.ctx, f.db, f.scope, f.rid(), cheque.RegisterInput{
		ChequeType: cheque.TypeIssued, DocDate: "2026-10-02", DueDate: "2026-10-12", PartnerCode: "AP-0007", ChequeNo: "50001234",
		BankCode: "SCB", BankBranchCode: "0456", BookCode: "SCB-01", Amount: "32000.00", Description: "จ่ายค่าเหล็กแผ่นตามใบสั่งซื้อ PO6909-0118",
	})
	if err != nil {
		t.Fatal(err)
	}
	if n := f.count(`SELECT count(*) FROM cheque_registers WHERE company='01' AND status=0 AND home_amount=amount`); n != 3 {
		t.Fatalf("registers stored = %d, want 3", n)
	}
	f.wantState(a.ID, cheque.InOnHand, "")
	f.wantState(c.ID, cheque.OutOnHand, "SCB-01")
	_, err = f.doc(cheque.FlagReceivedClear, "2026-10-14", "SCB-01", nil, cheque.DocLineInput{ChequeID: a.ID})
	wantKey(t, err, "cheque_err_not_due")

	// 2. นำฝาก 419: ไม่ลงบัญชี, สมุดของเอกสารไปอยู่ในเช็ค, เลขที่ออกตาม prefix
	deposit := f.mustDoc(cheque.FlagDeposit, "2026-10-03", "SCB-01", nil, cheque.DocLineInput{ChequeID: a.ID}, cheque.DocLineInput{ChequeID: b.ID})
	if deposit.DocNo != "CQDP6910-0001" || deposit.GLStatus != cheque.GLStatusNotRequired || deposit.TotalAmount != "23650.75" {
		t.Fatalf("deposit = %+v", deposit)
	}
	f.wantState(a.ID, cheque.InDeposited, "SCB-01")
	if n := f.count(`SELECT count(*) FROM gl_records WHERE company='01' AND kind='journals'`); n != 0 {
		t.Fatalf("deposit must not post, journals = %d", n)
	}

	// 3. D8: แก้ฟิลด์หลักของเช็คที่ถูกอ้างแล้วต้องยืนยันก่อน; ยืนยันแล้วบันทึก event edit
	reg := cheque.RegisterInput{ID: a.ID, Version: a.Version + 1, ChequeType: cheque.TypeReceived, DocDate: "2026-10-01", DueDate: "2026-10-15",
		PartnerCode: "AR-0012", ChequeNo: "10234567", BankCode: "KBANK", BankBranchCode: "0123", BookCode: "SCB-01", Amount: "15250.75",
		Description: "รับชำระค่าเหล็กเส้นตามใบแจ้งหนี้ IV6909-0042 (แก้เลขเช็ค)"}
	reg.ChequeNo = "10234577"
	_, err = cheque.UpdateRegister(f.ctx, f.db, f.scope, f.rid(), reg)
	if e := wantKey(t, err, "cheque_err_referenced"); e.Args[0] != deposit.DocNo {
		t.Fatalf("referenced args = %v", e.Args)
	}
	reg.Confirm = true
	edited, err := cheque.UpdateRegister(f.ctx, f.db, f.scope, f.rid(), reg)
	if err != nil || edited.ChequeNo != "10234577" {
		t.Fatalf("confirmed edit = %+v, %v", edited, err)
	}
	if n := f.count(`SELECT count(*) FROM cheque_events WHERE company='01' AND cheque_id=$1 AND event_kind='edit'`, a.ID); n != 1 {
		t.Fatalf("edit events = %d", n)
	}
	if _, err = cheque.CancelRegister(f.ctx, f.db, f.scope, f.rid(), cheque.TypeReceived, a.ID, edited.Version); err == nil {
		t.Fatal("a cheque used by a document must not be deleted")
	} else {
		wantKey(t, err, "cheque_err_referenced_delete")
	}
	// ประเภทรับ/จ่ายแก้ไม่ได้แม้ยืนยัน: Champ แยกตาราง BCChqIn/BCChqOut (SQLSERVER_Script.sql:7482, :7536)
	asIssued := reg
	asIssued.ChequeType, asIssued.Version, asIssued.PartnerCode = cheque.TypeIssued, edited.Version, "AP-0007"
	_, err = cheque.UpdateRegister(f.ctx, f.db, f.scope, f.rid(), asIssued)
	wantKey(t, err, "cheque_err_not_found")
	f.wantState(a.ID, cheque.InDeposited, "SCB-01")

	// 4. เช็ครับผ่าน 420 + ค่าธรรมเนียม: Dr ธนาคาร (บัญชีของสมุด) ยอดสุทธิ + Dr ค่าธรรมเนียม / Cr เช็ครับ ยอดเช็ค
	clearRID := f.rid()
	clearIn := cheque.DocInput{Flag: cheque.FlagReceivedClear, DocDate: "2026-10-15", BookCode: "SCB-01",
		Lines:  []cheque.DocLineInput{{ChequeID: a.ID, FeeAmount: "25.00"}, {ChequeID: b.ID}},
		Totals: &cheque.DocTotals{Total: "23650.75", Fee: "25.00", Net: "23625.75"}}
	clear, err := cheque.CreateDoc(f.ctx, f.db, glposting.Ledger{}, f.scope, clearRID, clearIn)
	if err != nil {
		t.Fatal(err)
	}
	if clear.GLStatus != cheque.GLStatusPosted || clear.JournalDocNo == "" {
		t.Fatalf("clear = %+v", clear)
	}
	clearJournal := f.wantJournal(clear, dr("11120", "23625.75"), dr("53110", "25.00"), cr("11310", "23650.75"))
	f.wantState(a.ID, cheque.InCleared, "SCB-01")
	f.wantState(b.ID, cheque.InCleared, "SCB-01")
	replay, err := cheque.CreateDoc(f.ctx, f.db, glposting.Ledger{}, f.scope, clearRID, clearIn)
	if err != nil || replay.DocNo != clear.DocNo || f.count(`SELECT count(*) FROM cheque_docs WHERE company='01' AND transflag=420`) != 1 {
		t.Fatalf("replayed request must return the first result without a new document: %+v, %v", replay, err)
	}

	// 5. เช็ครับคืน 421 (สมุดจากหัวเอกสาร), แล้วยกเลิกเอกสารผ่านไม่ได้เพราะมีเอกสารที่ใหม่กว่า
	ret := f.mustDoc(cheque.FlagReceivedReturn, "2026-10-18", "SCB-01", nil, cheque.DocLineInput{ChequeID: a.ID})
	f.wantJournal(ret, dr("11320", "15250.75"), cr("11120", "15250.75"))
	f.wantState(a.ID, cheque.InReturned, "SCB-01")
	if _, err = f.cancel(clear, "ธนาคารแจ้งยอดผิด"); err == nil {
		t.Fatal("cancel with a newer document must fail")
	} else if e := wantKey(t, err, "cheque_err_newer_event"); e.Args[1] != ret.DocNo {
		t.Fatalf("newer event args = %v", e.Args)
	}

	// 6. นำเช็คเข้าใหม่ 422: ต้องเป็นสมุดเดิมของเช็ค
	_, err = f.doc(cheque.FlagRedeposit, "2026-10-20", "KBANK-02", nil, cheque.DocLineInput{ChequeID: a.ID})
	wantKey(t, err, "cheque_err_book_mismatch")
	redeposit := f.mustDoc(cheque.FlagRedeposit, "2026-10-20", "SCB-01", nil, cheque.DocLineInput{ChequeID: a.ID, FeeAmount: "10.00"})
	f.wantJournal(redeposit, dr("11120", "15240.75"), dr("53110", "10.00"), cr("11320", "15250.75"))
	f.wantState(a.ID, cheque.InRedeposited, "SCB-01")

	// 7. ยกเลิกเช็ครับชนิด 3 (นำเข้าใหม่แล้ว): Cr บัญชีของสมุด (ไม่ใช้บัญชีสำรอง)
	cancelled := f.mustDoc(cheque.FlagReceivedCancel, "2026-10-25", "SCB-01", intRef(cheque.CancelRedeposited), cheque.DocLineInput{ChequeID: a.ID})
	f.wantJournal(cancelled, dr("11410", "15250.75"), cr("11120", "15250.75"))
	f.wantState(a.ID, cheque.InCancelled, "SCB-01")

	// 8. ยกเลิกย้อนทีละใบ: สถานะกลับตามประวัติ ใบสำคัญถูกกลับรายการ (ใบเดิมคงอยู่)
	if _, err = f.cancel(cancelled, ""); err == nil {
		t.Fatal("cancel without reason must fail")
	} else {
		wantKey(t, err, "cheque_err_reason_required")
	}
	for _, step := range []struct {
		doc    cheque.Doc
		status int
	}{{cancelled, cheque.InRedeposited}, {redeposit, cheque.InReturned}, {ret, cheque.InCleared}, {clear, cheque.InDeposited}} {
		done, err := f.cancel(step.doc, "ธนาคารแจ้งยกเลิกรายการ")
		if err != nil {
			t.Fatalf("cancel %s: %v", step.doc.DocNo, err)
		}
		if done.Status != "cancelled" || done.CancelReason != "ธนาคารแจ้งยกเลิกรายการ" {
			t.Fatalf("cancelled doc = %+v", done)
		}
		f.wantState(a.ID, step.status, "SCB-01")
		if n := f.count(`SELECT count(*) FROM cheque_doc_journals WHERE company='01' AND transflag=$1 AND doc_no=$2 AND voided_at IS NULL`, step.doc.Transflag, step.doc.DocNo); n != 0 {
			t.Fatalf("%s still has an active journal", step.doc.DocNo)
		}
	}
	f.wantState(b.ID, cheque.InDeposited, "SCB-01")
	if status := f.count(`SELECT count(*) FROM gl_records WHERE company='01' AND kind='journals' AND id=$1 AND payload->>'status'='reversed'`, clearJournal); status != 1 {
		t.Fatal("the cleared voucher must be reversed, not deleted")
	}
	if n := f.count(`SELECT count(*) FROM gl_records WHERE company='01' AND kind='journals' AND code=$1`, clear.JournalDocNo+"-R"); n != 1 {
		t.Fatalf("reversal voucher %s-R not found", clear.JournalDocNo)
	}
	if n := f.count(`SELECT count(*) FROM (SELECT account_code FROM gl_lines WHERE company='01' GROUP BY account_code HAVING SUM(debit-credit)<>0) x`); n != 0 {
		t.Fatalf("after cancelling every document all accounts must net to zero, %d do not", n)
	}

	// 9. แก้เอกสาร: ถอนผลเดิม (กลับรายการ) แล้วลงบัญชีรอบใหม่
	again := f.mustDoc(cheque.FlagReceivedClear, "2026-10-16", "SCB-01", nil, cheque.DocLineInput{ChequeID: a.ID, FeeAmount: "15.00"})
	firstJournal := f.wantJournal(again, dr("11120", "15235.75"), dr("53110", "15.00"), cr("11310", "15250.75"))
	updated, err := cheque.UpdateDoc(f.ctx, f.db, glposting.Ledger{}, f.scope, f.rid(), cheque.DocInput{
		Flag: cheque.FlagReceivedClear, DocNo: again.DocNo, Version: again.Version, DocDate: "2026-10-16", BookCode: "SCB-01",
		Lines: []cheque.DocLineInput{{ChequeID: a.ID, FeeAmount: "20.00"}, {ChequeID: b.ID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != again.Version+1 || updated.TotalAmount != "23650.75" || updated.FeeAmount != "20.00" {
		t.Fatalf("updated = %+v", updated)
	}
	if second := f.wantJournal(updated, dr("11120", "23630.75"), dr("53110", "20.00"), cr("11310", "23650.75")); second == firstJournal {
		t.Fatal("edit must post a new voucher")
	}
	if n := f.count(`SELECT count(*) FROM cheque_doc_journals WHERE company='01' AND transflag=420 AND doc_no=$1`, again.DocNo); n != 2 {
		t.Fatalf("revisions = %d, want 2", n)
	}
	f.wantState(b.ID, cheque.InCleared, "SCB-01")

	// 10. ไม่มีรูปแบบลงบัญชี (ขายลด 424): เอกสารบันทึกได้ สถานะ no_mapping
	// เช็คในมือระบุสมุดได้ (cqfrmcheqin.cpp:391-400,488) และขายลดที่ระบุสมุดเลือกได้เฉพาะเช็คของสมุดนั้น (cqfrmcheqinsale.cpp:676-681)
	d, err := cheque.CreateRegister(f.ctx, f.db, f.scope, f.rid(), cheque.RegisterInput{
		ChequeType: cheque.TypeReceived, DocDate: "2026-10-05", DueDate: "2026-11-30", PartnerCode: "AR-0012", ChequeNo: "10234590",
		BankCode: "KBANK", BankBranchCode: "0123", BookCode: "KBANK-02", Amount: "5000.00", Description: "รับชำระค่าท่อเหล็กตามใบแจ้งหนี้ IV6910-0007",
	})
	if err != nil {
		t.Fatal(err)
	}
	f.wantState(d.ID, cheque.InOnHand, "KBANK-02")
	_, err = f.doc(cheque.FlagDiscount, "2026-10-21", "SCB-01", nil, cheque.DocLineInput{ChequeID: d.ID})
	wantKey(t, err, "cheque_err_book_mismatch")
	discount := f.mustDoc(cheque.FlagDiscount, "2026-10-21", "KBANK-02", nil, cheque.DocLineInput{ChequeID: d.ID, FeeAmount: "50.00"})
	if discount.GLStatus != cheque.GLStatusNoMapping {
		t.Fatalf("discount gl_status = %q", discount.GLStatus)
	}
	f.wantState(d.ID, cheque.InDiscounted, "KBANK-02")

	// 10.1 แก้ขายลดของเช็คที่อยู่ในมือพร้อมสมุด: ถอนผลเดิมคืนเช็คเป็นในมือและล้างสมุด (CQFrmBaseTransStatus.cpp:188-190)
	// แล้วต้องบันทึกรอบใหม่ได้ (เดิมติด cheque_err_book_mismatch เพราะสมุดของเช็คว่างแล้ว)
	discount, err = cheque.UpdateDoc(f.ctx, f.db, glposting.Ledger{}, f.scope, f.rid(), cheque.DocInput{
		Flag: cheque.FlagDiscount, DocNo: discount.DocNo, Version: discount.Version, DocDate: "2026-10-21", BookCode: "KBANK-02",
		Lines: []cheque.DocLineInput{{ChequeID: d.ID, FeeAmount: "60.00"}},
	})
	if err != nil {
		t.Fatalf("edit the discount of an on-hand cheque with a book: %v", err)
	}
	if discount.FeeAmount != "60.00" || discount.NetAmount != "4940.00" {
		t.Fatalf("edited discount = %+v", discount)
	}
	f.wantState(d.ID, cheque.InDiscounted, "KBANK-02")

	// 10.2 ขายลดไม่บังคับสมุด (cqfrmcheqinsale.cpp:374-382): เช็คในมือที่ไม่มีสมุด → สถานะ 5 ไม่มีสมุด และยังแก้ทะเบียนต่อได้
	g := f.receive("10234595", "3200.00", "2026-10-06", "2026-12-15")
	noBook := f.mustDoc(cheque.FlagDiscount, "2026-10-22", "", nil, cheque.DocLineInput{ChequeID: g.ID, FeeAmount: "20.00"})
	if noBook.BookCode != "" || noBook.NetAmount != "3180.00" {
		t.Fatalf("discount without a book = %+v", noBook)
	}
	f.wantState(g.ID, cheque.InDiscounted, "")
	gNow, err := cheque.GetRegister(f.ctx, f.db, "01", g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = cheque.UpdateRegister(f.ctx, f.db, f.scope, f.rid(), cheque.RegisterInput{ID: g.ID, Version: gNow.Version,
		ChequeType: cheque.TypeReceived, DocDate: "2026-10-06", DueDate: "2026-12-15", PartnerCode: "AR-0012", ChequeNo: "10234595",
		BankCode: "KBANK", BankBranchCode: "0123", Amount: "3200.00", Description: "รับชำระค่าข้อต่อท่อตามใบแจ้งหนี้ IV6910-0011"}); err != nil {
		t.Fatalf("edit the register of a discounted cheque without a book: %v", err)
	}
	f.wantState(g.ID, cheque.InDiscounted, "")

	// 11. เช็คจ่ายผ่าน 425 ต้องระบุสมุด (cqfrmcheqoutpass.cpp:362-366) → ไม่มีอะไรค้าง
	_, err = f.doc(cheque.FlagIssuedClear, "2026-10-12", "", nil, cheque.DocLineInput{ChequeID: c.ID})
	wantKey(t, err, "cheque_err_book_required")
	f.wantState(c.ID, cheque.OutOnHand, "SCB-01")
	if n := f.count(`SELECT count(*) FROM cheque_docs WHERE company='01' AND transflag=425`); n != 0 {
		t.Fatalf("refused document left %d rows", n)
	}

	// 12. งวดปิด: บันทึกไม่ได้ และไม่มีอะไรค้าง
	f.exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{locked}','true') WHERE company='01' AND kind='periods' AND code='2026-10'`)
	_, err = f.doc(cheque.FlagIssuedClear, "2026-10-12", "SCB-01", nil, cheque.DocLineInput{ChequeID: c.ID})
	if user, ok := gl.AsUserError(err); !ok || user.Code != "source_period_locked" {
		t.Fatalf("locked period: %v", err)
	}
	if _, err = f.cancel(discount, "ขายลดผิดธนาคาร"); err == nil {
		t.Fatal("cancel inside a locked period must fail")
	}
	f.wantState(c.ID, cheque.OutOnHand, "SCB-01")
	f.wantState(d.ID, cheque.InDiscounted, "KBANK-02")
	f.exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{locked}','false') WHERE company='01' AND kind='periods' AND code='2026-10'`)
	paid := f.mustDoc(cheque.FlagIssuedClear, "2026-10-12", "SCB-01", nil, cheque.DocLineInput{ChequeID: c.ID})
	f.wantJournal(paid, dr("21310", "32000.00"), cr("11120", "32000.00"))
	f.wantState(c.ID, cheque.OutCleared, "SCB-01")

	// 13. ทุกใบสำคัญเดบิต = เครดิต
	if n := f.count(`SELECT count(*) FROM (SELECT journal_id FROM gl_lines WHERE company='01' GROUP BY journal_id HAVING SUM(debit)<>SUM(credit)) x`); n != 0 {
		t.Fatalf("%d vouchers are unbalanced", n)
	}

	// 14. สมุดเงินฝากที่เช็คหรือเอกสารเช็คอ้างอยู่ลบไม่ได้
	var version int64
	if err = f.db.QueryRowContext(f.ctx, `SELECT version FROM gl_subledger_bank_accounts WHERE company='01' AND code='KBANK-02'`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	f.glFail(gl.Command{Resource: gl.BankAccountKind, Action: "delete", ID: "KBANK-02", Version: version}, "bank_book_in_use")

	// 14.1 ลบเช็คที่เอกสารสถานะถูกยกเลิกหมดแล้วได้ (Champ: ยกเลิกลบ Sub CQFrmBaseTransStatus.cpp:323-335, ลบเช็ค cqfrmcheqin.cpp:462-470)
	// BC เก็บแถว+ประวัติไว้ (deleted_at + event delete) แต่ไม่นับในทะเบียน และเลขเช็คเดิมบันทึกใหม่ได้
	e := f.receive("10234601", "7800.00", "2026-10-06", "2026-10-10")
	eDeposit := f.mustDoc(cheque.FlagDeposit, "2026-10-07", "KBANK-02", nil, cheque.DocLineInput{ChequeID: e.ID})
	if _, err = f.cancel(eDeposit, "นำฝากผิดสมุด"); err != nil {
		t.Fatal(err)
	}
	f.wantState(e.ID, cheque.InOnHand, "")
	eNow, err := cheque.GetRegister(f.ctx, f.db, "01", e.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = cheque.CancelRegister(f.ctx, f.db, f.scope, f.rid(), cheque.TypeReceived, e.ID, eNow.Version); err != nil {
		t.Fatalf("delete after every status document was cancelled: %v", err)
	}
	if n := f.count(`SELECT count(*) FROM cheque_registers WHERE company='01' AND id=$1 AND deleted_at IS NOT NULL AND deleted_by IS NOT NULL`, e.ID); n != 1 {
		t.Fatal("deleted cheque row must stay with deleted_at/deleted_by")
	}
	if n := f.count(`SELECT count(*) FROM cheque_events WHERE company='01' AND cheque_id=$1 AND (event_kind='delete' OR (event_kind='status_doc' AND voided_at IS NOT NULL) OR event_kind='create')`, e.ID); n != 3 {
		t.Fatalf("history of the deleted cheque = %d events, want create + voided deposit + delete", n)
	}
	if _, err = cheque.GetRegister(f.ctx, f.db, "01", e.ID); err == nil {
		t.Fatal("a deleted cheque must not open")
	} else {
		wantKey(t, err, "cheque_err_not_found")
	}
	if page, err := cheque.ListRegisters(f.ctx, f.db, "01", cheque.RegisterFilter{ChequeType: cheque.TypeReceived, Query: "10234601", Limit: 20}); err != nil || len(page.Items) != 0 {
		t.Fatalf("register list shows a deleted cheque: %+v, %v", page.Items, err)
	}
	if kept, err := cheque.GetDoc(f.ctx, f.db, "01", cheque.FlagDeposit, eDeposit.DocNo); err != nil || kept.Status != "cancelled" {
		t.Fatalf("the cancelled deposit must stay viewable: %+v, %v", kept, err)
	}
	f.receive("10234601", "7800.00", "2026-10-08", "2026-10-10")

	// 15. คำนวณยอดเช็คใหม่: ทะเบียนที่เพี้ยนจากประวัติถูกพบและปรับกลับพร้อมเหตุผล
	f.exec(`UPDATE cheque_registers SET status=0, book_code=NULL WHERE company='01' AND id=$1`, b.ID)
	check, err := cheque.CheckRecalculate(f.ctx, f.db, "01")
	if err != nil || len(check.Mismatches) != 1 || check.Mismatches[0].ID != b.ID || check.Mismatches[0].DerivedStatus != cheque.InCleared || check.Mismatches[0].DerivedBook != "SCB-01" {
		t.Fatalf("check = %+v, %v", check, err)
	}
	if _, err = cheque.ApplyRecalculate(f.ctx, f.db, f.scope, f.rid(), ""); err == nil {
		t.Fatal("apply without reason must fail")
	}
	if _, err = cheque.ApplyRecalculate(f.ctx, f.db, f.scope, f.rid(), "ปรับสถานะเช็คตามประวัติเอกสาร"); err != nil {
		t.Fatal(err)
	}
	f.wantState(b.ID, cheque.InCleared, "SCB-01")
	if check, err = cheque.CheckRecalculate(f.ctx, f.db, "01"); err != nil || len(check.Mismatches) != 0 {
		t.Fatalf("after apply = %+v, %v", check, err)
	}
}

// TestEnsureSchemaRelaxesReceivedBookCheck: ฐานที่สร้างตารางทะเบียนเช็คไว้ก่อน (CHECK เดิมห้ามเช็ครับสถานะ 5 ไม่มีสมุด)
// ต้องได้ CHECK ใหม่จาก EnsureSchema เอง เพราะ CREATE TABLE IF NOT EXISTS ไม่แตะตารางที่มีอยู่แล้ว
func TestEnsureSchemaRelaxesReceivedBookCheck(t *testing.T) {
	f := newChequeIT(t)
	var schema string
	if err := f.db.QueryRowContext(f.ctx, `SELECT current_schema()`).Scan(&schema); err != nil {
		t.Fatal(err)
	}
	f.exec(`ALTER TABLE cheque_registers DROP CONSTRAINT IF EXISTS cheque_registers_in_book_check`)
	f.exec(`ALTER TABLE cheque_registers ADD CHECK (cheque_type = 1 OR status IN (0,4) OR book_code IS NOT NULL)`)

	// การเชื่อมต่อใหม่ = EnsureSchema รันจริงอีกครั้ง (ไม่ติด cache ของ *sql.DB เดิม); รันซ้ำสองรอบต้องไม่ error
	u, err := url.Parse(os.Getenv("BC_GL_TEST_POSTGRES_DSN"))
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for i := 0; i < 2; i++ {
		if err = cheque.EnsureSchema(f.ctx, db); err != nil {
			t.Fatalf("ensure schema run %d: %v", i+1, err)
		}
	}

	insert := `INSERT INTO cheque_registers (company, id, cheque_type, doc_no, doc_date, due_date, partner_code, cheque_no, bank_code, bank_branch_code,
		status, amount, home_amount, created_by, updated_by)
		VALUES ('01', gen_random_uuid(), 0, $1, '2026-10-01', '2026-10-15', 'AR-0012', $2, 'KBANK', '0123', $3, 1800.00, 1800.00, 'cq-it', 'cq-it')`
	if _, err = f.db.ExecContext(f.ctx, insert, "CHQIN6910-0901", "10239901", cheque.InDiscounted); err != nil {
		t.Fatalf("discounted received cheque without a book must be allowed: %v", err)
	}
	if _, err = f.db.ExecContext(f.ctx, insert, "CHQIN6910-0902", "10239902", cheque.InDeposited); err == nil {
		t.Fatal("deposited received cheque without a book must still be refused")
	}
	if n := f.count(`SELECT count(*) FROM pg_constraint WHERE conrelid = 'cheque_registers'::regclass AND contype = 'c'
		AND pg_get_constraintdef(oid) LIKE '%book_code IS NOT NULL%' AND pg_get_constraintdef(oid) NOT LIKE '%doc_date%'`); n != 1 {
		t.Fatalf("received-cheque book checks = %d, want exactly the new one", n)
	}
}
