//go:build integration

package generalledger

import (
	"strings"
	"testing"
)

// The four registers added with the company switches (Champ BCAllocate / BCJob / BCPart / BCSide).
var newDimensionKinds = []string{AllocateKind, JobKind, PartKind, SideKind}

// testPrefix gives every register its own code prefix (DE, PR, AL, JO, PA, SI) so one code never reads like another register's.
func testPrefix(d dimensionKind) string { return strings.ToUpper(d.errKey[:2]) }

func switchOf(s *DimensionSettingsInput, kind string) *bool {
	switch kind {
	case DepartmentKind:
		return &s.DepartSeparate
	case ProjectKind:
		return &s.ProjectSeparate
	case AllocateKind:
		return &s.AllocateSeparate
	case JobKind:
		return &s.JobSeparate
	case PartKind:
		return &s.PartSeparate
	case SideKind:
		return &s.SideSeparate
	}
	panic("no switch for " + kind)
}

func setBudgetDimension(b *Budget, kind, code string) {
	switch kind {
	case DepartmentKind:
		b.DepartmentCode = code
	case ProjectKind:
		b.ProjectCode = code
	case AllocateKind:
		b.AllocateCode = code
	case JobKind:
		b.JobCode = code
	case PartKind:
		b.PartCode = code
	case SideKind:
		b.SideCode = code
	}
}

func (f *pgIntegrityFixture) dimensionSettings() DimensionSettingsView {
	f.t.Helper()
	view, err := f.store.Projection().DimensionSettings(f.ctx, f.scope)
	if err != nil {
		f.t.Fatal(err)
	}
	return view
}

// setDimensionSwitch turns one company switch on or off and keeps the others.
func (f *pgIntegrityFixture) setDimensionSwitch(kind string, on bool) {
	f.t.Helper()
	view := f.dimensionSettings()
	input := view.DimensionSettingsInput
	*switchOf(&input, kind) = on
	f.run(Command{Resource: DimensionSettingsKind, Action: "update", ID: dimensionSettingsID, Version: view.Version, DimensionSettings: &input})
}

// setAccountDimension sets the flag, "เลือกทั้งหมด" and the allowed codes of one register on a fixture account.
func (f *pgIntegrityFixture) setAccountDimension(account, kind string, separate, all bool, codes ...string) {
	f.t.Helper()
	f.updateAccount(account, func(a *Account) {
		dim := a.accountDimension(kind)
		*dim.separate, *dim.all, *dim.codes = separate, all, codes
	})
}

