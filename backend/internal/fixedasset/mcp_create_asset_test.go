package fixedasset_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	fa "smlcloudplatform/internal/fixedasset"
	"smlcloudplatform/internal/fixedasset/mcp"
)

// An agent creating a passenger car through fa_create_asset must get the tax cap too.
// Lives in this test binary (not package mcp) so fa_records is created by one process only.
func TestMCPCreateAssetPassengerCarTaxCap(t *testing.T) {
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN to isolated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	company := fmt.Sprintf("FA_MCP_TEST_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM fa_records WHERE company = $1`, company)
		_ = db.Close()
	})
	connect := func(string) (*sql.DB, error) { return db, nil }
	ctx := context.Background()
	scope := fa.Scope{Holding: "TEST_HOLDING", Company: company, Branch: "HQ", Actor: "MCP_AGENT"}

	h := mcp.NewMCPHandler(fa.NewStore(connect), nil, fa.NewReporter(connect))
	created, err := h.HandleToolCall(ctx, scope, "fa_create_asset", map[string]interface{}{
		"assetcode": "CAR-MCP", "name_th": "รถยนต์นั่งผู้บริหาร", "assettypecode": "VEHICLE_PASSENGER",
		"cost": 1500000.0, "usefullifeyears": 5.0, "deprecpercent": 20.0,
		"purchasedate": "2026-01-01", "passengercartaxcap": true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if asset := created.(*fa.Asset); !asset.PassengerCarTaxCap {
		t.Fatalf("passengercartaxcap was dropped: %+v", asset)
	}

	rep, err := fa.NewReporter(connect).GetTaxReconciliationReport(ctx, scope, "2026")
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Rows) != 1 {
		t.Fatalf("want 1 report row, got %+v", rep.Rows)
	}
	book := decimal.RequireFromString(rep.Rows[0]["accounting_deprec"])
	tax := decimal.RequireFromString(rep.Rows[0]["tax_deprec"])
	if !tax.LessThan(book) || tax.Sub(decimal.RequireFromString("200000")).Abs().GreaterThan(decimal.RequireFromString("0.05")) {
		t.Fatalf("MCP-created passenger car must be capped at 20%% of 1,000,000: book %s tax %s", book, tax)
	}
}
