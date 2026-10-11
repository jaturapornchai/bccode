//go:build integration

package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
)

// ทดสอบกับ PostgreSQL จริง: BC_GL_TEST_POSTGRES_DSN ชี้ไปฐานที่ทิ้งได้ (สร้าง schema ชั่วคราวต่อเทสต์)
const erpHolding = "H"

// erpDocDay คือวันที่ของเอกสารในเทสต์ — seedLedger สร้างปีบัญชีจากปีของวันนี้ เอกสารกับปีบัญชีจึงตรงกันเสมอไม่ว่ารันปีไหน
var erpDocDay = time.Date(2026, 10, 9, 9, 0, 0, 0, time.UTC)

func erpTestDB(t *testing.T) (*sql.DB, *gl.Postgres) {
	t.Helper()
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN to isolated PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "bc_erp_test_" + fmt.Sprintf("%x", uuid.New())
	if _, err = admin.Exec(`CREATE SCHEMA ` + pq.QuoteIdentifier(schema)); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(`DROP SCHEMA ` + pq.QuoteIdentifier(schema) + ` CASCADE`); err != nil {
			t.Error(err)
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
	projection := gl.NewPostgres(func(holding string) (*sql.DB, error) {
		if holding != erpHolding {
			return nil, fmt.Errorf("holding denied")
		}
		return db, nil
	})
	ctx := context.Background()
	if err = gl.EnsureSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err = ensureSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	return db, projection
}

func erpScope() gl.Scope {
	return gl.Scope{Holding: erpHolding, Company: testCompany, Branch: testBranch, Actor: testActor}
}

func erpIdentity(id string) gl.Identity {
	return gl.Identity{ID: id, HoldingCode: erpHolding, BusinessCode: testCompany, Version: 1}
}

func erpChange(t *testing.T, kind, id, code string, value any) gl.Change {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return gl.Change{Kind: kind, ID: id, Code: code, Payload: string(payload)}
}

// seedLedger วางผังบัญชี ปีบัญชี สมุดรายวัน รูปแบบลงบัญชีของ purchase/pay และคู่ค้า S101 ไว้ในฐานทดสอบ
func seedLedger(t *testing.T, db *sql.DB, p *gl.Postgres) {
	t.Helper()
	changes := []gl.Change{}
	for _, row := range []struct {
		code, kind, normal string
		cash               bool
	}{
		{"1000", "asset", "debit", true},
		{"1300", "asset", "debit", false},
		{"1400", "asset", "debit", false},
		{"2100", "liability", "credit", false},
		{"3000", "equity", "credit", false},
	} {
		a := gl.Account{Identity: erpIdentity("account-" + row.code), AccountCode: row.code, Names: []gl.Name{{Code: "th", Name: "บัญชี " + row.code}}, AccountType: row.kind, NormalBalance: row.normal, AllowPosting: true, IsActive: true, IsCash: row.cash}
		changes = append(changes, erpChange(t, "accounts", a.ID, a.AccountCode, a))
	}
	y := erpDocDay.Year()
	year := gl.FiscalYear{Identity: erpIdentity(fmt.Sprintf("year-%d", y)), Code: fmt.Sprint(y), StartDate: fmt.Sprintf("%d-01-01", y), EndDate: fmt.Sprintf("%d-12-31", y), IsActive: true, Scale: 2, RetainedEarningsAccount: "3000"}
	changes = append(changes, erpChange(t, "fiscal-years", year.ID, year.Code, year))
	for _, code := range []string{"UV", "PV"} {
		book := gl.Master{Identity: erpIdentity("book-" + code), Kind: "journal-books", Code: code, Name: "สมุดรายวัน " + code, BookType: gl.BookTypeGeneral, IsActive: true}
		changes = append(changes, erpChange(t, "journal-books", book.ID, book.Code, book))
	}
	purchase := gl.Master{Identity: erpIdentity("mapping-purchase"), Kind: "mappings", Code: "MAP-PURCHASE", Name: "ซื้อสินค้า", BookCode: "UV", SourceDoc: "purchase", IsActive: true, Rules: []gl.MappingRule{
		{AccountCode: "1400", Side: gl.MappingSideDebit, Source: gl.MappingSourceGoods},
		{AccountCode: "1300", Side: gl.MappingSideDebit, Source: gl.MappingSourceVat},
		{AccountCode: "2100", Side: gl.MappingSideCredit, Source: gl.MappingSourceTotal},
	}}
	changes = append(changes, erpChange(t, "mappings", purchase.ID, purchase.Code, purchase))
	pay := gl.Master{Identity: erpIdentity("mapping-pay"), Kind: "mappings", Code: "MAP-PAY", Name: "จ่ายชำระหนี้", BookCode: "PV", SourceDoc: "pay", IsActive: true, Rules: []gl.MappingRule{
		{AccountCode: "2100", Side: gl.MappingSideDebit, Source: gl.MappingSourceTotal},
		{AccountCode: "1000", Side: gl.MappingSideCredit, Source: gl.MappingSourceTotal},
	}}
	changes = append(changes, erpChange(t, "mappings", pay.ID, pay.Code, pay))
	event := gl.Event{ID: "erp-seed-1", HoldingCode: erpHolding, BusinessCode: testCompany, Sequence: 1, RequestID: "erp-seed-1", Actor: "integration-test", Action: "test", OccurredAt: time.Date(2026, 1, 1, 0, 0, 1, 0, time.UTC), Changes: changes}
	if err := p.Project(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	partner, _ := json.Marshal(map[string]any{"code": "S101", "name": "บริษัท สยามวัสดุภัณฑ์ จำกัด"})
	if _, err := db.Exec(`INSERT INTO gl_subledger_partners (company, code, version, payload) VALUES ($1, 'S101', 1, $2)`, testCompany, partner); err != nil {
		t.Fatal(err)
	}
}

func purchaseDoc(docNo string, qty, price string) *ErpTransactionDoc {
	goods := amt(qty).Mul(amt(price))
	vat := money(goods.Mul(amt("0.07")))
	return &ErpTransactionDoc{
		DocNo:          docNo,
		DocDateTime:    erpDocDay.Format(time.RFC3339),
		CustCode:       "S101",
		CustName:       "บริษัท สยามวัสดุภัณฑ์ จำกัด",
		TotalBeforeVat: goods,
		TotalVatValue:  vat,
		TotalAmount:    goods.Add(vat),
		Details:        []ErpDetailItem{{ItemCode: "ITM-001", Qty: amt(qty), Price: amt(price), SumAmount: goods, WhCode: "WH01"}},
	}
}

func scalarString(t *testing.T, db *sql.DB, query string, args ...any) string {
	t.Helper()
	var v sql.NullString
	if err := db.QueryRow(query, args...).Scan(&v); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return v.String
}

func requireDecimal(t *testing.T, label, got, want string) {
	t.Helper()
	g, err := decimal.NewFromString(got)
	if err != nil {
		t.Fatalf("%s: %q is not a decimal", label, got)
	}
	if !g.Equal(decimal.RequireFromString(want)) {
		t.Fatalf("%s = %s, want %s", label, got, want)
	}
}

// ledgerSums คืนยอดเดบิต/เครดิตรวมของใบสำคัญที่ผ่านรายการแล้วทั้งบริษัท — ต้องเท่ากันเสมอ
func ledgerSums(t *testing.T, db *sql.DB) (debit, credit string) {
	t.Helper()
	rows, err := db.Query(`SELECT COALESCE(SUM((l->>'debit')::numeric), 0)::text, COALESCE(SUM((l->>'credit')::numeric), 0)::text
		FROM gl_records r, jsonb_array_elements(r.payload->'lines') l
		WHERE r.company=$1 AND r.kind='journals' AND r.payload->>'status' IN ('posted','reversed')`, testCompany)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("no sums")
	}
	if err = rows.Scan(&debit, &credit); err != nil {
		t.Fatal(err)
	}
	return debit, credit
}

func TestIntegrationPurchaseCreateEditDeleteKeepsLedgerStockAndSubledgerInStep(t *testing.T) {
	db, projection := erpTestDB(t)
	seedLedger(t, db, projection)
	ctx := context.Background()
	store := NewStore()
	scope := erpScope()
	dt := mustDocType(t, "purchase")

	// สร้าง: ซื้อ 10 ชิ้น × 100 = 1,000 + VAT 70 = 1,070
	saved, err := store.SaveDoc(ctx, db, scope, dt, purchaseDoc("", "10", "100"), false)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if saved.GLStatus != glStatusPosted || saved.GLDocNo == "" {
		t.Fatalf("create gl status = %q docno = %q", saved.GLStatus, saved.GLDocNo)
	}
	docNo := saved.DocNo
	debit, credit := ledgerSums(t, db)
	requireDecimal(t, "debit after create", debit, "1070")
	requireDecimal(t, "credit after create", credit, "1070")
	requireDecimal(t, "stock qty after create", scalarString(t, db, `SELECT balanceqty::text FROM stock_ledger WHERE businesscode=$1 AND docno=$2`, testCompany, docNo), "10")
	requireDecimal(t, "ap amount after create", scalarString(t, db, `SELECT amount::text FROM gl_subledger_documents WHERE company=$1 AND id=$2`, testCompany, "AP-"+docNo), "1070")
	if n := scalarString(t, db, `SELECT count(*)::text FROM erp_doc_journals WHERE businesscode=$1 AND docno=$2 AND voided_at IS NULL`, testCompany, docNo); n != "1" {
		t.Fatalf("active journals = %s, want 1", n)
	}

	// แก้: เปลี่ยนเป็น 20 ชิ้น × 100 = 2,000 + VAT 140 = 2,140 → ใบเดิมถูกกลับรายการ ใบใหม่ revision 2
	edit := purchaseDoc(docNo, "20", "100")
	edited, err := store.SaveDoc(ctx, db, scope, dt, edit, true)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if edited.GLStatus != glStatusPosted || edited.GLDocNo == saved.GLDocNo {
		t.Fatalf("edit gl status = %q docno = %q (was %q)", edited.GLStatus, edited.GLDocNo, saved.GLDocNo)
	}
	debit, credit = ledgerSums(t, db)
	// 1,070 + reversal 1,070 + 2,140 ทั้งสองฝั่ง
	requireDecimal(t, "debit after edit", debit, "4280")
	requireDecimal(t, "credit after edit", credit, "4280")
	if n := scalarString(t, db, `SELECT count(*)::text FROM erp_doc_journals WHERE businesscode=$1 AND docno=$2`, testCompany, docNo); n != "2" {
		t.Fatalf("journal revisions = %s, want 2", n)
	}
	if n := scalarString(t, db, `SELECT count(*)::text FROM erp_doc_journals WHERE businesscode=$1 AND docno=$2 AND voided_at IS NULL`, testCompany, docNo); n != "1" {
		t.Fatalf("active journals after edit = %s, want 1", n)
	}
	requireDecimal(t, "stock qty after edit", scalarString(t, db, `SELECT balanceqty::text FROM stock_ledger WHERE businesscode=$1 AND docno=$2`, testCompany, docNo), "20")
	requireDecimal(t, "ap amount after edit", scalarString(t, db, `SELECT amount::text FROM gl_subledger_documents WHERE company=$1 AND id=$2`, testCompany, "AP-"+docNo), "2140")
	if n := scalarString(t, db, `SELECT count(*)::text FROM gl_subledger_documents WHERE company=$1 AND ledger='ap'`, testCompany); n != "1" {
		t.Fatalf("ap rows after edit = %s, want 1", n)
	}

	// ลบ: กลับรายการใบที่ใช้งานอยู่ สต็อกและหนี้ของเอกสารหายไป แต่ใบสำคัญเดิมยังอยู่เป็นหลักฐาน
	if err = store.DeleteDoc(ctx, db, scope, dt, docNo); err != nil {
		t.Fatalf("delete: %v", err)
	}
	debit, credit = ledgerSums(t, db)
	requireDecimal(t, "debit after delete", debit, "6420")
	requireDecimal(t, "credit after delete", credit, "6420")
	if n := scalarString(t, db, `SELECT count(*)::text FROM erp_doc_journals WHERE businesscode=$1 AND docno=$2 AND voided_at IS NULL`, testCompany, docNo); n != "0" {
		t.Fatalf("active journals after delete = %s, want 0", n)
	}
	if n := scalarString(t, db, `SELECT count(*)::text FROM stock_ledger WHERE businesscode=$1 AND docno=$2`, testCompany, docNo); n != "0" {
		t.Fatalf("stock rows after delete = %s, want 0", n)
	}
	if n := scalarString(t, db, `SELECT count(*)::text FROM gl_subledger_documents WHERE company=$1 AND ledger='ap'`, testCompany); n != "0" {
		t.Fatalf("ap rows after delete = %s, want 0", n)
	}
	if err = store.DeleteDoc(ctx, db, scope, dt, docNo); err != sql.ErrNoRows {
		t.Fatalf("second delete = %v, want sql.ErrNoRows", err)
	}
	// เลขที่เอกสารที่เคยใช้แล้วใช้ซ้ำไม่ได้
	if _, err = store.SaveDoc(ctx, db, scope, dt, purchaseDoc(docNo, "1", "100"), false); userErrorCode(err) != "erp_doc_duplicate" {
		t.Fatalf("reuse of deleted docno = %v, want erp_doc_duplicate", err)
	}
}

// journalLines คือรายการบัญชีของใบสำคัญหนึ่งใบ: รหัสบัญชี → "เดบิต/เครดิต" (ทศนิยมจาก numeric ตรงตัว)
func journalLines(t *testing.T, db *sql.DB, glDocNo string) map[string]string {
	t.Helper()
	rows, err := db.Query(`SELECT l->>'accountcode', (l->>'debit')::numeric::text, (l->>'credit')::numeric::text
		FROM gl_records r, jsonb_array_elements(r.payload->'lines') l
		WHERE r.company=$1 AND r.kind='journals' AND r.payload->>'docno'=$2`, testCompany, glDocNo)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	lines := map[string]string{}
	for rows.Next() {
		var code, debit, credit string
		if err = rows.Scan(&code, &debit, &credit); err != nil {
			t.Fatal(err)
		}
		lines[code] = amt(debit).String() + "/" + amt(credit).String()
	}
	return lines
}

// Rule 20: ยอดหลายบรรทัด + VAT ผ่าน PostgreSQL numeric ไปกลับต้องตรงตัว (0.1 + 0.2 = 0.3, จำนวน 8 ตำแหน่ง, ยอดหลักล้าน)
// และใบสำคัญที่ลงจากเอกสารต้องเดบิต = เครดิต; ยอดหัวเอกสารที่มีเศษ .005 ปัดครึ่งออกจากศูนย์ตอนลงบัญชี
func TestIntegrationDecimalAmountsRoundTripAndPostBalanced(t *testing.T) {
	db, projection := erpTestDB(t)
	seedLedger(t, db, projection)
	ctx := context.Background()
	store := NewStore()
	scope := erpScope()
	dt := mustDocType(t, "purchase")

	var doc ErpTransactionDoc
	// คำขอจากหน้าจอ: ตัวเลขและข้อความทศนิยมปนกัน — ก่อนภาษี 0.10 + 0.20 + 2,469,135.78 + 12.35 = 2,469,148.43;
	// VAT 7% = 172,840.3901 → ปัดระดับเอกสาร 172,840.39; รวม 2,641,988.82
	body := `{"docdatetime":"` + erpDocDay.Format(time.RFC3339) + `","custcode":"S101","custname":"บริษัท สยามวัสดุภัณฑ์ จำกัด",
		"totalbeforevat":"2469148.43","totalvatvalue":172840.39,"totalamount":"2641988.82","vattype":1,"vatrate":"7",
		"details":[
			{"itemcode":"ITM-001","qty":1,"price":0.1,"sumamount":"0.10","whcode":"WH01"},
			{"itemcode":"ITM-001","qty":"1","price":"0.20","sumamount":0.2,"whcode":"WH01"},
			{"itemcode":"ITM-002","qty":"2","price":"1234567.89","sumamount":"2469135.78","whcode":"WH01"},
			{"itemcode":"ITM-002","qty":"0.12345678","price":"100.00","sumamount":"12.35","whcode":"WH01"}]}`
	if err := json.Unmarshal([]byte(body), &doc); err != nil {
		t.Fatal(err)
	}
	saved, err := store.SaveDoc(ctx, db, scope, dt, &doc, false)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if saved.GLStatus != glStatusPosted {
		t.Fatalf("gl status = %q, want posted", saved.GLStatus)
	}

	got, err := store.GetDoc(ctx, db, testCompany, dt.Flag, saved.DocNo)
	if err != nil {
		t.Fatal(err)
	}
	for label, pair := range map[string][2]decimal.Decimal{
		"totalbeforevat": {got.TotalBeforeVat, amt("2469148.43")},
		"totalvatvalue":  {got.TotalVatValue, amt("172840.39")},
		"totalamount":    {got.TotalAmount, amt("2641988.82")},
		"0.1+0.2":        {got.Details[0].SumAmount.Add(got.Details[1].SumAmount), amt("0.3")},
		"line 3 price":   {got.Details[2].Price, amt("1234567.89")},
		"line 4 qty":     {got.Details[3].Qty, amt("0.12345678")},
	} {
		if !pair[0].Equal(pair[1]) {
			t.Errorf("%s round trip = %s, want %s", label, pair[0], pair[1])
		}
	}
	out, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"totalamount":"2641988.82"`, `"totalvatvalue":"172840.39"`, `"qty":"0.12345678"`, `"price":"1234567.89"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("read-back JSON must carry %s, got %s", want, out)
		}
	}

	lines := journalLines(t, db, saved.GLDocNo)
	want := map[string]string{"1400": "2469148.43/0", "1300": "172840.39/0", "2100": "0/2641988.82"}
	if len(lines) != len(want) {
		t.Fatalf("journal lines = %v, want %v", lines, want)
	}
	for code, amounts := range want {
		if lines[code] != amounts {
			t.Errorf("account %s = %s, want %s", code, lines[code], amounts)
		}
	}
	debit, credit := ledgerSums(t, db)
	requireDecimal(t, "debit", debit, "2641988.82")
	requireDecimal(t, "credit", credit, "2641988.82")

	// ยอดหัวเอกสาร 107.005 เก็บ numeric(18,8) ตรงตัว แล้วลงบัญชี 107.01 (ปัดครึ่งออกจากศูนย์) มูลค่าสินค้า = 107.01 − 7 = 100.01
	half := purchaseDoc("", "1", "100")
	half.TotalBeforeVat, half.TotalVatValue, half.TotalAmount = amt("100.005"), amt("7"), amt("107.005")
	halfSaved, err := store.SaveDoc(ctx, db, scope, dt, half, false)
	if err != nil {
		t.Fatalf("save .005: %v", err)
	}
	requireDecimal(t, "stored .005 total", scalarString(t, db, `SELECT totalamount::text FROM doc WHERE businesscode=$1 AND docno=$2`, testCompany, halfSaved.DocNo), "107.005")
	halfLines := journalLines(t, db, halfSaved.GLDocNo)
	if halfLines["2100"] != "0/107.01" || halfLines["1300"] != "7/0" || halfLines["1400"] != "100.01/0" {
		t.Errorf(".005 journal lines = %v, want 2100 0/107.01, 1300 7/0, 1400 100.01/0", halfLines)
	}
	debit, credit = ledgerSums(t, db)
	requireDecimal(t, "debit after .005", debit, "2642095.83")
	if !amt(debit).Equal(amt(credit)) {
		t.Errorf("ledger debit %s != credit %s", debit, credit)
	}
}