// The company switches (Champ BCConfigurations XSeparate, myfrmsysconfig.cpp:927-946): all off until the owner turns one on;
// saved with the version the screen read; switching one off is always allowed and never touches stored data.
func TestDimensionSettingsLifecycleIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)

	// 1. a company that never saved the switches has every one off and version 0 (Champ DDL DEFAULT 0)
	fresh, err := f.store.Projection().DimensionSettings(f.ctx, Scope{Holding: "H", Company: "NEVER", Actor: "tester"})
	if err != nil || fresh.Version != 0 || fresh.DimensionSettingsInput != (DimensionSettingsInput{}) {
		t.Fatalf("default switches = %+v %v", fresh, err)
	}

	// 2. the fixture turned department and project on; the other four stay off
	view := f.dimensionSettings()
	if view.DimensionSettingsInput != (DimensionSettingsInput{DepartSeparate: true, ProjectSeparate: true}) || view.Version == 0 {
		t.Fatalf("fixture switches = %+v", view)
	}

	// 3. a stale version is refused, the current one saves and moves the version; branch is stored but has no effect (OD-2)
	input := view.DimensionSettingsInput
	input.JobSeparate, input.BranchSeparate = true, true
	f.failCode(Command{Resource: DimensionSettingsKind, Action: "update", ID: dimensionSettingsID, Version: view.Version + 7, DimensionSettings: &input}, CodeStaleVersion, "")
	f.run(Command{Resource: DimensionSettingsKind, Action: "update", ID: dimensionSettingsID, Version: view.Version, DimensionSettings: &input})
	saved := f.dimensionSettings()
	if !saved.JobSeparate || !saved.BranchSeparate || !saved.DepartSeparate || saved.AllocateSeparate || saved.Version <= view.Version {
		t.Fatalf("saved switches = %+v (was version %d)", saved, view.Version)
	}

	// 4. no payload, another id or another action is refused
	for name, cmd := range map[string]Command{
		"no payload": {Resource: DimensionSettingsKind, Action: "update", ID: dimensionSettingsID, Version: saved.Version},
		"other id":   {Resource: DimensionSettingsKind, Action: "update", ID: "other", Version: saved.Version, DimensionSettings: &input},
		"delete":     {Resource: DimensionSettingsKind, Action: "delete", ID: dimensionSettingsID, Version: saved.Version},
	} {
		if _, err = f.execute(f.scope, cmd); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}

	// 5. a register can be filled while its switch is off (the owner prepares the codes first)
	f.createDimension(AllocateKind, "AL01", "ศูนย์ต้นทุนส่วนกลาง")

	// 6. switching a register off while a voucher uses it is allowed: the stored code stays and the draft stays editable,
	//    a NEW code is refused, and posting never re-checks (Champ posting does not look at dimensions, GLFrmPostTrans.cpp:722-800)
	f.createDimension(DepartmentKind, "D01", "ฝ่ายบัญชีและการเงิน")
	f.createDimension(DepartmentKind, "D02", "ฝ่ายขาย")
	draft := dimensionJournal(DepartmentKind, "JV-SW1", "D01")
	made := f.run(Command{Resource: "journals", Action: "create", Journal: &draft})
	f.setDimensionSwitch(DepartmentKind, false)
	if off := f.dimensionSettings(); off.DepartSeparate || !off.ProjectSeparate || !off.JobSeparate {
		t.Fatalf("only the department switch should move: %+v", off)
	}
	edit := f.journal(made.ID)
	edit.Description = "แก้คำอธิบายหลังปิดสวิตช์แผนก"
	f.run(Command{Resource: "journals", Action: "update", ID: edit.ID, Version: edit.Version, Journal: &edit})
	if kept := f.journal(made.ID); kept.Lines[0].DepartmentCode != "D01" {
		t.Fatalf("stored department lost after the switch went off: %+v", kept.Lines[0])
	}
	moved := f.journal(made.ID)
	moved.Lines[0].DepartmentCode = "D02"
	user := f.failCode(Command{Resource: "journals", Action: "update", ID: moved.ID, Version: moved.Version, Journal: &moved}, "journal_line_depart_disabled", "lines[0].departmentcode")
	if len(user.Args) != 2 || user.Args[0] != "1" || user.Args[1] != "D02" {
		t.Fatalf("disabled error args = %v", user.Args)
	}
	fresher := dimensionJournal(DepartmentKind, "JV-SW2", "D01")
	f.failCode(Command{Resource: "journals", Action: "create", Journal: &fresher}, "journal_line_depart_disabled", "lines[0].departmentcode")
	f.post(f.journal(made.ID))
}

