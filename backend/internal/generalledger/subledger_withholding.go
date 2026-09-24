package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/whtcert"
)

// SubledgerWithholding - ภาษีหัก ณ ที่จ่ายประกอบใบสำคัญ: 1 แถว = 1 รายการเงินได้บนหนังสือรับรอง 50 ทวิ
// ชื่อฟิลด์ตาม mydocs/datamodels/gl/wht.sql; ฐานภาษีเป็นค่าที่ผู้ใช้กรอกและแก้ได้เสมอ (Champ BCAPWTaxList.BaseOfTax)
// ไม่อนุมานจากบรรทัดบัญชี — ใบที่มีรายการนี้ รายงาน ภ.ง.ด. และ 50 ทวิ ใช้ค่าที่บันทึกไว้ตรง ๆ
// payer_* / payee_* = snapshot ผู้จ่าย(ผู้หัก) / ผู้รับเงิน(ผู้ถูกหัก) ณ วันบันทึก ตาม wht.sql (ไม่บังคับ — ว่าง = ใช้ทะเบียนปัจจุบันตอนพิมพ์)
// เลขผู้เสียภาษีเก็บตัวเลข 13 หลัก, เลขสาขาเก็บ 5 หลัก (ว่าง = ไม่ระบุ)
type SubledgerWithholding struct {
	ID              string  `json:"id"`
	Direction       int     `json:"wht_direction"` // 1=เราหักภาษีผู้รับเงิน, 2=ผู้จ่ายหักภาษีเรา
	FormType        string  `json:"form_type"`     // PND2 | PND3 | PND53 (ไม่มี PND1 — ไม่ทำระบบเงินเดือน)
	PartnerCode     string  `json:"partner_code"`
	BookNo          string  `json:"wht_book_no,omitempty"` // เล่มที่ บนหนังสือรับรอง 50 ทวิ (ไม่บังคับ)
	CertificateNo   string  `json:"wht_cert_no,omitempty"`
	CertificateDate string  `json:"certificate_date,omitempty"`
	PaymentDate     string  `json:"payment_date"`
	IncomeType      string  `json:"income_tax_type"` // รหัสประเภทเงินได้ของแบบ 50 ทวิ เช่น 3_tres, 40_2
	Description     string  `json:"income_description,omitempty"`
	Condition       int     `json:"condition_type"` // 1=หัก ณ ที่จ่าย 2=ออกให้ตลอดไป 3=ออกให้ครั้งเดียว
	Rate            Amount  `json:"wht_rate"`
	BaseAmount      Amount  `json:"base_amount"`
	TaxAmount       *Amount `json:"tax_amount,omitempty"` // null = ให้ระบบคำนวณ ฐาน × อัตรา
	PayerTaxID      string  `json:"payer_tax_id,omitempty"`
	PayerBranchNo   string  `json:"payer_branch_no,omitempty"`
	PayerName       string  `json:"payer_name,omitempty"`
	PayerAddress    string  `json:"payer_address,omitempty"`
	PayeeTaxID      string  `json:"payee_tax_id,omitempty"`
	PayeeBranchNo   string  `json:"payee_branch_no,omitempty"`
	PayeeName       string  `json:"payee_name,omitempty"`
	PayeeAddress    string  `json:"payee_address,omitempty"`
	Remark          string  `json:"remark,omitempty"`
}

// WithholdingParty - ผู้จ่ายหรือผู้รับเงินของรายการภาษีหัก (snapshot ใน wht.sql)
type WithholdingParty struct {
	TaxID    string
	BranchNo string
	Name     string
	Address  string
}

func (p WithholdingParty) blank() bool {
	return p.TaxID == "" && p.BranchNo == "" && p.Name == "" && p.Address == ""
}

// Payer - ผู้จ่ายเงิน/ผู้หักภาษี ตามที่บันทึก
func (w SubledgerWithholding) Payer() WithholdingParty {
	return WithholdingParty{w.PayerTaxID, w.PayerBranchNo, w.PayerName, w.PayerAddress}
}

