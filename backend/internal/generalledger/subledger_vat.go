package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// SubledgerVat - ภาษีมูลค่าเพิ่มประกอบใบสำคัญ: 1 แถว = 1 รายการใบกำกับภาษี (ไม่ใช่บรรทัดเดบิต/เครดิต)
// ชื่อฟิลด์ตาม mydocs/datamodels/gl/vat.sql; ฐานภาษีเป็นค่าที่ผู้ใช้กรอกและแก้ได้เสมอ (แนวเดียวกับ Champ BaseOfTax)
// ไม่อนุมานจากบรรทัดบัญชี — ภ.พ.30 และรายงานภาษีซื้อ/ขายใช้ค่าที่บันทึกไว้ตรง ๆ
// เก็บยอดบวกเสมอ: ใบลดหนี้ (document_type 3) รายงานเป็นผู้กลับเครื่องหมาย
type SubledgerVat struct {
	ID                  string  `json:"id"`
	TaxType             int     `json:"tax_type"`      // 1=ภาษีซื้อ, 2=ภาษีขาย
	DocumentType        int     `json:"document_type"` // 1=ใบกำกับภาษี, 2=ใบเพิ่มหนี้, 3=ใบลดหนี้
	TaxInvoiceNo        string  `json:"tax_invoice_no"`
	TaxInvoiceDate      string  `json:"tax_invoice_date"`
	OriginalInvoiceNo   string  `json:"original_invoice_no,omitempty"`
	OriginalInvoiceDate string  `json:"original_invoice_date,omitempty"`
	TaxPeriodYear       int     `json:"tax_period_year,omitempty"` // ค.ศ.; 0 คู่กับเดือน 0 = ยังไม่กำหนดงวดใช้สิทธิ
	TaxPeriodMonth      int     `json:"tax_period_month,omitempty"`
	PartnerCode         string  `json:"partner_code,omitempty"` // ไม่บังคับ — ถ้าระบุต้องเป็นคู่ค้าที่เปิดใช้งาน
	PartnerTaxID        string  `json:"partner_tax_id,omitempty"`
	PartnerBranchNo     string  `json:"partner_branch_no,omitempty"` // 00000 = สำนักงานใหญ่
	PartnerName         string  `json:"partner_name"`
	BaseAmount          Amount  `json:"base_amount"`
	ZeroRateAmount      Amount  `json:"zero_rate_amount"`
	ExemptAmount        Amount  `json:"exempt_amount"`
	Rate                Amount  `json:"vat_rate"`
	VatAmount           *Amount `json:"vat_amount,omitempty"`   // null = ให้ระบบคำนวณ ฐาน × อัตรา
	ClaimStatus         int     `json:"claim_status,omitempty"` // เฉพาะซื้อ: 1=ใช้สิทธิ 2=ต้องห้าม 3=รอใช้สิทธิ 4=ไม่ใช้สิทธิ
	ClaimReason         string  `json:"claim_reason,omitempty"`
	Remark              string  `json:"remark,omitempty"`
}

// NUMERIC(16,2) ใน vat.sql — ยอดต้องน้อยกว่า 10^14
var vatAmountLimit = decimal.New(1, 14)

// errVatRate - อัตราภาษีว่าง/ผิดรูปแบบ; ใช้ทั้งตอนตรวจรายการและตอนพบว่าคำสั่งไม่ส่ง vat_rate มา (requireVatRates)
func errVatRate() *UserError {
	return fieldError("vat_rate_invalid", "vat_rate", "อัตราภาษีมูลค่าเพิ่มต้องอยู่ระหว่าง 0–100 และมีทศนิยมไม่เกิน 2 ตำแหน่ง")
}

// requireVatRates - คำสั่งต้องส่ง vat_rate ทุกรายการ: การคัดลอกรายละเอียดผ่าน JSON (cloneJournalDetails) แปลงอัตราว่างเป็น "0"
// (Amount.MarshalJSON) ทำให้รายการที่ลืมอัตรากลายเป็น 0% เงียบ ๆ และยอดขายเข้าช่องฐานภาษีโดยไม่มีภาษีขาย (UAT V13 2026-09-24)
// — ต้องตรวจก่อนคัดลอก ไม่ใช่หลังคัดลอก
func requireVatRates(details *JournalDetails) error {
	if details == nil {
		return nil
	}
	for _, v := range details.Vats {
		if v.Rate == "" {
			return errVatRate()
		}
	}
	return nil
}

