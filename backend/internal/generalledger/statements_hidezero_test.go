package generalledger

import (
	"reflect"
	"strconv"
	"testing"
)

// แถวผลงบสำหรับทดสอบ: amounts = amount, prioramount (งบปกติ)
func hzRow(rowNo int, rowType string, indent int, amounts ...string) map[string]string {
	row := map[string]string{"rowno": strconv.Itoa(rowNo), "rowtype": rowType, "indent": strconv.Itoa(indent), "showzero": "false", "amount": "", "prioramount": ""}
	for i, key := range []string{"amount", "prioramount"} {
		if i < len(amounts) {
			row[key] = amounts[i]
		}
	}
	return row
}

func hzRowNos(rows []map[string]string) []string {
	out := []string{}
	for _, row := range rows {
		out = append(out, row["block"]+row["rowno"])
	}
	return out
}

// ข้อ 7 ประกาศกรมพัฒนาธุรกิจการค้า พ.ศ. 2566: แถวยอดศูนย์ทุกคอลัมน์ไม่แสดง เว้นแต่ติ๊กแสดงศูนย์ หรือปีก่อนมียอด
func TestHideZeroStatementRows(t *testing.T) {
	showZero := hzRow(60, "account", 2, "0.00", "0.00")
	showZero["showzero"] = "true"
	rows := []map[string]string{
		hzRow(10, "header", 0),
		hzRow(20, "header", 1),
		hzRow(30, "account", 2, "100.00", "0.00"), // ปีนี้มียอด
		hzRow(40, "account", 2, "0.00", "0.00"),   // ซ่อน
		hzRow(50, "account", 2, "0.00", "5.00"),   // ปีก่อนมียอด → แสดง
		showZero,
		hzRow(70, "subtotal", 1, "100.00", "5.00"),
		hzRow(80, "header", 1), // ทุกแถวยอดเงินในส่วนนี้ถูกซ่อน → ซ่อนหัวข้อ
		hzRow(90, "account", 2, "0.00", "0.00"),
		hzRow(100, "subtotal", 1, "0.00", "-0.00"),
		hzRow(110, "subtotal", 0, "100.00", "5.00"), // ย่อหน้าน้อยกว่าหัวข้อ 80 = อยู่นอกส่วนของหัวข้อนั้น
		hzRow(120, "blank", 0),
		hzRow(130, "header", 0), // หัวข้อซ้อน: ทั้งส่วนเป็นศูนย์ → ซ่อนทั้งสองระดับ
		hzRow(140, "header", 1),
		hzRow(150, "account", 2, "0.00"),
		hzRow(160, "divider", 0),
		hzRow(170, "formula", 2, "0.00", "0.00"),
		hzRow(180, "header", 0), // ไม่มีแถวยอดเงินในส่วน → คงไว้
	}
	got := hzRowNos(hideZeroStatementRows(rows, []string{"amount", "prioramount"}, nil))
	want := []string{"10", "20", "30", "50", "60", "70", "110", "120", "180"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rows = %v, want %v", got, want)
	}
	// คอลัมน์ปีก่อนไม่แสดง (ค่าว่าง) ไม่นับ; ข้อความที่ไม่ใช่ตัวเลขถือว่ามียอด (ไม่ซ่อนเงียบ)
	single := []map[string]string{hzRow(10, "account", 0, "0.00", ""), hzRow(20, "account", 0, "abc"), hzRow(30, "formula", 0, "0.01")}
	if got := hzRowNos(hideZeroStatementRows(single, []string{"amount", "prioramount"}, nil)); !reflect.DeepEqual(got, []string{"20", "30"}) {
		t.Fatalf("single column rows = %v", got)
	}
}

// บรรทัดว่าง/เส้นคั่น: ไม่ขึ้นต้น ไม่ซ้อนกัน ไม่ปิดท้าย; บรรทัดว่างในส่วนของหัวข้อไม่ตัดส่วนนั้น
func TestHideZeroStatementSpacers(t *testing.T) {
	rows := []map[string]string{
		hzRow(10, "blank", 0),
		hzRow(20, "header", 1),
		hzRow(30, "account", 2, "0.00"),
		hzRow(40, "blank", 0), // ย่อหน้า 0 แต่เป็นบรรทัดว่าง: ส่วนของหัวข้อ 20 ยังต่อไปถึงแถว 50
		hzRow(50, "account", 2, "7.00"),
		hzRow(60, "blank", 0),
		hzRow(70, "account", 0, "0.00"),
		hzRow(80, "divider", 0),
		hzRow(90, "account", 0, "1.00"),
		hzRow(100, "divider", 0),
		hzRow(110, "blank", 0),
	}
	got := hzRowNos(hideZeroStatementRows(rows, []string{"amount"}, nil))
	want := []string{"20", "40", "50", "60", "90"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rows = %v, want %v", got, want)
	}
	if got := hideZeroStatementRows(nil, []string{"amount"}, nil); len(got) != 0 {
		t.Fatalf("empty = %v", got)
	}
}

