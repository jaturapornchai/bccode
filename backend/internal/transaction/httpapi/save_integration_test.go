//go:build integration

package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	gl "smlcloudplatform/internal/generalledger"
)

// ทดสอบกับ PostgreSQL จริง: BC_GL_TEST_POSTGRES_DSN ชี้ไปฐานที่ทิ้งได้ (สร้าง schema ชั่วคราวต่อเทสต์)
const erpHolding = "H"

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
	year := gl.FiscalYear{Identity: erpIdentity("year-2026"), Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true, Scale: 2, RetainedEarningsAccount: "3000"}
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

func purchaseDoc(docNo string, qty, price float64) *ErpTransactionDoc {
	goods := qty * price
	return &ErpTransactionDoc{
		DocNo:          docNo,
		DocDateTime:    time.Now().UTC().Format(time.RFC3339),
		CustCode:       "S101",
		CustName:       "บริษัท สยามวัสดุภัณฑ์ จำกัด",
		TotalBeforeVat: goods,
		TotalVatValue:  goods * 0.07,
		TotalAmount:    goods * 1.07,
		Details:        []ErpDetailItem{{ItemCode: "ITM-001", Qty: qty, Price: price, SumAmount: goods, WhCode: "WH01"}},
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
	saved, err := store.SaveDoc(ctx, db, scope, dt, purchaseDoc("", 10, 100), false)
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
	edit := purchaseDoc(docNo, 20, 100)
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
	if _, err = store.SaveDoc(ctx, db, scope, dt, purchaseDoc(docNo, 1, 100), false); userErrorCode(err) != "erp_doc_duplicate" {
		t.Fatalf("reuse of deleted docno = %v, want erp_doc_duplicate", err)
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

	bought, err := store.SaveDoc(ctx, db, scope, purchase, purchaseDoc("", 10, 100), false)
	if err != nil {
		t.Fatalf("purchase: %v", err)
	}
	payment := &ErpTransactionDoc{DocDateTime: time.Now().UTC().Format(time.RFC3339), CustCode: "S101", DocRef: bought.DocNo, TotalAmount: 1070}
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
	if _, err = store.SaveDoc(ctx, db, scope, purchase, purchaseDoc(bought.DocNo, 5, 100), true); userErrorCode(err) != "erp_doc_settled" {
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
	instant := time.Date(2026, 10, 9, 9, 0, 0, 0, time.UTC).Format(time.RFC3339)
	stockIn := purchaseDoc("", 10, 100)
	stockIn.DocDateTime = instant
	if _, err := store.SaveDoc(ctx, db, scope, mustDocType(t, "purchase"), stockIn, false); err != nil {
		t.Fatalf("stock-in purchase: %v", err)
	}
	sale := &ErpTransactionDoc{DocDateTime: instant, CustCode: "C101", TotalAmount: 321, TotalBeforeVat: 300, TotalVatValue: 21,
		Details: []ErpDetailItem{{ItemCode: "ITM-001", Qty: 3, Price: 100, SumAmount: 300, WhCode: "WH01"}}}
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
	same := purchaseDoc(scalarString(t, db, `SELECT docno FROM stock_ledger WHERE businesscode=$1 AND docno LIKE 'PUR%'`, testCompany), 12, 100)
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

	day := func(d int) string { return time.Date(2026, 9, d, 8, 0, 0, 0, time.UTC).Format(time.RFC3339) }
	line := func(qty, price float64) []ErpDetailItem {
		return []ErpDetailItem{{ItemCode: "ITM-002", Qty: qty, Price: price, SumAmount: qty * price, WhCode: "WH01"}}
	}
	// วันที่ 10 รับ 10 @100, วันที่ 20 เบิก 4 → คงเหลือ 6
	if _, err := store.SaveDoc(ctx, db, scope, receive, &ErpTransactionDoc{DocNo: "RCV-1", DocDateTime: day(10), Details: line(10, 100)}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveDoc(ctx, db, scope, pickup, &ErpTransactionDoc{DocNo: "PCK-1", DocDateTime: day(20), Details: line(4, 0)}, false); err != nil {
		t.Fatal(err)
	}
	requireDecimal(t, "balance after pickup", scalarString(t, db, `SELECT balanceqty::text FROM stock_ledger WHERE businesscode=$1 AND docno='PCK-1'`, testCompany), "6")
	// ย้อนหลังวันที่ 15 รับเพิ่ม 5 @120 → ยอดของใบเบิกวันที่ 20 ต้องคำนวณใหม่เป็น 15 - 4 = 11
	if _, err := store.SaveDoc(ctx, db, scope, receive, &ErpTransactionDoc{DocNo: "RCV-2", DocDateTime: day(15), Details: line(5, 120)}, false); err != nil {
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
	instant := time.Date(2026, 10, 8, 9, 0, 0, 0, time.UTC).Format(time.RFC3339)

	saved, err := store.SaveDoc(ctx, db, scope, pr, &ErpTransactionDoc{DocDateTime: instant, CustCode: "S101", CustName: "บริษัท สยามวัสดุภัณฑ์ จำกัด",
		Description: "ขอซื้อกระดาษถ่ายเอกสาร A4", TotalBeforeVat: 1000, TotalVatValue: 70, TotalAmount: 1070,
		Details: []ErpDetailItem{{ItemCode: "ITM-001", Qty: 10, Price: 100, SumAmount: 1000, WhCode: "WH01"}}}, false)
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
		TotalBeforeVat: 1000, TotalVatValue: 70, TotalAmount: 1070,
		Details: []ErpDetailItem{{ItemCode: "ITM-001", Qty: 10, Price: 100, SumAmount: 1000, WhCode: "WH01"}}}, false)
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
