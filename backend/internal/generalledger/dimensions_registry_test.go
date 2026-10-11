package generalledger

import "testing"

func allSwitchesOn() DimensionSettings {
	return DimensionSettings{DimensionSettingsInput: DimensionSettingsInput{
		DepartSeparate: true, ProjectSeparate: true, AllocateSeparate: true, JobSeparate: true, PartSeparate: true, SideSeparate: true, BranchSeparate: true,
	}}
}

func allRegisters() dimensionRegisters {
	return dimensionRegisters{
		DepartmentKind: {"ACC": true},
		ProjectKind:    {"P01": true},
		AllocateKind:   {"AL1": true, "AL2": true},
		JobKind:        {"J1": true},
		PartKind:       {"PT1": true},
		SideKind:       {"SD1": true, "SD2": true},
	}
}

func allSeparateAccount(code string) Account {
	return Account{AccountCode: code,
		DepartSeparate: true, ProjectSeparate: true, AllocateSeparate: true, JobSeparate: true, PartSeparate: true, SideSeparate: true,
		DepartAllCodes: true, ProjectAllCodes: true, AllocateAllCodes: true, JobAllCodes: true, PartAllCodes: true, SideAllCodes: true,
	}
}

// The registry carries seven Champ line columns: department, project, allocate, job, part (หน่วยงาน), side (ฝ่าย) — the first six here,
// branch stays on the journal header (design OD-2).
func TestDimensionRegistryHasSixKindsInChampOrder(t *testing.T) {
	want := []string{DepartmentKind, ProjectKind, AllocateKind, JobKind, PartKind, SideKind}
	if len(dimensionRegistry) != len(want) {
		t.Fatalf("registry has %d kinds, want %d", len(dimensionRegistry), len(want))
	}
	for i, kind := range want {
		if dimensionRegistry[i].kind != kind {
			t.Fatalf("registry[%d]=%s, want %s", i, dimensionRegistry[i].kind, kind)
		}
		if !IsDimensionKind(kind) {
			t.Fatalf("%s is not recognised as a dimension kind", kind)
		}
	}
}

func TestEveryDimensionLineCodeIsCheckedLikeChamp(t *testing.T) {
	accounts := map[string]Account{"5100": allSeparateAccount("5100"), "1100": {AccountCode: "1100"}}
	ok := []Line{{AccountCode: "5100", DepartmentCode: "ACC", ProjectCode: "P01", AllocateCode: "AL1", JobCode: "J1", PartCode: "PT1", SideCode: "SD1"}}
	if err := checkLineSeparation(ok, accounts, allRegisters(), allSwitchesOn(), nil); err != nil {
		t.Fatalf("a line with every dimension rejected: %v", err)
	}
	cases := []struct {
		name string
		line Line
		code string
	}{
		{"allocate", Line{AccountCode: "5100", AllocateCode: "NOPE"}, "journal_line_allocate_not_found"},
		{"job", Line{AccountCode: "5100", JobCode: "NOPE"}, "journal_line_job_not_found"},
		{"part", Line{AccountCode: "5100", PartCode: "NOPE"}, "journal_line_part_not_found"},
		{"side", Line{AccountCode: "5100", SideCode: "NOPE"}, "journal_line_side_not_found"},
		{"allocate-not-separate", Line{AccountCode: "1100", AllocateCode: "AL1"}, "journal_line_allocate_not_separate"},
		{"job-not-separate", Line{AccountCode: "1100", JobCode: "J1"}, "journal_line_job_not_separate"},
		{"part-not-separate", Line{AccountCode: "1100", PartCode: "PT1"}, "journal_line_part_not_separate"},
		{"side-not-separate", Line{AccountCode: "1100", SideCode: "SD1"}, "journal_line_side_not_separate"},
	}
	for _, c := range cases {
		err := checkLineSeparation([]Line{c.line}, accounts, allRegisters(), allSwitchesOn(), nil)
		user, isUser := AsUserError(err)
		if !isUser || user.Code != c.code {
			t.Fatalf("%s: got %v, want %s", c.name, err, c.code)
		}
	}
}

// Champ OnCheckComboData (glfrmjournal.cpp:2815-2923): flag on and no "_ALL" row and no code rows = nothing may be chosen.
func TestFlagWithoutAllCodesAndWithoutListAllowsNothing(t *testing.T) {
	accounts := map[string]Account{
		"5100": {AccountCode: "5100", SideSeparate: true},
		"5200": {AccountCode: "5200", SideSeparate: true, SideCodes: []string{"SD1"}},
		"5300": {AccountCode: "5300", SideSeparate: true, SideAllCodes: true},
	}
	err := checkLineSeparation([]Line{{AccountCode: "5100", SideCode: "SD1"}}, accounts, allRegisters(), allSwitchesOn(), nil)
	if user, ok := AsUserError(err); !ok || user.Code != "journal_line_side_not_allowed" {
		t.Fatalf("empty list must allow nothing, got %v", err)
	}
	if err = checkLineSeparation([]Line{{AccountCode: "5200", SideCode: "SD1"}, {AccountCode: "5300", SideCode: "SD2"}, {AccountCode: "5100"}}, accounts, allRegisters(), allSwitchesOn(), nil); err != nil {
		t.Fatalf("listed code / all codes / empty dimension rejected: %v", err)
	}
	err = checkLineSeparation([]Line{{AccountCode: "5200", SideCode: "SD2"}}, accounts, allRegisters(), allSwitchesOn(), nil)
	if user, ok := AsUserError(err); !ok || user.Code != "journal_line_side_not_allowed" {
		t.Fatalf("a code outside the list must be refused, got %v", err)
	}
}

