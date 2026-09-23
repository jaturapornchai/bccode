package whtcert

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// withCheckDigit - เติมหลักตรวจสอบ (mod 11) ให้เลข 12 หลักแรก
func withCheckDigit(first12 string) string {
	sum := 0
	for i := 0; i < 12; i++ {
		sum += int(first12[i]-'0') * (13 - i)
	}
	return first12 + fmt.Sprint((11-sum%11)%10)
}

var (
	payerCompany = Party{
		Name:    "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด (สำนักงานใหญ่)",
		Address: "เลขที่ 88/12 หมู่ที่ 3 ถนนลาดหลุมแก้ว ตำบลคูบางหลวง อำเภอลาดหลุมแก้ว จังหวัดปทุมธานี 12140",
		TaxID:   withCheckDigit("010555801234"),
	}
	payeeTransport = Party{
		Name:    "บริษัท สยามขนส่งด่วน จำกัด",
		Address: "เลขที่ 159 ซอยบางนา-ตราด 25 แขวงบางนาใต้ เขตบางนา กรุงเทพมหานคร 10260",
		TaxID:   withCheckDigit("010556204567"),
	}
	payeePerson = Party{
		Name:    "นายสมศักดิ์ ใจดีงาม",
		Address: "เลขที่ 45/7 หมู่บ้านพฤกษา 12 ซอย 5 ถนนรังสิต-นครนายก ตำบลประชาธิปัตย์ อำเภอธัญบุรี จังหวัดปทุมธานี 12130",
		TaxID:   withCheckDigit("310120045678"),
	}
)

