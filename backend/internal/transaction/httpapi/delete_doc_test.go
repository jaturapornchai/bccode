package httpapi

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"smlcloudplatform/internal/doctype"
)

// เคยส่ง args ที่ตัดท้ายสองตัวให้ SELECT docno ทำให้คิวรีล้มเงียบ ๆ แล้วไม่ลบ stock_ledger ของเอกสารที่ถูกลบ
// เอกสารซื้อต้องเดินครบลำดับ: ล็อกบริษัท → หาเลขที่ → เอาบัญชีย่อยออก → กลับรายการใบสำคัญ → ถอนสต็อก → ตั้งธงลบ
func TestDeleteDocReleasesLedgerAndStockForBothLookupForms(t *testing.T) {
	const docNo = "PO2610-0001"
	dt := mustDocType(t, "purchase")
	for _, tc := range []struct {
		name       string
		docIDOrNo  string
		selectArgs []driver.Value
	}{
		{"by docno", docNo, []driver.Value{testCompany, dt.Flag, docNo}},
		{"by numeric id", "7", []driver.Value{testCompany, dt.Flag, 7, "7"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, _ := newRecordedMock(t)
			mock.ExpectBegin()
			mock.ExpectExec(`pg_advisory_xact_lock`).WithArgs(testCompany).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\), docdatetime FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND`).
				WithArgs(tc.selectArgs...).
				WillReturnRows(sqlmock.NewRows([]string{"docno", "iscancel", "approval_status", "docdatetime"}).AddRow(docNo, false, "", testDocDate))
			mock.ExpectQuery(`SELECT EXISTS\(\s+SELECT 1 FROM docref r`).
				WithArgs(testCompany, docNo, dt.Flag).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			expectDateOpen(mock, testDocDay, false)
			mock.ExpectQuery(`FROM gl_subledger_settlements`).
				WithArgs(testCompany, "AP-"+docNo, "", "").WillReturnRows(sqlmock.NewRows([]string{"bound"}).AddRow(0))
			expectNotReserved(mock, "AP-"+docNo)
			mock.ExpectExec(`DELETE FROM gl_subledger_documents WHERE company=\$1 AND id=\$2`).
				WithArgs(testCompany, "AP-"+docNo).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectQuery(`FROM erp_doc_journals WHERE businesscode=\$1 AND transflag=\$2 AND docno=\$3 AND voided_at IS NULL`).
				WithArgs(testCompany, dt.Flag, docNo).WillReturnError(sql.ErrNoRows)
			mock.ExpectQuery(`DELETE FROM stock_ledger WHERE businesscode=\$1 AND transflag=\$2 AND docno=\$3 RETURNING`).
				WithArgs(testCompany, dt.Flag, docNo).
				WillReturnRows(sqlmock.NewRows([]string{"itemcode", "whcode", "docdatetime", "periodkey"}))
			mock.ExpectExec(`UPDATE doc SET isdelete = true`).
				WithArgs(testActor, sqlmock.AnyArg(), testCompany, dt.Flag, docNo).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			if err := NewStore().DeleteDoc(context.Background(), db, testScope(), dt, tc.docIDOrNo); err != nil {
				t.Fatalf("DeleteDoc(%s) = %v", tc.docIDOrNo, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

// ลบเอกสารที่ไม่มีอยู่หรือถูกลบไปแล้วต้องได้ sql.ErrNoRows (หน้าเว็บแปลงเป็น 404) และไม่เขียนอะไรเลย
func TestDeleteDocTwiceIsNotFound(t *testing.T) {
	dt := mustDocType(t, "cashdeposit")
	db, mock, _ := newRecordedMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\), docdatetime FROM doc`).WithArgs(testCompany, dt.Flag, "CDP6910-0001").WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := NewStore().DeleteDoc(context.Background(), db, testScope(), dt, "CDP6910-0001")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("DeleteDoc of a missing document = %v, want sql.ErrNoRows", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// เอกสารที่ถูกตัดหนี้/จัดสรรแล้วลบไม่ได้ ต้องไม่ถอนอะไรเลยและบอกเหตุผลเป็นข้อความที่แปลได้
func TestDeleteDocRefusedWhenSettled(t *testing.T) {
	const docNo = "PO2610-0002"
	dt := mustDocType(t, "purchase")
	db, mock, _ := newRecordedMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(`pg_advisory_xact_lock`).WithArgs(testCompany).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\), docdatetime FROM doc`).WithArgs(testCompany, dt.Flag, docNo).
		WillReturnRows(sqlmock.NewRows([]string{"docno", "iscancel", "approval_status", "docdatetime"}).AddRow(docNo, false, "", testDocDate))
	mock.ExpectQuery(`SELECT EXISTS\(\s+SELECT 1 FROM docref r`).
		WithArgs(testCompany, docNo, dt.Flag).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	expectDateOpen(mock, testDocDay, false)
	mock.ExpectQuery(`FROM gl_subledger_settlements`).WithArgs(testCompany, "AP-"+docNo, "", "").
		WillReturnRows(sqlmock.NewRows([]string{"bound"}).AddRow(1))
	mock.ExpectRollback()

	err := NewStore().DeleteDoc(context.Background(), db, testScope(), dt, docNo)
	if code := userErrorCode(err); code != "erp_doc_settled" {
		t.Fatalf("DeleteDoc error code = %q (%v), want erp_doc_settled", code, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func mustDocType(t *testing.T, key string) doctype.DocType {
	t.Helper()
	dt, ok := doctype.ByKey(key)
	if !ok {
		t.Fatalf("doc type %q is not registered", key)
	}
	return dt
}
