package cheque

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"
)

// คำนวณยอดเช็คใหม่ (เมนู cheque-recalculate): สถานะในทะเบียนเป็น cache ของ event สถานะล่าสุดที่ยังไม่ถูกยกเลิก
// — ตรวจคืนรายการที่ cache ไม่ตรงประวัติ แล้ว "ปรับให้ตรง" เขียน cache + event update พร้อมเหตุผล (cheque-bank.md §4.5)

// Mismatch คือเช็คหนึ่งใบที่สถานะ/สมุดในทะเบียนไม่ตรงกับประวัติ
type Mismatch struct {
	ID            string `json:"id"`
	ChequeType    int    `json:"cheque_type"`
	ChequeNo      string `json:"cheque_no"`
	Status        int    `json:"status"`
	DerivedStatus int    `json:"derived_status"`
	BookCode      string `json:"book_code"`
	DerivedBook   string `json:"derived_book"`
}

// RecalcResult คือผลตรวจ: checked = จำนวนเช็คที่ตรวจ, mismatches = รายการที่ไม่ตรง (สูงสุด maxMismatches แถว)
type RecalcResult struct {
	Checked    int        `json:"checked"`
	Total      int        `json:"total"`
	Mismatches []Mismatch `json:"mismatches"`
}

const maxMismatches = 500

// recalcRow คือเช็คหนึ่งใบกับสถานะที่ derive จาก event ล่าสุดที่ยังไม่ถูกยกเลิก (ไม่มี event = สถานะ 0)
const recalcSQL = `SELECT r.id::text, r.cheque_type, r.cheque_no, r.status, COALESCE(r.book_code, ''),
	COALESCE(e.to_status, 0), COALESCE(e.book_code, '')
FROM cheque_registers r
LEFT JOIN LATERAL (SELECT ev.to_status, ev.book_code FROM cheque_events ev
	WHERE ev.company = r.company AND ev.cheque_id = r.id AND ev.voided_at IS NULL AND ev.to_status IS NOT NULL
	ORDER BY ev.event_date DESC, ev.id DESC LIMIT 1) e ON true
WHERE r.company = $1 AND r.deleted_at IS NULL`

// expectedBook: เช็ครับที่ต้องย้อนกลับเป็นเช็คในมือถูกล้างสมุดเหมือน Champ (CQFrmBaseTransStatus.cpp:188-190);
// เช็ครับในมืออยู่แล้วคงสมุดที่ระบุตอนบันทึก; สถานะอื่นคงสมุดในทะเบียน (ผู้ใช้อาจแก้หลังยืนยันตาม D8) และเติมจากประวัติเมื่อว่าง
func expectedBook(chequeType, stored, derived int, current, eventBook string) string {
	if chequeType == TypeReceived && derived == InOnHand {
		if stored != InOnHand {
			return ""
		}
		return current
	}
	if current != "" {
		return current
	}
	return eventBook
}

func scanMismatch(row rowScanner) (Mismatch, bool, error) {
	var m Mismatch
	var eventBook string
	if err := row.Scan(&m.ID, &m.ChequeType, &m.ChequeNo, &m.Status, &m.BookCode, &m.DerivedStatus, &eventBook); err != nil {
		return m, false, err
	}
	m.DerivedBook = expectedBook(m.ChequeType, m.Status, m.DerivedStatus, m.BookCode, eventBook)
	return m, m.Status != m.DerivedStatus || m.BookCode != m.DerivedBook, nil
}

// CheckRecalculate ตรวจทุกเช็คของบริษัท (อ่านอย่างเดียว) เรียงตามชนิด/เลขเช็ค
func CheckRecalculate(ctx context.Context, q querier, company string) (RecalcResult, error) {
	result := RecalcResult{Mismatches: []Mismatch{}}
	rows, err := q.QueryContext(ctx, recalcSQL+` ORDER BY r.cheque_type, r.cheque_no, r.id`, company)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		m, differs, err := scanMismatch(rows)
		if err != nil {
			return result, err
		}
		result.Checked++
		if !differs {
			continue
		}
		result.Total++
		if len(result.Mismatches) < maxMismatches {
			result.Mismatches = append(result.Mismatches, m)
		}
	}
	return result, rows.Err()
}

