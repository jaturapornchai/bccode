package whtcert

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// แบบยื่นรายการภาษีหัก ณ ที่จ่ายที่อ้างถึงในช่อง "ลำดับที่ ... ในแบบ"
const (
	FormPND1A        = "1a"
	FormPND1ASpecial = "1a_special"
	FormPND2         = "2"
	FormPND3         = "3"
	FormPND2A        = "2a"
	FormPND3A        = "3a"
	FormPND53        = "53"
)

// ประเภทเงินได้ = บรรทัดในตาราง (หนึ่งบรรทัดต่อหนึ่งประเภท ตามแบบของกรมสรรพากร)
const (
	Income401   = "40_1"
	Income402   = "40_2"
	Income403   = "40_3"
	Income404A  = "40_4a"
	IncomeDiv11 = "40_4b_1_1"
	IncomeDiv12 = "40_4b_1_2"
	IncomeDiv13 = "40_4b_1_3"
	IncomeDiv14 = "40_4b_1_4"
	IncomeDiv21 = "40_4b_2_1"
	IncomeDiv22 = "40_4b_2_2"
	IncomeDiv23 = "40_4b_2_3"
	IncomeDiv24 = "40_4b_2_4"
	IncomeDiv25 = "40_4b_2_5"
	Income3Tres = "3_tres"
	IncomeOther = "other"
)

// ผู้จ่ายเงิน
const (
	ConditionWithhold = "withhold" // (1) หัก ณ ที่จ่าย
	ConditionAlways   = "always"   // (2) ออกให้ตลอดไป
	ConditionOnce     = "once"     // (3) ออกให้ครั้งเดียว
	ConditionOther    = "other"    // (4) อื่น ๆ (ระบุ)
)

// Party - ผู้มีหน้าที่หักภาษี / ผู้ถูกหักภาษี
type Party struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	TaxID   string `json:"taxid"`   // เลขประจำตัวผู้เสียภาษีอากร 13 หลัก (บังคับ)
	TaxID10 string `json:"taxid10"` // เลขประจำตัวผู้เสียภาษีอากรแบบเดิม 10 หลัก (ถ้ามี)
}

// Income - หนึ่งบรรทัดในตาราง "ประเภทเงินได้พึงประเมินที่จ่าย"
type Income struct {
	Type     string `json:"type"`
	PaidDate string `json:"paiddate"` // "YYYY-MM-DD" (ค.ศ.) หรือ "YYYY" = ปีภาษี — พิมพ์เป็น พ.ศ.
	Amount   string `json:"amount"`   // จำนวนเงินที่จ่าย ทศนิยมไม่เกิน 2 ตำแหน่ง (string ห้าม float)
	Tax      string `json:"tax"`      // ภาษีที่หักและนำส่งไว้
	Note     string `json:"note"`     // (1.4) อัตรา / (2.5) และ 6. ข้อความระบุ — บรรทัดอื่นต้องว่าง
}

// Certificate - ข้อมูลหนึ่งใบ 50 ทวิ
type Certificate struct {
	BookNo        string   `json:"bookno"`
	RunNo         string   `json:"runno"`
	Payer         Party    `json:"payer"`
	Payee         Party    `json:"payee"`
	SequenceNo    string   `json:"sequenceno"` // ลำดับที่ในใบแนบ ภ.ง.ด.
	Form          string   `json:"form"`
	Incomes       []Income `json:"incomes"`
	FundGPF       string   `json:"fundgpf"` // เงินที่จ่ายเข้า กบข./กสจ./กองทุนสงเคราะห์ครูโรงเรียนเอกชน
	FundSSF       string   `json:"fundssf"` // กองทุนประกันสังคม
	FundPVD       string   `json:"fundpvd"` // กองทุนสำรองเลี้ยงชีพ
	Condition     string   `json:"condition"`
	ConditionNote string   `json:"conditionnote"`
	IssueDate     string   `json:"issuedate"`   // "YYYY-MM-DD" วันที่ออกหนังสือรับรอง
	ArchiveCopy   bool     `json:"archivecopy"` // เพิ่มหน้า "สำเนาคู่ฉบับ" ให้ผู้ออกเก็บไว้
	Replacement   bool     `json:"replacement"` // พิมพ์ "ใบแทน" ที่หัวเอกสาร
}

// ValidationError - ข้อมูลไม่ครบ/ไม่ถูกต้อง: Key เป็น language key ให้ handler แปลเป็นข้อความตามภาษาผู้ใช้
type ValidationError struct {
	Key   string
	Field string
}

func (e *ValidationError) Error() string { return e.Key + ": " + e.Field }

