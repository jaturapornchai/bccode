package handlers

import (
	"errors"
	"testing"

	"smlcloudplatform/internal/rdform"
	"smlcloudplatform/internal/taxaddress"
)

func computeDoc(t *testing.T, code string, doc rdform.Document) rdform.Document {
	t.Helper()
	if err := rdform.Validate(code, doc); err != nil {
		t.Fatalf("%s input invalid: %v", code, err)
	}
	if err := computeTaxForm(code, &doc); err != nil {
		t.Fatalf("%s compute: %v", code, err)
	}
	// ผลคำนวณต้องยังเป็นเอกสารที่ถูกต้องของแบบ (ไม่มีช่องแปลกปลอม) และพิมพ์ได้
	if err := rdform.Validate(code, doc); err != nil {
		t.Fatalf("%s computed document invalid: %v", code, err)
	}
	return doc
}

func expectValues(t *testing.T, got map[string]string, want map[string]string) {
	t.Helper()
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
}

func TestComputePP30(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]string
		want map[string]string
	}{
		{"payable", map[string]string{"sales_amount": "1,200.00", "sales_zero_rate": "100", "sales_exempt": "100", "output_tax": "70", "purchase_amount": "500", "input_tax": "35"},
			map[string]string{"sales_taxable": "1000.00", "tax_payable": "35.00", "tax_excess": "", "net_payable": "35.00", "net_result": "payable", "total_payable": "35.00", "total_excess": ""}},
		{"excess plus brought forward", map[string]string{"output_tax": "50", "input_tax": "70", "excess_brought_forward": "10"},
			map[string]string{"tax_payable": "", "tax_excess": "20.00", "net_excess": "30.00", "net_result": "excess", "total_excess": "30.00", "total_payable": ""}},
		{"forward larger than payable", map[string]string{"output_tax": "70", "input_tax": "50", "excess_brought_forward": "30"},
			map[string]string{"tax_payable": "20.00", "net_payable": "", "net_excess": "10.00", "net_result": "excess"}},
		{"surcharge above excess", map[string]string{"output_tax": "50", "input_tax": "60", "surcharge": "12", "penalty": "3"},
			map[string]string{"net_excess": "10.00", "total_payable": "5.00", "total_excess": ""}},
		{"payable with penalties", map[string]string{"output_tax": "100", "input_tax": "40", "surcharge": "1.50", "penalty": "200"},
			map[string]string{"net_payable": "60.00", "total_payable": "261.50"}},
		{"equal", map[string]string{"output_tax": "70", "input_tax": "70"},
			map[string]string{"tax_payable": "0.00", "tax_excess": "", "net_result": "", "total_payable": ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := computeDoc(t, "pp30", rdform.Document{Values: tc.in})
			expectValues(t, doc.Values, tc.want)
		})
	}
}

