package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/shopspring/decimal"
)

// VatCalcResult - ผลของปุ่ม "ภาษีแยกนอก" / "ภาษีรวมใน" ในกล่องภาษีของจอบันทึกข้อมูลรายวัน (Champ OnTaxAuto
// glfrmjournal.cpp:1673-1831) เป็นข้อความทศนิยมตามทศนิยมของปีบัญชี — base + vat = total เสมอ
type VatCalcResult struct {
	BaseAmount  Amount `json:"base_amount"`
	VatAmount   Amount `json:"vat_amount"`
	TotalAmount Amount `json:"total_amount"`
	Scale       int    `json:"scale"`
}

// defaultVatCalcScale - ทศนิยมเมื่อยังไม่รู้ปีบัญชี (ไม่ส่ง fiscalyear หรือไม่พบปี) — ตรงกับค่าเริ่มต้นของปีบัญชีใหม่
const defaultVatCalcScale = 2

// VatCalc แยกยอดภาษีมูลค่าเพิ่ม: ไม่บันทึกอะไร ใช้เป็นเครื่องคิดเลขของกล่องภาษีเท่านั้น
//   - exclusive (ภาษีแยกนอก): จำนวนเงิน = ฐานภาษี; ภาษี = ฐาน × อัตรา ÷ 100 ปัดแบบเดียวกับตอนบันทึก (subledger_vat.go vat)
//   - inclusive (ภาษีรวมใน): จำนวนเงิน = ยอดรวมภาษี; ฐาน = ยอดรวม × 100 ÷ (100 + อัตรา) ปัดตามทศนิยม; ภาษี = ยอดรวม − ฐาน
//     (Champ คำนวณภาษีจากฐานที่ปัดแล้วอีกรอบ ทำให้ ฐาน + ภาษี คลาดจากยอดรวมได้ 1 สตางค์ — ที่นี่ให้รวมกลับได้พอดีเสมอ)
//
// อัตราภาษีต้องส่งมาเสมอ ไม่มีค่าเริ่มต้น 7% (mydocs/specs/gl/tax-design-notes.md:19)
func VatCalc(mode, amount, rate string, scale int) (VatCalcResult, error) {
	if scale < 0 || scale > 8 {
		scale = defaultVatCalcScale
	}
	mode = strings.TrimSpace(mode)
	if mode != "exclusive" && mode != "inclusive" {
		return VatCalcResult{}, fieldError("vat_calc_mode_invalid", "mode", "กรุณาเลือกวิธีคำนวณ: ภาษีแยกนอก หรือ ภาษีรวมใน")
	}
	value, err := ParseAmount(strings.TrimSpace(amount))
	if err != nil || value.ValidateScale(scale) != nil || value.Decimal().IsNegative() || value.Decimal().GreaterThanOrEqual(vatAmountLimit) {
		return VatCalcResult{}, fieldError("vat_calc_amount_invalid", "amount", "จำนวนเงินต้องเป็นตัวเลขไม่ติดลบ ไม่เกิน 14 หลัก และทศนิยมตามปีบัญชี")
	}
	r, err := ParseAmount(strings.TrimSpace(rate))
	if err != nil || r.ValidateScale(2) != nil || r.Decimal().IsNegative() || r.Decimal().GreaterThan(hundred) {
		return VatCalcResult{}, errVatRate()
	}
	places := int32(scale)
	var base, vat, total decimal.Decimal
	if mode == "exclusive" {
		base = value.Decimal()
		vat = base.Mul(r.Decimal()).Div(hundred).Round(places)
		total = base.Add(vat)
	} else {
		total = value.Decimal()
		base = total.Mul(hundred).Div(hundred.Add(r.Decimal())).Round(places)
		vat = total.Sub(base)
	}
	if total.GreaterThanOrEqual(vatAmountLimit) {
		return VatCalcResult{}, fieldError("vat_calc_amount_invalid", "amount", "จำนวนเงินต้องเป็นตัวเลขไม่ติดลบ ไม่เกิน 14 หลัก และทศนิยมตามปีบัญชี")
	}
	return VatCalcResult{
		BaseAmount:  Amount(base.StringFixed(places)),
		VatAmount:   Amount(vat.StringFixed(places)),
		TotalAmount: Amount(total.StringFixed(places)),
		Scale:       scale,
	}, nil
}

// FiscalYearScale - ทศนิยมของปีบัญชีตามรหัส (ใช้ปัดผลคำนวณภาษีให้ตรงกับตอนบันทึก); ไม่ระบุหรือไม่พบปี = 2
func (s *PostgresStore) FiscalYearScale(ctx context.Context, scope Scope, code string) (int, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return defaultVatCalcScale, nil
	}
	db, err := s.pg.database(ctx, scope.Holding)
	if err != nil {
		return 0, err
	}
	var data []byte
	err = db.QueryRowContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND code=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, scope.Company, code).Scan(&data)
	if err == sql.ErrNoRows {
		return defaultVatCalcScale, nil
	}
	if err != nil {
		return 0, err
	}
	var year FiscalYear
	if err := json.Unmarshal(data, &year); err != nil {
		return 0, err
	}
	return year.Scale, nil
}
