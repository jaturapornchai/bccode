package generalledger

import (
	"testing"

	"github.com/shopspring/decimal"
)

func dec(value string) decimal.Decimal { return decimal.RequireFromString(value) }

func TestEvaluateStatementFormula(t *testing.T) {
	values := map[int]decimal.Decimal{10: dec("100.10"), 20: dec("200.20"), 30: dec("-50"), 40: dec("3")}
	cases := []struct{ formula, want string }{
		{"SUM(R10:R30)", "250.30"},
		{"sum(30..10)", "250.30"},
		{"R10 + R20 * 2", "500.50"},
		{"(R10 + R20) * 2", "600.60"},
		{"-R30", "50"},
		{"R20 / R40", "66.73333333"}, // ตัด (ไม่ปัด) ที่ 8 ตำแหน่ง
		{"R10 / 0", "0"},
		{"R10 * 0.333333333", "33.36666663"},
		{"R99 + R10", "100.10"}, // แถวที่ไม่มี = 0
		{"R10 + abc", "100.10"},
		{"", "0"},
	}
	for _, c := range cases {
		if got := evaluateStatementFormula(c.formula, values, 50); !got.Equal(dec(c.want)) {
			t.Errorf("%q = %s, want %s", c.formula, got, c.want)
		}
	}
	// แถวอ้างตัวเอง = 0 และช่วงรวมไม่นับแถวตัวเอง
	if got := evaluateStatementFormula("R10 + R50", map[int]decimal.Decimal{10: dec("1"), 50: dec("9")}, 50); !got.Equal(dec("1")) {
		t.Errorf("self reference = %s", got)
	}
	if got := evaluateStatementFormula("SUM(R10:R50)", map[int]decimal.Decimal{10: dec("1"), 50: dec("9")}, 50); !got.Equal(dec("1")) {
		t.Errorf("range including itself = %s", got)
	}
}

func TestEvaluateStatementBalanceVersusPeriod(t *testing.T) {
	balances := map[string]statementBalance{
		"1000": {accountType: "asset", balance: dec("150"), movement: dec("80")},
		"1900": {accountType: "asset", balance: dec("-20"), movement: dec("-5")},
		"4000": {accountType: "income", balance: dec("-300"), movement: dec("-120")},
		"5000": {accountType: "expense", balance: dec("100"), movement: dec("40")},
		statementCurrentEarnings: {accountType: "equity", balance: dec("-200"), movement: dec("-80")},
	}
	rows := []StatementRow{
		{RowNo: 10, RowType: "account", AccountCodes: []string{"1000", "missing"}},
		{RowNo: 15, RowType: "account", AccountCodes: []string{"1900"}, NormalBalance: "credit", ReverseSign: true},
		{RowNo: 20, RowType: "account", AccountCodes: []string{"4000"}},
		{RowNo: 30, RowType: "account", AccountCodes: []string{"5000"}},
		{RowNo: 40, RowType: "formula", Formula: "R20 - R30"},
		{RowNo: 50, RowType: "account", AccountCodes: []string{statementCurrentEarnings}},
		{RowNo: 60, RowType: "subtotal", Formula: "SUM(R10:R15)"},
		{RowNo: 70, RowType: "header", Title: "หัวข้อ"},
	}
	position := evaluateStatement(rows, balances, false)
	for rowNo, want := range map[int]string{10: "150", 15: "-20", 20: "300", 30: "100", 40: "200", 50: "200", 60: "130"} {
		if !position[rowNo].Equal(dec(want)) {
			t.Errorf("balance row %d = %s, want %s", rowNo, position[rowNo], want)
		}
	}
	if _, ok := position[70]; ok {
		t.Error("header row got an amount")
	}
	period := evaluateStatement(rows, balances, true)
	for rowNo, want := range map[int]string{10: "80", 20: "120", 30: "40", 40: "80", 50: "80"} {
		if !period[rowNo].Equal(dec(want)) {
			t.Errorf("period row %d = %s, want %s", rowNo, period[rowNo], want)
		}
	}
}

func TestPriorRange(t *testing.T) {
	cur := FiscalYear{Code: "2569", StartDate: "2026-01-01", EndDate: "2026-12-31"}
	prior := FiscalYear{Code: "2568", StartDate: "2025-01-01", EndDate: "2025-12-31"}
	short := FiscalYear{Code: "2568", StartDate: "2025-07-01", EndDate: "2025-12-31"} // ปีแรกของกิจการสั้นกว่า 12 เดือน
	cases := []struct {
		prior          FiscalYear
		from, to       string
		wantFrom, want string
		ok             bool
	}{
		{prior, "2026-01-01", "2026-12-31", "2025-01-01", "2025-12-31", true},
		{prior, "2026-04-01", "2026-06-30", "2025-04-01", "2025-06-30", true},
		{short, "2026-01-01", "2026-12-31", "2025-07-01", "2025-12-31", true},
		{short, "2026-01-01", "2026-03-31", "2025-07-01", "2025-03-31", false},
	}
	for _, c := range cases {
		from, to, ok := priorRange(cur, c.prior, c.from, c.to)
		if ok != c.ok || ok && (from != c.wantFrom || to != c.want) {
			t.Errorf("priorRange(%s,%s) = %s %s %v", c.from, c.to, from, to, ok)
		}
	}
	for in, want := range map[string]string{"2028-02-29": "2027-02-28", "2029-02-28": "2028-02-29", "2026-06-15": "2025-06-15", "2026-04-30": "2025-04-30"} {
		if got := oneYearEarlier(in); got != want {
			t.Errorf("oneYearEarlier(%s) = %s, want %s", in, got, want)
		}
	}
}
