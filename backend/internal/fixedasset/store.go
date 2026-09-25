package fixedasset

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	gl "smlcloudplatform/internal/generalledger"
)

var (
	ErrNotFound      = errors.New("ไม่พบข้อมูลที่ต้องการ")
	ErrStaleVersion  = errors.New("รายการนี้ถูกแก้ไขโดยผู้ใช้อื่นไปแล้ว กรุณาโหลดข้อมูลใหม่")
	ErrCodeDuplicate = errors.New("รหัสนี้มีอยู่ในระบบแล้ว")
)

// CodeSchedulePosted is the refusal to rebuild a depreciation schedule once any period of the asset
// is posted to GL: CalculateSchedule starts again from beginaccumdeprec and cannot continue from the
// amounts already in the ledger, so the rebuilt rows would no longer add up to cost − scrap − begin
// accum. The screen text is gl_err_fa_schedule_posted in languages.tsv (th column = Message).
const CodeSchedulePosted = "fa_schedule_posted"

var errSchedulePosted = &gl.UserError{
	Code:    CodeSchedulePosted,
	Status:  http.StatusConflict,
	Message: "สินทรัพย์นี้มีค่าเสื่อมราคาผ่านรายการเข้า GL แล้ว จึงแก้ราคาทุน ราคาซาก อัตราค่าเสื่อม % ค่าเสื่อมปีแรก อายุการใช้งาน วันเริ่มคิด หรือค่าเสื่อมสะสมยกมาไม่ได้ กรุณายกเลิกการผ่านรายการค่าเสื่อมราคาของสินทรัพย์นี้ก่อน แล้วทำรายการอีกครั้ง",
}

// Store keeps fixed-asset master data and depreciation schedules in the holding's PostgreSQL database.
type Store struct {
	records *records
	calc    *Calculator
}

func NewStore(connect Connector) *Store {
	return &Store{
		records: newRecords(connect),
		calc:    NewCalculator(),
	}
}

func digest(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func entityID(s Scope, resource, code string) string {
	return digest([]byte(s.Holding + "\x00" + s.Company + "\x00" + resource + "\x00" + code))
}

func identityFor(s Scope, resource, code string, old Identity, now time.Time) Identity {
	id := old.ID
	if id == "" {
		id = entityID(s, resource, code)
	}
	ver := old.Version + 1
	if ver < 1 {
		ver = 1
	}
	created := old.CreatedAt
	if created.IsZero() {
		created = now
	}
	createdBy := old.CreatedBy
	if createdBy == "" {
		createdBy = s.Actor
	}
	return Identity{
		ID:           id,
		HoldingCode:  s.Holding,
		BusinessCode: s.Company,
		Version:      ver,
		CreatedAt:    created,
		CreatedBy:    createdBy,
		UpdatedAt:    now,
		UpdatedBy:    s.Actor,
		IsDeleted:    false,
	}
}

// ---------------- Fixed Assets CRUD ----------------

func (s *Store) CreateAsset(ctx context.Context, scope Scope, asset Asset, now time.Time) (*Asset, error) {
	if CleanString(asset.AssetCode) == "" {
		return nil, fmt.Errorf("กรุณาระบุรหัสสินทรัพย์")
	}
	if asset.Cost.Decimal().LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("ราคาทุนสินทรัพย์ต้องมากกว่า 0")
	}
	if asset.StartCalcDate == "" {
		asset.StartCalcDate = asset.PurchaseDate
	}
	if asset.Status == "" {
		asset.Status = "active"
	}

	db, err := s.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	count, err := countRecords(ctx, tx, scope.Company, kindAsset, ` AND code = $3`, asset.AssetCode)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrCodeDuplicate
	}

	asset.Identity = identityFor(scope, kindAsset, asset.AssetCode, Identity{}, now)
	if err := putRecord(ctx, tx, scope.Company, kindAsset, asset.ID, asset.AssetCode, asset); err != nil {
		return nil, err
	}
	schedule, err := s.calc.CalculateSchedule(asset, "")
	if err != nil {
		return nil, err
	}
	if err := putSchedule(ctx, tx, scope, schedule, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &asset, nil
}

