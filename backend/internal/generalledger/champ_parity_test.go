package generalledger

import "testing"

func TestChampParityMasterCollections(t *testing.T) {
	if MasterCollections["journal-books"] != "gl_journal_books" {
		t.Fatalf("expected MasterCollections[journal-books] to be gl_journal_books, got %s", MasterCollections["journal-books"])
	}
}

// Champ CGLFrmJournal::OnCheckSeparate: a department/project on a line is allowed only when the account has
// "แยกแผนก"/"แยกโครงการ" ticked; an empty dimension is always fine (the flag does not make it mandatory).
func TestCheckLineSeparationLikeChamp(t *testing.T) {
	accounts := map[string]Account{
		"5100": {AccountCode: "5100", DepartSeparate: true, DepartAllCodes: true},
		"5200": {AccountCode: "5200", ProjectSeparate: true, ProjectAllCodes: true},
		"1100": {AccountCode: "1100"},
	}
	ok := []Line{
		{AccountCode: "5100", DepartmentCode: "ACC"},
		{AccountCode: "5200", ProjectCode: "P01"},
		{AccountCode: "1100"},
		{AccountCode: "5100"},
	}
	if err := checkLineSeparation(ok, accounts, testDimensions(), allSwitchesOn(), nil); err != nil {
		t.Fatalf("allowed lines rejected: %v", err)
	}
	cases := []struct {
		line Line
		want string
	}{
		{Line{AccountCode: "1100", DepartmentCode: "ACC"}, "บรรทัด 2: ผังบัญชี 1100 ไม่ได้กำหนดให้แยกแผนก"},
		{Line{AccountCode: "5100", ProjectCode: "P01"}, "บรรทัด 2: ผังบัญชี 5100 ไม่ได้กำหนดให้แยกโครงการ"},
	}
	for _, c := range cases {
		err := checkLineSeparation([]Line{{AccountCode: "1100"}, c.line}, accounts, testDimensions(), allSwitchesOn(), nil)
		if err == nil || err.Error() != c.want {
			t.Fatalf("got %v, want %q", err, c.want)
		}
	}
}

func testDimensions() dimensionRegisters {
	return dimensionRegisters{
		DepartmentKind: {"ACC": true, "HR": true, "OLD": false},
		ProjectKind:    {"P01": true, "P02": true},
	}
}

// Champ OnCheckComboData (glfrmjournal.cpp:2913-2917): a code must be in the register and, when the account lists
// codes, in that list; BC keeps an empty list as "every code" (Champ treats flag-on-without-row as nothing allowed).
func TestCheckLineSeparationRegisterAndAllowedList(t *testing.T) {
	accounts := map[string]Account{
		"5100": {AccountCode: "5100", DepartSeparate: true, ProjectSeparate: true, DepartAllCodes: true, ProjectAllCodes: true},
		"5300": {AccountCode: "5300", DepartSeparate: true, ProjectSeparate: true, DepartCodes: []string{"ACC"}, ProjectCodes: []string{"P01"}},
	}
	ok := []Line{
		{AccountCode: "5100", DepartmentCode: "HR", ProjectCode: "P02"},
		{AccountCode: "5300", DepartmentCode: "ACC", ProjectCode: "P01"},
		{AccountCode: "5300"},
	}
	if err := checkLineSeparation(ok, accounts, testDimensions(), allSwitchesOn(), nil); err != nil {
		t.Fatalf("allowed lines rejected: %v", err)
	}
	cases := []struct {
		line Line
		code string
		want string
	}{
		{Line{AccountCode: "5100", DepartmentCode: "XX"}, "journal_line_depart_not_found", "บรรทัด 1: ไม่พบแผนก “XX” ที่เปิดใช้งานในทะเบียนแผนก — เลือกรหัสจากรายการ"},
		{Line{AccountCode: "5100", DepartmentCode: "OLD"}, "journal_line_depart_not_found", "บรรทัด 1: ไม่พบแผนก “OLD” ที่เปิดใช้งานในทะเบียนแผนก — เลือกรหัสจากรายการ"},
		{Line{AccountCode: "5100", ProjectCode: "P99"}, "journal_line_project_not_found", "บรรทัด 1: ไม่พบโครงการ “P99” ที่เปิดใช้งานในทะเบียนโครงการ — เลือกรหัสจากรายการ"},
		{Line{AccountCode: "5300", DepartmentCode: "HR"}, "journal_line_depart_not_allowed", "บรรทัด 1: บัญชี 5300 กำหนดแผนกที่ใช้ได้ไว้ ไม่รวมแผนก “HR” — เลือกแผนกที่บัญชีนี้อนุญาต หรือแก้รายการที่ผังบัญชี"},
		{Line{AccountCode: "5300", ProjectCode: "P02"}, "journal_line_project_not_allowed", "บรรทัด 1: บัญชี 5300 กำหนดโครงการที่ใช้ได้ไว้ ไม่รวมโครงการ “P02” — เลือกโครงการที่บัญชีนี้อนุญาต หรือแก้รายการที่ผังบัญชี"},
	}
	for _, c := range cases {
		err := checkLineSeparation([]Line{c.line}, accounts, testDimensions(), allSwitchesOn(), nil)
		user, isUser := AsUserError(err)
		if !isUser || user.Code != c.code || user.Message != c.want {
			t.Fatalf("got %v, want %s %q", err, c.code, c.want)
		}
	}
}

func TestNormalizeCodeListSortsDedupesAndDropsBlanks(t *testing.T) {
	got := normalizeCodeList([]string{" HR ", "ACC", "HR", "", "  "})
	if len(got) != 2 || got[0] != "ACC" || got[1] != "HR" {
		t.Fatalf("got %v", got)
	}
	if normalizeCodeList([]string{" ", ""}) != nil || normalizeCodeList(nil) != nil {
		t.Fatal("blank list must normalise to nil so omitempty drops it")
	}
}

func TestAccountDimensionCodeListsNeedMatchingFlag(t *testing.T) {
	base := Account{AccountCode: "5100", Names: []Name{{Code: "th", Name: "ค่าใช้จ่าย"}}, AccountType: "expense", NormalBalance: "debit", Level: 1}
	withDepart := base
	withDepart.DepartCodes = []string{"ACC"}
	if user, ok := AsUserError(withDepart.Validate()); !ok || user.Code != "account_depart_codes_need_flag" || user.Field != "departcodes" {
		t.Fatalf("depart list without flag: %v", withDepart.Validate())
	}
	withProject := base
	withProject.ProjectCodes = []string{"P01"}
	if user, ok := AsUserError(withProject.Validate()); !ok || user.Code != "account_project_codes_need_flag" {
		t.Fatalf("project list without flag: %v", withProject.Validate())
	}
	withDepart.DepartSeparate = true
	withProject.ProjectSeparate = true
	if err := withDepart.Validate(); err != nil {
		t.Fatalf("depart list with flag: %v", err)
	}
	if err := withProject.Validate(); err != nil {
		t.Fatalf("project list with flag: %v", err)
	}
}
