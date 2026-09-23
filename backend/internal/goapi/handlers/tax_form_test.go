package handlers

import (
	"errors"
	"testing"

	"smlcloudplatform/internal/rdform"
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
	if annual[2]["income_type"] != "other_404" {
		t.Errorf("ภ.ง.ด.2ก ไม่มีค่าลิขสิทธิ์ → other_404, got %q", annual[2]["income_type"])
	}
	computeDoc(t, "pnd2a", rdform.Document{Values: map[string]string{}, Rows: annual})
}

func TestTaxFormTextHelpers(t *testing.T) {
	if a, b := splitPersonName("นางสาว มาลี ศรีสุข"); a != "นางสาว มาลี" || b != "ศรีสุข" {
		t.Errorf("splitPersonName = %q %q", a, b)
	}
	if a, b := splitPersonName("สมชาย"); a != "สมชาย" || b != "" {
		t.Errorf("splitPersonName single = %q %q", a, b)
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
