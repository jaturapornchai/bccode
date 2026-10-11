package report

import (
	"os"
	"strings"
	"testing"
)

func lookup(t *testing.T, menuID string) (*Definition, *Variant) {
	t.Helper()
	d, v, ok := testRegistry().Lookup(menuID)
	if !ok {
		t.Fatalf("menu %s not registered", menuID)
	}
	return d, v
}

func journalInput(extra map[string]string) map[string]string {
	in := map[string]string{"date_from": "2026-01-01", "date_to": "2026-01-31"}
	for k, v := range extra {
		in[k] = v
	}
	return in
}

func wantFailure(t *testing.T, failure *Failure, code, field string) {
	t.Helper()
	if failure == nil {
		t.Fatalf("want %s, got success", code)
	}
	if failure.Code != code || failure.Field != field {
		t.Fatalf("got %s field %q, want %s field %q", failure.Code, failure.Field, code, field)
	}
}

func TestResolveParamsAppliesDefaultsAndPresets(t *testing.T) {
	d, v := lookup(t, "gl-daily-report")
	values, failure := ResolveParams(d, v, journalInput(map[string]string{"books": " JV , SV ", "account_from": " 1100 "}))
	if failure != nil {
		t.Fatal(failure)
	}
	if values["post_status"] != "all" || values["summary"] != "false" || values["books"] != "JV,SV" || values["account_from"] != "1100" {
		t.Fatalf("values = %v", values)
	}
	d, v = lookup(t, "trial-balance")
	values, failure = ResolveParams(d, v, journalInput(nil))
	if failure != nil {
		t.Fatal(failure)
	}
	if values["post_status"] != "posted" || values["gross"] != "false" {
		t.Fatalf("trial balance preset not applied: %v", values)
	}
}

// a variant's preset is locked: the same value is accepted, any other value is 403 REPORT_VARIANT_FORBIDDEN
func TestResolveParamsLocksVariantPreset(t *testing.T) {
	d, v := lookup(t, "trial-balance-2")
	if _, failure := ResolveParams(d, v, journalInput(map[string]string{"gross": "true"})); failure != nil {
		t.Fatal(failure)
	}
	_, failure := ResolveParams(d, v, journalInput(map[string]string{"gross": "false"}))
	wantFailure(t, failure, "REPORT_VARIANT_FORBIDDEN", "gross")
	if failure.Status != 403 {
		t.Fatalf("status = %d", failure.Status)
	}
	d, v = lookup(t, "gl-daily-summary-report")
	_, failure = ResolveParams(d, v, journalInput(map[string]string{"summary": "false"}))
	wantFailure(t, failure, "REPORT_VARIANT_FORBIDDEN", "summary")
}

func TestResolveParamsRejectsUnknownAndOmittedParams(t *testing.T) {
	d, v := lookup(t, "general-ledger")
	_, failure := ResolveParams(d, v, journalInput(map[string]string{"company": "OTHER"}))
	wantFailure(t, failure, "REPORT_PARAM_INVALID", "company")
	_, failure = ResolveParams(d, v, journalInput(map[string]string{"by_project": "true"}))
	wantFailure(t, failure, "REPORT_PARAM_INVALID", "by_project")
}

func TestResolveParamsValidatesValues(t *testing.T) {
	d, v := lookup(t, "gl-daily-report")
	long := strings.Repeat("ก", 51)
	many := strings.TrimSuffix(strings.Repeat("A,", 201), ",")
	for _, tc := range []struct {
		name  string
		input map[string]string
		code  string
		field string
	}{
		{"required date", map[string]string{"date_from": "", "date_to": "2026-01-31"}, "REPORT_PARAM_REQUIRED", "date_from"},
		{"impossible date", journalInput(map[string]string{"date_to": "2026-02-30"}), "REPORT_PARAM_INVALID", "date_to"},
		{"short date", journalInput(map[string]string{"date_to": "2026-1-31"}), "REPORT_PARAM_INVALID", "date_to"},
		{"reversed dates", journalInput(map[string]string{"date_from": "2026-02-01"}), "REPORT_DATE_RANGE_INVALID", "date_from"},
		{"reversed accounts", journalInput(map[string]string{"account_from": "5000", "account_to": "1000"}), "REPORT_PARAM_INVALID", "account_from"},
		{"code too long", journalInput(map[string]string{"account_from": long}), "REPORT_PARAM_INVALID", "account_from"},
		{"code with control character", journalInput(map[string]string{"doc_no_from": "JV\x00"}), "REPORT_PARAM_INVALID", "doc_no_from"},
		{"code with comma", journalInput(map[string]string{"account_from": "1000,2000"}), "REPORT_PARAM_INVALID", "account_from"},
		{"too many codes", journalInput(map[string]string{"books": many}), "REPORT_PARAM_INVALID", "books"},
		{"enum outside options", journalInput(map[string]string{"post_status": "deleted"}), "REPORT_PARAM_INVALID", "post_status"},
		{"bool not true/false", journalInput(map[string]string{"day_totals": "yes"}), "REPORT_PARAM_INVALID", "day_totals"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, failure := ResolveParams(d, v, tc.input)
			wantFailure(t, failure, tc.code, tc.field)
		})
	}
}

// injection text is only ever a bound value: it passes as a plain code and the sources never build SQL from input
func TestCodeValuesStayParameters(t *testing.T) {
	d, v := lookup(t, "gl-department-report")
	values, failure := ResolveParams(d, v, map[string]string{"code_from": "D1' OR '1'='1"})
	if failure != nil || values["code_from"] != "D1' OR '1'='1" {
		t.Fatalf("values=%v failure=%v", values, failure)
	}
	for _, file := range []string{"source_gl.go", "source_gl_master.go", "run.go", "params.go"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "Sprintf") {
			t.Errorf("%s formats text with Sprintf; SQL must use $n parameters only", file)
		}
	}
	if !strings.Contains(masterQuery, "WHERE company=$1 AND kind=$2") {
		t.Fatal("master lists must filter by the session company")
	}
}

func TestCheckSort(t *testing.T) {
	d, _ := lookup(t, "gl-daily-report")
	if sort, failure := checkSort(d, ""); failure != nil || sort != "docno" {
		t.Fatalf("default sort = %q %v", sort, failure)
	}
	if sort, failure := checkSort(d, "date"); failure != nil || sort != "date" {
		t.Fatalf("sort = %q %v", sort, failure)
	}
	_, failure := checkSort(d, "amount; DROP TABLE gl_records")
	wantFailure(t, failure, "REPORT_SORT_INVALID", "sort")
}
