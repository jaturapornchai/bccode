package generalledger

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/goapi/language"
)

// APOpenQuery คือเงื่อนไขของหนี้เจ้าหนี้ค้างจ่ายของเจ้าหนี้หนึ่งราย ณ วันที่ (design ap-payment-prepare §4.2)
type APOpenQuery struct {
	Partner        string   // รหัสเจ้าหนี้ (บังคับ)
	AsOf           string   // YYYY-MM-DD
	Q              string   // คำค้นเลขที่เอกสาร ('' = ไม่กรอง)
	ExcludePrepare string   // เลขที่ใบเตรียมจ่ายที่กำลังแก้ — ไม่นับยอดจองของใบนั้นเอง
	IDs            []string // เฉพาะแถวหนี้เหล่านี้ (ว่าง = ทุกแถว)
	Limit, Offset  int
}

// APOpenItem คือหนี้ค้างจ่ายหนึ่งแถวของบัญชีย่อยเจ้าหนี้ (ยอดเป็นข้อความทศนิยมจาก NUMERIC ของ PostgreSQL)
// Remaining = ยอดคงเหลือ as-of, Reserved = ยอดที่ใบเตรียมจ่ายอื่นยังจองไว้, Available = ยอดที่ยังเตรียมจ่ายได้,
// PrepareUsed = ยอดที่ตัดหนี้แล้วจากใบ ExcludePrepare (as-of เดียวกัน) — เพดานของบรรทัดเดิม = Available + PrepareUsed
// ซึ่งไม่เกินเพดานเดียวกันที่คำนวณ ณ ปัจจุบัน (นับใบจ่ายที่ลงวันที่หลัง as-of ด้วย)
type APOpenItem struct {
	ID           string
	DocumentNo   string
	DocumentDate string
	DueDate      string
	Description  string
	Kind         int    // document_kind ของเอกสารจากจอบัญชี (0 = เอกสาร ERP)
	FromGL       bool   // สร้างจากจอบัญชี (มี balance_side)
	Amount       Amount // ยอดของเอกสารหนี้
	Remaining    Amount
	Reserved     Amount
	Available    Amount
	PrepareUsed  Amount
}

