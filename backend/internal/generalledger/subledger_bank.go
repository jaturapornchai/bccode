package generalledger

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (m *subledgerMutation) bankLine(b *SubledgerBankLine) error {
	if b.JournalID == "" {
		b.JournalID = m.journal.ID
	}
	if b.JournalID != m.journal.ID || b.LineNumber < 1 || b.LineNumber > len(m.journal.Lines) || (b.Direction != 1 && b.Direction != 2) {
		return fmt.Errorf("บรรทัดธนาคารต้องอยู่ในใบสำคัญนี้")
	}
	var account SubledgerBankAccount
	if err := m.loadJSON("gl_subledger_bank_accounts", "code", b.BankAccountCode, &account); err != nil || !account.IsActive {
		return fmt.Errorf("ไม่พบบัญชีธนาคารที่เปิดใช้งาน")
	}
	line := m.journal.Lines[b.LineNumber-1]
	if line.AccountCode != account.GLAccountCode {
		return fmt.Errorf("บัญชีธนาคารไม่ตรงบัญชี GL ในบรรทัด")
	}
	amount := line.Debit
	if b.Direction == 2 {
		amount = line.Credit
	}
	if err := m.amount(amount); err != nil {
		return fmt.Errorf("ทิศทางธนาคารไม่ตรงเดบิต/เครดิต")
	}
	var existing []byte
	var oldAmount string
	err := m.tx.QueryRowContext(m.ctx, `SELECT payload,amount::text FROM gl_subledger_bank_lines WHERE company=$1 AND journal_id=$2 AND line_number=$3`, m.scope.Company, b.JournalID, b.LineNumber).Scan(&existing, &oldAmount)
	if err == nil {
		var old SubledgerBankLine
		if json.Unmarshal(existing, &old) != nil || !jsonEqual(old, *b) {
			return fmt.Errorf("บรรทัดธนาคารเดิมมีรายละเอียดต่างกัน")
		}
		v, e := ParseAmount(oldAmount)
		if e != nil || !v.Decimal().Equal(amount.Decimal()) {
			return fmt.Errorf("ยอดบรรทัดธนาคารเปลี่ยนโดยไม่แก้ draft")
		}
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	raw, _ := json.Marshal(b)
	_, err = m.tx.ExecContext(m.ctx, `INSERT INTO gl_subledger_bank_lines(company,journal_id,line_number,bank_account_code,direction,amount,payload) VALUES($1,$2,$3,$4,$5,$6,$7)`, m.scope.Company, b.JournalID, b.LineNumber, b.BankAccountCode, b.Direction, string(amount), string(raw))
	return err
}

func (m *subledgerMutation) statement(s *SubledgerStatementLine) error {
	if !subledgerID(s.ID) || strings.TrimSpace(s.SourceKey) == "" || len(s.SourceKey) > 150 || !validDate(s.Date) || (s.ValueDate != "" && !validDate(s.ValueDate)) || (s.Direction != 1 && s.Direction != 2) || len(s.Description) > 500 || len(s.Reference) > 150 {
		return fmt.Errorf("Statement ต้องมี ID ต้นทาง วันที่ ทิศทางและยอดจริง")
	}
	if err := m.amount(s.Amount); err != nil {
		return err
	}
	if s.BalanceAfter != nil {
		if err := s.BalanceAfter.ValidateScale(m.scale); err != nil {
			return err
		}
	}
	var bank SubledgerBankAccount
	if err := m.loadJSON("gl_subledger_bank_accounts", "code", s.BankAccountCode, &bank); err != nil || !bank.IsActive {
		return fmt.Errorf("บัญชีธนาคารของ Statement ไม่พร้อมใช้งาน")
	}
	var raw []byte
	var owner string
	err := m.tx.QueryRowContext(m.ctx, `SELECT payload,owner_journal_id FROM gl_subledger_statements WHERE company=$1 AND (id=$2 OR (bank_account_code=$3 AND source_key=$4)) ORDER BY id FOR UPDATE`, m.scope.Company, s.ID, s.BankAccountCode, s.SourceKey).Scan(&raw, &owner)
	if err == nil {
		if _, e := m.statementOwner(owner); e != nil {
			return e
		}
		var old SubledgerStatementLine
		if json.Unmarshal(raw, &old) != nil {
			return fmt.Errorf("Statement เดิมไม่ถูกต้อง")
		}
		incoming := s.ID
		copy := *s
		copy.ID = old.ID
		if !jsonEqual(old, copy) {
			return fmt.Errorf("Statement ID/source_key เดิมมียอดหรือรายละเอียดต่างกัน")
		}
		if m.statementAliases == nil {
			m.statementAliases = map[string]string{}
		}
		m.statementAliases[incoming] = old.ID
		s.ID = old.ID
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	raw, err = json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = m.tx.ExecContext(m.ctx, `INSERT INTO gl_subledger_statements(company,id,bank_account_code,source_key,transaction_date,direction,amount,payload,created_at,created_by,owner_journal_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, m.scope.Company, s.ID, s.BankAccountCode, s.SourceKey, s.Date, s.Direction, string(s.Amount), string(raw), m.now, m.scope.Actor, m.journal.ID)
	return err
}

func (m *subledgerMutation) match(v *SubledgerMatch) error {
	if v.JournalID == "" {
		v.JournalID = m.journal.ID
	}
	if alias := m.statementAliases[v.StatementLineID]; alias != "" {
		v.StatementLineID = alias
	}
	if !subledgerID(v.ID) || v.LineNumber < 1 {
		return fmt.Errorf("ข้อมูลจับคู่ Statement ไม่ถูกต้อง")
	}
	if err := m.amount(v.Amount); err != nil {
		return err
	}
	j, err := m.referencedJournal(v.JournalID)
	if err != nil {
		return err
	}
	if j.ID != m.journal.ID && j.Status != "posted" {
		return fmt.Errorf("จับคู่ได้เฉพาะใบสำคัญที่ผ่านบัญชีแล้ว")
	}
	var statementOwner string
	if err = m.tx.QueryRowContext(m.ctx, `SELECT owner_journal_id FROM gl_subledger_statements WHERE company=$1 AND id=$2`, m.scope.Company, v.StatementLineID).Scan(&statementOwner); err != nil {
		return err
	}
	importedBy, err := m.statementOwner(statementOwner)
	if err != nil {
		return err
	}
	if !importedBy.IsDeleted && importedBy.Status != "reversed" {
		m.affected[importedBy.ID] = true
	}
	var statement SubledgerStatementLine
	if err = m.loadJSON("gl_subledger_statements", "id", v.StatementLineID, &statement); err != nil {
		return err
	}
	var bank string
	var direction int
	var amount string
	err = m.tx.QueryRowContext(m.ctx, `SELECT bank_account_code,direction,amount::text FROM gl_subledger_bank_lines WHERE company=$1 AND journal_id=$2 AND line_number=$3 FOR UPDATE`, m.scope.Company, v.JournalID, v.LineNumber).Scan(&bank, &direction, &amount)
	if err != nil {
		return fmt.Errorf("ไม่พบบรรทัดธนาคารของใบสำคัญ")
	}
	if bank != statement.BankAccountCode || direction != statement.Direction {
		return fmt.Errorf("จับคู่ได้เฉพาะบัญชีธนาคารและทิศทางเดียวกัน")
	}
	var account SubledgerBankAccount
	if err = m.loadJSON("gl_subledger_bank_accounts", "code", bank, &account); err != nil || !account.IsActive {
		return fmt.Errorf("บัญชีธนาคารปิดใช้งาน")
	}
	if _, err = m.controlAccount(account.GLAccountCode, "asset"); err != nil {
		return err
	}
	if same, e := m.existingEdge("gl_subledger_matches", v.ID, v); e != nil || same {
		return e
	}
	for _, side := range []struct {
		filter string
		args   []any
		cap    Amount
	}{
		{"statement_id=$2", []any{m.scope.Company, v.StatementLineID}, statement.Amount},
		{"journal_id=$2 AND line_number=$3", []any{m.scope.Company, v.JournalID, v.LineNumber}, Amount(amount)},
	} {
		var used string
		if err = m.tx.QueryRowContext(m.ctx, `SELECT COALESCE(SUM(amount),0)::text FROM gl_subledger_matches WHERE company=$1 AND reversed_at IS NULL AND `+side.filter, side.args...).Scan(&used); err != nil {
			return err
		}
		sum, e := ParseAmount(used)
		if e != nil {
			return e
		}
		if sum.Decimal().Add(v.Amount.Decimal()).GreaterThan(side.cap.Decimal()) {
			return fmt.Errorf("ยอดจับคู่รวมเกิน Statement หรือบรรทัดธนาคาร")
		}
	}
	raw, _ := json.Marshal(v)
	_, err = m.tx.ExecContext(m.ctx, `INSERT INTO gl_subledger_matches(company,id,owner_journal_id,statement_id,journal_id,line_number,amount,payload,created_at,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, m.scope.Company, v.ID, m.journal.ID, v.StatementLineID, v.JournalID, v.LineNumber, string(v.Amount), string(raw), m.now, m.scope.Actor)
	m.affected[v.JournalID] = true
	return err
}

// A statement remains bank evidence after its importing draft is removed or
// reversed. The original journal still determines its immutable branch scope.
func (m *subledgerMutation) statementOwner(id string) (Journal, error) {
	if id == m.journal.ID {
		return *m.journal, nil
	}
	var j Journal
	var raw []byte
	if err := m.tx.QueryRowContext(m.ctx, `SELECT payload||jsonb_build_object('id',id,'version',version) FROM gl_records WHERE company=$1 AND kind='journals' AND id=$2`, m.scope.Company, id).Scan(&raw); err != nil {
		return j, err
	}
	if err := json.Unmarshal(raw, &j); err != nil {
		return j, err
	}
	if m.scope.Branch != "" && j.BranchCode != m.scope.Branch {
		return Journal{}, ErrNotFound
	}
	return j, nil
}
