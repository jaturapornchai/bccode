package generalledger

import (
	"strconv"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

// งบกระแสเงินสดแบบย่อ: กำไร → เงินสดเพิ่มขึ้นสุทธิ → + เงินสดต้นงวด = เงินสดปลายงวด (แถว 40 อ้างแถวยอดต้นงวด 30)
func cashCheckTemplate(statementType, ending string) Master {
	return Master{StatementType: statementType, Rows: []StatementRow{
		{RowNo: 10, RowType: "account", Title: "กำไร (ขาดทุน) สุทธิ", AccountCodes: []string{statementCurrentEarnings}, NormalBalance: "credit"},
		{RowNo: 20, RowType: "formula", Title: "เงินสดเพิ่มขึ้นสุทธิ", Formula: "R10"},
		{RowNo: 30, RowType: "account", Title: "เงินสดต้นงวด", AccountCodes: []string{"1000"}, NormalBalance: "debit", AmountBasis: "opening"},
		{RowNo: 40, RowType: "subtotal", Title: "เงินสดปลายงวด", Formula: ending},
		{RowNo: 50, RowType: "header", Title: "หัวข้อ"},
	}}
}

func cashCheckBalances(cashClosing string) map[string]statementBalance {
	return map[string]statementBalance{
		"1000":                   {accountType: "asset", opening: dec("100"), movement: dec("50"), balance: dec(cashClosing)},
		"4000":                   {accountType: "income", movement: dec("-80"), balance: dec("-80")},
		"5000":                   {accountType: "expense", movement: dec("30"), balance: dec("30")},
		statementCurrentEarnings: {accountType: "equity", movement: dec("-50"), balance: dec("-50")},
	}
}

func runCashChecks(t *testing.T, template Master, balances map[string]statementBalance, period ReportPeriod) ([]ReportCheck, []string) {
	t.Helper()
	endings, warnings := statementCashEndings(template)
	if len(warnings) != 0 {
		t.Fatalf("template warnings = %v", warnings)
	}
	values := evaluateStatement(template.Rows, balances, statementPeriodic(template.StatementType))
	return statementCashChecks(endings, balances, values, period, 2)
}

func TestStatementCashChecksMatchAndMismatch(t *testing.T) {
	current := ReportPeriod{Key: "amount", FiscalYear: "2027"}
	checks, warnings := runCashChecks(t, cashCheckTemplate("cash_flow", "R20 + R30"), cashCheckBalances("150"), current)
	want := ReportCheck{Key: "amount", FiscalYear: "2027", RowNo: 40, Title: "เงินสดปลายงวด", Statement: "150.00", Book: "150.00", Difference: "0.00", Matched: true}
	if len(checks) != 1 || checks[0] != want || len(warnings) != 0 {
		t.Fatalf("matched checks = %+v warnings = %v", checks, warnings)
	}
	// ยอดคงเหลือตามบัญชีมีรายการที่งบไม่ได้นับ (เช่น ไม่ได้เลือกบัญชีในบรรทัดใด) → ไม่ตรง + คำเตือนภาษาไทย
	checks, warnings = runCashChecks(t, cashCheckTemplate("cash_flow", "R20 + R30"), cashCheckBalances("170.5"), current)
	if len(checks) != 1 || checks[0].Matched || checks[0].Book != "170.50" || checks[0].Difference != "-20.50" {
		t.Fatalf("mismatch checks = %+v", checks)
	}
	if len(warnings) != 1 || warnings[0] != "เงินสดปลายงวดตามงบ (บรรทัด 40 เงินสดปลายงวด) 150.00 ไม่ตรงกับยอดคงเหลือตามบัญชี 170.50 ผลต่าง -20.50 กรุณาตรวจว่าเลือกบัญชีครบทุกบรรทัด" {
		t.Fatalf("mismatch warnings = %q", warnings)
	}
}

func TestStatementCashChecksPriorPeriodPrefix(t *testing.T) {
	checks, warnings := runCashChecks(t, cashCheckTemplate("cash_flow", "R20 + R30"), cashCheckBalances("160"), ReportPeriod{Key: "prioramount", FiscalYear: "2026"})
	if len(checks) != 1 || checks[0].Key != "prioramount" || checks[0].FiscalYear != "2026" || checks[0].Difference != "-10.00" {
		t.Fatalf("prior checks = %+v", checks)
	}
	if len(warnings) != 1 || !strings.HasPrefix(warnings[0], "ปี 2026: เงินสดปลายงวดตามงบ (บรรทัด 40 เงินสดปลายงวด)") {
		t.Fatalf("prior warnings = %q", warnings)
	}
}

// ช่วง SUM เลือกแถวแบบเดียวกับการคำนวณ (ทั้ง R10:R30 และ 10..30) และไม่นับแถวตัวเอง
func TestStatementCashChecksSumRangeReference(t *testing.T) {
	for _, formula := range []string{"SUM(R10:R30)", "sum(10..30)", "SUM(R30:R10)"} {
		template := Master{StatementType: "cash_flow", Rows: []StatementRow{
			{RowNo: 10, RowType: "account", AccountCodes: []string{statementCurrentEarnings}, NormalBalance: "credit"},
			{RowNo: 20, RowType: "account", Title: "เงินสดต้นงวด", AccountCodes: []string{"1000"}, NormalBalance: "debit", AmountBasis: "opening"},
			{RowNo: 30, RowType: "subtotal", Title: "เงินสดปลายงวด", Formula: formula},
		}}
		endings, _ := statementCashEndings(template)
		if len(endings) != 1 || endings[0].row.RowNo != 30 || len(endings[0].opening) != 1 || endings[0].opening[0].RowNo != 20 {
			t.Fatalf("%s endings = %+v", formula, endings)
		}
		checks, _ := runCashChecks(t, template, cashCheckBalances("150"), ReportPeriod{Key: "amount", FiscalYear: "2027"})
		if len(checks) != 1 || !checks[0].Matched || checks[0].Statement != "150.00" {
			t.Fatalf("%s checks = %+v", formula, checks)
		}
	}
	// ช่วงที่ไม่ครอบแถวยอดต้นงวด = ไม่ใช่บรรทัดยอดปลายงวด
	template := cashCheckTemplate("cash_flow", "SUM(R10:R20)")
	if endings, _ := statementCashEndings(template); len(endings) != 0 {
		t.Fatalf("range without opening row = %+v", endings)
	}
}

// ยอดตามบัญชีใช้ด้านปกติ/กลับเครื่องหมายของบรรทัดยอดต้นงวดเหมือนตอนคำนวณงบ
func TestStatementCashChecksSignRules(t *testing.T) {
	template := cashCheckTemplate("cash_flow", "R20 + R30")
	template.Rows[2].NormalBalance, template.Rows[2].ReverseSign = "credit", true
	checks, _ := runCashChecks(t, template, cashCheckBalances("150"), ReportPeriod{Key: "amount", FiscalYear: "2027"})
	if len(checks) != 1 || !checks[0].Matched || checks[0].Book != "150.00" {
		t.Fatalf("signed checks = %+v", checks)
	}
	// ปัดผลต่างตามทศนิยมของรูปแบบ: 0.004 ถือว่าตรง, 0.005 ปัดเป็น 0.01 ไม่ตรง
	for closing, matched := range map[string]bool{"150.004": true, "149.996": true, "150.005": false} {
		checks, _ = runCashChecks(t, cashCheckTemplate("cash_flow", "R20 + R30"), cashCheckBalances(closing), ReportPeriod{Key: "amount", FiscalYear: "2027"})
		if checks[0].Matched != matched {
			t.Fatalf("closing %s matched = %v", closing, checks[0].Matched)
		}
	}
}

func TestStatementCashChecksNotApplicable(t *testing.T) {
	// งบอื่นที่ไม่ใช่งบกระแสเงินสด ไม่มีผลตรวจ (เช่น งบต้นทุนการผลิตก็มีแถวยอดต้นงวด)
	for _, statementType := range []string{"pnl", "production_cost", "balance_sheet", "custom", ""} {
		if endings, warnings := statementCashEndings(cashCheckTemplate(statementType, "R20 + R30")); len(endings) != 0 || len(warnings) != 0 {
			t.Fatalf("%q endings = %+v %v", statementType, endings, warnings)
		}
	}
	// สูตรที่ไม่อ้างแถวยอดต้นงวดโดยตรง (อ้างผ่านแถวสูตรอื่น หรืออ้างตัวเอง) ไม่ตรวจ
	for _, formula := range []string{"R20", "R40", "", "R20 R30"} {
		if endings, _ := statementCashEndings(cashCheckTemplate("cash_flow", formula)); len(endings) != 0 {
			t.Fatalf("%q endings = %+v", formula, endings)
		}
	}
	// ไม่มีแถวยอดต้นงวดเลย
	template := cashCheckTemplate("cash_flow", "R20 + R30")
	template.Rows[2].AmountBasis = ""
	if endings, warnings := statementCashEndings(template); len(endings) != 0 || len(warnings) != 0 {
		t.Fatalf("no opening rows = %+v %v", endings, warnings)
	}
}

// ยังไม่ได้เลือกบัญชีเงินสดในบรรทัดยอดต้นงวด: ตรวจกับบัญชีไม่ได้ → คำเตือนแทนผลตรวจ (ไม่แสดง "ตรง/ไม่ตรง" กับยอดศูนย์)
func TestStatementCashChecksOpeningWithoutAccounts(t *testing.T) {
	template := cashCheckTemplate("cash_flow", "R20 + R30")
	template.Rows[2].AccountCodes = nil
	endings, warnings := statementCashEndings(template)
	if len(endings) != 0 || len(warnings) != 1 || warnings[0] != "บรรทัด 40 เงินสดปลายงวด: ยังไม่ได้เลือกบัญชีเงินสดในบรรทัดยอดต้นงวดที่สูตรอ้างถึง จึงยังตรวจเงินสดปลายงวดกับยอดคงเหลือตามบัญชีไม่ได้" {
		t.Fatalf("no accounts = %+v %q", endings, warnings)
	}
}

func TestStatementFormulaRefsFollowEvaluation(t *testing.T) {
	values := map[int]decimal.Decimal{10: dec("1"), 20: dec("2"), 30: dec("3")}
	cases := map[string]string{
		"R10 + R30":       "10,30",
		"SUM(R10:R20)":    "10,20",
		"SUM(R10:R50)":    "10,20,30", // ไม่นับแถวตัวเอง (50) และแถวที่ไม่มีค่า
		"R99 + R10":       "10",
		"R50 + R20":       "20",
		"R10 R30":         "10", // ขาดเครื่องหมาย: R30 ไม่ถูกคำนวณ จึงไม่นับเป็นการอ้าง
		"(R10 - R20) * 2": "10,20",
	}
	for formula, want := range cases {
		refs := map[int]bool{}
		statementFormula(formula, values, 50, refs)
		got := []string{}
		for _, rowNo := range []int{10, 20, 30, 50, 99} {
			if refs[rowNo] {
				got = append(got, strconv.Itoa(rowNo))
			}
		}
		if strings.Join(got, ",") != want {
			t.Errorf("%q refs = %v, want %s", formula, got, want)
		}
	}
}

func cashEndingRows(endings []statementCashEnding) string {
	out := []string{}
	for _, ending := range endings {
		opening := []string{}
		for _, row := range ending.opening {
			opening = append(opening, strconv.Itoa(row.RowNo))
		}
		out = append(out, strconv.Itoa(ending.row.RowNo)+"<"+strings.Join(opening, "+"))
	}
	return strings.Join(out, ",")
}

// ผลต่างของยอดคงเหลือในส่วนเงินทุนหมุนเวียน (ต้นงวด − ปลายงวด แบบเดียวกับงบต้นทุนขาย) และบรรทัด "เงินสดปลายงวด − ต้นงวด"
// ไม่ใช่เงินสดปลายงวด: ตรวจเฉพาะบรรทัดที่รวมเงินสดต้นงวดกับเงินสดเปลี่ยนแปลงสุทธิ
func TestStatementCashChecksIgnoreBalanceChangeRows(t *testing.T) {
	template := Master{StatementType: "cash_flow", Rows: []StatementRow{
		{RowNo: 10, RowType: "header", Title: "กระแสเงินสดจากกิจกรรมดำเนินงาน"},
		{RowNo: 20, RowType: "account", Title: "กำไร (ขาดทุน) สุทธิ", AccountCodes: []string{statementCurrentEarnings}, NormalBalance: "credit"},
		{RowNo: 41, RowType: "account", Title: "บวก ลูกหนี้การค้าต้นงวด", AccountCodes: []string{"1100"}, NormalBalance: "debit", AmountBasis: "opening"},
		{RowNo: 42, RowType: "account", Title: "หัก ลูกหนี้การค้าปลายงวด", AccountCodes: []string{"1100"}, NormalBalance: "debit", AmountBasis: "closing", ReverseSign: true},
		{RowNo: 51, RowType: "account", Title: "สินค้าคงเหลือต้นงวด", AccountCodes: []string{"1200"}, NormalBalance: "debit", AmountBasis: "opening"},
		{RowNo: 52, RowType: "account", Title: "สินค้าคงเหลือปลายงวด", AccountCodes: []string{"1200"}, NormalBalance: "debit", AmountBasis: "closing"},
		{RowNo: 55, RowType: "formula", Title: "สินค้าคงเหลือลดลง (เพิ่มขึ้น)", Formula: "R51 - R52"},
		{RowNo: 60, RowType: "subtotal", Title: "กำไรและการเปลี่ยนแปลงในลูกหนี้", Formula: "SUM(R20:R42)"},
		{RowNo: 70, RowType: "subtotal", Title: "เงินสดสุทธิจากกิจกรรมดำเนินงาน", Formula: "R60 + R55"},
		{RowNo: 160, RowType: "formula", Title: "เงินสดเพิ่มขึ้น (ลดลง) สุทธิ", Formula: "R70"},
		{RowNo: 170, RowType: "account", Title: "เงินสด ณ วันต้นงวด", AccountCodes: []string{"1000"}, NormalBalance: "debit", AmountBasis: "opening"},
		{RowNo: 180, RowType: "subtotal", Title: "เงินสด ณ วันปลายงวด", Formula: "R160 + R170"},
		{RowNo: 185, RowType: "account", Title: "เงินสดปลายงวดตามบัญชี", AccountCodes: []string{"1000"}, NormalBalance: "debit", AmountBasis: "closing"},
		{RowNo: 190, RowType: "formula", Title: "เงินสดเปลี่ยนแปลงตามบัญชี", Formula: "R185 - R170"},
	}}
	endings, warnings := statementCashEndings(template)
	if got := cashEndingRows(endings); got != "180<170" || len(warnings) != 0 {
		t.Fatalf("endings = %s warnings = %q", got, warnings)
	}
	// กำไร 150, ลูกหนี้ 500 → 300 (ลดลง 200), สินค้า 50 → 30 (ลดลง 20): เงินสดจากการดำเนินงาน 370, เงินสด 100 → 470
	balances := map[string]statementBalance{
		"1000":                   {accountType: "asset", opening: dec("100"), movement: dec("370"), balance: dec("470")},
		"1100":                   {accountType: "asset", opening: dec("500"), movement: dec("-200"), balance: dec("300")},
		"1200":                   {accountType: "asset", opening: dec("50"), movement: dec("-20"), balance: dec("30")},
		statementCurrentEarnings: {accountType: "equity", movement: dec("-150"), balance: dec("-150")},
	}
	checks, warnings := runCashChecks(t, template, balances, ReportPeriod{Key: "amount", FiscalYear: "2027"})
	want := ReportCheck{Key: "amount", FiscalYear: "2027", RowNo: 180, Title: "เงินสด ณ วันปลายงวด", Statement: "470.00", Book: "470.00", Difference: "0.00", Matched: true}
	if len(checks) != 1 || checks[0] != want || len(warnings) != 0 {
		t.Fatalf("checks = %+v warnings = %q", checks, warnings)
	}
}

// แถวที่รวมเฉพาะยอดต้นงวด (รวมเงินสดต้นงวดหลายบัญชี หรือยกยอดต้นงวดมาแสดงซ้ำ) = ยอดต้นงวดรวม ไม่ใช่เงินสดปลายงวด;
// บรรทัดที่รวมยอดต้นงวดรวมกับเงินสดเปลี่ยนแปลงสุทธิเป็นบรรทัดที่ตรวจ และยอดตามบัญชีรวมทุกบัญชีเงินสดที่อยู่ใต้ยอดรวมนั้น
func TestStatementCashChecksOpeningAggregates(t *testing.T) {
	template := Master{StatementType: "cash_flow", Rows: []StatementRow{
		{RowNo: 10, RowType: "account", Title: "เงินสดเพิ่มขึ้นสุทธิ", AccountCodes: []string{"4000"}, NormalBalance: "credit", AmountBasis: "movement"},
		{RowNo: 20, RowType: "account", Title: "เงินสดในมือต้นงวด", AccountCodes: []string{"1000"}, NormalBalance: "debit", AmountBasis: "opening"},
		{RowNo: 30, RowType: "account", Title: "เงินฝากธนาคารต้นงวด", AccountCodes: []string{"1010"}, NormalBalance: "debit", AmountBasis: "opening"},
		{RowNo: 40, RowType: "subtotal", Title: "รวมเงินสดต้นงวด", Formula: "SUM(R20:R30)"},
		{RowNo: 45, RowType: "formula", Title: "เงินสด ณ วันต้นงวด", Formula: "R40"},
		{RowNo: 50, RowType: "subtotal", Title: "เงินสดปลายงวด", Formula: "R10 + R45"},
	}}
	endings, warnings := statementCashEndings(template)
	if got := cashEndingRows(endings); got != "50<20+30" || len(warnings) != 0 {
		t.Fatalf("endings = %s warnings = %q", got, warnings)
	}
	balances := map[string]statementBalance{
		"1000": {accountType: "asset", opening: dec("100"), movement: dec("50"), balance: dec("150")},
		"1010": {accountType: "asset", opening: dec("200"), movement: dec("100"), balance: dec("300")},
		"4000": {accountType: "income", movement: dec("-150"), balance: dec("-150")},
	}
	checks, _ := runCashChecks(t, template, balances, ReportPeriod{Key: "amount", FiscalYear: "2027"})
	if len(checks) != 1 || checks[0].RowNo != 50 || checks[0].Statement != "450.00" || checks[0].Book != "450.00" || !checks[0].Matched {
		t.Fatalf("checks = %+v", checks)
	}
	// บรรทัดยอดต้นงวดรวมที่ยังไม่เลือกบัญชีเลย: เตือนที่บรรทัดยอดปลายงวดครั้งเดียว
	template.Rows[1].AccountCodes, template.Rows[2].AccountCodes = nil, nil
	if endings, warnings = statementCashEndings(template); len(endings) != 0 || len(warnings) != 1 || !strings.HasPrefix(warnings[0], "บรรทัด 50 เงินสดปลายงวด: ยังไม่ได้เลือกบัญชีเงินสด") {
		t.Fatalf("no accounts = %s %q", cashEndingRows(endings), warnings)
	}
}

// เงินสดและรายการเทียบเท่าเงินสดสุทธิจากเงินเบิกเกินบัญชี (TAS 7): ยอดตามบัญชีต้องคำนวณด้วยสูตรของผู้ใช้เอง (เงินสด − เงินเบิกเกินบัญชี)
// ไม่ใช่บวกทุกบรรทัดยอดต้นงวด — เดิมได้ 80 + 10 = 90 ทั้งที่งบถูกต้อง 70
func TestStatementCashChecksKeepFormulaSigns(t *testing.T) {
	rows := func(opening, ending string) []StatementRow {
		return []StatementRow{
			{RowNo: 20, RowType: "account", Title: "ลูกหนี้การค้าลดลง", AccountCodes: []string{"1100"}, NormalBalance: "credit", AmountBasis: "movement"},
			{RowNo: 160, RowType: "formula", Title: "เงินสดและรายการเทียบเท่าเงินสดเพิ่มขึ้นสุทธิ", Formula: "R20"},
			{RowNo: 171, RowType: "account", Title: "เงินสดและเงินฝากธนาคารต้นงวด", AccountCodes: []string{"1000"}, NormalBalance: "debit", AmountBasis: "opening"},
			{RowNo: 172, RowType: "account", Title: "หัก เงินเบิกเกินบัญชีต้นงวด", AccountCodes: []string{"2100"}, NormalBalance: "credit", AmountBasis: "opening"},
			{RowNo: 175, RowType: "formula", Title: "เงินสดและรายการเทียบเท่าเงินสดต้นงวด", Formula: opening},
			{RowNo: 180, RowType: "subtotal", Title: "เงินสดและรายการเทียบเท่าเงินสดปลายงวด", Formula: ending},
		}
	}
	// รับชำระหนี้ 50: เข้าเงินฝาก 30 และชำระเงินเบิกเกินบัญชี 20 → เงินสด 50 → 80, เงินเบิกเกินบัญชี 30 → 10, สุทธิ 20 → 70
	balances := map[string]statementBalance{
		"1000": {accountType: "asset", opening: dec("50"), movement: dec("30"), balance: dec("80")},
		"2100": {accountType: "liability", opening: dec("-30"), movement: dec("20"), balance: dec("-10")},
		"1100": {accountType: "asset", opening: dec("50"), movement: dec("-50"), balance: dec("0")},
	}
	for _, c := range []struct{ opening, ending string }{{"R171 - R172", "R160 + R175"}, {"R171", "R160 + R171 - R172"}} {
		template := Master{StatementType: "cash_flow", Rows: rows(c.opening, c.ending)}
		endings, warnings := statementCashEndings(template)
		if got := cashEndingRows(endings); got != "180<171+172" || len(warnings) != 0 {
			t.Fatalf("%s: endings = %s warnings = %q", c.ending, got, warnings)
		}
		checks, warnings := runCashChecks(t, template, balances, ReportPeriod{Key: "amount", FiscalYear: "2027"})
		want := ReportCheck{Key: "amount", FiscalYear: "2027", RowNo: 180, Title: "เงินสดและรายการเทียบเท่าเงินสดปลายงวด", Statement: "70.00", Book: "70.00", Difference: "0.00", Matched: true}
		if len(checks) != 1 || checks[0] != want || len(warnings) != 0 {
			t.Fatalf("%s: checks = %+v warnings = %q", c.ending, checks, warnings)
		}
	}
}

// ผู้ผลิต (วิธีอ้อม): สินค้าคงเหลือปลายงวดรวมจากวัตถุดิบ + สินค้าสำเร็จรูปในแถวสูตร แล้วผลต่าง (ต้นงวด − ปลายงวดรวม) อยู่ในเงินทุนหมุนเวียน
// แถวผลต่างนั้นไม่ใช่เงินสดปลายงวด — เดิมถูกตรวจและขึ้น "ไม่ตรงกัน" ผลต่าง -260.00
func TestStatementCashChecksIgnoreChangesFromClosingAggregates(t *testing.T) {
	template := Master{StatementType: "cash_flow", Rows: []StatementRow{
		{RowNo: 20, RowType: "account", Title: "กำไร (ขาดทุน) สุทธิ", AccountCodes: []string{statementCurrentEarnings}, NormalBalance: "credit"},
		{RowNo: 51, RowType: "account", Title: "สินค้าคงเหลือต้นงวด", AccountCodes: []string{"1200", "1210"}, NormalBalance: "debit", AmountBasis: "opening"},
		{RowNo: 52, RowType: "account", Title: "วัตถุดิบปลายงวด", AccountCodes: []string{"1200"}, NormalBalance: "debit", AmountBasis: "closing"},
		{RowNo: 53, RowType: "account", Title: "สินค้าสำเร็จรูปปลายงวด", AccountCodes: []string{"1210"}, NormalBalance: "debit", AmountBasis: "closing"},
		{RowNo: 55, RowType: "subtotal", Title: "สินค้าคงเหลือปลายงวด", Formula: "R52 + R53"},
		{RowNo: 56, RowType: "formula", Title: "สินค้าคงเหลือ (เพิ่มขึ้น) ลดลง", Formula: "R51 - R55"},
		{RowNo: 70, RowType: "subtotal", Title: "เงินสดสุทธิจากกิจกรรมดำเนินงาน", Formula: "R20 + R56"},
		{RowNo: 160, RowType: "formula", Title: "เงินสดเพิ่มขึ้น (ลดลง) สุทธิ", Formula: "R70"},
		{RowNo: 170, RowType: "account", Title: "เงินสด ณ วันต้นงวด", AccountCodes: []string{"1000"}, NormalBalance: "debit", AmountBasis: "opening"},
		{RowNo: 180, RowType: "subtotal", Title: "เงินสด ณ วันปลายงวด", Formula: "R160 + R170"},
		// เงินสดเปลี่ยนแปลงตามบัญชี = (ยอดปลายงวดรวม) − ต้นงวด: ผลต่างอีกแบบที่อ่านผ่านยอดปลายงวดรวม
		{RowNo: 185, RowType: "account", Title: "เงินสดในมือปลายงวด", AccountCodes: []string{"1000"}, NormalBalance: "debit", AmountBasis: "closing"},
		{RowNo: 186, RowType: "subtotal", Title: "เงินสดปลายงวดตามบัญชี", Formula: "R185"},
		{RowNo: 190, RowType: "formula", Title: "เงินสดเปลี่ยนแปลงตามบัญชี", Formula: "R186 - R170"},
	}}
	endings, warnings := statementCashEndings(template)
	if got := cashEndingRows(endings); got != "180<170" || len(warnings) != 0 {
		t.Fatalf("endings = %s warnings = %q", got, warnings)
	}
	// กำไร 200, สินค้า 300 → 280 (ลดลง 20): เงินสดจากการดำเนินงาน 220, เงินสด 340 → 560
	balances := map[string]statementBalance{
		"1000":                   {accountType: "asset", opening: dec("340"), movement: dec("220"), balance: dec("560")},
		"1200":                   {accountType: "asset", opening: dec("100"), movement: dec("-30"), balance: dec("70")},
		"1210":                   {accountType: "asset", opening: dec("200"), movement: dec("10"), balance: dec("210")},
		statementCurrentEarnings: {accountType: "equity", movement: dec("-200"), balance: dec("-200")},
	}
	checks, warnings := runCashChecks(t, template, balances, ReportPeriod{Key: "amount", FiscalYear: "2027"})
	want := ReportCheck{Key: "amount", FiscalYear: "2027", RowNo: 180, Title: "เงินสด ณ วันปลายงวด", Statement: "560.00", Book: "560.00", Difference: "0.00", Matched: true}
	if len(checks) != 1 || checks[0] != want || len(warnings) != 0 {
		t.Fatalf("checks = %+v warnings = %q", checks, warnings)
	}
}

// ตัวเลขในข้อความเตือนมีจุลภาคหลักพัน อ่านง่ายเหมือนการ์ดตรวจยอด; ค่าติดลบ/ศูนย์/ทศนิยม 0 ตำแหน่ง
func TestStatementWarningAmount(t *testing.T) {
	for _, c := range []struct {
		amount string
		scale  int32
		want   string
	}{
		{"1888400", 2, "1,888,400.00"}, {"-95280", 2, "-95,280.00"}, {"69.75", 2, "69.75"}, {"0", 2, "0.00"},
		{"999", 0, "999"}, {"1000", 0, "1,000"}, {"-1234567.891", 2, "-1,234,567.89"}, {"100000", 2, "100,000.00"},
	} {
		if got := statementWarningAmount(dec(c.amount), c.scale); got != c.want {
			t.Errorf("statementWarningAmount(%s, %d) = %s, want %s", c.amount, c.scale, got, c.want)
		}
	}
}
