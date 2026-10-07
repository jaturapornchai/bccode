package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shopspring/decimal"
)

// Batch posting (Champ 5512 ผ่านรายการ): the screen filters draft journals, checks each one with the
// single-journal post path (action "check", nothing saved) and posts the ticked ones all-or-nothing
// in one transaction (action "create"), recorded as one posting batch with a running number.
const (
	PostBatchKind           = "post-batches"
	PostBatchMaxJournals    = 200
	PostBatchRemarkMaxRunes = 500
)

// PostBatchInput: filters for "check" (all optional; blank = no bound) and the journals for "create".
type PostBatchInput struct {
	BookCode  string          `json:"bookcode,omitempty"`
	DateFrom  string          `json:"datefrom,omitempty"`
	DateTo    string          `json:"dateto,omitempty"`
	DocNoFrom string          `json:"docnofrom,omitempty"`
	DocNoTo   string          `json:"docnoto,omitempty"`
	Items     []PostBatchItem `json:"items,omitempty"`
	Remark    string          `json:"remark,omitempty"`
}

// PostBatchItem is one draft to post at the version the user checked (stale = whole batch refused).
type PostBatchItem struct {
	ID      string `json:"id"`
	Version int64  `json:"version"`
}

// PostBatch is the stored batch (gl_records kind post-batches, code = batch no). Each journal it
// posted carries the batch no in postbatchno.
type PostBatch struct {
	Identity   `bson:",inline"`
	BatchNo    string             `json:"batchno"`
	PostedAt   time.Time          `json:"postedat"`
	PostedBy   string             `json:"postedby"`
	Count      int                `json:"count"`
	Remark     string             `json:"remark"`
	BranchCode string             `json:"branchcode,omitempty"`
	Journals   []PostBatchJournal `json:"journals"`
}

type PostBatchJournal struct {
	ID    string `json:"id"`
	DocNo string `json:"docno"`
}

// PostCheck answers action "check": the matching drafts in Champ order (book, date, doc no) with the
// result of the post rules per journal. More = more drafts match than one batch may post.
type PostCheck struct {
	Rows  []PostCheckRow `json:"rows"`
	More  bool           `json:"more"`
	Limit int            `json:"limit"`
}

type PostCheckRow struct {
	ID          string `json:"id"`
	Version     int64  `json:"version"`
	DocNo       string `json:"docno"`
	Date        string `json:"date"`
	BookCode    string `json:"bookcode"`
	FiscalYear  string `json:"fiscalyear"`
	BranchCode  string `json:"branchcode"`
	Description string `json:"description"`
	Debit       string `json:"debit"`
	Credit      string `json:"credit"`
	OK          bool   `json:"ok"`
	Code        string `json:"code,omitempty"`
	Message     string `json:"message,omitempty"`
}

func validatePostBatchFilter(f PostBatchInput) error {
	switch {
	case f.DateFrom != "" && !validDate(f.DateFrom):
		return fieldError("post_batch_date_invalid", "datefrom", "วันที่เริ่มต้นไม่ถูกต้อง กรุณาเลือกวันที่ให้ครบ วัน เดือน ปี")
	case f.DateTo != "" && !validDate(f.DateTo):
		return fieldError("post_batch_date_invalid", "dateto", "วันที่สิ้นสุดไม่ถูกต้อง กรุณาเลือกวันที่ให้ครบ วัน เดือน ปี")
	case f.DateFrom != "" && f.DateTo != "" && f.DateFrom > f.DateTo:
		return fieldError("post_batch_date_range_invalid", "dateto", "วันที่สิ้นสุดต้องไม่ก่อนวันที่เริ่มต้น")
	case f.DocNoFrom != "" && f.DocNoTo != "" && f.DocNoFrom > f.DocNoTo:
		return fieldError("post_batch_docno_range_invalid", "docnoto", "เลขที่เอกสารสิ้นสุดต้องไม่น้อยกว่าเลขที่เริ่มต้น")
	}
	return nil
}

