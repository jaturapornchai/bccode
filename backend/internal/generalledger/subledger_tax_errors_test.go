package generalledger

import (
	"encoding/json"
	"testing"
)

// UAT S21 2026-09-24: รายการภาษีหักที่ผิดต้องบอกว่าช่องไหนผิด (code + field) ไม่ใช่ข้อความรวมข้อความเดียว
func TestSubledgerWithholdingReportsFieldOfEachError(t *testing.T) {
	m := &subledgerMutation{scale: 2}
	valid := func() SubledgerWithholding {
		return SubledgerWithholding{ID: "W1", Direction: 1, FormType: "PND53", PartnerCode: "TRANS", PaymentDate: "2026-09-15",
			IncomeType: "3_tres", Condition: 1, Rate: "3", BaseAmount: "10000"}
	}
	negative := Amount("-1")
	cases := []struct {
		name, code, field string
		mutate            func(*SubledgerWithholding)
	}{
		{"no id", "wht_id_invalid", "id", func(w *SubledgerWithholding) { w.ID = "" }},
		{"direction 3", "wht_direction_invalid", "wht_direction", func(w *SubledgerWithholding) { w.Direction = 3 }},
		{"PND1 (ไม่ทำระบบเงินเดือน)", "wht_form_invalid", "form_type", func(w *SubledgerWithholding) { w.FormType = "PND1" }},
		{"no partner", "wht_partner_required", "partner_code", func(w *SubledgerWithholding) { w.PartnerCode = "" }},
		{"bad payment date", "wht_payment_date_invalid", "payment_date", func(w *SubledgerWithholding) { w.PaymentDate = "2026-02-30" }},
		{"bad certificate date", "wht_certificate_date_invalid", "certificate_date", func(w *SubledgerWithholding) { w.CertificateDate = "15/09/2569" }},
		{"empty income type", "wht_income_type_required", "income_tax_type", func(w *SubledgerWithholding) { w.IncomeType = "  " }},
		{"condition 0", "wht_condition_invalid", "condition_type", func(w *SubledgerWithholding) { w.Condition = 0 }},
		{"rate over 100", "wht_rate_invalid", "wht_rate", func(w *SubledgerWithholding) { w.Rate = "100.01" }},
		{"negative base", "wht_base_invalid", "base_amount", func(w *SubledgerWithholding) { w.BaseAmount = "-5" }},
		{"negative tax", "wht_tax_invalid", "tax_amount", func(w *SubledgerWithholding) { w.TaxAmount = &negative }},
	}
	for _, tc := range cases {
		w := valid()
		tc.mutate(&w)
		err := m.withholding(&w)
		user, ok := AsUserError(err)
		if !ok || user.Code != tc.code || user.Field != tc.field || user.Message == "" {
			t.Errorf("%s: err = %#v, want code %s field %s", tc.name, err, tc.code, tc.field)
		}
		if ok && user.ToAppError().Field != tc.field {
			t.Errorf("%s: field lost in API error: %+v", tc.name, user.ToAppError())
		}
	}
}

// UAT V13 2026-09-24: คำสั่งที่ไม่ส่ง vat_rate ต้องถูกปฏิเสธ — การคัดลอกผ่าน JSON แปลงค่าว่างเป็น "0" จนกลายเป็น VAT 0% เงียบ ๆ
func TestCloneJournalDetailsRejectsMissingVatRate(t *testing.T) {
	var details JournalDetails
	if err := json.Unmarshal([]byte(`{"vats":[{"id":"V1","tax_type":2,"document_type":1,"tax_invoice_no":"IV6910-001","tax_invoice_date":"2026-10-05",
		"tax_period_year":2026,"tax_period_month":10,"partner_name":"บริษัท ก่อสร้างมั่นคง จำกัด","base_amount":"1000.00"}]}`), &details); err != nil {
		t.Fatal(err)
	}
	_, err := cloneJournalDetails(&details)
	user, ok := AsUserError(err)
	if !ok || user.Code != "vat_rate_invalid" || user.Field != "vat_rate" || user.Message != "อัตราภาษีมูลค่าเพิ่มต้องอยู่ระหว่าง 0–100 และมีทศนิยมไม่เกิน 2 ตำแหน่ง" {
		t.Fatalf("missing vat_rate: err = %#v", err)
	}
	// อัตรา 0 ที่ผู้ใช้ส่งมาจริง (ส่งออก) ยังผ่าน และคงเป็น "0"
	details.Vats[0].Rate = "0"
	cloned, err := cloneJournalDetails(&details)
	if err != nil || cloned.Vats[0].Rate != "0" {
		t.Fatalf("explicit 0%% rate: cloned=%+v err=%v", cloned, err)
	}
}

// ข้อความตรวจ VAT ต้องมี code (แปลภาษาอื่นผ่าน gl_err_<code>) — เดิมเป็น fmt.Errorf ไทยล้วน ภาษาอังกฤษก็ได้ไทย (UAT V13)
func TestSubledgerVatErrorsCarryCodes(t *testing.T) {
	m := &subledgerMutation{scale: 2}
	v := validSaleVat()
	v.TaxPeriodYear, v.TaxPeriodMonth = 0, 0
	user, ok := AsUserError(m.vat(&v))
	if !ok || user.Code != "vat_period_required" || user.Field != "tax_period_month" {
		t.Fatalf("sale without period = %#v", user)
	}
	v = validSaleVat()
	v.PartnerTaxID = "12345"
	if user, ok = AsUserError(m.vat(&v)); !ok || user.Code != "vat_partner_tax_id_invalid" {
		t.Fatalf("bad tax id = %#v", user)
	}
}

