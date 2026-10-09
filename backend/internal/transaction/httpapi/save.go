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

// rowQuerier คือส่วนของ *sql.DB / *sql.Tx ที่การออกเลขที่เอกสารใช้ — ออกเลขในธุรกรรมเดียวกับการบันทึกเพื่อไม่ให้ซ้ำกัน
type rowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// serialized: เอกสารที่ลงบัญชีหรือเคลื่อนไหวสต็อกต้องคิวต่อกันทั้งบริษัท — สายสต็อกของสินค้าหนึ่งคำนวณจากแถวก่อนหน้า
// และการออกเลขที่ใบสำคัญต้องไม่แย่งกัน; ใช้ล็อกเดียวกับคำสั่งบัญชี (gl.LockCompanyTx) เพื่อไม่ให้เกิด deadlock ข้ามกัน
func serialized(dt doctype.DocType) bool {
	return doctype.PostsToGL(dt.Flag) || movesStock(dt)
}

// normalizeLines ให้ทุกบรรทัดมีเลขบรรทัดบวกและไม่ซ้ำกัน — เลขบรรทัดเป็นส่วนหนึ่งของกุญแจแถวสต็อก ซ้ำแล้วบันทึกไม่ได้
func normalizeLines(details []ErpDetailItem) {
	used := make(map[int]bool, len(details))
	for i := range details {
		n := details[i].LineNumber
		if n <= 0 {
			n = i + 1
		}
		for used[n] {
			n++
		}
		used[n] = true
		details[i].LineNumber = n
	}
}