// Payee - ผู้รับเงิน/ผู้ถูกหักภาษี ตามที่บันทึก
func (w SubledgerWithholding) Payee() WithholdingParty {
	return WithholdingParty{w.PayeeTaxID, w.PayeeBranchNo, w.PayeeName, w.PayeeAddress}
}

// PartnerIsPayer - คู่ค้าของรายการเป็นผู้จ่ายเงิน (ทิศทาง 2 ผู้จ่ายหักภาษีเรา); ทิศทาง 1 คู่ค้าเป็นผู้รับเงิน บริษัทเป็นผู้จ่าย
func (w SubledgerWithholding) PartnerIsPayer() bool { return w.Direction == 2 }

var withholdingForms = map[string]bool{"PND2": true, "PND3": true, "PND53": true}

// pnd2ZeroPIN - เลขผู้เสียภาษีของผู้รับเงิน ภ.ง.ด.2 ที่ไม่มีเลข (ไฟล์ยื่นรับเฉพาะดอกเบี้ย — internal/rdfile/check.go zeroPINAllowed)
const pnd2ZeroPIN = "0000000000000"

var hundred = decimal.NewFromInt(100)

// errWithholdingRateRequired - คำสั่งที่ไม่ส่ง wht_rate: การคัดลอกผ่าน JSON แปลงค่าว่างเป็น "0" จนกลายเป็นหัก 0% ภาษี 0 เงียบ ๆ
func errWithholdingRateRequired() *UserError {
	return fieldError("wht_rate_required", "wht_rate", "กรุณาระบุอัตราภาษีหัก ณ ที่จ่าย (ร้อยละ 0–100) — ถ้าไม่ได้หักภาษีให้ใส่ 0")
}

// requireWithholdingRates - ทุกรายการภาษีหักต้องส่ง wht_rate มา (แบบเดียวกับ requireVatRates) — ตรวจก่อน cloneJournalDetails คัดลอก
func requireWithholdingRates(details *JournalDetails) error {
	if details == nil {
		return nil
	}
	for _, w := range details.Withholdings {
		if strings.TrimSpace(string(w.Rate)) == "" {
			return errWithholdingRateRequired()
		}
	}
	return nil
}

