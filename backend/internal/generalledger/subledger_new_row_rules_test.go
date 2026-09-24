package generalledger

import (
	"net/http"
	"strings"
	"testing"
	"unicode/utf8"
)

// ทะเบียนคู่ค้า: ช่องไม่บังคับสำหรับไฟล์ยื่นภาษี + รหัส 20 ตัวอักษรเฉพาะคู่ค้าใหม่ + หลักตรวจสอบเลขผู้เสียภาษี
func TestValidatePartnerWriteNewRules(t *testing.T) {
	base := func() SubledgerPartner {
		return SubledgerPartner{Code: "CUST-TH-001", Name: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", TaxID: "0105558012349", IsCustomer: true, IsActive: true}
	}
	p := base()
	p.TitleName, p.AddrDistrict, p.AddrProvince, p.AddrPostcode = " บริษัท ", " เขตบางนา ", " กรุงเทพมหานคร ", " 10260 "
	normalizePartner(&p)
	if err := validatePartner(&p); err != nil {
		t.Fatalf("valid RD-file fields must pass: %v", err)
	}
	if p.TitleName != "บริษัท" || p.AddrDistrict != "เขตบางนา" || p.AddrProvince != "กรุงเทพมหานคร" || p.AddrPostcode != "10260" {
		t.Fatalf("RD-file fields not trimmed: %+v", p)
	}
	p = base()
	p.TitleName, p.AddrDistrict, p.AddrProvince = strings.Repeat("ก", 100), strings.Repeat("ก", 50), strings.Repeat("ก", 50)
	if err := validatePartner(&p); err != nil {
		t.Fatalf("limits 100/50/50 Thai runes must pass: %v", err)
	}

	// รหัสคู่ค้า 21 ตัว: คู่ค้าใหม่ไม่รับ (partners.partner_code VARCHAR(20)) แต่คู่ค้าเดิมที่รหัสยาวยังแก้ข้อมูลอื่นได้ (รหัสแก้ไม่ได้)
	p = base()
	p.Code = strings.Repeat("A", 21)
	expectFieldError(t, "new partner code 21 runes", validatePartnerWrite(&p, true), "code_too_long", "partner_code")
	if err := validatePartnerWrite(&p, false); err != nil {
		t.Fatalf("existing long-code partner must stay writable: %v", err)
	}
	p.Code = strings.Repeat("ก", 20)
	if err := validatePartnerWrite(&p, true); err != nil {
		t.Fatalf("20 Thai runes must pass: %v", err)
	}

	p = base()
	p.TaxID = "0105558012340" // หลักตรวจสอบที่ถูกคือ 9
	err := validatePartnerWrite(&p, true)
	expectFieldError(t, "tax id checksum", err, "partner_tax_id_checksum", "tax_id")
	if user, _ := AsUserError(err); user == nil || user.HTTPStatus() != http.StatusBadRequest {
		t.Fatalf("checksum error must be 400: %#v", err)
	}
	p.TaxID = ""
	if err := validatePartnerWrite(&p, true); err != nil {
		t.Fatalf("blank tax id must stay allowed: %v", err)
	}
	for _, id := range []string{"0105558012349", "0105558012357", "0105558001011"} {
		if !checkThaiTaxID(id) {
			t.Errorf("checkThaiTaxID(%s) = false, want true", id)
		}
	}
}

// ช่วงงวดภาษีซื้อแยกตามประเภทเอกสาร (ม.82/9, 82/10, ม.77/1(22), ประกาศฯ ฉบับที่ 4) + ปี พ.ศ. + หลักตรวจสอบ
func TestVatNewRowRulesByDocumentType(t *testing.T) {
	m := &subledgerMutation{scale: 2}
	purchase := func(docType int, date string, year, month int) SubledgerVat {
		v := validSaleVat()
		v.TaxType, v.ClaimStatus, v.DocumentType, v.TaxInvoiceNo, v.TaxInvoiceDate = 1, 1, docType, "DN6901-0007", date
		v.TaxPeriodYear, v.TaxPeriodMonth = year, month
		if docType != 1 {
			v.OriginalInvoiceNo, v.OriginalInvoiceDate = "IV6812-0042", "2025-12-15"
		}
		return v
	}
	saleBE := validSaleVat()
	saleBE.TaxPeriodYear = 2569
	badChecksum := validSaleVat()
	badChecksum.PartnerTaxID = "0-1055-58012-34-0"
	cases := []struct {
		name        string
		v           SubledgerVat
		code, field string // code ว่าง = ต้องผ่าน
	}{
		{"debit note in 6th month after issue", purchase(2, "2026-01-05", 2026, 7), "", ""},
		{"debit note 7th month exceeded", purchase(2, "2026-01-05", 2026, 8), "vat_claim_window_exceeded", "tax_period_month"},
		{"debit note before issue month", purchase(2, "2026-02-01", 2026, 1), "vat_claim_before_invoice_month", "tax_period_month"},
		{"credit note same month", purchase(3, "2026-01-05", 2026, 1), "", ""},
		{"credit note received 12 months later has no 6-month cap", purchase(3, "2026-01-05", 2027, 1), "", ""},
		{"credit note before its month", purchase(3, "2026-02-01", 2026, 1), "vat_credit_note_period_before_note", "tax_period_month"},
		{"credit note Buddhist year", purchase(3, "2026-01-05", 2569, 1), "vat_period_year_buddhist", "tax_period_year"},
		{"sale Buddhist year", saleBE, "vat_period_year_buddhist", "tax_period_year"},
		{"partner tax id checksum", badChecksum, "vat_partner_tax_id_checksum", "partner_tax_id"},
	}
	for _, tc := range cases {
		v := tc.v
		err := m.vat(&v)
		if tc.code == "" {
			if err != nil {
				t.Errorf("%s: unexpected error %v", tc.name, err)
			}
			continue
		}
		expectFieldError(t, tc.name, err, tc.code, tc.field)
		if user, _ := AsUserError(err); user != nil && user.HTTPStatus() != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", tc.name, user.HTTPStatus())
		}
	}
}

