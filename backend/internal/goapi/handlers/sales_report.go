package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

// SalesReportRequest - request สำหรับรายงานขาย
type SalesReportRequest struct {
	HoldingCode   string   `json:"holdingcode"`
	FromDate      string   `json:"fromdate"`
	ToDate        string   `json:"todate"`
	BranchCodes   []string `json:"branchcodes,omitempty"`
	ProductCodes  []string `json:"productcodes,omitempty"`
	SortAscending bool     `json:"sortascending"`
	ReportType    string   `json:"reporttype"` // "header" or "detail"
	Limit         int      `json:"limit,omitempty"`
	Offset        int      `json:"offset,omitempty"`
}

// SalesReportByDocumentHandler - รายงานขายแยกตามเอกสาร (ปลอดภัยจาก SQL injection)
func SalesReportByDocumentHandler(c echo.Context) error {
	var req SalesReportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	// Validate required fields
	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode is required",
			"code":  "MISSING_HOLDING_CODE",
		})
	}

	if req.FromDate == "" || req.ToDate == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "from_date and to_date are required",
			"code":  "MISSING_DATE_RANGE",
		})
	}

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid from_date format. Use YYYY-MM-DD",
			"code":  "INVALID_FROM_DATE",
		})
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid to_date format. Use YYYY-MM-DD",
			"code":  "INVALID_TO_DATE",
		})
	}

	// Add 1 day to toDate to include the entire last day
	toDate = toDate.AddDate(0, 0, 1)

	// Set default limits
	if req.Limit <= 0 || req.Limit > 10000 {
		req.Limit = 1000
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Connect to database
	db, err := mypg.PgSqlFastConnect(req.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var query string
	var args []any

	// Build appropriate query based on report type
	if req.ReportType == "detail" {
		query, args = buildSalesReportDetailQuery(fromDate, toDate, req.BranchCodes, req.ProductCodes, req.SortAscending, req.Limit, req.Offset)
	} else {
		query, args = buildSalesReportHeaderQuery(fromDate, toDate, req.BranchCodes, req.ProductCodes, req.SortAscending, req.Limit, req.Offset)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}
	defer rows.Close()

	// Scan results
	results := make([]map[string]any, 0)
	columns, err := rows.Columns()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get columns",
			"code":  "COLUMN_ERROR",
		})
	}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":      "success",
		"data":        results,
		"count":       len(results),
		"limit":       req.Limit,
		"offset":      req.Offset,
		"report_type": req.ReportType,
	})
}

