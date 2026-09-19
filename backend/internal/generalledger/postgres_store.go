package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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
	if len(cmd.RequestID) < 16 || len(cmd.RequestID) > 80 || !validCode(cmd.RequestID) {
		return Result{}, fmt.Errorf("รหัสคำขอไม่ถูกต้อง กรุณาลองบันทึกอีกครั้ง")
	}
	if collectionName(cmd.Resource) == "" && cmd.Resource != "processes" {
		return Result{}, fmt.Errorf("ไม่รองรับรายการบัญชีนี้")
	}
	if !contains([]string{"create", "update", "delete", "post", "reverse", "lock", "unlock", "close", "year-end", "recalculate", "reprocess"}, cmd.Action) {
		return Result{}, fmt.Errorf("ไม่รองรับคำสั่งบัญชีนี้")
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
	err := tx.QueryRowContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind=$2 AND id=$3 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company, kind, id).Scan(&payload)
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
		return s.mutateFiscalYear(ctx, tx, scope, cmd, now)
	case "journals":
		return s.mutateJournal(ctx, tx, scope, cmd, now)
	case "processes":
		return s.mutateProcess(ctx, tx, scope, cmd, now)
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
		if err := acc.Validate(); err != nil {
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
		if err := next.Validate(); err != nil {
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

func (s *PostgresStore) mutateFiscalYear(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if cmd.Action == "create" {
		if cmd.FiscalYear == nil {
			return nil, fmt.Errorf("ไม่พบข้อมูลปีบัญชี")
		}
		fy := *cmd.FiscalYear
		if fy.Code == "" || fy.StartDate == "" || fy.EndDate == "" {
			return nil, fmt.Errorf("กรุณาระบุรหัสและช่วงวันที่ของปีบัญชี")
		}
		if fy.StartDate > fy.EndDate {
			return nil, fmt.Errorf("วันที่เริ่มต้นต้องไม่มากกว่าวันที่สิ้นสุด")
		}
		if fy.ProfitLossAccount == fy.RetainedEarningsAccount {
			return nil, fmt.Errorf("บัญชีกำไรขาดทุนและกำไรสะสมต้องเป็นคนละบัญชี")
		}
		if err := s.checkDuplicateCode(ctx, tx, scope.Company, "fiscal-years", fy.Code, ""); err != nil {
			return nil, err
		}
		fy.Identity = newIdentity(scope, now)
		data, err := json.Marshal(fy)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: "fiscal-years", ID: fy.ID, Code: fy.Code, Payload: string(data)}}, nil
	}

	var old FiscalYear
	if err := s.loadRecord(ctx, tx, scope.Company, "fiscal-years", cmd.ID, &old); err != nil {
		return nil, err
	}
	if err := checkVersion(cmd, old.Identity); err != nil {
		return nil, err
	}

	if cmd.Action == "update" {
		if cmd.FiscalYear == nil {
			return nil, fmt.Errorf("ไม่พบข้อมูลปีบัญชี")
		}
		next := *cmd.FiscalYear
		if next.Code != old.Code {
			return nil, fmt.Errorf("รหัสปีบัญชีแก้ไม่ได้")
		}
		if next.ProfitLossAccount == next.RetainedEarningsAccount {
			return nil, fmt.Errorf("บัญชีกำไรขาดทุนและกำไรสะสมต้องเป็นคนละบัญชี")
		}
		next.Identity = updateIdentity(scope, old.Identity, now)
		data, err := json.Marshal(next)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: "fiscal-years", ID: next.ID, Code: next.Code, Payload: string(data)}}, nil
	}

	return nil, fmt.Errorf("ไม่รองรับคำสั่งนี้สำหรับปีบัญชี")
}

