//go:build integration

package generalledger

import (
	"testing"
)

// งบการเงินตามรูปแบบ (backend) + คอลัมน์ปีก่อน บน PostgreSQL จริง: ปี 2026 ปิดบัญชีและยกยอดไป 2027 แล้ว
// งบกำไรขาดทุนปีที่ปิดแล้วต้องไม่เป็นศูนย์ (ไม่นับรายการปิดบัญชี), งบฐานะการเงินรวมกำไรที่ยังไม่ปิด (__current_earnings__)
func TestStatementReportComparativeIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	post := func(doc, date, year, debitAccount, creditAccount, amount string) {
		j := Journal{DocNo: doc, Date: date, BookCode: "JV", FiscalYear: year, Description: "รายการทดสอบงบการเงิน", Kind: "manual", BranchCode: "B1",
			Lines: []Line{{AccountCode: debitAccount, Debit: Amount(amount), Credit: Amount("0")}, {AccountCode: creditAccount, Debit: Amount("0"), Credit: Amount(amount)}}}
		f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID))
	}
	post("JV26-1", "2026-01-10", "2026", "1000", "4000", "100")
	post("JV26-2", "2026-07-01", "2026", "5000", "1000", "30.25")
	year := f.years["2026"]
	closing := f.journal(f.run(Command{Resource: "processes", Action: "close", ID: "2026", Version: year.Version, DocNo: "CLOSE26", Date: "2026-12-31", Reason: "ปิดบัญชีสิ้นปี"}).ID)
	f.post(closing)
	opening := f.journal(f.run(Command{Resource: "processes", Action: "year-end", ID: "2026", Version: year.Version, TargetYear: "2027", DocNo: "OPEN27", Date: "2027-01-01", Reason: "ยกยอดไปปีถัดไป"}).ID)
	if opening.Status == "draft" {
		f.post(opening)
	}
	post("JV27-1", "2027-03-10", "2027", "1000", "4000", "200")
	post("JV27-2", "2027-06-10", "2027", "5000", "1000", "50")

	position := Master{Kind: "statement-templates", Code: "FS-POS", Name: "งบแสดงฐานะการเงิน", StatementType: "balance_sheet", IsActive: true,
		GlobalStyle: &StatementGlobalStyle{Scale: 2, ComparisonType: "previous_year"},
		Rows: []StatementRow{
			{ID: "a", RowNo: 10, RowType: "account", Title: "สินทรัพย์", AccountCodes: []string{"1000"}},
			{ID: "b", RowNo: 20, RowType: "account", Title: "กำไรสะสม", AccountCodes: []string{"3100", "3200"}},
			{ID: "c", RowNo: 30, RowType: "account", Title: "กำไร (ขาดทุน) สุทธิประจำงวด", AccountCodes: []string{statementCurrentEarnings}},
			{ID: "d", RowNo: 40, RowType: "subtotal", Title: "รวมส่วนของผู้ถือหุ้น", Formula: "R20 + R30"},
			{ID: "e", RowNo: 50, RowType: "header", Title: "หัวข้อ"},
		}}
	income := Master{Kind: "statement-templates", Code: "FS-PL", Name: "งบกำไรขาดทุน", StatementType: "pnl", IsActive: true,
		GlobalStyle: &StatementGlobalStyle{Scale: 2, ComparisonType: "previous_year"},
		Rows: []StatementRow{
			{ID: "a", RowNo: 10, RowType: "account", Title: "รายได้", AccountCodes: []string{"4000"}},
			{ID: "b", RowNo: 20, RowType: "account", Title: "ค่าใช้จ่าย", AccountCodes: []string{"5000"}},
			{ID: "c", RowNo: 30, RowType: "formula", Title: "กำไร (ขาดทุน) สุทธิ", Formula: "R10 - R20"},
		}}
	plain := income
	plain.Code, plain.GlobalStyle = "FS-PL-ONE", &StatementGlobalStyle{Scale: 2, ComparisonType: "none"}
	for _, m := range []Master{position, income, plain} {
		m := m
		f.run(Command{Resource: "statement-templates", Action: "create", Master: &m})
	}

	check := func(template, from, to string, want map[string][2]string, wantPeriods int) Report {
		t.Helper()
		report, err := f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2027", From: from, To: to, Template: template})
		if err != nil {
			t.Fatalf("%s: %v", template, err)
		}
		if len(report.Periods) != wantPeriods {
			t.Fatalf("%s periods = %+v", template, report.Periods)
		}
		for _, row := range report.Rows {
			w, ok := want[row["rowno"]]
			if !ok {
				continue
			}
			if !sameAmount(row["amount"], w[0]) || wantPeriods == 2 && !sameAmount(row["prioramount"], w[1]) {
				t.Errorf("%s row %s = %s / %s, want %s / %s", template, row["rowno"], row["amount"], row["prioramount"], w[0], w[1])
			}
		}
		return report
	}
	// ฐานะการเงินสิ้นปี 2027: เงินสด 69.75 + 150; กำไรสะสมรับกำไรปี 2026 ที่ปิดแล้ว 69.75; กำไรปี 2027 ยังไม่ปิด 150
	pos := check("FS-POS", "", "", map[string][2]string{"10": {"219.75", "69.75"}, "20": {"69.75", "69.75"}, "30": {"150", "0"}, "40": {"219.75", "69.75"}}, 2)
	if pos.Periods[1].FiscalYear != "2026" || pos.Periods[1].From != "2026-01-01" || pos.Periods[1].To != "2026-12-31" {
		t.Fatalf("prior period = %+v", pos.Periods[1])
	}
	for _, row := range pos.Rows {
		if row["rowno"] == "50" && (row["amount"] != "" || row["rowtype"] != "header") {
			t.Fatalf("header row = %v", row)
		}
		// แสดงผลปัดตามทศนิยมของรูปแบบ (2 ตำแหน่ง) เสมอ
		if row["rowno"] == "30" && (row["amount"] != "150.00" || row["prioramount"] != "0.00") {
			t.Fatalf("fixed-scale amounts = %v", row)
		}
	}
	// กำไรขาดทุนปี 2026 ที่ปิดบัญชีแล้วยังต้องแสดงรายได้/ค่าใช้จ่ายจริง (ไม่นับรายการปิดบัญชี)
	check("FS-PL", "", "", map[string][2]string{"10": {"200", "100"}, "20": {"50", "30.25"}, "30": {"150", "69.75"}}, 2)
	// ช่วงกลางปี = เฉพาะความเคลื่อนไหวในช่วง ไม่ใช่ยอดสะสมจากต้นปี; ปีก่อนช่วงเดียวกัน (ก.ค.–ธ.ค. 2026)
	check("FS-PL", "2027-06-01", "2027-12-31", map[string][2]string{"10": {"0", "0"}, "20": {"50", "30.25"}, "30": {"-50", "-30.25"}}, 2)
	// ไม่ตั้งเปรียบเทียบ = คอลัมน์เดียว
	one := check("FS-PL-ONE", "", "", map[string][2]string{"30": {"150", ""}}, 1)
	if len(one.Columns) != 4 {
		t.Fatalf("single-period columns = %+v", one.Columns)
	}
	// ปีแรกไม่มีปีก่อน: ยังได้ผล พร้อมคำเตือน
	first, err := f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2026", Template: "FS-PL"})
	if err != nil || len(first.Periods) != 1 || len(first.Warnings) != 1 {
		t.Fatalf("first year = %+v %v", first, err)
	}
	if _, err = f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2027", Template: "NOPE"}); err == nil {
		t.Fatal("unknown template accepted")
	}
	if _, err = f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: "2027"}); err == nil {
		t.Fatal("missing template accepted")
	}
}
