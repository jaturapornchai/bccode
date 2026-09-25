package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// PostgresStore implements a 100% PostgreSQL-based General Ledger Store.
// It directly executes commands with ACID transactions, zero Kafka, and zero MongoDB.
type PostgresStore struct {
	pg *Postgres
}

func NewPostgresStore(pg *Postgres) *PostgresStore {
	return &PostgresStore{pg: pg}
}

func (s *PostgresStore) Projection() *Postgres {
	return s.pg
}

func (s *PostgresStore) Get(ctx context.Context, scope Scope, resource, id string) (json.RawMessage, error) {
	return s.pg.Get(ctx, scope, resource, id)
}

func (s *PostgresStore) List(ctx context.Context, scope Scope, resource, query string, page, limit int, filter ListFilter) (Page, error) {
	return s.pg.List(ctx, scope, resource, query, page, limit, filter)
}

func (s *PostgresStore) Report(ctx context.Context, scope Scope, name string, q ReportQuery) (Report, error) {
	return s.pg.Report(ctx, scope, name, q)
}

func (s *PostgresStore) Ready(ctx context.Context, scope Scope) error {
	return nil // Pure PostgreSQL is always synchronous and ready!
}

func (s *PostgresStore) JournalForAuthorization(ctx context.Context, scope Scope, id string) (Journal, error) {
	raw, err := s.pg.Get(ctx, scope, "journals", id)
	if err != nil {
		return Journal{}, err
	}
	var j Journal
	if err := json.Unmarshal(raw, &j); err != nil {
		return Journal{}, err
	}
	return j, nil
}

