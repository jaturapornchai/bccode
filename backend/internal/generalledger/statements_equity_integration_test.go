//go:build integration

package generalledger

import (
	"strings"
	"testing"
)

// งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้นบน PostgreSQL จริง: ปี 2026 ปิดบัญชีและยกยอดไป 2027 แล้ว, ปี 2027 เพิ่มทุน จ่ายปันผล
// มีกำไรที่ยังไม่ปิด และมีรายการกำไรสะสมที่ไม่ได้จัดบรรทัด → ต้นงวด + ทุกบรรทัด = ปลายงวด และรวมตรงกับเงินสด (สินทรัพย์เดียว)
func TestEquityStatementIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	dividend := Account{AccountCode: "3300", AccountType: "equity", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "เงินปันผลจ่าย"}}}
	f.run(Command{Resource: "accounts", Action: "create", Account: &dividend})
	post := func(doc, date, year, debitAccount, creditAccount, amount string) {
		j := Journal{DocNo: doc, Date: date, BookCode: "JV", FiscalYear: year, Description: "รายการทดสอบงบส่วนของผู้ถือหุ้น", Kind: "manual", BranchCode: "B1",
			Lines: []Line{{AccountCode: debitAccount, Debit: Amount(amount), Credit: Amount("0")}, {AccountCode: creditAccount, Debit: Amount("0"), Credit: Amount(amount)}}}
		f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID))
	}
	post("JV26-0", "2026-01-05", "2026", "1000", "3000", "1000")
	post("JV26-1", "2026-01-10", "2026", "1000", "4000", "100")
	post("JV26-2", "2026-07-01", "2026", "5000", "1000", "30.25")
	year := f.years["2026"]
	f.post(f.journal(f.run(Command{Resource: "processes", Action: "close", ID: "2026", Version: year.Version, DocNo: "CLOSE26", Date: "2026-12-31", Reason: "ปิดบัญชีสิ้นปี"}).ID))
	opening := f.journal(f.run(Command{Resource: "processes", Action: "year-end", ID: "2026", Version: year.Version, TargetYear: "2027", DocNo: "OPEN27", Date: "2027-01-01", Reason: "ยกยอดไปปีถัดไป"}).ID)
	if opening.Status == "draft" {
		f.post(opening)
	}
	post("JV27-1", "2027-02-01", "2027", "1000", "3000", "500")
	post("JV27-2", "2027-03-10", "2027", "1000", "4000", "200")
	post("JV27-3", "2027-04-01", "2027", "3300", "1000", "20")
	post("JV27-4", "2027-06-10", "2027", "5000", "1000", "50")
	post("JV27-5", "2027-08-01", "2027", "1000", "3100", "5")

	equity := Master{Kind: "statement-templates", Code: "EQ", Name: "งบการเปลี่ยนแปลงส่วนของผู้ถือหุ้น", StatementType: "equity", IsActive: true,
		GlobalStyle: &StatementGlobalStyle{Scale: 2, ComparisonType: "previous_year"},
		Columns: []StatementColumn{
			{ID: "cap", Title: "ทุนที่ชำระแล้ว", AccountCodes: []string{"3000"}},
			{ID: "premium", Title: "ส่วนเกินมูลค่าหุ้น"},
			{ID: "re", Title: "กำไร (ขาดทุน) สะสม", AccountCodes: []string{"3100", "3200", "3300", statementCurrentEarnings}},
		},
		Rows: []StatementRow{
			{ID: "a", RowNo: 10, RowType: "account", Title: "ยอดคงเหลือ ณ ต้นงวด {year}", AmountBasis: "opening"},
			{ID: "b", RowNo: 20, RowType: "account", Title: "การเพิ่ม (ลด) หุ้นสามัญ", AccountCodes: []string{"3000"}},
			{ID: "c", RowNo: 30, RowType: "account", Title: "เงินปันผลจ่าย", AccountCodes: []string{"3300"}},
			{ID: "d", RowNo: 40, RowType: "account", Title: "กำไร (ขาดทุน) สุทธิ", AccountCodes: []string{statementCurrentEarnings}},
			{ID: "e", RowNo: 50, RowType: "account", Title: "รายการอื่นที่ยังไม่ได้จัดประเภท", AmountBasis: "other"},
			{ID: "f", RowNo: 60, RowType: "account", Title: "ยอดคงเหลือ ณ ปลายงวด {year}", AmountBasis: "closing"},
		}}
	unmapped := equity
	unmapped.Code, unmapped.GlobalStyle = "EQ-NOCE", &StatementGlobalStyle{Scale: 2, ComparisonType: "none"}
	unmapped.Columns = []StatementColumn{{ID: "re", Title: "กำไร (ขาดทุน) สะสม", AccountCodes: []string{"3100"}}}
	empty := equity
	empty.Code, empty.Columns = "EQ-EMPTY", []StatementColumn{{ID: "cap", Title: "ทุนที่ชำระแล้ว"}}
	for _, m := range []Master{equity, unmapped, empty} {
		m := m
		f.run(Command{Resource: "statement-templates", Action: "create", Master: &m})
	}

	report, err := f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2027", Template: "EQ"})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Warnings) != 0 || len(report.Periods) != 2 || report.Periods[0].FiscalYear != "2027" || report.Periods[1].FiscalYear != "2026" {
		t.Fatalf("warnings/periods = %v %+v", report.Warnings, report.Periods)
	}
	// คอลัมน์ที่ยังไม่เลือกบัญชี (ส่วนเกินมูลค่าหุ้น) ไม่แสดง: ลำดับ รายการ หมายเหตุ ทุน กำไรสะสม รวม
	if len(report.Columns) != 6 || report.Columns[3].Label != "ทุนที่ชำระแล้ว" || report.Columns[4].Key != "c2" || report.Columns[5].Key != "total" {
		t.Fatalf("columns = %+v", report.Columns)
	}
	want := map[string][3]string{ // block|rowno → ทุน, กำไรสะสม, รวม
		"prioramount|10": {"0", "0", "0"}, "prioramount|20": {"1000", "0", "1000"}, "prioramount|40": {"0", "69.75", "69.75"},
		"prioramount|50": {"0", "0", "0"}, "prioramount|60": {"1000", "69.75", "1069.75"},
		"amount|10": {"1000", "69.75", "1069.75"}, "amount|20": {"500", "0", "500"}, "amount|30": {"0", "-20", "-20"},
		"amount|40": {"0", "150", "150"}, "amount|50": {"0", "5", "5"}, "amount|60": {"1500", "204.75", "1704.75"},
	}
	if len(report.Rows) != 12 || report.Rows[0]["block"] != "prioramount" || report.Rows[0]["title"] != "ยอดคงเหลือ ณ ต้นงวด 2026" || report.Rows[11]["title"] != "ยอดคงเหลือ ณ ปลายงวด 2027" {
		t.Fatalf("rows = %v", report.Rows)
	}
	for _, row := range report.Rows {
		w, ok := want[row["block"]+"|"+row["rowno"]]
		if !ok {
			continue
		}
		if !sameAmount(row["c1"], w[0]) || !sameAmount(row["c2"], w[1]) || !sameAmount(row["total"], w[2]) {
			t.Errorf("%s row %s = %s / %s / %s, want %v", row["block"], row["rowno"], row["c1"], row["c2"], row["total"], w)
		}
	}
	// รวมส่วนของผู้ถือหุ้นปลายปีต้องเท่ากับสินทรัพย์ (เงินสด) ในงบแสดงฐานะการเงิน
	var cash string
	if err = f.db.QueryRowContext(f.ctx, `SELECT SUM(debit-credit)::text FROM gl_lines WHERE company='C' AND fiscal_year='2027' AND account_code='1000'`).Scan(&cash); err != nil || !sameAmount(cash, "1704.75") {
		t.Fatalf("cash = %s %v", cash, err)
	}

	// ไม่ได้ใส่กำไรที่ยังไม่ปิด: ยังได้ผล แต่ต้องเตือน (ไม่เงียบ)
	bad, err := f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2026", Template: "EQ-NOCE"})
	if err != nil || len(bad.Periods) != 1 {
		t.Fatalf("EQ-NOCE = %+v %v", bad, err)
	}
	joined := strings.Join(bad.Warnings, "\n")
	for _, text := range []string{"ยังไม่ได้เลือก “รวมกำไร (ขาดทุน) ที่ยังไม่ปิดบัญชี” ในคอลัมน์ใด", "ปี 2026: คอลัมน์ กำไร (ขาดทุน) สะสม มีรายการปิดบัญชีที่ไม่หักล้างกัน 69.75"} {
		if !strings.Contains(joined, text) {
			t.Errorf("missing warning %q in\n%s", text, joined)
		}
	}
	if _, err = f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2027", Template: "EQ-EMPTY"}); err == nil {
		t.Fatal("equity statement without mapped columns accepted")
	}
}
