package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
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

func (s *Store) NextDocNo(ctx context.Context, db *sql.DB, company string, transflag int, prefix string) (string, error) {
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
		description, whcode, locationcode, docref
	) VALUES (
		$1, $2, $3, $4, $5,
		$6, $7, $8, $9, $10, $11,
		$12, $13, $14, $15
	)`

	docRefGlobal := strings.TrimSpace(doc.DocRef)
	for i, item := range doc.Details {
		lineNo := item.LineNumber
		if lineNo <= 0 {
			lineNo = i + 1
		}
		desc := item.Description
		if desc == "" {
			desc = item.ItemName
		}
		itemRef := strings.TrimSpace(item.DocRef)
		if itemRef == "" {
			itemRef = docRefGlobal
		}
		_, err = tx.ExecContext(ctx, insertDetailQuery,
			company, transflag, docNo, docDate, lineNo,
			item.ItemCode, item.Barcode, item.UnitCode, item.Qty, item.Price, item.SumAmount,
			desc, item.WhCode, item.LocationCode, itemRef,
		)
		if err != nil {
			return nil, fmt.Errorf("insert docdetail line %d: %w", lineNo, err)
		}
	}

	// บันทึกความสัมพันธ์การอ้างอิงเอกสาร (Document Reference Linkage)
	if docRefGlobal != "" {
		refTransFlag := doc.DocRefTransFlag
		if refTransFlag == 0 {
			if strings.HasPrefix(docRefGlobal, "PO") {
				refTransFlag = 6
			} else if strings.HasPrefix(docRefGlobal, "SO") {
				refTransFlag = 36
			} else if strings.HasPrefix(docRefGlobal, "QT") {
				refTransFlag = 34
			} else if strings.HasPrefix(docRefGlobal, "PR") {
				refTransFlag = 2
			} else if strings.HasPrefix(docRefGlobal, "INV") {
				refTransFlag = 44
			} else if strings.HasPrefix(docRefGlobal, "PUR") {
				refTransFlag = 12
			}
		}
		_, _ = tx.ExecContext(ctx, `DELETE FROM docref WHERE businesscode=$1 AND docno=$2`, company, docNo)
		_, _ = tx.ExecContext(ctx, `INSERT INTO docref (businesscode, docno, docnotransflag, docnoref, docnoreftransflag)
			VALUES ($1, $2, $3, $4, $5)`, company, docNo, transflag, docRefGlobal, refTransFlag)

		// อัปเดตเอกสารต้นทาง (PO, SO, QT, INV, PUR) ให้เป็นอ้างอิงแล้วและปิดเอกสาร
		_, _ = tx.ExecContext(ctx, `UPDATE doc SET isref = true, isclosed = true, modified_at = $1
			WHERE businesscode = $2 AND docno = $3`, now, company, docRefGlobal)
	}

	// Auto-posting GL สำหรับการซื้อสินค้า (Purchase: transflag=12) หรือ flag PostToGL
	if transflag == 12 || (doc.PostToGL && transflag != 44 && transflag != 28 && transflag != 29) {
		if err := s.postPurchaseToGL(ctx, tx, company, branch, doc, docNo, docDate, actor, now); err != nil {
			fmt.Printf("❌ [Auto-GL] postPurchaseToGL error for %s: %v\n", docNo, err)
		} else {
			fmt.Printf("✅ [Auto-GL] postPurchaseToGL success for %s\n", docNo)
		}
	}

	// Auto-posting GL และ AR สำหรับการขายสินค้า (Sale Invoice: transflag=44)
	if transflag == 44 {
		if err := s.postSaleToGL(ctx, tx, company, branch, doc, docNo, docDate, actor, now); err != nil {
			fmt.Printf("❌ [Auto-GL] postSaleToGL error for %s: %v\n", docNo, err)
		} else {
			fmt.Printf("✅ [Auto-GL] postSaleToGL success for %s\n", docNo)
		}
	}

	// Auto-posting GL และ AR Settlement สำหรับใบเสร็จรับเงิน (Receipt/Paid: transflag=28)
	if transflag == 28 {
		if err := s.postReceiptToGLAndAR(ctx, tx, company, branch, doc, docNo, docDate, actor, now); err != nil {
			fmt.Printf("❌ [Auto-GL] postReceiptToGLAndAR error for %s: %v\n", docNo, err)
		} else {
			fmt.Printf("✅ [Auto-GL] postReceiptToGLAndAR success for %s\n", docNo)
		}
	}

	// Auto-posting GL และ AP Settlement สำหรับใบสำคัญจ่าย (Payment/Pay: transflag=29)
	if transflag == 29 {
		if err := s.postPaymentToGLAndAP(ctx, tx, company, branch, doc, docNo, docDate, actor, now); err != nil {
			fmt.Printf("❌ [Auto-GL] postPaymentToGLAndAP error for %s: %v\n", docNo, err)
		} else {
			fmt.Printf("✅ [Auto-GL] postPaymentToGLAndAP success for %s\n", docNo)
		}
	}

	// Real-time Stock Movement สำหรับใบซื้อ (12) และใบขาย (44)
	if transflag == 12 || transflag == 44 {
		if err := s.postStockMovement(ctx, tx, company, branch, doc, docNo, transflag, docDate, now); err != nil {
			fmt.Printf("❌ [Stock] postStockMovement error for %s (transflag=%d): %v\n", docNo, transflag, err)
		} else {
			fmt.Printf("✅ [Stock] postStockMovement success for %s (transflag=%d)\n", docNo, transflag)
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

func (s *Store) postPurchaseToGL(ctx context.Context, tx *sql.Tx, company, branch string, doc *ErpTransactionDoc, docNo string, docDate time.Time, actor string, now time.Time) error {
	var fiscalYear string
	dateStr := docDate.Format("2006-01-02")
	err := tx.QueryRowContext(ctx, `SELECT code FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND payload->>'startdate'<=$2 AND payload->>'enddate'>=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company, dateStr).Scan(&fiscalYear)
	if err != nil {
		fiscalYear = "2569"
	}

	totalAmount := doc.TotalAmount
	vatAmount := doc.TotalVatValue
	if vatAmount == 0 && doc.TotalBeforeVat > 0 {
		vatAmount = totalAmount - doc.TotalBeforeVat
	}
	goodsAmount := totalAmount - vatAmount
	if goodsAmount <= 0 {
		goodsAmount = totalAmount
	}

	totalAmountStr := fmt.Sprintf("%.2f", totalAmount)
	goodsAmountStr := fmt.Sprintf("%.2f", goodsAmount)
	vatAmountStr := fmt.Sprintf("%.2f", vatAmount)

	glDocNo := "UV-" + docNo
	journalID := fmt.Sprintf("PUR-%s", docNo)

	lines := []map[string]any{
		{
			"accountcode": "1141",
			"accountname": "สินค้าสำเร็จรูป",
			"debit":       goodsAmountStr,
			"credit":      "0.00",
			"description": fmt.Sprintf("ซื้อสินค้าตามเอกสาร %s", docNo),
		},
	}
	if vatAmount > 0 {
		lines = append(lines, map[string]any{
			"accountcode": "1151",
			"accountname": "ภาษีซื้อ",
			"debit":       vatAmountStr,
			"credit":      "0.00",
			"description": fmt.Sprintf("ภาษีซื้อ 7%% ตามเอกสาร %s", docNo),
		})
	}
	lines = append(lines, map[string]any{
		"accountcode": "2111",
		"accountname": "เจ้าหนี้การค้า - ในประเทศ",
		"debit":       "0.00",
		"credit":      totalAmountStr,
		"description": fmt.Sprintf("ตั้งเจ้าหนี้ %s ตามเอกสาร %s", doc.CustCode, docNo),
	})

	payloadMap := map[string]any{
		"id":          journalID,
		"date":        dateStr,
		"kind":        "manual",
		"docno":       glDocNo,
		"bookcode":    "UV",
		"branchcode":  branch,
		"fiscalyear":  fiscalYear,
		"description": fmt.Sprintf("บันทึกซื้อสินค้าเชื่อ เลขที่ %s (เจ้าหนี้ %s)", docNo, doc.CustCode),
		"status":      "posted",
		"version":     2,
		"postedat":    now.Format(time.RFC3339Nano),
		"postedby":    actor,
		"createdat":   now.Format(time.RFC3339Nano),
		"createdby":   actor,
		"lines":       lines,
	}

	payloadJSON, err := json.Marshal(payloadMap)
	if err != nil {
		return err
	}

	// บันทึก gl_records (status: posted)
	_, err = tx.ExecContext(ctx, `INSERT INTO gl_records (company, kind, id, code, version, payload)
		VALUES ($1, 'journals', $2, $3, 2, $4)
		ON CONFLICT (company, kind, id) DO UPDATE SET version=2, payload=$4`,
		company, journalID, glDocNo, payloadJSON)
	if err != nil {
		return err
	}

	// บันทึก gl_lines
	_, _ = tx.ExecContext(ctx, `DELETE FROM gl_lines WHERE company=$1 AND journal_id=$2`, company, journalID)

	lineNo := 1
	// Line 1: Goods (Asset)
	_, _ = tx.ExecContext(ctx, `INSERT INTO gl_lines (
		company, journal_id, line_no, doc_no, entry_date, fiscal_year, book_code, branch_code,
		department_code, project_code, kind, currency, scale, account_code, account_name,
		account_type, normal_balance, is_cash, description, cash_flow, debit, credit
	) VALUES (
		$1, $2, $3, $4, $5, $6, 'UV', $7,
		'', '', 'manual', 'THB', 2, '1141', 'สินค้าสำเร็จรูป',
		'asset', 'debit', false, $8, '', $9, 0
	)`, company, journalID, lineNo, glDocNo, dateStr, fiscalYear, branch,
		fmt.Sprintf("ซื้อสินค้าตามเอกสาร %s", docNo), goodsAmount)
	lineNo++

	// Line 2: VAT Buy (if any)
	if vatAmount > 0 {
		_, _ = tx.ExecContext(ctx, `INSERT INTO gl_lines (
			company, journal_id, line_no, doc_no, entry_date, fiscal_year, book_code, branch_code,
			department_code, project_code, kind, currency, scale, account_code, account_name,
			account_type, normal_balance, is_cash, description, cash_flow, debit, credit
		) VALUES (
			$1, $2, $3, $4, $5, $6, 'UV', $7,
			'', '', 'manual', 'THB', 2, '1151', 'ภาษีซื้อ',
			'asset', 'debit', false, $8, '', $9, 0
		)`, company, journalID, lineNo, glDocNo, dateStr, fiscalYear, branch,
			fmt.Sprintf("ภาษีซื้อ 7%% ตามเอกสาร %s", docNo), vatAmount)
		lineNo++
	}

	// Line 3: AP Creditor
	_, _ = tx.ExecContext(ctx, `INSERT INTO gl_lines (
		company, journal_id, line_no, doc_no, entry_date, fiscal_year, book_code, branch_code,
		department_code, project_code, kind, currency, scale, account_code, account_name,
		account_type, normal_balance, is_cash, description, cash_flow, debit, credit
	) VALUES (
		$1, $2, $3, $4, $5, $6, 'UV', $7,
		'', '', 'manual', 'THB', 2, '2111', 'เจ้าหนี้การค้า - ในประเทศ',
		'liability', 'credit', false, $8, '', 0, $9
	)`, company, journalID, lineNo, glDocNo, dateStr, fiscalYear, branch,
		fmt.Sprintf("ตั้งเจ้าหนี้ %s ตามเอกสาร %s", doc.CustCode, docNo), totalAmount)

	// บันทึก AP Subledger (ตาราง gl_subledger_documents)
	if strings.TrimSpace(doc.CustCode) != "" {
		var partnerExists bool
		_ = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_partners WHERE company=$1 AND code=$2)`, company, doc.CustCode).Scan(&partnerExists)
		if partnerExists {
			subledgerDocID := fmt.Sprintf("AP-%s", docNo)
			subPayload, _ := json.Marshal(map[string]any{
				"doc_no":      docNo,
				"doc_date":    dateStr,
				"due_date":    dateStr,
				"journal_id":  journalID,
				"total":       totalAmountStr,
				"description": fmt.Sprintf("ซื้อสินค้าเชื่อ เอกสาร %s", docNo),
			})
			_, _ = tx.ExecContext(ctx, `INSERT INTO gl_subledger_documents (
				company, id, ledger, partner_code, branch_code, control_account_code, amount, side, version, created_journal_id, payload
			) VALUES (
				$1, $2, 'ap', $3, $4, '2111', $5, 2, 1, $6, $7
			) ON CONFLICT (company, id) DO UPDATE SET amount=$5, payload=$7`,
				company, subledgerDocID, doc.CustCode, branch, totalAmount, journalID, subPayload)
		}
	}

	return nil
}

func (s *Store) postSaleToGL(ctx context.Context, tx *sql.Tx, company, branch string, doc *ErpTransactionDoc, docNo string, docDate time.Time, actor string, now time.Time) error {
	var fiscalYear string
	dateStr := docDate.Format("2006-01-02")
	err := tx.QueryRowContext(ctx, `SELECT code FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND payload->>'startdate'<=$2 AND payload->>'enddate'>=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company, dateStr).Scan(&fiscalYear)
	if err != nil {
		fiscalYear = "2569"
	}

	totalAmount := doc.TotalAmount
	vatAmount := doc.TotalVatValue
	if vatAmount == 0 && doc.TotalBeforeVat > 0 {
		vatAmount = totalAmount - doc.TotalBeforeVat
	}
	goodsAmount := totalAmount - vatAmount
	if goodsAmount <= 0 {
		goodsAmount = totalAmount
	}

	totalAmountStr := fmt.Sprintf("%.2f", totalAmount)
	goodsAmountStr := fmt.Sprintf("%.2f", goodsAmount)
	vatAmountStr := fmt.Sprintf("%.2f", vatAmount)

	glDocNo := "SV-" + docNo
	journalID := fmt.Sprintf("SAL-%s", docNo)

	lines := []map[string]any{
		{
			"accountcode": "1131",
			"accountname": "ลูกหนี้การค้า - ในประเทศ",
			"debit":       totalAmountStr,
			"credit":      "0.00",
			"description": fmt.Sprintf("ตั้งลูกหนี้ %s ตามเอกสาร %s", doc.CustCode, docNo),
		},
		{
			"accountcode": "4111",
			"accountname": "รายได้จากการขายสินค้า - มีภาษีมูลค่าเพิ่ม 7%",
			"debit":       "0.00",
			"credit":      goodsAmountStr,
			"description": fmt.Sprintf("ขายสินค้าตามเอกสาร %s", docNo),
		},
	}
	if vatAmount > 0 {
		lines = append(lines, map[string]any{
			"accountcode": "2131",
			"accountname": "ภาษีขาย",
			"debit":       "0.00",
			"credit":      vatAmountStr,
			"description": fmt.Sprintf("ภาษีขาย 7%% ตามเอกสาร %s", docNo),
		})
	}

	payloadMap := map[string]any{
		"id":          journalID,
		"date":        dateStr,
		"kind":        "manual",
		"docno":       glDocNo,
		"bookcode":    "SV",
		"branchcode":  branch,
		"fiscalyear":  fiscalYear,
		"description": fmt.Sprintf("บันทึกขายสินค้าเชื่อ เลขที่ %s (ลูกหนี้ %s)", docNo, doc.CustCode),
		"status":      "posted",
		"version":     2,
		"postedat":    now.Format(time.RFC3339Nano),
		"postedby":    actor,
		"createdat":   now.Format(time.RFC3339Nano),
		"createdby":   actor,
		"lines":       lines,
	}

	payloadJSON, err := json.Marshal(payloadMap)
	if err != nil {
		return err
	}

	// บันทึก gl_records (status: posted)
	_, err = tx.ExecContext(ctx, `INSERT INTO gl_records (company, kind, id, code, version, payload)
		VALUES ($1, 'journals', $2, $3, 2, $4)
		ON CONFLICT (company, kind, id) DO UPDATE SET version=2, payload=$4`,
		company, journalID, glDocNo, payloadJSON)
	if err != nil {
		return err
	}

	// บันทึก gl_lines
	_, _ = tx.ExecContext(ctx, `DELETE FROM gl_lines WHERE company=$1 AND journal_id=$2`, company, journalID)

	lineNo := 1
	// Line 1: Trade Accounts Receivable (Asset, Debit)
	_, _ = tx.ExecContext(ctx, `INSERT INTO gl_lines (
		company, journal_id, line_no, doc_no, entry_date, fiscal_year, book_code, branch_code,
		department_code, project_code, kind, currency, scale, account_code, account_name,
		account_type, normal_balance, is_cash, description, cash_flow, debit, credit
	) VALUES (
		$1, $2, $3, $4, $5, $6, 'SV', $7,
		'', '', 'manual', 'THB', 2, '1131', 'ลูกหนี้การค้า - ในประเทศ',
		'asset', 'debit', false, $8, '', $9, 0
	)`, company, journalID, lineNo, glDocNo, dateStr, fiscalYear, branch,
		fmt.Sprintf("ตั้งลูกหนี้ %s ตามเอกสาร %s", doc.CustCode, docNo), totalAmount)
	lineNo++

	// Line 2: Sales Revenue (Income, Credit)
	_, _ = tx.ExecContext(ctx, `INSERT INTO gl_lines (
		company, journal_id, line_no, doc_no, entry_date, fiscal_year, book_code, branch_code,
		department_code, project_code, kind, currency, scale, account_code, account_name,
		account_type, normal_balance, is_cash, description, cash_flow, debit, credit
	) VALUES (
		$1, $2, $3, $4, $5, $6, 'SV', $7,
		'', '', 'manual', 'THB', 2, '4111', 'รายได้จากการขายสินค้า - มีภาษีมูลค่าเพิ่ม 7%',
		'income', 'credit', false, $8, '', 0, $9
	)`, company, journalID, lineNo, glDocNo, dateStr, fiscalYear, branch,
		fmt.Sprintf("ขายสินค้าตามเอกสาร %s", docNo), goodsAmount)
	lineNo++

	// Line 3: Output VAT (Liability, Credit)
	if vatAmount > 0 {
		_, _ = tx.ExecContext(ctx, `INSERT INTO gl_lines (
			company, journal_id, line_no, doc_no, entry_date, fiscal_year, book_code, branch_code,
			department_code, project_code, kind, currency, scale, account_code, account_name,
			account_type, normal_balance, is_cash, description, cash_flow, debit, credit
		) VALUES (
			$1, $2, $3, $4, $5, $6, 'SV', $7,
			'', '', 'manual', 'THB', 2, '2131', 'ภาษีขาย',
			'liability', 'credit', false, $8, '', 0, $9
		)`, company, journalID, lineNo, glDocNo, dateStr, fiscalYear, branch,
			fmt.Sprintf("ภาษีขาย 7%% ตามเอกสาร %s", docNo), vatAmount)
		lineNo++
	}

	// บันทึก AR Subledger (ตาราง gl_subledger_documents)
	if strings.TrimSpace(doc.CustCode) != "" {
		var partnerExists bool
		_ = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM gl_subledger_partners WHERE company=$1 AND code=$2)`, company, doc.CustCode).Scan(&partnerExists)
		if partnerExists {
			subledgerDocID := fmt.Sprintf("AR-%s", docNo)
			subPayload, _ := json.Marshal(map[string]any{
				"doc_no":      docNo,
				"doc_date":    dateStr,
				"due_date":    dateStr,
				"journal_id":  journalID,
				"total":       totalAmountStr,
				"description": fmt.Sprintf("ขายสินค้าเชื่อ เอกสาร %s", docNo),
			})
			_, _ = tx.ExecContext(ctx, `INSERT INTO gl_subledger_documents (
				company, id, ledger, partner_code, branch_code, control_account_code, amount, side, version, created_journal_id, payload
			) VALUES (
				$1, $2, 'ar', $3, $4, '1131', $5, 1, 1, $6, $7
			) ON CONFLICT (company, id) DO UPDATE SET amount=$5, payload=$7`,
				company, subledgerDocID, doc.CustCode, branch, totalAmount, journalID, subPayload)
		}
	}

	return nil
}

func (s *Store) postReceiptToGLAndAR(ctx context.Context, tx *sql.Tx, company, branch string, doc *ErpTransactionDoc, docNo string, docDate time.Time, actor string, now time.Time) error {
	var fiscalYear string
	dateStr := docDate.Format("2006-01-02")
	err := tx.QueryRowContext(ctx, `SELECT code FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND payload->>'startdate'<=$2 AND payload->>'enddate'>=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company, dateStr).Scan(&fiscalYear)
	if err != nil {
		fiscalYear = "2569"
	}

	totalAmount := doc.TotalAmount
	totalAmountStr := fmt.Sprintf("%.2f", totalAmount)

	glDocNo := "RV-" + docNo
	journalID := fmt.Sprintf("RCP-%s", docNo)

	lines := []map[string]any{
		{
			"accountcode": "1121",
			"accountname": "เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย",
			"debit":       totalAmountStr,
			"credit":      "0.00",
			"description": fmt.Sprintf("รับชำระเงินจากลูกหนี้ %s ตามเอกสาร %s", doc.CustCode, docNo),
		},
		{
			"accountcode": "1131",
			"accountname": "ลูกหนี้การค้า - ในประเทศ",
			"debit":       "0.00",
			"credit":      totalAmountStr,
			"description": fmt.Sprintf("ตัดยอดลูกหนี้ %s ตามเอกสาร %s", doc.CustCode, docNo),
		},
	}

	payloadMap := map[string]any{
		"id":          journalID,
		"date":        dateStr,
		"kind":        "manual",
		"docno":       glDocNo,
		"bookcode":    "RV",
		"branchcode":  branch,
		"fiscalyear":  fiscalYear,
		"description": fmt.Sprintf("ใบเสร็จรับเงิน เลขที่ %s (ลูกหนี้ %s)", docNo, doc.CustCode),
		"status":      "posted",
		"version":     2,
		"postedat":    now.Format(time.RFC3339Nano),
		"postedby":    actor,
		"createdat":   now.Format(time.RFC3339Nano),
		"createdby":   actor,
		"lines":       lines,
	}

	payloadJSON, err := json.Marshal(payloadMap)
	if err != nil {
		return err
	}

	// บันทึก gl_records (status: posted)
	_, err = tx.ExecContext(ctx, `INSERT INTO gl_records (company, kind, id, code, version, payload)
		VALUES ($1, 'journals', $2, $3, 2, $4)
		ON CONFLICT (company, kind, id) DO UPDATE SET version=2, payload=$4`,
		company, journalID, glDocNo, payloadJSON)
	if err != nil {
		return err
	}

	// บันทึก gl_lines
	_, _ = tx.ExecContext(ctx, `DELETE FROM gl_lines WHERE company=$1 AND journal_id=$2`, company, journalID)

	lineNo := 1
	// Line 1: Bank (Debit)
	_, _ = tx.ExecContext(ctx, `INSERT INTO gl_lines (
		company, journal_id, line_no, doc_no, entry_date, fiscal_year, book_code, branch_code,
		department_code, project_code, kind, currency, scale, account_code, account_name,
		account_type, normal_balance, is_cash, description, cash_flow, debit, credit
	) VALUES (
		$1, $2, $3, $4, $5, $6, 'RV', $7,
		'', '', 'manual', 'THB', 2, '1121', 'เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย',
		'asset', 'debit', true, $8, '', $9, 0
	)`, company, journalID, lineNo, glDocNo, dateStr, fiscalYear, branch,
		fmt.Sprintf("รับชำระเงินตามเอกสาร %s", docNo), totalAmount)
	lineNo++

	// Line 2: Accounts Receivable (Credit)
	_, _ = tx.ExecContext(ctx, `INSERT INTO gl_lines (
		company, journal_id, line_no, doc_no, entry_date, fiscal_year, book_code, branch_code,
		department_code, project_code, kind, currency, scale, account_code, account_name,
		account_type, normal_balance, is_cash, description, cash_flow, debit, credit
	) VALUES (
		$1, $2, $3, $4, $5, $6, 'RV', $7,
		'', '', 'manual', 'THB', 2, '1131', 'ลูกหนี้การค้า - ในประเทศ',
		'asset', 'credit', false, $8, '', 0, $9
	)`, company, journalID, lineNo, glDocNo, dateStr, fiscalYear, branch,
		fmt.Sprintf("ตัดยอดลูกหนี้ %s ตามเอกสาร %s", doc.CustCode, docNo), totalAmount)

	// บันทึก AR Subledger (ตาราง gl_subledger_documents & settlements)
	if strings.TrimSpace(doc.CustCode) != "" {
		// ตรวจสอบและ upsert partner ถ้ายังไม่มี
		partnerPayload, _ := json.Marshal(map[string]any{
			"code": doc.CustCode,
			"name": doc.CustCode,
			"type": "customer",
		})
		_, _ = tx.ExecContext(ctx, `INSERT INTO gl_subledger_partners (company, code, version, payload)
			VALUES ($1, $2, 1, $3)
			ON CONFLICT (company, code) DO NOTHING`, company, doc.CustCode, partnerPayload)

		receiptDocID := fmt.Sprintf("RCP-%s", docNo)
		subPayload, _ := json.Marshal(map[string]any{
			"doc_no":      docNo,
			"doc_date":    dateStr,
			"journal_id":  journalID,
			"total":       totalAmountStr,
			"description": fmt.Sprintf("รับชำระเงิน เอกสาร %s", docNo),
		})

		_, _ = tx.ExecContext(ctx, `INSERT INTO gl_subledger_documents (
			company, id, ledger, partner_code, branch_code, control_account_code, amount, side, version, created_journal_id, payload
		) VALUES (
			$1, $2, 'ar', $3, $4, '1131', $5, 2, 1, $6, $7
		) ON CONFLICT (company, id) DO UPDATE SET amount=$5, payload=$7`,
			company, receiptDocID, doc.CustCode, branch, totalAmount, journalID, subPayload)

		// ตัดหนี้ Settlement หากมี docref
		refDoc := strings.TrimSpace(doc.DocRef)
		if refDoc != "" {
			var debtDocID string
			possibleDebtIDs := []string{fmt.Sprintf("AR-%s", refDoc), refDoc}
			for _, pid := range possibleDebtIDs {
				var foundID string
				err := tx.QueryRowContext(ctx, `SELECT id FROM gl_subledger_documents WHERE company=$1 AND id=$2 AND ledger='ar' AND side=1`, company, pid).Scan(&foundID)
				if err == nil && foundID != "" {
					debtDocID = foundID
					break
				}
			}

			if debtDocID != "" {
				settlementID := fmt.Sprintf("SET-AR-%s", docNo)
				settlePayload, _ := json.Marshal(map[string]any{
					"debt_id":    debtDocID,
					"payment_id": receiptDocID,
					"docno":      docNo,
					"docref":     refDoc,
					"amount":     totalAmountStr,
				})
				_, _ = tx.ExecContext(ctx, `INSERT INTO gl_subledger_settlements (
					company, id, journal_id, debt_id, payment_id, amount, settlement_date, payload, created_at, created_by
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
				) ON CONFLICT (company, id) DO UPDATE SET amount=$6, payload=$8`,
					company, settlementID, journalID, debtDocID, receiptDocID, totalAmount, dateStr, settlePayload, now, actor)
			}
		}
	}

	return nil
}

