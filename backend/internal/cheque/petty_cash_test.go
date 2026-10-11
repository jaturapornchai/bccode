package cheque

import (
	"context"
	"strings"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestSavePettyCashValidation(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	scope := Scope{Holding: "rungrueng", Company: "01", Actor: "tester"}

	// 1. ตรวจสอบ code ว่าง
	_, err := SavePettyCash(context.Background(), db, scope, []PettyCashInput{
		{Code: "", Name: "เงินสดย่อยแผนกขาย", Amount: "10000.00", BeginAmount: "5000.00"},
	})
	if err == nil {
		t.Fatal("expected error for empty code")
	}

	// 2. ตรวจสอบ name ว่าง
	_, err = SavePettyCash(context.Background(), db, scope, []PettyCashInput{
		{Code: "PC01", Name: "  ", Amount: "10000.00", BeginAmount: "5000.00"},
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}

	// 3. ตรวจสอบ code ซ้ำในชุดเดียวกัน
	_, err = SavePettyCash(context.Background(), db, scope, []PettyCashInput{
		{Code: "PC01", Name: "สดย่อย 1", Amount: "10000.00", BeginAmount: "5000.00"},
		{Code: "pc01", Name: "สดย่อย 2", Amount: "20000.00", BeginAmount: "10000.00"},
	})
	if err == nil {
		t.Fatal("expected error for duplicate code")
	}

	// 4. ตรวจสอบยอดเงิน amount ติดลบ
	_, err = SavePettyCash(context.Background(), db, scope, []PettyCashInput{
		{Code: "PC01", Name: "สดย่อย 1", Amount: "-500.00", BeginAmount: "0.00"},
	})
	if err == nil {
		t.Fatal("expected error for negative amount")
	}

	// 5. ตรวจสอบ code ยาวเกิน 20 ตัวอักษร
	_, err = SavePettyCash(context.Background(), db, scope, []PettyCashInput{
		{Code: strings.Repeat("A", 21), Name: "สดย่อย 1", Amount: "1000.00", BeginAmount: "0.00"},
	})
	if err == nil {
		t.Fatal("expected error for code too long")
	}
}

func TestListPettyCashSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"code", "name", "amount", "begin_amount", "balance", "version"}).
		AddRow("PC01", "เงินสดย่อยแผนกขาย", "10000.00", "5000.00", "5000.00", 1).
		AddRow("PC02", "เงินสดย่อยสำนักงาน", "20000.00", "15000.00", "15000.00", 1)

	mock.ExpectQuery(`SELECT code, name, amount::numeric\(18,2\)::text, begin_amount::numeric\(18,2\)::text, balance::numeric\(18,2\)::text, version FROM bc_petty_cash`).
		WithArgs("01", MaxPettyCashPerCompany).
		WillReturnRows(rows)

	items, err := ListPettyCash(context.Background(), db, "01")
	if err != nil {
		t.Fatalf("ListPettyCash: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Code != "PC01" || items[0].Amount != "10000.00" || items[0].Balance != "5000.00" {
		t.Fatalf("unexpected item 0: %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
