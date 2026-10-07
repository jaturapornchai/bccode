//go:build integration

package generalledger

import (
	"encoding/json"
	"strings"
	"testing"
)

func (f *pgIntegrityFixture) dimensions(kind, from, to string) ([]string, int64) {
	f.t.Helper()
	page, err := f.store.List(f.ctx, f.scope, kind, "", 1, 100, ListFilter{CodeFrom: from, CodeTo: to})
	if err != nil {
		f.t.Fatal(err)
	}
	codes := []string{}
	for _, raw := range page.Items {
		var m Master
		if err := json.Unmarshal(raw, &m); err != nil {
			f.t.Fatal(err)
		}
		codes = append(codes, m.Code)
	}
	return codes, page.Total
}

func (f *pgIntegrityFixture) dimension(kind, code string) Master {
	f.t.Helper()
	page, err := f.store.List(f.ctx, f.scope, kind, "", 1, 100, ListFilter{CodeFrom: code})
	if err != nil || len(page.Items) != 1 {
		f.t.Fatalf("%s %s: %v, %d rows", kind, code, err, len(page.Items))
	}
	var m Master
	if err = json.Unmarshal(page.Items[0], &m); err != nil {
		f.t.Fatal(err)
	}
	return m
}

func (f *pgIntegrityFixture) createDimension(kind, code, name string) Result {
	f.t.Helper()
	return f.run(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: code, Name: name, IsActive: true}})
}

// dimensionJournal is a draft whose first line names the code in the register under test (department or project).
func dimensionJournal(kind, doc, code string) Journal {
	j := Journal{DocNo: doc, Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "ค่าใช้จ่ายสำนักงานประจำเดือน", Kind: "manual", BranchCode: "B1",
		Lines: []Line{{AccountCode: "5000", Debit: Amount("100"), Credit: Amount("0")}, {AccountCode: "1000", Debit: Amount("0"), Credit: Amount("100")}}}
	if kind == DepartmentKind {
		j.Lines[0].DepartmentCode = code
	} else {
		j.Lines[0].ProjectCode = code
	}
	return j
}

