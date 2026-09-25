//go:build integration

package fixedasset_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	fa "smlcloudplatform/internal/fixedasset"
	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
)

// TestFixedAssetJournalsUseFiscalYearOfVoucherDate posts to a company whose fiscal years are coded
// in พ.ศ. and start in April (bug 2026-09-25: the ค.ศ. schedule year "2026" was sent as the GL
// fiscal-year code, so post-gl failed with "กรุณาตั้งค่าปีบัญชีก่อนบันทึกรายการ"). Every voucher must
// land in the fiscal year that contains its own date; PostgreSQL is checked after every step.
func TestFixedAssetJournalsUseFiscalYearOfVoucherDate(t *testing.T) {
	db := faIntegrationDB(t)
	ctx := context.Background()
	connect := func(string) (*sql.DB, error) { return db, nil }
	ledger := gl.NewPostgresStore(gl.NewPostgres(connect))
	poster := fa.NewGLPoster(connect, ledger, func(ctx context.Context, s fa.Scope, branch string) error {
		return glhttp.CheckJournalBranch(ctx, connect, gl.Scope{Holding: s.Holding, Company: s.Company, Branch: s.Branch, Actor: s.Actor}, branch)
	})
	store := fa.NewStore(connect)
	headOffice := fa.Scope{Holding: "RUNGRUENG_FA", Company: "01", Branch: "00000", Actor: "fa-it"}
	glScope := gl.Scope{Holding: "RUNGRUENG_FA", Company: "01", Branch: "00000", Actor: "fa-it"}
	now := time.Date(2026, 9, 25, 3, 0, 0, 0, time.UTC)
	nonce := time.Now().UnixNano()
	setup := 0
	run := func(cmd gl.Command) {
		t.Helper()
		setup++
		cmd.RequestID = fmt.Sprintf("fa-fy-setup-%03d-%d", setup, nonce)
		if _, err := ledger.Execute(ctx, glScope, cmd); err != nil {
			t.Fatalf("GL setup %s %s: %v", cmd.Resource, cmd.Action, err)
		}
	}
	count := func(query string, args ...any) int {
		t.Helper()
		var n int
		if err := db.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
			t.Fatalf("%s: %v", query, err)
		}
		return n
	}
	journals := func() int {
		return count(`SELECT count(*) FROM gl_records WHERE company='01' AND kind='journals'`)
	}
	journalYear := func(docNo string) string {
		t.Helper()
		var year string
		if err := db.QueryRowContext(ctx, `SELECT payload->>'fiscalyear' FROM gl_records WHERE company='01' AND kind='journals' AND payload->>'docno'=$1`, docNo).Scan(&year); err != nil {
			t.Fatalf("journal %s: %v", docNo, err)
		}
		return year
	}
	unposted := func(year string, period int) int {
		return count(`SELECT count(*) FROM fa_records WHERE company='01' AND kind='depreciations' AND payload->>'fiscalyear'=$1 AND (payload->>'period')::int=$2 AND NOT COALESCE((payload->>'isposted')::boolean,false)`, year, period)
	}

	// 1. Chart of accounts and three fiscal years coded in พ.ศ., April to March: 2568 is not in
	// use any more, 2569 and 2570 are open.
	for _, a := range []struct {
		code, parent, name, kind, normal string
		level                            int
		posting                          bool
	}{
		{"10000", "", "สินทรัพย์", "asset", "debit", 1, false},
		{"12000", "10000", "สินทรัพย์ไม่หมุนเวียน", "asset", "debit", 2, false},
		{"12100", "12000", "ที่ดิน อาคารและอุปกรณ์", "asset", "debit", 3, false},
		{"12110", "12100", "อาคารและอุปกรณ์", "asset", "debit", 4, true},
		{"12120", "12100", "ค่าเสื่อมราคาสะสม-อาคารและอุปกรณ์", "asset", "credit", 4, true},
		{"30000", "", "ส่วนของเจ้าของ", "equity", "credit", 1, false},
		{"31000", "30000", "กำไรสะสม", "equity", "credit", 2, true},
		{"32000", "30000", "กำไร(ขาดทุน)สุทธิประจำงวด", "equity", "credit", 2, true},
		{"40000", "", "รายได้", "income", "credit", 1, false},
		{"42000", "40000", "รายได้อื่น", "income", "credit", 2, false},
		{"42100", "42000", "กำไร(ขาดทุน)จากการจำหน่ายสินทรัพย์", "income", "credit", 3, true},
		{"50000", "", "ค่าใช้จ่าย", "expense", "debit", 1, false},
		{"52000", "50000", "ค่าใช้จ่ายในการขายและบริหาร", "expense", "debit", 2, false},
		{"52100", "52000", "ค่าเสื่อมราคา-อาคารและอุปกรณ์", "expense", "debit", 3, true},
	} {
		run(gl.Command{Resource: "accounts", Action: "create", Account: &gl.Account{
			AccountCode: a.code, ParentAccountCode: a.parent, Names: []gl.Name{{Code: "th", Name: a.name}},
			AccountType: a.kind, NormalBalance: a.normal, Level: a.level, AllowPosting: a.posting, IsActive: true,
		}})
	}
	for _, y := range []gl.FiscalYear{
		{Code: "2568", StartDate: "2025-04-01", EndDate: "2026-03-31", IsActive: false},
		{Code: "2569", StartDate: "2026-04-01", EndDate: "2027-03-31", IsActive: true},
		{Code: "2570", StartDate: "2027-04-01", EndDate: "2028-03-31", IsActive: true},
	} {
		y.Scale, y.ProfitLossAccount, y.RetainedEarningsAccount = 2, "32000", "31000"
		run(gl.Command{Resource: "fiscal-years", Action: "create", FiscalYear: &y})
	}
	if n := count(`SELECT count(*) FROM gl_records WHERE company='01' AND kind='fiscal-years' AND code IN ('2568','2569','2570')`); n != 3 {
		t.Fatalf("expected the 3 พ.ศ. fiscal years, got %d", n)
	}

	// 2. A packing machine bought 1 January 2026: its schedule is keyed by ค.ศ. calendar month.
	if _, err := store.CreateAssetType(ctx, headOffice, fa.AssetType{
		TypeCode: "MACHINERY", Names: []fa.Name{{Code: "th", Name: "เครื่องจักรและอุปกรณ์"}},
		AssetAccountCode: "12110", AccumDeprecAccountCode: "12120", DeprecExpenseAccountCode: "52100",
	}, now); err != nil {
		t.Fatal(err)
	}
	machine := fa.Asset{
		AssetCode: "FA-2569-010", Names: []fa.Name{{Code: "th", Name: "เครื่องบรรจุปูนซีเมนต์อัตโนมัติ"}},
		AssetTypeCode: "MACHINERY", Cost: "120000.00", ScrapValue: "1.00", UsefulLifeYears: 5, DeprecPercent: "20.00",
		PurchaseDate: "2026-01-01", StartCalcDate: "2026-01-01", Status: "active",
	}
	if _, err := store.CreateAsset(ctx, headOffice, machine, now); err != nil {
		t.Fatal(err)
	}
	if n := unposted("2026", 4); n != 1 {
		t.Fatalf("expected the April 2026 schedule row, got %d", n)
	}

	// 3. Refusals come before any write and point at the field to fix: the period's year is not in
	// use, the voucher date has no fiscal year, the date is malformed, the year is พ.ศ., or the date
	// is in another fiscal year than the period.
	_, err := poster.PostDepreciation(ctx, headOffice, "2026", 1, "", "", "", now)
	mustUserError(t, err, "fa_fiscal_year_closed", "fiscalyear", "2026-01-31", "2568")
	_, err = poster.PostDepreciation(ctx, headOffice, "2027", 1, "2028-06-30", "", "", now)
	mustUserError(t, err, "fa_fiscal_year_not_found", "date", "2028-06-30", "ปีบัญชีและบัญชีปิดปี")
	_, err = poster.PostDepreciation(ctx, headOffice, "2026", 1, "31/01/2026", "", "", now)
	mustUserError(t, err, "fa_date_invalid", "date")
	_, err = poster.PostDepreciation(ctx, headOffice, "2569", 4, "", "", "", now)
	mustUserError(t, err, "fa_post_year_invalid", "fiscalyear", "ค.ศ.")
	_, err = poster.PostDepreciation(ctx, headOffice, "2027", 3, "2027-04-05", "", "", now)
	mustUserError(t, err, "fa_date_outside_period_year", "date", "2570", "2569", "2027-03-31")
	if n := journals(); n != 0 {
		t.Fatalf("refused postings must not write a journal, found %d", n)
	}
	if n := unposted("2026", 1); n != 1 {
		t.Fatalf("refused postings must leave January unposted, got %d", n)
	}

	// 4. The screen sends no date: the voucher is dated the period's last day and lands in the
	// พ.ศ. fiscal year of that day.
	april, err := poster.PostDepreciation(ctx, headOffice, "2026", 4, "", "", "", now)
	if err != nil {
		t.Fatalf("PostDepreciation April: %v", err)
	}
	if april.Date != "2026-04-30" {
		t.Fatalf("a blank date must be the period end 2026-04-30, got %s", april.Date)
	}
	if april.FiscalYear != "2569" || journalYear(april.DocNo) != "2569" {
		t.Fatalf("April journal %s: returned year %q, stored year %q", april.DocNo, april.FiscalYear, journalYear(april.DocNo))
	}
	var balanced, matchesSchedule bool
	if err := db.QueryRowContext(ctx, `SELECT SUM(debit)=SUM(credit),
		SUM(debit)=(SELECT SUM((payload->>'perioddeprec')::numeric) FROM fa_records WHERE company='01' AND kind='depreciations' AND payload->>'fiscalyear'='2026' AND (payload->>'period')::int=4)
		FROM gl_lines WHERE company='01' AND doc_no=$1`, april.DocNo).Scan(&balanced, &matchesSchedule); err != nil {
		t.Fatal(err)
	}
	if !balanced || !matchesSchedule {
		t.Fatalf("April lines balanced=%v equal the schedule=%v", balanced, matchesSchedule)
	}
	if n := count(`SELECT count(*) FROM fa_records WHERE company='01' AND kind='depreciations' AND payload->>'journaldocno'=$1 AND (payload->>'isposted')::boolean`, april.DocNo); n != 1 {
		t.Fatalf("the April row must point at %s, got %d", april.DocNo, n)
	}

	// 5. One ค.ศ. year spans two fiscal years: January 2027 is still 2569, April 2027 is 2570.
	// March 2027 posted in May (today is in 2570) still goes into 2569, dated 31 March.
	march, err := poster.PostDepreciation(ctx, headOffice, "2027", 3, "", "", "", time.Date(2027, 5, 10, 3, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("PostDepreciation March 2027: %v", err)
	}
	if march.Date != "2027-03-31" || journalYear(march.DocNo) != "2569" {
		t.Fatalf("March 2027 posted in May: date %s year %s; want 2027-03-31 in 2569", march.Date, journalYear(march.DocNo))
	}
	january, err := poster.PostDepreciation(ctx, headOffice, "2027", 1, "2027-01-31", "", "", now)
	if err != nil {
		t.Fatalf("PostDepreciation January 2027: %v", err)
	}
	april27, err := poster.PostDepreciation(ctx, headOffice, "2027", 4, "2027-04-30", "", "", now)
	if err != nil {
		t.Fatalf("PostDepreciation April 2027: %v", err)
	}
	if journalYear(january.DocNo) != "2569" || journalYear(april27.DocNo) != "2570" {
		t.Fatalf("January 2027 → %s, April 2027 → %s; want 2569 and 2570", journalYear(january.DocNo), journalYear(april27.DocNo))
	}

	// 6. Reversal goes into the year of the reversal date and frees the schedule row again.
	if err := poster.ReverseDepreciation(ctx, headOffice, april.DocNo, "ผ่านรายการค่าเสื่อมราคาผิดงวด", now); err != nil {
		t.Fatalf("ReverseDepreciation: %v", err)
	}
	if got := journalYear("REV-" + april.DocNo); got != "2569" {
		t.Fatalf("the reversal dated %s must be in 2569, got %q", now.Format("2006-01-02"), got)
	}
	if n := unposted("2026", 4); n != 1 {
		t.Fatalf("the reversed April row must be unposted again, got %d", n)
	}
	// A reversal date outside every fiscal year is explained in Thai, and nothing changes.
	before := journals()
	err = poster.ReverseDepreciation(ctx, headOffice, april27.DocNo, "ผ่านรายการค่าเสื่อมราคาผิดงวด", time.Date(2028, 5, 2, 3, 0, 0, 0, time.UTC))
	mustUserError(t, err, "fa_fiscal_year_not_found", "", "2028-05-02")
	if n := journals(); n != before {
		t.Fatalf("a refused reversal must not write a journal, %d → %d", before, n)
	}
	if n := count(`SELECT count(*) FROM gl_records WHERE company='01' AND kind='journals' AND payload->>'docno'=$1 AND payload->>'status'='posted'`, april27.DocNo); n != 1 {
		t.Fatal("the April 2027 journal must stay posted")
	}

	// 7. Disposal: refused in the year no longer in use, posted into 2570 by its date otherwise.
	writeOff := fa.AssetDisposal{AssetCode: machine.AssetCode, DisposalDate: "2026-03-15", DisposalType: "write_off", GainLossAccountCode: "42100", Reason: "เครื่องบรรจุปูนซีเมนต์ชำรุดใช้งานไม่ได้"}
	_, _, err = poster.DisposeAsset(ctx, headOffice, writeOff, now)
	mustUserError(t, err, "fa_fiscal_year_closed", "disposaldate", "2568")
	if n := count(`SELECT count(*) FROM fa_records WHERE company='01' AND kind='assets' AND code=$1 AND payload->>'status'='active'`, machine.AssetCode); n != 1 {
		t.Fatal("a refused disposal must leave the machine active")
	}
	writeOff.DisposalDate = "2027-06-30"
	_, disposalJournal, err := poster.DisposeAsset(ctx, headOffice, writeOff, now)
	if err != nil {
		t.Fatalf("DisposeAsset: %v", err)
	}
	if disposalJournal.FiscalYear != "2570" || journalYear(disposalJournal.DocNo) != "2570" {
		t.Fatalf("write-off journal %s: returned year %q, stored year %q", disposalJournal.DocNo, disposalJournal.FiscalYear, journalYear(disposalJournal.DocNo))
	}
	if n := count(`SELECT count(*) FROM fa_records WHERE company='01' AND kind='assets' AND code=$1 AND payload->>'status'='disposed'`, machine.AssetCode); n != 1 {
		t.Fatal("the written-off machine must be disposed")
	}
	if n := count(`SELECT count(*) FROM gl_records WHERE company='01' AND kind='journals' AND payload->>'fiscalyear' NOT IN ('2569','2570')`); n != 0 {
		t.Fatalf("every journal must carry a พ.ศ. fiscal-year code, %d do not", n)
	}
}
