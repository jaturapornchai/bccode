package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// Journal books are a user-defined master (kind "journal-books"). Business rules
// read the book's booktype, never its code: companies name their books freely.
// booktype follows mydocs/datamodels/gl/journalbook.sql (book_type SMALLINT 1..6).
const (
	BookTypeGeneral  = 1 // ทั่วไป
	BookTypePayment  = 2 // จ่าย
	BookTypeReceipt  = 3 // รับ
	BookTypeSales    = 4 // ขาย
	BookTypePurchase = 5 // ซื้อ
	BookTypeOpening  = 6 // ยอดยกมา

	bookNameMaxRunes = 100 // journal_books.name_th / name_en VARCHAR(100)
)

func validBookType(t int) bool { return t >= BookTypeGeneral && t <= BookTypeOpening }

// DefaultJournalBook is one of the standard books a new company starts with.
type DefaultJournalBook struct {
	Code, Name, NameEn string
	BookType           int
}

// DefaultJournalBooks returns the market-standard set named in the spec comment of
// journalbook.sql: JV general, PV payment, RV receipt, SV sales, UV purchase.
func DefaultJournalBooks() []DefaultJournalBook {
	return []DefaultJournalBook{
		{Code: "JV", Name: "สมุดรายวันทั่วไป", NameEn: "General Journal", BookType: BookTypeGeneral},
		{Code: "PV", Name: "สมุดรายวันจ่ายเงิน", NameEn: "Payment Journal", BookType: BookTypePayment},
		{Code: "RV", Name: "สมุดรายวันรับเงิน", NameEn: "Receipt Journal", BookType: BookTypeReceipt},
		{Code: "SV", Name: "สมุดรายวันขาย", NameEn: "Sales Journal", BookType: BookTypeSales},
		{Code: "UV", Name: "สมุดรายวันซื้อ", NameEn: "Purchase Journal", BookType: BookTypePurchase},
	}
}

// ---- user-facing errors (codes get gl_err_<code> rows in languages.tsv) ----

var (
	errJournalBookNameRequired   = fieldError("journal_book_name_required", "name", "กรุณาระบุชื่อสมุดรายวันภาษาไทย")
	errJournalBookNameTooLong    = fieldError("journal_book_name_too_long", "name", fmt.Sprintf("ชื่อสมุดรายวันภาษาไทยยาวได้ไม่เกิน %d ตัวอักษร กรุณาย่อชื่อให้สั้นลง", bookNameMaxRunes))
	errJournalBookNameEnTooLong  = fieldError("journal_book_name_en_too_long", "nameen", fmt.Sprintf("ชื่อสมุดรายวันภาษาอังกฤษยาวได้ไม่เกิน %d ตัวอักษร กรุณาย่อชื่อให้สั้นลง", bookNameMaxRunes))
	errJournalBookTypeInvalid    = fieldError("journal_book_type_invalid", "booktype", "กรุณาเลือกประเภทสมุดรายวัน: ทั่วไป จ่าย รับ ขาย ซื้อ หรือยอดยกมา")
	errJournalBookRequired       = fieldError("journal_book_required", "bookcode", "กรุณาเลือกสมุดรายวันของเอกสาร")
	errJournalBookImmutable      = fieldError("journal_book_immutable", "bookcode", "เปลี่ยนสมุดรายวันของเอกสารที่บันทึกแล้วไม่ได้ ถ้าเลือกสมุดผิด ให้ลบใบร่างนี้แล้วบันทึกใหม่ในสมุดที่ถูกต้อง")
	errJournalKindImmutable      = fieldError("journal_kind_immutable", "kind", "เปลี่ยนประเภทรายการของเอกสารที่บันทึกแล้วไม่ได้ ถ้าเลือกประเภทผิด ให้ลบใบร่างนี้แล้วบันทึกใหม่")
	errJournalDocNoImmutable     = fieldError("journal_docno_immutable", "docno", "เลขที่เอกสารที่บันทึกแล้วแก้ไม่ได้ ถ้าต้องการเลขใหม่ ให้ลบใบร่างนี้แล้วบันทึกใหม่")
	errJournalBranchRequired     = fieldError("journal_branch_required", "branchcode", "กรุณาเลือกสาขาของเอกสาร (คุณเข้าระบบระดับบริษัท ระบบจึงเลือกสาขาให้อัตโนมัติไม่ได้)")
	errGeneralJournalBookMissing = userError("journal_book_general_missing", "ยังไม่มีสมุดรายวันประเภททั่วไปที่เปิดใช้งาน กรุณาไปที่หน้ากำหนดสมุดรายวัน สร้างหรือเปิดใช้งานสมุดประเภททั่วไป แล้วทำรายการอีกครั้ง")
)

