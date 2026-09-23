package fixedasset

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
)

// Connector opens the PostgreSQL database of one holding.
type Connector func(holding string) (*sql.DB, error)

const (
	kindAsset        = "assets"
	kindType         = "types"
	kindDepreciation = "depreciations"
	kindDisposal     = "disposals"
)

// Fixed-asset documents share one table per holding database, mirroring gl_records:
// key columns for lookups, the full document in payload.
const recordsSchema = `
CREATE TABLE IF NOT EXISTS fa_records (
	company    text        NOT NULL,
	kind       text        NOT NULL,
	id         text        NOT NULL,
	code       text        NOT NULL,
	payload    jsonb       NOT NULL,
	updated_at timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (company, kind, id)
);
CREATE INDEX IF NOT EXISTS idx_fa_records_code ON fa_records (company, kind, code);
CREATE INDEX IF NOT EXISTS idx_fa_records_asset_code ON fa_records (company, kind, (payload->>'assetcode'));`

type records struct {
	connect Connector
	ready   sync.Map
}

func newRecords(connect Connector) *records {
	return &records{connect: connect}
}

func (r *records) db(ctx context.Context, holding string) (*sql.DB, error) {
	db, err := r.connect(holding)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, fmt.Errorf("ไม่พบฐานข้อมูลของกลุ่มบริษัท")
	}
	if _, ok := r.ready.Load(holding); ok {
		return db, nil
	}
	if _, err := db.ExecContext(ctx, recordsSchema); err != nil {
		return nil, err
	}
	r.ready.Store(holding, true)
	return db, nil
}

type queryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func putRecord(ctx context.Context, q queryer, company, kind, id, code string, doc any) error {
	payload, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	_, err = q.ExecContext(ctx, `INSERT INTO fa_records (company, kind, id, code, payload, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (company, kind, id) DO UPDATE SET code = EXCLUDED.code, payload = EXCLUDED.payload, updated_at = now()`,
		company, kind, id, code, payload)
	return err
}

// queryRecords returns live (not soft-deleted) documents; where/order start with a space and use $3 onwards.
func queryRecords[T any](ctx context.Context, q queryer, company, kind, where, order string, args ...any) ([]T, error) {
	query := `SELECT payload FROM fa_records WHERE company = $1 AND kind = $2 AND NOT COALESCE((payload->>'isdeleted')::boolean, false)` + where + order
	rows, err := q.QueryContext(ctx, query, append([]any{company, kind}, args...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []T{}
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var item T
		if err := json.Unmarshal(payload, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func firstRecord[T any](ctx context.Context, q queryer, company, kind, where string, args ...any) (*T, error) {
	items, err := queryRecords[T](ctx, q, company, kind, where, " LIMIT 1", args...)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	return &items[0], nil
}

func countRecords(ctx context.Context, q queryer, company, kind, where string, args ...any) (int64, error) {
	rows, err := q.QueryContext(ctx, `SELECT count(*) FROM fa_records WHERE company = $1 AND kind = $2 AND NOT COALESCE((payload->>'isdeleted')::boolean, false)`+where, append([]any{company, kind}, args...)...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var total int64
	if rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, err
		}
	}
	return total, rows.Err()
}

// lockRecord reads one live document and locks its row until the transaction ends.
func lockRecord[T any](ctx context.Context, q queryer, company, kind, where string, args ...any) (*T, error) {
	items, err := queryRecords[T](ctx, q, company, kind, where, " LIMIT 1 FOR UPDATE", args...)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	return &items[0], nil
}
