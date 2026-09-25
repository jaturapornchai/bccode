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

func openTaxReviewDB(t *testing.T, company string) (*sql.DB, func(query string, args ...any)) {
	t.Helper()
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set BC_GL_TEST_POSTGRES_DSN")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := generalledger.EnsureSchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		for _, table := range []string{"gl_lines", "gl_records", "gl_subledger_partners"} {
			if _, err := db.Exec(`DELETE FROM `+table+` WHERE company=$1`, company); err != nil {
				t.Fatal(err)
			}
		}
	}
	cleanup()
	t.Cleanup(cleanup)
	return db, func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatalf("%v\n%s", err, query)
		}
	}
}

// รายงานภาษีซื้อ + ภ.พ.30: ใบกำกับฉบับเดียวกัน (ผู้ออก + เลขที่ + วันที่) ในใบสำคัญสองใบ → ทั้งสองแถวเตือน, สรุปนับ 2, ภ.พ.30 มีหมายเหตุ
// ผู้ขายอีกรายใช้เลขเดียวกัน → ไม่เตือน (ม.86/4: เลขที่ไม่ซ้ำเฉพาะภายในผู้ออกรายเดียว)
func TestVatDuplicateInvoiceReachesRegisterAndPP30(t *testing.T) {
	const company = "VATDUP"
	db, exec := openTaxReviewDB(t, company)
	ctx := context.Background()
	insert := func(id, docNo, vats string) {
		exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals',$2,$3,1,$4::jsonb)`, company, id, docNo,
			`{"docno":"`+docNo+`","date":"2026-09-10","status":"posted","branchcode":"00000","details":{"vats":`+vats+`}}`)
	}
	purchase := func(id, taxID, invoice string) string {
		return `[{"id":"` + id + `","tax_type":1,"document_type":1,"tax_invoice_no":"` + invoice + `","tax_invoice_date":"2026-09-08","tax_period_year":2026,"tax_period_month":9,"claim_status":1,` +
			`"partner_name":"บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด","partner_tax_id":"` + taxID + `","partner_branch_no":"00000","base_amount":"20000","zero_rate_amount":"0","exempt_amount":"0","vat_rate":"7","vat_amount":"1400"}]`
	}
	insert("J1", "PV6909-001", purchase("P1", "0105558012349", "IV6909-0077"))
	insert("J2", "PV6909-014", purchase("P2", "0105558012349", "iv6909-0077"))
	insert("J3", "PV6909-015", purchase("P3", "0105561234567", "IV6909-0077"))

	records, err := generalledger.VatRecordsForPeriod(ctx, db, company, 2026, 9, 1)
	if err != nil {
		t.Fatal(err)
	}
	rows, summary := buildVatRegister(records)
	dups := map[string]int{}
	for i, r := range records {
		dups[r.DocNo] = len(rows[i].DuplicateDocNos)
	}
	if len(rows) != 3 || summary.DuplicateCount != 2 || dups["PV6909-001"] != 1 || dups["PV6909-014"] != 1 || dups["PV6909-015"] != 0 {
		t.Fatalf("register rows = %+v summary = %+v", rows, summary)
	}
	// ยอดไม่ถูกตัดเอง — เตือนให้ตรวจเท่านั้น (ผู้ใช้ตัดสินจากเอกสารจริง)
	if summary.VatAmount != "4200.00" {
		t.Fatalf("duplicate warning must not change totals: %+v", summary)
	}
	pp30, err := newFormFiller("pp30")
	if err != nil {
		t.Fatal(err)
	}
	notes, err := fillVatForm(ctx, db, company, 2026, 9, pp30)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range notes {
		if n.Key == "tax_form_note_duplicate_invoice" {
			found = n.Count == 2
		}
	}
	if !found {
		t.Fatalf("pp30 notes = %+v, want tax_form_note_duplicate_invoice count 2", notes)
	}
}