func TestIntegrationSettledPurchaseCannotBeEditedOrDeleted(t *testing.T) {
	db, projection := erpTestDB(t)
	seedLedger(t, db, projection)
	ctx := context.Background()
	store := NewStore()
	scope := erpScope()
	purchase := mustDocType(t, "purchase")
	pay := mustDocType(t, "pay")

	bought, err := store.SaveDoc(ctx, db, scope, purchase, purchaseDoc("", "10", "100"), false)
	if err != nil {
		t.Fatalf("purchase: %v", err)
	}
	// ใบจ่ายชำระหนี้ตัดหนี้รายบรรทัด (debt_document_id) — docref ที่หัวเหลือไว้เป็นการอ้างอิงเอกสาร
	payment := &ErpTransactionDoc{DocDateTime: erpDocDay.Format(time.RFC3339), CustCode: "S101", DocRef: bought.DocNo, TotalAmount: amt("1070"),
		Details: []ErpDetailItem{{LineNumber: 1, DebtDocumentID: "AP-" + bought.DocNo, DocRef: bought.DocNo, SumAmount: amt("1070")}}}
	paid, err := store.SaveDoc(ctx, db, scope, pay, payment, false)
	if err != nil {
		t.Fatalf("pay: %v", err)
	}
	if paid.GLStatus != glStatusPosted {
		t.Fatalf("pay gl status = %q", paid.GLStatus)
	}
	if n := scalarString(t, db, `SELECT count(*)::text FROM gl_subledger_settlements WHERE company=$1 AND debt_id=$2`, testCompany, "AP-"+bought.DocNo); n != "1" {
		t.Fatalf("settlements against debt = %s, want 1", n)
	}

	before, _ := ledgerSums(t, db)
	if _, err = store.SaveDoc(ctx, db, scope, purchase, purchaseDoc(bought.DocNo, "5", "100"), true); userErrorCode(err) != "erp_doc_settled" {
		t.Fatalf("edit of settled purchase = %v, want erp_doc_settled", err)
	}
	// ใบจ่ายอ้างอิงใบซื้อ (docref) — ลบถูกหยุดที่ตรวจเอกสารปลายทางก่อนถึงตรวจการตัดหนี้
	if err = store.DeleteDoc(ctx, db, scope, purchase, bought.DocNo); userErrorCode(err) != "erp_doc_has_downstream" {
		t.Fatalf("delete of settled purchase = %v, want erp_doc_has_downstream", err)
	}
	after, _ := ledgerSums(t, db)
	requireDecimal(t, "ledger untouched by refused edit", after, before)
	requireDecimal(t, "stock untouched by refused edit", scalarString(t, db, `SELECT balanceqty::text FROM stock_ledger WHERE businesscode=$1 AND docno=$2`, testCompany, bought.DocNo), "10")

	// ยกเลิกใบจ่ายก่อน (ปลดการตัดหนี้) แล้วใบซื้อจึงลบได้
	if err = store.DeleteDoc(ctx, db, scope, pay, paid.DocNo); err != nil {
		t.Fatalf("delete payment: %v", err)
	}
	if err = store.DeleteDoc(ctx, db, scope, purchase, bought.DocNo); err != nil {
		t.Fatalf("delete purchase after payment removed: %v", err)
	}
}

