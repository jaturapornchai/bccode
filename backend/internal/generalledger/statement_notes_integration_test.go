//go:build integration

package generalledger

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// หมายเหตุประกอบงบการเงินบน PostgreSQL จริง: หนึ่งรายการต่อบริษัทต่อปีบัญชี (code = ปีบัญชี) ผ่านคำสั่ง create/update/delete ปกติ
// ตรวจแถวใน gl_records ทีละขั้น (กฎ UAT + PostgreSQL)
func TestStatementNotesLifecycleIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	userCode := func(cmd Command) string {
		t.Helper()
		_, err := f.execute(f.scope, cmd)
		user, ok := AsUserError(err)
		if !ok {
			t.Fatalf("%s/%s: want user error, got %v", cmd.Resource, cmd.Action, err)
		}
		return user.Code
	}
	type stored struct {
		version   int64
		isDeleted bool
		notes     []StatementNote
	}
	row := func(id string) stored {
		t.Helper()
		var s stored
		var payload []byte
		if err := f.db.QueryRowContext(f.ctx, `SELECT version, COALESCE((payload->>'isdeleted')::boolean,false), payload FROM gl_records WHERE company=$1 AND kind='statement-notes' AND id=$2`, f.scope.Company, id).Scan(&s.version, &s.isDeleted, &payload); err != nil {
			t.Fatal(err)
		}
		var m Master
		if err := json.Unmarshal(payload, &m); err != nil {
			t.Fatal(err)
		}
		s.notes = m.Notes
		return s
	}
	countRows := func() int {
		t.Helper()
		var n int
		if err := f.db.QueryRowContext(f.ctx, `SELECT count(*) FROM gl_records WHERE company=$1 AND kind='statement-notes'`, f.scope.Company).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	// ปีบัญชีต้องมีอยู่จริง
	if code := userCode(Command{Resource: "statement-notes", Action: "create", Master: &Master{Code: "2030", Notes: []StatementNote{{NoteNo: "1", Title: "ข้อมูลทั่วไป"}}}}); code != "statement_notes_fiscal_year_not_found" {
		t.Fatalf("unknown year code = %s", code)
	}
	if countRows() != 0 {
		t.Fatal("rejected create wrote a row")
	}

	// Create → แถวถูกสร้างพร้อมเลขที่/หัวข้อที่ตัดช่องว่างแล้ว และ id ของหมายเหตุ
	body := "งบการเงินฉบับนี้จัดทำขึ้นตามมาตรฐานการรายงานทางการเงินสำหรับกิจการที่ไม่มีส่วนได้เสียสาธารณะ (TFRS for NPAEs)\nแสดงหน่วยเงินเป็นบาท"
	create := Master{Code: "2026", Name: "หมายเหตุประกอบงบการเงิน", IsActive: true, Notes: []StatementNote{{NoteNo: " 1 ", Title: "ข้อมูลทั่วไป", Body: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด"}, {NoteNo: "2", Title: "เกณฑ์ในการจัดทำและนำเสนองบการเงิน", Body: body}}}
	created := f.run(Command{Resource: "statement-notes", Action: "create", Master: &create})
	got := row(created.ID)
	if countRows() != 1 || got.version != 1 || got.isDeleted || len(got.notes) != 2 || got.notes[0].NoteNo != "1" || got.notes[0].ID == "" || got.notes[1].Body != body {
		t.Fatalf("created row = %+v", got)
	}
	raw, err := f.store.Get(f.ctx, f.scope, "statement-notes", created.ID)
	if err != nil {
		t.Fatal(err)
	}
	var read Master
	if err = json.Unmarshal(raw, &read); err != nil || read.Code != "2026" || len(read.Notes) != 2 {
		t.Fatalf("get = %s (%v)", raw, err)
	}
	page, err := f.store.List(f.ctx, f.scope, "statement-notes", "", 1, 100, ListFilter{})
	if err != nil || page.Total != 1 {
		t.Fatalf("list = %+v (%v)", page, err)
	}

	// ปีเดียวกันมีได้รายการเดียว: ข้อความบอกให้โหลดหมายเหตุล่าสุด (รหัสคือปีบัญชี ผู้ใช้เปลี่ยนเองไม่ได้ — ห้ามบอก "ใช้รหัสอื่น")
	if code := userCode(Command{Resource: "statement-notes", Action: "create", Master: &Master{Code: "2026", Notes: []StatementNote{{NoteNo: "1", Title: "ข้อมูลทั่วไป"}}}}); code != CodeDuplicateCode {
		t.Fatalf("second record for 2026 = %s", code)
	}
	if _, err := f.execute(f.scope, Command{Resource: "statement-notes", Action: "create", Master: &Master{Code: "2026", Notes: []StatementNote{{NoteNo: "1", Title: "ข้อมูลทั่วไป"}}}}); err == nil || err.Error() != "ปีบัญชีนี้มีหมายเหตุประกอบงบการเงินอยู่แล้ว กรุณาโหลดหมายเหตุล่าสุดก่อนแก้ไข" {
		t.Fatalf("second record message = %v", err)
	}

	// Update ที่ผิด (เลขที่ซ้ำ) ไม่เปลี่ยนแถวเดิม
	bad := read
	bad.Notes = append(append([]StatementNote{}, read.Notes...), StatementNote{NoteNo: "2", Title: "ซ้ำ"})
	if code := userCode(Command{Resource: "statement-notes", Action: "update", ID: created.ID, Version: created.Version, Master: &bad}); code != "statement_note_no_duplicate" {
		t.Fatalf("duplicate noteno update = %s", code)
	}
	if after := row(created.ID); after.version != 1 || len(after.notes) != 2 {
		t.Fatalf("rejected update changed the row: %+v", after)
	}

	// Update → แถวเดิม version 2 มีหมายเหตุ 3 ข้อ (ไม่ใช่แถวใหม่)
	next := read
	next.Notes = append(append([]StatementNote{}, read.Notes...), StatementNote{NoteNo: "3", Title: "สรุปนโยบายการบัญชี", Body: "รับรู้รายได้เมื่อส่งมอบสินค้า"})
	updated := f.run(Command{Resource: "statement-notes", Action: "update", ID: created.ID, Version: created.Version, Master: &next})
	if got = row(created.ID); updated.ID != created.ID || got.version != 2 || len(got.notes) != 3 || got.notes[0].ID != read.Notes[0].ID || countRows() != 1 {
		t.Fatalf("updated row = %+v (result %+v)", got, updated)
	}
	if code := userCode(Command{Resource: "statement-notes", Action: "update", ID: created.ID, Version: created.Version, Master: &next}); code != CodeStaleVersion {
		t.Fatalf("stale update = %s", code)
	}

	// Delete → แถวถูกทำเครื่องหมายลบ อ่านไม่เจอ และสร้างปีเดิมใหม่ได้
	f.run(Command{Resource: "statement-notes", Action: "delete", ID: created.ID, Version: updated.Version, Reason: "ลบหมายเหตุ"})
	if got = row(created.ID); !got.isDeleted {
		t.Fatalf("deleted row = %+v", got)
	}
	if _, err = f.store.Get(f.ctx, f.scope, "statement-notes", created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get deleted = %v", err)
	}
	again := Master{Code: "2026", Notes: []StatementNote{{NoteNo: "1", Title: "ข้อมูลทั่วไป"}}}
	if recreated := f.run(Command{Resource: "statement-notes", Action: "create", Master: &again}); recreated.ID == created.ID {
		t.Fatal("recreate reused the deleted id")
	}

	// ข้อมูลหลักชนิดอื่นไม่เก็บ notes; ไม่ติ๊ก "แสดงคอลัมน์หมายเหตุประกอบงบ" ต้องเก็บ false ได้จริง (เดิม omitempty ทิ้งค่า)
	hideNotes := false
	template := Master{Code: "FS-N", Name: "งบกำไรขาดทุน", StatementType: "pnl", GlobalStyle: &StatementGlobalStyle{Scale: 2, ShowNoteColumn: &hideNotes}, Notes: []StatementNote{{NoteNo: "1", Title: "ไม่ควรเก็บ"}}}
	tr := f.run(Command{Resource: "statement-templates", Action: "create", Master: &template})
	var notesInTemplate bool
	var showNoteColumn sql.NullString
	if err = f.db.QueryRowContext(f.ctx, `SELECT payload ? 'notes', payload->'globalstyle'->>'shownotecolumn' FROM gl_records WHERE company=$1 AND id=$2`, f.scope.Company, tr.ID).Scan(&notesInTemplate, &showNoteColumn); err != nil || notesInTemplate {
		t.Fatalf("statement template stored notes: %v %v", notesInTemplate, err)
	}
	if !showNoteColumn.Valid || showNoteColumn.String != "false" {
		t.Fatalf("shownotecolumn stored = %+v, want false", showNoteColumn)
	}
}

// ปีบัญชีที่มีหมายเหตุประกอบงบการเงินลบไม่ได้ (หมายเหตุผูกปีด้วย code): ไม่งั้นหมายเหตุค้างเปิดจากจอไม่ได้และกลับมาเองเมื่อสร้างปีรหัสเดิม;
// ลบหมายเหตุแล้วลบปีได้ตามปกติ — ตรวจแถวใน gl_records ทีละขั้น
func TestStatementNotesBlockFiscalYearDeleteIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	yearDeleted := func() bool {
		t.Helper()
		var deleted bool
		if err := f.db.QueryRowContext(f.ctx, `SELECT COALESCE((payload->>'isdeleted')::boolean,false) FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND code='2027'`, f.scope.Company).Scan(&deleted); err != nil {
			t.Fatal(err)
		}
		return deleted
	}
	notes := Master{Code: "2027", Notes: []StatementNote{{NoteNo: "1", Title: "ข้อมูลทั่วไป", Body: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด"}}}
	created := f.run(Command{Resource: "statement-notes", Action: "create", Master: &notes})
	year := f.years["2027"]
	_, err := f.execute(f.scope, Command{Resource: "fiscal-years", Action: "delete", ID: year.ID, Version: year.Version, Reason: "ลบปีบัญชี"})
	if err == nil || err.Error() != "ปีบัญชีมีข้อมูลอ้างอิง ลบไม่ได้" {
		t.Fatalf("delete year with notes = %v", err)
	}
	if yearDeleted() {
		t.Fatal("fiscal year 2027 was deleted while its notes exist")
	}
	f.run(Command{Resource: "statement-notes", Action: "delete", ID: created.ID, Version: created.Version, Reason: "ลบหมายเหตุ"})
	f.run(Command{Resource: "fiscal-years", Action: "delete", ID: year.ID, Version: year.Version, Reason: "ลบปีบัญชี"})
	if !yearDeleted() {
		t.Fatal("fiscal year 2027 not deleted after its notes were deleted")
	}
}

// งบการเงินเตือนเมื่อบรรทัดอ้างเลขหมายเหตุที่ไม่มีในหมายเหตุประกอบงบการเงินของปีที่ออกงบ (ทั้งงบปกติและงบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น)
func TestStatementNoteReferenceWarningsIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	rows := []StatementRow{
		{ID: "a", RowNo: 10, RowType: "account", Title: "รายได้จากการขาย", NoteNo: "1", AccountCodes: []string{"4000"}},
		{ID: "b", RowNo: 20, RowType: "account", Title: "ค่าใช้จ่ายในการขายและบริหาร", NoteNo: " 2 ", AccountCodes: []string{"5000"}},
		{ID: "c", RowNo: 30, RowType: "formula", Title: "กำไร (ขาดทุน) สุทธิ", NoteNo: "1", Formula: "R10 - R20"},
	}
	templates := []Master{
		{Code: "PL-N", Name: "งบกำไรขาดทุน", StatementType: "pnl", IsActive: true, GlobalStyle: &StatementGlobalStyle{Scale: 2}, Rows: rows},
		{Code: "EQ-N", Name: "งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น", StatementType: "equity", IsActive: true, GlobalStyle: &StatementGlobalStyle{Scale: 2},
			Columns: []StatementColumn{{ID: "cap", Title: "ทุนที่ชำระแล้ว", AccountCodes: []string{"3000"}}},
			Rows:    []StatementRow{{ID: "a", RowNo: 10, RowType: "account", Title: "ยอดคงเหลือ ณ ต้นงวด", NoteNo: "1", AmountBasis: "opening"}, {ID: "b", RowNo: 20, RowType: "account", Title: "ยอดคงเหลือ ณ ปลายงวด", NoteNo: "2", AmountBasis: "closing"}}},
		{Code: "PL-NONE", Name: "งบกำไรขาดทุน", StatementType: "pnl", IsActive: true, GlobalStyle: &StatementGlobalStyle{Scale: 2}, Rows: []StatementRow{{ID: "a", RowNo: 10, RowType: "account", Title: "รายได้จากการขาย", AccountCodes: []string{"4000"}}}},
	}
	for _, m := range templates {
		m := m
		f.run(Command{Resource: "statement-templates", Action: "create", Master: &m})
	}
	noteWarnings := func(template string) []string {
		t.Helper()
		report, err := f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2026", Template: template})
		if err != nil {
			t.Fatalf("%s: %v", template, err)
		}
		out := []string{}
		for _, warning := range report.Warnings {
			if strings.Contains(warning, "หมายเหตุ") {
				out = append(out, warning)
			}
		}
		return out
	}
	expect := func(want ...string) {
		t.Helper()
		for _, template := range []string{"PL-N", "EQ-N"} {
			if got := noteWarnings(template); strings.Join(got, "|") != strings.Join(want, "|") {
				t.Fatalf("%s warnings = %q, want %q", template, got, want)
			}
		}
		if got := noteWarnings("PL-NONE"); len(got) != 0 {
			t.Fatalf("template without note references warned: %q", got)
		}
	}
	missingRecord := "ยังไม่มีหมายเหตุประกอบงบการเงินของปี 2026 แต่มีบรรทัดอ้างหมายเหตุ"

	// ไม่มีหมายเหตุของปีนี้ = คำเตือนเดียว; หมายเหตุของปีอื่นไม่นับ (ปีปัจจุบันเท่านั้น)
	expect(missingRecord)
	other := Master{Code: "2027", Notes: []StatementNote{{NoteNo: "1", Title: "ข้อมูลทั่วไป"}, {NoteNo: "2", Title: "เกณฑ์ในการจัดทำและนำเสนองบการเงิน"}}}
	f.run(Command{Resource: "statement-notes", Action: "create", Master: &other})
	expect(missingRecord)

	// มีหมายเหตุ 1 แต่ไม่มี 2 = เตือนเฉพาะเลขที่ขาด (เลขที่ซ้ำในงบเตือนครั้งเดียว)
	notes := Master{Code: "2026", Notes: []StatementNote{{NoteNo: "1", Title: "ข้อมูลทั่วไป"}}}
	created := f.run(Command{Resource: "statement-notes", Action: "create", Master: &notes})
	expect("หมายเหตุ 2 ที่อ้างในงบยังไม่มีในหมายเหตุประกอบงบการเงินปี 2026")

	// ครบทุกเลขที่ = ไม่มีคำเตือนเรื่องหมายเหตุ
	notes.Notes = append(notes.Notes, StatementNote{NoteNo: "2", Title: "เกณฑ์ในการจัดทำและนำเสนองบการเงิน"})
	updated := f.run(Command{Resource: "statement-notes", Action: "update", ID: created.ID, Version: created.Version, Master: &notes})
	expect()

	// ตรวจเฉพาะบรรทัดที่พิมพ์จริง และบรรทัดที่อ้างหลายข้อ: แถวค่าใช้จ่ายเป็นศูนย์ถูกซ่อน (hidezerorows) จึงไม่ต้องมีหมายเหตุ 9;
	// แถวรายได้อ้าง "1, 2" ครบทั้งสองข้อ = ไม่มีคำเตือน
	sale := Journal{DocNo: "JV26-N", Date: "2026-03-10", BookCode: "JV", FiscalYear: "2026", Description: "ขายวัสดุก่อสร้าง รับชำระเงินสดหน้าร้าน", Kind: "manual", BranchCode: "B1",
		Lines: []Line{{AccountCode: "1000", Debit: Amount("100"), Credit: Amount("0")}, {AccountCode: "4000", Debit: Amount("0"), Credit: Amount("100")}}}
	f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &sale}).ID))
	hideZero := Master{Code: "PL-NHZ", Name: "งบกำไรขาดทุน", StatementType: "pnl", IsActive: true, GlobalStyle: &StatementGlobalStyle{Scale: 2, HideZeroRows: true}, Rows: []StatementRow{
		{ID: "a", RowNo: 10, RowType: "account", Title: "รายได้จากการขาย", NoteNo: "1, 2", AccountCodes: []string{"4000"}},
		{ID: "b", RowNo: 20, RowType: "account", Title: "ค่าใช้จ่ายในการขายและบริหาร", NoteNo: "9", AccountCodes: []string{"5000"}},
	}}
	f.run(Command{Resource: "statement-templates", Action: "create", Master: &hideZero})
	if got := noteWarnings("PL-NHZ"); len(got) != 0 {
		t.Fatalf("PL-NHZ note warnings = %q", got)
	}
	printed, err := f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2026", Template: "PL-NHZ"})
	if err != nil || len(printed.Rows) != 1 || printed.Rows[0]["rowno"] != "10" || printed.Rows[0]["noteno"] != "1, 2" {
		t.Fatalf("PL-NHZ rows = %v (%v)", printed.Rows, err)
	}

	// ไม่ติ๊ก "แสดงคอลัมน์หมายเหตุประกอบงบ" = งบไม่พิมพ์เลขที่หมายเหตุ จึงไม่เตือนเรื่องหมายเหตุ 9 ที่ยังไม่มี (ทั้งงบปกติและงบส่วนของผู้ถือหุ้น)
	hideColumn := false
	for _, m := range []Master{
		{Code: "PL-NOCOL", Name: "งบกำไรขาดทุน", StatementType: "pnl", IsActive: true, GlobalStyle: &StatementGlobalStyle{Scale: 2, ShowNoteColumn: &hideColumn},
			Rows: []StatementRow{{ID: "a", RowNo: 10, RowType: "account", Title: "รายได้จากการขาย", NoteNo: "9", AccountCodes: []string{"4000"}}}},
		{Code: "EQ-NOCOL", Name: "งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น", StatementType: "equity", IsActive: true, GlobalStyle: &StatementGlobalStyle{Scale: 2, ShowNoteColumn: &hideColumn},
			Columns: []StatementColumn{{ID: "cap", Title: "ทุนที่ชำระแล้ว", AccountCodes: []string{"3000"}}},
			Rows:    []StatementRow{{ID: "a", RowNo: 10, RowType: "account", Title: "ยอดคงเหลือ ณ ต้นงวด", NoteNo: "9", AmountBasis: "opening"}}},
	} {
		m := m
		f.run(Command{Resource: "statement-templates", Action: "create", Master: &m})
		if got := noteWarnings(m.Code); len(got) != 0 {
			t.Fatalf("%s with the note column hidden warned: %q", m.Code, got)
		}
	}
	showColumn := true
	shown := Master{Code: "PL-COL", Name: "งบกำไรขาดทุน", StatementType: "pnl", IsActive: true, GlobalStyle: &StatementGlobalStyle{Scale: 2, ShowNoteColumn: &showColumn},
		Rows: []StatementRow{{ID: "a", RowNo: 10, RowType: "account", Title: "รายได้จากการขาย", NoteNo: "9", AccountCodes: []string{"4000"}}}}
	f.run(Command{Resource: "statement-templates", Action: "create", Master: &shown})
	if got := noteWarnings("PL-COL"); len(got) != 1 || got[0] != "หมายเหตุ 9 ที่อ้างในงบยังไม่มีในหมายเหตุประกอบงบการเงินปี 2026" {
		t.Fatalf("PL-COL note warnings = %q", got)
	}

	// ลบหมายเหตุแล้ว = นับว่าไม่มี
	f.run(Command{Resource: "statement-notes", Action: "delete", ID: created.ID, Version: updated.Version, Reason: "ลบหมายเหตุ"})
	expect(missingRecord)
}