func (s *PostgresStore) Execute(ctx context.Context, scope Scope, cmd Command) (Result, error) {
	if scope.Holding == "" || scope.Company == "" || scope.Actor == "" {
		return Result{}, fmt.Errorf("กรุณาเลือกบริษัทและเข้าสู่ระบบก่อนใช้งานบัญชี")
	}
	// Codes are trimmed + NFC-normalised before the request hash, so a retry with a pasted
	// trailing space is the same request; NUL is refused before it reaches jsonb.
	normalizeCommandCodes(&cmd)
	if err := rejectNUL(cmd); err != nil {
		return Result{}, err
	}
	if cmd.Resource == "budgets" && cmd.Action == "spread" {
		periods, err := s.budgetSpreadPeriods(ctx, scope, cmd.Budget)
		if err != nil {
			return Result{}, err
		}
		return spreadBudget(cmd, periods)
	}
	if !validRequestID(cmd.RequestID) {
		return Result{}, fmt.Errorf("รหัสคำขอไม่ถูกต้อง กรุณาลองบันทึกอีกครั้ง")
	}
	if !supportedRecord(cmd.Resource) && cmd.Resource != "processes" {
		return Result{}, fmt.Errorf("ไม่รองรับรายการบัญชีนี้")
	}
	if !contains([]string{"reconcile", "review", "create", "update", "delete", "post", "reverse", "lock", "unlock", "close", "year-end", "recalculate", "reprocess"}, cmd.Action) {
		return Result{}, fmt.Errorf("ไม่รองรับคำสั่งบัญชีนี้")
	}
	if cmd.Action == "reconcile" && cmd.Resource != "journals" {
		return Result{}, fmt.Errorf("กระทบยอดได้เฉพาะใบสำคัญ")
	}
	if cmd.Action != "create" && cmd.ID == "" {
		return Result{}, fmt.Errorf("กรุณาเลือกรายการบัญชี")
	}

	db, err := s.pg.database(ctx, scope.Holding)
	if err != nil {
		return Result{}, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()

	if err = lockCompany(ctx, tx, scope.Company); err != nil {
		return Result{}, err
	}

	// 1. Check idempotency by RequestID
	rawCmd, err := json.Marshal(cmd)
	if err != nil {
		return Result{}, err
	}
	hash := digest(rawCmd)

	var savedEventPayload []byte
	err = tx.QueryRowContext(ctx, `SELECT payload FROM gl_events WHERE company=$1 AND payload->>'requestid'=$2`, scope.Company, cmd.RequestID).Scan(&savedEventPayload)
	if err == nil {
		var savedEvent Event
		if err := json.Unmarshal(savedEventPayload, &savedEvent); err == nil {
			if savedEvent.RequestHash != hash {
				return Result{}, fmt.Errorf("รหัสคำขอนี้ถูกใช้แล้วด้วยข้อมูลที่ต่างกัน")
			}
			res := Result{Sequence: savedEvent.Sequence}
			if cmd.Action == "review" {
				res.ID = cmd.ID
				res.Version = cmd.Version
			}
			if cmd.Resource == "processes" && (cmd.Action == "close" || cmd.Action == "year-end") {
				for _, change := range savedEvent.Changes {
					if change.Kind == "journals" {
						res.CreatedJournals++
					}
				}
			}
			if len(savedEvent.Changes) > 0 {
				res.ID = savedEvent.Changes[0].ID
				var idt Identity
				_ = json.Unmarshal([]byte(savedEvent.Changes[0].Payload), &idt)
				res.Version = idt.Version
			}
			return res, nil
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, err
	}

	if result, found, err := s.existingSourceJournal(ctx, tx, scope, cmd); err != nil {
		return Result{}, err
	} else if found {
		return result, nil
	}

	now := time.Now().UTC()

	// 2. Mutate changes in PostgreSQL
	changes, err := s.applyMutation(ctx, tx, scope, cmd, now)
	if err != nil {
		return Result{}, err
	}

	// 3. Get next sequence
	if _, err = tx.ExecContext(ctx, `INSERT INTO gl_projection_state(company) VALUES ($1) ON CONFLICT DO NOTHING`, scope.Company); err != nil {
		return Result{}, err
	}
	var sequence int64
	if err = tx.QueryRowContext(ctx, `SELECT sequence FROM gl_projection_state WHERE company=$1 FOR UPDATE`, scope.Company).Scan(&sequence); err != nil {
		return Result{}, err
	}
	sequence++

	eventID := fmt.Sprintf("%s-%012d", scope.Company, sequence)
	event := Event{
		ID:           eventID,
		HoldingCode:  scope.Holding,
		BusinessCode: scope.Company,
		Sequence:     sequence,
		RequestID:    cmd.RequestID,
		RequestHash:  hash,
		Action:       cmd.Resource + ":" + cmd.Action,
		Actor:        scope.Actor,
		Reason:       cmd.Reason,
		OccurredAt:   now,
		Changes:      changes,
	}

	// 4. Project event into gl_records and gl_lines directly
	if err = applyChanges(ctx, tx, event); err != nil {
		return Result{}, err
	}

	if err = s.recordSourceJournal(ctx, tx, scope, cmd, changes, sequence, now); err != nil {
		return Result{}, err
	}

	eventHash, data, err := eventDigest(event)
	if err != nil {
		return Result{}, err
	}

	if _, err = tx.ExecContext(ctx, `INSERT INTO gl_events(company,id,sequence,event_hash,occurred_at,payload) VALUES ($1,$2,$3,$4,$5,$6::jsonb)`,
		scope.Company, event.ID, event.Sequence, eventHash, event.OccurredAt, string(data)); err != nil {
		return Result{}, err
	}

	if _, err = tx.ExecContext(ctx, `UPDATE gl_projection_state SET sequence=$2 WHERE company=$1`, scope.Company, event.Sequence); err != nil {
		return Result{}, err
	}

	if err = tx.Commit(); err != nil {
		return Result{}, err
	}

	res := Result{Sequence: event.Sequence, ProjectionPending: false}
	if cmd.Action == "review" {
		res.ID = cmd.ID
		res.Version = cmd.Version
	}
	if len(changes) > 0 {
		res.ID = changes[0].ID
		var idt Identity
		_ = json.Unmarshal([]byte(changes[0].Payload), &idt)
		res.Version = idt.Version
	}
	if cmd.Resource == "processes" && (cmd.Action == "close" || cmd.Action == "year-end") {
		for _, ch := range changes {
			if ch.Kind == "journals" {
				res.CreatedJournals++
			}
		}
	}
	return res, nil
}

func (s *PostgresStore) loadRecord(ctx context.Context, tx *sql.Tx, company, kind, id string, target any) error {
	var payload []byte
	err := tx.QueryRowContext(ctx, `SELECT payload || jsonb_build_object('id', id, 'version', version) FROM gl_records WHERE company=$1 AND kind=$2 AND id=$3 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company, kind, id).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, target)
}

func (s *PostgresStore) checkDuplicateCode(ctx context.Context, tx *sql.Tx, company, kind, code, excludeID string) error {
	var existsID string
	err := tx.QueryRowContext(ctx, `SELECT id FROM gl_records WHERE company=$1 AND kind=$2 AND code=$3 AND id<>$4 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) LIMIT 1`, company, kind, code, excludeID).Scan(&existsID)
	if err == nil {
		if kind == "accounts" {
			return userError(CodeDuplicateCode, "รหัสบัญชีนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น")
		}
		if kind == "journals" {
			return userError(CodeDuplicateCode, "เลขที่เอกสารนี้ถูกใช้แล้ว กรุณาใช้เลขที่อื่น")
		}
		return userError(CodeDuplicateCode, "รหัสนี้ถูกใช้แล้ว กรุณาใช้รหัสอื่น")
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func (s *PostgresStore) applyMutation(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	switch cmd.Resource {
	case "accounts":
		return s.mutateAccount(ctx, tx, scope, cmd, now)
	case "fiscal-years":
		changes, err := s.mutatePGFiscalYear(ctx, tx, scope, cmd, now)
		if err != nil || cmd.Action != "create" {
			return changes, err
		}
		// A company's first fiscal year is its ledger setup: the standard journal books come with it.
		books, err := defaultJournalBookChanges(ctx, tx, scope, now)
		if err != nil {
			return nil, err
		}
		return append(changes, books...), nil
	case "journals":
		if cmd.Action == "reconcile" {
			return s.mutateReconciliation(ctx, tx, scope, cmd, now)
		}
		if cmd.Action == "review" {
			return s.mutateReview(ctx, tx, scope, cmd, now)
		}
		return s.mutateJournal(ctx, tx, scope, cmd, now)
	case "processes":
		return s.mutateProcess(ctx, tx, scope, cmd, now)
	case "budgets":
		return s.mutateBudget(ctx, tx, scope, cmd, now)
	default:
		return s.mutateMaster(ctx, tx, scope, cmd, now)
	}
}

func (s *PostgresStore) mutateAccount(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if cmd.Action == "create" {
		if cmd.Account == nil {
			return nil, fmt.Errorf("ไม่พบข้อมูลผังบัญชี")
		}
		acc := *cmd.Account
		if err := s.validatePGAccount(ctx, tx, scope, &acc, nil); err != nil {
			return nil, err
		}
		if err := s.checkDuplicateCode(ctx, tx, scope.Company, "accounts", acc.AccountCode, ""); err != nil {
			return nil, err
		}
		acc.Identity = newIdentity(scope, now)
		data, err := json.Marshal(acc)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: "accounts", ID: acc.ID, Code: acc.AccountCode, Payload: string(data)}}, nil
	}

	var old Account
	if err := s.loadRecord(ctx, tx, scope.Company, "accounts", cmd.ID, &old); err != nil {
		return nil, err
	}
	if err := checkVersion(cmd, old.Identity); err != nil {
		return nil, err
	}

	if cmd.Action == "delete" {
		if err := s.deletePGAccountGuard(ctx, tx, scope, old.AccountCode); err != nil {
			return nil, err
		}
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM gl_lines WHERE company=$1 AND account_code=$2`, scope.Company, old.AccountCode).Scan(&count); err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, userError(CodeReferenced, "บัญชีนี้มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบผังบัญชีเด็ดขาด กรุณาปิดใช้งานแทนการลบ")
		}
		old.Identity = updateIdentity(scope, old.Identity, now)
		old.IsDeleted = true
		data, err := json.Marshal(old)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: "accounts", ID: old.ID, Code: old.AccountCode, Payload: string(data)}}, nil
	}

	if cmd.Action == "update" {
		if cmd.Account == nil {
			return nil, fmt.Errorf("ไม่พบข้อมูลผังบัญชี")
		}
		next := *cmd.Account
		if next.AccountCode != old.AccountCode {
			return nil, userError(CodeImmutableCode, "รหัสบัญชีแก้ไม่ได้ กรุณาสร้างบัญชีใหม่")
		}
		if err := s.validatePGAccount(ctx, tx, scope, &next, &old); err != nil {
			return nil, err
		}
		next.Identity = updateIdentity(scope, old.Identity, now)
		data, err := json.Marshal(next)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: "accounts", ID: next.ID, Code: next.AccountCode, Payload: string(data)}}, nil
	}

	return nil, fmt.Errorf("ไม่รองรับคำสั่งนี้สำหรับผังบัญชี")
}

