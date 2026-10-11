package cheque

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/lib/pq"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var codeCols = []string{"kind", "code", "name", "bank_code", "user_group", "is_active", "version"}

var testScope = Scope{Company: "01", Actor: "somchai"}

func wantErrorKey(t *testing.T, err error, key string) *Error {
	t.Helper()
	e, ok := AsError(err)
	if !ok {
		t.Fatalf("error = %v, want cheque.Error %q", err, key)
	}
	if e.Key != key {
		t.Fatalf("error key = %q, want %q", e.Key, key)
	}
	return e
}

func TestNormalizeCodeRejectsInvalidInput(t *testing.T) {
	long := func(n int) string { return strings.Repeat("ก", n) }
	for _, tc := range []struct {
		name  string
		in    CodeInput
		key   string
		field string
	}{
		{"unknown kind", CodeInput{Kind: "branch", Code: "A", Name: "x"}, "bank_code_err_kind_invalid", "kind"},
		{"empty kind", CodeInput{Code: "A", Name: "x"}, "bank_code_err_kind_invalid", "kind"},
		{"blank code", CodeInput{Kind: KindBank, Code: "   ", Name: "x"}, "bank_code_err_code_required", "code"},
		{"code over 10 characters", CodeInput{Kind: KindBank, Code: long(11), Name: "x"}, "bank_code_err_code_too_long", "code"},
		{"code with control character", CodeInput{Kind: KindBank, Code: "A\tB", Name: "x"}, "bank_code_err_code_invalid", "code"},
		{"blank name", CodeInput{Kind: KindBank, Code: "BBL", Name: " \t"}, "bank_code_err_name_required", "name"},
		{"name over 50 characters", CodeInput{Kind: KindBank, Code: "BBL", Name: long(51)}, "bank_code_err_name_too_long", "name"},
		{"user group over 255 characters", CodeInput{Kind: KindBank, Code: "BBL", Name: "x", UserGroup: long(256)}, "bank_code_err_group_too_long", "user_group"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := normalizeCode(tc.in)
			e := wantErrorKey(t, err, tc.key)
			if e.Field != tc.field || e.Status != 400 {
				t.Fatalf("field/status = %q/%d, want %q/400", e.Field, e.Status, tc.field)
			}
		})
	}
}

// ขอบเขตความยาวนับเป็นตัวอักษร (ไทยมีสระ/วรรณยุกต์) ไม่ใช่ไบต์: 10 และ 50 ตัวอักษรต้องผ่าน
func TestNormalizeCodeAcceptsBoundaryLengthsAndTrims(t *testing.T) {
	in, err := normalizeCode(CodeInput{Kind: KindBank, Code: " " + strings.Repeat("ก", 10) + " ", Name: strings.Repeat("ข", 50), BankCode: "BBL"})
	if err != nil {
		t.Fatal(err)
	}
	if in.Code != strings.Repeat("ก", 10) {
		t.Fatalf("code not trimmed: %q", in.Code)
	}
	if in.BankCode != "" {
		t.Fatalf("bank filter must only apply to branch codes, got %q", in.BankCode)
	}
	in, err = normalizeCode(CodeInput{Kind: KindBankBranch, Code: "0123", Name: "สีลม", BankCode: " BBL "})
	if err != nil || in.BankCode != "BBL" {
		t.Fatalf("branch bank filter = %q, err %v", in.BankCode, err)
	}
}

func TestLikePatternEscapesWildcards(t *testing.T) {
	if got, want := likePattern(`50%_a\b`), `%50\%\_a\\b%`; got != want {
		t.Fatalf("likePattern = %q, want %q", got, want)
	}
}