func (s *Store) postPaymentToGLAndAP(ctx context.Context, tx *sql.Tx, company, branch string, doc *ErpTransactionDoc, docNo string, docDate time.Time, actor string, now time.Time) error {
	var fiscalYear string
	dateStr := docDate.Format("2006-01-02")
	err := tx.QueryRowContext(ctx, `SELECT code FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND payload->>'startdate'<=$2 AND payload->>'enddate'>=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, company, dateStr).Scan(&fiscalYear)
	if err != nil {
		fiscalYear = "2569"
	}

	totalAmount := doc.TotalAmount
	totalAmountStr := fmt.Sprintf("%.2f", totalAmount)

	glDocNo := "PV-" + docNo
	journalID := fmt.Sprintf("PAY-%s", docNo)

	lines := []map[string]any{
		{
			"accountcode": "2111",
			"accountname": "เจ้าหนี้การค้า - ในประเทศ",
			"debit":       totalAmountStr,
			"credit":      "0.00",
			"description": fmt.Sprintf("จ่ายชำระหนี้ %s ตามเอกสาร %s", doc.CustCode, docNo),
		},
		{
			"accountcode": "1121",
			"accountname": "เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย",
			"debit":       "0.00",
			"credit":      totalAmountStr,
			"description": fmt.Sprintf("จ่ายเงินผ่านบัญชีกระแสรายวัน ตามเอกสาร %s", docNo),
		},
	}

	payloadMap := map[string]any{
		"id":          journalID,
		"date":        dateStr,
		"kind":        "manual",
		"docno":       glDocNo,
		"bookcode":    "PV",
		"branchcode":  branch,
		"fiscalyear":  fiscalYear,
		"description": fmt.Sprintf("ใบสำคัญจ่าย เลขที่ %s (เจ้าหนี้ %s)", docNo, doc.CustCode),
		"status":      "posted",
		"version":     2,
		"postedat":    now.Format(time.RFC3339Nano),
		"postedby":    actor,
		"createdat":   now.Format(time.RFC3339Nano),
		"createdby":   actor,
		"lines":       lines,
	}

	payloadJSON, err := json.Marshal(payloadMap)
	if err != nil {
		return err
	}

	// บันทึก gl_records (status: posted)
	_, err = tx.ExecContext(ctx, `INSERT INTO gl_records (company, kind, id, code, version, payload)
		VALUES ($1, 'journals', $2, $3, 2, $4)
		ON CONFLICT (company, kind, id) DO UPDATE SET version=2, payload=$4`,
		company, journalID, glDocNo, payloadJSON)
	if err != nil {
		return err
	}

	// บันทึก gl_lines
	_, _ = tx.ExecContext(ctx, `DELETE FROM gl_lines WHERE company=$1 AND journal_id=$2`, company, journalID)

	lineNo := 1
	// Line 1: Accounts Payable (Debit)
	_, _ = tx.ExecContext(ctx, `INSERT INTO gl_lines (
		company, journal_id, line_no, doc_no, entry_date, fiscal_year, book_code, branch_code,
		department_code, project_code, kind, currency, scale, account_code, account_name,
		account_type, normal_balance, is_cash, description, cash_flow, debit, credit
	) VALUES (
		$1, $2, $3, $4, $5, $6, 'PV', $7,
		'', '', 'manual', 'THB', 2, '2111', 'เจ้าหนี้การค้า - ในประเทศ',
		'liability', 'debit', false, $8, '', $9, 0
	)`, company, journalID, lineNo, glDocNo, dateStr, fiscalYear, branch,
		fmt.Sprintf("จ่ายชำระเจ้าหนี้ %s ตามเอกสาร %s", doc.CustCode, docNo), totalAmount)
	lineNo++

	// Line 2: Bank (Credit)
	_, _ = tx.ExecContext(ctx, `INSERT INTO gl_lines (
		company, journal_id, line_no, doc_no, entry_date, fiscal_year, book_code, branch_code,
		department_code, project_code, kind, currency, scale, account_code, account_name,
		account_type, normal_balance, is_cash, description, cash_flow, debit, credit
	) VALUES (
		$1, $2, $3, $4, $5, $6, 'PV', $7,
		'', '', 'manual', 'THB', 2, '1121', 'เงินฝากกระแสรายวัน - ธนาคารกสิกรไทย',
		'asset', 'credit', true, $8, '', 0, $9
	)`, company, journalID, lineNo, glDocNo, dateStr, fiscalYear, branch,
		fmt.Sprintf("จ่ายเงินชำระหนี้ตามเอกสาร %s", docNo), totalAmount)

	// บันทึก AP Subledger (ตาราง gl_subledger_documents & settlements)
	if strings.TrimSpace(doc.CustCode) != "" {
		// ตรวจสอบและ upsert partner ถ้ายังไม่มี
		partnerPayload, _ := json.Marshal(map[string]any{
			"code": doc.CustCode,
			"name": doc.CustCode,
			"type": "supplier",
		})
		_, _ = tx.ExecContext(ctx, `INSERT INTO gl_subledger_partners (company, code, version, payload)
			VALUES ($1, $2, 1, $3)
			ON CONFLICT (company, code) DO NOTHING`, company, doc.CustCode, partnerPayload)

		payDocID := fmt.Sprintf("PAY-%s", docNo)
		subPayload, _ := json.Marshal(map[string]any{
			"doc_no":      docNo,
			"doc_date":    dateStr,
			"journal_id":  journalID,
			"total":       totalAmountStr,
			"description": fmt.Sprintf("จ่ายชำระเงิน เอกสาร %s", docNo),
		})

		_, _ = tx.ExecContext(ctx, `INSERT INTO gl_subledger_documents (
			company, id, ledger, partner_code, branch_code, control_account_code, amount, side, version, created_journal_id, payload
		) VALUES (
			$1, $2, 'ap', $3, $4, '2111', $5, 1, 1, $6, $7
		) ON CONFLICT (company, id) DO UPDATE SET amount=$5, payload=$7`,
			company, payDocID, doc.CustCode, branch, totalAmount, journalID, subPayload)

		// ตัดหนี้ Settlement หากมี docref
		refDoc := strings.TrimSpace(doc.DocRef)
		if refDoc != "" {
			var debtDocID string
			possibleDebtIDs := []string{fmt.Sprintf("AP-%s", refDoc), refDoc}
			for _, pid := range possibleDebtIDs {
				var foundID string
				err := tx.QueryRowContext(ctx, `SELECT id FROM gl_subledger_documents WHERE company=$1 AND id=$2 AND ledger='ap' AND side=2`, company, pid).Scan(&foundID)
				if err == nil && foundID != "" {
					debtDocID = foundID
					break
				}
			}

			if debtDocID != "" {
				settlementID := fmt.Sprintf("SET-AP-%s", docNo)
				settlePayload, _ := json.Marshal(map[string]any{
					"debt_id":    debtDocID,
					"payment_id": payDocID,
					"docno":      docNo,
					"docref":     refDoc,
					"amount":     totalAmountStr,
				})
				_, _ = tx.ExecContext(ctx, `INSERT INTO gl_subledger_settlements (
					company, id, journal_id, debt_id, payment_id, amount, settlement_date, payload, created_at, created_by
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
				) ON CONFLICT (company, id) DO UPDATE SET amount=$6, payload=$8`,
					company, settlementID, journalID, debtDocID, payDocID, totalAmount, dateStr, settlePayload, now, actor)
			}
		}
	}

	return nil
}

func (s *Store) postStockMovement(ctx context.Context, tx *sql.Tx, company, branch string, doc *ErpTransactionDoc, docNo string, transflag int, docDate time.Time, now time.Time) error {
	if len(doc.Details) == 0 {
		return nil
	}

	// ลบรายการเดิมของ docno ใน stock_ledger ก่อน
	_, _ = tx.ExecContext(ctx, `DELETE FROM stock_ledger WHERE businesscode=$1 AND docno=$2`, company, docNo)

	direction := 1
	if transflag == 44 || transflag == 20 { // Sale หรือ Out
		direction = -1
	}

	periodKey := docDate.Format("2006-01")

	for i, detail := range doc.Details {
		itemCode := strings.TrimSpace(detail.ItemCode)
		if itemCode == "" {
			continue
		}
		whCode := strings.TrimSpace(detail.WhCode)
		if whCode == "" {
			whCode = "WH01"
		}
		qty := detail.Qty
		if qty < 0 {
			qty = -qty
		}
		price := detail.Price
		if price < 0 {
			price = 0
		}
		docValue := qty * price

		// ค้นหายอดคงเหลือล่าสุดของสินค้านี้ในคลังนี้
		var lastBalQty, lastBalAmount, lastAvgCost float64
		row := tx.QueryRowContext(ctx, `SELECT balanceqty, balanceamount, avgcost
			FROM stock_ledger
			WHERE businesscode=$1 AND itemcode=$2 AND whcode=$3
			ORDER BY docdatetime DESC, behindindex DESC, docno DESC, linenumber DESC
			LIMIT 1`, company, itemCode, whCode)
		_ = row.Scan(&lastBalQty, &lastBalAmount, &lastAvgCost)

		var unitCost, amount, newBalQty, newBalAmount, avgCost float64
		if direction == 1 { // รับเข้า (Purchase)
			unitCost = price
			amount = docValue
			newBalQty = lastBalQty + qty
			newBalAmount = lastBalAmount + amount
			if newBalQty > 0 {
				avgCost = newBalAmount / newBalQty
			} else {
				avgCost = unitCost
			}
		} else { // จ่ายออก (Sale)
			if lastAvgCost > 0 {
				unitCost = lastAvgCost
			} else {
				unitCost = price
			}
			amount = qty * unitCost
			newBalQty = lastBalQty - qty
			newBalAmount = lastBalAmount - amount
			if newBalAmount < 0 {
				newBalAmount = 0
			}
			avgCost = lastAvgCost
			if avgCost <= 0 && newBalQty > 0 {
				avgCost = newBalAmount / newBalQty
			}
		}

		lineNo := i + 1
		if detail.LineNumber > 0 {
			lineNo = detail.LineNumber
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO stock_ledger (
			businesscode, itemcode, whcode, docdatetime, behindindex, docno, linenumber,
			periodkey, locationcode, barcode, unitcode, docref, transflag, direction,
			docqty, unitstand, unitdivide, price, docvalue, qty, unitcost, amount,
			balanceqty, balanceamount, avgcost, calculatedat
		) VALUES (
			$1, $2, $3, $4, 0, $5, $6,
			$7, $8, $9, $10, $11, $12, $13,
			$14, 1, 1, $15, $16, $14, $17, $18,
			$19, $20, $21, $22
		)`, company, itemCode, whCode, docDate, docNo, lineNo,
			periodKey, detail.LocationCode, detail.Barcode, detail.UnitCode, doc.DocRef, transflag, direction,
			qty, price, docValue, unitCost, amount,
			newBalQty, newBalAmount, avgCost, now); err != nil {
			fmt.Printf("❌ [Stock] stock_ledger insert error: %v\n", err)
		}

		// บันทึก stock_period_balance
		if _, err := tx.ExecContext(ctx, `INSERT INTO stock_period_balance (
			businesscode, itemcode, whcode, periodkey, closeqty, closeamount, closeavgcost, hastrans, updatedat
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, true, $8
		) ON CONFLICT (businesscode, itemcode, whcode, periodkey)
		DO UPDATE SET closeqty = $5, closeamount = $6, closeavgcost = $7, hastrans = true, updatedat = $8`,
			company, itemCode, whCode, periodKey, newBalQty, newBalAmount, avgCost, now); err != nil {
			fmt.Printf("❌ [Stock] stock_period_balance insert error: %v\n", err)
		}
	}

	return nil
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

	var docNo string
	_ = tx.QueryRowContext(ctx, fmt.Sprintf("SELECT docno FROM doc %s", where), args[:len(args)-2]...).Scan(&docNo)
	if docNo != "" {
		_, _ = tx.ExecContext(ctx, "DELETE FROM stock_ledger WHERE businesscode=$1 AND docno=$2", company, docNo)
	}

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