func (s *PostgresStore) mutateMaster(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	kind := cmd.Resource
	if kind == "periods" {
		return s.mutatePeriod(ctx, tx, scope, cmd, now)
	}
	if cmd.Action == "create" || cmd.Action == "lock" {
		if cmd.Master == nil {
			return nil, fmt.Errorf("ไม่พบข้อมูลหลัก")
		}
		m := *cmd.Master
		m.Kind = kind
		if err := validateMasterCode(&m, false); err != nil {
			return nil, err
		}
		if cmd.Action == "lock" {
			m.Locked = true
		}
		if err := s.checkDuplicateCode(ctx, tx, scope.Company, kind, m.Code, ""); err != nil {
			return nil, err
		}
		m.Identity = newIdentity(scope, now)
		data, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: kind, ID: m.ID, Code: m.Code, Payload: string(data)}}, nil
	}

	var old Master
	if err := s.loadRecord(ctx, tx, scope.Company, kind, cmd.ID, &old); err != nil {
		return nil, err
	}
	if err := checkVersion(cmd, old.Identity); err != nil {
		return nil, err
	}

	if cmd.Action == "delete" {
		if kind == "journal-books" {
			if err := guardJournalBookDelete(ctx, tx, scope.Company, old); err != nil {
				return nil, err
			}
		}
		old.Identity = updateIdentity(scope, old.Identity, now)
		old.IsDeleted = true
		data, err := json.Marshal(old)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: kind, ID: old.ID, Code: old.Code, Payload: string(data)}}, nil
	}

	if cmd.Action == "update" || cmd.Action == "unlock" {
		if cmd.Master == nil {
			return nil, fmt.Errorf("ไม่พบข้อมูลหลัก")
		}
		next := *cmd.Master
		next.Kind = kind
		if cmd.Action == "unlock" {
			next.Locked = false
		}
		if err := validateMasterCode(&next, old.BookType == 0); err != nil {
			return nil, err
		}
		if kind == "journal-books" {
			if err := guardJournalBookUpdate(ctx, tx, scope.Company, old, next); err != nil {
				return nil, err
			}
		}
		if next.Code != old.Code {
			if err := s.checkDuplicateCode(ctx, tx, scope.Company, kind, next.Code, old.ID); err != nil {
				return nil, err
			}
		}
		next.Identity = updateIdentity(scope, old.Identity, now)
		data, err := json.Marshal(next)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: kind, ID: next.ID, Code: next.Code, Payload: string(data)}}, nil
	}

	return nil, fmt.Errorf("ไม่รองรับคำสั่งนี้สำหรับข้อมูลหลัก")
}