func TestComputePP30Branches(t *testing.T) {
	doc := computeDoc(t, "pp30", rdform.Document{Values: map[string]string{}, Rows: []map[string]string{
		{"branch": "00000", "sales_amount": "1000", "output_tax": "70", "purchase_amount": "400", "input_tax": "28"},
		{"branch": "00001", "sales_amount": "500", "sales_exempt": "100", "output_tax": "28", "input_tax": "0"},
	}})
	expectValues(t, doc.Rows[1], map[string]string{"sales_taxable": "400.00", "tax_net": "28.00"})
	expectValues(t, doc.Values, map[string]string{"sales_amount": "1500.00", "sales_exempt": "100.00", "sales_taxable": "1400.00",
		"output_tax": "98.00", "input_tax": "28.00", "tax_payable": "70.00", "net_payable": "70.00"})
	// ยื่นปกติ: ช่องยื่นเพิ่มเติมต้องว่าง ไม่ใช่ "0.00" (UAT 2026-09-24)
	for _, k := range pp30AdditionalColumns {
		if v, ok := doc.Values[k]; ok {
			t.Errorf("normal filing %s = %q, want blank", k, v)
		}
	}

	// ยื่นเพิ่มเติม: สาขาที่กรอกยอดขายต่ำไป รวมขึ้นหน้าแบบ ช่องที่ไม่มีสาขากรอกยังว่าง
	extra := computeDoc(t, "pp30", rdform.Document{Values: map[string]string{}, Rows: []map[string]string{
		{"branch": "00000", "sales_amount": "1000", "output_tax": "70", "add_sales_under": "100"},
		{"branch": "00001", "sales_amount": "500", "output_tax": "35", "add_sales_under": "50.50"},
	}})
	expectValues(t, extra.Values, map[string]string{"add_sales_under": "150.50", "add_purchase_over": ""})
	if _, ok := extra.Values["add_purchase_over"]; ok {
		t.Error("add_purchase_over must stay blank when no branch filled it")
	}

	// สาขาล้างยอดยื่นเพิ่มเติมออกหมด (กลับเป็นยื่นปกติ) — ยอดรวมเดิม 150.50 บนหน้าแบบต้องหาย ไม่ค้างไว้
	for _, r := range extra.Rows {
		delete(r, "add_sales_under")
	}
	cleared := computeDoc(t, "pp30", extra)
	if v, ok := cleared.Values["add_sales_under"]; ok {
		t.Errorf("stale additional-filing total after branches cleared = %q, want blank", v)
	}
	expectValues(t, cleared.Values, map[string]string{"sales_amount": "1500.00", "output_tax": "105.00"})
}

// TestPrepareTaxDocumentRecomputesTotals - บันทึก/พิมพ์ใช้ยอดรวมที่คำนวณใหม่เสมอ: ยอดรวมที่ถูกแก้จาก browser ต้องไม่รอด
func TestPrepareTaxDocumentRecomputesTotals(t *testing.T) {
	tampered := rdform.Document{Values: map[string]string{"total_income": "1.00", "total_tax": "1.00", "total_payable": "1.00"}, Rows: []map[string]string{
		{"name": "บริษัท รุ่งเรืองขนส่ง จำกัด", "l1_amount": "100000", "l1_tax": "3000.00", "l2_amount": "66353.50", "l2_tax": "3317.07"},
	}}
	doc, err := prepareTaxDocument("pnd53", tampered)
	if err != nil {
		t.Fatal(err)
	}
	expectValues(t, doc.Values, map[string]string{"total_income": "166353.50", "total_tax": "6317.07", "total_payable": "6317.07"})
	if err := rdform.Validate("pnd53", doc); err != nil {
		t.Fatalf("prepared document must still be valid: %v", err)
	}
	if _, err := prepareTaxDocument("pnd53", rdform.Document{Rows: []map[string]string{{"l1_tax": "abc"}}}); !errors.Is(err, rdform.ErrInvalidValue) {
		t.Fatalf("bad money must be rejected, got %v", err)
	}
}