// buildSalesReportHeaderQuery - สร้าง query สำหรับรายงานขายแบบสรุป (ป้องกัน SQL injection)
func buildSalesReportHeaderQuery(fromDate, toDate time.Time, branchCodes, productCodes []string, sortAscending bool, limit, offset int) (string, []any) {
	args := make([]any, 0)
	argIndex := 1

	query := `
SELECT
  CAST(MAX(p.docdatetime) + INTERVAL '7 hour' AS DATE) as docdate,
  TO_CHAR(MAX(p.docdatetime) + INTERVAL '7 hour', 'HH24:MI:SS') as doctime,
  MAX(p.docdatetime) + INTERVAL '7 hour' as docdatetime,
  p.docno,
  COALESCE(doc.custcode, '') as debtorcode,
  COALESCE(d.name0, '') as debtorname,
  COALESCE(MAX(doc.totalamount), 0) as totalqty,
  (SUM(p.totalqty * p.price) * -1) as totalamount,
  AVG(p.price) as price,
  AVG(p.averagecost) as averagecost,
  (SUM(p.calcamount) * -1) as calcamount,
  ((SUM(p.totalqty * p.price) * -1) - (SUM(p.calcamount) * -1)) as grossprofit
FROM public.processstockcost p
LEFT JOIN public.doc doc ON p.docno = doc.docno
LEFT JOIN public.debtor d ON doc.custcode = d.code
WHERE p.transflag = 44
  AND p.docdatetime >= $1 AND p.docdatetime < $2`

	args = append(args, fromDate, toDate)
	argIndex = 3

	// Add branch filter (parameterized)
	if len(branchCodes) > 0 {
		query += fmt.Sprintf(" AND p.whcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(branchCodes))
		argIndex++
	}

	// Add product filter via subquery (parameterized)
	if len(productCodes) > 0 {
		query += fmt.Sprintf(`
  AND p.docno IN (
    SELECT DISTINCT sub.docno
    FROM public.processstockcost sub
    WHERE sub.transflag = 44
      AND sub.docdatetime >= $1 AND sub.docdatetime < $2
      AND sub.itemcode = ANY($%d)`, argIndex)
		args = append(args, pq.Array(productCodes))
		argIndex++

		// Add branch filter to subquery if present
		if len(branchCodes) > 0 {
			query += fmt.Sprintf(" AND sub.whcode = ANY($%d)", argIndex-2) // reuse branch param
		}
		query += ")"
	}

	query += "\nGROUP BY p.docno, doc.custcode, d.name0"

	if sortAscending {
		query += "\nORDER BY MAX(p.docdatetime) ASC, p.docno"
	} else {
		query += "\nORDER BY MAX(p.docdatetime) DESC, p.docno"
	}

	query += fmt.Sprintf("\nLIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	return query, args
}

// buildSalesReportDetailQuery - สร้าง query สำหรับรายงานขายแบบรายละเอียด (ป้องกัน SQL injection)
func buildSalesReportDetailQuery(fromDate, toDate time.Time, branchCodes, productCodes []string, sortAscending bool, limit, offset int) (string, []any) {
	args := make([]any, 0)
	argIndex := 1

	query := `
SELECT
  CAST(p.docdatetime + INTERVAL '7 hour' AS DATE) as docdate,
  TO_CHAR(p.docdatetime + INTERVAL '7 hour', 'HH24:MI:SS') as doctime,
  p.docdatetime + INTERVAL '7 hour' as docdatetime,
  p.docno,
  p.linenumber,
  p.itemcode,
  COALESCE(item_lookup.itemname, '') as itemname,
  p.barcode,
  COALESCE(unit_lookup.unitname, '') as unitname,
  (p.totalqty * -1) as totalqty,
  (p.totalqty * p.price * -1) as totalamount,
  (p.calcamount * -1) as calcamount,
  p.price,
  p.averagecost,
  ((p.totalqty * p.price * -1) - (p.calcamount * -1)) as grossprofit
FROM public.processstockcost p
LEFT JOIN LATERAL (
  SELECT STRING_AGG(DISTINCT pb.name0, ', ') AS itemname
  FROM public.productbarcode pb
  WHERE pb.itemcode = p.itemcode
    AND pb.barcoderefunitstand = 1
    AND pb.barcoderefunitdivide = 1
) item_lookup ON TRUE
LEFT JOIN LATERAL (
  SELECT STRING_AGG(DISTINCT pb.unitname, ', ') AS unitname
  FROM public.productbarcode pb
  WHERE pb.barcode = p.barcode
) unit_lookup ON TRUE
WHERE p.transflag = 44
  AND p.docdatetime >= $1 AND p.docdatetime < $2`

	args = append(args, fromDate, toDate)
	argIndex = 3

	// Add branch filter (parameterized)
	if len(branchCodes) > 0 {
		query += fmt.Sprintf(" AND p.whcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(branchCodes))
		argIndex++
	}

	// Add product filter via subquery (parameterized)
	if len(productCodes) > 0 {
		query += fmt.Sprintf(`
  AND p.docno IN (
    SELECT DISTINCT sub.docno
    FROM public.processstockcost sub
    WHERE sub.transflag = 44
      AND sub.docdatetime >= $1 AND sub.docdatetime < $2
      AND sub.itemcode = ANY($%d)`, argIndex)
		args = append(args, pq.Array(productCodes))
		argIndex++

		// Add branch filter to subquery if present
		if len(branchCodes) > 0 {
			query += fmt.Sprintf(" AND sub.whcode = ANY($%d)", argIndex-2)
		}
		query += ")"
	}

	if sortAscending {
		query += "\nORDER BY p.docdatetime ASC, p.docno, p.linenumber"
	} else {
		query += "\nORDER BY p.docdatetime DESC, p.docno, p.linenumber"
	}

	query += fmt.Sprintf("\nLIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	return query, args
}

// SalesReportSummaryHandler - สรุปยอดขายรวม
func SalesReportSummaryHandler(c echo.Context) error {
	var req SalesReportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" || req.FromDate == "" || req.ToDate == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode, from_date and to_date are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid from_date format",
			"code":  "INVALID_FROM_DATE",
		})
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid to_date format",
			"code":  "INVALID_TO_DATE",
		})
	}
	toDate = toDate.AddDate(0, 0, 1)

	db, err := mypg.PgSqlFastConnect(req.HoldingCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query, args := buildSalesReportSummaryQuery(fromDate, toDate, req.BranchCodes, req.ProductCodes)

	var totalDocuments int64
	var totalAmount, totalCost, totalProfit float64

	err = db.QueryRowContext(ctx, query, args...).Scan(
		&totalDocuments, &totalAmount, &totalCost, &totalProfit,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data": map[string]any{
			"total_documents": totalDocuments,
			"totalamount":     totalAmount,
			"totalcost":       totalCost,
			"total_profit":    totalProfit,
			"profit_margin":   calculateMargin(totalAmount, totalProfit),
		},
	})
}

// buildSalesReportSummaryQuery - สร้าง query สำหรับสรุปยอดขาย
func buildSalesReportSummaryQuery(fromDate, toDate time.Time, branchCodes, productCodes []string) (string, []any) {
	args := make([]any, 0)
	argIndex := 1

	query := `
SELECT
  COUNT(DISTINCT p.docno) as total_documents,
  COALESCE(SUM(p.totalqty * p.price) * -1, 0) as totalamount,
  COALESCE(SUM(p.calcamount) * -1, 0) as totalcost,
  COALESCE((SUM(p.totalqty * p.price) * -1) - (SUM(p.calcamount) * -1), 0) as total_profit
FROM public.processstockcost p
WHERE p.transflag = 44
  AND p.docdatetime >= $1 AND p.docdatetime < $2`

	args = append(args, fromDate, toDate)
	argIndex = 3

	if len(branchCodes) > 0 {
		query += fmt.Sprintf(" AND p.whcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(branchCodes))
		argIndex++
	}

	if len(productCodes) > 0 {
		query += fmt.Sprintf(" AND p.itemcode = ANY($%d)", argIndex)
		args = append(args, pq.Array(productCodes))
	}

	return query, args
}

// calculateMargin - คำนวณ profit margin
func calculateMargin(amount, profit float64) float64 {
	if amount == 0 {
		return 0
	}
	return (profit / amount) * 100
}