// normalizeWithholdingSnapshot - ตัดช่องว่าง/ขีดของเลขผู้เสียภาษี เติม 0 หน้าเลขสาขา แล้วตรวจทีละช่อง
func normalizeWithholdingSnapshot(w *SubledgerWithholding) error {
	for _, text := range []*string{&w.BookNo, &w.PayerName, &w.PayerAddress, &w.PayeeName, &w.PayeeAddress, &w.Remark} {
		*text = strings.TrimSpace(*text)
	}
	w.PayerTaxID, w.PayeeTaxID = normalizeTaxID(w.PayerTaxID), normalizeTaxID(w.PayeeTaxID)
	w.PayerBranchNo, w.PayeeBranchNo = normalizeTaxBranch(w.PayerBranchNo), normalizeTaxBranch(w.PayeeBranchNo)
	switch {
	case w.PayerTaxID != "" && !subledgerTaxID.MatchString(w.PayerTaxID):
		return fieldError("wht_payer_tax_id_invalid", "payer_tax_id", "เลขประจำตัวผู้เสียภาษีของผู้จ่ายเงิน (ผู้หักภาษี) ต้องเป็นตัวเลข 13 หลัก (มีขีดหรือเว้นวรรคได้ ระบบตัดให้เอง)")
	case w.PayeeTaxID != "" && !subledgerTaxID.MatchString(w.PayeeTaxID):
		return fieldError("wht_payee_tax_id_invalid", "payee_tax_id", "เลขประจำตัวผู้เสียภาษีของผู้รับเงิน (ผู้ถูกหักภาษี) ต้องเป็นตัวเลข 13 หลัก (มีขีดหรือเว้นวรรคได้ ระบบตัดให้เอง)")
	case w.PayerBranchNo != "" && !subledgerTaxBranch.MatchString(w.PayerBranchNo):
		return fieldError("wht_payer_branch_invalid", "payer_branch_no", "เลขสาขาของผู้จ่ายเงินต้องเป็นตัวเลขไม่เกิน 5 หลัก เช่น 0 = สำนักงานใหญ่ (เว้นว่างได้ถ้าไม่ระบุ)")
	case w.PayeeBranchNo != "" && !subledgerTaxBranch.MatchString(w.PayeeBranchNo):
		return fieldError("wht_payee_branch_invalid", "payee_branch_no", "เลขสาขาของผู้รับเงินต้องเป็นตัวเลขไม่เกิน 5 หลัก เช่น 0 = สำนักงานใหญ่ (เว้นว่างได้ถ้าไม่ระบุ)")
	case utf8.RuneCountInString(w.PayerName) > 255:
		return fieldError("wht_payer_name_too_long", "payer_name", "ชื่อผู้จ่ายเงิน (ผู้หักภาษี) ต้องไม่เกิน 255 ตัวอักษร")
	case utf8.RuneCountInString(w.PayeeName) > 255:
		return fieldError("wht_payee_name_too_long", "payee_name", "ชื่อผู้รับเงิน (ผู้ถูกหักภาษี) ต้องไม่เกิน 255 ตัวอักษร")
	case utf8.RuneCountInString(w.BookNo) > 50: // ขนาดเดียวกับเลขที่หนังสือรับรอง (wht_cert_no VARCHAR(50))
		return fieldError("wht_book_no_too_long", "wht_book_no", "เล่มที่ของหนังสือรับรองการหักภาษี ณ ที่จ่ายต้องไม่เกิน 50 ตัวอักษร")
	case utf8.RuneCountInString(w.Remark) > 500:
		return fieldError("wht_remark_too_long", "remark", "หมายเหตุรายการภาษีหัก ณ ที่จ่ายต้องไม่เกิน 500 ตัวอักษร")
	}
	return nil
}

// fillPartnerSnapshot - ฝั่งคู่ค้าที่ผู้ใช้ไม่ได้ระบุเลย (ว่างทั้งชุด) เก็บข้อมูลทะเบียนคู่ค้า ณ วันบันทึก
// ใบที่บันทึกแล้วจึงคงชื่อ/เลข/ที่อยู่เดิมแม้ทะเบียนคู่ค้าถูกแก้ภายหลัง (แก้รายการเองได้ทาง reconcile พร้อมเหตุผล);
// ถ้าผู้ใช้ระบุบางช่อง (เช่น เว้นสาขาว่างโดยตั้งใจ) ไม่เติมช่องที่เหลือ — ฝั่งบริษัทไม่เติม (ทะเบียนบริษัทอยู่นอก GL) ใช้ทะเบียนตอนพิมพ์
//
// ชื่อ/ที่อยู่เก็บแบบเต็มตามหนังสือรับรอง 50 ทวิ (แบบ approve_wh3_081156.pdf: ชื่อ "ให้ระบุว่าเป็น บุคคล นิติบุคคล บริษัท ..."
// ที่อยู่ "ให้ระบุ ... ตำบล/แขวง อำเภอ/เขต จังหวัด") — เดิมเก็บแค่ชื่อไม่มีคำนำหน้าและที่อยู่ไม่มีอำเภอ/จังหวัด/รหัสไปรษณีย์ (UAT 2026-09-24)
// รายงาน/ใบแนบ ภ.ง.ด. แยกคำนำหน้าและอำเภอ/จังหวัดกลับเป็นช่องของตัวเอง (tax_report.go applyWithholdingPartySnapshot)
func fillPartnerSnapshot(w *SubledgerWithholding, p SubledgerPartner) {
	normalizePartner(&p)
	name, address := PartnerFullName(p.TitleName, p.Name), PartnerFullAddress(p.Address, p.AddrDistrict, p.AddrProvince, p.AddrPostcode)
	if utf8.RuneCountInString(name) > 255 { // ช่องชื่อในรายการรับได้ 255 ตัวอักษร — ชื่อยาวสุดในทะเบียนไม่ให้ติด error ที่ผู้ใช้ไม่ได้พิมพ์
		name = p.Name
	}
	if w.PartnerIsPayer() {
		if w.Payer().blank() {
			w.PayerTaxID, w.PayerBranchNo, w.PayerName, w.PayerAddress = p.TaxID, p.TaxBranch, name, address
		}
		return
	}
	if w.Payee().blank() {
		w.PayeeTaxID, w.PayeeBranchNo, w.PayeeName, w.PayeeAddress = p.TaxID, p.TaxBranch, name, address
	}
}