// งบส่วนของผู้ถือหุ้น: กรองทีละชุดปี (หัวข้อไม่ลามไปชุดถัดไป) และตัดคอลัมน์องค์ประกอบที่เป็นศูนย์ทั้งงบ คงคอลัมน์รวม
func TestHideZeroEquityBlocksAndColumns(t *testing.T) {
	eq := func(block string, rowNo int, rowType string, indent int, c1, c2, c3 string) map[string]string {
		row := map[string]string{"block": block, "rowno": strconv.Itoa(rowNo), "rowtype": rowType, "indent": strconv.Itoa(indent), "showzero": "false", "c1": c1, "c2": c2, "c3": c3, "total": ""}
		if statementAmountRow(rowType) {
			row["total"] = dec(c1).Add(dec(c2)).Add(dec(c3)).StringFixed(2)
		}
		return row
	}
	rows := []map[string]string{
		eq("prioramount", 10, "account", 0, "0.00", "0.00", "0.00"),
		eq("prioramount", 50, "header", 0, "", "", ""),
		eq("prioramount", 60, "account", 1, "0.00", "0.00", "0.00"),
		eq("prioramount", 70, "account", 1, "0.00", "0.00", "0.00"),
		eq("amount", 10, "account", 0, "1000.00", "0.00", "0.00"),
		eq("amount", 50, "header", 0, "", "", ""),
		eq("amount", 60, "account", 1, "500.00", "0.00", "0.00"),
		eq("amount", 70, "account", 1, "0.00", "0.00", "-3.00"),
	}
	columns := []ReportColumn{textColumn("rowno", "ลำดับ"), textColumn("title", "รายการ"), textColumn("noteno", "หมายเหตุ"),
		amountColumn("c1", "ทุน"), amountColumn("c2", "ส่วนเกินมูลค่าหุ้น"), amountColumn("c3", "กำไรสะสม"), amountColumn("total", "รวม")}
	kept := hideZeroStatementRows(rows, statementAmountKeys(columns), nil)
	if got, want := hzRowNos(kept), []string{"amount10", "amount50", "amount60", "amount70"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("rows = %v, want %v", got, want)
	}
	keys := func(columns []ReportColumn) []string {
		out := []string{}
		for _, column := range columns {
			out = append(out, column.Key)
		}
		return out
	}
	if got := keys(hideZeroEquityColumns(columns, kept)); !reflect.DeepEqual(got, []string{"rowno", "title", "noteno", "c1", "c3", "total"}) {
		t.Fatalf("columns = %v", got)
	}
	// ทุกคอลัมน์เป็นศูนย์: เหลือคอลัมน์รวมเสมอ
	if got := keys(hideZeroEquityColumns(columns, rows[:4])); !reflect.DeepEqual(got, []string{"rowno", "title", "noteno", "total"}) {
		t.Fatalf("all-zero columns = %v", got)
	}
	if !reflect.DeepEqual(statementAmountKeys(columns), []string{"c1", "c2", "c3", "total"}) {
		t.Fatalf("amount keys = %v", statementAmountKeys(columns))
	}
}

