package httpapi

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Store struct{}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) ListDocs(ctx context.Context, db *sql.DB, company string, transflag int, q string, offset, limit int) ([]ErpTransactionDoc, int, error) {
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
		COALESCE(cancelreason, ''), created_at, modified_at
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
			&d.Remark, &createdAt, &modifiedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan doc row: %w", err)
		}
		d.ID = strconv.Itoa(idInt)
		d.CreatedAt = createdAt
		d.ModifiedAt = modifiedAt
		if d.IsCancel {
			d.Status = 2
		} else if d.IsClosed {
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
		COALESCE(cancelreason, ''), created_at, modified_at
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
		&d.Remark, &createdAt, &modifiedAt,
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
	} else if d.IsClosed {
		d.Status = 1
	} else {
		d.Status = 0
	}

	// Fetch detail items
	detailQuery := `SELECT id, linenumber, COALESCE(itemcode, ''), COALESCE(barcode, ''),
		COALESCE(unitcode, ''), COALESCE(totalqty, 0), COALESCE(price, 0),
		COALESCE(sumamount, 0), COALESCE(description, ''), COALESCE(whcode, ''), COALESCE(locationcode, '')
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
		); err != nil {
			return nil, fmt.Errorf("scan doc detail: %w", err)
		}
		item.ItemName = item.Description
		d.Details = append(d.Details, item)
	}
	return &d, nil
}

func (s *Store) NextDocNo(ctx context.Context, db *sql.DB, company string, transflag int, prefix string) (string, error) {
	now := time.Now()
	// ปี พ.ศ. สองหลัก + เดือน เช่น 6910
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

func (s *Store) SaveDoc(ctx context.Context, db *sql.DB, company, branch string, transflag int, prefix string, doc *ErpTransactionDoc, actor string, isEdit bool) (*ErpTransactionDoc, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

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
		generated, err := s.NextDocNo(ctx, db, company, transflag, prefix)
		if err != nil {
			return nil, fmt.Errorf("generate docno: %w", err)
		}
		docNo = generated
		doc.DocNo = docNo
	}

	if isEdit {
		// Update header
		updateQuery := `UPDATE doc SET
			docdatetime = $1, custcode = $2, totalamount = $3,
			totalvalue_doc = $4, totaldiscount_doc = $5, totalvatvalue_doc = $6,
			totalbeforevat_doc = $7, totalaftervat_doc = $8,
			cancelreason = $9, branchid = $10, modifier_code = $11, modified_at = $12
			WHERE businesscode = $13 AND transflag = $14 AND docno = $15`

		_, err = tx.ExecContext(ctx, updateQuery,
			docDate, doc.CustCode, doc.TotalAmount,
			doc.TotalValue, doc.TotalDiscount, doc.TotalVatValue,
			doc.TotalBeforeVat, doc.TotalAfterVat,
			doc.Remark, branch, actor, now,
			company, transflag, docNo,
		)
		if err != nil {
			return nil, fmt.Errorf("update doc: %w", err)
		}

		// Delete existing details
		_, err = tx.ExecContext(ctx, "DELETE FROM docdetail WHERE businesscode = $1 AND transflag = $2 AND docno = $3", company, transflag, docNo)
		if err != nil {
			return nil, fmt.Errorf("delete doc details: %w", err)
		}
	} else {
		// Insert header
		insertQuery := `INSERT INTO doc (
			businesscode, transflag, docno, custcode, docdatetime,
			totalamount, totalvalue_doc, totaldiscount_doc, totalvatvalue_doc,
			totalbeforevat_doc, totalaftervat_doc, branchid, cancelreason,
			creator_code, created_at, modifier_code, modified_at, checksum
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16, $17, ''
		)`

		_, err = tx.ExecContext(ctx, insertQuery,
			company, transflag, docNo, doc.CustCode, docDate,
			doc.TotalAmount, doc.TotalValue, doc.TotalDiscount, doc.TotalVatValue,
			doc.TotalBeforeVat, doc.TotalAfterVat, branch, doc.Remark,
			actor, now, actor, now,
		)
		if err != nil {
			return nil, fmt.Errorf("insert doc: %w", err)
		}
	}

	// Insert details
	insertDetailQuery := `INSERT INTO docdetail (
		businesscode, transflag, docno, docdatetime, linenumber,
		itemcode, barcode, unitcode, totalqty, price, sumamount,
		description, whcode, locationcode
	) VALUES (
		$1, $2, $3, $4, $5,
		$6, $7, $8, $9, $10, $11,
		$12, $13, $14
	)`

	for i, item := range doc.Details {
		lineNo := item.LineNumber
		if lineNo <= 0 {
			lineNo = i + 1
		}
		desc := item.Description
		if desc == "" {
			desc = item.ItemName
		}
		_, err = tx.ExecContext(ctx, insertDetailQuery,
			company, transflag, docNo, docDate, lineNo,
			item.ItemCode, item.Barcode, item.UnitCode, item.Qty, item.Price, item.SumAmount,
			desc, item.WhCode, item.LocationCode,
		)
		if err != nil {
			return nil, fmt.Errorf("insert docdetail line %d: %w", lineNo, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	doc.DocDateTime = docDate.Format(time.RFC3339)
	doc.TransFlag = transflag
	doc.BranchCode = branch
	return doc, nil
}

func (s *Store) DeleteDoc(ctx context.Context, db *sql.DB, company string, transflag int, docIdOrNo string, actor string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	where := "WHERE businesscode = $1 AND transflag = $2 AND "
	var args []any
	if idNum, err := strconv.Atoi(docIdOrNo); err == nil && idNum > 0 {
		where += "(id = $3 OR docno = $4)"
		args = []any{company, transflag, idNum, docIdOrNo}
	} else {
		where += "docno = $3"
		args = []any{company, transflag, docIdOrNo}
	}

	// Soft delete
	updateDoc := fmt.Sprintf("UPDATE doc SET isdelete = true, modifier_code = $%d, modified_at = $%d %s", len(args)+1, len(args)+2, where)
	args = append(args, actor, time.Now().UTC())

	res, err := tx.ExecContext(ctx, updateDoc, args...)
	if err != nil {
		return fmt.Errorf("delete doc: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}
