package cheque

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"time"

	"github.com/lib/pq"

	authmodels "smlcloudplatform/internal/authentication/models"
)

// ชนิดของ event ในประวัติเช็ค (append-only)
const (
	eventCreate       = "create"        // สร้างทะเบียน (to_status = สถานะเริ่มต้น)
	eventUpdate       = "update"        // แก้ฟิลด์ที่แก้ได้เสมอ / ปรับสถานะให้ตรงประวัติ (recalculate)
	eventEdit         = "edit"          // แก้ฟิลด์หลักของเช็คที่ถูกอ้างแล้ว หลังผู้ใช้ยืนยัน (D8)
	eventManualStatus = "manual_status" // แก้สถานะตรงในจอทะเบียน หลังผู้ใช้ยืนยัน (D8; ไม่ลงบัญชี)
	eventStatusDoc    = "status_doc"    // เอกสารเปลี่ยนสถานะ 419..426
	eventDelete       = "delete"        // ลบเช็คออกจากทะเบียน (แถวยังอยู่พร้อม deleted_at; ไม่มี to_status จึงไม่กระทบสถานะที่ derive)
)

// Event คือหนึ่งแถวในประวัติของเช็ค
type Event struct {
	ID         int64           `json:"id"`
	ChequeID   string          `json:"cheque_id"`
	ChequeNo   string          `json:"cheque_no,omitempty"`
	EventKind  string          `json:"event_kind"`
	DocFlag    int             `json:"doc_flag,omitempty"`
	DocNo      string          `json:"doc_no,omitempty"`
	EventDate  string          `json:"event_date"`
	FromStatus *int            `json:"from_status"`
	ToStatus   *int            `json:"to_status"`
	BookCode   string          `json:"book_code"`
	Detail     json.RawMessage `json:"detail"`
	Reason     string          `json:"reason"`
	Actor      string          `json:"actor"`
	CreatedAt  time.Time       `json:"created_at"`
	VoidedAt   *time.Time      `json:"voided_at,omitempty"`
	VoidedBy   string          `json:"voided_by,omitempty"`
	VoidReason string          `json:"void_reason,omitempty"`
}

const eventColumns = `e.id, e.cheque_id::text, r.cheque_no, e.event_kind, COALESCE(e.doc_flag, 0), COALESCE(e.doc_no, ''),
	to_char(e.event_date, 'YYYY-MM-DD'), e.from_status, e.to_status, COALESCE(e.book_code, ''), e.detail, e.reason, e.actor,
	e.created_at, e.voided_at, COALESCE(e.voided_by, ''), COALESCE(e.void_reason, '')`

func scanEvent(row rowScanner) (Event, error) {
	var (
		e        Event
		from, to sql.NullInt64
		voidedAt sql.NullTime
		detail   []byte
	)
	if err := row.Scan(&e.ID, &e.ChequeID, &e.ChequeNo, &e.EventKind, &e.DocFlag, &e.DocNo, &e.EventDate, &from, &to,
		&e.BookCode, &detail, &e.Reason, &e.Actor, &e.CreatedAt, &voidedAt, &e.VoidedBy, &e.VoidReason); err != nil {
		return Event{}, err
	}
	e.FromStatus, e.ToStatus = nullInt(from), nullInt(to)
	if voidedAt.Valid {
		e.VoidedAt = &voidedAt.Time
	}
	e.Detail = json.RawMessage(detail)
	return e, nil
}

func nullInt(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	n := int(v.Int64)
	return &n
}

// eventRow คือ event ที่จะเพิ่ม (doc_flag/doc_no เฉพาะเอกสารสถานะ; from/to เฉพาะ event ที่เปลี่ยนสถานะ)
type eventRow struct {
	chequeID, kind, docNo, date, book, reason string
	docFlag                                   int
	from, to                                  *int
	detail                                    any
}

func insertEvent(ctx context.Context, tx *sql.Tx, scope Scope, ev eventRow) error {
	detail := []byte("{}")
	if ev.detail != nil {
		var err error
		if detail, err = json.Marshal(ev.detail); err != nil {
			return err
		}
	}
	var docFlag any
	if ev.docFlag != 0 {
		docFlag = ev.docFlag
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO cheque_events (company, cheque_id, event_kind, doc_flag, doc_no, event_date, from_status, to_status, book_code, detail, reason, actor)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		scope.Company, ev.chequeID, ev.kind, docFlag, nullable(ev.docNo), ev.date, intOrNil(ev.from), intOrNil(ev.to), nullable(ev.book), detail, ev.reason, scope.Actor)
	return err
}