// แถวที่บันทึกก่อนมีกฎ: ไม่เปลี่ยน = ไม่ถูกบล็อกตอนแก้ใบร่าง/reconcile; แก้แถวหรือผ่านรายการ = ตรวจ
func TestVatUnchangedRowSkipsNewRules(t *testing.T) {
	legacy := validSaleVat()
	legacy.TaxPeriodYear = 2569 // บันทึกไว้ก่อนมีกฎปี พ.ศ.
	amount := Amount("7000")
	legacy.VatAmount = &amount
	m := &subledgerMutation{scale: 2}
	m.rememberVats(&JournalDetails{Vats: []SubledgerVat{legacy}})

	v := legacy
	if err := m.vat(&v); err != nil {
		t.Fatalf("unchanged legacy row must not block the save: %v", err)
	}
	v = legacy
	v.VatAmount = nil // จอส่งยอดภาษีว่าง → คำนวณได้เท่าเดิม = ไม่เปลี่ยน
	if err := m.vat(&v); err != nil {
		t.Fatalf("recomputed identical row must not block the save: %v", err)
	}
	v = legacy
	v.Remark = "แก้หมายเหตุ"
	expectFieldError(t, "edited legacy row", m.vat(&v), "vat_period_year_buddhist", "tax_period_year")
	v = legacy
	v.ID = "V2"
	expectFieldError(t, "new row", m.vat(&v), "vat_period_year_buddhist", "tax_period_year")
	m.strict = true
	v = legacy
	expectFieldError(t, "posting checks every row", m.vat(&v), "vat_period_year_buddhist", "tax_period_year")
}

// ความยาวข้อความตาม journal_entries/journal_lines.sql — ตรวจเฉพาะข้อความใหม่/ที่แก้
func TestJournalTextLimitsCheckOnlyNewText(t *testing.T) {
	long := func(n int) string { return strings.Repeat("ก", n) }
	valid := func() Journal {
		return Journal{Description: long(500), Reference: long(50), Lines: []Line{{Description: long(255)}, {Description: "ซื้อปูนซีเมนต์"}}}
	}
	if err := checkJournalTextLimits(valid(), nil); err != nil {
		t.Fatalf("limits 500/50/255 Thai runes must pass: %v", err)
	}
	j := valid()
	j.Description = long(501)
	expectFieldError(t, "header 501", checkJournalTextLimits(j, nil), "journal_description_too_long", "description")
	j = valid()
	j.Reference = long(51)
	expectFieldError(t, "reference 51", checkJournalTextLimits(j, nil), "journal_reference_too_long", "reference")
	j = valid()
	j.Lines[1].Description = long(256)
	err := checkJournalTextLimits(j, nil)
	expectFieldError(t, "line 256", err, "journal_line_description_too_long", "lines[1].description")
	if !strings.Contains(err.Error(), "บรรทัดที่ 2") {
		t.Fatalf("line message must name the line: %q", err.Error())
	}

	old := Journal{Description: long(600), Reference: long(60), Lines: []Line{{Description: long(300)}, {Description: "ซื้อปูนซีเมนต์"}}}
	next := old
	next.Lines = append([]Line(nil), old.Lines...)
	next.Lines[1].Description = "ซื้อเหล็กเส้น"
	if err := checkJournalTextLimits(next, &old); err != nil {
		t.Fatalf("draft saved before the limits must stay editable: %v", err)
	}
	next.Description = long(601)
	expectFieldError(t, "edited long header", checkJournalTextLimits(next, &old), "journal_description_too_long", "description")

	if got := utf8.RuneCountInString(truncateRunes("กลับรายการ: "+long(255), JournalLineDescriptionMaxRunes)); got != JournalLineDescriptionMaxRunes {
		t.Fatalf("truncateRunes = %d runes, want %d", got, JournalLineDescriptionMaxRunes)
	}
	if got := truncateRunes("สั้น", 10); got != "สั้น" {
		t.Fatalf("short text changed: %q", got)
	}
}

