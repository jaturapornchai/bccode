package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	serviceconfig "smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/mypg"
	build "smlcloudplatform/internal/goapi/process/build"
	"smlcloudplatform/internal/product/projection"
)

type projectionSource interface {
	Products(context.Context, string, string, []string) (map[string]MongoProductModel, error)
	Barcodes(context.Context, string, string, []string) (map[string]models.MongoProductBarcodeModel, error)
}
type mongoProjectionSource struct{ db *mongo.Database }

func (s mongoProjectionSource) Products(ctx context.Context, holding, business string, codes []string) (map[string]MongoProductModel, error) {
	rows, err := s.db.Collection("products", options.Collection().SetReadPreference(readpref.Primary())).Find(ctx,
		bson.M{"holdingcode": holding, "businesscode": business, "code": bson.M{"$in": codes}, "deletedat": bson.M{"$exists": false}},
		options.Find().SetProjection(bson.M{"holdingcode": 1, "businesscode": 1, "guidfixed": 1, "code": 1, "names": 1, "unitcode": 1, "unitnames": 1}))
	if err != nil {
		return nil, err
	}
	defer rows.Close(ctx)
	result := make(map[string]MongoProductModel)
	for rows.Next(ctx) {
		var p MongoProductModel
		if err := rows.Decode(&p); err != nil {
			return nil, err
		}
		if p.HoldingCode != holding || p.BusinessCode != business || p.Code == "" || p.UnitCode == "" {
			return nil, errors.New("invalid current product identity or base unit")
		}
		if _, found := result[p.Code]; found {
			return nil, errors.New("duplicate current product identity")
		}
		result[p.Code] = p
	}
	return result, rows.Err()
}
func (s mongoProjectionSource) Barcodes(ctx context.Context, holding, business string, codes []string) (map[string]models.MongoProductBarcodeModel, error) {
	rows, err := s.db.Collection("productbarcodes", options.Collection().SetReadPreference(readpref.Primary())).Find(ctx,
		bson.M{"holdingcode": holding, "businesscode": business, "barcode": bson.M{"$in": codes}, "deletedat": bson.M{"$exists": false}})
	if err != nil {
		return nil, err
	}
	defer rows.Close(ctx)
	result := make(map[string]models.MongoProductBarcodeModel)
	for rows.Next(ctx) {
		var p models.MongoProductBarcodeModel
		if err := rows.Decode(&p); err != nil {
			return nil, err
		}
		if err := normalizeProductBarcodeIdentity(&p, true); err != nil {
			return nil, err
		}
		if p.HoldingCode != holding || p.BusinessCode != business {
			return nil, errors.New("current barcode crossed company boundary")
		}
		if _, found := result[p.Barcode]; found {
			return nil, errors.New("duplicate current barcode identity")
		}
		result[p.Barcode] = p
	}
	return result, rows.Err()
}

const productFenceDDL = `CREATE TABLE IF NOT EXISTS product_projection_fences (
 holding_code text NOT NULL, businesscode text NOT NULL, itemcode text NOT NULL,
 aggregateuid text NOT NULL, version bigint NOT NULL CHECK (version > 0),
 deleted boolean NOT NULL, updatedat timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY (holding_code,businesscode,itemcode,aggregateuid)
)`

type productFenceSetup struct {
	sync.Mutex
	ready bool
}

var productFenceTables sync.Map

func ensureProductFence(ctx context.Context, db *sql.DB) error {
	entry, _ := productFenceTables.LoadOrStore(db, &productFenceSetup{})
	mu := entry.(*productFenceSetup)
	mu.Lock()
	defer mu.Unlock()
	// Once set for a live connection pool, schema creation is not repeated per message.
	if mu.ready {
		return nil
	}
	if _, err := db.ExecContext(ctx, productFenceDDL); err != nil {
		return err
	}
	mu.ready = true
	return nil
}

// PostgreSQL's shared lock table holds about max_locks_per_transaction x
// max_connections entries (6400 by default), so one bulk message must not claim
// thousands of row locks. A large batch takes the exclusive company lock instead,
// exactly like a company rebuild, and therefore never deadlocks with row holders.
const maxRowLocksPerTransaction = 256

