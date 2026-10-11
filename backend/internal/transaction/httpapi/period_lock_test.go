package httpapi

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
)

// วันที่ของเอกสารทดสอบ: 2026-10-08 00:00 UTC = 07:00 ตามเวลาไทย จึงเป็นวันทำการ 2026-10-08
const testDocDay = "2026-10-08"

var testDocDate = time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)

// expectDateOpen คาดการตรวจงวดของวันที่หนึ่งวัน (gl.CheckSourceDateOpen): ปีบัญชีที่ครอบคลุม → ปีบัญชีเปิดอยู่ → งวดล็อกหรือไม่
func expectDateOpen(mock sqlmock.Sqlmock, day string, locked bool) {
	mock.ExpectQuery(`SELECT code FROM gl_records WHERE company=\$1 AND kind='fiscal-years'`).
		WithArgs(testCompany, day).WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow("2569"))
	mock.ExpectQuery(`FROM gl_records WHERE company=\$1 AND kind='fiscal-years' AND code=\$2`).
		WithArgs(testCompany, "2569").
		WillReturnRows(sqlmock.NewRows([]string{"payload"}).AddRow([]byte(`{"code":"2569","isactive":true,"closed":false}`)))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM gl_records WHERE company=\$1 AND kind='periods'`).
		WithArgs(testCompany, day).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(locked))
}

func requirePeriodLocked(t *testing.T, err error, day string) {
	t.Helper()
	user, ok := gl.AsUserError(err)
	if !ok || user.Code != "source_period_locked" || len(user.Args) != 1 || user.Args[0] != day {
		t.Fatalf("error = %v, want source_period_locked for %s", err, day)
	}
	if !strings.Contains(user.Message, day) || !strings.Contains(user.Message, "ถูกล็อก") {
		t.Errorf("period-locked message must be the Thai text with the date, got %q", user.Message)
	}
}

func expectStateRow(mock sqlmock.Sqlmock, flag int, docNo string, cancelled bool, approval string) {
	mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\), docdatetime FROM doc .* FOR UPDATE`).
		WithArgs(testCompany, flag, docNo).
		WillReturnRows(sqlmock.NewRows([]string{"docno", "iscancel", "approval_status", "docdatetime"}).AddRow(docNo, cancelled, approval, testDocDate))
}

func expectNoDownstream(mock sqlmock.Sqlmock, flag int, docNo string) {
	mock.ExpectQuery(`SELECT EXISTS\(\s+SELECT 1 FROM docref r`).
		WithArgs(testCompany, docNo, flag).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
}

// expectGetDoc คาดการอ่านเอกสารกลับหลังบันทึก (หัว + การอ้างอิง + รายการ)
func expectGetDoc(mock sqlmock.Sqlmock, flag int, docNo string, cancelled bool, approval, reason string) {
	stamp := time.Date(2026, 10, 8, 2, 0, 0, 0, time.UTC)
	docColumns := []string{"id", "docno", "docdatetime", "transflag", "custcode", "totalamount", "totalvalue_doc",
		"totaldiscount_doc", "totalvatvalue_doc", "totalbeforevat_doc", "totalaftervat_doc", "iscancel", "isclosed",
		"branchid", "custname", "description", "remark", "created_at", "modified_at",
		"approval_status", "approved_by", "approved_at", "cancelled_by", "cancelled_at", "cancelreason"}
	var cancelledBy string
	var cancelledAt any
	if cancelled {
		cancelledBy, cancelledAt = testActor, stamp
	}
	mock.ExpectQuery(`FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND NOT COALESCE\(isdelete, false\) AND docno = \$3 LIMIT 1`).
		WithArgs(testCompany, flag, docNo).
		WillReturnRows(sqlmock.NewRows(docColumns).AddRow(1, docNo, "2026-10-08T00:00:00Z", flag, "V101",
			1070.0, 1000.0, 0.0, 70.0, 1000.0, 1070.0, cancelled, false, testBranch, "ห้างหุ้นส่วนจำกัด ศรีสยามวัสดุ", "ขอซื้อวัสดุสำนักงาน", "", stamp, stamp,
			approval, "", nil, cancelledBy, cancelledAt, reason))
	mock.ExpectQuery(`SELECT docnoref, docnoreftransflag FROM docref`).
		WithArgs(testCompany, docNo).WillReturnRows(sqlmock.NewRows([]string{"docnoref", "docnoreftransflag"}))
	mock.ExpectQuery(`FROM docdetail WHERE businesscode = \$1 AND transflag = \$2 AND docno = \$3`).
		WithArgs(testCompany, flag, docNo).
		WillReturnRows(sqlmock.NewRows([]string{"id", "linenumber", "itemcode", "barcode", "unitcode", "totalqty",
			"price", "sumamount", "description", "whcode", "locationcode", "docref"}))
}

