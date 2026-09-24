//go:build integration

package generalledger

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// ใบกำกับฉบับเดียวกันบันทึกซ้ำในใบสำคัญเดียวกัน → DuplicateDocNos มีเลขที่ของตัวเอง; แยกแถวต่างสถานะใช้สิทธิไม่นับว่าซ้ำ
func TestPostgresVatSameVoucherDuplicate(t *testing.T) {
	f := newPGIntegrityFixture(t)
	var gotMonth func(taxType, month int) map[string]string
	create := func(doc string, vats ...SubledgerVat) Journal {
		t.Helper()
		j := Journal{DocNo: doc, Date: "2026-01-15", BookCode: "PV", FiscalYear: "2026", Description: "ซื้อวัสดุก่อสร้าง " + doc, Kind: "manual", BranchCode: "B1",
			Lines:   []Line{{AccountCode: "5000", Debit: "2140", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "2140"}},
			Details: &JournalDetails{Vats: vats}}
		return f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID))
	}
	// PV21: ใบ IV-600 คีย์ซ้ำ 2 แถว สถานะใช้สิทธิเดียวกัน
	create("PV21", vatItem("A1", 1, 1, "IV-600", "2026-01-10", 2026, 1, 1, "1000", nil), vatItem("A2", 1, 1, " iv-600", "2026-01-10", 2026, 1, 1, "1000", nil))
	// PV22: ใบ IV-700 แยกเป็นใช้สิทธิ + ต้องห้าม (สถานะต่างกัน) = ไม่ซ้ำ
	create("PV22", vatItem("B1", 1, 1, "IV-700", "2026-01-11", 2026, 1, 1, "1000", nil), vatItem("B2", 1, 1, "IV-700", "2026-01-11", 2026, 1, 2, "1000", nil))
	// PV23: ใบ IV-600 อีกครั้งในใบสำคัญอื่น → แถวของ PV21 ต้องเห็นทั้ง PV21 และ PV23
	create("PV23", vatItem("C1", 1, 1, "IV-600", "2026-01-10", 2026, 1, 1, "1000", nil))
	// ภาษีขาย SO-9 ซ้ำ 2 แถวในใบเดียว
	create("SV21", vatItem("D1", 2, 1, "SO-9", "2026-01-12", 2026, 1, 0, "1000", nil), vatItem("D2", 2, 1, "SO-9", "2026-01-12", 2026, 1, 0, "1000", nil))
	// PV24: ใบ IV-800 ใช้สิทธิ 2 แถวในใบเดียวกันคนละงวด (ม.ค. + ก.พ.) = ใช้ภาษีซื้อซ้ำใน ภ.พ.30 สองเดือน (review 2026-09-24)
	create("PV24", vatItem("E1", 1, 1, "IV-800", "2026-01-12", 2026, 1, 1, "1000", nil), vatItem("E2", 1, 1, "IV-800", "2026-01-12", 2026, 2, 1, "1000", nil))

	got := func(taxType int) map[string]string { return gotMonth(taxType, 1) }
	gotMonth = func(taxType, month int) map[string]string {
		t.Helper()
		records, err := VatRecordsForPeriod(context.Background(), f.db, "C", 2026, month, taxType)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]string{}
		for _, r := range records {
			out[r.DocNo+"/"+r.ID] = strings.Join(r.DuplicateDocNos, ",")
		}
		return out
	}
	purchases := got(1)
	for key, want := range map[string]string{"PV21/A1": "PV21,PV23", "PV21/A2": "PV21,PV23", "PV22/B1": "", "PV23/C1": "PV21", "PV24/E1": "PV24"} {
		if v, ok := purchases[key]; !ok || v != want {
			t.Errorf("purchase %s duplicates = %q (present %v), want %q", key, v, ok, want)
		}
	}
	if _, ok := purchases["PV22/B2"]; ok {
		t.Error("forbidden input VAT row must not be in the claimed register")
	}
	if v := gotMonth(1, 2)["PV24/E2"]; v != "PV24" {
		t.Errorf("same voucher, next tax month: duplicates = %q, want PV24", v)
	}
	sales := got(2)
	for key, want := range map[string]string{"SV21/D1": "SV21", "SV21/D2": "SV21"} {
		if sales[key] != want {
			t.Errorf("sale %s duplicates = %q want %q", key, sales[key], want)
		}
	}
}

