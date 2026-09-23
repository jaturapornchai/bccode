package generalledger

import (
	"context"
	"database/sql"
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

// WithheldFromCompanyTotal - ภาษีที่ผู้จ่ายเงินหักบริษัทไว้ (wht_direction=2) ตามรายการภาษีหักของใบสำคัญที่ผ่านรายการ
// ที่วันที่จ่ายในหลักฐานอยู่ในช่วง from..to (YYYY-MM-DD) — ใช้เป็นเครดิตภาษีของ ภ.ง.ด.50 ข้อ 3.(3) / ภ.ง.ด.51 ข้อ 5.(1)
// (คู่มือวิธีกรอกแบบของกรมสรรพากร: "ตามหลักฐานที่ถูกหักไว้" — docs/kms/21-thai-tax-form-references.md §2)
// ยอดภาษีเป็นยอดที่บันทึก (ช่องว่างถูกเติม ฐาน × อัตรา ตอนบันทึกแล้ว) รวมด้วย decimal; ใบที่ไม่มีรายการภาษีหักไม่นับ (ไม่ใช่หลักฐาน)
func WithheldFromCompanyTotal(ctx context.Context, db *sql.DB, company, from, to string) (decimal.Decimal, int, error) {
	total, count := decimal.Zero, 0
	if !validDate(from) || !validDate(to) {
		return total, count, fmt.Errorf("invalid withholding period %s..%s", from, to)
	}
	rows, err := db.QueryContext(ctx, `
SELECT COALESCE(w.item->>'tax_amount','')
FROM gl_records r
CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(r.payload->'details'->'withholdings') = 'array'
  THEN r.payload->'details'->'withholdings' ELSE '[]'::jsonb END) AS w(item)
WHERE r.company = $1 AND r.kind = 'journals'
  AND r.payload->>'status' = 'posted'
  AND NOT COALESCE((r.payload->>'isdeleted')::boolean, false)
  AND w.item @> '{"wht_direction":2}'::jsonb
  AND w.item->>'payment_date' BETWEEN $2 AND $3`, company, from, to)
	if err != nil {
		return total, count, fmt.Errorf("read withheld tax: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return total, count, fmt.Errorf("scan withheld tax: %w", err)
		}
		tax, err := decimal.NewFromString(raw)
		if err != nil {
			return total, count, fmt.Errorf("withheld tax amount %q: %w", raw, err)
		}
		total, count = total.Add(tax), count+1
	}
	return total, count, rows.Err()
}
