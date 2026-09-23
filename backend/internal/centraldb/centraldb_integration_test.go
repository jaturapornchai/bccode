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
