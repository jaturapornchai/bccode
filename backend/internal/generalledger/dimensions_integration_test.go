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

// enableDimensions turns on the company switches of the registers (design §3.4); the other switches keep their value.
func (f *pgIntegrityFixture) enableDimensions(kinds ...string) {
	f.t.Helper()
	view, err := f.store.Projection().DimensionSettings(f.ctx, f.scope)
	if err != nil {
		f.t.Fatal(err)
	}
	input := view.DimensionSettingsInput
	for _, kind := range kinds {
		setSwitch(&input, kind)
	}
	f.run(Command{Resource: DimensionSettingsKind, Action: "update", ID: dimensionSettingsID, Version: view.Version, DimensionSettings: &input})
}

// ensureDimension registers a code a fixture journal names (a line may only use active codes of the register).
func (f *pgIntegrityFixture) ensureDimension(kind, code string) {
	f.t.Helper()
	page, err := f.store.List(f.ctx, f.scope, kind, "", 1, 1, ListFilter{CodeFrom: code})
	if err != nil {
		f.t.Fatal(err)
	}
	if len(page.Items) == 0 {
		name := "ฝ่ายบริหารทั่วไป"
		if kind == ProjectKind {
			name = "โครงการปรับปรุงสำนักงานใหญ่"
		}
		f.createDimension(kind, code, name)
	}
}

func (f *pgIntegrityFixture) createDimension(kind, code, name string) Result {
	f.t.Helper()
	return f.run(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: code, Name: name, IsActive: true}})
}