// SaveDoc บันทึกเอกสารหนึ่งใบเป็นธุรกรรมเดียว: หัว/รายการ/การอ้างอิง + สต็อก + ใบสำคัญบัญชี + บัญชีย่อย
// ทั้งหมดสำเร็จหรือถอยกลับพร้อมกัน; แก้ไข = กลับรายการผลเดิมก่อนแล้วลงใหม่
func (s *Store) SaveDoc(ctx context.Context, db *sql.DB, scope gl.Scope, dt doctype.DocType, doc *ErpTransactionDoc, isEdit bool) (*ErpTransactionDoc, error) {
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
	company, branch, transflag, actor := scope.Company, scope.Branch, dt.Flag, scope.Actor

	now := time.Now().UTC()
	docDate := now
	if doc.DocDateTime != "" {
		if t, err := time.Parse(time.RFC3339, doc.DocDateTime); err == nil {
			docDate = t
		} else if t, err := time.Parse("2006-01-02", doc.DocDateTime); err == nil {
			docDate = t
		}
	}

	docNo := strings.TrimSpace(doc.DocNo)
	if docNo == "" {
		if isEdit {
			return nil, sql.ErrNoRows
		}
		generated, err := s.NextDocNo(ctx, tx, company, transflag, dt.Prefix)
		if err != nil {
			return nil, fmt.Errorf("generate docno: %w", err)
		}
		docNo = generated
	}
	doc.DocNo = docNo
	normalizeLines(doc.Details)

	var docID int
	if isEdit {
		var currentCancel bool
		var currentApproval string
		err := tx.QueryRowContext(ctx, `SELECT COALESCE(iscancel, false), COALESCE(approval_status, '')
			FROM doc WHERE businesscode = $1 AND transflag = $2 AND docno = $3 AND NOT COALESCE(isdelete, false) FOR UPDATE`,
			company, transflag, docNo).Scan(&currentCancel, &currentApproval)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("update doc %s: %w", docNo, sql.ErrNoRows)
		}
		if err != nil {
			return nil, fmt.Errorf("check doc status: %w", err)
		}
		if currentCancel || currentApproval == "approved" {
			return nil, docLockedError()
		}

		if doctype.PostsToGL(transflag) {
			if err = s.releasePosting(ctx, tx, scope, dt, docNo, reasonEdit, now); err != nil {
				return nil, err
			}
		}
		updateQuery := `UPDATE doc SET
			docdatetime = $1, custcode = $2, totalamount = $3,
			totalvalue_doc = $4, totaldiscount_doc = $5, totalvatvalue_doc = $6,
			totalbeforevat_doc = $7, totalaftervat_doc = $8,
			custname = $9, description = $10, remark = $11,
			branchid = $12, modifier_code = $13, modified_at = $14
			WHERE businesscode = $15 AND transflag = $16 AND docno = $17 AND NOT COALESCE(isdelete, false)
			RETURNING id`
		err = tx.QueryRowContext(ctx, updateQuery,
			docDate, doc.CustCode, doc.TotalAmount,
			doc.TotalValue, doc.TotalDiscount, doc.TotalVatValue,
			doc.TotalBeforeVat, doc.TotalAfterVat,
			doc.CustName, doc.Description, doc.Remark,
			branch, actor, now,
			company, transflag, docNo,
		).Scan(&docID)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("update doc %s: %w", docNo, sql.ErrNoRows)
		}
		if err != nil {
			return nil, fmt.Errorf("update doc: %w", err)
		}
		if _, err = tx.ExecContext(ctx, "DELETE FROM docdetail WHERE businesscode = $1 AND transflag = $2 AND docno = $3", company, transflag, docNo); err != nil {
			return nil, fmt.Errorf("delete doc details: %w", err)
		}
	} else {
		// เลขที่ที่เคยใช้แล้วใช้ซ้ำไม่ได้ แม้เอกสารนั้นถูกลบ — ใบสำคัญและสมุดสต็อกของมันยังเป็นหลักฐานอ้างเลขนี้อยู่
		var taken bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM doc WHERE businesscode = $1 AND transflag = $2 AND docno = $3)`, company, transflag, docNo).Scan(&taken); err != nil {
			return nil, fmt.Errorf("check docno: %w", err)
		}
		if taken {
			return nil, duplicateDocError(docNo)
		}
		insertQuery := `INSERT INTO doc (
			businesscode, transflag, docno, custcode, docdatetime,
			totalamount, totalvalue_doc, totaldiscount_doc, totalvatvalue_doc,
			totalbeforevat_doc, totalaftervat_doc, branchid,
			custname, description, remark,
			creator_code, created_at, modifier_code, modified_at, checksum
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19, ''
		) RETURNING id`
		err = tx.QueryRowContext(ctx, insertQuery,
			company, transflag, docNo, doc.CustCode, docDate,
			doc.TotalAmount, doc.TotalValue, doc.TotalDiscount, doc.TotalVatValue,
			doc.TotalBeforeVat, doc.TotalAfterVat, branch,
			doc.CustName, doc.Description, doc.Remark,
			actor, now, actor, now,
		).Scan(&docID)
		if err != nil {
			return nil, fmt.Errorf("insert doc: %w", err)
		}
	}

	insertDetailQuery := `INSERT INTO docdetail (
		businesscode, transflag, docno, docdatetime, linenumber,
		itemcode, barcode, unitcode, totalqty, price, sumamount,
		description, whcode, locationcode, docref
	) VALUES (
		$1, $2, $3, $4, $5,
		$6, $7, $8, $9, $10, $11,
		$12, $13, $14, $15
	)`
	docRefGlobal := strings.TrimSpace(doc.DocRef)
	for _, item := range doc.Details {
		desc := item.Description
		if desc == "" {
			desc = item.ItemName
		}
		itemRef := strings.TrimSpace(item.DocRef)
		if itemRef == "" {
			itemRef = docRefGlobal
		}
		if _, err = tx.ExecContext(ctx, insertDetailQuery,
			company, transflag, docNo, docDate, item.LineNumber,
			item.ItemCode, item.Barcode, item.UnitCode, item.Qty, item.Price, item.SumAmount,
			desc, item.WhCode, item.LocationCode, itemRef,
		); err != nil {
			return nil, fmt.Errorf("insert docdetail line %d: %w", item.LineNumber, err)
		}
	}

	// บันทึกความสัมพันธ์การอ้างอิงเอกสาร (Document Reference Linkage)
	if docRefGlobal != "" {
		refTransFlag := doc.DocRefTransFlag
		if refTransFlag == 0 {
			if refType, ok := doctype.ByDocNo(docRefGlobal); ok {
				refTransFlag = refType.Flag
			}
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM docref WHERE businesscode=$1 AND docno=$2`, company, docNo); err != nil {
			return nil, fmt.Errorf("reset docref: %w", err)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO docref (businesscode, docno, docnotransflag, docnoref, docnoreftransflag)
			VALUES ($1, $2, $3, $4, $5)`, company, docNo, transflag, docRefGlobal, refTransFlag); err != nil {
			return nil, fmt.Errorf("insert docref: %w", err)
		}
		// อัปเดตเอกสารต้นทาง (PO, SO, QT, INV, PUR) ให้เป็นอ้างอิงแล้วและปิดเอกสาร
		if _, err = tx.ExecContext(ctx, `UPDATE doc SET isref = true, isclosed = true, modified_at = $1
			WHERE businesscode = $2 AND docno = $3`, now, company, docRefGlobal); err != nil {
			return nil, fmt.Errorf("close referenced doc: %w", err)
		}
	}

	if movesStock(dt) {
		if err = s.replaceStockMovement(ctx, tx, company, dt, doc, docNo, docDate, now); err != nil {
			return nil, err
		}
	}
	posted := postingResult{}
	if doctype.PostsToGL(transflag) {
		if posted, err = s.postDocument(ctx, tx, scope, dt, doc, docNo, docDate, now); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	doc.ID = strconv.Itoa(docID)
	doc.DocDateTime = docDate.Format(time.RFC3339)
	doc.TransFlag = transflag
	doc.BranchCode = branch
	doc.GLStatus, doc.GLDocNo = posted.status, posted.docNo
	return doc, nil
}

// DeleteDoc ลบเอกสาร (ตั้งธง isdelete) พร้อมกลับรายการใบสำคัญ เอาแถวบัญชีย่อยของเอกสารออก และคำนวณสายสต็อกใหม่
// เอกสารที่ถูกตัดหนี้/จัดสรรแล้วลบไม่ได้ (erp_doc_settled); ไม่พบหรือถูกลบไปแล้ว = sql.ErrNoRows
func (s *Store) DeleteDoc(ctx context.Context, db *sql.DB, scope gl.Scope, dt doctype.DocType, docIdOrNo string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	if serialized(dt) {
		if err = gl.LockCompanyTx(ctx, tx, scope.Company); err != nil {
			return ledgerFail(err)
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
	var currentCancel bool
	var currentApproval string
	query := fmt.Sprintf("SELECT docno, COALESCE(iscancel, false), COALESCE(approval_status, '') FROM doc %s FOR UPDATE", where)
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&docNo, &currentCancel, &currentApproval); err != nil {
		return err
	}
	if currentCancel || currentApproval == "approved" {
		return docLockedError()
	}

	// Reject if any live downstream document exists
	var hasDownstream bool
	checkDownstreamQuery := `SELECT EXISTS(
		SELECT 1 FROM docref r
		JOIN doc d ON d.businesscode = r.businesscode AND d.docno = r.docno AND d.transflag = r.docnotransflag
		WHERE r.businesscode = $1 AND r.docnoref = $2 AND (r.docnoreftransflag = $3 OR r.docnoreftransflag = 0)
		  AND NOT COALESCE(d.isdelete, false) AND NOT COALESCE(d.iscancel, false)
	)`
	if err := tx.QueryRowContext(ctx, checkDownstreamQuery, scope.Company, docNo, dt.Flag).Scan(&hasDownstream); err != nil {
		return fmt.Errorf("check downstream: %w", err)
	}
	if hasDownstream {
		return docHasDownstreamError()
	}

	now := time.Now().UTC()
	if doctype.PostsToGL(dt.Flag) {
		if err = s.releasePosting(ctx, tx, scope, dt, docNo, reasonDelete, now); err != nil {
			return err
		}
	}
	if movesStock(dt) {
		if err = s.replaceStockMovement(ctx, tx, scope.Company, dt, nil, docNo, time.Time{}, now); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE doc SET isdelete = true, modifier_code = $1, modified_at = $2
		WHERE businesscode = $3 AND transflag = $4 AND docno = $5`, scope.Actor, now, scope.Company, dt.Flag, docNo); err != nil {
		return fmt.Errorf("delete doc: %w", err)
	}
	return tx.Commit()
}