// normalizeVat - ตัดช่องว่าง/ขีดตามรูปมาตรฐานของ vat.sql; ใช้กับทั้งแถวที่ส่งมาและแถวเดิม (rememberVats) ให้เทียบกันได้ตรง ๆ
func normalizeVat(v *SubledgerVat) {
	v.TaxInvoiceNo = strings.TrimSpace(v.TaxInvoiceNo)
	v.OriginalInvoiceNo = strings.TrimSpace(v.OriginalInvoiceNo)
	// เลขที่วางจากใบกำกับภาษีมีขีด/ช่องว่างได้ และสาขาพิมพ์สั้นได้ ("0" = 00000) — เก็บรูปมาตรฐานตาม vat.sql
	v.PartnerTaxID = normalizeTaxID(v.PartnerTaxID)
	v.PartnerBranchNo = normalizeTaxBranch(v.PartnerBranchNo)
	v.PartnerName = strings.TrimSpace(v.PartnerName)
	v.ClaimReason = strings.TrimSpace(v.ClaimReason)
	v.Remark = strings.TrimSpace(v.Remark)
	if v.DocumentType == 0 {
		v.DocumentType = 1 // DEFAULT 1 ตาม vat.sql
	}
}

// vat - ตรวจรายการภาษีมูลค่าเพิ่มตาม CHECK ของ vat.sql; ภาษีว่าง = ฐาน × อัตรา ÷ 100 ปัดครึ่งขึ้นตามทศนิยมของปีบัญชี
// (ผู้ใช้พิมพ์ยอดภาษีเองได้ตามใบกำกับจริง — ใช้ค่าที่พิมพ์)
// กฎที่เพิ่มหลังมีข้อมูลแล้ว (checkVatRowRules) ตรวจเฉพาะแถวใหม่/แถวที่แก้ และตอนผ่านรายการ — ดู vatRowNeedsRules
func (m *subledgerMutation) vat(v *SubledgerVat) error {
	normalizeVat(v)
	// ตรวจทีละช่อง: code + field ของช่องที่ผิด ให้จอชี้ช่องได้ (เดิมรวมหลายสาเหตุเป็น vat_invalid ไม่มี field ผู้ใช้ไม่รู้จะแก้ตรงไหน)
	switch {
	case !subledgerID(v.ID):
		return fieldError("vat_id_invalid", "id", "รายการภาษีมูลค่าเพิ่มไม่มีรหัสรายการ หรือรหัสรายการไม่ถูกต้อง")
	case v.TaxType != 1 && v.TaxType != 2:
		return fieldError("vat_tax_type_invalid", "tax_type", "กรุณาเลือกประเภทภาษี: ภาษีซื้อ หรือ ภาษีขาย")
	case v.DocumentType < 1 || v.DocumentType > 3:
		return fieldError("vat_document_type_invalid", "document_type", "กรุณาเลือกประเภทเอกสาร: ใบกำกับภาษี ใบเพิ่มหนี้ หรือใบลดหนี้")
	case v.TaxInvoiceNo == "" || utf8.RuneCountInString(v.TaxInvoiceNo) > 50:
		return fieldError("vat_invoice_no_invalid", "tax_invoice_no", "กรุณาระบุเลขที่ใบกำกับภาษี (ไม่เกิน 50 ตัวอักษร)")
	case !validDate(v.TaxInvoiceDate):
		return fieldError("vat_invoice_date_invalid", "tax_invoice_date", "กรุณาระบุวันที่ใบกำกับภาษีให้ถูกต้อง")
	case utf8.RuneCountInString(v.OriginalInvoiceNo) > 50:
		return fieldError("vat_original_invoice_no_too_long", "original_invoice_no", "เลขที่ใบกำกับภาษีเดิมต้องไม่เกิน 50 ตัวอักษร")
	case v.OriginalInvoiceDate != "" && !validDate(v.OriginalInvoiceDate):
		return fieldError("vat_original_invoice_date_invalid", "original_invoice_date", "วันที่ใบกำกับภาษีเดิมไม่ถูกต้อง")
	case v.PartnerName == "" || utf8.RuneCountInString(v.PartnerName) > 255:
		return fieldError("vat_partner_name_invalid", "partner_name", "กรุณาระบุชื่อผู้ขายหรือผู้ซื้อตามใบกำกับภาษี (ไม่เกิน 255 ตัวอักษร)")
	case utf8.RuneCountInString(v.ClaimReason) > 500:
		return fieldError("vat_claim_reason_too_long", "claim_reason", "เหตุผลการใช้สิทธิภาษีซื้อต้องไม่เกิน 500 ตัวอักษร")
	case utf8.RuneCountInString(v.Remark) > 500:
		return fieldError("vat_remark_too_long", "remark", "หมายเหตุรายการภาษีมูลค่าเพิ่มต้องไม่เกิน 500 ตัวอักษร")
	}
	if v.PartnerTaxID != "" && !subledgerTaxID.MatchString(v.PartnerTaxID) {
		return fieldError("vat_partner_tax_id_invalid", "partner_tax_id", "เลขประจำตัวผู้เสียภาษีของคู่ค้าต้องเป็นตัวเลข 13 หลัก")
	}
	if v.PartnerBranchNo != "" && !subledgerTaxBranch.MatchString(v.PartnerBranchNo) {
		return fieldError("vat_partner_branch_invalid", "partner_branch_no", "เลขสาขาของคู่ค้าต้องเป็นตัวเลข 5 หลัก (00000 = สำนักงานใหญ่)")
	}
	if v.DocumentType != 1 && (v.OriginalInvoiceNo == "" || v.OriginalInvoiceDate == "") {
		return fieldError("vat_original_invoice_required", "original_invoice_no", "ใบเพิ่มหนี้/ใบลดหนี้ต้องระบุเลขที่และวันที่ใบกำกับภาษีเดิม")
	}
	hasPeriod := v.TaxPeriodYear != 0 || v.TaxPeriodMonth != 0
	if hasPeriod && (v.TaxPeriodYear < 1900 || v.TaxPeriodYear > 9999 || v.TaxPeriodMonth < 1 || v.TaxPeriodMonth > 12) {
		return fieldError("vat_period_invalid", "tax_period_month", "งวดภาษีต้องระบุทั้งปี (ค.ศ.) และเดือน 1–12")
	}
	switch {
	case v.TaxType == 1 && (v.ClaimStatus < 1 || v.ClaimStatus > 4):
		return fieldError("vat_claim_status_required", "claim_status", "ภาษีซื้อต้องระบุสถานะการใช้สิทธิ")
	case v.TaxType == 2 && v.ClaimStatus != 0:
		return fieldError("vat_sale_claim_status", "claim_status", "ภาษีขายไม่มีสถานะการใช้สิทธิ")
	case v.TaxType == 2 && !hasPeriod:
		return fieldError("vat_period_required", "tax_period_month", "ภาษีขายต้องระบุงวดภาษี")
	case v.ClaimStatus == 1 && !hasPeriod:
		return fieldError("vat_claim_period_required", "tax_period_month", "ภาษีซื้อที่ใช้สิทธิต้องระบุงวดภาษี")
	}
	rate := v.Rate.Decimal()
	if err := v.Rate.ValidateScale(2); err != nil || v.Rate == "" || rate.IsNegative() || rate.GreaterThan(hundred) {
		return errVatRate()
	}
	if v.VatAmount == nil {
		vat := amountFromDecimal(v.BaseAmount.Decimal().Mul(rate).Div(hundred).Round(int32(m.scale)))
		v.VatAmount = &vat
	}
	for _, a := range []struct {
		field  string
		amount Amount
	}{{"base_amount", v.BaseAmount}, {"zero_rate_amount", v.ZeroRateAmount}, {"exempt_amount", v.ExemptAmount}, {"vat_amount", *v.VatAmount}} {
		d := a.amount.Decimal()
		if err := a.amount.ValidateScale(m.scale); err != nil || d.IsNegative() || d.GreaterThanOrEqual(vatAmountLimit) {
			return fieldError("vat_amount_invalid", a.field, "ยอดเงินภาษีมูลค่าเพิ่มต้องไม่ติดลบ ไม่เกิน 14 หลัก และทศนิยมตามปีบัญชี")
		}
	}
	var p SubledgerPartner
	if v.PartnerCode != "" {
		if !validCode(v.PartnerCode) || m.loadJSON("gl_subledger_partners", "code", v.PartnerCode, &p) != nil || !p.IsActive {
			return fieldError("vat_partner_not_found", "partner_code", "ไม่พบคู่ค้าของรายการภาษีมูลค่าเพิ่ม หรือคู่ค้าปิดใช้งาน")
		}
	}
	if !m.vatRowNeedsRules(*v) {
		return nil
	}
	// เติมจากทะเบียนก่อนแล้วค่อยตรวจ (review 2026-09-24): เลขเก่าในทะเบียนที่ผิดหลักตรวจสอบเคยผ่านตอนบันทึกร่าง
	// แล้วติดตอนผ่านรายการด้วยข้อความ "เว้นว่างได้" ซึ่งพาวนกลับมาเติมเลขเดิม — เลขที่มาจากทะเบียนชี้ให้แก้ที่ทะเบียนคู่ค้า
	if v.PartnerCode != "" {
		fillVatPartnerTaxID(v, p)
	}
	if err := checkVatRowRules(v); err != nil {
		if user, ok := AsUserError(err); ok && user.Code == "vat_partner_tax_id_checksum" && v.PartnerCode != "" && v.PartnerTaxID == normalizeTaxID(p.TaxID) {
			return fieldError("vat_partner_registry_tax_id_checksum", "partner_tax_id", "เลขประจำตัวผู้เสียภาษีของคู่ค้าในทะเบียนคู่ค้าไม่ถูกต้อง (หลักสุดท้ายไม่ตรงกับ 12 หลักแรก) — ระบบเติมเลขนี้จากทะเบียนคู่ค้า กรุณาแก้เลขในทะเบียนคู่ค้าให้ถูกต้อง แล้วลบเลขในรายการนี้ให้ระบบเติมใหม่ หรือพิมพ์เลขที่ถูกต้องตามใบกำกับภาษีในรายการนี้แทน")
		}
		return err
	}
	return nil
}

