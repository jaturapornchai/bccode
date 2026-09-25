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

	journalStatus := func(docNo string) string {
		t.Helper()
		var status string
		if err := db.QueryRowContext(ctx, `SELECT payload->>'status' FROM gl_records WHERE company='01' AND kind='journals' AND payload->>'docno'=$1`, docNo).Scan(&status); err != nil {
			t.Fatalf("journal %s: %v", docNo, err)
		}
		return status
	}
	journalDate := func(docNo string) string {
		t.Helper()
		var date string
		if err := db.QueryRowContext(ctx, `SELECT payload->>'date' FROM gl_records WHERE company='01' AND kind='journals' AND payload->>'docno'=$1`, docNo).Scan(&date); err != nil {
			t.Fatalf("journal %s: %v", docNo, err)
		}
		return date
	}
	setYearClosed := func(code string, closed bool) {
		t.Helper()
		if _, err := db.ExecContext(ctx, `UPDATE gl_records SET payload = jsonb_set(payload, '{closed}', to_jsonb($2::boolean)) WHERE company='01' AND kind='fiscal-years' AND code=$1`, code, closed); err != nil {
			t.Fatal(err)
		}
	}

	// 6. Reversal while the original's fiscal year is open: dated on the original's date, so April
	// nets to zero in April (Champ re-transfers into the same period), and the row is free again.
	if err := poster.ReverseDepreciation(ctx, headOffice, april.DocNo, "ผ่านรายการค่าเสื่อมราคาผิดงวด", now); err != nil {
		t.Fatalf("ReverseDepreciation: %v", err)
	}
	if got, year := journalDate("REV-"+april.DocNo), journalYear("REV-"+april.DocNo); got != april.Date || year != "2569" {
		t.Fatalf("the reversal of %s is dated %s in %q, want %s in 2569", april.DocNo, got, year, april.Date)
	}
	if n := unposted("2026", 4); n != 1 {
		t.Fatalf("the reversed April row must be unposted again, got %d", n)
	}

	// 6b. April posted again after its reversal: the reversed journal keeps its number for audit
	// (replaying its GL request would post nothing), so the new journal is <number>-2, and April
	// holds exactly one month of depreciation while no other month is touched.
	aprilAgain, err := poster.PostDepreciation(ctx, headOffice, "2026", 4, "", "", "", now)
	if err != nil {
		t.Fatalf("PostDepreciation April again: %v", err)
	}
	if aprilAgain.DocNo != april.DocNo+"-2" || journalYear(aprilAgain.DocNo) != "2569" {
		t.Fatalf("April posted again as %s in %s, want %s-2 in 2569", aprilAgain.DocNo, journalYear(aprilAgain.DocNo), april.DocNo)
	}
	if journalStatus(april.DocNo) != "reversed" || journalStatus(aprilAgain.DocNo) != "posted" {
		t.Fatalf("statuses %s=%s %s=%s", april.DocNo, journalStatus(april.DocNo), aprilAgain.DocNo, journalStatus(aprilAgain.DocNo))
	}
	if n := count(`SELECT count(*) FROM fa_records WHERE company='01' AND kind='depreciations' AND payload->>'journaldocno'=$1 AND (payload->>'isposted')::boolean`, aprilAgain.DocNo); n != 1 {
		t.Fatalf("the April row must point at %s, got %d", aprilAgain.DocNo, n)
	}
	var aprilIsOneMonth, otherMonthsUntouched bool
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT SUM(debit)-SUM(credit) FROM gl_lines WHERE company='01' AND account_code='52100' AND doc_no IN ($1,$2,$3) AND entry_date BETWEEN '2026-04-01' AND '2026-04-30')
			= (SELECT SUM((payload->>'perioddeprec')::numeric) FROM fa_records WHERE company='01' AND kind='depreciations' AND payload->>'fiscalyear'='2026' AND (payload->>'period')::int=4),
		NOT EXISTS(SELECT 1 FROM gl_lines WHERE company='01' AND doc_no IN ($1,$2,$3) AND entry_date NOT BETWEEN '2026-04-01' AND '2026-04-30')`,
		april.DocNo, "REV-"+april.DocNo, aprilAgain.DocNo).Scan(&aprilIsOneMonth, &otherMonthsUntouched); err != nil || !aprilIsOneMonth || !otherMonthsUntouched {
		t.Fatalf("April must hold one month of depreciation (%v) and no other month a line (%v): %v", aprilIsOneMonth, otherMonthsUntouched, err)
	}
	// Posting a finished period again with its number, or a reversed number, is refused on the
	// number field; a generated number moves on instead.
	_, err = poster.PostDepreciation(ctx, headOffice, "2026", 5, "", april.DocNo, "", now)
	mustUserError(t, err, "fa_docno_reversed", "docno", april.DocNo)
	_, err = poster.PostDepreciation(ctx, headOffice, "2026", 5, "", aprilAgain.DocNo, "", now)
	mustUserError(t, err, "fa_docno_in_use", "docno", aprilAgain.DocNo)

	// A second posting of the same period while one is running is refused, not posted twice.
	lock, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var held bool
	if err := lock.QueryRowContext(ctx, `SELECT pg_try_advisory_lock(hashtext('fa-depreciation-post'), hashtext('01:2026:5'))`).Scan(&held); err != nil || !held {
		t.Fatalf("could not hold the May lock: %v %v", held, err)
	}
	before := journals()
	_, err = poster.PostDepreciation(ctx, headOffice, "2026", 5, "", "", "", now)
	mustUserError(t, err, "fa_period_post_in_progress", "", "5/2026")
	if n := journals(); n != before || unposted("2026", 5) != 1 {
		t.Fatalf("a refused concurrent posting must write nothing: journals %d → %d, May unposted %d", before, n, unposted("2026", 5))
	}
	if _, err := lock.ExecContext(ctx, `SELECT pg_advisory_unlock(hashtext('fa-depreciation-post'), hashtext('01:2026:5'))`); err != nil {
		t.Fatal(err)
	}
	_ = lock.Close()
	if _, err := poster.PostDepreciation(ctx, headOffice, "2026", 5, "", "", "", now); err != nil {
		t.Fatalf("May after the other posting finished: %v", err)
	}

	// Reversing the second April journal keeps the original's date again (the Thai-calendar
	// "today" is only for a date that can no longer change, see 6c).
	if err := poster.ReverseDepreciation(ctx, headOffice, aprilAgain.DocNo, "ผ่านรายการค่าเสื่อมราคาซ้ำ", time.Date(2026, 10, 1, 18, 30, 0, 0, time.UTC)); err != nil {
		t.Fatalf("ReverseDepreciation April again: %v", err)
	}
	if got := journalDate("REV-" + aprilAgain.DocNo); got != aprilAgain.Date {
		t.Fatalf("reversal of %s dated %q, want the original's date %s", aprilAgain.DocNo, got, aprilAgain.Date)
	}
	if n := unposted("2026", 4); n != 1 {
		t.Fatalf("April must be unposted after the second reversal, got %d", n)
	}

	// A reversal made in the GL screen leaves the row marked posted; reverse-gl then only frees
	// the row and writes no journal.
	third, err := poster.PostDepreciation(ctx, headOffice, "2026", 4, "", "", "", now)
	if err != nil || third.DocNo != april.DocNo+"-3" {
		t.Fatalf("third April posting: %v %+v", err, third)
	}
	var thirdID string
	var thirdVersion int64
	if err := db.QueryRowContext(ctx, `SELECT id, version FROM gl_records WHERE company='01' AND kind='journals' AND payload->>'docno'=$1`, third.DocNo).Scan(&thirdID, &thirdVersion); err != nil {
		t.Fatal(err)
	}
	run(gl.Command{Resource: "journals", Action: "reverse", ID: thirdID, Version: thirdVersion, DocNo: "REV-" + third.DocNo, Date: "2026-10-05", Reason: "กลับรายการจากหน้าสมุดรายวัน"})
	written := journals()
	if err := poster.ReverseDepreciation(ctx, headOffice, third.DocNo, "ปลดสถานะค่าเสื่อมราคาที่กลับรายการแล้ว", now); err != nil {
		t.Fatalf("ReverseDepreciation after a GL-screen reversal: %v", err)
	}
	if n := journals(); n != written {
		t.Fatalf("freeing the rows must not write a journal, %d → %d", written, n)
	}
	if n := unposted("2026", 4); n != 1 {
		t.Fatalf("April must be unposted after the GL-screen reversal, got %d", n)
	}

	// 6c. The original's year is closed: the reversal goes to today by the Thai calendar, and the
	// period cannot be posted again into the closed year, so nothing is counted twice.
	setYearClosed("2570", true)
	before = journals()
	err = poster.ReverseDepreciation(ctx, headOffice, april27.DocNo, "ผ่านรายการค่าเสื่อมราคาผิดงวด", time.Date(2028, 5, 2, 3, 0, 0, 0, time.UTC))
	mustUserError(t, err, "fa_fiscal_year_not_found", "", "2028-05-02")
	if n := journals(); n != before || journalStatus(april27.DocNo) != "posted" {
		t.Fatalf("a refused reversal must write nothing: journals %d → %d, %s %s", before, n, april27.DocNo, journalStatus(april27.DocNo))
	}
	next := gl.FiscalYear{Code: "2571", StartDate: "2028-04-01", EndDate: "2029-03-31", IsActive: true, Scale: 2, ProfitLossAccount: "32000", RetainedEarningsAccount: "31000"}
	run(gl.Command{Resource: "fiscal-years", Action: "create", FiscalYear: &next})
	if err := poster.ReverseDepreciation(ctx, headOffice, april27.DocNo, "ผ่านรายการค่าเสื่อมราคาผิดงวด", time.Date(2028, 5, 1, 18, 30, 0, 0, time.UTC)); err != nil {
		t.Fatalf("ReverseDepreciation of a closed-year journal: %v", err)
	}
	if got, year := journalDate("REV-"+april27.DocNo), journalYear("REV-"+april27.DocNo); got != "2028-05-02" || year != "2571" {
		t.Fatalf("reversal of the closed-year %s dated %s in %q, want the Thai date 2028-05-02 in 2571", april27.DocNo, got, year)
	}
	_, err = poster.PostDepreciation(ctx, headOffice, "2027", 4, "", "", "", now)
	mustUserError(t, err, "fa_fiscal_year_closed", "fiscalyear", "2570")
	setYearClosed("2570", false) // step 7 files a disposal into 2570

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
	if n := count(`SELECT count(*) FROM gl_records WHERE company='01' AND kind='journals' AND payload->>'fiscalyear' NOT IN ('2569','2570','2571')`); n != 0 {
		t.Fatalf("every journal must carry a พ.ศ. fiscal-year code, %d do not", n)
	}
}
