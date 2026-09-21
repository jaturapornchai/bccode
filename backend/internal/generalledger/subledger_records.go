package generalledger

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var subledgerTaxID = regexp.MustCompile(`^[0-9]{13}$`)
var subledgerTaxBranch = regexp.MustCompile(`^[0-9]{5}$`)

func (m *subledgerMutation) partner(p *SubledgerPartner) error {
	if !validCode(p.Code) || strings.TrimSpace(p.Name) == "" || len(p.Name) > 255 || (!p.IsCustomer && !p.IsSupplier) || (p.TaxID != "" && !subledgerTaxID.MatchString(p.TaxID)) || (p.TaxBranch != "" && !subledgerTaxBranch.MatchString(p.TaxBranch)) {
		return fmt.Errorf("ข้อมูลคู่ค้าไม่ถูกต้อง")
	}
	var old SubledgerPartner
	err := m.loadJSON("gl_subledger_partners", "code", p.Code, &old)
	if err == nil {
		given := p.Version
		p.Version = old.Version
		if jsonEqual(old, *p) {
			return nil
		}
		if given != old.Version {
			return fmt.Errorf("ข้อมูลคู่ค้าถูกแก้ไขแล้ว กรุณาโหลดใหม่")
		}
		var used bool
		if err = m.tx.QueryRowContext(m.ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_documents WHERE company=$1 AND partner_code=$2)`, m.scope.Company, p.Code).Scan(&used); err != nil {
			return err
		}
		if used && (old.IsCustomer != p.IsCustomer || old.IsSupplier != p.IsSupplier || old.TaxID != p.TaxID) {
			return fmt.Errorf("คู่ค้ามีเอกสารแล้ว ห้ามเปลี่ยนบทบาทหรือเลขภาษี")
		}
		p.Version++
	} else if errors.Is(err, sql.ErrNoRows) {
		if p.Version != 0 {
			return ErrNotFound
		}
		p.Version = 1
	} else {
		return err
	}
	return m.saveMetadata("gl_subledger_partners", p.Code, p.Version, p)
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
	if d.BranchCode == "" {
		d.BranchCode = m.journal.BranchCode
	}
	if !subledgerID(d.ID) || !contains([]string{"ar", "ap"}, d.Ledger) || !validCode(d.PartnerCode) || !validDate(d.Date) || (d.DueDate != "" && (!validDate(d.DueDate) || d.DueDate < d.Date)) || strings.TrimSpace(d.DocumentNo) == "" || d.Kind < 1 || d.Kind > 5 || (d.Side != 1 && d.Side != 2) || ((d.Kind == 1 || d.Kind == 3) && d.Side != 1) || ((d.Kind == 2 || d.Kind == 4) && d.Side != 2) {
		return fmt.Errorf("ข้อมูลเอกสารลูกหนี้/เจ้าหนี้ไม่ถูกต้อง")
	}
	if m.scope.Branch != "" && d.BranchCode != m.scope.Branch {
		return ErrNotFound
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
