//go:build integration

package generalledger

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func (f *pgIntegrityFixture) books() map[string]Master {
	f.t.Helper()
	page, err := f.store.List(f.ctx, f.scope, "journal-books", "", 1, 100, ListFilter{})
	if err != nil {
		f.t.Fatal(err)
	}
	out := map[string]Master{}
	for _, raw := range page.Items {
		var m Master
		if err := json.Unmarshal(raw, &m); err != nil {
			f.t.Fatal(err)
		}
		out[m.Code] = m
	}
	return out
}

func (f *pgIntegrityFixture) failCode(cmd Command, code, field string) *UserError {
	f.t.Helper()
	_, err := f.execute(f.scope, cmd)
	if err == nil {
		f.t.Fatalf("%s/%s accepted, want %s", cmd.Resource, cmd.Action, code)
	}
	user, ok := AsUserError(err)
	if !ok || user.Code != code || (field != "" && user.Field != field) || user.Message == "" {
		f.t.Fatalf("%s/%s: got %#v, want code=%s field=%s", cmd.Resource, cmd.Action, err, code, field)
	}
	return user
}

func (f *pgIntegrityFixture) bookJournal(doc, book string) Journal {
	return Journal{DocNo: doc, Date: "2026-01-10", BookCode: book, FiscalYear: "2026", Description: "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. เป็นเงินสด", Kind: "manual", BranchCode: "B1",
		Lines: []Line{{AccountCode: "5000", Debit: Amount("1250.50"), Credit: Amount("0")}, {AccountCode: "1000", Debit: Amount("0"), Credit: Amount("1250.50")}}}
}

