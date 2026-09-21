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
)

func validateJournalSource(j Journal) error {
	if j.SourceType < 0 || j.SourceType > 6 {
		return fmt.Errorf("ประเภทข้อมูลต้นทางไม่ถูกต้อง")
	}
	system, record := strings.TrimSpace(j.SourceSystem), strings.TrimSpace(j.SourceRecordID)
	if (system == "") != (record == "") || utf8.RuneCountInString(j.SourceSystem) > 100 || utf8.RuneCountInString(j.SourceRecordID) > 150 {
		return fmt.Errorf("กรุณาระบุระบบต้นทางและรหัสรายการต้นทางคู่กัน")
	}
	if system != j.SourceSystem || record != j.SourceRecordID {
		return fmt.Errorf("รหัสต้นทางต้องไม่มีช่องว่างหัวท้าย")
	}
	if j.SourceType > 1 && system == "" {
		return fmt.Errorf("ข้อมูลนำเข้าต้องมีระบบต้นทางและรหัสรายการต้นทาง")
	}
	return nil
}

func sourceJournalHash(j Journal) (string, error) {
	j.Identity = Identity{}
	j.Status = ""
	j.PostedAt = nil
	j.PostedBy = ""
	if j.SourceType == 0 {
		j.SourceType = 1
	}
	data, err := json.Marshal(j)
	if err != nil {
		return "", err
	}
	return digest(data), nil
}

// Source identity supplements request idempotency under the same company lock.
func (s *PostgresStore) existingSourceJournal(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command) (Result, bool, error) {
	if cmd.Resource != "journals" || cmd.Action != "create" || cmd.Journal == nil {
		return Result{}, false, nil
	}
	if err := validateJournalSource(*cmd.Journal); err != nil {
		return Result{}, false, err
	}
	if cmd.Journal.SourceSystem == "" {
		return Result{}, false, nil
	}
	hash, err := sourceJournalHash(*cmd.Journal)
	if err != nil {
		return Result{}, false, err
	}
	var result Result
	var savedHash, branch string
	var deleted bool
	err = tx.QueryRowContext(ctx, `SELECT s.journal_id,s.version,s.sequence,s.request_hash,COALESCE((r.payload->>'isdeleted')::boolean,false),COALESCE(r.payload->>'branchcode','')
 FROM gl_source_journals s JOIN gl_records r ON r.company=s.company AND r.kind='journals' AND r.id=s.journal_id
 WHERE s.company=$1 AND s.source_system=$2 AND s.source_record_id=$3`, scope.Company, cmd.Journal.SourceSystem, cmd.Journal.SourceRecordID).Scan(&result.ID, &result.Version, &result.Sequence, &savedHash, &deleted, &branch)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{}, false, nil
	}
	if err != nil {
		return Result{}, false, err
	}
	if scope.Branch != "" && branch != scope.Branch {
		return Result{}, false, ErrNotFound
	}
	if deleted || savedHash != hash {
		return Result{}, false, fmt.Errorf("รายการต้นทางนี้มีอยู่แล้ว กรุณาเปิดใบเดิมและแก้ไขตามสิทธิ์ ห้ามนำเข้าทับ")
	}
	return result, true, nil
}

func (s *PostgresStore) recordSourceJournal(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, changes []Change, sequence int64, now time.Time) error {
	if cmd.Resource != "journals" || cmd.Action != "create" || cmd.Journal == nil || cmd.Journal.SourceSystem == "" {
		return nil
	}
	hash, err := sourceJournalHash(*cmd.Journal)
	if err != nil {
		return err
	}
	for _, change := range changes {
		if change.Kind != "journals" {
			continue
		}
		var identity Identity
		if err = json.Unmarshal([]byte(change.Payload), &identity); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO gl_source_journals(company,source_system,source_record_id,journal_id,request_hash,sequence,version,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, scope.Company, cmd.Journal.SourceSystem, cmd.Journal.SourceRecordID, change.ID, hash, sequence, identity.Version, now)
		return err
	}
	return fmt.Errorf("ไม่พบใบสำคัญของข้อมูลต้นทาง")
}

func preserveJournalSource(old, next Journal) error {
	if old.SourceType != next.SourceType || old.SourceSystem != next.SourceSystem || old.SourceRecordID != next.SourceRecordID {
		return fmt.Errorf("ห้ามเปลี่ยนรหัสต้นทางของใบสำคัญเดิม")
	}
	return validateJournalSource(next)
}
