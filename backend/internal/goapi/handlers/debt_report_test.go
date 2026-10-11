package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"smlcloudplatform/internal/goapi/language"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

var debtLedgerReportCodes = []string{"ap_movement", "ap_status", "ap_outstanding", "ap_daily_payment", "ar_movement", "ar_status", "ar_outstanding", "ar_credit_limit"}

// TestBuildDebtReportQueryScopesToOneCompany - เดิม WHERE (d.company = $1 OR d.company = '01' OR $1 = ”)
// ทำให้ผู้ใช้บริษัท 02 เห็นเอกสารของบริษัท 01 และรหัสว่างถูกแทนเป็น 01 — ต้องกรองบริษัทเดียว และรหัสว่าง = ข้อผิดพลาด
func TestBuildDebtReportQueryScopesToOneCompany(t *testing.T) {
	fromDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for _, code := range debtLedgerReportCodes {
		t.Run(code, func(t *testing.T) {
			q, args, err := BuildDebtReportQuery(code, "02", fromDate, toDate, "", 20, 0)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(q, "WHERE d.company = $1 AND d.ledger") || strings.Contains(q, "'01'") || strings.Contains(q, "$1 = ''") {
				t.Fatalf("company filter is not strict:\n%s", q)
			}
			if len(args) == 0 || args[0] != "02" {
				t.Fatalf("first arg = %v, want company 02", args)
			}
			for _, empty := range []string{"", "   "} {
				if _, _, err := BuildDebtReportQuery(code, empty, fromDate, toDate, "", 20, 0); !errors.Is(err, errReportCompanyRequired) {
					t.Fatalf("company %q: err = %v, want errReportCompanyRequired", empty, err)
				}
			}
		})
	}
}

// TestReportCompanyRequiredFailIsCodedBadRequest - รหัสบริษัทว่างที่หลุดถึงตัวสร้าง query ตอบ 400 พร้อมรหัสและข้อความตามภาษา
func TestReportCompanyRequiredFailIsCodedBadRequest(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	ctx := e.NewContext(httptest.NewRequest(http.MethodPost, "/?lang=en", nil), rec)
	if err := reportCompanyRequiredFail(ctx); err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	want := language.Text("company_required", "en")
	if rec.Code != http.StatusBadRequest || body["code"] != "COMPANY_REQUIRED" || body["message"] != want || want == "company_required" {
		t.Fatalf("status=%d body=%v want message %q", rec.Code, body, want)
	}
}

// TestCompanyReportsRejectMissingCompany - ผู้ใช้ที่ยังไม่เลือกบริษัทต้องไม่ได้ข้อมูลบริษัทใด และข้อความแปลตามภาษาผู้ใช้
func TestCompanyReportsRejectMissingCompany(t *testing.T) {
	user := &msmodels.UserInfo{HoldingCode: "H1"}
	const dates = `"holdingcode":"H1","fromdate":"2026-09-01","todate":"2026-09-30"`
	cases := map[string]struct {
		handler echo.HandlerFunc
		body    string
	}{
		"debt":              {DebtReportHandler, `{"holdingcode":"H1","reportcode":"ar_status"}`},
		"customer labels":   {CustomerLabelsHandler, `{"holdingcode":"H1"}`},
		"sales by document": {SalesReportByDocumentHandler, `{` + dates + `}`},
		"sales summary":     {SalesReportSummaryHandler, `{` + dates + `}`},
	}
	want := language.Text("company_required", "en")
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec := callTaxHandlerLang(t, tc.handler, "/", "en", tc.body, user)
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusConflict || body["code"] != "COMPANY_REQUIRED" || body["message"] != want {
				t.Fatalf("status=%d body=%v want message %q", rec.Code, body, want)
			}
		})
	}
}