// vatRowNeedsRules - แถวที่ไม่เปลี่ยนจากก่อนคำสั่งนี้ (id เดียวกัน ค่าเท่ากันหลัง normalize และคำนวณภาษีแล้ว) ไม่ต้องผ่านกฎใหม่
// เพื่อให้ใบร่างที่บันทึกก่อนมีกฎยังแก้ส่วนอื่น/ลบได้ และการกลับรายการ (reverseJournal ไม่เรียก vat) ไม่ติดแถวเก่า;
// ผ่านรายการ (m.strict) ตรวจทุกแถว เพราะแถวที่ผ่านบัญชีจะเข้ารายงานภาษีและ ภ.พ.30 ทันที — ใบร่างเก่าที่ผิดกฎต้องแก้ก่อนผ่านรายการ
func (m *subledgerMutation) vatRowNeedsRules(v SubledgerVat) bool {
	if m.strict {
		return true
	}
	prev, ok := m.previousVats[v.ID]
	return !ok || !jsonEqual(prev, v)
}

// rememberVats - เก็บรายการภาษีมูลค่าเพิ่มก่อนคำสั่ง (ใบร่างเดิม หรือใบที่ผ่านบัญชีตอน reconcile) ไว้เทียบว่าแถวไหนใหม่/ถูกแก้
func (m *subledgerMutation) rememberVats(d *JournalDetails) {
	if d == nil {
		return
	}
	m.previousVats = map[string]SubledgerVat{}
	for _, v := range d.Vats {
		normalizeVat(&v) // เทียบในรูปเดียวกับค่าที่เพิ่ง normalize ของคำสั่งนี้
		m.previousVats[v.ID] = v
	}
}

