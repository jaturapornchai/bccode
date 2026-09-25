package fixedasset

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"

	gl "smlcloudplatform/internal/generalledger"
)

// fakeLedger stands in for *generalledger.Store in tests. It reproduces just
// the two guarantees GLPoster relies on from the real Store.Execute:
//   - a "create"/"post" call replayed with the exact same RequestID and
//     payload is idempotent (returns the cached result, never creates a
//     second journal);
//   - a "create"/"post" call reusing a RequestID with a *different* payload
//     is rejected, never silently applied.
//
// This lets the P0 fix (posting through generalledger.Store.Execute instead
// of writing gl_journals/gl_lines directly) be tested without a live PostgreSQL
// ledger database.
type fakeLedger struct {
	hashes   map[string]string
	results  map[string]gl.Result
	journals map[string]*gl.Journal
	created  int
	posted   int
	reversed int
	failErr  error
}

// fakeFiscalYears are coded in พ.ศ. like a real Thai company, so a journal that still carries the
// ค.ศ. schedule year is refused here as it is by the real GL (bug 2026-09-25).
var fakeFiscalYears = []gl.FiscalYear{
	{Code: "2569", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true},
	{Code: "2570", StartDate: "2027-01-01", EndDate: "2027-12-31", IsActive: true},
}

func newFakeLedger() *fakeLedger {
	return &fakeLedger{
		hashes:   map[string]string{},
		results:  map[string]gl.Result{},
		journals: map[string]*gl.Journal{},
	}
}

func (f *fakeLedger) Execute(_ context.Context, _ gl.Scope, cmd gl.Command) (gl.Result, error) {
	if f.failErr != nil {
		return gl.Result{}, f.failErr
	}
	raw, err := json.Marshal(cmd)
	if err != nil {
		return gl.Result{}, err
	}
	sum := sha256.Sum256(raw)
	hash := fmt.Sprintf("%x", sum)
	if prev, ok := f.hashes[cmd.RequestID]; ok {
		if prev != hash {
			return gl.Result{}, fmt.Errorf("รหัสคำขอนี้ถูกใช้กับข้อมูลอื่นแล้ว")
		}
		return f.results[cmd.RequestID], nil
	}
	f.hashes[cmd.RequestID] = hash

	switch cmd.Action {
	case "create":
		if !fakeYearHolds(cmd.Journal.FiscalYear, cmd.Journal.Date) {
			return gl.Result{}, fmt.Errorf("วันที่อยู่นอกปีบัญชีที่เปิดใช้งาน")
		}
		f.created++
		id := fmt.Sprintf("fake-journal-%d", f.created)
		j := *cmd.Journal
		j.ID = id
		j.Version = 1
		j.Status = "draft"
		f.journals[id] = &j
		res := gl.Result{ID: id, Version: 1, Sequence: int64(f.created + f.posted)}
		f.results[cmd.RequestID] = res
		return res, nil
	case "post":
		j, ok := f.journals[cmd.ID]
		if !ok || j.Version != cmd.Version {
			return gl.Result{}, fmt.Errorf("รายการนี้เปลี่ยนไปแล้ว กรุณาโหลดข้อมูลล่าสุด")
		}
		f.posted++
		j.Status = "posted"
		j.Version++
		res := gl.Result{ID: j.ID, Version: j.Version, Sequence: int64(f.created + f.posted)}
		f.results[cmd.RequestID] = res
		return res, nil
	case "reverse":
		f.reversed++
		res := gl.Result{ID: "fake-reversal-" + cmd.DocNo, Version: 1, Sequence: int64(f.created + f.posted + f.reversed)}
		f.results[cmd.RequestID] = res
		return res, nil
	}
	return gl.Result{}, fmt.Errorf("unsupported action in fake ledger: %s", cmd.Action)
}

// fakeYearHolds mirrors generalledger Journal.Validate: the code must be the year that holds the date.
func fakeYearHolds(code, date string) bool {
	for _, y := range fakeFiscalYears {
		if y.Code == code && y.StartDate <= date && date <= y.EndDate {
			return true
		}
	}
	return false
}

