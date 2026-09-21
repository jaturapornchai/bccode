package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type ReviewInput struct {
	Status          int    `json:"status"`
	Note            string `json:"note"`
	ExpectedEventNo *int64 `json:"expectedEventNo"`
}
type ReviewEvent struct {
	EventNo    int64     `json:"eventno"`
	Version    int64     `json:"version"`
	Status     int       `json:"status"`
	Note       string    `json:"note"`
	ReviewedBy string    `json:"reviewedby"`
	ReviewedAt time.Time `json:"reviewedat"`
}
type JournalReview struct {
	JournalID string        `json:"journalid"`
	Version   int64         `json:"version"`
	Status    int           `json:"status"`
	EventNo   int64         `json:"eventno"`
	Events    []ReviewEvent `json:"events"`
}

func (r *ReviewInput) validate() error {
	if r == nil || r.Status < 1 || r.Status > 3 || r.ExpectedEventNo == nil || *r.ExpectedEventNo < 0 {
		return fmt.Errorf("ข้อมูลผลตรวจไม่ถูกต้อง")
	}
	if utf8.RuneCountInString(r.Note) > 4000 {
		return fmt.Errorf("หมายเหตุผลตรวจยาวเกิน 4000 ตัวอักษร")
	}
	if r.Status == 2 && strings.TrimSpace(r.Note) == "" {
		return fmt.Errorf("กรุณาระบุข้อแตกต่างที่พบ")
	}
	return nil
}
func (s *PostgresStore) mutateReview(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if err := cmd.Review.validate(); err != nil {
		return nil, err
	}
	var j Journal
	if err := s.loadRecord(ctx, tx, scope.Company, "journals", cmd.ID, &j); err != nil {
		return nil, err
	}
	if scope.Branch != "" && j.BranchCode != scope.Branch {
		return nil, ErrNotFound
	}
	if err := checkVersion(cmd, j.Identity); err != nil {
		return nil, err
	}
	var latest int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(event_no),0) FROM gl_journal_review_events WHERE company=$1 AND journal_id=$2 AND journal_version=$3`, scope.Company, j.ID, j.Version).Scan(&latest); err != nil {
		return nil, err
	}
	if latest != *cmd.Review.ExpectedEventNo {
		return nil, userError(CodeStaleVersion, "ผลตรวจเปลี่ยนไปแล้ว กรุณาโหลดข้อมูลล่าสุด")
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO gl_journal_review_events(company,journal_id,journal_version,status,note,reviewed_by,reviewed_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, scope.Company, j.ID, j.Version, cmd.Review.Status, strings.TrimSpace(cmd.Review.Note), scope.Actor, now)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
func (p *Postgres) JournalReview(ctx context.Context, scope Scope, id string) (JournalReview, error) {
	raw, err := p.Get(ctx, scope, "journals", id)
	if err != nil {
		return JournalReview{}, err
	}
	var j Journal
	if err = json.Unmarshal(raw, &j); err != nil {
		return JournalReview{}, err
	}
	result := JournalReview{JournalID: id, Version: j.Version, Status: 1, Events: []ReviewEvent{}}
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return result, err
	}
	rows, err := db.QueryContext(ctx, `SELECT event_no,journal_version,status,note,reviewed_by,reviewed_at FROM gl_journal_review_events WHERE company=$1 AND journal_id=$2 ORDER BY event_no DESC LIMIT 100`, scope.Company, id)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var e ReviewEvent
		if err = rows.Scan(&e.EventNo, &e.Version, &e.Status, &e.Note, &e.ReviewedBy, &e.ReviewedAt); err != nil {
			return result, err
		}
		result.Events = append(result.Events, e)
		if e.Version == j.Version && result.EventNo == 0 {
			result.EventNo = e.EventNo
			result.Status = e.Status
		}
	}
	return result, rows.Err()
}