// buddhistYearFloor - ปีงวดภาษีตั้งแต่ 2400 ถือว่าเป็น พ.ศ. (ค.ศ. 2400 ยังไม่เกิดขึ้น ส่วน พ.ศ. ปัจจุบันอยู่ช่วง 25xx)
const buddhistYearFloor = 2400

// buddhistDate - วันที่ YYYY-MM-DD ที่ปีเป็น พ.ศ. (≥ buddhistYearFloor); ว่าง/รูปแบบผิด = false (ตรวจรูปแบบที่ validDate แล้ว)
// ใช้กับวันที่ในรายละเอียดภาษี: รายงานภาษีกรองงวดด้วยปี ค.ศ. แถวปี พ.ศ. จึงหายจากแบบยื่นโดยไม่มีคำเตือน
func buddhistDate(value string) bool {
	t, err := time.Parse("2006-01-02", value)
	return err == nil && t.Year() >= buddhistYearFloor
}

// checkVatRowRules - กฎที่ตรวจเฉพาะแถวใหม่/แถวที่แก้ (vatRowNeedsRules): ปีงวดเป็น พ.ศ., ช่วงงวดใช้สิทธิภาษีซื้อ, หลักตรวจสอบเลขผู้เสียภาษี
func checkVatRowRules(v *SubledgerVat) error {
	// ไม่แปลง พ.ศ. → ค.ศ. เอง: ผู้ใช้ต้องเห็นและแก้เอง ไม่ให้งวดภาษีถูกเปลี่ยนเงียบ ๆ
	switch {
	case v.TaxPeriodYear >= buddhistYearFloor:
		return fieldError("vat_period_year_buddhist", "tax_period_year", "ปีงวดภาษีให้ใช้ ค.ศ. เช่น 2026 (ถ้ากรอกปี พ.ศ. ให้ลบ 543)")
	case buddhistDate(v.TaxInvoiceDate):
		return fieldError("vat_invoice_date_buddhist", "tax_invoice_date", "วันที่ใบกำกับภาษีให้ใช้ปี ค.ศ. เช่น 2026-01-10 (ถ้ากรอกปี พ.ศ. ให้ลบ 543)")
	case buddhistDate(v.OriginalInvoiceDate):
		return fieldError("vat_original_invoice_date_buddhist", "original_invoice_date", "วันที่ใบกำกับภาษีเดิมให้ใช้ปี ค.ศ. เช่น 2025-12-15 (ถ้ากรอกปี พ.ศ. ให้ลบ 543)")
	}
	if err := checkPurchaseClaimWindow(v); err != nil {
		return err
	}
	if !checkThaiTaxID(v.PartnerTaxID) {
		return fieldError("vat_partner_tax_id_checksum", "partner_tax_id", "เลขประจำตัวผู้เสียภาษีของคู่ค้าในรายการภาษีมูลค่าเพิ่มไม่ถูกต้อง — หลักสุดท้าย (หลักตรวจสอบ) ไม่ตรงกับ 12 หลักแรก มักเกิดจากพิมพ์ผิดหรือสลับตัวเลข กรุณาตรวจกับใบกำกับภาษีแล้วพิมพ์ใหม่ (เว้นว่างได้ ระบบเติมจากทะเบียนคู่ค้าให้ถ้าเลือกคู่ค้าไว้)")
	}
	return nil
}