func journalBookNotFound(code string) error {
	return fieldError("journal_book_not_found", "bookcode", fmt.Sprintf("ไม่พบสมุดรายวัน “%s” ในบริษัทนี้ กรุณาเลือกสมุดที่มีอยู่ หรือสร้างสมุดนี้ที่หน้ากำหนดสมุดรายวันก่อน", code))
}

func journalBookTypeMissing(code string) error {
	return fieldError("journal_book_type_missing", "bookcode", fmt.Sprintf("สมุดรายวัน “%s” ยังไม่ได้กำหนดประเภท กรุณาไปที่หน้ากำหนดสมุดรายวัน เลือกประเภทสมุด (ทั่วไป จ่าย รับ ขาย ซื้อ หรือยอดยกมา) แล้วบันทึกเอกสารอีกครั้ง", code))
}

func journalBookInactive(code string) error {
	return fieldError("journal_book_inactive", "bookcode", fmt.Sprintf("สมุดรายวัน “%s” ปิดใช้งานอยู่ กรุณาเลือกสมุดอื่นที่เปิดใช้งาน หรือเปิดใช้งานสมุดนี้ที่หน้ากำหนดสมุดรายวัน", code))
}

func journalBookInUseDelete(code string) error {
	return conflictError("journal_book_in_use_delete", "code", fmt.Sprintf("สมุดรายวัน “%s” มีเอกสารบันทึกอยู่แล้ว หรือมีรูปแบบการเชื่อมบัญชีเลือกสมุดนี้อยู่ ลบไม่ได้ ถ้าไม่ต้องการใช้ต่อ ให้ปิดใช้งานแทน", code))
}

func journalBookInUseCode(code string) error {
	return conflictError("journal_book_in_use_code", "code", fmt.Sprintf("สมุดรายวัน “%s” มีเอกสารบันทึกอยู่แล้ว หรือมีรูปแบบการเชื่อมบัญชีเลือกสมุดนี้อยู่ เปลี่ยนรหัสไม่ได้ ถ้าต้องการรหัสใหม่ ให้สร้างสมุดใหม่แล้วปิดใช้งานสมุดเดิม", code))
}

func journalBookInUseType(code string) error {
	return conflictError("journal_book_in_use_type", "booktype", fmt.Sprintf("สมุดรายวัน “%s” มีเอกสารบันทึกอยู่แล้ว หรือมีรูปแบบการเชื่อมบัญชีเลือกสมุดนี้อยู่ เปลี่ยนประเภทไม่ได้ ถ้าต้องการประเภทอื่น ให้สร้างสมุดใหม่", code))
}

// JournalBranchOutsideSession: the voucher names a branch other than the one the user logged into.
func JournalBranchOutsideSession(branch, session string) error {
	return fieldError("journal_branch_outside_session", "branchcode", fmt.Sprintf("สาขาของเอกสาร “%s” ไม่ตรงกับสาขาที่เข้าระบบ “%s” กรุณาเว้นว่างเพื่อใช้สาขาที่เข้าระบบ หรือเข้าระบบใหม่ในสาขาที่ต้องการ", branch, session))
}

// JournalBranchNotFound: the branch is not an active branch of the company in the central registry.
func JournalBranchNotFound(branch string) error {
	return fieldError("journal_branch_not_found", "branchcode", fmt.Sprintf("ไม่พบสาขา “%s” ที่เปิดใช้งานในบริษัทนี้ กรุณาเลือกสาขาจากรายการ", branch))
}

