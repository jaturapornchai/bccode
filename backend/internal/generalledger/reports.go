package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type reportContext struct {
	tx     *sql.Tx
	scope  Scope
	query  ReportQuery
	fiscal FiscalYear
	args   []any
	filter string
}

func (p *Postgres) Report(ctx context.Context, scope Scope, name string, q ReportQuery) (Report, error) {
	if name == "ar-outstanding" || name == "ap-outstanding" || name == "bank-unmatched" {
		return p.SubledgerReport(ctx, scope, name, q)
	}
	empty := Report{Columns: []ReportColumn{}, Rows: []map[string]string{}, Totals: map[string]string{}, Warnings: []string{}}
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return empty, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return empty, err
	}
	defer tx.Rollback()
	rc, err := newReportContext(ctx, tx, scope, q)
	if err != nil {
		return empty, err
	}
	var report Report
	switch name {
	case "ledger":
		report, err = rc.ledger(ctx)
	case "trialbalance", "workingpaper":
		report, err = rc.trialBalance(ctx, name == "workingpaper")
	case "pnl":
		report, err = rc.profitLoss(ctx)
	case "balancesheet":
		report, err = rc.balanceSheet(ctx)
	case "annual-balances":
		report, err = rc.annualBalances(ctx)
	case "allocate":
		report, err = rc.allocation(ctx)
	case "daily-check":
		report, err = rc.dailyCheck(ctx)
	case "cashflow":
		report, err = rc.cashFlow(ctx)
	case "cashflowforecast":
		report, err = rc.cashFlowForecast(ctx)
	case "project-pnl", "dimensionpnl", "projectsummary":
		report, err = rc.dimensionProfit(ctx, name)
	case "dashboard", "executivesummary", "financialgraphs":
		report, err = rc.summary(ctx, name)
	case "gljournal":
		report, err = rc.glJournal(ctx)
	case "budgetcomparison":
		report, err = rc.budgetComparison(ctx)
	default:
		return empty, fmt.Errorf("รายงานนี้ยังไม่มีรูปแบบที่ยืนยันแล้ว")
	}
	if err != nil {
		return empty, err
	}
	report.AsOf = rc.query.To
	if report.Warnings == nil {
		report.Warnings = []string{}
	}
	if err = tx.Commit(); err != nil {
		return empty, err
	}
	return report, nil
}