func TestBuildDebtReportQuery(t *testing.T) {
	fromDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		reportCode string
		partyCode  string
		hasDate    bool
		wantSub    string
		wantArgs   int
	}{
		{
			reportCode: "ap_movement",
			partyCode:  "AP001",
			hasDate:    true,
			wantSub:    "d.ledger = 'ap'",
			wantArgs:   6, // company, fromDate, toDate, partyCode, limit, offset
		},
		{
			reportCode: "ap_status",
			partyCode:  "AP002",
			hasDate:    false,
			wantSub:    "d.ledger = 'ap'",
			wantArgs:   4, // company, partyCode, limit, offset
		},
		{
			reportCode: "ap_outstanding",
			partyCode:  "",
			hasDate:    false,
			wantSub:    "HAVING (d.amount - COALESCE(SUM(s.amount), 0)) > 0",
			wantArgs:   3, // company, limit, offset
		},
		{
			reportCode: "ap_daily_payment",
			partyCode:  "AP003",
			hasDate:    true,
			wantSub:    "d.ledger = 'ap' AND d.side = 1",
			wantArgs:   6, // company, fromDate, toDate, partyCode, limit, offset
		},
		{
			reportCode: "ar_movement",
			partyCode:  "AR001",
			hasDate:    true,
			wantSub:    "d.ledger = 'ar'",
			wantArgs:   6, // company, fromDate, toDate, partyCode, limit, offset
		},
		{
			reportCode: "ar_status",
			partyCode:  "AR002",
			hasDate:    false,
			wantSub:    "d.ledger = 'ar'",
			wantArgs:   4, // company, partyCode, limit, offset
		},
		{
			reportCode: "ar_outstanding",
			partyCode:  "",
			hasDate:    false,
			wantSub:    "HAVING (d.amount - COALESCE(SUM(s.amount), 0)) > 0",
			wantArgs:   3, // company, limit, offset
		},
		{
			reportCode: "ar_credit_limit",
			partyCode:  "AR003",
			hasDate:    false,
			wantSub:    "creditday",
			wantArgs:   4, // company, partyCode, limit, offset
		},
	}

	for _, tt := range tests {
		t.Run(tt.reportCode, func(t *testing.T) {
			fd := time.Time{}
			td := time.Time{}
			if tt.hasDate {
				fd = fromDate
				td = toDate
			}

			q, args, err := BuildDebtReportQuery(tt.reportCode, "biz01", fd, td, tt.partyCode, 100, 0)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tt.reportCode, err)
			}
			if !strings.Contains(q, tt.wantSub) {
				t.Errorf("%s: query missing substring %q\nQuery: %s", tt.reportCode, tt.wantSub, q)
			}
			if len(args) != tt.wantArgs {
				t.Errorf("%s: expected %d args, got %d", tt.reportCode, tt.wantArgs, len(args))
			}
		})
	}

	// Test invalid report code
	_, _, err := BuildDebtReportQuery("invalid_report", "biz01", time.Time{}, time.Time{}, "", 100, 0)
	if err == nil {
		t.Fatal("expected error for invalid report code, got nil")
	}
}

func TestAgingAgeFollowsChampProDiffYear(t *testing.T) {
	asOf := time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		ref  time.Time
		want int
	}{
		{asOf, 1},                     // ถึงกำหนดวันนี้ = วันที่ 1
		{asOf.AddDate(0, 0, -6), 7},   // ครบ 7 วัน
		{asOf.AddDate(0, 0, -7), 8},   // วันที่ 8
		{asOf.AddDate(0, 0, 1), -1},   // ยังไม่ถึงกำหนด
		{asOf.AddDate(0, 0, 30), -30}, // ยังไม่ถึงกำหนดอีก 30 วัน
		{time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC), 957},
	}
	for _, tt := range tests {
		if got := agingAge(asOf, tt.ref); got != tt.want {
			t.Errorf("agingAge(%s, %s) = %d, want %d", asOf.Format("2006-01-02"), tt.ref.Format("2006-01-02"), got, tt.want)
		}
	}
}

func TestAgingBucketChampBoundaries(t *testing.T) {
	gaps := []AgingRange{{1, 7}, {10, 14}, {15, 21}, {22, 28}}
	tests := []struct {
		name   string
		age    int
		ranges []AgingRange
		over   int
		want   string
	}{
		{"negative = not yet due", -3, defaultAgingRanges, 28, "notinrange"},
		{"zero = not in range", 0, defaultAgingRanges, 28, "notinrange"},
		{"first day of range 1", 1, defaultAgingRanges, 28, "range1"},
		{"last day of range 1", 7, defaultAgingRanges, 28, "range1"},
		{"first day of range 2", 8, defaultAgingRanges, 28, "range2"},
		{"range 3", 15, defaultAgingRanges, 28, "range3"},
		{"last day of range 4", 28, defaultAgingRanges, 28, "range4"},
		{"over 28", 29, defaultAgingRanges, 28, "overrange"},
		{"gap between ranges", 8, gaps, 28, "notinrange"},
		{"gap between range 4 and over", 30, defaultAgingRanges, 30, "notinrange"},
		{"above wider over", 31, defaultAgingRanges, 30, "overrange"},
		{"range starting at 0 still excludes not-yet-due", -1, []AgingRange{{0, 7}, {8, 14}, {15, 21}, {22, 28}}, 28, "notinrange"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := agingBucket(tt.age, tt.ranges, tt.over); got != tt.want {
				t.Errorf("agingBucket(%d) = %s, want %s", tt.age, got, tt.want)
			}
		})
	}
}