// checkPostBatch runs the real post path for every matching draft inside a savepoint that is
// rolled back, so the result is exactly what posting would decide — and nothing is saved.
func (s *PostgresStore) checkPostBatch(ctx context.Context, scope Scope, cmd Command) (Result, error) {
	filter := PostBatchInput{}
	if cmd.PostBatch != nil {
		filter = *cmd.PostBatch
	}
	if err := validatePostBatchFilter(filter); err != nil {
		return Result{}, err
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
	journals, err := draftJournals(ctx, tx, scope, filter)
	if err != nil {
		return Result{}, err
	}
	check := PostCheck{Rows: []PostCheckRow{}, Limit: PostBatchMaxJournals}
	if len(journals) > PostBatchMaxJournals {
		journals, check.More = journals[:PostBatchMaxJournals], true
	}
	now := time.Now().UTC()
	for _, j := range journals {
		row := PostCheckRow{ID: j.ID, Version: j.Version, DocNo: j.DocNo, Date: j.Date, BookCode: j.BookCode, FiscalYear: j.FiscalYear, BranchCode: j.BranchCode, Description: j.Description}
		debit, credit := decimal.Zero, decimal.Zero
		for _, line := range j.Lines {
			debit, credit = debit.Add(line.Debit.Decimal()), credit.Add(line.Credit.Decimal())
		}
		row.Debit, row.Credit = debit.String(), credit.String()
		if _, err = tx.ExecContext(ctx, `SAVEPOINT gl_post_check`); err != nil {
			return Result{}, err
		}
		issue := s.tryPost(ctx, tx, scope, j, now)
		if _, err = tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT gl_post_check`); err != nil {
			return Result{}, err
		}
		if _, err = tx.ExecContext(ctx, `RELEASE SAVEPOINT gl_post_check`); err != nil {
			return Result{}, err
		}
		if issue == nil {
			row.OK = true
		} else if code, message, userFacing := postIssue(issue); userFacing {
			row.Code, row.Message = code, message
		} else {
			return Result{}, issue
		}
		check.Rows = append(check.Rows, row)
	}
	return Result{PostCheck: &check}, nil
}

// tryPost = the single-journal post (mutateJournal "post") plus its projection into gl_lines.
func (s *PostgresStore) tryPost(ctx context.Context, tx *sql.Tx, scope Scope, j Journal, now time.Time) error {
	changes, err := s.mutateJournal(ctx, tx, scope, Command{Resource: "journals", Action: "post", ID: j.ID, Version: j.Version}, now)
	if err != nil {
		return err
	}
	return applyChanges(ctx, tx, Event{HoldingCode: scope.Holding, BusinessCode: scope.Company, Changes: changes})
}

func draftJournals(ctx context.Context, tx *sql.Tx, scope Scope, f PostBatchInput) ([]Journal, error) {
	rows, err := tx.QueryContext(ctx, `SELECT payload || jsonb_build_object('id',id,'version',version) FROM gl_records
		WHERE company=$1 AND kind='journals' AND payload->>'status'='draft' AND NOT COALESCE((payload->>'isdeleted')::boolean,false)
		AND ($2='' OR payload->>'branchcode'=$2) AND ($3='' OR payload->>'bookcode'=$3)
		AND ($4='' OR payload->>'date'>=$4) AND ($5='' OR payload->>'date'<=$5)
		AND ($6='' OR code>=$6) AND ($7='' OR code<=$7)
		ORDER BY payload->>'bookcode', payload->>'date', code, id LIMIT $8`,
		scope.Company, scope.Branch, f.BookCode, f.DateFrom, f.DateTo, f.DocNoFrom, f.DocNoTo, PostBatchMaxJournals+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	journals := []Journal{}
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var j Journal
		if err := json.Unmarshal(data, &j); err != nil {
			return nil, err
		}
		journals = append(journals, j)
	}
	return journals, rows.Err()
}

// postBatch is post-batches action "create": every journal goes through the single-journal post and
// gets its own audit event (as if posted one by one), then the batch record is the last event.
// The caller's transaction makes the whole batch all-or-nothing.
func (s *PostgresStore) postBatch(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, hash string, now time.Time) (Event, error) {
	if cmd.Action != "create" {
		return Event{}, fmt.Errorf("ไม่รองรับคำสั่งบัญชีนี้")
	}
	if cmd.PostBatch == nil || len(cmd.PostBatch.Items) == 0 {
		return Event{}, fieldError("post_batch_empty", "items", "กรุณาเลือกเอกสารที่จะผ่านรายการอย่างน้อย 1 ใบ")
	}
	in := cmd.PostBatch
	if len(in.Items) > PostBatchMaxJournals {
		return Event{}, &UserError{Code: "post_batch_too_many", Field: "items", Status: http.StatusBadRequest, Args: []string{strconv.Itoa(PostBatchMaxJournals)},
			Message: fmt.Sprintf("ผ่านรายการได้ครั้งละไม่เกิน %d ใบ กรุณาแบ่งเป็นหลายชุด", PostBatchMaxJournals)}
	}
	remark := strings.TrimSpace(in.Remark)
	if utf8.RuneCountInString(remark) > PostBatchRemarkMaxRunes {
		return Event{}, &UserError{Code: "post_batch_remark_too_long", Field: "remark", Status: http.StatusBadRequest, Args: []string{strconv.Itoa(PostBatchRemarkMaxRunes)},
			Message: fmt.Sprintf("หมายเหตุชุดผ่านรายการต้องไม่เกิน %d ตัวอักษร", PostBatchRemarkMaxRunes)}
	}
	seen := map[string]bool{}
	for _, item := range in.Items {
		if item.ID == "" || item.Version < 1 || seen[item.ID] {
			return Event{}, fieldError("post_batch_item_invalid", "items", "รายการเอกสารที่เลือกซ้ำหรือไม่ครบ กรุณากดตรวจสอบใหม่แล้วเลือกอีกครั้ง")
		}
		seen[item.ID] = true
	}
	batchNo, err := nextPostBatchNo(ctx, tx, scope.Company)
	if err != nil {
		return Event{}, err
	}
	batch := PostBatch{BatchNo: batchNo, PostedAt: now, PostedBy: scope.Actor, Remark: remark, BranchCode: scope.Branch, Journals: []PostBatchJournal{}}
	for index, item := range in.Items {
		post := Command{Resource: "journals", Action: "post", ID: item.ID, Version: item.Version, RequestID: fmt.Sprintf("%s#%d", cmd.RequestID, index+1), Reason: cmd.Reason, postBatchNo: batchNo}
		changes, err := s.mutateJournal(ctx, tx, scope, post, now)
		if err == nil {
			var raw []byte
			if raw, err = json.Marshal(post); err == nil {
				_, err = s.appendEvent(ctx, tx, scope, post, digest(raw), changes, now)
			}
		}
		if err != nil {
			return Event{}, postBatchItemError(ctx, tx, scope, item.ID, err)
		}
		batch.Journals = append(batch.Journals, PostBatchJournal{ID: item.ID, DocNo: changes[0].Code})
	}
	batch.Count = len(batch.Journals)
	batch.Identity = newIdentity(scope, now)
	changes, err := pgChange(PostBatchKind, batch.ID, batch.BatchNo, batch)
	if err != nil {
		return Event{}, err
	}
	return s.appendEvent(ctx, tx, scope, cmd, hash, changes, now)
}

// nextPostBatchNo: running number per company (000001, 000002, …); the company lock of the
// command transaction serialises concurrent batches.
func nextPostBatchNo(ctx context.Context, tx *sql.Tx, company string) (string, error) {
	var last string
	err := tx.QueryRowContext(ctx, `SELECT code FROM gl_records WHERE company=$1 AND kind=$2 ORDER BY length(code) DESC, code DESC LIMIT 1`, company, PostBatchKind).Scan(&last)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Sprintf("%06d", 1), nil
	}
	if err != nil {
		return "", err
	}
	n, err := strconv.ParseInt(last, 10, 64)
	if err != nil {
		return "", fmt.Errorf("เลขที่ชุดผ่านรายการล่าสุดไม่ถูกต้อง: %w", err)
	}
	return fmt.Sprintf("%06d", n+1), nil
}

// postBatchItemError names the journal that stopped the batch; nothing of the batch is saved.
func postBatchItemError(ctx context.Context, tx *sql.Tx, scope Scope, id string, err error) error {
	_, message, userFacing := postIssue(err)
	if !userFacing {
		return err
	}
	status := http.StatusConflict
	if user, ok := AsUserError(err); ok {
		status = user.HTTPStatus()
	}
	docNo := id
	var code string
	if tx.QueryRowContext(ctx, `SELECT code FROM gl_records WHERE company=$1 AND kind='journals' AND id=$2`, scope.Company, id).Scan(&code) == nil && code != "" {
		docNo = code
	}
	return &UserError{Code: "post_batch_item_failed", Status: status, Args: []string{docNo, message},
		Message: fmt.Sprintf("ยังไม่ได้ผ่านรายการชุดนี้ เพราะเอกสาร %s: %s", docNo, message)}
}

// postIssue: the text of a failure the user can fix (rule of a journal), or false for a system fault.
func postIssue(err error) (code, message string, userFacing bool) {
	if user, ok := AsUserError(err); ok {
		return user.Code, user.Message, true
	}
	if errors.Is(err, ErrNotFound) {
		return "not_found", "ไม่พบเอกสารนี้ในบริษัทหรือสาขาที่เลือก หรือเอกสารถูกลบไปแล้ว", true
	}
	message = err.Error()
	if strings.IndexFunc(message, func(r rune) bool { return r >= 0x0E00 && r <= 0x0E7F }) >= 0 && !strings.Contains(message, "SQLSTATE") {
		return "invalid_request", message, true
	}
	return "", "", false
}
