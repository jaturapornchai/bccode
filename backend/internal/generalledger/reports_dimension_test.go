package generalledger

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestParseReportCodeList(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want []string
	}{
		{"", []string{}},
		{" , ,", []string{}},
		{"D1", []string{"D1"}},
		{" D1 ", []string{"D1"}},
		{"D2,D1", []string{"D1", "D2"}},
		{"D1, D2 ,D1,,D2", []string{"D1", "D2"}},
		{"ฝ่ายขาย,ฝ่ายผลิต", []string{"ฝ่ายขาย", "ฝ่ายผลิต"}},
	} {
		got, err := parseReportCodeList(tc.raw, "departmentcode")
		if err != nil || got == nil || !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parseReportCodeList(%q) = %#v, %v, want %#v", tc.raw, got, err, tc.want)
		}
	}
}

func TestParseReportCodeListCap(t *testing.T) {
	codes := make([]string, maxReportDimensionCodes+1)
	for i := range codes {
		codes[i] = fmt.Sprintf("P%03d", i)
	}
	if got, err := parseReportCodeList(strings.Join(codes[:maxReportDimensionCodes], ","), "projectcode"); err != nil || len(got) != maxReportDimensionCodes {
		t.Fatalf("%d codes = %d, %v", maxReportDimensionCodes, len(got), err)
	}
	// repeated codes count once, so a long raw list of few distinct codes is fine
	if got, err := parseReportCodeList(strings.Repeat("P1,", 500), "projectcode"); err != nil || len(got) != 1 {
		t.Fatalf("repeated code = %v, %v", got, err)
	}
	_, err := parseReportCodeList(strings.Join(codes, ","), "projectcode")
	user, ok := err.(*UserError)
	if !ok || user.Code != "report_dimension_too_many" || user.Field != "projectcode" || user.Status != 400 || !reflect.DeepEqual(user.Args, []string{"200", "201"}) {
		t.Fatalf("201 codes = %#v", err)
	}
}
