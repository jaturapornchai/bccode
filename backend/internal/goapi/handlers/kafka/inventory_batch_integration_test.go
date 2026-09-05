//go:build integration

package kafka

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"smlcloudplatform/internal/goapi/models"
)

func TestBarcodeBatchIntegration(t *testing.T) {
	dsn := os.Getenv("BC_BARCODE_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_BARCODE_TEST_POSTGRES_DSN to an isolated PostgreSQL instance")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	schema := "bc_barcode_test_" + primitive.NewObjectID().Hex()
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	defer func() {
		clean, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if _, err := db.ExecContext(clean, "DROP SCHEMA "+pq.QuoteIdentifier(schema)+" CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	if _, err := db.ExecContext(ctx, "SET search_path TO "+pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "CREATE TABLE productbarcode (holding_code text, businesscode text, barcode text, price1 numeric(20,4) NOT NULL CHECK(price1>=0), PRIMARY KEY(holding_code,businesscode,barcode))"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO productbarcode VALUES ('H','A','SKU','0.1000'),('H','B','SKU','0.2000')"); err != nil {
		t.Fatal(err)
	}
	total := func() string {
		var value string
		if err := db.QueryRowContext(ctx, "SELECT SUM(price1)::text FROM productbarcode").Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	if total() != "0.3000" {
		t.Fatal("initial exact reconciliation failed")
	}
	docs := []models.MongoProductBarcodeModel{{HoldingCode: "H", BusinessCode: "A", Barcode: "SKU"}}
	columns := []string{"holding_code", "businesscode", "barcode", "price1"}
	if err := replaceProductBarcodeRows(ctx, db, docs, columns, [][]any{{"H", "A", "SKU", "-1"}}); err == nil {
		t.Fatal("invalid copy accepted")
	}
	if total() != "0.3000" {
		t.Fatal("failed COPY removed original rows")
	}
	for i := 0; i < 2; i++ {
		if err := replaceProductBarcodeRows(ctx, db, docs, columns, [][]any{{"H", "A", "SKU", "0.3000"}}); err != nil {
			t.Fatal(err)
		}
		if total() != "0.5000" {
			t.Fatal("duplicate replay changed exact reconciliation")
		}
		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM productbarcode").Scan(&count); err != nil || count != 2 {
			t.Fatalf("unexpected row count %d: %v", count, err)
		}
	}
	var other string
	if err := db.QueryRowContext(ctx, "SELECT price1::text FROM productbarcode WHERE businesscode='B'").Scan(&other); err != nil || other != "0.2000" {
		t.Fatal("other company changed")
	}
}
