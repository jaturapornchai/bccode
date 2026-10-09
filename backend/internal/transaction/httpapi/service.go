package httpapi

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
)

type Store struct{}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) ListDocs(ctx context.Context, db *sql.DB, company string, transflag int, q string, status string, offset, limit int) ([]ErpTransactionDoc, int, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	whereClause := "WHERE businesscode = $1 AND transflag = $2 AND NOT COALESCE(isdelete, false)"
	args := []any{company, transflag}
	argIdx := 3

	if strings.TrimSpace(q) != "" {
		search := "%" + strings.TrimSpace(q) + "%"
		whereClause += fmt.Sprintf(" AND (docno ILIKE $%d OR custcode ILIKE $%d)", argIdx, argIdx)
		args = append(args, search)
		argIdx++
	}

	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "pending":
		whereClause += " AND NOT COALESCE(iscancel, false) AND COALESCE(approval_status, '') = ''"
	case "approved":
		whereClause += " AND NOT COALESCE(iscancel, false) AND COALESCE(approval_status, '') = 'approved'"
	case "cancelled":
		whereClause += " AND COALESCE(iscancel, false)"
	case "active":
		whereClause += " AND NOT COALESCE(iscancel, false)"
	}

	countQuery := "SELECT count(*) FROM doc " + whereClause
	var total int
	if err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count docs: %w", err)
	}

	query := fmt.Sprintf(`SELECT id, docno, COALESCE(to_char(docdatetime, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), ''),
		transflag, COALESCE(custcode, ''), COALESCE(totalamount, 0),
		COALESCE(totalvalue_doc, 0), COALESCE(totaldiscount_doc, 0),
		COALESCE(totalvatvalue_doc, 0), COALESCE(totalbeforevat_doc, 0),
		COALESCE(totalaftervat_doc, 0), COALESCE(iscancel, false),
		COALESCE(isclosed, false), COALESCE(branchid, ''),
		COALESCE(custname, ''), COALESCE(description, ''), COALESCE(remark, ''), created_at, modified_at,
		COALESCE(approval_status, ''), COALESCE(approved_by, ''), approved_at,
		COALESCE(cancelled_by, ''), cancelled_at, COALESCE(cancelreason, '')
		FROM doc %s ORDER BY docdatetime DESC, docno DESC LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query docs: %w", err)
	}
	defer rows.Close()

	var docs []ErpTransactionDoc
	for rows.Next() {
		var d ErpTransactionDoc
		var idInt int
		var createdAt, modifiedAt *time.Time
		if err := rows.Scan(
			&idInt, &d.DocNo, &d.DocDateTime,
			&d.TransFlag, &d.CustCode, &d.TotalAmount,
			&d.TotalValue, &d.TotalDiscount,
			&d.TotalVatValue, &d.TotalBeforeVat,
			&d.TotalAfterVat, &d.IsCancel,
			&d.IsClosed, &d.BranchCode,
			&d.CustName, &d.Description, &d.Remark, &createdAt, &modifiedAt,
			&d.ApprovalStatus, &d.ApprovedBy, &d.ApprovedAt,
			&d.CancelledBy, &d.CancelledAt, &d.CancelReason,
		); err != nil {
			return nil, 0, fmt.Errorf("scan doc row: %w", err)
		}
		d.ID = strconv.Itoa(idInt)
		d.CreatedAt = createdAt
		d.ModifiedAt = modifiedAt
		if d.IsCancel {
			d.Status = 2
		} else if d.ApprovalStatus == "approved" {
			d.Status = 1
		} else {
			d.Status = 0
		}
		d.Details = []ErpDetailItem{}
		docs = append(docs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if docs == nil {
		docs = []ErpTransactionDoc{}
	}
	return docs, total, nil
}

func (s *Store) GetDoc(ctx context.Context, db *sql.DB, company string, transflag int, docIdOrNo string) (*ErpTransactionDoc, error) {
	where := "WHERE businesscode = $1 AND transflag = $2 AND NOT COALESCE(isdelete, false) AND "
	var args []any
	if idNum, err := strconv.Atoi(docIdOrNo); err == nil && idNum > 0 {
		where += "(id = $3 OR docno = $4)"
		args = []any{company, transflag, idNum, docIdOrNo}
	} else {
		where += "docno = $3"
		args = []any{company, transflag, docIdOrNo}
	}

	query := fmt.Sprintf(`SELECT id, docno, COALESCE(to_char(docdatetime, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), ''),
		transflag, COALESCE(custcode, ''), COALESCE(totalamount, 0),
		COALESCE(totalvalue_doc, 0), COALESCE(totaldiscount_doc, 0),
		COALESCE(totalvatvalue_doc, 0), COALESCE(totalbeforevat_doc, 0),
		COALESCE(totalaftervat_doc, 0), COALESCE(iscancel, false),
		COALESCE(isclosed, false), COALESCE(branchid, ''),
		COALESCE(custname, ''), COALESCE(description, ''), COALESCE(remark, ''), created_at, modified_at,
		COALESCE(approval_status, ''), COALESCE(approved_by, ''), approved_at,
		COALESCE(cancelled_by, ''), cancelled_at, COALESCE(cancelreason, '')
		FROM doc %s LIMIT 1`, where)

	var d ErpTransactionDoc
	var idInt int
	var createdAt, modifiedAt *time.Time
	err := db.QueryRowContext(ctx, query, args...).Scan(
		&idInt, &d.DocNo, &d.DocDateTime,
		&d.TransFlag, &d.CustCode, &d.TotalAmount,
		&d.TotalValue, &d.TotalDiscount,
		&d.TotalVatValue, &d.TotalBeforeVat,
		&d.TotalAfterVat, &d.IsCancel,
		&d.IsClosed, &d.BranchCode,
		&d.CustName, &d.Description, &d.Remark, &createdAt, &modifiedAt,
		&d.ApprovalStatus, &d.ApprovedBy, &d.ApprovedAt,
		&d.CancelledBy, &d.CancelledAt, &d.CancelReason,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get doc: %w", err)
	}
	d.ID = strconv.Itoa(idInt)
	d.CreatedAt = createdAt
	d.ModifiedAt = modifiedAt
	if d.IsCancel {
		d.Status = 2
	} else if d.ApprovalStatus == "approved" {
		d.Status = 1
	} else {
		d.Status = 0
	}

	// Fetch docref if any
	var refDocNo string
	var refTransFlag int
	if err := db.QueryRowContext(ctx, `SELECT docnoref, docnoreftransflag FROM docref WHERE businesscode=$1 AND docno=$2 LIMIT 1`, company, d.DocNo).Scan(&refDocNo, &refTransFlag); err == nil {
		d.DocRef = refDocNo
		d.DocRefTransFlag = refTransFlag
	}

	// Fetch detail items
	detailQuery := `SELECT id, linenumber, COALESCE(itemcode, ''), COALESCE(barcode, ''),
		COALESCE(unitcode, ''), COALESCE(totalqty, 0), COALESCE(price, 0),
		COALESCE(sumamount, 0), COALESCE(description, ''), COALESCE(whcode, ''), COALESCE(locationcode, ''),
		COALESCE(docref, '')
		FROM docdetail WHERE businesscode = $1 AND transflag = $2 AND docno = $3
		ORDER BY linenumber`

	rows, err := db.QueryContext(ctx, detailQuery, company, transflag, d.DocNo)
	if err != nil {
		return nil, fmt.Errorf("get doc details: %w", err)
	}
	defer rows.Close()

	d.Details = []ErpDetailItem{}
	for rows.Next() {
		var item ErpDetailItem
		if err := rows.Scan(
			&item.ID, &item.LineNumber, &item.ItemCode, &item.Barcode,
			&item.UnitCode, &item.Qty, &item.Price,
			&item.SumAmount, &item.Description, &item.WhCode, &item.LocationCode,
			&item.DocRef,
		); err != nil {
			return nil, fmt.Errorf("scan doc detail: %w", err)
		}
		item.ItemName = item.Description
		d.Details = append(d.Details, item)
	}
	return &d, nil
}

func (s *Store) NextDocNo(ctx context.Context, db rowQuerier, company string, transflag int, prefix string) (string, error) {
	now := time.Now()
	thaiYear2D := (now.Year() + 543) % 100
	periodPrefix := fmt.Sprintf("%s%02d%02d", prefix, thaiYear2D, int(now.Month()))

	var maxDocNo sql.NullString
	query := `SELECT docno FROM doc WHERE businesscode = $1 AND transflag = $2 AND docno LIKE $3 ORDER BY docno DESC LIMIT 1`
	err := db.QueryRowContext(ctx, query, company, transflag, periodPrefix+"-%").Scan(&maxDocNo)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	nextSeq := 1
	if maxDocNo.Valid && maxDocNo.String != "" {
		parts := strings.Split(maxDocNo.String, "-")
		if len(parts) >= 2 {
			if seq, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
				nextSeq = seq + 1
			}
		}
	}
	return fmt.Sprintf("%s-%04d", periodPrefix, nextSeq), nil
}

// isApprovableModule ตรวจสอบประเภทเอกสารที่มีขั้นตอนการอนุมัติ (PR, PO, QT, SO)
func isApprovableModule(transflag int) bool {
	switch transflag {
	case 2, 6, 34, 36:
		return true
	default:
		return false
	}
}

// isCancellableModule ตรวจสอบประเภทเอกสารที่ยกเลิกได้ (รองรับทุกโมดูลในระบบ)
func isCancellableModule(transflag int) bool {
	_, ok := doctype.ByFlag(transflag)
	return ok
}

// ApproveDoc อนุมัติเอกสารหนึ่งใบ (เปลี่ยนสถานะจาก '' เป็น 'approved' พร้อมบันทึกผู้/เวลาอนุมัติ และลง audit log)
func (s *Store) ApproveDoc(ctx context.Context, db *sql.DB, scope gl.Scope, dt doctype.DocType, docIdOrNo string) (*ErpTransactionDoc, error) {
	if !isApprovableModule(dt.Flag) {
		return nil, docNotPendingError()
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	where := "WHERE businesscode = $1 AND transflag = $2 AND NOT COALESCE(isdelete, false) AND "
	args := []any{scope.Company, dt.Flag}
	if idNum, err := strconv.Atoi(docIdOrNo); err == nil && idNum > 0 {
		where += "(id = $3 OR docno = $4)"
		args = append(args, idNum, docIdOrNo)
	} else {
		where += "docno = $3"
		args = append(args, docIdOrNo)
	}

	var docNo string
	var isCancel bool
	var approvalStatus string
	query := fmt.Sprintf("SELECT docno, COALESCE(iscancel, false), COALESCE(approval_status, '') FROM doc %s FOR UPDATE", where)
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&docNo, &isCancel, &approvalStatus); err != nil {
		return nil, err
	}

	if isCancel {
		return nil, docAlreadyCancelledError()
	}
	if approvalStatus != "" {
		return nil, docNotPendingError()
	}

	now := time.Now().UTC()
	actor := scope.Actor
	updateQuery := `UPDATE doc SET
		approval_status = 'approved',
		approved_by = $1,
		approved_at = $2,
		modifier_code = $1,
		modified_at = $2
		WHERE businesscode = $3 AND transflag = $4 AND docno = $5`
	if _, err := tx.ExecContext(ctx, updateQuery, actor, now, scope.Company, dt.Flag, docNo); err != nil {
		return nil, fmt.Errorf("update doc approval: %w", err)
	}

	eventQuery := `INSERT INTO erp_doc_events (
		businesscode, transflag, docno, action, from_status, to_status, reason, actor, created_at
	) VALUES ($1, $2, $3, 'approve', '', 'approved', '', $4, $5)`
	if _, err := tx.ExecContext(ctx, eventQuery, scope.Company, dt.Flag, docNo, actor, now); err != nil {
		return nil, fmt.Errorf("record doc event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return s.GetDoc(ctx, db, scope.Company, dt.Flag, docNo)
}

// UnapproveDoc ยกเลิกการอนุมัติเอกสาร (เปลี่ยนสถานะจาก 'approved' กลับเป็น '')
func (s *Store) UnapproveDoc(ctx context.Context, db *sql.DB, scope gl.Scope, dt doctype.DocType, docIdOrNo string, reason string) (*ErpTransactionDoc, error) {
	if !isApprovableModule(dt.Flag) {
		return nil, docNotApprovedError()
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	where := "WHERE businesscode = $1 AND transflag = $2 AND NOT COALESCE(isdelete, false) AND "
	args := []any{scope.Company, dt.Flag}
	if idNum, err := strconv.Atoi(docIdOrNo); err == nil && idNum > 0 {
		where += "(id = $3 OR docno = $4)"
		args = append(args, idNum, docIdOrNo)
	} else {
		where += "docno = $3"
		args = append(args, docIdOrNo)
	}

	var docNo string
	var isCancel bool
	var approvalStatus string
	query := fmt.Sprintf("SELECT docno, COALESCE(iscancel, false), COALESCE(approval_status, '') FROM doc %s FOR UPDATE", where)
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&docNo, &isCancel, &approvalStatus); err != nil {
		return nil, err
	}

	if isCancel {
		return nil, docAlreadyCancelledError()
	}
	if approvalStatus != "approved" {
		return nil, docNotApprovedError()
	}

	// ตรวจสอบเอกสารปลายทางที่มีชีวิตอยู่
	var hasDownstream bool
	checkDownstreamQuery := `SELECT EXISTS(
		SELECT 1 FROM docref r
		JOIN doc d ON d.businesscode = r.businesscode AND d.docno = r.docno AND d.transflag = r.docnotransflag
		WHERE r.businesscode = $1 AND r.docnoref = $2 AND (r.docnoreftransflag = $3 OR r.docnoreftransflag = 0)
		  AND NOT COALESCE(d.isdelete, false) AND NOT COALESCE(d.iscancel, false)
	)`
	if err := tx.QueryRowContext(ctx, checkDownstreamQuery, scope.Company, docNo, dt.Flag).Scan(&hasDownstream); err != nil {
		return nil, fmt.Errorf("check downstream: %w", err)
	}
	if hasDownstream {
		return nil, docHasDownstreamError()
	}

	now := time.Now().UTC()
	actor := scope.Actor
	updateQuery := `UPDATE doc SET
		approval_status = '',
		approved_by = '',
		approved_at = NULL,
		modifier_code = $1,
		modified_at = $2
		WHERE businesscode = $3 AND transflag = $4 AND docno = $5`
	if _, err := tx.ExecContext(ctx, updateQuery, actor, now, scope.Company, dt.Flag, docNo); err != nil {
		return nil, fmt.Errorf("update doc unapprove: %w", err)
	}

	eventQuery := `INSERT INTO erp_doc_events (
		businesscode, transflag, docno, action, from_status, to_status, reason, actor, created_at
	) VALUES ($1, $2, $3, 'unapprove', 'approved', '', $4, $5, $6)`
	if _, err := tx.ExecContext(ctx, eventQuery, scope.Company, dt.Flag, docNo, strings.TrimSpace(reason), actor, now); err != nil {
		return nil, fmt.Errorf("record doc event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return s.GetDoc(ctx, db, scope.Company, dt.Flag, docNo)
}

// CancelDoc ยกเลิกเอกสารทั้งฉบับ (ต้องระบุเหตุผล)
func (s *Store) CancelDoc(ctx context.Context, db *sql.DB, scope gl.Scope, dt doctype.DocType, docIdOrNo string, reason string) (*ErpTransactionDoc, error) {
	if !isCancellableModule(dt.Flag) {
		return nil, docAlreadyCancelledError()
	}
	trimmedReason := strings.TrimSpace(reason)
	if trimmedReason == "" {
		return nil, cancelReasonRequiredError()
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if serialized(dt) {
		if err = gl.LockCompanyTx(ctx, tx, scope.Company); err != nil {
			return nil, ledgerFail(err)
		}
	}

	where := "WHERE businesscode = $1 AND transflag = $2 AND NOT COALESCE(isdelete, false) AND "
	args := []any{scope.Company, dt.Flag}
	if idNum, err := strconv.Atoi(docIdOrNo); err == nil && idNum > 0 {
		where += "(id = $3 OR docno = $4)"
		args = append(args, idNum, docIdOrNo)
	} else {
		where += "docno = $3"
		args = append(args, docIdOrNo)
	}

	var docNo string
	var isCancel bool
	var approvalStatus string
	query := fmt.Sprintf("SELECT docno, COALESCE(iscancel, false), COALESCE(approval_status, '') FROM doc %s FOR UPDATE", where)
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&docNo, &isCancel, &approvalStatus); err != nil {
		return nil, err
	}

	if isCancel {
		return nil, docAlreadyCancelledError()
	}

	// PR cancel ต้องเป็นสถานะรออนุมัติ (ถ้าอนุมัติแล้วต้องยกเลิกอนุมัติก่อน)
	if dt.Flag == 2 && approvalStatus != "" {
		return nil, docNotPendingError()
	}

	// ตรวจสอบเอกสารปลายทางที่มีชีวิตอยู่
	var hasDownstream bool
	checkDownstreamQuery := `SELECT EXISTS(
		SELECT 1 FROM docref r
		JOIN doc d ON d.businesscode = r.businesscode AND d.docno = r.docno AND d.transflag = r.docnotransflag
		WHERE r.businesscode = $1 AND r.docnoref = $2 AND (r.docnoreftransflag = $3 OR r.docnoreftransflag = 0)
		  AND NOT COALESCE(d.isdelete, false) AND NOT COALESCE(d.iscancel, false)
	)`
	if err := tx.QueryRowContext(ctx, checkDownstreamQuery, scope.Company, docNo, dt.Flag).Scan(&hasDownstream); err != nil {
		return nil, fmt.Errorf("check downstream: %w", err)
	}
	if hasDownstream {
		return nil, docHasDownstreamError()
	}

	now := time.Now().UTC()
	if doctype.PostsToGL(dt.Flag) {
		if err = s.releasePosting(ctx, tx, scope, dt, docNo, reasonCancel, now); err != nil {
			return nil, err
		}
	}
	if movesStock(dt) {
		if err = s.replaceStockMovement(ctx, tx, scope.Company, dt, nil, docNo, time.Time{}, now); err != nil {
			return nil, err
		}
	}

	actor := scope.Actor
	updateQuery := `UPDATE doc SET
		iscancel = true,
		cancelreason = $1,
		cancelled_by = $2,
		cancelled_at = $3,
		modifier_code = $2,
		modified_at = $3
		WHERE businesscode = $4 AND transflag = $5 AND docno = $6`
	if _, err := tx.ExecContext(ctx, updateQuery, trimmedReason, actor, now, scope.Company, dt.Flag, docNo); err != nil {
		return nil, fmt.Errorf("update doc cancel: %w", err)
	}

	eventQuery := `INSERT INTO erp_doc_events (
		businesscode, transflag, docno, action, from_status, to_status, reason, actor, created_at
	) VALUES ($1, $2, $3, 'cancel', $4, 'cancelled', $5, $6, $7)`
	if _, err := tx.ExecContext(ctx, eventQuery, scope.Company, dt.Flag, docNo, approvalStatus, trimmedReason, actor, now); err != nil {
		return nil, fmt.Errorf("record doc event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return s.GetDoc(ctx, db, scope.Company, dt.Flag, docNo)
}

// ApproveBatchDocs อนุมัติเอกสารหลายใบพร้อมกันแบบทีละใบ (ไม่ล้มเหลวยกชุดตามแบบ Champ)
func (s *Store) ApproveBatchDocs(ctx context.Context, db *sql.DB, scope gl.Scope, dt doctype.DocType, ids []string) ([]BatchApproveItemResult, error) {
	results := make([]BatchApproveItemResult, 0, len(ids))
	for _, id := range ids {
		trimmedID := strings.TrimSpace(id)
		if trimmedID == "" {
			continue
		}
		_, err := s.ApproveDoc(ctx, db, scope, dt, trimmedID)
		if err == nil {
			results = append(results, BatchApproveItemResult{
				ID:      trimmedID,
				Success: true,
			})
		} else {
			code := "error"
			msg := err.Error()
			if uerr, isUser := gl.AsUserError(err); isUser {
				code = uerr.Code
				msg = uerr.Message
			}
			results = append(results, BatchApproveItemResult{
				ID:      trimmedID,
				Success: false,
				Code:    code,
				Message: msg,
			})
		}
	}
	return results, nil
}

// GetDocEvents ดึงประวัติการอนุมัติ/ยกเลิกของเอกสาร
func (s *Store) GetDocEvents(ctx context.Context, db *sql.DB, company string, transflag int, docIdOrNo string) ([]ErpDocEvent, error) {
	doc, err := s.GetDoc(ctx, db, company, transflag, docIdOrNo)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, sql.ErrNoRows
	}

	query := `SELECT id, businesscode, transflag, docno, action, from_status, to_status, reason, actor, created_at
		FROM erp_doc_events
		WHERE businesscode = $1 AND transflag = $2 AND docno = $3
		ORDER BY id ASC`
	rows, err := db.QueryContext(ctx, query, company, transflag, doc.DocNo)
	if err != nil {
		return nil, fmt.Errorf("query doc events: %w", err)
	}
	defer rows.Close()

	var events []ErpDocEvent
	for rows.Next() {
		var ev ErpDocEvent
		if err := rows.Scan(
			&ev.ID, &ev.BusinessCode, &ev.TransFlag, &ev.DocNo,
			&ev.Action, &ev.FromStatus, &ev.ToStatus, &ev.Reason,
			&ev.Actor, &ev.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan doc event: %w", err)
		}
		events = append(events, ev)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if events == nil {
		events = []ErpDocEvent{}
	}
	return events, nil
}
