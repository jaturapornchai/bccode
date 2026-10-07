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
		"5100": {AccountCode: "5100", DepartSeparate: true},
		"5200": {AccountCode: "5200", ProjectSeparate: true},
		"1100": {AccountCode: "1100"},
	}
	ok := []Line{
		{AccountCode: "5100", DepartmentCode: "ACC"},
		{AccountCode: "5200", ProjectCode: "P01"},
		{AccountCode: "1100"},
		{AccountCode: "5100"},
	}
	if err := checkLineSeparation(ok, accounts); err != nil {
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
		err := checkLineSeparation([]Line{{AccountCode: "1100"}, c.line}, accounts)
		if err == nil || err.Error() != c.want {
			t.Fatalf("got %v, want %q", err, c.want)
		}
	}
}
