package httpapi

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
)

// statementRecorder จดทุกคำสั่ง SQL ที่ส่งถึงฐานข้อมูล เพื่อยืนยันว่าเอกสารที่ไม่ลงบัญชี/ไม่ตัดสต็อกไม่แตะตารางบัญชีและสต็อกเลย
type statementRecorder struct {
	mu         sync.Mutex
	statements []string
}

func (r *statementRecorder) add(query string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statements = append(r.statements, query)
}

func (r *statementRecorder) touched(fragment string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, q := range r.statements {
		if strings.Contains(q, fragment) {
			return true
		}
	}
	return false
}

type recordingConnector struct {
	inner driver.Driver
	dsn   string
	rec   *statementRecorder
}

func (c recordingConnector) Connect(context.Context) (driver.Conn, error) {
	conn, err := c.inner.Open(c.dsn)
	if err != nil {
		return nil, err
	}
	return recordingConn{Conn: conn, rec: c.rec}, nil
}

func (c recordingConnector) Driver() driver.Driver { return c.inner }

type recordingConn struct {
	driver.Conn
	rec *statementRecorder
}

func (c recordingConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return c.Conn.(driver.ConnBeginTx).BeginTx(ctx, opts)
}

func (c recordingConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.rec.add(query)
	return c.Conn.(driver.ExecerContext).ExecContext(ctx, query, args)
}

func (c recordingConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	c.rec.add(query)
	return c.Conn.(driver.QueryerContext).QueryContext(ctx, query, args)
}

func newRecordedMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *statementRecorder) {
	t.Helper()
	dsn := "transaction_" + t.Name()
	mockDB, mock, err := sqlmock.NewWithDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}
	rec := &statementRecorder{}
	db := sql.OpenDB(recordingConnector{inner: mockDB.Driver(), dsn: dsn, rec: rec})
	t.Cleanup(func() {
		db.Close()
		mockDB.Close()
	})
	return db, mock, rec
}

func withAny(fixed []driver.Value, anyCount int) []driver.Value {
	args := append([]driver.Value(nil), fixed...)
	for i := 0; i < anyCount; i++ {
		args = append(args, sqlmock.AnyArg())
	}
	return args
}

const (
	testCompany = "01"
	testBranch  = "00000"
	testActor   = "demo"
)

func testScope() gl.Scope {
	return gl.Scope{Holding: "rungrueng", Company: testCompany, Branch: testBranch, Actor: testActor}
}

func userErrorCode(err error) string {
	if user, ok := gl.AsUserError(err); ok {
		return user.Code
	}
	return ""
}

func newTestDoc() *ErpTransactionDoc {
	return &ErpTransactionDoc{
		DocDateTime: "2026-10-08",
		CustCode:    "C101",
		TotalAmount: 5350,
		Details: []ErpDetailItem{{
			ItemCode: "P-1001", UnitCode: "PCS", Qty: 1, Price: 5000, SumAmount: 5000, WhCode: "WH01",
		}},
	}
}

// expectCreate คาดคำสั่งของการสร้างเอกสารใหม่ที่ไม่ลงบัญชี/ไม่ตัดสต็อก: ออกเลขที่ ตรวจเลขซ้ำ หัวเอกสาร และรายการ 1 บรรทัด
func expectCreate(mock sqlmock.Sqlmock, flag int) {
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT docno FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND docno LIKE \$3`).
		WithArgs(testCompany, flag, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"docno"}))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND docno = \$3\)`).
		WithArgs(testCompany, flag, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`INSERT INTO doc \(`).
		WithArgs(withAny([]driver.Value{testCompany, flag}, 17)...).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectExec(`INSERT INTO docdetail \(`).
		WithArgs(withAny([]driver.Value{testCompany, flag}, 13)...).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func firstDocNoPattern(prefix string) *regexp.Regexp {
	return regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `\d{4}-0001$`)
}

func TestParseTransFlagFollowsRegistry(t *testing.T) {
	for _, dt := range doctype.All() {
		flag, prefix, err := parseTransFlag(" " + strings.ToUpper(dt.Key) + " ")
		if err != nil || flag != dt.Flag || prefix != dt.Prefix {
			t.Errorf("parseTransFlag(%q) = %d,%q,%v; want %d,%q", dt.Key, flag, prefix, err, dt.Flag, dt.Prefix)
		}
	}
	for _, key := range []string{"", "doc", "receipt", "payment", "bank", "debitnote", "stock-prickup-product",
		"etax", "withholdingtax", "purchasereceiptadjustment", "petty-cash-limit"} {
		if flag, _, err := parseTransFlag(key); err == nil {
			t.Errorf("parseTransFlag(%q) = %d; want error", key, flag)
		}
	}
}