// รายงานภาษีหัก ณ ที่จ่ายอ่านเฉพาะรายการที่บันทึกในใบสำคัญ (แบบ Champ BCAPWTaxList; เลิกเดาจากชื่อบัญชี 2026-09-25):
// ใบโอนยอดระหว่างบัญชีภาษีหัก, ใบที่ไม่ได้บันทึกรายการ (จ่าย/ถูกหัก) และยอดในบัญชีที่เกินรายการที่บันทึก ไม่เป็นแถว
func TestWithholdingReportRecordedOnly(t *testing.T) {
	const company = "WHTREV"
	db, exec := openTaxReviewDB(t, company)
	ctx := context.Background()
	account := func(id, code, kind, name string) {
		exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'accounts',$2,$3,1,$4::jsonb)`, company, id, code,
			`{"accountcode":"`+code+`","accounttype":"`+kind+`","names":[{"code":"th","name":"`+name+`"}]}`)
	}
	account("A1", "2150", "liability", "ภ.ง.ด.3 ค้างจ่าย")
	account("A2", "2151", "liability", "ภ.ง.ด.53 ค้างจ่าย")
	account("A3", "1153", "asset", "ภาษีเงินได้ถูกหัก ณ ที่จ่าย")
	account("A4", "1121", "asset", "เงินฝากธนาคาร")
	account("A5", "5221", "expense", "ค่าบริการ")
	account("A6", "4121", "income", "รายได้จากการให้บริการ")
	exec(`INSERT INTO gl_subledger_partners(company,code,version,payload) VALUES($1,'SERVE',1,'{"partner_code":"SERVE","name_th":"บริษัท บริการอาคารดี จำกัด","tax_id":"0105561234567","is_supplier":true,"is_active":true}')`, company)
	journal := func(id, docNo, payload string) {
		exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals',$2,$3,1,$4::jsonb)`, company, id, docNo, payload)
	}
	line := `INSERT INTO gl_lines(company,journal_id,line_no,doc_no,entry_date,fiscal_year,book_code,branch_code,department_code,project_code,kind,currency,scale,account_code,account_name,account_type,normal_balance,is_cash,description,cash_flow,debit,credit)
		VALUES($1,$2,$3,$4,$5,'2569','PV','00000','','','manual','THB',2,$6,$6,$7,'debit',false,$8,'',$9,$10)`

	// โอนยอดจากบัญชี ภ.ง.ด.3 ไปบัญชี ภ.ง.ด.53 — ไม่ใช่การหักภาษี
	journal("J1", "JV6910-T01", `{"status":"posted","kind":"manual","description":"โอนภาษีหักค้างจ่ายให้ตรงแบบยื่น"}`)
	exec(line, company, "J1", 1, "JV6910-T01", "2026-10-05", "2150", "liability", "โอนออก ภ.ง.ด.3", "300", "0")
	exec(line, company, "J1", 2, "JV6910-T01", "2026-10-05", "2151", "liability", "โอนเข้า ภ.ง.ด.53", "0", "300")

	// จ่ายค่าบริการ ลงบัญชีภาษีหัก แต่ไม่ได้บันทึกรายการ → ไม่เป็นแถว
	journal("J2", "PV6910-W02", `{"status":"posted","kind":"manual","description":"จ่ายค่าบริการทำความสะอาด"}`)
	exec(line, company, "J2", 1, "PV6910-W02", "2026-10-08", "5221", "expense", "ค่าบริการ", "10000", "0")
	exec(line, company, "J2", 2, "PV6910-W02", "2026-10-08", "1121", "asset", "จ่ายเงิน", "0", "9700")
	exec(line, company, "J2", 3, "PV6910-W02", "2026-10-08", "2151", "liability", "ภาษีหัก 3%", "0", "300")

	// ใบเดียวจ่ายสองราย หักรวม 450 ในบัญชี ภ.ง.ด.53 แต่บันทึกรายการไว้รายเดียว (ภาษี 300) → แถวเดียวตามที่บันทึก
	journal("J3", "PV6910-W03", `{"status":"posted","kind":"manual","description":"จ่ายค่าบริการสองราย","details":{"withholdings":[{"id":"W1","wht_direction":1,"form_type":"PND53","partner_code":"SERVE","payment_date":"2026-10-10","income_tax_type":"3_tres","condition_type":1,"wht_rate":"3","base_amount":"10000","tax_amount":"300"}]}}`)
	exec(line, company, "J3", 1, "PV6910-W03", "2026-10-10", "5221", "expense", "ค่าบริการ", "15000", "0")
	exec(line, company, "J3", 2, "PV6910-W03", "2026-10-10", "1121", "asset", "จ่ายเงิน", "0", "14550")
	exec(line, company, "J3", 3, "PV6910-W03", "2026-10-10", "2151", "liability", "ภาษีหัก 3%", "0", "450")

	// ฝั่งถูกหัก: รับชำระค่าบริการ ลูกค้าหัก 3% ไม่ได้บันทึกรายการ → ไม่เป็นแถว
	journal("J4", "RV6910-W04", `{"status":"posted","kind":"manual","description":"รับชำระค่าบริการ ลูกค้าหักภาษี ณ ที่จ่าย 3%"}`)
	exec(line, company, "J4", 1, "RV6910-W04", "2026-10-12", "1121", "asset", "รับเงิน", "9700", "0")
	exec(line, company, "J4", 2, "RV6910-W04", "2026-10-12", "1153", "asset", "ภาษีถูกหัก 3%", "300", "0")
	exec(line, company, "J4", 3, "RV6910-W04", "2026-10-12", "4121", "income", "ค่าบริการ", "0", "10000")

	oct53, err := buildWithholdingReport(ctx, db, company, 2026, 10, "paid", []string{"53"})
	if err != nil {
		t.Fatal(err)
	}
	// เฉพาะรายการที่บันทึก (W03 ภาษี 300) — ใบโอนยอด, ใบที่ไม่ได้บันทึก และยอดส่วนเกินในบัญชีไม่ถูกเดาเป็นแถว
	if len(oct53.Rows) != 1 || oct53.Rows[0].DocNo != "PV6910-W03" || oct53.Rows[0].WhtAmount != "300.00" || oct53.Rows[0].BaseAmount != "10000.00" ||
		oct53.Rows[0].RatePercent != "3.00" || oct53.Rows[0].PartnerCode != "SERVE" || oct53.Summary.WhtTotal != "300.00" || oct53.NoteKey != "" {
		t.Fatalf("PND53 Oct = %+v", oct53)
	}
	if oct3, err := buildWithholdingReport(ctx, db, company, 2026, 10, "paid", []string{"3"}); err != nil || len(oct3.Rows) != 0 || oct3.NoteKey != "tax_wht_note_no_records" {
		t.Fatalf("PND3 rows = %+v err=%v", oct3, err)
	}
	received, err := buildWithholdingReport(ctx, db, company, 2026, 10, "received", nil)
	if err != nil || len(received.Rows) != 0 || received.NoteKey != "tax_wht_note_no_records" {
		t.Fatalf("received = %+v err=%v", received, err)
	}
}