// The four new registers behave like department and project: trimmed code and name, Champ widths (job code 15, the others 10;
// names 50 / 255), a user group, from–to list, and no delete or recode while a voucher, an account list or a budget names the code.
func TestNewDimensionRegistersLifecycleIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.enableDimensions(newDimensionKinds...)
	for _, kind := range newDimensionKinds {
		d := dimensionKinds[kind]
		prefix := testPrefix(d)
		code := func(n string) string { return prefix + n }

		// 1. created with code, name and user group trimmed
		created := f.run(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: "  " + code("01") + " ", Name: "  ทีม" + d.label + "หลัก ", UserGroup: " ทีมบัญชี ", IsActive: true}})
		first := f.dimension(kind, code("01"))
		if first.ID != created.ID || first.Code != code("01") || first.Name != "ทีม"+d.label+"หลัก" || first.UserGroup != "ทีมบัญชี" || !first.IsActive {
			t.Fatalf("%s stored = %+v", kind, first)
		}

		// 2. duplicate code; the same code is free in another register
		f.failCode(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: code("01"), Name: "ซ้ำ", IsActive: true}}, CodeDuplicateCode, "")

		// 3. widths of Champ's columns, counted in runes
		f.failCode(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: strings.Repeat("X", d.codeMax+1), Name: "รหัสยาวเกิน", IsActive: true}}, "code_too_long", "code")
		f.failCode(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: code("98"), Name: "   ", IsActive: true}}, "dimension_name_required", "name")
		f.failCode(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: code("98"), Name: strings.Repeat("ก", d.nameMax+1), IsActive: true}}, "dimension_name_too_long", "name")
		f.failCode(Command{Resource: kind, Action: "create", Master: &Master{Kind: kind, Code: code("98"), Name: "ชื่อปกติ", UserGroup: strings.Repeat("ก", dimensionUserGroupMaxRunes+1), IsActive: true}}, "dimension_usergroup_too_long", "usergroup")
		widest := strings.Repeat("X", d.codeMax)
		f.createDimension(kind, widest, strings.Repeat("ก", d.nameMax)) // exactly the Champ widths

		// 4. from–to list
		f.createDimension(kind, code("02"), "ทีมที่สอง")
		f.createDimension(kind, code("03"), "ทีมที่สาม")
		f.createDimension(kind, code("10"), "ทีมที่สิบ")
		if got, n := f.dimensions(kind, code("02"), code("10")); strings.Join(got, ",") != strings.Join([]string{code("02"), code("03"), code("10")}, ",") || n != 3 {
			t.Fatalf("%s range = %v (%d)", kind, got, n)
		}
		if got, n := f.dimensions(kind, code("99"), ""); len(got) != 0 || n != 0 {
			t.Fatalf("%s unknown exact code = %v (%d)", kind, got, n)
		}

		// 5. a draft voucher line blocks delete and recode, deleting the draft frees the code
		f.setAccountDimension("5000", kind, true, true)
		draft := dimensionJournal(kind, "JV-"+prefix+"1", code("01"))
		made := f.run(Command{Resource: "journals", Action: "create", Journal: &draft})
		f.failCode(Command{Resource: kind, Action: "delete", ID: first.ID, Version: first.Version}, "dimension_in_use_delete", "code")
		recoded := first
		recoded.Code = code("11")
		f.failCode(Command{Resource: kind, Action: "update", ID: first.ID, Version: first.Version, Master: &recoded}, "dimension_in_use_code", "code")
		saved := f.journal(made.ID)
		f.run(Command{Resource: "journals", Action: "delete", ID: saved.ID, Version: saved.Version})
		current := f.dimension(kind, code("01"))
		f.run(Command{Resource: kind, Action: "delete", ID: current.ID, Version: current.Version})

		// 6. an account whose allowed list names the code blocks delete until the list drops it
		f.setAccountDimension("1000", kind, true, false, code("02"))
		second := f.dimension(kind, code("02"))
		f.failCode(Command{Resource: kind, Action: "delete", ID: second.ID, Version: second.Version}, "dimension_in_use_delete", "code")
		f.setAccountDimension("1000", kind, false, false)
		f.run(Command{Resource: kind, Action: "delete", ID: second.ID, Version: second.Version})

		// 7. a budget limited to the code blocks delete
		budget := Budget{Code: "BG-" + prefix, Name: "งบประมาณประจำปี", FiscalYear: "2026", BranchCode: "B1", Lines: []BudgetLine{{AccountCode: "5000", Periods: budgetMonths("10")}}}
		setBudgetDimension(&budget, kind, code("03"))
		f.run(Command{Resource: "budgets", Action: "create", Budget: &budget})
		third := f.dimension(kind, code("03"))
		f.failCode(Command{Resource: kind, Action: "delete", ID: third.ID, Version: third.Version}, "dimension_in_use_delete", "code")

		// 8. the registers never see each other's records
		for _, other := range newDimensionKinds {
			if other != kind {
				if got, _ := f.dimensions(other, code("03"), ""); len(got) != 0 {
					t.Fatalf("%s code %s leaked into %s: %v", kind, code("03"), other, got)
				}
			}
		}
	}
}

