package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
)

// DebtReportRequest defines payload for AP/AR reports matching Champ specifications.
type DebtReportRequest struct {
	HoldingCode string `json:"holdingcode"`
	ReportCode  string `json:"reportcode"` // ap_movement, ap_status, ap_outstanding, ap_daily_payment, ar_movement, ar_status, ar_outstanding, ar_credit_limit
	FromDate    string `json:"fromdate,omitempty"`
	ToDate      string `json:"todate,omitempty"`
	PartyCode   string `json:"partycode,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// DebtReportHandler handles AP and AR reports from PostgreSQL (per-holding db).
func DebtReportHandler(c echo.Context) error {
	var req DebtReportRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	req.HoldingCode = strings.TrimSpace(req.HoldingCode)
	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode is required",
			"code":  "MISSING_HOLDING_CODE",
		})
	}

	req.ReportCode = strings.TrimSpace(req.ReportCode)
	if req.ReportCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "reportcode is required",
			"code":  "MISSING_REPORT_CODE",
		})
	}

	businessCode, scopeErr := reportCompanyScope(c, req.HoldingCode)
	if scopeErr != nil {
		return scopeErr.respond(c)
	}

	// Default limits
	if req.Limit <= 0 || req.Limit > 10000 {
		req.Limit = 1000
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Parse date range if provided
	var fromDate, toDate time.Time
	if req.FromDate != "" && req.ToDate != "" {
		var err error
		fromDate, err = time.Parse("2006-01-02", req.FromDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid from_date format. Use YYYY-MM-DD",
				"code":  "INVALID_FROM_DATE",
			})
		}
		toDate, err = time.Parse("2006-01-02", req.ToDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid to_date format. Use YYYY-MM-DD",
				"code":  "INVALID_TO_DATE",
			})
		}
		// Include entire toDate
		toDate = toDate.AddDate(0, 0, 1)
	}

	query, args, err := BuildDebtReportQuery(req.ReportCode, businessCode, fromDate, toDate, req.PartyCode, req.Limit, req.Offset)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
			"code":  "INVALID_REPORT_CODE",
		})
	}

	db, err := mypg.PgSqlFastConnect(req.HoldingCode)
	if err != nil {
		logger.Error("DebtReport: PgSqlFastConnect failed for %s: %v", req.HoldingCode, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error("DebtReport: query failed for %s/%s [%s]: %v", req.HoldingCode, businessCode, req.ReportCode, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get columns",
			"code":  "COLUMN_ERROR",
		})
	}

	results := make([]map[string]any, 0)
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
		"report_code": req.ReportCode,
	})
}

// BuildDebtReportQuery constructs parameterized SQL queries for the 8 AP and AR reports.
func BuildDebtReportQuery(reportCode, businessCode string, fromDate, toDate time.Time, partyCode string, limit, offset int) (string, []any, error) {
	hasDate := !fromDate.IsZero() && !toDate.IsZero()

	switch reportCode {
	case "ap_movement":
		query := `
SELECT
  TO_CHAR(t.docdate + INTERVAL '7 hour', 'YYYY-MM-DD') AS docdate,
  t.docno,
  t.creditorcode,
  COALESCE(
    CASE WHEN c.names IS NOT NULL AND jsonb_typeof(c.names) = 'array'
         THEN (SELECT nm->>'name' FROM jsonb_array_elements(c.names) nm WHERE nm->>'code' = 'th' LIMIT 1)
         ELSE '' END,
    t.creditorcode
  ) AS creditorname,
  t.transflag,
  CASE
    WHEN t.transflag = 36 THEN 'ซื้อเชื่อ'
    WHEN t.transflag = 38 THEN 'จ่ายชำระหนี้'
    WHEN t.transflag = 40 THEN 'ลดหนี้เจ้าหนี้'
    ELSE 'รายการเจ้าหนี้'
  END AS transname,
  COALESCE(t.totalamount, 0) AS totalamount,
  COALESCE(t.paidamount, 0) AS paidamount,
  COALESCE(t.balanceamount, 0) AS balanceamount
FROM public.creditortransaction t
LEFT JOIN public.creditor c ON c.code = t.creditorcode
WHERE t.iscancel = false`
		args := []any{}
		argIdx := 1
		if hasDate {
			query += fmt.Sprintf(" AND t.docdate >= $%d AND t.docdate < $%d", argIdx, argIdx+1)
			args = append(args, fromDate, toDate)
			argIdx += 2
		}
		if partyCode != "" {
			query += fmt.Sprintf(" AND t.creditorcode = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(" ORDER BY t.docdate ASC, t.docno ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ap_status":
		query := `
SELECT
  c.code AS creditorcode,
  COALESCE(
    CASE WHEN c.names IS NOT NULL AND jsonb_typeof(c.names) = 'array'
         THEN (SELECT nm->>'name' FROM jsonb_array_elements(c.names) nm WHERE nm->>'code' = 'th' LIMIT 1)
         ELSE '' END,
    c.code
  ) AS creditorname,
  COALESCE(c.creditday, 0) AS creditday,
  COALESCE(c.phoneprimary, '') AS phone,
  COALESCE(SUM(t.totalamount), 0) AS totalamount,
  COALESCE(SUM(t.paidamount), 0) AS paidamount,
  COALESCE(SUM(t.balanceamount), c.creditorbalanceamount, 0) AS balanceamount
FROM public.creditor c
LEFT JOIN public.creditortransaction t ON t.creditorcode = c.code AND t.iscancel = false`
		args := []any{}
		argIdx := 1
		if partyCode != "" {
			query += fmt.Sprintf(" WHERE c.code = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(`
GROUP BY c.code, c.names, c.creditday, c.phoneprimary, c.creditorbalanceamount
ORDER BY c.code ASC
LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ap_outstanding":
		query := `
SELECT
  TO_CHAR(t.docdate + INTERVAL '7 hour', 'YYYY-MM-DD') AS docdate,
  t.docno,
  t.creditorcode,
  COALESCE(
    CASE WHEN c.names IS NOT NULL AND jsonb_typeof(c.names) = 'array'
         THEN (SELECT nm->>'name' FROM jsonb_array_elements(c.names) nm WHERE nm->>'code' = 'th' LIMIT 1)
         ELSE '' END,
    t.creditorcode
  ) AS creditorname,
  COALESCE(t.totalamount, 0) AS totalamount,
  COALESCE(t.paidamount, 0) AS paidamount,
  COALESCE(t.balanceamount, 0) AS balanceamount
FROM public.creditortransaction t
LEFT JOIN public.creditor c ON c.code = t.creditorcode
WHERE t.iscancel = false
  AND t.balanceamount > 0`
		args := []any{}
		argIdx := 1
		if partyCode != "" {
			query += fmt.Sprintf(" AND t.creditorcode = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(" ORDER BY t.docdate ASC, t.docno ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ap_daily_payment":
		query := `
SELECT
  TO_CHAR(t.docdate + INTERVAL '7 hour', 'YYYY-MM-DD') AS docdate,
  t.docno,
  t.creditorcode,
  COALESCE(
    CASE WHEN c.names IS NOT NULL AND jsonb_typeof(c.names) = 'array'
         THEN (SELECT nm->>'name' FROM jsonb_array_elements(c.names) nm WHERE nm->>'code' = 'th' LIMIT 1)
         ELSE '' END,
    t.creditorcode
  ) AS creditorname,
  COALESCE(t.paidamount, 0) AS paidamount,
  COALESCE(t.balanceamount, 0) AS balanceamount
FROM public.creditortransaction t
LEFT JOIN public.creditor c ON c.code = t.creditorcode
WHERE t.iscancel = false
  AND t.paidamount > 0`
		args := []any{}
		argIdx := 1
		if hasDate {
			query += fmt.Sprintf(" AND t.docdate >= $%d AND t.docdate < $%d", argIdx, argIdx+1)
			args = append(args, fromDate, toDate)
			argIdx += 2
		}
		if partyCode != "" {
			query += fmt.Sprintf(" AND t.creditorcode = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(" ORDER BY t.docdate ASC, t.docno ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ar_movement":
		query := `
SELECT
  TO_CHAR(t.docdate + INTERVAL '7 hour', 'YYYY-MM-DD') AS docdate,
  t.docno,
  t.creditorcode AS debtorcode,
  COALESCE(
    CASE WHEN d.names IS NOT NULL AND jsonb_typeof(d.names) = 'array'
         THEN (SELECT nm->>'name' FROM jsonb_array_elements(d.names) nm WHERE nm->>'code' = 'th' LIMIT 1)
         ELSE '' END,
    t.creditorcode
  ) AS debtorname,
  t.transflag,
  CASE
    WHEN t.transflag = 44 THEN 'ขายเชื่อ'
    WHEN t.transflag = 48 THEN 'รับชำระหนี้'
    WHEN t.transflag = 50 THEN 'ลดหนี้ลูกหนี้'
    ELSE 'รายการลูกหนี้'
  END AS transname,
  COALESCE(t.totalamount, 0) AS totalamount,
  COALESCE(t.paidamount, 0) AS paidamount,
  COALESCE(t.balanceamount, 0) AS balanceamount
FROM public.debtortransaction t
LEFT JOIN public.debtor d ON d.code = t.creditorcode
WHERE t.iscancel = false`
		args := []any{}
		argIdx := 1
		if hasDate {
			query += fmt.Sprintf(" AND t.docdate >= $%d AND t.docdate < $%d", argIdx, argIdx+1)
			args = append(args, fromDate, toDate)
			argIdx += 2
		}
		if partyCode != "" {
			query += fmt.Sprintf(" AND t.creditorcode = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(" ORDER BY t.docdate ASC, t.docno ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ar_status":
		query := `
SELECT
  d.code AS debtorcode,
  COALESCE(
    CASE WHEN d.names IS NOT NULL AND jsonb_typeof(d.names) = 'array'
         THEN (SELECT nm->>'name' FROM jsonb_array_elements(d.names) nm WHERE nm->>'code' = 'th' LIMIT 1)
         ELSE '' END,
    d.code
  ) AS debtorname,
  COALESCE(d.creditday, 0) AS creditday,
  COALESCE(d.phoneprimary, '') AS phone,
  COALESCE(SUM(t.totalamount), 0) AS totalamount,
  COALESCE(SUM(t.paidamount), 0) AS paidamount,
  COALESCE(SUM(t.balanceamount), d.debtorbalanceamount, 0) AS balanceamount
FROM public.debtor d
LEFT JOIN public.debtortransaction t ON t.creditorcode = d.code AND t.iscancel = false`
		args := []any{}
		argIdx := 1
		if partyCode != "" {
			query += fmt.Sprintf(" WHERE d.code = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(`
GROUP BY d.code, d.names, d.creditday, d.phoneprimary, d.debtorbalanceamount
ORDER BY d.code ASC
LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ar_outstanding":
		query := `
SELECT
  TO_CHAR(t.docdate + INTERVAL '7 hour', 'YYYY-MM-DD') AS docdate,
  t.docno,
  t.creditorcode AS debtorcode,
  COALESCE(
    CASE WHEN d.names IS NOT NULL AND jsonb_typeof(d.names) = 'array'
         THEN (SELECT nm->>'name' FROM jsonb_array_elements(d.names) nm WHERE nm->>'code' = 'th' LIMIT 1)
         ELSE '' END,
    t.creditorcode
  ) AS debtorname,
  COALESCE(t.totalamount, 0) AS totalamount,
  COALESCE(t.paidamount, 0) AS paidamount,
  COALESCE(t.balanceamount, 0) AS balanceamount
FROM public.debtortransaction t
LEFT JOIN public.debtor d ON d.code = t.creditorcode
WHERE t.iscancel = false
  AND t.balanceamount > 0`
		args := []any{}
		argIdx := 1
		if partyCode != "" {
			query += fmt.Sprintf(" AND t.creditorcode = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(" ORDER BY t.docdate ASC, t.docno ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ar_credit_limit":
		query := `
SELECT
  d.code AS debtorcode,
  COALESCE(
    CASE WHEN d.names IS NOT NULL AND jsonb_typeof(d.names) = 'array'
         THEN (SELECT nm->>'name' FROM jsonb_array_elements(d.names) nm WHERE nm->>'code' = 'th' LIMIT 1)
         ELSE '' END,
    d.code
  ) AS debtorname,
  COALESCE(SUM(t.balanceamount), d.debtorbalanceamount, 0) AS balanceamount,
  COALESCE(d.creditday, 0) AS creditday
FROM public.debtor d
LEFT JOIN public.debtortransaction t ON t.creditorcode = d.code AND t.iscancel = false`
		args := []any{}
		argIdx := 1
		if partyCode != "" {
			query += fmt.Sprintf(" WHERE d.code = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(`
GROUP BY d.code, d.names, d.debtorbalanceamount, d.creditday
ORDER BY d.code ASC
LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	default:
		return "", nil, fmt.Errorf("unknown report code: %s", reportCode)
	}
}
