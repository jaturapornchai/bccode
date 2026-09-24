//go:build integration

package handlers

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"testing"

	"smlcloudplatform/internal/centraldb/centraldbtest"
	"smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/rdform"
)

// TestTaxFilingStorage ตรวจการบันทึกแบบยื่นภาษีกับ PostgreSQL จริงทีละขั้น:
// สร้าง → ซ้ำงวด → แก้ตาม version → version เก่า → ประวัติ → รายการ → เปิด → ลบ
//
// ฐานเปล่าก็รันได้ (test สร้างข้อมูลเอง) — verify.sh postgres ส่งตัวแปรนี้ให้อัตโนมัติ
//
//	BC_TAXFORM_TEST_POSTGRES_DSN='postgres://postgres@127.0.0.1:5432/taxform_it?sslmode=disable'
func TestTaxFilingStorage(t *testing.T) {
	dsn := os.Getenv("BC_TAXFORM_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_TAXFORM_TEST_POSTGRES_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	if err := ensureTaxFilingSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	const company = "IT01"
	cleanup := func() {
		db.ExecContext(ctx, `DELETE FROM tax_filing_history WHERE filing_id IN (SELECT id FROM tax_filings WHERE company_code=$1)`, company)
		db.ExecContext(ctx, `DELETE FROM tax_filings WHERE company_code=$1`, company)
	}
	cleanup()
	defer cleanup()

	doc := rdform.Document{Values: map[string]string{"filing_type": "normal", "output_tax": "70.00", "input_tax": "35.00"}}
	f := TaxFiling{Code: "pp30", Year: 2026, Month: 9, Document: &doc}
	if err := saveTaxFiling(ctx, db, company, "demo", &f); err != nil {
		t.Fatalf("create: %v", err)
	}
	var stored string
	if err := db.QueryRowContext(ctx, `SELECT document->'values'->>'output_tax' FROM tax_filings WHERE id=$1 AND version=1`, f.ID).Scan(&stored); err != nil || stored != "70.00" {
		t.Fatalf("PG after create: %q %v (ยอดต้องเก็บเป็นสตริงทศนิยม)", stored, err)
	}

	dup := TaxFiling{Code: "pp30", Year: 2026, Month: 9, Document: &doc}
	if err := saveTaxFiling(ctx, db, company, "demo", &dup); !errors.Is(err, errTaxFilingDuplicate) {
		t.Fatalf("duplicate period err = %v", err)
	}

	doc.Values["output_tax"] = "77.00"
	if err := saveTaxFiling(ctx, db, company, "demo2", &f); err != nil || f.Version != 2 || f.UpdatedBy != "demo2" {
		t.Fatalf("update: v=%d by=%s err=%v", f.Version, f.UpdatedBy, err)
	}
	var count int
	db.QueryRowContext(ctx, `SELECT count(*) FROM tax_filings WHERE company_code=$1`, company).Scan(&count)
	if count != 1 {
		t.Fatalf("rows after update = %d, want 1 (แก้แถวเดิม ไม่สร้างใหม่)", count)
	}

	stale := f
	stale.Version = 1
	if err := saveTaxFiling(ctx, db, company, "demo", &stale); !errors.Is(err, errTaxFilingConflict) {
		t.Fatalf("stale version err = %v", err)
	}
	db.QueryRowContext(ctx, `SELECT count(*) FROM tax_filing_history WHERE filing_id=$1`, f.ID).Scan(&count)
	if count != 2 {
		t.Fatalf("history = %d, want 2", count)
	}

	// ยื่นเพิ่มเติมเป็นอีกฉบับของงวดเดียวกัน
	extra := TaxFiling{Code: "pp30", Year: 2026, Month: 9, FilingSeq: 1, Document: &doc}
	if err := saveTaxFiling(ctx, db, company, "demo", &extra); err != nil {
		t.Fatalf("additional filing: %v", err)
	}
	list, err := listTaxFilings(ctx, db, company, "pp30", 2026)
	if err != nil || len(list) != 2 || list[0].FilingSeq != 0 {
		t.Fatalf("list = %+v err=%v", list, err)
	}
	loaded, err := loadTaxFiling(ctx, db, company, f.ID)
	if err != nil || loaded.Version != 2 || loaded.Document.Values["output_tax"] != "77.00" {
		t.Fatalf("load = %+v err=%v", loaded, err)
	}
	if _, err := loadTaxFiling(ctx, db, "OTHER", f.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("other company must not load: %v", err)
	}

	// ลบ: id ที่ไม่มี (หรือของบริษัทอื่น) = not found (404) ไม่ใช่ "มีผู้อื่นแก้" (409); version เก่า = conflict; ถูกต้อง = หายจริงใน PG
	if err := deleteTaxFiling(ctx, db, company, f.ID+100000, 1); !errors.Is(err, errTaxFilingNotFound) {
		t.Fatalf("delete missing id err = %v, want not found", err)
	}
	if err := deleteTaxFiling(ctx, db, "OTHER", f.ID, f.Version); !errors.Is(err, errTaxFilingNotFound) {
		t.Fatalf("delete other company err = %v, want not found", err)
	}
	if err := deleteTaxFiling(ctx, db, company, f.ID, 1); !errors.Is(err, errTaxFilingConflict) {
		t.Fatalf("delete stale version err = %v, want conflict", err)
	}
	if err := deleteTaxFiling(ctx, db, company, f.ID, f.Version); err != nil {
		t.Fatalf("delete: %v", err)
	}
	db.QueryRowContext(ctx, `SELECT count(*) FROM tax_filings WHERE id=$1`, f.ID).Scan(&count)
	if count != 0 {
		t.Fatalf("filing still in PG after delete: %d", count)
	}
}

// TestCopyProfileSameFamily - หัวแบบ (ที่อยู่/ผู้ลงนาม) ยกจากฉบับล่าสุดของแบบตระกูลเดียวกันเท่านั้น:
// แบบของบริษัทห้ามยกที่อยู่ของผู้ยื่นบุคคลธรรมดา (ภ.ง.ด.93/94) และกลับกัน
func TestCopyProfileSameFamily(t *testing.T) {
	dsn := os.Getenv("BC_TAXFORM_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_TAXFORM_TEST_POSTGRES_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	if err := ensureTaxFilingSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	const company = "IT04"
	cleanup := func() {
		db.ExecContext(ctx, `DELETE FROM tax_filing_history WHERE filing_id IN (SELECT id FROM tax_filings WHERE company_code=$1)`, company)
		db.ExecContext(ctx, `DELETE FROM tax_filings WHERE company_code=$1`, company)
	}
	cleanup()
	defer cleanup()
	save := func(code string, year, month int, values map[string]string) {
		t.Helper()
		doc := rdform.Document{Values: values}
		if err := saveTaxFiling(ctx, db, company, "demo", &TaxFiling{Code: code, Year: year, Month: month, Document: &doc}); err != nil {
			t.Fatal(err)
		}
	}
	save("pnd53", 2026, 8, map[string]string{"filing_type": "normal", "addr_road": "ถนนพระราม 2", "signer_name": "นายกรรมการ บริษัท"})
	// ฉบับล่าสุดเป็นแบบบุคคลธรรมดา — ต้องไม่ปนเข้าแบบของบริษัท
	save("pnd94", 2026, 0, map[string]string{"tax_id": "3100500123456", "name": "นายสมชาย ใจดี", "addr_road": "ถนนสุขุมวิท", "signer_name": "นายสมชาย ใจดี"})

	pp30, err := newFormFiller("pp30")
	if err != nil {
		t.Fatal(err)
	}
	if err := pp30.copyProfile(ctx, db, company, "pp30", false); err != nil {
		t.Fatal(err)
	}
	expectValues(t, pp30.values, map[string]string{"addr_road": "ถนนพระราม 2", "signer_name": "นายกรรมการ บริษัท", "tax_id": "", "name": ""})

	pnd94, err := newFormFiller("pnd94")
	if err != nil {
		t.Fatal(err)
	}
	if err := pnd94.copyProfile(ctx, db, company, "pnd94", true); err != nil {
		t.Fatal(err)
	}
	expectValues(t, pnd94.values, map[string]string{"addr_road": "ถนนสุขุมวิท", "tax_id": "3100500123456", "name": "นายสมชาย ใจดี"})

	pnd93, err := newFormFiller("pnd93")
	if err != nil {
		t.Fatal(err)
	}
	if err := pnd93.copyProfile(ctx, db, company, "pnd93", true); err != nil {
		t.Fatal(err)
	}
	if len(pnd93.values) != 0 {
		t.Fatalf("pnd93 must not copy from pnd53/pnd94: %v", pnd93.values)
	}
}

// TestTaxFormFillFromLedger - บริษัทที่ยังไม่มีรายการเปิดแบบได้พร้อมหมายเหตุ (ไม่ error) และ ภ.พ.30 ดึงยอดจากรายการภาษีที่บันทึกจริง
func TestTaxFormFillFromLedger(t *testing.T) {
	dsn := os.Getenv("BC_TAXFORM_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_TAXFORM_TEST_POSTGRES_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	doc := rdform.Document{Values: map[string]string{}}
	notes, err := fillWithholdingForm(ctx, db, "IT01", "pnd53", 2026, 9, &doc)
	if err != nil || len(notes) != 1 || notes[0].Key != "tax_form_note_no_withholding" {
		t.Fatalf("pnd53 notes=%+v err=%v", notes, err)
	}
	f, err := newFormFiller("pp30")
	if err != nil {
		t.Fatal(err)
	}
	notes, err = fillVatForm(ctx, db, "IT01", 2026, 9, f)
	if err != nil || notes[0].Key != "tax_form_note_no_vat" || f.values["output_tax"] != "0.00" {
		t.Fatalf("pp30 notes=%+v values=%v err=%v", notes, f.values, err)
	}
	cit, err := newFormFiller("pnd50")
	if err != nil {
		t.Fatal(err)
	}
	notes, err = fillCitForm(ctx, db, "IT01", "pnd50", 2026, cit)
	if err != nil || notes[0].Key != "tax_form_note_no_fiscal_year" || cit.values["period_start_year_be"] != "2569" || cit.values["period_end_day"] != "31" {
		t.Fatalf("pnd50 notes=%+v values=%v err=%v", notes, cit.values, err)
	}

	// ภ.พ.30 จากรายการภาษีของใบสำคัญที่ผ่านบัญชี: ขาย 100,000 + 0% 20,000, ลดหนี้ 1,000, ซื้อใช้สิทธิ 50,000 (ต้องห้าม/ร่างไม่นับ)
	const company = "IT02"
	cleanup := func() { db.ExecContext(ctx, `DELETE FROM gl_records WHERE company=$1`, company) }
	cleanup()
	defer cleanup()
	insert := func(id, status, vats string) {
		t.Helper()
		payload := `{"docno":"` + id + `","date":"2026-09-05","status":"` + status + `","details":{"vats":` + vats + `}}`
		if _, err := db.ExecContext(ctx, `INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals',$2,$2,1,$3::jsonb)`, company, id, payload); err != nil {
			t.Fatal(err)
		}
	}
	const party = `"partner_name":"บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด","partner_tax_id":"0105558012349","partner_branch_no":"00000","tax_period_year":2026,"tax_period_month":9,"vat_rate":"7","exempt_amount":"0"`
	insert("UV1", "posted", `[{"id":"S1","tax_type":2,"document_type":1,"tax_invoice_no":"IV1","tax_invoice_date":"2026-09-05",`+party+`,"base_amount":"100000","zero_rate_amount":"20000","vat_amount":"7000"}]`)
	insert("UV2", "posted", `[{"id":"S2","tax_type":2,"document_type":3,"tax_invoice_no":"CN1","tax_invoice_date":"2026-09-20","original_invoice_no":"IV1","original_invoice_date":"2026-09-05",`+party+`,"base_amount":"1000","zero_rate_amount":"0","vat_amount":"70"}]`)
	insert("SV1", "posted", `[{"id":"P1","tax_type":1,"document_type":1,"tax_invoice_no":"PI1","tax_invoice_date":"2026-09-10","claim_status":1,`+party+`,"base_amount":"50000","zero_rate_amount":"0","vat_amount":"3500"},
		{"id":"P2","tax_type":1,"document_type":1,"tax_invoice_no":"PI2","tax_invoice_date":"2026-09-10","claim_status":2,`+party+`,"base_amount":"10000","zero_rate_amount":"0","vat_amount":"700"}]`)
	insert("UV9", "draft", `[{"id":"S9","tax_type":2,"document_type":1,"tax_invoice_no":"IV9","tax_invoice_date":"2026-09-25",`+party+`,"base_amount":"9000","zero_rate_amount":"0","vat_amount":"630"}]`)

	pp30, err := newFormFiller("pp30")
	if err != nil {
		t.Fatal(err)
	}
	notes, err = fillVatForm(ctx, db, company, 2026, 9, pp30)
	if err != nil || notes[0].Key != "tax_form_note_vat_records" || notes[0].Count != 3 {
		t.Fatalf("pp30 notes=%+v err=%v", notes, err)
	}
	pp30.values["excess_brought_forward"] = "400"
	filled := rdform.Document{Values: pp30.values}
	if err := computeTaxForm("pp30", &filled); err != nil {
		t.Fatal(err)
	}
	expectValues(t, filled.Values, map[string]string{"sales_amount": "119000.00", "sales_zero_rate": "20000.00", "sales_taxable": "99000.00",
		"output_tax": "6930.00", "purchase_amount": "50000.00", "input_tax": "3500.00", "tax_payable": "3430.00", "net_payable": "3030.00", "total_payable": "3030.00"})
}

// TestCitCreditsFromLedger ตรวจเครดิตภาษีของ ภ.ง.ด.50/51 จาก PostgreSQL จริง (docs/kms/21-thai-tax-form-references.md §2–§3):
// ภาษีที่บริษัทถูกหักตามรายการภาษีหักของใบที่ผ่านรายการ (ช่วงตามวันที่จ่ายในหลักฐาน ไม่ใช่วันลงบัญชี) และยอดชำระเพิ่มเติมของ ภ.ง.ด.51 ที่บันทึกไว้
func TestCitCreditsFromLedger(t *testing.T) {
	dsn := os.Getenv("BC_TAXFORM_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_TAXFORM_TEST_POSTGRES_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	if err := generalledger.EnsureSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := ensureTaxFilingSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	const company = "IT03"
	cleanup := func() {
		db.ExecContext(ctx, `DELETE FROM gl_lines WHERE company=$1`, company)
		db.ExecContext(ctx, `DELETE FROM gl_records WHERE company=$1`, company)
		db.ExecContext(ctx, `DELETE FROM tax_filing_history WHERE filing_id IN (SELECT id FROM tax_filings WHERE company_code=$1)`, company)
		db.ExecContext(ctx, `DELETE FROM tax_filings WHERE company_code=$1`, company)
	}
	cleanup()
	defer cleanup()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	// ใบสำคัญ: สถานะ, วันลงบัญชี, รายการภาษีหัก (ทิศทาง, วันที่จ่ายในหลักฐาน, ยอดภาษี)
	journal := func(id, status, entryDate, withholdings string) {
		t.Helper()
		exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals',$2,$2,1,$3::jsonb)`, company, id,
			`{"docno":"`+id+`","date":"`+entryDate+`","status":"`+status+`","details":{"withholdings":[`+withholdings+`]}}`)
	}
	wht := func(id string, direction int, paidDate, tax string) string {
		return `{"id":"` + id + `","wht_direction":` + strconv.Itoa(direction) + `,"form_type":"PND53","partner_code":"C001","payment_date":"` + paidDate +
			`","income_tax_type":"3_tres","condition_type":1,"wht_rate":"3","base_amount":"10000","tax_amount":"` + tax + `"}`
	}
	journal("RV1", "posted", "2026-03-10", wht("W1", 2, "2026-03-10", "300.00")+","+wht("W2", 1, "2026-03-10", "90.00")) // ทิศทาง 1 = เราหักผู้อื่น → ไม่นับ
	journal("RV2", "posted", "2026-08-15", "")                                                                           // ไม่มีรายการภาษีหัก = ไม่มีหลักฐาน → ไม่นับ
	journal("RV3", "posted", "2026-01-05", wht("W3", 2, "2025-12-28", "50.00"))                                          // ถูกหักปีก่อน ลงบัญชีปีนี้ → ไม่นับ
	journal("RV4", "posted", "2027-01-03", wht("W4", 2, "2026-12-30", "20.25"))                                          // ถูกหักปลายปี ลงบัญชีปีหน้า → นับใน ภ.ง.ด.50
	journal("RV5", "draft", "2026-04-01", wht("W5", 2, "2026-04-01", "70.00"))                                           // ร่างยังไม่ผ่านรายการ → ไม่นับ

	pnd51 := func(year, seq int, sign, balance string) {
		t.Helper()
		doc := rdform.Document{Values: map[string]string{"filing_type": "normal", "r2_6_sign": sign, "r2_6_balance": balance}}
		if err := saveTaxFiling(ctx, db, company, "demo", &TaxFiling{Code: "pnd51", Year: year, FilingSeq: seq, Document: &doc}); err != nil {
			t.Fatal(err)
		}
	}
	pnd51(2026, 0, "payable", "60000.50")
	pnd51(2026, 1, "payable", "1000.00") // ยื่นเพิ่มเติม: ยอดชำระเพิ่มของฉบับนี้
	pnd51(2026, 2, "overpaid", "500.00") // ชำระไว้เกิน = ไม่ได้ชำระ → ไม่นับ
	pnd51(2025, 0, "payable", "999.00")  // คนละปี

	f50, err := newFormFiller("pnd50")
	if err != nil {
		t.Fatal(err)
	}
	notes, err := fillCitForm(ctx, db, company, "pnd50", 2026, f50)
	if err != nil {
		t.Fatal(err)
	}
	expectValues(t, f50.values, map[string]string{"less_wht": "320.25", "less_pnd51_paid": "61000.50"})
	expectNote(t, notes, TaxFormNote{Key: "tax_form_note_cit_wht_credit", Count: 2, Amount: "320.25"})
	expectNote(t, notes, TaxFormNote{Key: "tax_form_note_cit_pnd51_paid", Count: 2, Amount: "61000.50"})
	doc := rdform.Document{Values: f50.values}
	if err := computeTaxForm("pnd50", &doc); err != nil {
		t.Fatal(err)
	}
	expectValues(t, doc.Values, map[string]string{"less_total": "61320.75", "tax_balance": ""})

	f51, err := newFormFiller("pnd51")
	if err != nil {
		t.Fatal(err)
	}
	notes, err = fillCitForm(ctx, db, company, "pnd51", 2026, f51)
	if err != nil {
		t.Fatal(err)
	}
	expectValues(t, f51.values, map[string]string{"r2_5_1_wht": "300.00", "less_pnd51_paid": ""})
	expectNote(t, notes, TaxFormNote{Key: "tax_form_note_cit_wht_credit", Count: 1, Amount: "300.00"})
}

func expectNote(t *testing.T, notes []TaxFormNote, want TaxFormNote) {
	t.Helper()
	for _, n := range notes {
		if n == want {
			return
		}
	}
	t.Fatalf("note %+v not found in %+v", want, notes)
}

// หัวแบบจากทะเบียนบริษัทบนฐานควบคุมกลางจริง: ที่อยู่สำหรับภาษี + โทรศัพท์ + ที่อยู่บรรทัดเดียวสำหรับ 50 ทวิ;
// บริษัทที่ยังไม่กรอกที่อยู่ = ไม่มี address/addressline (จอแสดงว่ายังไม่ระบุ ไม่เดา)
func TestQueryCompanyHeaderTaxAddress(t *testing.T) {
	db := centraldbtest.New(t)
	ctx := context.Background()
	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name) VALUES ('rungrueng','กลุ่มกิจการรุ่งเรืองกรุ๊ป')`)
	centraldbtest.Exec(t, db, `INSERT INTO companies (holding_code, code, name, tax_id, addr_building, addr_room, addr_floor, addr_no, addr_road,
		addr_subdistrict, addr_district, addr_province, addr_postcode, phone) VALUES
		('rungrueng','01','บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด','0105558012349','สาทรซิตี้ทาวเวอร์','1201','12','175','สาทรใต้',
		 'ทุ่งมหาเมฆ','สาทร','กรุงเทพมหานคร','10120',' 02-123-4567 '),
		('rungrueng','02','บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด','','','','','','','','','','','')`)

	header, err := queryCompanyHeader(ctx, db, "rungrueng", "01")
	if err != nil {
		t.Fatal(err)
	}
	want := "อาคารสาทรซิตี้ทาวเวอร์ ห้องเลขที่ 1201 ชั้นที่ 12 เลขที่ 175 ถนนสาทรใต้ แขวงทุ่งมหาเมฆ เขตสาทร กรุงเทพมหานคร 10120"
	if header.TaxID != "0105558012349" || header.Phone != "02-123-4567" || header.Address == nil || header.Address.Postcode != "10120" || header.AddressLine != want {
		t.Fatalf("header %+v line %q", header, header.AddressLine)
	}

	blank, err := queryCompanyHeader(ctx, db, "rungrueng", "02")
	if err != nil || blank.Address != nil || blank.AddressLine != "" || blank.Phone != "" || blank.Name == "" {
		t.Fatalf("blank registry: %+v err=%v", blank, err)
	}
	missing, err := queryCompanyHeader(ctx, db, "rungrueng", "99")
	if err != nil || missing.Code != "99" || missing.Name != "" {
		t.Fatalf("missing company: %+v err=%v", missing, err)
	}
}
