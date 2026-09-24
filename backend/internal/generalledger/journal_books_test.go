package generalledger

import (
	"net/http"
	"strings"
	"testing"
)

func userCode(t *testing.T, err error) *UserError {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	user, ok := AsUserError(err)
	if !ok {
		t.Fatalf("expected *UserError, got %T %v", err, err)
	}
	if user.Message == "" {
		t.Fatalf("error %s has no Thai message", user.Code)
	}
	return user
}

// Thai codes: vowel signs and tone marks are combining marks (\p{M}), which the old
// ^[\p{L}\p{N}] regexp rejected — ลุงจืดสร้างสมุด "สมุดซื้อ" ไม่ได้.
func TestCheckCodeAcceptsThaiCodes(t *testing.T) {
	for _, code := range []string{"ขายสด-๐๑", "ค่าน้ำ/ไฟ", "สมุดซื้อ", "กข(1)", "JV", "A.B_C#1:2", "๑๒๓", "เงินสด"} {
		if err := checkCode(code, "code", "รหัส", codeMaxRunes); err != nil {
			t.Fatalf("%q rejected: %v", code, err)
		}
	}
}

func TestCheckCodeRejectsWithFieldAndReason(t *testing.T) {
	cases := []struct{ code, want string }{
		{"", "code_required"},
		{"ัก", "code_invalid_start"}, // leading Thai vowel sign (mark)
		{"่ก", "code_invalid_start"}, // leading tone mark
		{"-JV", "code_invalid_start"},
		{"JV 01", "code_has_space"},
		{"JV\u00a001", "code_has_space"}, // no-break space pasted from Excel
		{"JV\t01", "code_has_space"},
		{"JV\x0001", "code_invisible_char"},
		{"JV\u200b01", "code_invisible_char"}, // zero-width space
		{"JV\xff", "code_invisible_char"},     // invalid UTF-8
		{"a|b", "code_invalid_char"},
		{`a"b`, "code_invalid_char"},
		{"a'b", "code_invalid_char"},
		{"a,b", "code_invalid_char"},
		{"a;b", "code_invalid_char"},
		{"a@b", "code_invalid_char"},
	}
	for _, c := range cases {
		user := userCode(t, checkCode(c.code, "code", "รหัสสมุดรายวัน", BookCodeMaxRunes))
		if user.Code != c.want || user.Field != "code" {
			t.Fatalf("%q: got code=%s field=%s want %s on code", c.code, user.Code, user.Field, c.want)
		}
		if !strings.Contains(user.Message, "รหัสสมุดรายวัน") {
			t.Fatalf("%q: message does not name the field: %s", c.code, user.Message)
		}
	}
	if user := userCode(t, checkCode("JV 01", "code", "รหัส", 15)); !strings.Contains(user.Message, "รหัสห้ามมีช่องว่าง") {
		t.Fatalf("space message = %s", user.Message)
	}
}

// Limits count runes (characters), not bytes: one Thai letter is 3 bytes in UTF-8.
func TestCodeLimitsCountRunes(t *testing.T) {
	limits := []struct {
		name string
		max  int
	}{{"book", BookCodeMaxRunes}, {"docno", DocNoMaxRunes}, {"account", AccountCodeMaxRunes}}
	for _, l := range limits {
		if err := checkCode(strings.Repeat("ก", l.max), "code", "รหัส", l.max); err != nil {
			t.Fatalf("%s: %d Thai runes rejected: %v", l.name, l.max, err)
		}
		user := userCode(t, checkCode(strings.Repeat("ก", l.max+1), "code", "รหัส", l.max))
		if user.Code != "code_too_long" || !strings.Contains(user.Message, "ไม่เกิน") {
			t.Fatalf("%s: over limit got %s %s", l.name, user.Code, user.Message)
		}
	}
	if BookCodeMaxRunes != 15 || DocNoMaxRunes != 30 || AccountCodeMaxRunes != 20 {
		t.Fatalf("limits drifted from mydocs DDL: book=%d docno=%d account=%d", BookCodeMaxRunes, DocNoMaxRunes, AccountCodeMaxRunes)
	}
	// ผ่านจุดตรวจของ model จริง
	if user := userCode(t, (Account{AccountCode: strings.Repeat("๑", 21), Names: []Name{{Code: "th", Name: "เงินสด"}}}).Validate()); user.Field != "accountcode" || user.Code != "code_too_long" {
		t.Fatalf("account code 21 runes: %s/%s", user.Field, user.Code)
	}
	if err := (Account{AccountCode: strings.Repeat("๑", 20), Names: []Name{{Code: "th", Name: "เงินสด"}}, AccountType: "asset", NormalBalance: "debit"}).Validate(); err != nil && strings.Contains(err.Error(), "รหัสบัญชี") {
		t.Fatalf("account code 20 runes rejected: %v", err)
	}
	if user := userCode(t, (Journal{DocNo: strings.Repeat("ข", 31), Date: "2026-01-01", Description: "ขายสด"}).Validate(FiscalYear{}, nil)); user.Field != "docno" || user.Code != "code_too_long" {
		t.Fatalf("docno 31 runes: %s/%s", user.Field, user.Code)
	}
}

