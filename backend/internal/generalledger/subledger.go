package generalledger

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
)

//go:embed subledger.sql
var subledgerSchema string

type subledgerMutation struct {
	ctx              context.Context
	tx               *sql.Tx
	store            *PostgresStore
	scope            Scope
	journal          *Journal
	now              time.Time
	scale            int
	affected         map[string]bool
	statementAliases map[string]string
	reversalDate     string
	reallocatedDocs  []string
}

func cloneJournalDetails(details *JournalDetails) (*JournalDetails, error) {
	if details == nil {
		return &JournalDetails{}, nil
	}
	raw, err := json.Marshal(details)
	if err != nil {
		return nil, err
	}
	var copy JournalDetails
	err = json.Unmarshal(raw, &copy)
	return &copy, err
}

func (s *PostgresStore) syncJournalDetails(ctx context.Context, tx *sql.Tx, scope Scope, next, old *Journal, action string, now time.Time, effectiveDate ...string) ([]Change, error) {
	// Statement branch ownership must survive later removal of inline details.
	if old != nil && old.BranchCode != next.BranchCode {
		var ownsStatement bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_statements WHERE company=$1 AND owner_journal_id=$2)`, scope.Company, next.ID).Scan(&ownsStatement); err != nil {
			return nil, err
		}
		if ownsStatement {
			return nil, fmt.Errorf("ใบสำคัญที่นำเข้า Statement แล้วไม่สามารถเปลี่ยนสาขา")
		}
	}
	if next.Details == nil && (old == nil || old.Details == nil) {
		return nil, nil
	}
	details, err := cloneJournalDetails(next.Details)
	if err != nil {
		return nil, err
	}
	next.Details = details
	fiscal, err := s.pgYear(ctx, tx, scope, next.FiscalYear)
	if err != nil {
		return nil, err
	}
	m := &subledgerMutation{ctx: ctx, tx: tx, store: s, scope: scope, journal: next, now: now, scale: fiscal.Scale, affected: map[string]bool{}}
	if old != nil && old.Details != nil {
		if err = m.touchDetails(old.Details); err != nil {
			return nil, err
		}
	}
	if len(effectiveDate) > 0 {
		m.reversalDate = effectiveDate[0]
	}
	if action == "reverse" {
		if err = m.reverseJournal(); err != nil {
			return nil, err
		}
		return m.relatedChanges()
	}
	if old != nil && old.Status == "draft" && (action == "update" || action == "delete") {
		if err = m.clearDraftLinks(); err != nil {
			return nil, err
		}
	}
	if action == "delete" {
		if err = m.audit("delete", old.Details); err != nil {
			return nil, err
		}
		return m.relatedChanges()
	}
	if len(details.Withdrawals) > 0 {
		return nil, fmt.Errorf("use reconcile for withdrawal")
	}
	if err = m.apply(details, false); err != nil {
		return nil, err
	}
	if err = m.touchDetails(details); err != nil {
		return nil, err
	}
	if action == "post" {
		if err = m.completeDocuments(); err != nil {
			return nil, err
		}
	}
	if err = m.audit(action, details); err != nil {
		return nil, err
	}
	return m.relatedChanges()
}

func (m *subledgerMutation) touchDetails(d *JournalDetails) error {
	settlements, matches := []string{}, []string{}
	for _, v := range d.Settlements {
		settlements = append(settlements, v.ID)
	}
	for _, v := range d.Matches {
		matches = append(matches, v.ID)
	}
	rows, err := m.tx.QueryContext(m.ctx, `SELECT DISTINCT a.journal_id FROM gl_subledger_allocations a JOIN gl_subledger_settlements s ON s.company=a.company AND a.document_id IN(s.debt_id,s.payment_id) WHERE a.company=$1 AND a.reversed_at IS NULL AND s.reversed_at IS NULL AND s.id=ANY($2) UNION SELECT journal_id FROM gl_subledger_matches WHERE company=$1 AND id=ANY($3) AND reversed_at IS NULL ORDER BY 1`, m.scope.Company, pq.Array(settlements), pq.Array(matches))
	if err != nil {
		return err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, id := range ids {
		if _, err = m.referencedJournal(id); err != nil {
			return err
		}
		m.affected[id] = true
	}
	return nil
}

func (m *subledgerMutation) apply(d *JournalDetails, reconcile bool) error {
	if len(d.Partners)+len(d.BankAccounts)+len(d.Documents)+len(d.Allocations)+len(d.Settlements)+len(d.BankLines)+len(d.StatementLines)+len(d.Matches)+len(d.Withdrawals)+len(d.Withholdings) > 2000 {
		return fmt.Errorf("รายละเอียดประกอบมากเกิน 2000 รายการต่อคำขอ")
	}
	if err := m.lockReferences(d); err != nil {
		return err
	}
	for i := range d.Partners {
		if err := m.partner(&d.Partners[i]); err != nil {
			return err
		}
	}
	for i := range d.BankAccounts {
		if err := m.bankAccount(&d.BankAccounts[i]); err != nil {
			return err
		}
	}
	for i := range d.Documents {
		if err := m.document(&d.Documents[i]); err != nil {
			return err
		}
	}
	for _, w := range d.Withdrawals {
		if !reconcile {
			return fmt.Errorf("ถอนการจับคู่ได้เฉพาะ reconcile")
		}
		if err := m.withdraw(w); err != nil {
			return err
		}
	}
	for i := range d.Allocations {
		if err := m.allocation(&d.Allocations[i]); err != nil {
			return err
		}
	}
	for i := range d.BankLines {
		if err := m.bankLine(&d.BankLines[i]); err != nil {
			return err
		}
	}
	for i := range d.StatementLines {
		if err := m.statement(&d.StatementLines[i]); err != nil {
			return err
		}
	}
	for i := range d.Settlements {
		if err := m.settlement(&d.Settlements[i]); err != nil {
			return err
		}
	}
	for i := range d.Matches {
		if err := m.match(&d.Matches[i]); err != nil {
			return err
		}
	}
	for i := range d.Withholdings {
		if err := m.withholding(&d.Withholdings[i]); err != nil {
			return err
		}
	}
	return m.allocationCaps()
}

func (m *subledgerMutation) lockReferences(d *JournalDetails) error {
	docs, statements, journals := map[string]bool{}, map[string]bool{}, map[string]bool{m.journal.ID: true}
	for _, v := range d.Documents {
		docs[v.ID] = true
	}
	for _, v := range d.Allocations {
		docs[v.DocumentID] = true
		if v.JournalID != "" {
			journals[v.JournalID] = true
		}
	}
	for _, v := range d.Settlements {
		docs[v.DebtDocumentID] = true
		docs[v.PaymentDocumentID] = true
	}
	for _, v := range d.StatementLines {
		statements[v.ID] = true
	}
	for _, v := range d.Matches {
		statements[v.StatementLineID] = true
		if v.JournalID != "" {
			journals[v.JournalID] = true
		}
	}
	for _, entry := range []struct {
		table string
		ids   map[string]bool
	}{{"gl_subledger_documents", docs}, {"gl_subledger_statements", statements}, {"gl_records", journals}} {
		keys := make([]string, 0, len(entry.ids))
		for id := range entry.ids {
			keys = append(keys, id)
		}
		sort.Strings(keys)
		filter := ""
		if entry.table == "gl_records" {
			filter = " AND kind='journals'"
		}
		rows, err := m.tx.QueryContext(m.ctx, `SELECT id FROM `+entry.table+` WHERE company=$1 AND id=ANY($2)`+filter+` ORDER BY id FOR UPDATE`, m.scope.Company, pq.Array(keys))
		if err != nil {
			return err
		}
		for rows.Next() {
			var id string
			if err = rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *subledgerMutation) audit(action string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = m.tx.ExecContext(m.ctx, `INSERT INTO gl_subledger_audit(company,journal_id,action,actor,occurred_at,payload) VALUES($1,$2,$3,$4,$5,$6)`, m.scope.Company, m.journal.ID, action, m.scope.Actor, m.now, string(raw))
	return err
}

func (m *subledgerMutation) amount(a Amount) error {
	if err := a.ValidateScale(m.scale); err != nil {
		return err
	}
	if !a.Decimal().IsPositive() {
		return fmt.Errorf("ยอดจัดสรรต้องมากกว่าศูนย์")
	}
	return nil
}
func subledgerID(id string) bool {
	return len(id) <= 100 && strings.TrimSpace(id) == id && id != "" && !strings.ContainsAny(id, "\x00\r\n")
}
func subledgerCurrency(currency *string) error {
	if *currency == "" {
		*currency = "THB"
	}
	if *currency != "THB" {
		return fmt.Errorf("รายละเอียดประกอบรองรับเฉพาะ THB")
	}
	return nil
}

func (m *subledgerMutation) referencedJournal(id string) (Journal, error) {
	if id == "" || id == m.journal.ID {
		return *m.journal, nil
	}
	var j Journal
	if err := m.store.loadRecord(m.ctx, m.tx, m.scope.Company, "journals", id, &j); err != nil {
		return j, err
	}
	if j.IsDeleted || j.Status == "reversed" || (m.scope.Branch != "" && j.BranchCode != m.scope.Branch) {
		return Journal{}, ErrNotFound
	}
	return j, nil
}

func (m *subledgerMutation) clearDraftLinks() error {
	var used bool
	if err := m.tx.QueryRowContext(m.ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_matches WHERE company=$1 AND journal_id=$2 AND owner_journal_id<>$2)`, m.scope.Company, m.journal.ID).Scan(&used); err != nil {
		return err
	}
	if used {
		return fmt.Errorf("รายการนี้มีการจับคู่จากใบสำคัญอื่นแล้ว")
	}
	for _, entry := range []struct{ table, column string }{{"gl_subledger_matches", "owner_journal_id"}, {"gl_subledger_settlements", "journal_id"}, {"gl_subledger_allocations", "journal_id"}, {"gl_subledger_bank_lines", "journal_id"}} {
		if _, err := m.tx.ExecContext(m.ctx, `DELETE FROM `+entry.table+` WHERE company=$1 AND `+entry.column+`=$2`, m.scope.Company, m.journal.ID); err != nil {
			return err
		}
	}
	keep := []string{}
	if !m.journal.IsDeleted && m.journal.Details != nil {
		for _, d := range m.journal.Details.Documents {
			keep = append(keep, d.ID)
		}
		for _, a := range m.journal.Details.Allocations {
			keep = append(keep, a.DocumentID)
		}
		for _, s := range m.journal.Details.Settlements {
			keep = append(keep, s.DebtDocumentID, s.PaymentDocumentID)
		}
	}
	// Remove only abandoned, unreferenced draft documents; shared documents and
	// any document with retained relation history remain authoritative evidence.
	_, err := m.tx.ExecContext(m.ctx, `DELETE FROM gl_subledger_documents d WHERE d.company=$1 AND d.created_journal_id=$2 AND NOT(d.id=ANY($3)) AND NOT EXISTS(SELECT 1 FROM gl_subledger_allocations a WHERE a.company=d.company AND a.document_id=d.id) AND NOT EXISTS(SELECT 1 FROM gl_subledger_settlements s WHERE s.company=d.company AND d.id IN(s.debt_id,s.payment_id))`, m.scope.Company, m.journal.ID, pq.Array(keep))
	return err
}