// apBalanceCTE คือสูตรเดียวของยอดหนี้เจ้าหนี้และยอดจอง (APOpenItems + APReservations) ต่อจาก subledgerReadCTE
// ($1 company, $2 branch, $3 q, $4 as-of) และใช้สูตรยอดคงเหลือเดียวกับรายงานอายุหนี้:
// หนี้จาก ERP นับเมื่อใบสำคัญที่สร้างหนี้ผ่านรายการแล้ว, หนี้จากจอบัญชีนับจากการจัดสรรที่ผ่านรายการ, การตัดหนี้นับเฉพาะที่ผ่านรายการ
// as-of ทั้งฝั่ง debt_id และ payment_id; เฟสนี้เฉพาะฝั่งหนี้ (normalize side ตามแหล่ง)
// เพิ่ม $5 partner (ว่าง = ทุกราย), $6 ใบเตรียมจ่ายที่ไม่นับยอดจอง, $7 ids (ว่าง = ทุกแถว)
// open_items: posted = ยอดหนี้ที่ผ่านบัญชี, remaining = posted − การตัดหนี้; reserved: ยอดที่ใบเตรียมจ่ายที่ยังไม่ยกเลิกจองไว้และใบจ่ายยังไม่ใช้
const apBalanceCTE = `, posted_alloc AS (
 SELECT a.document_id, SUM(a.amount) AS amt FROM active_allocations a JOIN posted_journals j ON j.id=a.journal_id GROUP BY 1
), posted_settle AS (
 SELECT x.document_id, SUM(s.amount) AS amt FROM active_settlements s CROSS JOIN LATERAL (VALUES (s.debt_id),(s.payment_id)) x(document_id) GROUP BY 1
), debt AS (
 SELECT d.id, d.amount, d.payload, d.payload->>'balance_side' IS NOT NULL AS from_gl, cj.id IS NOT NULL AS created_posted
 FROM gl_subledger_documents d LEFT JOIN posted_journals cj ON cj.id=d.created_journal_id
 WHERE d.company=$1 AND d.ledger='ap' AND ($5='' OR d.partner_code=$5) AND ($2='' OR d.branch_code=$2)
 AND CASE WHEN d.payload->>'balance_side' IS NOT NULL THEN d.side=1 ELSE d.side=2 END
 AND ($3='' OR d.id ILIKE $3 OR COALESCE(d.payload->>'doc_no', d.payload->>'document_no') ILIKE $3)
 AND (cardinality($7::text[])=0 OR d.id=ANY($7::text[]))
), used AS (
 SELECT s.payload->>'prepare_docno' AS prepare_docno, s.debt_id, SUM(s.amount) AS amt FROM active_settlements s
 WHERE s.payload->>'prepare_docno' IS NOT NULL GROUP BY 1,2
), reserved AS (
 SELECT l.debt_document_id, SUM(GREATEST(l.payamount-COALESCE(u.amt,0),0)) AS amt,
 MIN(l.docno) FILTER (WHERE l.payamount>COALESCE(u.amt,0)) AS prepare_docno, MIN(l.document_no) AS document_no
 FROM ap_payment_prepare_lines l JOIN ap_payment_prepare p ON p.company=l.company AND p.docno=l.docno AND NOT p.iscancel
 LEFT JOIN used u ON u.prepare_docno=l.docno AND u.debt_id=l.debt_document_id
 WHERE l.company=$1 AND l.docno<>$6 AND (cardinality($7::text[])=0 OR l.debt_document_id=ANY($7::text[])) GROUP BY 1
), posted_items AS (
 SELECT d.*, CASE WHEN d.from_gl THEN COALESCE(pa.amt,0) WHEN d.created_posted THEN d.amount ELSE 0 END AS posted, COALESCE(ps.amt,0) AS settled
 FROM debt d LEFT JOIN posted_alloc pa ON pa.document_id=d.id LEFT JOIN posted_settle ps ON ps.document_id=d.id
), open_items AS (
 SELECT p.*, p.posted-p.settled AS remaining FROM posted_items p
)`

// apOpenItemsSQL: $8 limit, $9 offset; ระบุ ids (โหมดตรวจของการบันทึก) = คืนทุกแถวที่ขอแม้ยอดคงเหลือ ≤ 0 เพื่อให้ผู้เรียกใช้สูตรเดียวกันทุกบรรทัด
const apOpenItemsSQL = apBalanceCTE + `
SELECT o.id, COALESCE(o.payload->>'doc_no', o.payload->>'document_no', o.id),
 COALESCE(NULLIF(o.payload->>'doc_date',''), o.payload->>'document_date')::date::text AS document_date,
 COALESCE(NULLIF(o.payload->>'due_date',''), NULLIF(o.payload->>'doc_date',''), o.payload->>'document_date')::date::text AS due_date,
 COALESCE(o.payload->>'description',''), COALESCE((o.payload->>'document_kind')::int,0), o.from_gl,
 o.amount::text, o.remaining::text, COALESCE(r.amt,0)::text, (o.remaining-COALESCE(r.amt,0))::text, COALESCE(su.amt,0)::text, COUNT(*) OVER()
FROM open_items o LEFT JOIN reserved r ON r.debt_document_id=o.id LEFT JOIN used su ON su.prepare_docno=$6 AND su.debt_id=o.id
WHERE o.remaining > 0 OR cardinality($7::text[]) > 0
ORDER BY due_date, document_date, 2, o.id
LIMIT $8 OFFSET $9`