func (s *PostgresStore) mutateMaster(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	kind := cmd.Resource
	if cmd.Action == "create" || cmd.Action == "lock" {
		if cmd.Master == nil {
			return nil, fmt.Errorf("ไม่พบข้อมูลหลัก")
		}
		m := *cmd.Master
		m.Kind = kind
		if m.Code == "" {
			return nil, fmt.Errorf("กรุณาระบุรหัส")
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
		if err := s.validateJournalLines(ctx, tx, scope, &j); err != nil {
			return nil, err
		}
		if err := s.checkDuplicateCode(ctx, tx, scope.Company, "journals", j.DocNo, ""); err != nil {
			return nil, err
		}
		j.Status = "draft"
		j.Identity = newIdentity(scope, now)
		data, err := json.Marshal(j)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: "journals", ID: j.ID, Code: j.DocNo, Payload: string(data)}}, nil
	}

	var old Journal
	if err := s.loadRecord(ctx, tx, scope.Company, "journals", cmd.ID, &old); err != nil {
		return nil, err
	}
	if err := checkVersion(cmd, old.Identity); err != nil {
		return nil, err
	}

	if cmd.Action == "delete" {
		if old.Status == "posted" {
			return nil, fmt.Errorf("ห้ามลบรายการที่ผ่านบัญชีแล้ว กรุณาสร้างใบกลับรายการแทน")
		}
		old.Identity = updateIdentity(scope, old.Identity, now)
		old.IsDeleted = true
		data, err := json.Marshal(old)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: "journals", ID: old.ID, Code: old.DocNo, Payload: string(data)}}, nil
	}

	if cmd.Action == "update" {
		if old.Status == "posted" {
			return nil, fmt.Errorf("ห้ามแก้ไขรายการที่ผ่านบัญชีแล้ว")
		}
		if cmd.Journal == nil {
			return nil, fmt.Errorf("ไม่พบข้อมูลเอกสารรายวัน")
		}
		next := *cmd.Journal
		if next.DocNo != old.DocNo {
			return nil, fmt.Errorf("เลขที่เอกสารแก้ไม่ได้")
		}
		if err := s.validateJournalLines(ctx, tx, scope, &next); err != nil {
			return nil, err
		}
		next.Status = old.Status
		next.Identity = updateIdentity(scope, old.Identity, now)
		data, err := json.Marshal(next)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: "journals", ID: next.ID, Code: next.DocNo, Payload: string(data)}}, nil
	}

	if cmd.Action == "post" {
		if old.Status == "posted" {
			return nil, fmt.Errorf("รายการนี้ผ่านบัญชีแล้ว")
		}
		old.Status = "posted"
		tNow := now
		old.PostedAt = &tNow
		old.PostedBy = scope.Actor
		old.Identity = updateIdentity(scope, old.Identity, now)
		data, err := json.Marshal(old)
		if err != nil {
			return nil, err
		}
		return []Change{{Kind: "journals", ID: old.ID, Code: old.DocNo, Payload: string(data)}}, nil
	}

	if cmd.Action == "reverse" {
		if old.Status != "posted" {
			return nil, fmt.Errorf("กลับรายการได้เฉพาะรายการที่ผ่านบัญชีแล้วเท่านั้น")
		}
		if cmd.DocNo == "" || cmd.Date == "" {
			return nil, fmt.Errorf("กรุณาระบุเลขที่และวันที่ของใบกลับรายการ")
		}
		if err := s.checkDuplicateCode(ctx, tx, scope.Company, "journals", cmd.DocNo, ""); err != nil {
			return nil, err
		}

		revLines := make([]Line, len(old.Lines))
		for i, l := range old.Lines {
			revLines[i] = Line{
				AccountCode:    l.AccountCode,
				AccountName:    l.AccountName,
				Description:    "กลับรายการ: " + l.Description,
				Debit:          l.Credit, // สลับ Dr/Cr
				Credit:         l.Debit,
				DepartmentCode: l.DepartmentCode,
				ProjectCode:    l.ProjectCode,
				CashFlow:       l.CashFlow,
			}
		}

		revDoc := Journal{
			DocNo:       cmd.DocNo,
			Date:        cmd.Date,
			BookCode:    old.BookCode,
			FiscalYear:  old.FiscalYear,
			Description: "กลับรายการของเอกสาร " + old.DocNo + ": " + cmd.Reason,
			Reference:   old.DocNo,
			Kind:        old.Kind,
			Status:      "posted",
			Lines:       revLines,
			ReversalOf:  old.DocNo,
			Reason:      cmd.Reason,
			Identity:    newIdentity(scope, now),
		}

		tNow := now
		revDoc.PostedAt = &tNow
		revDoc.PostedBy = scope.Actor

		revData, err := json.Marshal(revDoc)
		if err != nil {
			return nil, err
		}

		old.Status = "reversed"
		old.Reason = cmd.Reason
		old.Identity = updateIdentity(scope, old.Identity, now)
		oldData, err := json.Marshal(old)
		if err != nil {
			return nil, err
		}

		return []Change{
			{Kind: "journals", ID: old.ID, Code: old.DocNo, Payload: string(oldData)},
			{Kind: "journals", ID: revDoc.ID, Code: revDoc.DocNo, Payload: string(revData)},
		}, nil
	}

	return nil, fmt.Errorf("ไม่รองรับคำสั่งนี้สำหรับเอกสารรายวัน")
}