// ยอดรวม/สูตรที่เป็นศูนย์แต่รวมจากแถวที่ยังแสดง ต้องแสดง (กู้ยืมแล้วคืนในปีเดียวกัน) — ไม่งั้นงบที่พิมพ์มีรายการแต่ไม่มียอดรวม;
// ยอดรวมที่แถวที่อ้างถูกซ่อนหมดยังซ่อนตามเดิม และยอดรวมที่กลับมาแสดงทำให้ยอดรวมชั้นที่อ้างมันแสดงด้วย (ไม่ขึ้นกับลำดับแถว)
func TestHideZeroKeepsTotalsOfVisibleRows(t *testing.T) {
	rows := []map[string]string{
		hzRow(5, "formula", 0, "0.00"), // อ้างแถว 150 ข้างล่าง: แสดงเมื่อ 150 กลับมาแสดง
		hzRow(130, "header", 0),
		hzRow(140, "account", 1, "1000000.00"),
		hzRow(145, "account", 1, "-1000000.00"),
		hzRow(150, "subtotal", 0, "0.00"),
		hzRow(160, "header", 0),
		hzRow(170, "account", 1, "0.00"),
		hzRow(180, "subtotal", 0, "0.00"), // แถวที่อ้างถูกซ่อนหมด → ซ่อน
		hzRow(190, "formula", 0, "0.00"),  // อ้างแถวที่ไม่มีในชุด → ซ่อน
	}
	refs := map[int][]int{5: {150}, 150: {140, 145}, 180: {170}, 190: {999}}
	got := hzRowNos(hideZeroStatementRows(rows, []string{"amount"}, refs))
	if want := []string{"5", "130", "140", "145", "150"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("rows = %v, want %v", got, want)
	}
	// ไม่ส่ง refs = ซ่อนยอดรวมศูนย์ทุกแถวเหมือนเดิม
	if got := hzRowNos(hideZeroStatementRows(rows, []string{"amount"}, nil)); !reflect.DeepEqual(got, []string{"130", "140", "145"}) {
		t.Fatalf("rows without refs = %v", got)
	}
	// refs จากรูปแบบงบจริง: แถวสูตรเชื่อมกันเป็นทอด ๆ ตามลำดับการคำนวณ
	template := []StatementRow{
		{RowNo: 140, RowType: "account"}, {RowNo: 145, RowType: "account"},
		{RowNo: 150, RowType: "subtotal", Formula: "SUM(R140:R149)"},
		{RowNo: 160, RowType: "formula", Formula: "R150 + R70"},
	}
	if refs := statementFormulaRefs(template); !reflect.DeepEqual(refs, map[int][]int{150: {140, 145}, 160: {150}}) {
		t.Fatalf("refs = %v", refs)
	}
}

// หัวข้อที่ไม่เหลือรายการต้องไม่ค้างบนงบ เมื่อแถวถัดมาย่อหน้าเท่าหัวข้อ (ตามแม่แบบ BS-DBD / CASH-FLOW-IND / EQ-DBD)
func TestHideZeroNestedSectionEndsAtSameIndent(t *testing.T) {
	equity := func(values map[int]string) []map[string]string {
		amount := func(rowNo int) string {
			if value, ok := values[rowNo]; ok {
				return value
			}
			return "0.00"
		}
		return []map[string]string{
			hzRow(550, "header", 1), hzRow(560, "header", 2), hzRow(570, "header", 3),
			hzRow(580, "account", 3, amount(580)), hzRow(590, "account", 2, amount(590)), hzRow(600, "account", 2, amount(600)),
			hzRow(610, "header", 2), hzRow(620, "header", 3),
			hzRow(630, "account", 4, amount(630)), hzRow(640, "account", 4, amount(640)),
			hzRow(650, "account", 3, amount(650)), hzRow(660, "account", 2, amount(660)), hzRow(670, "account", 2, amount(670)),
			hzRow(680, "subtotal", 1, amount(680)), hzRow(690, "subtotal", 0, amount(690)),
		}
	}
	cases := []struct {
		name   string
		values map[int]string
		want   []string
	}{
		// ไม่มีทุนสำรองตามกฎหมาย: "จัดสรรแล้ว" ไม่ค้างเหนือ "ยังไม่ได้จัดสรร"
		{"no legal reserve", map[int]string{580: "1000.00", 650: "250.00", 680: "1250.00", 690: "1250.00"}, []string{"550", "560", "570", "580", "610", "650", "680", "690"}},
		// กำไรสะสมเป็นศูนย์ทั้งส่วน: "กำไร (ขาดทุน) สะสม" ไม่ค้างเหนือ "ส่วนได้เสีย - ทุนอื่น"
		{"no retained earnings", map[int]string{580: "1000.00", 660: "500.00", 680: "1500.00", 690: "1500.00"}, []string{"550", "560", "570", "580", "660", "680", "690"}},
		// มีทุนสำรอง: หัวข้อทั้งสองระดับแสดง; หัวข้อ "ทุนจดทะเบียน" ที่แถวถัดไปย่อหน้าเท่ากัน (ส่วนแบบแบน) ยังแสดงคู่กับทุนที่ชำระแล้ว
		{"with legal reserve", map[int]string{580: "1000.00", 630: "100.00", 650: "150.00", 680: "1250.00", 690: "1250.00"}, []string{"550", "560", "570", "580", "610", "620", "630", "650", "680", "690"}},
	}
	for _, c := range cases {
		if got := hzRowNos(hideZeroStatementRows(equity(c.values), []string{"amount"}, nil)); !reflect.DeepEqual(got, c.want) {
			t.Fatalf("%s: rows = %v, want %v", c.name, got, c.want)
		}
	}

	// งบกระแสเงินสดที่ไม่มีกิจกรรมจัดหาเงิน: หัวข้อ 130 ไม่ค้างเหนือบรรทัดเงินสดเพิ่มขึ้นสุทธิ
	cashFlow := []map[string]string{
		hzRow(10, "header", 0), hzRow(20, "account", 1, "100.00"), hzRow(30, "account", 1, "0.00"), hzRow(70, "subtotal", 0, "100.00"),
		hzRow(80, "blank", 0), hzRow(90, "header", 0), hzRow(100, "account", 1, "-40.00"), hzRow(110, "subtotal", 0, "-40.00"),
		hzRow(120, "blank", 0), hzRow(130, "header", 0), hzRow(140, "account", 1, "0.00"), hzRow(150, "subtotal", 0, "0.00"),
		hzRow(160, "formula", 0, "60.00"), hzRow(170, "account", 0, "500.00"), hzRow(180, "subtotal", 0, "560.00"),
	}
	if got, want := hzRowNos(hideZeroStatementRows(cashFlow, []string{"amount"}, map[int][]int{150: {140}})), []string{"10", "20", "70", "80", "90", "100", "110", "120", "160", "170", "180"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("cash flow rows = %v, want %v", got, want)
	}

	// งบส่วนของผู้ถือหุ้น ปีที่ไม่มีความเคลื่อนไหว: หัวข้อ 50 ไม่ค้างเหนือยอดคงเหลือปลายงวด (ย่อหน้าเท่ากัน)
	eq := []map[string]string{
		hzRow(10, "account", 0, "1000.00"), hzRow(40, "subtotal", 0, "1000.00"), hzRow(50, "header", 0),
		hzRow(60, "account", 1, "0.00"), hzRow(130, "account", 1, "0.00"), hzRow(140, "account", 0, "1000.00"),
	}
	if got := hzRowNos(hideZeroStatementRows(eq, []string{"amount"}, nil)); !reflect.DeepEqual(got, []string{"10", "40", "140"}) {
		t.Fatalf("equity rows = %v", got)
	}
	// ส่วนแบบแบน (หัวข้อและรายการย่อหน้าเท่ากันทั้งงบ): หัวข้อยังคุมรายการจนถึงหัวข้อถัดไปเหมือนเดิม
	flat := []map[string]string{
		hzRow(10, "header", 0), hzRow(20, "account", 0, "0.00"), hzRow(30, "account", 0, "5.00"),
		hzRow(40, "header", 0), hzRow(50, "account", 0, "0.00"),
	}
	if got := hzRowNos(hideZeroStatementRows(flat, []string{"amount"}, nil)); !reflect.DeepEqual(got, []string{"10", "30"}) {
		t.Fatalf("flat rows = %v", got)
	}
}

