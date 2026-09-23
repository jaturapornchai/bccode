//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"smlcloudplatform/internal/generalledger"
)

// TestWithholdingReportUsesRecordedBase - ใบที่บันทึกฐานภาษีไว้ รายงาน ภ.ง.ด.53 ต้องใช้ฐานที่บันทึก ไม่ใช่ยอดจ่ายรวม VAT
// ใบที่ไม่ได้บันทึก ยังประมาณจากบรรทัดบัญชีและบอกว่าเป็นค่าประมาณ (inferred) — ใช้ฐานเปล่าได้ (สร้าง schema เอง)
func TestWithholdingReportUsesRecordedBase(t *testing.T) {
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
	const company = "WHTREC"
	cleanup := func() {
		for _, table := range []string{"gl_lines", "gl_records", "gl_subledger_partners"} {
			if _, err := db.Exec(`DELETE FROM `+table+` WHERE company=$1`, company); err != nil {
				t.Fatal(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatalf("%v\n%s", err, query)
		}
	}
	exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'accounts','A1','21410',1,'{"accountcode":"21410","accounttype":"liability","names":[{"code":"th","name":"ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย ภ.ง.ด.53"}]}')`, company)
	exec(`INSERT INTO gl_subledger_partners(company,code,version,payload) VALUES($1,'TRANS',1,'{"partner_code":"TRANS","name_th":"บริษัท ขนส่งไทยเร็ว จำกัด","tax_id":"0105558012349","address":"12 ถนนบางนา-ตราด กรุงเทพฯ","is_supplier":true,"is_active":true}')`, company)
	// PV1: จ่าย 107,000 รวม VAT หัก 3% จากฐานก่อน VAT ที่ผู้ใช้บันทึก 100,000 → 3,000
	exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals','J1','PV1',1,'{"description":"จ่ายค่าขนส่ง","details":{"partners":[],"withholdings":[{"id":"W1","wht_direction":1,"form_type":"PND53","partner_code":"TRANS","payment_date":"2026-09-15","income_tax_type":"3_tres","income_description":"ค่าขนส่งสินค้า","condition_type":1,"wht_rate":"3","base_amount":"100000","tax_amount":"3000"}]}}')`, company)
	// PV2: ไม่ได้บันทึกรายการภาษี — ฐานประมาณจากเดบิตฝั่งตรงข้าม
	exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals','J2','PV2',1,'{"description":"จ่ายค่าเช่า"}')`, company)
	line := `INSERT INTO gl_lines(company,journal_id,line_no,doc_no,entry_date,fiscal_year,book_code,branch_code,department_code,project_code,kind,currency,scale,account_code,account_name,account_type,normal_balance,is_cash,description,cash_flow,debit,credit)
		VALUES($1,$2,$3,$4,$5,'2569','PV','00000','','','manual','THB',2,$6,$6,$7,$8,false,$9,'',$10,$11)`
	exec(line, company, "J1", 1, "PV1", "2026-09-15", "52300", "expense", "debit", "ค่าขนส่ง", "107000", "0")
	exec(line, company, "J1", 2, "PV1", "2026-09-15", "21410", "liability", "credit", "ภาษีหัก 3%", "0", "3000")
	exec(line, company, "J1", 3, "PV1", "2026-09-15", "11110", "asset", "debit", "เงินสด", "0", "104000")
	exec(line, company, "J2", 1, "PV2", "2026-09-20", "52100", "expense", "debit", "ค่าเช่า", "10000", "0")
	exec(line, company, "J2", 2, "PV2", "2026-09-20", "21410", "liability", "credit", "ภาษีหัก 5%", "0", "500")
	exec(line, company, "J2", 3, "PV2", "2026-09-20", "11110", "asset", "debit", "เงินสด", "0", "9500")

	report, err := buildWithholdingReport(ctx, db, company, 2026, 9, "paid", []string{"53"})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Rows) != 2 {
		t.Fatalf("rows = %+v", report.Rows)
	}
	pv1, pv2 := report.Rows[0], report.Rows[1]
	if pv1.TaxBaseSource != "recorded" || pv1.BaseAmount != "100000.00" || pv1.WhtAmount != "3000.00" || pv1.RatePercent != "3.00" || pv1.IncomeType != "3_tres" || pv1.PartnerName != "บริษัท ขนส่งไทยเร็ว จำกัด" || pv1.PaidDate != "2026-09-15" {
		t.Fatalf("recorded row = %+v", pv1)
	}
	if pv2.TaxBaseSource != "inferred" || pv2.BaseAmount != "10000.00" || pv2.WhtAmount != "500.00" {
		t.Fatalf("inferred row = %+v", pv2)
	}
	if report.Summary.BaseTotal != "110000.00" || report.Summary.WhtTotal != "3500.00" {
		t.Fatalf("summary = %+v", report.Summary)
	}
	// แบบ ภ.ง.ด.3 ต้องไม่เห็นรายการที่บันทึกเป็น ภ.ง.ด.53
	if other, err := buildWithholdingReport(ctx, db, company, 2026, 9, "paid", []string{"3"}); err != nil || len(other.Rows) != 0 {
		t.Fatalf("PND3 rows = %+v err=%v", other.Rows, err)
	}
}