func (s *PostgresStore) validateJournalLines(ctx context.Context, tx *sql.Tx, scope Scope, j *Journal) error {
	if len(j.Lines) < 2 {
		return fmt.Errorf("รายการบัญชีต้องมีอย่างน้อย 2 บรรทัด")
	}

	// ตรวจสอบ Period Lock ในงวดที่คีย์
	var isLocked bool
	err := tx.QueryRowContext(ctx, `SELECT COALESCE((payload->>'locked')::boolean,false) FROM gl_records WHERE company=$1 AND kind='periods' AND payload->>'startdate'<=$2 AND payload->>'enddate'>=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) LIMIT 1`, scope.Company, j.Date).Scan(&isLocked)
	if err == nil && isLocked {
		return fmt.Errorf("งวดบัญชีของวันที่ %s ถูกล็อกแล้ว ไม่อนุญาตให้บันทึกหรือแก้ไข", j.Date)
	}

	totalDr := decimal.Zero
	totalCr := decimal.Zero

	for idx, l := range j.Lines {
		if l.AccountCode == "" {
			return fmt.Errorf("บรรทัดที่ %d: กรุณาระบุรหัสบัญชี", idx+1)
		}
		dr, _ := decimal.NewFromString(string(l.Debit))
		cr, _ := decimal.NewFromString(string(l.Credit))
		if dr.IsNegative() || cr.IsNegative() {
			return fmt.Errorf("บรรทัดที่ %d: จำนวนเงินต้องไม่ติดลบ", idx+1)
		}
		if dr.IsZero() && cr.IsZero() {
			return fmt.Errorf("บรรทัดที่ %d: ต้องระบุจำนวนเงินเดบิตหรือเครดิต", idx+1)
		}
		if !dr.IsZero() && !cr.IsZero() {
			return fmt.Errorf("บรรทัดที่ %d: ไม่สามารถระบุทั้งเดบิตและเครดิตในบรรทัดเดียวกันได้", idx+1)
		}
		totalDr = totalDr.Add(dr)
		totalCr = totalCr.Add(cr)
	}

	if !totalDr.Equal(totalCr) {
		return fmt.Errorf("ยอดเดบิต (%s) และเครดิต (%s) ไม่เท่ากัน ต่างกัน %s", totalDr.String(), totalCr.String(), totalDr.Sub(totalCr).Abs().String())
	}
	return nil
}

func (s *PostgresStore) mutateProcess(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if cmd.Action == "recalculate" {
		return []Change{}, nil // Rebuilds lines and balances directly
	}
	return []Change{}, nil
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
