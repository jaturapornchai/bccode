package cheque

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var contactCols = []string{"contact_code", "name", "position", "address", "telephone", "is_contact"}

func TestNormalizeContactsRejectsInvalidRows(t *testing.T) {
	long := func(n int) string { return strings.Repeat("ก", n) }
	for _, tc := range []struct {
		name  string
		in    []BankContactInput
		key   string
		field string
		arg   string
	}{
		{"missing code names the line", []BankContactInput{{ContactCode: "C1", Name: "คุณสมศรี"}, {Name: "คุณสมชาย"}}, "bank_contact_err_required", "contacts[1].contact_code", "2"},
		{"missing name", []BankContactInput{{ContactCode: "C1", Name: "  "}}, "bank_contact_err_required", "contacts[0].name", "1"},
		{"control character in code", []BankContactInput{{ContactCode: "C\n1", Name: "x"}}, "bank_contact_err_text_invalid", "contacts[0].contact_code", "1"},
		{"code over 10", []BankContactInput{{ContactCode: long(11), Name: "x"}}, "bank_contact_err_too_long", "contacts[0].contact_code", "1"},
		{"name over 50", []BankContactInput{{ContactCode: "C1", Name: long(51)}}, "bank_contact_err_too_long", "contacts[0].name", "1"},
		{"position over 50", []BankContactInput{{ContactCode: "C1", Name: "x", Position: long(51)}}, "bank_contact_err_too_long", "contacts[0].position", "1"},
		{"address over 255", []BankContactInput{{ContactCode: "C1", Name: "x", Address: long(256)}}, "bank_contact_err_too_long", "contacts[0].address", "1"},
		{"telephone over 30", []BankContactInput{{ContactCode: "C1", Name: "x", Telephone: long(31)}}, "bank_contact_err_too_long", "contacts[0].telephone", "1"},
		{"duplicate code ignoring case", []BankContactInput{{ContactCode: "ab1", Name: "x"}, {ContactCode: " AB1 ", Name: "y"}}, "bank_contact_err_duplicate_code", "contacts[1].contact_code", "AB1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := normalizeContacts(tc.in)
			e := wantErrorKey(t, err, tc.key)
			if e.Field != tc.field || len(e.Args) != 1 || e.Args[0] != tc.arg {
				t.Fatalf("field/args = %q/%v, want %q/[%s]", e.Field, e.Args, tc.field, tc.arg)
			}
		})
	}
}

func TestNormalizeContactsLimitsTheSet(t *testing.T) {
	rows := make([]BankContactInput, MaxContactsPerBranch+1)
	for i := range rows {
		rows[i] = BankContactInput{ContactCode: fmt.Sprint(i), Name: "x"}
	}
	_, err := normalizeContacts(rows)
	wantErrorKey(t, err, "bank_contact_err_too_many")
	if _, err = normalizeContacts(rows[:MaxContactsPerBranch]); err != nil {
		t.Fatalf("exactly %d contacts must be accepted: %v", MaxContactsPerBranch, err)
	}
}