// TestPayeeKeyWithoutIdentity - แถวที่ไม่มีทั้งเลขผู้เสียภาษีและชื่อ ห้ามรวมเป็นผู้มีเงินได้รายเดียวกัน
func TestPayeeKeyWithoutIdentity(t *testing.T) {
	rows := []TaxWithholdingRow{
		{JournalID: "J1", PartnerCode: "S001", RatePercent: "3", BaseAmount: "1000.00", WhtAmount: "30.00", DocDate: "2026-09-01"},
		{JournalID: "J2", PartnerCode: "S002", RatePercent: "3", BaseAmount: "2000.00", WhtAmount: "60.00", DocDate: "2026-09-02"},
		{JournalID: "J3", PartnerCode: "S001", RatePercent: "3", BaseAmount: "500.00", WhtAmount: "15.00", DocDate: "2026-09-03"},
		{JournalID: "J4", RatePercent: "3", BaseAmount: "100.00", WhtAmount: "3.00", DocDate: "2026-09-04"},
		{JournalID: "J5", RatePercent: "3", BaseAmount: "200.00", WhtAmount: "6.00", DocDate: "2026-09-05"},
	}
	out := payeeRows(rows, false)
	if len(out) != 4 {
		t.Fatalf("payees = %d, want 4 (S001 รวมกัน, S002, J4, J5 แยก): %+v", len(out), out)
	}
	expectValues(t, out[0], map[string]string{"l1_amount": "1000.00", "l2_amount": "500.00"})
	expectValues(t, out[2], map[string]string{"l1_amount": "100.00", "l2_amount": ""})

	annual := pnd2Rows([]TaxWithholdingRow{
		{JournalID: "J1", PartnerCode: "A", IncomeType: "40_4a", BaseAmount: "10.00", WhtAmount: "1.50"},
		{JournalID: "J2", PartnerCode: "B", IncomeType: "40_4a", BaseAmount: "20.00", WhtAmount: "3.00"},
	}, true)
	if len(annual) != 2 {
		t.Fatalf("ภ.ง.ด.2ก rows = %d, want 2 (คนละคู่ค้า)", len(annual))
	}
}

func TestComputeWithholdingCover(t *testing.T) {
	doc := computeDoc(t, "pnd53", rdform.Document{Values: map[string]string{"surcharge": "10"}, Rows: []map[string]string{
		{"name": "บริษัท ก จำกัด", "l1_amount": "100000", "l1_tax": "3000", "l2_amount": "20000.50", "l2_tax": "1000.03"},
		{"name": "บริษัท ข จำกัด", "l1_amount": "5000", "l1_tax": "250"},
	}})
	expectValues(t, doc.Values, map[string]string{"total_income": "125000.50", "total_tax": "4250.03", "total_payable": "4260.03"})

	// ยื่นทางสื่อบันทึก (ไม่มีใบแนบในระบบ) — ยอดที่ผู้ใช้กรอกต้องไม่ถูกทับ
	manual := computeDoc(t, "pnd3", rdform.Document{Values: map[string]string{"total_income": "900", "total_tax": "27"}})
	expectValues(t, manual.Values, map[string]string{"total_income": "900", "total_tax": "27", "total_payable": "27.00"})
}

func TestComputePnd2Cover(t *testing.T) {
	rows := []map[string]string{
		{"income_type": "interest", "name": "นาย ก", "l1_amount": "10000", "l1_tax": "1500"},
		{"income_type": "interest", "name": "นาย ข", "l1_amount": "2000", "l1_tax": "300"},
		{"income_type": "dividend", "name": "นาย ค", "l1_amount": "5000", "l1_tax": "500"},
	}
	doc := computeDoc(t, "pnd2", rdform.Document{Values: map[string]string{"surcharge": "5"}, Rows: rows})
	expectValues(t, doc.Values, map[string]string{"interest_count": "2", "interest_income": "12000.00", "interest_tax": "1800.00",
		"dividend_count": "1", "royalty_count": "0", "total_count": "3", "total_income": "17000.00", "total_tax": "2300.00", "total_payable": "2305.00"})

	// แถวที่ยังไม่เลือกประเภทเงินได้ไม่อยู่ในบรรทัดใด แต่ต้องอยู่ใน 6. รวม (ยอดภาษีที่นำส่งไม่ตกหล่น)
	blank := computeDoc(t, "pnd2", rdform.Document{Values: map[string]string{}, Rows: append(append([]map[string]string{}, rows...),
		map[string]string{"income_type": "", "name": "นาย ง", "l1_amount": "100", "l1_tax": "15"})})
	expectValues(t, blank.Values, map[string]string{"interest_count": "2", "total_count": "4", "total_income": "17100.00", "total_tax": "2315.00"})

	annual := computeDoc(t, "pnd2a", rdform.Document{Values: map[string]string{}, Rows: rows})
	expectValues(t, annual.Values, map[string]string{"interest_count": "2", "total_count": "3", "total_tax": "2300.00"})
	if _, ok := annual.Values["total_payable"]; ok {
		t.Error("ภ.ง.ด.2ก ไม่มีช่อง total_payable")
	}
}