// ApplyRecalculate ปรับ cache ของเช็คที่ไม่ตรงทั้งหมดใน tx เดียว: ล็อกแถวเช็คก่อนแล้วคำนวณซ้ำ (เอกสารสถานะที่บันทึกพร้อมกัน
// ต้องรอแถวเดียวกัน) แล้วบันทึก event update ที่มีค่าเดิม→ค่าใหม่ + เหตุผล; ไม่ลงบัญชี
func ApplyRecalculate(ctx context.Context, db *sql.DB, scope Scope, requestID, reason string) (RecalcResult, error) {
	reason, err := cleanText(reason, "reason", maxDescriptionLen, true)
	if err != nil {
		if e, ok := AsError(err); ok && e.Key == "cheque_err_required_field" {
			return RecalcResult{}, badRequest("cheque_err_reason_required", "reason")
		}
		return RecalcResult{}, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return RecalcResult{}, err
	}
	defer tx.Rollback()
	var out RecalcResult
	if replayed, err := claimRequest(ctx, tx, scope, requestID, "recalculate:apply", &out); err != nil || replayed {
		return out, err
	}
	found, err := CheckRecalculate(ctx, tx, scope.Company)
	if err != nil {
		return RecalcResult{}, err
	}
	out = RecalcResult{Checked: found.Checked, Mismatches: []Mismatch{}}
	ids := mismatchIDs(found.Mismatches)
	for len(ids) > 0 {
		fixed, err := applyMismatches(ctx, tx, scope, ids, reason)
		if err != nil {
			return RecalcResult{}, err
		}
		out.Mismatches = append(out.Mismatches, fixed...)
		if found.Total <= maxMismatches {
			break
		}
		// มีมากกว่าหนึ่งชุด: ตรวจใหม่หลังปรับชุดแรก (ชุดที่ปรับแล้วตรงแล้ว จึงไม่วนซ้ำ)
		if found, err = CheckRecalculate(ctx, tx, scope.Company); err != nil {
			return RecalcResult{}, err
		}
		ids = mismatchIDs(found.Mismatches)
	}
	out.Total = len(out.Mismatches)
	if err = completeRequest(ctx, tx, scope, requestID, out); err != nil {
		return RecalcResult{}, err
	}
	return out, tx.Commit()
}

func mismatchIDs(items []Mismatch) []string {
	ids := make([]string, len(items))
	for i, m := range items {
		ids[i] = m.ID
	}
	return ids
}

// applyMismatches ล็อกเช็คตามลำดับ id (กัน deadlock) คำนวณซ้ำหลังล็อก แล้วปรับเฉพาะใบที่ยังไม่ตรง
func applyMismatches(ctx context.Context, tx *sql.Tx, scope Scope, ids []string, reason string) ([]Mismatch, error) {
	args := make([]any, 0, len(ids)+1)
	args = append(args, scope.Company)
	marks := make([]string, len(ids))
	for i, id := range ids {
		args = append(args, id)
		marks[i] = "$" + strconv.Itoa(i+2)
	}
	if _, err := tx.ExecContext(ctx, `SELECT 1 FROM cheque_registers WHERE company = $1 AND id IN (`+strings.Join(marks, ",")+`) ORDER BY id FOR UPDATE`, args...); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, recalcSQL+` AND r.id IN (`+strings.Join(marks, ",")+`) ORDER BY r.cheque_type, r.cheque_no, r.id`, args...)
	if err != nil {
		return nil, err
	}
	pending := []Mismatch{}
	for rows.Next() {
		m, differs, err := scanMismatch(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		if differs {
			pending = append(pending, m)
		}
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	today := businessDay(time.Now())
	for _, m := range pending {
		if _, err = tx.ExecContext(ctx, `UPDATE cheque_registers SET status = $3, book_code = $4, version = version + 1, updated_at = now(), updated_by = $5
			WHERE company = $1 AND id = $2`, scope.Company, m.ID, m.DerivedStatus, nullable(m.DerivedBook), scope.Actor); err != nil {
			return nil, err
		}
		detail := map[string]change{}
		if m.Status != m.DerivedStatus {
			detail["status"] = change{From: strconv.Itoa(m.Status), To: strconv.Itoa(m.DerivedStatus)}
		}
		if m.BookCode != m.DerivedBook {
			detail["book_code"] = change{From: m.BookCode, To: m.DerivedBook}
		}
		if err = insertEvent(ctx, tx, scope, eventRow{chequeID: m.ID, kind: eventUpdate, date: today, book: m.DerivedBook, reason: reason, detail: detail}); err != nil {
			return nil, err
		}
	}
	return pending, nil
}
