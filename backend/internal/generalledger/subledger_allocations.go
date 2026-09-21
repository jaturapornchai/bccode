package generalledger

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func (m *subledgerMutation) allocation(a *SubledgerAllocation) error {
	if a.JournalID == "" {
		a.JournalID = m.journal.ID
	}
	if !subledgerID(a.ID) || a.JournalID != m.journal.ID || a.LineNumber < 1 || a.LineNumber > len(m.journal.Lines) {
		return fmt.Errorf("การจัดสรรต้องอ้างบรรทัดของใบสำคัญนี้")
	}
	if err := m.amount(a.Amount); err != nil {
		return err
	}
	var doc SubledgerDocument
	if err := m.loadJSON("gl_subledger_documents", "id", a.DocumentID, &doc); err != nil {
		return fmt.Errorf("ไม่พบเอกสารหนี้ที่จัดสรร")
	}
	if a.Ledger != doc.Ledger || (m.scope.Branch != "" && doc.BranchCode != m.scope.Branch) {
		return ErrNotFound
	}
	if err := m.activePartner(doc.PartnerCode, doc.Ledger); err != nil {
		return err
	}
	kind := "asset"
	if doc.Ledger == "ap" {
		kind = "liability"
	}
	if _, err := m.controlAccount(doc.ControlAccountCode, kind); err != nil {
		return err
	}
	line := m.journal.Lines[a.LineNumber-1]
	if line.AccountCode != doc.ControlAccountCode {
		return fmt.Errorf("การจัดสรรไม่ตรงบัญชีคุมของเอกสาร")
	}
	debitSide := (doc.Ledger == "ar" && doc.Side == 1) || (doc.Ledger == "ap" && doc.Side == 2)
	available := line.Credit.Decimal()
	if debitSide {
		available = line.Debit.Decimal()
	}
	if !available.IsPositive() || a.Amount.Decimal().GreaterThan(available) {
		return fmt.Errorf("ยอด/ทิศทางจัดสรรเกินบรรทัดบัญชีคุม")
	}
	if same, err := m.existingEdge("gl_subledger_allocations", a.ID, a); err != nil || same {
		return err
	}
	raw, _ := json.Marshal(a)
	_, err := m.tx.ExecContext(m.ctx, `INSERT INTO gl_subledger_allocations(company,id,document_id,journal_id,line_number,amount,payload,created_at,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, m.scope.Company, a.ID, a.DocumentID, a.JournalID, a.LineNumber, string(a.Amount), string(raw), m.now, m.scope.Actor)
	return err
}

func (m *subledgerMutation) existingEdge(table, id string, value any) (bool, error) {
	var raw []byte
	var reversed sql.NullTime
	err := m.tx.QueryRowContext(m.ctx, `SELECT payload,reversed_at FROM `+table+` WHERE company=$1 AND id=$2`, m.scope.Company, id).Scan(&raw, &reversed)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var old any
	var next any
	fresh, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	if json.Unmarshal(raw, &old) != nil || json.Unmarshal(fresh, &next) != nil || !jsonEqual(old, next) {
		return false, fmt.Errorf("ID เดิมมีรายละเอียดต่างกัน กรุณาใช้ ID ใหม่")
	}
	if reversed.Valid {
		return false, fmt.Errorf("ID นี้ถูกถอนแล้ว ห้ามนำกลับมาใช้")
	}
	return true, nil
}

func (m *subledgerMutation) allocationCaps() error {
	var over bool
	q := `SELECT EXISTS(SELECT 1 FROM gl_subledger_allocations a JOIN gl_subledger_documents d ON d.company=a.company AND d.id=a.document_id LEFT JOIN gl_records r ON r.company=a.company AND r.kind='journals' AND r.id=a.journal_id WHERE a.company=$1 AND a.reversed_at IS NULL AND (a.journal_id=$2 OR (COALESCE(r.payload->>'status','') IN ('draft','posted') AND NOT COALESCE((r.payload->>'isdeleted')::boolean,false))) GROUP BY d.id,d.amount HAVING SUM(a.amount)>d.amount)`
	if err := m.tx.QueryRowContext(m.ctx, q, m.scope.Company, m.journal.ID).Scan(&over); err != nil {
		return err
	}
	if over {
		return fmt.Errorf("ยอดจัดสรรรวมเกินยอดเอกสารหนี้")
	}
	rows, err := m.tx.QueryContext(m.ctx, `SELECT line_number,SUM(amount)::text FROM gl_subledger_allocations WHERE company=$1 AND journal_id=$2 AND reversed_at IS NULL GROUP BY line_number ORDER BY line_number`, m.scope.Company, m.journal.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var line int
		var amount string
		if err = rows.Scan(&line, &amount); err != nil {
			return err
		}
		if line < 1 || line > len(m.journal.Lines) {
			return fmt.Errorf("บรรทัดจัดสรรไม่มีในใบสำคัญ")
		}
		sum, e := ParseAmount(amount)
		if e != nil {
			return e
		}
		l := m.journal.Lines[line-1]
		if sum.Decimal().GreaterThan(l.Debit.Decimal().Add(l.Credit.Decimal())) {
			return fmt.Errorf("ยอดจัดสรรหลายเอกสารรวมกันเกินบรรทัด GL")
		}
	}
	return rows.Err()
}

func (m *subledgerMutation) completeDocuments() error {
	var incomplete bool
	err := m.tx.QueryRowContext(m.ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_documents d WHERE d.company=$1 AND (d.id=ANY($3) OR d.created_journal_id=$2 OR EXISTS(SELECT 1 FROM gl_subledger_allocations mine WHERE mine.company=d.company AND mine.document_id=d.id AND mine.journal_id=$2 AND mine.reversed_at IS NULL)) AND d.amount<>COALESCE((SELECT SUM(a.amount) FROM gl_subledger_allocations a LEFT JOIN gl_records r ON r.company=a.company AND r.kind='journals' AND r.id=a.journal_id WHERE a.company=d.company AND a.document_id=d.id AND a.reversed_at IS NULL AND (a.journal_id=$2 OR (r.payload->>'status' IN ('draft','posted') AND NOT COALESCE((r.payload->>'isdeleted')::boolean,false)))),0))`, m.scope.Company, m.journal.ID, pq.Array(m.reallocatedDocs)).Scan(&incomplete)
	if err != nil {
		return err
	}
	if incomplete {
		return fmt.Errorf("ก่อนผ่านรายการ ยอดจัดสรรทุกใบสำคัญต้องครบยอดเอกสารหนี้")
	}
	return nil
}