func invalid(key, field string) error { return &ValidationError{Key: key, Field: field} }

var (
	moneyPattern = regexp.MustCompile(`^[0-9]{1,13}(\.[0-9]{1,2})?$`)
	datePattern  = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
	yearPattern  = regexp.MustCompile(`^[0-9]{4}$`)
	digitsOnly   = regexp.MustCompile(`[\s\-]`)
)

var thaiMonths = []string{"มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
	"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"}

// incomeRowOrder - ลำดับบรรทัดตามแบบ ใช้เรียงผลลัพธ์และตรวจชนิดที่รองรับ
var incomeRowOrder = []string{Income401, Income402, Income403, Income404A, IncomeDiv11, IncomeDiv12, IncomeDiv13,
	IncomeDiv14, IncomeDiv21, IncomeDiv22, IncomeDiv23, IncomeDiv24, IncomeDiv25, Income3Tres, IncomeOther}

// incomeNoteField - บรรทัดที่มีช่อง "ระบุ" บนแบบ
var incomeNoteField = map[string]lineField{
	IncomeDiv14: fieldDividendRate,
	IncomeDiv25: fieldDividendNote,
	IncomeOther: fieldOtherNote,
}

// normalizedIncome - บรรทัดที่ตรวจแล้ว
type normalizedIncome struct {
	Income
	amount, tax decimal.Decimal
	dateText    string
}

// normalized - ใบที่ตรวจแล้ว พร้อมยอดรวมที่ backend คำนวณเอง
type normalized struct {
	Certificate
	incomes               []normalizedIncome
	payerID, payeeID      string
	payerID10, payeeID10  string
	totalAmount, totalTax decimal.Decimal
	funds                 [3]decimal.Decimal
	issued                time.Time
}

// Normalize - ตรวจความถูกต้องทั้งใบ แล้วคำนวณยอดรวมและตัวอักษร (ห้ามเชื่อยอดรวมจาก client)
func normalize(c Certificate) (normalized, error) {
	n := normalized{Certificate: c}
	var err error
	if strings.TrimSpace(c.Payer.Name) == "" {
		return n, invalid("wht_cert_payer_name_required", "payer.name")
	}
	if strings.TrimSpace(c.Payer.Address) == "" {
		return n, invalid("wht_cert_payer_address_required", "payer.address")
	}
	if strings.TrimSpace(c.Payee.Name) == "" {
		return n, invalid("wht_cert_payee_name_required", "payee.name")
	}
	if strings.TrimSpace(c.Payee.Address) == "" {
		return n, invalid("wht_cert_payee_address_required", "payee.address")
	}
	if n.payerID, err = thaiTaxID(c.Payer.TaxID, 13, "payer.taxid"); err != nil {
		return n, err
	}
	if n.payeeID, err = thaiTaxID(c.Payee.TaxID, 13, "payee.taxid"); err != nil {
		return n, err
	}
	if n.payerID10, err = optionalTaxID10(c.Payer.TaxID10, "payer.taxid10"); err != nil {
		return n, err
	}
	if n.payeeID10, err = optionalTaxID10(c.Payee.TaxID10, "payee.taxid10"); err != nil {
		return n, err
	}
	if _, ok := formChecks[c.Form]; !ok {
		return n, invalid("wht_cert_form_invalid", "form")
	}
	if _, ok := conditionChecks[c.Condition]; !ok {
		return n, invalid("wht_cert_condition_invalid", "condition")
	}
	if c.Condition == ConditionOther && strings.TrimSpace(c.ConditionNote) == "" {
		return n, invalid("wht_cert_condition_note_required", "conditionnote")
	}
	if c.Condition != ConditionOther && strings.TrimSpace(c.ConditionNote) != "" {
		return n, invalid("wht_cert_condition_note_unexpected", "conditionnote")
	}
	if n.issued, err = time.Parse("2006-01-02", c.IssueDate); err != nil || !datePattern.MatchString(c.IssueDate) {
		return n, invalid("wht_cert_issue_date_invalid", "issuedate")
	}
	for i, raw := range []string{c.FundGPF, c.FundSSF, c.FundPVD} {
		if n.funds[i], err = optionalMoney(raw, []string{"fundgpf", "fundssf", "fundpvd"}[i]); err != nil {
			return n, err
		}
	}

	if len(c.Incomes) == 0 {
		return n, invalid("wht_cert_income_required", "incomes")
	}
	byType := map[string]normalizedIncome{}
	for i, in := range c.Incomes {
		field := fmt.Sprintf("incomes[%d]", i)
		if _, ok := incomeRowBaselines[in.Type]; !ok {
			return n, invalid("wht_cert_income_type_invalid", field+".type")
		}
		if _, dup := byType[in.Type]; dup {
			return n, invalid("wht_cert_income_type_duplicate", field+".type")
		}
		row := normalizedIncome{Income: in}
		if row.amount, err = requiredMoney(in.Amount, field+".amount"); err != nil {
			return n, err
		}
		if row.tax, err = requiredMoney(in.Tax, field+".tax"); err != nil {
			return n, err
		}
		if row.tax.GreaterThan(row.amount) {
			return n, invalid("wht_cert_tax_exceeds_amount", field+".tax")
		}
		if row.dateText, err = paidDateText(in.PaidDate, field+".paiddate"); err != nil {
			return n, err
		}
		_, hasNote := incomeNoteField[in.Type]
		note := strings.TrimSpace(in.Note)
		if hasNote && note == "" {
			return n, invalid("wht_cert_income_note_required", field+".note")
		}
		if !hasNote && note != "" {
			return n, invalid("wht_cert_income_note_unexpected", field+".note")
		}
		byType[in.Type] = row
	}
	for _, t := range incomeRowOrder {
		if row, ok := byType[t]; ok {
			n.incomes = append(n.incomes, row)
			n.totalAmount = n.totalAmount.Add(row.amount)
			n.totalTax = n.totalTax.Add(row.tax)
		}
	}
	return n, nil
}