func (f *fakeLedger) List(_ context.Context, _ gl.Scope, resource string, query string, _ int, _ int, _ gl.ListFilter) (gl.Page, error) {
	page := gl.Page{Items: []json.RawMessage{}}
	if resource == "fiscal-years" {
		for _, y := range fakeFiscalYears {
			raw, _ := json.Marshal(y)
			page.Items = append(page.Items, raw)
		}
		return page, nil
	}
	if resource == "journal-books" {
		// The company's general book is deliberately not named "JV": the poster must pick by booktype.
		for _, b := range []gl.Master{
			{Kind: "journal-books", Code: "AA-SALE", Name: "สมุดรายวันขาย", BookType: gl.BookTypeSales, IsActive: true},
			{Kind: "journal-books", Code: "GJ-OLD", Name: "สมุดรายวันทั่วไป (เลิกใช้)", BookType: gl.BookTypeGeneral, IsActive: false},
			{Kind: "journal-books", Code: "GJ", Name: "สมุดรายวันทั่วไป", BookType: gl.BookTypeGeneral, IsActive: true},
		} {
			raw, _ := json.Marshal(b)
			page.Items = append(page.Items, raw)
		}
		return page, nil
	}
	for _, j := range f.journals {
		if j.DocNo == query {
			raw, _ := json.Marshal(j)
			page.Items = append(page.Items, raw)
		}
	}
	return page, nil
}

// allowBranch accepts every voucher branch; the header-branch rule itself is covered by
// generalledger/httpapi TestCheckJournalBranch and gl_poster_integration_test.go.
func allowBranch(context.Context, Scope, string) error { return nil }

// testConnector opens an isolated PostgreSQL database for the fixed-asset store;
// every row it writes is removed by company code when the test ends.
func testConnector(t *testing.T) (Connector, string) {
	t.Helper()
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN to isolated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	company := fmt.Sprintf("FA_TEST_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM fa_records WHERE company = $1`, company)
		_ = db.Close()
	})
	return func(string) (*sql.DB, error) { return db, nil }, company
}