// ลงบัญชีและตัดสต็อกตามทะเบียนประเภทเอกสาร (doctype) ที่เดียว: ลงบัญชี 7 ประเภท เคลื่อนไหวสต็อกทุกประเภทที่ทะเบียนระบุรับเข้า/จ่ายออก
// ประเภทที่ปรับสต็อกตามบรรทัด (โอนคลัง ปรับปรุงสต็อก) ยังไม่ตัด — ทำเมื่อลุงจืดสั่ง
func TestSideEffectsFollowRegistry(t *testing.T) {
	posting := map[string]bool{"purchase": true, "purchase-return": true, "expense": true, "paid": true, "pay": true,
		"sale-invoice": true, "sale-invoice-return": true}
	stock := map[string]bool{"stock-balance": true, "stock-receive-product": true, "stock-pickup-product": true,
		"stock-return-product": true, "goodsreceive": true, "purchase": true, "purchase-return": true,
		"sale-invoice": true, "sale-invoice-return": true}
	for _, dt := range doctype.All() {
		if got := doctype.PostsToGL(dt.Flag); got != posting[dt.Key] {
			t.Errorf("%s: posts to GL = %v, want %v", dt.Key, got, posting[dt.Key])
		}
		if got := movesStock(dt); got != stock[dt.Key] {
			t.Errorf("%s: moves stock = %v, want %v", dt.Key, got, stock[dt.Key])
		}
		if got, want := serialized(dt), posting[dt.Key] || stock[dt.Key]; got != want {
			t.Errorf("%s: serialized = %v, want %v", dt.Key, got, want)
		}
	}
}

