package generalledger

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"smlcloudplatform/internal/whtcert"
)

var subledgerTaxID = regexp.MustCompile(`^[0-9]{13}$`)
var subledgerTaxBranch = regexp.MustCompile(`^[0-9]{5}$`)
var subledgerDigits = regexp.MustCompile(`^[0-9]+$`)

// normalizeTaxID - เลขประจำตัวผู้เสียภาษีเก็บเป็นตัวเลข 13 หลักล้วน: เลขที่วางจากใบกำกับภาษีมักมีขีด/ช่องว่าง
// (0-1055-58012-34-9) จึงตัดขีดทุกแบบและช่องว่างทิ้ง; อักขระอื่น (ตัวอักษร จุด) ไม่ตัด ให้ถูกปฏิเสธ — ไม่เดาว่าเป็นเลขอะไร
func normalizeTaxID(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || unicode.Is(unicode.Pd, r) || r == '\u2212' {
			return -1
		}
		return r
	}, value)
}

// normalizeTaxBranch - เลขสาขาเก็บ 5 หลัก: ตัวเลขที่สั้นกว่าเติม 0 ข้างหน้า ("0" → "00000" สำนักงานใหญ่, "1" → "00001");
// ว่างคงว่าง (= ไม่ระบุ) และค่าที่ไม่ใช่ตัวเลขล้วนคงไว้ให้ถูกปฏิเสธพร้อมข้อความของช่อง
func normalizeTaxBranch(value string) string {
	v := normalizeTaxID(value)
	if v == "" || len(v) >= 5 || !subledgerDigits.MatchString(v) {
		return v
	}
	return strings.Repeat("0", 5-len(v)) + v
}

func normalizePartner(p *SubledgerPartner) {
	for _, text := range []*string{&p.Name, &p.Address, &p.TitleName, &p.AddrDistrict, &p.AddrProvince, &p.AddrPostcode} {
		*text = strings.TrimSpace(*text)
	}
	p.TaxID = normalizeTaxID(p.TaxID)
	p.TaxBranch = normalizeTaxBranch(p.TaxBranch)
}

// PartnerCodeMaxRunes - gl_partners.partner_code VARCHAR(20) (mydocs/datamodels/gl/partners.sql, ar.sql, ap.sql)
const PartnerCodeMaxRunes = 20

var subledgerPostcode = regexp.MustCompile(`^[0-9]{5}$`)

// samePartnerData - ข้อมูลคู่ค้าเหมือนกันทุกช่อง ไม่นับ version
func samePartnerData(a, b SubledgerPartner) bool {
	a.Version, b.Version = 0, 0
	return jsonEqual(a, b)
}

// validatePartner - กติกาที่ใช้กับทุกแถวคู่ค้า (รวมแถวเดิมที่ผู้ใช้ไม่ได้แก้); กติกาที่เพิ่มภายหลังอยู่ที่ validatePartnerWrite
// รหัสตรวจทีละสาเหตุด้วย checkCode (ช่องว่าง/อักขระ/ความยาว) ให้ข้อความบอกวิธีแก้ — ยอมรับ 60 ตัวอักษรไว้สำหรับรหัสเดิมในทะเบียน
func validatePartner(p *SubledgerPartner) error {
	if err := checkCode(p.Code, "partner_code", "รหัสคู่ค้า", codeMaxRunes); err != nil {
		return err
	}
	switch {
	case p.Name == "":
		return fieldError("partner_name_required", "name_th", "กรุณาระบุชื่อคู่ค้า")
	case utf8.RuneCountInString(p.Name) > 255:
		return fieldError("partner_name_too_long", "name_th", "ชื่อคู่ค้าต้องไม่เกิน 255 ตัวอักษร")
	case !p.IsCustomer && !p.IsSupplier:
		return fieldError("partner_role_required", "is_customer", "กรุณาเลือกบทบาทของคู่ค้าอย่างน้อยหนึ่งอย่าง: ลูกหนี้ หรือ เจ้าหนี้")
	case p.TaxID != "" && !subledgerTaxID.MatchString(p.TaxID):
		return fieldError("partner_tax_id_invalid", "tax_id", "เลขประจำตัวผู้เสียภาษีของคู่ค้าต้องเป็นตัวเลข 13 หลัก (มีขีดหรือเว้นวรรคได้ ระบบตัดให้เอง)")
	case p.TaxBranch != "" && !subledgerTaxBranch.MatchString(p.TaxBranch):
		return fieldError("partner_branch_invalid", "tax_branch_no", "เลขสาขาของคู่ค้าต้องเป็นตัวเลขไม่เกิน 5 หลัก เช่น 0 หรือ 00000 = สำนักงานใหญ่, 1 = สาขา 00001 (เว้นว่างได้ถ้าไม่ระบุ)")
	case utf8.RuneCountInString(p.TitleName) > 100:
		return fieldError("partner_title_name_too_long", "title_name", "คำนำหน้าชื่อคู่ค้าต้องไม่เกิน 100 ตัวอักษร เช่น บริษัท ห้างหุ้นส่วนจำกัด นาย — กรุณาย่อให้สั้นลง (เว้นว่างได้)")
	case utf8.RuneCountInString(p.AddrDistrict) > 50:
		return fieldError("partner_addr_district_too_long", "addr_district", "อำเภอ/เขตของคู่ค้าต้องไม่เกิน 50 ตัวอักษร — ใส่เฉพาะชื่ออำเภอหรือเขต เช่น เขตบางนา (เว้นว่างได้)")
	case utf8.RuneCountInString(p.AddrProvince) > 50:
		return fieldError("partner_addr_province_too_long", "addr_province", "จังหวัดของคู่ค้าต้องไม่เกิน 50 ตัวอักษร — ใส่เฉพาะชื่อจังหวัด เช่น กรุงเทพมหานคร (เว้นว่างได้)")
	case p.AddrPostcode != "" && !subledgerPostcode.MatchString(p.AddrPostcode):
		return fieldError("partner_addr_postcode_invalid", "addr_postcode", "รหัสไปรษณีย์ของคู่ค้าต้องเป็นตัวเลข 5 หลัก เช่น 10110 (เว้นว่างได้ถ้าไม่ระบุ)")
	}
	return nil
}