// PartnerFullName - คำนำหน้า + ชื่อตามทะเบียนคู่ค้า; ชื่อที่ขึ้นต้นด้วยคำนำหน้าอยู่แล้ว (เช่น "บริษัท ... จำกัด") ไม่เติมซ้ำ
func PartnerFullName(title, name string) string {
	title, name = strings.Join(strings.Fields(title), " "), strings.Join(strings.Fields(name), " ")
	if title == "" || title == "-" || name == "" || strings.HasPrefix(name, title) {
		return name
	}
	return title + " " + name
}

// PartnerFullAddress - ที่อยู่ + อำเภอ/เขต + จังหวัด + รหัสไปรษณีย์ตามทะเบียนคู่ค้า; ส่วนที่มีในที่อยู่แล้วไม่เติมซ้ำ
func PartnerFullAddress(address, district, province, postcode string) string {
	full := strings.Join(strings.Fields(address), " ")
	for _, part := range []string{district, province, postcode} {
		if part = strings.Join(strings.Fields(part), " "); part != "" && !strings.Contains(full, part) {
			full = strings.TrimSpace(full + " " + part)
		}
	}
	return full
}

// withholding - ตรวจรายการภาษีหัก; ยอดภาษีว่าง = ฐาน × อัตรา ÷ 100 ปัดครึ่งขึ้นตามทศนิยมของปีบัญชี
// (ผู้ใช้พิมพ์ยอดภาษีเองได้ — ใช้ค่าที่พิมพ์ตามแบบ Champ แต่ต้องไม่เกินฐาน)
func (m *subledgerMutation) withholding(w *SubledgerWithholding) error {
	w.CertificateNo = strings.TrimSpace(w.CertificateNo)
	w.IncomeType = strings.TrimSpace(w.IncomeType)
	w.Description = strings.TrimSpace(w.Description)
	// ตรวจทีละช่อง: ข้อความบอกว่าช่องไหนผิด + field ให้จอชี้ช่องได้ (เดิมรวมเป็นข้อความเดียว ผู้ใช้ไม่รู้จะแก้ตรงไหน — UAT S21 2026-09-24)
	switch {
	case !subledgerID(w.ID):
		return fieldError("wht_id_invalid", "id", "รายการภาษีหัก ณ ที่จ่ายไม่มีรหัสรายการ หรือรหัสรายการไม่ถูกต้อง")
	case w.Direction != 1 && w.Direction != 2:
		return fieldError("wht_direction_invalid", "wht_direction", "กรุณาเลือกทิศทางภาษีหัก ณ ที่จ่าย: เราหักภาษีผู้รับเงิน หรือ ผู้จ่ายเงินหักภาษีเรา")
	case !withholdingForms[w.FormType]:
		return fieldError("wht_form_invalid", "form_type", "แบบยื่นภาษีหัก ณ ที่จ่ายต้องเป็น ภ.ง.ด.2 ภ.ง.ด.3 หรือ ภ.ง.ด.53")
	case !validCode(w.PartnerCode):
		return fieldError("wht_partner_required", "partner_code", "กรุณาเลือกคู่ค้าของรายการภาษีหัก ณ ที่จ่าย")
	case !validDate(w.PaymentDate):
		return fieldError("wht_payment_date_invalid", "payment_date", "กรุณาระบุวันที่จ่ายเงินของรายการภาษีหัก ณ ที่จ่ายให้ถูกต้อง")
	case w.CertificateDate != "" && !validDate(w.CertificateDate):
		return fieldError("wht_certificate_date_invalid", "certificate_date", "วันที่หนังสือรับรองการหักภาษี ณ ที่จ่ายไม่ถูกต้อง")
	case w.IncomeType == "" || utf8.RuneCountInString(w.IncomeType) > 50:
		return fieldError("wht_income_type_required", "income_tax_type", "กรุณาเลือกประเภทเงินได้ของรายการภาษีหัก ณ ที่จ่าย")
	case utf8.RuneCountInString(w.Description) > 255:
		return fieldError("wht_description_too_long", "income_description", "คำอธิบายเงินได้ต้องไม่เกิน 255 ตัวอักษร")
	case utf8.RuneCountInString(w.CertificateNo) > 50:
		return fieldError("wht_certificate_no_too_long", "wht_cert_no", "เลขที่หนังสือรับรองการหักภาษี ณ ที่จ่ายต้องไม่เกิน 50 ตัวอักษร")
	case w.Condition < 1 || w.Condition > 3:
		return fieldError("wht_condition_invalid", "condition_type", "กรุณาเลือกเงื่อนไขการหักภาษี: หัก ณ ที่จ่าย ออกให้ตลอดไป หรือออกให้ครั้งเดียว")
	}
	if err := normalizeWithholdingSnapshot(w); err != nil {
		return err
	}
	if strings.TrimSpace(string(w.Rate)) == "" {
		return errWithholdingRateRequired()
	}
	rate := w.Rate.Decimal()
	if err := w.Rate.ValidateScale(2); err != nil || rate.IsNegative() || rate.GreaterThan(hundred) {
		return fieldError("wht_rate_invalid", "wht_rate", "อัตราภาษีหัก ณ ที่จ่ายต้องอยู่ระหว่าง 0–100 และมีทศนิยมไม่เกิน 2 ตำแหน่ง")
	}
	base := w.BaseAmount.Decimal()
	if err := w.BaseAmount.ValidateScale(m.scale); err != nil || base.IsNegative() {
		return fieldError("wht_base_invalid", "base_amount", "ฐานภาษีต้องไม่ติดลบ และทศนิยมตามปีบัญชี")
	}
	if w.TaxAmount == nil {
		tax := amountFromDecimal(base.Mul(rate).Div(hundred).Round(int32(m.scale)))
		w.TaxAmount = &tax
	}
	tax := w.TaxAmount.Decimal()
	if err := w.TaxAmount.ValidateScale(m.scale); err != nil || tax.IsNegative() || tax.GreaterThan(base) {
		return fieldError("wht_tax_invalid", "tax_amount", "ภาษีที่หักต้องไม่ติดลบและไม่เกินฐานภาษี")
	}
	var p SubledgerPartner
	if err := m.loadJSON("gl_subledger_partners", "code", w.PartnerCode, &p); err != nil || !p.IsActive {
		return fieldError("wht_partner_not_found", "partner_code", "ไม่พบคู่ค้าของรายการภาษีหัก ณ ที่จ่าย หรือคู่ค้าปิดใช้งาน")
	}
	prev, hadPrev := m.previousWithholdings[w.ID]
	// แถวที่ไม่เปลี่ยนจากก่อนคำสั่ง (ใบร่าง/ใบที่ผ่านบัญชีที่บันทึกก่อนมีกฎ) ไม่ต้องผ่านหลักตรวจสอบเลขผู้เสียภาษี — ผ่านรายการตรวจทุกแถว
	checkRules := m.strict || !hadPrev || !jsonEqual(prev, *w)
	if hadPrev {
		clearStaleWithholdingSnapshot(w, prev)
	}
	// เติมจากทะเบียนก่อนแล้วค่อยตรวจ: เลขเก่าในทะเบียน (บันทึกก่อนมีกฎหลักตรวจสอบ) เคยผ่านตอนบันทึกร่าง แล้วไปติดตอนผ่านรายการ
	// ด้วยข้อความ "เว้นว่างได้ ระบบเติมจากทะเบียน" ซึ่งพาวนกลับมาเติมเลขเดิม (review 2026-09-24)
	fillPartnerSnapshot(w, p)
	if checkRules {
		return checkWithholdingRowRules(w, normalizeTaxID(p.TaxID))
	}
	return nil
}

