//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/shopspring/decimal"
)

// TestTaxWithholdingReportFromGL ตรวจรายงานภาษีหัก ณ ที่จ่ายกับฐานทดสอบที่มีข้อมูล GL จริง
// (seed ผ่าน backend/cmd/glseed — ใบจ่ายเจ้าหนี้หัก 3% และใบค่าเช่าหัก 5% ลงบัญชี ภ.ง.ด.53)
//
// ตั้งสนามทดสอบ:
//
//	docker run -d --name bc-gl-wht-test-pg18 -e POSTGRES_HOST_AUTH_METHOD=trust \
//	  -e POSTGRES_DB=restore -p 127.0.0.1:55443:5432 postgres:18-alpine
//	(restore ฐาน rungrueng + รัน glseed -apply ตาม docs)
//	BC_GL_TEST_POSTGRES_DSN='postgres://postgres@127.0.0.1:55443/rungrueng?sslmode=disable'
func TestTaxWithholdingReportFromGL(t *testing.T) {
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN to a GL test database with seeded WHT journals")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	// 1) paid ทุกแบบยื่น งวด ม.ค.: ใบจ่ายเจ้าหนี้หัก 3% (3,210/107,000)
	report, err := buildWithholdingReport(ctx, db, "01", 2026, 1, "paid", nil)
	if err != nil {
		t.Fatalf("paid report: %v", err)
	}
	rows := report.Rows
	if len(rows) != 1 {
		t.Fatalf("paid rows=%d คาด 1 (rows=%+v)", len(rows), rows)
	}
	if report.Summary.BaseTotal != "107000.00" || report.Summary.WhtTotal != "3210.00" {
		t.Fatalf("paid totals base=%s wht=%s คาด 107000.00/3210.00", report.Summary.BaseTotal, report.Summary.WhtTotal)
	}
	if report.Summary.NetTotal != "103790.00" || rows[0].NetAmount != "103790.00" {
		t.Fatalf("paid net total=%s row=%s คาด 103790.00", report.Summary.NetTotal, rows[0].NetAmount)
	}
	if report.Summary.WhtTotalText != "สามพันสองร้อยสิบบาทถ้วน" || report.Summary.PayeeCount != 1 {
		t.Fatalf("paid summary text=%q payees=%d", report.Summary.WhtTotalText, report.Summary.PayeeCount)
	}
	if len(report.Summary.ByRate) != 1 || report.Summary.ByRate[0].RatePercent != "3.00" {
		t.Fatalf("paid byrate=%+v คาด 1 กลุ่ม 3.00%%", report.Summary.ByRate)
	}
	if report.Note != "" {
		t.Fatalf("ไม่ควรมี note เมื่อพบบัญชี: %s", report.Note)
	}

	pv1, ok := rows[0], rows[0].DocNo == "PV6901-S001"
	if !ok {
		t.Fatalf("แถวแรกไม่ใช่ PV6901-S001: %+v", rows[0])
	}
	if pv1.PartnerCode != "SUPP-TH-001" || pv1.TaxID != "0105558002001" {
		t.Fatalf("PV6901-S001 partner=%q taxid=%q", pv1.PartnerCode, pv1.TaxID)
	}
	if pv1.WhtAmount != "3210.00" || pv1.BaseAmount != "107000.00" || pv1.RatePercent != "3.00" {
		t.Fatalf("PV6901-S001 tax=%s base=%s rate=%s คาด 3210.00/107000.00/3.00", pv1.WhtAmount, pv1.BaseAmount, pv1.RatePercent)
	}

	// 1.1) งวด ก.พ.: ใบค่าเช่าหัก 5% (1,000/20,000)
	feb, err := buildWithholdingReport(ctx, db, "01", 2026, 2, "paid", nil)
	if err != nil {
		t.Fatalf("paid feb: %v", err)
	}
	if len(feb.Rows) != 1 || feb.Summary.BaseTotal != "20000.00" || feb.Summary.WhtTotal != "1000.00" {
		t.Fatalf("paid feb rows=%d base=%s wht=%s คาด 1/20000.00/1000.00", len(feb.Rows), feb.Summary.BaseTotal, feb.Summary.WhtTotal)
	}

	// 2) กรองเฉพาะแบบยื่น ภ.ง.ด.53 ต้องได้ใบจ่ายเจ้าหนี้นิติบุคคล
	form53, err := buildWithholdingReport(ctx, db, "01", 2026, 1, "paid", []string{"53"})
	if err != nil {
		t.Fatalf("form 53: %v", err)
	}
	if len(form53.Rows) != 1 || form53.Summary.WhtTotal != "3210.00" {
		t.Fatalf("form 53 rows=%d wht=%s คาด 1/3210.00", len(form53.Rows), form53.Summary.WhtTotal)
	}

	// 3) ภ.ง.ด.3 ในงวดนี้ไม่มีรายการ (แถวร่างจะไม่เข้า GL lines) — คืนว่างโดยไม่ error
	form3, err := buildWithholdingReport(ctx, db, "01", 2026, 1, "paid", []string{"3"})
	if err != nil {
		t.Fatalf("form 3: %v", err)
	}
	if len(form3.Rows) != 0 || form3.Summary.WhtTotal != "0.00" {
		t.Fatalf("form 3 rows=%d wht=%s คาด 0/0.00", len(form3.Rows), form3.Summary.WhtTotal)
	}

	// 4) received: ยังไม่มีใบที่เราถูกหักในงวดนี้ — ว่างโดยไม่ error
	received, err := buildWithholdingReport(ctx, db, "01", 2026, 1, "received", nil)
	if err != nil {
		t.Fatalf("received: %v", err)
	}
	if len(received.Rows) != 0 {
		t.Fatalf("received rows=%d คาด 0", len(received.Rows))
	}

	// 5) แบบยื่นนอก whitelist ต้องถูกปฏิเสธ ไม่ใช่รัน query
	if _, err := buildWithholdingReport(ctx, db, "01", 2026, 1, "paid", []string{"99"}); err == nil {
		t.Fatal("ระบบยอมรับแบบยื่นที่ไม่อยู่ใน whitelist")
	}

	// 6) paging ตัดหน้าแต่ยอดรวมยังเป็นทั้งงวด
	if page := pageWithholdingRows(rows, 1, 5); len(page) != 0 {
		t.Fatalf("offset เกินจำนวนแถวต้องได้หน้าว่าง: %+v", page)
	}
}