func TestFixedAssets_CPACycle(t *testing.T) {
	connect, company := testConnector(t)
	ctx := context.Background()
	scope := Scope{Holding: "TEST_HOLDING", Company: company, Branch: "HQ", Actor: "CPA_AUDITOR"}
	now := time.Now().UTC()

	asset := Asset{
		AssetCode:                "EQ-2026-001",
		Names:                    []Name{{Code: "th", Name: "เครื่องจักรผลิตสินค้า A"}},
		AssetTypeCode:            "MACHINERY",
		Cost:                     Amount("600000.00"),
		ScrapValue:               Amount("1.00"),
		UsefulLifeYears:          5,
		DeprecPercent:            Amount("20.00"),
		PurchaseDate:             "2026-01-01",
		StartCalcDate:            "2026-01-01",
		FirstYearPercent:         Amount("0.00"),
		AssetAccountCode:         "120101",
		AccumDeprecAccountCode:   "129101",
		DeprecExpenseAccountCode: "520103",
		Status:                   "active",
	}

	store := NewStore(connect)
	created, err := store.CreateAsset(ctx, scope, asset, now)
	if err != nil {
		t.Fatalf("CreateAsset failed: %v", err)
	}
	if _, err := store.CreateAsset(ctx, scope, asset, now); err != ErrCodeDuplicate {
		t.Fatalf("expected duplicate asset code to be rejected, got %v", err)
	}
	got, err := store.GetAsset(ctx, scope, created.AssetCode)
	if err != nil || got.ID != created.ID || !got.Cost.Decimal().Equal(decimal.RequireFromString("600000")) {
		t.Fatalf("GetAsset by code = %+v, %v", got, err)
	}
	items, total, err := store.ListAssets(ctx, scope, "เครื่องจักร", "", "", 1, 10)
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("ListAssets search = %d/%d, %v", len(items), total, err)
	}

	sched, err := store.GetAssetDepreciationSchedule(ctx, scope, asset.AssetCode)
	if err != nil {
		t.Fatalf("GetAssetDepreciationSchedule failed: %v", err)
	}
	if len(sched) < 60 {
		t.Fatalf("expected at least 60 months of depreciation, got %d", len(sched))
	}
	totDep := decimal.Zero
	for _, item := range sched {
		totDep = totDep.Add(item.PeriodDeprec.Decimal())
	}
	if expected := asset.Cost.Decimal().Sub(asset.ScrapValue.Decimal()); !totDep.Equal(expected) {
		t.Fatalf("expected total depreciation %s, got %s", expected, totDep)
	}

	ledger := newFakeLedger()
	poster := NewGLPoster(connect, ledger, allowBranch)
	journal, err := poster.PostDepreciation(ctx, scope, "2026", 1, "2026-01-31", "", "", now)
	if err != nil {
		t.Fatalf("PostDepreciation failed: %v", err)
	}
	dr, cr := decimal.Zero, decimal.Zero
	for _, l := range journal.Lines {
		dr = dr.Add(l.Debit)
		cr = cr.Add(l.Credit)
	}
	if !dr.Equal(cr) || dr.IsZero() {
		t.Fatalf("GL journal out of balance: Dr %s Cr %s", dr, cr)
	}
	if journal.FiscalYear != "2569" {
		t.Fatalf("the journal must carry the พ.ศ. fiscal-year code of its date, got %q", journal.FiscalYear)
	}
	if journal.BookCode != "GJ" || journal.DocNo != "GJ-FA-2026-01" || journal.BranchCode != "HQ" {
		t.Fatalf("depreciation must go to the active general book chosen by type, numbered from it, in the session branch; got book %q docno %q branch %q", journal.BookCode, journal.DocNo, journal.BranchCode)
	}
	if ledger.created != 1 || ledger.posted != 1 {
		t.Fatalf("expected exactly 1 create + 1 post, got created=%d posted=%d", ledger.created, ledger.posted)
	}
	if _, err := poster.PostDepreciation(ctx, scope, "2026", 1, "2026-01-31", "", "", now); err == nil {
		t.Fatalf("posting the same period twice must fail: no unposted rows remain")
	}
	if err := store.DeleteAsset(ctx, scope, created.ID, 0, now); err == nil {
		t.Fatalf("an asset with posted depreciation must not be deletable")
	}

	failing := newFakeLedger()
	failing.failErr = fmt.Errorf("ledger unavailable")
	if _, err := NewGLPoster(connect, failing, allowBranch).PostDepreciation(ctx, scope, "2026", 2, "2026-02-28", "", "", now); err == nil {
		t.Fatalf("expected PostDepreciation to fail when the GL engine fails")
	}
	db, _ := connect("")
	stillOpen, err := queryRecords[DepreciationScheduleItem](ctx, db, scope.Company, kindDepreciation, ` AND payload->>'fiscalyear' = '2026' AND (payload->>'period')::int = 2 AND NOT COALESCE((payload->>'isposted')::boolean, false)`, "")
	if err != nil || len(stillOpen) != 1 {
		t.Fatalf("failed GL post must leave period 2 unposted, got %d rows, %v", len(stillOpen), err)
	}

	disposal := AssetDisposal{
		AssetCode:             asset.AssetCode,
		DisposalDate:          "2026-06-30",
		DisposalType:          "sale",
		SalePrice:             Amount("550000.00"),
		VatAmount:             Amount("38500.00"),
		SettlementAccountCode: "110101",
		GainLossAccountCode:   "420101",
		VatAccountCode:        "210301",
		Reason:                "ขายเครื่องจักรเก่า",
	}
	disposer := NewGLPoster(connect, newFakeLedger(), allowBranch)
	resDisp, resJourn, err := disposer.DisposeAsset(ctx, scope, disposal, now)
	if err != nil {
		t.Fatalf("DisposeAsset failed: %v", err)
	}
	dispDr, dispCr := decimal.Zero, decimal.Zero
	for _, l := range resJourn.Lines {
		dispDr = dispDr.Add(l.Debit)
		dispCr = dispCr.Add(l.Credit)
	}
	if !dispDr.Equal(dispCr) {
		t.Fatalf("disposal journal out of balance: Dr %s Cr %s", dispDr, dispCr)
	}
	if resDisp.GainLoss.Decimal().IsZero() {
		t.Fatalf("expected a non-zero gain/loss, got %s", resDisp.GainLoss)
	}
	disposed, err := store.GetAsset(ctx, scope, created.ID)
	if err != nil || disposed.Status != "disposed" {
		t.Fatalf("asset status after disposal = %+v, %v", disposed, err)
	}
	if _, _, err := disposer.DisposeAsset(ctx, scope, disposal, now); err == nil {
		t.Fatalf("a disposed asset must not be disposed twice")
	}

	if _, err := NewReporter(connect).GetAssetScheduleReport(ctx, scope, "2026", 12, ""); err != nil {
		t.Fatalf("schedule report failed: %v", err)
	}
	if _, err := NewReporter(connect).GetTaxReconciliationReport(ctx, scope, "2026"); err != nil {
		t.Fatalf("tax reconciliation report failed: %v", err)
	}
}