func (s *PostgresStore) mutateJournal(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if cmd.Action == "create" {
		if cmd.Journal == nil {
			return nil, fmt.Errorf("ไม่พบข้อมูลเอกสารรายวัน")
		}
		j := *cmd.Journal
		if j.PostedAt != nil || j.PostedBy != "" || j.ReversalOf != "" || j.IsDeleted || (j.Status != "" && j.Status != "draft") {
			return nil, fmt.Errorf("สถานะผ่านรายการและข้อมูลกลับรายการกำหนดโดยระบบเท่านั้น")
		}
		j.Lines = append([]Line(nil), j.Lines...)
		// The new-journal form starts with no branch; it belongs to the branch the session works in.
		// Without this a branch-scoped session rejected every new journal as "not found" (UAT 2026-09-24).
		if strings.TrimSpace(j.BranchCode) == "" {
			j.BranchCode = scope.Branch
		}
		if err := s.validateJournalLines(ctx, tx, scope, &j, true); err != nil {
			return nil, err
		}
		if err := checkJournalTextLimits(j, nil); err != nil {
			return nil, err
		}
		if err := s.checkDuplicateCode(ctx, tx, scope.Company, "journals", j.DocNo, ""); err != nil {
			return nil, err
		}
		j.Status = "draft"
		j.Identity = newIdentity(scope, now)
		related, err := s.syncJournalDetails(ctx, tx, scope, &j, nil, "create", now)
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(j)
		if err != nil {
			return nil, err
		}
		return append([]Change{{Kind: "journals", ID: j.ID, Code: j.DocNo, Payload: string(data)}}, related...), nil
	}

	var old Journal
	if err := s.loadRecord(ctx, tx, scope.Company, "journals", cmd.ID, &old); err != nil {
		return nil, err
	}
	if scope.Branch != "" && old.BranchCode != scope.Branch {
		return nil, ErrNotFound
	}
	if err := checkVersion(cmd, old.Identity); err != nil {
		return nil, err
	}
	previous := old

	if cmd.Action == "delete" {
		if err := s.protectGeneratedOpening(ctx, tx, scope, old); err != nil {
			return nil, err
		}
		if _, err := s.openPGYear(ctx, tx, scope, old.FiscalYear); err != nil {
			return nil, err
		}
		if err := s.checkOpenDate(ctx, tx, scope, old.Date); err != nil {
			return nil, err
		}
		if old.Status != "draft" {
			return nil, fmt.Errorf("ห้ามลบรายการที่ผ่านบัญชีแล้ว กรุณาสร้างใบกลับรายการแทน")
		}
		old.Identity = updateIdentity(scope, old.Identity, now)
		old.IsDeleted = true
		related, err := s.syncJournalDetails(ctx, tx, scope, &old, &previous, "delete", now)
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(old)
		if err != nil {
			return nil, err
		}
		return append([]Change{{Kind: "journals", ID: old.ID, Code: old.DocNo, Payload: string(data)}}, related...), nil
	}

	if cmd.Action == "update" {
		if err := s.protectGeneratedOpening(ctx, tx, scope, old); err != nil {
			return nil, err
		}
		if _, err := s.openPGYear(ctx, tx, scope, old.FiscalYear); err != nil {
			return nil, err
		}
		if err := s.checkOpenDate(ctx, tx, scope, old.Date); err != nil {
			return nil, err
		}
		if old.Status != "draft" {
			return nil, fmt.Errorf("ห้ามแก้ไขรายการที่ผ่านบัญชีแล้ว")
		}
		if cmd.Journal == nil {
			return nil, fmt.Errorf("ไม่พบข้อมูลเอกสารรายวัน")
		}
		next := *cmd.Journal
		next.Lines = append([]Line(nil), next.Lines...)
		// Same default as create: a form whose local state still has branch "" must save into the
		// session branch instead of being refused as "not found"; an explicit other branch is still refused below.
		if strings.TrimSpace(next.BranchCode) == "" {
			next.BranchCode = scope.Branch
		}
		if err := preserveJournalSource(old, next); err != nil {
			return nil, err
		}
		switch {
		case next.DocNo != old.DocNo:
			return nil, errJournalDocNoImmutable
		case next.BookCode != old.BookCode:
			return nil, errJournalBookImmutable
		case next.Kind != old.Kind:
			return nil, errJournalKindImmutable
		}
		if err := s.validateJournalLines(ctx, tx, scope, &next, false); err != nil {
			return nil, err
		}
		if err := checkJournalTextLimits(next, &old); err != nil {
			return nil, err
		}
		next.Status = old.Status
		next.PostedAt = old.PostedAt
		next.PostedBy = old.PostedBy
		next.ReversalOf = old.ReversalOf
		next.Identity = updateIdentity(scope, old.Identity, now)
		related, err := s.syncJournalDetails(ctx, tx, scope, &next, &previous, "update", now)
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(next)
		if err != nil {
			return nil, err
		}
		return append([]Change{{Kind: "journals", ID: next.ID, Code: next.DocNo, Payload: string(data)}}, related...), nil
	}

	if cmd.Action == "post" {
		if old.Status != "draft" {
			return nil, fmt.Errorf("รายการนี้ผ่านบัญชีแล้ว")
		}
		if err := s.validateJournalLines(ctx, tx, scope, &old, false); err != nil {
			return nil, err
		}
		old.Status = "posted"
		tNow := now
		old.PostedAt = &tNow
		old.PostedBy = scope.Actor
		old.Identity = updateIdentity(scope, old.Identity, now)
		related, err := s.syncJournalDetails(ctx, tx, scope, &old, &previous, "post", now)
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(old)
		if err != nil {
			return nil, err
		}
		return append([]Change{{Kind: "journals", ID: old.ID, Code: old.DocNo, Payload: string(data)}}, related...), nil
	}

	if cmd.Action == "reverse" {
		// แยกข้อความตามสาเหตุ — เดิมรวม 5 สาเหตุเป็นข้อความเดียว ผู้ใช้ที่กรอกครบแล้วไม่รู้ว่าต้องแก้อะไร (UAT S08/S09/V14 2026-09-24)
		switch {
		case old.Status != "posted":
			return nil, userError("reverse_requires_posted", "กลับรายการได้เฉพาะรายการที่ผ่านบัญชีแล้วเท่านั้น")
		case old.Kind == "reversal" || old.Kind == "closing":
			return nil, userError("reverse_kind_not_allowed", "ใบกลับรายการและใบปิดบัญชีไม่สามารถกลับรายการซ้ำได้")
		case cmd.DocNo == "":
			return nil, fieldError("reverse_doc_no_invalid", "docno", "กรุณาระบุเลขที่ใบกลับรายการให้ถูกต้อง")
		case checkCode(cmd.DocNo, "docno", "เลขที่ใบกลับรายการ", DocNoMaxRunes) != nil:
			return nil, checkCode(cmd.DocNo, "docno", "เลขที่ใบกลับรายการ", DocNoMaxRunes)
		case !validDate(cmd.Date):
			return nil, fieldError("reverse_date_invalid", "date", "กรุณาระบุวันที่ใบกลับรายการให้ถูกต้อง")
		case cmd.Date < old.Date:
			return nil, fieldError("reverse_date_before_original", "date", "วันที่กลับรายการต้องไม่ก่อนวันที่ของรายการเดิม")
		case strings.TrimSpace(cmd.Reason) == "":
			return nil, fieldError("reverse_reason_required", "reason", "กรุณาระบุเหตุผลการกลับรายการ")
		case utf8.RuneCountInString(strings.TrimSpace(cmd.Reason)) > ReversalReasonMaxRunes:
			return nil, fieldError("reverse_reason_too_long", "reason", fmt.Sprintf("เหตุผลการกลับรายการต้องไม่เกิน %d ตัวอักษร — กรุณาย่อให้สั้นลง", ReversalReasonMaxRunes))
		}
		if err := s.checkDuplicateCode(ctx, tx, scope.Company, "journals", cmd.DocNo, ""); err != nil {
			return nil, err
		}

		revLines := make([]Line, len(old.Lines))
		// ข้อความที่ระบบต่อท้ายต้นฉบับอาจยาวเกินคอลัมน์ — ตัดให้พอดี ไม่ให้การกลับรายการติดความยาวของใบเดิม
		for i, l := range old.Lines {
			revLines[i] = Line{
				AccountCode:    l.AccountCode,
				AccountName:    l.AccountName,
				Description:    truncateRunes("กลับรายการ: "+l.Description, JournalLineDescriptionMaxRunes),
				Debit:          l.Credit, // สลับ Dr/Cr
				Credit:         l.Debit,
				DepartmentCode: l.DepartmentCode,
				ProjectCode:    l.ProjectCode,
				CashFlow:       l.CashFlow,
			}
		}

		year, err := s.pgYearAt(ctx, tx, scope, cmd.Date)
		if err != nil {
			return nil, err
		}
		revDoc := Journal{
			DocNo:       cmd.DocNo,
			Date:        cmd.Date,
			BookCode:    old.BookCode,
			FiscalYear:  year.Code,
			BranchCode:  old.BranchCode,
			Description: truncateRunes("กลับรายการของเอกสาร "+old.DocNo+": "+cmd.Reason, JournalDescriptionMaxRunes),
			Reference:   old.DocNo,
			Kind:        "reversal",
			Status:      "posted",
			Lines:       revLines,
			ReversalOf:  old.ID,
			Reason:      cmd.Reason,
			Identity:    newIdentity(scope, now),
		}

		tNow := now
		revDoc.PostedAt = &tNow
		revDoc.PostedBy = scope.Actor

		// The reversal stays in the original's book even if that book was deactivated since.
		if err := s.validateJournalLines(ctx, tx, scope, &revDoc, false); err != nil {
			return nil, err
		}
		revData, err := json.Marshal(revDoc)
		if err != nil {
			return nil, err
		}

		old.Status = "reversed"
		old.Reason = cmd.Reason
		old.Identity = updateIdentity(scope, old.Identity, now)
		related, err := s.syncJournalDetails(ctx, tx, scope, &old, &previous, "reverse", now, cmd.Date)
		if err != nil {
			return nil, err
		}
		oldData, err := json.Marshal(old)
		if err != nil {
			return nil, err
		}

		return append([]Change{
			{Kind: "journals", ID: old.ID, Code: old.DocNo, Payload: string(oldData)},
			{Kind: "journals", ID: revDoc.ID, Code: revDoc.DocNo, Payload: string(revData)},
		}, related...), nil
	}

	return nil, fmt.Errorf("ไม่รองรับคำสั่งนี้สำหรับเอกสารรายวัน")
}