// TestTaxVatQueriesRunOnRealSchema ยิง query ยอดรวมภาษีซื้อ/ขาย และ ภ.พ.30 กับ DDL จริงของตาราง ERP
// (ตาราง ERP มีเฉพาะใน bcai_projection ไม่มีในฐาน Holding — ดู docs/kms/bugs/2026-09-23-vat-report-reads-missing-erp-tables.md)
// (คอลัมน์ยอดเงิน ERP ยังเป็น double precision — ต้องแปลงเป็น numeric ได้โดยไม่ error)
func TestTaxVatQueriesRunOnRealSchema(t *testing.T) {
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN to a restored Holding database")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	// ยอดขายทดสอบ (DSN ต้องมีแถว double: 0.1/0.2/0.3 และ 100.005/7.0035/107.0085 เดือน ม.ค. 2026)
	// ปัดต่อใบก่อนรวม: 0.10+100.01 / 0.20+7.00 / 0.30+107.01 — ไม่มีเศษ float หลุดออกมา
	query, args := buildVatRegisterSummaryQuery("sale", 2026, 1)
	var count int
	var before, vat, total decimal.Decimal
	if err := db.QueryRowContext(ctx, query, args...).Scan(&count, &before, &vat, &total); err != nil {
		t.Fatalf("sale summary: %v", err)
	}
	if got := []string{moneyText(before), moneyText(vat), moneyText(total)}; count != 2 || got[0] != "100.11" || got[1] != "7.20" || got[2] != "107.31" {
		t.Fatalf("sale summary count=%d amounts=%v คาด 2 [100.11 7.20 107.31]", count, got)
	}
	query, args = buildVatRegisterSummaryQuery("purchase", 2026, 1)
	if err := db.QueryRowContext(ctx, query, args...).Scan(&count, &before, &vat, &total); err != nil {
		t.Fatalf("purchase summary: %v", err)
	}
	for name, build := range map[string]func(int, int) (string, []any){"sales": buildPP30SalesQuery, "purchase": buildPP30PurchaseQuery} {
		query, args := build(2026, 1)
		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			t.Fatalf("pp30 %s: %v", name, err)
		}
		rows.Close()
	}
}
