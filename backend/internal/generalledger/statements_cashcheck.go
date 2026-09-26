package generalledger

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// งบกระแสเงินสด: ตรวจว่าเงินสดปลายงวดตามงบเท่ากับยอดคงเหลือตามบัญชี ไม่ให้ออกงบที่กระทบยอดไม่ได้โดยไม่รู้ตัว.
// ความสัมพันธ์มาจากสูตรของผู้ใช้เองเท่านั้น (ไม่เดาบัญชีเงินสดจากรหัส/ชื่อบัญชี):
//   - บรรทัดยอดต้นงวด = แถวบัญชีที่เลือกฐาน "ยอดต้นงวด" (amountbasis opening) — ผู้ใช้เลือกบัญชีเงินสดไว้ที่นี่
//   - แถวสูตรที่อ่านเฉพาะบรรทัดยอดต้นงวด (เช่น รวมเงินสดต้นงวดหลายบัญชี หรือยกยอดต้นงวดมาแสดงซ้ำ) = ยอดต้นงวดรวม ไม่ใช่ยอดปลายงวด
//   - บรรทัดยอดปลายงวด = แถวสูตร/รวมยอดที่อ่านยอดต้นงวด (หรือยอดต้นงวดรวม) คู่กับแถวอื่น (Rnn หรือช่วง SUM ตามกติกาเดียวกับการคำนวณ)
//   - แถวสูตรที่อ่านแถวบัญชีฐาน "ยอดปลายงวด" ด้วย = ผลต่างของยอดคงเหลือ (เช่น สินค้าคงเหลือ/ลูกหนี้ ต้นงวด − ปลายงวด
//     ในส่วนเงินทุนหมุนเวียน หรือ เงินสดปลายงวด − ต้นงวด) ไม่ใช่เงินสดปลายงวด จึงไม่ตรวจ; แถวสูตรที่อ่านเฉพาะยอดปลายงวด = ยอดปลายงวดรวม
//     นับเหมือนแถวยอดปลายงวด (ผลต่างที่อ่านผ่านยอดรวมนี้ก็ไม่ตรวจ)
//   - ยอดตามบัญชี = คำนวณสูตรของบรรทัดยอดปลายงวดเอง โดยให้บรรทัดยอดต้นงวดที่อ้างเป็นยอดคงเหลือปลายงวดของบัญชีในบรรทัดนั้น
//     (ด้านปกติ/กลับเครื่องหมายเดียวกับตอนคำนวณงบ) และแถวบัญชีอื่นเป็นศูนย์ — คงเครื่องหมายของสูตร เช่น เงินสด − เงินเบิกเกินบัญชี
// ตรวจจากค่าเต็มความละเอียดก่อนซ่อนแถวศูนย์ จึงไม่ขึ้นกับ hidezerorows

// ReportCheck ผลตรวจยอดของงบกับบัญชี 1 บรรทัดต่อ 1 งวด (key = คอลัมน์ยอดเงิน amount / prioramount)
type ReportCheck struct {
	Key        string `json:"key"`
	FiscalYear string `json:"fiscalyear"`
	RowNo      int    `json:"rowno"`
	Title      string `json:"title"`
	Statement  string `json:"statement"`
	Book       string `json:"book"`
	Difference string `json:"difference"`
	Matched    bool   `json:"matched"`
}

type statementCashEnding struct {
	row     StatementRow
	opening []StatementRow // บรรทัดยอดต้นงวดที่สูตรอ้าง เรียงตามลำดับบรรทัด
	rows    []StatementRow // แถวทั้งรูปแบบงบ ใช้คำนวณยอดตามบัญชีผ่านสูตรของผู้ใช้เอง
}