// Journal books are a master the user maintains: create → use → delete blocked → deactivate,
// Thai codes accepted, processes pick the book by booktype (never by code).
func TestJournalBooksLifecycleIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)

	// 1. the first fiscal year created the standard books with their types (SV ขาย, UV ซื้อ)
	books := f.books()
	for _, want := range DefaultJournalBooks() {
		if got := books[want.Code]; got.BookType != want.BookType || !got.IsActive || got.Name != want.Name {
			t.Fatalf("default book %s = %+v, want type %d", want.Code, got, want.BookType)
		}
	}
	if len(books) != len(DefaultJournalBooks()) {
		t.Fatalf("second fiscal year duplicated books: %d", len(books))
	}

	// 2. a Thai-coded book (trimmed on save)
	created := f.run(Command{Resource: "journal-books", Action: "create", Master: &Master{Kind: "journal-books", Code: "  สมุดซื้อ ", Name: "สมุดรายวันซื้อเงินสด", NameEn: "Cash Purchase Journal", BookType: BookTypePurchase, IsActive: true}})
	book := f.books()["สมุดซื้อ"]
	if book.ID != created.ID || book.BookType != BookTypePurchase {
		t.Fatalf("Thai book not stored normalised: %+v", book)
	}
	f.deny(f.scope, Command{Resource: "journal-books", Action: "create", Master: &Master{Kind: "journal-books", Code: "สมุดซื้อ", Name: "ซ้ำ", BookType: BookTypePurchase, IsActive: true}})
	f.failCode(Command{Resource: "journal-books", Action: "create", Master: &Master{Kind: "journal-books", Code: "ัซื้อ", Name: "ขึ้นต้นด้วยสระ", BookType: BookTypePurchase, IsActive: true}}, "code_invalid_start", "code")
	f.failCode(Command{Resource: "journal-books", Action: "create", Master: &Master{Kind: "journal-books", Code: "สมุด ซื้อ", Name: "มีช่องว่าง", BookType: BookTypePurchase, IsActive: true}}, "code_has_space", "code")
	f.failCode(Command{Resource: "journal-books", Action: "create", Master: &Master{Kind: "journal-books", Code: "สมุดรายวันซื้อเชื่อ", Name: "ยาวเกิน", BookType: BookTypePurchase, IsActive: true}}, "code_too_long", "code")
	f.failCode(Command{Resource: "journal-books", Action: "create", Master: &Master{Kind: "journal-books", Code: "ไม่มีประเภท", Name: "ไม่ระบุประเภท", IsActive: true}}, "journal_book_type_invalid", "booktype")

	// 3. use it with a Thai document number
	j := f.bookJournal("ซื้อสด-๐๑", "สมุดซื้อ")
	used := f.run(Command{Resource: "journals", Action: "create", Journal: &j})
	if got := f.journal(used.ID); got.BookCode != "สมุดซื้อ" || got.DocNo != "ซื้อสด-๐๑" {
		t.Fatalf("journal stored %q/%q", got.BookCode, got.DocNo)
	}

	// 4. a referenced book cannot be deleted, recoded or retyped
	f.failCode(Command{Resource: "journal-books", Action: "delete", ID: book.ID, Version: book.Version}, "journal_book_in_use_delete", "code")
	recoded := book
	recoded.Code = "ซื้อ2"
	f.failCode(Command{Resource: "journal-books", Action: "update", ID: book.ID, Version: book.Version, Master: &recoded}, "journal_book_in_use_code", "code")
	retyped := book
	retyped.BookType = BookTypeGeneral
	f.failCode(Command{Resource: "journal-books", Action: "update", ID: book.ID, Version: book.Version, Master: &retyped}, "journal_book_in_use_type", "booktype")

	// 5. rename + deactivate are always allowed
	renamed := book
	renamed.Name = "สมุดรายวันซื้อเงินสด (เลิกใช้)"
	renamed.IsActive = false
	f.run(Command{Resource: "journal-books", Action: "update", ID: book.ID, Version: book.Version, Master: &renamed})
	if got := f.books()["สมุดซื้อ"]; got.IsActive || got.Name != renamed.Name {
		t.Fatalf("deactivate/rename not stored: %+v", got)
	}

	// 6. inactive book: no new documents, but the existing draft still posts
	j2 := f.bookJournal("ซื้อสด-๐๒", "สมุดซื้อ")
	f.failCode(Command{Resource: "journals", Action: "create", Journal: &j2}, "journal_book_inactive", "bookcode")
	f.post(f.journal(used.ID))

	// 7. unknown book / unused book delete
	j3 := f.bookJournal("ซื้อสด-๐๓", "ไม่มีสมุดนี้")
	f.failCode(Command{Resource: "journals", Action: "create", Journal: &j3}, "journal_book_not_found", "bookcode")
	rv := f.books()["RV"]
	f.run(Command{Resource: "journal-books", Action: "delete", ID: rv.ID, Version: rv.Version})
	if _, ok := f.books()["RV"]; ok {
		t.Fatal("unused book not deleted")
	}

	// 8. legacy book without a type: never inferred from the code "PV"; user is sent to the setup screen
	if _, err := f.db.ExecContext(f.ctx, `UPDATE gl_records SET payload = payload - 'booktype' WHERE company=$1 AND kind='journal-books' AND code='PV'`, f.scope.Company); err != nil {
		t.Fatal(err)
	}
	j4 := f.bookJournal("PV-0001", "PV")
	user := f.failCode(Command{Resource: "journals", Action: "create", Journal: &j4}, "journal_book_type_missing", "bookcode")
	if !strings.Contains(user.Message, "PV") || !strings.Contains(user.Message, "กำหนดสมุดรายวัน") {
		t.Fatalf("legacy book message: %s", user.Message)
	}
	pv := f.books()["PV"]
	pv.Name = "สมุดรายวันจ่ายเงิน (ปรับชื่อ)"
	f.run(Command{Resource: "journal-books", Action: "update", ID: pv.ID, Version: pv.Version, Master: &pv}) // rename still works
	pv = f.books()["PV"]
	pv.BookType = BookTypePayment
	f.run(Command{Resource: "journal-books", Action: "update", ID: pv.ID, Version: pv.Version, Master: &pv})
	f.run(Command{Resource: "journals", Action: "create", Journal: &j4})

	// 9. NUL is a per-field 400 before the jsonb insert
	j5 := f.bookJournal("JV-NUL", "JV")
	j5.Lines[1].Description = "ค่าน้ำ\x00"
	user = f.failCode(Command{Resource: "journals", Action: "create", Journal: &j5}, "text_contains_nul", "lines[1].description")
	if user.HTTPStatus() != http.StatusBadRequest {
		t.Fatalf("NUL status %d", user.HTTPStatus())
	}

	// 10. processes choose by type: JV off, "GJ" is the only active general book → closing uses GJ
	jv := f.books()["JV"]
	jv.IsActive = false
	f.run(Command{Resource: "journal-books", Action: "update", ID: jv.ID, Version: jv.Version, Master: &jv})
	// drafts would block the close — post them first
	page, err := f.store.List(f.ctx, f.scope, "journals", "", 1, 50, ListFilter{Status: "draft"})
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range page.Items {
		var draft Journal
		_ = json.Unmarshal(raw, &draft)
		f.post(draft)
	}
	year := f.years["2026"]
	missing := Command{Resource: "processes", Action: "close", ID: "2026", Version: year.Version, DocNo: "CLOSE26", Date: "2026-12-31", Reason: "ปิดบัญชีสิ้นปี"}
	f.failCode(missing, "journal_book_general_missing", "")
	f.run(Command{Resource: "journal-books", Action: "create", Master: &Master{Kind: "journal-books", Code: "GJ", Name: "สมุดรายวันทั่วไป (ปิดบัญชี)", BookType: BookTypeGeneral, IsActive: true}})
	missing.RequestID = uuid.NewString()
	closeResult := f.run(missing)
	closing := f.journal(closeResult.ID)
	if closing.Kind != "closing" || closing.BookCode != "GJ" {
		t.Fatalf("closing used book %q (kind %s)", closing.BookCode, closing.Kind)
	}
	f.post(closing)

	// 11. year-end opening prefers a type-6 book
	f.run(Command{Resource: "journal-books", Action: "create", Master: &Master{Kind: "journal-books", Code: "ยกมา", Name: "สมุดรายวันยอดยกมา", BookType: BookTypeOpening, IsActive: true}})
	result := f.run(Command{Resource: "processes", Action: "year-end", ID: "2026", Version: year.Version, TargetYear: "2027", DocNo: "OPEN27", Date: "2027-01-01", Reason: "ยกยอดไปปีถัดไป", RequestID: uuid.NewString()})
	if opening := f.journal(result.ID); opening.Kind != "opening" || opening.BookCode != "ยกมา" {
		t.Fatalf("opening used book %q (kind %s)", opening.BookCode, opening.Kind)
	}
}

