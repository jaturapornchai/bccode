//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/doctype"
)

// TestSalesReportsNeverMixCompanies - รายงานขายตามเอกสาร/กำไรขั้นต้น/สรุปยอดขาย บน PostgreSQL จริง (SQL เดียวกับ handler):
// เดิม join public.debtor ที่ฐานปัจจุบันไม่มี (ล้มทั้งรายงาน) และ join ด้วยรหัสลูกค้าอย่างเดียวไม่มีบริษัท
// บริษัท 01/02 ใช้เลขที่เอกสารและรหัสลูกค้าเดียวกัน: แต่ละบริษัทต้องเห็นแค่ชื่อลูกค้าและยอดของตัวเอง ยอดเงินเป็นข้อความทศนิยมตรงตัว
// query อ้าง public.* ตรง ๆ จึงสร้างฐานแยก sales_scope_it (ไม่ชนตาราง doc/stock_ledger จริงของแพ็กเกจอื่นในฐานทดสอบเดียวกัน)
func TestSalesReportsNeverMixCompanies(t *testing.T) {
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN")
	}
	ctx := context.Background()
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Close() })
	dropDatabase := func() {
		if _, err := admin.Exec(`DROP DATABASE IF EXISTS sales_scope_it WITH (FORCE)`); err != nil {
			t.Fatal(err)
		}
	}
	dropDatabase()
	if _, err := admin.Exec(`CREATE DATABASE sales_scope_it`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(dropDatabase)

	scoped, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	scoped.Path = "/sales_scope_it"
	db, err := sql.Open("postgres", scoped.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	// คอลัมน์ที่รายงานอ่าน ชนิดเดียวกับ transaction/httpapi/schema.go (doc, docdetail, stock_ledger)
	if _, err := db.Exec(`
CREATE TABLE public.doc (businesscode text NOT NULL DEFAULT '', transflag integer, docno text, custcode text,
	custname text NOT NULL DEFAULT '', isdelete boolean NOT NULL DEFAULT false);
CREATE TABLE public.docdetail (businesscode text NOT NULL DEFAULT '', transflag integer, docno text, linenumber integer,
	itemcode text, description text);
CREATE TABLE public.stock_ledger (
	businesscode text NOT NULL, itemcode text NOT NULL, whcode text NOT NULL, docdatetime timestamptz NOT NULL,
	docno text NOT NULL, linenumber integer NOT NULL, barcode text NOT NULL DEFAULT '', unitcode text NOT NULL DEFAULT '',
	transflag integer NOT NULL, docqty numeric(18,8) NOT NULL DEFAULT 0, price numeric(18,8) NOT NULL DEFAULT 0,
	docvalue numeric(18,8) NOT NULL DEFAULT 0, unitcost numeric(18,8) NOT NULL, amount numeric(18,8) NOT NULL);`); err != nil {
		t.Fatal(err)
	}

	at := time.Date(2026, 9, 15, 3, 0, 0, 0, time.UTC)
	type line struct{ value, cost string }
	type companyData struct {
		custname, item              string
		lines                       []line
		total, cost, profit, margin string
	}
	companies := map[string]companyData{
		// 0.10 + 0.20 = 0.30 ตรงตัว (float64 ได้ 0.30000000000000004)
		"01": {custname: "ร้านมิตรภาพค้าส่ง", item: "น้ำดื่มตราช้าง 600 มล. แพ็ก 12 ขวด",
			lines: []line{{"0.10", "0.05"}, {"0.20", "0.05"}}, total: "0.30", cost: "0.10", profit: "0.20", margin: "66.67"},
		"02": {custname: "ห้างหุ้นส่วนจำกัด ชัยเจริญพาณิชย์", item: "น้ำดื่มตราช้าง 600 มล. ลังละ 24 ขวด",
			lines: []line{{"1070.00", "700.00"}}, total: "1070.00", cost: "700.00", profit: "370.00", margin: "34.58"},
	}
	const docNo, custCode = "IV6909-0001", "AR-0001"
	for company, c := range companies {
		if _, err := db.Exec(`INSERT INTO doc(businesscode, transflag, docno, custcode, custname) VALUES($1, $2, $3, $4, $5)`,
			company, doctype.SaleInvoice, docNo, custCode, c.custname); err != nil {
			t.Fatal(err)
		}
		for i, l := range c.lines {
			if _, err := db.Exec(`INSERT INTO docdetail(businesscode, transflag, docno, linenumber, itemcode, description) VALUES($1, $2, $3, $4, 'WTR-600', $5)`,
				company, doctype.SaleInvoice, docNo, i+1, c.item); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`INSERT INTO stock_ledger(businesscode, itemcode, whcode, docdatetime, docno, linenumber, unitcode, transflag,
				docqty, price, docvalue, unitcost, amount) VALUES($1, 'WTR-600', 'WH-01', $2, $3, $4, 'ขวด', $5, 1, $6::numeric, $6::numeric, $7::numeric, $7::numeric)`,
				company, at, docNo, i+1, doctype.SaleInvoice, l.value, l.cost); err != nil {
				t.Fatal(err)
			}
		}
	}
	// เอกสารคนละประเภท (ซื้อ) เลขที่เดียวกันในบริษัท 01: join หัวเอกสารต้องตรงประเภทด้วย ไม่งั้นแถวสต็อกถูกนับซ้ำ
	if _, err := db.Exec(`INSERT INTO doc(businesscode, transflag, docno, custcode, custname) VALUES('01', $1, $2, 'AP-0001', 'บริษัท สยามเครื่องดื่ม จำกัด')`,
		doctype.Purchase, docNo); err != nil {
		t.Fatal(err)
	}

	from, to := at.AddDate(0, 0, -1), at.AddDate(0, 0, 1)
	for company, want := range companies {
		query, args := buildSalesReportHeaderQuery(from, to, company, nil, nil, true, 100, 0)
		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			t.Fatalf("company %s header: %v", company, err)
		}
		header, err := scanSalesReportRows(rows)
		rows.Close()
		if err != nil {
			t.Fatalf("company %s header rows: %v", company, err)
		}
		if len(header) != 1 {
			t.Fatalf("company %s header = %+v, want one document", company, header)
		}
		got := header[0]
		if got["docno"] != docNo || got["debtorcode"] != custCode || got["debtorname"] != want.custname ||
			got["totalamount"] != want.total || got["calcamount"] != want.cost || got["grossprofit"] != want.profit || got["marginpercent"] != want.margin {
			t.Errorf("company %s header = %+v, want %s %s %s/%s/%s %s%%", company, got, docNo, want.custname, want.total, want.cost, want.profit, want.margin)
		}
		if _, ok := got["price"]; ok {
			t.Errorf("company %s header still returns AVG price: %+v", company, got)
		}

		query, args = buildSalesReportDetailQuery(from, to, company, nil, nil, true, 100, 0)
		rows, err = db.QueryContext(ctx, query, args...)
		if err != nil {
			t.Fatalf("company %s detail: %v", company, err)
		}
		detail, err := scanSalesReportRows(rows)
		rows.Close()
		if err != nil {
			t.Fatalf("company %s detail rows: %v", company, err)
		}
		if len(detail) != len(want.lines) {
			t.Fatalf("company %s detail = %+v, want %d lines", company, detail, len(want.lines))
		}
		for _, d := range detail {
			if d["itemname"] != want.item || d["unitname"] != "ขวด" {
				t.Errorf("company %s detail line = %+v, want item %q unit ขวด", company, d, want.item)
			}
		}

		query, args = buildSalesReportSummaryQuery(from, to, company, nil, nil)
		var documents int64
		var amount, cost, profit decimal.Decimal
		if err := db.QueryRowContext(ctx, query, args...).Scan(&documents, &amount, &cost, &profit); err != nil {
			t.Fatalf("company %s summary: %v", company, err)
		}
		summary := salesSummaryData(documents, amount, cost, profit)
		if summary["total_documents"] != int64(1) || summary["totalamount"] != want.total || summary["totalcost"] != want.cost ||
			summary["total_profit"] != want.profit || summary["profit_margin"] != want.margin {
			t.Errorf("company %s summary = %+v", company, summary)
		}
	}

	// บริษัทที่ไม่มีการขาย ต้องได้รายงานว่าง ไม่ใช่ยอดบริษัท 01
	query, args := buildSalesReportHeaderQuery(from, to, "03", nil, nil, true, 100, 0)
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		t.Fatal(err)
	}
	header, err := scanSalesReportRows(rows)
	rows.Close()
	if err != nil || len(header) != 0 {
		t.Errorf("company 03 header = %+v, %v; want empty", header, err)
	}
	query, args = buildSalesReportSummaryQuery(from, to, "03", nil, nil)
	var documents int64
	var amount, cost, profit decimal.Decimal
	if err := db.QueryRowContext(ctx, query, args...).Scan(&documents, &amount, &cost, &profit); err != nil {
		t.Fatal(err)
	}
	if summary := salesSummaryData(documents, amount, cost, profit); summary["total_documents"] != int64(0) || summary["totalamount"] != "0.00" {
		t.Errorf("company 03 summary = %+v, want zero", summary)
	}
}
