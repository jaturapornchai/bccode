package httpapi

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

	"smlcloudplatform/internal/doctype"
	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/pkg/microservice"
)

type ReceiptAdjustmentItem struct {
	DocNo      string `json:"docno"`
	DocDate    string `json:"docdate"`
	VendorCode string `json:"vendor_code"`
	VendorName string `json:"vendor_name"`
	IRRefNo    string `json:"ir_ref_no"`
	IRRefDate  string `json:"ir_ref_date"`
	GRQty      string `json:"gr_qty"`
	IRQty      string `json:"ir_qty"`
	IsMatched  bool   `json:"is_matched"`
	IsClosed   bool   `json:"is_closed"`
}

type ReceiptAdjustmentFilter struct {
	FromDate   string
	ToDate     string
	FromDocNo  string
	ToDocNo    string
	FromVendor string
	ToVendor   string
	FromDept   string
	ToDept     string
}

type ProcessReceiptAdjustmentsPayload struct {
	DocNos []string `json:"docnos"`
}

// ListReceiptAdjustments ค้นหาใบรับสินค้า (Goods Receive - transflag 401) ที่ยังค้างอยู่
// เพื่อนำมาตรวจสอบและปรับปรุงสถานะเมื่อตั้งหนี้หรือรับบิลครบแล้ว เทียบเท่า OnBnProcess ใน POFrmAdjGRStatus.cpp
func (s *Store) ListReceiptAdjustments(ctx context.Context, db *sql.DB, company string, filter ReceiptAdjustmentFilter) ([]ReceiptAdjustmentItem, error) {
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, fmt.Sprintf("d.businesscode = $%d", argIdx))
	args = append(args, company)
	argIdx++

	whereClauses = append(whereClauses, fmt.Sprintf("d.transflag = $%d", argIdx))
	args = append(args, doctype.GoodsReceive)
	argIdx++

	whereClauses = append(whereClauses, "NOT d.isdelete", "NOT d.iscancel", "NOT d.isclosed")

	if filter.FromDate != "" {
		if filter.ToDate != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("d.docdatetime::date BETWEEN $%d AND $%d", argIdx, argIdx+1))
			args = append(args, filter.FromDate, filter.ToDate)
			argIdx += 2
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("d.docdatetime::date >= $%d", argIdx))
			args = append(args, filter.FromDate)
			argIdx++
		}
	} else if filter.ToDate != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.docdatetime::date <= $%d", argIdx))
		args = append(args, filter.ToDate)
		argIdx++
	}

	if filter.FromDocNo != "" {
		if filter.ToDocNo != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("d.docno BETWEEN $%d AND $%d", argIdx, argIdx+1))
			args = append(args, filter.FromDocNo, filter.ToDocNo)
			argIdx += 2
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("d.docno >= $%d", argIdx))
			args = append(args, filter.FromDocNo)
			argIdx++
		}
	} else if filter.ToDocNo != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.docno <= $%d", argIdx))
		args = append(args, filter.ToDocNo)
		argIdx++
	}

	if filter.FromVendor != "" {
		if filter.ToVendor != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("d.custcode BETWEEN $%d AND $%d", argIdx, argIdx+1))
			args = append(args, filter.FromVendor, filter.ToVendor)
			argIdx += 2
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("d.custcode >= $%d", argIdx))
			args = append(args, filter.FromVendor)
			argIdx++
		}
	} else if filter.ToVendor != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.custcode <= $%d", argIdx))
		args = append(args, filter.ToVendor)
		argIdx++
	}

	if filter.FromDept != "" {
		if filter.ToDept != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("d.branchid BETWEEN $%d AND $%d", argIdx, argIdx+1))
			args = append(args, filter.FromDept, filter.ToDept)
			argIdx += 2
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("d.branchid = $%d", argIdx))
			args = append(args, filter.FromDept)
			argIdx++
		}
	}

	query := fmt.Sprintf(`
		SELECT d.docno,
		       COALESCE(d.docdatetime, now())::timestamptz,
		       COALESCE(d.custcode, ''),
		       COALESCE(d.custname, ''),
		       COALESCE((
		           SELECT r.docno
		           FROM docref r
		           JOIN doc ir ON ir.businesscode = r.businesscode AND ir.docno = r.docno
		           WHERE r.businesscode = d.businesscode
		             AND r.docnoref = d.docno
		             AND NOT ir.isdelete
		             AND NOT ir.iscancel
		           ORDER BY ir.docdatetime DESC, ir.docno DESC
		           LIMIT 1
		       ), ''),
		       COALESCE((
		           SELECT ir.docdatetime::text
		           FROM docref r
		           JOIN doc ir ON ir.businesscode = r.businesscode AND ir.docno = r.docno
		           WHERE r.businesscode = d.businesscode
		             AND r.docnoref = d.docno
		             AND NOT ir.isdelete
		             AND NOT ir.iscancel
		           ORDER BY ir.docdatetime DESC, ir.docno DESC
		           LIMIT 1
		       ), ''),
		       COALESCE((
		           SELECT TRIM(TRAILING '.' FROM TRIM(TRAILING '0' FROM ROUND(SUM(totalqty), 4)::text))
		           FROM docdetail
		           WHERE businesscode = d.businesscode
		             AND docno = d.docno
		             AND NOT iscancel
		       ), '0'),
		       COALESCE((
		           SELECT TRIM(TRAILING '.' FROM TRIM(TRAILING '0' FROM ROUND(SUM(ir_det.totalqty), 4)::text))
		           FROM docdetail ir_det
		           JOIN doc ir ON ir.businesscode = ir_det.businesscode AND ir.docno = ir_det.docno
		           WHERE ir_det.businesscode = d.businesscode
		             AND NOT ir.isdelete
		             AND NOT ir.iscancel
		             AND NOT ir_det.iscancel
		             AND (
		                 ir_det.docref = d.docno
		                 OR ir_det.docno IN (
		                     SELECT r.docno
		                     FROM docref r
		                     WHERE r.businesscode = d.businesscode
		                       AND r.docnoref = d.docno
		                 )
		             )
		       ), '0'),
		       d.isclosed
		FROM doc d
		WHERE %s
		ORDER BY d.docdatetime ASC, d.docno ASC
		LIMIT 500
	`, strings.Join(whereClauses, " AND "))

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list receipt adjustments: %w", err)
	}
	defer rows.Close()

	var items []ReceiptAdjustmentItem
	for rows.Next() {
		var docno, custcode, custname, irRefNo, irRefDateRaw, grQty, irQty string
		var docDate time.Time
		var isClosed bool
		if err := rows.Scan(&docno, &docDate, &custcode, &custname, &irRefNo, &irRefDateRaw, &grQty, &irQty, &isClosed); err != nil {
			return nil, err
		}

		irRefDate := ""
		if irRefDateRaw != "" {
			if t, err := time.Parse(time.RFC3339, irRefDateRaw); err == nil {
				irRefDate = t.Format("2006-01-02")
			} else if len(irRefDateRaw) >= 10 {
				irRefDate = irRefDateRaw[:10]
			}
		}

		if grQty == "" {
			grQty = "0"
		}
		if irQty == "" {
			irQty = "0"
		}

		isMatched := irRefNo != "" && irQty != "0"

		items = append(items, ReceiptAdjustmentItem{
			DocNo:      docno,
			DocDate:    docDate.Format("2006-01-02"),
			VendorCode: custcode,
			VendorName: custname,
			IRRefNo:    irRefNo,
			IRRefDate:  irRefDate,
			GRQty:      grQty,
			IRQty:      irQty,
			IsMatched:  isMatched,
			IsClosed:   isClosed,
		})
	}
	if items == nil {
		items = []ReceiptAdjustmentItem{}
	}
	return items, nil
}