// Champ hides the column when the company switch is off, so a new value cannot be keyed; a value the voucher already holds stays.
func TestSwitchOffRefusesNewCodeButKeepsOldValue(t *testing.T) {
	accounts := map[string]Account{"5100": allSeparateAccount("5100")}
	off := DimensionSettings{} // every switch off, like a company that never saved the settings
	err := checkLineSeparation([]Line{{AccountCode: "5100", JobCode: "J1"}}, accounts, allRegisters(), off, nil)
	if user, ok := AsUserError(err); !ok || user.Code != "journal_line_job_disabled" {
		t.Fatalf("new code while switch off must be refused, got %v", err)
	}
	old := []Line{{AccountCode: "5100", JobCode: "J1"}}
	if err = checkLineSeparation([]Line{{AccountCode: "5100", JobCode: "J1"}}, accounts, allRegisters(), off, old); err != nil {
		t.Fatalf("unchanged old value must stay editable while switch off: %v", err)
	}
	if err = checkLineSeparation([]Line{{AccountCode: "5100"}}, accounts, allRegisters(), off, nil); err != nil {
		t.Fatalf("a line without dimensions never depends on the switches: %v", err)
	}
	only := DimensionSettings{DimensionSettingsInput: DimensionSettingsInput{JobSeparate: true}}
	err = checkLineSeparation([]Line{{AccountCode: "5100", DepartmentCode: "ACC"}}, accounts, allRegisters(), only, nil)
	if user, ok := AsUserError(err); !ok || user.Code != "journal_line_depart_disabled" {
		t.Fatalf("each register has its own switch, got %v", err)
	}
}

func TestDimensionSettingsEnabledPerKind(t *testing.T) {
	s := DimensionSettings{DimensionSettingsInput: DimensionSettingsInput{PartSeparate: true}}
	for _, d := range dimensionRegistry {
		if got, want := s.enabled(d.kind), d.kind == PartKind; got != want {
			t.Fatalf("enabled(%s)=%v, want %v", d.kind, got, want)
		}
	}
	if (DimensionSettings{}).enabled(JobKind) {
		t.Fatal("a company that never saved the switches has every one off")
	}
}

func TestAccountAllCodesNeedsFlagAndDropsList(t *testing.T) {
	base := Account{AccountCode: "5100", Names: []Name{{Code: "th", Name: "ค่าใช้จ่าย"}}, AccountType: "expense", NormalBalance: "debit", Level: 1}
	noFlag := base
	noFlag.JobAllCodes = true
	if user, ok := AsUserError(noFlag.Validate()); !ok || user.Code != "account_job_codes_need_flag" {
		t.Fatalf("all-codes without the flag: %v", noFlag.Validate())
	}
	noFlag.JobSeparate = true
	if err := noFlag.Validate(); err != nil {
		t.Fatalf("all-codes with the flag: %v", err)
	}
	cmd := Command{Account: &Account{AccountCode: "5100", PartSeparate: true, PartAllCodes: true, PartCodes: []string{"PT1", "PT2"}}}
	normalizeCommandCodes(&cmd)
	if cmd.Account.PartCodes != nil {
		t.Fatalf("all-codes already covers every code, list must be dropped: %v", cmd.Account.PartCodes)
	}
}

func TestValidateDimensionMasterLengthsPerKind(t *testing.T) {
	long := func(n int) string {
		out := make([]rune, n)
		for i := range out {
			out[i] = 'ก'
		}
		return string(out)
	}
	cases := []struct {
		kind string
		m    Master
		ok   bool
	}{
		{AllocateKind, Master{Kind: AllocateKind, Code: "AL1", Name: long(50)}, true},
		{AllocateKind, Master{Kind: AllocateKind, Code: "AL1", Name: long(51)}, false},
		{JobKind, Master{Kind: JobKind, Code: "J1", Name: long(255)}, true},
		{JobKind, Master{Kind: JobKind, Code: "J1", Name: long(256)}, false},
		{PartKind, Master{Kind: PartKind, Code: "PT1", Name: "ฝ่ายผลิต"}, true},
		{SideKind, Master{Kind: SideKind, Code: "SD1", Name: ""}, false},
		{SideKind, Master{Kind: SideKind, Code: "SD1", Name: "ฝ่ายขาย", UserGroup: long(256)}, false},
	}
	for i, c := range cases {
		m := c.m
		err := validateDimensionMaster(&m)
		if (err == nil) != c.ok {
			t.Fatalf("case %d (%s): err=%v, want ok=%v", i, c.kind, err, c.ok)
		}
	}
}
