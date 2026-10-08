package inventory

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"smlcloudplatform/internal/doctype"
)

// ป้ายประเภทเอกสารในรายงานเคลื่อนไหวสต็อกต้องตรงกับทะเบียน (รับคืนจากการขาย = ใบรับคืนสินค้า/ลดหนี้)
func TestStockMovementReportLabelsFollowRegistry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	pattern := ""
	for _, flag := range []int{doctype.Purchase, doctype.SaleInvoice, doctype.StockPickup, doctype.SaleInvoiceReturn, doctype.PurchaseReturn} {
		pattern += `[\s\S]*` + regexp.QuoteMeta(fmt.Sprintf("WHEN s.transflag = %d THEN", flag))
	}
	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(pattern).WithArgs("01", day, day, "P-1001").
		WillReturnRows(sqlmock.NewRows([]string{"docdate"}))

	rows, err := NewInventoryCostingService(db).GetStockMovementReport(context.Background(), "01", "P-1001", day, day)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Errorf("rows = %v, want none", rows)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