// เทียบเวลาด้วย Equal (ขอบวันตามเวลาไทยไม่ขึ้นกับ *time.Location ที่ใช้สร้าง)
type timeArg struct{ want time.Time }

func (a timeArg) Match(v driver.Value) bool {
	got, ok := v.(time.Time)
	return ok && got.Equal(a.want)
}

// งวดล็อก (Champ IsPeriodClose) ต้องหยุดทุกคำสั่งเขียนก่อนเขียนอะไรลงฐาน ทุกประเภทเอกสาร ไม่ใช่เฉพาะที่ลงบัญชี
func TestSaveDocCreateStopsInLockedPeriod(t *testing.T) {
	dt := mustDocType(t, "cashdeposit")
	db, mock, _ := newRecordedMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT docno FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND docno LIKE \$3`).
		WithArgs(testCompany, dt.Flag, sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"docno"}))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND docno = \$3\)`).
		WithArgs(testCompany, dt.Flag, sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	expectDateOpen(mock, testDocDay, true)
	mock.ExpectRollback()

	_, err := NewStore().SaveDoc(context.Background(), db, testScope(), dt, newTestDoc(), false)
	requirePeriodLocked(t, err, testDocDay)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// แก้ไขเอกสาร = ย้ายออกจากวันที่เดิมไปวันที่ใหม่: งวดของวันเดิมหรือวันใหม่ล็อกอย่างใดอย่างหนึ่ง = ห้ามแก้
func TestSaveDocEditChecksStoredAndNewDates(t *testing.T) {
	dt := mustDocType(t, "cashdeposit")
	const docNo = "CDP6910-0001"
	storedDate := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name         string
		storedLocked bool
		newLocked    bool
		lockedDay    string
	}{
		{"stored date locked", true, false, "2026-09-30"},
		{"new date locked", false, true, testDocDay},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, _ := newRecordedMock(t)
			mock.ExpectBegin()
			mock.ExpectQuery(`SELECT COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\), docdatetime\s+FROM doc WHERE`).
				WithArgs(testCompany, dt.Flag, docNo).
				WillReturnRows(sqlmock.NewRows([]string{"iscancel", "approval_status", "docdatetime"}).AddRow(false, "", storedDate))
			expectDateOpen(mock, "2026-09-30", tc.storedLocked)
			if !tc.storedLocked {
				expectDateOpen(mock, testDocDay, tc.newLocked)
			}
			mock.ExpectRollback()

			doc := newTestDoc()
			doc.DocNo = docNo
			_, err := NewStore().SaveDoc(context.Background(), db, testScope(), dt, doc, true)
			requirePeriodLocked(t, err, tc.lockedDay)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDeleteDocStopsInLockedPeriod(t *testing.T) {
	dt := mustDocType(t, "cashdeposit")
	const docNo = "CDP6910-0002"
	db, mock, _ := newRecordedMock(t)
	mock.ExpectBegin()
	expectStateRow(mock, dt.Flag, docNo, false, "")
	expectNoDownstream(mock, dt.Flag, docNo)
	expectDateOpen(mock, testDocDay, true)
	mock.ExpectRollback()

	err := NewStore().DeleteDoc(context.Background(), db, testScope(), dt, docNo)
	requirePeriodLocked(t, err, testDocDay)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestApproveUnapproveCancelStopInLockedPeriod(t *testing.T) {
	ctx := context.Background()
	pr := mustDocType(t, "purchase-requisition")
	po := mustDocType(t, "purchase-order")
	store := NewStore()

	t.Run("approve", func(t *testing.T) {
		db, mock, _ := newRecordedMock(t)
		mock.ExpectBegin()
		expectStateRow(mock, pr.Flag, "PR6910-0001", false, "")
		expectDateOpen(mock, testDocDay, true)
		mock.ExpectRollback()
		_, err := store.ApproveDoc(ctx, db, testScope(), pr, "PR6910-0001")
		requirePeriodLocked(t, err, testDocDay)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("unapprove", func(t *testing.T) {
		db, mock, _ := newRecordedMock(t)
		mock.ExpectBegin()
		expectStateRow(mock, pr.Flag, "PR6910-0002", false, "approved")
		expectNoDownstream(mock, pr.Flag, "PR6910-0002")
		expectDateOpen(mock, testDocDay, true)
		mock.ExpectRollback()
		_, err := store.UnapproveDoc(ctx, db, testScope(), pr, "PR6910-0002", "ราคาที่ขอซื้อไม่ถูกต้อง")
		requirePeriodLocked(t, err, testDocDay)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		db, mock, _ := newRecordedMock(t)
		mock.ExpectBegin()
		expectStateRow(mock, po.Flag, "PO6910-0003", false, "approved")
		expectNoDownstream(mock, po.Flag, "PO6910-0003")
		expectDateOpen(mock, testDocDay, true)
		mock.ExpectRollback()
		_, err := store.CancelDoc(ctx, db, testScope(), po, "PO6910-0003", "ผู้ขายแจ้งยกเลิกการส่งสินค้า")
		requirePeriodLocked(t, err, testDocDay)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
}

func TestUnapproveDocReturnsPurchaseRequisitionToPending(t *testing.T) {
	pr := mustDocType(t, "purchase-requisition")
	const docNo = "PR6910-0004"
	db, mock, _ := newRecordedMock(t)
	mock.ExpectBegin()
	expectStateRow(mock, pr.Flag, docNo, false, "approved")
	expectNoDownstream(mock, pr.Flag, docNo)
	expectDateOpen(mock, testDocDay, false)
	mock.ExpectExec(`UPDATE doc SET\s+approval_status = '',\s+approved_by = '',\s+approved_at = NULL`).
		WithArgs(testActor, sqlmock.AnyArg(), testCompany, pr.Flag, docNo).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO erp_doc_events .* VALUES \(\$1, \$2, \$3, 'unapprove', 'approved', '', \$4, \$5, \$6\)`).
		WithArgs(testCompany, pr.Flag, docNo, "ราคาที่ขอซื้อไม่ถูกต้อง", testActor, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	expectGetDoc(mock, pr.Flag, docNo, false, "", "")

	doc, err := NewStore().UnapproveDoc(context.Background(), db, testScope(), pr, docNo, "  ราคาที่ขอซื้อไม่ถูกต้อง  ")
	if err != nil {
		t.Fatalf("UnapproveDoc = %v", err)
	}
	if doc.ApprovalStatus != "" || doc.IsCancel || doc.Status != 0 {
		t.Errorf("unapproved document = status %d approval %q cancel %v; want pending", doc.Status, doc.ApprovalStatus, doc.IsCancel)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCancelPendingPurchaseRequisition(t *testing.T) {
	pr := mustDocType(t, "purchase-requisition")
	const docNo = "PR6910-0005"
	const reason = "ฝ่ายผลิตยกเลิกความต้องการวัสดุ"
	db, mock, rec := newRecordedMock(t)
	mock.ExpectBegin()
	expectStateRow(mock, pr.Flag, docNo, false, "")
	expectNoDownstream(mock, pr.Flag, docNo)
	expectDateOpen(mock, testDocDay, false)
	mock.ExpectExec(`UPDATE doc SET\s+iscancel = true`).
		WithArgs(reason, testActor, sqlmock.AnyArg(), testCompany, pr.Flag, docNo).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO erp_doc_events .* VALUES \(\$1, \$2, \$3, 'cancel', \$4, 'cancelled', \$5, \$6, \$7\)`).
		WithArgs(testCompany, pr.Flag, docNo, "", reason, testActor, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	expectGetDoc(mock, pr.Flag, docNo, true, "", reason)

	doc, err := NewStore().CancelDoc(context.Background(), db, testScope(), pr, docNo, reason)
	if err != nil {
		t.Fatalf("CancelDoc = %v", err)
	}
	if !doc.IsCancel || doc.Status != 2 || doc.CancelReason != reason {
		t.Errorf("cancelled requisition = status %d cancel %v reason %q", doc.Status, doc.IsCancel, doc.CancelReason)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
	if rec.touchedLedger() || rec.touched("stock_") || rec.touched("pg_advisory") {
		t.Error("cancelling a requisition must not touch ledger, stock or the company lock")
	}
}

// คำสั่งหลายใบ: ใบที่งวดล็อกได้ code ของกฎบัญชี, ไม่พบเอกสาร = erp_err_doc_not_found,
// ความผิดของระบบได้ข้อความกลาง erp_err_internal โดยไม่เปิดรายละเอียดภายใน; ใบอื่นทำต่อ
func TestBatchResultsCarryCodedPerRowErrors(t *testing.T) {
	ctx := context.Background()
	pr := mustDocType(t, "purchase-requisition")
	db, mock, _ := newRecordedMock(t)

	mock.ExpectBegin()
	expectStateRow(mock, pr.Flag, "PR6910-0010", false, "")
	expectDateOpen(mock, testDocDay, true)
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\)`).WithArgs(testCompany, pr.Flag, "PR6910-0011").
		WillReturnError(errors.New("pq: connection reset by peer"))
	mock.ExpectRollback()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\)`).WithArgs(testCompany, pr.Flag, "PR6910-0012").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	var logged []string
	results := NewStore().ApproveBatchDocs(ctx, db, testScope(), pr, []string{"PR6910-0010", "  ", "PR6910-0011", "PR6910-0012"}, "en",
		func(line string) { logged = append(logged, line) })
	if len(results) != 3 {
		t.Fatalf("results = %+v, want 3 rows (blank ids skipped)", results)
	}
	want := []struct{ id, code, text string }{
		{"PR6910-0010", "source_period_locked", "The date 2026-10-08 is in a locked accounting period."},
		{"PR6910-0011", "erp_err_internal", "The request could not be completed."},
		{"PR6910-0012", "erp_err_doc_not_found", "The document was not found."},
	}
	for i, w := range want {
		got := results[i]
		if got.ID != w.id || got.Success || got.Code != w.code || !strings.Contains(got.Message, w.text) {
			t.Errorf("row %d = %+v, want id %s code %s message containing %q", i, got, w.id, w.code, w.text)
		}
		if strings.Contains(got.Message, "pq:") || strings.Contains(got.Message, "connection reset") {
			t.Errorf("row %d leaks the internal error: %q", i, got.Message)
		}
	}
	// เฉพาะความผิดของระบบ (PR6910-0011) ลง log ฝั่งเซิร์ฟเวอร์พร้อมรายละเอียดจริง; ความผิดที่ผู้ใช้แก้ได้/ไม่พบเอกสารไม่ลง
	if len(logged) != 1 || !strings.Contains(logged[0], "PR6910-0011") || !strings.Contains(logged[0], "connection reset by peer") {
		t.Errorf("server log = %q, want one line with PR6910-0011 and the internal error", logged)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// runBatch: ความผิดที่ไม่ใช่ของผู้ใช้ยังได้ code erp_err_internal + ข้อความกลาง (สัญญา {id, success, code, message} เดิม)
// และรายละเอียดภายในลง log ฝั่งเซิร์ฟเวอร์เท่านั้น ไม่หลุดถึงผู้ใช้
func TestRunBatchLogsInternalErrorAndKeepsGenericCode(t *testing.T) {
	var logged []string
	results := runBatch([]string{"RE6910-0001", "RE6910-0002"}, "en", func(line string) { logged = append(logged, line) }, func(id string) error {
		if id == "RE6910-0001" {
			return fmt.Errorf("commit tx: %w", errors.New("pq: deadlock detected"))
		}
		return nil
	})
	if len(results) != 2 {
		t.Fatalf("results = %+v, want 2 rows", results)
	}
	failed := results[0]
	if failed.ID != "RE6910-0001" || failed.Success || failed.Code != "erp_err_internal" || failed.Message != langText("erp_err_internal", "en") {
		t.Fatalf("failed row = %+v, want erp_err_internal with the generic message", failed)
	}
	if strings.Contains(failed.Message, "deadlock") {
		t.Fatalf("message leaks the internal error: %q", failed.Message)
	}
	if !results[1].Success || results[1].Code != "" {
		t.Fatalf("second row = %+v, want success", results[1])
	}
	if len(logged) != 1 || !strings.Contains(logged[0], "RE6910-0001") || !strings.Contains(logged[0], "pq: deadlock detected") {
		t.Fatalf("server log = %q, want one line with the document and the internal error", logged)
	}
	// ไม่มี logger (nil) ต้องไม่ panic
	if got := runBatch([]string{"RE6910-0001"}, "th", nil, func(string) error { return errors.New("boom") }); got[0].Code != "erp_err_internal" {
		t.Fatalf("nil logger row = %+v", got[0])
	}
}

// runBatch: ความผิดของระบบที่ห่อด้วย ledgerError (เช่น lock timeout จากเครื่องบัญชี) ตอบเป็น unavailable ข้อความกลาง
// แต่ต้องลง log ด้วย — ตัดสินจากตัว error ไม่ใช่ code ที่ตอบ; ความผิดที่ผู้ใช้แก้ได้และไม่พบเอกสารไม่ลง log
func TestRunBatchLogsWrappedLedgerFailure(t *testing.T) {
	var logged []string
	failures := map[string]error{
		"RE6910-0001": ledgerFail(errors.New("pq: canceling statement due to lock timeout")),
		"RE6910-0002": docLockedError(),
		"RE6910-0003": sql.ErrNoRows,
		"RE6910-0004": ledgerFail(docNotPendingError()),
	}
	results := runBatch([]string{"RE6910-0001", "RE6910-0002", "RE6910-0003", "RE6910-0004"}, "en", func(line string) { logged = append(logged, line) }, func(id string) error {
		return failures[id]
	})
	if len(results) != 4 {
		t.Fatalf("results = %+v, want 4 rows", results)
	}
	if got := results[0]; got.Success || got.Code != "unavailable" || strings.Contains(got.Message, "lock timeout") {
		t.Fatalf("wrapped ledger failure row = %+v, want unavailable with a generic message", got)
	}
	if got := results[1]; got.Code != "erp_doc_locked" {
		t.Fatalf("user error row = %+v, want erp_doc_locked", got)
	}
	if got := results[2]; got.Code != "erp_err_doc_not_found" {
		t.Fatalf("not found row = %+v, want erp_err_doc_not_found", got)
	}
	if got := results[3]; got.Code != "erp_doc_not_pending" {
		t.Fatalf("wrapped user error row = %+v, want erp_doc_not_pending", got)
	}
	if len(logged) != 1 || !strings.Contains(logged[0], "RE6910-0001") || !strings.Contains(logged[0], "lock timeout") {
		t.Fatalf("server log = %q, want exactly one line for the wrapped internal failure", logged)
	}
}

func TestCancelBatchStopsLockedRowInThai(t *testing.T) {
	po := mustDocType(t, "purchase-order")
	db, mock, _ := newRecordedMock(t)
	mock.ExpectBegin()
	expectStateRow(mock, po.Flag, "PO6910-0020", false, "")
	expectNoDownstream(mock, po.Flag, "PO6910-0020")
	expectDateOpen(mock, testDocDay, true)
	mock.ExpectRollback()

	results := NewStore().CancelBatchDocs(context.Background(), db, testScope(), po, []string{"PO6910-0020"}, "สั่งซ้ำ", "th", nil)
	if len(results) != 1 || results[0].Success || results[0].Code != "source_period_locked" ||
		!strings.Contains(results[0].Message, "ถูกล็อก") || !strings.Contains(results[0].Message, testDocDay) {
		t.Fatalf("results = %+v", results)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestSaveDeliveryDatesStopsInLockedPeriod(t *testing.T) {
	si := mustDocType(t, "sale-invoice")
	db, mock, _ := newRecordedMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`pg_advisory_xact_lock`).WithArgs(testCompany).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT docdatetime FROM doc`).WithArgs(testCompany, si.Flag, "INV6910-0001").
		WillReturnRows(sqlmock.NewRows([]string{"docdatetime"}).AddRow(testDocDate))
	expectDateOpen(mock, testDocDay, true)
	mock.ExpectRollback()

	_, err := NewStore().SaveDeliveryDates(context.Background(), db, testScope(), si,
		[]SaveDeliveryDateItem{{DocNo: "INV6910-0001", DeliveryDate: "2026-10-16", CreditDelivery: 30, NewDueDate: "2026-11-15"}})
	requirePeriodLocked(t, err, testDocDay)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestProcessReceiptAdjustmentsStopsInLockedPeriod(t *testing.T) {
	db, mock, _ := newRecordedMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`pg_advisory_xact_lock`).WithArgs(testCompany).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT docdatetime FROM doc`).WithArgs(testCompany, "GR6910-0001", doctype.GoodsReceive).
		WillReturnRows(sqlmock.NewRows([]string{"docdatetime"}).AddRow(testDocDate))
	expectDateOpen(mock, testDocDay, true)
	mock.ExpectRollback()

	_, err := NewStore().ProcessReceiptAdjustments(context.Background(), db, testScope(), []string{"GR6910-0001"})
	requirePeriodLocked(t, err, testDocDay)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// postDocument ตรวจงวดก่อนทุกทางออก: งวดล็อกต้องหยุดก่อนหารูปแบบลงบัญชี; งวดเปิดแต่ไม่มีรูปแบบ = no_mapping
func TestPostDocumentChecksPeriodBeforeMapping(t *testing.T) {
	ctx := context.Background()
	pur := mustDocType(t, "purchase")
	for _, tc := range []struct {
		name   string
		locked bool
	}{{"locked", true}, {"open", false}} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, _ := newRecordedMock(t)
			mock.ExpectBegin()
			expectDateOpen(mock, testDocDay, tc.locked)
			if !tc.locked {
				mock.ExpectQuery(`FROM gl_records WHERE company=\$1 AND kind='mappings'`).
					WithArgs(testCompany, pur.Key).WillReturnError(sql.ErrNoRows)
			}
			mock.ExpectRollback()

			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			result, err := NewStore().postDocument(ctx, tx, testScope(), pur, newTestDoc(), "PUR6910-0001", testDocDate, time.Now())
			_ = tx.Rollback()
			if tc.locked {
				requirePeriodLocked(t, err, testDocDay)
			} else if err != nil || result.status != glStatusNoMapping {
				t.Errorf("open period without mapping = %+v, %v; want no_mapping", result, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

// ค้นรายการตามช่วงวันที่ (รวมวันปลาย ตามปฏิทินไทย) รหัสลูกค้า/ผู้จำหน่ายตรงตัว และสถานะ; คืนผู้บันทึก
// สถานะแต่ละค่าต้องได้เงื่อนไข SQL ตรงตัว (ว่าง = ไม่กรองสถานะ) ทั้งคิวรีนับและคิวรีรายการ
func TestListDocsFiltersByDateRangePartnerAndStatus(t *testing.T) {
	pr := mustDocType(t, "purchase-requisition")
	from, _ := dayStart("2026-10-01")
	toNext, _ := dayStart("2026-11-01")
	filterArgs := []driver.Value{testCompany, pr.Flag, timeArg{from}, timeArg{toNext}, "V101"}
	stamp := time.Date(2026, 10, 8, 2, 0, 0, 0, time.UTC)
	baseWhere := "WHERE businesscode = $1 AND transflag = $2 AND NOT COALESCE(isdelete, false) AND docdatetime >= $3 AND docdatetime < $4 AND custcode = $5"
	for _, tc := range []struct{ status, where string }{
		{"pending", " AND NOT COALESCE(iscancel, false) AND COALESCE(approval_status, '') = ''"},
		{"approved", " AND NOT COALESCE(iscancel, false) AND COALESCE(approval_status, '') = 'approved'"},
		{"cancelled", " AND COALESCE(iscancel, false)"},
		{"active", " AND NOT COALESCE(iscancel, false)"},
		{"", ""},
	} {
		t.Run("status="+tc.status, func(t *testing.T) {
			db, mock, _ := newRecordedMock(t)
			where := baseWhere + tc.where
			mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT count(*) FROM doc "+where) + "$").
				WithArgs(filterArgs...).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery(regexp.QuoteMeta("COALESCE(creator_code, '') FROM doc "+where+" ORDER BY docdatetime DESC, docno DESC LIMIT $6 OFFSET $7") + "$").
				WithArgs(append(filterArgs, 20, 40)...).
				WillReturnRows(sqlmock.NewRows([]string{"id", "docno", "docdatetime", "transflag", "custcode", "totalamount", "totalvalue_doc",
					"totaldiscount_doc", "totalvatvalue_doc", "totalbeforevat_doc", "totalaftervat_doc", "iscancel", "isclosed",
					"branchid", "custname", "description", "remark", "created_at", "modified_at",
					"approval_status", "approved_by", "approved_at", "cancelled_by", "cancelled_at", "cancelreason", "creator_code"}).
					AddRow(3, "PR6910-0003", "2026-10-31T00:00:00Z", pr.Flag, "V101", 1070.0, 1000.0, 0.0, 70.0, 1000.0, 1070.0, false, false,
						testBranch, "ห้างหุ้นส่วนจำกัด ศรีสยามวัสดุ", "", "", stamp, stamp, "", "", nil, "", nil, "", "somchai"))

			docs, total, err := NewStore().ListDocs(context.Background(), db, testCompany, pr.Flag,
				DocListFilter{FromDate: "2026-10-01", ToDate: "2026-10-31", CustCode: " V101 ", Status: tc.status}, 40, 20)
			if err != nil || total != 1 || len(docs) != 1 || docs[0].CreatedBy != "somchai" {
				t.Fatalf("ListDocs = %+v, %d, %v", docs, total, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestListDocsRejectsInvalidDateBeforeQuerying(t *testing.T) {
	for _, tc := range []struct {
		filter DocListFilter
		field  string
	}{
		{DocListFilter{FromDate: "2026-13-01"}, "fromdate"},
		{DocListFilter{ToDate: "31/10/2569"}, "todate"},
	} {
		// db = nil: ถ้าคิวรีก่อนตรวจวันที่จะ panic
		_, _, err := NewStore().ListDocs(context.Background(), nil, testCompany, 2, tc.filter, 0, 20)
		user, ok := gl.AsUserError(err)
		if !ok || user.Code != "erp_list_date_invalid" || user.Field != tc.field || user.HTTPStatus() != http.StatusBadRequest {
			t.Errorf("ListDocs(%+v) error = %v, want erp_list_date_invalid on %s (400)", tc.filter, err, tc.field)
		}
	}
}

// เหตุผลที่ระบบเขียนเอง (update, adjust_status) เก็บเป็น key แล้วแปลตามภาษาของผู้ดู; เหตุผลที่ผู้ใช้พิมพ์ (cancel, unapprove)
// คืนตามเดิมเสมอ แม้พิมพ์ให้ขึ้นต้นด้วย erp_event_ เพื่อปลอมเป็นข้อความระบบ
func TestEventReasonText(t *testing.T) {
	for _, tc := range []struct{ action, reason, lang, want string }{
		{"adjust_status", "erp_event_gr_closed", "th", "ปิดใบรับสินค้า (ปรับปรุงสถานะใบรับสินค้า)"},
		{"adjust_status", "erp_event_gr_closed", "en", "Goods receipt closed (status adjustment)"},
		{"update", "erp_event_due_date_changed:2026-11-15", "en", "Due date changed to 2026-11-15"},
		{"cancel", "ผู้ขายแจ้งยกเลิก", "en", "ผู้ขายแจ้งยกเลิก"},
		{"adjust_status", "erp_event_unknown_key", "th", "erp_event_unknown_key"},
		{"cancel", "erp_event_gr_closed", "th", "erp_event_gr_closed"},
		{"cancel", "erp_event_due_date_changed:2026-01-01", "en", "erp_event_due_date_changed:2026-01-01"},
		{"unapprove", "erp_event_gr_closed", "en", "erp_event_gr_closed"},
		{"approve", "erp_event_gr_closed", "th", "erp_event_gr_closed"},
	} {
		if got := eventReasonText(tc.action, tc.reason, tc.lang); got != tc.want {
			t.Errorf("eventReasonText(%s, %q, %s) = %q, want %q", tc.action, tc.reason, tc.lang, got, tc.want)
		}
	}
}

// ปรับปรุงสถานะใบรับสินค้าเขียนประวัติ action = adjust_status — CHECK ของตารางต้องรับค่านี้ทั้งตอนสร้างและตอนปรับตารางเดิม
func TestEventActionCheckAllowsAdjustStatus(t *testing.T) {
	if got := strings.Count(docSchema, "'approve','unapprove','cancel','update','adjust_status'"); got != 2 {
		t.Fatalf("erp_doc_events action CHECK lists adjust_status %d times, want 2 (CREATE TABLE and ALTER)", got)
	}
}