// statementCashEndings หาบรรทัดยอดปลายงวดของงบกระแสเงินสด; บรรทัดที่อ้างเฉพาะบรรทัดยอดต้นงวดที่ยังไม่เลือกบัญชี
// ตรวจกับบัญชีไม่ได้ จึงคืนเป็นคำเตือนแทน (ยอดตามบัญชีเป็นศูนย์เพราะไม่มีบัญชี ไม่ใช่เพราะเงินสดเป็นศูนย์)
func statementCashEndings(template Master) ([]statementCashEnding, []string) {
	if template.StatementType != "cash_flow" {
		return nil, nil
	}
	opening := map[int]StatementRow{}
	closing := map[int]bool{}
	for _, row := range template.Rows {
		if row.RowType != "account" {
			continue
		}
		switch row.AmountBasis {
		case "opening":
			opening[row.RowNo] = row
		case "closing":
			closing[row.RowNo] = true
		}
	}
	endings, warnings := []statementCashEnding{}, []string{}
	if len(opening) == 0 {
		return endings, warnings
	}
	// covers = บรรทัดยอดต้นงวดที่แต่ละแถวแทนอยู่: แถวยอดต้นงวดเอง และแถวสูตรที่รวมเฉพาะยอดต้นงวด (ยอดต้นงวดรวม)
	covers := map[int][]int{}
	for rowNo := range opening {
		covers[rowNo] = []int{rowNo}
	}
	refs := statementFormulaRefs(template.Rows)
	for _, row := range template.Rows {
		if row.RowType != "formula" && row.RowType != "subtotal" {
			continue
		}
		under, other, withClosing := map[int]bool{}, false, false
		for _, ref := range refs[row.RowNo] {
			if closing[ref] {
				withClosing = true
				continue
			}
			rowNos, ok := covers[ref]
			if !ok {
				other = true
				continue
			}
			for _, rowNo := range rowNos {
				under[rowNo] = true
			}
		}
		if withClosing {
			// แถวที่รวมเฉพาะยอดปลายงวด (เช่น สินค้าคงเหลือปลายงวด = วัตถุดิบ + สินค้าสำเร็จรูป) = ยอดปลายงวดรวม: แถวที่อ่านมันคู่กับยอดต้นงวด
			// เป็นผลต่างของยอดคงเหลือเหมือนอ่านแถวยอดปลายงวดตรง ๆ; แถวผลต่างเอง (ปนต้นงวด/แถวอื่น) ไม่ทำเครื่องหมาย ไม่งั้นลามถึงบรรทัดเงินสดปลายงวด
			if len(under) == 0 && !other {
				closing[row.RowNo] = true
			}
			continue
		}
		if len(under) == 0 {
			continue
		}
		rowNos := make([]int, 0, len(under))
		for rowNo := range under {
			rowNos = append(rowNos, rowNo)
		}
		sort.Ints(rowNos)
		if !other {
			covers[row.RowNo] = rowNos
			continue
		}
		ending, accounts := statementCashEnding{row: row, rows: template.Rows}, 0
		for _, rowNo := range rowNos {
			ending.opening = append(ending.opening, opening[rowNo])
			accounts += len(opening[rowNo].AccountCodes)
		}
		if accounts == 0 {
			warnings = append(warnings, fmt.Sprintf("%s: ยังไม่ได้เลือกบัญชีเงินสดในบรรทัดยอดต้นงวดที่สูตรอ้างถึง จึงยังตรวจเงินสดปลายงวดกับยอดคงเหลือตามบัญชีไม่ได้", statementRowLabel(row)))
			continue
		}
		endings = append(endings, ending)
	}
	return endings, warnings
}

// statementCashChecks ตรวจหนึ่งงวด: values = ค่าเต็มความละเอียดของงบ, balances = ยอดบัญชีของงวดเดียวกัน
func statementCashChecks(endings []statementCashEnding, balances map[string]statementBalance, values map[int]decimal.Decimal, period ReportPeriod, scale int32) ([]ReportCheck, []string) {
	checks, warnings := []ReportCheck{}, []string{}
	for _, ending := range endings {
		// ยอดตามบัญชีคำนวณด้วยสูตรของผู้ใช้เอง (เครื่องหมาย/สัมประสิทธิ์ เช่น เงินสด − เงินเบิกเกินบัญชี และกลับเครื่องหมายของแถวสูตร):
		// บรรทัดยอดต้นงวดที่สูตรอ้าง = ยอดคงเหลือปลายงวดของบัญชีในบรรทัดนั้น, แถวบัญชีอื่นทุกแถว = 0
		bookValues := map[int]decimal.Decimal{}
		for _, row := range ending.rows {
			if row.RowType == "account" {
				bookValues[row.RowNo] = decimal.Zero
			}
		}
		for _, row := range ending.opening {
			bookValues[row.RowNo] = statementAccountRowValue(row, balances, "closing", true)
		}
		evaluateStatementFormulas(ending.rows, bookValues)
		book := bookValues[ending.row.RowNo]
		statement := values[ending.row.RowNo]
		difference := statement.Sub(book)
		check := ReportCheck{Key: period.Key, FiscalYear: period.FiscalYear, RowNo: ending.row.RowNo, Title: ending.row.Title,
			Statement: statement.StringFixed(scale), Book: book.StringFixed(scale), Difference: difference.StringFixed(scale),
			Matched: difference.Round(scale).IsZero()}
		checks = append(checks, check)
		if !check.Matched {
			prefix := ""
			if period.Key != "amount" {
				prefix = "ปี " + period.FiscalYear + ": "
			}
			warnings = append(warnings, fmt.Sprintf("%sเงินสดปลายงวดตามงบ (%s) %s ไม่ตรงกับยอดคงเหลือตามบัญชี %s ผลต่าง %s กรุณาตรวจว่าเลือกบัญชีครบทุกบรรทัด",
				prefix, statementRowLabel(ending.row), check.Statement, check.Book, check.Difference))
		}
	}
	return checks, warnings
}

func statementRowLabel(row StatementRow) string {
	label := "บรรทัด " + strconv.Itoa(row.RowNo)
	if title := strings.TrimSpace(row.Title); title != "" {
		label += " " + title
	}
	return label
}