func newReportContext(ctx context.Context, tx *sql.Tx, scope Scope, q ReportQuery) (reportContext, error) {
	r := reportContext{tx: tx, scope: scope, query: q}
	if scope.Company == "" {
		return r, fmt.Errorf("กรุณาเลือกบริษัท")
	}
	if scope.Branch != "" {
		if q.BranchCode != "" && q.BranchCode != scope.Branch {
			return r, fmt.Errorf("ไม่มีสิทธิ์รายงานของสาขานี้")
		}
		q.BranchCode = scope.Branch
	}
	if q.From != "" && !validDate(q.From) || q.To != "" && !validDate(q.To) {
		return r, fmt.Errorf("วันที่รายงานไม่ถูกต้อง")
	}
	date := q.To
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	rows, err := tx.QueryContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND NOT COALESCE((payload->>'isdeleted')::boolean,false) AND (($2<>'' AND code=$2) OR ($2='' AND (payload->>'startdate')::date<=$3::date AND (payload->>'enddate')::date>=$3::date)) ORDER BY code LIMIT 2`, scope.Company, q.FiscalYear, date)
	if err != nil {
		return r, err
	}
	count := 0
	for rows.Next() {
		var data []byte
		if err = rows.Scan(&data); err != nil {
			rows.Close()
			return r, err
		}
		if err = json.Unmarshal(data, &r.fiscal); err != nil {
			rows.Close()
			return r, err
		}
		count++
	}
	rowErr := rows.Err()
	rows.Close()
	if rowErr != nil {
		return r, rowErr
	}
	if count != 1 {
		return r, fmt.Errorf("กรุณาเลือกปีบัญชีให้ชัดเจนก่อนออกรายงาน")
	}
	if q.From == "" {
		q.From = r.fiscal.StartDate
	}
	if q.To == "" {
		q.To = r.fiscal.EndDate
	}
	if q.From > q.To || q.From < r.fiscal.StartDate || q.To > r.fiscal.EndDate {
		return r, fmt.Errorf("ช่วงรายงานต้องอยู่ในปีบัญชีที่เลือก")
	}
	q.FiscalYear = r.fiscal.Code
	q.Page, q.Limit = pageBounds(q.Page, q.Limit)
	r.query = q
	r.args = []any{scope.Company, q.FiscalYear, q.From, q.To, q.AccountCode, q.BranchCode, q.DepartmentCode, q.ProjectCode, q.BookCode}
	r.filter = `company=$1 AND fiscal_year=$2 AND $3::date<=$4::date AND entry_date<=$4::date AND ($5='' OR account_code=$5) AND ($6='' OR branch_code=$6) AND ($7='' OR department_code=$7) AND ($8='' OR project_code=$8) AND ($9='' OR book_code=$9)`
	return r, nil
}

func textColumn(key, label string) ReportColumn { return ReportColumn{Key: key, Label: label} }
func amountColumn(key, label string) ReportColumn {
	return ReportColumn{Key: key, Label: label, Amount: true}
}

type reportTotal struct {
	key  string
	expr string
}

func (r reportContext) run(ctx context.Context, cte, order string, columns []ReportColumn, totalKeys []string, metadata ...string) (Report, error) {
	totals := make([]reportTotal, len(totalKeys))
	for i, key := range totalKeys {
		totals[i] = reportTotal{key: key, expr: key + "::numeric"}
	}
	return r.runWithTotals(ctx, cte, order, columns, totals, metadata...)
}

// The rows CTE must expose only text columns and, where needed, private sort
// columns. PostgreSQL performs all accounting arithmetic before ::text scanning.
// Consolidates total count, window totals, and paginated rows into a single statement.
func (r reportContext) runWithTotals(ctx context.Context, cte, order string, columns []ReportColumn, totals []reportTotal, metadata ...string) (Report, error) {
	result := Report{Columns: columns, Rows: []map[string]string{}, Totals: map[string]string{}, Warnings: []string{}}
	for _, t := range totals {
		result.Totals[t.key] = "0"
	}
	selected := make([]string, len(columns))
	nullCols := make([]string, len(columns))
	for i, column := range columns {
		selected[i] = column.Key
		nullCols[i] = "NULL::text"
	}
	// Source identifiers and sort keys do not become visible/CSV columns.
	for _, key := range metadata {
		selected = append(selected, key)
		nullCols = append(nullCols, "NULL::text")
	}
	windowSums := ""
	fallbackSums := ""
	for _, t := range totals {
		expr := t.expr
		if expr == "" {
			expr = t.key + "::numeric"
		}
		windowSums += ", COALESCE(SUM(" + expr + ") OVER(), 0)::text AS __sum_" + t.key
		fallbackSums += ", COALESCE(SUM(" + expr + "), 0)::text"
	}
	limitArg := fmt.Sprintf("$%d", len(r.args)+1)
	offsetArg := fmt.Sprintf("$%d", len(r.args)+2)
	nullColsStr := ""
	if len(nullCols) > 0 {
		nullColsStr = ", " + strings.Join(nullCols, ", ")
	}

	query := cte + `, numbered AS (
		SELECT count(*) OVER() AS __total_rows` + windowSums + `, ` + strings.Join(selected, ", ") + ` FROM result
	),
	page AS (
		SELECT * FROM numbered ORDER BY ` + order + ` LIMIT ` + limitArg + ` OFFSET ` + offsetArg + `
	)
	SELECT 0 AS __is_fallback, * FROM page
	UNION ALL
	SELECT 1 AS __is_fallback, count(*)` + fallbackSums + nullColsStr + ` FROM result WHERE NOT EXISTS (SELECT 1 FROM page)`

	args := append(append([]any{}, r.args...), r.query.Limit, (r.query.Page-1)*r.query.Limit)
	rows, err := r.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	hasReadTotals := false
	for rows.Next() {
		var isFallback int
		var totalRows int64
		sumVals := make([]string, len(totals))
		colVals := make([]sql.NullString, len(selected))

		targets := make([]any, 0, 2+len(totals)+len(columns))
		targets = append(targets, &isFallback, &totalRows)
		for i := range sumVals {
			targets = append(targets, &sumVals[i])
		}
		for i := range colVals {
			targets = append(targets, &colVals[i])
		}

		if err = rows.Scan(targets...); err != nil {
			return result, err
		}

		if !hasReadTotals {
			result.TotalRows = totalRows
			for i, t := range totals {
				result.Totals[t.key] = sumVals[i]
			}
			hasReadTotals = true
		}

		if isFallback == 0 {
			record := make(map[string]string, len(columns))
			for i, col := range columns {
				record[col.Key] = colVals[i].String
			}
			for i, key := range metadata {
				if !strings.HasPrefix(key, "__") {
					record[key] = colVals[len(columns)+i].String
				}
			}
			result.Rows = append(result.Rows, record)
		}
	}
	return result, rows.Err()
}

func (r reportContext) setTotals(ctx context.Context, report *Report, query string, keys ...string) error {
	values := make([]string, len(keys))
	targets := make([]any, len(keys))
	for i := range values {
		targets[i] = &values[i]
	}
	if err := r.tx.QueryRowContext(ctx, query, r.args...).Scan(targets...); err != nil {
		return err
	}
	for i, key := range keys {
		report.Totals[key] = values[i]
	}
	return nil
}

func (r reportContext) base() string {
	return `WITH filtered AS (SELECT * FROM gl_lines WHERE ` + r.filter + `)`
}
func (r reportContext) trialCTE() string {
	return r.base() + `, balances AS (
    SELECT account_code,MAX(account_name) AS account_name,MAX(account_type) AS account_type,
      SUM(CASE WHEN entry_date<$3::date OR kind='opening' THEN debit-credit ELSE 0 END) AS opening,
      SUM(CASE WHEN entry_date>=$3::date AND kind<>'opening' THEN debit ELSE 0 END) AS debit,
      SUM(CASE WHEN entry_date>=$3::date AND kind<>'opening' THEN credit ELSE 0 END) AS credit,
      SUM(debit-credit) AS balance FROM filtered GROUP BY account_code
    )`
}

func (r reportContext) trialBalance(ctx context.Context, working bool) (Report, error) {
	columns := []ReportColumn{textColumn("accountcode", "รหัสบัญชี"), textColumn("accountname", "ชื่อบัญชี"), textColumn("accounttype", "หมวดบัญชี"), amountColumn("openingdebit", "ยกมาเดบิต"), amountColumn("openingcredit", "ยกมาเครดิต"), amountColumn("debit", "เคลื่อนไหวเดบิต"), amountColumn("credit", "เคลื่อนไหวเครดิต"), amountColumn("endingdebit", "คงเหลือเดบิต"), amountColumn("endingcredit", "คงเหลือเครดิต"), amountColumn("balance", "สุทธิเดบิตลบเครดิต")}
	cte := r.trialCTE() + `, result AS (SELECT account_code AS accountcode,account_name AS accountname,account_type AS accounttype,GREATEST(opening,0)::text AS openingdebit,GREATEST(-opening,0)::text AS openingcredit,debit::text,credit::text,GREATEST(balance,0)::text AS endingdebit,GREATEST(-balance,0)::text AS endingcredit,balance::text`
	totals := []string{"openingdebit", "openingcredit", "debit", "credit", "endingdebit", "endingcredit", "balance"}
	if working {
		cte += `,(CASE WHEN account_type IN ('income','expense') THEN GREATEST(balance,0) ELSE 0 END)::text AS pnldebit,(CASE WHEN account_type IN ('income','expense') THEN GREATEST(-balance,0) ELSE 0 END)::text AS pnlcredit,(CASE WHEN account_type IN ('asset','liability','equity') THEN GREATEST(balance,0) ELSE 0 END)::text AS bsdebit,(CASE WHEN account_type IN ('asset','liability','equity') THEN GREATEST(-balance,0) ELSE 0 END)::text AS bscredit`
		columns = append(columns, amountColumn("pnldebit", "กำไรขาดทุนเดบิต"), amountColumn("pnlcredit", "กำไรขาดทุนเครดิต"), amountColumn("bsdebit", "ฐานะการเงินเดบิต"), amountColumn("bscredit", "ฐานะการเงินเครดิต"))
		totals = append(totals, "pnldebit", "pnlcredit", "bsdebit", "bscredit")
	}
	cte += ` FROM balances)`
	result, err := r.run(ctx, cte, "accountcode", columns, totals)
	result.Totals["difference"] = result.Totals["balance"]
	return result, err
}

func (r reportContext) ledger(ctx context.Context) (Report, error) {
	cte := r.base() + `, openings AS (SELECT account_code,SUM(debit-credit) AS opening FROM filtered WHERE entry_date<$3::date OR kind='opening' GROUP BY account_code), movements AS (SELECT * FROM filtered WHERE entry_date>=$3::date AND kind<>'opening'), result AS (
      SELECT m.account_code AS accountcode,m.account_name AS accountname,m.entry_date::text AS date,m.doc_no AS docno,m.book_code AS bookcode,m.branch_code AS branchcode,m.department_code AS departmentcode,m.project_code AS projectcode,m.description,
      m.debit::text,m.credit::text,COALESCE(o.opening,0)::text AS opening,(COALESCE(o.opening,0)+SUM(m.debit-m.credit) OVER(PARTITION BY m.account_code ORDER BY m.entry_date,m.doc_no,m.journal_id,m.line_no ROWS UNBOUNDED PRECEDING))::text AS balance,m.journal_id AS journalid,m.line_no::text AS __line_no FROM movements m LEFT JOIN openings o USING(account_code))`
	cols := []ReportColumn{textColumn("accountcode", "รหัสบัญชี"), textColumn("accountname", "ชื่อบัญชี"), textColumn("date", "วันที่"), textColumn("docno", "เลขที่เอกสาร"), textColumn("bookcode", "สมุดรายวัน"), textColumn("branchcode", "สาขา"), textColumn("departmentcode", "แผนก"), textColumn("projectcode", "โครงการ"), textColumn("description", "รายละเอียด"), amountColumn("opening", "ยอดยกมา"), amountColumn("debit", "เดบิต"), amountColumn("credit", "เครดิต"), amountColumn("balance", "ยอดคงเหลือ")}
	result, err := r.run(ctx, cte, "accountcode,date,docno,journalid,__line_no::integer", cols, []string{"debit", "credit"}, "journalid", "__line_no")
	if err != nil {
		return result, err
	}
	err = r.setTotals(ctx, &result, r.base()+` SELECT COALESCE(SUM(CASE WHEN entry_date<$3::date OR kind='opening' THEN debit-credit ELSE 0 END),0)::text,COALESCE(SUM(debit-credit),0)::text FROM filtered`, "opening", "balance")
	return result, err
}

func (r reportContext) pnlCTE() string {
	return r.base() + `, pnl AS (SELECT account_code,MAX(account_name) AS account_name,MAX(account_type) AS account_type,SUM(debit) AS debit,SUM(credit) AS credit,SUM(debit-credit) AS balance FROM filtered WHERE entry_date>=$3::date AND kind NOT IN ('opening','closing') AND account_type IN ('income','expense') GROUP BY account_code)`
}
func (r reportContext) profitLoss(ctx context.Context) (Report, error) {
	cte := r.pnlCTE() + `, result AS (SELECT account_code AS accountcode,account_name AS accountname,account_type AS accounttype,debit::text,credit::text,balance::text,(CASE WHEN account_type='income' THEN -balance ELSE balance END)::text AS amount FROM pnl)`
	totals := []reportTotal{
		{key: "debit", expr: "debit::numeric"},
		{key: "credit", expr: "credit::numeric"},
		{key: "revenue", expr: "CASE WHEN accounttype='income' THEN -balance::numeric ELSE 0 END"},
		{key: "expense", expr: "CASE WHEN accounttype='expense' THEN balance::numeric ELSE 0 END"},
		{key: "profit", expr: "-balance::numeric"},
	}
	columns := []ReportColumn{textColumn("accountcode", "รหัสบัญชี"), textColumn("accountname", "ชื่อบัญชี"), textColumn("accounttype", "หมวดบัญชี"), amountColumn("debit", "เดบิต"), amountColumn("credit", "เครดิต"), amountColumn("balance", "สุทธิเดบิตลบเครดิต"), amountColumn("amount", "จำนวนเงิน")}
	return r.runWithTotals(ctx, cte, "accounttype,accountcode", columns, totals)
}

func (r reportContext) balanceSheet(ctx context.Context) (Report, error) {
	cte := r.trialCTE() + `, statements AS (
      SELECT account_code AS accountcode,account_name AS accountname,account_type AS accounttype,CASE WHEN account_type='asset' THEN balance ELSE -balance END AS amount FROM balances WHERE account_type IN ('asset','liability','equity')
      UNION ALL SELECT '__current_earnings__','กำไรขาดทุนที่ยังไม่ปิดเข้ากำไรสะสม','equity',COALESCE(SUM(-balance),0) FROM balances WHERE account_type IN ('income','expense')
    ), result AS (SELECT accountcode,accountname,accounttype,amount::text FROM statements)`
	totals := []reportTotal{
		{key: "assets", expr: "CASE WHEN accounttype='asset' THEN amount::numeric ELSE 0 END"},
		{key: "liabilities", expr: "CASE WHEN accounttype='liability' THEN amount::numeric ELSE 0 END"},
		{key: "equity", expr: "CASE WHEN accounttype='equity' THEN amount::numeric ELSE 0 END"},
		{key: "currentearnings", expr: "CASE WHEN accountcode='__current_earnings__' THEN amount::numeric ELSE 0 END"},
		{key: "difference", expr: "CASE WHEN accounttype='asset' THEN amount::numeric ELSE -amount::numeric END"},
	}
	columns := []ReportColumn{textColumn("accountcode", "รหัสบัญชี"), textColumn("accountname", "ชื่อบัญชี"), textColumn("accounttype", "หมวดบัญชี"), amountColumn("amount", "ยอดคงเหลือ")}
	return r.runWithTotals(ctx, cte, "accounttype,accountcode", columns, totals)
}

func (r reportContext) annualBalances(ctx context.Context) (Report, error) {
	cte := r.base() + `, monthly AS (SELECT account_code,MAX(account_name) AS account_name,date_trunc('month',entry_date)::date AS month,SUM(CASE WHEN kind='opening' THEN debit-credit ELSE 0 END) AS opening,SUM(CASE WHEN kind<>'opening' THEN debit ELSE 0 END) AS debit,SUM(CASE WHEN kind<>'opening' THEN credit ELSE 0 END) AS credit FROM filtered GROUP BY account_code,date_trunc('month',entry_date)), result AS (
      SELECT account_code AS accountcode,account_name AS accountname,to_char(month,'YYYY-MM') AS month,opening::text,debit::text,credit::text,(opening+debit-credit)::text AS movement,(SUM(opening+debit-credit) OVER(PARTITION BY account_code ORDER BY month ROWS UNBOUNDED PRECEDING))::text AS balance FROM monthly)
    `
	return r.run(ctx, cte, "accountcode,month", []ReportColumn{textColumn("accountcode", "รหัสบัญชี"), textColumn("accountname", "ชื่อบัญชี"), textColumn("month", "เดือน"), amountColumn("opening", "ยอดยกมา"), amountColumn("debit", "เดบิต"), amountColumn("credit", "เครดิต"), amountColumn("movement", "เคลื่อนไหวสุทธิ"), amountColumn("balance", "ยอดสะสม")}, []string{"opening", "debit", "credit", "movement"})
}

// allocation reports the cost-allocation setup (gl_allocations) together with
// the amount of the source account that falls into this period, so the operator
// can verify each rule splits exactly 100 percent before posting. The setup
// lives in gl_records (written by the synchronous PostgreSQL store) and the movement
// comes from gl_lines, which is the only place accounting arithmetic may run.
func (r reportContext) allocation(ctx context.Context) (Report, error) {
	columns := []ReportColumn{
		textColumn("alloccode", "รหัสการปันส่วน"),
		textColumn("allocname", "ชื่อ/รายละเอียด"),
		textColumn("accountcode", "บัญชีต้นทุน"),
		textColumn("accountname", "ชื่อบัญชี"),
		textColumn("target", "ปลายทางที่ปันส่วน"),
		textColumn("departmentcode", "แผนก"),
		textColumn("projectcode", "โครงการ"),
		amountColumn("rate", "อัตรา %"),
		amountColumn("sourceamount", "ยอดต้นทุนที่ปันส่วน"),
		amountColumn("allocatedamount", "ยอดปันส่วน"),
	}
	cte := r.base() + `, source AS (
      SELECT account_code, SUM(debit-credit) AS amount FROM filtered GROUP BY account_code
    ), alloc AS (
      SELECT rec.code AS alloccode, rec.payload->>'name' AS allocname, rec.payload->>'accountcode' AS accountcode,
        rule.value->>'branchcode' AS branchcode, rule.value->>'departmentcode' AS departmentcode,
        rule.value->>'projectcode' AS projectcode, rule.value->>'accountcode' AS targetaccount,
        COALESCE((rule.value->>'rate')::numeric, 0) AS rate
      FROM gl_records rec CROSS JOIN LATERAL jsonb_array_elements(COALESCE(rec.payload->'allocaterules','[]'::jsonb)) rule
      WHERE rec.company=$1 AND rec.kind='allocations' AND NOT COALESCE((rec.payload->>'isdeleted')::boolean,false)
        AND COALESCE((rec.payload->>'isactive')::boolean,true)
    ), result AS (
      SELECT a.alloccode, a.allocname, a.accountcode, COALESCE(acc.account_name,'') AS accountname,
        TRIM(BOTH ' /' FROM CONCAT_WS(' / ',
          NULLIF(a.targetaccount,''), NULLIF(a.branchcode,''))) AS target,
        a.departmentcode, a.projectcode,
        a.rate::text AS rate,
        COALESCE(src.amount,0)::text AS sourceamount,
        (a.rate/100 * COALESCE(src.amount,0))::text AS allocatedamount
      FROM alloc a
      LEFT JOIN source src ON src.account_code = a.accountcode
      LEFT JOIN LATERAL (SELECT MAX(account_name) AS account_name FROM filtered WHERE account_code = a.accountcode) acc ON TRUE
    )`
	result, err := r.runWithTotals(ctx, cte, "alloccode,departmentcode,projectcode,target", columns, []reportTotal{
		{key: "rate", expr: "rate::numeric"},
		{key: "sourceamount", expr: "sourceamount::numeric"},
		{key: "allocatedamount", expr: "allocatedamount::numeric"},
	})
	return result, err
}

func (r reportContext) glJournal(ctx context.Context) (Report, error) {
	cte := r.base() + `, movements AS (
		SELECT * FROM filtered WHERE entry_date >= $3::date AND entry_date <= $4::date AND kind <> 'opening'
	), result AS (
		SELECT entry_date::text AS date, doc_no AS docno, book_code AS bookcode,
		       account_code AS accountcode, account_name AS accountname,
		       description, debit::text, credit::text,
		       branch_code AS branchcode, department_code AS departmentcode,
		       project_code AS projectcode, journal_id AS journalid, line_no::text AS __line_no
		FROM movements
	)`
	cols := []ReportColumn{
		textColumn("date", "วันที่"),
		textColumn("docno", "เลขที่เอกสาร"),
		textColumn("bookcode", "สมุดรายวัน"),
		textColumn("accountcode", "รหัสบัญชี"),
		textColumn("accountname", "ชื่อบัญชี"),
		textColumn("description", "คำอธิบาย"),
		amountColumn("debit", "เดบิต"),
		amountColumn("credit", "เครดิต"),
		textColumn("branchcode", "สาขา"),
		textColumn("departmentcode", "แผนก"),
		textColumn("projectcode", "โครงการ"),
	}
	return r.run(ctx, cte, "date,docno,journalid,__line_no::integer", cols, []string{"debit", "credit"}, "journalid", "__line_no")
}

// budgetComparison (Champ 5530 รายงานเปรียบเทียบงบประมาณ, GLRepBudgetCompView) compares each
// budget line (budget × account) with the actual movement of that account on posted ledger
// lines inside the budget's own dimensions ('' = any), excluding opening and closing entries.
// Actual follows the account's normal balance recorded on the line (credit-normal accounts
// count credit - debit), so income and expense budgets both read as positive amounts.
// Budget = the periods of the fiscal year whose start date falls inside the report range
// (budgetPeriodsInRange); variance = budget - actual as in Champ.
func (r reportContext) budgetComparison(ctx context.Context) (Report, error) {
	r.args = append(append([]any{}, r.args...), pq.Array(budgetPeriodsInRange(r.fiscal, r.query.From, r.query.To)), NormalizeCode(r.query.BudgetCode))
	cte := r.base() + `, budget_rows AS (
		SELECT b.code AS budget_code, b.name AS budget_name, b.status AS budget_status,
			b.branch_code, b.department_code, b.project_code, l.account_code,
			SUM(CASE WHEN l.period_no = ANY($10::smallint[]) THEN l.amount ELSE 0 END) AS budget_amount
		FROM gl_budgets b
		JOIN gl_budget_lines l ON l.company=b.company AND l.budget_code=b.code
		WHERE b.company=$1 AND b.fiscal_year=$2 AND ($11='' OR b.code=$11)
			AND ($5='' OR l.account_code=$5) AND ($6='' OR b.branch_code=$6)
			AND ($7='' OR b.department_code=$7) AND ($8='' OR b.project_code=$8)
		GROUP BY b.code, b.name, b.status, b.branch_code, b.department_code, b.project_code, l.account_code
	), compared AS (
		SELECT br.*, COALESCE(acc.payload->>'accounttype', '') AS account_type,
			COALESCE((SELECT n->>'name' FROM jsonb_array_elements(COALESCE(acc.payload->'names','[]'::jsonb)) n WHERE n->>'code'='th' LIMIT 1), '') AS account_name,
			COALESCE(act.actual, 0) AS actual_amount
		FROM budget_rows br
		LEFT JOIN gl_records acc ON acc.company=$1 AND acc.kind='accounts' AND acc.code=br.account_code AND NOT COALESCE((acc.payload->>'isdeleted')::boolean, false)
		LEFT JOIN LATERAL (
			SELECT SUM(CASE WHEN f.normal_balance='credit' THEN f.credit-f.debit ELSE f.debit-f.credit END) AS actual
			FROM filtered f
			WHERE f.account_code=br.account_code AND f.entry_date>=$3::date AND f.kind NOT IN ('opening','closing')
				AND (br.branch_code='' OR f.branch_code=br.branch_code)
				AND (br.department_code='' OR f.department_code=br.department_code)
				AND (br.project_code='' OR f.project_code=br.project_code)
		) act ON true
	), result AS (
		SELECT account_code AS accountcode, account_name AS accountname, account_type AS accounttype,
			budget_code AS budgetcode, budget_name AS budgetname, budget_status AS budgetstatus,
			branch_code AS branchcode, department_code AS departmentcode, project_code AS projectcode,
			budget_amount::text AS budgetamount, actual_amount::text AS actualamount,
			(budget_amount - actual_amount)::text AS variance,
			(CASE WHEN budget_amount > 0 THEN ROUND(actual_amount / budget_amount * 100, 2) ELSE 0 END)::numeric(20,2)::text AS percentused
		FROM compared
	)`
	cols := []ReportColumn{
		textColumn("accountcode", "รหัสบัญชี"),
		textColumn("accountname", "ชื่อบัญชี"),
		textColumn("accounttype", "หมวดบัญชี"),
		textColumn("budgetcode", "รหัสงบประมาณ"),
		textColumn("budgetname", "ชื่องบประมาณ"),
		textColumn("budgetstatus", "สถานะงบประมาณ"),
		textColumn("branchcode", "สาขา"),
		textColumn("departmentcode", "แผนก"),
		textColumn("projectcode", "โครงการ"),
		amountColumn("budgetamount", "งบประมาณ"),
		amountColumn("actualamount", "ใช้จริง"),
		amountColumn("variance", "ผลต่างคงเหลือ"),
		amountColumn("percentused", "ร้อยละที่ใช้ (%)"),
	}
	totals := []reportTotal{
		{key: "budgetamount", expr: "budgetamount::numeric"},
		{key: "actualamount", expr: "actualamount::numeric"},
		{key: "variance", expr: "variance::numeric"},
	}
	// Champ orders the comparison by account, then budget code.
	return r.runWithTotals(ctx, cte, "accountcode,budgetcode", cols, totals)
}

// budgetPeriodsInRange lists the period numbers of the fiscal year whose start date is
// inside [from, to] (a period counts whole when its first day is in the report range).
func budgetPeriodsInRange(y FiscalYear, from, to string) []int64 {
	out := []int64{}
	for i, start := range fiscalPeriodStarts(y) {
		if start >= from && start <= to {
			out = append(out, int64(i+1))
		}
	}
	return out
}
