//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"fmt"
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

// TestWithholdingReceivedReport - รายงานภาษีถูกหัก: ชื่อบัญชีมาตรฐาน "ภาษีเงินได้ถูกหัก ณ ที่จ่าย" ต้องหาเจอ
// และใบที่บันทึกรายการภาษีถูกหักไว้ต้องเข้ารายงานแม้ลงบัญชีที่ชื่อไม่ตรงรูปแบบ (หลักฐานที่บันทึกชนะการหาจากชื่อ)
func TestWithholdingReceivedReport(t *testing.T) {
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
	const company = "WHTRCV"
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
	exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'accounts','A1','1153',1,'{"accountcode":"1153","accounttype":"asset","names":[{"code":"th","name":"ภาษีเงินได้ถูกหัก ณ ที่จ่าย (WHT ถูกหัก)"}]}')`, company)
	exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'accounts','A2','11990',1,'{"accountcode":"11990","accounttype":"asset","names":[{"code":"th","name":"ลูกหนี้อื่น"}]}')`, company)
	exec(`INSERT INTO gl_subledger_partners(company,code,version,payload) VALUES($1,'CUST',1,'{"partner_code":"CUST","name_th":"บริษัท ก่อสร้างมั่นคง จำกัด","tax_id":"0105560001234","is_customer":true,"is_active":true}')`, company)
	recorded := `{"description":"%s","details":{"withholdings":[{"id":"W1","wht_direction":2,"form_type":"PND53","partner_code":"CUST","payment_date":"2026-09-%s","income_tax_type":"3_tres","income_description":"ค่าบริการ","condition_type":1,"wht_rate":"3","base_amount":"%s","tax_amount":"%s"}]}}`
	// RV1: ลูกค้าหัก 3% ของ 10,000 บันทึกรายการไว้ ลงบัญชีภาษีถูกหักมาตรฐาน
	exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals','J1','RV1',1,$2)`, company, fmt.Sprintf(recorded, "รับชำระค่าบริการ", "05", "10000", "300"))
	// RV2: ไม่ได้บันทึกรายการภาษี — ประมาณจากบรรทัดบัญชีภาษีถูกหัก
	exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals','J2','RV2',1,'{"description":"รับชำระค่าขนส่ง"}')`, company)
	// RV3: บันทึกรายการไว้ แต่ลงยอดหักในบัญชีลูกหนี้อื่น (ชื่อไม่ตรงรูปแบบ) — ต้องยังเข้ารายงาน
	exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals','J3','RV3',1,$2)`, company, fmt.Sprintf(recorded, "รับชำระค่าเช่า", "25", "5000", "250"))
	line := `INSERT INTO gl_lines(company,journal_id,line_no,doc_no,entry_date,fiscal_year,book_code,branch_code,department_code,project_code,kind,currency,scale,account_code,account_name,account_type,normal_balance,is_cash,description,cash_flow,debit,credit)
		VALUES($1,$2,$3,$4,$5,'2569','RV','00000','','','manual','THB',2,$6,$6,$7,$8,false,$9,'',$10,$11)`
	exec(line, company, "J1", 1, "RV1", "2026-09-05", "11110", "asset", "debit", "เงินสด", "9700", "0")
	exec(line, company, "J1", 2, "RV1", "2026-09-05", "1153", "asset", "debit", "ภาษีถูกหัก 3%", "300", "0")
	exec(line, company, "J1", 3, "RV1", "2026-09-05", "41100", "income", "credit", "ค่าบริการ", "0", "10000")
	exec(line, company, "J2", 1, "RV2", "2026-09-10", "11110", "asset", "debit", "เงินสด", "1980", "0")
	exec(line, company, "J2", 2, "RV2", "2026-09-10", "1153", "asset", "debit", "ภาษีถูกหัก 1%", "20", "0")
	exec(line, company, "J2", 3, "RV2", "2026-09-10", "41200", "income", "credit", "ค่าขนส่ง", "0", "2000")
	exec(line, company, "J3", 1, "RV3", "2026-09-25", "11110", "asset", "debit", "เงินสด", "4750", "0")
	exec(line, company, "J3", 2, "RV3", "2026-09-25", "11990", "asset", "debit", "ภาษีถูกหัก 5%", "250", "0")
	exec(line, company, "J3", 3, "RV3", "2026-09-25", "41300", "income", "credit", "ค่าเช่า", "0", "5000")

	report, err := buildWithholdingReport(ctx, db, company, 2026, 9, "received", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Rows) != 3 || report.Note != "" {
		t.Fatalf("rows = %+v note=%q", report.Rows, report.Note)
	}
	rv1, rv2, rv3 := report.Rows[0], report.Rows[1], report.Rows[2]
	if rv1.DocNo != "RV1" || rv1.TaxBaseSource != "recorded" || rv1.WhtAmount != "300.00" || rv1.PartnerName != "บริษัท ก่อสร้างมั่นคง จำกัด" {
		t.Fatalf("RV1 = %+v", rv1)
	}
	if rv2.DocNo != "RV2" || rv2.TaxBaseSource != "inferred" || rv2.WhtAmount != "20.00" || rv2.BaseAmount != "2000.00" {
		t.Fatalf("RV2 = %+v", rv2)
	}
	if rv3.DocNo != "RV3" || rv3.TaxBaseSource != "recorded" || rv3.WhtAmount != "250.00" || rv3.BaseAmount != "5000.00" || rv3.DocDate != "2026-09-25" {
		t.Fatalf("RV3 = %+v", rv3)
	}
	if report.Summary.WhtTotal != "570.00" || report.Summary.BaseTotal != "17000.00" {
		t.Fatalf("summary = %+v", report.Summary)
	}
	// เดือนอื่นต้องว่าง และบอกเหตุผล
	if other, err := buildWithholdingReport(ctx, db, company, 2026, 10, "received", nil); err != nil || len(other.Rows) != 0 {
		t.Fatalf("Oct rows = %+v err=%v", other.Rows, err)
	}
}
