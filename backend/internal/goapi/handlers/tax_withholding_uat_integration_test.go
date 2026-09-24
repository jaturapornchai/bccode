//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/rdform"
)

// TestWithholdingReportUATRegressions2026_09_24 - ข้อผิดพลาดจาก UAT ภาษีหัก ณ ที่จ่าย 2026-09-24 (ตรวจบน PostgreSQL จริง):
//   - S25/S26: ใบโอนยอดระหว่างบัญชีภาษีหัก (Dr ภ.ง.ด.3 / Cr ภ.ง.ด.53) ห้ามเป็นแถวประมาณ — เดิมทำให้ภาษีใน ภ.ง.ด.53 นับซ้ำ
//   - S24: ใบยอดยกมา (kind opening) ห้ามเข้ารายงาน ทั้งแถวประมาณและรายการที่บันทึก — เดิมเป็นแถวอัตรา 100% ใน ม.ค.
//   - S14: บัญชีรายได้ชื่อ "…มีภาษีมูลค่าเพิ่ม 7%" เป็นฐานภาษี ไม่ใช่บรรทัดภาษี — เดิมได้ฐาน 0 สุทธิติดลบ
//   - S10/S12: แถวเรียงตามวันที่จ่าย ไม่ใช่วันที่ใบ และใบแนบ ภ.ง.ด.53 พิมพ์ตามลำดับนั้น
func TestWithholdingReportUATRegressions2026_09_24(t *testing.T) {
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
	const company = "WHTUAT"
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
	account := func(id, code, kind, name string) {
		exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'accounts',$2,$3,1,$4::jsonb)`, company, id, code,
			`{"accountcode":"`+code+`","accounttype":"`+kind+`","names":[{"code":"th","name":"`+name+`"}]}`)
	}
	// ชื่อบัญชีตามผัง seed rungrueng — รหัสบัญชีไม่ถูกใช้เป็นเงื่อนไข
	account("A1", "2142", "liability", "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.3")
	account("A2", "2143", "liability", "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย - ภ.ง.ด.53")
	account("A3", "1153", "asset", "ภาษีเงินได้ถูกหัก ณ ที่จ่าย")
	account("A4", "4121", "income", "รายได้จากการให้บริการ - มีภาษีมูลค่าเพิ่ม 7%")
	account("A5", "2131", "liability", "ภาษีขาย")
	account("A6", "1121", "asset", "เงินฝากธนาคาร")
	account("A7", "5221", "expense", "ค่าบริการ")
	exec(`INSERT INTO gl_subledger_partners(company,code,version,payload) VALUES($1,'SERVE',1,'{"partner_code":"SERVE","name_th":"บริษัท บริการอาคารดี จำกัด","tax_id":"0105561234567","is_supplier":true,"is_active":true}')`, company)
	journal := func(id, docNo, payload string) {
		exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals',$2,$3,1,$4::jsonb)`, company, id, docNo, payload)
	}
	recorded := func(paid, base, tax string) string {
		return `"details":{"withholdings":[{"id":"W1","wht_direction":1,"form_type":"PND53","partner_code":"SERVE","payment_date":"` + paid +
			`","income_tax_type":"3_tres","income_description":"ค่าบริการทำความสะอาด","condition_type":1,"wht_rate":"2","base_amount":"` + base + `","tax_amount":"` + tax + `"}]}`
	}
	line := `INSERT INTO gl_lines(company,journal_id,line_no,doc_no,entry_date,fiscal_year,book_code,branch_code,department_code,project_code,kind,currency,scale,account_code,account_name,account_type,normal_balance,is_cash,description,cash_flow,debit,credit)
		VALUES($1,$2,$3,$4,$5,'2569','PV','00000','','',$6,'THB',2,$7,$7,$8,'debit',false,$9,'',$10,$11)`

	// S25: PV6912-W05 จ่ายค่าบริการ บันทึกเป็น ภ.ง.ด.53 แต่ลงยอดหักในบัญชี ภ.ง.ด.3 → PV6912-W06 โอนยอดไปบัญชี ภ.ง.ด.53
	journal("J5", "PV6912-W05", `{"status":"posted","kind":"manual","description":"จ่ายค่าบริการทำความสะอาด",`+recorded("2026-12-05", "10000", "200")+`}`)
	exec(line, company, "J5", 1, "PV6912-W05", "2026-12-05", "manual", "5221", "expense", "ค่าบริการ", "10000", "0")
	exec(line, company, "J5", 2, "PV6912-W05", "2026-12-05", "manual", "1121", "asset", "จ่ายเงิน", "0", "9800")
	exec(line, company, "J5", 3, "PV6912-W05", "2026-12-05", "manual", "2142", "liability", "ภาษีหัก 2%", "0", "200")
	journal("J6", "PV6912-W06", `{"status":"posted","kind":"manual","description":"โอนภาษีหัก ณ ที่จ่ายไปบัญชี ภ.ง.ด.53 ให้ตรงแบบยื่น"}`)
	exec(line, company, "J6", 1, "PV6912-W06", "2026-12-20", "manual", "2142", "liability", "โอนออก ภ.ง.ด.3", "200", "0")
	exec(line, company, "J6", 2, "PV6912-W06", "2026-12-20", "manual", "2143", "liability", "โอนเข้า ภ.ง.ด.53", "0", "200")
	dec53, err := buildWithholdingReport(ctx, db, company, 2026, 12, "paid", []string{"53"})
	if err != nil {
		t.Fatal(err)
	}
	if len(dec53.Rows) != 1 || dec53.Rows[0].DocNo != "PV6912-W05" || dec53.Summary.BaseTotal != "10000.00" || dec53.Summary.WhtTotal != "200.00" {
		t.Fatalf("S25 Dec PND53 rows = %+v summary = %+v", dec53.Rows, dec53.Summary)
	}
	if dec3, err := buildWithholdingReport(ctx, db, company, 2026, 12, "paid", []string{"3"}); err != nil || len(dec3.Rows) != 0 {
		t.Fatalf("S25 Dec PND3 rows = %+v err=%v", dec3.Rows, err)
	}
	doc := rdform.Document{Values: map[string]string{}}
	if _, err := fillWithholdingForm(ctx, db, company, "pnd53", 2026, 12, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Rows) != 1 || doc.Rows[0]["l1_tax"] != "200.00" || doc.Rows[0]["l2_tax"] != "" {
		t.Fatalf("S25 PND53 attachment = %+v", doc.Rows)
	}

	// S24: ใบยอดยกมา 1 ม.ค. — หนี้ภาษีหักค้างจ่ายของ ธ.ค. ปีก่อน + ภาษีถูกหักยกมา + รายการที่บันทึกไว้ในใบยกมา → ไม่ใช่การหักภาษีของ ม.ค.
	journal("J1", "JV6901-W01", `{"status":"posted","kind":"opening","description":"ยอดยกมา ภาษีหัก ณ ที่จ่ายค้างจ่าย ภ.ง.ด.53 ธ.ค. 68",`+recorded("2026-01-01", "42500", "850")+`}`)
	exec(line, company, "J1", 1, "JV6901-W01", "2026-01-01", "opening", "1121", "asset", "ยอดยกมา", "1150", "0")
	exec(line, company, "J1", 2, "JV6901-W01", "2026-01-01", "opening", "2143", "liability", "ยอดยกมา ภ.ง.ด.53", "0", "850")
	exec(line, company, "J1", 3, "JV6901-W01", "2026-01-01", "opening", "1153", "asset", "ยอดยกมา ภาษีถูกหัก", "300", "0")
	exec(line, company, "J1", 4, "JV6901-W01", "2026-01-01", "opening", "1121", "asset", "ยอดยกมา", "0", "600")
	for _, direction := range []string{"paid", "received"} {
		jan, err := buildWithholdingReport(ctx, db, company, 2026, 1, direction, nil)
		if err != nil || len(jan.Rows) != 0 {
			t.Fatalf("S24 Jan %s rows = %+v err=%v", direction, jan.Rows, err)
		}
	}

	// S14: รับชำระค่าบริการ ลูกค้าหัก 3% ไม่ได้บันทึกรายการ — ฐาน = รายได้ 20,000 (ไม่รวมภาษีขาย 1,400)
	journal("J2", "RV6910-W02", `{"status":"posted","kind":"manual","description":"รับชำระค่าบริการ ลูกค้าหักภาษี ณ ที่จ่าย 3%"}`)
	exec(line, company, "J2", 1, "RV6910-W02", "2026-10-14", "manual", "1121", "asset", "รับเงิน", "20800", "0")
	exec(line, company, "J2", 2, "RV6910-W02", "2026-10-14", "manual", "1153", "asset", "ภาษีถูกหัก 3%", "600", "0")
	exec(line, company, "J2", 3, "RV6910-W02", "2026-10-14", "manual", "4121", "income", "ค่าบริการ", "0", "20000")
	exec(line, company, "J2", 4, "RV6910-W02", "2026-10-14", "manual", "2131", "liability", "ภาษีขาย", "0", "1400")
	oct, err := buildWithholdingReport(ctx, db, company, 2026, 10, "received", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(oct.Rows) != 1 || oct.Rows[0].BaseAmount != "20000.00" || oct.Rows[0].WhtAmount != "600.00" || oct.Rows[0].RatePercent != "3.00" || oct.Rows[0].NetAmount != "19400.00" {
		t.Fatalf("S14 Oct received rows = %+v", oct.Rows)
	}

	// S10/S12: ลงบัญชี 31 ต.ค. จ่ายจริง 2 พ.ย. (W11) กับลงบัญชีและจ่าย 1 พ.ย. (W01) — เรียงตามวันที่จ่าย
	journal("J3", "PV6910-W11", `{"status":"posted","kind":"manual","description":"จ่ายค่าบริการ ต.ค.",`+recorded("2026-11-02", "5000", "100")+`}`)
	exec(line, company, "J3", 1, "PV6910-W11", "2026-10-31", "manual", "5221", "expense", "ค่าบริการ", "5000", "0")
	exec(line, company, "J3", 2, "PV6910-W11", "2026-10-31", "manual", "2143", "liability", "ภาษีหัก 2%", "0", "100")
	exec(line, company, "J3", 3, "PV6910-W11", "2026-10-31", "manual", "1121", "asset", "จ่ายเงิน", "0", "4900")
	journal("J4", "PV6911-W01", `{"status":"posted","kind":"manual","description":"จ่ายค่าบริการ พ.ย.",`+recorded("2026-11-01", "3000", "60")+`}`)
	exec(line, company, "J4", 1, "PV6911-W01", "2026-11-01", "manual", "5221", "expense", "ค่าบริการ", "3000", "0")
	exec(line, company, "J4", 2, "PV6911-W01", "2026-11-01", "manual", "2143", "liability", "ภาษีหัก 2%", "0", "60")
	exec(line, company, "J4", 3, "PV6911-W01", "2026-11-01", "manual", "1121", "asset", "จ่ายเงิน", "0", "2940")
	nov, err := buildWithholdingReport(ctx, db, company, 2026, 11, "paid", []string{"53"})
	if err != nil {
		t.Fatal(err)
	}
	if len(nov.Rows) != 2 || nov.Rows[0].DocNo != "PV6911-W01" || nov.Rows[1].DocNo != "PV6910-W11" {
		t.Fatalf("S10 Nov order = %+v", nov.Rows)
	}
	novDoc := rdform.Document{Values: map[string]string{}}
	if _, err := fillWithholdingForm(ctx, db, company, "pnd53", 2026, 11, &novDoc); err != nil {
		t.Fatal(err)
	}
	if len(novDoc.Rows) != 1 || novDoc.Rows[0]["l1_date"] != thaiDate("2026-11-01") || novDoc.Rows[0]["l2_date"] != thaiDate("2026-11-02") {
		t.Fatalf("S12 PND53 attachment order = %+v", novDoc.Rows)
	}
}
