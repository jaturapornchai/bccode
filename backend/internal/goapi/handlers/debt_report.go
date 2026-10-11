package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/doctype"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

// DebtReportRequest defines payload for AP/AR reports matching Champ specifications.
type DebtReportRequest struct {
	HoldingCode string `json:"holdingcode"`
	ReportCode  string `json:"reportcode"` // ap_movement, ap_status, ap_outstanding, ap_daily_payment, ap_aging, ar_movement, ar_status, ar_outstanding, ar_credit_limit, ar_aging
	FromDate    string `json:"fromdate,omitempty"`
	ToDate      string `json:"todate,omitempty"`
	PartyCode   string `json:"partycode,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
	// Aging only (ar_aging/ap_aging): Champ condition dialog IDD_AP_REP_AGING.
	AsOfDate string       `json:"asofdate,omitempty"` // default today (Asia/Bangkok)
	AgeBasis string       `json:"agebasis,omitempty"` // duedate (default) | docdate
	Ranges   []AgingRange `json:"ranges,omitempty"`   // 4 inclusive day ranges; default Champ 1-7/8-14/15-21/22-28
	OverDays *int         `json:"overdays,omitempty"` // "เกินกว่า" day; default 28
}

// AgingRange is one inclusive day range of the aging report (Champ "ช่วงที่ 1-4").
type AgingRange struct {
	From int `json:"from"`
	To   int `json:"to"`
}

const (
	agingRangeCount = 4
	agingMaxDays    = 36500
)

// Champ defaults when the ini has no value: APRepAgingDlg.cpp:271-279, AR_Rep_AgeByArDlg.cpp:383-391.
var defaultAgingRanges = []AgingRange{{1, 7}, {8, 14}, {15, 21}, {22, 28}}

const defaultAgingOverDays = 28

var agingRangeColumns = [agingRangeCount]string{"range1", "range2", "range3", "range4"}

// agingMoneyColumns - money columns of an aging row, in Champ column order (APRepAgingView.cpp DefaultData).
// earnestamount (Champ EARNEST_AMT) stays 0: no subledger document type identifies advance payments.
var agingMoneyColumns = []string{"notinrange", "range1", "range2", "range3", "range4", "overrange", "debtdecrease", "earnestamount", "totaldue"}

type agingCondition struct {
	AsOf     time.Time
	Basis    string
	Ranges   []AgingRange
	OverDays int
}

func isAgingReportCode(code string) bool {
	return code == "ar_aging" || code == "ap_aging"
}

// normalizeAgingCondition applies Champ defaults and returns the language key of the first invalid value.
func normalizeAgingCondition(req DebtReportRequest, now time.Time) (agingCondition, string) {
	cond := agingCondition{Basis: strings.ToLower(strings.TrimSpace(req.AgeBasis))}
	asOf := strings.TrimSpace(req.AsOfDate)
	if asOf == "" {
		asOf = now.In(ictZone).Format("2006-01-02")
	}
	parsed, err := time.Parse("2006-01-02", asOf)
	if err != nil {
		return cond, "aging_condition_invalid"
	}
	cond.AsOf = parsed
	if cond.Basis == "" {
		cond.Basis = "duedate"
	}
	if cond.Basis != "duedate" && cond.Basis != "docdate" {
		return cond, "aging_condition_invalid"
	}
	cond.Ranges = req.Ranges
	if len(cond.Ranges) == 0 {
		cond.Ranges = defaultAgingRanges
	}
	cond.OverDays = defaultAgingOverDays
	if req.OverDays != nil {
		cond.OverDays = *req.OverDays
	} else if len(req.Ranges) == agingRangeCount {
		cond.OverDays = req.Ranges[agingRangeCount-1].To
	}
	if !validAgingRanges(cond.Ranges, cond.OverDays) {
		return cond, "aging_ranges_invalid"
	}
	return cond, ""
}

// validAgingRanges - Champ IsStartReport needs every range filled; ranges must also ascend without overlap
// and "เกินกว่า" may not fall inside range 4, otherwise one age would land in two columns.
func validAgingRanges(ranges []AgingRange, overDays int) bool {
	if len(ranges) != agingRangeCount {
		return false
	}
	for i, r := range ranges {
		if r.From < 0 || r.From > r.To || r.To > agingMaxDays || (i > 0 && r.From <= ranges[i-1].To) {
			return false
		}
	}
	return overDays >= ranges[agingRangeCount-1].To && overDays <= agingMaxDays
}

// agingAge - Champ ProDiffYear (APRepAgingView.cpp:201-240): days from the reference date to the
// as-of date, counting the reference day itself once the date is reached (same day = 1).
func agingAge(asOf, ref time.Time) int {
	days := int(asOf.Sub(ref) / (24 * time.Hour))
	if days >= 0 {
		days++
	}
	return days
}

// agingBucket - Champ bucket rules (APRepAgingView.cpp:415-462, AR_Rep_AgeByArView.cpp ~582): not yet due
// (age < 1) is "not in range"; ranges are inclusive; age above overDays is the over column. Ages that fall
// in a gap between user ranges go to "not in range" so every row still adds up to its total.
func agingBucket(age int, ranges []AgingRange, overDays int) string {
	if age < 1 {
		return "notinrange"
	}
	for i, r := range ranges {
		if i < agingRangeCount && age >= r.From && age <= r.To {
			return agingRangeColumns[i]
		}
	}
	if age > overDays {
		return "overrange"
	}
	return "notinrange"
}

type agingTotals struct {
	amounts  map[string]decimal.Decimal
	docCount int64
}

// add places one (reference date, side) group of remaining balances. Balance-reducing documents
// (credit notes, unapplied payments) are shown negative in the debt-decrease column, as Champ does for credit notes.
func (t *agingTotals) add(cond agingCondition, refDate string, isDebt bool, amountText string, count int64) error {
	amount, err := decimal.NewFromString(amountText)
	if err != nil {
		return err
	}
	if t.amounts == nil {
		t.amounts = map[string]decimal.Decimal{}
	}
	column := "debtdecrease"
	if isDebt {
		ref, err := time.Parse("2006-01-02", refDate)
		if err != nil {
			return err
		}
		column = agingBucket(agingAge(cond.AsOf, ref), cond.Ranges, cond.OverDays)
	} else {
		amount = amount.Neg()
	}
	t.amounts[column] = t.amounts[column].Add(amount)
	t.amounts["totaldue"] = t.amounts["totaldue"].Add(amount)
	t.docCount += count
	return nil
}

func (t agingTotals) row() map[string]any {
	row := map[string]any{"doccount": t.docCount}
	for _, key := range agingMoneyColumns {
		row[key] = agingMoneyText(t.amounts[key])
	}
	return row
}

// agingMoneyText - numeric text with 2 decimals, never rounding away stored precision.
func agingMoneyText(value decimal.Decimal) string {
	if value.Equal(value.Round(2)) {
		return value.StringFixed(2)
	}
	return value.String()
}

type agingQuery struct {
	rows, totals, count string
	args                []any // company, as-of, basis, party
	rowArgs             []any // args + limit, offset
}

// buildAgingQuery - remaining balance as of the date per document, grouped by partner and reference date.
// Bucketing happens in Go (agingBucket) so the Champ rules live in one tested place.
func buildAgingQuery(reportCode, company string, cond agingCondition, partyCode string, limit, offset int) (agingQuery, error) {
	ledger := ""
	debtSide := "d.side = 1"
	switch reportCode {
	case "ar_aging":
		ledger = "ar"
	case "ap_aging":
		ledger = "ap"
		// GL-recorded documents keep balance_side (1 = debt) in side; ERP postings keep the control-account
		// side, where a payable is a credit (posting.go subledgerSpecs "purchase" side 2).
		debtSide = "CASE WHEN d.payload->>'balance_side' IS NOT NULL THEN d.side = 1 ELSE d.side = 2 END"
	default:
		return agingQuery{}, fmt.Errorf("unknown aging report code: %s", reportCode)
	}
	// Only journals posted as of the date count, as in the GL as-of outstanding (generalledger subledgerReadCTE):
	// draft documents/settlements and journals reversed by then drop out, so the aging ties to the control account.
	// GL-recorded documents (balance_side) owe their posted allocations; ERP documents owe their amount once
	// the creating journal is posted.
	base := `WITH posted_journals AS (
 SELECT r.id FROM gl_records r
 WHERE r.company = $1 AND r.kind = 'journals' AND NOT COALESCE((r.payload->>'isdeleted')::boolean, false)
  AND r.payload->>'date' <= to_char($2::date, 'YYYY-MM-DD')
  AND (r.payload->>'postedat' IS NULL OR (r.payload->>'postedat')::timestamptz < (($2::date + 1)::timestamp AT TIME ZONE 'Asia/Bangkok'))
  AND (r.payload->>'status' = 'posted' OR (r.payload->>'status' = 'reversed' AND NOT EXISTS (
   SELECT 1 FROM gl_records v WHERE v.company = $1 AND v.kind = 'journals' AND v.payload->>'reversalof' = r.id
    AND v.payload->>'date' <= to_char($2::date, 'YYYY-MM-DD')
    AND (v.payload->>'postedat' IS NULL OR (v.payload->>'postedat')::timestamptz < (($2::date + 1)::timestamp AT TIME ZONE 'Asia/Bangkok')))))
), posted_allocations AS (
 SELECT a.document_id, SUM(a.amount) AS amount
 FROM gl_subledger_allocations a JOIN posted_journals j ON j.id = a.journal_id
 WHERE a.company = $1 AND a.created_at < (($2::date + 1)::timestamp AT TIME ZONE 'Asia/Bangkok')
  AND (a.reversed_at IS NULL OR a.reversed_at >= (($2::date + 1)::timestamp AT TIME ZONE 'Asia/Bangkok') OR a.reversed_effective_date > $2::date)
 GROUP BY a.document_id
), posted_settlements AS (
 SELECT x.document_id, SUM(s.amount) AS amount
 FROM gl_subledger_settlements s JOIN posted_journals j ON j.id = s.journal_id
 CROSS JOIN LATERAL (VALUES (s.debt_id), (s.payment_id)) x(document_id)
 WHERE s.company = $1 AND s.created_at < (($2::date + 1)::timestamp AT TIME ZONE 'Asia/Bangkok') AND s.settlement_date <= $2::date
  AND (s.reversed_at IS NULL OR s.reversed_at >= (($2::date + 1)::timestamp AT TIME ZONE 'Asia/Bangkok'))
 GROUP BY x.document_id
), docs AS (
 SELECT d.partner_code, ` + debtSide + ` AS isdebt,
  COALESCE(NULLIF(d.payload->>'doc_date',''), NULLIF(d.payload->>'document_date',''))::date AS docdate,
  COALESCE(NULLIF(d.payload->>'due_date',''), NULLIF(d.payload->>'doc_date',''), NULLIF(d.payload->>'document_date',''))::date AS duedate,
  CASE WHEN d.payload->>'balance_side' IS NOT NULL THEN COALESCE(pa.amount, 0)
   WHEN cj.id IS NOT NULL THEN d.amount ELSE 0 END - COALESCE(ps.amount, 0) AS remaining
 FROM gl_subledger_documents d
 LEFT JOIN posted_journals cj ON cj.id = d.created_journal_id
 LEFT JOIN posted_allocations pa ON pa.document_id = d.id
 LEFT JOIN posted_settlements ps ON ps.document_id = d.id
 WHERE d.company = $1 AND d.ledger = '` + ledger + `' AND ($4 = '' OR d.partner_code = $4)
), aged AS (
 SELECT partner_code, isdebt, CASE WHEN $3 = 'docdate' THEN docdate ELSE duedate END AS refdate, remaining
 FROM docs WHERE docdate <= $2::date AND remaining <> 0
)
`
	args := []any{company, cond.AsOf.Format("2006-01-02"), cond.Basis, partyCode}
	return agingQuery{
		rows: base + `, parties AS (
 SELECT partner_code FROM aged GROUP BY partner_code ORDER BY partner_code ASC LIMIT $5 OFFSET $6
)
SELECT a.partner_code AS partycode,
 COALESCE(NULLIF(p.payload->>'name_th',''), NULLIF(p.payload->>'name',''), a.partner_code) AS partyname,
 to_char(a.refdate, 'YYYY-MM-DD') AS refdate, a.isdebt, SUM(a.remaining)::text AS amount, COUNT(*) AS doccount
FROM aged a
JOIN parties t ON t.partner_code = a.partner_code
LEFT JOIN gl_subledger_partners p ON p.company = $1 AND p.code = a.partner_code
GROUP BY a.partner_code, p.payload, a.refdate, a.isdebt
ORDER BY a.partner_code ASC, a.refdate ASC`,
		totals:  base + `SELECT to_char(refdate, 'YYYY-MM-DD'), isdebt, SUM(remaining)::text, COUNT(*) FROM aged GROUP BY refdate, isdebt`,
		count:   base + `SELECT COUNT(DISTINCT partner_code) FROM aged`,
		args:    args,
		rowArgs: append(append([]any{}, args...), limit, offset),
	}, nil
}

// debtAgingReport - รายงานอายุลูกหนี้/เจ้าหนี้ (Champ 5093/5016) ตามวันที่ ช่วงอายุ และวิธีนับที่ผู้ใช้กำหนด
func debtAgingReport(c echo.Context, req DebtReportRequest, businessCode string) error {
	cond, invalidKey := normalizeAgingCondition(req, time.Now())
	if invalidKey != "" {
		code := "AGING_CONDITION_INVALID"
		if invalidKey == "aging_ranges_invalid" {
			code = "AGING_RANGES_INVALID"
		}
		message := language.Text(invalidKey, taxRequestLanguage(c))
		return c.JSON(http.StatusBadRequest, map[string]any{"success": false, "code": code, "error": message, "message": message})
	}
	q, err := buildAgingQuery(req.ReportCode, businessCode, cond, strings.TrimSpace(req.PartyCode), req.Limit, req.Offset)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error(), "code": "INVALID_REPORT_CODE"})
	}
	db, err := mypg.PgSqlFastConnect(req.HoldingCode)
	if err != nil {
		logger.Error("DebtAging: PgSqlFastConnect failed for %s: %v", req.HoldingCode, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database connection failed", "code": "DB_CONNECTION_ERROR"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	data, totals, partyCount, err := runAgingQuery(ctx, db, q, cond)
	if err != nil {
		logger.Error("DebtAging: query failed for %s/%s [%s]: %v", req.HoldingCode, businessCode, req.ReportCode, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Query execution failed", "code": "QUERY_ERROR"})
	}
	return c.JSON(http.StatusOK, map[string]any{
		"status":      "success",
		"data":        data,
		"totals":      totals,
		"count":       len(data),
		"totalrows":   partyCount,
		"limit":       req.Limit,
		"offset":      req.Offset,
		"report_code": req.ReportCode,
		"asofdate":    cond.AsOf.Format("2006-01-02"),
		"agebasis":    cond.Basis,
		"ranges":      cond.Ranges,
		"overdays":    cond.OverDays,
	})
}

// runAgingQuery reads the page, the all-partner totals and the partner count in one snapshot.
func runAgingQuery(ctx context.Context, db *sql.DB, q agingQuery, cond agingCondition) ([]map[string]any, map[string]any, int64, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, nil, 0, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, q.rows, q.rowArgs...)
	if err != nil {
		return nil, nil, 0, err
	}
	data := []map[string]any{}
	var current agingTotals
	var currentCode, currentName string
	flush := func() {
		if currentCode == "" {
			return
		}
		row := current.row()
		row["partycode"], row["partyname"] = currentCode, currentName
		data = append(data, row)
	}
	for rows.Next() {
		var code, name, refDate, amount string
		var isDebt bool
		var count int64
		if err = rows.Scan(&code, &name, &refDate, &isDebt, &amount, &count); err != nil {
			rows.Close()
			return nil, nil, 0, err
		}
		if code != currentCode {
			flush()
			current, currentCode, currentName = agingTotals{}, code, name
		}
		if err = current.add(cond, refDate, isDebt, amount, count); err != nil {
			rows.Close()
			return nil, nil, 0, err
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, nil, 0, err
	}
	flush()

	totalRows, err := tx.QueryContext(ctx, q.totals, q.args...)
	if err != nil {
		return nil, nil, 0, err
	}
	var totals agingTotals
	for totalRows.Next() {
		var refDate, amount string
		var isDebt bool
		var count int64
		if err = totalRows.Scan(&refDate, &isDebt, &amount, &count); err != nil {
			totalRows.Close()
			return nil, nil, 0, err
		}
		if err = totals.add(cond, refDate, isDebt, amount, count); err != nil {
			totalRows.Close()
			return nil, nil, 0, err
		}
	}
	err = totalRows.Err()
	totalRows.Close()
	if err != nil {
		return nil, nil, 0, err
	}

	var partyCount int64
	if err = tx.QueryRowContext(ctx, q.count, q.args...).Scan(&partyCount); err != nil {
		return nil, nil, 0, err
	}
	return data, totals.row(), partyCount, tx.Commit()
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

	if isAgingReportCode(req.ReportCode) {
		return debtAgingReport(c, req, businessCode)
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