// The five line rules of the design (§3.5) for the four new registers, in Champ's order: switch, account flag, register,
// allowed list ("เลือกทั้งหมด" or the code in the list; an empty list allows nothing).
func TestJournalLineNewDimensionRulesIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	for _, kind := range newDimensionKinds {
		d := dimensionKinds[kind]
		prefix := testPrefix(d)
		code := func(n string) string { return prefix + n }
		field := "lines[0]." + d.jsonKey
		failure := func(suffix string) string { return dimensionErrorCode("journal_line_", d, suffix) }
		create := func(j Journal) Command { return Command{Resource: "journals", Action: "create", Journal: &j} }
		f.createDimension(kind, code("01"), "ทีม"+d.label+"บัญชี")
		f.createDimension(kind, code("02"), "ทีม"+d.label+"ขาย")
		f.createDimension(kind, code("03"), "ทีม"+d.label+"ผลิต")
		f.deactivateDimension(kind, code("03"))

		// 1. the company switch is off (default): a new code is refused although the register has it
		f.failCode(create(dimensionJournal(kind, "JV-"+prefix+"A", code("01"))), failure("_disabled"), field)

		// 2. switch on, the account has no "แยก…" flag
		f.setDimensionSwitch(kind, true)
		f.failCode(create(dimensionJournal(kind, "JV-"+prefix+"A", code("01"))), failure("_not_separate"), field)

		// 3. flag on, no "เลือกทั้งหมด", no list: Champ allows nothing
		f.setAccountDimension("5000", kind, true, false)
		f.failCode(create(dimensionJournal(kind, "JV-"+prefix+"A", code("01"))), failure("_not_allowed"), field)

		// 4. "เลือกทั้งหมด": every active code; an unknown or inactive one is refused with the line number and the code
		f.setAccountDimension("5000", kind, true, true)
		made := f.run(create(dimensionJournal(kind, "JV-"+prefix+"B", code("01"))))
		user := f.failCode(create(dimensionJournal(kind, "JV-"+prefix+"U", code("99"))), failure("_not_found"), field)
		if len(user.Args) != 2 || user.Args[0] != "1" || user.Args[1] != code("99") || !strings.Contains(user.Message, "บรรทัด 1") {
			t.Fatalf("%s unknown code error = %+v", kind, user)
		}
		f.failCode(create(dimensionJournal(kind, "JV-"+prefix+"I", code("03"))), failure("_not_found"), field)

		// 5. a list narrows the codes, on create and on editing a draft
		f.setAccountDimension("5000", kind, true, false, code("01"))
		user = f.failCode(create(dimensionJournal(kind, "JV-"+prefix+"O", code("02"))), failure("_not_allowed"), field)
		if len(user.Args) != 3 || user.Args[1] != "5000" || user.Args[2] != code("02") {
			t.Fatalf("%s not-allowed error args = %v", kind, user.Args)
		}
		f.run(create(dimensionJournal(kind, "JV-"+prefix+"N", code("01"))))
		outside := f.journal(made.ID)
		*outside.Lines[0].lineCodePtr(kind) = code("02")
		f.failCode(Command{Resource: "journals", Action: "update", ID: outside.ID, Version: outside.Version, Journal: &outside}, failure("_not_allowed"), field)

		// 6. switch off again: the stored code stays through an edit, a new code is refused, posting is not re-checked
		f.setDimensionSwitch(kind, false)
		kept := f.journal(made.ID)
		kept.Description = "แก้คำอธิบายหลังปิดสวิตช์"
		f.run(Command{Resource: "journals", Action: "update", ID: kept.ID, Version: kept.Version, Journal: &kept})
		moved := f.journal(made.ID)
		*moved.Lines[0].lineCodePtr(kind) = code("02")
		f.failCode(Command{Resource: "journals", Action: "update", ID: moved.ID, Version: moved.Version, Journal: &moved}, failure("_disabled"), field)
		posted := f.post(f.journal(made.ID))
		if got := posted.Lines[0].dimensionCode(kind); got != code("01") {
			t.Fatalf("%s posted voucher lost its code: %q", kind, got)
		}
		if n := f.countRows(`SELECT count(*) FROM gl_lines WHERE company='C' AND `+d.column+`=$1`, code("01")); n != 1 {
			t.Fatalf("%s ledger lines carrying %s = %d, want 1", kind, code("01"), n)
		}
	}
}