// JournalBranchRequired: a company-wide session must choose the voucher's branch itself.
func JournalBranchRequired() error { return errJournalBranchRequired }

// JournalBookImmutable / JournalKindImmutable are shared with the HTTP layer, which rejects
// the change before authorizing against the stored journal.
func JournalBookImmutable() error { return errJournalBookImmutable }
func JournalKindImmutable() error { return errJournalKindImmutable }

// ---- master validation ----

// validateJournalBookMaster checks the journal-books record contract:
// { code ≤15, name (Thai, required ≤100), nameen (≤100), booktype 1..6, isactive }.
// allowUntyped lets a legacy book without booktype be renamed or deactivated as it is.
func validateJournalBookMaster(m *Master, allowUntyped bool) error {
	if err := checkCode(m.Code, "code", "รหัสสมุดรายวัน", BookCodeMaxRunes); err != nil {
		return err
	}
	m.Name = strings.TrimSpace(m.Name)
	m.NameEn = strings.TrimSpace(m.NameEn)
	switch {
	case m.Name == "":
		return errJournalBookNameRequired
	case utf8.RuneCountInString(m.Name) > bookNameMaxRunes:
		return errJournalBookNameTooLong
	case utf8.RuneCountInString(m.NameEn) > bookNameMaxRunes:
		return errJournalBookNameEnTooLong
	case !validBookType(m.BookType) && !(allowUntyped && m.BookType == 0):
		return errJournalBookTypeInvalid
	}
	return nil
}

// journalBookInUse reports whether any live journal or account-mapping master refers to the book
// code (the spec FK journal_entries.book_code → journal_books is ON DELETE RESTRICT; a mapping's
// bookcode is where its generated vouchers go, so renaming/retyping/deleting the book breaks it too).
func journalBookInUse(ctx context.Context, tx *sql.Tx, company, code string) (bool, error) {
	var used bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind IN ('journals','mappings') AND payload->>'bookcode'=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false))`, company, code).Scan(&used)
	return used, err
}

// guardJournalBookUpdate allows name/nameen/isactive edits always; code and booktype are
// frozen once a journal uses the book. A legacy book without booktype may still get one.
func guardJournalBookUpdate(ctx context.Context, tx *sql.Tx, company string, old, next Master) error {
	if next.Code == old.Code && (next.BookType == old.BookType || old.BookType == 0) {
		return nil
	}
	used, err := journalBookInUse(ctx, tx, company, old.Code)
	if err != nil || !used {
		return err
	}
	if next.Code != old.Code {
		return journalBookInUseCode(old.Code)
	}
	return journalBookInUseType(old.Code)
}

func guardJournalBookDelete(ctx context.Context, tx *sql.Tx, company string, old Master) error {
	used, err := journalBookInUse(ctx, tx, company, old.Code)
	if err != nil {
		return err
	}
	if used {
		return journalBookInUseDelete(old.Code)
	}
	return nil
}

// ---- journal validation ----

func loadJournalBook(ctx context.Context, tx *sql.Tx, company, code string) (Master, bool, error) {
	var payload []byte
	err := tx.QueryRowContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='journal-books' AND code=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) ORDER BY id LIMIT 1`, company, code).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return Master{}, false, nil
	}
	if err != nil {
		return Master{}, false, err
	}
	var book Master
	if err = json.Unmarshal(payload, &book); err != nil {
		return Master{}, false, err
	}
	return book, true, nil
}

// checkJournalBook validates journal.bookcode against the company's journal-books master.
// requireActive is true for new vouchers; drafts, posting and reversals of an existing
// voucher keep working after its book is deactivated (deactivate is the alternative to delete).
func checkJournalBook(ctx context.Context, tx *sql.Tx, company string, j Journal, requireActive bool) error {
	if j.BookCode == "" {
		return errJournalBookRequired
	}
	book, found, err := loadJournalBook(ctx, tx, company, j.BookCode)
	if err != nil {
		return err
	}
	if !found {
		return journalBookNotFound(j.BookCode)
	}
	if !validBookType(book.BookType) {
		// Legacy books were created without a type; never infer it from the code.
		return journalBookTypeMissing(j.BookCode)
	}
	if requireActive && !book.IsActive {
		return journalBookInactive(j.BookCode)
	}
	return nil
}