// ProcessReceiptAdjustments ปรับปรุงสถานะใบรับสินค้า (GR) เป็นปิดเอกสาร (isclosed = true)
// เทียบเท่า OnBnOk ใน POFrmAdjGRStatus.cpp
func (s *Store) ProcessReceiptAdjustments(ctx context.Context, db *sql.DB, scope gl.Scope, docNos []string) (int, error) {
	if len(docNos) == 0 {
		return 0, nil
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if err := gl.LockCompanyTx(ctx, tx, scope.Company); err != nil {
		return 0, ledgerFail(err)
	}

	updatedCount := 0
	for _, rawNo := range docNos {
		docno := strings.TrimSpace(rawNo)
		if docno == "" {
			continue
		}

		var docDate sql.NullTime
		err := tx.QueryRowContext(ctx, `SELECT docdatetime FROM doc
			WHERE businesscode = $1 AND docno = $2 AND transflag = $3 AND NOT isdelete AND NOT iscancel FOR UPDATE`,
			scope.Company, docno, doctype.GoodsReceive).Scan(&docDate)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return 0, fmt.Errorf("lock gr %s: %w", docno, err)
		}
		if err := checkDocDates(ctx, tx, scope, docDate); err != nil {
			return 0, err
		}

		res, err := tx.ExecContext(ctx, `
			UPDATE doc
			SET isclosed = true,
			    isref = true,
			    modifier_code = $3,
			    modified_at = now()
			WHERE businesscode = $1
			  AND docno = $2
			  AND transflag = $4
			  AND NOT isdelete
			  AND NOT iscancel
		`, scope.Company, docno, scope.Actor, doctype.GoodsReceive)
		if err != nil {
			return 0, fmt.Errorf("update gr status %s: %w", docno, err)
		}

		rowsAff, _ := res.RowsAffected()
		if rowsAff > 0 {
			updatedCount += int(rowsAff)

			// Record audit event in erp_doc_events (เหตุผลเก็บเป็น key แปลตามภาษาของผู้ดูตอนอ่าน)
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO erp_doc_events (businesscode, transflag, docno, action, to_status, reason, actor, created_at)
				VALUES ($1, $2, $3, 'adjust_status', 'closed', $4, $5, now())
			`, scope.Company, doctype.GoodsReceive, docno, eventReason(eventGRClosed), scope.Actor); err != nil {
				return 0, fmt.Errorf("record gr adjustment event %s: %w", docno, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit gr adjustment: %w", err)
	}

	return updatedCount, nil
}

// Handlers for microservice router
func (h *Http) listReceiptAdjustments(request microservice.IContext) error {
	access := h.authorize(request, actionRead)
	if access == nil {
		return nil
	}

	filter := ReceiptAdjustmentFilter{
		FromDate:   strings.TrimSpace(request.QueryParam("from_date")),
		ToDate:     strings.TrimSpace(request.QueryParam("to_date")),
		FromDocNo:  strings.TrimSpace(request.QueryParam("from_docno")),
		ToDocNo:    strings.TrimSpace(request.QueryParam("to_docno")),
		FromVendor: strings.TrimSpace(request.QueryParam("from_vendor")),
		ToVendor:   strings.TrimSpace(request.QueryParam("to_vendor")),
		FromDept:   strings.TrimSpace(request.QueryParam("from_dept")),
		ToDept:     strings.TrimSpace(request.QueryParam("to_dept")),
	}

	items, err := h.store.ListReceiptAdjustments(access.ctx, access.db, access.scope.Company, filter)
	if err != nil {
		return docFailure(request, err, false)
	}

	return success(request, map[string]any{
		"success": true,
		"data":    items,
		"count":   len(items),
	})
}

func (h *Http) processReceiptAdjustments(request microservice.IContext) error {
	access := h.authorize(request, actionUpdate)
	if access == nil {
		return nil
	}

	var payload ProcessReceiptAdjustmentsPayload
	if err := json.NewDecoder(request.Request().Body).Decode(&payload); err != nil {
		return fail(request, http.StatusBadRequest, "erp_err_invalid_payload")
	}

	if len(payload.DocNos) == 0 {
		return fail(request, http.StatusBadRequest, "erp_err_select_items")
	}

	count, err := h.store.ProcessReceiptAdjustments(access.ctx, access.db, access.scope, payload.DocNos)
	if err != nil {
		return docFailure(request, err, false)
	}

	return success(request, map[string]any{
		"success":        true,
		"adjusted_count": count,
		"message":        langText("erp_msg_receipt_adjusted", requestLanguage(request), strconv.Itoa(count)),
	})
}