// thaiTaxID - เลขประจำตัว 13 หลัก ตรวจหลักตรวจสอบ (mod 11) แบบเดียวกับบัตรประชาชน/เลขนิติบุคคล
func thaiTaxID(raw string, length int, field string) (string, error) {
	id := digitsOnly.ReplaceAllString(raw, "")
	if len(id) != length || strings.Trim(id, "0123456789") != "" {
		return "", invalid("wht_cert_taxid_invalid", field)
	}
	sum := 0
	for i := 0; i < 12; i++ {
		sum += int(id[i]-'0') * (13 - i)
	}
	if (11-sum%11)%10 != int(id[12]-'0') {
		return "", invalid("wht_cert_taxid_checksum", field)
	}
	return id, nil
}

func optionalTaxID10(raw, field string) (string, error) {
	id := digitsOnly.ReplaceAllString(raw, "")
	if id == "" {
		return "", nil
	}
	if len(id) != 10 || strings.Trim(id, "0123456789") != "" {
		return "", invalid("wht_cert_taxid_invalid", field)
	}
	return id, nil
}

func requiredMoney(raw, field string) (decimal.Decimal, error) {
	raw = strings.TrimSpace(raw)
	if !moneyPattern.MatchString(raw) {
		return decimal.Zero, invalid("wht_cert_amount_invalid", field)
	}
	return decimal.RequireFromString(raw), nil
}

func optionalMoney(raw, field string) (decimal.Decimal, error) {
	if strings.TrimSpace(raw) == "" {
		return decimal.Zero, nil
	}
	return requiredMoney(raw, field)
}

// paidDateText - "2026-01-15" → "15/01/2569", "2026" (ปีภาษี) → "2569"
func paidDateText(raw, field string) (string, error) {
	switch {
	case yearPattern.MatchString(raw):
		var year int
		fmt.Sscanf(raw, "%d", &year)
		if year < 1900 || year > 2200 {
			return "", invalid("wht_cert_paid_date_invalid", field)
		}
		return fmt.Sprintf("%d", year+543), nil
	case datePattern.MatchString(raw):
		t, err := time.Parse("2006-01-02", raw)
		if err != nil {
			return "", invalid("wht_cert_paid_date_invalid", field)
		}
		return fmt.Sprintf("%02d/%02d/%d", t.Day(), int(t.Month()), t.Year()+543), nil
	}
	return "", invalid("wht_cert_paid_date_invalid", field)
}

// splitMoney - 1234567.8 → ("1,234,567", "80") สำหรับช่องบาท|สตางค์
func splitMoney(v decimal.Decimal) (string, string) {
	whole, satang, _ := strings.Cut(v.StringFixed(2), ".")
	var b strings.Builder
	for i, ch := range whole {
		if i > 0 && (len(whole)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(ch)
	}
	return b.String(), satang
}

// moneyText - ยอดเงินพร้อมจุลภาค เช่น "12,500.00" (ช่องกองทุนที่ไม่มีช่องสตางค์แยก)
func moneyText(v decimal.Decimal) string {
	baht, satang := splitMoney(v)
	return baht + "." + satang
}
