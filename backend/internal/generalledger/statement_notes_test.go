package generalledger

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// หมายเหตุประกอบงบการเงิน (แบบ 2 ข้อ 5): เลขที่ต้องมีและไม่ซ้ำในปีเดียวกัน, หัวข้อต้องมี, ความยาวนับเป็นตัวอักษร (ภาษาไทย 3 byte ต่อตัว)
func TestNormalizeStatementNotes(t *testing.T) {
	notes, err := normalizeStatementNotes([]StatementNote{
		{NoteNo: " 1 ", Title: "  ข้อมูลทั่วไป ", Body: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด\nสำนักงานใหญ่  "},
		{ID: "n2", NoteNo: "2", Title: "เกณฑ์ในการจัดทำและนำเสนองบการเงิน"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if notes[0].NoteNo != "1" || notes[0].Title != "ข้อมูลทั่วไป" || notes[0].ID == "" {
		t.Fatalf("trimmed note = %+v", notes[0])
	}
	// เนื้อหาเก็บตามที่ผู้ใช้พิมพ์ (บรรทัดใหม่/ช่องว่างท้ายบรรทัดไม่ถูกตัด)
	if notes[0].Body != "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด\nสำนักงานใหญ่  " || notes[1].ID != "n2" || notes[1].Body != "" {
		t.Fatalf("notes = %+v", notes)
	}

	thai := func(n int) string { return strings.Repeat("ก", n) }
	cases := []struct {
		name  string
		notes []StatementNote
		code  string
		field string
	}{
		{"missing noteno", []StatementNote{{NoteNo: "  ", Title: "ข้อมูลทั่วไป"}}, "statement_note_no_required", "notes[0].noteno"},
		{"duplicate noteno after trim", []StatementNote{{NoteNo: "1", Title: "ก"}, {NoteNo: " 1", Title: "ข"}}, "statement_note_no_duplicate", "notes[1].noteno"},
		{"noteno 11 Thai runes", []StatementNote{{NoteNo: "๑๒๓๔๕๖๗๘๙๐๑", Title: "ก"}}, "statement_note_no_too_long", "notes[0].noteno"},
		{"missing title", []StatementNote{{NoteNo: "1", Title: " \t "}}, "statement_note_title_required", "notes[0].title"},
		{"title 201 Thai runes", []StatementNote{{NoteNo: "1", Title: thai(201)}}, "statement_note_title_too_long", "notes[0].title"},
		{"body 20001 Thai runes", []StatementNote{{NoteNo: "1", Title: "ก", Body: thai(20001)}}, "statement_note_body_too_long", "notes[0].body"},
	}
	for _, c := range cases {
		_, err := normalizeStatementNotes(c.notes)
		user, ok := AsUserError(err)
		if !ok || user.Code != c.code || user.Field != c.field || user.HTTPStatus() != 400 {
			t.Fatalf("%s: got %v (%+v)", c.name, err, user)
		}
		if !strings.ContainsAny(user.Message, "กขคงจฉชซฌญฎฏฐฑฒณดตถทธนบปผฝพฟภมยรลวศษสหฬอฮ") {
			t.Fatalf("%s: message is not Thai: %q", c.name, user.Message)
		}
	}

	// ขอบเขตพอดีผ่าน: เลขที่ 10 ตัวอักษรไทย (30 byte), หัวข้อ 200, เนื้อหา 20000 ตัวอักษรไทย
	if _, err := normalizeStatementNotes([]StatementNote{{NoteNo: "๑๒๓๔๕๖๗๘๙๐", Title: thai(200), Body: thai(20000)}}); err != nil {
		t.Fatalf("limits counted in bytes: %v", err)
	}

	many := make([]StatementNote, 0, statementNotesMax+1)
	for i := 1; i <= statementNotesMax+1; i++ {
		many = append(many, StatementNote{NoteNo: strconv.Itoa(i), Title: "หัวข้อ"})
	}
	if _, err := normalizeStatementNotes(many[:statementNotesMax]); err != nil {
		t.Fatalf("100 notes rejected: %v", err)
	}
	if user, ok := AsUserError(func() error { _, err := normalizeStatementNotes(many); return err }()); !ok || user.Code != "statement_notes_too_many" {
		t.Fatalf("101 notes accepted or wrong code: %+v", user)
	}
}

func TestStatementNotesMasterContract(t *testing.T) {
	if MasterCollections["statement-notes"] == "" || !supportedRecord("statement-notes") {
		t.Fatal("statement-notes is not a stored master kind")
	}
	m := Master{Kind: "statement-notes", Code: ""}
	if user, ok := AsUserError(validateMasterCode(&m, false)); !ok || user.Code != "code_required" || !strings.Contains(user.Message, "ปีบัญชีของหมายเหตุประกอบงบการเงิน") {
		t.Fatalf("empty year code: %+v", user)
	}
	// NUL ในเนื้อหาหมายเหตุบอกชื่อช่องเป็นภาษาไทย
	cmd := Command{Resource: "statement-notes", Action: "create", Master: &Master{Code: "2569", Notes: []StatementNote{{NoteNo: "1", Title: "ข้อมูลทั่วไป", Body: "ข้อความ\x00"}}}}
	if user, ok := AsUserError(rejectNUL(cmd)); !ok || user.Field != "notes[0].body" || !strings.HasPrefix(user.Message, "เนื้อหาหมายเหตุ") {
		t.Fatalf("NUL in note body: %+v", user)
	}
}

// id ของหมายเหตุมาจาก client: ซ้ำหรือยาวเกิน 64 ตัวอักษร = ออก id ใหม่ (จอแก้ไขผูกหมายเหตุด้วย id); id ที่ถูกต้องคงเดิม
func TestNormalizeStatementNoteIDs(t *testing.T) {
	long := strings.Repeat("ก", statementNoteIDMaxRunes+1)
	notes, err := normalizeStatementNotes([]StatementNote{
		{ID: "n1", NoteNo: "1", Title: "ข้อมูลทั่วไป"},
		{ID: " n1 ", NoteNo: "2", Title: "เกณฑ์ในการจัดทำและนำเสนองบการเงิน"},
		{ID: long, NoteNo: "3", Title: "สรุปนโยบายการบัญชี"},
		{ID: strings.Repeat("ก", statementNoteIDMaxRunes), NoteNo: "4", Title: "ประมาณการทางบัญชี"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if notes[0].ID != "n1" || notes[1].ID == "n1" || notes[1].ID == "" || notes[2].ID == long || notes[3].ID != strings.Repeat("ก", statementNoteIDMaxRunes) {
		t.Fatalf("ids = %q %q %q %q", notes[0].ID, notes[1].ID, notes[2].ID, notes[3].ID)
	}
	seen := map[string]bool{}
	for _, note := range notes {
		if seen[note.ID] {
			t.Fatalf("duplicate id %q", note.ID)
		}
		seen[note.ID] = true
	}
}

// บรรทัดงบอ้างหมายเหตุได้หลายข้อ ("4, 5" / "4 และ 5" / "4;5") และหัวข้อย่อย "5.1" (เขียนรวมในหมายเหตุหลัก 5 ได้)
func TestStatementNoteReferences(t *testing.T) {
	rows := []map[string]string{
		{"noteno": "4, 5"}, {"noteno": " 5 และ 6 "}, {"noteno": "7;8"}, {"noteno": ""}, {"noteno": "5.1"}, {"noteno": "9.2"}, {"noteno": "4"},
	}
	referenced := statementReportNoteNos(rows)
	if want := []string{"4", "5", "6", "7", "8", "5.1", "9.2"}; !reflect.DeepEqual(referenced, want) {
		t.Fatalf("references = %q, want %q", referenced, want)
	}
	present := map[string]bool{"4": true, "5": true, "6": true, "7": true, "9.1": true}
	got := missingStatementNotes(referenced, present, "2026")
	want := []string{"หมายเหตุ 8 ที่อ้างในงบยังไม่มีในหมายเหตุประกอบงบการเงินปี 2026", "หมายเหตุ 9.2 ที่อ้างในงบยังไม่มีในหมายเหตุประกอบงบการเงินปี 2026"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("missing = %q, want %q", got, want)
	}
	if got := missingStatementNotes([]string{"4", "5"}, map[string]bool{"4": true, "5": true}, "2026"); len(got) != 0 {
		t.Fatalf("all present warned: %q", got)
	}
	if got := statementReportNoteNos(nil); len(got) != 0 {
		t.Fatalf("no rows = %q", got)
	}
}

// เลขที่หมายเหตุเทียบกันแบบปกติ: จุดท้าย ("5." = "5"), เลขไทย ("๕" = "5"), ช่องว่างไม่ตัดบรรทัดจาก Word/Excel ("4<NBSP>5" = 4 และ 5)
// และหัวข้อย่อยหลายชั้น ("5.1.2" นับว่ามีเมื่อมี "5.1") — เดิมเตือน "ยังไม่มี" ทั้งที่มีหมายเหตุนั้นแล้ว
func TestStatementNoteNumbersCompareCanonical(t *testing.T) {
	notes, err := normalizeStatementNotes([]StatementNote{{NoteNo: " 5. ", Title: "ลูกหนี้การค้า"}, {NoteNo: "๖", Title: "สินค้าคงเหลือ"}, {NoteNo: "7.1", Title: "ที่ดิน อาคารและอุปกรณ์"}})
	if err != nil {
		t.Fatal(err)
	}
	// เก็บตามที่พิมพ์ (เลขไทยคงไว้) แต่ตัดจุดท้าย เพราะหัวข้อที่พิมพ์เติม ". " ต่อท้ายเลขที่เอง
	if notes[0].NoteNo != "5" || notes[1].NoteNo != "๖" {
		t.Fatalf("stored noteno = %q %q", notes[0].NoteNo, notes[1].NoteNo)
	}
	for _, dup := range [][2]string{{"5", "๕"}, {"5.", "5"}, {"5.1", "๕.๑."}} {
		_, err := normalizeStatementNotes([]StatementNote{{NoteNo: dup[0], Title: "ก"}, {NoteNo: dup[1], Title: "ข"}})
		if user, ok := AsUserError(err); !ok || user.Code != "statement_note_no_duplicate" || user.Field != "notes[1].noteno" {
			t.Fatalf("%q vs %q: got %v", dup[0], dup[1], err)
		}
	}
	if _, err := normalizeStatementNotes([]StatementNote{{NoteNo: ".", Title: "ก"}}); err == nil {
		t.Fatal("noteno of only a dot accepted")
	}

	referenced := statementReportNoteNos([]map[string]string{{"noteno": "5"}, {"noteno": "4\u00a05"}, {"noteno": "๖"}, {"noteno": "7.1.2"}, {"noteno": "6."}, {"noteno": "8.1.2"}})
	if want := []string{"5", "4", "6", "7.1.2", "8.1.2"}; !reflect.DeepEqual(referenced, want) {
		t.Fatalf("references = %q, want %q", referenced, want)
	}
	present := map[string]bool{}
	for _, note := range append(notes, StatementNote{NoteNo: "4"}, StatementNote{NoteNo: "8.2"}) {
		present[canonicalStatementNoteNo(note.NoteNo)] = true
	}
	want := []string{"หมายเหตุ 8.1.2 ที่อ้างในงบยังไม่มีในหมายเหตุประกอบงบการเงินปี 2026"}
	if got := missingStatementNotes(referenced, present, "2026"); !reflect.DeepEqual(got, want) {
		t.Fatalf("missing = %q, want %q", got, want)
	}
}

// ไม่ติ๊ก "แสดงคอลัมน์หมายเหตุประกอบงบ" = งบไม่พิมพ์เลขที่หมายเหตุ จึงไม่มีเลขที่ให้ตรวจ; ไม่ระบุ = แสดงตามค่าเริ่ม
func TestStatementPrintedNoteNos(t *testing.T) {
	rows := []map[string]string{{"noteno": "9"}}
	hide, show := false, true
	for name, c := range map[string]struct {
		style *StatementGlobalStyle
		want  int
	}{"no style": {nil, 1}, "unset": {&StatementGlobalStyle{}, 1}, "shown": {&StatementGlobalStyle{ShowNoteColumn: &show}, 1}, "hidden": {&StatementGlobalStyle{ShowNoteColumn: &hide}, 0}} {
		if got := statementPrintedNoteNos(Master{GlobalStyle: c.style}, rows); len(got) != c.want {
			t.Fatalf("%s: printed note numbers = %q", name, got)
		}
	}
}