// กลับรายการ (ประกาศฯ ฉบับที่ 111: งวด = เดือนที่จ่ายเงิน): กลับในเดือนเดียวกัน → หักล้าง ไม่แสดงทั้งคู่;
// กลับในเดือนหลัง → ต้นฉบับอยู่ในเดือนของตัวเองพร้อม ReversedMonth และเดือนที่กลับไม่มีแถวติดลบ
// ใบที่ไม่ได้บันทึกรายการ (R3/R4, บัญชีที่ชื่ออ้างหลายแบบ M1, ภาษีออกแทน B1) ไม่เป็นแถวในแบบใด — ผู้ใช้ต้องบันทึกรายการในใบสำคัญ
func TestWithholdingReportReversalsAndForms2026_09_24(t *testing.T) {
	const company = "WHTREV2"
	db, exec := openTaxReviewDB(t, company)
	ctx := context.Background()
	account := func(id, code, kind, name string) {
		exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'accounts',$2,$3,1,$4::jsonb)`, company, id, code,
			`{"accountcode":"`+code+`","accounttype":"`+kind+`","names":[{"code":"th","name":"`+name+`"}]}`)
	}
	account("A1", "2150", "liability", "ภ.ง.ด.3 ค้างจ่าย")
	account("A2", "2151", "liability", "ภ.ง.ด.53 ค้างจ่าย")
	account("A4", "1121", "asset", "เงินฝากธนาคาร")
	account("A5", "5221", "expense", "ค่าบริการ")
	account("A7", "2152", "liability", "ภาษีเงินได้หัก ณ ที่จ่ายค้างจ่าย ภ.ง.ด.3/ภ.ง.ด.53")
	account("A8", "5390", "expense", "ภาษีเงินได้หัก ณ ที่จ่ายออกแทน")
	exec(`INSERT INTO gl_subledger_partners(company,code,version,payload) VALUES($1,'SERVE',1,'{"partner_code":"SERVE","name_th":"บริษัท บริการอาคารดี จำกัด","tax_id":"0105561234567","is_supplier":true,"is_active":true}')`, company)
	journal := func(id, docNo, payload string) {
		exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals',$2,$3,1,$4::jsonb)`, company, id, docNo, payload)
	}
	line := `INSERT INTO gl_lines(company,journal_id,line_no,doc_no,entry_date,fiscal_year,book_code,branch_code,department_code,project_code,kind,currency,scale,account_code,account_name,account_type,normal_balance,is_cash,description,cash_flow,debit,credit)
		VALUES($1,$2,$3,$4,$5,'2569','PV','00000','','','manual','THB',2,$6,$6,$7,'debit',false,$8,'',$9,$10)`
	payment := func(id, docNo, date, whtAccount string) {
		exec(line, company, id, 1, docNo, date, "5221", "expense", "ค่าบริการ", "10000", "0")
		exec(line, company, id, 2, docNo, date, "1121", "asset", "จ่ายเงิน", "0", "9700")
		exec(line, company, id, 3, docNo, date, whtAccount, "liability", "ภาษีหัก 3%", "0", "300")
	}
	// ใบกลับรายการตามที่ระบบสร้าง: kind reversal + reversalof + บรรทัดสลับฝั่ง (ถูกข้ามด้วย whtNonTaxJournalKinds)
	reversal := func(id, docNo, of, date string) {
		journal(id, docNo, `{"status":"posted","kind":"reversal","reversalof":"`+of+`","date":"`+date+`","description":"กลับรายการ"}`)
		exec(line, company, id, 1, docNo, date, "5221", "expense", "กลับรายการ: ค่าบริการ", "0", "10000")
		exec(line, company, id, 2, docNo, date, "1121", "asset", "กลับรายการ: จ่ายเงิน", "9700", "0")
		exec(line, company, id, 3, docNo, date, "2151", "liability", "กลับรายการ: ภาษีหัก 3%", "300", "0")
	}
	recorded := func(id, paid string) string {
		return `{"withholdings":[{"id":"` + id + `","wht_direction":1,"form_type":"PND53","partner_code":"SERVE","payment_date":"` + paid +
			`","income_tax_type":"3_tres","condition_type":1,"wht_rate":"3","base_amount":"10000","tax_amount":"300"}]}`
	}

	// R1: บันทึกรายการ จ่าย 10 ต.ค. กลับรายการ 20 ต.ค. (เดือนเดียวกัน) → ไม่มีแถว
	journal("R1", "PV6910-R01", `{"status":"reversed","kind":"manual","date":"2026-10-10","details":`+recorded("W1", "2026-10-10")+`}`)
	payment("R1", "PV6910-R01", "2026-10-10", "2151")
	reversal("R1R", "PV6910-R01R", "R1", "2026-10-20")
	// R2: บันทึกรายการ จ่าย 11 ต.ค. กลับรายการ 5 พ.ย. → แถวใน ต.ค. ReversedMonth 2026-11; พ.ย. ไม่มีแถว
	journal("R2", "PV6910-R02", `{"status":"reversed","kind":"manual","date":"2026-10-11","details":`+recorded("W2", "2026-10-11")+`}`)
	payment("R2", "PV6910-R02", "2026-10-11", "2151")
	reversal("R2R", "PV6911-R02R", "R2", "2026-11-05")
	// R3: ไม่ได้บันทึกรายการ (inferred) กลับรายการ 3 พ.ย. → แถวใน ต.ค. ReversedMonth 2026-11
	journal("R3", "PV6910-R03", `{"status":"reversed","kind":"manual","date":"2026-10-12"}`)
	payment("R3", "PV6910-R03", "2026-10-12", "2151")
	reversal("R3R", "PV6911-R03R", "R3", "2026-11-03")
	// R4: inferred กลับรายการ 25 ต.ค. (เดือนเดียวกัน) → ไม่มีแถว
	journal("R4", "PV6910-R04", `{"status":"reversed","kind":"manual","date":"2026-10-13"}`)
	payment("R4", "PV6910-R04", "2026-10-13", "2151")
	reversal("R4R", "PV6910-R04R", "R4", "2026-10-25")
	// M1: บัญชีภาษีหักที่ชื่ออ้างทั้ง ภ.ง.ด.3 และ ภ.ง.ด.53 ไม่ได้บันทึกรายการ → แบบว่าง ไม่เติมลงแบบใด
	journal("M1", "PV6910-M01", `{"status":"posted","kind":"manual","date":"2026-10-14"}`)
	payment("M1", "PV6910-M01", "2026-10-14", "2152")
	// B1: บริษัทออกภาษีให้ผู้รับเงิน: Dr ภาษีเงินได้หัก ณ ที่จ่ายออกแทน 30.93 / Cr ภ.ง.ด.53 ค้างจ่าย 30.93
	journal("B1", "JV6910-B01", `{"status":"posted","kind":"manual","date":"2026-10-31"}`)
	exec(line, company, "B1", 1, "JV6910-B01", "2026-10-31", "5390", "expense", "ภาษีออกแทนผู้รับเงิน", "30.93", "0")
	exec(line, company, "B1", 2, "JV6910-B01", "2026-10-31", "2151", "liability", "ภาษีหัก ณ ที่จ่ายออกแทน", "0", "30.93")

	byDoc := func(rows []TaxWithholdingRow) map[string][]TaxWithholdingRow {
		out := map[string][]TaxWithholdingRow{}
		for _, r := range rows {
			out[r.DocNo] = append(out[r.DocNo], r)
		}
		return out
	}
	oct53, err := buildWithholdingReport(ctx, db, company, 2026, 10, "paid", []string{"53"})
	if err != nil {
		t.Fatal(err)
	}
	docs := byDoc(oct53.Rows)
	for _, gone := range []string{"PV6910-R01", "PV6910-R01R", "PV6910-R04", "PV6910-R04R", "PV6910-M01"} {
		if len(docs[gone]) != 0 {
			t.Fatalf("%s must not be in the PND53 report: %+v", gone, docs[gone])
		}
	}
	if r2 := docs["PV6910-R02"]; len(r2) != 1 || r2[0].ReversedMonth != "2026-11" || r2[0].WhtAmount != "300.00" || r2[0].BaseAmount != "10000.00" {
		t.Fatalf("R2 reversed later = %+v", r2)
	}
	if oct53.Summary.WhtTotal != "300.00" || len(oct53.Rows) != 1 {
		t.Fatalf("PND53 Oct: total %s rows %+v", oct53.Summary.WhtTotal, oct53.Rows)
	}
	oct3, err := buildWithholdingReport(ctx, db, company, 2026, 10, "paid", []string{"3"})
	if err != nil || len(oct3.Rows) != 0 {
		t.Fatalf("PND3 Oct rows=%+v err=%v", oct3.Rows, err)
	}
	all, err := buildWithholdingReport(ctx, db, company, 2026, 10, "paid", nil)
	if err != nil || len(all.Rows) != 1 || all.Rows[0].DocNo != "PV6910-R02" {
		t.Fatalf("all forms Oct rows=%+v err=%v", all.Rows, err)
	}
	nov, err := buildWithholdingReport(ctx, db, company, 2026, 11, "paid", nil)
	if err != nil || len(nov.Rows) != 0 {
		t.Fatalf("Nov (reversal month) rows=%+v err=%v", nov.Rows, err)
	}
	// แบบ ภ.ง.ด.53 ที่เติมจากรายการที่บันทึก: หมายเหตุกลับรายการเดือนหลัง 1 รายการ (R2)
	notes, err := fillWithholdingForm(ctx, db, company, "pnd53", 2026, 10, &rdform.Document{Values: map[string]string{}})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, n := range notes {
		got[n.Key] = n.Count
	}
	if got["tax_form_note_wht_reversed_later"] != 1 {
		t.Fatalf("PND53 prefill notes = %+v", notes)
	}
}

