package fixedasset

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	gl "smlcloudplatform/internal/generalledger"
)

func toBsonD(v interface{}) bson.D {
	data, _ := bson.Marshal(v)
	var doc bson.D
	_ = bson.Unmarshal(data, &doc)
	return doc
}

// fakeLedger stands in for *generalledger.Store in tests. It reproduces just
// the two guarantees GLPoster relies on from the real Store.Execute:
//   - a "create"/"post" call replayed with the exact same RequestID and
//     payload is idempotent (returns the cached result, never creates a
//     second journal);
//   - a "create"/"post" call reusing a RequestID with a *different* payload
//     is rejected, never silently applied.
//
// This lets the P0 fix (posting through generalledger.Store.Execute instead
// of writing gl_journals/gl_lines directly) be tested without a live MongoDB
// replica set + Kafka broker, which Store.Execute itself requires.
type fakeLedger struct {
	hashes   map[string]string
	results  map[string]gl.Result
	journals map[string]*gl.Journal
	created  int
	posted   int
	reversed int
	failErr  error
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

func TestFixedAssets_CPACycle(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	scope := Scope{
		Holding: "TEST_HOLDING",
		Company: "TEST_CO",
		Branch:  "HQ",
		Actor:   "CPA_AUDITOR",
	}
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

	mt.Run("1_create_asset", func(mt *mtest.T) {
		db := mt.DB
		store := NewStore(db)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, db.Name()+".fixed_assets", mtest.FirstBatch, bson.D{{Key: "n", Value: 0}}),
			mtest.CreateCursorResponse(0, db.Name()+".fixed_assets", mtest.NextBatch),
			mtest.CreateSuccessResponse(), // InsertOne fixed_assets
			mtest.CreateSuccessResponse(), // InsertMany asset_depreciations
		)

		created, err := store.CreateAsset(context.Background(), scope, asset, now)
		if err != nil {
			mt.Fatalf("CreateAsset failed: %v", err)
		}
		if created.AssetCode != "EQ-2026-001" {
			mt.Errorf("expected asset code EQ-2026-001, got %s", created.AssetCode)
		}
	})

	mt.Run("2_calculate_schedule", func(mt *mtest.T) {
		calc := NewCalculator()
		sched, err := calc.CalculateSchedule(asset, "2031-12-31")
		if err != nil {
			mt.Fatalf("CalculateSchedule failed: %v", err)
		}
		if len(sched) < 60 {
			mt.Errorf("expected at least 60 months of depreciation, got %d", len(sched))
		}

		totDep := decimal.Zero
		for _, item := range sched {
			totDep = totDep.Add(item.PeriodDeprec.Decimal())
		}
		expectedDep := asset.Cost.Decimal().Sub(asset.ScrapValue.Decimal())
		if !totDep.Equal(expectedDep) {
			mt.Errorf("expected total deprecation %s, got %s", expectedDep, totDep)
		}
	})

	mt.Run("3_post_depreciation_to_gl", func(mt *mtest.T) {
		db := mt.DB
		ledger := newFakeLedger()
		poster := NewGLPoster(db, ledger)

		mockDeprecItem := DepreciationScheduleItem{
			AssetCode:    asset.AssetCode,
			FiscalYear:   "2026",
			Period:       1,
			PeriodDeprec: Amount("10000.00"),
			Days:         31,
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, db.Name()+".asset_depreciations", mtest.FirstBatch, toBsonD(mockDeprecItem)),
			mtest.CreateCursorResponse(0, db.Name()+".asset_depreciations", mtest.NextBatch),
			mtest.CreateCursorResponse(1, db.Name()+".fixed_assets", mtest.FirstBatch, toBsonD(asset)),
			mtest.CreateCursorResponse(0, db.Name()+".fixed_assets", mtest.NextBatch),
			mtest.CreateSuccessResponse(), // UpdateMany asset_depreciations
		)

		journal, err := poster.PostDepreciation(context.Background(), scope, "2026", 1, "2026-01-31", "JV-FA-2026-01", now)
		if err != nil {
			mt.Fatalf("PostDepreciation failed: %v", err)
		}
		if journal.DocNo != "JV-FA-2026-01" {
			mt.Errorf("expected docno JV-FA-2026-01, got %s", journal.DocNo)
		}
		if journal.Status != "posted" {
			mt.Errorf("expected journal status posted, got %s", journal.Status)
		}

		dr := decimal.Zero
		cr := decimal.Zero
		for _, l := range journal.Lines {
			dr = dr.Add(l.Debit)
			cr = cr.Add(l.Credit)
		}
		if !dr.Equal(cr) {
			mt.Fatalf("GL Journal out of balance: Dr %s != Cr %s", dr, cr)
		}
		if !dr.Equal(decimal.NewFromFloat(10000)) {
			mt.Errorf("expected journal amount 10000, got Dr %s", dr)
		}

		// The GL engine (real Store.Execute, mirrored here by fakeLedger) must
		// have received exactly one create + one post: this is what makes
		// re-posting the same period idempotent instead of duplicating the
		// journal (P0-2).
		if ledger.created != 1 || ledger.posted != 1 {
			mt.Fatalf("expected exactly 1 create + 1 post through the GL engine, got created=%d posted=%d", ledger.created, ledger.posted)
		}
	})

	mt.Run("3b_post_depreciation_gl_failure_not_swallowed", func(mt *mtest.T) {
		db := mt.DB
		ledger := newFakeLedger()
		ledger.failErr = fmt.Errorf("postgresql projection unavailable: connection refused")
		poster := NewGLPoster(db, ledger)

		mockDeprecItem := DepreciationScheduleItem{
			AssetCode:    asset.AssetCode,
			FiscalYear:   "2026",
			Period:       2,
			PeriodDeprec: Amount("10000.00"),
			Days:         28,
		}

		// Deliberately no "UpdateMany asset_depreciations" mock response queued:
		// if PostDepreciation reached step 6 (marking items posted) despite the
		// GL engine failing, the test would fail with "no responses remaining"
		// instead of the expected error, proving the failure is not swallowed.
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, db.Name()+".asset_depreciations", mtest.FirstBatch, toBsonD(mockDeprecItem)),
			mtest.CreateCursorResponse(0, db.Name()+".asset_depreciations", mtest.NextBatch),
			mtest.CreateCursorResponse(1, db.Name()+".fixed_assets", mtest.FirstBatch, toBsonD(asset)),
			mtest.CreateCursorResponse(0, db.Name()+".fixed_assets", mtest.NextBatch),
		)

		_, err := poster.PostDepreciation(context.Background(), scope, "2026", 2, "2026-02-28", "JV-FA-2026-02", now)
		if err == nil {
			mt.Fatalf("expected PostDepreciation to return an error when the GL engine fails, got nil")
		}
		if ledger.created != 0 {
			mt.Fatalf("expected the failed create to not register as a posted journal, got created=%d", ledger.created)
		}
	})

	mt.Run("4_dispose_asset_with_gain", func(mt *mtest.T) {
		db := mt.DB
		ledger := newFakeLedger()
		poster := NewGLPoster(db, ledger)

		disposal := AssetDisposal{
			AssetCode:             asset.AssetCode,
			DisposalDate:          "2026-06-30",
			DisposalType:          "sale",
			SalePrice:             Amount("550000.00"),
			VatAmount:             Amount("38500.00"),
			SettlementAccountCode: "110101",
			GainLossAccountCode:   "420101",
			Reason:                "ขายเครื่องจักรเก่า",
		}

		dispDeprecMock := DepreciationScheduleItem{
			AssetCode:   asset.AssetCode,
			AccumDeprec: Amount("100000.00"),
			StopDate:    "2026-06-30",
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, db.Name()+".fixed_assets", mtest.FirstBatch, toBsonD(asset)),
			mtest.CreateCursorResponse(0, db.Name()+".fixed_assets", mtest.NextBatch),
			mtest.CreateCursorResponse(1, db.Name()+".asset_depreciations", mtest.FirstBatch, toBsonD(dispDeprecMock)),
			mtest.CreateCursorResponse(0, db.Name()+".asset_depreciations", mtest.NextBatch),
			mtest.CreateSuccessResponse(), // InsertOne asset_disposals
			mtest.CreateSuccessResponse(), // UpdateOne fixed_assets
		)

		resDisp, resJourn, err := poster.DisposeAsset(context.Background(), scope, disposal, now)
		if err != nil {
			mt.Fatalf("DisposeAsset failed: %v", err)
		}

		// Gain = 550,000 - (600,000 - 100,000) = 50,000
		if !resDisp.GainLoss.Decimal().Equal(decimal.NewFromFloat(50000)) {
			mt.Errorf("expected gain of 50,000, got %s", resDisp.GainLoss)
		}

		// Verify Disposal Journal Balance
		dispDr := decimal.Zero
		dispCr := decimal.Zero
		for _, l := range resJourn.Lines {
			dispDr = dispDr.Add(l.Debit)
			dispCr = dispCr.Add(l.Credit)
		}
		if !dispDr.Equal(dispCr) {
			mt.Fatalf("Disposal Journal out of balance: Dr %s != Cr %s", dispDr, dispCr)
		}
		if ledger.created != 1 || ledger.posted != 1 {
			mt.Fatalf("expected exactly 1 create + 1 post through the GL engine, got created=%d posted=%d", ledger.created, ledger.posted)
		}
	})
}

// TestGLPoster_PostJournalIdempotent exercises postJournal (the low-level
// helper both PostDepreciation and DisposeAsset use to talk to the GL engine)
// directly, independent of Mongo mocking. It proves the deterministic
// RequestID + Store.Execute contract: replaying the exact same journal input
// (e.g. a network retry between "create" and "post" succeeding on Mongo but
// the caller not observing it) must not create a second journal.
func TestGLPoster_PostJournalIdempotent(t *testing.T) {
	ledger := newFakeLedger()
	poster := NewGLPoster(nil, ledger)
	scope := Scope{Holding: "TEST_HOLDING", Company: "TEST_CO", Branch: "HQ", Actor: "CPA_AUDITOR"}

	journalInput := &gl.Journal{
		DocNo:       "JV-FA-2026-03",
		Date:        "2026-03-31",
		BookCode:    "JV",
		FiscalYear:  "2026",
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