// sampleCertificates - ตัวอย่างหลายแบบ ครอบคลุมทุกช่องเลือกบนแบบฟอร์มรวมกัน
func sampleCertificates() map[string]Certificate {
	return map[string]Certificate{
		"01-pnd53-service-withhold": {
			BookNo: "1", RunNo: "2569/0001", Payer: payerCompany, Payee: payeeTransport, SequenceNo: "1",
			Form: FormPND53, Condition: ConditionWithhold, IssueDate: "2026-01-15",
			Incomes: []Income{{Type: Income3Tres, PaidDate: "2026-01-15", Amount: "25000.00", Tax: "750.00"}},
		},
		"02-pnd53-rent-grossup-always": {
			BookNo: "1", RunNo: "2569/0002",
			Payer: payerCompany,
			Payee: Party{Name: "บริษัท สยามพาณิชย์ โฮลดิ้ง จำกัด", Address: "เลขที่ 1 อาคารสยามพาณิชย์ ชั้น 12 ถนนสุขุมวิท แขวงคลองเตย เขตคลองเตย กรุงเทพมหานคร 10110",
				TaxID: withCheckDigit("010553700891")},
			SequenceNo: "2", Form: FormPND53, Condition: ConditionAlways, IssueDate: "2026-01-31",
			Incomes: []Income{{Type: Income3Tres, PaidDate: "2026-01-31", Amount: "10526.32", Tax: "526.32"}},
		},
		"03-pnd3-contractor-once": {
			RunNo: "2569/0003", Payer: payerCompany, Payee: payeePerson, SequenceNo: "1",
			Form: FormPND3, Condition: ConditionOnce, IssueDate: "2026-02-05",
			Incomes: []Income{
				{Type: Income3Tres, PaidDate: "2026-02-05", Amount: "48500.00", Tax: "1455.00"},
				{Type: IncomeOther, PaidDate: "2026-02-05", Amount: "3000.00", Tax: "30.00", Note: "ค่าขนส่งวัสดุก่อสร้าง (หัก 1%)"},
			},
		},
		"04-pnd2-interest-dividend-other": {
			BookNo: "2", RunNo: "2569/0101", Payer: payerCompany, Payee: payeePerson, SequenceNo: "3",
			Form: FormPND2, Condition: ConditionOther, ConditionNote: "หักจากเงินปันผลค้างจ่าย", IssueDate: "2026-04-30",
			Incomes: []Income{
				{Type: Income404A, PaidDate: "2026-04-30", Amount: "12000.00", Tax: "1800.00"},
				{Type: IncomeDiv11, PaidDate: "2026-04-30", Amount: "50000.00", Tax: "5000.00"},
				{Type: IncomeDiv12, PaidDate: "2026-04-30", Amount: "20000.00", Tax: "2000.00"},
				{Type: IncomeDiv13, PaidDate: "2026-04-30", Amount: "10000.00", Tax: "1000.00"},
				{Type: IncomeDiv14, PaidDate: "2026-04-30", Amount: "8000.00", Tax: "800.00", Note: "15"},
			},
		},
		"05-pnd2a-dividend-no-credit": {
			BookNo: "2", RunNo: "2569/0102", Payer: payerCompany, Payee: payeePerson, SequenceNo: "7",
			Form: FormPND2A, Condition: ConditionWithhold, IssueDate: "2027-01-20",
			Incomes: []Income{
				{Type: IncomeDiv21, PaidDate: "2026", Amount: "30000.00", Tax: "3000.00"},
				{Type: IncomeDiv22, PaidDate: "2026", Amount: "15000.50", Tax: "1500.05"},
				{Type: IncomeDiv23, PaidDate: "2026", Amount: "9000.00", Tax: "900.00"},
				{Type: IncomeDiv24, PaidDate: "2026", Amount: "4000.00", Tax: "400.00"},
				{Type: IncomeDiv25, PaidDate: "2026", Amount: "2500.00", Tax: "250.00", Note: "เงินปันผลจากกำไรที่ได้รับส่งเสริมการลงทุน (BOI)"},
			},
		},
		"06-pnd3a-royalty-archive": {
			RunNo: "2569/0201", Payer: payerCompany, Payee: payeePerson, SequenceNo: "12",
			Form: FormPND3A, Condition: ConditionWithhold, IssueDate: "2026-06-30", ArchiveCopy: true,
			Incomes: []Income{{Type: Income403, PaidDate: "2026-06-30", Amount: "1234567.89", Tax: "61728.39"}},
		},
		"07-pnd1a-special-replacement": {
			BookNo: "3", RunNo: "2569/0301", Payer: payerCompany, Payee: payeePerson, SequenceNo: "5",
			Form: FormPND1ASpecial, Condition: ConditionWithhold, IssueDate: "2027-02-10", Replacement: true,
			Incomes: []Income{{Type: Income402, PaidDate: "2026", Amount: "120000.00", Tax: "6000.00"}},
		},
		// ตรวจตำแหน่ง: กรอกทุกบรรทัด ทุกช่องข้อความ กองทุน และเลข 10 หลัก ในใบเดียว (ใบจริงไม่ใช้แบบนี้)
		"08-layout-every-field": {
			BookNo: "999", RunNo: "2569/99999",
			Payer:      Party{Name: payerCompany.Name, Address: payerCompany.Address, TaxID: payerCompany.TaxID, TaxID10: "3031234567"},
			Payee:      Party{Name: payeePerson.Name, Address: payeePerson.Address, TaxID: payeePerson.TaxID, TaxID10: "1234567890"},
			SequenceNo: "9999", Form: FormPND1A, Condition: ConditionWithhold, IssueDate: "2027-02-15", ArchiveCopy: true,
			FundGPF: "12500.00", FundSSF: "9000.00", FundPVD: "36000.00",
			Incomes: []Income{
				{Type: Income401, PaidDate: "2026", Amount: "600000.00", Tax: "18500.00"},
				{Type: Income402, PaidDate: "2026-03-31", Amount: "15000.00", Tax: "450.00"},
				{Type: Income403, PaidDate: "2026-04-30", Amount: "20000.00", Tax: "600.00"},
				{Type: Income404A, PaidDate: "2026-05-31", Amount: "3000.00", Tax: "450.00"},
				{Type: IncomeDiv11, PaidDate: "2026-05-31", Amount: "1000.00", Tax: "100.00"},
				{Type: IncomeDiv12, PaidDate: "2026-05-31", Amount: "1100.00", Tax: "110.00"},
				{Type: IncomeDiv13, PaidDate: "2026-05-31", Amount: "1200.00", Tax: "120.00"},
				{Type: IncomeDiv14, PaidDate: "2026-05-31", Amount: "1300.00", Tax: "130.00", Note: "23"},
				{Type: IncomeDiv21, PaidDate: "2026-06-30", Amount: "1400.00", Tax: "140.00"},
				{Type: IncomeDiv22, PaidDate: "2026-06-30", Amount: "1500.00", Tax: "150.00"},
				{Type: IncomeDiv23, PaidDate: "2026-06-30", Amount: "1600.00", Tax: "160.00"},
				{Type: IncomeDiv24, PaidDate: "2026-06-30", Amount: "1700.00", Tax: "170.00"},
				{Type: IncomeDiv25, PaidDate: "2026-06-30", Amount: "1800.00", Tax: "180.00", Note: "อื่น ๆ ตามมติที่ประชุมผู้ถือหุ้น"},
				{Type: Income3Tres, PaidDate: "2026-07-31", Amount: "99999.99", Tax: "2999.99"},
				{Type: IncomeOther, PaidDate: "2026-08-31", Amount: "500.00", Tax: "5.00", Note: "ค่าสินไหมทดแทน"},
			},
		},
	}
}