// ภาษีขายที่กลับรายการหลังงวด (ส.ค. → กลับ 3 ก.ย.): ภ.พ.30 ส.ค. ยังรวมยอดพร้อมหมายเหตุ ม.83/4, ภ.พ.30 ก.ย. ไม่หักแต่เตือนให้ดูแท็บยกเลิกข้ามงวด,
// มุมมองยกเลิกข้ามงวดของ ก.ย. แสดงงวดเดิม เลขที่ใบกลับรายการ และเหตุผล
func TestVatReversedAfterPeriodFormsAndView(t *testing.T) {
	const company = "VATREV"
	db, exec := openTaxReviewDB(t, company)
	ctx := context.Background()
	const buyer = `"partner_name":"บริษัท หอมกรุ่น คอฟฟี่ แอนด์ เบเกอรี่ จำกัด","partner_tax_id":"0105558012349","partner_branch_no":"00000"`
	sale := func(id, invoice, base, vat string) string {
		return `{"id":"` + id + `","tax_type":2,"document_type":1,"tax_invoice_no":"` + invoice + `","tax_invoice_date":"2026-08-20","tax_period_year":2026,"tax_period_month":8,` +
			buyer + `,"base_amount":"` + base + `","zero_rate_amount":"0","exempt_amount":"0","vat_rate":"7","vat_amount":"` + vat + `"}`
	}
	insert := func(id, docNo, date, status, extra string) {
		exec(`INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,'journals',$2,$3,1,$4::jsonb)`, company, id, docNo,
			`{"docno":"`+docNo+`","date":"`+date+`","status":"`+status+`"`+extra+`}`)
	}
	insert("J1", "UV6908-001", "2026-08-20", "reversed", `,"kind":"manual","details":{"vats":[`+sale("S1", "IV6908-001", "1000", "70")+`]}`)
	insert("J2", "REV-UV6908-001", "2026-09-03", "posted", `,"kind":"reversal","reversalof":"J1","reason":"ออกใบกำกับภาษีผิดราย"`)
	insert("J3", "UV6908-002", "2026-08-25", "posted", `,"kind":"manual","details":{"vats":[`+sale("S3", "IV6908-002", "500", "35")+`]}`)

	pp30 := func(month int) (map[string]int, string) {
		t.Helper()
		f, err := newFormFiller("pp30")
		if err != nil {
			t.Fatal(err)
		}
		notes, err := fillVatForm(ctx, db, company, 2026, month, f)
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]int{}
		for _, n := range notes {
			got[n.Key] = n.Count
		}
		return got, f.values["output_tax"]
	}
	if notes, output := pp30(8); output != "105.00" || notes["tax_form_note_vat_reversed_later"] != 1 || notes["tax_form_note_vat_cross_period_cancel"] != 0 {
		t.Fatalf("PP30 Aug output=%s notes=%v", output, notes)
	}
	if notes, output := pp30(9); output != "" && output != "0.00" || notes["tax_form_note_vat_cross_period_cancel"] != 1 || notes["tax_form_note_vat_reversed_later"] != 0 {
		t.Fatalf("PP30 Sep output=%q notes=%v", output, notes)
	}

	cancelled, err := generalledger.VatCrossPeriodCancellations(ctx, db, company, 2026, 9, 2)
	if err != nil {
		t.Fatal(err)
	}
	rows, summary := buildVatRegister(cancelled)
	if len(rows) != 1 || rows[0].TaxInvoiceNo != "IV6908-001" || rows[0].TaxMonth != "2026-08" || rows[0].ReversalDocNo != "REV-UV6908-001" || summary.VatAmount != "70.00" {
		t.Fatalf("cross-period rows = %+v summary = %+v", rows, summary)
	}
	if note := vatRegisterRowNote(vatViewReversedLater, rows[0], "th"); note != "ยื่นในงวด สิงหาคม 2569 — กลับรายการด้วยใบ REV-UV6908-001 — ออกใบกำกับภาษีผิดราย" {
		t.Fatalf("cross-period row note = %q", note)
	}
}
