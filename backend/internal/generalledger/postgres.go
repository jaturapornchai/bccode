package generalledger

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

//go:embed schema.sql
var postgresSchema string

type Postgres struct {
	resolve func(string) (*sql.DB, error)
}

func NewPostgres(resolve func(string) (*sql.DB, error)) *Postgres {
	return &Postgres{resolve: resolve}
}

var _ Projection = (*Postgres)(nil)

func (p *Postgres) database(ctx context.Context, holding string) (*sql.DB, error) {
	if strings.TrimSpace(holding) == "" {
		return nil, fmt.Errorf("กรุณาระบุกลุ่มบริษัท")
	}
	db, err := p.resolve(holding)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, fmt.Errorf("ฐานประมวลผลยังไม่พร้อม")
	}
	if err := EnsureSchema(ctx, db); err != nil {
		return nil, err
	}
	return db, nil
}

var (
	schemaMu    sync.Mutex
	schemaReady = map[*sql.DB]bool{}
)

// EnsureSchema สร้างตาราง GL (schema.sql + subledger.sql) ในฐานของ holding ครั้งแรกที่ใช้ — ใช้ร่วมกับรายงานที่อ่าน
// gl_* ตรง (เช่น ภาษีหัก ณ ที่จ่าย) เพื่อให้กลุ่มกิจการใหม่เห็นรายงานว่างแทนที่จะ error เพราะยังไม่มีตาราง
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	schemaMu.Lock()
	defer schemaMu.Unlock()
	if schemaReady[db] {
		return nil
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('bc_general_ledger_schema'))`); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, postgresSchema); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, subledgerSchema); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	schemaReady[db] = true
	return nil
}

func lockCompany(ctx context.Context, tx *sql.Tx, company string) error {
	if strings.TrimSpace(company) == "" {
		return fmt.Errorf("กรุณาระบุบริษัท")
	}
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('bc_general_ledger_company'),hashtext($1))`, company)
	return err
}

func eventDigest(event Event) (string, []byte, error) {
	event.Delivered = false
	event.OccurredAt = event.OccurredAt.UTC().Truncate(time.Millisecond)
	if len(event.Changes) == 0 {
		event.Changes = []Change{}
	}
	data, err := json.Marshal(event)
	if err != nil {
		return "", nil, err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), data, nil
}