// validateJournalLines checks a voucher before it is stored; requireActiveBook is true for new
// vouchers only (an existing voucher stays editable/postable after its book is deactivated).
func (s *PostgresStore) validateJournalLines(ctx context.Context, tx *sql.Tx, scope Scope, j *Journal, requireActiveBook bool) error {
	if err := validateJournalSource(*j); err != nil {
		return err
	}
	if scope.Branch != "" && j.BranchCode != scope.Branch {
		return ErrNotFound
	}
	year, err := s.pgYear(ctx, tx, scope, j.FiscalYear)
	if err != nil {
		return err
	}
	accounts, err := loadLineAccounts(ctx, tx, scope.Company, j.Lines)
	if err != nil {
		return err
	}
	if err = j.Validate(year, accounts); err != nil {
		return err
	}
	if err = checkJournalBook(ctx, tx, scope.Company, *j, requireActiveBook); err != nil {
		return err
	}
	if j.Kind == "opening" && j.Date != year.StartDate {
		return fmt.Errorf("ยอดยกมาต้องลงวันที่เริ่มปีบัญชี")
	}
	if err = s.checkOpenDate(ctx, tx, scope, j.Date); err != nil {
		return err
	}
	for i := range j.Lines {
		j.Lines[i].AccountName = accounts[j.Lines[i].AccountCode].ThaiName()
	}
	return nil
}

func newIdentity(scope Scope, now time.Time) Identity {
	return Identity{
		ID:           uuid.NewString(),
		HoldingCode:  scope.Holding,
		BusinessCode: scope.Company,
		Version:      1,
		CreatedAt:    now,
		CreatedBy:    scope.Actor,
		UpdatedAt:    now,
		UpdatedBy:    scope.Actor,
		IsDeleted:    false,
	}
}

func updateIdentity(scope Scope, old Identity, now time.Time) Identity {
	old.Version++
	old.UpdatedAt = now
	old.UpdatedBy = scope.Actor
	return old
}
