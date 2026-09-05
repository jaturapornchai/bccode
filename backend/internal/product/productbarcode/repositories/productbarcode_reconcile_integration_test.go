//go:build integration

package repositories_test

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/product/projection"
	"smlcloudplatform/pkg/microservice"
)

func TestLegacyBarcodeReconcileIntegration(t *testing.T) {
	dsn := os.Getenv("BC_BARCODE_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Close() })
	schema := "bc_legacy_barcode_test_" + primitive.NewObjectID().Hex()
	if _, err := admin.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + pq.QuoteIdentifier(schema) + " CASCADE"); err != nil {
			t.Error(err)
		}
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	pool, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	ddl := "CREATE TABLE productbarcode (holding_code text,businesscode text,barcode text,parid text,itemcode text,unitcode text,standvalue bigint,dividevalue bigint,itemtype integer,materialtype integer,balanceqty numeric(20,4) NOT NULL DEFAULT 0,balanceamount numeric(20,4) NOT NULL DEFAULT 0,averagecost numeric(20,4) NOT NULL DEFAULT 0"
	for _, field := range []string{"brandcode", "categorycode", "classcode", "designcode", "gradecode", "groupcode", "groupsubonecode", "groupsubtwocode", "modelcode", "patterncode"} {
		ddl += "," + pq.QuoteIdentifier(field) + " text"
	}
	for _, field := range []string{"names", "unitnames", "brandnames", "categorynames", "classnames", "designnames", "gradenames", "groupnames", "groupsubonenames", "groupsubtwonames", "modelnames", "patternnames", "bom"} {
		ddl += "," + pq.QuoteIdentifier(field) + " jsonb"
	}
	ddl += ", PRIMARY KEY(holding_code,businesscode,barcode))"
	if _, err := pool.ExecContext(ctx, ddl); err != nil {
		t.Fatal(err)
	}
	orm, err := gorm.Open(postgres.New(postgres.Config{Conn: pool}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	repo := repositories.NewProductBarcodePGRepository(microservice.NewPersisterWithDB(orm))
	load := func(context.Context) (*models.ProductBarcodePg, error) {
		return &models.ProductBarcodePg{UnitCode: "CURRENT", ItemCode: "SKU", StandValue: 1, DivideValue: 1}, nil
	}
	if _, err := repo.ReconcileInCompany(ctx, "H", "A", "BAR", load); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.ExecContext(ctx, "UPDATE productbarcode SET balanceqty='12.3400',balanceamount='56.7800',averagecost='4.6000'"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ReconcileInCompany(ctx, "H", "A", "BAR", load); err != nil {
		t.Fatal(err)
	}
	var totals string
	if err := pool.QueryRowContext(ctx, "SELECT balanceqty::text||','||balanceamount::text||','||averagecost::text FROM productbarcode WHERE holding_code='H' AND businesscode='A' AND barcode='BAR'").Scan(&totals); err != nil || totals != "12.3400,56.7800,4.6000" {
		t.Fatalf("replay changed accounting values: %q %v", totals, err)
	}
	if _, err := repo.ReconcileInCompany(ctx, "H", "B", "BAR", load); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("source unavailable")
	if _, err := repo.ReconcileInCompany(ctx, "H", "A", "BAR", func(context.Context) (*models.ProductBarcodePg, error) { return nil, failure }); !errors.Is(err, failure) {
		t.Fatal("source failure hidden")
	}
	// A GoAPI/rebuild lock and the legacy GORM repository must coordinate before
	// loading source data, not just serialize writes after stale reads.
	tx, err := pool.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, projection.ExclusiveSQL, projection.LockKey("barcode", "H", "A", "")); err != nil {
		t.Fatal(err)
	}
	lockCtx, stop := context.WithTimeout(ctx, 150*time.Millisecond)
	called := false
	_, err = repo.ReconcileInCompany(lockCtx, "H", "A", "BAR", func(context.Context) (*models.ProductBarcodePg, error) { called = true; return load(ctx) })
	stop()
	tx.Rollback()
	if err == nil || called {
		t.Fatalf("read source before shared company lock: called=%v error=%v", called, err)
	}
	if _, err := repo.ReconcileInCompany(ctx, "H", "A", "BAR", func(context.Context) (*models.ProductBarcodePg, error) { return nil, nil }); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := pool.QueryRowContext(ctx, "SELECT count(*) FROM productbarcode WHERE businesscode='A'").Scan(&count); err != nil || count != 0 {
		t.Fatal("deleted source remains projected")
	}
	if err := pool.QueryRowContext(ctx, "SELECT count(*) FROM productbarcode WHERE businesscode='B'").Scan(&count); err != nil || count != 1 {
		t.Fatal("other company was modified")
	}
}