// A budget limited to a register code (Champ GLFrmBudget.cpp:421-530): the code must be in the register, every account must
// carry the "แยก…" flag and accept the code; a register whose switch is off is not read (the stored code is kept).
func TestBudgetDimensionCodesIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	for _, d := range dimensionRegistry {
		kind := d.kind
		prefix := testPrefix(d)
		code := func(n string) string { return prefix + n }
		field := d.jsonKey
		failure := func(suffix string) string { return dimensionErrorCode("budget_", d, suffix) }
		budget := func(name, account, dimension string) Command {
			b := Budget{Code: "BG-" + prefix + name, Name: "งบประมาณ" + d.label, FiscalYear: "2026", BranchCode: "B1", Lines: []BudgetLine{{AccountCode: account, Periods: budgetMonths("10")}}}
			setBudgetDimension(&b, kind, dimension)
			return Command{Resource: "budgets", Action: "create", Budget: &b}
		}
		f.createDimension(kind, code("01"), "ทีม"+d.label+"บัญชี")
		f.createDimension(kind, code("02"), "ทีม"+d.label+"ขาย")

		// 1. switch off: the register is not read, the budget saves and keeps its code
		f.setDimensionSwitch(kind, false)
		f.run(budget("0", "4000", code("99")))
		if got := f.budget("BG-" + prefix + "0").dimensionCode(kind); got != code("99") {
			t.Fatalf("%s budget with the switch off kept %q", kind, got)
		}

		// 2. switch on: an unknown code is refused, an account without the flag is refused
		f.setDimensionSwitch(kind, true)
		user := f.failCode(budget("1", "4000", code("99")), failure("_not_found"), field)
		if len(user.Args) != 1 || user.Args[0] != code("99") {
			t.Fatalf("%s not-found args = %v", kind, user.Args)
		}
		f.setAccountDimension("4000", kind, false, false)
		user = f.failCode(budget("1", "4000", code("01")), failure("_not_separate"), field)
		if len(user.Args) != 1 || user.Args[0] != "4000" {
			t.Fatalf("%s not-separate args = %v", kind, user.Args)
		}

		// 3. flagged with a list: a code outside the list is refused, one inside is saved
		f.setAccountDimension("4000", kind, true, false, code("01"))
		user = f.failCode(budget("1", "4000", code("02")), failure("_not_allowed"), field)
		if len(user.Args) != 2 || user.Args[0] != "4000" || user.Args[1] != code("02") {
			t.Fatalf("%s not-allowed args = %v", kind, user.Args)
		}
		f.run(budget("2", "4000", code("01")))

		// 4. "เลือกทั้งหมด": any registered code; a budget without the dimension is never checked
		f.setAccountDimension("4000", kind, true, true)
		f.run(budget("3", "4000", code("02")))
		if got := f.budget("BG-" + prefix + "3").dimensionCode(kind); got != code("02") {
			t.Fatalf("%s budget kept %q, want %s", kind, got, code("02"))
		}
		f.run(budget("4", "4000", ""))
	}
}

// Year-end (design §3.9): closing and transfer lines keep all six codes of the balance they move, and the real close writes
// the same codes the preview showed.
func TestClosePreviewAndCloseCarryAllSixDimensionsIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.enableDimensions(newDimensionKinds...)
	codes := map[string]string{}
	for _, d := range dimensionRegistry {
		codes[d.kind] = testPrefix(d) + "01"
		f.ensureDimension(d.kind, codes[d.kind])
		for _, account := range []string{"1000", "4000"} {
			f.setAccountDimension(account, d.kind, true, true)
		}
	}
	sale := Journal{DocNo: "S-ALL6", Date: "2026-03-10", BookCode: "JV", FiscalYear: "2026", Description: "ขายสินค้าเงินสด", Kind: "manual", BranchCode: "B1",
		Lines: []Line{{AccountCode: "1000", Debit: Amount("100"), Credit: Amount("0")}, {AccountCode: "4000", Debit: Amount("0"), Credit: Amount("100")}}}
	for i := range sale.Lines {
		for kind, code := range codes {
			*sale.Lines[i].lineCodePtr(kind) = code
		}
	}
	f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &sale}).ID))

	preview, err := f.previewClose("2026-12-31", "CLOSE-ALL6")
	if err != nil || len(preview.Branches) != 1 {
		t.Fatalf("preview = %+v %v", preview, err)
	}
	for _, line := range preview.Branches[0].Lines {
		got := ClosePreviewLine{DepartmentCode: line.DepartmentCode, ProjectCode: line.ProjectCode, AllocateCode: line.AllocateCode, JobCode: line.JobCode, PartCode: line.PartCode, SideCode: line.SideCode}
		want := ClosePreviewLine{DepartmentCode: codes[DepartmentKind], ProjectCode: codes[ProjectKind], AllocateCode: codes[AllocateKind], JobCode: codes[JobKind], PartCode: codes[PartKind], SideCode: codes[SideKind]}
		if got != want {
			t.Fatalf("preview line %s (transfer=%v) codes = %+v, want %+v", line.AccountCode, line.Transfer, got, want)
		}
	}

	year := f.years["2026"]
	f.run(Command{Resource: "processes", Action: "close", ID: "2026", Version: year.Version, DocNo: "CLOSE-ALL6", Date: "2026-12-31", Reason: "close"})
	drafts := f.closingDrafts()
	if len(drafts) != 1 || len(drafts[0].Lines) != len(preview.Branches[0].Lines) {
		t.Fatalf("closing drafts = %+v", drafts)
	}
	for _, line := range drafts[0].Lines {
		for kind, code := range codes {
			if got := line.dimensionCode(kind); got != code {
				t.Fatalf("closing line %s lost %s: %q, want %q", line.AccountCode, kind, got, code)
			}
		}
	}
}