func (p *Postgres) Project(ctx context.Context, event Event) error {
	if event.ID == "" || event.BusinessCode == "" || event.Sequence < 1 || (len(event.Changes) == 0 && event.Action != "processes:recalculate") {
		return fmt.Errorf("เหตุการณ์บัญชีไม่ครบถ้วน")
	}
	hash, data, err := eventDigest(event)
	if err != nil {
		return err
	}
	db, err := p.database(ctx, event.HoldingCode)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = lockCompany(ctx, tx, event.BusinessCode); err != nil {
		return err
	}
	var savedHash string
	err = tx.QueryRowContext(ctx, `SELECT event_hash FROM gl_events WHERE company=$1 AND id=$2`, event.BusinessCode, event.ID).Scan(&savedHash)
	if err == nil {
		if savedHash != hash {
			return fmt.Errorf("เหตุการณ์ซ้ำมีข้อมูลไม่ตรงกัน")
		}
		return tx.Commit()
	}
	if err != sql.ErrNoRows {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO gl_projection_state(company) VALUES ($1) ON CONFLICT DO NOTHING`, event.BusinessCode); err != nil {
		return err
	}
	var sequence int64
	if err = tx.QueryRowContext(ctx, `SELECT sequence FROM gl_projection_state WHERE company=$1 FOR UPDATE`, event.BusinessCode).Scan(&sequence); err != nil {
		return err
	}
	if event.Sequence != sequence+1 {
		return fmt.Errorf("ลำดับบัญชียังไม่ต่อเนื่อง: ต้องการ %d ได้รับ %d", sequence+1, event.Sequence)
	}
	if err = applyChanges(ctx, tx, event); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO gl_events(company,id,sequence,event_hash,occurred_at,payload) VALUES ($1,$2,$3,$4,$5,$6::jsonb)`, event.BusinessCode, event.ID, event.Sequence, hash, event.OccurredAt, string(data)); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE gl_projection_state SET sequence=$2 WHERE company=$1`, event.BusinessCode, event.Sequence); err != nil {
		return err
	}
	return tx.Commit()
}

func supportedRecord(kind string) bool {
	return kind == "accounts" || kind == "fiscal-years" || kind == "journals" || MasterCollections[kind] != ""
}

func applyChanges(ctx context.Context, tx *sql.Tx, event Event) error {
	// Store every snapshot first: a command can introduce masters and a journal
	// atomically without depending on the order of Change entries.
	for _, change := range event.Changes {
		if !supportedRecord(change.Kind) || change.ID == "" {
			return fmt.Errorf("ชนิดข้อมูลบัญชีไม่ถูกต้อง")
		}
		var identity Identity
		if err := json.Unmarshal([]byte(change.Payload), &identity); err != nil {
			return fmt.Errorf("ข้อมูลเหตุการณ์บัญชีไม่ถูกต้อง: %w", err)
		}
		if identity.ID != change.ID || identity.HoldingCode != event.HoldingCode || identity.BusinessCode != event.BusinessCode || identity.Version < 1 {
			return fmt.Errorf("ขอบเขตข้อมูลเหตุการณ์บัญชีไม่ตรงกัน")
		}
		result, err := tx.ExecContext(ctx, `INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES($1,$2,$3,$4,$5,$6::jsonb)
          ON CONFLICT(company,kind,id) DO UPDATE SET code=EXCLUDED.code,version=EXCLUDED.version,payload=EXCLUDED.payload WHERE gl_records.version<EXCLUDED.version`, event.BusinessCode, change.Kind, change.ID, change.Code, identity.Version, change.Payload)
		if err != nil {
			return err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if n != 1 {
			return fmt.Errorf("รุ่นข้อมูลบัญชีเก่าหรือซ้ำ")
		}
	}
	for _, change := range event.Changes {
		if change.Kind != "journals" {
			continue
		}
		var journal Journal
		if err := json.Unmarshal([]byte(change.Payload), &journal); err != nil {
			return err
		}
		if err := projectJournal(ctx, tx, event.BusinessCode, journal); err != nil {
			return err
		}
	}
	return nil
}

func projectJournal(ctx context.Context, tx *sql.Tx, company string, journal Journal) error {
	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM gl_lines WHERE company=$1 AND journal_id=$2`, company, journal.ID).Scan(&existing); err != nil {
		return err
	}
	contributes := journal.Status == "posted" || journal.Status == "reversed"
	if !contributes || journal.IsDeleted {
		if existing != 0 {
			return fmt.Errorf("ห้ามลบยอดบัญชีที่ผ่านรายการแล้ว ให้สร้างรายการกลับบัญชี")
		}
		return nil
	}
	if existing != 0 {
		// A reversal changes the original's status only. Keep its original
		// lines and verify their accounting identity before accepting it.
		return verifyProjectedJournal(ctx, tx, company, journal, existing)
	}
	var fiscalData []byte
	if err := tx.QueryRowContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND code=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company, journal.FiscalYear).Scan(&fiscalData); err != nil {
		return fmt.Errorf("ไม่พบปีบัญชีของรายการ: %w", err)
	}
	var fiscal FiscalYear
	if err := json.Unmarshal(fiscalData, &fiscal); err != nil {
		return err
	}
	maxLines := 500
	if journal.Kind == "closing" || journal.Kind == "opening" {
		maxLines = ProcessBalanceRowLimit
	}
	if !validDate(journal.Date) || journal.Date < fiscal.StartDate || journal.Date > fiscal.EndDate || len(journal.Lines) < 2 || len(journal.Lines) > maxLines {
		return fmt.Errorf("ขอบเขตปีและสกุลเงินบัญชีไม่ถูกต้อง")
	}
	accounts, err := loadLineAccounts(ctx, tx, company, journal.Lines)
	if err != nil {
		return err
	}
	projected := make([]projectedLine, 0, len(journal.Lines))
	debit, credit := Amount("0").Decimal(), Amount("0").Decimal()
	for index, line := range journal.Lines {
		if err := line.Debit.ValidateScale(fiscal.Scale); err != nil {
			return err
		}
		if err := line.Credit.ValidateScale(fiscal.Scale); err != nil {
			return err
		}
		d, c := line.Debit.Decimal(), line.Credit.Decimal()
		if d.IsNegative() || c.IsNegative() || d.IsZero() == c.IsZero() {
			return fmt.Errorf("เดบิตเครดิตของบรรทัดบัญชีไม่ถูกต้อง")
		}
		debit = debit.Add(d)
		credit = credit.Add(c)
		account, ok := accounts[line.AccountCode]
		if !ok {
			return fmt.Errorf("ไม่พบบัญชี %s", line.AccountCode)
		}
		if !account.AllowPosting {
			return fmt.Errorf("บัญชี %s ไม่อนุญาตลงรายการ", line.AccountCode)
		}
		name := line.AccountName
		if name == "" {
			name = account.ThaiName()
		}
		projected = append(projected, projectedLine{Company: company, JournalID: journal.ID, LineNo: index + 1, DocNo: journal.DocNo, Date: journal.Date, FiscalYear: journal.FiscalYear, BookCode: journal.BookCode, BranchCode: journal.BranchCode, DepartmentCode: line.DepartmentCode, ProjectCode: line.ProjectCode, Kind: journal.Kind, Currency: projectionCurrency, Scale: fiscal.Scale, AccountCode: line.AccountCode, AccountName: name, AccountType: account.AccountType, NormalBalance: account.NormalBalance, IsCash: account.IsCash, Description: line.Description, CashFlow: line.CashFlow, Debit: d.String(), Credit: c.String()})
	}
	if !debit.Equal(credit) {
		return fmt.Errorf("ยอดเดบิตและเครดิตไม่เท่ากัน")
	}
	return insertProjectedLines(ctx, tx, projected)
}

