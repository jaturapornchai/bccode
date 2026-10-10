package httpapi

import (
	"context"
	"regexp"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestListBillingNotesForPrint(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	store := &Store{}
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"docno", "docdatetime", "custcode", "custname", "totalamount", "remark", "creator_code", "count",
	}).AddRow("BL001", now, "AR001", "Customer 1", "25000.00", "รอบวางบิล 15", "SALE01", 3)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT d.docno,`)).
		WithArgs("TEST_COMP", 26).
		WillReturnRows(rows)

	filter := BillingNotePrintFilter{}
	items, err := store.ListBillingNotesForPrint(context.Background(), db, "TEST_COMP", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].DocNo != "BL001" {
		t.Errorf("expected docno BL001, got %s", items[0].DocNo)
	}
	if items[0].TotalAmount != "25000.00" {
		t.Errorf("expected total amount 25000.00, got %s", items[0].TotalAmount)
	}
	if items[0].InvCount != 3 {
		t.Errorf("expected 3 invoices, got %d", items[0].InvCount)
	}
}
