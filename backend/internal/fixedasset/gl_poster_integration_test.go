//go:build integration

package fixedasset_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"

	fa "smlcloudplatform/internal/fixedasset"
	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
)

// faCentralFixtureSQL is the slice of the central tenancy registry the header-branch check reads.
const faCentralFixtureSQL = `
CREATE TABLE holdings (code TEXT PRIMARY KEY, name TEXT NOT NULL, is_active BOOLEAN NOT NULL DEFAULT true);
CREATE TABLE companies (holding_code TEXT NOT NULL REFERENCES holdings(code), code TEXT NOT NULL, name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true, PRIMARY KEY (holding_code, code));
CREATE TABLE branches (holding_code TEXT NOT NULL, company_code TEXT NOT NULL, code TEXT NOT NULL, name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true, PRIMARY KEY (holding_code, company_code, code));
INSERT INTO holdings VALUES ('RUNGRUENG_FA', 'กลุ่มกิจการรุ่งเรืองกรุ๊ป', true);
INSERT INTO companies VALUES ('RUNGRUENG_FA', '01', 'บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด', true);
INSERT INTO branches VALUES ('RUNGRUENG_FA', '01', '00000', 'สำนักงานใหญ่', true),
    ('RUNGRUENG_FA', '01', '00001', 'สาขาลาดหลุมแก้ว', true);`

// faIntegrationDB opens a throwaway schema on BC_GL_TEST_POSTGRES_DSN (tools/verify.sh postgres
// target: postgres:18-alpine); the GL and fixed-asset tables are created by the code under test.
func faIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN to isolated PostgreSQL")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("bc_fa_it_%d", time.Now().UnixNano())
	if _, err := admin.Exec(`CREATE SCHEMA ` + pq.QuoteIdentifier(schema)); err != nil {
		admin.Close()
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(`DROP SCHEMA ` + pq.QuoteIdentifier(schema) + ` CASCADE`); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
		admin.Close()
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(faCentralFixtureSQL); err != nil {
		t.Fatalf("load central fixture: %v", err)
	}
	return db
}

func mustUserError(t *testing.T, err error, code, field string, contains ...string) {
	t.Helper()
	user, ok := gl.AsUserError(err)
	if !ok || user.Code != code || user.Field != field {
		t.Fatalf("want %s on %s, got %#v (%v)", code, field, user, err)
	}
	for _, want := range contains {
		if !strings.Contains(user.Message, want) {
			t.Fatalf("message %q lacks %q", user.Message, want)
		}
	}
}