// pnd2ZeroPINIncome - ประเภทเงินได้ที่ใช้เลขผู้รับเงิน 0 ทั้ง 13 หลักได้: Format กลาง ภ.ง.ด.2 V2.0 ช่อง D4 "กรณียื่นด้วยสื่อฯ
// เฉพาะการจ่ายดอกเบี้ยเงินฝาก หากไม่มีให้ระบุ 0000000000000" (https://www.rd.go.th/fileadmin/user_upload/WHT/Download/FormatPND2V2_0.pdf)
// — ตรงกับ zeroPINAllowed ใน internal/rdfile/check.go (INC_TYPE 2 = 40(4)(ก))
const pnd2ZeroPINIncome = whtcert.Income404A

// checkWithholdingRowRules - กฎที่ตรวจเฉพาะแถวใหม่/แถวที่แก้ และตอนผ่านรายการ (เหมือน checkVatRowRules): ปี พ.ศ. ในวันที่,
// หลักตรวจสอบเลขผู้เสียภาษีของทั้งสองฝั่งหลังเติมจากทะเบียนคู่ค้า; registryTaxID = เลขในทะเบียนคู่ค้าของแถว (normalize แล้ว)
func checkWithholdingRowRules(w *SubledgerWithholding, registryTaxID string) error {
	// วันที่จ่ายกำหนดเดือนที่ต้องยื่น ภ.ง.ด. (รายงานกรองด้วยปี ค.ศ.) — ปี พ.ศ. ทำให้รายการหายจากแบบและไฟล์ยื่นเงียบ ๆ; ไม่แปลงปีให้เอง
	switch {
	case buddhistDate(w.PaymentDate):
		return fieldError("wht_payment_date_buddhist", "payment_date", "วันที่จ่ายเงินให้ใช้ปี ค.ศ. เช่น 2026-09-15 (ถ้ากรอกปี พ.ศ. ให้ลบ 543) — วันที่จ่ายเป็นตัวกำหนดเดือนที่ต้องยื่นแบบ ภ.ง.ด.")
	case buddhistDate(w.CertificateDate):
		return fieldError("wht_certificate_date_buddhist", "certificate_date", "วันที่หนังสือรับรองการหักภาษี ณ ที่จ่ายให้ใช้ปี ค.ศ. เช่น 2026-09-15 (ถ้ากรอกปี พ.ศ. ให้ลบ 543)")
	}
	partnerPays := w.PartnerIsPayer()
	switch {
	case !checkThaiTaxID(w.PayerTaxID) && partnerPays && w.PayerTaxID == registryTaxID:
		return errWithholdingRegistryTaxID("payer_tax_id")
	case !checkThaiTaxID(w.PayerTaxID):
		return fieldError("wht_payer_tax_id_checksum", "payer_tax_id", "เลขประจำตัวผู้เสียภาษีของผู้จ่ายเงิน (ผู้หักภาษี) ไม่ถูกต้อง — หลักสุดท้าย (หลักตรวจสอบ) ไม่ตรงกับ 12 หลักแรก มักเกิดจากพิมพ์ผิดหรือสลับตัวเลข กรุณาตรวจกับเอกสารแล้วพิมพ์ใหม่ (เว้นว่างได้ถ้าไม่ระบุ)")
	case w.FormType == "PND2" && w.PayeeTaxID == pnd2ZeroPIN:
		// ภ.ง.ด.2 ผู้รับเงินไม่มีเลข ใช้ 0 ทั้ง 13 หลักได้เฉพาะดอกเบี้ย (ไม่ผ่านสูตรหลักตรวจสอบ) — ประเภทอื่นไปติดตอนสร้างไฟล์ยื่นหลังผ่านรายการแล้ว
		if w.IncomeType != pnd2ZeroPINIncome {
			return fieldError("wht_payee_zero_tax_id_interest_only", "payee_tax_id", "เลขผู้เสียภาษี 0000000000000 ใช้ได้เฉพาะ ภ.ง.ด.2 ที่จ่ายดอกเบี้ย มาตรา 40(4)(ก) ตามรูปแบบไฟล์ยื่นของกรมสรรพากร — ประเภทเงินได้อื่นกรุณากรอกเลขประจำตัวผู้เสียภาษี 13 หลักของผู้รับเงิน")
		}
	case !checkThaiTaxID(w.PayeeTaxID) && !partnerPays && w.PayeeTaxID == registryTaxID:
		return errWithholdingRegistryTaxID("payee_tax_id")
	case !checkThaiTaxID(w.PayeeTaxID):
		return fieldError("wht_payee_tax_id_checksum", "payee_tax_id", "เลขประจำตัวผู้เสียภาษีของผู้รับเงิน (ผู้ถูกหักภาษี) ไม่ถูกต้อง — หลักสุดท้าย (หลักตรวจสอบ) ไม่ตรงกับ 12 หลักแรก มักเกิดจากพิมพ์ผิดหรือสลับตัวเลข กรุณาตรวจกับบัตรประชาชนหรือหนังสือรับรองนิติบุคคลแล้วพิมพ์ใหม่ (เว้นว่างได้ ระบบเติมจากทะเบียนคู่ค้า)")
	}
	return nil
}

