package generalledger

import (
	"encoding/json"
	"strings"
	"testing"
)

// ใบกำกับภาษีขายที่ถูกต้องตาม vat.sql — ใช้เป็นฐานของกรณีทดสอบ (ไม่มีรหัสคู่ค้า จึงไม่ต้องใช้ฐานข้อมูล)
func validSaleVat() SubledgerVat {
	return SubledgerVat{ID: "V1", TaxType: 2, DocumentType: 1, TaxInvoiceNo: "IV2609-001", TaxInvoiceDate: "2026-09-10",
		TaxPeriodYear: 2026, TaxPeriodMonth: 9, PartnerTaxID: "0105558012349", PartnerBranchNo: "00000",
		PartnerName: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", BaseAmount: "100000", Rate: "7"}
}

func TestSubledgerVatComputesAndKeepsTypedAmount(t *testing.T) {
	m := &subledgerMutation{scale: 2}
	v := validSaleVat()
	if err := m.vat(&v); err != nil {
		t.Fatal(err)
	}
	if v.VatAmount == nil || *v.VatAmount != "7000" {
		t.Fatalf("auto vat = %v", v.VatAmount)
	}
	// ฐาน 1,234.57 × 7% = 86.4199 → ปัดตามทศนิยมปีบัญชีเป็น 86.42
	v = validSaleVat()
	v.BaseAmount = "1234.57"
	if err := m.vat(&v); err != nil || *v.VatAmount != "86.42" {
		t.Fatalf("rounded vat = %v err=%v", v.VatAmount, err)
	}
	// ยอดภาษีที่ผู้ใช้พิมพ์ตามใบกำกับจริงต้องคงไว้ (ไม่มีกฎ vat ≤ ฐาน ใน vat.sql)
	typed := Amount("86.41")
	v = validSaleVat()
	v.BaseAmount, v.VatAmount = "1234.57", &typed
	if err := m.vat(&v); err != nil || *v.VatAmount != "86.41" {
		t.Fatalf("typed vat = %v err=%v", v.VatAmount, err)
	}
	// ไม่ระบุประเภทเอกสาร = ใบกำกับภาษี (DEFAULT 1)
	v = validSaleVat()
	v.DocumentType = 0
	if err := m.vat(&v); err != nil || v.DocumentType != 1 {
		t.Fatalf("default document type = %d err=%v", v.DocumentType, err)
	}
	// ภาษีซื้อรอใช้สิทธิไม่ต้องมีงวด
	v = validSaleVat()
	v.TaxType, v.ClaimStatus, v.TaxPeriodYear, v.TaxPeriodMonth, v.ClaimReason = 1, 3, 0, 0, "ใบกำกับมาถึงช้า รอใช้สิทธิเดือนถัดไป"
	if err := m.vat(&v); err != nil {
		t.Fatalf("pending purchase: %v", err)
	}
}

func TestSubledgerVatRejectsRowsOutsideVatSQLChecks(t *testing.T) {
	m := &subledgerMutation{scale: 2}
	negative, over := Amount("-1"), Amount("1.001")
	cases := map[string]func(*SubledgerVat){
		"no id":                      func(v *SubledgerVat) { v.ID = "" },
		"bad tax type":               func(v *SubledgerVat) { v.TaxType = 3 },
		"bad document type":          func(v *SubledgerVat) { v.DocumentType = 4 },
		"blank invoice no":           func(v *SubledgerVat) { v.TaxInvoiceNo = "   " },
		"invoice no over 50":         func(v *SubledgerVat) { v.TaxInvoiceNo = strings.Repeat("9", 51) },
		"bad invoice date":           func(v *SubledgerVat) { v.TaxInvoiceDate = "2026-02-30" },
		"blank partner name":         func(v *SubledgerVat) { v.PartnerName = " " },
		"tax id 12 digits":           func(v *SubledgerVat) { v.PartnerTaxID = "010555801234" },
		"branch 4 digits":            func(v *SubledgerVat) { v.PartnerBranchNo = "0000" },
		"credit note without origin": func(v *SubledgerVat) { v.DocumentType = 3 },
		"debit note without date":    func(v *SubledgerVat) { v.DocumentType, v.OriginalInvoiceNo = 2, "IV2608-010" },
		"period month only":          func(v *SubledgerVat) { v.TaxPeriodYear = 0 },
		"period month 13":            func(v *SubledgerVat) { v.TaxPeriodMonth = 13 },
		"sale without period":        func(v *SubledgerVat) { v.TaxPeriodYear, v.TaxPeriodMonth = 0, 0 },
		"sale with claim status":     func(v *SubledgerVat) { v.ClaimStatus = 1 },
		"purchase without status":    func(v *SubledgerVat) { v.TaxType = 1 },
		"claimed without period":     func(v *SubledgerVat) { v.TaxType, v.ClaimStatus, v.TaxPeriodYear, v.TaxPeriodMonth = 1, 1, 0, 0 },
		"no rate":                    func(v *SubledgerVat) { v.Rate = "" },
		"rate over 100":              func(v *SubledgerVat) { v.Rate = "100.01" },
		"rate 3 decimals":            func(v *SubledgerVat) { v.Rate = "7.125" },
		"negative base":              func(v *SubledgerVat) { v.BaseAmount = "-1" },
		"negative zero rate":         func(v *SubledgerVat) { v.ZeroRateAmount = "-0.01" },
		"exempt over scale":          func(v *SubledgerVat) { v.ExemptAmount = "1.001" },
		"base over numeric(16,2)":    func(v *SubledgerVat) { v.BaseAmount = "100000000000000" },
		"negative vat":               func(v *SubledgerVat) { v.VatAmount = &negative },
		"vat over scale":             func(v *SubledgerVat) { v.VatAmount = &over },
		"bad partner code":           func(v *SubledgerVat) { v.PartnerCode = "bad code!" },
		"remark over 500":            func(v *SubledgerVat) { v.Remark = strings.Repeat("ก", 501) },
	}
	for name, mutate := range cases {
		v := validSaleVat()
		mutate(&v)
		if err := m.vat(&v); err == nil {
			t.Errorf("%s: expected rejection", name)
		}
	}
	// ใบลดหนี้ที่อ้างใบกำกับเดิมครบ ผ่าน
	v := validSaleVat()
	v.DocumentType, v.OriginalInvoiceNo, v.OriginalInvoiceDate = 3, "IV2608-010", "2026-08-20"
	if err := m.vat(&v); err != nil {
		t.Fatalf("credit note with origin: %v", err)
	}
}

// ชื่อฟิลด์ JSON เป็นสัญญากับ frontend และตัวสร้างแบบยื่น — ห้ามเปลี่ยน
func TestSubledgerVatJSONContract(t *testing.T) {
	m := &subledgerMutation{scale: 2}
	v := validSaleVat()
	if err := m.vat(&v); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(JournalDetails{Vats: []SubledgerVat{v}})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"vats":[`, `"id":"V1"`, `"tax_type":2`, `"document_type":1`, `"tax_invoice_no":"IV2609-001"`, `"tax_invoice_date":"2026-09-10"`,
		`"tax_period_year":2026`, `"tax_period_month":9`, `"partner_tax_id":"0105558012349"`, `"partner_branch_no":"00000"`, `"partner_name":`,
		`"base_amount":"100000"`, `"zero_rate_amount":"0"`, `"exempt_amount":"0"`, `"vat_rate":"7"`, `"vat_amount":"7000"`} {
		if !strings.Contains(string(raw), key) {
			t.Fatalf("missing %s in %s", key, raw)
		}
	}
	if strings.Contains(string(raw), `"claim_status"`) {
		t.Fatalf("sales must not carry claim_status: %s", raw)
	}
	var back JournalDetails
	if err = json.Unmarshal(raw, &back); err != nil || len(back.Vats) != 1 || *back.Vats[0].VatAmount != "7000" || back.Vats[0].TaxPeriodMonth != 9 {
		t.Fatalf("round trip = %+v err=%v", back, err)
	}
	// ยอดเงินเป็นตัวเลข JSON ต้องถูกปฏิเสธ (ห้าม float)
	if err = json.Unmarshal([]byte(`{"vats":[{"id":"V1","base_amount":100.5}]}`), &back); err == nil {
		t.Fatal("JSON number amount must be rejected")
	}
}
