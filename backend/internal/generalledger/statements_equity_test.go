package generalledger

import (
	"strings"
	"testing"
)

// ฐานยอดของแถว: ต้นงวด/ปลายงวด/ความเคลื่อนไหว ทับค่าเริ่มของชนิดงบ (ใช้กับงบต้นทุนขาย "สินค้าต้นงวด/ปลายงวด" และงบกระแสเงินสด)
func TestEvaluateStatementAmountBasis(t *testing.T) {
	balances := map[string]statementBalance{
		"1300": {accountType: "asset", opening: dec("40"), balance: dec("65"), movement: dec("25")},
	}
	rows := []StatementRow{
		{RowNo: 10, RowType: "account", AccountCodes: []string{"1300"}, AmountBasis: "opening"},
		{RowNo: 20, RowType: "account", AccountCodes: []string{"1300"}, AmountBasis: "closing"},
		{RowNo: 30, RowType: "account", AccountCodes: []string{"1300"}, AmountBasis: "movement"},
		{RowNo: 40, RowType: "account", AccountCodes: []string{"1300"}},
		{RowNo: 50, RowType: "formula", Formula: "R10 + R30 - R20"},
	}
	for periodic, defaultValue := range map[bool]string{true: "25", false: "65"} {
		values := evaluateStatement(rows, balances, periodic)
		for rowNo, want := range map[int]string{10: "40", 20: "65", 30: "25", 40: defaultValue, 50: "0"} {
			if !values[rowNo].Equal(dec(want)) {
				t.Errorf("periodic=%v row %d = %s, want %s", periodic, rowNo, values[rowNo], want)
			}
		}
	}
	if !statementPeriodic("cash_flow") || !statementPeriodic("equity") || statementPeriodic("balance_sheet") || statementPeriodic("custom") {
		t.Fatal("statementPeriodic")
	}
}

// งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น: ต้นงวด + ทุกบรรทัด + รายการอื่น = ปลายงวด ในทุกคอลัมน์
func TestEvaluateEquityStatement(t *testing.T) {
	balances := map[string]statementBalance{
		"3000":                   {accountType: "equity", opening: dec("-1000"), balance: dec("-1500"), movement: dec("-500")},
		"3100":                   {accountType: "equity", opening: dec("-69.75"), balance: dec("-74.75"), movement: dec("-5")},
		"3300":                   {accountType: "equity", balance: dec("20"), movement: dec("20")},
		statementCurrentEarnings: {accountType: "equity", balance: dec("-150"), movement: dec("-150")},
	}
	columns := []StatementColumn{
		{ID: "cap", Title: "ทุนที่ชำระแล้ว", AccountCodes: []string{"3000"}},
		{ID: "re", Title: "กำไรสะสม", AccountCodes: []string{"3100", "3300", statementCurrentEarnings, "3300"}},
	}
	rows := []StatementRow{
		{RowNo: 10, RowType: "account", AmountBasis: "opening"},
		{RowNo: 20, RowType: "account", AccountCodes: []string{"3000"}},
		{RowNo: 30, RowType: "account", AccountCodes: []string{"3300"}},
		{RowNo: 40, RowType: "account", AccountCodes: []string{statementCurrentEarnings}},
		{RowNo: 50, RowType: "account", AmountBasis: "other"},
		{RowNo: 60, RowType: "account", AmountBasis: "closing"},
		{RowNo: 70, RowType: "formula", Formula: "R10 + R20 + R30 + R40 + R50 - R60"},
	}
	values, warnings := evaluateEquityStatement(rows, columns, balances)
	if len(warnings) != 0 || len(equityTemplateWarnings(rows, columns)) != 0 {
		t.Fatalf("warnings = %v %v", warnings, equityTemplateWarnings(rows, columns))
	}
	want := []map[int]string{
		{10: "1000", 20: "500", 30: "0", 40: "0", 50: "0", 60: "1500", 70: "0"},
		{10: "69.75", 20: "0", 30: "-20", 40: "150", 50: "5", 60: "204.75", 70: "0"},
	}
	for i := range columns {
		for rowNo, amount := range want[i] {
			if !values[i][rowNo].Equal(dec(amount)) {
				t.Errorf("column %s row %d = %s, want %s", columns[i].ID, rowNo, values[i][rowNo], amount)
			}
		}
	}

	// ไม่มีบรรทัดรายการอื่น + ไม่ได้ใส่กำไรที่ยังไม่ปิดไว้ในคอลัมน์ → เตือน ไม่เงียบ
	columns[1].AccountCodes = []string{"3100", "3300"}
	_, warnings = evaluateEquityStatement(rows[:4], columns, balances)
	template := equityTemplateWarnings(append(rows[:4:4], StatementRow{RowNo: 80, RowType: "account", Title: "ซ้ำ", AccountCodes: []string{"3000"}}), append(columns, StatementColumn{Title: "อื่น", AccountCodes: []string{"3000"}}))
	joined := strings.Join(append(warnings, template...), "\n")
	for _, text := range []string{"ยังไม่อยู่ในบรรทัดใด", "ยังไม่ได้เลือก “รวมกำไร (ขาดทุน) ที่ยังไม่ปิดบัญชี” ในคอลัมน์ใด", "หลายคอลัมน์", "หลายบรรทัด", "ไม่อยู่ในคอลัมน์ใด จึงไม่แสดงยอด"} {
		if !strings.Contains(joined, text) {
			t.Errorf("missing warning %q in\n%s", text, joined)
		}
	}
	// รายการปิดบัญชีเข้ากำไรสะสม แต่กำไรงวดอยู่คนละคอลัมน์ → รายการปิดบัญชีไม่หักล้าง
	closed := map[string]statementBalance{"3100": {accountType: "equity", balance: dec("-69.75")}}
	_, warnings = evaluateEquityStatement(rows, []StatementColumn{{Title: "กำไรสะสม", AccountCodes: []string{"3100"}}}, closed)
	if len(warnings) != 1 || !strings.Contains(warnings[0], "ปิดบัญชีที่ไม่หักล้าง") {
		t.Fatalf("closing warning = %v", warnings)
	}
}