func TestComputePP36AndPBT40(t *testing.T) {
	pp36 := computeDoc(t, "pp36", rdform.Document{Values: map[string]string{"vat_amount": "700", "surcharge": "7"}})
	expectValues(t, pp36.Values, map[string]string{"total_payable": "707.00", "vat_amount_text": "เจ็ดร้อยบาทถ้วน", "total_payable_text": "เจ็ดร้อยเจ็ดบาทถ้วน"})

	pbt := computeDoc(t, "pbt40", rdform.Document{Values: map[string]string{"local_tax": "30", "surcharge": "5"}, Sheets: []map[string]string{
		{"biz8_real_estate_receipts": "10000", "biz8_real_estate_tax": "300"},
		{"biz8_real_estate_receipts": "5000", "biz8_real_estate_tax": "150", "biz10_factoring_tax": "50"},
	}})
	expectValues(t, pbt.Sheets[1], map[string]string{"est_total_tax": "200.00"})
	expectValues(t, pbt.Values, map[string]string{"biz8_real_estate_receipts": "15000.00", "biz8_real_estate_tax": "450.00", "biz10_factoring_tax": "50.00",
		"total_sbt_tax": "500.00", "total_with_surcharge_penalty": "505.00", "total_payable": "535.00"})
}

func TestComputeRejectsBadMoneyWithRow(t *testing.T) {
	doc := rdform.Document{Values: map[string]string{}, Rows: []map[string]string{{"l1_amount": "1"}, {"l1_amount": "abc"}}}
	err := computeTaxForm("pnd53", &doc)
	var fe *rdform.FieldError
	if !errors.As(err, &fe) || fe.Row != 2 || fe.Key != "l1_amount" || !errors.Is(err, rdform.ErrInvalidValue) {
		t.Fatalf("err = %v, want row 2 l1_amount", err)
	}
}