// ---- book selection for generated vouchers ----

// ChooseBookCode returns the lowest code of an active, typed, live book whose booktype is
// the first match in the preference order (e.g. opening: 6 then 1).
func ChooseBookCode(books []Master, types ...int) (string, error) {
	for _, want := range types {
		best := ""
		for _, b := range books {
			if b.IsDeleted || !b.IsActive || b.BookType != want || b.Code == "" {
				continue
			}
			if best == "" || b.Code < best {
				best = b.Code
			}
		}
		if best != "" {
			return best, nil
		}
	}
	return "", errGeneralJournalBookMissing
}

func loadJournalBooks(ctx context.Context, tx *sql.Tx, company string) ([]Master, error) {
	rows, err := tx.QueryContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='journal-books' AND NOT COALESCE((payload->>'isdeleted')::boolean,false) ORDER BY code,id`, company)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	books := []Master{}
	for rows.Next() {
		var payload []byte
		if err = rows.Scan(&payload); err != nil {
			return nil, err
		}
		var b Master
		if err = json.Unmarshal(payload, &b); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

// processBookCode picks the book for closing (general) or year-end opening
// (opening type if the company keeps one, else general) vouchers.
func processBookCode(ctx context.Context, tx *sql.Tx, company string, opening bool) (string, error) {
	books, err := loadJournalBooks(ctx, tx, company)
	if err != nil {
		return "", err
	}
	if opening {
		return ChooseBookCode(books, BookTypeOpening, BookTypeGeneral)
	}
	return ChooseBookCode(books, BookTypeGeneral)
}

// defaultJournalBookChanges creates the standard books together with a company's first
// fiscal year — the moment a company sets up its ledger. It never runs again once the
// company has any journal-books record (live or deleted), so books the user removed or
// renamed are not resurrected and existing untyped books are left for the user to type.
// A company that already has journals but no book rows (legacy data) gets nothing either:
// its journals may use JV/PV/RV/SV/UV with a meaning other than the defaults (old data had
// SV=purchase, UV=sale), and typing those codes here would infer the type from the code —
// the user sets the types on the journal-books screen instead.
func defaultJournalBookChanges(ctx context.Context, tx *sql.Tx, scope Scope, now time.Time) ([]Change, error) {
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind IN ('journal-books','journals'))`, scope.Company).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, nil
	}
	changes := make([]Change, 0, len(DefaultJournalBooks()))
	for _, d := range DefaultJournalBooks() {
		m := Master{Kind: "journal-books", Code: d.Code, Name: d.Name, NameEn: d.NameEn, BookType: d.BookType, IsActive: true, Identity: newIdentity(scope, now)}
		data, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		changes = append(changes, Change{Kind: "journal-books", ID: m.ID, Code: m.Code, Payload: string(data)})
	}
	return changes, nil
}

var masterCodeLabels = map[string]string{
	"account-groups": "รหัสกลุ่มผังบัญชี", "product-account-groups": "รหัสกลุ่มบัญชีสินค้า", "mappings": "รหัสรูปแบบการเชื่อม",
	"budgets": "รหัสงบประมาณ", "forecast": "รหัสประมาณการกระแสเงินสด", "allocations": "รหัสการปันส่วน",
	"statement-templates": "รหัสรูปแบบงบการเงิน",
}

// validateMasterCode checks the code of an auxiliary master; journal books also get their
// whole record contract checked (allowUntypedBook: updating a legacy book that has no type yet).
func validateMasterCode(m *Master, allowUntypedBook bool) error {
	if m.Kind == "journal-books" {
		return validateJournalBookMaster(m, allowUntypedBook)
	}
	label, ok := masterCodeLabels[m.Kind]
	if !ok {
		label = "รหัส"
	}
	return checkCode(m.Code, "code", label, codeMaxRunes)
}