// APOpenItems คือฟังก์ชันเดียวที่คำนวณหนี้เจ้าหนี้ค้างจ่าย/ยอดจอง/ยอดที่เตรียมจ่ายได้ — จอเตรียมจ่าย, ใบจ่ายชำระ และการตรวจ
// ยอดจองเรียกที่นี่ ห้ามเขียนสูตรซ้ำ; คืนแถวในหน้า + จำนวนทั้งหมด (ตาราง ap_payment_prepare* สร้างโดย schema ของเอกสาร ERP)
func APOpenItems(ctx context.Context, tx *sql.Tx, scope Scope, query APOpenQuery) ([]APOpenItem, int, error) {
	partner := strings.TrimSpace(query.Partner)
	if partner == "" {
		return nil, 0, fmt.Errorf("ap open items: partner required")
	}
	if !validDate(query.AsOf) {
		return nil, 0, fmt.Errorf("ap open items: invalid as-of date %q", query.AsOf)
	}
	limit, offset := query.Limit, query.Offset
	if limit < 1 || limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	q := SearchPattern(query.Q)
	ids := query.IDs
	if ids == nil {
		ids = []string{}
	}
	items, total, err := queryAPOpenItems(ctx, tx, scope.Company, scope.Branch, q, query.AsOf, partner, query.ExcludePrepare, ids, limit, offset)
	if err != nil || query.AsOf == reservationAsOf || len(items) == 0 {
		return items, total, err
	}
	// ใบจ่ายที่ลงวันที่หลัง as-of ตัดหนี้ไปแล้ว as-of มองไม่เห็น: ยอดที่ยังเตรียมจ่าย/จ่ายได้ต้องไม่เกินยอด ณ ปัจจุบันด้วย
	// (นับทุกการตัดหนี้ไม่ว่าลงวันที่ใด) — เพดานของบรรทัดเดิม (Available + PrepareUsed) จึงเป็นค่าที่ต่ำกว่าของสองช่วงเวลา
	nowIDs := make([]string, len(items))
	for i, item := range items {
		nowIDs[i] = item.ID
	}
	now, _, err := queryAPOpenItems(ctx, tx, scope.Company, scope.Branch, "", reservationAsOf, partner, query.ExcludePrepare, nowIDs, len(nowIDs), 0)
	if err != nil {
		return nil, 0, err
	}
	nowLimit := make(map[string]decimal.Decimal, len(now))
	for _, n := range now {
		nowLimit[n.ID] = n.Available.Decimal().Add(n.PrepareUsed.Decimal())
	}
	for i := range items {
		used := items[i].PrepareUsed.Decimal()
		if limit, ok := nowLimit[items[i].ID]; ok && items[i].Available.Decimal().Add(used).GreaterThan(limit) {
			items[i].Available = amountFromDecimal(limit.Sub(used))
		}
	}
	return items, total, nil
}