func TestPayeeRowsForAttachments(t *testing.T) {
	rows := []TaxWithholdingRow{}
	for i := 0; i < 4; i++ {
		rows = append(rows, TaxWithholdingRow{TaxID: "0-1055-12345-67-8", PartnerName: "บริษัท รุ่งเรืองขนส่ง จำกัด", Address: "99/1 ถนนพระราม 2   แขวงแสมดำ เขตบางขุนเทียน กรุงเทพมหานคร 10150",
			Description: "ค่าขนส่ง", RatePercent: "1.00", BaseAmount: "1000.00", WhtAmount: "10.00", Condition: 1, PaidDate: "2026-09-15"})
	}
	rows = append(rows, TaxWithholdingRow{PartnerName: "นายสมชาย ใจดี", IncomeType: "40_2", RatePercent: "3", BaseAmount: "500.00", WhtAmount: "15.00", DocDate: "2026-09-30T00:00:00Z"})

	out := payeeRows(rows, false)
	if len(out) != 3 {
		t.Fatalf("rows = %d, want 3 (3 รายการ/แถว แล้วขึ้นแถวใหม่ของรายเดิม)", len(out))
	}
	expectValues(t, out[0], map[string]string{"tax_id": "0105512345678", "l3_date": "15/09/2569", "l1_rate": "1", "l1_income_type": "ค่าขนส่ง"})
	expectValues(t, out[1], map[string]string{"l1_amount": "1000.00", "l2_amount": ""})
	expectValues(t, out[2], map[string]string{"tax_id": "", "l1_income_type": "ค่านายหน้า 40(2)", "l1_date": "30/09/2569", "l1_condition": "1"})
	if out[0]["address2"] == "" {
		t.Error("ที่อยู่ยาวต้องแยก 2 บรรทัด")
	}
	doc := computeDoc(t, "pnd53", rdform.Document{Values: map[string]string{}, Rows: out})
	expectValues(t, doc.Values, map[string]string{"total_income": "4500.00", "total_tax": "55.00"})

	person := payeeRows(rows[4:], true)
	expectValues(t, person[0], map[string]string{"name": "นายสมชาย", "surname": "ใจดี"})

	// review 2026-09-24: มาตรา 3 เตรส ที่ไม่มีคำอธิบาย ไม่ใช่ "ค่าอะไร" ([F53] ช่อง 13) → ว่าง + หมายเหตุ ไม่เขียนชื่อมาตราลงแบบ
	tres := payeeRows([]TaxWithholdingRow{{TaxID: "0105558012349", PartnerName: "บริษัท ขนส่งไทยเร็ว จำกัด", IncomeType: "3_tres", RatePercent: "1", BaseAmount: "1000.00", WhtAmount: "10.00", Condition: 1, PaidDate: "2026-09-15"},
		{TaxID: "0105558012349", PartnerName: "บริษัท ขนส่งไทยเร็ว จำกัด", IncomeType: "3_tres", Description: "ค่าขนส่ง", RatePercent: "1", BaseAmount: "500.00", WhtAmount: "5.00", Condition: 1, PaidDate: "2026-09-16"}}, false)
	expectValues(t, tres[0], map[string]string{"l1_income_type": "", "l2_income_type": "ค่าขนส่ง"})
	if notes := missingIncomeTypeNotes("pnd53", tres); len(notes) != 1 || notes[0].Key != "tax_form_note_missing_income_type" || notes[0].Count != 1 {
		t.Errorf("missing income type notes = %+v", notes)
	}
	computeDoc(t, "pnd3", rdform.Document{Values: map[string]string{}, Rows: person})
}

func TestPnd2Rows(t *testing.T) {
	rows := []TaxWithholdingRow{
		{TaxID: "1234567890123", PartnerName: "นายเอ บี", IncomeType: "40_4a", RatePercent: "15", BaseAmount: "1000.00", WhtAmount: "150.00", PaidDate: "2026-01-10"},
		{TaxID: "1234567890123", PartnerName: "นายเอ บี", IncomeType: "40_4a", RatePercent: "15", BaseAmount: "2000.00", WhtAmount: "300.00", PaidDate: "2026-06-10"},
		{TaxID: "1234567890123", PartnerName: "นายเอ บี", IncomeType: "40_4b_1", RatePercent: "10", BaseAmount: "500.00", WhtAmount: "50.00", PaidDate: "2026-06-10"},
		{PartnerName: "นายซี", IncomeType: "40_3", RatePercent: "15", BaseAmount: "100.00", WhtAmount: "15.00", PaidDate: "2026-06-11"},
	}
	monthly := pnd2Rows(rows, false)
	if len(monthly) != 4 || monthly[3]["income_type"] != "royalty" || monthly[0]["l1_date"] != "10/01/2569" {
		t.Fatalf("monthly = %+v", monthly)
	}
	computeDoc(t, "pnd2", rdform.Document{Values: map[string]string{}, Rows: monthly})

	annual := pnd2Rows(rows, true)
	if len(annual) != 3 {
		t.Fatalf("annual rows = %d, want 3 (ดอกเบี้ยของรายเดียวกันรวมเป็นแถวเดียว)", len(annual))
	}
	expectValues(t, annual[0], map[string]string{"income_type": "interest", "l1_amount": "3000.00", "l1_tax": "450.00", "l1_rate": "15"})
	// review 2026-09-24: ภ.ง.ด.2ก ไม่มีบรรทัด 40(3) — ไม่เดาเป็น "40(4) อื่น ๆ" ให้ผู้ใช้เลือกเอง
	if annual[2]["income_type"] != "" {
		t.Errorf("ภ.ง.ด.2ก ค่าลิขสิทธิ์ → ว่าง (ผู้ใช้เลือก), got %q", annual[2]["income_type"])
	}
	annualDoc := computeDoc(t, "pnd2a", rdform.Document{Values: map[string]string{}, Rows: annual})
	expectValues(t, annualDoc.Values, map[string]string{"total_count": "3", "total_income": "3600.00"})
	if notes := missingIncomeTypeNotes("pnd2a", annual); len(notes) != 1 || notes[0].Count != 1 {
		t.Errorf("missing income type notes = %+v", notes)
	}
	// ภ.ง.ด.2 รับเฉพาะ 40(3)/40(4) (Format กลาง ภ.ง.ด.2 ช่อง 14) — 40(2)/มาตรา 3 เตรส/อื่น ๆ ไม่ถูกเขียนเป็น 40(4) อื่น ๆ เงียบ ๆ
	for _, code := range []string{"40_2", "3_tres", "other", "40_1"} {
		if kind := pnd2IncomeType(code, false); kind != "" {
			t.Errorf("pnd2IncomeType(%q) = %q, want empty", code, kind)
		}
	}
}