// fillVatPartnerTaxID - แถวที่เลือกคู่ค้าแต่ไม่ได้กรอกเลขผู้เสียภาษี/สาขา เติมจากทะเบียนคู่ค้า ณ วันบันทึก (ไม่ทับค่าที่ผู้ใช้พิมพ์)
// สาขาเติมเฉพาะเมื่อเลขผู้เสียภาษีของแถวตรงกับทะเบียน — เลขคนละรายกับทะเบียนแปลว่าผู้ใช้กรอกผู้ออกใบกำกับอื่น สาขาของทะเบียนจึงไม่ใช่ของใบนี้
func fillVatPartnerTaxID(v *SubledgerVat, p SubledgerPartner) {
	normalizePartner(&p)
	if p.TaxID == "" {
		return
	}
	if v.PartnerTaxID == "" {
		v.PartnerTaxID = p.TaxID
	}
	if v.PartnerTaxID == p.TaxID && v.PartnerBranchNo == "" {
		v.PartnerBranchNo = p.TaxBranch
	}
}

// purchaseClaimMonths - ภาษีซื้อที่มิได้นำไปหักในเดือนที่ออกใบกำกับภาษี ใช้สิทธิได้อีกไม่เกิน 6 เดือนถัดไป (ดู checkPurchaseClaimWindow)
const purchaseClaimMonths = 6