// review 2026-09-24: วันที่ในรายละเอียดภาษีที่เป็นปี พ.ศ. ถูกรับไว้ แล้วแถวหายจาก ภ.ง.ด./ไฟล์ยื่น/ทะเบียนภาษีที่กรองปี ค.ศ.
func TestTaxDetailDatesRejectBuddhistYear(t *testing.T) {
	m := &subledgerMutation{scale: 2}
	sale := validSaleVat()
	sale.TaxInvoiceDate = "2569-09-10"
	expectFieldError(t, "sale invoice date B.E.", m.vat(&sale), "vat_invoice_date_buddhist", "tax_invoice_date")
	credit := validSaleVat()
	credit.DocumentType, credit.OriginalInvoiceNo, credit.OriginalInvoiceDate = 3, "IV2608-010", "2569-08-20"
	expectFieldError(t, "original invoice date B.E.", m.vat(&credit), "vat_original_invoice_date_buddhist", "original_invoice_date")

	w := SubledgerWithholding{ID: "W1", Direction: 1, FormType: "PND53", PartnerCode: "TRANS", PaymentDate: "2569-09-15", IncomeType: "3_tres", Condition: 1}
	expectFieldError(t, "payment date B.E.", checkWithholdingRowRules(&w, ""), "wht_payment_date_buddhist", "payment_date")
	w.PaymentDate, w.CertificateDate = "2026-09-15", "2569-09-30"
	expectFieldError(t, "certificate date B.E.", checkWithholdingRowRules(&w, ""), "wht_certificate_date_buddhist", "certificate_date")
	w.CertificateDate = "2026-09-30"
	if err := checkWithholdingRowRules(&w, ""); err != nil {
		t.Fatalf("C.E. dates rejected: %v", err)
	}
}

// review 2026-09-24: ภ.ง.ด.2 ใช้เลข 0 ทั้ง 13 หลักได้เฉพาะดอกเบี้ย 40(4)(ก) (Format กลาง ภ.ง.ด.2 ช่อง D4) ไม่ใช่ทุกประเภทเงินได้
func TestPND2ZeroPayeeOnlyForInterest(t *testing.T) {
	w := SubledgerWithholding{ID: "W1", Direction: 1, FormType: "PND2", PartnerCode: "DEP", PaymentDate: "2026-09-15", IncomeType: "40_4a", Condition: 1, PayeeTaxID: pnd2ZeroPIN}
	if err := checkWithholdingRowRules(&w, ""); err != nil {
		t.Fatalf("interest with zero PIN rejected: %v", err)
	}
	for _, income := range []string{"40_4b_1_1", "40_3", "3_tres", "other"} {
		w.IncomeType = income
		expectFieldError(t, "zero PIN "+income, checkWithholdingRowRules(&w, ""), "wht_payee_zero_tax_id_interest_only", "payee_tax_id")
	}
	w.FormType, w.IncomeType = "PND3", "40_4a"
	expectFieldError(t, "zero PIN outside PND2", checkWithholdingRowRules(&w, ""), "wht_payee_tax_id_checksum", "payee_tax_id")
}

// review 2026-09-24: เลขที่เติมจากทะเบียนคู่ค้าถูกตรวจด้วย และชี้ให้แก้ที่ทะเบียน (เดิมบอก "เว้นว่างได้" ซึ่งเติมเลขผิดเดิมกลับมา)
func TestRegistryTaxIDChecksumPointsToRegistry(t *testing.T) {
	const badRegistry = "0105558002001" // เลขใน glseed เดิม ไม่ผ่าน mod 11
	payee := SubledgerWithholding{ID: "W1", Direction: 1, FormType: "PND53", PaymentDate: "2026-09-15", IncomeType: "3_tres", Condition: 1}
	fillPartnerSnapshot(&payee, SubledgerPartner{Code: "SUPP-TH-001", Name: "บริษัท ปูนไทยค้าส่ง จำกัด", TaxID: badRegistry})
	expectFieldError(t, "payee from registry", checkWithholdingRowRules(&payee, badRegistry), "wht_partner_registry_tax_id_checksum", "payee_tax_id")
	payer := SubledgerWithholding{ID: "W2", Direction: 2, FormType: "PND53", PaymentDate: "2026-09-15", IncomeType: "3_tres", Condition: 1}
	fillPartnerSnapshot(&payer, SubledgerPartner{Code: "SUPP-TH-001", Name: "บริษัท ปูนไทยค้าส่ง จำกัด", TaxID: badRegistry})
	expectFieldError(t, "payer from registry", checkWithholdingRowRules(&payer, badRegistry), "wht_partner_registry_tax_id_checksum", "payer_tax_id")
	typed := SubledgerWithholding{ID: "W3", Direction: 1, FormType: "PND53", PaymentDate: "2026-09-15", IncomeType: "3_tres", Condition: 1, PayeeTaxID: "0105558012340"}
	expectFieldError(t, "typed bad number", checkWithholdingRowRules(&typed, badRegistry), "wht_payee_tax_id_checksum", "payee_tax_id")
}