func TestListCodesFiltersOrdersAndPages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(`SELECT count\(\*\) FROM bank_codes WHERE company = \$1 AND kind = \$2 AND \(code ILIKE \$3 OR name ILIKE \$3\) AND \(bank_code = \$4 OR bank_code IS NULL\) AND is_active`).
		WithArgs("01", KindBankBranch, "%สีลม%", "BBL").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(`FROM bank_codes WHERE .* ORDER BY code LIMIT \$5 OFFSET \$6`).
		WithArgs("01", KindBankBranch, "%สีลม%", "BBL", 2, 2).
		WillReturnRows(sqlmock.NewRows(codeCols).AddRow(KindBankBranch, "0123", "สาขาสีลม", "BBL", "", true, 1))

	page, err := ListCodes(context.Background(), db, "01", KindBankBranch, CodeFilter{Query: " สีลม ", BankCode: "BBL", ActiveOnly: true, Page: 2, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 3 || page.Page != 2 || page.Limit != 2 || len(page.Items) != 1 || page.Items[0].Code != "0123" {
		t.Fatalf("page = %+v", page)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListCodesCapsLimitAndIgnoresBankFilterOnOtherKinds(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(`SELECT count\(\*\) FROM bank_codes WHERE company = \$1 AND kind = \$2$`).
		WithArgs("01", KindBank).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`ORDER BY code LIMIT \$3 OFFSET \$4`).
		WithArgs("01", KindBank, maxCodeLimit, 0).WillReturnRows(sqlmock.NewRows(codeCols))

	page, err := ListCodes(context.Background(), db, "01", KindBank, CodeFilter{BankCode: "BBL", Limit: 100000})
	if err != nil {
		t.Fatal(err)
	}
	if page.Limit != maxCodeLimit || page.Page != 1 || page.Items == nil || len(page.Items) != 0 {
		t.Fatalf("page = %+v", page)
	}
	if _, err = ListCodes(context.Background(), db, "01", "nope", CodeFilter{}); err == nil {
		t.Fatal("unknown kind must be rejected")
	} else {
		wantErrorKey(t, err, "bank_code_err_kind_invalid")
	}
}

func TestCreateCodeInsertsVersionOneAndChecksBranchBank(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT 1 FROM bank_codes WHERE company = \$1 AND kind = \$2 AND code = \$3 FOR SHARE`).
		WithArgs("01", KindBank, "BBL").WillReturnRows(sqlmock.NewRows([]string{"one"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO bank_codes`).
		WithArgs("01", KindBankBranch, "0123", "สาขาสีลม", "BBL", "บัญชี", true, "somchai").
		WillReturnRows(sqlmock.NewRows(codeCols).AddRow(KindBankBranch, "0123", "สาขาสีลม", "BBL", "บัญชี", true, 1))
	mock.ExpectCommit()

	created, err := CreateCode(context.Background(), db, testScope, CodeInput{Kind: KindBankBranch, Code: "0123", Name: "สาขาสีลม", BankCode: "BBL", UserGroup: "บัญชี"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Version != 1 || created.BankCode != "BBL" || !created.IsActive {
		t.Fatalf("created = %+v", created)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCodeRejectsUnknownBankForBranch(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR SHARE`).WithArgs("01", KindBank, "ZZZ").WillReturnRows(sqlmock.NewRows([]string{"one"}))
	mock.ExpectRollback()

	_, err := CreateCode(context.Background(), db, testScope, CodeInput{Kind: KindBankBranch, Code: "0123", Name: "สาขาสีลม", BankCode: "ZZZ"})
	e := wantErrorKey(t, err, "bank_code_err_not_found")
	if e.Status != 404 || e.Field != "bank_code" || len(e.Args) != 1 || e.Args[0] != "ZZZ" {
		t.Fatalf("error = %+v", e)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCodeDuplicateIsUserError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO bank_codes`).WillReturnError(&pq.Error{Code: "23505"})
	mock.ExpectRollback()

	_, err := CreateCode(context.Background(), db, testScope, CodeInput{Kind: KindBank, Code: "BBL", Name: "ธนาคารกรุงเทพ"})
	e := wantErrorKey(t, err, "bank_code_err_duplicate")
	if e.Status != 409 || e.Field != "code" {
		t.Fatalf("error = %+v", e)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateCodeUnexpectedDatabaseErrorIsNotUserError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO bank_codes`).WillReturnError(errors.New("connection reset"))
	mock.ExpectRollback()

	_, err := CreateCode(context.Background(), db, testScope, CodeInput{Kind: KindBank, Code: "BBL", Name: "ธนาคารกรุงเทพ"})
	if _, isUser := AsError(err); isUser || err == nil {
		t.Fatalf("a database failure must not be reported as a user error: %v", err)
	}
}

func TestUpdateCodeBumpsVersionAndKeepsActiveWhenOmitted(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR UPDATE`).WithArgs("01", KindBank, "BBL").
		WillReturnRows(sqlmock.NewRows(codeCols).AddRow(KindBank, "BBL", "ธนาคารกรุงเทพ", "", "", false, 4))
	mock.ExpectQuery(`UPDATE bank_codes SET name = \$4, bank_code = \$5, user_group = \$6, is_active = \$7,\s+version = version \+ 1`).
		WithArgs("01", KindBank, "BBL", "ธนาคารกรุงเทพ จำกัด (มหาชน)", nil, "", false, "somchai").
		WillReturnRows(sqlmock.NewRows(codeCols).AddRow(KindBank, "BBL", "ธนาคารกรุงเทพ จำกัด (มหาชน)", "", "", false, 5))
	mock.ExpectCommit()

	updated, err := UpdateCode(context.Background(), db, testScope, CodeInput{Kind: KindBank, Code: "BBL", Name: "ธนาคารกรุงเทพ จำกัด (มหาชน)", Version: 4})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 5 || updated.IsActive {
		t.Fatalf("updated = %+v", updated)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateCodeStaleVersionIsConflictAndWritesNothing(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR UPDATE`).WillReturnRows(sqlmock.NewRows(codeCols).AddRow(KindBank, "BBL", "ธนาคารกรุงเทพ", "", "", true, 7))
	mock.ExpectRollback()

	_, err := UpdateCode(context.Background(), db, testScope, CodeInput{Kind: KindBank, Code: "BBL", Name: "ใหม่", Version: 6})
	e := wantErrorKey(t, err, "cheque_err_version_conflict")
	if e.Status != 409 {
		t.Fatalf("status = %d, want 409", e.Status)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateCodeMissingRowIsNotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR UPDATE`).WillReturnRows(sqlmock.NewRows(codeCols))
	mock.ExpectRollback()

	_, err := UpdateCode(context.Background(), db, testScope, CodeInput{Kind: KindBank, Code: "BBL", Name: "x", Version: 1})
	if e := wantErrorKey(t, err, "bank_code_err_not_found"); e.Status != 404 {
		t.Fatalf("status = %d, want 404", e.Status)
	}
}

// รหัสธนาคารที่ถูกอ้างโดยสมุดเงินฝาก (jsonb) ลบไม่ได้ และต้องไม่มี DELETE ส่งไปที่ฐาน
func TestDeleteCodeInUseIsRefused(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR UPDATE`).WithArgs("01", KindBank, "BBL").
		WillReturnRows(sqlmock.NewRows(codeCols).AddRow(KindBank, "BBL", "ธนาคารกรุงเทพ", "", "", true, 2))
	mock.ExpectQuery(`to_regclass`).WillReturnRows(sqlmock.NewRows([]string{"t"}).AddRow("bank_codes").AddRow("bank_contacts").AddRow("gl_subledger_bank_accounts"))
	mock.ExpectQuery(`EXISTS \(SELECT 1 FROM bank_codes WHERE company = \$1 AND kind = 'bank_branch' AND bank_code = \$2\) OR EXISTS \(SELECT 1 FROM bank_contacts WHERE company = \$1 AND bank_code = \$2\) OR EXISTS \(SELECT 1 FROM gl_subledger_bank_accounts WHERE company = \$1 AND payload->>'bank_code' = \$2\)$`).
		WithArgs("01", "BBL").WillReturnRows(sqlmock.NewRows([]string{"used"}).AddRow(true))
	mock.ExpectRollback()

	_, err := DeleteCode(context.Background(), db, testScope, CodeInput{Kind: KindBank, Code: "BBL", Version: 2})
	e := wantErrorKey(t, err, "bank_code_err_in_use")
	if e.Status != 409 || len(e.Args) != 1 || e.Args[0] != "BBL" {
		t.Fatalf("error = %+v", e)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteCodeUnusedIsDeleted(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR UPDATE`).WillReturnRows(sqlmock.NewRows(codeCols).AddRow(KindBankAccountType, "SAV", "ออมทรัพย์", "", "", true, 1))
	mock.ExpectQuery(`to_regclass`).WillReturnRows(sqlmock.NewRows([]string{"t"})) // ยังไม่มีตารางสมุดเงินฝากในฐานนี้
	mock.ExpectExec(`DELETE FROM bank_codes WHERE company = \$1 AND kind = \$2 AND code = \$3`).
		WithArgs("01", KindBankAccountType, "SAV").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	deleted, err := DeleteCode(context.Background(), db, testScope, CodeInput{Kind: KindBankAccountType, Code: "SAV", Version: 1})
	if err != nil || deleted.Code != "SAV" {
		t.Fatalf("deleted = %+v, err %v", deleted, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteCodeWithoutConsumersSkipsUsageProbe(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery(`FOR UPDATE`).WillReturnRows(sqlmock.NewRows(codeCols).AddRow(KindBankService, "SV1", "บริการ", "", "", true, 1))
	mock.ExpectExec(`DELETE FROM bank_codes`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if _, err := DeleteCode(context.Background(), db, testScope, CodeInput{Kind: KindBankService, Code: "SV1", Version: 1}); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// ทุกชนิดรหัสที่มีผู้ใช้ต้องมี probe ครบ และ SQL ของ probe ใช้เฉพาะพารามิเตอร์ $1/$2 ไม่ฝังข้อมูลผู้ใช้
func TestCodeUsageProbesAreParameterized(t *testing.T) {
	for kind, probes := range codeUsage {
		if !ValidCodeKind(kind) {
			t.Fatalf("usage declared for unknown kind %q", kind)
		}
		for _, p := range probes {
			if !strings.Contains(p.condition, "$2") || strings.ContainsAny(p.table, " ;$'\"") {
				t.Fatalf("unsafe usage probe %+v for %s", p, kind)
			}
		}
	}
	for _, kind := range []string{KindBank, KindBankBranch, KindBankAccountType, KindChequeExtStatus} {
		if len(codeUsage[kind]) == 0 {
			t.Fatalf("kind %s has consumers but no usage probe", kind)
		}
	}
}