func TestRenderSamples(t *testing.T) {
	dir := os.Getenv("WHT_CERT_SAMPLE_DIR") // ตั้งเพื่อบันทึกไฟล์ตัวอย่างไว้เปิดดู
	for name, cert := range sampleCertificates() {
		t.Run(name, func(t *testing.T) {
			pdf, err := Render(cert)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if err := api.Validate(bytes.NewReader(pdf), pdfConfig()); err != nil {
				t.Fatalf("invalid pdf: %v", err)
			}
			pages, err := api.PageCount(bytes.NewReader(pdf), pdfConfig())
			if err != nil {
				t.Fatal(err)
			}
			want := 2
			if cert.ArchiveCopy {
				want = 3
			}
			if pages != want {
				t.Fatalf("pages = %d, want %d", pages, want)
			}
			fields, err := api.FormFields(bytes.NewReader(pdf), pdfConfig())
			if err == nil && len(fields) > 0 {
				t.Fatalf("output still has %d AcroForm fields", len(fields))
			}
			if dir != "" {
				if err := os.WriteFile(filepath.Join(dir, name+".pdf"), pdf, 0o644); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestSamplesCoverEveryOption(t *testing.T) {
	forms, conditions, incomes := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, c := range sampleCertificates() {
		forms[c.Form], conditions[c.Condition] = true, true
		for _, in := range c.Incomes {
			incomes[in.Type] = true
		}
	}
	if len(forms) != len(formChecks) || len(conditions) != len(conditionChecks) || len(incomes) != len(incomeRowBaselines) {
		t.Fatalf("samples cover forms %d/%d, conditions %d/%d, income rows %d/%d",
			len(forms), len(formChecks), len(conditions), len(conditionChecks), len(incomes), len(incomeRowBaselines))
	}
}

func TestNormalizeTotalsAndOrder(t *testing.T) {
	c := sampleCertificates()["03-pnd3-contractor-once"]
	c.Incomes[0], c.Incomes[1] = c.Incomes[1], c.Incomes[0] // ส่งมาสลับลำดับ ต้องเรียงตามแบบ
	n, err := normalize(c)
	if err != nil {
		t.Fatal(err)
	}
	if n.incomes[0].Type != Income3Tres || n.totalAmount.StringFixed(2) != "51500.00" || n.totalTax.StringFixed(2) != "1485.00" {
		t.Fatalf("order/totals wrong: %s %s %s", n.incomes[0].Type, n.totalAmount, n.totalTax)
	}
	if got := BahtText(n.totalTax); got != "หนึ่งพันสี่ร้อยแปดสิบห้าบาทถ้วน" {
		t.Fatalf("baht text = %q", got)
	}
	if baht, satang := splitMoney(n.totalAmount); baht != "51,500" || satang != "00" {
		t.Fatalf("split = %s|%s", baht, satang)
	}
}

func TestNormalizeRejects(t *testing.T) {
	base := sampleCertificates()["01-pnd53-service-withhold"]
	cases := map[string]struct {
		mutate func(*Certificate)
		key    string
	}{
		"bad checksum":       {func(c *Certificate) { c.Payee.TaxID = c.Payee.TaxID[:12] + "0" }, "wht_cert_taxid_checksum"},
		"short id":           {func(c *Certificate) { c.Payer.TaxID = "12345" }, "wht_cert_taxid_invalid"},
		"tax over amount":    {func(c *Certificate) { c.Incomes[0].Tax = "25000.01" }, "wht_cert_tax_exceeds_amount"},
		"float-ish amount":   {func(c *Certificate) { c.Incomes[0].Amount = "1e5" }, "wht_cert_amount_invalid"},
		"three decimals":     {func(c *Certificate) { c.Incomes[0].Amount = "10.005" }, "wht_cert_amount_invalid"},
		"negative":           {func(c *Certificate) { c.Incomes[0].Tax = "-1" }, "wht_cert_amount_invalid"},
		"duplicate row":      {func(c *Certificate) { c.Incomes = append(c.Incomes, c.Incomes[0]) }, "wht_cert_income_type_duplicate"},
		"unknown row":        {func(c *Certificate) { c.Incomes[0].Type = "40_9" }, "wht_cert_income_type_invalid"},
		"note on plain row":  {func(c *Certificate) { c.Incomes[0].Note = "x" }, "wht_cert_income_note_unexpected"},
		"missing other note": {func(c *Certificate) { c.Incomes[0].Type = IncomeOther }, "wht_cert_income_note_required"},
		"condition other":    {func(c *Certificate) { c.Condition = ConditionOther }, "wht_cert_condition_note_required"},
		"bad form":           {func(c *Certificate) { c.Form = "1" }, "wht_cert_form_invalid"},
		"bad issue date":     {func(c *Certificate) { c.IssueDate = "2026-02-30" }, "wht_cert_issue_date_invalid"},
		"bad paid date":      {func(c *Certificate) { c.Incomes[0].PaidDate = "15/01/2569" }, "wht_cert_paid_date_invalid"},
		"no incomes":         {func(c *Certificate) { c.Incomes = nil }, "wht_cert_income_required"},
		"no payer address":   {func(c *Certificate) { c.Payer.Address = " " }, "wht_cert_payer_address_required"},
		"name too long": {func(c *Certificate) {
			c.Payee.Name = strings.Repeat("บริษัท ยาวมากเป็นพิเศษ ", 8)
		},
			"wht_cert_text_too_long"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := base
			c.Incomes = append([]Income(nil), base.Incomes...)
			tc.mutate(&c)
			_, err := Render(c)
			var ve *ValidationError
			if !errors.As(err, &ve) || ve.Key != tc.key {
				t.Fatalf("err = %v, want key %s", err, tc.key)
			}
		})
	}
}

// TestThaiShaping - สระบน + วรรณยุกต์ต้องไม่ทับกัน: glyph ของ "ปั้ม" ต้องมีวรรณยุกต์ยกสูงกว่าสระ ั
func TestThaiShaping(t *testing.T) {
	if err := prepare(); err != nil {
		t.Fatal(err)
	}
	glyphs, width := textShaper.shape("ปั้ม", fontSize)
	if len(glyphs) < 4 || width <= 0 {
		t.Fatalf("glyphs=%d width=%v", len(glyphs), width)
	}
	if glyphs[0].text == "" {
		t.Fatal("first glyph missing cluster text for ToUnicode")
	}
}