func TestNormalizeCodeTrimsAndComposes(t *testing.T) {
	if got := NormalizeCode("  Cafe\u0301 "); got != "Caf\u00e9" {
		t.Fatalf("NormalizeCode = %q", got)
	}
	if got := NormalizeCode("\tสมุดซื้อ\n"); got != "สมุดซื้อ" {
		t.Fatalf("NormalizeCode Thai = %q", got)
	}
	cmd := Command{DocNo: " X ", Journal: &Journal{DocNo: " ขาย-01 ", BookCode: " SV ", Lines: []Line{{AccountCode: " 1111 "}}}, Master: &Master{Code: " ซื้อ "}}
	original := cmd.Journal
	normalizeCommandCodes(&cmd)
	if cmd.Journal.DocNo != "ขาย-01" || cmd.Journal.BookCode != "SV" || cmd.Journal.Lines[0].AccountCode != "1111" || cmd.Master.Code != "ซื้อ" || cmd.DocNo != "X" {
		t.Fatalf("normalizeCommandCodes: %+v %+v", cmd.Journal, cmd.Master)
	}
	if original.DocNo != " ขาย-01 " || original.Lines[0].AccountCode != " 1111 " {
		t.Fatal("normalizeCommandCodes mutated the caller's journal")
	}
}

func TestRejectNULNamesFieldAndReturns400(t *testing.T) {
	cmd := Command{Resource: "journals", Action: "create", Journal: &Journal{DocNo: "JV-01", Lines: []Line{{}, {}, {Description: "ค่าน้ำ\x00"}}}}
	user := userCode(t, rejectNUL(cmd))
	if user.Code != "text_contains_nul" || user.Field != "lines[2].description" || user.HTTPStatus() != http.StatusBadRequest {
		t.Fatalf("got code=%s field=%s status=%d", user.Code, user.Field, user.HTTPStatus())
	}
	if !strings.Contains(user.Message, "บรรทัดที่ 3 คำอธิบาย") {
		t.Fatalf("message does not point at the line: %s", user.Message)
	}
	user = userCode(t, rejectNUL(Command{Master: &Master{Kind: "journal-books", Code: "SV", Name: "สมุด\x00ขาย"}}))
	if user.Field != "name" || !strings.HasPrefix(user.Message, "ชื่อ") {
		t.Fatalf("master name NUL: field=%s msg=%s", user.Field, user.Message)
	}
	if err := rejectNUL(Command{Journal: &Journal{Description: "ปกติ", Lines: []Line{{Description: "ok"}}}}); err != nil {
		t.Fatalf("clean command rejected: %v", err)
	}
}

func TestValidateJournalBookMaster(t *testing.T) {
	ok := Master{Kind: "journal-books", Code: "สมุดซื้อ", Name: "  สมุดรายวันซื้อ  ", NameEn: " Purchase Journal ", BookType: BookTypePurchase, IsActive: true}
	if err := validateJournalBookMaster(&ok, false); err != nil {
		t.Fatalf("valid Thai book rejected: %v", err)
	}
	if ok.Name != "สมุดรายวันซื้อ" || ok.NameEn != "Purchase Journal" {
		t.Fatalf("names not trimmed: %q %q", ok.Name, ok.NameEn)
	}
	cases := []struct {
		m            Master
		allowUntyped bool
		code, field  string
	}{
		{Master{Code: "สมุดรายวันซื้อเชื่อ1", Name: "ก", BookType: 1}, false, "code_too_long", "code"},
		{Master{Code: "SV", Name: "  ", BookType: 4}, false, "journal_book_name_required", "name"},
		{Master{Code: "SV", Name: strings.Repeat("ข", 101), BookType: 4}, false, "journal_book_name_too_long", "name"},
		{Master{Code: "SV", Name: "ขาย", NameEn: strings.Repeat("s", 101), BookType: 4}, false, "journal_book_name_en_too_long", "nameen"},
		{Master{Code: "SV", Name: "ขาย"}, false, "journal_book_type_invalid", "booktype"},
		{Master{Code: "SV", Name: "ขาย", BookType: 7}, true, "journal_book_type_invalid", "booktype"},
	}
	for _, c := range cases {
		m := c.m
		user := userCode(t, validateJournalBookMaster(&m, c.allowUntyped))
		if user.Code != c.code || user.Field != c.field {
			t.Fatalf("%+v: got %s/%s want %s/%s", c.m, user.Code, user.Field, c.code, c.field)
		}
	}
	// legacy book without a type: renaming or deactivating stays possible
	legacy := Master{Code: "GJ", Name: "สมุดเก่า"}
	if err := validateJournalBookMaster(&legacy, true); err != nil {
		t.Fatalf("legacy untyped edit rejected: %v", err)
	}
}