func (s *Store) UpdateAsset(ctx context.Context, scope Scope, id string, asset Asset, expectedVersion int64, now time.Time) (*Asset, error) {
	db, err := s.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	old, err := lockRecord[Asset](ctx, tx, scope.Company, kindAsset, ` AND id = $3`, id)
	if err != nil {
		return nil, err
	}
	if expectedVersion > 0 && old.Version != expectedVersion {
		return nil, ErrStaleVersion
	}

	asset.Identity = identityFor(scope, kindAsset, old.AssetCode, old.Identity, now)
	asset.AssetCode = old.AssetCode // Cannot change asset code
	if err := putRecord(ctx, tx, scope.Company, kindAsset, asset.ID, asset.AssetCode, asset); err != nil {
		return nil, err
	}

	// Regenerate the schedule, as the recalculate action does, when a calculation input changed;
	// refused (whole update rolled back) once a period is posted to GL.
	if scheduleInputsChanged(*old, asset) {
		if err := s.recalculateSchedule(ctx, tx, scope, asset, now); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &asset, nil
}

// scheduleInputsChanged reports whether an edit touches any Asset field Calculator.CalculateSchedule
// reads; keep this list in step with calculator.go. Amounts compare by value ("20" == "20.00").
func scheduleInputsChanged(old, asset Asset) bool {
	return !asset.Cost.Decimal().Equal(old.Cost.Decimal()) ||
		!asset.ScrapValue.Decimal().Equal(old.ScrapValue.Decimal()) ||
		!asset.DeprecPercent.Decimal().Equal(old.DeprecPercent.Decimal()) ||
		!asset.FirstYearPercent.Decimal().Equal(old.FirstYearPercent.Decimal()) ||
		!asset.BeginAccumDeprec.Decimal().Equal(old.BeginAccumDeprec.Decimal()) ||
		asset.UsefulLifeYears != old.UsefulLifeYears ||
		asset.StartCalcDate != old.StartCalcDate
}

func (s *Store) DeleteAsset(ctx context.Context, scope Scope, id string, expectedVersion int64, now time.Time) error {
	db, err := s.records.db(ctx, scope.Holding)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	old, err := lockRecord[Asset](ctx, tx, scope.Company, kindAsset, ` AND id = $3`, id)
	if err != nil {
		return err
	}
	if expectedVersion > 0 && old.Version != expectedVersion {
		return ErrStaleVersion
	}

	// Check if any depreciation is already posted to GL
	postedCount, err := countPostedDepreciation(ctx, tx, scope, old.AssetCode)
	if err != nil {
		return err
	}
	if postedCount > 0 {
		return fmt.Errorf("สินทรัพย์นี้มีการผ่านรายการค่าเสื่อมราคาเข้า GL แล้ว ไม่สามารถลบได้ กรุณายกเลิกการผ่านรายการก่อน")
	}

	old.Identity = identityFor(scope, kindAsset, old.AssetCode, old.Identity, now)
	old.IsDeleted = true
	if err := putRecord(ctx, tx, scope.Company, kindAsset, old.ID, old.AssetCode, old); err != nil {
		return err
	}
	if err := deleteUnpostedSchedule(ctx, tx, scope, old.AssetCode); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) GetAsset(ctx context.Context, scope Scope, id string) (*Asset, error) {
	db, err := s.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	// Accept either the document id or the asset code.
	return firstRecord[Asset](ctx, db, scope.Company, kindAsset, ` AND (id = $3 OR code = $3) ORDER BY (id = $3) DESC`, id)
}

func (s *Store) ListAssets(ctx context.Context, scope Scope, q, typeCode, status string, page, limit int) ([]Asset, int64, error) {
	db, err := s.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, 0, err
	}
	// Search is a literal substring on code, serial number and names (never a regex from user input).
	where := ` AND ($3 = '' OR payload->>'assettypecode' = $3) AND ($4 = '' OR payload->>'status' = $4)` +
		` AND ($5 = '' OR strpos(lower(code || ' ' || COALESCE(payload->>'serialnumber', '') || ' ' || COALESCE(jsonb_path_query_array(payload, '$.names[*].name')::text, '')), lower($5)) > 0)`
	total, err := countRecords(ctx, db, scope.Company, kindAsset, where, typeCode, status, q)
	if err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	items, err := queryRecords[Asset](ctx, db, scope.Company, kindAsset, where, ` ORDER BY code LIMIT $6 OFFSET $7`, typeCode, status, q, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ---------------- Depreciation Schedule ----------------

func depreciationKey(item DepreciationScheduleItem) string {
	return fmt.Sprintf("%s-%s-%d", item.AssetCode, item.FiscalYear, item.Period)
}

// putSchedule stores calculated schedule rows; callers make sure no period of the asset is posted.
func putSchedule(ctx context.Context, q queryer, scope Scope, schedule []DepreciationScheduleItem, now time.Time) error {
	for _, item := range schedule {
		key := depreciationKey(item)
		item.Identity = identityFor(scope, kindDepreciation, key, Identity{}, now)
		if err := putRecord(ctx, q, scope.Company, kindDepreciation, item.ID, key, item); err != nil {
			return err
		}
	}
	return nil
}

func deleteUnpostedSchedule(ctx context.Context, q queryer, scope Scope, assetCode string) error {
	_, err := q.ExecContext(ctx, `DELETE FROM fa_records WHERE company = $1 AND kind = $2 AND payload->>'assetcode' = $3 AND NOT COALESCE((payload->>'isposted')::boolean, false)`, scope.Company, kindDepreciation, assetCode)
	return err
}

func countPostedDepreciation(ctx context.Context, q queryer, scope Scope, assetCode string) (int64, error) {
	return countRecords(ctx, q, scope.Company, kindDepreciation, ` AND payload->>'assetcode' = $3 AND COALESCE((payload->>'isposted')::boolean, false)`, assetCode)
}

// recalculateSchedule rebuilds the whole schedule from the asset; it refuses (errSchedulePosted)
// while any period is posted, because a rebuild from scratch would not follow on from the ledger.
func (s *Store) recalculateSchedule(ctx context.Context, q queryer, scope Scope, asset Asset, now time.Time) error {
	postedCount, err := countPostedDepreciation(ctx, q, scope, asset.AssetCode)
	if err != nil {
		return err
	}
	if postedCount > 0 {
		return errSchedulePosted
	}
	if err := deleteUnpostedSchedule(ctx, q, scope, asset.AssetCode); err != nil {
		return err
	}
	schedule, err := s.calc.CalculateSchedule(asset, "")
	if err != nil {
		return err
	}
	return putSchedule(ctx, q, scope, schedule, now)
}

func (s *Store) RecalculateAssetSchedule(ctx context.Context, scope Scope, assetCode string, now time.Time) error {
	db, err := s.records.db(ctx, scope.Holding)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	asset, err := firstRecord[Asset](ctx, tx, scope.Company, kindAsset, ` AND code = $3`, assetCode)
	if err != nil {
		return err
	}
	if err := s.recalculateSchedule(ctx, tx, scope, *asset, now); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) GetAssetDepreciationSchedule(ctx context.Context, scope Scope, assetCode string) ([]DepreciationScheduleItem, error) {
	db, err := s.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	return queryRecords[DepreciationScheduleItem](ctx, db, scope.Company, kindDepreciation, ` AND payload->>'assetcode' = $3`, ` ORDER BY payload->>'fiscalyear', (payload->>'period')::int`, assetCode)
}

// ---------------- Asset Types CRUD ----------------

func (s *Store) CreateAssetType(ctx context.Context, scope Scope, at AssetType, now time.Time) (*AssetType, error) {
	if CleanString(at.TypeCode) == "" {
		return nil, fmt.Errorf("กรุณาระบุรหัสประเภทสินทรัพย์")
	}
	db, err := s.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	count, err := countRecords(ctx, db, scope.Company, kindType, ` AND code = $3`, at.TypeCode)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrCodeDuplicate
	}
	at.Identity = identityFor(scope, kindType, at.TypeCode, Identity{}, now)
	at.IsActive = true
	if err := putRecord(ctx, db, scope.Company, kindType, at.ID, at.TypeCode, at); err != nil {
		return nil, err
	}
	return &at, nil
}

func (s *Store) ListAssetTypes(ctx context.Context, scope Scope) ([]AssetType, error) {
	db, err := s.records.db(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	return queryRecords[AssetType](ctx, db, scope.Company, kindType, "", ` ORDER BY code`)
}
