package httpapi

import (
	"context"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
)

func TestApproveDocTransitions(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	t.Run("non-approvable module rejected", func(t *testing.T) {
		sir, _ := doctype.ByKey("stock-issue-request")
		_, err := store.ApproveDoc(ctx, nil, testScope(), sir, "SIR001")
		if uerr, ok := gl.AsUserError(err); !ok || uerr.Code != "erp_doc_not_pending" {
			t.Fatalf("ApproveDoc on non-approvable module = %v, want erp_doc_not_pending", err)
		}
	})

	t.Run("PR already approved rejected", func(t *testing.T) {
		pr, _ := doctype.ByKey("purchase-requisition")
		db, mock, _ := newRecordedMock(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\) FROM doc .* FOR UPDATE`).
			WithArgs(testCompany, pr.Flag, "PR001").
			WillReturnRows(sqlmock.NewRows([]string{"docno", "iscancel", "approval_status"}).AddRow("PR001", false, "approved"))
		mock.ExpectRollback()

		_, err := store.ApproveDoc(ctx, db, testScope(), pr, "PR001")
		if uerr, ok := gl.AsUserError(err); !ok || uerr.Code != "erp_doc_not_pending" {
			t.Fatalf("ApproveDoc on already approved = %v, want erp_doc_not_pending", err)
		}
	})

	t.Run("PR already cancelled rejected", func(t *testing.T) {
		pr, _ := doctype.ByKey("purchase-requisition")
		db, mock, _ := newRecordedMock(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\) FROM doc .* FOR UPDATE`).
			WithArgs(testCompany, pr.Flag, "PR002").
			WillReturnRows(sqlmock.NewRows([]string{"docno", "iscancel", "approval_status"}).AddRow("PR002", true, ""))
		mock.ExpectRollback()

		_, err := store.ApproveDoc(ctx, db, testScope(), pr, "PR002")
		if uerr, ok := gl.AsUserError(err); !ok || uerr.Code != "erp_doc_already_cancelled" {
			t.Fatalf("ApproveDoc on cancelled = %v, want erp_doc_already_cancelled", err)
		}
	})

	t.Run("pending PR approved successfully", func(t *testing.T) {
		pr, _ := doctype.ByKey("purchase-requisition")
		stamp := time.Date(2026, 10, 10, 8, 0, 0, 0, time.UTC)
		docColumns := []string{"id", "docno", "docdatetime", "transflag", "custcode", "totalamount", "totalvalue_doc",
			"totaldiscount_doc", "totalvatvalue_doc", "totalbeforevat_doc", "totalaftervat_doc", "iscancel", "isclosed",
			"branchid", "custname", "description", "remark", "created_at", "modified_at",
			"approval_status", "approved_by", "approved_at", "cancelled_by", "cancelled_at", "cancelreason"}
		db, mock, _ := newRecordedMock(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\) FROM doc .* FOR UPDATE`).
			WithArgs(testCompany, pr.Flag, "PR003").
			WillReturnRows(sqlmock.NewRows([]string{"docno", "iscancel", "approval_status"}).AddRow("PR003", false, ""))
		mock.ExpectExec(`UPDATE doc SET\s+approval_status = 'approved'`).
			WithArgs(testActor, sqlmock.AnyArg(), testCompany, pr.Flag, "PR003").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO erp_doc_events`).
			WithArgs(testCompany, pr.Flag, "PR003", testActor, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		// GetDoc mock
		mock.ExpectQuery(`FROM doc WHERE businesscode = \$1 AND transflag = \$2 AND NOT COALESCE\(isdelete, false\) AND docno = \$3 LIMIT 1`).
			WithArgs(testCompany, pr.Flag, "PR003").
			WillReturnRows(sqlmock.NewRows(docColumns).AddRow(1, "PR003", "2026-10-10T08:00:00Z", pr.Flag, "V01",
				100.0, 100.0, 0.0, 0.0, 100.0, 100.0, false, false, testBranch, "ผู้จำหน่าย", "ขอซื้อ", "", stamp, stamp,
				"approved", testActor, stamp, "", nil, ""))
		mock.ExpectQuery(`SELECT docnoref, docnoreftransflag FROM docref`).
			WithArgs(testCompany, "PR003").
			WillReturnRows(sqlmock.NewRows([]string{"docnoref", "docnoreftransflag"}))
		mock.ExpectQuery(`FROM docdetail WHERE businesscode = \$1 AND transflag = \$2 AND docno = \$3`).
			WithArgs(testCompany, pr.Flag, "PR003").
			WillReturnRows(sqlmock.NewRows([]string{"id", "linenumber", "itemcode", "barcode", "unitcode", "totalqty",
				"price", "sumamount", "description", "whcode", "locationcode", "docref"}))

		approved, err := store.ApproveDoc(ctx, db, testScope(), pr, "PR003")
		if err != nil {
			t.Fatalf("ApproveDoc error = %v", err)
		}
		if approved.Status != 1 || approved.ApprovalStatus != "approved" {
			t.Fatalf("ApproveDoc status = %d (%s), want 1 (approved)", approved.Status, approved.ApprovalStatus)
		}
	})
}