// validatePartnerWrite - กติกาที่ตรวจเฉพาะตอนเขียนทะเบียน (คู่ค้าใหม่ หรือข้อมูลต่างจากทะเบียน/snapshot เดิมของใบ):
// คู่ค้าเดิมในทะเบียนที่บันทึกก่อนมีกติกา ยังใช้ในใบใหม่ แก้ใบร่าง และกลับรายการได้ ไม่ติดค้าง
//   - รหัสไม่เกิน 20 ตัวอักษรตาม partners.sql — เฉพาะรหัสใหม่ (isNew): รหัสเป็นคีย์ แก้ไม่ได้ รหัสเดิมที่ยาวกว่ายังแก้ข้อมูลอื่นได้
//   - เลขผู้เสียภาษี 13 หลักต้องผ่านหลักตรวจสอบ (checkThaiTaxID) — 50 ทวิ และไฟล์ยื่นภาษีด้วยสื่อบังคับอยู่แล้ว เดิมผู้ใช้รู้ตอนพิมพ์หลังผ่านบัญชี
func validatePartnerWrite(p *SubledgerPartner, isNew bool) error {
	if isNew {
		if err := checkCode(p.Code, "partner_code", "รหัสคู่ค้า", PartnerCodeMaxRunes); err != nil {
			return err
		}
	}
	if !checkThaiTaxID(p.TaxID) {
		return fieldError("partner_tax_id_checksum", "tax_id", "เลขประจำตัวผู้เสียภาษีของคู่ค้าไม่ถูกต้อง — หลักสุดท้าย (หลักตรวจสอบ) ไม่ตรงกับ 12 หลักแรก มักเกิดจากพิมพ์ผิดหรือสลับตัวเลข กรุณาตรวจกับหนังสือรับรองนิติบุคคล บัตรประชาชน หรือใบกำกับภาษีแล้วพิมพ์ใหม่ (เว้นว่างได้ถ้าไม่มีเลข)")
	}
	return nil
}

// checkThaiTaxID - เลขประจำตัวผู้เสียภาษี 13 หลักผ่านหลักตรวจสอบ; ว่าง = ผ่าน (ช่องที่ธุรกิจยอมให้ว่าง)
// หลักที่ 13 เป็นเลขตรวจสอบ (Check Digit) ตามประกาศกรมพัฒนาธุรกิจการค้า พ.ศ. 2548 (เลขนิติบุคคล) และกรมการปกครอง (เลขประจำตัวประชาชน);
// สูตร mod 11 ใช้ตัวเดียวกับหนังสือรับรอง 50 ทวิ/ไฟล์ยื่นด้วยสื่อ (whtcert.ValidThaiTaxID) — docs/kms/21-thai-tax-form-references.md
func checkThaiTaxID(id string) bool {
	return id == "" || whtcert.ValidThaiTaxID(id)
}

