package generalledger

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น (แบบ 2 ประกาศกรมพัฒนาธุรกิจการค้า พ.ศ. 2566): คอลัมน์ = องค์ประกอบส่วนของผู้ถือหุ้น
// ที่ผู้ใช้เลือกบัญชีเอง (ห้ามเดาจากรหัส/ชื่อบัญชี), แถวของรูปแบบซ้ำเป็นชุดต่อปี (ปีก่อนแล้วปีปัจจุบัน) โดย {year} ในชื่อแถว
// แทนด้วยรหัสปีบัญชี. แถวบัญชีคำนวณแยกทีละคอลัมน์: opening/closing = ยอดต้น/ปลายงวดของบัญชีในคอลัมน์,
// ความเคลื่อนไหว = บัญชีของแถวที่อยู่ในคอลัมน์นั้น, other = ส่วนที่เหลือของคอลัมน์ที่ยังไม่อยู่ในบรรทัดใด
// → ยอดต้นงวด + ทุกบรรทัดความเคลื่อนไหว + other = ยอดปลายงวด เสมอ

type equityBlock struct{ key, year, from, to string }

func (r reportContext) equityStatement(ctx context.Context, template Master, scale int32) (Report, error) {
	columns := make([]StatementColumn, 0, len(template.Columns))
	for _, column := range template.Columns {
		// ข้อ 7 ของประกาศ: องค์ประกอบที่กิจการไม่มีไม่ต้องแสดง — คอลัมน์ที่ยังไม่เลือกบัญชีจึงไม่แสดง
		if len(column.AccountCodes) > 0 {
			columns = append(columns, column)
		}
	}
	if len(columns) == 0 {
		return Report{}, fmt.Errorf("กรุณากำหนดคอลัมน์องค์ประกอบส่วนของผู้ถือหุ้นและเลือกบัญชีอย่างน้อย 1 คอลัมน์")
	}
	templateWarnings := equityTemplateWarnings(template.Rows, columns)
	report := Report{
		Columns: []ReportColumn{textColumn("rowno", "ลำดับ"), textColumn("title", "รายการ"), textColumn("noteno", "หมายเหตุ")},
		Rows:    []map[string]string{}, Totals: map[string]string{}, Warnings: []string{},
	}
	for i, column := range columns {
		report.Columns = append(report.Columns, amountColumn(equityColumnKey(i), column.Title))
	}
	report.Columns = append(report.Columns, amountColumn("total", "รวมส่วนของผู้ถือหุ้น"))
	current := equityBlock{"amount", r.fiscal.Code, r.query.From, r.query.To}
	report.Periods = []ReportPeriod{{Key: current.key, FiscalYear: current.year, From: current.from, To: current.to}}
	blocks := []equityBlock{current}
	if template.GlobalStyle != nil && template.GlobalStyle.ComparisonType == "previous_year" {
		year, from, to, found, err := r.priorPeriod(ctx)
		if err != nil {
			return Report{}, err
		}
		if found {
			prior := equityBlock{"prioramount", year, from, to}
			blocks = []equityBlock{prior, current} // แบบ 2 เรียงปีก่อนก่อน
			report.Periods = append(report.Periods, ReportPeriod{Key: prior.key, FiscalYear: prior.year, From: prior.from, To: prior.to})
		} else {
			report.Warnings = append(report.Warnings, "ไม่พบปีบัญชีก่อนหน้า จึงไม่มีคอลัมน์เปรียบเทียบ")
		}
	}
	for _, block := range blocks {
		balances, err := r.statementBalances(ctx, block.year, block.from, block.to)
		if err != nil {
			return Report{}, err
		}
		values, warnings := evaluateEquityStatement(template.Rows, columns, balances)
		for _, warning := range warnings {
			report.Warnings = append(report.Warnings, "ปี "+block.year+": "+warning)
		}
		for _, row := range template.Rows {
			out := statementRowMap(row, strings.ReplaceAll(row.Title, "{year}", block.year))
			out["block"], out["total"] = block.key, ""
			total := decimal.Zero
			for i := range columns {
				out[equityColumnKey(i)] = ""
				if statementAmountRow(row.RowType) {
					out[equityColumnKey(i)] = values[i][row.RowNo].StringFixed(scale)
					total = total.Add(values[i][row.RowNo])
				}
			}
			if statementAmountRow(row.RowType) {
				out["total"] = total.StringFixed(scale)
			}
			report.Rows = append(report.Rows, out)
		}
	}
	if statementHideZero(template) {
		report.Rows = hideZeroStatementRows(report.Rows, statementAmountKeys(report.Columns), statementFormulaRefs(template.Rows))
		report.Columns = hideZeroEquityColumns(report.Columns, report.Rows)
	}
	// หมายเหตุตรวจเฉพาะบรรทัดที่พิมพ์จริง (หลังซ่อนแถวศูนย์); ลำดับคำเตือน: การตั้งค่ารูปแบบ → หมายเหตุ → รายปี
	noteWarnings, err := r.statementNoteWarnings(ctx, statementPrintedNoteNos(template, report.Rows))
	if err != nil {
		return Report{}, err
	}
	report.Warnings = append(append(templateWarnings, noteWarnings...), report.Warnings...)
	report.TotalRows = int64(len(report.Rows))
	return report, nil
}

func equityColumnKey(index int) string { return "c" + strconv.Itoa(index+1) }

func statementAccountLabel(code string) string {
	if code == statementCurrentEarnings {
		return "กำไร (ขาดทุน) ที่ยังไม่ปิดบัญชี"
	}
	return code
}