// รายการ VAT ที่ผิดต้องบอกช่องที่ผิดแยกตามสาเหตุ — เดิมหกสาเหตุรวมเป็น vat_invalid ไม่มี field จอชี้ช่องไม่ได้
func TestSubledgerVatReportsFieldOfEachError(t *testing.T) {
	m := &subledgerMutation{scale: 2}
	long := func(n int) string {
		b := make([]rune, n)
		for i := range b {
			b[i] = 'ก'
		}
		return string(b)
	}
	cases := []struct {
		name, code, field string
		mutate            func(*SubledgerVat)
	}{
		{"no id", "vat_id_invalid", "id", func(v *SubledgerVat) { v.ID = "" }},
		{"tax type 3", "vat_tax_type_invalid", "tax_type", func(v *SubledgerVat) { v.TaxType = 3 }},
		{"document type 4", "vat_document_type_invalid", "document_type", func(v *SubledgerVat) { v.DocumentType = 4 }},
		{"blank invoice no", "vat_invoice_no_invalid", "tax_invoice_no", func(v *SubledgerVat) { v.TaxInvoiceNo = "  " }},
		{"invoice no 51 chars", "vat_invoice_no_invalid", "tax_invoice_no", func(v *SubledgerVat) { v.TaxInvoiceNo = long(51) }},
		{"invoice date 30 Feb", "vat_invoice_date_invalid", "tax_invoice_date", func(v *SubledgerVat) { v.TaxInvoiceDate = "2026-02-30" }},
		{"original no 51 chars", "vat_original_invoice_no_too_long", "original_invoice_no", func(v *SubledgerVat) { v.OriginalInvoiceNo = long(51) }},
		{"original date Thai format", "vat_original_invoice_date_invalid", "original_invoice_date", func(v *SubledgerVat) { v.OriginalInvoiceDate = "10/09/2569" }},
		{"blank partner name", "vat_partner_name_invalid", "partner_name", func(v *SubledgerVat) { v.PartnerName = "" }},
		{"claim reason 501", "vat_claim_reason_too_long", "claim_reason", func(v *SubledgerVat) { v.ClaimReason = long(501) }},
		{"remark 501", "vat_remark_too_long", "remark", func(v *SubledgerVat) { v.Remark = long(501) }},
	}
	for _, tc := range cases {
		v := validSaleVat()
		tc.mutate(&v)
		user, ok := AsUserError(m.vat(&v))
		if !ok || user.Code != tc.code || user.Field != tc.field || user.Message == "" {
			t.Errorf("%s: err = %#v, want code %s field %s", tc.name, user, tc.code, tc.field)
		}
	}
}

// ภาษีซื้อที่ใช้สิทธิ (claim_status 1): งวดต้องอยู่ระหว่างเดือนที่ออกใบกำกับ ถึง 6 เดือนถัดไป
// (ม.82/3 วรรคสอง + ประกาศอธิบดีฯ VAT ฉบับที่ 4 ข้อ 2 แก้โดยฉบับที่ 76) — ใบกำกับ ม.ค. 2569 ใช้สิทธิได้ ม.ค.–ก.ค. 2026 (ค.ศ.)
func TestPurchaseVatClaimWindow(t *testing.T) {
	m := &subledgerMutation{scale: 2}
	purchase := func(invoiceDate string, year, month, claim int) SubledgerVat {
		v := validSaleVat()
		v.TaxType, v.ClaimStatus, v.TaxInvoiceNo, v.TaxInvoiceDate = 1, claim, "IV6901-0042", invoiceDate
		v.TaxPeriodYear, v.TaxPeriodMonth = year, month
		return v
	}
	cases := []struct {
		name        string
		v           SubledgerVat
		code, field string // code ว่าง = ต้องผ่าน
	}{
		{"same month", purchase("2026-01-31", 2026, 1, 1), "", ""},
		{"deferred 6 months (Jul)", purchase("2026-01-05", 2026, 7, 1), "", ""},
		{"cross year: Sep invoice → Mar next year", purchase("2025-09-30", 2026, 3, 1), "", ""},
		{"period before invoice month", purchase("2026-02-01", 2026, 1, 1), "vat_claim_before_invoice_month", "tax_period_month"},
		{"7th month (Aug) exceeded", purchase("2026-01-05", 2026, 8, 1), "vat_claim_window_exceeded", "tax_period_month"},
		{"Buddhist year typed as period", purchase("2026-01-05", 2569, 1, 1), "vat_period_year_buddhist", "tax_period_year"},
		{"forbidden input VAT is not checked", purchase("2026-01-05", 2026, 12, 2), "", ""},
		{"pending claim without period is not checked", purchase("2026-01-05", 0, 0, 3), "", ""},
		{"not claimed is not checked", purchase("2026-01-05", 2027, 1, 4), "", ""},
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
		user, ok := AsUserError(err)
		if !ok || user.Code != tc.code || user.Field != tc.field || user.Message == "" {
			t.Errorf("%s: err = %#v, want %s on %s", tc.name, err, tc.code, tc.field)
		}
	}
	// ภาษีขายไม่อยู่ใต้กติกานี้ (ยังไม่มีมติ) — งวดห่างจากใบกำกับเกิน 6 เดือนยังบันทึกได้เหมือนเดิม
	sale := validSaleVat()
	sale.TaxPeriodYear, sale.TaxPeriodMonth = 2027, 6
	if err := m.vat(&sale); err != nil {
		t.Fatalf("sale VAT must not be checked by the input-tax claim window: %v", err)
	}
}