// checkPurchaseClaimWindow - ภาษีซื้อที่ "ใช้สิทธิในงวดนี้" (claim_status 1) แยกตามประเภทเอกสาร (document_type)
//
// ใบลดหนี้ (3): งวดต้องไม่ก่อนเดือนของวันที่ใบลดหนี้ — ไม่มีเพดาน 6 เดือน
//   - ม.82/10 วรรคท้าย (https://www.rd.go.th/5206.html): ผู้ได้รับใบลดหนี้ต้องนำภาษีไปหักออกจากภาษีซื้อ "ในเดือนภาษีที่ได้รับใบลดหนี้นั้น"
//     และคำสั่ง ป.80/2542 ข้อ 5 (https://www.rd.go.th/3574.html) ใช้ถ้อยคำเดียวกันเมื่อได้รับในเดือนอื่น — เป็นการลดภาษีซื้อที่บังคับ
//     ไม่ใช่สิทธิที่เลื่อนได้ จึงไม่มีช่วง 6 เดือน; ระบบไม่มีวันที่ได้รับเอกสาร จึงถือว่างวดที่เลือกคือเดือนที่ได้รับ ซึ่งต้องไม่ก่อนวันที่ออก
//
// ใบกำกับภาษี (1) และใบเพิ่มหนี้ (2): เดือนที่ออกเอกสาร ถึงเดือนที่ 6 ถัดจากนั้น
//   - ม.77/1(22) (https://www.rd.go.th/5205.html) "ใบกำกับภาษี" หมายความรวมถึงใบเพิ่มหนี้ ใบลดหนี้ → ช่วงของประกาศฯ ฉบับที่ 4 ใช้กับใบเพิ่มหนี้ด้วย;
//     ม.82/9 วรรคท้าย ผู้ได้รับใบเพิ่มหนี้ถือเป็นภาษีซื้อ "ในเดือนภาษีที่ได้รับใบเพิ่มหนี้นั้น"
//   - ม.82/3 วรรคสอง ประมวลรัษฎากร (https://www.rd.go.th/5206.html): ภาษีซื้อที่มิได้นำไปหักในเดือนภาษีเพราะมีเหตุจำเป็นตามที่อธิบดีกำหนด
//     ให้หักในเดือนภาษีหลังจากนั้นได้ตามหลักเกณฑ์ที่อธิบดีกำหนด "แต่ต้องไม่เกินสามปีนับจากวันที่ได้มีการออกใบกำกับภาษี"
//   - ประกาศอธิบดีฯ เกี่ยวกับภาษีมูลค่าเพิ่ม (ฉบับที่ 4) ข้อ 2 แก้โดยฉบับที่ 76 ใช้บังคับ 1 พ.ค. 2541 (https://www.rd.go.th/3417.html):
//     "ต้องไม่เกินหกเดือนนับแต่เดือนถัดจากเดือนที่ออกใบกำกับภาษี"
//
// ใบกำกับ ม.ค. → ใช้สิทธิได้งวด ม.ค. (ปกติ) ถึง ก.ค.; เกินนั้นต้องบันทึกในงวดของเดือนที่ออกใบกำกับ แล้วยื่น ภ.พ.30 เพิ่มเติม
// ของเดือนนั้น (ม.83/4) และขอคืนเป็นเงินสด (ม.84/1) — ระบบไม่ยอมให้เลื่อนงวดเกินช่วงเงียบ ๆ เพราะ ภ.พ.30 ที่ยื่นจะผิดกฎหมาย
// ปีของงวดภาษีและวันที่ใบกำกับเป็น ค.ศ. ทั้งคู่ (vat.sql) — ปีงวดที่เป็น พ.ศ. ถูกปฏิเสธก่อนถึงที่นี่ (vat_period_year_buddhist)
// สถานะอื่น (ต้องห้าม/รอใช้สิทธิ/ไม่ใช้สิทธิ) ไม่ตรวจจนกว่าจะเปลี่ยนเป็นใช้สิทธิ; ภาษีขาย (ม.82/9, 82/10 ฝั่งผู้ออกเอกสาร) ไม่ตรวจ (ยังไม่มีมติลุงจืด)
// ตรวจเฉพาะแถวใหม่/แถวที่แก้ และตอนผ่านรายการ (vatRowNeedsRules) ทั้งตอนบันทึกใบร่างและตอนแก้หลังผ่านบัญชี (reconcile)
func checkPurchaseClaimWindow(v *SubledgerVat) error {
	if v.TaxType != 1 || v.ClaimStatus != 1 {
		return nil
	}
	issued, err := time.Parse("2006-01-02", v.TaxInvoiceDate)
	if err != nil {
		return fieldError("vat_invoice_date_invalid", "tax_invoice_date", "กรุณาระบุวันที่ใบกำกับภาษีให้ถูกต้อง")
	}
	invoiceMonth := issued.Year()*12 + int(issued.Month()) - 1
	claimMonth := v.TaxPeriodYear*12 + v.TaxPeriodMonth - 1
	if v.DocumentType == 3 {
		if claimMonth < invoiceMonth {
			return fieldError("vat_credit_note_period_before_note", "tax_period_month",
				"ใบลดหนี้ต้องนำไปลดภาษีซื้อในงวดของเดือนที่ได้รับใบลดหนี้ — เลือกงวดตั้งแต่เดือนของวันที่ใบลดหนี้เป็นต้นไป หรือแก้วันที่ใบลดหนี้ให้ถูกต้อง")
		}
		return nil
	}
	switch {
	case claimMonth < invoiceMonth:
		return fieldError("vat_claim_before_invoice_month", "tax_period_month",
			"งวดที่ใช้สิทธิภาษีซื้อต้องไม่ก่อนเดือนที่ออกใบกำกับภาษีหรือใบเพิ่มหนี้ — เลือกงวดตั้งแต่เดือนของวันที่เอกสารเป็นต้นไป หรือแก้วันที่เอกสารให้ถูกต้อง")
	case claimMonth > invoiceMonth+purchaseClaimMonths:
		return fieldError("vat_claim_window_exceeded", "tax_period_month",
			"ภาษีซื้อตามใบกำกับภาษีหรือใบเพิ่มหนี้ใช้สิทธิได้ไม่เกิน 6 เดือนนับแต่เดือนถัดจากเดือนที่ออกเอกสาร — ให้บันทึกในงวดของเดือนที่ออกเอกสาร แล้วยื่น ภ.พ.30 เพิ่มเติมของเดือนนั้น")
	}
	return nil
}