// TestFixedAssetGLPostingOnRealLedger posts depreciation and disposals through the real GL engine
// and checks PostgreSQL after every step: accounts come only from the asset / asset type /
// disposal input, the voucher number comes from the user-defined general book, the header branch
// follows the GL rule, and amounts stay exact decimals (bug 2026-09-24).
func TestFixedAssetGLPostingOnRealLedger(t *testing.T) {
	db := faIntegrationDB(t)
	ctx := context.Background()
	connect := func(string) (*sql.DB, error) { return db, nil }
	ledger := gl.NewPostgresStore(gl.NewPostgres(connect))
	poster := fa.NewGLPoster(connect, ledger, func(ctx context.Context, s fa.Scope, branch string) error {
		return glhttp.CheckJournalBranch(ctx, connect, gl.Scope{Holding: s.Holding, Company: s.Company, Branch: s.Branch, Actor: s.Actor}, branch)
	})
	store := fa.NewStore(connect)
	companyWide := fa.Scope{Holding: "RUNGRUENG_FA", Company: "01", Actor: "fa-it"}
	headOffice := companyWide
	headOffice.Branch = "00000"
	glScope := gl.Scope{Holding: "RUNGRUENG_FA", Company: "01", Branch: "00000", Actor: "fa-it"}
	now := time.Date(2026, 9, 24, 3, 0, 0, 0, time.UTC)
	nonce := time.Now().UnixNano()
	setup := 0
	run := func(cmd gl.Command) {
		t.Helper()
		setup++
		cmd.RequestID = fmt.Sprintf("fa-it-setup-%03d-%d", setup, nonce)
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

	// 1. Multi-level Thai chart of accounts, the first fiscal year (which creates the standard
	// books) and a second general book "GJ" that sorts before "JV".
	for _, a := range []struct {
		code, parent, name, kind, normal string
		level                            int
		posting                          bool
	}{
		{"10000", "", "สินทรัพย์", "asset", "debit", 1, false},
		{"11000", "10000", "สินทรัพย์หมุนเวียน", "asset", "debit", 2, false},
		{"11100", "11000", "เงินสดและรายการเทียบเท่าเงินสด", "asset", "debit", 3, false},
		{"11110", "11100", "เงินสดในมือ", "asset", "debit", 4, true},
		{"12000", "10000", "สินทรัพย์ไม่หมุนเวียน", "asset", "debit", 2, false},
		{"12100", "12000", "ที่ดิน อาคารและอุปกรณ์", "asset", "debit", 3, false},
		{"12110", "12100", "อาคารและอุปกรณ์", "asset", "debit", 4, true},
		{"12120", "12100", "ค่าเสื่อมราคาสะสม-อาคารและอุปกรณ์", "asset", "credit", 4, true},
		{"12130", "12100", "ยานพาหนะ", "asset", "debit", 4, true},
		{"12140", "12100", "ค่าเสื่อมราคาสะสม-ยานพาหนะ", "asset", "credit", 4, true},
		{"20000", "", "หนี้สิน", "liability", "credit", 1, false},
		{"21000", "20000", "หนี้สินหมุนเวียน", "liability", "credit", 2, false},
		{"21100", "21000", "ภาษีค้างจ่าย", "liability", "credit", 3, false},
		{"21110", "21100", "ภาษีขาย", "liability", "credit", 4, true},
		{"30000", "", "ส่วนของเจ้าของ", "equity", "credit", 1, false},
		{"31000", "30000", "กำไรสะสม", "equity", "credit", 2, true},
		{"32000", "30000", "กำไร(ขาดทุน)สุทธิประจำงวด", "equity", "credit", 2, true},
		{"40000", "", "รายได้", "income", "credit", 1, false},
		{"42000", "40000", "รายได้อื่น", "income", "credit", 2, false},
		{"42100", "42000", "กำไร(ขาดทุน)จากการจำหน่ายสินทรัพย์", "income", "credit", 3, true},
		{"50000", "", "ค่าใช้จ่าย", "expense", "debit", 1, false},
		{"52000", "50000", "ค่าใช้จ่ายในการขายและบริหาร", "expense", "debit", 2, false},
		{"52100", "52000", "ค่าเสื่อมราคา-อาคารและอุปกรณ์", "expense", "debit", 3, true},
		{"52200", "52000", "ค่าเสื่อมราคา-ยานพาหนะ", "expense", "debit", 3, true},
	} {
		run(gl.Command{Resource: "accounts", Action: "create", Account: &gl.Account{
			AccountCode: a.code, ParentAccountCode: a.parent, Names: []gl.Name{{Code: "th", Name: a.name}},
			AccountType: a.kind, NormalBalance: a.normal, Level: a.level, AllowPosting: a.posting, IsActive: true,
			IsCash: a.code == "11110",
		}})
	}
	run(gl.Command{Resource: "fiscal-years", Action: "create", FiscalYear: &gl.FiscalYear{
		Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", Scale: 2, IsActive: true,
		ProfitLossAccount: "32000", RetainedEarningsAccount: "31000",
	}})
	run(gl.Command{Resource: "journal-books", Action: "create", Master: &gl.Master{
		Kind: "journal-books", Code: "GJ", Name: "สมุดรายวันทั่วไป (สินทรัพย์ถาวร)", BookType: gl.BookTypeGeneral, IsActive: true,
	}})

	// 2. Asset register: the packing machine inherits every account from its type; the delivery
	// truck belongs to a type without accounts and has none of its own yet.
	if _, err := store.CreateAssetType(ctx, headOffice, fa.AssetType{
		TypeCode: "MACHINERY", Names: []fa.Name{{Code: "th", Name: "เครื่องจักรและอุปกรณ์"}},
		AssetAccountCode: "12110", AccumDeprecAccountCode: "12120", DeprecExpenseAccountCode: "52100",
	}, now); err != nil {
		t.Fatal(err)
	}
	machine := fa.Asset{
		AssetCode: "เครื่องบรรจุปูนซีเมนต์-01", Names: []fa.Name{{Code: "th", Name: "เครื่องบรรจุปูนซีเมนต์อัตโนมัติ"}},
		AssetTypeCode: "MACHINERY", Cost: "240000.00", ScrapValue: "1.00", UsefulLifeYears: 5, DeprecPercent: "20.00",
		PurchaseDate: "2026-01-01", StartCalcDate: "2026-01-01", Status: "active",
	}
	truck := fa.Asset{
		AssetCode: "FA-2569-002", Names: []fa.Name{{Code: "th", Name: "รถกระบะส่งวัสดุก่อสร้าง"}},
		AssetTypeCode: "VEHICLE", BranchCode: "00001", Cost: "650000.00", ScrapValue: "1.00", UsefulLifeYears: 5,
		DeprecPercent: "20.00", PurchaseDate: "2026-01-01", StartCalcDate: "2026-01-01", Status: "active",
	}
	if _, err := store.CreateAsset(ctx, headOffice, machine, now); err != nil {
		t.Fatal(err)
	}
	createdTruck, err := store.CreateAsset(ctx, headOffice, truck, now)
	if err != nil {
		t.Fatal(err)
	}
	unpostedJan := `SELECT count(*) FROM fa_records WHERE company='01' AND kind='depreciations' AND payload->>'fiscalyear'='2026' AND (payload->>'period')::int=1 AND NOT COALESCE((payload->>'isposted')::boolean,false)`
	if n := count(unpostedJan); n != 2 {
		t.Fatalf("expected 2 unposted January depreciation rows, got %d", n)
	}

	// 3. Depreciation from a company-wide session: the header branch is required and must exist.
	_, err = poster.PostDepreciation(ctx, companyWide, "2026", 1, "2026-01-31", "", "", now)
	mustUserError(t, err, "journal_branch_required", "branchcode")
	_, err = poster.PostDepreciation(ctx, companyWide, "2026", 1, "2026-01-31", "", "99999", now)
	mustUserError(t, err, "journal_branch_not_found", "branchcode", "99999")
	// The truck has no expense account anywhere: ask, never fall back to a guessed code.
	_, err = poster.PostDepreciation(ctx, companyWide, "2026", 1, "2026-01-31", "", "00000", now)
	mustUserError(t, err, "fa_account_required", "deprecexpenseaccountcode", "FA-2569-002", "ประเภทสินทรัพย์")
	if n := journals(); n != 0 {
		t.Fatalf("a refused posting must not write a journal, found %d", n)
	}
	if n := count(unpostedJan); n != 2 {
		t.Fatalf("a refused posting must leave January unposted, got %d", n)
	}

	truck.AssetAccountCode, truck.AccumDeprecAccountCode, truck.DeprecExpenseAccountCode = "12130", "12140", "52200"
	if _, err := store.UpdateAsset(ctx, headOffice, createdTruck.ID, truck, createdTruck.Version, now); err != nil {
		t.Fatal(err)
	}
	journal, err := poster.PostDepreciation(ctx, companyWide, "2026", 1, "2026-01-31", "", "00000", now)
	if err != nil {
		t.Fatalf("PostDepreciation: %v", err)
	}
	if journal.DocNo != "GJ-FA-2026-01" || journal.BookCode != "GJ" || journal.BranchCode != "00000" {
		t.Fatalf("journal header = docno %q book %q branch %q", journal.DocNo, journal.BookCode, journal.BranchCode)
	}
	var lines, bookOK, branchOK int
	var balanced, matchesSchedule bool
	if err := db.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE book_code='GJ'), count(*) FILTER (WHERE branch_code='00000'),
		SUM(debit)=SUM(credit),
		SUM(debit)=(SELECT SUM((payload->>'perioddeprec')::numeric) FROM fa_records WHERE company='01' AND kind='depreciations' AND payload->>'fiscalyear'='2026' AND (payload->>'period')::int=1)
		FROM gl_lines WHERE company='01' AND doc_no='GJ-FA-2026-01'`).Scan(&lines, &bookOK, &branchOK, &balanced, &matchesSchedule); err != nil {
		t.Fatal(err)
	}
	if lines != 4 || bookOK != 4 || branchOK != 4 || !balanced || !matchesSchedule {
		t.Fatalf("depreciation lines=%d book=%d branch=%d balanced=%v equals schedule=%v", lines, bookOK, branchOK, balanced, matchesSchedule)
	}
	var order string
	if err := db.QueryRowContext(ctx, `SELECT string_agg(account_code, ',' ORDER BY line_no) FROM gl_lines WHERE company='01' AND doc_no='GJ-FA-2026-01'`).Scan(&order); err != nil {
		t.Fatal(err)
	}
	if order != "52100,12120,52200,12140" {
		t.Fatalf("lines must use the type account for the machine, the truck's own accounts, in a stable order; got %s", order)
	}
	if n := count(`SELECT count(*) FROM fa_records WHERE company='01' AND kind='depreciations' AND payload->>'journaldocno'='GJ-FA-2026-01' AND (payload->>'isposted')::boolean`); n != 2 {
		t.Fatalf("both January rows must point at the journal, got %d", n)
	}

	// 4. Disposals from the head-office session.
	_, _, err = poster.DisposeAsset(ctx, headOffice, fa.AssetDisposal{AssetCode: truck.AssetCode, DisposalDate: "2026-06-30", DisposalType: "write_off", GainLossAccountCode: "42100"}, now)
	mustUserError(t, err, "journal_branch_outside_session", "branchcode", "FA-2569-002", "00001", "00000")

	sale := fa.AssetDisposal{
		AssetCode: machine.AssetCode, DisposalDate: "2026-06-30", DisposalType: "sale",
		SalePrice: "200000.00", VatAmount: "14000.00", SettlementAccountCode: "11110", GainLossAccountCode: "42100",
		Reason: "ขายเครื่องบรรจุปูนซีเมนต์เก่าให้ผู้รับซื้อ",
	}
	_, _, err = poster.DisposeAsset(ctx, headOffice, sale, now)
	mustUserError(t, err, "fa_account_required", "vataccountcode", "ภาษีขาย", "14000.00 บาท")
	sale.VatAccountCode = "21110"
	// "GJ-DISP-" + a 25-rune Thai asset code = 33 runes > doc_no VARCHAR(30).
	_, _, err = poster.DisposeAsset(ctx, headOffice, sale, now)
	mustUserError(t, err, "code_too_long", "journaldocno", "GJ-DISP-เครื่องบรรจุปูนซีเมนต์-01", "33", "ระบุเลขที่ใบสำคัญเอง")
	if n := count(`SELECT count(*) FROM gl_records WHERE company='01' AND kind='journals' AND code LIKE '%DISP%'`); n != 0 {
		t.Fatalf("refused disposals must not write journals, found %d", n)
	}
	if n := count(`SELECT count(*) FROM fa_records WHERE company='01' AND kind='assets' AND code=$1 AND payload->>'status'='active'`, machine.AssetCode); n != 1 {
		t.Fatal("refused disposals must leave the machine active")
	}

	sale.JournalDocNo = "GJ-DISP-2569-0001"
	disposal, disposalJournal, err := poster.DisposeAsset(ctx, headOffice, sale, now)
	if err != nil {
		t.Fatalf("DisposeAsset: %v", err)
	}
	if disposalJournal.DocNo != "GJ-DISP-2569-0001" || disposalJournal.BranchCode != "00000" || disposal.JournalDocNo != "GJ-DISP-2569-0001" {
		t.Fatalf("sale journal = %+v / disposal %+v", disposalJournal, disposal)
	}
	var saleLines string
	var saleHeaderOK bool
	if err := db.QueryRowContext(ctx, `SELECT string_agg(account_code || ':' || debit::numeric(18,2) || ':' || credit::numeric(18,2), ',' ORDER BY line_no),
		bool_and(branch_code='00000' AND book_code='GJ'), SUM(debit)=SUM(credit)
		FROM gl_lines WHERE company='01' AND doc_no='GJ-DISP-2569-0001'`).Scan(&saleLines, &saleHeaderOK, &balanced); err != nil {
		t.Fatal(err)
	}
	accum := disposal.AccumDeprecAtDisposal.Decimal().StringFixed(2)
	loss := disposal.GainLoss.Decimal().Abs().StringFixed(2)
	want := "11110:214000.00:0.00,12120:" + accum + ":0.00,42100:" + loss + ":0.00,12110:0.00:240000.00,21110:0.00:14000.00"
	if saleLines != want || !saleHeaderOK || !balanced {
		t.Fatalf("sale lines = %s (branch/book ok=%v balanced=%v), want %s", saleLines, saleHeaderOK, balanced, want)
	}
	if n := count(`SELECT count(*) FROM fa_records WHERE company='01' AND kind='assets' AND code=$1 AND payload->>'status'='disposed'`, machine.AssetCode); n != 1 {
		t.Fatal("the sold machine must be disposed")
	}

	// A write-off receives nothing, so no settlement or VAT account is needed; the generated
	// number starts with the chosen general book and the voucher lands in the asset's branch.
	_, writeOff, err := poster.DisposeAsset(ctx, companyWide, fa.AssetDisposal{AssetCode: truck.AssetCode, DisposalDate: "2026-06-30", DisposalType: "write_off", GainLossAccountCode: "42100", Reason: "รถกระบะเสียหายจากอุบัติเหตุ ซ่อมไม่คุ้ม"}, now)
	if err != nil {
		t.Fatalf("write-off: %v", err)
	}
	if writeOff.DocNo != "GJ-DISP-FA-2569-002" || writeOff.BranchCode != "00001" {
		t.Fatalf("write-off journal = docno %q branch %q", writeOff.DocNo, writeOff.BranchCode)
	}
	if n := count(`SELECT count(*) FROM gl_lines WHERE company='01' AND doc_no='GJ-DISP-FA-2569-002' AND branch_code='00001' AND book_code='GJ' AND account_code<>'11110'`); n != 3 {
		t.Fatalf("write-off must post accum, loss and cost lines in branch 00001 without a settlement line, got %d", n)
	}
	if n := count(`SELECT count(*) FROM gl_records WHERE company='01' AND kind='journals' AND payload->>'status'='posted'`); n != 3 {
		t.Fatalf("expected 3 posted journals (depreciation, sale, write-off), got %d", n)
	}
}
