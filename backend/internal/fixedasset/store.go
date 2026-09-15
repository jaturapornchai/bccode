package fixedasset

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNotFound      = errors.New("ไม่พบข้อมูลที่ต้องการ")
	ErrStaleVersion  = errors.New("รายการนี้ถูกแก้ไขโดยผู้ใช้อื่นไปแล้ว กรุณาโหลดข้อมูลใหม่")
	ErrCodeDuplicate = errors.New("รหัสนี้มีอยู่ในระบบแล้ว")
)

type Store struct {
	db      *mongo.Database
	calc    *Calculator
	indexMu sync.Mutex
	indexed bool
}

func NewStore(db *mongo.Database) *Store {
	return &Store{
		db:   db,
		calc: NewCalculator(),
	}
}

func scopeFilter(s Scope) bson.M {
	return bson.M{"holdingcode": s.Holding, "businesscode": s.Company}
}

func scopedID(s Scope, id string) bson.M {
	f := scopeFilter(s)
	f["_id"] = id
	return f
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

func (s *Store) EnsureIndexes(ctx context.Context) error {
	s.indexMu.Lock()
	defer s.indexMu.Unlock()
	if s.indexed {
		return nil
	}

	// Index for fixed_assets
	_, err := s.db.Collection("fixed_assets").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "holdingcode", Value: 1}, {Key: "businesscode", Value: 1}, {Key: "assetcode", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "holdingcode", Value: 1}, {Key: "businesscode", Value: 1}, {Key: "status", Value: 1}},
		},
	})
	if err != nil {
		return err
	}

	// Index for asset_types
	_, err = s.db.Collection("asset_types").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "holdingcode", Value: 1}, {Key: "businesscode", Value: 1}, {Key: "typecode", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// Index for asset_depreciations
	_, err = s.db.Collection("asset_depreciations").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "holdingcode", Value: 1}, {Key: "businesscode", Value: 1}, {Key: "assetcode", Value: 1}, {Key: "fiscalyear", Value: 1}, {Key: "period", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return err
	}

	s.indexed = true
	return nil
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

	f := scopeFilter(scope)
	f["assetcode"] = asset.AssetCode
	f["isdeleted"] = false
	count, err := s.db.Collection("fixed_assets").CountDocuments(ctx, f)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrCodeDuplicate
	}

	asset.Identity = identityFor(scope, "assets", asset.AssetCode, Identity{}, now)
	_, err = s.db.Collection("fixed_assets").InsertOne(ctx, asset)
	if err != nil {
		return nil, err
	}

	// Calculate initial schedule
	schedule, err := s.calc.CalculateSchedule(asset, "")
	if err == nil && len(schedule) > 0 {
		var docs []interface{}
		for _, item := range schedule {
			item.Identity = identityFor(scope, "depreciations", fmt.Sprintf("%s-%s-%d", item.AssetCode, item.FiscalYear, item.Period), Identity{}, now)
			docs = append(docs, item)
		}
		_, _ = s.db.Collection("asset_depreciations").InsertMany(ctx, docs)
	}

	return &asset, nil
}

func (s *Store) UpdateAsset(ctx context.Context, scope Scope, id string, asset Asset, expectedVersion int64, now time.Time) (*Asset, error) {
	var old Asset
	err := s.db.Collection("fixed_assets").FindOne(ctx, scopedID(scope, id)).Decode(&old)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if old.IsDeleted {
		return nil, ErrNotFound
	}
	if expectedVersion > 0 && old.Version != expectedVersion {
		return nil, ErrStaleVersion
	}

	asset.Identity = identityFor(scope, "assets", old.AssetCode, old.Identity, now)
	asset.AssetCode = old.AssetCode // Cannot change asset code

	_, err = s.db.Collection("fixed_assets").ReplaceOne(ctx, scopedID(scope, id), asset)
	if err != nil {
		return nil, err
	}

	// Recalculate unposted schedule if cost, useful life, scrap, or dates changed
	if !asset.Cost.Decimal().Equal(old.Cost.Decimal()) || asset.UsefulLifeYears != old.UsefulLifeYears || !asset.ScrapValue.Decimal().Equal(old.ScrapValue.Decimal()) || asset.StartCalcDate != old.StartCalcDate {
		_ = s.RecalculateAssetSchedule(ctx, scope, asset.AssetCode, now)
	}

	return &asset, nil
}

func (s *Store) DeleteAsset(ctx context.Context, scope Scope, id string, expectedVersion int64, now time.Time) error {
	var old Asset
	err := s.db.Collection("fixed_assets").FindOne(ctx, scopedID(scope, id)).Decode(&old)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return err
	}
	if expectedVersion > 0 && old.Version != expectedVersion {
		return ErrStaleVersion
	}

	// Check if any depreciation is already posted to GL
	fDep := scopeFilter(scope)
	fDep["assetcode"] = old.AssetCode
	fDep["isposted"] = true
	fDep["isdeleted"] = false
	postedCount, err := s.db.Collection("asset_depreciations").CountDocuments(ctx, fDep)
	if err != nil {
		return err
	}
	if postedCount > 0 {
		return fmt.Errorf("สินทรัพย์นี้มีการผ่านรายการค่าเสื่อมราคาเข้า GL แล้ว ไม่สามารถลบได้ กรุณายกเลิกการผ่านรายการก่อน")
	}

	old.Identity = identityFor(scope, "assets", old.AssetCode, old.Identity, now)
	old.IsDeleted = true

	_, err = s.db.Collection("fixed_assets").ReplaceOne(ctx, scopedID(scope, id), old)
	if err != nil {
		return err
	}

	// Delete unposted depreciations
	fDel := scopeFilter(scope)
	fDel["assetcode"] = old.AssetCode
	fDel["isposted"] = false
	_, _ = s.db.Collection("asset_depreciations").DeleteMany(ctx, fDel)

	return nil
}