func (p *Postgres) Version(ctx context.Context, scope Scope) (int64, error) {
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return 0, err
	}
	var version int64
	err = db.QueryRowContext(ctx, `SELECT sequence FROM gl_projection_state WHERE company=$1`, scope.Company).Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return version, err
}

func pageBounds(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	if page > 1000000 {
		page = 1000000
	}
	return page, limit
}

func (p *Postgres) List(ctx context.Context, scope Scope, kind, search string, page, limit int, filter ListFilter) (Page, error) {
	page, limit = pageBounds(page, limit)
	result := Page{Items: []json.RawMessage{}, Page: page, Limit: limit}
	if !supportedRecord(kind) {
		return result, fmt.Errorf("ไม่พบชนิดข้อมูลบัญชี")
	}
	if filter.BookCode != "" && !contains([]string{"JV", "UV", "SV", "RV", "PV"}, filter.BookCode) || filter.Kind != "" && !contains([]string{"manual", "opening", "closing", "reversal", "mapping"}, filter.Kind) || filter.Status != "" && !contains([]string{"draft", "posted", "reversed", "void"}, filter.Status) {
		return result, fmt.Errorf("ตัวกรองรายการบัญชีไม่ถูกต้อง")
	}
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return result, err
	}
	// Search is a literal substring on code, names, description, and reference (never matching JSON keys or boolean flags).
	where := `company=$1 AND kind=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) AND ($3='' OR strpos(lower(code || ' ' || COALESCE(payload->>'name', '') || ' ' || COALESCE(payload->>'description', '') || ' ' || COALESCE(payload->>'reference', '') || ' ' || COALESCE(jsonb_path_query_array(payload, '$.names[*].name')::text, '')),lower($3))>0) AND ($4='' OR kind NOT IN ('journals','budgets','forecast') OR payload->>'branchcode'=$4) AND (kind<>'journals' OR (($5='' OR payload->>'bookcode'=$5) AND ($6='' OR payload->>'kind'=$6) AND ($7='' OR payload->>'status'=$7)))`
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM gl_records WHERE `+where, scope.Company, kind, search, scope.Branch, filter.BookCode, filter.Kind, filter.Status).Scan(&result.Total); err != nil {
		return result, err
	}
	rows, err := db.QueryContext(ctx, `SELECT payload || jsonb_build_object('id', id, 'version', version) FROM gl_records WHERE `+where+` ORDER BY code,id LIMIT $8 OFFSET $9`, scope.Company, kind, search, scope.Branch, filter.BookCode, filter.Kind, filter.Status, limit, (page-1)*limit)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var data []byte
		if err = rows.Scan(&data); err != nil {
			return result, err
		}
		result.Items = append(result.Items, json.RawMessage(data))
	}
	return result, rows.Err()
}

func (p *Postgres) Get(ctx context.Context, scope Scope, kind, id string) (json.RawMessage, error) {
	if !supportedRecord(kind) {
		return nil, fmt.Errorf("ไม่พบชนิดข้อมูลบัญชี")
	}
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return nil, err
	}
	var data []byte
	err = db.QueryRowContext(ctx, `SELECT payload || jsonb_build_object('id', id, 'version', version) FROM gl_records WHERE company=$1 AND kind=$2 AND id=$3 AND NOT COALESCE((payload->>'isdeleted')::boolean,false) AND ($4='' OR kind NOT IN ('journals','budgets','forecast') OR payload->>'branchcode'=$4)`, scope.Company, kind, id, scope.Branch).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return json.RawMessage(data), err
}

// Rebuild replaces derived records and lines atomically from the audit stream.
// Readers keep their committed snapshot; the immutable events and watermark stay.
func (p *Postgres) Rebuild(ctx context.Context, scope Scope) error {
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = lockCompany(ctx, tx, scope.Company); err != nil {
		return err
	}
	var version int64
	err = tx.QueryRowContext(ctx, `SELECT sequence FROM gl_projection_state WHERE company=$1`, scope.Company).Scan(&version)
	if err == sql.ErrNoRows {
		return tx.Commit()
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM gl_lines WHERE company=$1`, scope.Company); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM gl_records WHERE company=$1`, scope.Company); err != nil {
		return err
	}
	// Keyset batches bound memory; rows must close before statements on tx.
	for after := int64(0); after < version; {
		rows, err := tx.QueryContext(ctx, `SELECT sequence,event_hash,payload FROM gl_events WHERE company=$1 AND sequence>$2 ORDER BY sequence LIMIT 100`, scope.Company, after)
		if err != nil {
			return err
		}
		events := []Event{}
		for rows.Next() {
			var sequence int64
			var hash string
			var data []byte
			if err = rows.Scan(&sequence, &hash, &data); err != nil {
				rows.Close()
				return err
			}
			var event Event
			if err = json.Unmarshal(data, &event); err != nil {
				rows.Close()
				return err
			}
			actual, _, err := eventDigest(event)
			if err != nil || actual != hash || sequence != event.Sequence || sequence != after+int64(len(events))+1 {
				rows.Close()
				return fmt.Errorf("ประวัติบัญชีไม่ต่อเนื่องหรือไม่ตรงกับต้นฉบับ")
			}
			events = append(events, event)
		}
		rowErr := rows.Err()
		rows.Close()
		if rowErr != nil {
			return rowErr
		}
		if len(events) == 0 {
			return fmt.Errorf("ประวัติบัญชีขาดหาย")
		}
		for _, event := range events {
			if err = applyChanges(ctx, tx, event); err != nil {
				return err
			}
			after = event.Sequence
		}
		if after > version {
			return fmt.Errorf("ลำดับประวัติบัญชีเกินยอดประมวลผล")
		}
	}
	return tx.Commit()
}
