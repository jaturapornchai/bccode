package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
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
	case "statement":
		report, err = rc.statement(ctx)
	default:
		return empty, fmt.Errorf("รายงานนี้ยังไม่มีรูปแบบที่ยืนยันแล้ว")
	}
	if err != nil {
		return empty, err
	}
	report = rc.finish(report)
	if err = tx.Commit(); err != nil {
		return empty, err
	}
	return report, nil
}

// finish ค่าท้ายที่ทุกรายงานได้เหมือนกัน — ใช้ทั้ง reports/<name> และงบแต่ละงบใน reports/statement-set (statement_set.go)
func (r reportContext) finish(report Report) Report {
	report.AsOf = r.query.To
	if report.Warnings == nil {
		report.Warnings = []string{}
	}
	return report
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
	q.AccountFrom, q.AccountTo = NormalizeCode(q.AccountFrom), NormalizeCode(q.AccountTo)
	// เทียบแบบไบต์ (Go string compare) ให้ตรงกับ account_code COLLATE "C" ใน r.filter
	if q.AccountFrom != "" && q.AccountTo != "" && q.AccountFrom > q.AccountTo {
		return r, fieldError("report_account_range_invalid", "accountfrom", "รหัสบัญชีเริ่มต้นต้องไม่มากกว่ารหัสบัญชีสิ้นสุด")
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
	r.args = []any{scope.Company, q.FiscalYear, q.From, q.To, q.AccountCode, q.BranchCode, q.DepartmentCode, q.ProjectCode, q.BookCode, q.AccountFrom, q.AccountTo}
	// $10/$11 = ช่วงรหัสบัญชีแบบ Champ (รวมปลายทั้งสอง) เทียบแบบไบต์ให้ผลเหมือนกันทุก locale ของฐานข้อมูล
	r.filter = `company=$1 AND fiscal_year=$2 AND $3::date<=$4::date AND entry_date<=$4::date AND ($5='' OR account_code=$5) AND ($6='' OR branch_code=$6) AND ($7='' OR department_code=$7) AND ($8='' OR project_code=$8) AND ($9='' OR book_code=$9)` + accountRangeFilter("account_code")
	return r, nil
}

// accountRangeFilter is the account code range condition on $10 (from) and $11 (to) for any
// query that runs with r.args — every such query must reference both parameters.
func accountRangeFilter(column string) string {
	return ` AND ($10='' OR ` + column + ` COLLATE "C">=$10) AND ($11='' OR ` + column + ` COLLATE "C"<=$11)`
}

func textColumn(key, label string) ReportColumn { return ReportColumn{Key: key, Label: label} }
func amountColumn(key, label string) ReportColumn {
	return ReportColumn{Key: key, Label: label, Amount: true}
}

// debitCreditColumns = คู่ เดบิต | เครดิต ใต้หัวกลุ่มเดียว (หัวตาราง 2 ชั้นของ Champ)
func debitCreditColumns(debitKey, creditKey, group string) []ReportColumn {
	return []ReportColumn{{Key: debitKey, Label: "เดบิต", Amount: true, Group: group}, {Key: creditKey, Label: "เครดิต", Amount: true, Group: group}}
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

// RECURSIVE allows the chart tree CTE of accountBalanceReport; it changes nothing for plain CTEs.
func (r reportContext) base() string {
	return `WITH RECURSIVE filtered AS (SELECT * FROM gl_lines WHERE ` + r.filter + `)`
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

// trialBalance = Champ 5506 รายงานงบทดลอง: ยอดยกมา · ยอดประจำงวด · ยอดสะสม แต่ละกลุ่มหักกลบลงฝั่งเดียว
// (GLRepTrialBalanceView.cpp:654-697); working = Champ 5499 รายงานกระดาษทำการ: งบทดลอง · งบกำไรขาดทุน · งบดุล
// (GLRepWorkingPaperView.cpp:181-195) ป้ายตามตัวอย่างรายงานจริง bcspec/Report_ระบบบัญชี
func (r reportContext) trialBalance(ctx context.Context, working bool) (Report, error) {
	columns := []ReportColumn{textColumn("accountcode", "รหัสผังบัญชี"), textColumn("accountname", "รายการ")}
	cte := r.trialCTE() + `, leaf AS (SELECT account_code,account_name,account_type,GREATEST(balance,0) AS endingdebit,GREATEST(-balance,0) AS endingcredit`
	if working {
		cte += `,(CASE WHEN account_type IN ('income','expense') THEN GREATEST(balance,0) ELSE 0 END) AS pnldebit,(CASE WHEN account_type IN ('income','expense') THEN GREATEST(-balance,0) ELSE 0 END) AS pnlcredit,(CASE WHEN account_type IN ('asset','liability','equity') THEN GREATEST(balance,0) ELSE 0 END) AS bsdebit,(CASE WHEN account_type IN ('asset','liability','equity') THEN GREATEST(-balance,0) ELSE 0 END) AS bscredit`
		columns = append(append(append(columns, debitCreditColumns("endingdebit", "endingcredit", "งบทดลอง")...), debitCreditColumns("pnldebit", "pnlcredit", "งบกำไรขาดทุน")...), debitCreditColumns("bsdebit", "bscredit", "งบดุล")...)
	} else {
		cte += `,GREATEST(opening,0) AS openingdebit,GREATEST(-opening,0) AS openingcredit,GREATEST(debit-credit,0) AS debit,GREATEST(credit-debit,0) AS credit`
		columns = append(append(append(columns, debitCreditColumns("openingdebit", "openingcredit", "ยอดยกมา")...), debitCreditColumns("debit", "credit", "ยอดประจำงวด")...), debitCreditColumns("endingdebit", "endingcredit", "ยอดสะสม")...)
	}
	cte += ` FROM balances)`
	result, err := r.accountBalanceReport(ctx, cte, columns, r.query.HideZero)
	if err != nil {
		return result, err
	}
	total := func(key string) decimal.Decimal { return Amount(result.Totals[key]).Decimal() }
	result.Totals["difference"] = total("endingdebit").Sub(total("endingcredit")).String()
	if working {
		// Champ พิมพ์แถว "กำไร(ขาดทุน)สุทธิ" ท้ายกระดาษทำการ = ส่วนต่างที่ทำให้งบกำไรขาดทุนดุล
		result.Totals["profit"] = total("pnlcredit").Sub(total("pnldebit")).String()
	}
	return result, nil
}

// accountBalanceReport renders the chart of accounts as Champ prints its trial balance family
// (GLRepTrialBalanceView.cpp:445-509, งบทดลอง/กระดาษทำการ/แยกประเภทแบบสรุป): the leaf CTE must expose
// account_code, account_name, account_type and every amount column of columns as numeric.
//   - every posting account of the chart inside the account filters is listed, zero or not, plus any
//     account that has lines; hideZero drops accounts whose amounts are all zero
//     (Champ "ไม่แสดงรายการผังบัญชีที่มียอดเท่ากับศูนย์")
//   - each control account above a listed account gets a header row (rowtype=header, code + name,
//     no amounts) before its descendants
//   - with ControlTotals it also gets a subtotal row (rowtype=control) after its descendants; the
//     frontend prints it name-only in bold like Champ (option "ไม่แสดงผลรวมบัญชีคุม" turns it off)
//
// Header and control rows never count in the report totals.
func (r reportContext) accountBalanceReport(ctx context.Context, leafCTE string, columns []ReportColumn, hideZero bool) (Report, error) {
	amounts := []string{}
	for _, column := range columns {
		if column.Amount {
			amounts = append(amounts, column.Key)
		}
	}
	names := map[string]string{"accountcode": "account_code", "accountname": "account_name", "accounttype": "account_type"}
	selectFrom := func(alias string, blankAmounts bool) string {
		parts := make([]string, len(columns))
		for i, column := range columns {
			switch {
			case column.Amount && blankAmounts:
				parts[i] = "''::text AS " + column.Key
			case column.Amount:
				parts[i] = alias + column.Key + "::text AS " + column.Key
			default:
				parts[i] = alias + names[column.Key] + " AS " + column.Key
			}
		}
		return strings.Join(parts, ",")
	}
	listed := make([]string, len(amounts))
	sums := make([]string, len(amounts))
	zero := make([]string, len(amounts))
	totals := make([]reportTotal, len(amounts))
	for i, key := range amounts {
		listed[i] = "COALESCE(l." + key + ",0) AS " + key
		sums[i] = "SUM(p." + key + ") AS " + key
		zero[i] = key + "=0"
		totals[i] = reportTotal{key: key, expr: "CASE WHEN rowtype='account' THEN " + key + "::numeric ELSE 0 END"}
	}
	kept := `SELECT * FROM listed`
	if hideZero {
		kept += ` WHERE NOT (` + strings.Join(zero, " AND ") + `)`
	}
	subtotals := ""
	if r.query.ControlTotals {
		subtotals = `
		UNION ALL
		SELECT ` + selectFrom("c.", false) + `,'control',(cardinality(t.parents)+1)::text,t.path||chr(2) FROM (SELECT s.*,ch.account_name,ch.account_type FROM controls s JOIN chart ch ON ch.code=s.account_code) c JOIN tree t ON t.code=c.account_code`
	}
	// path = รหัสบัญชีแม่ทุกชั้นคั่นด้วย chr(1): หัวบัญชีคุม (path) มาก่อนลูกทุกตัว (path||chr(1)||…) และแถวรวม (path||chr(2)) ตามหลังเมื่อเทียบแบบไบต์
	cte := leafCTE + `, chart AS (
		SELECT DISTINCT ON (code) code, COALESCE(payload->>'parentaccountcode','') AS parent, COALESCE(payload->>'accounttype','') AS account_type,
			COALESCE((SELECT n->>'name' FROM jsonb_array_elements(COALESCE(payload->'names','[]'::jsonb)) n WHERE n->>'code'='th' LIMIT 1),'') AS account_name,
			COALESCE((payload->>'allowposting')::boolean,true) AS allowposting
		FROM gl_records WHERE company=$1 AND kind='accounts' AND NOT COALESCE((payload->>'isdeleted')::boolean,false) ORDER BY code,id
	), tree AS (
		SELECT c.code, ARRAY[]::text[] AS parents, c.code AS path FROM chart c WHERE NOT EXISTS (SELECT 1 FROM chart p WHERE p.code=c.parent)
		UNION ALL
		SELECT c.code, t.parents||t.code, t.path||chr(1)||c.code FROM tree t JOIN chart c ON c.parent=t.code WHERE cardinality(t.parents)<64 AND c.code<>ALL(t.parents)
	), postable AS (
		SELECT c.* FROM chart c WHERE c.allowposting AND NOT EXISTS (SELECT 1 FROM chart k WHERE k.parent=c.code) AND ($5='' OR c.code=$5)` + accountRangeFilter("c.code") + `
	), listed AS (
		SELECT COALESCE(l.account_code,c.code) AS account_code,COALESCE(NULLIF(l.account_name,''),c.account_name) AS account_name,COALESCE(NULLIF(l.account_type,''),c.account_type) AS account_type,` + strings.Join(listed, ",") + `
		FROM leaf l FULL JOIN postable c ON c.code=l.account_code
	), kept AS (` + kept + `
	), placed AS (
		SELECT k.*, COALESCE(t.parents,ARRAY[]::text[]) AS parents, COALESCE(t.path,k.account_code) AS path FROM kept k LEFT JOIN tree t ON t.code=k.account_code
	), controls AS (
		SELECT a.code AS account_code, ` + strings.Join(sums, ",") + ` FROM placed p CROSS JOIN LATERAL unnest(p.parents) AS a(code) GROUP BY a.code
	), result AS (
		SELECT ` + selectFrom("p.", false) + `,'account' AS rowtype,(cardinality(p.parents)+1)::text AS accountlevel,p.path AS __path FROM placed p
		UNION ALL
		SELECT ` + selectFrom("h.", true) + `,'header',(cardinality(t.parents)+1)::text,t.path FROM (SELECT ch.code AS account_code,ch.account_name,ch.account_type FROM controls s JOIN chart ch ON ch.code=s.account_code) h JOIN tree t ON t.code=h.account_code` + subtotals + `
	)`
	return r.runWithTotals(ctx, cte, `__path COLLATE "C"`, columns, totals, "rowtype", "accountlevel", "__path")
}

func (r reportContext) ledger(ctx context.Context) (Report, error) {
	if r.query.Summary {
		return r.ledgerSummary(ctx)
	}
	// ไม่แสดงยอดยกมา (Champ 5534 m_iNotTransfer): ไม่มีคอลัมน์ยอดยกมา และยอดคงเหลือสะสมเริ่มจากศูนย์ (GLRepGL2View.cpp:392-394)
	opening := "COALESCE(o.opening,0)"
	if r.query.HideOpening {
		opening = "0"
	}
	// Champ 5534 (GLRepGL2View.cpp:187-1099): ต่อบัญชี = แถว "ยอดยกมา" → รายการ → "ยอดยกไป" (ยอดคงเหลือลงฝั่งตรงข้ามให้ดุล)
	// → "ยอดรวม" (เดบิต = เครดิต); rowtype opening/line/carry/total — ข้อความของแถวพิเศษแปลที่ frontend;
	// accountfirst = "1" ที่แถวแรกของบัญชี ให้จอแสดงรหัส/ชื่อบัญชีครั้งเดียวแบบหัวบัญชีของ Champ
	accounts := `accts a FULL JOIN openings o USING(account_code) WHERE a.account_code IS NOT NULL OR o.opening<>0`
	if r.query.HideOpening {
		accounts = `accts a LEFT JOIN openings o USING(account_code)`
	}
	special := func(rowtype string, sort int, debit, credit, balance string) string {
		return `SELECT la.account_code,la.account_name,'','','','','','','',(` + debit + `)::text,(` + credit + `)::text,` + balance + `,'` + rowtype + `','','` + fmt.Sprint(sort) + `',NULL::text FROM ledger_accounts la`
	}
	ending := `(la.opening+la.debit-la.credit)`
	rows := []string{
		`SELECT m.account_code,m.account_name,m.entry_date::text,m.doc_no,m.book_code,m.branch_code,m.department_code,m.project_code,m.description,m.debit::text,m.credit::text,
      (la.opening+SUM(m.debit-m.credit) OVER(PARTITION BY m.account_code ORDER BY m.entry_date,m.doc_no,m.journal_id,m.line_no ROWS UNBOUNDED PRECEDING))::text,'line',m.journal_id,'2',m.line_no::text FROM movements m JOIN ledger_accounts la USING(account_code)`,
		special("carry", 3, `GREATEST(-`+ending+`,0)`, `GREATEST(`+ending+`,0)`, `''`),
		special("total", 4, `GREATEST(la.opening,0)+la.debit+GREATEST(-`+ending+`,0)`, `GREATEST(-la.opening,0)+la.credit+GREATEST(`+ending+`,0)`, `''`),
	}
	if !r.query.HideOpening {
		rows = append(rows, special("opening", 1, `GREATEST(la.opening,0)`, `GREATEST(-la.opening,0)`, `la.opening::text`))
	}
	cte := r.base() + `, openings AS (SELECT account_code,MAX(account_name) AS account_name,SUM(debit-credit) AS opening FROM filtered WHERE entry_date<$3::date OR kind='opening' GROUP BY account_code),
      movements AS (SELECT * FROM filtered WHERE entry_date>=$3::date AND kind<>'opening'),
      accts AS (SELECT account_code,MAX(account_name) AS account_name,SUM(debit) AS debit,SUM(credit) AS credit FROM movements GROUP BY account_code),
      ledger_accounts AS (SELECT account_code,COALESCE(a.account_name,o.account_name) AS account_name,` + opening + ` AS opening,COALESCE(a.debit,0) AS debit,COALESCE(a.credit,0) AS credit FROM ` + accounts + `),
      ledger_rows(accountcode,accountname,date,docno,bookcode,branchcode,departmentcode,projectcode,description,debit,credit,balance,rowtype,journalid,__sort,__line_no) AS (` + strings.Join(rows, "\n      UNION ALL ") + `),
      result AS (SELECT *,CASE WHEN ROW_NUMBER() OVER(PARTITION BY accountcode ORDER BY __sort,date,docno,journalid,__line_no::integer)=1 THEN '1' ELSE '' END AS accountfirst FROM ledger_rows)`
	cols := []ReportColumn{textColumn("accountcode", "ผังบัญชี"), textColumn("accountname", "ชื่อผังบัญชี"), textColumn("date", "วันที่"), textColumn("bookcode", "สมุดรายวัน"), textColumn("docno", "เลขที่เอกสาร"), textColumn("description", "คำอธิบาย"), amountColumn("debit", "เดบิต"), amountColumn("credit", "เครดิต"), amountColumn("balance", "ยอดคงเหลือ"), textColumn("branchcode", "สาขา"), textColumn("departmentcode", "แผนก"), textColumn("projectcode", "โครงการ")}
	lineOnly := func(key string) reportTotal {
		return reportTotal{key: key, expr: "CASE WHEN rowtype='line' THEN " + key + "::numeric ELSE 0 END"}
	}
	result, err := r.runWithTotals(ctx, cte, "accountcode,__sort,date,docno,journalid,__line_no::integer", cols, []reportTotal{lineOnly("debit"), lineOnly("credit")}, "rowtype", "accountfirst", "journalid", "__sort", "__line_no")
	if err != nil {
		return result, err
	}
	if r.query.HideOpening {
		err = r.setTotals(ctx, &result, r.base()+` SELECT COALESCE(SUM(debit-credit) FILTER(WHERE entry_date>=$3::date AND kind<>'opening'),0)::text FROM filtered`, "balance")
		return result, err
	}
	err = r.setTotals(ctx, &result, r.base()+` SELECT COALESCE(SUM(CASE WHEN entry_date<$3::date OR kind='opening' THEN debit-credit ELSE 0 END),0)::text,COALESCE(SUM(debit-credit),0)::text FROM filtered`, "opening", "balance")
	return result, err
}

// ledgerSummary = Champ 5535 รายงานบัญชีแยกประเภทแบบสรุป (GLRepGLSumView): ยอดสะสมสุทธิ ณ วันที่ (to)
// ของแต่ละบัญชีตั้งแต่ต้นปีบัญชีรวมยอดยกมา ลงฝั่งเดบิตหรือเครดิตฝั่งเดียว พร้อมผลรวมบัญชีคุมตามระดับ
func (r reportContext) ledgerSummary(ctx context.Context) (Report, error) {
	cte := r.base() + `, leaf AS (SELECT account_code,MAX(account_name) AS account_name,MAX(account_type) AS account_type,GREATEST(SUM(debit-credit),0) AS debit,GREATEST(-SUM(debit-credit),0) AS credit FROM filtered GROUP BY account_code)`
	columns := []ReportColumn{textColumn("accountcode", "รหัสผังบัญชี"), textColumn("accountname", "รายการ"), amountColumn("debit", "เดบิต"), amountColumn("credit", "เครดิต")}
	return r.accountBalanceReport(ctx, cte, columns, false)
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
		textColumn("alloccode", "รหัสการจัดสรร"),
		textColumn("allocname", "ชื่อ/รายละเอียด"),
		textColumn("accountcode", "บัญชีต้นทุน"),
		textColumn("accountname", "ชื่อบัญชี"),
		textColumn("target", "ปลายทางที่จัดสรร"),
		textColumn("departmentcode", "แผนก"),
		textColumn("projectcode", "โครงการ"),
		amountColumn("rate", "อัตรา %"),
		amountColumn("sourceamount", "ยอดต้นทุนที่จัดสรร"),
		amountColumn("allocatedamount", "ยอดจัดสรร"),
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
	if r.query.Summary {
		return r.glJournalSummary(ctx)
	}
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
		textColumn("bookcode", "สมุดรายวัน"),
		textColumn("docno", "เลขที่เอกสาร"),
		textColumn("accountcode", "ผังบัญชี"),
		textColumn("accountname", "ชื่อบัญชี"),
		textColumn("description", "คำอธิบาย"),
		textColumn("departmentcode", "แผนก"),
		textColumn("projectcode", "โครงการ"),
		amountColumn("debit", "เดบิต"),
		amountColumn("credit", "เครดิต"),
		textColumn("branchcode", "สาขา"),
	}
	return r.run(ctx, cte, "date,docno,journalid,__line_no::integer", cols, []string{"debit", "credit"}, "journalid", "__line_no")
}

// glJournalSummary = Champ 5533 รายงานข้อมูลรายวันแบบสรุป (GLRepTrans2View): แถวละเอกสาร — วันที่ · สมุดรายวัน ·
// เลขที่ · เลขที่อ้างอิง · คำอธิบาย · จำนวนเงิน (ยอดเดบิตของบรรทัดที่ผ่านเงื่อนไข) จากเอกสารที่ลงบัญชีแล้วใน gl_lines
func (r reportContext) glJournalSummary(ctx context.Context) (Report, error) {
	cte := r.base() + `, docs AS (
		SELECT journal_id, MIN(entry_date) AS entry_date, MIN(doc_no) AS doc_no, MIN(book_code) AS book_code, SUM(debit) AS amount
		FROM filtered WHERE entry_date >= $3::date AND kind <> 'opening' GROUP BY journal_id
	), result AS (
		SELECT d.entry_date::text AS date, d.book_code AS bookcode, d.doc_no AS docno,
		       COALESCE(j.payload->>'reference','') AS referenceno, COALESCE(j.payload->>'description','') AS description,
		       d.amount::text AS amount, d.journal_id AS journalid
		FROM docs d LEFT JOIN gl_records j ON j.company=$1 AND j.kind='journals' AND j.id=d.journal_id
	)`
	cols := []ReportColumn{
		textColumn("date", "วันที่เอกสาร"),
		textColumn("bookcode", "สมุดรายวัน"),
		textColumn("docno", "เลขที่เอกสาร"),
		textColumn("referenceno", "เอกสารอ้างอิง"),
		textColumn("description", "คำอธิบาย"),
		amountColumn("amount", "จำนวนเงิน"),
	}
	return r.run(ctx, cte, "date,docno,journalid", cols, []string{"amount"}, "journalid")
}

// budgetComparison (Champ 5530 รายงานเปรียบเทียบงบประมาณ, GLRepBudgetCompView) compares each
// budget line (budget × account) with the actual movement of that account on posted ledger
// lines inside the budget's own dimensions ('' = any), excluding opening and closing entries.
// Actual follows the account's normal balance recorded on the line (credit-normal accounts
// count credit - debit), so income and expense budgets both read as positive amounts.
// Budget = the periods of the fiscal year whose start date falls inside the report range
// (budgetPeriodsInRange); variance = budget - actual as in Champ.
func (r reportContext) budgetComparison(ctx context.Context) (Report, error) {
	// $12 = งวดงบที่เริ่มในช่วงรายงาน, $13 = รหัสงบประมาณ ('' = ทุกงบ) ต่อท้าย r.args ($1-$11)
	r.args = append(append([]any{}, r.args...), pq.Array(budgetPeriodsInRange(r.fiscal, r.query.From, r.query.To)), NormalizeCode(r.query.BudgetCode))
	cte := r.base() + `, budget_rows AS (
		SELECT b.code AS budget_code, b.name AS budget_name, b.status AS budget_status,
			b.branch_code, b.department_code, b.project_code, l.account_code,
			SUM(CASE WHEN l.period_no = ANY($12::smallint[]) THEN l.amount ELSE 0 END) AS budget_amount
		FROM gl_budgets b
		JOIN gl_budget_lines l ON l.company=b.company AND l.budget_code=b.code
		WHERE b.company=$1 AND b.fiscal_year=$2 AND ($13='' OR b.code=$13)
			AND ($5='' OR l.account_code=$5)` + accountRangeFilter("l.account_code") + ` AND ($6='' OR b.branch_code=$6)
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
