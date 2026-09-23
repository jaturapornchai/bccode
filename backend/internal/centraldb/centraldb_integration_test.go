//go:build integration

package centraldb_test

import (
	"context"
	"testing"

	"smlcloudplatform/internal/centraldb"
	"smlcloudplatform/internal/centraldb/centraldbtest"
)

func TestEnsureSchemaIsIdempotent(t *testing.T) {
	db := centraldbtest.New(t)
	if err := centraldb.EnsureSchema(context.Background(), db); err != nil {
		t.Fatalf("second EnsureSchema: %v", err)
	}
	var tables int
	if err := db.QueryRow(`SELECT count(*) FROM information_schema.tables WHERE table_schema='public'`).Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if tables != 11 {
		t.Fatalf("tables = %d", tables)
	}
}

// TestEnsureSchemaAddsBusinessCodeToOldAccessLog - ตาราง access log รุ่นเก่า (ไม่มี business_code) บน prod
// ต้องได้คอลัมน์เพิ่มตอน start ไม่งั้นเลือกบริษัทแล้ว log error ทุกครั้ง
func TestEnsureSchemaAddsBusinessCodeToOldAccessLog(t *testing.T) {
	db := centraldbtest.New(t)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `DROP TABLE shop_user_access_logs`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE shop_user_access_logs (
		id BIGSERIAL PRIMARY KEY, holding_code TEXT NOT NULL, username TEXT NOT NULL DEFAULT '',
		ip TEXT NOT NULL DEFAULT '', last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	if err := centraldb.EnsureSchema(ctx, db); err != nil {
		t.Fatalf("EnsureSchema on old table: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO shop_user_access_logs (holding_code, business_code, username, ip) VALUES ('h1', '01', 'u1', '127.0.0.1')`); err != nil {
		t.Fatalf("insert with business_code: %v", err)
	}
}
