package httpapi

import (
	"context"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// เอกสารที่อนุมัติแล้วหรือยกเลิกแล้วแก้ไข/ลบไม่ได้ (ต้องยกเลิกการอนุมัติก่อน) — หยุดก่อนตรวจงวดและก่อนเขียนอะไรลงฐาน
func TestSaveAndDeleteRefuseApprovedOrCancelledDocs(t *testing.T) {
	po := mustDocType(t, "purchase-order")
	const docNo = "PO6910-0030"
	for _, state := range []struct {
		name      string
		cancelled bool
		approval  string
	}{
		{"approved", false, "approved"},
		{"cancelled", true, ""},
	} {
		t.Run("edit "+state.name, func(t *testing.T) {
			db, mock, _ := newRecordedMock(t)
			mock.ExpectBegin()
			mock.ExpectQuery(`SELECT COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\), docdatetime\s+FROM doc WHERE`).
				WithArgs(testCompany, po.Flag, docNo).
				WillReturnRows(sqlmock.NewRows([]string{"iscancel", "approval_status", "docdatetime"}).AddRow(state.cancelled, state.approval, testDocDate))
			mock.ExpectRollback()

			doc := newTestDoc()
			doc.DocNo = docNo
			_, err := NewStore().SaveDoc(context.Background(), db, testScope(), po, doc, true)
			if code := userErrorCode(err); code != "erp_doc_locked" {
				t.Fatalf("SaveDoc on %s document = %q (%v), want erp_doc_locked", state.name, code, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})

		t.Run("delete "+state.name, func(t *testing.T) {
			db, mock, _ := newRecordedMock(t)
			mock.ExpectBegin()
			expectStateRow(mock, po.Flag, docNo, state.cancelled, state.approval)
			mock.ExpectRollback()

			err := NewStore().DeleteDoc(context.Background(), db, testScope(), po, docNo)
			if code := userErrorCode(err); code != "erp_doc_locked" {
				t.Fatalf("DeleteDoc on %s document = %q (%v), want erp_doc_locked", state.name, code, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}