func equityRowBasis(row StatementRow) string {
	switch row.AmountBasis {
	case "opening", "closing", "other":
		return row.AmountBasis
	}
	return "movement"
}

// ตรวจการตั้งค่าที่ทำให้ยอดนับซ้ำหรือตกหล่น (เหมือนกันทุกปี จึงตรวจครั้งเดียว)
func equityTemplateWarnings(rows []StatementRow, columns []StatementColumn) []string {
	warnings := []string{}
	columnOf := map[string]string{}
	for _, column := range columns {
		for _, code := range column.AccountCodes {
			if previous, ok := columnOf[code]; ok && previous != column.Title {
				warnings = append(warnings, fmt.Sprintf("บัญชี %s อยู่ในหลายคอลัมน์ (%s, %s) ยอดจะถูกนับซ้ำ", statementAccountLabel(code), previous, column.Title))
				continue
			}
			columnOf[code] = column.Title
		}
	}
	if _, ok := columnOf[statementCurrentEarnings]; !ok {
		warnings = append(warnings, "ยังไม่ได้เลือก “รวมกำไร (ขาดทุน) ที่ยังไม่ปิดบัญชี” ในคอลัมน์ใด ยอดรวมจะไม่ตรงกับงบฐานะการเงิน")
	}
	rowOf := map[string]string{}
	for _, row := range rows {
		if row.RowType != "account" || equityRowBasis(row) != "movement" {
			continue
		}
		for _, code := range row.AccountCodes {
			if previous, ok := rowOf[code]; ok && previous != row.Title {
				warnings = append(warnings, fmt.Sprintf("บัญชี %s อยู่ในหลายบรรทัด (%s, %s) ยอดจะถูกนับซ้ำ", statementAccountLabel(code), previous, row.Title))
				continue
			}
			rowOf[code] = row.Title
			if _, ok := columnOf[code]; !ok {
				warnings = append(warnings, fmt.Sprintf("บัญชี %s ในบรรทัด %s ไม่อยู่ในคอลัมน์ใด จึงไม่แสดงยอด", statementAccountLabel(code), row.Title))
			}
		}
	}
	return warnings
}

func evaluateEquityStatement(rows []StatementRow, columns []StatementColumn, balances map[string]statementBalance) ([]map[int]decimal.Decimal, []string) {
	values := make([]map[int]decimal.Decimal, len(columns))
	warnings := []string{}
	for i, column := range columns {
		inColumn := map[string]bool{}
		var opening, closing, movement decimal.Decimal // ด้านปกติตามประเภทบัญชี (ส่วนของเจ้าของ = เครดิตเป็นบวก)
		for _, code := range column.AccountCodes {
			if inColumn[code] {
				continue
			}
			inColumn[code] = true
			if account, ok := balances[code]; ok {
				opening = opening.Add(statementSigned(account, account.opening, ""))
				closing = closing.Add(statementSigned(account, account.balance, ""))
				movement = movement.Add(statementSigned(account, account.movement, ""))
			}
		}
		result := map[int]decimal.Decimal{}
		assigned := decimal.Zero
		otherRow, hasOther, otherReverse := 0, false, false
		for _, row := range rows {
			if row.RowType != "account" {
				continue
			}
			basis := equityRowBasis(row)
			if basis == "other" {
				if !hasOther {
					otherRow, hasOther, otherReverse = row.RowNo, true, row.ReverseSign
				}
				continue
			}
			sum := decimal.Zero
			for code := range inColumn {
				if basis == "movement" && !contains(row.AccountCodes, code) {
					continue
				}
				account, ok := balances[code]
				if !ok {
					continue
				}
				switch basis {
				case "opening":
					sum = sum.Add(statementSigned(account, account.opening, row.NormalBalance))
				case "closing":
					sum = sum.Add(statementSigned(account, account.balance, row.NormalBalance))
				default:
					sum = sum.Add(statementSigned(account, account.movement, row.NormalBalance))
					assigned = assigned.Add(statementSigned(account, account.movement, ""))
				}
			}
			if row.ReverseSign {
				sum = sum.Neg()
			}
			result[row.RowNo] = sum
		}
		// ส่วนที่เหลือรวมรายการปิดบัญชีที่ไม่หักล้างในคอลัมน์ด้วย เพื่อให้ยอดต้นงวด + บรรทัดทั้งหมด = ยอดปลายงวดเสมอ
		rest := closing.Sub(opening).Sub(assigned)
		if hasOther {
			result[otherRow] = rest
			if otherReverse {
				result[otherRow] = rest.Neg()
			}
		} else if !rest.IsZero() {
			warnings = append(warnings, fmt.Sprintf("คอลัมน์ %s มีรายการ %s ที่ยังไม่อยู่ในบรรทัดใด กรุณาเลือกบัญชีให้บรรทัด หรือเพิ่มบรรทัดรายการอื่น", column.Title, statementWarningAmount(rest, 2)))
		}
		if closingNet := closing.Sub(opening).Sub(movement); !closingNet.IsZero() {
			warnings = append(warnings, fmt.Sprintf("คอลัมน์ %s มีรายการปิดบัญชีที่ไม่หักล้างกัน %s กรุณาเลือก “รวมกำไร (ขาดทุน) ที่ยังไม่ปิดบัญชี” ในคอลัมน์เดียวกับบัญชีกำไรสะสม", column.Title, statementWarningAmount(closingNet, 2)))
		}
		evaluateStatementFormulas(rows, result)
		values[i] = result
	}
	return values, warnings
}