func lockProjectionRows(ctx context.Context, tx *sql.Tx, kind, holding, business string, codes []string) error {
	if len(codes) > maxRowLocksPerTransaction {
		_, err := tx.ExecContext(ctx, projection.ExclusiveSQL, projection.LockKey(kind, holding, business, ""))
		return err
	}
	if _, err := tx.ExecContext(ctx, projection.CompanySharedSQL, projection.LockKey(kind, holding, business, "")); err != nil {
		return err
	}
	for _, code := range codes {
		if _, err := tx.ExecContext(ctx, projection.ExclusiveSQL, projection.LockKey(kind, holding, business, code)); err != nil {
			return err
		}
	}
	return nil
}

// Re-read the primary MongoDB source after acquiring the row lock. This also
// makes unversioned legacy messages and delete/recreate signals safe: their
// old snapshot never becomes the new PostgreSQL state.
func reconcileProductSignal(ctx context.Context, db *sql.DB, source projectionSource, signal MongoProductModel) error {
	if signal.Projection != nil {
		meta := signal.Projection
		if meta.Version < 1 || meta.EventUID == "" || signal.GuidFixed == "" ||
			meta.AggregateUID != projection.AggregateKey(signal.HoldingCode, signal.BusinessCode, signal.GuidFixed) {
			return fmt.Errorf("%w: invalid product projection fence identity", projection.ErrRejected)
		}
	}
	if err := ensureProductFence(ctx, db); err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := lockProjectionRows(ctx, tx, "product", signal.HoldingCode, signal.BusinessCode, []string{signal.Code}); err != nil {
		return err
	}
	if signal.Projection != nil {
		var last int64
		err := tx.QueryRowContext(ctx, "SELECT version FROM product_projection_fences WHERE holding_code=$1 AND businesscode=$2 AND itemcode=$3 AND aggregateuid=$4",
			signal.HoldingCode, signal.BusinessCode, signal.Code, signal.Projection.AggregateUID).Scan(&last)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if last >= signal.Projection.Version {
			return tx.Commit()
		}
	}
	current, err := source.Products(ctx, signal.HoldingCode, signal.BusinessCode, []string{signal.Code})
	if err != nil {
		return err
	}
	p, active := current[signal.Code]
	if !active {
		_, err = tx.ExecContext(ctx, "DELETE FROM product WHERE holding_code=$1 AND businesscode=$2 AND itemcode=$3", signal.HoldingCode, signal.BusinessCode, signal.Code)
	} else {
		name, unitName := p.Code, p.UnitCode
		if len(p.Names) > 0 && p.Names[0].Name != nil && *p.Names[0].Name != "" {
			name = *p.Names[0].Name
		}
		if len(p.UnitNames) > 0 && p.UnitNames[0].Name != nil && *p.UnitNames[0].Name != "" {
			unitName = *p.UnitNames[0].Name
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO product (holding_code,businesscode,itemcode,name0,unitcode,unitname) VALUES ($1,$2,$3,$4,$5,$6)
   ON CONFLICT (holding_code,businesscode,itemcode) DO UPDATE SET name0=EXCLUDED.name0,unitcode=EXCLUDED.unitcode,unitname=EXCLUDED.unitname`,
			signal.HoldingCode, signal.BusinessCode, signal.Code, name, p.UnitCode, unitName)
	}
	if err != nil {
		return err
	}
	if meta := signal.Projection; meta != nil {
		_, err = tx.ExecContext(ctx, `INSERT INTO product_projection_fences (holding_code,businesscode,itemcode,aggregateuid,version,deleted) VALUES ($1,$2,$3,$4,$5,$6)
   ON CONFLICT (holding_code,businesscode,itemcode,aggregateuid) DO UPDATE SET version=EXCLUDED.version,deleted=EXCLUDED.deleted,updatedat=now()`,
			signal.HoldingCode, signal.BusinessCode, signal.Code, meta.AggregateUID, meta.Version, !active)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

const barcodeProjectionUpsert = `INSERT INTO productbarcode (holding_code,businesscode,barcode,itemcode,name0,checksum,groupcode,groupnames,unitcode,unitname,price1,barcoderefunitstand,barcoderefunitdivide,itemtype)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
 ON CONFLICT (holding_code,businesscode,barcode) DO UPDATE SET itemcode=EXCLUDED.itemcode,name0=EXCLUDED.name0,checksum=EXCLUDED.checksum,
 groupcode=EXCLUDED.groupcode,groupnames=EXCLUDED.groupnames,unitcode=EXCLUDED.unitcode,unitname=EXCLUDED.unitname,price1=EXCLUDED.price1,
 barcoderefunitstand=EXCLUDED.barcoderefunitstand,barcoderefunitdivide=EXCLUDED.barcoderefunitdivide,itemtype=EXCLUDED.itemtype`

func reconcileBarcodeSignals(ctx context.Context, db *sql.DB, source projectionSource, signals []models.MongoProductBarcodeModel) error {
	if len(signals) == 0 {
		return nil
	}
	holding, business, err := normalizeProductBarcodeBatch(signals, false)
	if err != nil {
		return err
	}
	seen := make(map[string]bool)
	for _, p := range signals {
		seen[p.Barcode] = true
	}
	codes := make([]string, 0, len(seen))
	for code := range seen {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := lockProjectionRows(ctx, tx, "barcode", holding, business, codes); err != nil {
		return err
	}
	current, err := source.Barcodes(ctx, holding, business, codes)
	if err != nil {
		return err
	}
	for _, code := range codes {
		p, active := current[code]
		if !active {
			if _, err = tx.ExecContext(ctx, "DELETE FROM productbarcode WHERE holding_code=$1 AND businesscode=$2 AND barcode=$3", holding, business, code); err != nil {
				return err
			}
			continue
		}
		if _, err = tx.ExecContext(ctx, barcodeProjectionUpsert, productBarcodeProjectionRecord(p)...); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func productBarcodeProjectionRecord(p models.MongoProductBarcodeModel) []any {
	name, group, unit := "", "", ""
	if len(p.Names) > 0 {
		name = p.Names[0].Name
	}
	if len(p.GroupNames) > 0 {
		group = p.GroupNames[0].Name
	}
	if len(p.ItemUnitNames) > 0 {
		unit = p.ItemUnitNames[0].Name
	}
	// Preserve the existing projection representation; money conversion is a separate contract change.
	price := 0.0
	if len(p.Prices) > 0 {
		price = p.Prices[0].Price
	}
	checksum := myglobal.CalculateMD5(fmt.Sprintf("%v", p))
	return []any{p.HoldingCode, p.BusinessCode, p.Barcode, p.ItemCode, name, checksum, p.GroupCode, group, p.ItemUnitCode, unit, price, p.StandValue, p.DivideValue, p.ItemType}
}

func projectionRuntime(holding string) (*sql.DB, projectionSource, error) {
	client := myglobal.SafeMongoConnectFast()
	if client == nil {
		return nil, nil, errors.New("MongoDB unavailable for projection reconciliation")
	}
	build.DatabaseChecker(holding, false)
	db, err := mypg.PgSqlFastConnect(holding)
	return db, mongoProjectionSource{client.Database(serviceconfig.NewServiceConfig().MongodbDatabaseName())}, err
}
func consumeProductSignal(msg string) error {
	var p MongoProductModel
	if err := json.Unmarshal([]byte(msg), &p); err != nil {
		return fmt.Errorf("%w: %v", projection.ErrRejected, err)
	}
	if err := normalizeProductSignal(&p); err != nil {
		return err
	}
	db, source, err := projectionRuntime(p.HoldingCode)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return reconcileProductSignal(ctx, db, source, p)
}
func consumeBarcodeSignals(msg string, batch bool) error {
	var signals []models.MongoProductBarcodeModel
	if batch {
		if err := json.Unmarshal([]byte(msg), &signals); err != nil {
			return fmt.Errorf("%w: %v", projection.ErrRejected, err)
		}
	} else {
		var p models.MongoProductBarcodeModel
		if err := json.Unmarshal([]byte(msg), &p); err != nil {
			return fmt.Errorf("%w: %v", projection.ErrRejected, err)
		}
		signals = []models.MongoProductBarcodeModel{p}
	}
	if len(signals) == 0 {
		return nil
	}
	holding, _, err := normalizeProductBarcodeBatch(signals, false)
	if err != nil {
		return err
	}
	db, source, err := projectionRuntime(holding)
	if err != nil {
		return err
	}
	for start := 0; start < len(signals); start += BATCH_SIZE {
		end := min(start+BATCH_SIZE, len(signals))
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err = reconcileBarcodeSignals(ctx, db, source, signals[start:end])
		cancel()
		if err != nil {
			return err
		}
	}
	return nil
}