// dimensionJournal is a draft whose first line names the code in the register under test (any of the six).
func dimensionJournal(kind, doc, code string) Journal {
	j := Journal{DocNo: doc, Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Description: "ค่าใช้จ่ายสำนักงานประจำเดือน", Kind: "manual", BranchCode: "B1",
		Lines: []Line{{AccountCode: "5000", Debit: Amount("100"), Credit: Amount("0")}, {AccountCode: "1000", Debit: Amount("0"), Credit: Amount("100")}}}
	*j.Lines[0].lineCodePtr(kind) = code
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

// updateAccount saves the fixture account with the given edit and keeps its version current.
func (f *pgIntegrityFixture) updateAccount(code string, edit func(a *Account)) {
	f.t.Helper()
	a := f.accounts[code]
	edit(&a)
	r := f.run(Command{Resource: "accounts", Action: "update", ID: a.ID, Version: a.Version, Account: &a})
	a.Version = r.Version
	f.accounts[code] = a
}

func (f *pgIntegrityFixture) failAccountUpdate(code string, edit func(a *Account), wantCode, wantField string) *UserError {
	f.t.Helper()
	a := f.accounts[code]
	edit(&a)
	return f.failCode(Command{Resource: "accounts", Action: "update", ID: a.ID, Version: a.Version, Account: &a}, wantCode, wantField)
}

func (f *pgIntegrityFixture) deactivateDimension(kind, code string) {
	f.t.Helper()
	m := f.dimension(kind, code)
	m.IsActive = false
	f.run(Command{Resource: kind, Action: "update", ID: m.ID, Version: m.Version, Master: &m})
}

// A journal line may only name an active code of its register (Champ OnCheckComboData glfrmjournal.cpp:2913-2917) and,
// when the account lists codes, one of them. An empty list means every code; the flag-off rule still comes first.
func TestJournalLineDimensionCodesIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	for _, kind := range []string{DepartmentKind, ProjectKind} {
		prefix, jsonField := "D", "lines[0].departmentcode"
		wantNotFound, wantNotAllowed := "journal_line_depart_not_found", "journal_line_depart_not_allowed"
		setList := func(a *Account, codes []string) { a.DepartCodes, a.DepartAllCodes = codes, len(codes) == 0 }
		if kind == ProjectKind {
			prefix, jsonField = "P", "lines[0].projectcode"
			wantNotFound, wantNotAllowed = "journal_line_project_not_found", "journal_line_project_not_allowed"
			setList = func(a *Account, codes []string) { a.ProjectCodes, a.ProjectAllCodes = codes, len(codes) == 0 }
		}
		code := func(n string) string { return prefix + n }
		f.createDimension(kind, code("01"), "ฝ่ายบัญชีและการเงิน")
		f.createDimension(kind, code("02"), "ฝ่ายขาย")
		f.createDimension(kind, code("03"), "ฝ่ายผลิต")
		f.deactivateDimension(kind, code("03"))
		doc := func(n string) string { return "JV-" + prefix + n }

		// 1. a code that is not in the register, or is inactive, is refused with the line number and the code
		unknown := dimensionJournal(kind, doc("U"), code("99"))
		user := f.failCode(Command{Resource: "journals", Action: "create", Journal: &unknown}, wantNotFound, jsonField)
		if len(user.Args) != 2 || user.Args[0] != "1" || user.Args[1] != code("99") || !strings.Contains(user.Message, "บรรทัด 1") {
			t.Fatalf("%s unknown code error = %+v", kind, user)
		}
		inactive := dimensionJournal(kind, doc("I"), code("03"))
		f.failCode(Command{Resource: "journals", Action: "create", Journal: &inactive}, wantNotFound, jsonField)

		// 2. "เลือกทั้งหมด" on the account (the fixture ticks it) = every active register code
		anyA := dimensionJournal(kind, doc("A"), code("01"))
		f.run(Command{Resource: "journals", Action: "create", Journal: &anyA})
		anyB := dimensionJournal(kind, doc("B"), code("02"))
		made := f.run(Command{Resource: "journals", Action: "create", Journal: &anyB})

		// 3. a list narrows the codes: outside the list refused on create and on editing a draft, inside accepted
		f.updateAccount("5000", func(a *Account) { setList(a, []string{code("01")}) })
		outside := dimensionJournal(kind, doc("O"), code("02"))
		user = f.failCode(Command{Resource: "journals", Action: "create", Journal: &outside}, wantNotAllowed, jsonField)
		if len(user.Args) != 3 || user.Args[1] != "5000" || user.Args[2] != code("02") {
			t.Fatalf("%s not-allowed error args = %+v", kind, user.Args)
		}
		inside := dimensionJournal(kind, doc("N"), code("01"))
		f.run(Command{Resource: "journals", Action: "create", Journal: &inside})
		draft := f.journal(made.ID) // saved earlier with the second code, before the list existed
		f.failCode(Command{Resource: "journals", Action: "update", ID: draft.ID, Version: draft.Version, Journal: &draft}, wantNotAllowed, jsonField)
		draft.Lines[0].DepartmentCode, draft.Lines[0].ProjectCode = "", ""
		f.run(Command{Resource: "journals", Action: "update", ID: draft.ID, Version: draft.Version, Journal: &draft})

		// 4. ticking "เลือกทั้งหมด" again (empty list) opens the account to every active code
		f.updateAccount("5000", func(a *Account) { setList(a, nil) })
		again := dimensionJournal(kind, doc("E"), code("02"))
		f.run(Command{Resource: "journals", Action: "create", Journal: &again})
	}

	// 5. the flag rule still comes first: a registered code on an account that does not separate is refused as before
	f.updateAccount("5000", func(a *Account) {
		a.DepartSeparate, a.ProjectSeparate, a.DepartAllCodes, a.ProjectAllCodes = false, false, false, false
	})
	flagOff := dimensionJournal(DepartmentKind, "JV-OFF", "D01")
	if _, err := f.execute(f.scope, Command{Resource: "journals", Action: "create", Journal: &flagOff}); err == nil || err.Error() != "บรรทัด 1: ผังบัญชี 5000 ไม่ได้กำหนดให้แยกแผนก" {
		t.Fatalf("flag off with a registered department: %v", err)
	}
}

// A saved journal pattern is held to the same register rules as a voucher line.
func TestJournalTemplateDimensionCodesIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.createDimension(DepartmentKind, "D01", "ฝ่ายบัญชีและการเงิน")
	f.createDimension(ProjectKind, "P01", "โครงการก่อสร้างโรงงานแห่งที่สอง")
	f.run(Command{Resource: JournalTemplateKind, Action: "create", Master: rentTemplate()})
	unknown := rentTemplate()
	unknown.Template.Lines[0].DepartmentCode = "D99"
	f.failCode(Command{Resource: JournalTemplateKind, Action: "create", Master: unknown}, "journal_line_depart_not_found", "lines[0].departmentcode")
	f.updateAccount("5000", func(a *Account) { a.ProjectCodes, a.ProjectAllCodes = []string{"P01"}, false })
	f.createDimension(ProjectKind, "P02", "โครงการปรับปรุงคลังสินค้า")
	outside := rentTemplate()
	outside.Template.Lines[0].ProjectCode = "P02"
	f.failCode(Command{Resource: JournalTemplateKind, Action: "create", Master: outside}, "journal_line_project_not_allowed", "lines[0].projectcode")
}