// errWithholdingRegistryTaxID - เลขผู้เสียภาษีที่ผิดหลักตรวจสอบเป็นเลขเดียวกับทะเบียนคู่ค้า (เติมจากทะเบียน หรือบันทึกไว้จากทะเบียน):
// เว้นว่างไม่ช่วย เพราะระบบเติมเลขเดิมกลับมา — ต้องแก้ที่ทะเบียนคู่ค้า หรือพิมพ์เลขที่ถูกในรายการนี้
func errWithholdingRegistryTaxID(field string) *UserError {
	return fieldError("wht_partner_registry_tax_id_checksum", field, "เลขประจำตัวผู้เสียภาษีของคู่ค้าในทะเบียนคู่ค้าไม่ถูกต้อง (หลักสุดท้ายไม่ตรงกับ 12 หลักแรก) — ระบบเติมเลขนี้จากทะเบียนคู่ค้า กรุณาแก้เลขในทะเบียนคู่ค้าให้ถูกต้อง แล้วลบเลขในรายการนี้ให้ระบบเติมใหม่ หรือพิมพ์เลขที่ถูกต้องในรายการนี้แทน")
}

// rememberWithholdings - เก็บรายการภาษีหักก่อนคำสั่ง (ใบร่างเดิม หรือใบที่ผ่านบัญชีตอน reconcile) ไว้เทียบว่าแถวไหนเปลี่ยนคู่ค้า/ทิศทาง
func (m *subledgerMutation) rememberWithholdings(d *JournalDetails) {
	if d == nil {
		return
	}
	m.previousWithholdings = map[string]SubledgerWithholding{}
	for _, w := range d.Withholdings {
		_ = normalizeWithholdingSnapshot(&w) // เทียบในรูปเดียวกับค่าที่เพิ่ง normalize ของคำสั่งนี้
		m.previousWithholdings[w.ID] = w
	}
}

