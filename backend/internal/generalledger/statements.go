package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// งบการเงินตามรูปแบบที่ผู้ใช้ออกแบบ (Champ 5505 "รูปแบบงบการเงิน") คำนวณที่ backend ทั้งหมด:
// ยอดบัญชีมาจาก gl_lines (ยอดสะสมไม่เก็บ — mydocs/specs/gl/spec.md), แถวบัญชีรวมยอดตามรหัสบัญชีที่ผู้ใช้เลือกเอง
// (ห้ามเดาบรรทัดงบจากรหัส/ชื่อบัญชี), แถวสูตรคำนวณด้วย decimal ไม่มี float. คอลัมน์ปีก่อนเปรียบเทียบ = ปีบัญชีก่อนหน้า
// ช่วงเดียวกัน (ข้อมูลเปรียบเทียบตามรูปแบบงบการเงินของกรมพัฒนาธุรกิจการค้า; Champ ทำได้ด้วย YEAR() ทีละช่อง).

const statementCurrentEarnings = "__current_earnings__"

type statementBalance struct {
	accountType string
	opening     decimal.Decimal // ยอดต้นงวด = ยอดยกมา + รายการก่อนวันเริ่มงวด
	balance     decimal.Decimal // ยอดคงเหลือเดบิตลบเครดิต ณ วันสิ้นงวด (รวมยอดยกมาและรายการปิดบัญชี)
	movement    decimal.Decimal // ความเคลื่อนไหวในงวด ไม่รวมยอดยกมาและรายการปิดบัญชี
}

// ReportPeriod บอกช่วงของแต่ละคอลัมน์ยอดเงิน ให้หน้าจอพิมพ์หัวงบ ("ณ วันที่" / "สำหรับงวด") จากวันที่จริง
type ReportPeriod struct {
	Key        string `json:"key"`
	FiscalYear string `json:"fiscalyear"`
	From       string `json:"from"`
	To         string `json:"to"`
}

