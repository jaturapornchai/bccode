package httpapi

import (
	"context"
	"regexp"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	gl "smlcloudplatform/internal/generalledger"
)

func TestListReceiptAdjustments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	store := &Store{}
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"docno", "docdatetime", "custcode", "custname", "ir_ref_no", "ir_ref_date", "gr_qty", "ir_qty", "isclosed",
	}).AddRow("GR001", now, "AP001", "Supplier 1", "PUR001", now.Format(time.RFC3339), "10.0000", "10.0000", false)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT d.docno,`)).
		WithArgs("TEST_COMP", 401).
		WillReturnRows(rows)

	filter := ReceiptAdjustmentFilter{}
	items, err := store.ListReceiptAdjustments(context.Background(), db, "TEST_COMP", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].DocNo != "GR001" {
		t.Errorf("expected docno GR001, got %s", items[0].DocNo)
	}
	if !items[0].IsMatched {
		t.Errorf("expected matched to be true")
	}
	if items[0].GRQty != "10.0000" || items[0].IRQty != "10.0000" {
		t.Errorf("expected quantities 10.0000, got %s / %s", items[0].GRQty, items[0].IRQty)
	}
}

func TestProcessReceiptAdjustments(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	store := &Store{}
	scope := gl.Scope{
		Company: "TEST_COMP",
		Actor:   "U01",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`pg_advisory_xact_lock`).
		WithArgs("TEST_COMP").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`UPDATE doc SET isclosed = true.*`).
		WithArgs("TEST_COMP", "GR001", "U01", 401).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`INSERT INTO erp_doc_events .*`).
		WithArgs("TEST_COMP", 401, "GR001", "ปรับปรุงสถานะปิดใบรับสินค้า (GR Closed)", "U01").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	count, err := store.ProcessReceiptAdjustments(context.Background(), db, scope, []string{"GR001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 updated, got %d", count)
	}
}
