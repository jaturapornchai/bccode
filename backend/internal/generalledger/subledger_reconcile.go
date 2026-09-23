package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

func (s *PostgresStore) mutateReconciliation(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	var current Journal
	if err := s.loadRecord(ctx, tx, scope.Company, "journals", cmd.ID, &current); err != nil {
		return nil, err
	}
	if current.IsDeleted || current.Status != "posted" || (scope.Branch != "" && scope.Branch != current.BranchCode) {
		return nil, ErrNotFound
	}
	if err := checkVersion(cmd, current.Identity); err != nil {
		return nil, err
	}
	if cmd.Journal == nil || cmd.Journal.Details == nil {
		return nil, fmt.Errorf("กรุณาระบุรายละเอียดกระทบยอด")
	}
	patch, err := cloneJournalDetails(cmd.Journal.Details)
	if err != nil {
		return nil, err
	}
	if len(patch.Partners)+len(patch.BankAccounts)+len(patch.Documents)+len(patch.BankLines) > 0 {
		return nil, fmt.Errorf("หลังผ่านบัญชีเพิ่มได้เฉพาะ Statement การตัดยอด การจับคู่ การถอน ภาษีหัก ณ ที่จ่าย และภาษีมูลค่าเพิ่ม")
	}
	if len(patch.StatementLines)+len(patch.Settlements)+len(patch.Matches)+len(patch.Withdrawals)+len(patch.Withholdings)+len(patch.Vats) == 0 {
		return nil, fmt.Errorf("ไม่มีรายละเอียดกระทบยอด")
	}
	if err = s.checkOpenDate(ctx, tx, scope, current.Date); err != nil {
		return nil, err
	}
	fiscal, err := s.openPGYear(ctx, tx, scope, current.FiscalYear)
	if err != nil {
		return nil, err
	}
	current.Details, err = cloneJournalDetails(current.Details)
	if err != nil {
		return nil, err
	}
	m := &subledgerMutation{ctx: ctx, tx: tx, store: s, scope: scope, journal: &current, now: now, scale: fiscal.Scale, affected: map[string]bool{}}
	if err = m.apply(patch, true); err != nil {
		return nil, err
	}
	if len(patch.Allocations) > 0 || len(m.reallocatedDocs) > 0 {
		if len(patch.Allocations) == 0 || len(m.reallocatedDocs) == 0 {
			return nil, fmt.Errorf("ถอนและจัดสรรใหม่ในคำขอเดียวกัน")
		}
		if err = m.completeDocuments(); err != nil {
			return nil, err
		}
	}
	// Preserve the original financial evidence and append only newly submitted evidence.
	current.Details.Allocations = mergeSubledgerRows(current.Details.Allocations, patch.Allocations, func(v SubledgerAllocation) string { return v.ID })
	current.Details.StatementLines = mergeSubledgerRows(current.Details.StatementLines, patch.StatementLines, func(v SubledgerStatementLine) string { return v.ID })
	current.Details.Settlements = mergeSubledgerRows(current.Details.Settlements, patch.Settlements, func(v SubledgerSettlement) string { return v.ID })
	current.Details.Matches = mergeSubledgerRows(current.Details.Matches, patch.Matches, func(v SubledgerMatch) string { return v.ID })
	// ภาษีหัก ณ ที่จ่าย/ภาษีมูลค่าเพิ่มเป็นรายละเอียดประกอบ ไม่ใช่ยอด GL — แก้ได้เสมอแม้ผ่านบัญชีแล้ว: แทนทั้งชุด และเก็บค่าเดิมไว้ใน audit
	if len(patch.Withholdings)+len(patch.Vats) > 0 {
		reason := strings.TrimSpace(cmd.Reason)
		if reason == "" || len([]rune(reason)) > 500 {
			return nil, fmt.Errorf("ระบุเหตุผลการแก้ฐานภาษีหลังผ่านบัญชี (ไม่เกิน 500 ตัวอักษร)")
		}
		if len(patch.Withholdings) > 0 {
			if err = m.audit("withholding_replace", map[string]any{"before": current.Details.Withholdings, "after": patch.Withholdings, "reason": reason}); err != nil {
				return nil, err
			}
			current.Details.Withholdings = patch.Withholdings
		}
		if len(patch.Vats) > 0 {
			if err = m.audit("vat_replace", map[string]any{"before": current.Details.Vats, "after": patch.Vats, "reason": reason}); err != nil {
				return nil, err
			}
			current.Details.Vats = patch.Vats
		}
	}
	current.Details.Withdrawals = mergeSubledgerRows(current.Details.Withdrawals, patch.Withdrawals, func(v SubledgerWithdrawal) string { return v.Kind + ":" + v.ID })
	current.Identity = updateIdentity(scope, current.Identity, now)
	raw, err := json.Marshal(current)
	if err != nil {
		return nil, err
	}
	related, err := m.relatedChanges()
	if err != nil {
		return nil, err
	}
	if err = m.audit("reconcile", patch); err != nil {
		return nil, err
	}
	return append([]Change{{Kind: "journals", ID: current.ID, Code: current.DocNo, Payload: string(raw)}}, related...), nil
}

