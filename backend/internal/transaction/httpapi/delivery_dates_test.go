package httpapi

import (
	"context"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"smlcloudplatform/internal/doctype"
)

func TestListDeliveryDateInvoices(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	docDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT docno, .* FROM doc WHERE businesscode = \$1 AND transflag = \$2 .*`).
		WithArgs("test_company", 44, "C001").
		WillReturnRows(sqlmock.NewRows([]string{"docno", "docdatetime", "perioddatetime", "custcode", "custname"}).
			AddRow("INV001", docDate, dueDate, "C001", "ลูกค้าทดสอบ จำกัด"))

	items, err := store.ListDeliveryDateInvoices(ctx, db, "test_company", 44, "C001")
	if err != nil {
		t.Fatalf("ListDeliveryDateInvoices failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].DocNo != "INV001" {
		t.Errorf("expected docno INV001, got %s", items[0].DocNo)
	}
	if items[0].CreditOfBill != 30 {
		t.Errorf("expected 30 credit days, got %d", items[0].CreditOfBill)
	}
	if items[0].DeliveryDate != "2026-10-01" {
		t.Errorf("expected delivery date 2026-10-01, got %s", items[0].DeliveryDate)
	}
}

func TestSaveDeliveryDates(t *testing.T) {
	ctx := context.Background()
	store := NewStore()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	si, _ := doctype.ByKey("sale-invoice")
	scope := testScope()

	mock.ExpectBegin()
	mock.ExpectExec(`pg_advisory_xact_lock`).
		WithArgs(scope.Company).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`UPDATE doc SET perioddatetime = \$1, modified_at = now\(\), modifier_code = \$2 WHERE businesscode = \$3 AND transflag = \$4 AND docno = \$5 .*`).
		WithArgs(sqlmock.AnyArg(), scope.Actor, scope.Company, si.Flag, "INV001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM information_schema.tables WHERE table_name = 'gl_subledger_documents'\)`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectExec(`UPDATE gl_subledger_documents SET payload = jsonb_set\(COALESCE\(payload::jsonb, '{}'::jsonb\), '{due_date}', to_jsonb\(\$1::text\)\) WHERE company = \$2 AND \(id = \$3 OR id = 'AR-' \|\| \$3 OR id = 'AP-' \|\| \$3\)`).
		WithArgs("2026-11-15", scope.Company, "INV001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`INSERT INTO erp_doc_events .*`).
		WithArgs(scope.Company, si.Flag, "INV001", "ปรับปรุงวันครบกำหนดเป็น 2026-11-15", scope.Actor).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	items := []SaveDeliveryDateItem{
		{
			DocNo:          "INV001",
			DeliveryDate:   "2026-10-16",
			CreditDelivery: 30,
			NewDueDate:     "2026-11-15",
		},
	}

	count, err := store.SaveDeliveryDates(ctx, db, scope, si, items)
	if err != nil {
		t.Fatalf("SaveDeliveryDates failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 updated, got %d", count)
	}
}
