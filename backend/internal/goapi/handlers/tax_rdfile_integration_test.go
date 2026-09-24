//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"smlcloudplatform/internal/generalledger"
	msmodels "smlcloudplatform/pkg/microservice/models"
)

// TestTaxRdFileFromSavedFiling - ไฟล์ยื่นด้วยสื่อ (Format กลาง V2.0) ครบวงจรบน PostgreSQL จริง:
// ทะเบียนคู่ค้า (คำนำหน้า/อำเภอ/จังหวัด/รหัสไปรษณีย์) → รายงานภาษีหัก ณ ที่จ่าย → prefill ใบแนบ → บันทึกฉบับ
// → POST /api/report/tax/form/rdfile ได้ไฟล์ตรงรูปแบบ; version เก่า = 409, บริษัทอื่น = 404, ข้อมูลไม่ครบ = 400 พร้อมจุดที่ต้องแก้
// และการสร้างไฟล์ต้องไม่เขียนอะไรลงฐานข้อมูล
//
//	BC_TAXFORM_TEST_POSTGRES_DSN='postgres://postgres@127.0.0.1:5432/taxform_it?sslmode=disable'
func TestTaxRdFileFromSavedFiling(t *testing.T) {
	dsn := os.Getenv("BC_TAXFORM_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_TAXFORM_TEST_POSTGRES_DSN")
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
	if err := ensureTaxFilingSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	const company = "RDF1" // ผู้ใช้เลือกบริษัท rdf1 → normalizeBusinessCode = RDF1
	cleanup := func() {
		for _, table := range []string{"gl_lines", "gl_records", "gl_subledger_partners"} {
			if _, err := db.Exec(`DELETE FROM `+table+` WHERE company=$1`, company); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := db.Exec(`DELETE FROM tax_filing_history WHERE filing_id IN (SELECT id FROM tax_filings WHERE company_code=$1)`, company); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`DELETE FROM tax_filings WHERE company_code=$1`, company); err != nil {
			t.Fatal(err)
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
	exec(`INSERT INTO gl_subledger_partners(company,code,version,payload) VALUES($1,'TRANS',1,'{"partner_code":"TRANS","name_th":"บริษัท ขนส่งไทยเร็ว จำกัด","title_name":"บริษัท","tax_id":"0105562045671","address":"12 ถนนบางนา-ตราด กรุงเทพฯ","is_supplier":true,"is_active":true}')`, company)
	exec(`INSERT INTO gl_subledger_partners(company,code,version,payload) VALUES($1,'SOMCHAI',1,'{"partner_code":"SOMCHAI","name_th":"นายสมชาย ใจดี","title_name":"นาย","tax_id":"3101701291901","address":"99/1 ซอยสุขุมวิท 101","addr_district":"เขตบางนา","addr_province":"กรุงเทพมหานคร","addr_postcode":"10260","is_supplier":true,"is_active":true}')`, company)
	exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals','J1','PV1',1,'{"status":"posted","details":{"withholdings":[{"id":"W1","wht_direction":1,"form_type":"PND53","partner_code":"TRANS","payment_date":"2026-09-15","income_tax_type":"3_tres","income_description":"ค่าขนส่งสินค้า","condition_type":1,"wht_rate":"3","base_amount":"100000","tax_amount":"3000"}]}}')`, company)
	exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals','J2','PV2',1,'{"status":"posted","details":{"withholdings":[{"id":"W2","wht_direction":1,"form_type":"PND3","partner_code":"SOMCHAI","payment_date":"2026-09-20","income_tax_type":"3_tres","income_description":"ค่าจ้างทำของ","condition_type":1,"wht_rate":"3","base_amount":"20000","tax_amount":"600"}]}}')`, company)
	exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals','J3','PV3',1,'{"status":"posted","details":{"withholdings":[{"id":"W3","wht_direction":1,"form_type":"PND2","partner_code":"SOMCHAI","payment_date":"2026-09-25","income_tax_type":"40_4a","income_description":"ดอกเบี้ยเงินกู้ยืม","condition_type":1,"wht_rate":"15","base_amount":"10000","tax_amount":"1500"}]}}')`, company)
	line := `INSERT INTO gl_lines(company,journal_id,line_no,doc_no,entry_date,fiscal_year,book_code,branch_code,department_code,project_code,kind,currency,scale,account_code,account_name,account_type,normal_balance,is_cash,description,cash_flow,debit,credit)
		VALUES($1,$2,$3,$4,$5,'2569','PV','00000','','','manual','THB',2,$6,$6,$7,$8,false,$9,'',$10,$11)`
	exec(line, company, "J1", 1, "PV1", "2026-09-15", "52300", "expense", "debit", "ค่าขนส่ง", "100000", "0")
	exec(line, company, "J1", 2, "PV1", "2026-09-15", "21410", "liability", "credit", "ภาษีหัก 3%", "0", "3000")
	exec(line, company, "J1", 3, "PV1", "2026-09-15", "11110", "asset", "debit", "เงินสด", "0", "97000")
	exec(line, company, "J2", 1, "PV2", "2026-09-20", "52400", "expense", "debit", "ค่าจ้างทำของ", "20000", "0")
	exec(line, company, "J2", 2, "PV2", "2026-09-20", "21420", "liability", "credit", "ภาษีหัก 3%", "0", "600")
	exec(line, company, "J2", 3, "PV2", "2026-09-20", "11110", "asset", "debit", "เงินสด", "0", "19400")
	exec(line, company, "J3", 1, "PV3", "2026-09-25", "53100", "expense", "debit", "ดอกเบี้ยจ่าย", "10000", "0")
	exec(line, company, "J3", 2, "PV3", "2026-09-25", "21430", "liability", "credit", "ภาษีหัก 15%", "0", "1500")
	exec(line, company, "J3", 3, "PV3", "2026-09-25", "11110", "asset", "debit", "เงินสด", "0", "8500")

	stubWhtCompany(t, CompanyHeader{Name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", TaxID: "0105558012349"})
	original := taxRdFileDB
	taxRdFileDB = func(context.Context, *taxFormCall) (*sql.DB, error) { return db, nil }
	t.Cleanup(func() { taxRdFileDB = original })
	user := msmodels.UserInfo{Username: "tester", HoldingCode: "h1", BusinessCode: "rdf1", UID: "u1"}
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, ictZone)

	// 1) รายงานอ่านคำนำหน้า/ที่อยู่แยกช่องจากทะเบียนคู่ค้า
	report, err := buildWithholdingReport(ctx, db, company, 2026, 9, "paid", []string{"3"})
	if err != nil || len(report.Rows) != 1 {
		t.Fatalf("pnd3 report = %+v err=%v", report.Rows, err)
	}
	if r := report.Rows[0]; r.Title != "นาย" || r.District != "เขตบางนา" || r.Province != "กรุงเทพมหานคร" || r.Postcode != "10260" {
		t.Fatalf("partner title/address not read: %+v", r)
	}

	// 2) ภ.ง.ด.3: prefill → บันทึก → ไฟล์
	doc3, _, err := prefillTaxForm(ctx, db, "h1", company, "pnd3", 2026, 9, now)
	if err != nil || len(doc3.Rows) != 1 {
		t.Fatalf("prefill pnd3 rows=%v err=%v", doc3.Rows, err)
	}
	if r := doc3.Rows[0]; r["title"] != "นาย" || r["name"] != "นายสมชาย" || r["surname"] != "ใจดี" || r["addr_district"] != "เขตบางนา" ||
		r["addr_province"] != "กรุงเทพมหานคร" || r["addr_postcode"] != "10260" || r["l1_date"] != "20/09/2569" {
		t.Fatalf("pnd3 prefill row = %v", r)
	}
	doc3.Values["tax_section"], doc3.Values["media_ref_no"] = "3tres", "ZXC1234567"
	prepared3, err := prepareTaxDocument("pnd3", doc3)
	if err != nil {
		t.Fatalf("prepare pnd3: %v", err)
	}
	f3 := TaxFiling{Code: "pnd3", Year: 2026, Month: 9, Document: &prepared3}
	if err := saveTaxFiling(ctx, db, company, "tester", &f3); err != nil {
		t.Fatalf("save pnd3: %v", err)
	}
	rec := callRdFile(t, f3.ID, f3.Version, &user)
	if rec.Code != http.StatusOK {
		t.Fatalf("pnd3 status %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="PND3_0105558012349_000000_2569_09_00_00.txt"` {
		t.Fatalf("Content-Disposition = %q", got)
	}
	if rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Fatalf("headers = %v", rec.Header())
	}
	lines := rdFileBodyLines(t, rec.Body.String(), 2)
	assertRdLine(t, lines[0], 25, "H|0000|0105558012349|000000|1|PND3|0105558012349|000000|สำนักงานใหญ่|1|0|0||09|2569||00|1|20000.00|600.00|0.00|600.00|0.00|ZXC1234567|1")
	assertRdLine(t, lines[1], 38, rdJoin([]string{"D", "1", "000000", "3101701291901", "0000000000", "นาย", "สมชาย", "ใจดี",
		"20092569", "3.00", "20000.00", "600.00", "ค่าจ้างทำของ", "1", rdEmptyItem, rdEmptyItem}, 9, "เขตบางนา", "กรุงเทพมหานคร", "10260"))

	// การสร้างไฟล์อ่านอย่างเดียว: ฉบับยังเป็น version เดิม ประวัติไม่เพิ่ม
	var version, history int
	if err := db.QueryRow(`SELECT version FROM tax_filings WHERE id=$1`, f3.ID).Scan(&version); err != nil || version != f3.Version {
		t.Fatalf("PG version after export = %d err=%v, want %d", version, err, f3.Version)
	}
	if err := db.QueryRow(`SELECT count(*) FROM tax_filing_history WHERE filing_id=$1`, f3.ID).Scan(&history); err != nil || history != 1 {
		t.Fatalf("PG history after export = %d err=%v, want 1", history, err)
	}

	// 3) version เก่า/ใหม่กว่าที่บันทึก = 409; ฉบับของบริษัทอื่น = 404
	if rec := callRdFile(t, f3.ID, f3.Version+1, &user); rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "tax_form_version_conflict") {
		t.Fatalf("stale version: %d %s", rec.Code, rec.Body.String())
	}
	other := user
	other.BusinessCode = "rdf2"
	if rec := callRdFile(t, f3.ID, f3.Version, &other); rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "tax_form_not_found") {
		t.Fatalf("other company: %d %s", rec.Code, rec.Body.String())
	}

	// 4) ภ.ง.ด.53: เลขอ้างอิงการยื่นด้วยสื่อยกมาจากฉบับก่อน; ลบคำนำหน้า + เลขอ้างอิง + ไม่เลือกมาตรา = 400 พร้อมจุดที่ต้องแก้
	doc53, _, err := prefillTaxForm(ctx, db, "h1", company, "pnd53", 2026, 9, now)
	if err != nil || len(doc53.Rows) != 1 {
		t.Fatalf("prefill pnd53 rows=%v err=%v", doc53.Rows, err)
	}
	if doc53.Values["media_ref_no"] != "ZXC1234567" || doc53.Rows[0]["title"] != "บริษัท" {
		t.Fatalf("pnd53 prefill media_ref_no=%q row=%v", doc53.Values["media_ref_no"], doc53.Rows[0])
	}
	delete(doc53.Values, "media_ref_no")
	delete(doc53.Rows[0], "title")
	prepared53, err := prepareTaxDocument("pnd53", doc53)
	if err != nil {
		t.Fatalf("prepare pnd53: %v", err)
	}
	f53 := TaxFiling{Code: "pnd53", Year: 2026, Month: 9, Document: &prepared53}
	if err := saveTaxFiling(ctx, db, company, "tester", &f53); err != nil {
		t.Fatalf("save pnd53: %v", err)
	}
	rec = callRdFile(t, f53.ID, f53.Version, &user)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("pnd53 incomplete status %d body=%s", rec.Code, rec.Body.String())
	}
	var bad struct {
		Code   string `json:"code"`
		Total  int    `json:"total"`
		Issues []struct {
			Key, Field string
			Row        int
		} `json:"issues"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &bad); err != nil || bad.Code != "tax_rdfile_invalid" || bad.Total != 3 {
		t.Fatalf("issues body = %s err=%v", rec.Body.String(), err)
	}
	got := map[string]bool{}
	for _, i := range bad.Issues {
		got[fmt.Sprintf("%s@%s#%d", i.Key, i.Field, i.Row)] = true
	}
	for _, want := range []string{"tax_rdfile_section_required@tax_section#0", "tax_rdfile_user_id_required@media_ref_no#0", "tax_rdfile_title_required@title#1"} {
		if !got[want] {
			t.Fatalf("missing issue %s in %v", want, got)
		}
	}

	// แก้ครบแล้วบันทึกเป็น version ถัดไป → ไฟล์ออกได้; ชื่อ (FNAME) ตัดคำนำหน้าออก
	prepared53.Values["tax_section"], prepared53.Values["media_ref_no"] = "3tres", "ZXC1234567"
	prepared53.Rows[0]["title"] = "บริษัท"
	if err := saveTaxFiling(ctx, db, company, "tester", &f53); err != nil || f53.Version != 2 {
		t.Fatalf("save pnd53 v2: v=%d err=%v", f53.Version, err)
	}
	rec = callRdFile(t, f53.ID, f53.Version, &user)
	if rec.Code != http.StatusOK {
		t.Fatalf("pnd53 status %d body=%s", rec.Code, rec.Body.String())
	}
	lines = rdFileBodyLines(t, rec.Body.String(), 2)
	assertRdLine(t, lines[0], 25, "H|0000|0105558012349|000000|1|PND53|0105558012349|000000|สำนักงานใหญ่|1|0|0||09|2569||00|1|100000.00|3000.00|0.00|3000.00|0.00|ZXC1234567|1")
	assertRdLine(t, lines[1], 38, rdJoin([]string{"D", "1", "000000", "0105562045671", "0000000000", "บริษัท", "ขนส่งไทยเร็ว จำกัด", "",
		"15092569", "3.00", "100000.00", "3000.00", "ค่าขนส่งสินค้า", "1", rdEmptyItem, rdEmptyItem}, 12))

	// 5) ภ.ง.ด.2: ประเภทเงินได้ interest → 2, ACC_NO ว่างได้
	doc2, _, err := prefillTaxForm(ctx, db, "h1", company, "pnd2", 2026, 9, now)
	if err != nil || len(doc2.Rows) != 1 || doc2.Rows[0]["title"] != "นาย" || doc2.Rows[0]["income_type"] != "interest" {
		t.Fatalf("prefill pnd2 rows=%v err=%v", doc2.Rows, err)
	}
	prepared2, err := prepareTaxDocument("pnd2", doc2)
	if err != nil {
		t.Fatalf("prepare pnd2: %v", err)
	}
	f2 := TaxFiling{Code: "pnd2", Year: 2026, Month: 9, Document: &prepared2}
	if err := saveTaxFiling(ctx, db, company, "tester", &f2); err != nil {
		t.Fatalf("save pnd2: %v", err)
	}
	rec = callRdFile(t, f2.ID, f2.Version, &user)
	if rec.Code != http.StatusOK {
		t.Fatalf("pnd2 status %d body=%s", rec.Code, rec.Body.String())
	}
	lines = rdFileBodyLines(t, rec.Body.String(), 2)
	assertRdLine(t, lines[0], 22, "H|0000|0105558012349|000000|1|PND2|0105558012349|000000|สำนักงานใหญ่||09|2569||00|1|10000.00|1500.00|0.00|1500.00|0.00|ZXC1234567|1")
	assertRdLine(t, lines[1], 27, rdJoin([]string{"D", "1", "000000", "3101701291901", "0000000000", "", "นาย", "สมชาย", "ใจดี",
		"25092569", "15.00", "10000.00", "1500.00", "2", "1"}, 12))
}

func callRdFile(t *testing.T, id int64, version int, user *msmodels.UserInfo) *httptest.ResponseRecorder {
	t.Helper()
	return callTaxReportHandler(t, TaxFormRdFileHandler, fmt.Sprintf(`{"id":%d,"version":%d,"rdfile":{}}`, id, version), user)
}

// rdFileBodyLines - เนื้อไฟล์ต้องขึ้นต้น BOM, คั่นบรรทัดด้วย CRLF และไม่มี CRLF ปิดท้าย
func rdFileBodyLines(t *testing.T, body string, want int) []string {
	t.Helper()
	if !strings.HasPrefix(body, "\uFEFF") {
		t.Fatalf("missing BOM: %q", body[:min(len(body), 12)])
	}
	body = strings.TrimPrefix(body, "\uFEFF")
	if strings.HasSuffix(body, "\n") || strings.HasSuffix(body, "\r") {
		t.Fatal("file must not end with CRLF")
	}
	lines := strings.Split(body, "\r\n")
	if len(lines) != want {
		t.Fatalf("lines = %d want %d: %q", len(lines), want, body)
	}
	for _, l := range lines {
		if strings.ContainsAny(l, "\r\n") {
			t.Fatalf("bare CR/LF inside line %q", l)
		}
	}
	return lines
}

// rdEmptyItem - รายการเงินได้ที่ไม่มี (ไฟล์ต้องเขียนเป็นศูนย์ ไม่ใช่ช่องว่าง)
const rdEmptyItem = "00000000|0.00|0.00|0.00||"

// rdJoin - บรรทัดที่คาดไว้: ช่องหลัก + ช่องที่อยู่ว่าง n ช่อง + ช่องท้าย (ไม่ต้องนับ | เอง)
func rdJoin(head []string, emptyAddr int, tail ...string) string {
	parts := append(append(head, make([]string, emptyAddr)...), tail...)
	return strings.Join(parts, "|")
}

func assertRdLine(t *testing.T, line string, fields int, want string) {
	t.Helper()
	if n := len(strings.Split(line, "|")); n != fields {
		t.Fatalf("fields = %d want %d: %s", n, fields, line)
	}
	if line != want {
		t.Fatalf("line\n got %s\nwant %s", line, want)
	}
}