// VatRecord - รายการภาษีมูลค่าเพิ่มพร้อมใบสำคัญต้นทาง สำหรับรายงานภาษีซื้อ/ขาย และ ภ.พ.30
type VatRecord struct {
	JournalID string `json:"journalid"`
	DocNo     string `json:"docno"`
	DocDate   string `json:"docdate"`
	// DuplicateDocNos - เลขที่ใบสำคัญที่ยังมีผล (ร่าง/ผ่านบัญชี) ซึ่งบันทึกใบกำกับภาษีฉบับเดียวกัน (vatInvoiceKeySQL) — ว่าง = ไม่ซ้ำ
	// มีเลขที่ใบสำคัญของแถวนี้เองด้วย = ใบกำกับเดียวกันถูกบันทึกเกิน 1 แถวในใบสำคัญนี้ (สถานะใช้สิทธิเดียวกัน)
	// — แยกแถวต่างสถานะ (เช่น ใช้สิทธิบางส่วน + ต้องห้ามบางส่วน) ไม่นับว่าซ้ำ
	// เป็นคำเตือนให้ตรวจเท่านั้น ไม่บล็อกการบันทึก: ใบกำกับฉบับหนึ่งใช้สิทธิได้ครั้งเดียว (ม.82/5, ประกาศฯ ฉบับที่ 42 สำเนาเป็นภาษีซื้อต้องห้าม)
	// แต่ระบบแยกไม่ได้ว่าเป็นการคีย์ซ้ำ หรือเป็นใบคนละฉบับที่ผู้ออกใช้เลขเดียวกัน — ผู้ใช้ตัดสินจากเอกสารจริง
	DuplicateDocNos []string `json:"duplicatedocnos"`
	SubledgerVat
}

// vatInvoiceKeySQL - ตัวตนของใบกำกับภาษีหนึ่งฉบับ = ประเภทภาษี + ผู้ออกใบกำกับ + เลขที่ (ตัดช่องว่างหัวท้าย ยุบช่องว่าง ตัวพิมพ์ใหญ่) + วันที่ใบกำกับ
// ม.86/4 (https://www.rd.go.th/5208.html): ใบกำกับภาษีมี (2) ชื่อ ที่อยู่ เลขประจำตัวผู้เสียภาษีของ "ผู้ออก" และ (4) "หมายเลขลำดับของใบกำกับภาษี
// และหมายเลขลำดับของเล่มถ้ามี" → เลขที่ไม่ซ้ำเฉพาะภายในผู้ออกรายเดียว: ผู้ขายต่างรายใช้เลขเดียวกันได้ และสาขาต่างกันเป็นคนละผู้ออก
// ใบแทน (คำสั่ง ป.86/2542 ข้อ 25 https://www.rd.go.th/3568.html) ได้เลขใหม่แต่คงวันที่เดิม จึงไม่ชนกับฉบับเดิม — วันที่อยู่ในคีย์เพื่อไม่เตือนเลขที่ผู้ออกเริ่มนับใหม่
//   - ภาษีซื้อ: ผู้ออก = เลขผู้เสียภาษีคู่ค้า (ตัวเลขล้วน) + สาขา (ว่าง = 00000 เฉพาะในคีย์); ไม่มีเลขผู้เสียภาษีใช้รหัสคู่ค้า ไม่มีอีกใช้ชื่อคู่ค้า
//   - ภาษีขาย: ผู้ออก = บริษัทเราเอง (company อยู่ใน WHERE แล้ว) + สาขาของใบสำคัญ
//
// chr(31) (unit separator) คั่นส่วนต่าง ๆ เพื่อไม่ให้ "A|B" + "C" ชนกับ "A" + "B|C"
const vatInvoiceKeySQL = `concat_ws(chr(31),
    v.item->>'tax_type',
    CASE WHEN v.item->>'tax_type' = '2' THEN 'branch:' || COALESCE(r.payload->>'branchcode', '')
    ELSE CASE
      WHEN regexp_replace(COALESCE(v.item->>'partner_tax_id', ''), '[^0-9]', '', 'g') <> ''
        THEN 'tax:' || regexp_replace(v.item->>'partner_tax_id', '[^0-9]', '', 'g')
      WHEN btrim(COALESCE(v.item->>'partner_code', '')) <> '' THEN 'code:' || btrim(v.item->>'partner_code')
      ELSE 'name:' || upper(btrim(regexp_replace(COALESCE(v.item->>'partner_name', ''), '\s+', ' ', 'g')))
    END || chr(31) || COALESCE(NULLIF(btrim(COALESCE(v.item->>'partner_branch_no', '')), ''), '00000')
    END,
    upper(btrim(regexp_replace(COALESCE(v.item->>'tax_invoice_no', ''), '\s+', ' ', 'g'))),
    LEFT(COALESCE(v.item->>'tax_invoice_date', ''), 10))`

