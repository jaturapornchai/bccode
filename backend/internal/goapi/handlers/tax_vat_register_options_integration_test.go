//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"smlcloudplatform/internal/generalledger"
)

// TestVatRegisterSortClaimRemark - ตัวเลือกทะเบียนภาษีซื้อ/ขายแบบ Champ 5536/5537: คอลัมน์หมายเหตุ, "เรียงตาม" 4 แบบ (ORDER BY ที่ SQL),
// "แสดงยอดขอคืนภาษี" (ขอคืน/ไม่ขอคืน/ทั้งหมด + ช่องภาษีแยกขอคืน/ไม่ขอคืน) — ตัวเลือกกรองเฉพาะทะเบียน ยอด ภ.พ.30 ต้องเท่าเดิม (ใช้สิทธิเท่านั้น)
func TestVatRegisterSortClaimRemark(t *testing.T) {
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
	const company = "VATREGOPTS"
	cleanup := func() {
		if _, err := db.Exec(`DELETE FROM gl_records WHERE company=$1`, company); err != nil {
			t.Fatal(err)
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	const party = `"partner_name":"บริษัท ศรีสยามวัสดุภัณฑ์ จำกัด","partner_tax_id":"0105561098765","partner_branch_no":"00000","document_type":1,"vat_rate":"7"`
	insert := func(id, docNo, date string, vats ...string) {
		t.Helper()
		payload := `{"docno":"` + docNo + `","date":"` + date + `","status":"posted","details":{"vats":[` + strings.Join(vats, ",") + `]}}`
		if _, err := db.Exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals',$2,$3,1,$4::jsonb)`, company, id, docNo, payload); err != nil {
			t.Fatal(err)
		}
	}
	// ภาษีซื้อ งวด ก.ย. 2569: PI-20/PI-10 ใช้สิทธิ, PI-30 ต้องห้าม, PI-05 ไม่ใช้สิทธิ — เลขที่ใบสำคัญ/วันที่ไขว้กันให้ทุกแบบการเรียงได้ลำดับต่างกัน
	insert("J1", "SV-C", "2026-09-03",
		`{"id":"P1","tax_type":1,"tax_invoice_no":"PI-20","tax_invoice_date":"2026-09-10",`+party+`,"base_amount":"1000","zero_rate_amount":"0","exempt_amount":"0","vat_amount":"70","tax_period_year":2026,"tax_period_month":9,"claim_status":1,"remark":"ค่าเช่าโกดังลาดกระบัง เดือน ก.ย. 2569"}`)
	insert("J2", "SV-B", "2026-09-25",
		`{"id":"P2","tax_type":1,"tax_invoice_no":"PI-10","tax_invoice_date":"2026-09-05",`+party+`,"base_amount":"2000","zero_rate_amount":"0","exempt_amount":"0","vat_amount":"140","tax_period_year":2026,"tax_period_month":9,"claim_status":1}`)
	insert("J3", "SV-A", "2026-09-15",
		`{"id":"P3","tax_type":1,"tax_invoice_no":"PI-30","tax_invoice_date":"2026-09-12",`+party+`,"base_amount":"500","zero_rate_amount":"0","exempt_amount":"0","vat_amount":"35","tax_period_year":2026,"tax_period_month":9,"claim_status":2,"claim_reason":"ค่ารับรอง"}`,
		`{"id":"P4","tax_type":1,"tax_invoice_no":"PI-05","tax_invoice_date":"2026-09-20",`+party+`,"base_amount":"300","zero_rate_amount":"0","exempt_amount":"0","vat_amount":"21","tax_period_year":2026,"tax_period_month":9,"claim_status":4}`)
	insert("J4", "UV-1", "2026-09-08",
		`{"id":"S1","tax_type":2,"tax_invoice_no":"IV-1","tax_invoice_date":"2026-09-08",`+party+`,"base_amount":"4000","zero_rate_amount":"0","exempt_amount":"0","vat_amount":"280","tax_period_year":2026,"tax_period_month":9}`)

	register := func(taxType int, q generalledger.VatRegisterQuery) ([]TaxVatRegisterRow, TaxVatRegisterSummary) {
		t.Helper()
		records, err := generalledger.VatRegisterRecords(ctx, db, company, 2026, 9, taxType, q)
		if err != nil {
			t.Fatal(err)
		}
		return buildVatRegister(records)
	}
	invoiceNos := func(rows []TaxVatRegisterRow) string {
		nos := make([]string, 0, len(rows))
		for _, r := range rows {
			nos = append(nos, r.TaxInvoiceNo)
		}
		return strings.Join(nos, ",")
	}

	// เรียงตาม (ทุกสถานะ เพื่อเห็นครบ 4 แถว)
	for _, tc := range []struct{ sort, want string }{
		{"", "PI-10,PI-20,PI-30,PI-05"},
		{generalledger.VatSortInvoiceDate, "PI-10,PI-20,PI-30,PI-05"},
		{generalledger.VatSortDocumentDate, "PI-20,PI-30,PI-05,PI-10"},
		{generalledger.VatSortInvoiceNo, "PI-05,PI-10,PI-20,PI-30"},
		{generalledger.VatSortDocumentNo, "PI-30,PI-05,PI-10,PI-20"},
	} {
		rows, summary := register(1, generalledger.VatRegisterQuery{Sort: tc.sort, Claim: generalledger.VatClaimAll})
		if got := invoiceNos(rows); got != tc.want {
			t.Errorf("sort %q: invoices = %s, want %s", tc.sort, got, tc.want)
		}
		// ยอดรายวันเรียงตามวันที่ใบกำกับเสมอ ไม่ว่าแถวเรียงแบบใด
		dates := make([]string, 0, len(summary.Days))
		for _, d := range summary.Days {
			dates = append(dates, d.Date)
		}
		if got := strings.Join(dates, ","); got != "2026-09-05,2026-09-10,2026-09-12,2026-09-20" {
			t.Errorf("sort %q: days = %s", tc.sort, got)
		}
	}
	if _, err := generalledger.VatRegisterRecords(ctx, db, company, 2026, 9, 1, generalledger.VatRegisterQuery{Sort: "p.doc_no DESC"}); err == nil {
		t.Error("sort outside the whitelist must be rejected")
	}

	// แสดงยอดขอคืนภาษี: ค่าเริ่มต้น = ขอคืน (ชุดเดียวกับ ภ.พ.30); ช่องภาษีแยกขอคืน/ไม่ขอคืนรวมได้เท่ายอดภาษี
	for _, tc := range []struct{ claim, want, vat, claimed, notClaimed string }{
		{"", "PI-10,PI-20", "210.00", "210.00", "0.00"},
		{generalledger.VatClaimClaimed, "PI-10,PI-20", "210.00", "210.00", "0.00"},
		{generalledger.VatClaimNotClaimed, "PI-30,PI-05", "56.00", "0.00", "56.00"},
		{generalledger.VatClaimAll, "PI-10,PI-20,PI-30,PI-05", "266.00", "210.00", "56.00"},
	} {
		rows, summary := register(1, generalledger.VatRegisterQuery{Claim: tc.claim})
		if got := invoiceNos(rows); got != tc.want {
			t.Errorf("claim %q: invoices = %s, want %s", tc.claim, got, tc.want)
		}
		if summary.VatAmount != tc.vat || summary.VatClaimed != tc.claimed || summary.VatNotClaimed != tc.notClaimed {
			t.Errorf("claim %q: vat=%s claimed=%s notclaimed=%s, want %s %s %s", tc.claim, summary.VatAmount, summary.VatClaimed, summary.VatNotClaimed, tc.vat, tc.claimed, tc.notClaimed)
		}
	}
	rows, _ := register(1, generalledger.VatRegisterQuery{Claim: generalledger.VatClaimAll})
	for _, r := range rows {
		claimedRow := r.TaxInvoiceNo == "PI-10" || r.TaxInvoiceNo == "PI-20"
		if claimedRow && (r.VatClaimed != r.VatAmount || r.VatNotClaimed != "0.00") || !claimedRow && (r.VatNotClaimed != r.VatAmount || r.VatClaimed != "0.00") {
			t.Errorf("row %s: vat=%s claimed=%s notclaimed=%s", r.TaxInvoiceNo, r.VatAmount, r.VatClaimed, r.VatNotClaimed)
		}
		// หมายเหตุที่ผู้ใช้บันทึกในรายการภาษี = คอลัมน์ "หมายเหตุ"
		wantRemark := ""
		if r.TaxInvoiceNo == "PI-20" {
			wantRemark = "ค่าเช่าโกดังลาดกระบัง เดือน ก.ย. 2569"
		}
		if r.Remark != wantRemark {
			t.Errorf("row %s: remark = %q, want %q", r.TaxInvoiceNo, r.Remark, wantRemark)
		}
	}
	// ภาษีขายไม่มีสถานะการใช้สิทธิ: ตัวเลือกขอคืนต้องถูกปฏิเสธ; ภาษีทั้งหมดอยู่ช่องขอคืน
	if _, err := generalledger.VatRegisterRecords(ctx, db, company, 2026, 9, 2, generalledger.VatRegisterQuery{Claim: generalledger.VatClaimAll}); err == nil {
		t.Error("claim filter on output VAT must be rejected")
	}
	if sales, summary := register(2, generalledger.VatRegisterQuery{Sort: generalledger.VatSortDocumentNo}); invoiceNos(sales) != "IV-1" || summary.VatClaimed != "280.00" || summary.VatNotClaimed != "0.00" {
		t.Errorf("sales = %s claimed=%s notclaimed=%s", invoiceNos(sales), summary.VatClaimed, summary.VatNotClaimed)
	}

	// ภ.พ.30: ภาษีซื้อเฉพาะที่ใช้สิทธิเหมือนเดิม — ตัวเลือกของทะเบียนไม่กระทบยอดยื่น
	sales, err := generalledger.VatRecordsForPeriod(ctx, db, company, 2026, 9, 2)
	if err != nil {
		t.Fatal(err)
	}
	purchases, err := generalledger.VatRecordsForPeriod(ctx, db, company, 2026, 9, 1)
	if err != nil {
		t.Fatal(err)
	}
	pp30 := sumPP30(sales, purchases)
	if pp30.inputVat.StringFixed(2) != "210.00" || pp30.purchaseTaxable.StringFixed(2) != "3000.00" || pp30.outputVat.StringFixed(2) != "280.00" || len(purchases) != 2 {
		t.Errorf("pp30 input vat=%s purchase base=%s output vat=%s purchases=%d", pp30.inputVat.StringFixed(2), pp30.purchaseTaxable.StringFixed(2), pp30.outputVat.StringFixed(2), len(purchases))
	}
}
