package httpapi

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"smlcloudplatform/pkg/microservice"
)

type BillingNotePrintItem struct {
	DocNo       string `json:"docno"`
	DocDate     string `json:"docdate"`
	CustCode    string `json:"custcode"`
	CustName    string `json:"custname"`
	TotalAmount string `json:"total_amount"`
	Remark      string `json:"remark"`
	SaleCode    string `json:"sale_code"`
	InvCount    int    `json:"inv_count"`
}

type BillingNotePrintFilter struct {
	FromDate  string
	ToDate    string
	FromDocNo string
	ToDocNo   string
	FromCust  string
	ToCust    string
	SaleCode  string
}

// ListBillingNotesForPrint ค้นหารายการใบวางบิลเพื่อนำมาพิมพ์เป็นชุด เทียบเท่า OnBttnShowBill ใน ARFrmPrintBillAutoView.cpp
func (s *Store) ListBillingNotesForPrint(ctx context.Context, db *sql.DB, company string, filter BillingNotePrintFilter) ([]BillingNotePrintItem, error) {
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, fmt.Sprintf("d.businesscode = $%d", argIdx))
	args = append(args, company)
	argIdx++

	whereClauses = append(whereClauses, fmt.Sprintf("d.transflag = $%d", argIdx))
	args = append(args, 26) // doctype billingnote
	argIdx++

	whereClauses = append(whereClauses, "NOT d.isdelete", "NOT d.iscancel")

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

	if filter.FromCust != "" {
		if filter.ToCust != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("d.custcode BETWEEN $%d AND $%d", argIdx, argIdx+1))
			args = append(args, filter.FromCust, filter.ToCust)
			argIdx += 2
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("d.custcode >= $%d", argIdx))
			args = append(args, filter.FromCust)
			argIdx++
		}
	} else if filter.ToCust != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.custcode <= $%d", argIdx))
		args = append(args, filter.ToCust)
		argIdx++
	}

	if filter.SaleCode != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.creator_code = $%d", argIdx))
		args = append(args, filter.SaleCode)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT d.docno,
		       COALESCE(d.docdatetime, now())::timestamptz,
		       COALESCE(d.custcode, ''),
		       COALESCE(d.custname, ''),
		       COALESCE(TRIM(TRAILING '.' FROM TRIM(TRAILING '0' FROM ROUND(d.totalamount, 2)::text)), '0'),
		       COALESCE(d.remark, ''),
		       COALESCE(d.creator_code, ''),
		       COALESCE((
		           SELECT COUNT(*)
		           FROM docdetail det
		           WHERE det.businesscode = d.businesscode
		             AND det.docno = d.docno
		             AND NOT det.iscancel
		       ), 0)
		FROM doc d
		WHERE %s
		ORDER BY d.docdatetime ASC, d.docno ASC, d.custcode ASC
		LIMIT 1000
	`, strings.Join(whereClauses, " AND "))

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list billing notes: %w", err)
	}
	defer rows.Close()

	var items []BillingNotePrintItem
	for rows.Next() {
		var docno, custcode, custname, totalAmount, remark, saleCode string
		var docDate time.Time
		var invCount int
		if err := rows.Scan(&docno, &docDate, &custcode, &custname, &totalAmount, &remark, &saleCode, &invCount); err != nil {
			return nil, err
		}

		if totalAmount == "" {
			totalAmount = "0"
		}

		items = append(items, BillingNotePrintItem{
			DocNo:       docno,
			DocDate:     docDate.Format("2006-01-02"),
			CustCode:    custcode,
			CustName:    custname,
			TotalAmount: totalAmount,
			Remark:      remark,
			SaleCode:    saleCode,
			InvCount:    invCount,
		})
	}
	if items == nil {
		items = []BillingNotePrintItem{}
	}
	return items, nil
}

func (h *Http) listBillingNotesForPrint(request microservice.IContext) error {
	access := h.authorize(request, actionRead)
	if access == nil {
		return nil
	}

	filter := BillingNotePrintFilter{
		FromDate:  strings.TrimSpace(request.QueryParam("from_date")),
		ToDate:    strings.TrimSpace(request.QueryParam("to_date")),
		FromDocNo: strings.TrimSpace(request.QueryParam("from_docno")),
		ToDocNo:   strings.TrimSpace(request.QueryParam("to_docno")),
		FromCust:  strings.TrimSpace(request.QueryParam("from_cust")),
		ToCust:    strings.TrimSpace(request.QueryParam("to_cust")),
		SaleCode:  strings.TrimSpace(request.QueryParam("sale_code")),
	}

	items, err := h.store.ListBillingNotesForPrint(access.ctx, access.db, access.scope.Company, filter)
	if err != nil {
		return docFailure(request, err, false)
	}

	return success(request, map[string]any{
		"success": true,
		"data":    items,
		"count":   len(items),
	})
}