// partner - ทะเบียนคู่ค้าที่ฝังในใบสำคัญ (details.partners): ใบสำคัญเก็บ snapshot ของตัวเอง ทะเบียนเก็บค่าปัจจุบัน
//   - แถวที่ข้อมูลเหมือน snapshot เดิมของใบนี้ = ผู้ใช้ไม่ได้แก้ → ไม่เขียนทะเบียน (ใบร่างที่ฝังฉบับเก่ายังบันทึก/ผ่านรายการได้
//     แม้ใบอื่นแก้ทะเบียนไปแล้ว และไม่เอาข้อมูลเก่าไปทับทะเบียน)
//   - version 0 = ผู้ใช้พิมพ์รหัสเอง ไม่ได้โหลดทะเบียน → ข้อมูลที่พิมพ์เป็นข้อมูลล่าสุดของทะเบียน
//   - version > 0 ที่ไม่ตรงทะเบียน และข้อมูลต่างจากทะเบียน = มีคนแก้ทะเบียนหลังผู้ใช้โหลด (optimistic lock) → ปฏิเสธ
//   - มีเอกสารแล้ว: เพิ่มบทบาทและแก้เลขภาษีได้ (รายการ VAT/ภาษีหักเก็บเลข/ชื่อของตัวเองไว้แล้ว ไม่เปลี่ยนตาม)
//     แต่ยกเลิกบทบาทที่ยังมีเอกสารลูกหนี้/เจ้าหนี้ของบทบาทนั้นไม่ได้
func (m *subledgerMutation) partner(p *SubledgerPartner) error {
	normalizePartner(p)
	if err := validatePartner(p); err != nil {
		return err
	}
	var old SubledgerPartner
	err := m.loadJSON("gl_subledger_partners", "code", p.Code, &old)
	if errors.Is(err, sql.ErrNoRows) {
		if p.Version != 0 {
			return ErrNotFound
		}
		if err := validatePartnerWrite(p, true); err != nil {
			return err
		}
		p.Version = 1
		return m.saveMetadata("gl_subledger_partners", p.Code, p.Version, p)
	}
	if err != nil {
		return err
	}
	given := p.Version
	if samePartnerData(old, *p) {
		p.Version = old.Version
		return nil
	}
	if prev, ok := m.previousPartners[p.Code]; ok && samePartnerData(prev, *p) {
		p.Version = prev.Version
		return nil
	}
	if given != 0 && given != old.Version {
		return conflictError("partner_version_conflict", "partner_code", "ข้อมูลคู่ค้ารหัสนี้ในทะเบียนถูกแก้ไขหลังจากที่คุณเปิดข้อมูล — ตรวจข้อมูลคู่ค้าให้ถูกต้อง แล้วลบแถวคู่ค้านี้และเพิ่มแถวใหม่ด้วยรหัสเดิมเพื่อบันทึกทับ")
	}
	if err := validatePartnerWrite(p, false); err != nil {
		return err
	}
	if old.IsCustomer && !p.IsCustomer {
		used, err := m.partnerHasDocuments(p.Code, "ar")
		if err != nil {
			return err
		}
		if used {
			return conflictError("partner_customer_role_in_use", "is_customer", "คู่ค้านี้มีเอกสารลูกหนี้แล้ว จึงยกเลิกบทบาทลูกหนี้ไม่ได้ — เลือกลูกหนี้ไว้ตามเดิม (เพิ่มบทบาทเจ้าหนี้ได้)")
		}
	}
	if old.IsSupplier && !p.IsSupplier {
		used, err := m.partnerHasDocuments(p.Code, "ap")
		if err != nil {
			return err
		}
		if used {
			return conflictError("partner_supplier_role_in_use", "is_supplier", "คู่ค้านี้มีเอกสารเจ้าหนี้แล้ว จึงยกเลิกบทบาทเจ้าหนี้ไม่ได้ — เลือกเจ้าหนี้ไว้ตามเดิม (เพิ่มบทบาทลูกหนี้ได้)")
		}
	}
	p.Version = old.Version + 1
	return m.saveMetadata("gl_subledger_partners", p.Code, p.Version, p)
}