// แม่แบบ BS-DBD ฝั่งหนี้สินและส่วนของผู้ถือหุ้น (frontend generateStarterTemplates): "รวมหนี้สิน" (540) ย่อหน้า 1 เท่า "รวมส่วนของผู้ถือหุ้น"
// ส่วนของหัวข้อ 320 จึงครอบถึงส่วนของผู้ถือหุ้น — กิจการที่ไม่มีหนี้สินต้องยังเห็นหัวข้อ 320 คู่กับยอดรวม 690
func TestHideZeroBalanceSheetWithoutLiabilities(t *testing.T) {
	side := func(payable string) []map[string]string {
		liabilities := payable
		total := dec("100000").Add(dec(payable)).StringFixed(2)
		return []map[string]string{
			hzRow(320, "header", 0), hzRow(330, "header", 1),
			hzRow(340, "account", 2, "0.00"), hzRow(350, "account", 2, payable), hzRow(440, "subtotal", 1, liabilities),
			hzRow(450, "header", 1), hzRow(460, "account", 2, "0.00"), hzRow(530, "subtotal", 1, "0.00"),
			hzRow(540, "subtotal", 1, liabilities),
			hzRow(550, "header", 1), hzRow(560, "header", 2), hzRow(570, "header", 3),
			hzRow(580, "account", 3, "100000.00"), hzRow(590, "account", 2, "0.00"), hzRow(680, "subtotal", 1, "100000.00"),
			hzRow(690, "subtotal", 0, total),
		}
	}
	refs := map[int][]int{440: {340, 350}, 530: {460}, 540: {440, 530}, 680: {580, 590}, 690: {540, 680}}
	if got, want := hzRowNos(hideZeroStatementRows(side("0.00"), []string{"amount"}, refs)), []string{"320", "550", "560", "570", "580", "680", "690"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("no liabilities rows = %v, want %v", got, want)
	}
	if got, want := hzRowNos(hideZeroStatementRows(side("500.00"), []string{"amount"}, refs)), []string{"320", "330", "350", "440", "540", "550", "560", "570", "580", "680", "690"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("with payables rows = %v, want %v", got, want)
	}
}