func TestIntegrationDocWithoutMappingSavesUnposted(t *testing.T) {
	db, projection := erpTestDB(t)
	seedLedger(t, db, projection)
	ctx := context.Background()
	store := NewStore()
	scope := erpScope()
	// sale-invoice ไม่มีรูปแบบลงบัญชีในชุดทดสอบ: บันทึกได้ สต็อกตัดออก แต่ไม่มีใบสำคัญและไม่ error
	dt := mustDocType(t, "sale-invoice")
	instant := erpDocDay.Format(time.RFC3339)
	stockIn := purchaseDoc("", "10", "100")
	stockIn.DocDateTime = instant
	if _, err := store.SaveDoc(ctx, db, scope, mustDocType(t, "purchase"), stockIn, false); err != nil {
		t.Fatalf("stock-in purchase: %v", err)
	}
	sale := &ErpTransactionDoc{DocDateTime: instant, CustCode: "C101", TotalAmount: amt("321"), TotalBeforeVat: amt("300"), TotalVatValue: amt("21"),
		Details: []ErpDetailItem{{ItemCode: "ITM-001", Qty: amt("3"), Price: amt("100"), SumAmount: amt("300"), WhCode: "WH01"}}}
	saved, err := store.SaveDoc(ctx, db, scope, dt, sale, false)
	if err != nil {
		t.Fatalf("sale: %v", err)
	}
	if saved.GLStatus != glStatusNoMapping || saved.GLDocNo != "" {
		t.Fatalf("gl status = %q docno = %q, want no_mapping", saved.GLStatus, saved.GLDocNo)
	}
	requireDecimal(t, "stock after sale", scalarString(t, db, `SELECT balanceqty::text FROM stock_ledger WHERE businesscode=$1 AND docno=$2`, testCompany, saved.DocNo), "7")
	// หน้าจอส่งเวลาเดียวกันทั้งวัน (T09:00:00Z): ใบซื้อกับใบขายวันเดียวกันต้องเรียงตามลำดับที่บันทึก ไม่ใช่ตามเลขที่เอกสาร (INV มาก่อน PUR)
	// และการแก้ใบซื้อภายหลังต้องคงตำแหน่งเดิม ไม่ย้ายไปต่อท้ายใบขาย
	requireDecimal(t, "purchase balance", scalarString(t, db, `SELECT balanceqty::text FROM stock_ledger WHERE businesscode=$1 AND docno LIKE 'PUR%'`, testCompany), "10")
	same := purchaseDoc(scalarString(t, db, `SELECT docno FROM stock_ledger WHERE businesscode=$1 AND docno LIKE 'PUR%'`, testCompany), "12", "100")
	same.DocDateTime = instant
	if _, err = store.SaveDoc(ctx, db, scope, mustDocType(t, "purchase"), same, true); err != nil {
		t.Fatalf("edit purchase: %v", err)
	}
	requireDecimal(t, "sale balance after purchase edit", scalarString(t, db, `SELECT balanceqty::text FROM stock_ledger WHERE businesscode=$1 AND docno=$2`, testCompany, saved.DocNo), "9")
	if err = store.DeleteDoc(ctx, db, scope, dt, saved.DocNo); err != nil {
		t.Fatalf("delete sale: %v", err)
	}
}