func (m *subledgerMutation) documentFunding(id string) (decimal.Decimal, error) {
	var amount string
	err := m.tx.QueryRowContext(m.ctx, `SELECT COALESCE(SUM(a.amount),0)::text FROM gl_subledger_allocations a LEFT JOIN gl_records r ON r.company=a.company AND r.kind='journals' AND r.id=a.journal_id WHERE a.company=$1 AND a.document_id=$2 AND a.reversed_at IS NULL AND (a.journal_id=$3 OR r.payload->>'status'='posted')`, m.scope.Company, id, m.journal.ID).Scan(&amount)
	if err != nil {
		return decimal.Zero, err
	}
	v, err := ParseAmount(amount)
	return v.Decimal(), err
}

func (m *subledgerMutation) settlement(s *SubledgerSettlement) error {
	if !subledgerID(s.ID) || !validDate(s.Date) || s.Date < m.journal.Date {
		return fmt.Errorf("ข้อมูลตัดยอดหนี้/วันที่ไม่ถูกต้อง")
	}
	if err := m.amount(s.Amount); err != nil {
		return err
	}
	if err := m.store.checkOpenDate(m.ctx, m.tx, m.scope, s.Date); err != nil {
		return err
	}
	var debt, payment SubledgerDocument
	if err := m.loadJSON("gl_subledger_documents", "id", s.DebtDocumentID, &debt); err != nil {
		return err
	}
	if err := m.loadJSON("gl_subledger_documents", "id", s.PaymentDocumentID, &payment); err != nil {
		return err
	}
	if debt.Side != 1 || payment.Side != 2 || debt.Ledger != s.Ledger || payment.Ledger != s.Ledger || debt.PartnerCode != s.PartnerCode || payment.PartnerCode != s.PartnerCode || debt.Currency != payment.Currency || debt.ControlAccountCode != payment.ControlAccountCode || s.Date < debt.Date || s.Date < payment.Date {
		return fmt.Errorf("เอกสารตัดยอดต้องเป็นหนี้/ผลชำระของคู่ค้าและประเภทเดียวกัน")
	}
	if m.scope.Branch != "" && (debt.BranchCode != m.scope.Branch || payment.BranchCode != m.scope.Branch) {
		return ErrNotFound
	}
	if err := m.activePartner(s.PartnerCode, s.Ledger); err != nil {
		return err
	}

	if same, err := m.existingEdge("gl_subledger_settlements", s.ID, s); err != nil || same {
		return err
	}
	for _, id := range []string{s.DebtDocumentID, s.PaymentDocumentID} {
		funded, err := m.documentFunding(id)
		if err != nil {
			return err
		}
		var used string
		err = m.tx.QueryRowContext(m.ctx, `SELECT COALESCE(SUM(amount),0)::text FROM gl_subledger_settlements WHERE company=$1 AND (debt_id=$2 OR payment_id=$2) AND reversed_at IS NULL`, m.scope.Company, id).Scan(&used)
		if err != nil {
			return err
		}
		n, e := ParseAmount(used)
		if e != nil {
			return e
		}
		if n.Decimal().Add(s.Amount.Decimal()).GreaterThan(funded) {
			return fmt.Errorf("ยอดตัดชำระเกินหนี้/ยอดรับจ่ายที่ผ่านบัญชีแล้ว")
		}
		if err = m.touchDocument(id); err != nil {
			return err
		}
	}
	raw, _ := json.Marshal(s)
	_, err := m.tx.ExecContext(m.ctx, `INSERT INTO gl_subledger_settlements(company,id,journal_id,debt_id,payment_id,amount,settlement_date,payload,created_at,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, m.scope.Company, s.ID, m.journal.ID, s.DebtDocumentID, s.PaymentDocumentID, string(s.Amount), s.Date, string(raw), m.now, m.scope.Actor)
	return err
}

func (m *subledgerMutation) activePartner(code, ledger string) error {
	var p SubledgerPartner
	if err := m.loadJSON("gl_subledger_partners", "code", code, &p); err != nil || !p.IsActive || (ledger == "ar" && !p.IsCustomer) || (ledger == "ap" && !p.IsSupplier) {
		return fmt.Errorf("คู่ค้าปิดใช้งานหรือไม่มีบทบาทลูกหนี้/เจ้าหนี้")
	}
	return nil
}

func (m *subledgerMutation) touchDocument(id string) error {
	rows, err := m.tx.QueryContext(m.ctx, `SELECT DISTINCT journal_id FROM gl_subledger_allocations WHERE company=$1 AND document_id=$2 AND reversed_at IS NULL ORDER BY journal_id`, m.scope.Company, id)
	if err != nil {
		return err
	}
	var ids []string
	for rows.Next() {
		var jid string
		if err = rows.Scan(&jid); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, jid)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, jid := range ids {
		if _, err = m.referencedJournal(jid); err != nil {
			return err
		}
		m.affected[jid] = true
	}
	return nil
}
