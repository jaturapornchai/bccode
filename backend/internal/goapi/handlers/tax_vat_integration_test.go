//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"smlcloudplatform/internal/generalledger"
)

// TestVatReportsReadRecordedVat - รายงานภาษีขาย/ซื้อ และ ภ.พ.30 อ่านจากรายละเอียดภาษีมูลค่าเพิ่มของใบสำคัญที่ผ่านบัญชี
// (แทนตาราง ERP ที่ไม่มีในฐาน holding — docs/kms/bugs/2026-09-23-vat-report-reads-missing-erp-tables.md) ใช้ฐานเปล่าได้ (สร้าง schema เอง)
func TestVatReportsReadRecordedVat(t *testing.T) {
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() }) // ปิดหลัง cleanup ข้อมูล (Cleanup ทำงานแบบ LIFO)
	ctx := context.Background()
	if err := generalledger.EnsureSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	const company = "VATREP"
	cleanup := func() {
		if _, err := db.Exec(`DELETE FROM gl_records WHERE company=$1`, company); err != nil {
			t.Fatal(err)
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	insert := func(id, docNo, date, status string, vats string) {
		t.Helper()
		payload := `{"docno":"` + docNo + `","date":"` + date + `","status":"` + status + `","details":{"vats":` + vats + `}}`
		if _, err := db.Exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals',$2,$3,1,$4::jsonb)`, company, id, docNo, payload); err != nil {
			t.Fatal(err)
		}
	}
	const buyer = `"partner_name":"บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด","partner_tax_id":"0105558012349","partner_branch_no":"00000"`
	// UV1: ขายเชื่อ ก.ย. 2569 ฐาน 100,000 VAT 7,000 + ส่งออก 0% 20,000
	insert("J1", "UV1", "2026-09-05", "posted", `[{"id":"S1","tax_type":2,"document_type":1,"tax_invoice_no":"IV6909-001","tax_invoice_date":"2026-09-05","tax_period_year":2026,"tax_period_month":9,`+buyer+`,"base_amount":"100000","zero_rate_amount":"20000","exempt_amount":"0","vat_rate":"7","vat_amount":"7000"}]`)
	// UV2: ใบลดหนี้ รับคืนสินค้า ฐาน 1,000 VAT 70 → หักออกจากยอดขาย
	insert("J2", "UV2", "2026-09-20", "posted", `[{"id":"S2","tax_type":2,"document_type":3,"tax_invoice_no":"CN6909-001","tax_invoice_date":"2026-09-20","original_invoice_no":"IV6909-001","original_invoice_date":"2026-09-05","tax_period_year":2026,"tax_period_month":9,`+buyer+`,"base_amount":"1000","zero_rate_amount":"0","exempt_amount":"0","vat_rate":"7","vat_amount":"70"}]`)
	// SV1: ซื้อ ใช้สิทธิ ก.ย. ฐาน 50,000 VAT 3,500 + ซื้อต้องห้าม (ไม่นับ) + ซื้อใช้สิทธิงวด ต.ค. (ไม่นับ)
	insert("J3", "SV1", "2026-09-10", "posted", `[
		{"id":"P1","tax_type":1,"document_type":1,"tax_invoice_no":"PI-001","tax_invoice_date":"2026-09-10","tax_period_year":2026,"tax_period_month":9,"claim_status":1,`+buyer+`,"base_amount":"50000","zero_rate_amount":"0","exempt_amount":"0","vat_rate":"7","vat_amount":"3500"},
		{"id":"P2","tax_type":1,"document_type":1,"tax_invoice_no":"PI-002","tax_invoice_date":"2026-09-10","tax_period_year":2026,"tax_period_month":9,"claim_status":2,`+buyer+`,"base_amount":"10000","zero_rate_amount":"0","exempt_amount":"0","vat_rate":"7","vat_amount":"700"},
		{"id":"P3","tax_type":1,"document_type":1,"tax_invoice_no":"PI-003","tax_invoice_date":"2026-09-11","tax_period_year":2026,"tax_period_month":10,"claim_status":1,`+buyer+`,"base_amount":"2000","zero_rate_amount":"0","exempt_amount":"0","vat_rate":"7","vat_amount":"140"}]`)
	// UV9 ร่าง + UV8 กลับรายการแล้ว — ต้องไม่อยู่ในรายงาน
	insert("J9", "UV9", "2026-09-25", "draft", `[{"id":"S9","tax_type":2,"document_type":1,"tax_invoice_no":"IV6909-009","tax_invoice_date":"2026-09-25","tax_period_year":2026,"tax_period_month":9,`+buyer+`,"base_amount":"9000","zero_rate_amount":"0","exempt_amount":"0","vat_rate":"7","vat_amount":"630"}]`)
	insert("J8", "UV8", "2026-09-26", "reversed", `[{"id":"S8","tax_type":2,"document_type":1,"tax_invoice_no":"IV6909-008","tax_invoice_date":"2026-09-26","tax_period_year":2026,"tax_period_month":9,`+buyer+`,"base_amount":"8000","zero_rate_amount":"0","exempt_amount":"0","vat_rate":"7","vat_amount":"560"}]`)

	sales, err := generalledger.VatRecordsForPeriod(ctx, db, company, 2026, 9, 2)
	if err != nil {
		t.Fatal(err)
	}
	rows, summary := buildVatRegister(sales)
	if len(rows) != 2 || rows[0].TaxInvoiceNo != "IV6909-001" || rows[0].AmountBeforeVat != "120000.00" || rows[0].VatAmount != "7000.00" || rows[0].TotalAmount != "127000.00" ||
		rows[1].TaxInvoiceNo != "CN6909-001" || rows[1].AmountBeforeVat != "-1000.00" || rows[1].VatAmount != "-70.00" || rows[0].TaxID != "0105558012349" || rows[0].BranchNo != "00000" {
		t.Fatalf("sale register rows = %+v", rows)
	}
	if summary.AmountBeforeVat != "119000.00" || summary.VatAmount != "6930.00" || summary.TotalAmount != "125930.00" {
		t.Fatalf("sale register summary = %+v", summary)
	}

	purchases, err := generalledger.VatRecordsForPeriod(ctx, db, company, 2026, 9, 1)
	if err != nil {
		t.Fatal(err)
	}
	if rows, summary := buildVatRegister(purchases); len(rows) != 1 || rows[0].TaxInvoiceNo != "PI-001" || summary.VatAmount != "3500.00" {
		t.Fatalf("purchase register = %+v %+v", rows, summary)
	}

	totals := sumPP30(sales, purchases)
	got := []string{moneyText(totals.salesTaxable), moneyText(totals.salesZeroRated), moneyText(totals.salesExempt), moneyText(totals.outputVat),
		moneyText(totals.purchaseTaxable), moneyText(totals.inputVat)}
	want := []string{"99000.00", "20000.00", "0.00", "6930.00", "50000.00", "3500.00"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pp30 = %v want %v", got, want)
		}
	}

	// บริษัทอื่น/งวดอื่นต้องว่าง (holding ใหม่ที่ยังไม่มีรายการภาษีเห็นรายงานว่าง ไม่ใช่ error)
	if other, err := generalledger.VatRecordsForPeriod(ctx, db, "NOPE", 2026, 9, 2); err != nil || len(other) != 0 {
		t.Fatalf("other company = %+v err=%v", other, err)
	}
}