func (r reportContext) statement(ctx context.Context) (Report, error) {
	code := strings.TrimSpace(r.query.Template)
	if code == "" {
		return Report{}, fmt.Errorf("กรุณาเลือกรูปแบบงบการเงิน")
	}
	var payload []byte
	err := r.tx.QueryRowContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='statement-templates' AND code=$2 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`, r.scope.Company, code).Scan(&payload)
	if err == sql.ErrNoRows {
		return Report{}, fmt.Errorf("ไม่พบรูปแบบงบการเงินนี้")
	}
	if err != nil {
		return Report{}, err
	}
	var template Master
	if err = json.Unmarshal(payload, &template); err != nil {
		return Report{}, err
	}
	periodic := statementPeriodic(template.StatementType)
	// คำนวณเต็มความละเอียด (สูตรคูณ/หารตัดที่ 8 ตำแหน่ง) แล้วแสดงผลปัดครึ่งขึ้นห่างศูนย์ที่ทศนิยมของรูปแบบงบ ทีละแถว
	scale := statementScale(template)
	if template.StatementType == "equity" {
		return r.equityStatement(ctx, template, scale)
	}
	current, err := r.statementBalances(ctx, r.fiscal.Code, r.query.From, r.query.To)
	if err != nil {
		return Report{}, err
	}
	report := Report{
		Columns: []ReportColumn{textColumn("rowno", "ลำดับ"), textColumn("title", "รายการ"), textColumn("noteno", "หมายเหตุ"), amountColumn("amount", r.fiscal.Code)},
		Rows:    []map[string]string{}, Totals: map[string]string{}, Warnings: []string{},
		Periods: []ReportPeriod{{Key: "amount", FiscalYear: r.fiscal.Code, From: r.query.From, To: r.query.To}},
	}
	values := evaluateStatement(template.Rows, current, periodic)
	// งบกระแสเงินสด: ตรวจเงินสดปลายงวดกับยอดคงเหลือตามบัญชีจากค่าเต็มความละเอียด ก่อนซ่อนแถวศูนย์ (statements_cashcheck.go)
	cashEndings, cashWarnings := statementCashEndings(template)
	report.Warnings = append(report.Warnings, cashWarnings...)
	report.Checks, cashWarnings = statementCashChecks(cashEndings, current, values, report.Periods[0], scale)
	report.Warnings = append(report.Warnings, cashWarnings...)
	var prior map[int]decimal.Decimal
	if template.GlobalStyle != nil && template.GlobalStyle.ComparisonType == "previous_year" {
		year, from, to, found, err := r.priorPeriod(ctx)
		if err != nil {
			return Report{}, err
		}
		if found {
			balances, err := r.statementBalances(ctx, year, from, to)
			if err != nil {
				return Report{}, err
			}
			prior = evaluateStatement(template.Rows, balances, periodic)
			report.Columns = append(report.Columns, amountColumn("prioramount", year))
			report.Periods = append(report.Periods, ReportPeriod{Key: "prioramount", FiscalYear: year, From: from, To: to})
			priorChecks, priorWarnings := statementCashChecks(cashEndings, balances, prior, report.Periods[1], scale)
			report.Checks = append(report.Checks, priorChecks...)
			report.Warnings = append(report.Warnings, priorWarnings...)
		} else {
			report.Warnings = append(report.Warnings, "ไม่พบปีบัญชีก่อนหน้า จึงไม่มีคอลัมน์เปรียบเทียบ")
		}
	}
	for _, row := range template.Rows {
		out := statementRowMap(row, row.Title)
		out["amount"], out["prioramount"] = "", ""
		if statementAmountRow(row.RowType) {
			out["amount"] = values[row.RowNo].StringFixed(scale)
			if prior != nil {
				out["prioramount"] = prior[row.RowNo].StringFixed(scale)
			}
		}
		report.Rows = append(report.Rows, out)
	}
	if statementHideZero(template) {
		report.Rows = hideZeroStatementRows(report.Rows, statementAmountKeys(report.Columns), statementFormulaRefs(template.Rows))
	}
	// หมายเหตุที่บรรทัดที่พิมพ์จริงอ้างถึงต้องมีในหมายเหตุประกอบงบการเงินของปีที่ออกงบ (statement_notes.go) — ตรวจหลังซ่อนแถวศูนย์
	noteWarnings, err := r.statementNoteWarnings(ctx, statementPrintedNoteNos(template, report.Rows))
	if err != nil {
		return Report{}, err
	}
	report.Warnings = append(noteWarnings, report.Warnings...)
	report.TotalRows = int64(len(report.Rows))
	return report, nil
}

// statementScale ทศนิยมที่แสดงของรูปแบบงบ (ไม่ตั้ง = 2) — ชุดงบ (statement_set.go) ส่งค่าเดียวกันให้จอพิมพ์
func statementScale(template Master) int32 {
	if template.GlobalStyle != nil && template.GlobalStyle.Scale > 0 {
		return int32(template.GlobalStyle.Scale)
	}
	return 2
}

func statementAmountRow(rowType string) bool {
	return rowType == "account" || rowType == "formula" || rowType == "subtotal"
}

// งบช่วงเวลาใช้ความเคลื่อนไหวในงวดเป็นค่าเริ่ม (ยกเว้นแถวที่เลือก amountbasis เอง); งบฐานะการเงินและแบบกำหนดเองใช้ยอดคงเหลือ
func statementPeriodic(statementType string) bool {
	switch statementType {
	case "pnl", "production_cost", "cash_flow", "equity":
		return true
	}
	return false
}

func statementRowMap(row StatementRow, title string) map[string]string {
	return map[string]string{
		"rowno": strconv.Itoa(row.RowNo), "title": title, "noteno": row.NoteNo, "rowtype": row.RowType,
		"indent": strconv.Itoa(row.Style.Indent), "fontweight": row.Style.FontWeight, "fontstyle": row.Style.FontStyle, "underline": row.Style.Underline,
		"showzero": strconv.FormatBool(row.ShowZero),
	}
}

// ยอดต่อบัญชีของปีบัญชีหนึ่ง กรองสาขา/แผนก/โครงการเหมือนรายงานอื่น; รหัสบัญชีและสมุดรายวันไม่ใช้กรองงบการเงิน
func (r reportContext) statementBalances(ctx context.Context, fiscalYear, from, to string) (map[string]statementBalance, error) {
	rows, err := r.tx.QueryContext(ctx, `SELECT account_code,MAX(account_type),COALESCE(SUM(debit-credit),0)::text,
      COALESCE(SUM(CASE WHEN entry_date>=$3::date AND kind NOT IN ('opening','closing') THEN debit-credit ELSE 0 END),0)::text,
      COALESCE(SUM(CASE WHEN entry_date<$3::date OR kind='opening' THEN debit-credit ELSE 0 END),0)::text
    FROM gl_lines WHERE company=$1 AND fiscal_year=$2 AND entry_date<=$4::date AND ($5='' OR branch_code=$5) AND ($6='' OR department_code=$6) AND ($7='' OR project_code=$7)
    GROUP BY account_code`, r.scope.Company, fiscalYear, from, to, r.query.BranchCode, r.query.DepartmentCode, r.query.ProjectCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	balances := map[string]statementBalance{}
	earnings := statementBalance{accountType: "equity"}
	for rows.Next() {
		var code, accountType, balance, movement, opening string
		if err = rows.Scan(&code, &accountType, &balance, &movement, &opening); err != nil {
			return nil, err
		}
		item := statementBalance{accountType: accountType, balance: decimal.RequireFromString(balance), movement: decimal.RequireFromString(movement), opening: decimal.RequireFromString(opening)}
		balances[code] = item
		// กำไรขาดทุนที่ยังไม่ปิดเข้ากำไรสะสม: แบบเดียวกับแถว __current_earnings__ ของรายงานงบแสดงฐานะการเงิน
		if accountType == "income" || accountType == "expense" {
			earnings.balance = earnings.balance.Add(item.balance)
			earnings.movement = earnings.movement.Add(item.movement)
			earnings.opening = earnings.opening.Add(item.opening)
		}
	}
	balances[statementCurrentEarnings] = earnings
	return balances, rows.Err()
}

// ปีบัญชีก่อนหน้า = ปีที่สิ้นสุดล่าสุดก่อนวันเริ่มปีปัจจุบัน; ช่วงเทียบ = ทั้งปีถ้าเลือกทั้งปี ไม่งั้นเลื่อนวันที่ย้อนไป 1 ปี
func (r reportContext) priorPeriod(ctx context.Context) (string, string, string, bool, error) {
	var payload []byte
	err := r.tx.QueryRowContext(ctx, `SELECT payload FROM gl_records WHERE company=$1 AND kind='fiscal-years' AND NOT COALESCE((payload->>'isdeleted')::boolean,false) AND (payload->>'enddate')::date<$2::date ORDER BY (payload->>'enddate')::date DESC, code DESC LIMIT 1`, r.scope.Company, r.fiscal.StartDate).Scan(&payload)
	if err == sql.ErrNoRows {
		return "", "", "", false, nil
	}
	if err != nil {
		return "", "", "", false, err
	}
	var prior FiscalYear
	if err = json.Unmarshal(payload, &prior); err != nil {
		return "", "", "", false, err
	}
	from, to, ok := priorRange(r.fiscal, prior, r.query.From, r.query.To)
	return prior.Code, from, to, ok, nil
}

func priorRange(current, prior FiscalYear, from, to string) (string, string, bool) {
	priorFrom, priorTo := prior.StartDate, prior.EndDate
	if from != current.StartDate {
		priorFrom = maxDate(oneYearEarlier(from), prior.StartDate)
	}
	if to != current.EndDate {
		priorTo = minDate(oneYearEarlier(to), prior.EndDate)
	}
	return priorFrom, priorTo, priorFrom <= priorTo
}

// วันเดียวกันของปีก่อน; วันสิ้นเดือนคงเป็นวันสิ้นเดือน (28 ก.พ. ↔ 29 ก.พ.)
func oneYearEarlier(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	monthEnd := t.AddDate(0, 0, 1).Day() == 1
	first := time.Date(t.Year()-1, t.Month(), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	if monthEnd || t.Day() > last.Day() {
		return last.Format("2006-01-02")
	}
	return first.AddDate(0, 0, t.Day()-1).Format("2006-01-02")
}

func maxDate(a, b string) string {
	if a > b {
		return a
	}
	return b
}
func minDate(a, b string) string {
	if a < b {
		return a
	}
	return b
}

// แถวบัญชีคำนวณก่อน แล้วแถวสูตร/รวมยอดตามลำดับในรูปแบบ (แถวสูตรที่อ้างแถวสูตรข้างล่างได้ 0 เหมือนเดิม)
func evaluateStatement(rows []StatementRow, balances map[string]statementBalance, periodic bool) map[int]decimal.Decimal {
	values := map[int]decimal.Decimal{}
	for _, row := range rows {
		if row.RowType != "account" {
			continue
		}
		values[row.RowNo] = statementAccountRowValue(row, balances, row.AmountBasis, periodic)
	}
	evaluateStatementFormulas(rows, values)
	return values
}

// ยอดของแถวบัญชีตามฐานที่ระบุ ด้านปกติและกลับเครื่องหมายตามแถว (ใช้ทั้งคำนวณงบและตรวจยอดเงินสดปลายงวดกับบัญชี)
func statementAccountRowValue(row StatementRow, balances map[string]statementBalance, basis string, periodic bool) decimal.Decimal {
	sum := decimal.Zero
	for _, code := range row.AccountCodes {
		if account, ok := balances[code]; ok {
			sum = sum.Add(statementSigned(account, statementBasisValue(account, basis, periodic), row.NormalBalance))
		}
	}
	if row.ReverseSign {
		sum = sum.Neg()
	}
	return sum
}

// ยอดของบัญชีตามฐานที่แถวเลือก: ต้นงวด / ปลายงวด / ความเคลื่อนไหว; ไม่ระบุ = ตามชนิดงบ
func statementBasisValue(account statementBalance, basis string, periodic bool) decimal.Decimal {
	switch basis {
	case "opening":
		return account.opening
	case "closing":
		return account.balance
	case "movement":
		return account.movement
	}
	if periodic {
		return account.movement
	}
	return account.balance
}

// ด้านปกติของแถว: ไม่ระบุ = ตามประเภทบัญชี (สินทรัพย์/ค่าใช้จ่ายเดบิต นอกนั้นเครดิต); ด้านเครดิตแสดงเครดิตเป็นบวก
func statementSigned(account statementBalance, value decimal.Decimal, normal string) decimal.Decimal {
	if normal == "" {
		normal = "credit"
		if account.accountType == "asset" || account.accountType == "expense" {
			normal = "debit"
		}
	}
	if normal == "credit" {
		return value.Neg()
	}
	return value
}

func evaluateStatementFormulas(rows []StatementRow, values map[int]decimal.Decimal) {
	for _, row := range rows {
		if row.RowType != "formula" && row.RowType != "subtotal" {
			continue
		}
		sum := decimal.Zero
		if strings.TrimSpace(row.Formula) != "" {
			sum = evaluateStatementFormula(row.Formula, values, row.RowNo)
		}
		if row.ReverseSign {
			sum = sum.Neg()
		}
		values[row.RowNo] = sum
	}
}

// statementFormulaRefs แถวที่สูตรของแต่ละแถวสูตร/รวมยอดอ่านจริง (เรียงตามเลขบรรทัด) ตามลำดับเดียวกับ evaluateStatementFormulas:
// แถวบัญชีทุกแถว + แถวสูตรที่คำนวณก่อนหน้า — ใช้ทั้งการซ่อนแถวศูนย์และการหาบรรทัดเงินสดปลายงวด ให้ตรงกับการคำนวณ
func statementFormulaRefs(rows []StatementRow) map[int][]int {
	values := map[int]decimal.Decimal{}
	for _, row := range rows {
		if row.RowType == "account" {
			values[row.RowNo] = decimal.Zero
		}
	}
	refs := map[int][]int{}
	for _, row := range rows {
		if row.RowType != "formula" && row.RowType != "subtotal" {
			continue
		}
		if strings.TrimSpace(row.Formula) != "" {
			used := map[int]bool{}
			statementFormula(row.Formula, values, row.RowNo, used)
			rowNos := make([]int, 0, len(used))
			for rowNo := range used {
				rowNos = append(rowNos, rowNo)
			}
			sort.Ints(rowNos)
			refs[row.RowNo] = rowNos
		}
		values[row.RowNo] = decimal.Zero
	}
	return refs
}

var (
	statementRangePattern = regexp.MustCompile(`(?i)^SUM\s*\(\s*R?(\d+)\s*(?::|\.\.)\s*R?(\d+)\s*\)$`)
	statementTokenPattern = regexp.MustCompile(`[A-Za-z]+[0-9]*|[0-9]+(?:\.[0-9]+)?|\+|-|\*|/|\(|\)`)
	statementRowPattern   = regexp.MustCompile(`^[Rr]\d+$`)
	statementNumber       = regexp.MustCompile(`^\d+(\.\d+)?$`)
)

// สูตร: SUM(R10:R50) ทั้งช่วง หรือนิพจน์ + - * / วงเล็บ อ้างแถว Rnn และตัวเลข; คูณ/หารตัดเหลือ 8 ตำแหน่ง, หารศูนย์ได้ 0
func evaluateStatementFormula(formula string, values map[int]decimal.Decimal, current int) decimal.Decimal {
	return statementFormula(formula, values, current, nil)
}

// statementFormula คำนวณสูตร และถ้าส่ง refs มาด้วยจะบันทึกแถวที่สูตรใช้จริง (แถวในช่วง SUM / แถว Rnn ที่ถูกอ่าน และมีค่าใน values)
// — การตรวจยอดงบกระแสเงินสดหาแถวที่สูตรอ้างด้วยกติกาเดียวกับการคำนวณ
func statementFormula(formula string, values map[int]decimal.Decimal, current int, refs map[int]bool) decimal.Decimal {
	clean := strings.TrimSpace(formula)
	if match := statementRangePattern.FindStringSubmatch(clean); match != nil {
		from, _ := strconv.Atoi(match[1])
		to, _ := strconv.Atoi(match[2])
		if from > to {
			from, to = to, from
		}
		sum := decimal.Zero
		for rowNo, value := range values {
			if rowNo >= from && rowNo <= to && rowNo != current {
				sum = sum.Add(value)
				if refs != nil {
					refs[rowNo] = true
				}
			}
		}
		return sum
	}
	p := formulaParser{tokens: statementTokenPattern.FindAllString(clean, -1), values: values, current: current, refs: refs}
	return p.expr()
}

type formulaParser struct {
	tokens  []string
	pos     int
	values  map[int]decimal.Decimal
	current int
	refs    map[int]bool
}

func (p *formulaParser) peek() string {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return ""
}

func (p *formulaParser) expr() decimal.Decimal {
	result := p.term()
	for op := p.peek(); op == "+" || op == "-"; op = p.peek() {
		p.pos++
		right := p.term()
		if op == "+" {
			result = result.Add(right)
		} else {
			result = result.Sub(right)
		}
	}
	return result
}

func (p *formulaParser) term() decimal.Decimal {
	result := p.primary()
	for op := p.peek(); op == "*" || op == "/"; op = p.peek() {
		p.pos++
		right := p.primary()
		if op == "*" {
			result = result.Mul(right).Truncate(8)
		} else if right.IsZero() {
			result = decimal.Zero
		} else {
			result, _ = result.QuoRem(right, 8)
		}
	}
	return result
}

func (p *formulaParser) primary() decimal.Decimal {
	token := p.peek()
	if token == "" {
		return decimal.Zero
	}
	p.pos++
	switch {
	case token == "+":
		return p.primary()
	case token == "-":
		return p.primary().Neg()
	case token == "(":
		value := p.expr()
		if p.peek() == ")" {
			p.pos++
		}
		return value
	case statementRowPattern.MatchString(token):
		rowNo, _ := strconv.Atoi(token[1:])
		if rowNo == p.current {
			return decimal.Zero
		}
		value, ok := p.values[rowNo]
		if ok && p.refs != nil {
			p.refs[rowNo] = true
		}
		return value
	case statementNumber.MatchString(token):
		value, err := decimal.NewFromString(token)
		if err != nil {
			return decimal.Zero
		}
		return value
	}
	return decimal.Zero
}