// TestGLPoster_PostJournalIdempotent exercises postJournal (the low-level
// helper both PostDepreciation and DisposeAsset use to talk to the GL engine)
// directly, without a database. It proves the deterministic
// RequestID + Store.Execute contract: replaying the exact same journal input
// (e.g. a network retry between "create" and "post" succeeding but
// the caller not observing it) must not create a second journal.
func TestGLPoster_PostJournalIdempotent(t *testing.T) {
	ledger := newFakeLedger()
	poster := NewGLPoster(nil, ledger, nil)
	scope := Scope{Holding: "TEST_HOLDING", Company: "TEST_CO", Branch: "HQ", Actor: "CPA_AUDITOR"}

	journalInput := &gl.Journal{
		DocNo:       "JV-FA-2026-03",
		Date:        "2026-03-31",
		BookCode:    "JV",
		FiscalYear:  "2569",
		Description: "บันทึกค่าเสื่อมราคาสินทรัพย์ประจำงวด 3/2026",
		Reference:   "FA-2026-03",
		BranchCode:  scope.Branch,
		Kind:        "manual",
		Lines: []gl.Line{
			{AccountCode: "520103", Description: "ค่าเสื่อมราคาประจำงวด 3/2026", Debit: gl.Amount("15000"), Credit: gl.Amount("0")},
			{AccountCode: "129101", Description: "ค่าเสื่อมราคาสะสมประจำงวด 3/2026", Debit: gl.Amount("0"), Credit: gl.Amount("15000")},
		},
	}

	first, firstPosted, err := poster.postJournal(context.Background(), scope, journalInput)
	if err != nil {
		t.Fatalf("first postJournal failed: %v", err)
	}
	second, secondPosted, err := poster.postJournal(context.Background(), scope, journalInput)
	if err != nil {
		t.Fatalf("second (retry) postJournal failed: %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf("retry returned a different journal id: got %s want %s", second.ID, first.ID)
	}
	if secondPosted.Version != firstPosted.Version {
		t.Fatalf("retry produced a different posted version: got %d want %d", secondPosted.Version, firstPosted.Version)
	}
	if ledger.created != 1 || ledger.posted != 1 {
		t.Fatalf("retrying the same docno must not create/post a duplicate journal, got created=%d posted=%d", ledger.created, ledger.posted)
	}

	// A retry that changes the journal content under the same docno must be
	// rejected, never silently merged into the existing journal.
	tampered := *journalInput
	tampered.Description = "แก้ไขคำอธิบายโดยไม่ได้ตั้งใจ"
	if _, _, err := poster.postJournal(context.Background(), scope, &tampered); err == nil {
		t.Fatalf("expected an error when retrying the same docno with different content, got nil")
	}
}

// Book codes are user-defined: the poster picks the active general (booktype 1) book with the
// lowest code and explains in Thai when the company has none (2026-09-24).
func TestGLPosterChoosesGeneralBookByType(t *testing.T) {
	scope := Scope{Holding: "H", Company: "C", Branch: "B", Actor: "tester"}
	code, err := NewGLPoster(nil, newFakeLedger(), nil).generalBookCode(context.Background(), scope)
	if err != nil || code != "GJ" {
		t.Fatalf("general book = %q, %v; want GJ (active type 1, not the inactive GJ-OLD or the sales book)", code, err)
	}
	empty := &noBooksLedger{fakeLedger: newFakeLedger()}
	_, err = NewGLPoster(nil, empty, nil).generalBookCode(context.Background(), scope)
	user, ok := gl.AsUserError(err)
	if !ok || user.Code != "journal_book_general_missing" || !strings.Contains(user.Message, "กำหนดสมุดรายวัน") {
		t.Fatalf("missing general book error = %v", err)
	}
}

type noBooksLedger struct{ *fakeLedger }

func (l *noBooksLedger) List(ctx context.Context, scope gl.Scope, resource, query string, page, limit int, filter gl.ListFilter) (gl.Page, error) {
	if resource == "journal-books" {
		return gl.Page{Items: []json.RawMessage{}}, nil
	}
	return l.fakeLedger.List(ctx, scope, resource, query, page, limit, filter)
}
