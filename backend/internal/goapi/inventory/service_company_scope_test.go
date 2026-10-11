package inventory

import (
	"context"
	"errors"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// รายงานสต็อกต้องกรองบริษัทของ session อย่างเดียว: เดิม WHERE มี OR s.businesscode = '01' และ OR $1 ว่าง
// ทำให้บริษัท 02 เห็นสต็อกบริษัท 01 และรหัสว่างได้ทุกบริษัท (ผลบนฐานจริงอยู่ใน service_company_scope_integration_test.go)
func TestInventoryReportsRejectEmptyCompany(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := NewInventoryCostingService(db)
	ctx := context.Background()
	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	if _, err := svc.GetStockMovementReport(ctx, "", "", day, day); !errors.Is(err, ErrCompanyRequired) {
		t.Errorf("stock movement: err = %v, want ErrCompanyRequired", err)
	}
	if _, err := svc.GetStockCard(ctx, "rungrueng", "", "P-1001", day, day); !errors.Is(err, ErrCompanyRequired) {
		t.Errorf("stock card: err = %v, want ErrCompanyRequired", err)
	}
	if _, err := svc.GetInventoryValuation(ctx, "rungrueng", ""); !errors.Is(err, ErrCompanyRequired) {
		t.Errorf("valuation: err = %v, want ErrCompanyRequired", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestInventoryReportsFilterOnlySessionCompany(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := NewInventoryCostingService(db)
	ctx := context.Background()
	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`WHERE s\.businesscode = \$1\s+AND s\.docdatetime::date >= \$2::date`).
		WithArgs("02", day, day, "").WillReturnRows(sqlmock.NewRows([]string{"docdate"}))
	if _, err := svc.GetStockMovementReport(ctx, "02", "", day, day); err != nil {
		t.Fatal(err)
	}

	// สต็อกการ์ด: stock_ledger กรองด้วยบริษัทเป็นแหล่งเดียว — บริษัทที่ไม่มีแถวได้รายงานว่าง ไม่ถอยไปตาราง holding เดิม
	mock.ExpectQuery(`WHERE s\.businesscode = \$1\s+AND s\.itemcode = \$2`).
		WithArgs("02", "P-1001", day, day).WillReturnRows(sqlmock.NewRows([]string{"docdatetime"}))
	card, err := svc.GetStockCard(ctx, "rungrueng", "02", "P-1001", day, day)
	if err != nil || len(card.Entries) != 0 {
		t.Fatalf("stock card = %+v, %v; want empty report", card, err)
	}

	// มูลค่าคงเหลือ: ยอดปิดงวดล่าสุดของบริษัทนี้ และชื่อสินค้าจาก docdetail ของบริษัทเดียวกัน
	mock.ExpectQuery(`FROM stock_period_balance WHERE businesscode = \$1\s+ORDER BY itemcode, whcode, periodkey DESC[\s\S]*FROM docdetail WHERE businesscode = \$1 `).
		WithArgs("02").WillReturnRows(sqlmock.NewRows([]string{"itemcode"}))
	valuation, err := svc.GetInventoryValuation(ctx, "rungrueng", "02")
	if err != nil || len(valuation.Items) != 0 {
		t.Fatalf("valuation = %+v, %v; want empty report", valuation, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// คำสั่งอ่านสมุดสต็อกล้ม = ตอบข้อผิดพลาด ห้ามกลืนแล้วถอยไปตารางเดิมที่กรองแค่ holding (ได้สต็อกทุกบริษัท) หรือคืนรายงานว่างเงียบ ๆ
func TestInventoryReportsReturnQueryErrors(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc := NewInventoryCostingService(db)
	ctx := context.Background()
	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	failure := errors.New("relation stock_ledger is unavailable")

	mock.ExpectQuery(`FROM stock_ledger s`).WillReturnError(failure)
	if _, err := svc.GetStockCard(ctx, "rungrueng", "02", "P-1001", day, day); !errors.Is(err, failure) {
		t.Errorf("stock card: err = %v, want %v", err, failure)
	}
	mock.ExpectQuery(`FROM stock_period_balance`).WillReturnError(failure)
	if _, err := svc.GetInventoryValuation(ctx, "rungrueng", "02"); !errors.Is(err, failure) {
		t.Errorf("valuation: err = %v, want %v", err, failure)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