// clearStaleWithholdingSnapshot - แถวเดิม (id เดียวกัน) ที่เปลี่ยนคู่ค้าหรือทิศทาง: ฝั่งที่ยังเป็นค่า snapshot เดิมทุกช่อง
// คือข้อมูลของคู่ค้าเดิม/ฝั่งเดิมที่จอส่งกลับมา ไม่ใช่ข้อมูลที่ผู้ใช้กรอกใหม่ — ล้างทิ้งให้ fillPartnerSnapshot เติมจากทะเบียนคู่ค้าใหม่
// (เดิมคงชื่อ/เลขผู้เสียภาษีของคู่ค้าเดิมไว้ → ภ.ง.ด./ไฟล์ยื่น/50 ทวิ ออกในนามคู่ค้าเดิม); เปลี่ยนแค่คู่ค้า ฝั่งบริษัทคงเดิม,
// เปลี่ยนทิศทาง ผู้จ่าย/ผู้รับสลับบทบาทจึงล้างทั้งสองฝั่ง; ฝั่งที่ผู้ใช้พิมพ์ค่าใหม่มาในคำสั่งเดียวกันไม่ถูกแตะ
func clearStaleWithholdingSnapshot(w *SubledgerWithholding, prev SubledgerWithholding) {
	if prev.PartnerCode == w.PartnerCode && prev.Direction == w.Direction {
		return
	}
	directionChanged := prev.Direction != w.Direction
	if (directionChanged || w.PartnerIsPayer()) && !w.Payer().blank() && w.Payer() == prev.Payer() {
		w.PayerTaxID, w.PayerBranchNo, w.PayerName, w.PayerAddress = "", "", "", ""
	}
	if (directionChanged || !w.PartnerIsPayer()) && !w.Payee().blank() && w.Payee() == prev.Payee() {
		w.PayeeTaxID, w.PayeeBranchNo, w.PayeeName, w.PayeeAddress = "", "", "", ""
	}
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

// RecordedWithholding - รายการภาษีหักหนึ่งรายการ (details.withholdings[id]) ของใบสำคัญที่ผ่านบัญชีแล้ว พร้อมทะเบียนคู่ค้าปัจจุบัน
// สำหรับออกหนังสือรับรอง 50 ทวิ: ใช้ snapshot ที่บันทึกก่อน ช่องที่ว่างจึงใช้ทะเบียน — ไม่พบ/ยังไม่ผ่านบัญชี/ถูกลบ = ErrNotFound
func RecordedWithholding(ctx context.Context, db *sql.DB, company, journalID, itemID string) (SubledgerWithholding, SubledgerPartner, error) {
	var item SubledgerWithholding
	var partner SubledgerPartner
	var rawItem, rawPartner []byte
	err := db.QueryRowContext(ctx, `
SELECT w.item, p.payload
FROM gl_records r
CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(r.payload->'details'->'withholdings') = 'array'
  THEN r.payload->'details'->'withholdings' ELSE '[]'::jsonb END) AS w(item)
LEFT JOIN gl_subledger_partners p ON p.company = r.company AND p.code = w.item->>'partner_code'
WHERE r.company = $1 AND r.kind = 'journals' AND r.id = $2
  AND r.payload->>'status' = 'posted'
  AND NOT COALESCE((r.payload->>'isdeleted')::boolean, false)
  AND w.item->>'id' = $3
LIMIT 1`, company, journalID, itemID).Scan(&rawItem, &rawPartner)
	if errors.Is(err, sql.ErrNoRows) {
		return item, partner, ErrNotFound
	}
	if err != nil {
		return item, partner, fmt.Errorf("read recorded withholding: %w", err)
	}
	if err = json.Unmarshal(rawItem, &item); err != nil {
		return item, partner, fmt.Errorf("parse recorded withholding %s/%s: %w", journalID, itemID, err)
	}
	if len(rawPartner) > 0 {
		if err = json.Unmarshal(rawPartner, &partner); err != nil {
			return item, partner, fmt.Errorf("parse partner %s: %w", item.PartnerCode, err)
		}
	}
	return item, partner, nil
}