// Departments and projects (Champ BCDepartment / BCProject): create → duplicate → length limits → from–to list →
// delete blocked while a draft, a posted voucher, a template or a budget names the code → delete once unused.
func TestDimensionRegistersLifecycleIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	for _, kind := range []string{DepartmentKind, ProjectKind} {
		prefix := "D"
		if kind == ProjectKind {
			prefix = "P"
		}
		code := func(n string) string { return prefix + n }

		// 1. created with the code and name trimmed; only the contract fields are stored
		created := f.createDimension(kind, "  "+code("01")+" ", "  ฝ่ายบัญชีและการเงิน ")
		first := f.dimension(kind, code("01"))
		if first.ID != created.ID || first.Code != code("01") || first.Name != "ฝ่ายบัญชีและการเงิน" || !first.IsActive {
			t.Fatalf("%s stored = %+v", kind, first)
		}
		var stored string
		if err := f.db.QueryRow(`SELECT payload::text FROM gl_records WHERE company='C' AND kind=$1 AND id=$2`, kind, created.ID).Scan(&stored); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(stored, `"accountcode"`) || strings.Contains(stored, `"booktype"`) {
			t.Fatalf("%s payload carries foreign master fields: %s", kind, stored)
		}

		// 2. duplicate code in the same company; the same code is free in the other register
		f.failCode(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: code("01"), Name: "ซ้ำ", IsActive: true}}, CodeDuplicateCode, "")

		// 3. code and name limits (Champ VARCHAR(10) / VARCHAR(50)), counted in runes, Thai marks included
		f.failCode(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: "", Name: "ไม่มีรหัส", IsActive: true}}, "code_required", "code")
		f.failCode(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: prefix + "12345678901", Name: "รหัสยาวเกิน", IsActive: true}}, "code_too_long", "code")
		f.failCode(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: prefix + " 9", Name: "รหัสมีช่องว่าง", IsActive: true}}, "code_has_space", "code")
		f.failCode(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: code("98"), Name: "   ", IsActive: true}}, "dimension_name_required", "name")
		f.failCode(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: code("98"), Name: strings.Repeat("กิ", 26), IsActive: true}}, "dimension_name_too_long", "name")
		f.createDimension(kind, prefix+"123456789", strings.Repeat("กิ", 25)) // exactly 10 and 50 runes
		longName := first
		longName.Name = strings.Repeat("ก", 51)
		f.failCode(Command{Resource: kind, Action: "update", ID: first.ID, Version: first.Version, Master: &longName}, "dimension_name_too_long", "name")
		if got := f.dimension(kind, code("01")); got.Name != first.Name || got.Version != first.Version {
			t.Fatalf("%s rejected rename changed the record: %+v", kind, got)
		}

		// 4. from–to list: from alone = exact match, from+to = inclusive range, to alone ignored, ordered by code
		f.createDimension(kind, code("02"), "ฝ่ายขาย")
		f.createDimension(kind, code("03"), "ฝ่ายผลิต")
		f.createDimension(kind, code("10"), "ฝ่ายบริหาร")
		all, total := f.dimensions(kind, "", "")
		wantAll := []string{code("01"), code("02"), code("03"), code("10"), prefix + "123456789"}
		if strings.Join(all, ",") != strings.Join(wantAll, ",") || total != int64(len(wantAll)) {
			t.Fatalf("%s all = %v (%d), want %v", kind, all, total, wantAll)
		}
		if got, n := f.dimensions(kind, " "+code("02")+" ", ""); strings.Join(got, ",") != code("02") || n != 1 {
			t.Fatalf("%s exact = %v (%d)", kind, got, n)
		}
		if got, n := f.dimensions(kind, code("02"), code("10")); strings.Join(got, ",") != strings.Join([]string{code("02"), code("03"), code("10")}, ",") || n != 3 {
			t.Fatalf("%s range = %v (%d)", kind, got, n)
		}
		if got, n := f.dimensions(kind, "", code("02")); len(got) != len(wantAll) || n != int64(len(wantAll)) {
			t.Fatalf("%s to-only must be ignored like Champ, got %v (%d)", kind, got, n)
		}
		if got, n := f.dimensions(kind, code("99"), ""); len(got) != 0 || n != 0 {
			t.Fatalf("%s unknown exact code = %v (%d)", kind, got, n)
		}

		// 5. a draft voucher line uses D01 → cannot delete or recode; rename and deactivate stay allowed
		draft := dimensionJournal(kind, "JV-"+prefix+"1", code("01"))
		made := f.run(Command{Resource: "journals", Action: "create", Journal: &draft})
		f.failCode(Command{Resource: kind, Action: "delete", ID: first.ID, Version: first.Version}, "dimension_in_use_delete", "code")
		recoded := first
		recoded.Code = code("11")
		f.failCode(Command{Resource: kind, Action: "update", ID: first.ID, Version: first.Version, Master: &recoded}, "dimension_in_use_code", "code")
		renamed := first
		renamed.Name = "ฝ่ายบัญชี (เลิกใช้)"
		renamed.IsActive = false
		f.run(Command{Resource: kind, Action: "update", ID: first.ID, Version: first.Version, Master: &renamed})
		if got := f.dimension(kind, code("01")); got.Name != renamed.Name || got.IsActive {
			t.Fatalf("%s rename/deactivate not stored: %+v", kind, got)
		}
		// deleting the draft frees the code
		saved := f.journal(made.ID)
		f.run(Command{Resource: "journals", Action: "delete", ID: saved.ID, Version: saved.Version})
		current := f.dimension(kind, code("01"))
		f.run(Command{Resource: kind, Action: "delete", ID: current.ID, Version: current.Version})
		if got, _ := f.dimensions(kind, code("01"), ""); len(got) != 0 {
			t.Fatalf("%s unused code not deleted: %v", kind, got)
		}

		// 6. a posted voucher (ledger lines) blocks delete for good
		posted := dimensionJournal(kind, "JV-"+prefix+"2", code("02"))
		madePosted := f.run(Command{Resource: "journals", Action: "create", Journal: &posted})
		f.post(f.journal(madePosted.ID))
		second := f.dimension(kind, code("02"))
		f.failCode(Command{Resource: kind, Action: "delete", ID: second.ID, Version: second.Version}, "dimension_in_use_delete", "code")
		if n := f.countRows(`SELECT count(*) FROM gl_lines WHERE company='C' AND `+dimensionKinds[kind].column+`=$1`, code("02")); n != 1 {
			t.Fatalf("%s expected 1 ledger line using %s, got %d", kind, code("02"), n)
		}

		// 7. a saved journal template blocks delete (its lines name the code) — a template is a pattern, never a voucher
		tpl := rentTemplate()
		if kind == DepartmentKind {
			tpl.Template.Lines[0].DepartmentCode, tpl.Template.Lines[0].ProjectCode = code("03"), ""
		} else {
			tpl.Template.Lines[0].DepartmentCode, tpl.Template.Lines[0].ProjectCode = "", code("03")
		}
		f.run(Command{Resource: JournalTemplateKind, Action: "create", Master: tpl})
		third := f.dimension(kind, code("03"))
		f.failCode(Command{Resource: kind, Action: "delete", ID: third.ID, Version: third.Version}, "dimension_in_use_delete", "code")

		// 8. a budget header blocks delete (gl_budgets keeps its own department/project columns)
		budget := Budget{Code: "BG-" + prefix, Name: "งบประมาณประจำปี", FiscalYear: "2026", BranchCode: "B1", Lines: []BudgetLine{{AccountCode: "5000", Periods: budgetMonths("10")}}}
		if kind == DepartmentKind {
			budget.DepartmentCode = code("10")
		} else {
			budget.ProjectCode = code("10")
		}
		f.run(Command{Resource: "budgets", Action: "create", Budget: &budget})
		tenth := f.dimension(kind, code("10"))
		f.failCode(Command{Resource: kind, Action: "delete", ID: tenth.ID, Version: tenth.Version}, "dimension_in_use_delete", "code")

		// 9. an unused code recodes and deletes freely
		last := f.dimension(kind, prefix+"123456789")
		moved := last
		moved.Code = code("55")
		f.run(Command{Resource: kind, Action: "update", ID: last.ID, Version: last.Version, Master: &moved})
		moved = f.dimension(kind, code("55"))
		f.run(Command{Resource: kind, Action: "delete", ID: moved.ID, Version: moved.Version})
		left, _ := f.dimensions(kind, "", "")
		if strings.Join(left, ",") != strings.Join([]string{code("02"), code("03"), code("10")}, ",") {
			t.Fatalf("%s remaining = %v", kind, left)
		}
	}

	// the two registers never see each other's records
	if got, _ := f.dimensions(DepartmentKind, "P02", ""); len(got) != 0 {
		t.Fatalf("project code leaked into departments: %v", got)
	}
}
