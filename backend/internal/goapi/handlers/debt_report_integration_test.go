//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"smlcloudplatform/internal/generalledger"
)

// TestDebtReportsNeverMixCompanies - รายงานเจ้าหนี้/ลูกหนี้ทั้ง 8 รายงานบน PostgreSQL จริง:
// ผู้ใช้บริษัท 02 ต้องไม่เห็นเอกสารของบริษัท 01 (เดิมมี OR d.company = '01') และกลับกัน
func TestDebtReportsNeverMixCompanies(t *testing.T) {
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	if err := generalledger.EnsureSchema(ctx, db); err != nil {
		t.Fatal(err)
	}

	type companyData struct{ company, partner, marker string }
	companies := []companyData{
		{company: "01", partner: "DRSCOPE-P01", marker: "DRSCOPE-01"},
		{company: "02", partner: "DRSCOPE-P02", marker: "DRSCOPE-02"},
	}
	cleanup := func() {
		for _, c := range companies {
			if _, err := db.Exec(`DELETE FROM gl_subledger_documents WHERE company = $1 AND partner_code = $2`, c.company, c.partner); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`DELETE FROM gl_subledger_partners WHERE company = $1 AND code = $2`, c.company, c.partner); err != nil {
				t.Fatal(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)

	for _, c := range companies {
		partner := fmt.Sprintf(`{"name":"บริษัท ศรีสยามค้าส่ง %s จำกัด","creditday":30,"creditmoney":"50000"}`, c.marker)
		if _, err := db.Exec(`INSERT INTO gl_subledger_partners(company, code, version, payload) VALUES($1, $2, 1, $3::jsonb)`, c.company, c.partner, partner); err != nil {
			t.Fatal(err)
		}
		// side: ลูกหนี้ 1 = ตั้งหนี้ 2 = รับชำระ / เจ้าหนี้ 2 = ตั้งหนี้ 1 = จ่ายชำระ
		docs := []struct {
			ledger string
			side   int
			suffix string
			amount string
		}{
			{"ar", 1, "IV", "10700.00"},
			{"ar", 2, "RE", "1000.00"},
			{"ap", 2, "PI", "5350.00"},
			{"ap", 1, "PS", "500.00"},
		}
		for _, d := range docs {
			id := c.marker + "-" + d.ledger + "-" + d.suffix
			payload := fmt.Sprintf(`{"doc_no":%q,"doc_date":"2026-09-15"}`, id)
			if _, err := db.Exec(`INSERT INTO gl_subledger_documents(company, id, ledger, partner_code, branch_code, control_account_code, amount, side, version, created_journal_id, payload)
VALUES($1, $2, $3, $4, '00000', '11300', $5::numeric, $6, 1, $7, $8::jsonb)`, c.company, id, d.ledger, c.partner, d.amount, d.side, "J-"+id, payload); err != nil {
				t.Fatal(err)
			}
		}
	}

	from := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for _, code := range debtLedgerReportCodes {
		for i, own := range companies {
			other := companies[1-i]
			t.Run(code+"/"+own.company, func(t *testing.T) {
				query, args, err := BuildDebtReportQuery(code, own.company, from, to, "", 1000, 0)
				if err != nil {
					t.Fatal(err)
				}
				rows, err := db.QueryContext(ctx, query, args...)
				if err != nil {
					t.Fatal(err)
				}
				defer rows.Close()
				columns, err := rows.Columns()
				if err != nil {
					t.Fatal(err)
				}
				ownRows := 0
				for rows.Next() {
					values := make([]any, len(columns))
					ptrs := make([]any, len(columns))
					for i := range values {
						ptrs[i] = &values[i]
					}
					if err := rows.Scan(ptrs...); err != nil {
						t.Fatal(err)
					}
					var text strings.Builder
					for _, v := range values {
						if b, ok := v.([]byte); ok {
							v = string(b)
						}
						fmt.Fprint(&text, v, "|")
					}
					line := text.String()
					if strings.Contains(line, other.marker) || strings.Contains(line, other.partner) {
						t.Fatalf("company %s report shows company %s data: %s", own.company, other.company, line)
					}
					if strings.Contains(line, own.partner) {
						ownRows++
					}
				}
				if err := rows.Err(); err != nil {
					t.Fatal(err)
				}
				if ownRows == 0 {
					t.Fatalf("company %s report has no rows of its own partner %s", own.company, own.partner)
				}
			})
		}
	}
}