// queryAPOpenItems runs apOpenItemsSQL once and scans every row (rows are closed before the caller queries again on tx).
func queryAPOpenItems(ctx context.Context, tx *sql.Tx, company, branch, q, asOf, partner, excludePrepare string, ids []string, limit, offset int) ([]APOpenItem, int, error) {
	rows, err := tx.QueryContext(ctx, subledgerReadCTE+apOpenItemsSQL, company, branch, q, asOf, partner, excludePrepare, pq.Array(ids), limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("ap open items: %w", err)
	}
	defer rows.Close()
	items, total := []APOpenItem{}, 0
	for rows.Next() {
		var item APOpenItem
		if err = rows.Scan(&item.ID, &item.DocumentNo, &item.DocumentDate, &item.DueDate, &item.Description, &item.Kind, &item.FromGL,
			&item.Amount, &item.Remaining, &item.Reserved, &item.Available, &item.PrepareUsed, &total); err != nil {
			return nil, 0, fmt.Errorf("scan ap open item: %w", err)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

// maxSearchRunes คือความยาวคำค้นสูงสุด (ตัวอักษร ไม่ใช่ byte)
const maxSearchRunes = 200

// SearchPattern คือคำค้นแบบ ILIKE '%คำ%' (ค่าว่าง = ไม่กรอง) ยาวไม่เกิน maxSearchRunes ตัวอักษร —
// ตัดตาม rune เพราะการตัดตาม byte ทำให้อักษรไทยขาดกลางตัว แล้ว PostgreSQL ปฏิเสธข้อความ UTF-8 ที่ไม่ครบ
func SearchPattern(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return ""
	}
	if r := []rune(q); len(r) > maxSearchRunes {
		q = string(r[:maxSearchRunes])
	}
	return "%" + q + "%"
}

// APReservation คือหนี้เจ้าหนี้หนึ่งแถวที่ใบเตรียมจ่ายที่ยังไม่ยกเลิกยังจองยอดไว้ (สูตร apBalanceCTE นับทุกรายการถึงปัจจุบัน)
// Exists = แถวหนี้ยังอยู่ในบัญชีย่อย, Posted = ยอดหนี้ที่ผ่านบัญชี, PrepareNo = ใบเตรียมจ่ายแรกที่ยังจองอยู่
type APReservation struct {
	DebtID, DocumentNo, PrepareNo string
	Exists                        bool
	Posted, Remaining, Reserved   Amount
}

// apReservationsSQL: เฉพาะหนี้ที่ยังมียอดจอง (แถวหนี้ที่ถูกลบไปแล้วก็คืน — Exists=false); $8 limit, $9 offset
const apReservationsSQL = apBalanceCTE + `
SELECT r.debt_document_id, COALESCE(o.payload->>'doc_no', o.payload->>'document_no', r.document_no), COALESCE(r.prepare_docno,''),
 o.id IS NOT NULL, COALESCE(o.posted,0)::text, COALESCE(o.remaining,0)::text, r.amt::text
FROM reserved r LEFT JOIN open_items o ON o.id=r.debt_document_id
WHERE r.amt > 0
ORDER BY 1
LIMIT $8 OFFSET $9`

// reservationAsOf: ยอดจอง/ยอดคงเหลือ ณ ปัจจุบัน นับทุกรายการที่บันทึกแล้วไม่ว่าลงวันที่ใด
const reservationAsOf = "9999-12-31"

// APReservations คืนยอดจองของใบเตรียมจ่ายต่อหนี้ใน ids (ทุกเจ้าหนี้ ทุกสาขา) — ใช้ตรวจว่าการลบ/ลดยอดหนี้ไม่ทำให้ยอดที่จองไว้จ่ายไม่ได้
// (ฝั่งจอบัญชี: subledgerMutation.checkReservations, ฝั่งเอกสาร ERP: cleanSubledger)
func APReservations(ctx context.Context, tx *sql.Tx, company string, ids []string) (map[string]APReservation, error) {
	out := map[string]APReservation{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := tx.QueryContext(ctx, subledgerReadCTE+apReservationsSQL, company, "", "", reservationAsOf, "", "", pq.Array(ids), len(ids), 0)
	if err != nil {
		return nil, fmt.Errorf("ap reservations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var r APReservation
		if err = rows.Scan(&r.DebtID, &r.DocumentNo, &r.PrepareNo, &r.Exists, &r.Posted, &r.Remaining, &r.Reserved); err != nil {
			return nil, fmt.Errorf("scan ap reservation: %w", err)
		}
		out[r.DebtID] = r
	}
	return out, rows.Err()
}

// snapshotReservations จำยอดจองของหนี้เจ้าหนี้ที่ใบสำคัญนี้สร้างหรือจัดสรรยอดให้ ก่อนคำสั่งเปลี่ยนใบ — ยอดผ่านบัญชีของหนี้ลดได้
// ผ่านใบสำคัญที่สร้าง/จัดสรรหนี้นั้นเท่านั้น (กลับรายการ, ถอนการจัดสรร, ลบเอกสาร)
func (m *subledgerMutation) snapshotReservations() error {
	rows, err := m.tx.QueryContext(m.ctx, `SELECT DISTINCT l.debt_document_id FROM ap_payment_prepare_lines l
		JOIN ap_payment_prepare p ON p.company=l.company AND p.docno=l.docno AND NOT p.iscancel
		WHERE l.company=$1 AND (EXISTS(SELECT 1 FROM gl_subledger_documents d WHERE d.company=$1 AND d.id=l.debt_document_id AND d.created_journal_id=$2)
		OR EXISTS(SELECT 1 FROM gl_subledger_allocations a WHERE a.company=$1 AND a.document_id=l.debt_document_id AND a.journal_id=$2))
		ORDER BY 1`, m.scope.Company, m.journal.ID)
	if err != nil {
		return fmt.Errorf("read reserved ap debts: %w", err)
	}
	var ids []string
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return fmt.Errorf("scan reserved ap debt: %w", err)
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	m.reservedBefore, err = APReservations(m.ctx, m.tx, m.scope.Company, ids)
	return err
}

// checkReservations: หนี้ที่ใบเตรียมจ่ายยังจองไว้ต้องไม่ถูกลบ หรือถูกลดยอดผ่านบัญชีจนยอดคงเหลือต่ำกว่ายอดจอง (design ap-payment-prepare
// §4.1; ฝั่งเอกสาร ERP ตรวจใน cleanSubledger) — การตัดหนี้ด้วยเอกสารอื่นไม่บล็อก (design §6 ข้อ 10: ตรวจด้วย available สดเสมอ)
func (m *subledgerMutation) checkReservations() error {
	if len(m.reservedBefore) == 0 {
		return nil
	}
	ids := make([]string, 0, len(m.reservedBefore))
	for id := range m.reservedBefore {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	after, err := APReservations(m.ctx, m.tx, m.scope.Company, ids)
	if err != nil {
		return err
	}
	for _, id := range ids {
		now, reserved := after[id]
		if !reserved {
			continue
		}
		reduced := !now.Exists || now.Posted.Decimal().LessThan(m.reservedBefore[id].Posted.Decimal())
		if reduced && now.Remaining.Decimal().LessThan(now.Reserved.Decimal()) {
			return PrepareReservedError(now.DocumentNo, now.PrepareNo)
		}
	}
	return nil
}

// finish ตรวจยอดจองของใบเตรียมจ่ายแล้วคืนใบสำคัญอื่นที่เปลี่ยนตามคำสั่งนี้ — ทุกคำสั่งบัญชีย่อยจบที่นี่
func (m *subledgerMutation) finish() ([]Change, error) {
	if err := m.checkReservations(); err != nil {
		return nil, err
	}
	return m.relatedChanges()
}

// PrepareReservedError: เอกสารหนี้ {0} ยังถูกจองในใบเตรียมจ่าย {1} — ข้อความบอกวิธีแก้ (ยกเลิก/แก้ใบเตรียมจ่ายก่อน), 409
func PrepareReservedError(docNo, prepareNo string) *UserError {
	args := []string{docNo, prepareNo}
	message := language.Text("gl_err_ap_prepare_reserved", "th")
	for i, arg := range args {
		message = strings.ReplaceAll(message, "{"+strconv.Itoa(i)+"}", arg)
	}
	return &UserError{Code: "ap_prepare_reserved", Message: message, Args: args}
}

// DimensionExists บอกว่ารหัสแผนก/โครงการมีอยู่ในทะเบียนและยังใช้งาน (kind = DepartmentKind | ProjectKind)
func DimensionExists(ctx context.Context, tx *sql.Tx, company, kind, code string) (bool, error) {
	found, err := loadDimensionCodes(ctx, tx, company, kind, []string{code})
	if err != nil {
		return false, err
	}
	return found[code], nil
}