// เติมเลขผู้เสียภาษี/สาขาจากทะเบียนคู่ค้า + กฎใหม่ตรวจเฉพาะแถวใหม่/ที่แก้ — ตรวจค่าใน PostgreSQL ทีละขั้น
func TestPostgresVatRegistryFillAndLegacyRows(t *testing.T) {
	f := newPGIntegrityFixture(t)
	storedVats := func(id string) []SubledgerVat {
		t.Helper()
		var raw []byte
		if err := f.db.QueryRow(`SELECT COALESCE(payload->'details'->'vats','[]'::jsonb) FROM gl_records WHERE company='C' AND kind='journals' AND id=$1`, id).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var rows []SubledgerVat
		if err := json.Unmarshal(raw, &rows); err != nil {
			t.Fatal(err)
		}
		return rows
	}
	status := func(id string) string {
		t.Helper()
		var s string
		if err := f.db.QueryRow(`SELECT payload->>'status' FROM gl_records WHERE company='C' AND kind='journals' AND id=$1`, id).Scan(&s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	// ใบร่างที่บันทึกก่อนมีกฎ: เขียนค่าที่ผิดกฎใหม่ลง PostgreSQL ตรง ๆ (ปีงวด พ.ศ.)
	injectLegacyYear := func(id string, index int) {
		t.Helper()
		res, err := f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload, ARRAY['details','vats',$2::text,'tax_period_year'], '2569'::jsonb) WHERE company='C' AND kind='journals' AND id=$1`, id, index)
		if err != nil {
			t.Fatal(err)
		}
		if n, _ := res.RowsAffected(); n != 1 {
			t.Fatalf("legacy inject touched %d rows", n)
		}
	}

	row := vatItem("V1", 1, 1, "IV-6901-88", "2026-01-10", 2026, 1, 1, "1000", nil)
	row.PartnerCode, row.PartnerTaxID, row.PartnerBranchNo = "SUP-VAT", "", ""
	typed := vatItem("V2", 1, 1, "IV-6901-89", "2026-01-10", 2026, 1, 1, "1000", nil)
	typed.PartnerCode, typed.PartnerTaxID, typed.PartnerBranchNo = "SUP-VAT", "0105558012357", ""
	j := Journal{DocNo: "PV6901-031", Date: "2026-01-15", BookCode: "PV", FiscalYear: "2026", Description: "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. เป็นเงินเชื่อ", Kind: "manual", BranchCode: "B1",
		Lines: []Line{{AccountCode: "5000", Debit: "2140", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "2140"}},
		Details: &JournalDetails{
			Partners: []SubledgerPartner{{Code: "SUP-VAT", Name: "บริษัท ปูนกรุงไทย จำกัด", TaxID: "0105558012349", TaxBranch: "1", IsSupplier: true, IsActive: true}},
			Vats:     []SubledgerVat{row, typed},
		}}
	created := f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID)
	// ขั้น 1: แถวที่ไม่กรอกเลข → เติมจากทะเบียน; แถวที่กรอกเลขอื่น → คงเลขที่พิมพ์ ไม่เติมสาขาของทะเบียน
	vats := storedVats(created.ID)
	if vats[0].PartnerTaxID != "0105558012349" || vats[0].PartnerBranchNo != "00001" {
		t.Fatalf("step 1 registry fill = %q/%q", vats[0].PartnerTaxID, vats[0].PartnerBranchNo)
	}
	if vats[1].PartnerTaxID != "0105558012357" || vats[1].PartnerBranchNo != "" {
		t.Fatalf("step 1 typed tax id overwritten = %q/%q", vats[1].PartnerTaxID, vats[1].PartnerBranchNo)
	}

	// ขั้น 2: เลขผู้เสียภาษีผิดหลักตรวจสอบในแถวใหม่ → 400 ไม่เขียน
	bad := f.journal(created.ID)
	extra := vatItem("V3", 1, 1, "IV-6901-90", "2026-01-10", 2026, 1, 1, "1000", nil)
	extra.PartnerTaxID = "0105558012340"
	bad.Details.Vats = append(bad.Details.Vats, extra)
	user := f.failCode(Command{Resource: "journals", Action: "update", ID: bad.ID, Version: bad.Version, Journal: &bad}, "vat_partner_tax_id_checksum", "partner_tax_id")
	if user.HTTPStatus() != http.StatusBadRequest {
		t.Fatalf("step 2 status %d, want 400", user.HTTPStatus())
	}
	if n := len(storedVats(created.ID)); n != 2 {
		t.Fatalf("step 2 rejected row was written: %d rows", n)
	}

	// ขั้น 3: แถวเดิมที่ผิดกฎใหม่ (ปี พ.ศ.) ไม่บล็อกการแก้ส่วนอื่นของใบร่าง
	injectLegacyYear(created.ID, 0)
	legacy := f.journal(created.ID)
	legacy.Description = "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. จำนวน 500 ถุง เป็นเงินเชื่อ"
	f.run(Command{Resource: "journals", Action: "update", ID: legacy.ID, Version: legacy.Version, Journal: &legacy})
	if vats := storedVats(created.ID); vats[0].TaxPeriodYear != 2569 {
		t.Fatalf("step 3 legacy row changed: %+v", vats[0])
	}
	// ขั้น 4: ผ่านรายการตรวจทุกแถว → ปฏิเสธ ใบยังเป็นร่าง
	legacy = f.journal(created.ID)
	f.failCode(Command{Resource: "journals", Action: "post", ID: legacy.ID, Version: legacy.Version}, "vat_period_year_buddhist", "tax_period_year")
	if s := status(created.ID); s != "draft" {
		t.Fatalf("step 4 status = %s, want draft", s)
	}
	// ขั้น 5: แก้แถวนั้น → ต้องใช้ ค.ศ.
	legacy.Details.Vats[0].Remark = "แก้หมายเหตุ"
	f.failCode(Command{Resource: "journals", Action: "update", ID: legacy.ID, Version: legacy.Version, Journal: &legacy}, "vat_period_year_buddhist", "tax_period_year")
	legacy.Details.Vats[0].TaxPeriodYear = 2026
	f.run(Command{Resource: "journals", Action: "update", ID: legacy.ID, Version: legacy.Version, Journal: &legacy})
	posted := f.post(f.journal(created.ID))
	if s := status(posted.ID); s != "posted" {
		t.Fatalf("step 5 status = %s, want posted", s)
	}

	// ขั้น 6: ใบผ่านบัญชีที่มีแถวเก่าผิดกฎ → reconcile เพิ่มแถวใหม่ได้ (แถวเก่าไม่เปลี่ยน); แก้แถวเก่า → ปฏิเสธ
	injectLegacyYear(posted.ID, 1)
	posted = f.journal(posted.ID)
	add := vatItem("V4", 1, 1, "IV-6901-91", "2026-01-10", 2026, 1, 1, "500", nil)
	f.run(subledgerReconcile(f, posted, JournalDetails{Vats: append(append([]SubledgerVat{}, posted.Details.Vats...), add)}))
	if vats := storedVats(posted.ID); len(vats) != 3 || vats[1].TaxPeriodYear != 2569 {
		t.Fatalf("step 6 reconcile rows = %+v", vats)
	}
	posted = f.journal(posted.ID)
	changed := append([]SubledgerVat{}, posted.Details.Vats...)
	changed[1].Remark = "แก้หลังผ่านบัญชี"
	f.failCode(subledgerReconcile(f, posted, JournalDetails{Vats: changed}), "vat_period_year_buddhist", "tax_period_year")

	// ขั้น 7: กลับรายการใบที่มีแถวเก่าผิดกฎได้เสมอ
	posted = f.journal(posted.ID)
	f.run(Command{Resource: "journals", Action: "reverse", ID: posted.ID, Version: posted.Version, DocNo: "REV6901-031", Date: "2026-01-20", Reason: "บันทึกผิดงวด"})
	if s := status(posted.ID); s != "reversed" {
		t.Fatalf("step 7 status = %s, want reversed", s)
	}
}

// หลักตรวจสอบเลขผู้เสียภาษีของภาษีหัก ณ ที่จ่าย: แถวใหม่ตรวจ, ภ.ง.ด.2 ใช้ 0 ทั้ง 13 หลักได้, แถวเก่าที่ไม่แก้ไม่บล็อก
func TestPostgresWithholdingTaxIDChecksum(t *testing.T) {
	f := newPGIntegrityFixture(t)
	base := func(payee string, form string) Journal {
		return Journal{DocNo: "PV6901-041", Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Description: "จ่ายค่าขนส่ง หักภาษี ณ ที่จ่าย 1%",
			Lines: []Line{{AccountCode: "5000", Debit: "10000", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "10000"}},
			Details: &JournalDetails{
				Partners:     []SubledgerPartner{{Code: "S001", Name: "บริษัท ขนส่งไทยเร็ว จำกัด", TaxID: "0105558012349", IsSupplier: true, IsActive: true}},
				Withholdings: []SubledgerWithholding{{ID: "W1", Direction: 1, FormType: form, PartnerCode: "S001", PaymentDate: "2026-01-10", IncomeType: "3_tres", Condition: 1, Rate: "1", BaseAmount: "10000", PayeeTaxID: payee, PayeeName: "บริษัท ขนส่งไทยเร็ว จำกัด"}},
			}}
	}
	bad := base("0105558012340", "PND53")
	f.failCode(Command{Resource: "journals", Action: "create", Journal: &bad}, "wht_payee_tax_id_checksum", "payee_tax_id")
	var count int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM gl_records WHERE company='C' AND kind='journals' AND code='PV6901-041'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rejected withholding voucher written: %d (err=%v)", count, err)
	}
	zeroDividend := base("0000000000000", "PND2")
	zeroDividend.Details.Withholdings[0].IncomeType = "40_4b_1_1"
	f.failCode(Command{Resource: "journals", Action: "create", Journal: &zeroDividend}, "wht_payee_zero_tax_id_interest_only", "payee_tax_id")
	zero := base("0000000000000", "PND2")
	zero.Details.Withholdings[0].IncomeType = "40_4a"
	created := f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &zero}).ID)
	if w := created.Details.Withholdings[0]; w.PayeeTaxID != "0000000000000" {
		t.Fatalf("PND2 zero PIN = %q", w.PayeeTaxID)
	}
	// แถวเก่าที่เลขผิดหลักตรวจสอบ (บันทึกก่อนมีกฎ) — แก้ส่วนอื่นของใบร่างได้
	if _, err := f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload, '{details,withholdings,0,payee_tax_id}', '"0105558012340"'::jsonb) WHERE company='C' AND kind='journals' AND id=$1`, created.ID); err != nil {
		t.Fatal(err)
	}
	legacy := f.journal(created.ID)
	legacy.Description = "จ่ายค่าขนส่งสินค้า หักภาษี ณ ที่จ่าย 1%"
	f.run(Command{Resource: "journals", Action: "update", ID: legacy.ID, Version: legacy.Version, Journal: &legacy})
	legacy = f.journal(created.ID)
	legacy.Details.Withholdings[0].Remark = "แก้หมายเหตุ"
	f.failCode(Command{Resource: "journals", Action: "update", ID: legacy.ID, Version: legacy.Version, Journal: &legacy}, "wht_payee_tax_id_checksum", "payee_tax_id")
}

// ทะเบียนคู่ค้า: ช่องไฟล์ยื่นภาษีเก็บใน payload และคืนใน list; หลักตรวจสอบ + รหัส 20 ตัวเฉพาะคู่ค้าใหม่
func TestPostgresPartnerRDFieldsAndChecksum(t *testing.T) {
	f := newPGIntegrityFixture(t)
	voucher := func(doc string, p SubledgerPartner) Journal {
		return Journal{DocNo: doc, Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Description: "บันทึกคู่ค้า",
			Lines:   []Line{{AccountCode: "5000", Debit: "100", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "100"}},
			Details: &JournalDetails{Partners: []SubledgerPartner{p}}}
	}
	partnerCount := func(code string) int {
		t.Helper()
		var n int
		if err := f.db.QueryRow(`SELECT COUNT(*) FROM gl_subledger_partners WHERE company='C' AND code=$1`, code).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	bad := voucher("JV6901-051", SubledgerPartner{Code: "S-BAD", Name: "ร้านวัสดุบ้านแข็งแรง", TaxID: "0105558012340", IsSupplier: true, IsActive: true})
	f.failCode(Command{Resource: "journals", Action: "create", Journal: &bad}, "partner_tax_id_checksum", "tax_id")
	long := voucher("JV6901-052", SubledgerPartner{Code: "SUPPLIER-TH-000000001", Name: "ร้านวัสดุบ้านแข็งแรง", IsSupplier: true, IsActive: true})
	f.failCode(Command{Resource: "journals", Action: "create", Journal: &long}, "code_too_long", "partner_code")
	postcode := voucher("JV6901-053", SubledgerPartner{Code: "S-PC", Name: "ร้านวัสดุบ้านแข็งแรง", AddrPostcode: "1026", IsSupplier: true, IsActive: true})
	f.failCode(Command{Resource: "journals", Action: "create", Journal: &postcode}, "partner_addr_postcode_invalid", "addr_postcode")
	if partnerCount("S-BAD")+partnerCount("SUPPLIER-TH-000000001")+partnerCount("S-PC") != 0 {
		t.Fatal("rejected partners were written")
	}

	good := voucher("JV6901-054", SubledgerPartner{Code: "S-RD", Name: "ร้านวัสดุบ้านแข็งแรง", TaxID: "0105558012357", TitleName: "ร้าน", AddrDistrict: "เขตบางนา",
		AddrProvince: "กรุงเทพมหานคร", AddrPostcode: "10260", IsSupplier: true, IsActive: true})
	f.run(Command{Resource: "journals", Action: "create", Journal: &good})
	var raw []byte
	if err := f.db.QueryRow(`SELECT payload FROM gl_subledger_partners WHERE company='C' AND code='S-RD'`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var stored map[string]any
	if err := json.Unmarshal(raw, &stored); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"title_name": "ร้าน", "addr_district": "เขตบางนา", "addr_province": "กรุงเทพมหานคร", "addr_postcode": "10260"} {
		if stored[key] != want {
			t.Errorf("stored %s = %v, want %s", key, stored[key], want)
		}
	}
	page, err := f.store.pg.SubledgerList(f.ctx, f.scope, "partners", "S-RD", 1, 50, "")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range page.Items {
		var p SubledgerPartner
		if err := json.Unmarshal(item, &p); err != nil {
			t.Fatal(err)
		}
		if p.Code == "S-RD" {
			found = p.TitleName == "ร้าน" && p.AddrDistrict == "เขตบางนา" && p.AddrProvince == "กรุงเทพมหานคร" && p.AddrPostcode == "10260"
		}
	}
	if !found {
		t.Fatalf("partner list does not return the RD-file fields: %d items", len(page.Items))
	}
}

// สมุดที่รูปแบบการเชื่อมบัญชีเลือกไว้ ลบ/เปลี่ยนรหัสไม่ได้ (409) จนกว่าจะลบรูปแบบการเชื่อมนั้น
func TestJournalBookInUseByMapping(t *testing.T) {
	f := newPGIntegrityFixture(t)
	created := f.run(Command{Resource: "journal-books", Action: "create", Master: &Master{Kind: "journal-books", Code: "สมุดขายส่ง", Name: "สมุดรายวันขายส่ง", BookType: BookTypeSales, IsActive: true}})
	book := f.books()["สมุดขายส่ง"]
	if book.ID != created.ID {
		t.Fatalf("book id = %s, want %s", book.ID, created.ID)
	}
	mapping := f.run(Command{Resource: "mappings", Action: "create", Master: &Master{Code: "MAP-SALE-01", Name: "ขายวัสดุก่อสร้างเงินเชื่อ", BookCode: "สมุดขายส่ง", IsActive: true}})
	user := f.failCode(Command{Resource: "journal-books", Action: "delete", ID: book.ID, Version: book.Version}, "journal_book_in_use_delete", "code")
	if user.HTTPStatus() != http.StatusConflict || !strings.Contains(user.Message, "รูปแบบการเชื่อมบัญชี") {
		t.Fatalf("in-use by mapping: status %d message %q", user.HTTPStatus(), user.Message)
	}
	recoded := book
	recoded.Code = "สมุดขายใหม่"
	f.failCode(Command{Resource: "journal-books", Action: "update", ID: book.ID, Version: book.Version, Master: &recoded}, "journal_book_in_use_code", "code")

	var m Master
	if err := f.db.QueryRow(`SELECT version FROM gl_records WHERE company='C' AND kind='mappings' AND id=$1`, mapping.ID).Scan(&m.Version); err != nil {
		t.Fatal(err)
	}
	f.run(Command{Resource: "mappings", Action: "delete", ID: mapping.ID, Version: m.Version})
	book = f.books()["สมุดขายส่ง"]
	f.run(Command{Resource: "journal-books", Action: "delete", ID: book.ID, Version: book.Version})
	var live int
	if err := f.db.QueryRow(`SELECT COUNT(*) FROM gl_records WHERE company='C' AND kind='journal-books' AND code='สมุดขายส่ง' AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`).Scan(&live); err != nil || live != 0 {
		t.Fatalf("book still live after mapping removed: %d (err=%v)", live, err)
	}
}

// review 2026-09-24: เลขผู้เสียภาษีเก่าในทะเบียนคู่ค้าที่ผิดหลักตรวจสอบ — แถวที่เว้นเลขไว้ต้องถูกปฏิเสธตั้งแต่บันทึกร่าง
// ด้วยข้อความที่ชี้ไปแก้ทะเบียนคู่ค้า (เดิมบันทึกร่างผ่าน แล้วติดตอนผ่านรายการด้วยข้อความที่พาวนเติมเลขเดิม)
// + วันที่จ่ายภาษีหักเป็นปี พ.ศ. ถูกปฏิเสธ (เดิมรับไว้ แล้วรายการหายจาก ภ.ง.ด./ไฟล์ยื่น)
func TestPostgresRegistryTaxIDAndBuddhistPaymentDate(t *testing.T) {
	f := newPGIntegrityFixture(t)
	partner := SubledgerPartner{Code: "SUPP-TH-001", Name: "บริษัท ปูนไทยค้าส่ง จำกัด", TaxID: "0105558012349", IsSupplier: true, IsActive: true}
	seed := Journal{DocNo: "JV6901-001", Date: "2026-01-10", BookCode: "JV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Description: "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. เป็นเงินเชื่อ",
		Lines:   []Line{{AccountCode: "5000", Debit: "1000", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "1000"}},
		Details: &JournalDetails{Partners: []SubledgerPartner{partner}}}
	f.run(Command{Resource: "journals", Action: "create", Journal: &seed})
	// ทะเบียนที่บันทึกก่อนมีกฎหลักตรวจสอบ (เลขเดิมใน glseed) — เขียนลง PostgreSQL ตรง ๆ
	res, err := f.db.Exec(`UPDATE gl_subledger_partners SET payload = jsonb_set(payload, '{tax_id}', '"0105558002001"') WHERE company = 'C' AND code = 'SUPP-TH-001'`)
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := res.RowsAffected(); n != 1 {
		t.Fatalf("registry inject touched %d rows", n)
	}
	count := func(doc string) int {
		t.Helper()
		var n int
		if err := f.db.QueryRow(`SELECT COUNT(*) FROM gl_records WHERE company='C' AND kind='journals' AND code=$1`, doc).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	vat := vatItem("V1", 1, 1, "IV-6901-501", "2026-01-10", 2026, 1, 1, "1000", nil)
	vat.PartnerCode, vat.PartnerTaxID, vat.PartnerBranchNo = "SUPP-TH-001", "", ""
	vatJournal := Journal{DocNo: "PV6901-501", Date: "2026-01-10", BookCode: "PV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Description: "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. จำนวน 10 ถุง เป็นเงินเชื่อ",
		Lines:   []Line{{AccountCode: "5000", Debit: "1070", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "1070"}},
		Details: &JournalDetails{Vats: []SubledgerVat{vat}}}
	f.failCode(Command{Resource: "journals", Action: "create", Journal: &vatJournal}, "vat_partner_registry_tax_id_checksum", "partner_tax_id")
	if n := count("PV6901-501"); n != 0 {
		t.Fatalf("VAT voucher with a bad registry tax id was written: %d", n)
	}

	wht := SubledgerWithholding{ID: "W1", Direction: 1, FormType: "PND53", PartnerCode: "SUPP-TH-001", PaymentDate: "2026-01-10", IncomeType: "3_tres", Description: "ค่าขนส่งสินค้า", Condition: 1, Rate: "1", BaseAmount: "1000"}
	whtJournal := Journal{DocNo: "PV6901-502", Date: "2026-01-10", BookCode: "PV", FiscalYear: "2026", Kind: "manual", BranchCode: "B1", Description: "จ่ายค่าขนส่งสินค้า หักภาษี ณ ที่จ่าย 1%",
		Lines:   []Line{{AccountCode: "5000", Debit: "1000", Credit: "0"}, {AccountCode: "1000", Debit: "0", Credit: "1000"}},
		Details: &JournalDetails{Withholdings: []SubledgerWithholding{wht}}}
	f.failCode(Command{Resource: "journals", Action: "create", Journal: &whtJournal}, "wht_partner_registry_tax_id_checksum", "payee_tax_id")
	if n := count("PV6901-502"); n != 0 {
		t.Fatalf("withholding voucher with a bad registry tax id was written: %d", n)
	}

	// แก้ทะเบียนแล้ว บันทึกได้และเติมเลขใหม่; วันที่จ่ายปี พ.ศ. ถูกปฏิเสธ
	if _, err := f.db.Exec(`UPDATE gl_subledger_partners SET payload = jsonb_set(payload, '{tax_id}', '"0105558012349"') WHERE company = 'C' AND code = 'SUPP-TH-001'`); err != nil {
		t.Fatal(err)
	}
	whtJournal.Details.Withholdings[0].PaymentDate = "2569-01-10"
	f.failCode(Command{Resource: "journals", Action: "create", Journal: &whtJournal}, "wht_payment_date_buddhist", "payment_date")
	whtJournal.Details.Withholdings[0].PaymentDate = "2026-01-10"
	created := f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &whtJournal}).ID)
	if w := created.Details.Withholdings[0]; w.PayeeTaxID != "0105558012349" || w.PaymentDate != "2026-01-10" {
		t.Fatalf("stored withholding = %+v", w)
	}
}