func TestTaxFormTextHelpers(t *testing.T) {
	// ไม่รู้คำนำหน้า: คำสุดท้ายเป็นชื่อสกุล (แยกคำนำหน้าแยกคำกับชื่อกลางไม่ออก)
	if a, b := splitPersonName("นางสาว มาลี ศรีสุข", ""); a != "นางสาว มาลี" || b != "ศรีสุข" {
		t.Errorf("splitPersonName = %q %q", a, b)
	}
	if a, b := splitPersonName("สมชาย", "นาย"); a != "สมชาย" || b != "" {
		t.Errorf("splitPersonName single = %q %q", a, b)
	}
	// review 2026-09-24: Format กลาง ภ.ง.ด.3 ช่อง 8 / ภ.ง.ด.2 ช่อง 9 — ชื่อกลางอยู่ในช่องชื่อสกุล ("ชื่อกลาง+1 ช่องว่าง+นามสกุล")
	for _, tc := range []struct{ full, title, name, surname string }{
		{"Mr. John Michael Smith", "Mr.", "Mr. John", "Michael Smith"},
		{"นายสมชาย ณ อยุธยา", "นาย", "นายสมชาย", "ณ อยุธยา"},
		{"นาย สมชาย ใจดี", "นาย", "นาย สมชาย", "ใจดี"},
		{"นายสมชาย ใจดี", "นาย", "นายสมชาย", "ใจดี"},
		{"ว่าที่ ร.ต. สมชาย ใจดี", "ว่าที่ ร.ต.", "ว่าที่ ร.ต. สมชาย", "ใจดี"},
		{"สมชาย ใจดี", "-", "สมชาย", "ใจดี"},
	} {
		if a, b := splitPersonName(tc.full, tc.title); a != tc.name || b != tc.surname {
			t.Errorf("splitPersonName(%q, %q) = %q %q, want %q %q", tc.full, tc.title, a, b, tc.name, tc.surname)
		}
	}
	if a, b := splitAddress("สั้น", 60); a != "สั้น" || b != "" {
		t.Errorf("splitAddress short = %q %q", a, b)
	}
	if thaiDate("bad") != "" || thaiDate("2026-01-05") != "05/01/2569" {
		t.Error("thaiDate")
	}
	if rateText("1.50") != "1.5" || rateText("x") != "" {
		t.Error("rateText")
	}
	for _, tc := range []struct {
		values map[string]string
		seq    int
		ok     bool
	}{
		{map[string]string{"filing_type": "normal"}, 0, true},
		{map[string]string{"filing_type": "additional", "additional_no": "2"}, 2, true},
		{map[string]string{"filing_type": "additional", "additional_no": "0"}, 0, false},
		{map[string]string{"filing_type": "additional"}, 0, false},
	} {
		if seq, ok := filingSeq(tc.values); seq != tc.seq || ok != tc.ok {
			t.Errorf("filingSeq(%v) = %d %v", tc.values, seq, ok)
		}
	}
}