// ประเภทที่ไม่ลงบัญชีและไม่ตัดสต็อกบันทึกได้โดยไม่แตะตารางบัญชี/สต็อก และไม่ต้องล็อกบริษัท
func TestSaveDocStoreOnlyTypesTouchNeitherGLNorStock(t *testing.T) {
	for _, dt := range doctype.All() {
		dt := dt
		if serialized(dt) {
			continue
		}
		t.Run(dt.Key, func(t *testing.T) {
			db, mock, rec := newRecordedMock(t)
			expectCreate(mock, dt.Flag)
			mock.ExpectCommit()

			saved, err := NewStore().SaveDoc(context.Background(), db, testScope(), dt, newTestDoc(), false)
			if err != nil {
				t.Fatal(err)
			}
			if !firstDocNoPattern(dt.Prefix).MatchString(saved.DocNo) || saved.TransFlag != dt.Flag {
				t.Errorf("saved docno %q flag %d; want prefix %q flag %d", saved.DocNo, saved.TransFlag, dt.Prefix, dt.Flag)
			}
			if saved.GLStatus != "" || saved.GLDocNo != "" {
				t.Errorf("a store-only document must not report a posting result, got %q %q", saved.GLStatus, saved.GLDocNo)
			}
			if rec.touched("gl_") || rec.touched("stock_") || rec.touched("erp_doc_journals") || rec.touched("pg_advisory") {
				t.Error("store-only document touched ledger, stock or lock statements")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSaveDocRejectsReusedDocNoIncludingDeleted(t *testing.T) {
	dt, _ := doctype.ByKey("cashdeposit")
	db, mock, _ := newRecordedMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND docno = \$3\)`).
		WithArgs(testCompany, dt.Flag, "CDP6910-0001").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	doc := newTestDoc()
	doc.DocNo = "CDP6910-0001"
	_, err := NewStore().SaveDoc(context.Background(), db, testScope(), dt, doc, false)
	user, ok := gl.AsUserError(err)
	if !ok || user.Code != "erp_doc_duplicate" || len(user.Args) != 1 || user.Args[0] != "CDP6910-0001" {
		t.Fatalf("SaveDoc error = %v, want erp_doc_duplicate with the docno", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestNormalizeLinesGivesEveryLineAUniquePositiveNumber(t *testing.T) {
	details := []ErpDetailItem{{LineNumber: 0}, {LineNumber: 2}, {LineNumber: 2}, {LineNumber: -4}, {LineNumber: 1}}
	normalizeLines(details)
	seen := map[int]bool{}
	for i, d := range details {
		if d.LineNumber <= 0 || seen[d.LineNumber] {
			t.Fatalf("line %d got number %d (all: %+v)", i, d.LineNumber, details)
		}
		seen[d.LineNumber] = true
	}
	if details[1].LineNumber != 2 {
		t.Errorf("an unused explicit number must be kept, got %d", details[1].LineNumber)
	}
}

func TestSaveDocInfersReferenceTypeFromDocNo(t *testing.T) {
	cases := []struct {
		key, ref string
		refFlag  int
	}{
		{"apdebitnote", "PO6910-0007", doctype.PurchaseOrder},
		{"apcreditnote", "PUR6910-0003", doctype.Purchase},
		{"cashdeposit", "PRET6910-0002", doctype.PurchaseReturn},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			dt, ok := doctype.ByKey(tc.key)
			if !ok {
				t.Fatalf("key %q not registered", tc.key)
			}
			if serialized(dt) {
				t.Fatalf("%s posts or moves stock; this test covers store-only types", tc.key)
			}
			db, mock, rec := newRecordedMock(t)
			expectCreate(mock, dt.Flag)
			mock.ExpectExec(`DELETE FROM docref WHERE businesscode=\$1 AND docno=\$2`).
				WithArgs(testCompany, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectExec(`INSERT INTO docref`).
				WithArgs(testCompany, sqlmock.AnyArg(), dt.Flag, tc.ref, tc.refFlag).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec(`UPDATE doc SET isref = true, isclosed = true`).
				WithArgs(sqlmock.AnyArg(), testCompany, tc.ref).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			doc := newTestDoc()
			doc.DocRef = tc.ref
			if _, err := NewStore().SaveDoc(context.Background(), db, testScope(), dt, doc, false); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
			if rec.touched("gl_") || rec.touched("stock_") {
				t.Error("referenced document must still be store-only")
			}
		})
	}
}

// วงจร CRUD ของเอกสารประเภทที่แยกออกมาใหม่ ทุกคำสั่งต้องกรองด้วยเลขประเภทของตัวเอง
func TestCashDepositCRUDUsesItsOwnFlagAndPrefix(t *testing.T) {
	dt, ok := doctype.ByKey("cashdeposit")
	if !ok {
		t.Fatal("cashdeposit not registered")
	}
	ctx := context.Background()
	store := NewStore()
	db, mock, rec := newRecordedMock(t)

	// Create
	expectCreate(mock, dt.Flag)
	mock.ExpectCommit()
	saved, err := store.SaveDoc(ctx, db, testScope(), dt, newTestDoc(), false)
	if err != nil {
		t.Fatal(err)
	}
	docNo := saved.DocNo
	if !firstDocNoPattern("CDP").MatchString(docNo) {
		t.Fatalf("docno %q must use the CDP prefix", docNo)
	}
	if saved.ID != "7" {
		t.Errorf("create response must carry the new row id, got %q", saved.ID)
	}

	// Read
	stamp := time.Date(2026, 10, 8, 2, 0, 0, 0, time.UTC)
	docColumns := []string{"id", "docno", "docdatetime", "transflag", "custcode", "totalamount", "totalvalue_doc",
		"totaldiscount_doc", "totalvatvalue_doc", "totalbeforevat_doc", "totalaftervat_doc", "iscancel", "isclosed",
		"branchid", "custname", "description", "remark", "created_at", "modified_at",
		"approval_status", "approved_by", "approved_at", "cancelled_by", "cancelled_at", "cancelreason"}
	docRow := func() *sqlmock.Rows {
		return sqlmock.NewRows(docColumns).AddRow(7, docNo, "2026-10-08T00:00:00Z", dt.Flag, "C101",
			5350.0, 5000.0, 0.0, 350.0, 5000.0, 5350.0, false, false, testBranch, "บริษัท รุ่งเรืองก่อสร้าง จำกัด", "ซื้อปูนซีเมนต์", "ส่งภายในสัปดาห์นี้", stamp, stamp,
			"", "", nil, "", nil, "")
	}
	mock.ExpectQuery(`FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND NOT COALESCE\(isdelete, false\) AND docno = \$3 LIMIT 1`).
		WithArgs(testCompany, dt.Flag, docNo).WillReturnRows(docRow())
	mock.ExpectQuery(`SELECT docnoref, docnoreftransflag FROM docref`).
		WithArgs(testCompany, docNo).WillReturnRows(sqlmock.NewRows([]string{"docnoref", "docnoreftransflag"}))
	mock.ExpectQuery(`FROM docdetail WHERE businesscode = \$1 AND transflag = \$2 AND docno = \$3`).
		WithArgs(testCompany, dt.Flag, docNo).
		WillReturnRows(sqlmock.NewRows([]string{"id", "linenumber", "itemcode", "barcode", "unitcode", "totalqty",
			"price", "sumamount", "description", "whcode", "locationcode", "docref"}).
			AddRow(1, 1, "P-1001", "", "PCS", 1.0, 5000.0, 5000.0, "", "WH01", "", ""))
	got, err := store.GetDoc(ctx, db, testCompany, dt.Flag, docNo)
	if err != nil || got == nil {
		t.Fatalf("GetDoc = %+v, %v", got, err)
	}
	if got.DocNo != docNo || got.TransFlag != dt.Flag || len(got.Details) != 1 || got.Details[0].ItemCode != "P-1001" {
		t.Errorf("GetDoc returned %+v", got)
	}
	if got.CustName != "บริษัท รุ่งเรืองก่อสร้าง จำกัด" || got.Description != "ซื้อปูนซีเมนต์" || got.Remark != "ส่งภายในสัปดาห์นี้" {
		t.Errorf("GetDoc must return custname/description/remark separately, got %q / %q / %q", got.CustName, got.Description, got.Remark)
	}

	// List
	mock.ExpectQuery(`SELECT count\(\*\) FROM doc WHERE businesscode = \$1 AND transflag = \$2`).
		WithArgs(testCompany, dt.Flag, "%CDP%").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`FROM doc WHERE businesscode = \$1 AND transflag = \$2 .* LIMIT \$4 OFFSET \$5`).
		WithArgs(testCompany, dt.Flag, "%CDP%", 20, 0).WillReturnRows(docRow())
	list, total, err := store.ListDocs(ctx, db, testCompany, dt.Flag, "CDP", "", 0, 20)
	if err != nil || total != 1 || len(list) != 1 || list[0].DocNo != docNo {
		t.Fatalf("ListDocs = %+v, %d, %v", list, total, err)
	}

	// Update
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\)\s+FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND docno = \$3 AND NOT COALESCE\(isdelete, false\) FOR UPDATE`).
		WithArgs(testCompany, dt.Flag, docNo).WillReturnRows(sqlmock.NewRows([]string{"iscancel", "approval_status"}).AddRow(false, ""))
	mock.ExpectQuery(`UPDATE doc SET docdatetime = \$1`).
		WithArgs(append(withAny(nil, 14), testCompany, dt.Flag, docNo)...).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectExec(`DELETE FROM docdetail WHERE businesscode = \$1 AND transflag = \$2 AND docno = \$3`).
		WithArgs(testCompany, dt.Flag, docNo).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO docdetail \(`).
		WithArgs(withAny([]driver.Value{testCompany, dt.Flag, docNo}, 12)...).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	edit := newTestDoc()
	edit.DocNo = docNo
	edit.Details[0].Price = 5200
	if _, err := store.SaveDoc(ctx, db, testScope(), dt, edit, true); err != nil {
		t.Fatal(err)
	}

	// Delete (soft delete)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\) FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND NOT COALESCE\(isdelete, false\) AND docno = \$3 FOR UPDATE`).
		WithArgs(testCompany, dt.Flag, docNo).WillReturnRows(sqlmock.NewRows([]string{"docno", "iscancel", "approval_status"}).AddRow(docNo, false, ""))
	mock.ExpectQuery(`SELECT EXISTS\(\s+SELECT 1 FROM docref r`).
		WithArgs(testCompany, docNo, dt.Flag).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(`UPDATE doc SET isdelete = true, modifier_code = \$1, modified_at = \$2\s+WHERE businesscode = \$3 AND transflag = \$4 AND docno = \$5`).
		WithArgs(testActor, sqlmock.AnyArg(), testCompany, dt.Flag, docNo).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := store.DeleteDoc(ctx, db, testScope(), dt, docNo); err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
	if rec.touched("gl_") || rec.touched("stock_") {
		t.Error("cash deposit CRUD must not touch GL or stock tables")
	}
}