func intOrNil(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func intPtr(v int) *int { return &v }

// businessDay คือวันที่ทำรายการตามปฏิทินของ holding (Asia/Bangkok) — วันที่ของ event ที่ไม่ได้มาจากเอกสาร
func businessDay(now time.Time) string {
	return now.In(authmodels.HoldingLocation("")).Format("2006-01-02")
}

// derivedState คือสถานะ/สมุดที่ควรเป็นตาม event สถานะล่าสุดที่ยังไม่ถูกยกเลิก (event_date DESC, id DESC เทียบ GetChqInStatus);
// ไม่มี event = สถานะ 0
type derivedState struct {
	status   int
	book     string
	date     string // วันที่ของ event นั้น
	hasEvent bool
}

func deriveStates(ctx context.Context, q querier, company string, ids []string) (map[string]derivedState, error) {
	rows, err := q.QueryContext(ctx, `SELECT DISTINCT ON (cheque_id) cheque_id::text, to_status, COALESCE(book_code, ''), to_char(event_date, 'YYYY-MM-DD')
		FROM cheque_events WHERE company = $1 AND cheque_id = ANY($2::uuid[]) AND voided_at IS NULL AND to_status IS NOT NULL
		ORDER BY cheque_id, event_date DESC, id DESC`, company, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	states := map[string]derivedState{}
	for rows.Next() {
		var (
			id string
			st derivedState
		)
		if err = rows.Scan(&id, &st.status, &st.book, &st.date); err != nil {
			return nil, err
		}
		st.hasEvent = true
		states[id] = st
	}
	return states, rows.Err()
}

// restoreStates คำนวณสถานะของเช็คใหม่จากประวัติหลัง void event (ยกเลิก/แก้เอกสารสถานะ) แล้วเขียน cache ในทะเบียน
func restoreStates(ctx context.Context, tx *sql.Tx, scope Scope, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	states, err := deriveStates(ctx, tx, scope.Company, ids)
	if err != nil {
		return err
	}
	// เขียนเรียงตาม id เหมือน lockDocCheques/ApplyRecalculate (RETURNING ของ voidDocEvents ไม่เรียง) กันสองคำขอล็อกแถวสลับลำดับจน deadlock
	sort.Strings(ids)
	for _, id := range ids {
		st := states[id]
		if _, err = tx.ExecContext(ctx, `UPDATE cheque_registers SET status = $3::smallint,
				book_code = CASE WHEN cheque_type = 0 AND $3::smallint = 0 THEN NULL WHEN $4::text <> '' THEN $4::text ELSE book_code END,
				version = version + 1, updated_at = now(), updated_by = $5
			WHERE company = $1 AND id = $2`, scope.Company, id, st.status, st.book, scope.Actor); err != nil {
			return err
		}
	}
	return nil
}

// newerEvent หา event สถานะที่ใหม่กว่าและยังไม่ถูกยกเลิกของเช็คในเอกสาร (Champ IsChqReference): คืนเลขเช็คและเอกสารที่ใหม่กว่า
func newerEvent(ctx context.Context, tx *sql.Tx, company string, flag int, docNo string) (string, string, error) {
	var chequeNo, newer string
	err := tx.QueryRowContext(ctx, `SELECT r.cheque_no, COALESCE(n.doc_no, '')
		FROM cheque_events e
		JOIN cheque_events n ON n.company = e.company AND n.cheque_id = e.cheque_id AND n.voided_at IS NULL
			AND n.event_kind = 'status_doc' AND (n.event_date, n.id) > (e.event_date, e.id)
		JOIN cheque_registers r ON r.company = e.company AND r.id = e.cheque_id
		WHERE e.company = $1 AND e.doc_flag = $2 AND e.doc_no = $3 AND e.event_kind = 'status_doc' AND e.voided_at IS NULL
		ORDER BY n.event_date, n.id LIMIT 1`, company, flag, docNo).Scan(&chequeNo, &newer)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	return chequeNo, newer, err
}

// voidDocEvents ยกเลิก event ของเอกสารสถานะ (เก็บแถวเดิมไว้เป็นหลักฐาน) แล้วคืนเช็คที่ได้รับผล
func voidDocEvents(ctx context.Context, tx *sql.Tx, scope Scope, flag int, docNo, reason string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `UPDATE cheque_events SET voided_at = now(), voided_by = $4, void_reason = $5
		WHERE company = $1 AND doc_flag = $2 AND doc_no = $3 AND event_kind = 'status_doc' AND voided_at IS NULL
		RETURNING cheque_id::text`, scope.Company, flag, docNo, scope.Actor, reason)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ListDocEvents คือประวัติ event ของเอกสารสถานะหนึ่งฉบับ (รวมรอบแก้ไขที่ถูกยกเลิกแล้ว) เรียงตามเวลา
func ListDocEvents(ctx context.Context, q querier, company string, flag int, docNo string) ([]Event, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+eventColumns+` FROM cheque_events e
		JOIN cheque_registers r ON r.company = e.company AND r.id = e.cheque_id
		WHERE e.company = $1 AND e.doc_flag = $2 AND e.doc_no = $3 ORDER BY e.id`, company, flag, docNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Event{}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

// lastEvent คือ event ล่าสุดที่ยังไม่ถูกยกเลิกของเช็ค ("การปรับปรุงเอกสาร" ในจอทะเบียน)
func lastEvent(ctx context.Context, q querier, company, id string) (*Event, error) {
	e, err := scanEvent(q.QueryRowContext(ctx, `SELECT `+eventColumns+` FROM cheque_events e
		JOIN cheque_registers r ON r.company = e.company AND r.id = e.cheque_id
		WHERE e.company = $1 AND e.cheque_id = $2 AND e.voided_at IS NULL ORDER BY e.id DESC LIMIT 1`, company, id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}