func (s *Store) GetAsset(ctx context.Context, scope Scope, id string) (*Asset, error) {
	var asset Asset
	err := s.db.Collection("fixed_assets").FindOne(ctx, scopedID(scope, id)).Decode(&asset)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Try by AssetCode
			f := scopeFilter(scope)
			f["assetcode"] = id
			f["isdeleted"] = false
			err2 := s.db.Collection("fixed_assets").FindOne(ctx, f).Decode(&asset)
			if err2 != nil {
				return nil, ErrNotFound
			}
			return &asset, nil
		}
		return nil, err
	}
	if asset.IsDeleted {
		return nil, ErrNotFound
	}
	return &asset, nil
}

func (s *Store) ListAssets(ctx context.Context, scope Scope, q, typeCode, status string, page, limit int) ([]Asset, int64, error) {
	f := scopeFilter(scope)
	f["isdeleted"] = false

	if typeCode != "" {
		f["assettypecode"] = typeCode
	}
	if status != "" {
		f["status"] = status
	}
	if q != "" {
		f["$or"] = []bson.M{
			{"assetcode": bson.M{"$regex": q, "$options": "i"}},
			{"names.name": bson.M{"$regex": q, "$options": "i"}},
			{"serialnumber": bson.M{"$regex": q, "$options": "i"}},
		}
	}

	total, err := s.db.Collection("fixed_assets").CountDocuments(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	skip := int64((page - 1) * limit)

	opts := options.Find().SetSort(bson.D{{Key: "assetcode", Value: 1}}).SetSkip(skip).SetLimit(int64(limit))
	cur, err := s.db.Collection("fixed_assets").Find(ctx, f, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var items []Asset
	if err = cur.All(ctx, &items); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ---------------- Depreciation Schedule ----------------

func (s *Store) RecalculateAssetSchedule(ctx context.Context, scope Scope, assetCode string, now time.Time) error {
	var asset Asset
	f := scopeFilter(scope)
	f["assetcode"] = assetCode
	f["isdeleted"] = false
	if err := s.db.Collection("fixed_assets").FindOne(ctx, f).Decode(&asset); err != nil {
		return err
	}

	// Fetch already posted items to preserve them
	fPosted := scopeFilter(scope)
	fPosted["assetcode"] = assetCode
	fPosted["isposted"] = true
	fPosted["isdeleted"] = false
	cur, err := s.db.Collection("asset_depreciations").Find(ctx, fPosted)
	if err != nil {
		return err
	}
	var postedItems []DepreciationScheduleItem
	_ = cur.All(ctx, &postedItems)

	// Remove unposted items
	fUnposted := scopeFilter(scope)
	fUnposted["assetcode"] = assetCode
	fUnposted["isposted"] = false
	_, err = s.db.Collection("asset_depreciations").DeleteMany(ctx, fUnposted)
	if err != nil {
		return err
	}

	// Calculate new schedule
	schedule, err := s.calc.CalculateSchedule(asset, "")
	if err != nil {
		return err
	}

	postedMap := make(map[string]bool)
	for _, p := range postedItems {
		key := fmt.Sprintf("%s-%d", p.FiscalYear, p.Period)
		postedMap[key] = true
	}

	var newDocs []interface{}
	for _, item := range schedule {
		key := fmt.Sprintf("%s-%d", item.FiscalYear, item.Period)
		if postedMap[key] {
			continue // Skip already posted
		}
		item.Identity = identityFor(scope, "depreciations", fmt.Sprintf("%s-%s-%d", item.AssetCode, item.FiscalYear, item.Period), Identity{}, now)
		newDocs = append(newDocs, item)
	}

	if len(newDocs) > 0 {
		_, err = s.db.Collection("asset_depreciations").InsertMany(ctx, newDocs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) GetAssetDepreciationSchedule(ctx context.Context, scope Scope, assetCode string) ([]DepreciationScheduleItem, error) {
	f := scopeFilter(scope)
	f["assetcode"] = assetCode
	f["isdeleted"] = false

	opts := options.Find().SetSort(bson.D{{Key: "fiscalyear", Value: 1}, {Key: "period", Value: 1}})
	cur, err := s.db.Collection("asset_depreciations").Find(ctx, f, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var items []DepreciationScheduleItem
	if err = cur.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// ---------------- Asset Types CRUD ----------------

func (s *Store) CreateAssetType(ctx context.Context, scope Scope, at AssetType, now time.Time) (*AssetType, error) {
	if CleanString(at.TypeCode) == "" {
		return nil, fmt.Errorf("กรุณาระบุรหัสประเภทสินทรัพย์")
	}

	f := scopeFilter(scope)
	f["typecode"] = at.TypeCode
	f["isdeleted"] = false
	count, err := s.db.Collection("asset_types").CountDocuments(ctx, f)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrCodeDuplicate
	}

	at.Identity = identityFor(scope, "types", at.TypeCode, Identity{}, now)
	at.IsActive = true
	_, err = s.db.Collection("asset_types").InsertOne(ctx, at)
	if err != nil {
		return nil, err
	}
	return &at, nil
}

func (s *Store) ListAssetTypes(ctx context.Context, scope Scope) ([]AssetType, error) {
	f := scopeFilter(scope)
	f["isdeleted"] = false

	opts := options.Find().SetSort(bson.D{{Key: "typecode", Value: 1}})
	cur, err := s.db.Collection("asset_types").Find(ctx, f, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var items []AssetType
	if err = cur.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}