func TestListBankContactsNeedsBankAndBranch(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	_, err := ListBankContacts(context.Background(), db, "01", " ", "0123")
	if e := wantErrorKey(t, err, "bank_contact_err_bank_required"); e.Field != "bank_code" {
		t.Fatalf("field = %q", e.Field)
	}
	_, err = ListBankContacts(context.Background(), db, "01", "BBL", "")
	if e := wantErrorKey(t, err, "bank_contact_err_branch_required"); e.Field != "bank_branch_code" {
		t.Fatalf("field = %q", e.Field)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListBankContactsOrdersByLineAndCapsRows(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(`FROM bank_contacts WHERE company = \$1 AND bank_code = \$2 AND bank_branch_code = \$3\s+ORDER BY line_no, contact_code LIMIT \$4 OFFSET \$5`).
		WithArgs("01", "BBL", "0123", MaxContactsPerBranch, 0).
		WillReturnRows(sqlmock.NewRows(contactCols).
			AddRow("C1", "คุณสมศรี ใจดี", "ผู้จัดการสาขา", "175 ถนนสีลม กรุงเทพฯ", "02-231-4333", true).
			AddRow("C2", "คุณวิชัย", "", "", "", false))

	items, err := ListBankContacts(context.Background(), db, "01", "BBL", "0123")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Name != "คุณสมศรี ใจดี" || items[1].IsContact {
		t.Fatalf("items = %+v", items)
	}
}

func TestSaveBankContactsReplacesTheWholeSetInOneTransaction(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	off := false
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR SHARE`).WithArgs("01", KindBank, "BBL").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	mock.ExpectQuery(`FOR SHARE`).WithArgs("01", KindBankBranch, "0123").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO bank_contacts .* ON CONFLICT \(company, bank_code, bank_branch_code, contact_code\) DO UPDATE`).
		WithArgs("01", "BBL", "0123", "C1", 1, "คุณสมศรี ใจดี", "ผู้จัดการสาขา", "175 ถนนสีลม", "02-231-4333", true, "somchai").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO bank_contacts`).
		WithArgs("01", "BBL", "0123", "C2", 2, "คุณวิชัย", "", "", "", false, "somchai").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM bank_contacts WHERE company = \$1 AND bank_code = \$2 AND bank_branch_code = \$3 AND contact_code <> ALL`).
		WithArgs("01", "BBL", "0123", `{"C1","C2"}`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`FROM bank_contacts WHERE`).
		WillReturnRows(sqlmock.NewRows(contactCols).AddRow("C1", "คุณสมศรี ใจดี", "ผู้จัดการสาขา", "175 ถนนสีลม", "02-231-4333", true).AddRow("C2", "คุณวิชัย", "", "", "", false))
	mock.ExpectCommit()

	saved, err := SaveBankContacts(context.Background(), db, testScope, " BBL ", "0123", []BankContactInput{
		{ContactCode: "C1", Name: "คุณสมศรี ใจดี", Position: "ผู้จัดการสาขา", Address: "175 ถนนสีลม", Telephone: "02-231-4333"},
		{ContactCode: "C2", Name: "คุณวิชัย", IsContact: &off},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(saved) != 2 {
		t.Fatalf("saved = %+v", saved)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// ส่งชุดว่าง = ล้างผู้ติดต่อทั้งสาขา (DELETE ด้วยอาร์เรย์ว่าง) ไม่ใช่ไม่ทำอะไร
func TestSaveBankContactsEmptySetClearsTheBranch(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR SHARE`).WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	mock.ExpectQuery(`FOR SHARE`).WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	mock.ExpectExec(`DELETE FROM bank_contacts`).WithArgs("01", "BBL", "0123", "{}").WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectQuery(`FROM bank_contacts WHERE`).WillReturnRows(sqlmock.NewRows(contactCols))
	mock.ExpectCommit()

	saved, err := SaveBankContacts(context.Background(), db, testScope, "BBL", "0123", nil)
	if err != nil || len(saved) != 0 {
		t.Fatalf("saved = %+v, err %v", saved, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSaveBankContactsUnknownBranchWritesNothing(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR SHARE`).WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	mock.ExpectQuery(`FOR SHARE`).WithArgs("01", KindBankBranch, "9999").WillReturnRows(sqlmock.NewRows([]string{"one"}))
	mock.ExpectRollback()

	_, err := SaveBankContacts(context.Background(), db, testScope, "BBL", "9999", []BankContactInput{{ContactCode: "C1", Name: "x"}})
	if e := wantErrorKey(t, err, "bank_code_err_not_found"); e.Field != "bank_branch_code" || e.Status != 404 {
		t.Fatalf("error = %+v", e)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSaveBankContactsValidationFailsBeforeOpeningATransaction(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	_, err := SaveBankContacts(context.Background(), db, testScope, "BBL", "0123", []BankContactInput{{ContactCode: "C1"}})
	wantErrorKey(t, err, "bank_contact_err_required")
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
