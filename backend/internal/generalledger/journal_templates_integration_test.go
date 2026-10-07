//go:build integration

package generalledger

import (
	"encoding/json"
	"strings"
	"testing"
)

func (f *pgIntegrityFixture) journalTemplates() []Master {
	f.t.Helper()
	page, err := f.store.List(f.ctx, f.scope, JournalTemplateKind, "", 1, 100, ListFilter{})
	if err != nil {
		f.t.Fatal(err)
	}
	out := []Master{}
	for _, raw := range page.Items {
		var m Master
		if err := json.Unmarshal(raw, &m); err != nil {
			f.t.Fatal(err)
		}
		out = append(out, m)
	}
	return out
}

func rentTemplate() *Master {
	return &Master{Name: "  ค่าเช่าสำนักงานรายเดือน ", Template: &JournalTemplate{
		BookCode: "JV", Reference: "สัญญาเช่า-01", RefDate: "2026-01-01", Description: "ค่าเช่าสำนักงาน", TaxType: 1,
		Lines: []JournalTemplateLine{
			{AccountCode: "5000", Description: "ค่าเช่า", DepartmentCode: "D01", ProjectCode: "P01"},
			{AccountCode: "1000", Description: "จ่ายเงินสด"},
		},
	}}
}

// A journal template is a Champ "รูปแบบรายการรายวัน": created from the screen, listed for the picker, never a voucher.
func TestJournalTemplatesLifecycleIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)

	// 1. created with the typed name trimmed and a system number; only the pattern is stored
	first := f.run(Command{Resource: JournalTemplateKind, Action: "create", Master: rentTemplate()})
	templates := f.journalTemplates()
	if len(templates) != 1 || templates[0].ID != first.ID {
		t.Fatalf("templates = %+v", templates)
	}
	rent := templates[0]
	if rent.Code != "T0001" || rent.Name != "ค่าเช่าสำนักงานรายเดือน" || !rent.IsActive || rent.Kind != "" && rent.Kind != JournalTemplateKind {
		t.Fatalf("stored template header = %+v", rent)
	}
	if got := rent.Template; got == nil || got.BookCode != "JV" || got.Reference != "สัญญาเช่า-01" || got.RefDate != "2026-01-01" || got.TaxType != 1 || len(got.Lines) != 2 || got.Lines[0].DepartmentCode != "D01" || got.Lines[1].AccountCode != "1000" {
		t.Fatalf("stored pattern = %+v", got)
	}
	var stored string
	if err := f.db.QueryRow(`SELECT payload::text FROM gl_records WHERE company='C' AND kind=$1 AND id=$2`, JournalTemplateKind, first.ID).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"debit"`, `"credit"`, `"docno"`, `"date"`} {
		if strings.Contains(stored, forbidden) {
			t.Fatalf("a template must not keep %s: %s", forbidden, stored)
		}
	}

	// 2. a second template is numbered after the first; the same name is allowed (Champ keeps any text)
	second := f.run(Command{Resource: JournalTemplateKind, Action: "create", Master: rentTemplate()})
	if got := f.journalTemplates(); len(got) != 2 || got[1].Code != "T0002" || got[1].ID != second.ID {
		t.Fatalf("second template = %+v", got)
	}

	// 3. unrelated master fields a client sends are dropped, not parked in the record
	noisy := rentTemplate()
	noisy.AccountCode, noisy.Amount, noisy.Rules, noisy.Rows = "1000", Amount("99.99"), []MappingRule{{AccountCode: "1000", Side: "debit", Source: "x"}}, []StatementRow{{ID: "r", Title: "x"}}
	third := f.run(Command{Resource: JournalTemplateKind, Action: "create", Master: noisy})
	for _, m := range f.journalTemplates() {
		if m.ID == third.ID && (m.AccountCode != "" || len(m.Rules) != 0 || len(m.Rows) != 0 || m.Amount != "" && m.Amount != Amount("0")) {
			t.Fatalf("junk master fields stored: %+v", m)
		}
	}

	// 4. what Champ's name dialog and a voucher line would refuse
	blankName := rentTemplate()
	blankName.Name = "   "
	f.failCode(Command{Resource: JournalTemplateKind, Action: "create", Master: blankName}, "journal_template_name_required", "name")
	longName := rentTemplate()
	longName.Name = strings.Repeat("ก", JournalTemplateNameMaxRunes+1)
	f.failCode(Command{Resource: JournalTemplateKind, Action: "create", Master: longName}, "journal_template_name_too_long", "name")
	exactName := rentTemplate()
	exactName.Name = strings.Repeat("ก", JournalTemplateNameMaxRunes)
	f.run(Command{Resource: JournalTemplateKind, Action: "create", Master: exactName})
	noBody := &Master{Name: "ไม่มีเนื้อหา"}
	f.deny(f.scope, Command{Resource: JournalTemplateKind, Action: "create", Master: noBody})
	noLines := rentTemplate()
	noLines.Template.Lines = nil
	f.deny(f.scope, Command{Resource: JournalTemplateKind, Action: "create", Master: noLines})
	blankAccount := rentTemplate()
	blankAccount.Template.Lines[1].AccountCode = ""
	f.failCode(Command{Resource: JournalTemplateKind, Action: "create", Master: blankAccount}, "code_required", "lines[1].accountcode")
	unknownAccount := rentTemplate()
	unknownAccount.Template.Lines[1].AccountCode = "9999"
	_, err := f.execute(f.scope, Command{Resource: JournalTemplateKind, Action: "create", Master: unknownAccount})
	if err == nil || !strings.Contains(err.Error(), "บรรทัด 2") {
		t.Fatalf("unknown account: %v", err)
	}
	unknownBook := rentTemplate()
	unknownBook.Template.BookCode = "ไม่มีสมุด"
	f.failCode(Command{Resource: JournalTemplateKind, Action: "create", Master: unknownBook}, "journal_book_not_found", "bookcode")
	badTax := rentTemplate()
	badTax.Template.TaxType = 3
	f.deny(f.scope, Command{Resource: JournalTemplateKind, Action: "create", Master: badTax})
	if got := len(f.journalTemplates()); got != 4 {
		t.Fatalf("a refused template was stored: %d templates", got)
	}

	// 5. rename keeps the number; a code sent by the client is ignored; a stale version is refused
	renamed := rentTemplate()
	renamed.Name = "ค่าเช่าสำนักงาน (แก้ไข)"
	renamed.Code = "T9999"
	f.run(Command{Resource: JournalTemplateKind, Action: "update", ID: rent.ID, Version: rent.Version, Master: renamed})
	after := f.journalTemplates()[0]
	if after.Code != "T0001" || after.Name != "ค่าเช่าสำนักงาน (แก้ไข)" || after.Version != rent.Version+1 {
		t.Fatalf("renamed template = %+v", after)
	}
	f.deny(f.scope, Command{Resource: JournalTemplateKind, Action: "update", ID: rent.ID, Version: rent.Version, Master: renamed})

	// 6. delete hides it and the number is never reused
	f.run(Command{Resource: JournalTemplateKind, Action: "delete", ID: after.ID, Version: after.Version})
	for _, m := range f.journalTemplates() {
		if m.ID == after.ID {
			t.Fatal("deleted template still listed")
		}
	}
	next := f.run(Command{Resource: JournalTemplateKind, Action: "create", Master: rentTemplate()})
	for _, m := range f.journalTemplates() {
		if m.ID == next.ID && m.Code != "T0005" {
			t.Fatalf("code after delete = %s, want T0005 (no reuse)", m.Code)
		}
	}

	// 7. templates are not journals: nothing was posted or projected
	var lines, journals int
	if err := f.db.QueryRow(`SELECT (SELECT count(*) FROM gl_lines WHERE company='C'), (SELECT count(*) FROM gl_records WHERE company='C' AND kind='journals')`).Scan(&lines, &journals); err != nil {
		t.Fatal(err)
	}
	if lines != 0 || journals != 0 {
		t.Fatalf("templates created %d gl_lines and %d journals", lines, journals)
	}

	// 8. another company never sees them
	other := f.scope
	other.Company = "C2"
	page, err := f.store.List(f.ctx, other, JournalTemplateKind, "", 1, 100, ListFilter{})
	if err != nil || len(page.Items) != 0 {
		t.Fatalf("other company templates: %v %+v", err, page)
	}
}