func TestIntegrationBackDatedStockDocumentRebuildsLaterBalances(t *testing.T) {
	db, projection := erpTestDB(t)
	seedLedger(t, db, projection)
	ctx := context.Background()
	store := NewStore()
	scope := erpScope()
	receive := mustDocType(t, "stock-receive-product")
	pickup := mustDocType(t, "stock-pickup-product")

	day := func(d int) string {
		return time.Date(erpDocDay.Year(), 9, d, 8, 0, 0, 0, time.UTC).Format(time.RFC3339)
	}
	line := func(qty, price string) []ErpDetailItem {
		return []ErpDetailItem{{ItemCode: "ITM-002", Qty: amt(qty), Price: amt(price), SumAmount: amt(qty).Mul(amt(price)), WhCode: "WH01"}}
	}
	// วันที่ 10 รับ 10 @100, วันที่ 20 เบิก 4 → คงเหลือ 6
	if _, err := store.SaveDoc(ctx, db, scope, receive, &ErpTransactionDoc{DocNo: "RCV-1", DocDateTime: day(10), Details: line("10", "100")}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveDoc(ctx, db, scope, pickup, &ErpTransactionDoc{DocNo: "PCK-1", DocDateTime: day(20), Details: line("4", "0")}, false); err != nil {
		t.Fatal(err)
	}
	requireDecimal(t, "balance after pickup", scalarString(t, db, `SELECT balanceqty::text FROM stock_ledger WHERE businesscode=$1 AND docno='PCK-1'`, testCompany), "6")
	// ย้อนหลังวันที่ 15 รับเพิ่ม 5 @120 → ยอดของใบเบิกวันที่ 20 ต้องคำนวณใหม่เป็น 15 - 4 = 11
	if _, err := store.SaveDoc(ctx, db, scope, receive, &ErpTransactionDoc{DocNo: "RCV-2", DocDateTime: day(15), Details: line("5", "120")}, false); err != nil {
		t.Fatal(err)
	}
	requireDecimal(t, "balance of later pickup after back-dated receipt", scalarString(t, db, `SELECT balanceqty::text FROM stock_ledger WHERE businesscode=$1 AND docno='PCK-1'`, testCompany), "11")
	// ลบใบรับย้อนหลังออก → กลับเป็น 6
	if err := store.DeleteDoc(ctx, db, scope, receive, "RCV-2"); err != nil {
		t.Fatal(err)
	}
	requireDecimal(t, "balance after back-dated receipt removed", scalarString(t, db, `SELECT balanceqty::text FROM stock_ledger WHERE businesscode=$1 AND docno='PCK-1'`, testCompany), "6")
}

// docState อ่านสถานะอนุมัติ/ยกเลิก/ปิดจากฐานจริง (ไม่เชื่อค่าที่ store คืน)
func docState(t *testing.T, db *sql.DB, flag int, docNo string) string {
	t.Helper()
	return scalarString(t, db, `SELECT COALESCE(approval_status, '') || '|' || COALESCE(iscancel, false)::text || '|' || COALESCE(isclosed, false)::text
		FROM doc WHERE businesscode = $1 AND transflag = $2 AND docno = $3`, testCompany, flag, docNo)
}

// docEvents อ่านประวัติเอกสารเรียงตามลำดับที่บันทึก เป็น action|from_status|to_status|reason|actor ต่อแถว
func docEvents(t *testing.T, db *sql.DB, flag int, docNo string) []string {
	t.Helper()
	rows, err := db.Query(`SELECT action || '|' || from_status || '|' || to_status || '|' || reason || '|' || actor
		FROM erp_doc_events WHERE businesscode = $1 AND transflag = $2 AND docno = $3 ORDER BY id`, testCompany, flag, docNo)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var events []string
	for rows.Next() {
		var row string
		if err := rows.Scan(&row); err != nil {
			t.Fatal(err)
		}
		events = append(events, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return events
}

func requireEvents(t *testing.T, label string, got, want []string) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("%s events = %q, want %q", label, got, want)
	}
}

// อนุมัติ → ยกเลิกอนุมัติ → ยกเลิกใบขอซื้อ แต่ละขั้นเปลี่ยนสถานะเอกสารและเขียน erp_doc_events หนึ่งแถว;
// ปรับปรุงสถานะใบรับสินค้าเขียนแถว adjust_status ผ่าน CHECK ของตารางจริง
func TestApproveUnapproveCancelWritesEvents(t *testing.T) {
	db, projection := erpTestDB(t)
	seedLedger(t, db, projection)
	ctx := context.Background()
	store := NewStore()
	scope := erpScope()
	pr := mustDocType(t, "purchase-requisition")
	instant := erpDocDay.AddDate(0, 0, -1).Format(time.RFC3339)

	saved, err := store.SaveDoc(ctx, db, scope, pr, &ErpTransactionDoc{DocDateTime: instant, CustCode: "S101", CustName: "บริษัท สยามวัสดุภัณฑ์ จำกัด",
		Description: "ขอซื้อกระดาษถ่ายเอกสาร A4", TotalBeforeVat: amt("1000"), TotalVatValue: amt("70"), TotalAmount: amt("1070"),
		Details: []ErpDetailItem{{ItemCode: "ITM-001", Qty: amt("10"), Price: amt("100"), SumAmount: amt("1000"), WhCode: "WH01"}}}, false)
	if err != nil {
		t.Fatalf("save purchase requisition: %v", err)
	}
	docNo := saved.DocNo
	if got := docState(t, db, pr.Flag, docNo); got != "|false|false" {
		t.Fatalf("new requisition state = %q", got)
	}
	requireEvents(t, "after save", docEvents(t, db, pr.Flag, docNo), nil)

	if _, err = store.ApproveDoc(ctx, db, scope, pr, docNo); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if got := docState(t, db, pr.Flag, docNo); got != "approved|false|false" {
		t.Fatalf("approved state = %q", got)
	}
	approve := "approve||approved||" + testActor
	requireEvents(t, "after approve", docEvents(t, db, pr.Flag, docNo), []string{approve})

	if _, err = store.UnapproveDoc(ctx, db, scope, pr, docNo, "  ผู้ขอซื้อขอแก้จำนวน  "); err != nil {
		t.Fatalf("unapprove: %v", err)
	}
	if got := docState(t, db, pr.Flag, docNo); got != "|false|false" {
		t.Fatalf("unapproved state = %q", got)
	}
	unapprove := "unapprove|approved||ผู้ขอซื้อขอแก้จำนวน|" + testActor
	requireEvents(t, "after unapprove", docEvents(t, db, pr.Flag, docNo), []string{approve, unapprove})

	if _, err = store.CancelDoc(ctx, db, scope, pr, docNo, "ฝ่ายจัดซื้อยกเลิกคำขอ"); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if got := docState(t, db, pr.Flag, docNo); got != "|true|false" {
		t.Fatalf("cancelled state = %q", got)
	}
	if got := scalarString(t, db, `SELECT cancelreason FROM doc WHERE businesscode = $1 AND transflag = $2 AND docno = $3`, testCompany, pr.Flag, docNo); got != "ฝ่ายจัดซื้อยกเลิกคำขอ" {
		t.Fatalf("cancelreason = %q", got)
	}
	requireEvents(t, "after cancel", docEvents(t, db, pr.Flag, docNo), []string{approve, unapprove, "cancel||cancelled|ฝ่ายจัดซื้อยกเลิกคำขอ|" + testActor})
	// ยกเลิกแล้วอนุมัติซ้ำไม่ได้ และไม่มีแถวประวัติเพิ่ม
	if _, err = store.ApproveDoc(ctx, db, scope, pr, docNo); err == nil {
		t.Fatal("approve after cancel succeeded")
	}
	if got := len(docEvents(t, db, pr.Flag, docNo)); got != 3 {
		t.Fatalf("events after refused approve = %d, want 3", got)
	}

	gr := mustDocType(t, "goodsreceive")
	receipt, err := store.SaveDoc(ctx, db, scope, gr, &ErpTransactionDoc{DocDateTime: instant, CustCode: "S101", CustName: "บริษัท สยามวัสดุภัณฑ์ จำกัด",
		TotalBeforeVat: amt("1000"), TotalVatValue: amt("70"), TotalAmount: amt("1070"),
		Details: []ErpDetailItem{{ItemCode: "ITM-001", Qty: amt("10"), Price: amt("100"), SumAmount: amt("1000"), WhCode: "WH01"}}}, false)
	if err != nil {
		t.Fatalf("save goods receipt: %v", err)
	}
	count, err := store.ProcessReceiptAdjustments(ctx, db, scope, []string{receipt.DocNo})
	if err != nil || count != 1 {
		t.Fatalf("ProcessReceiptAdjustments = %d, %v; want 1", count, err)
	}
	if got := docState(t, db, gr.Flag, receipt.DocNo); got != "|false|true" {
		t.Fatalf("adjusted goods receipt state = %q", got)
	}
	requireEvents(t, "after receipt adjustment", docEvents(t, db, gr.Flag, receipt.DocNo), []string{"adjust_status||closed|" + eventGRClosed + "|" + testActor})
}

// seedReceivables เพิ่มบัญชีลูกหนี้/รายได้/ภาษีขาย รูปแบบลงบัญชีของ sale-invoice/paid และลูกค้า C101 ต่อจาก seedLedger
// (แยกไว้ เพราะเทสต์อื่นต้องการ sale-invoice ที่ยังไม่มีรูปแบบลงบัญชี)
func seedReceivables(t *testing.T, db *sql.DB, p *gl.Postgres) {
	t.Helper()
	changes := []gl.Change{}
	for _, row := range []struct{ code, kind, normal string }{
		{"1100", "asset", "debit"},
		{"2200", "liability", "credit"},
		{"4000", "income", "credit"},
	} {
		a := gl.Account{Identity: erpIdentity("account-" + row.code), AccountCode: row.code, Names: []gl.Name{{Code: "th", Name: "บัญชี " + row.code}}, AccountType: row.kind, NormalBalance: row.normal, AllowPosting: true, IsActive: true}
		changes = append(changes, erpChange(t, "accounts", a.ID, a.AccountCode, a))
	}
	for _, code := range []string{"SV", "RV"} {
		book := gl.Master{Identity: erpIdentity("book-" + code), Kind: "journal-books", Code: code, Name: "สมุดรายวัน " + code, BookType: gl.BookTypeGeneral, IsActive: true}
		changes = append(changes, erpChange(t, "journal-books", book.ID, book.Code, book))
	}
	sale := gl.Master{Identity: erpIdentity("mapping-sale"), Kind: "mappings", Code: "MAP-SALE", Name: "ขายเชื่อ", BookCode: "SV", SourceDoc: "sale-invoice", IsActive: true, Rules: []gl.MappingRule{
		{AccountCode: "1100", Side: gl.MappingSideDebit, Source: gl.MappingSourceTotal},
		{AccountCode: "4000", Side: gl.MappingSideCredit, Source: gl.MappingSourceGoods},
		{AccountCode: "2200", Side: gl.MappingSideCredit, Source: gl.MappingSourceVat},
	}}
	changes = append(changes, erpChange(t, "mappings", sale.ID, sale.Code, sale))
	receipt := gl.Master{Identity: erpIdentity("mapping-paid"), Kind: "mappings", Code: "MAP-PAID", Name: "รับชำระหนี้", BookCode: "RV", SourceDoc: "paid", IsActive: true, Rules: []gl.MappingRule{
		{AccountCode: "1000", Side: gl.MappingSideDebit, Source: gl.MappingSourceTotal},
		{AccountCode: "1100", Side: gl.MappingSideCredit, Source: gl.MappingSourceTotal},
	}}
	changes = append(changes, erpChange(t, "mappings", receipt.ID, receipt.Code, receipt))
	salesReturn := gl.Master{Identity: erpIdentity("mapping-sale-return"), Kind: "mappings", Code: "MAP-SR", Name: "รับคืนสินค้า ลดหนี้", BookCode: "SV", SourceDoc: "sale-invoice-return", IsActive: true, Rules: []gl.MappingRule{
		{AccountCode: "4000", Side: gl.MappingSideDebit, Source: gl.MappingSourceGoods},
		{AccountCode: "2200", Side: gl.MappingSideDebit, Source: gl.MappingSourceVat},
		{AccountCode: "1100", Side: gl.MappingSideCredit, Source: gl.MappingSourceTotal},
	}}
	changes = append(changes, erpChange(t, "mappings", salesReturn.ID, salesReturn.Code, salesReturn))
	event := gl.Event{ID: "erp-seed-2", HoldingCode: erpHolding, BusinessCode: testCompany, Sequence: 2, RequestID: "erp-seed-2", Actor: "integration-test", Action: "test", OccurredAt: time.Date(2026, 1, 1, 0, 0, 2, 0, time.UTC), Changes: changes}
	if err := p.Project(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	for code, name := range map[string]string{"C101": "บริษัท ไทยเจริญพาณิชย์ จำกัด", "C102": "ห้างหุ้นส่วนจำกัด เชียงใหม่ค้าวัสดุ"} {
		partner, _ := json.Marshal(map[string]any{"code": code, "name": name})
		if _, err := db.Exec(`INSERT INTO gl_subledger_partners (company, code, version, payload) VALUES ($1, $2, 1, $3)`, testCompany, code, partner); err != nil {
			t.Fatal(err)
		}
	}
}

// subledgerOutstanding คือยอดค้างของเอกสารบัญชีย่อย id (หนี้ หรือเครดิตค้างของใบรับคืน) ตามสูตรเดียวกับการตัดหนี้
func subledgerOutstanding(t *testing.T, db *sql.DB, id string) string {
	t.Helper()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	v, err := debtOutstanding(context.Background(), tx, testCompany, id)
	if err != nil {
		t.Fatal(err)
	}
	return v.String()
}

// ใบเสร็จรับเงินตัดหนี้ใบกำกับได้ไม่เกินยอดคงเหลือ (Champ arfrmbcreceipt1view.cpp:1760): ใบที่เกินถูกปฏิเสธพร้อมยอดสูงสุด
// และไม่มีอะไรถูกเขียน; ตัดพอดียอดคงเหลือ → หนี้คงเหลือเป็นศูนย์ (เดิมยอดเกินถูกตัดเต็มจำนวน หนี้ติดลบ)
func TestIntegrationReceiptCannotSettleMoreThanInvoiceOutstanding(t *testing.T) {
	db, projection := erpTestDB(t)
	seedLedger(t, db, projection)
	seedReceivables(t, db, projection)
	ctx := context.Background()
	store := NewStore()
	scope := erpScope()
	paid := mustDocType(t, "paid")
	day := erpDocDay.Format(time.RFC3339)

	invoice, err := store.SaveDoc(ctx, db, scope, mustDocType(t, "sale-invoice"), &ErpTransactionDoc{DocDateTime: day, CustCode: "C101",
		CustName: "บริษัท ไทยเจริญพาณิชย์ จำกัด", TotalBeforeVat: amt("1000"), TotalVatValue: amt("70"), TotalAmount: amt("1070")}, false)
	if err != nil {
		t.Fatalf("sale invoice: %v", err)
	}
	if invoice.GLStatus != glStatusPosted {
		t.Fatalf("invoice gl status = %q", invoice.GLStatus)
	}
	debtID := "AR-" + invoice.DocNo
	outstanding := func() string { return subledgerOutstanding(t, db, debtID) }
	receipt := func(amount string) *ErpTransactionDoc {
		return &ErpTransactionDoc{DocDateTime: day, CustCode: "C101", DocRef: invoice.DocNo, TotalAmount: amt(amount)}
	}

	if _, err = store.SaveDoc(ctx, db, scope, paid, receipt("600"), false); err != nil {
		t.Fatalf("first receipt: %v", err)
	}
	requireDecimal(t, "outstanding after first receipt", outstanding(), "470")

	debitBefore, _ := ledgerSums(t, db)
	_, err = store.SaveDoc(ctx, db, scope, paid, receipt("500"), false)
	user, ok := gl.AsUserError(err)
	if !ok || user.Code != "erp_settle_exceeds_remaining" {
		t.Fatalf("receipt over the remainder = %v, want erp_settle_exceeds_remaining", err)
	}
	if len(user.Args) != 2 || user.Args[0] != invoice.DocNo || user.Args[1] != "470.00" {
		t.Fatalf("error args = %q, want [%s 470.00]", user.Args, invoice.DocNo)
	}
	if !strings.Contains(user.Message, "470.00") || strings.Contains(user.Message, "gl_err_") {
		t.Fatalf("error message = %q, want the Thai text with the outstanding amount", user.Message)
	}
	// ปฏิเสธทั้งธุรกรรม: ไม่มีเอกสาร ใบสำคัญ บัญชีย่อย หรือการตัดหนี้ของใบที่สองค้างอยู่
	if n := scalarString(t, db, `SELECT count(*)::text FROM doc WHERE businesscode=$1 AND transflag=$2`, testCompany, paid.Flag); n != "1" {
		t.Fatalf("receipt documents after refused receipt = %s, want 1", n)
	}
	if n := scalarString(t, db, `SELECT count(*)::text FROM gl_subledger_settlements WHERE company=$1 AND debt_id=$2`, testCompany, debtID); n != "1" {
		t.Fatalf("settlements after refused receipt = %s, want 1", n)
	}
	if n := scalarString(t, db, `SELECT count(*)::text FROM gl_subledger_documents WHERE company=$1 AND id LIKE 'RCP-%'`, testCompany); n != "1" {
		t.Fatalf("receipt subledger rows after refused receipt = %s, want 1", n)
	}
	debitAfter, _ := ledgerSums(t, db)
	requireDecimal(t, "ledger untouched by refused receipt", debitAfter, debitBefore)
	requireDecimal(t, "outstanding after refused receipt", outstanding(), "470")

	// ตัดพอดียอดคงเหลือ → หนี้หมด และตัดเพิ่มอีกแม้ 0.01 ไม่ได้
	if _, err = store.SaveDoc(ctx, db, scope, paid, receipt("470"), false); err != nil {
		t.Fatalf("exact receipt: %v", err)
	}
	requireDecimal(t, "outstanding after exact receipt", outstanding(), "0")
	requireDecimal(t, "settled total", scalarString(t, db, `SELECT SUM(amount)::text FROM gl_subledger_settlements WHERE company=$1 AND debt_id=$2`, testCompany, debtID), "1070")
	_, err = store.SaveDoc(ctx, db, scope, paid, receipt("0.01"), false)
	if user, ok := gl.AsUserError(err); !ok || user.Code != "erp_settle_exceeds_remaining" || user.Args[1] != "0.00" {
		t.Fatalf("receipt on a settled invoice = %v, want erp_settle_exceeds_remaining with 0.00", err)
	}
}

// ใบรับคืนสินค้าอ้างใบกำกับที่ลูกค้าชำระแล้วต้องบันทึกได้ (Champ BILLFrmInvCreditNote4.cpp:4336-4350 ไม่ปฏิเสธ): ตัดหนี้เท่าที่ยังค้าง
// ส่วนเกินคงเป็นเครดิตค้างของใบรับคืน (CN-) — หนี้ไม่ติดลบ และไม่มีแถวตัดหนี้ยอดศูนย์
func TestIntegrationSalesReturnSettlesOnlyOutstandingAndKeepsExcessAsCredit(t *testing.T) {
	db, projection := erpTestDB(t)
	seedLedger(t, db, projection)
	seedReceivables(t, db, projection)
	ctx := context.Background()
	store := NewStore()
	scope := erpScope()
	sale := mustDocType(t, "sale-invoice")
	paid := mustDocType(t, "paid")
	salesReturn := mustDocType(t, "sale-invoice-return")
	day := erpDocDay.Format(time.RFC3339)
	invoiceFor := func(label string) *ErpTransactionDoc {
		t.Helper()
		saved, err := store.SaveDoc(ctx, db, scope, sale, &ErpTransactionDoc{DocDateTime: day, CustCode: "C101", CustName: "บริษัท ไทยเจริญพาณิชย์ จำกัด",
			TotalBeforeVat: amt("1000"), TotalVatValue: amt("70"), TotalAmount: amt("1070")}, false)
		if err != nil {
			t.Fatalf("%s invoice: %v", label, err)
		}
		return saved
	}
	settle := func(dt doctype.DocType, ref string, before, vat, total string) *ErpTransactionDoc {
		t.Helper()
		saved, err := store.SaveDoc(ctx, db, scope, dt, &ErpTransactionDoc{DocDateTime: day, CustCode: "C101", DocRef: ref,
			TotalBeforeVat: amt(before), TotalVatValue: amt(vat), TotalAmount: amt(total)}, false)
		if err != nil {
			t.Fatalf("%s against %s: %v", dt.Key, ref, err)
		}
		return saved
	}
	settlementCount := func(id string) string {
		return scalarString(t, db, `SELECT count(*)::text FROM gl_subledger_settlements WHERE company=$1 AND id=$2`, testCompany, id)
	}

	// ชำระครบแล้วค่อยรับคืน: บันทึกได้ ไม่มีการตัดหนี้ ยอดรับคืนทั้งหมดเป็นเครดิตค้างของลูกค้า
	paidUp := invoiceFor("paid-up")
	settle(paid, paidUp.DocNo, "0", "0", "1070")
	fullReturn := settle(salesReturn, paidUp.DocNo, "500", "35", "535")
	if fullReturn.GLStatus != glStatusPosted {
		t.Fatalf("return gl status = %q", fullReturn.GLStatus)
	}
	if n := settlementCount("SET-CN-" + fullReturn.DocNo); n != "0" {
		t.Fatalf("settlement rows of a return on a paid invoice = %s, want 0", n)
	}
	requireDecimal(t, "paid invoice stays settled", subledgerOutstanding(t, db, "AR-"+paidUp.DocNo), "0")
	requireDecimal(t, "return credit left open", subledgerOutstanding(t, db, "CN-"+fullReturn.DocNo), "535")

	// ชำระบางส่วน (ค้าง 270) แล้วรับคืน 535: ตัดหนี้ 270 พอดี ที่เหลือ 265 เป็นเครดิตค้าง
	partly := invoiceFor("partly-paid")
	settle(paid, partly.DocNo, "0", "0", "800")
	partReturn := settle(salesReturn, partly.DocNo, "500", "35", "535")
	requireDecimal(t, "capped settlement", scalarString(t, db, `SELECT amount::text FROM gl_subledger_settlements WHERE company=$1 AND id=$2`, testCompany, "SET-CN-"+partReturn.DocNo), "270")
	requireDecimal(t, "partly paid invoice after return", subledgerOutstanding(t, db, "AR-"+partly.DocNo), "0")
	requireDecimal(t, "excess return credit", subledgerOutstanding(t, db, "CN-"+partReturn.DocNo), "265")

	// รับคืนน้อยกว่ายอดค้าง: ตัดเต็มจำนวนตามเดิม
	open := invoiceFor("open")
	smallReturn := settle(salesReturn, open.DocNo, "100", "7", "107")
	requireDecimal(t, "open invoice after small return", subledgerOutstanding(t, db, "AR-"+open.DocNo), "963")
	requireDecimal(t, "small return fully used", subledgerOutstanding(t, db, "CN-"+smallReturn.DocNo), "0")
}

// ใบเสร็จของลูกค้าอีกรายอ้างเลขที่ใบกำกับของ C101 ต้องไม่ตัดหนี้ของ C101 (จอบัญชีบังคับคู่ค้า/บัญชีคุมเดียวกัน
// subledger_allocations.go settlement) และไม่เปิดยอดค้างของ C101 ให้เห็นผ่านข้อความผิดพลาด
func TestIntegrationReceiptOfAnotherPartnerDoesNotSettleDebt(t *testing.T) {
	db, projection := erpTestDB(t)
	seedLedger(t, db, projection)
	seedReceivables(t, db, projection)
	ctx := context.Background()
	store := NewStore()
	scope := erpScope()
	day := erpDocDay.Format(time.RFC3339)
	invoice, err := store.SaveDoc(ctx, db, scope, mustDocType(t, "sale-invoice"), &ErpTransactionDoc{DocDateTime: day, CustCode: "C101",
		CustName: "บริษัท ไทยเจริญพาณิชย์ จำกัด", TotalBeforeVat: amt("1000"), TotalVatValue: amt("70"), TotalAmount: amt("1070")}, false)
	if err != nil {
		t.Fatalf("sale invoice: %v", err)
	}
	other, err := store.SaveDoc(ctx, db, scope, mustDocType(t, "paid"), &ErpTransactionDoc{DocDateTime: day, CustCode: "C102", DocRef: invoice.DocNo, TotalAmount: amt("5000")}, false)
	if err != nil {
		t.Fatalf("receipt of another customer: %v", err)
	}
	if n := scalarString(t, db, `SELECT count(*)::text FROM gl_subledger_settlements WHERE company=$1 AND payment_id=$2`, testCompany, "RCP-"+other.DocNo); n != "0" {
		t.Fatalf("settlements by another customer's receipt = %s, want 0", n)
	}
	requireDecimal(t, "C101 invoice untouched", subledgerOutstanding(t, db, "AR-"+invoice.DocNo), "1070")
}

// ใบจ่ายชำระอ้างเลขที่เอกสารที่สร้างจากจอบัญชี (มี balance_side: 1 = หนี้, 2 = ผลชำระ) ต้องตัดได้เฉพาะเอกสารหนี้
// แบบเงื่อนไข debt.Side==1 ของจอบัญชี (subledger_allocations.go settlement) — เดิมใบจ่ายชำระ (ตัด side 2) ไปตัดเอกสารจ่ายชำระของจอบัญชี
// และตัดใบตั้งหนี้ของจอบัญชีไม่ได้; ยอดเกินหนี้คงเหลือของเอกสารจอบัญชีก็ถูกปฏิเสธเหมือนกัน
func TestIntegrationPaymentSettlesOnlyGLDebtDocuments(t *testing.T) {
	db, projection := erpTestDB(t)
	seedLedger(t, db, projection)
	ctx := context.Background()
	store := NewStore()
	scope := erpScope()
	pay := mustDocType(t, "pay")
	purchase, err := store.SaveDoc(ctx, db, scope, mustDocType(t, "purchase"), purchaseDoc("", "1", "100"), false)
	if err != nil {
		t.Fatalf("purchase: %v", err)
	}
	// ใบสำคัญที่ผ่านรายการแล้วของใบซื้อใช้เป็นที่มาของยอดเอกสารจอบัญชีทั้งสองใบ
	var journalID, branch, control string
	if err = db.QueryRow(`SELECT created_journal_id, branch_code, control_account_code FROM gl_subledger_documents WHERE company=$1 AND id=$2`,
		testCompany, "AP-"+purchase.DocNo).Scan(&journalID, &branch, &control); err != nil {
		t.Fatal(err)
	}
	for i, row := range []struct {
		id, amount string
		side       int
	}{{"INV-GL-001", "500", 1}, {"PMT-GL-001", "300", 2}} {
		// payload เหมือนที่จอบัญชีเขียน (SubledgerDocument): มีเลขที่และวันที่เอกสารเสมอ
		payload, _ := json.Marshal(map[string]any{"id": row.id, "balance_side": row.side, "amount": row.amount, "document_no": row.id, "document_date": erpDocDay.Format("2006-01-02")})
		if _, err = db.Exec(`INSERT INTO gl_subledger_documents (company, id, ledger, partner_code, branch_code, control_account_code, amount, side, version, created_journal_id, payload)
			VALUES ($1, $2, 'ap', 'S101', $3, $4, $5, $6, 1, $7, $8)`, testCompany, row.id, branch, control, row.amount, row.side, journalID, payload); err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(`INSERT INTO gl_subledger_allocations (company, id, document_id, journal_id, line_number, amount, payload, created_at, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, '{}', now(), 'integration-test')`, testCompany, "ALLOC-"+row.id, row.id, journalID, i+1, row.amount); err != nil {
			t.Fatal(err)
		}
	}
	payFor := func(ref string, amount string) (*ErpTransactionDoc, error) {
		return store.SaveDoc(ctx, db, scope, pay, &ErpTransactionDoc{DocDateTime: erpDocDay.Format(time.RFC3339), CustCode: "S101", TotalAmount: amt(amount),
			Details: []ErpDetailItem{{LineNumber: 1, DebtDocumentID: ref, DocRef: ref, SumAmount: amt(amount)}}}, false)
	}
	settlements := func(id string) string {
		return scalarString(t, db, `SELECT count(*)::text FROM gl_subledger_settlements WHERE company=$1 AND (debt_id=$2 OR payment_id=$2)`, testCompany, id)
	}

	// เอกสารฝั่งจ่ายของจอบัญชีไม่ใช่หนี้ที่ใบจ่ายตัดได้ — บรรทัดที่อ้างถึงถูกปฏิเสธทั้งใบ
	if _, err = payFor("PMT-GL-001", "100"); userErrorCode(err) != "ap_pay_settle_mismatch" {
		t.Fatalf("pay against a GL payment document = %v, want ap_pay_settle_mismatch", err)
	}
	if n := settlements("PMT-GL-001"); n != "0" {
		t.Fatalf("settlements on a GL payment document = %s, want 0", n)
	}
	requireDecimal(t, "GL payment document untouched", subledgerOutstanding(t, db, "PMT-GL-001"), "300")

	if _, err = payFor("INV-GL-001", "200"); err != nil {
		t.Fatalf("pay against a GL debt document: %v", err)
	}
	if n := settlements("INV-GL-001"); n != "1" {
		t.Fatalf("settlements on a GL debt document = %s, want 1", n)
	}
	requireDecimal(t, "GL debt after payment", subledgerOutstanding(t, db, "INV-GL-001"), "300")

	_, err = payFor("INV-GL-001", "300.01")
	if user, ok := gl.AsUserError(err); !ok || user.Code != "erp_settle_exceeds_remaining" || len(user.Args) != 2 || user.Args[1] != "300.00" {
		t.Fatalf("pay over a GL debt remainder = %v, want erp_settle_exceeds_remaining with 300.00", err)
	}
	requireDecimal(t, "GL debt after refused payment", subledgerOutstanding(t, db, "INV-GL-001"), "300")
}