func TestNormalizeAgingCondition(t *testing.T) {
	now := time.Date(2026, 10, 10, 18, 30, 0, 0, time.UTC) // 01:30 วันที่ 11 เวลาไทย
	intPtr := func(v int) *int { return &v }

	cond, key := normalizeAgingCondition(DebtReportRequest{}, now)
	if key != "" {
		t.Fatalf("defaults rejected: %s", key)
	}
	if got := cond.AsOf.Format("2006-01-02"); got != "2026-10-11" {
		t.Errorf("default as-of = %s, want Bangkok today 2026-10-11", got)
	}
	if cond.Basis != "duedate" || cond.OverDays != 28 || len(cond.Ranges) != 4 || cond.Ranges[0] != (AgingRange{1, 7}) || cond.Ranges[3] != (AgingRange{22, 28}) {
		t.Errorf("Champ defaults not applied: %+v", cond)
	}

	cond, key = normalizeAgingCondition(DebtReportRequest{AsOfDate: "2026-09-30", AgeBasis: "docdate", Ranges: []AgingRange{{1, 30}, {31, 60}, {61, 90}, {91, 120}}}, now)
	if key != "" || cond.OverDays != 120 || cond.Basis != "docdate" {
		t.Errorf("custom ranges: key=%q cond=%+v", key, cond)
	}

	invalid := []struct {
		name string
		req  DebtReportRequest
		want string
	}{
		{"bad date", DebtReportRequest{AsOfDate: "30/09/2026"}, "aging_condition_invalid"},
		{"bad basis", DebtReportRequest{AgeBasis: "paydate"}, "aging_condition_invalid"},
		{"three ranges", DebtReportRequest{Ranges: []AgingRange{{1, 7}, {8, 14}, {15, 21}}}, "aging_ranges_invalid"},
		{"from after to", DebtReportRequest{Ranges: []AgingRange{{7, 1}, {8, 14}, {15, 21}, {22, 28}}}, "aging_ranges_invalid"},
		{"overlap", DebtReportRequest{Ranges: []AgingRange{{1, 7}, {7, 14}, {15, 21}, {22, 28}}}, "aging_ranges_invalid"},
		{"negative from", DebtReportRequest{Ranges: []AgingRange{{-1, 7}, {8, 14}, {15, 21}, {22, 28}}}, "aging_ranges_invalid"},
		{"over inside range 4", DebtReportRequest{OverDays: intPtr(27)}, "aging_ranges_invalid"},
		{"too many days", DebtReportRequest{OverDays: intPtr(agingMaxDays + 1)}, "aging_ranges_invalid"},
	}
	for _, tt := range invalid {
		if _, got := normalizeAgingCondition(tt.req, now); got != tt.want {
			t.Errorf("%s: key = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestAgingTotalsSplitsDebtAndReductionsExactly(t *testing.T) {
	cond := agingCondition{AsOf: time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC), Basis: "duedate", Ranges: defaultAgingRanges, OverDays: 28}
	var totals agingTotals
	add := func(ref string, isDebt bool, amount string, count int64) {
		t.Helper()
		if err := totals.add(cond, ref, isDebt, amount, count); err != nil {
			t.Fatal(err)
		}
	}
	add("2026-10-20", true, "100.10", 1)        // ยังไม่ถึงกำหนด
	add("2026-10-11", true, "0.10", 1)          // วันที่ 1 -> ช่วง 1
	add("2026-10-04", true, "0.20", 2)          // วันที่ 8 -> ช่วง 2
	add("2026-09-13", true, "1000.00000000", 1) // วันที่ 29 -> เกิน 28
	add("2026-10-01", false, "50.05", 1)        // ใบลดหนี้/รับเงินที่ยังไม่ตัด
	if err := totals.add(cond, "bad", true, "1", 1); err == nil {
		t.Fatal("expected error for bad reference date")
	}
	row := totals.row()
	want := map[string]string{"notinrange": "100.10", "range1": "0.10", "range2": "0.20", "range3": "0.00", "range4": "0.00", "overrange": "1000.00", "debtdecrease": "-50.05", "earnestamount": "0.00", "totaldue": "1050.35"}
	for key, value := range want {
		if row[key] != value {
			t.Errorf("%s = %v, want %s", key, row[key], value)
		}
	}
	if row["doccount"] != int64(6) {
		t.Errorf("doccount = %v, want 6", row["doccount"])
	}
	if got := reportMoneyText(decimal.RequireFromString("1.23456")); got != "1.23456" {
		t.Errorf("precision lost: %s", got)
	}
}

func TestBuildAgingQuery(t *testing.T) {
	cond := agingCondition{AsOf: time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC), Basis: "docdate", Ranges: defaultAgingRanges, OverDays: 28}
	ar, err := buildAgingQuery("ar_aging", "01", cond, "C101", 50, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(ar.args) != 4 || ar.args[0] != "01" || ar.args[1] != "2026-09-30" || ar.args[2] != "docdate" || ar.args[3] != "C101" {
		t.Errorf("args = %v", ar.args)
	}
	if len(ar.rowArgs) != 6 || ar.rowArgs[4] != 50 || ar.rowArgs[5] != 100 {
		t.Errorf("row args = %v", ar.rowArgs)
	}
	for _, sub := range []string{
		"d.ledger = 'ar'",
		"d.company = $1",
		"s.settlement_date <= $2::date",
		"s.reversed_at >= (($2::date + 1)::timestamp AT TIME ZONE 'Asia/Bangkok')",
		"NULLIF(d.payload->>'document_date','')",
		"CASE WHEN $3 = 'docdate' THEN docdate ELSE duedate END",
		"ORDER BY partner_code ASC LIMIT $5 OFFSET $6",
		"d.side = 1 AS isdebt",
		// only journals posted as of the date count (draft and reversed-by-then journals drop out)
		"r.kind = 'journals' AND NOT COALESCE((r.payload->>'isdeleted')::boolean, false)",
		"r.payload->>'date' <= to_char($2::date, 'YYYY-MM-DD')",
		"(r.payload->>'postedat')::timestamptz < (($2::date + 1)::timestamp AT TIME ZONE 'Asia/Bangkok')",
		"r.payload->>'status' = 'posted' OR (r.payload->>'status' = 'reversed' AND NOT EXISTS",
		"v.payload->>'reversalof' = r.id",
		"FROM gl_subledger_allocations a JOIN posted_journals j ON j.id = a.journal_id",
		"a.reversed_effective_date > $2::date",
		"FROM gl_subledger_settlements s JOIN posted_journals j ON j.id = s.journal_id",
		"s.created_at < (($2::date + 1)::timestamp AT TIME ZONE 'Asia/Bangkok') AND s.settlement_date <= $2::date",
		"CASE WHEN d.payload->>'balance_side' IS NOT NULL THEN COALESCE(pa.amount, 0)",
		"WHEN cj.id IS NOT NULL THEN d.amount ELSE 0 END - COALESCE(ps.amount, 0) AS remaining",
		"LEFT JOIN posted_journals cj ON cj.id = d.created_journal_id",
	} {
		if !strings.Contains(ar.rows, sub) {
			t.Errorf("rows query missing %q", sub)
		}
	}
	for name, query := range map[string]string{"totals": ar.totals, "count": ar.count} {
		if !strings.Contains(query, "JOIN posted_journals j ON j.id = s.journal_id") || !strings.Contains(query, "LEFT JOIN posted_journals cj") {
			t.Errorf("%s query must use posted journals only", name)
		}
	}
	if strings.Contains(ar.rows, "d.amount - COALESCE") {
		t.Error("document amount must not count before its journal is posted")
	}
	if strings.Contains(ar.rows, "'01' OR") || strings.Contains(ar.totals, "LIMIT") || strings.Contains(ar.count, "LIMIT") {
		t.Error("aging must stay inside the company and totals must cover every partner")
	}
	ap, err := buildAgingQuery("ap_aging", "01", cond, "", 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ap.rows, "d.ledger = 'ap'") || !strings.Contains(ap.rows, "WHEN d.payload->>'balance_side' IS NOT NULL THEN d.side = 1 ELSE d.side = 2") {
		t.Errorf("ap query lost ledger or debt side rule:\n%s", ap.rows)
	}
	if _, err := buildAgingQuery("ar_status", "01", cond, "", 50, 0); err == nil {
		t.Error("expected error for non-aging report code")
	}
}
