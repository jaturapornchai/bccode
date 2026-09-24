package generalledger

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

// ชื่อคู่ค้าไทย 103 ตัวอักษร (≈ 300 ไบต์) — เดิมนับไบต์ จึงถูกปฏิเสธทั้งที่ไม่เกิน 255 ตัวอักษร (save-audit 2026-09-24)
const thaiPartnerName = "ห้างหุ้นส่วนจำกัด รุ่งเรืองการค้าไทย วัสดุก่อสร้างและเหล็กเส้นครบวงจร สาขาลาดหลุมแก้ว จังหวัดปทุมธานี"

func expectFieldError(t *testing.T, name string, err error, code, field string) {
	t.Helper()
	user, ok := AsUserError(err)
	if !ok || user.Code != code || user.Field != field || user.Message == "" {
		t.Errorf("%s: err = %#v, want code %s field %s", name, err, code, field)
	}
}

func TestNormalizeTaxIDAndBranch(t *testing.T) {
	for in, want := range map[string]string{
		"0-1055-58012-34-9":   "0105558012349",
		" 0 1055 58012 34 9 ": "0105558012349",
		"0–1055–58012–34–9":   "0105558012349", // en dash จากเอกสาร Word
		"0105558012349":       "0105558012349",
		"":                    "",
		"0-1055-58012-34-X":   "010555801234X", // ตัวอักษรไม่ตัด → ถูกปฏิเสธที่ regex
	} {
		if got := normalizeTaxID(in); got != want {
			t.Errorf("normalizeTaxID(%q) = %q, want %q", in, got, want)
		}
	}
	for in, want := range map[string]string{
		"0":      "00000",
		"1":      "00001",
		" 12 ":   "00012",
		"00000":  "00000",
		"":       "", // ว่าง = ไม่ระบุ ห้ามเติมเป็นสำนักงานใหญ่เอง
		"HQ":     "HQ",
		"000001": "000001",
	} {
		if got := normalizeTaxBranch(in); got != want {
			t.Errorf("normalizeTaxBranch(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidatePartnerReportsEachField(t *testing.T) {
	if utf8.RuneCountInString(thaiPartnerName) > 255 || len(thaiPartnerName) <= 255 {
		t.Fatalf("fixture must be <=255 runes but >255 bytes: %d runes %d bytes", utf8.RuneCountInString(thaiPartnerName), len(thaiPartnerName))
	}
	valid := func() SubledgerPartner {
		return SubledgerPartner{Code: "CUST-TH-001", Name: thaiPartnerName, TaxID: "0-1055-58012-34-9", TaxBranch: "0", IsCustomer: true, IsActive: true}
	}
	p := valid()
	normalizePartner(&p)
	if err := validatePartner(&p); err != nil {
		t.Fatalf("Thai name + dashed tax id + branch 0 must pass: %v", err)
	}
	if p.TaxID != "0105558012349" || p.TaxBranch != "00000" {
		t.Fatalf("normalized tax id/branch = %q/%q", p.TaxID, p.TaxBranch)
	}
	cases := []struct {
		name, code, field string
		mutate            func(*SubledgerPartner)
	}{
		{"code with space", "code_has_space", "partner_code", func(p *SubledgerPartner) { p.Code = "bad code" }},
		{"code with bang", "code_invalid_char", "partner_code", func(p *SubledgerPartner) { p.Code = "bad!" }},
		{"title 101 runes", "partner_title_name_too_long", "title_name", func(p *SubledgerPartner) { p.TitleName = strings.Repeat("ก", 101) }},
		{"district 51 runes", "partner_addr_district_too_long", "addr_district", func(p *SubledgerPartner) { p.AddrDistrict = strings.Repeat("ก", 51) }},
		{"province 51 runes", "partner_addr_province_too_long", "addr_province", func(p *SubledgerPartner) { p.AddrProvince = strings.Repeat("ก", 51) }},
		{"postcode 4 digits", "partner_addr_postcode_invalid", "addr_postcode", func(p *SubledgerPartner) { p.AddrPostcode = "1011" }},
		{"postcode 6 digits", "partner_addr_postcode_invalid", "addr_postcode", func(p *SubledgerPartner) { p.AddrPostcode = "101100" }},
		{"postcode letters", "partner_addr_postcode_invalid", "addr_postcode", func(p *SubledgerPartner) { p.AddrPostcode = "1O11O" }},
		{"blank name", "partner_name_required", "name_th", func(p *SubledgerPartner) { p.Name = "   " }},
		{"name 256 runes", "partner_name_too_long", "name_th", func(p *SubledgerPartner) { p.Name = strings.Repeat("ก", 256) }},
		{"no role", "partner_role_required", "is_customer", func(p *SubledgerPartner) { p.IsCustomer, p.IsSupplier = false, false }},
		{"tax id 12 digits", "partner_tax_id_invalid", "tax_id", func(p *SubledgerPartner) { p.TaxID = "0-1055-58012-34" }},
		{"tax id letter", "partner_tax_id_invalid", "tax_id", func(p *SubledgerPartner) { p.TaxID = "0-1055-58012-34-X" }},
		{"branch 6 digits", "partner_branch_invalid", "tax_branch_no", func(p *SubledgerPartner) { p.TaxBranch = "000001" }},
		{"branch text", "partner_branch_invalid", "tax_branch_no", func(p *SubledgerPartner) { p.TaxBranch = "สำนักงานใหญ่" }},
	}
	for _, tc := range cases {
		p := valid()
		tc.mutate(&p)
		normalizePartner(&p)
		expectFieldError(t, tc.name, validatePartner(&p), tc.code, tc.field)
	}
	p = valid()
	p.Name = strings.Repeat("ก", 255)
	normalizePartner(&p)
	if err := validatePartner(&p); err != nil {
		t.Fatalf("255 Thai runes must pass: %v", err)
	}
}

// save-audit 2026-09-24: รายการภาษีหักที่ไม่ส่ง wht_rate เคยถูกบันทึกเป็น 0% ภาษี 0 เงียบ ๆ
func TestCloneJournalDetailsRejectsMissingWithholdingRate(t *testing.T) {
	var details JournalDetails
	if err := json.Unmarshal([]byte(`{"withholdings":[{"id":"W1","wht_direction":1,"form_type":"PND53","partner_code":"C001",
		"payment_date":"2026-09-01","income_tax_type":"3_tres","condition_type":1,"base_amount":"10000"}]}`), &details); err != nil {
		t.Fatal(err)
	}
	_, err := cloneJournalDetails(&details)
	expectFieldError(t, "missing wht_rate", err, "wht_rate_required", "wht_rate")
	// ส่งอัตรา 0 มาจริง (ไม่หัก) ยังผ่าน และคงเป็น "0"
	details.Withholdings[0].Rate = "0"
	cloned, err := cloneJournalDetails(&details)
	if err != nil || cloned.Withholdings[0].Rate != "0" {
		t.Fatalf("explicit 0%% rate: %v %v", err, cloned)
	}
	// ตัวตรวจรายการเองก็ปฏิเสธอัตราว่าง (ผู้เรียกที่ไม่ผ่าน clone)
	w := SubledgerWithholding{ID: "W1", Direction: 1, FormType: "PND53", PartnerCode: "C001", PaymentDate: "2026-09-01", IncomeType: "3_tres", Condition: 1, BaseAmount: "10000"}
	expectFieldError(t, "withholding() blank rate", (&subledgerMutation{scale: 2}).withholding(&w), "wht_rate_required", "wht_rate")
}

func TestWithholdingSnapshotNormalizedAndValidated(t *testing.T) {
	w := SubledgerWithholding{PayerTaxID: " 0-1055-58012-34-9 ", PayerBranchNo: "0", PayeeTaxID: "3 1012 00345 67 8", PayeeBranchNo: "",
		PayerName: "  บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด ", BookNo: " 12 ", Remark: strings.Repeat("ข", 500)}
	if err := normalizeWithholdingSnapshot(&w); err != nil {
		t.Fatalf("valid snapshot rejected: %v", err)
	}
	if w.PayerTaxID != "0105558012349" || w.PayerBranchNo != "00000" || w.PayeeTaxID != "3101200345678" || w.PayeeBranchNo != "" ||
		w.PayerName != "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด" || w.BookNo != "12" {
		t.Fatalf("normalized snapshot = %+v", w)
	}
	cases := []struct {
		name, code, field string
		mutate            func(*SubledgerWithholding)
	}{
		{"payer tax id", "wht_payer_tax_id_invalid", "payer_tax_id", func(w *SubledgerWithholding) { w.PayerTaxID = "123" }},
		{"payee tax id", "wht_payee_tax_id_invalid", "payee_tax_id", func(w *SubledgerWithholding) { w.PayeeTaxID = "01055580123490" }},
		{"payer branch", "wht_payer_branch_invalid", "payer_branch_no", func(w *SubledgerWithholding) { w.PayerBranchNo = "สาขา1" }},
		{"payee branch", "wht_payee_branch_invalid", "payee_branch_no", func(w *SubledgerWithholding) { w.PayeeBranchNo = "123456" }},
		{"payer name", "wht_payer_name_too_long", "payer_name", func(w *SubledgerWithholding) { w.PayerName = strings.Repeat("ก", 256) }},
		{"payee name", "wht_payee_name_too_long", "payee_name", func(w *SubledgerWithholding) { w.PayeeName = strings.Repeat("ก", 256) }},
		{"book no", "wht_book_no_too_long", "wht_book_no", func(w *SubledgerWithholding) { w.BookNo = strings.Repeat("๑", 51) }},
		{"remark 501 runes", "wht_remark_too_long", "remark", func(w *SubledgerWithholding) { w.Remark = strings.Repeat("ข", 501) }},
	}
	for _, tc := range cases {
		w := SubledgerWithholding{}
		tc.mutate(&w)
		expectFieldError(t, tc.name, normalizeWithholdingSnapshot(&w), tc.code, tc.field)
	}
}

func TestFillPartnerSnapshotBySide(t *testing.T) {
	partner := SubledgerPartner{Code: "TRANS", Name: " บริษัท ขนส่งไทยเร็ว จำกัด ", TaxID: "0105558012349", TaxBranch: "00001", Address: "88/12 ถนนลาดหลุมแก้ว ปทุมธานี 12140"}
	// ทิศทาง 1 เราหักภาษีผู้รับเงิน: คู่ค้า = ผู้รับเงิน (payee); ฝั่งผู้จ่าย (บริษัท) ไม่เติม
	w := SubledgerWithholding{Direction: 1}
	fillPartnerSnapshot(&w, partner)
	if w.PayeeName != "บริษัท ขนส่งไทยเร็ว จำกัด" || w.PayeeTaxID != "0105558012349" || w.PayeeBranchNo != "00001" || w.PayeeAddress == "" || !w.Payer().blank() {
		t.Fatalf("direction 1 snapshot = %+v", w)
	}
	// ทิศทาง 2 ผู้จ่ายหักภาษีเรา: คู่ค้า = ผู้จ่าย (payer)
	w = SubledgerWithholding{Direction: 2}
	fillPartnerSnapshot(&w, partner)
	if w.PayerTaxID != "0105558012349" || !w.Payee().blank() {
		t.Fatalf("direction 2 snapshot = %+v", w)
	}
	// ผู้ใช้ระบุบางช่องเอง (เว้นสาขาว่างโดยตั้งใจ) → ไม่เติมช่องที่เหลือ
	w = SubledgerWithholding{Direction: 1, PayeeName: "ชื่อตามหนังสือรับรอง"}
	fillPartnerSnapshot(&w, partner)
	if w.PayeeName != "ชื่อตามหนังสือรับรอง" || w.PayeeTaxID != "" || w.PayeeBranchNo != "" {
		t.Fatalf("partial snapshot overwritten: %+v", w)
	}
}

// ความยาวนับตัวอักษร: ข้อความไทยยาวเกินจำนวนไบต์เดิมแต่ไม่เกินจำนวนตัวอักษรต้องผ่านการตรวจความยาว
func TestStatementAndWithdrawalLengthsCountRunes(t *testing.T) {
	m := &subledgerMutation{scale: 2}
	valid := func() SubledgerStatementLine {
		return SubledgerStatementLine{ID: "S1", BankAccountCode: "BANK", SourceKey: "KBANK-2026-01-11-0001", Date: "2026-01-11", Direction: 1, Amount: "0"}
	}
	// ยอด 0 ทำให้หยุดที่ช่องยอดเงิน (ก่อนแตะฐานข้อมูล) — ถ้าหยุดที่ช่องยอด แปลว่าผ่านการตรวจความยาวแล้ว
	s := valid()
	s.Description, s.Reference, s.SourceKey = strings.Repeat("โอนเงิน", 70), strings.Repeat("อ", 150), strings.Repeat("ก", 150)
	expectFieldError(t, "Thai 490-rune description", m.statement(&s), "statement_amount_invalid", "amount")
	cases := []struct {
		name, code, field string
		mutate            func(*SubledgerStatementLine)
	}{
		{"no id", "statement_id_invalid", "id", func(s *SubledgerStatementLine) { s.ID = "" }},
		{"no source key", "statement_source_key_required", "source_key", func(s *SubledgerStatementLine) { s.SourceKey = " " }},
		{"source key 151", "statement_source_key_too_long", "source_key", func(s *SubledgerStatementLine) { s.SourceKey = strings.Repeat("ก", 151) }},
		{"bad date", "statement_date_invalid", "transaction_date", func(s *SubledgerStatementLine) { s.Date = "11/01/2569" }},
		{"bad value date", "statement_value_date_invalid", "value_date", func(s *SubledgerStatementLine) { s.ValueDate = "2026-13-01" }},
		{"direction 3", "statement_direction_invalid", "direction", func(s *SubledgerStatementLine) { s.Direction = 3 }},
		{"description 501", "statement_description_too_long", "description", func(s *SubledgerStatementLine) { s.Description = strings.Repeat("ก", 501) }},
		{"reference 151", "statement_reference_too_long", "bank_reference", func(s *SubledgerStatementLine) { s.Reference = strings.Repeat("ก", 151) }},
	}
	for _, tc := range cases {
		s := valid()
		s.Amount = "10"
		tc.mutate(&s)
		expectFieldError(t, tc.name, m.statement(&s), tc.code, tc.field)
	}

	// เหตุผลการถอนไทย 500 ตัว (1,500 ไบต์) ผ่านการตรวจความยาว — ประเภทผิดทำให้หยุดก่อนแตะฐานข้อมูล
	err := m.withdraw(SubledgerWithdrawal{Kind: "other", ID: "SET1", Reason: strings.Repeat("ก", 500)})
	if err == nil || err.Error() != "ประเภทการถอนไม่ถูกต้อง" {
		t.Fatalf("500 Thai runes reason must pass length check, got %v", err)
	}
	expectFieldError(t, "reason 501", m.withdraw(SubledgerWithdrawal{Kind: "settlement", ID: "SET1", Reason: strings.Repeat("ก", 501)}), "withdraw_reason_too_long", "reason")
	expectFieldError(t, "blank reason", m.withdraw(SubledgerWithdrawal{Kind: "settlement", ID: "SET1", Reason: "  "}), "withdraw_reason_required", "reason")
	expectFieldError(t, "no id", m.withdraw(SubledgerWithdrawal{Kind: "settlement", Reason: "ถอน"}), "withdraw_id_invalid", "id")
}

// ชุดว่างกับไม่ส่งคีย์ต้องแยกกันได้หลัง decode (reconcile อ่านก่อน clone)
func TestExplicitEmptyTaxArraysSurviveDecode(t *testing.T) {
	var cleared, absent JournalDetails
	if err := json.Unmarshal([]byte(`{"withholdings":[],"vats":[]}`), &cleared); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{}`), &absent); err != nil {
		t.Fatal(err)
	}
	if cleared.Withholdings == nil || cleared.Vats == nil || absent.Withholdings != nil || absent.Vats != nil {
		t.Fatalf("decode lost [] vs absent: cleared=%#v absent=%#v", cleared, absent)
	}
}

// UAT 2026-09-24: snapshot ภาษีหักเก็บชื่อไม่มีคำนำหน้า และที่อยู่ไม่มีอำเภอ/จังหวัด/รหัสไปรษณีย์ → หนังสือรับรอง 50 ทวิ ได้ค่าไม่ครบ
// (แบบ 50 ทวิ approve_wh3_081156.pdf: ที่อยู่ "ให้ระบุ ... ตำบล/แขวง อำเภอ/เขต จังหวัด")
func TestFillPartnerSnapshotKeepsFullNameAndAddress(t *testing.T) {
	partner := SubledgerPartner{Code: "ZUAT-IND", TitleName: "นาย", Name: "สมชาย รับเหมาดี", TaxID: "3149900771234", TaxBranch: "00000",
		Address: "เลขที่ 12 หมู่ 3 ตำบลคูบางหลวง", AddrDistrict: "อำเภอลาดหลุมแก้ว", AddrProvince: "ปทุมธานี", AddrPostcode: "12140"}
	w := SubledgerWithholding{Direction: 1}
	fillPartnerSnapshot(&w, partner)
	if w.PayeeName != "นาย สมชาย รับเหมาดี" || w.PayeeAddress != "เลขที่ 12 หมู่ 3 ตำบลคูบางหลวง อำเภอลาดหลุมแก้ว ปทุมธานี 12140" || w.PayeeBranchNo != "00000" {
		t.Fatalf("payee snapshot = %q / %q / %q", w.PayeeName, w.PayeeAddress, w.PayeeBranchNo)
	}
	// ชื่อที่ขึ้นต้นด้วยคำนำหน้าอยู่แล้ว / ที่อยู่ที่มีจังหวัดอยู่แล้ว ไม่เติมซ้ำ
	if got := PartnerFullName("บริษัท", "บริษัท ปูนไทยค้าส่ง จำกัด"); got != "บริษัท ปูนไทยค้าส่ง จำกัด" {
		t.Fatalf("PartnerFullName duplicated the title: %q", got)
	}
	if got := PartnerFullAddress("88/12 ถนนลาดหลุมแก้ว ปทุมธานี 12140", "", "ปทุมธานี", "12140"); got != "88/12 ถนนลาดหลุมแก้ว ปทุมธานี 12140" {
		t.Fatalf("PartnerFullAddress duplicated parts: %q", got)
	}
	if got := PartnerFullName("-", "สมชาย ใจดี"); got != "สมชาย ใจดี" {
		t.Fatalf("placeholder title must not be prefixed: %q", got)
	}
}
