package fixedasset

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func toBsonD(v interface{}) bson.D {
	data, _ := bson.Marshal(v)
	var doc bson.D
	_ = bson.Unmarshal(data, &doc)
	return doc
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
			mtest.CreateCursorResponse(1, db.Name()+".fixed_assets", mtest.FirstBatch, bson.D{{"n", 0}}),
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
		poster := NewGLPoster(db, nil)

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
			mtest.CreateSuccessResponse(), // InsertOne gl_journals
			mtest.CreateSuccessResponse(), // UpdateMany asset_depreciations
		)

		journal, err := poster.PostDepreciation(context.Background(), scope, "2026", 1, "2026-01-31", "JV-FA-2026-01", now)
		if err != nil {
			mt.Fatalf("PostDepreciation failed: %v", err)
		}
		if journal.DocNo != "JV-FA-2026-01" {
			mt.Errorf("expected docno JV-FA-2026-01, got %s", journal.DocNo)
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
	})

	mt.Run("4_dispose_asset_with_gain", func(mt *mtest.T) {
		db := mt.DB
		poster := NewGLPoster(db, nil)

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
			mtest.CreateSuccessResponse(), // InsertOne gl_journals
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
	})
}