// ภ.ง.ด.50 รายการที่ 1 ข้อ 3–6 และ ภ.ง.ด.51 รายการที่ 2 ข้อ 5–8 ตามคู่มือวิธีกรอกแบบของกรมสรรพากร (docs/kms/21-thai-tax-form-references.md §4)
func TestComputeCorporateIncomeTax(t *testing.T) {
	cases := []struct {
		name, code string
		in, want   map[string]string
	}{
		{"pnd50 payable", "pnd50", map[string]string{"tax_computed": "150,000.00", "less_wht": "12000", "less_pnd51_paid": "60000.50", "surcharge": "900"},
			map[string]string{"less_total": "72000.50", "tax_balance": "77999.50", "tax_balance_type": "payable", "tax_net": "78899.50", "tax_net_type": "payable"}},
		{"pnd50 overpaid", "pnd50", map[string]string{"tax_computed": "10000", "less_wht": "4000", "less_pnd51_paid": "9000"},
			map[string]string{"less_total": "13000.00", "tax_balance": "3000.00", "tax_balance_type": "overpaid", "tax_net": "3000.00", "tax_net_type": "overpaid"}},
		{"pnd50 zero is payable", "pnd50", map[string]string{"tax_computed": "0.1", "less_wht": "0.1"},
			map[string]string{"less_total": "0.10", "tax_balance": "0.00", "tax_balance_type": "payable"}},
		// ยังไม่กรอกภาษีที่คำนวณได้: รวมเครดิตได้ แต่ห้ามขึ้น "ชำระไว้เกิน"
		{"pnd50 tax not entered", "pnd50", map[string]string{"less_wht": "5000"},
			map[string]string{"less_total": "5000.00", "tax_balance": "", "tax_balance_type": "", "tax_net": ""}},
		{"pnd50 no credits", "pnd50", map[string]string{"tax_computed": "800"},
			map[string]string{"less_total": "", "tax_balance": "800.00", "tax_balance_type": "payable", "tax_net": "800.00"}},
		{"pnd51 payable with surcharge", "pnd51", map[string]string{"r2_4_tax_computed": "45000", "r2_5_1_wht": "3000", "r2_5_3_prior_pnd51_paid": "20000", "r2_7_surcharge": "4400"},
			map[string]string{"r2_5_total_credits": "23000.00", "r2_6_balance": "22000.00", "r2_6_sign": "payable", "r2_8_total": "26400.00", "r2_8_sign": "payable"}},
		{"pnd51 overpaid", "pnd51", map[string]string{"r2_4_tax_computed": "1000", "r2_5_1_wht": "1500.25"},
			map[string]string{"r2_5_total_credits": "1500.25", "r2_6_balance": "500.25", "r2_6_sign": "overpaid", "r2_8_total": "500.25", "r2_8_sign": "overpaid"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := computeDoc(t, tc.code, rdform.Document{Values: tc.in})
			expectValues(t, doc.Values, tc.want)
		})
	}
}

func TestMissingTaxIDNotesRecheck(t *testing.T) {
	rows := []map[string]string{{"tax_id": "0105566123456"}, {"tax_id": ""}, {"tax_id": "123"}}
	notes := missingTaxIDNotes(rows)
	if len(notes) != 1 || notes[0].Key != "tax_form_note_missing_taxid" || notes[0].Count != 2 {
		t.Fatalf("missing notes = %+v", notes)
	}
	rows[1]["tax_id"], rows[2]["tax_id"] = "0-1055-66123-45-6", "0105566123456"
	if notes := missingTaxIDNotes(rows); notes == nil || len(notes) != 0 {
		t.Fatalf("after fixing every row the note must disappear (empty, not nil): %+v", notes)
	}
}

