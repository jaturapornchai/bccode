//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"smlcloudplatform/internal/rdform"
)

// A tax_filings table created with the old VARCHAR(20)/(100) columns is widened to TEXT by
// ensureTaxFilingSchema, keeps its rows, and then accepts a long company code and a long
// username/email; running the schema step again changes nothing.
//
//	BC_TAXFORM_TEST_POSTGRES_DSN='postgres://postgres@127.0.0.1:5432/taxform_it?sslmode=disable'
func TestTaxFilingColumnsWidenToText(t *testing.T) {
	dsn := os.Getenv("BC_TAXFORM_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_TAXFORM_TEST_POSTGRES_DSN")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	namespace := fmt.Sprintf("tax_width_%d", time.Now().UnixNano())
	if _, err = admin.Exec("CREATE SCHEMA " + namespace); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", namespace)
	u.RawQuery = query.Encode()
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close(); admin.Exec("DROP SCHEMA " + namespace + " CASCADE"); admin.Close() })
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `CREATE TABLE tax_filings (
    id BIGSERIAL PRIMARY KEY, company_code VARCHAR(20) NOT NULL, form_code VARCHAR(30) NOT NULL,
    period_year INT NOT NULL, period_month SMALLINT NOT NULL, filing_seq SMALLINT NOT NULL DEFAULT 0,
    document JSONB NOT NULL, version INT NOT NULL DEFAULT 1,
    created_by VARCHAR(100) NOT NULL DEFAULT '', created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(100) NOT NULL DEFAULT '', updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_tax_filings_period UNIQUE (company_code, form_code, period_year, period_month, filing_seq));
CREATE TABLE tax_filing_history (id BIGSERIAL PRIMARY KEY, filing_id BIGINT NOT NULL, version INT NOT NULL,
    document JSONB NOT NULL, saved_by VARCHAR(100) NOT NULL DEFAULT '', saved_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP);
INSERT INTO tax_filings (company_code, form_code, period_year, period_month, document, created_by, updated_by)
    VALUES ('01', 'pp30', 2026, 8, '{"values":{"output_tax":"70.00"}}', 'demo', 'demo');`)
	if err != nil {
		t.Fatal(err)
	}

	for run := 1; run <= 2; run++ {
		if err = ensureTaxFilingSchema(ctx, db); err != nil {
			t.Fatalf("schema run %d: %v", run, err)
		}
		rows, err := db.QueryContext(ctx, `SELECT table_name || '.' || column_name, data_type FROM information_schema.columns
			WHERE table_schema = current_schema() AND column_name IN ('company_code', 'created_by', 'updated_by', 'saved_by')`)
		if err != nil {
			t.Fatal(err)
		}
		types := map[string]string{}
		for rows.Next() {
			var column, dataType string
			if err = rows.Scan(&column, &dataType); err != nil {
				t.Fatal(err)
			}
			types[column] = dataType
		}
		rows.Close()
		if len(types) != 4 {
			t.Fatalf("run %d: columns = %v", run, types)
		}
		for column, dataType := range types {
			if dataType != "text" {
				t.Fatalf("run %d: %s is %s, want text", run, column, dataType)
			}
		}
	}

	var kept string
	if err = db.QueryRowContext(ctx, `SELECT document->'values'->>'output_tax' FROM tax_filings WHERE company_code='01' AND period_month=8`).Scan(&kept); err != nil || kept != "70.00" {
		t.Fatalf("existing filing after widening = %q (%v)", kept, err)
	}

	company := "company-" + strings.Repeat("x", 40)
	user := strings.Repeat("u", 120) + "@example.co.th"
	doc := rdform.Document{Values: map[string]string{"filing_type": "normal", "output_tax": "70.00"}}
	f := TaxFiling{Code: "pp30", Year: 2026, Month: 9, Document: &doc}
	if err = saveTaxFiling(ctx, db, company, user, &f); err != nil {
		t.Fatalf("save with long company code/username: %v", err)
	}
	var storedCompany, createdBy, savedBy string
	if err = db.QueryRowContext(ctx, `SELECT f.company_code, f.created_by, h.saved_by FROM tax_filings f
		JOIN tax_filing_history h ON h.filing_id = f.id WHERE f.id = $1`, f.ID).Scan(&storedCompany, &createdBy, &savedBy); err != nil {
		t.Fatal(err)
	}
	if storedCompany != company || createdBy != user || savedBy != user {
		t.Fatalf("stored %q/%q/%q, want the full values", storedCompany, createdBy, savedBy)
	}
}