// partnerHasDocuments - คู่ค้ามีเอกสารลูกหนี้ (ar) / เจ้าหนี้ (ap) ในบริษัทนี้แล้วหรือยัง
func (m *subledgerMutation) partnerHasDocuments(code, ledger string) (bool, error) {
	var used bool
	err := m.tx.QueryRowContext(m.ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_documents WHERE company=$1 AND partner_code=$2 AND ledger=$3)`, m.scope.Company, code, ledger).Scan(&used)
	return used, err
}

func (m *subledgerMutation) bankAccount(b *SubledgerBankAccount) error {
	if !validCode(b.Code) || strings.TrimSpace(b.BankName) == "" || strings.TrimSpace(b.AccountNumber) == "" || strings.TrimSpace(b.AccountName) == "" || !validCode(b.GLAccountCode) {
		return fmt.Errorf("ข้อมูลบัญชีธนาคารไม่ครบ")
	}
	if err := subledgerCurrency(&b.Currency); err != nil {
		return err
	}
	if _, err := m.controlAccount(b.GLAccountCode, "asset"); err != nil {
		return err
	}
	var old SubledgerBankAccount
	err := m.loadJSON("gl_subledger_bank_accounts", "code", b.Code, &old)
	if err == nil {
		given := b.Version
		b.Version = old.Version
		if jsonEqual(old, *b) {
			return nil
		}
		if given != old.Version {
			return fmt.Errorf("ข้อมูลบัญชีธนาคารถูกแก้ไขแล้ว")
		}
		var used bool
		if err = m.tx.QueryRowContext(m.ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_bank_lines WHERE company=$1 AND bank_account_code=$2 UNION ALL SELECT 1 FROM gl_subledger_statements WHERE company=$1 AND bank_account_code=$2)`, m.scope.Company, b.Code).Scan(&used); err != nil {
			return err
		}
		if used && (old.GLAccountCode != b.GLAccountCode || old.AccountNumber != b.AccountNumber || old.Currency != b.Currency) {
			return fmt.Errorf("บัญชีธนาคารมีรายการแล้ว ห้ามเปลี่ยนเลขบัญชีหรือบัญชี GL")
		}
		b.Version++
	} else if errors.Is(err, sql.ErrNoRows) {
		if b.Version != 0 {
			return ErrNotFound
		}
		b.Version = 1
	} else {
		return err
	}
	return m.saveMetadata("gl_subledger_bank_accounts", b.Code, b.Version, b)
}

