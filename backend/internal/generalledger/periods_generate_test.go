package generalledger

import (
	"strings"
	"testing"
	"time"
)

// assertCoversYear: kept + planned periods cover the fiscal year day by day — no gap, no overlap.
func assertCoversYear(t *testing.T, year FiscalYear, periods []Master) {
	t.Helper()
	byStart := map[string]Master{}
	for _, p := range periods {
		if _, dup := byStart[p.StartDate]; dup {
			t.Fatalf("two periods start on %s", p.StartDate)
		}
		byStart[p.StartDate] = p
	}
	day := year.StartDate
	for count := 0; count <= len(periods); count++ {
		p, ok := byStart[day]
		if !ok {
			t.Fatalf("gap: no period starts on %s", day)
		}
		if p.EndDate == year.EndDate {
			if count+1 != len(periods) {
				t.Fatalf("periods past year end: %+v", periods)
			}
			return
		}
		end, _ := time.Parse(time.DateOnly, p.EndDate)
		day = end.AddDate(0, 0, 1).Format(time.DateOnly)
	}
	t.Fatalf("periods do not reach year end %s", year.EndDate)
}

func TestPlanMonthlyPeriodsCalendarYear(t *testing.T) {
	year := FiscalYear{Code: "2569", StartDate: "2026-01-01", EndDate: "2026-12-31"}
	planned, err := planMonthlyPeriods(year, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(planned) != 12 || planned[0].Code != "2569-01" || planned[11].Code != "2569-12" {
		t.Fatalf("want 12 periods 2569-01..2569-12, got %+v", planned)
	}
	if planned[1].StartDate != "2026-02-01" || planned[1].EndDate != "2026-02-28" || planned[0].FiscalYear != "2569" || !planned[0].IsActive || planned[0].Kind != "periods" {
		t.Fatalf("february/fields wrong: %+v / %+v", planned[1], planned[0])
	}
	assertCoversYear(t, year, planned)
}

// A first year that starts mid-month gets a short first period; the rest are calendar (VAT) months.
func TestPlanMonthlyPeriodsMidMonthAndLeapYear(t *testing.T) {
	year := FiscalYear{Code: "Y1", StartDate: "2027-06-15", EndDate: "2028-06-14"}
	planned, err := planMonthlyPeriods(year, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(planned) != 13 || planned[0].EndDate != "2027-06-30" || planned[12].StartDate != "2028-06-01" || planned[12].Code != "Y1-13" {
		t.Fatalf("mid-month split wrong: %+v", planned)
	}
	if planned[8].StartDate != "2028-02-01" || planned[8].EndDate != "2028-02-29" {
		t.Fatalf("leap february wrong: %+v", planned[8])
	}
	assertCoversYear(t, year, planned)
}

// Months that already have exactly their period are skipped; numbering stays the month number.
func TestPlanMonthlyPeriodsSkipsExistingMonths(t *testing.T) {
	year := FiscalYear{Code: "2569", StartDate: "2026-01-01", EndDate: "2026-12-31"}
	existing := []Master{{Code: "P01", StartDate: "2026-01-01", EndDate: "2026-01-31"}, {Code: "P05", StartDate: "2026-05-01", EndDate: "2026-05-31"}}
	planned, err := planMonthlyPeriods(year, existing)
	if err != nil {
		t.Fatal(err)
	}
	if len(planned) != 10 || planned[0].Code != "2569-02" || planned[3].Code != "2569-06" {
		t.Fatalf("existing months not skipped: %+v", planned)
	}
	assertCoversYear(t, year, append(planned, existing...))
	if again, err := planMonthlyPeriods(year, append(planned, existing...)); err != nil || len(again) != 0 {
		t.Fatalf("complete year should plan nothing: %v %+v", err, again)
	}
}

// A period that cuts across a month would leave an overlap or a gap: the whole plan is refused.
func TestPlanMonthlyPeriodsRejectsOverlap(t *testing.T) {
	year := FiscalYear{Code: "2569", StartDate: "2026-01-01", EndDate: "2026-12-31"}
	for _, existing := range []Master{
		{Code: "Q1", StartDate: "2026-01-01", EndDate: "2026-03-31"},   // quarter spans three months
		{Code: "HALF", StartDate: "2026-04-10", EndDate: "2026-04-30"}, // part of a month
		{Code: "PREV", StartDate: "2025-12-01", EndDate: "2026-01-15"}, // reaches in from outside
	} {
		_, err := planMonthlyPeriods(year, []Master{existing})
		user, ok := AsUserError(err)
		if !ok || user.Code != "period_generate_overlap" || len(user.Args) != 1 || user.Args[0] != existing.Code {
			t.Fatalf("%s: want period_generate_overlap, got %v", existing.Code, err)
		}
	}
}

func TestPlanMonthlyPeriodsRejectsRunawayYear(t *testing.T) {
	_, err := planMonthlyPeriods(FiscalYear{Code: "X", StartDate: "2026-01-01", EndDate: "2062-12-31"}, nil)
	if user, ok := AsUserError(err); !ok || user.Code != "periods_too_many" {
		t.Fatalf("want periods_too_many, got %v", err)
	}
	if planned, err := planMonthlyPeriods(FiscalYear{Code: "X", StartDate: "2026-01-01", EndDate: "2027-12-31"}, nil); err != nil || len(planned) != MaxGeneratedPeriods {
		t.Fatalf("24 months must still plan: %v %d", err, len(planned))
	}
}

func TestValidatePostBatchFilter(t *testing.T) {
	for _, c := range []struct {
		filter PostBatchInput
		code   string
	}{
		{PostBatchInput{}, ""},
		{PostBatchInput{DateFrom: "2026-01-01", DateTo: "2026-01-31", DocNoFrom: "JV001", DocNoTo: "JV999", BookCode: "JV"}, ""},
		{PostBatchInput{DateFrom: "2026-13-01"}, "post_batch_date_invalid"},
		{PostBatchInput{DateFrom: "2026-02-01", DateTo: "2026-01-31"}, "post_batch_date_range_invalid"},
		{PostBatchInput{DocNoFrom: "JV9", DocNoTo: "JV1"}, "post_batch_docno_range_invalid"},
	} {
		err := validatePostBatchFilter(c.filter)
		if c.code == "" {
			if err != nil {
				t.Fatalf("%+v: want valid, got %v", c.filter, err)
			}
			continue
		}
		if user, ok := AsUserError(err); !ok || user.Code != c.code {
			t.Fatalf("%+v: want %s, got %v", c.filter, c.code, err)
		}
	}
}

func TestPostBatchSignatoriesRules(t *testing.T) {
	c1, c2, ap, err := postBatchSignatories(PostBatchInput{Check1: "  นางสาวสุดา มั่นคง ", Approver: "	นายสมชาย ใจดี "})
	if err != nil || c1 != "นางสาวสุดา มั่นคง" || c2 != "" || ap != "นายสมชาย ใจดี" {
		t.Fatalf("names must be trimmed and reviewers optional: %q %q %q %v", c1, c2, ap, err)
	}
	fifty := strings.Repeat("ก", PostBatchPersonMaxRunes)
	if _, _, _, err := postBatchSignatories(PostBatchInput{Check1: fifty, Check2: fifty, Approver: fifty}); err != nil {
		t.Fatalf("50 characters (150 bytes) must be allowed: %v", err)
	}
	for _, c := range []struct {
		input PostBatchInput
		code  string
		field string
	}{
		{PostBatchInput{}, "post_batch_approver_required", "approver"},
		{PostBatchInput{Approver: " 	 ", Check1: "นางสาวสุดา มั่นคง"}, "post_batch_approver_required", "approver"},
		{PostBatchInput{Approver: "นายสมชาย ใจดี", Check1: fifty + "ข"}, "post_batch_person_too_long", "check1"},
		{PostBatchInput{Approver: "นายสมชาย ใจดี", Check2: fifty + "ข"}, "post_batch_person_too_long", "check2"},
		{PostBatchInput{Approver: fifty + "ข"}, "post_batch_person_too_long", "approver"},
	} {
		_, _, _, err := postBatchSignatories(c.input)
		if user, ok := AsUserError(err); !ok || user.Code != c.code || user.Field != c.field {
			t.Fatalf("%+v: want %s on %s, got %v", c.input, c.code, c.field, err)
		}
	}
}

func TestPostIssueKeepsUserTextHidesSystemFaults(t *testing.T) {
	if code, msg, ok := postIssue(userError(CodeStaleVersion, "รายการนี้เปลี่ยนไปแล้ว")); !ok || code != CodeStaleVersion || msg == "" {
		t.Fatalf("user error lost: %s %s %v", code, msg, ok)
	}
	if code, _, ok := postIssue(ErrNotFound); !ok || code != "not_found" {
		t.Fatalf("not found should be a row issue: %s %v", code, ok)
	}
	if _, _, ok := postIssue(errString("pq: SQLSTATE 40001 serialization failure")); ok {
		t.Fatal("database fault must abort the check, not become a row message")
	}
}

type errString string

func (e errString) Error() string { return string(e) }
