package fixedasset

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	gl "smlcloudplatform/internal/generalledger"
)

// Regression coverage for docs/kms/bugs/2026-09-25-fa-edit-does-not-recalculate-schedule.md:
// saving an asset whose depreciation inputs changed must regenerate its schedule exactly like the
// "recalculate" action (both refused while a period is posted to GL); other edits must leave the
// schedule rows alone.

func recalcTestAsset(code string) Asset {
	return Asset{
		AssetCode:                code,
		Names:                    []Name{{Code: "th", Name: "รถกระบะบรรทุกสินค้า"}},
		AssetTypeCode:            "VEHICLE",
		PurchaseDate:             "2026-01-01",
		StartCalcDate:            "2026-01-01",
		Cost:                     Amount("120000.00"),
		ScrapValue:               Amount("1.00"),
		UsefulLifeYears:          5,
		DeprecPercent:            Amount("20.00"),
		FirstYearPercent:         Amount("0.00"),
		BeginAccumDeprec:         Amount("0.00"),
		AssetAccountCode:         "12130",
		AccumDeprecAccountCode:   "12131",
		DeprecExpenseAccountCode: "53130",
		Status:                   "active",
	}
}

type storedScheduleRow struct {
	UpdatedAt time.Time
	Payload   string
}

// storedScheduleRows reads the raw depreciation rows (id → column updated_at + payload) of one asset.
func storedScheduleRows(t *testing.T, db *sql.DB, company, assetCode string) map[string]storedScheduleRow {
	t.Helper()
	rows, err := db.Query(`SELECT id, updated_at, payload::text FROM fa_records WHERE company = $1 AND kind = $2 AND payload->>'assetcode' = $3`, company, kindDepreciation, assetCode)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[string]storedScheduleRow{}
	for rows.Next() {
		var id string
		var row storedScheduleRow
		if err := rows.Scan(&id, &row.UpdatedAt, &row.Payload); err != nil {
			t.Fatal(err)
		}
		out[id] = row
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func scheduleAmounts(items []DepreciationScheduleItem) map[string]string {
	out := make(map[string]string, len(items))
	for _, it := range items {
		out[depreciationKey(it)] = fmt.Sprintf("%s|%s|%s|%s", it.PeriodDeprec.Decimal().StringFixed(2), it.AccumDeprec.Decimal().StringFixed(2), it.NetBookValue.Decimal().StringFixed(2), it.StartDate)
	}
	return out
}

// assertScheduleMatches checks the stored unposted rows equal a fresh CalculateSchedule of want,
// ignoring the periods listed in posted (those rows are never regenerated).
func assertScheduleMatches(t *testing.T, store *Store, scope Scope, want Asset, posted map[string]bool) {
	t.Helper()
	calculated, err := store.calc.CalculateSchedule(want, "")
	if err != nil {
		t.Fatal(err)
	}
	expected := scheduleAmounts(calculated)
	for key := range posted {
		delete(expected, key)
	}
	stored, err := store.GetAssetDepreciationSchedule(context.Background(), scope, want.AssetCode)
	if err != nil {
		t.Fatal(err)
	}
	var unposted []DepreciationScheduleItem
	for _, it := range stored {
		if !it.IsPosted {
			unposted = append(unposted, it)
		}
	}
	got := scheduleAmounts(unposted)
	if len(got) != len(expected) {
		t.Fatalf("stored unposted periods = %d, want %d", len(got), len(expected))
	}
	for key, want := range expected {
		if got[key] != want {
			t.Fatalf("period %s stored %q, want %q (schedule not regenerated from the edited asset)", key, got[key], want)
		}
	}
}

// assertScheduleTotal checks the stored rows, posted or not, follow on from each other and add up to
// cost − scrap − beginaccumdeprec: what GL holds once every period is posted.
func assertScheduleTotal(t *testing.T, store *Store, scope Scope, asset Asset) {
	t.Helper()
	items, err := store.GetAssetDepreciationSchedule(context.Background(), scope, asset.AssetCode)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("schedule empty")
	}
	begin := asset.BeginAccumDeprec.Decimal()
	accum := begin
	for _, it := range items {
		accum = accum.Add(it.PeriodDeprec.Decimal())
		if !it.AccumDeprec.Decimal().Equal(accum) {
			t.Fatalf("period %s accum %s, want %s (does not follow on from the period before)", depreciationKey(it), it.AccumDeprec.Decimal().StringFixed(2), accum.StringFixed(2))
		}
	}
	want := asset.Cost.Decimal().Sub(asset.ScrapValue.Decimal()).Sub(begin)
	if total := accum.Sub(begin); !total.Equal(want) {
		t.Fatalf("lifetime depreciation %s, want cost − scrap − begin accum %s", total.StringFixed(2), want.StringFixed(2))
	}
}

// scheduleEdits changes every Asset field CalculateSchedule reads (calculator.go), one per case;
// AssetCode cannot change on update.
var scheduleEdits = []struct {
	name    string
	prepare func(*Asset) // optional change to the asset before it is created
	edit    func(*Asset)
}{
	{"deprecpercent", nil, func(a *Asset) { a.DeprecPercent = Amount("25.00") }},
	{"firstyearpercent", nil, func(a *Asset) { a.FirstYearPercent = Amount("40.00") }},
	{"beginaccumdeprec", nil, func(a *Asset) { a.BeginAccumDeprec = Amount("24000.00") }},
	{"cost", nil, func(a *Asset) { a.Cost = Amount("150000.00") }},
	{"scrapvalue", nil, func(a *Asset) { a.ScrapValue = Amount("20000.00") }},
	{"startcalcdate", nil, func(a *Asset) { a.StartCalcDate = "2026-03-15" }},
	// Rate left blank: the calculator derives it from the useful life.
	{"usefullifeyears", func(a *Asset) { a.DeprecPercent = Amount("0") }, func(a *Asset) { a.UsefulLifeYears = 4 }},
}

func TestUpdateAssetRegeneratesScheduleWhenInputsChange(t *testing.T) {
	connect, company := testConnector(t)
	ctx := context.Background()
	scope := Scope{Holding: "RUNGRUENG_FA", Company: company, Branch: "00000", Actor: "fa-edit-test"}
	store := NewStore(connect)
	created := time.Date(2026, 9, 25, 1, 0, 0, 0, time.UTC)

	for i, tc := range scheduleEdits {
		t.Run(tc.name, func(t *testing.T) {
			base := recalcTestAsset(fmt.Sprintf("FA-EDIT-%02d", i))
			if tc.prepare != nil {
				tc.prepare(&base)
			}
			asset, err := store.CreateAsset(ctx, scope, base, created)
			if err != nil {
				t.Fatalf("CreateAsset: %v", err)
			}
			edited := *asset
			tc.edit(&edited)
			updated, err := store.UpdateAsset(ctx, scope, asset.ID, edited, asset.Version, created.Add(time.Hour))
			if err != nil {
				t.Fatalf("UpdateAsset: %v", err)
			}
			assertScheduleMatches(t, store, scope, *updated, nil)
			assertScheduleTotal(t, store, scope, *updated)
		})
	}
}

func TestUpdateAssetNonScheduleEditKeepsScheduleRows(t *testing.T) {
	connect, company := testConnector(t)
	ctx := context.Background()
	db, _ := connect("")
	scope := Scope{Holding: "RUNGRUENG_FA", Company: company, Branch: "00000", Actor: "fa-edit-test"}
	store := NewStore(connect)
	created := time.Date(2026, 9, 25, 1, 0, 0, 0, time.UTC)

	asset, err := store.CreateAsset(ctx, scope, recalcTestAsset("FA-NAME-01"), created)
	if err != nil {
		t.Fatalf("CreateAsset: %v", err)
	}
	before := storedScheduleRows(t, db, company, asset.AssetCode)
	if len(before) == 0 {
		t.Fatal("create must generate the schedule")
	}

	edited := *asset
	edited.Names = []Name{{Code: "th", Name: "รถกระบะส่งปูนซีเมนต์ สาขาลาดหลุมแก้ว"}}
	edited.LocationCode = "LLK"
	edited.Notes = "ย้ายไปใช้ที่สาขาลาดหลุมแก้ว"
	edited.DeprecPercent = Amount("20") // same value written differently is not a change
	time.Sleep(20 * time.Millisecond)   // a rewrite would get a later updated_at column
	updated, err := store.UpdateAsset(ctx, scope, asset.ID, edited, asset.Version, created.Add(time.Hour))
	if err != nil {
		t.Fatalf("UpdateAsset: %v", err)
	}
	if updated.Version != asset.Version+1 || updated.ThaiName() != "รถกระบะส่งปูนซีเมนต์ สาขาลาดหลุมแก้ว" {
		t.Fatalf("asset not updated: version %d name %q", updated.Version, updated.ThaiName())
	}
	after := storedScheduleRows(t, db, company, asset.AssetCode)
	if len(after) != len(before) {
		t.Fatalf("schedule rows %d → %d after a name-only edit", len(before), len(after))
	}
	for id, row := range before {
		if got, ok := after[id]; !ok || !got.UpdatedAt.Equal(row.UpdatedAt) || got.Payload != row.Payload {
			t.Fatalf("schedule row %s was rewritten by a name-only edit", id)
		}
	}
}

// mustSchedulePosted checks err is the refusal a schedule rebuild gets while a period is posted.
func mustSchedulePosted(t *testing.T, err error) {
	t.Helper()
	user, ok := gl.AsUserError(err)
	if !ok || user.Code != CodeSchedulePosted || user.HTTPStatus() != http.StatusConflict {
		t.Fatalf("want %s (409), got %#v (%v)", CodeSchedulePosted, user, err)
	}
}

// Once a period is posted to GL, rebuilding the schedule from the asset would start again from
// beginaccumdeprec and not follow on from the posted amounts (lifetime total ≠ cost − scrap), so
// both a schedule-input edit and the recalculate action are refused until the posting is reversed;
// other edits still save and leave every row alone.
func TestScheduleRebuildRefusedWhilePeriodPosted(t *testing.T) {
	connect, company := testConnector(t)
	ctx := context.Background()
	db, _ := connect("")
	scope := Scope{Holding: "RUNGRUENG_FA", Company: company, Branch: "00000", Actor: "fa-edit-test"}
	store := NewStore(connect)
	poster := NewGLPoster(connect, newFakeLedger(), allowBranch)
	created := time.Date(2026, 9, 25, 1, 0, 0, 0, time.UTC)

	asset, err := store.CreateAsset(ctx, scope, recalcTestAsset("FA-POSTED-01"), created)
	if err != nil {
		t.Fatalf("CreateAsset: %v", err)
	}
	journal, err := poster.PostDepreciation(ctx, scope, "2026", 1, "2026-01-31", "", "", created)
	if err != nil {
		t.Fatalf("PostDepreciation: %v", err)
	}
	before := storedScheduleRows(t, db, company, asset.AssetCode)
	postedID := entityID(scope, kindDepreciation, fmt.Sprintf("%s-2026-1", asset.AssetCode))
	if !strings.Contains(before[postedID].Payload, `"isposted": true`) {
		t.Fatalf("period 1 not posted: %s", before[postedID].Payload)
	}
	assertScheduleTotal(t, store, scope, *asset)

	rowsUnchanged := func(t *testing.T) {
		t.Helper()
		after := storedScheduleRows(t, db, company, asset.AssetCode)
		if len(after) != len(before) {
			t.Fatalf("schedule rows %d → %d", len(before), len(after))
		}
		for id, row := range before {
			if got, ok := after[id]; !ok || !got.UpdatedAt.Equal(row.UpdatedAt) || got.Payload != row.Payload {
				t.Fatalf("schedule row %s was rewritten", id)
			}
		}
	}

	for _, tc := range scheduleEdits {
		t.Run("refuse "+tc.name, func(t *testing.T) {
			edited := *asset
			tc.edit(&edited)
			_, err := store.UpdateAsset(ctx, scope, asset.ID, edited, asset.Version, created.Add(time.Hour))
			mustSchedulePosted(t, err)
			rowsUnchanged(t)
			stored, err := store.GetAsset(ctx, scope, asset.ID)
			if err != nil {
				t.Fatal(err)
			}
			if stored.Version != asset.Version || stored.ThaiName() != asset.ThaiName() {
				t.Fatalf("refused update still saved the asset: version %d → %d", asset.Version, stored.Version)
			}
		})
	}
	t.Run("refuse recalculate", func(t *testing.T) {
		mustSchedulePosted(t, store.RecalculateAssetSchedule(ctx, scope, asset.AssetCode, created.Add(time.Hour)))
		rowsUnchanged(t)
	})

	// A name/location/notes edit is not a schedule input: it saves and leaves every row alone.
	edited := *asset
	edited.Names = []Name{{Code: "th", Name: "รถกระบะส่งปูนซีเมนต์ สาขาบางนา"}}
	edited.LocationCode = "BNA"
	edited.Notes = "ย้ายไปใช้ที่สาขาบางนา"
	renamed, err := store.UpdateAsset(ctx, scope, asset.ID, edited, asset.Version, created.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("name-only edit with a posted period: %v", err)
	}
	if renamed.Version != asset.Version+1 {
		t.Fatalf("name-only edit version %d, want %d", renamed.Version, asset.Version+1)
	}
	rowsUnchanged(t)

	// After the posting is reversed nothing is posted, so the rate edit goes through and rebuilds.
	if err := poster.ReverseDepreciation(ctx, scope, journal.DocNo, "แก้อัตราค่าเสื่อมราคา", created.Add(3*time.Hour)); err != nil {
		t.Fatalf("ReverseDepreciation: %v", err)
	}
	rerated := *renamed
	rerated.DeprecPercent = Amount("25.00")
	final, err := store.UpdateAsset(ctx, scope, renamed.ID, rerated, renamed.Version, created.Add(4*time.Hour))
	if err != nil {
		t.Fatalf("UpdateAsset after reversal: %v", err)
	}
	assertScheduleMatches(t, store, scope, *final, nil)
	assertScheduleTotal(t, store, scope, *final)

	// Posting the period again gets the next number: the reversed journal keeps its number for
	// audit, and replaying its GL request would mark the row posted while GL posts nothing.
	again, err := poster.PostDepreciation(ctx, scope, "2026", 1, "2026-01-31", "", "", created.Add(5*time.Hour))
	if err != nil {
		t.Fatalf("PostDepreciation after reversal and a rate edit: %v", err)
	}
	if again.DocNo != journal.DocNo+"-2" || again.Status != "posted" {
		t.Fatalf("re-posted journal %s (%s), want %s-2 posted", again.DocNo, again.Status, journal.DocNo)
	}
	reposted := storedScheduleRows(t, db, company, asset.AssetCode)[postedID].Payload
	if !strings.Contains(reposted, `"isposted": true`) || !strings.Contains(reposted, `"journaldocno": "`+again.DocNo+`"`) {
		t.Fatalf("period 1 must point at %s: %s", again.DocNo, reposted)
	}
}
