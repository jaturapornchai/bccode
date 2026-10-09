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
		COALESCE(custname, ''), COALESCE(description, ''), COALESCE(remark, ''), created_at, modified_at
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
		COALESCE(custname, ''), COALESCE(description, ''), COALESCE(remark, ''), created_at, modified_at
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