func (m *subledgerMutation) controlAccount(code, kind string) (Account, error) {
	var account Account
	var payload []byte
	err := m.tx.QueryRowContext(m.ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='accounts' AND code=$2`, m.scope.Company, code).Scan(&payload)
	if err != nil {
		return account, fmt.Errorf("ไม่พบบัญชีคุม %s", code)
	}
	if err = json.Unmarshal(payload, &account); err != nil {
		return account, err
	}
	if account.IsDeleted || !account.IsActive || !account.AllowPosting || account.AccountType != kind {
		return account, fmt.Errorf("บัญชีคุม %s ไม่ใช่บัญชี %s ที่ลงรายการได้", code, kind)
	}
	return account, nil
}

func (m *subledgerMutation) document(d *SubledgerDocument) error {
	if d.BranchCode = NormalizeCode(d.BranchCode); d.BranchCode == "" {
		d.BranchCode = m.journal.BranchCode
	}
	if !subledgerID(d.ID) || !contains([]string{"ar", "ap"}, d.Ledger) || !validCode(d.PartnerCode) || !validDate(d.Date) || (d.DueDate != "" && (!validDate(d.DueDate) || d.DueDate < d.Date)) || strings.TrimSpace(d.DocumentNo) == "" || d.Kind < 1 || d.Kind > 5 || (d.Side != 1 && d.Side != 2) || ((d.Kind == 1 || d.Kind == 3) && d.Side != 1) || ((d.Kind == 2 || d.Kind == 4) && d.Side != 2) {
		return fmt.Errorf("ข้อมูลเอกสารลูกหนี้/เจ้าหนี้ไม่ถูกต้อง")
	}
	if m.scope.Branch != "" && d.BranchCode != m.scope.Branch {
		return ErrNotFound
	}
	// หัวใบตรวจสาขากับทะเบียนสาขาแล้ว (httpapi checkJournalBranch) แต่ branch_code ของเอกสารเป็นข้อความอิสระ — เซสชันระดับบริษัท
	// เคยบันทึกสาขาที่ไม่มีจริงได้ เอกสารจึงไม่ขึ้นในรายงาน/การตัดยอดของสาขาใดเลย: เอกสารต้องอยู่สาขาเดียวกับใบสำคัญ
	if d.BranchCode != m.journal.BranchCode {
		return fieldError("subledger_document_branch_mismatch", "branch_code", fmt.Sprintf("สาขาของเอกสารลูกหนี้/เจ้าหนี้ “%s” ต้องเป็นสาขาเดียวกับใบสำคัญ “%s” — เว้นว่างเพื่อใช้สาขาของใบสำคัญ หรือแก้สาขาที่หัวใบสำคัญ", d.BranchCode, m.journal.BranchCode))
	}
	if err := subledgerCurrency(&d.Currency); err != nil {
		return err
	}
	if err := m.amount(d.Amount); err != nil {
		return err
	}
	var p SubledgerPartner
	if err := m.loadJSON("gl_subledger_partners", "code", d.PartnerCode, &p); err != nil || !p.IsActive || (d.Ledger == "ar" && !p.IsCustomer) || (d.Ledger == "ap" && !p.IsSupplier) {
		return fmt.Errorf("คู่ค้าไม่มีสิทธิ์เป็นลูกหนี้/เจ้าหนี้หรือปิดใช้งาน")
	}
	kind := "asset"
	if d.Ledger == "ap" {
		kind = "liability"
	}
	if _, err := m.controlAccount(d.ControlAccountCode, kind); err != nil {
		return err
	}
	var old SubledgerDocument
	err := m.loadJSON("gl_subledger_documents", "id", d.ID, &old)
	if err == nil {
		given := d.Version
		d.Version = old.Version
		if jsonEqual(old, *d) {
			return nil
		}
		if given != old.Version {
			return fmt.Errorf("เอกสารหนี้ถูกแก้ไขแล้ว")
		}
		var mutable bool
		err = m.tx.QueryRowContext(m.ctx, `SELECT created_journal_id=$3 AND NOT EXISTS(SELECT 1 FROM gl_subledger_allocations a JOIN gl_records r ON r.company=a.company AND r.kind='journals' AND r.id=a.journal_id WHERE a.company=$1 AND a.document_id=$2 AND r.payload->>'status' IN ('posted','reversed')) FROM gl_subledger_documents WHERE company=$1 AND id=$2`, m.scope.Company, d.ID, m.journal.ID).Scan(&mutable)
		if err != nil {
			return err
		}
		if !mutable || m.journal.Status != "draft" {
			return fmt.Errorf("เอกสารนี้ผูกบัญชีแล้ว ห้ามเปลี่ยนคู่ค้า/ยอด/ข้อมูลโดยตรง")
		}
		d.Version++
	} else if errors.Is(err, sql.ErrNoRows) {
		if d.Version != 0 {
			return ErrNotFound
		}
		d.Version = 1
	} else {
		return err
	}
	raw, err := json.Marshal(d)
	if err != nil {
		return err
	}
	_, err = m.tx.ExecContext(m.ctx, `INSERT INTO gl_subledger_documents(company,id,ledger,partner_code,branch_code,control_account_code,amount,side,version,created_journal_id,payload) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT(company,id) DO UPDATE SET ledger=EXCLUDED.ledger,partner_code=EXCLUDED.partner_code,branch_code=EXCLUDED.branch_code,control_account_code=EXCLUDED.control_account_code,amount=EXCLUDED.amount,side=EXCLUDED.side,version=EXCLUDED.version,payload=EXCLUDED.payload`, m.scope.Company, d.ID, d.Ledger, d.PartnerCode, d.BranchCode, d.ControlAccountCode, string(d.Amount), d.Side, d.Version, m.journal.ID, string(raw))
	return err
}

func (m *subledgerMutation) loadJSON(table, key, id string, target any) error {
	var data []byte
	err := m.tx.QueryRowContext(m.ctx, `SELECT payload FROM `+table+` WHERE company=$1 AND `+key+`=$2`, m.scope.Company, id).Scan(&data)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
func jsonEqual(a, b any) bool {
	left, e1 := json.Marshal(a)
	right, e2 := json.Marshal(b)
	return e1 == nil && e2 == nil && string(left) == string(right)
}
func (m *subledgerMutation) saveMetadata(table, code string, version int64, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = m.tx.ExecContext(m.ctx, `INSERT INTO `+table+`(company,code,version,payload) VALUES($1,$2,$3,$4) ON CONFLICT(company,code) DO UPDATE SET version=EXCLUDED.version,payload=EXCLUDED.payload`, m.scope.Company, code, version, string(raw))
	return err
}