// The allowed-code lists of an account (Champ BCAccSeparate): only with the matching flag, only register codes,
// stored trimmed / unique / sorted, and a code a list still names can be neither deleted nor recoded.
func TestAccountDimensionCodeListsIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	for _, kind := range []string{DepartmentKind, ProjectKind} {
		prefix, listField, wantNeedFlag, wantNotFound := "D", "departcodes", "account_depart_codes_need_flag", "account_depart_code_not_found"
		setList := func(a *Account, codes []string) { a.DepartCodes, a.DepartAllCodes = codes, len(codes) == 0 }
		getList := func(a Account) []string { return a.DepartCodes }
		setFlag := func(a *Account, on bool) { a.DepartSeparate = on }
		if kind == ProjectKind {
			prefix, listField, wantNeedFlag, wantNotFound = "P", "projectcodes", "account_project_codes_need_flag", "account_project_code_not_found"
			setList = func(a *Account, codes []string) { a.ProjectCodes, a.ProjectAllCodes = codes, len(codes) == 0 }
			getList = func(a Account) []string { return a.ProjectCodes }
			setFlag = func(a *Account, on bool) { a.ProjectSeparate = on }
		}
		code := func(n string) string { return prefix + n }
		f.createDimension(kind, code("01"), "ฝ่ายบัญชีและการเงิน")
		f.createDimension(kind, code("02"), "ฝ่ายขาย")
		f.createDimension(kind, code("03"), "ฝ่ายผลิต")
		f.deactivateDimension(kind, code("03"))

		// 1. a list needs the matching flag (reject, not silently dropped)
		f.failAccountUpdate("1000", func(a *Account) { setFlag(a, false); setList(a, []string{code("01")}) }, wantNeedFlag, listField)

		// 2. every code must be in the register and active
		user := f.failAccountUpdate("1000", func(a *Account) { setList(a, []string{code("01"), code("99")}) }, wantNotFound, listField)
		if len(user.Args) != 1 || user.Args[0] != code("99") {
			t.Fatalf("%s unknown list code args = %+v", kind, user.Args)
		}
		f.failAccountUpdate("1000", func(a *Account) { setList(a, []string{code("03")}) }, wantNotFound, listField)

		// 3. stored trimmed, unique and sorted
		f.updateAccount("1000", func(a *Account) { setList(a, []string{" " + code("02"), code("01"), code("02"), ""}) })
		var stored Account
		raw, err := f.store.Get(f.ctx, f.scope, "accounts", f.accounts["1000"].ID)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(raw, &stored); err != nil {
			t.Fatal(err)
		}
		if got := strings.Join(getList(stored), ","); got != code("01")+","+code("02") {
			t.Fatalf("%s stored list = %q", kind, got)
		}

		// 4. a code deactivated after it was listed stays valid for the account that already lists it
		f.deactivateDimension(kind, code("02"))
		f.updateAccount("1000", func(a *Account) {
			a.Names = []Name{{Code: "th", Name: "เงินสดและเงินฝากธนาคาร"}}
		})

		// 5. a listed code cannot be deleted or recoded; once the list drops it, it can
		listed := f.dimension(kind, code("01"))
		f.failCode(Command{Resource: kind, Action: "delete", ID: listed.ID, Version: listed.Version}, "dimension_in_use_delete", "code")
		recoded := listed
		recoded.Code = code("11")
		f.failCode(Command{Resource: kind, Action: "update", ID: listed.ID, Version: listed.Version, Master: &recoded}, "dimension_in_use_code", "code")
		f.updateAccount("1000", func(a *Account) { setList(a, nil) })
		listed = f.dimension(kind, code("01"))
		f.run(Command{Resource: kind, Action: "delete", ID: listed.ID, Version: listed.Version})
		if got, _ := f.dimensions(kind, code("01"), ""); len(got) != 0 {
			t.Fatalf("%s code unused by any list not deleted: %v", kind, got)
		}
	}
}
