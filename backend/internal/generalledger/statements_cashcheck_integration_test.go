//go:build integration

package generalledger

import (
	"encoding/json"
	"strings"
	"testing"
)

// งบกระแสเงินสด: เงินสดปลายงวดตามงบต้องตรวจกับยอดคงเหลือตามบัญชีเงินสด (บัญชี 1000) บน PostgreSQL จริง
// ปี 2026 ปิดบัญชีและยกยอดไป 2027 แล้ว: เงินสดสิ้นปี 2026 = 100 - 30.25 = 69.75, สิ้นปี 2027 = 69.75 + 200 - 50 = 219.75
func TestStatementCashFlowBookCheckIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	post := func(doc, date, year, debitAccount, creditAccount, amount string) {
		j := Journal{DocNo: doc, Date: date, BookCode: "JV", FiscalYear: year, Description: "รายการทดสอบงบกระแสเงินสด", Kind: "manual", BranchCode: "B1",
			Lines: []Line{{AccountCode: debitAccount, Debit: Amount(amount), Credit: Amount("0")}, {AccountCode: creditAccount, Debit: Amount("0"), Credit: Amount(amount)}}}
		f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID))
	}
	post("JV26-1", "2026-01-10", "2026", "1000", "4000", "100")
	post("JV26-2", "2026-07-01", "2026", "5000", "1000", "30.25")
	year := f.years["2026"]
	f.post(f.journal(f.run(Command{Resource: "processes", Action: "close", ID: "2026", Version: year.Version, DocNo: "CLOSE26", Date: "2026-12-31", Reason: "ปิดบัญชีสิ้นปี"}).ID))
	opening := f.journal(f.run(Command{Resource: "processes", Action: "year-end", ID: "2026", Version: year.Version, TargetYear: "2027", DocNo: "OPEN27", Date: "2027-01-01", Reason: "ยกยอดไปปีถัดไป"}).ID)
	if opening.Status == "draft" {
		f.post(opening)
	}
	post("JV27-1", "2027-03-10", "2027", "1000", "4000", "200")
	post("JV27-2", "2027-06-10", "2027", "5000", "1000", "50")

	cashFlow := func(code string, profitAccounts, openingAccounts []string) Master {
		return Master{Kind: "statement-templates", Code: code, Name: "งบกระแสเงินสด", StatementType: "cash_flow", IsActive: true,
			GlobalStyle: &StatementGlobalStyle{Scale: 2, ComparisonType: "previous_year", HideZeroRows: true},
			Rows: []StatementRow{
				{ID: "a", RowNo: 10, RowType: "account", Title: "กำไร (ขาดทุน) สุทธิประจำงวด", AccountCodes: profitAccounts, NormalBalance: "credit"},
				{ID: "b", RowNo: 15, RowType: "account", Title: "ปรับปรุง: ค่าเสื่อมราคา", NormalBalance: "debit"},
				{ID: "c", RowNo: 20, RowType: "formula", Title: "เงินสดเพิ่มขึ้น (ลดลง) สุทธิ", Formula: "R10"},
				{ID: "d", RowNo: 30, RowType: "account", Title: "เงินสด ณ วันต้นงวด", AccountCodes: openingAccounts, NormalBalance: "debit", AmountBasis: "opening"},
				{ID: "e", RowNo: 40, RowType: "subtotal", Title: "เงินสด ณ วันปลายงวด", Formula: "R20 + R30"},
			}}
	}
	income := Master{Kind: "statement-templates", Code: "PL-NOCHECK", Name: "งบกำไรขาดทุน", StatementType: "pnl", IsActive: true,
		GlobalStyle: &StatementGlobalStyle{Scale: 2, ComparisonType: "previous_year"},
		Rows: []StatementRow{
			{ID: "a", RowNo: 10, RowType: "account", Title: "รายได้", AccountCodes: []string{"4000"}},
			{ID: "b", RowNo: 20, RowType: "account", Title: "ยอดต้นงวด", AccountCodes: []string{"1000"}, AmountBasis: "opening"},
			{ID: "c", RowNo: 30, RowType: "formula", Title: "รวม", Formula: "R10 + R20"},
		}}
	for _, m := range []Master{
		cashFlow("CF-OK", []string{statementCurrentEarnings}, []string{"1000"}),
		cashFlow("CF-MISS", []string{"4000"}, []string{"1000"}), // ลืมค่าใช้จ่าย: กำไรเกินจริง เงินสดปลายงวดตามงบจึงไม่ตรงบัญชี
		cashFlow("CF-NOACC", []string{statementCurrentEarnings}, nil),
		income,
	} {
		m := m
		f.run(Command{Resource: "statement-templates", Action: "create", Master: &m})
	}
	report := func(template, fiscalYear, from, to string) Report {
		t.Helper()
		result, err := f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: fiscalYear, From: from, To: to, Template: template})
		if err != nil {
			t.Fatalf("%s: %v", template, err)
		}
		return result
	}
	sameChecks := func(label string, got []ReportCheck, want []ReportCheck) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s checks = %+v", label, got)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%s check %d = %+v, want %+v", label, i, got[i], want[i])
			}
		}
	}

	ok := report("CF-OK", "2027", "", "")
	sameChecks("CF-OK", ok.Checks, []ReportCheck{
		{Key: "amount", FiscalYear: "2027", RowNo: 40, Title: "เงินสด ณ วันปลายงวด", Statement: "219.75", Book: "219.75", Difference: "0.00", Matched: true},
		{Key: "prioramount", FiscalYear: "2026", RowNo: 40, Title: "เงินสด ณ วันปลายงวด", Statement: "69.75", Book: "69.75", Difference: "0.00", Matched: true},
	})
	if len(ok.Warnings) != 0 {
		t.Fatalf("CF-OK warnings = %q", ok.Warnings)
	}
	for _, row := range ok.Rows {
		if row["rowno"] == "15" {
			t.Fatalf("hidezerorows kept zero row 15: %v", row)
		}
	}

	miss := report("CF-MISS", "2027", "", "")
	sameChecks("CF-MISS", miss.Checks, []ReportCheck{
		{Key: "amount", FiscalYear: "2027", RowNo: 40, Title: "เงินสด ณ วันปลายงวด", Statement: "269.75", Book: "219.75", Difference: "50.00", Matched: false},
		{Key: "prioramount", FiscalYear: "2026", RowNo: 40, Title: "เงินสด ณ วันปลายงวด", Statement: "100.00", Book: "69.75", Difference: "30.25", Matched: false},
	})
	wantWarnings := []string{
		"เงินสดปลายงวดตามงบ (บรรทัด 40 เงินสด ณ วันปลายงวด) 269.75 ไม่ตรงกับยอดคงเหลือตามบัญชี 219.75 ผลต่าง 50.00 กรุณาตรวจว่าเลือกบัญชีครบทุกบรรทัด",
		"ปี 2026: เงินสดปลายงวดตามงบ (บรรทัด 40 เงินสด ณ วันปลายงวด) 100.00 ไม่ตรงกับยอดคงเหลือตามบัญชี 69.75 ผลต่าง 30.25 กรุณาตรวจว่าเลือกบัญชีครบทุกบรรทัด",
	}
	if strings.Join(miss.Warnings, "\n") != strings.Join(wantWarnings, "\n") {
		t.Fatalf("CF-MISS warnings = %q", miss.Warnings)
	}

	// ช่วงที่ทุกบรรทัดเป็นศูนย์: hidezerorows ซ่อนทุกแถว แต่ผลตรวจยังคำนวณจากค่าเต็ม (ปีแรกไม่มีปีก่อน = งวดเดียว)
	empty := report("CF-OK", "2026", "2026-01-01", "2026-01-05")
	if len(empty.Rows) != 0 {
		t.Fatalf("zero period rows = %v", empty.Rows)
	}
	sameChecks("CF-OK zero", empty.Checks, []ReportCheck{{Key: "amount", FiscalYear: "2026", RowNo: 40, Title: "เงินสด ณ วันปลายงวด", Statement: "0.00", Book: "0.00", Difference: "0.00", Matched: true}})
	// กลางปี: เงินสดปลายงวด = ยอดคงเหลือ ณ วันสิ้นช่วง (69.75 + 200 = 269.75 ณ 31 มี.ค. 2027; ปีก่อนช่วงเดียวกัน 100)
	mid := report("CF-OK", "2027", "2027-01-01", "2027-03-31")
	sameChecks("CF-OK mid-year", mid.Checks, []ReportCheck{
		{Key: "amount", FiscalYear: "2027", RowNo: 40, Title: "เงินสด ณ วันปลายงวด", Statement: "269.75", Book: "269.75", Difference: "0.00", Matched: true},
		{Key: "prioramount", FiscalYear: "2026", RowNo: 40, Title: "เงินสด ณ วันปลายงวด", Statement: "100.00", Book: "100.00", Difference: "0.00", Matched: true},
	})

	// ยังไม่เลือกบัญชีเงินสด: ไม่มีผลตรวจ แต่มีคำเตือนบอกให้เลือก
	noAccount := report("CF-NOACC", "2027", "", "")
	if len(noAccount.Checks) != 0 || len(noAccount.Warnings) != 1 || !strings.Contains(noAccount.Warnings[0], "ยังไม่ได้เลือกบัญชีเงินสด") {
		t.Fatalf("CF-NOACC = %+v %q", noAccount.Checks, noAccount.Warnings)
	}

	// งบที่ไม่ใช่งบกระแสเงินสด: ไม่มีผลตรวจ และ JSON ไม่มีคีย์ checks
	pnl := report("PL-NOCHECK", "2027", "", "")
	payload, err := json.Marshal(pnl)
	if err != nil {
		t.Fatal(err)
	}
	if len(pnl.Checks) != 0 || strings.Contains(string(payload), `"checks"`) {
		t.Fatalf("pnl checks = %s", payload)
	}
	payload, _ = json.Marshal(ok)
	if !strings.Contains(string(payload), `"checks":[{"key":"amount","fiscalyear":"2027","rowno":40,`) {
		t.Fatalf("cash-flow JSON = %s", payload)
	}
}