// บริษัทที่มีใบสำคัญอยู่แล้วแต่ไม่มีแถวสมุด (ข้อมูลเก่า SV=ซื้อ UV=ขาย) สร้างปีบัญชีใหม่ต้องไม่ได้สมุดมาตรฐานที่เดาประเภทจากรหัส
func TestDefaultJournalBooksSkippedWhenJournalsExist(t *testing.T) {
	f := newPGIntegrityFixture(t)
	draft := f.draft("SV6901-S001", "B1")
	if _, err := f.db.Exec(`DELETE FROM gl_records WHERE company='C' AND kind='journal-books'`); err != nil {
		t.Fatal(err)
	}
	// บริษัทเดิมที่มีเอกสารแต่ไม่มีทะเบียนสมุด: แก้ใบร่างต้องได้ข้อความที่บอกรหัสสมุดที่ขาดและชี้ไปหน้ากำหนดสมุดรายวัน
	user := f.failCode(Command{Resource: "journals", Action: "update", ID: draft.ID, Version: draft.Version, Journal: &draft}, "journal_book_not_found", "bookcode")
	if !strings.Contains(user.Message, "“"+draft.BookCode+"”") || !strings.Contains(user.Message, "กำหนดสมุดรายวัน") || user.HTTPStatus() != http.StatusBadRequest {
		t.Fatalf("missing book message = %q (status %d)", user.Message, user.HTTPStatus())
	}
	y := FiscalYear{Code: "2028", StartDate: "2028-01-01", EndDate: "2028-12-31", IsActive: true, Scale: 2, ProfitLossAccount: "3200", RetainedEarningsAccount: "3100"}
	f.run(Command{Resource: "fiscal-years", Action: "create", FiscalYear: &y})
	var count int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM gl_records WHERE company='C' AND kind='journal-books'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("legacy company got %d default books (err=%v) — types would be inferred from codes", count, err)
	}
}