func mergeSubledgerRows[T any](old, next []T, key func(T) string) []T {
	seen := map[string]bool{}
	for _, v := range old {
		seen[key(v)] = true
	}
	for _, v := range next {
		if !seen[key(v)] {
			old = append(old, v)
			seen[key(v)] = true
		}
	}
	return old
}

func (m *subledgerMutation) relatedChanges() ([]Change, error) {
	ids := make([]string, 0, len(m.affected))
	for id := range m.affected {
		if id != m.journal.ID {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	changes := make([]Change, 0, len(ids))
	for _, id := range ids {
		j, err := m.referencedJournal(id)
		if err != nil {
			return nil, err
		}
		j.Identity = updateIdentity(m.scope, j.Identity, m.now)
		raw, err := json.Marshal(j)
		if err != nil {
			return nil, err
		}
		changes = append(changes, Change{Kind: "journals", ID: j.ID, Code: j.DocNo, Payload: string(raw)})
	}
	return changes, nil
}

func (m *subledgerMutation) withdraw(w SubledgerWithdrawal) error {
	w.Reason = strings.TrimSpace(w.Reason)
	if !subledgerID(w.ID) || w.Reason == "" || len(w.Reason) > 500 {
		return fmt.Errorf("การถอนต้องระบุ ID และเหตุผลไม่เกิน 500 ตัวอักษร")
	}
	var table, owner string
	switch w.Kind {
	case "allocation":
		table = "gl_subledger_allocations"
		owner = "journal_id"
	case "settlement":
		table = "gl_subledger_settlements"
		owner = "journal_id"
	case "match":
		table = "gl_subledger_matches"
		owner = "owner_journal_id"
	default:
		return fmt.Errorf("ประเภทการถอนไม่ถูกต้อง")
	}
	var jid string
	var raw []byte
	var reversed sql.NullTime
	var reason sql.NullString
	err := m.tx.QueryRowContext(m.ctx, `SELECT `+owner+`,payload,reversed_at,reversal_reason FROM `+table+` WHERE company=$1 AND id=$2 FOR UPDATE`, m.scope.Company, w.ID).Scan(&jid, &raw, &reversed, &reason)
	if err != nil {
		return err
	}
	j, err := m.referencedJournal(jid)
	if err != nil {
		return err
	}
	if j.Status != "posted" && j.ID != m.journal.ID {
		return fmt.Errorf("ถอนรายละเอียดได้เฉพาะใบสำคัญที่ผ่านบัญชี")
	}
	if err = m.store.checkOpenDate(m.ctx, m.tx, m.scope, j.Date); err != nil {
		return err
	}
	if reversed.Valid {
		if reason.String == w.Reason {
			return nil
		}
		return fmt.Errorf("รายการนี้ถูกถอนแล้วด้วยเหตุผลอื่น")
	}
	m.affected[jid] = true
	switch w.Kind {
	case "allocation":
		if jid != m.journal.ID {
			return fmt.Errorf("withdraw allocation through its own journal")
		}
		var a SubledgerAllocation
		if err = json.Unmarshal(raw, &a); err != nil {
			return err
		}
		m.reallocatedDocs = append(m.reallocatedDocs, a.DocumentID)
		var used bool
		if err = m.tx.QueryRowContext(m.ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_settlements WHERE company=$1 AND (debt_id=$2 OR payment_id=$2) AND reversed_at IS NULL)`, m.scope.Company, a.DocumentID).Scan(&used); err != nil {
			return err
		}
		if used {
			return fmt.Errorf("ต้องถอนการตัดยอดของเอกสารก่อนถอนการจัดสรร")
		}
		if err = m.touchDocument(a.DocumentID); err != nil {
			return err
		}
	case "settlement":
		var s SubledgerSettlement
		if err = json.Unmarshal(raw, &s); err != nil {
			return err
		}
		if err = m.touchDocument(s.DebtDocumentID); err != nil {
			return err
		}
		if err = m.touchDocument(s.PaymentDocumentID); err != nil {
			return err
		}
	case "match":
		var v SubledgerMatch
		if err = json.Unmarshal(raw, &v); err != nil {
			return err
		}
		if _, err = m.referencedJournal(v.JournalID); err != nil {
			return err
		}
		m.affected[v.JournalID] = true
		var owner string
		if err = m.tx.QueryRowContext(m.ctx, `SELECT owner_journal_id FROM gl_subledger_statements WHERE company=$1 AND id=$2`, m.scope.Company, v.StatementLineID).Scan(&owner); err != nil {
			return err
		}
		imported, err := m.statementOwner(owner)
		if err != nil {
			return err
		}
		if !imported.IsDeleted && imported.Status != "reversed" {
			m.affected[owner] = true
		}
	}
	_, err = m.tx.ExecContext(m.ctx, `UPDATE `+table+` SET reversed_at=$3,reversed_by=$4,reversal_reason=$5 WHERE company=$1 AND id=$2 AND reversed_at IS NULL`, m.scope.Company, w.ID, m.now, m.scope.Actor, w.Reason)
	if err != nil {
		return err
	}
	return m.audit("withdraw", w)
}

func (m *subledgerMutation) reverseJournal() error {
	var used bool
	err := m.tx.QueryRowContext(m.ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_matches WHERE company=$1 AND (journal_id=$2 OR owner_journal_id=$2) AND reversed_at IS NULL) OR EXISTS(SELECT 1 FROM gl_subledger_settlements s WHERE s.company=$1 AND s.reversed_at IS NULL AND (s.journal_id=$2 OR EXISTS(SELECT 1 FROM gl_subledger_allocations a WHERE a.company=s.company AND a.journal_id=$2 AND a.reversed_at IS NULL AND a.document_id IN(s.debt_id,s.payment_id))))`, m.scope.Company, m.journal.ID).Scan(&used)
	if err != nil {
		return err
	}
	if used {
		return fmt.Errorf("ต้องถอนการตัดยอดหนี้และการจับคู่ธนาคารที่เกี่ยวข้องก่อนกลับรายการ")
	}
	_, err = m.tx.ExecContext(m.ctx, `UPDATE gl_subledger_allocations SET reversed_at=$3,reversed_by=$4,reversal_reason=$5,reversed_effective_date=$6::date WHERE company=$1 AND journal_id=$2 AND reversed_at IS NULL`, m.scope.Company, m.journal.ID, m.now, m.scope.Actor, m.journal.Reason, m.reversalDate)
	if err != nil {
		return err
	}
	return m.audit("reverse", m.journal.Details)
}
