package generalledger

import (
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// ประกาศกรมพัฒนาธุรกิจการค้า เรื่อง กำหนดรายการย่อที่ต้องมีในงบการเงิน พ.ศ. 2566 ข้อ 7: รายการที่กิจการไม่มีไม่ต้องแสดง.
// รูปแบบงบที่ตั้ง hidezerorows จึงตัดแถวออกจากผลงบที่ backend (ไม่ใช่ที่หน้าจอ) — แถวยอดเงินที่เป็นศูนย์ทุกคอลัมน์
// (ยกเว้นแถวที่ผู้ใช้ติ๊ก "แสดงศูนย์" และแถวสูตร/รวมยอดที่รวมจากแถวที่ยังแสดง), หัวข้อที่แถวยอดเงินข้างใต้ถูกซ่อนหมด และบรรทัดว่าง/เส้นคั่นที่ไม่มีอะไรให้คั่นแล้ว.
// งบส่วนของผู้ถือหุ้นทำทีละชุดปี (block) และตัดคอลัมน์องค์ประกอบที่เป็นศูนย์ทั้งงบด้วย

func statementHideZero(template Master) bool {
	return template.GlobalStyle != nil && template.GlobalStyle.HideZeroRows
}

// คีย์ยอดเงินที่แสดงของงบ = คอลัมน์ยอดเงินของรายงาน (งบปกติ amount/prioramount, งบส่วนของผู้ถือหุ้น c1..cn และ total)
func statementAmountKeys(columns []ReportColumn) []string {
	keys := []string{}
	for _, column := range columns {
		if column.Amount {
			keys = append(keys, column.Key)
		}
	}
	return keys
}

// hideZeroStatementRows ตัดแถวที่ไม่มียอดออก ทีละชุดแถวที่ต่อเนื่องและมี block เดียวกัน (งบปกติ block ว่างทั้งงบ = ชุดเดียว);
// refs = แถวที่สูตรของแต่ละแถวสูตร/รวมยอดอ้างถึง (statementFormulaRefs ของรูปแบบงบ ใช้ได้กับทุกชุดปีเพราะแถวชุดเดียวกัน)
func hideZeroStatementRows(rows []map[string]string, amountKeys []string, refs map[int][]int) []map[string]string {
	kept := make([]map[string]string, 0, len(rows))
	for start := 0; start < len(rows); {
		end := start + 1
		for end < len(rows) && rows[end]["block"] == rows[start]["block"] {
			end++
		}
		kept = append(kept, hideZeroStatementBlock(rows[start:end], amountKeys, refs)...)
		start = end
	}
	return kept
}

func hideZeroStatementBlock(rows []map[string]string, amountKeys []string, refs map[int][]int) []map[string]string {
	hidden := make([]bool, len(rows))
	position := map[int]int{}
	for i, row := range rows {
		hidden[i] = statementAmountRow(row["rowtype"]) && row["showzero"] != "true" && statementRowZero(row, amountKeys)
		if rowNo, err := strconv.Atoi(strings.TrimSpace(row["rowno"])); err == nil {
			if _, ok := position[rowNo]; !ok {
				position[rowNo] = i
			}
		}
	}
	// ยอดรวม/สูตรที่เป็นศูนย์แต่รวมจากแถวที่ยังแสดง (เช่น กู้ยืมแล้วคืนในปีเดียวกัน, ซื้อแล้วขายอุปกรณ์) ต้องแสดง
	// ไม่งั้นงบที่พิมพ์มีรายการแต่ไม่มียอดรวมของส่วนนั้น; ทำซ้ำจนนิ่งเพราะยอดรวมที่กลับมาแสดงอาจเป็นส่วนของยอดรวมชั้นถัดไป
	for changed := true; changed; {
		changed = false
		for i, row := range rows {
			if !hidden[i] || row["rowtype"] != "formula" && row["rowtype"] != "subtotal" {
				continue
			}
			rowNo, _ := strconv.Atoi(strings.TrimSpace(row["rowno"]))
			for _, ref := range refs[rowNo] {
				if j, ok := position[ref]; ok && !hidden[j] {
					hidden[i], changed = false, true
					break
				}
			}
		}
	}
	// หัวข้อซ่อนเมื่อส่วนของมันมีแถวยอดเงินและทุกแถวถูกซ่อน (บรรทัดว่าง/เส้นคั่นไม่ตัดส่วน). รูปแบบของส่วนดูจากแถวแรกหลังหัวข้อ:
	//   - แถวแรกย่อหน้าลึกกว่าหัวข้อ = ส่วนแบบซ้อน: ส่วนจบที่แถวแรกที่ย่อหน้าไม่ลึกกว่าหัวข้อ ไม่ว่าชนิดใด (เช่น "จัดสรรแล้ว"
	//     จบก่อน "ยังไม่ได้จัดสรร" ที่ย่อหน้าเท่ากัน, "กิจกรรมจัดหาเงิน" จบก่อนบรรทัดเงินสดเพิ่มขึ้นสุทธิ, "รวมสินทรัพย์หมุนเวียน")
	//   - แถวแรกย่อหน้าเท่าหัวข้อ = ส่วนแบบแบน: ส่วนจบที่หัวข้อที่ย่อหน้าเท่ากัน หรือแถวที่ย่อหน้าน้อยกว่าหัวข้อ
	headerHidden := make([]bool, len(rows))
	for i, row := range rows {
		if row["rowtype"] != "header" {
			continue
		}
		indent := statementRowIndent(row)
		amounts, visible, nested, first := 0, false, false, true
		for j := i + 1; j < len(rows); j++ {
			nextType := rows[j]["rowtype"]
			if statementSpacerRow(nextType) {
				continue
			}
			nextIndent := statementRowIndent(rows[j])
			if first {
				nested, first = nextIndent > indent, false
			}
			if nextIndent < indent || nextIndent == indent && (nested || nextType == "header") {
				break
			}
			if statementAmountRow(nextType) {
				amounts++
				visible = visible || !hidden[j]
			}
		}
		headerHidden[i] = amounts > 0 && !visible
	}
	kept := make([]map[string]string, 0, len(rows))
	afterSpacer := true // ต้นชุดยังไม่มีแถวที่แสดง: บรรทัดว่าง/เส้นคั่นนำหน้าไม่ต้องแสดง
	for i, row := range rows {
		if hidden[i] || headerHidden[i] {
			continue
		}
		spacer := statementSpacerRow(row["rowtype"])
		if spacer && afterSpacer {
			continue
		}
		kept = append(kept, row)
		afterSpacer = spacer
	}
	for len(kept) > 0 && statementSpacerRow(kept[len(kept)-1]["rowtype"]) {
		kept = kept[:len(kept)-1]
	}
	return kept
}

// งบส่วนของผู้ถือหุ้น: คอลัมน์องค์ประกอบที่เป็นศูนย์ทุกแถวทุกปีไม่ต้องแสดง — คงคอลัมน์รวมไว้เสมอ (คีย์ในแถวคงอยู่ได้)
func hideZeroEquityColumns(columns []ReportColumn, rows []map[string]string) []ReportColumn {
	kept := make([]ReportColumn, 0, len(columns))
	for _, column := range columns {
		if column.Amount && column.Key != "total" && statementColumnZero(rows, column.Key) {
			continue
		}
		kept = append(kept, column)
	}
	return kept
}

func statementColumnZero(rows []map[string]string, key string) bool {
	for _, row := range rows {
		if !statementRowZero(row, []string{key}) {
			return false
		}
	}
	return true
}

// ยอดในแถวปัดตามทศนิยมของรูปแบบแล้ว (StringFixed) จึงเทียบศูนย์ ณ ทศนิยมที่แสดง; ค่าว่าง = คอลัมน์นี้ไม่แสดงยอด
func statementRowZero(row map[string]string, keys []string) bool {
	for _, key := range keys {
		text := strings.TrimSpace(row[key])
		if text == "" {
			continue
		}
		value, err := decimal.NewFromString(text)
		if err != nil || !value.IsZero() {
			return false
		}
	}
	return true
}

func statementRowIndent(row map[string]string) int {
	indent, _ := strconv.Atoi(strings.TrimSpace(row["indent"]))
	return indent
}

func statementSpacerRow(rowType string) bool { return rowType == "blank" || rowType == "divider" }