func TestDefaultJournalBooksFollowSpecMeaning(t *testing.T) {
	want := map[string]int{"JV": BookTypeGeneral, "PV": BookTypePayment, "RV": BookTypeReceipt, "SV": BookTypeSales, "UV": BookTypePurchase}
	books := DefaultJournalBooks()
	if len(books) != len(want) {
		t.Fatalf("default books = %d", len(books))
	}
	for _, b := range books {
		if want[b.Code] != b.BookType || b.Name == "" || b.NameEn == "" {
			t.Fatalf("default book %+v", b)
		}
		m := Master{Kind: "journal-books", Code: b.Code, Name: b.Name, NameEn: b.NameEn, BookType: b.BookType}
		if err := validateJournalBookMaster(&m, false); err != nil {
			t.Fatalf("default book %s invalid: %v", b.Code, err)
		}
	}
}

// Processes and the fixed-asset poster pick the book by booktype, never by code.
func TestChooseBookCodeByType(t *testing.T) {
	books := []Master{
		{Code: "ทั่วไป", BookType: BookTypeGeneral, IsActive: true},
		{Code: "GJ", BookType: BookTypeGeneral, IsActive: true},
		{Code: "AA", BookType: BookTypeGeneral, IsActive: false},
		{Identity: Identity{IsDeleted: true}, Code: "AB", BookType: BookTypeGeneral, IsActive: true},
		{Code: "JV", BookType: 0, IsActive: true}, // legacy untyped: never inferred from its code
		{Code: "SV", BookType: BookTypeSales, IsActive: true},
	}
	if got, err := ChooseBookCode(books, BookTypeGeneral); err != nil || got != "GJ" {
		t.Fatalf("general = %q %v", got, err)
	}
	if got, err := ChooseBookCode(books, BookTypeOpening, BookTypeGeneral); err != nil || got != "GJ" {
		t.Fatalf("opening falls back to general = %q %v", got, err)
	}
	withOpening := append([]Master{{Code: "OB", BookType: BookTypeOpening, IsActive: true}}, books...)
	if got, err := ChooseBookCode(withOpening, BookTypeOpening, BookTypeGeneral); err != nil || got != "OB" {
		t.Fatalf("opening = %q %v", got, err)
	}
	user := userCode(t, func() error {
		_, err := ChooseBookCode([]Master{{Code: "JV", IsActive: true}, {Code: "SV", BookType: BookTypeSales, IsActive: true}}, BookTypeGeneral)
		return err
	}())
	if user.Code != "journal_book_general_missing" || !strings.Contains(user.Message, "กำหนดสมุดรายวัน") {
		t.Fatalf("missing general book: %s %s", user.Code, user.Message)
	}
}

func TestJournalBookErrorsExplainTheFix(t *testing.T) {
	checks := []struct {
		err         error
		code, field string
		mustContain string
	}{
		{journalBookNotFound("ขาย"), "journal_book_not_found", "bookcode", "ขาย"},
		{journalBookTypeMissing("GJ"), "journal_book_type_missing", "bookcode", "กำหนดสมุดรายวัน"},
		{journalBookInactive("SV"), "journal_book_inactive", "bookcode", "SV"},
		{journalBookInUseDelete("SV"), "journal_book_in_use_delete", "code", "ปิดใช้งาน"},
		{journalBookInUseCode("SV"), "journal_book_in_use_code", "code", "SV"},
		{journalBookInUseType("SV"), "journal_book_in_use_type", "booktype", "SV"},
	}
	for _, c := range checks {
		user := userCode(t, c.err)
		if user.Code != c.code || user.Field != c.field || !strings.Contains(user.Message, c.mustContain) {
			t.Fatalf("%s: field=%s msg=%s", c.code, user.Field, user.Message)
		}
	}
}