// VatRecordsForPeriod - VAT lines of POSTED vouchers of the company whose tax period is year/month.
// taxType 2 = sales (output VAT); taxType 1 = purchases with claim_status 1 (claimed in this period).
// เรียงตามวันที่ใบกำกับ เลขที่ใบสำคัญ และลำดับรายการในใบ; ใบร่าง/ลบ/กลับรายการแล้วไม่นับ
// พร้อมเลขที่ใบสำคัญอื่นที่บันทึกใบกำกับฉบับเดียวกัน (DuplicateDocNos) เทียบทุกงวดของบริษัท และเลขที่ของตัวเองเมื่อใบเดียวกันบันทึกซ้ำในใบสำคัญนี้ — ใบที่ยังมีผลคือ ไม่ถูกลบ, สถานะร่างหรือผ่านบัญชี,
// ไม่ใช่ใบกลับรายการ (ต้นฉบับที่กลับแล้วเป็น reversed จึงไม่เตือนเมื่อบันทึกใหม่) — ทั้งหมดใน SQL เดียวแบบ GROUP BY ไม่ query ทีละรายการ
func VatRecordsForPeriod(ctx context.Context, db *sql.DB, company string, year, month, taxType int) ([]VatRecord, error) {
	if taxType != 1 && taxType != 2 {
		return nil, fmt.Errorf("invalid vat tax type %d", taxType)
	}
	if year < 1900 || year > 9999 || month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid vat tax period %d-%d", year, month)
	}
	// jsonb containment เทียบตัวเลขตรง ๆ — แถวที่ไม่มีงวดหรือชนิดไม่ตรงจะไม่ถูกเลือก โดยไม่ต้อง cast ข้อความ
	filter := map[string]int{"tax_type": taxType, "tax_period_year": year, "tax_period_month": month}
	if taxType == 1 {
		filter["claim_status"] = 1
	}
	rawFilter, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
WITH live AS (
  SELECT r.id AS journal_id, r.code AS doc_no, r.payload->>'status' AS status, COALESCE(r.payload->>'date', '') AS doc_date,
    v.item, v.ord, `+vatInvoiceKeySQL+` AS invoice_key
  FROM gl_records r
  CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(r.payload->'details'->'vats') = 'array'
    THEN r.payload->'details'->'vats' ELSE '[]'::jsonb END) WITH ORDINALITY AS v(item, ord)
  WHERE r.company = $1 AND r.kind = 'journals'
    AND r.payload->>'status' IN ('draft', 'posted')
    AND NOT COALESCE((r.payload->>'isdeleted')::boolean, false)
    AND COALESCE(r.payload->>'kind', '') <> 'reversal'
), period AS (
  SELECT * FROM live WHERE status = 'posted' AND item @> $2::jsonb
), shared AS (
  SELECT invoice_key, array_agg(DISTINCT doc_no ORDER BY doc_no) AS doc_nos
  FROM live
  WHERE invoice_key IN (SELECT invoice_key FROM period)
  GROUP BY invoice_key
  HAVING COUNT(DISTINCT journal_id) > 1
), repeated AS (
  -- ใบเดียวกันซ้ำในใบสำคัญเดียวกัน นับทุกงวดภาษี: ใช้สิทธิ 2 แถวคนละเดือน = ใช้ภาษีซื้อซ้ำใน ภ.พ.30 สองเดือน (review 2026-09-24)
  SELECT journal_id, invoice_key, COALESCE(item->>'claim_status', '') AS claim_status
  FROM live
  WHERE (journal_id, invoice_key) IN (SELECT journal_id, invoice_key FROM period)
  GROUP BY journal_id, invoice_key, COALESCE(item->>'claim_status', '')
  HAVING COUNT(*) > 1
)
SELECT p.journal_id, p.doc_no, p.doc_date, p.item,
  ARRAY(SELECT DISTINCT d FROM unnest(COALESCE(array_remove(s.doc_nos, p.doc_no), '{}'::text[])
    || CASE WHEN rp.journal_id IS NULL THEN '{}'::text[] ELSE ARRAY[p.doc_no] END) AS d ORDER BY d)
FROM period p
LEFT JOIN shared s ON s.invoice_key = p.invoice_key
LEFT JOIN repeated rp ON rp.journal_id = p.journal_id AND rp.invoice_key = p.invoice_key
  AND rp.claim_status = COALESCE(p.item->>'claim_status', '')
ORDER BY p.item->>'tax_invoice_date', p.doc_no, p.ord`, company, string(rawFilter))
	if err != nil {
		return nil, fmt.Errorf("read vat records: %w", err)
	}
	defer rows.Close()
	records := []VatRecord{}
	for rows.Next() {
		var rec VatRecord
		var item []byte
		var duplicates pq.StringArray
		if err := rows.Scan(&rec.JournalID, &rec.DocNo, &rec.DocDate, &item, &duplicates); err != nil {
			return nil, fmt.Errorf("scan vat record: %w", err)
		}
		if err := json.Unmarshal(item, &rec.SubledgerVat); err != nil {
			return nil, fmt.Errorf("parse vat record of %s: %w", rec.DocNo, err)
		}
		rec.DuplicateDocNos = append([]string{}, duplicates...)
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read vat records: %w", err)
	}
	return records, nil
}
