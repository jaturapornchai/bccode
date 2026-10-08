package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/doctype"
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

// BuildDebtReportQuery constructs parameterized SQL queries for the 8 AP and AR reports from gl_subledger tables.
func BuildDebtReportQuery(reportCode, businessCode string, fromDate, toDate time.Time, partyCode string, limit, offset int) (string, []any, error) {
	hasDate := !fromDate.IsZero() && !toDate.IsZero()
	companyParam := businessCode
	if companyParam == "" {
		companyParam = "01"
	}

	switch reportCode {
	case "ap_movement":
		query := `
SELECT
  d.payload->>'doc_date' AS docdate,
  COALESCE(d.payload->>'doc_no', SUBSTRING(d.id FROM 5)) AS docno,
  d.partner_code AS creditorcode,
  COALESCE(p.payload->>'name', d.partner_code) AS creditorname,
  CASE WHEN d.side = 2 THEN ` + strconv.Itoa(doctype.Purchase) + ` ELSE ` + strconv.Itoa(doctype.Pay) + ` END AS transflag,
  CASE WHEN d.side = 2 THEN 'ซื้อเชื่อ' ELSE 'จ่ายชำระหนี้' END AS transname,
  CASE WHEN d.side = 2 THEN d.amount ELSE 0 END AS totalamount,
  CASE WHEN d.side = 1 THEN d.amount ELSE 0 END AS paidamount,
  d.amount AS balanceamount
FROM gl_subledger_documents d
LEFT JOIN gl_subledger_partners p ON p.company = d.company AND p.code = d.partner_code
WHERE (d.company = $1 OR d.company = '01' OR $1 = '') AND d.ledger = 'ap'`
		args := []any{companyParam}
		argIdx := 2
		if hasDate {
			query += fmt.Sprintf(" AND (d.payload->>'doc_date')::date >= $%d AND (d.payload->>'doc_date')::date < $%d", argIdx, argIdx+1)
			args = append(args, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
			argIdx += 2
		}
		if partyCode != "" {
			query += fmt.Sprintf(" AND d.partner_code = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(" ORDER BY docdate ASC, docno ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ap_status":
		query := `
SELECT
  d.partner_code AS creditorcode,
  COALESCE(p.payload->>'name', d.partner_code) AS creditorname,
  0 AS creditday,
  '' AS phone,
  COALESCE(SUM(CASE WHEN d.side = 2 THEN d.amount ELSE 0 END), 0) AS totalamount,
  COALESCE(SUM(CASE WHEN d.side = 1 THEN d.amount ELSE 0 END), 0) AS paidamount,
  COALESCE((SUM(CASE WHEN d.side = 2 THEN d.amount ELSE 0 END) - SUM(CASE WHEN d.side = 1 THEN d.amount ELSE 0 END)), 0) AS balanceamount
FROM gl_subledger_documents d
LEFT JOIN gl_subledger_partners p ON p.company = d.company AND p.code = d.partner_code
WHERE (d.company = $1 OR d.company = '01' OR $1 = '') AND d.ledger = 'ap'`
		args := []any{companyParam}
		argIdx := 2
		if partyCode != "" {
			query += fmt.Sprintf(" AND d.partner_code = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(`
GROUP BY d.partner_code, p.payload
ORDER BY d.partner_code ASC
LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ap_outstanding":
		query := `
SELECT
  d.payload->>'doc_date' AS docdate,
  COALESCE(d.payload->>'doc_no', SUBSTRING(d.id FROM 4)) AS docno,
  d.partner_code AS creditorcode,
  COALESCE(p.payload->>'name', d.partner_code) AS creditorname,
  d.amount AS totalamount,
  COALESCE(SUM(s.amount), 0) AS paidamount,
  (d.amount - COALESCE(SUM(s.amount), 0)) AS balanceamount
FROM gl_subledger_documents d
LEFT JOIN gl_subledger_partners p ON p.company = d.company AND p.code = d.partner_code
LEFT JOIN gl_subledger_settlements s ON s.company = d.company AND s.debt_id = d.id AND s.reversed_at IS NULL
WHERE (d.company = $1 OR d.company = '01' OR $1 = '') AND d.ledger = 'ap' AND d.side = 2`
		args := []any{companyParam}
		argIdx := 2
		if hasDate {
			query += fmt.Sprintf(" AND (d.payload->>'doc_date')::date >= $%d AND (d.payload->>'doc_date')::date < $%d", argIdx, argIdx+1)
			args = append(args, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
			argIdx += 2
		}
		if partyCode != "" {
			query += fmt.Sprintf(" AND d.partner_code = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(`
GROUP BY d.company, d.id, d.payload, d.partner_code, p.payload, d.amount
HAVING (d.amount - COALESCE(SUM(s.amount), 0)) > 0
ORDER BY docdate ASC, docno ASC
LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ap_daily_payment":
		query := `
SELECT
  d.payload->>'doc_date' AS docdate,
  COALESCE(d.payload->>'doc_no', SUBSTRING(d.id FROM 5)) AS docno,
  d.partner_code AS creditorcode,
  COALESCE(p.payload->>'name', d.partner_code) AS creditorname,
  d.amount AS paidamount,
  0 AS balanceamount
FROM gl_subledger_documents d
LEFT JOIN gl_subledger_partners p ON p.company = d.company AND p.code = d.partner_code
WHERE (d.company = $1 OR d.company = '01' OR $1 = '') AND d.ledger = 'ap' AND d.side = 1`
		args := []any{companyParam}
		argIdx := 2
		if hasDate {
			query += fmt.Sprintf(" AND (d.payload->>'doc_date')::date >= $%d AND (d.payload->>'doc_date')::date < $%d", argIdx, argIdx+1)
			args = append(args, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
			argIdx += 2
		}
		if partyCode != "" {
			query += fmt.Sprintf(" AND d.partner_code = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(" ORDER BY docdate ASC, docno ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ar_movement":
		query := `
SELECT
  d.payload->>'doc_date' AS docdate,
  COALESCE(d.payload->>'doc_no', SUBSTRING(d.id FROM 5)) AS docno,
  d.partner_code AS debtorcode,
  COALESCE(p.payload->>'name', d.partner_code) AS debtorname,
  CASE WHEN d.side = 1 THEN ` + strconv.Itoa(doctype.SaleInvoice) + ` ELSE ` + strconv.Itoa(doctype.Paid) + ` END AS transflag,
  CASE WHEN d.side = 1 THEN 'ขายเชื่อ' ELSE 'รับชำระหนี้' END AS transname,
  CASE WHEN d.side = 1 THEN d.amount ELSE 0 END AS totalamount,
  CASE WHEN d.side = 2 THEN d.amount ELSE 0 END AS paidamount,
  d.amount AS balanceamount
FROM gl_subledger_documents d
LEFT JOIN gl_subledger_partners p ON p.company = d.company AND p.code = d.partner_code
WHERE (d.company = $1 OR d.company = '01' OR $1 = '') AND d.ledger = 'ar'`
		args := []any{companyParam}
		argIdx := 2
		if hasDate {
			query += fmt.Sprintf(" AND (d.payload->>'doc_date')::date >= $%d AND (d.payload->>'doc_date')::date < $%d", argIdx, argIdx+1)
			args = append(args, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
			argIdx += 2
		}
		if partyCode != "" {
			query += fmt.Sprintf(" AND d.partner_code = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(" ORDER BY docdate ASC, docno ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ar_status":
		query := `
SELECT
  d.partner_code AS debtorcode,
  COALESCE(p.payload->>'name', d.partner_code) AS debtorname,
  0 AS creditday,
  '' AS phone,
  COALESCE(SUM(CASE WHEN d.side = 1 THEN d.amount ELSE 0 END), 0) AS totalamount,
  COALESCE(SUM(CASE WHEN d.side = 2 THEN d.amount ELSE 0 END), 0) AS paidamount,
  COALESCE((SUM(CASE WHEN d.side = 1 THEN d.amount ELSE 0 END) - SUM(CASE WHEN d.side = 2 THEN d.amount ELSE 0 END)), 0) AS balanceamount
FROM gl_subledger_documents d
LEFT JOIN gl_subledger_partners p ON p.company = d.company AND p.code = d.partner_code
WHERE (d.company = $1 OR d.company = '01' OR $1 = '') AND d.ledger = 'ar'`
		args := []any{companyParam}
		argIdx := 2
		if partyCode != "" {
			query += fmt.Sprintf(" AND d.partner_code = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(`
GROUP BY d.partner_code, p.payload
ORDER BY d.partner_code ASC
LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ar_outstanding":
		query := `
SELECT
  d.payload->>'doc_date' AS docdate,
  COALESCE(d.payload->>'doc_no', SUBSTRING(d.id FROM 4)) AS docno,
  d.partner_code AS debtorcode,
  COALESCE(p.payload->>'name', d.partner_code) AS debtorname,
  d.amount AS totalamount,
  COALESCE(SUM(s.amount), 0) AS paidamount,
  (d.amount - COALESCE(SUM(s.amount), 0)) AS balanceamount
FROM gl_subledger_documents d
LEFT JOIN gl_subledger_partners p ON p.company = d.company AND p.code = d.partner_code
LEFT JOIN gl_subledger_settlements s ON s.company = d.company AND s.debt_id = d.id AND s.reversed_at IS NULL
WHERE (d.company = $1 OR d.company = '01' OR $1 = '') AND d.ledger = 'ar' AND d.side = 1`
		args := []any{companyParam}
		argIdx := 2
		if hasDate {
			query += fmt.Sprintf(" AND (d.payload->>'doc_date')::date >= $%d AND (d.payload->>'doc_date')::date < $%d", argIdx, argIdx+1)
			args = append(args, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
			argIdx += 2
		}
		if partyCode != "" {
			query += fmt.Sprintf(" AND d.partner_code = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(`
GROUP BY d.company, d.id, d.payload, d.partner_code, p.payload, d.amount
HAVING (d.amount - COALESCE(SUM(s.amount), 0)) > 0
ORDER BY docdate ASC, docno ASC
LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	case "ar_credit_limit":
		query := `
SELECT
  d.partner_code AS debtorcode,
  COALESCE(p.payload->>'name', d.partner_code) AS debtorname,
  COALESCE((SUM(CASE WHEN d.side = 1 THEN d.amount ELSE 0 END) - SUM(CASE WHEN d.side = 2 THEN d.amount ELSE 0 END)), 0) AS balanceamount,
  0 AS creditday
FROM gl_subledger_documents d
LEFT JOIN gl_subledger_partners p ON p.company = d.company AND p.code = d.partner_code
WHERE (d.company = $1 OR d.company = '01' OR $1 = '') AND d.ledger = 'ar'`
		args := []any{companyParam}
		argIdx := 2
		if partyCode != "" {
			query += fmt.Sprintf(" AND d.partner_code = $%d", argIdx)
			args = append(args, partyCode)
			argIdx++
		}
		query += fmt.Sprintf(`
GROUP BY d.partner_code, p.payload
ORDER BY d.partner_code ASC
LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
		args = append(args, limit, offset)
		return query, args, nil

	default:
		return "", nil, fmt.Errorf("unknown report code: %s", reportCode)
	}
}
