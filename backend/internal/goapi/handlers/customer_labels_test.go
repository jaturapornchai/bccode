package handlers

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestQueryCustomerLabels(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"code", "name_th", "name_en", "tax_id", "tax_branch_no", "address",
		"shipping_address", "addr_district", "addr_province", "addr_postcode",
		"phone", "total_sales", "doc_count",
	}).AddRow(
		"C001", "บริษัท สยามพาณิชย์ จำกัด", "Siam Commercial Co., Ltd.", "0105558000000", "00000",
		"123 ถนนสุขุมวิท", "123 ถนนสุขุมวิท", "คลองเตย", "กรุงเทพฯ", "10110", "02-123-4567", "50000.00", 5,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT p.code,`)).
		WithArgs("01", 500, 0).
		WillReturnRows(rows)

	req := CustomerLabelsRequest{
		HoldingCode: "test",
		Limit:       500,
		Offset:      0,
	}

	items, err := queryCustomerLabels(context.Background(), db, "01", req)
	if err != nil {
		t.Fatalf("queryCustomerLabels returned error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].CustomerCode != "C001" {
		t.Errorf("expected customer code C001, got %s", items[0].CustomerCode)
	}
	if items[0].BillingAddress != "123 ถนนสุขุมวิท" {
		t.Errorf("expected billing address 123 ถนนสุขุมวิท, got %s", items[0].BillingAddress)
	}
	if items[0].TotalSales != "50000.00" {
		t.Errorf("expected total sales 50000.00, got %s", items[0].TotalSales)
	}
}

// TestQueryCustomerLabelsScopesToOneCompany - เดิม (p.company = $1 OR p.company = '01' OR $1 = ”)
// ทำให้ลูกค้าของบริษัท 01 ปนในป้ายของบริษัทอื่น และรหัสว่างดึงลูกค้าทุกบริษัท
func TestQueryCustomerLabelsScopesToOneCompany(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`FROM gl_subledger_partners p\s+WHERE p\.company = \$1 AND COALESCE`).
		WithArgs("02", 500, 0).
		WillReturnRows(sqlmock.NewRows([]string{"code"}))

	req := CustomerLabelsRequest{HoldingCode: "test", Limit: 500}
	if _, err := queryCustomerLabels(context.Background(), db, "02", req); err != nil {
		t.Fatalf("queryCustomerLabels returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

	for _, empty := range []string{"", "  "} {
		if _, err := queryCustomerLabels(context.Background(), db, empty, req); !errors.Is(err, errReportCompanyRequired) {
			t.Fatalf("company %q: err = %v, want errReportCompanyRequired", empty, err)
		}
	}
}