func TestUnapproveDocTransitions(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	pr, _ := doctype.ByKey("purchase-requisition")

	t.Run("not approved rejected", func(t *testing.T) {
		db, mock, _ := newRecordedMock(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\) FROM doc .* FOR UPDATE`).
			WithArgs(testCompany, pr.Flag, "PR001").
			WillReturnRows(sqlmock.NewRows([]string{"docno", "iscancel", "approval_status"}).AddRow("PR001", false, ""))
		mock.ExpectRollback()

		_, err := store.UnapproveDoc(ctx, db, testScope(), pr, "PR001", "แก้ไข")
		if uerr, ok := gl.AsUserError(err); !ok || uerr.Code != "erp_doc_not_approved" {
			t.Fatalf("UnapproveDoc on pending = %v, want erp_doc_not_approved", err)
		}
	})

	t.Run("has live downstream rejected", func(t *testing.T) {
		db, mock, _ := newRecordedMock(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\) FROM doc .* FOR UPDATE`).
			WithArgs(testCompany, pr.Flag, "PR002").
			WillReturnRows(sqlmock.NewRows([]string{"docno", "iscancel", "approval_status"}).AddRow("PR002", false, "approved"))
		mock.ExpectQuery(`SELECT EXISTS\(\s+SELECT 1 FROM docref r`).
			WithArgs(testCompany, "PR002", pr.Flag).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		mock.ExpectRollback()

		_, err := store.UnapproveDoc(ctx, db, testScope(), pr, "PR002", "แก้ไข")
		if uerr, ok := gl.AsUserError(err); !ok || uerr.Code != "erp_doc_has_downstream" {
			t.Fatalf("UnapproveDoc with downstream = %v, want erp_doc_has_downstream", err)
		}
	})
}

func TestCancelDocTransitions(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	po, _ := doctype.ByKey("purchase-order")

	t.Run("empty reason rejected", func(t *testing.T) {
		_, err := store.CancelDoc(ctx, nil, testScope(), po, "PO001", "   ")
		if uerr, ok := gl.AsUserError(err); !ok || uerr.Code != "erp_cancel_reason_required" {
			t.Fatalf("CancelDoc with empty reason = %v, want erp_cancel_reason_required", err)
		}
	})

	t.Run("PR cancel rejects approved PR", func(t *testing.T) {
		pr, _ := doctype.ByKey("purchase-requisition")
		db, mock, _ := newRecordedMock(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\) FROM doc .* FOR UPDATE`).
			WithArgs(testCompany, pr.Flag, "PR001").
			WillReturnRows(sqlmock.NewRows([]string{"docno", "iscancel", "approval_status"}).AddRow("PR001", false, "approved"))
		mock.ExpectRollback()

		_, err := store.CancelDoc(ctx, db, testScope(), pr, "PR001", "ไม่ต้องการแล้ว")
		if uerr, ok := gl.AsUserError(err); !ok || uerr.Code != "erp_doc_not_pending" {
			t.Fatalf("CancelDoc on approved PR = %v, want erp_doc_not_pending", err)
		}
	})

	t.Run("PO cancel rejects live downstream", func(t *testing.T) {
		db, mock, _ := newRecordedMock(t)
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT docno, COALESCE\(iscancel, false\), COALESCE\(approval_status, ''\) FROM doc .* FOR UPDATE`).
			WithArgs(testCompany, po.Flag, "PO002").
			WillReturnRows(sqlmock.NewRows([]string{"docno", "iscancel", "approval_status"}).AddRow("PO002", false, "approved"))
		mock.ExpectQuery(`SELECT EXISTS\(\s+SELECT 1 FROM docref r`).
			WithArgs(testCompany, "PO002", po.Flag).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		mock.ExpectRollback()

		_, err := store.CancelDoc(ctx, db, testScope(), po, "PO002", "ไม่ต้องการแล้ว")
		if uerr, ok := gl.AsUserError(err); !ok || uerr.Code != "erp_doc_has_downstream" {
			t.Fatalf("CancelDoc with downstream = %v, want erp_doc_has_downstream", err)
		}
	})
}

func TestScreenAliasesForApprovalAndCancellation(t *testing.T) {
	po, _ := doctype.ByKey("purchase-order")
	pr, _ := doctype.ByKey("purchase-requisition")
	sir, _ := doctype.ByKey("stock-issue-request")

	// Users with screen + action permission can access
	sessionApprove := glhttp.NewSessionAccess(gl.Scope{Holding: "H", Company: "C", Actor: "U"}, "purchase-order-approve", "purchase-order-approve:approve")
	if !allowsDocType(sessionApprove, po, actionApprove) {
		t.Error("purchase-order-approve:approve must grant approve on purchase-order")
	}

	sessionCancel := glhttp.NewSessionAccess(gl.Scope{Holding: "H", Company: "C", Actor: "U"}, "purchase-requisition-cancel", "purchase-requisition-cancel:cancel")
	if !allowsDocType(sessionCancel, pr, actionCancel) {
		t.Error("purchase-requisition-cancel:cancel must grant cancel on purchase-requisition")
	}

	sessionSIR := glhttp.NewSessionAccess(gl.Scope{Holding: "H", Company: "C", Actor: "U"}, "stock-issue-request-cancel", "stock-issue-request-cancel:cancel")
	if !allowsDocType(sessionSIR, sir, actionCancel) {
		t.Error("stock-issue-request-cancel:cancel must grant cancel on stock-issue-request")
	}
}