// R9: a company that kept department / project on its lines before the switches existed is moved once: the switches of the
// registers it uses go on, "เลือกทั้งหมด" is ticked on every flagged account that listed nothing. A dry run writes nothing.
func TestDimensionProvisionMovesAnExistingCompanyIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.createDimension(DepartmentKind, "D01", "ฝ่ายบัญชีและการเงิน")
	// the company as it was: flags on, switches off, no "เลือกทั้งหมด"; account 4000 already lists a department
	f.setDimensionSwitch(DepartmentKind, false)
	f.setDimensionSwitch(ProjectKind, false)
	f.setAccountDimension("1000", DepartmentKind, true, false)
	f.setAccountDimension("5000", ProjectKind, true, false)
	f.setAccountDimension("4000", DepartmentKind, true, false, "D01")

	before := f.rowCounts()
	plan, err := f.store.PlanDimensionProvision(f.ctx, f.scope)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(plan.Switches, ",") != DepartmentKind+","+ProjectKind || strings.Join(plan.Accounts, ",") != "1000,5000" || plan.Empty() {
		t.Fatalf("plan = switches %v accounts %v", plan.Switches, plan.Accounts)
	}
	if after := f.rowCounts(); after != before {
		t.Fatalf("planning wrote rows: before %v after %v", before, after)
	}

	if err = f.store.ApplyDimensionProvision(f.ctx, f.scope, plan); err != nil {
		t.Fatal(err)
	}
	if view := f.dimensionSettings(); !view.DepartSeparate || !view.ProjectSeparate || view.AllocateSeparate || view.JobSeparate || view.PartSeparate || view.SideSeparate {
		t.Fatalf("switches after provisioning = %+v", view)
	}
	account := func(code string) Account {
		t.Helper()
		return f.accountRecord(code)
	}
	if a := account("1000"); !a.DepartAllCodes || len(a.DepartCodes) != 0 {
		t.Fatalf("1000 department after provisioning = all %v codes %v", a.DepartAllCodes, a.DepartCodes)
	}
	if a := account("5000"); !a.ProjectAllCodes {
		t.Fatalf("5000 project after provisioning: all=%v", a.ProjectAllCodes)
	}
	if a := account("4000"); a.DepartAllCodes || strings.Join(a.DepartCodes, ",") != "D01" {
		t.Fatalf("4000 kept its own list: all %v codes %v", a.DepartAllCodes, a.DepartCodes)
	}
	if again, err := f.store.PlanDimensionProvision(f.ctx, f.scope); err != nil || !again.Empty() {
		t.Fatalf("second plan = %+v %v, want nothing to do", again, err)
	}

	// the voucher of the old way of working saves again
	j := dimensionJournal(DepartmentKind, "JV-PROV", "D01")
	f.run(Command{Resource: "journals", Action: "create", Journal: &j})
}
