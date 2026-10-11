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

	mock.ExpectQuery(`SELECT c.code, c.name, c.amount::numeric\(18,2\)::text, c.begin_amount::numeric\(18,2\)::text,.*FROM bc_petty_cash c.*LEFT JOIN bc_petty_cash_trans t`).
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

func TestCreatePettyCashTransValidation(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	scope := Scope{Holding: "rungrueng", Company: "01", Actor: "tester"}

	// 1. ตรวจสอบ doc_no ว่าง
	_, err := CreatePettyCashTrans(context.Background(), db, scope, PettyCashTransInput{
		DocNo:         "",
		DocDate:       "2026-10-11",
		PettyCashCode: "PC01",
		TransType:     1,
		Amount:        "500.00",
	})
	if err == nil {
		t.Fatal("expected error for empty doc_no")
	}

	// 2. ตรวจสอบ doc_date ไม่ถูกต้อง
	_, err = CreatePettyCashTrans(context.Background(), db, scope, PettyCashTransInput{
		DocNo:         "PC-EXP-001",
		DocDate:       "11/10/2026",
		PettyCashCode: "PC01",
		TransType:     1,
		Amount:        "500.00",
	})
	if err == nil {
		t.Fatal("expected error for invalid doc_date")
	}

	// 3. ตรวจสอบ trans_type ไม่ถูกต้อง (ต้องเป็น 1 หรือ 2)
	_, err = CreatePettyCashTrans(context.Background(), db, scope, PettyCashTransInput{
		DocNo:         "PC-EXP-001",
		DocDate:       "2026-10-11",
		PettyCashCode: "PC01",
		TransType:     3,
		Amount:        "500.00",
	})
	if err == nil {
		t.Fatal("expected error for invalid trans_type")
	}

	// 4. ตรวจสอบยอดเงิน amount เป็นศูนย์หรือติดลบ
	_, err = CreatePettyCashTrans(context.Background(), db, scope, PettyCashTransInput{
		DocNo:         "PC-EXP-001",
		DocDate:       "2026-10-11",
		PettyCashCode: "PC01",
		TransType:     1,
		Amount:        "0.00",
	})
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestListPettyCashTransSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"id", "doc_no", "doc_date", "petty_cash_code", "name",
		"trans_type", "amount", "description", "payee_payer", "created_at", "created_by",
	}).AddRow("uuid-1", "PC-REC-001", "2026-10-11", "PC01", "เงินสดย่อยแผนกขาย",
		1, "5000.00", "เบิกชดเชยเงินสดย่อย", "นายสมชาย", "2026-10-11 10:00:00+07", "tester")

	mock.ExpectQuery(`SELECT t.id::text, t.doc_no, t.doc_date::text, t.petty_cash_code, c.name.*FROM bc_petty_cash_trans t.*JOIN bc_petty_cash c`).
		WithArgs("01", "PC01", 500).
		WillReturnRows(rows)

	items, err := ListPettyCashTrans(context.Background(), db, "01", "PC01", "", "", 500)
	if err != nil {
		t.Fatalf("ListPettyCashTrans: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].DocNo != "PC-REC-001" || items[0].Amount != "5000.00" || items[0].TransType != 1 {
		t.Fatalf("unexpected item: %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
