package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
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

// The rows CTE must expose only text columns and, where needed, private sort
// columns. PostgreSQL performs all accounting arithmetic before ::text scanning.
func (r reportContext) run(ctx context.Context, cte, order string, columns []ReportColumn, totalKeys []string) (Report, error) {
	result := Report{Columns: columns, Rows: []map[string]string{}, Totals: map[string]string{}, Warnings: []string{}}
	aggregate := []string{"count(*)"}
	for _, key := range totalKeys {
		aggregate = append(aggregate, `COALESCE(SUM(`+key+`::numeric),0)::text`)
	}
	values := make([]string, len(totalKeys))
	targets := []any{&result.TotalRows}
	for i := range values {
		targets = append(targets, &values[i])
	}
	if err := r.tx.QueryRowContext(ctx, cte+` SELECT `+strings.Join(aggregate, ",")+` FROM result`, r.args...).Scan(targets...); err != nil {
		return result, err
	}
	for i, key := range totalKeys {
		result.Totals[key] = values[i]
	}
	selected := []string{}
	for _, column := range columns {
		selected = append(selected, column.Key)
	}
	args := append(append([]any{}, r.args...), r.query.Limit, (r.query.Page-1)*r.query.Limit)
	rows, err := r.tx.QueryContext(ctx, cte+` SELECT `+strings.Join(selected, ",")+` FROM result ORDER BY `+order+` LIMIT $10 OFFSET $11`, args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		values := make([]string, len(columns))
		targets := make([]any, len(columns))
		for i := range values {
			targets[i] = &values[i]
		}
		if err = rows.Scan(targets...); err != nil {
			return result, err
		}
		record := map[string]string{}
		for i, column := range columns {
			record[column.Key] = values[i]
		}
		result.Rows = append(result.Rows, record)
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
      m.debit::text,m.credit::text,COALESCE(o.opening,0)::text AS opening,(COALESCE(o.opening,0)+SUM(m.debit-m.credit) OVER(PARTITION BY m.account_code ORDER BY m.entry_date,m.doc_no,m.journal_id,m.line_no ROWS UNBOUNDED PRECEDING))::text AS balance,m.journal_id,m.line_no FROM movements m LEFT JOIN openings o USING(account_code))`
	cols := []ReportColumn{textColumn("accountcode", "รหัสบัญชี"), textColumn("accountname", "ชื่อบัญชี"), textColumn("date", "วันที่"), textColumn("docno", "เลขที่เอกสาร"), textColumn("bookcode", "สมุดรายวัน"), textColumn("branchcode", "สาขา"), textColumn("departmentcode", "แผนก"), textColumn("projectcode", "โครงการ"), textColumn("description", "รายละเอียด"), amountColumn("opening", "ยอดยกมา"), amountColumn("debit", "เดบิต"), amountColumn("credit", "เครดิต"), amountColumn("balance", "ยอดคงเหลือ")}
	result, err := r.run(ctx, cte, "accountcode,date,docno,journal_id,line_no", cols, []string{"debit", "credit"})
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
	result, err := r.run(ctx, cte, "accounttype,accountcode", []ReportColumn{textColumn("accountcode", "รหัสบัญชี"), textColumn("accountname", "ชื่อบัญชี"), textColumn("accounttype", "หมวดบัญชี"), amountColumn("debit", "เดบิต"), amountColumn("credit", "เครดิต"), amountColumn("balance", "สุทธิเดบิตลบเครดิต"), amountColumn("amount", "จำนวนเงิน")}, []string{"debit", "credit"})
	if err != nil {
		return result, err
	}
	err = r.setTotals(ctx, &result, r.pnlCTE()+` SELECT COALESCE(SUM(-balance) FILTER(WHERE account_type='income'),0)::text,COALESCE(SUM(balance) FILTER(WHERE account_type='expense'),0)::text,COALESCE(SUM(-balance),0)::text FROM pnl`, "revenue", "expense", "profit")
	return result, err
}

func (r reportContext) balanceSheet(ctx context.Context) (Report, error) {
	cte := r.trialCTE() + `, statements AS (
      SELECT account_code AS accountcode,account_name AS accountname,account_type AS accounttype,CASE WHEN account_type='asset' THEN balance ELSE -balance END AS amount FROM balances WHERE account_type IN ('asset','liability','equity')
      UNION ALL SELECT '__current_earnings__','กำไรขาดทุนที่ยังไม่ปิดเข้ากำไรสะสม','equity',COALESCE(SUM(-balance),0) FROM balances WHERE account_type IN ('income','expense')
    ), result AS (SELECT accountcode,accountname,accounttype,amount::text FROM statements)`
	result, err := r.run(ctx, cte, "accounttype,accountcode", []ReportColumn{textColumn("accountcode", "รหัสบัญชี"), textColumn("accountname", "ชื่อบัญชี"), textColumn("accounttype", "หมวดบัญชี"), amountColumn("amount", "ยอดคงเหลือ")}, nil)
	if err != nil {
		return result, err
	}
	err = r.setTotals(ctx, &result, cte+` SELECT COALESCE(SUM(amount::numeric) FILTER(WHERE accounttype='asset'),0)::text,COALESCE(SUM(amount::numeric) FILTER(WHERE accounttype='liability'),0)::text,COALESCE(SUM(amount::numeric) FILTER(WHERE accounttype='equity'),0)::text,COALESCE(SUM(amount::numeric) FILTER(WHERE accountcode='__current_earnings__'),0)::text,COALESCE(SUM(CASE WHEN accounttype='asset' THEN amount::numeric ELSE -amount::numeric END),0)::text FROM result`, "assets", "liabilities", "equity", "currentearnings", "difference")
	return result, err
}

func (r reportContext) annualBalances(ctx context.Context) (Report, error) {
	cte := r.base() + `, monthly AS (SELECT account_code,MAX(account_name) AS account_name,date_trunc('month',entry_date)::date AS month,SUM(CASE WHEN kind='opening' THEN debit-credit ELSE 0 END) AS opening,SUM(CASE WHEN kind<>'opening' THEN debit ELSE 0 END) AS debit,SUM(CASE WHEN kind<>'opening' THEN credit ELSE 0 END) AS credit FROM filtered GROUP BY account_code,date_trunc('month',entry_date)), result AS (
      SELECT account_code AS accountcode,account_name AS accountname,to_char(month,'YYYY-MM') AS month,opening::text,debit::text,credit::text,(opening+debit-credit)::text AS movement,(SUM(opening+debit-credit) OVER(PARTITION BY account_code ORDER BY month ROWS UNBOUNDED PRECEDING))::text AS balance FROM monthly)
    `
	return r.run(ctx, cte, "accountcode,month", []ReportColumn{textColumn("accountcode", "รหัสบัญชี"), textColumn("accountname", "ชื่อบัญชี"), textColumn("month", "เดือน"), amountColumn("opening", "ยอดยกมา"), amountColumn("debit", "เดบิต"), amountColumn("credit", "เครดิต"), amountColumn("movement", "เคลื่อนไหวสุทธิ"), amountColumn("balance", "ยอดสะสม")}, []string{"opening", "debit", "credit", "movement"})
}