// ทะเบียนบริษัทมีที่อยู่สำหรับภาษี → ล้างที่อยู่ที่ยกจากฉบับก่อนทั้งชุดแล้วใส่ของทะเบียน (ไม่ปนสองที่อยู่); ทะเบียนว่าง → คงค่าที่ยกมา;
// โทรศัพท์ใส่เฉพาะแบบที่มีช่องโทรศัพท์ (ภ.พ.30 มี, ภ.ง.ด.53 ไม่มี)
func TestApplyRegistryAddress(t *testing.T) {
	copied := map[string]string{"addr_no": "9/9", "addr_soi": "ซอยเดิม", "addr_road": "ถนนเดิม", "addr_province": "นนทบุรี", "phone": "02-000-0000", "signer_name": "นายกรรมการ บริษัท"}
	registry := CompanyHeader{Name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", Phone: "02-123-4567",
		Address: &taxaddress.Address{No: "88/12", Road: "ลาดหลุมแก้ว", Subdistrict: "คูบางหลวง", District: "ลาดหลุมแก้ว", Province: "ปทุมธานี", Postcode: "12140"}}

	pp30, err := newFormFiller("pp30")
	if err != nil {
		t.Fatal(err)
	}
	if !pp30.keys["phone"] || !pp30.keys["addr_soi"] {
		t.Fatal("pp30 spec must have phone and address keys")
	}
	for k, v := range copied {
		pp30.set(k, v)
	}
	pp30.applyRegistryAddress(registry)
	expectValues(t, pp30.values, map[string]string{"addr_no": "88/12", "addr_road": "ลาดหลุมแก้ว", "addr_subdistrict": "คูบางหลวง",
		"addr_district": "ลาดหลุมแก้ว", "addr_province": "ปทุมธานี", "addr_postcode": "12140", "addr_soi": "", "phone": "02-123-4567",
		"signer_name": "นายกรรมการ บริษัท"})
	if _, ok := pp30.values["addr_soi"]; ok {
		t.Fatal("copied soi from the last filing must be cleared when the registry has an address")
	}

	pnd53, err := newFormFiller("pnd53")
	if err != nil {
		t.Fatal(err)
	}
	if pnd53.keys["phone"] {
		t.Fatal("pnd53 spec unexpectedly has phone — update this test")
	}
	pnd53.applyRegistryAddress(registry)
	if _, ok := pnd53.values["phone"]; ok || pnd53.values["addr_no"] != "88/12" {
		t.Fatalf("pnd53 values %v", pnd53.values)
	}

	// ทะเบียนว่าง → พฤติกรรมเดิม (ค่าที่ยกจากฉบับก่อนอยู่ครบ)
	kept, err := newFormFiller("pp30")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range copied {
		kept.set(k, v)
	}
	kept.applyRegistryAddress(CompanyHeader{Name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด"})
	expectValues(t, kept.values, copied)

	// ฉบับก่อนยื่นในนามสาขา → ทะเบียน (สำนักงานใหญ่) ไม่ทับที่อยู่/โทรศัพท์ของสาขา; สำนักงานใหญ่ 00000 ยังใช้ทะเบียน
	branch, err := newFormFiller("pp30")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range copied {
		branch.set(k, v)
	}
	branch.set("branch_no", "00001")
	branch.applyRegistryAddress(registry)
	expectValues(t, branch.values, map[string]string{"addr_no": "9/9", "addr_soi": "ซอยเดิม", "addr_road": "ถนนเดิม", "addr_province": "นนทบุรี",
		"phone": "02-000-0000", "branch_no": "00001"})
	head, err := newFormFiller("pp30")
	if err != nil {
		t.Fatal(err)
	}
	head.set("branch_no", "00000")
	head.applyRegistryAddress(registry)
	if head.values["addr_no"] != "88/12" || head.values["phone"] != "02-123-4567" {
		t.Fatalf("head office filing must use the registry: %v", head.values)
	}
}
