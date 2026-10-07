package generalledger

import (
	"strings"
	"testing"
)

func templateLines(codes ...string) []JournalTemplateLine {
	lines := make([]JournalTemplateLine, len(codes))
	for i, code := range codes {
		lines[i] = JournalTemplateLine{AccountCode: code}
	}
	return lines
}

func TestCleanJournalTemplateTrimsAndKeepsPattern(t *testing.T) {
	got, err := cleanJournalTemplate(JournalTemplate{
		BookCode: " JV ", Reference: " สัญญาเช่า-01 ", RefDate: "2026-01-01", Description: " ค่าเช่าสำนักงาน ", TaxType: 1,
		Lines: []JournalTemplateLine{{AccountCode: " 5000 ", Description: " ค่าเช่า ", DepartmentCode: " D01 ", ProjectCode: "P01", CashFlow: "operating"}, {AccountCode: "1000"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.BookCode != "JV" || got.Reference != "สัญญาเช่า-01" || got.Description != "ค่าเช่าสำนักงาน" || got.TaxType != 1 || got.RefDate != "2026-01-01" {
		t.Fatalf("header not trimmed: %+v", got)
	}
	if first := got.Lines[0]; first.AccountCode != "5000" || first.Description != "ค่าเช่า" || first.DepartmentCode != "D01" || first.ProjectCode != "P01" || first.CashFlow != "operating" {
		t.Fatalf("line not trimmed: %+v", first)
	}
}

func TestCleanJournalTemplateRejectsWhatAVoucherWouldReject(t *testing.T) {
	long := strings.Repeat("ก", JournalLineDescriptionMaxRunes+1)
	cases := []struct {
		name string
		in   JournalTemplate
		code string // expected UserError code, "" = plain error
	}{
		{"no lines", JournalTemplate{}, ""},
		{"too many lines", JournalTemplate{Lines: templateLines(make([]string, JournalTemplateMaxLines+1)...)}, ""},
		{"blank account", JournalTemplate{Lines: templateLines("5000", "  ")}, "code_required"},
		{"account with space", JournalTemplate{Lines: templateLines("50 00")}, "code_has_space"},
		{"bad ref date", JournalTemplate{RefDate: "2026-02-30", Lines: templateLines("5000")}, "journal_refdate_invalid"},
		{"reference too long", JournalTemplate{Reference: strings.Repeat("A", JournalReferenceMaxRunes+1), Lines: templateLines("5000")}, "journal_reference_too_long"},
		{"description too long", JournalTemplate{Description: strings.Repeat("ก", JournalDescriptionMaxRunes+1), Lines: templateLines("5000")}, "journal_description_too_long"},
		{"line description too long", JournalTemplate{Lines: []JournalTemplateLine{{AccountCode: "5000", Description: long}}}, "journal_line_description_too_long"},
		{"tax type 3", JournalTemplate{TaxType: 3, Lines: templateLines("5000")}, ""},
		{"cash flow", JournalTemplate{Lines: []JournalTemplateLine{{AccountCode: "5000", CashFlow: "other"}}}, ""},
	}
	for _, tc := range cases {
		_, err := cleanJournalTemplate(tc.in)
		if err == nil {
			t.Fatalf("%s: accepted", tc.name)
		}
		user, isUser := AsUserError(err)
		if tc.code != "" && (!isUser || user.Code != tc.code) {
			t.Fatalf("%s: got %v, want code %s", tc.name, err, tc.code)
		}
		if tc.code == "" && isUser {
			t.Fatalf("%s: got user error %s, want a plain validation error", tc.name, user.Code)
		}
	}
}
