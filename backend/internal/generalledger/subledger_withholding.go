package generalledger

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/shopspring/decimal"
)

// SubledgerWithholding - ภาษีหัก ณ ที่จ่ายประกอบใบสำคัญ: 1 แถว = 1 รายการเงินได้บนหนังสือรับรอง 50 ทวิ
// ชื่อฟิลด์ตาม mydocs/datamodels/gl/wht.sql; ฐานภาษีเป็นค่าที่ผู้ใช้กรอกและแก้ได้เสมอ (Champ BCAPWTaxList.BaseOfTax)
// ไม่อนุมานจากบรรทัดบัญชี — ใบที่มีรายการนี้ รายงาน ภ.ง.ด. และ 50 ทวิ ใช้ค่าที่บันทึกไว้ตรง ๆ
type SubledgerWithholding struct {
	ID              string  `json:"id"`
	Direction       int     `json:"wht_direction"` // 1=เราหักภาษีผู้รับเงิน, 2=ผู้จ่ายหักภาษีเรา
	FormType        string  `json:"form_type"`     // PND2 | PND3 | PND53 (ไม่มี PND1 — ไม่ทำระบบเงินเดือน)
	PartnerCode     string  `json:"partner_code"`
	CertificateNo   string  `json:"wht_cert_no,omitempty"`
	CertificateDate string  `json:"certificate_date,omitempty"`
	PaymentDate     string  `json:"payment_date"`
	IncomeType      string  `json:"income_tax_type"` // รหัสประเภทเงินได้ของแบบ 50 ทวิ เช่น 3_tres, 40_2
	Description     string  `json:"income_description,omitempty"`
	Condition       int     `json:"condition_type"` // 1=หัก ณ ที่จ่าย 2=ออกให้ตลอดไป 3=ออกให้ครั้งเดียว
	Rate            Amount  `json:"wht_rate"`
	BaseAmount      Amount  `json:"base_amount"`
	TaxAmount       *Amount `json:"tax_amount,omitempty"` // null = ให้ระบบคำนวณ ฐาน × อัตรา
}

var withholdingForms = map[string]bool{"PND2": true, "PND3": true, "PND53": true}

var hundred = decimal.NewFromInt(100)

// withholding - ตรวจรายการภาษีหัก; ยอดภาษีว่าง = ฐาน × อัตรา ÷ 100 ปัดครึ่งขึ้นตามทศนิยมของปีบัญชี
// (ผู้ใช้พิมพ์ยอดภาษีเองได้ — ใช้ค่าที่พิมพ์ตามแบบ Champ แต่ต้องไม่เกินฐาน)
func (m *subledgerMutation) withholding(w *SubledgerWithholding) error {
	w.CertificateNo = strings.TrimSpace(w.CertificateNo)
	w.IncomeType = strings.TrimSpace(w.IncomeType)
	w.Description = strings.TrimSpace(w.Description)
	if !subledgerID(w.ID) || (w.Direction != 1 && w.Direction != 2) || !withholdingForms[w.FormType] || !validCode(w.PartnerCode) ||
		!validDate(w.PaymentDate) || (w.CertificateDate != "" && !validDate(w.CertificateDate)) || w.IncomeType == "" || len(w.IncomeType) > 50 ||
		utf8.RuneCountInString(w.Description) > 255 || utf8.RuneCountInString(w.CertificateNo) > 50 || w.Condition < 1 || w.Condition > 3 {
		return fmt.Errorf("ข้อมูลภาษีหัก ณ ที่จ่ายไม่ครบหรือไม่ถูกต้อง")
	}
	rate := w.Rate.Decimal()
	if err := w.Rate.ValidateScale(2); err != nil || rate.IsNegative() || rate.GreaterThan(hundred) {
		return fmt.Errorf("อัตราภาษีหัก ณ ที่จ่ายต้องอยู่ระหว่าง 0–100 และมีทศนิยมไม่เกิน 2 ตำแหน่ง")
	}
	base := w.BaseAmount.Decimal()
	if err := w.BaseAmount.ValidateScale(m.scale); err != nil || base.IsNegative() {
		return fmt.Errorf("ฐานภาษีต้องไม่ติดลบ และทศนิยมตามปีบัญชี")
	}
	if w.TaxAmount == nil {
		tax := amountFromDecimal(base.Mul(rate).Div(hundred).Round(int32(m.scale)))
		w.TaxAmount = &tax
	}
	tax := w.TaxAmount.Decimal()
	if err := w.TaxAmount.ValidateScale(m.scale); err != nil || tax.IsNegative() || tax.GreaterThan(base) {
		return fmt.Errorf("ภาษีที่หักต้องไม่ติดลบและไม่เกินฐานภาษี")
	}
	var p SubledgerPartner
	if err := m.loadJSON("gl_subledger_partners", "code", w.PartnerCode, &p); err != nil || !p.IsActive {
		return fmt.Errorf("ไม่พบคู่ค้าของรายการภาษีหัก ณ ที่จ่าย หรือคู่ค้าปิดใช้งาน")
	}
	return nil
}
