//go:build integration

package generalledger

import (
	"context"
	"sort"
	"strings"
	"testing"
)

// detailsFixture: ผังบัญชี 1000 · 4 > 4000 (ตั้งแยกแผนก) / 4100 · 5000, สมุดรายวัน JV + PV, ใบสำคัญที่ผ่านรายการแล้ว
//
//	G-001 JV 2026-04-01 "ขายสินค้า⏎บรรทัดสอง"  Dr 1000 / Cr 4000 100  แผนก D1 โครงการ P1
//	G-003 JV 2026-04-01 "รับชำระ"               Dr 1000 / Cr 4000  50  ไม่ระบุแผนก/โครงการ · ใบกำกับภาษี TI-9, TI-10
//	G-002 PV 2026-04-02 (ไม่มีคำอธิบาย)          Dr 1000 / Cr 4000  25  ไม่ระบุแผนก/โครงการ
//	G-004 JV 2026-04-03 "จ่ายค่าใช้จ่าย"          Dr 5000 / Cr 1000  40  แผนก D2 โครงการ P2
//	G-005 JV 2026-04-04 (ไม่มีคำอธิบาย)          Dr 1000 / Cr 4000  20  แผนก D2 โครงการ P1
func detailsFixture(t *testing.T) (*Postgres, Scope) {
	t.Helper()
	p, _ := glpgTestDB(t)
	company := "A"
	changes := []Change{}
	for _, row := range []struct {
		code, parent, kind, normal string
		level                      int
		posting, departSeparate    bool
	}{
		{"1000", "", "asset", "debit", 1, true, false},
		{"4", "", "income", "credit", 1, false, false},
		{"4000", "4", "income", "credit", 2, true, true},
		{"4100", "4", "income", "credit", 2, true, false},
		{"5000", "", "expense", "debit", 1, true, false},
	} {
		a := Account{Identity: glpgIdentity("account-"+row.code, company, 1), AccountCode: row.code, Names: []Name{{Code: "th", Name: "บัญชี " + row.code}}, AccountType: row.kind, ParentAccountCode: row.parent, NormalBalance: row.normal, AllowPosting: row.posting, DepartSeparate: row.departSeparate, IsActive: true, Level: row.level}
		changes = append(changes, glpgChange(t, "accounts", a.ID, a.AccountCode, a))
	}
	year := FiscalYear{Identity: glpgIdentity("year-2026", company, 1), Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", IsActive: true, Scale: 2}
	changes = append(changes, glpgChange(t, "fiscal-years", year.ID, year.Code, year))
	for _, code := range []string{"JV", "PV"} {
		book := Master{Identity: glpgIdentity("book-"+code, company, 1), Kind: "journal-books", Code: code, Name: "สมุด " + code, BookType: BookTypeGeneral, IsActive: true}
		changes = append(changes, glpgChange(t, "journal-books", book.ID, book.Code, book))
	}
	if err := p.Project(context.Background(), glpgEvent(company, 1, changes...)); err != nil {
		t.Fatal(err)
	}
	entry := func(id, book, date, description, debit, credit, amount, department, project string) Journal {
		j := glpgJournal(company, id, date, "manual", "posted", debit, credit, amount)
		j.BookCode, j.Description = book, description
		for i := range j.Lines {
			j.Lines[i].DepartmentCode, j.Lines[i].ProjectCode = department, project
		}
		return j
	}
	taxed := entry("G-003", "JV", "2026-04-01", "รับชำระ", "1000", "4000", "50", "", "")
	taxed.Details = &JournalDetails{Vats: []SubledgerVat{
		{TaxType: 2, DocumentType: 1, TaxInvoiceNo: "TI-9", TaxInvoiceDate: "2026-04-01"},
		{TaxType: 2, DocumentType: 1, TaxInvoiceNo: "TI-10", TaxInvoiceDate: "2026-04-02"},
	}}
	for i, j := range []Journal{
		entry("G-001", "JV", "2026-04-01", "ขายสินค้า\nบรรทัดสอง", "1000", "4000", "100", "D1", "P1"),
		taxed,
		entry("G-002", "PV", "2026-04-02", "", "1000", "4000", "25", "", ""),
		entry("G-004", "JV", "2026-04-03", "จ่ายค่าใช้จ่าย", "5000", "1000", "40", "D2", "P2"),
		entry("G-005", "JV", "2026-04-04", "", "1000", "4000", "20", "D2", "P1"),
	} {
		glpgSendJournal(t, p, int64(i+2), j)
	}
	return p, Scope{Holding: "H", Company: company}
}

func runDetailsReport(t *testing.T, p *Postgres, scope Scope, name string, q ReportQuery) Report {
	t.Helper()
	q.FiscalYear, q.Limit = "2026", 1000
	report, err := p.Report(context.Background(), scope, name, q)
	if err != nil {
		t.Fatalf("%s %+v: %v", name, q, err)
	}
	return report
}

// Champ 5534 (GLRepGL2View.cpp:886-926): the description column is the FIRST line of the voucher's description plus the
// line's dimension codes; the line's own text moves to its own column.
func TestLedgerDescriptionIsVoucherHeaderWithDimensionCodes(t *testing.T) {
	p, scope := detailsFixture(t)
	report := runDetailsReport(t, p, scope, "ledger", ReportQuery{AccountCode: "4000"})
	got := map[string][2]string{}
	for _, row := range report.Rows {
		if row["rowtype"] == "line" {
			got[row["docno"]] = [2]string{row["description"], row["linedescription"]}
		}
	}
	want := map[string][2]string{
		"G-001": {"ขายสินค้า,Dep: D1,Proj: P1,Bran: B1", "เครดิต"},
		"G-003": {"รับชำระ,Bran: B1", "เครดิต"},
		"G-002": {"Bran: B1", "เครดิต"},
		"G-005": {"Dep: D2,Proj: P1,Bran: B1", "เครดิต"},
	}
	if len(got) != len(want) {
		t.Fatalf("ledger lines = %v", got)
	}
	for docno, pair := range want {
		if got[docno] != pair {
			t.Errorf("%s description/linedescription = %q, want %q", docno, got[docno], pair)
		}
	}
	keys := make([]string, len(report.Columns))
	for i, column := range report.Columns {
		keys[i] = column.Key
	}
	if !strings.Contains(strings.Join(keys, ","), "balance,linedescription,branchcode") {
		t.Errorf("ledger columns = %v", keys)
	}
	// the opening / carry / total rows have no line text
	for _, row := range report.Rows {
		if row["rowtype"] != "line" && row["rowtype"] != "daytotal" && row["linedescription"] != "" {
			t.Errorf("%s row carries line text %q", row["rowtype"], row["linedescription"])
		}
	}
	// day totals (the existing option) still work next to the new column
	withDays := runDetailsReport(t, p, scope, "ledger", ReportQuery{AccountCode: "4000", DayTotals: true})
	days := 0
	for _, row := range withDays.Rows {
		if row["rowtype"] == "daytotal" {
			days++
		}
	}
	if days != 3 { // 04-01 (G-001 + G-003), 04-02, 04-04
		t.Errorf("ledger day totals = %d, want 3", days)
	}
}

// documentSequence lists the structural rows of the daily journal report: document numbers, "=" for a document total,
// "Σ<date>" for a day total.
func documentSequence(report Report) string {
	var out []string
	for _, row := range report.Rows {
		switch row["rowtype"] {
		case "document", "docsummary":
			out = append(out, row["docno"])
		case "doctotal":
			out = append(out, "=")
		case "daytotal":
			out = append(out, "Σ"+row["date"])
		}
	}
	return strings.Join(out, ",")
}

func documentNumbers(report Report) string {
	var out []string
	for _, row := range report.Rows {
		if row["rowtype"] == "document" || row["rowtype"] == "docsummary" {
			out = append(out, row["docno"])
		}
	}
	return strings.Join(out, ",")
}

func TestDailyJournalDocumentRangeOrderAndBooks(t *testing.T) {
	p, scope := detailsFixture(t)
	if got := documentNumbers(runDetailsReport(t, p, scope, "gljournal", ReportQuery{})); got != "G-001,G-002,G-003,G-004,G-005" {
		t.Fatalf("default order (document number) = %s", got)
	}
	for _, tc := range []struct {
		name  string
		query ReportQuery
		want  string
	}{
		{"order by date", ReportQuery{OrderBy: "date"}, "G-001,G-003,G-002,G-004,G-005"},
		{"order by document number", ReportQuery{OrderBy: "docno"}, "G-001,G-002,G-003,G-004,G-005"},
		{"range from-to", ReportQuery{DocNoFrom: "G-002", DocNoTo: "G-004"}, "G-002,G-003,G-004"},
		{"from alone is that one document (Champ DocNo = from)", ReportQuery{DocNoFrom: "G-004"}, "G-004"},
		{"range with order by date", ReportQuery{DocNoFrom: "G-002", DocNoTo: "G-004", OrderBy: "date"}, "G-003,G-002,G-004"},
		{"one book", ReportQuery{BookCode: "PV"}, "G-002"},
		{"several books", ReportQuery{BookCode: "JV,PV"}, "G-001,G-002,G-003,G-004,G-005"},
		{"messy book list", ReportQuery{BookCode: " PV , PV,, "}, "G-002"},
		{"other book only", ReportQuery{BookCode: "JV"}, "G-001,G-003,G-004,G-005"},
		{"book + range", ReportQuery{BookCode: "JV", DocNoFrom: "G-002", DocNoTo: "G-004"}, "G-003,G-004"},
		{"unknown book", ReportQuery{BookCode: "XX"}, ""},
	} {
		if got := documentNumbers(runDetailsReport(t, p, scope, "gljournal", tc.query)); got != tc.want {
			t.Errorf("gljournal %s = %q, want %q", tc.name, got, tc.want)
		}
		// 5533 (summary) takes the same options
		summary := tc.query
		summary.Summary = true
		if got := documentNumbers(runDetailsReport(t, p, scope, "gljournal", summary)); got != tc.want {
			t.Errorf("gljournal summary %s = %q, want %q", tc.name, got, tc.want)
		}
	}
	// the book list works for every report that reads the ledger lines, not only the daily journal
	ledger := runDetailsReport(t, p, scope, "ledger", ReportQuery{AccountCode: "4000", BookCode: "PV"})
	lines := 0
	for _, row := range ledger.Rows {
		if row["rowtype"] == "line" {
			lines++
			if row["docno"] != "G-002" {
				t.Errorf("ledger book PV row = %v", row)
			}
		}
	}
	if lines != 1 {
		t.Errorf("ledger book PV lines = %d, want 1", lines)
	}
	// the daily check reads journals (not ledger lines) and takes the same book list
	for _, tc := range []struct{ books, want string }{{"PV", "G-002"}, {"JV,PV", "G-001,G-002,G-003,G-004,G-005"}, {"XX", ""}} {
		var got []string
		for _, row := range runDetailsReport(t, p, scope, "daily-check", ReportQuery{BookCode: tc.books}).Rows {
			got = append(got, row["docno"])
		}
		sort.Strings(got)
		if joined := strings.Join(got, ","); joined != tc.want {
			t.Errorf("daily-check books %q = %q, want %q", tc.books, joined, tc.want)
		}
	}
	if _, err := p.Report(context.Background(), scope, "gljournal", ReportQuery{FiscalYear: "2026", DocNoTo: "G-004"}); err == nil {
		t.Error("document number 'to' without 'from' accepted")
	} else {
		setUserError(t, err, "report_docno_from_required")
	}
	if _, err := p.Report(context.Background(), scope, "gljournal", ReportQuery{FiscalYear: "2026", OrderBy: "amount"}); err == nil {
		t.Error("unknown order accepted")
	} else {
		setUserError(t, err, "report_orderby_invalid")
	}
}

func TestDailyJournalDayTotals(t *testing.T) {
	p, scope := detailsFixture(t)
	// 5532: the option only exists when sorted by date (GLRepTransDlg.cpp:291-299)
	plain := runDetailsReport(t, p, scope, "gljournal", ReportQuery{DayTotals: true})
	if strings.Contains(documentSequence(plain), "Σ") {
		t.Errorf("day totals while sorted by document number: %s", documentSequence(plain))
	}
	report := runDetailsReport(t, p, scope, "gljournal", ReportQuery{OrderBy: "date", DayTotals: true})
	const want = "G-001,=,G-003,=,Σ2026-04-01,G-002,=,Σ2026-04-02,G-004,=,Σ2026-04-03,G-005,=,Σ2026-04-04"
	if got := documentSequence(report); got != want {
		t.Fatalf("sequence = %s\nwant       %s", got, want)
	}
	type day struct{ lines, debit, credit string }
	wantDays := map[string]day{
		"2026-04-01": {"4", "150", "150"}, "2026-04-02": {"2", "25", "25"}, "2026-04-03": {"2", "40", "40"}, "2026-04-04": {"2", "20", "20"},
	}
	for _, row := range report.Rows {
		if row["rowtype"] != "daytotal" {
			continue
		}
		w := wantDays[row["date"]]
		if row["linecount"] != w.lines {
			t.Errorf("%s lines = %s, want %s", row["date"], row["linecount"], w.lines)
		}
		glpgDecimalEqual(t, row["debit"], w.debit)
		glpgDecimalEqual(t, row["credit"], w.credit)
		if row["docno"] != "" || row["journalid"] != "" {
			t.Errorf("day total row names a document: %v", row)
		}
	}
	// the grand total counts the document lines only — never the day total rows again
	glpgDecimalEqual(t, report.Totals["debit"], "235")
	glpgDecimalEqual(t, report.Totals["credit"], "235")
	plainTotals := runDetailsReport(t, p, scope, "gljournal", ReportQuery{OrderBy: "date"})
	glpgDecimalEqual(t, plainTotals.Totals["debit"], "235")
	if strings.Contains(documentSequence(plainTotals), "Σ") {
		t.Errorf("day totals without the option: %s", documentSequence(plainTotals))
	}

	// 5533: sorted by date Champ always prints the day total "รวม n รายการ" (GLRepTrans2View.cpp:186-198)
	summary := runDetailsReport(t, p, scope, "gljournal", ReportQuery{Summary: true, OrderBy: "date"})
	if got := documentSequence(summary); got != "G-001,G-003,Σ2026-04-01,G-002,Σ2026-04-02,G-004,Σ2026-04-03,G-005,Σ2026-04-04" {
		t.Fatalf("summary sequence = %s", got)
	}
	for _, row := range summary.Rows {
		if row["rowtype"] == "daytotal" && row["date"] == "2026-04-01" {
			if row["linecount"] != "2" {
				t.Errorf("summary 04-01 documents = %s, want 2", row["linecount"])
			}
			glpgDecimalEqual(t, row["amount"], "150")
		}
	}
	glpgDecimalEqual(t, summary.Totals["amount"], "235")
	if got := documentSequence(runDetailsReport(t, p, scope, "gljournal", ReportQuery{Summary: true})); strings.Contains(got, "Σ") {
		t.Errorf("summary sorted by document number has day totals: %s", got)
	}
}

// Champ 5532 prints the tax date and tax number of the document in its header row (GLRepTransView.cpp:305-327).
func TestDailyJournalHeaderShowsTaxInvoice(t *testing.T) {
	p, scope := detailsFixture(t)
	report := runDetailsReport(t, p, scope, "gljournal", ReportQuery{})
	keys := make([]string, len(report.Columns))
	for i, column := range report.Columns {
		keys[i] = column.Key
	}
	if !strings.Contains(strings.Join(keys, ","), "docno,taxinvoicedate,taxinvoiceno,referenceno") {
		t.Fatalf("columns = %v", keys)
	}
	for _, row := range report.Rows {
		switch {
		case row["rowtype"] == "document" && row["docno"] == "G-003":
			if row["taxinvoiceno"] != "TI-9, TI-10" || row["taxinvoicedate"] != "2026-04-01, 2026-04-02" {
				t.Errorf("G-003 header tax = %q / %q", row["taxinvoiceno"], row["taxinvoicedate"])
			}
		case row["rowtype"] == "document":
			if row["taxinvoiceno"] != "" || row["taxinvoicedate"] != "" {
				t.Errorf("%s header has tax %q / %q", row["docno"], row["taxinvoiceno"], row["taxinvoicedate"])
			}
		case row["taxinvoiceno"] != "" || row["taxinvoicedate"] != "":
			t.Errorf("%s row carries the document tax: %v", row["rowtype"], row)
		}
	}
}

// trialRows prints the trial balance rows as code:department:project, [code] for a control header, code* for a control
// total and code= for the "มียอดรวม" total of a department-separate account.
func trialRows(report Report) string {
	var out []string
	for _, row := range report.Rows {
		switch row["rowtype"] {
		case "header":
			out = append(out, "["+row["accountcode"]+"]")
		case "control":
			out = append(out, row["accountcode"]+"*")
		case "dimtotal":
			out = append(out, row["accountcode"]+"=")
		default:
			out = append(out, row["accountcode"]+":"+row["departmentcode"]+":"+row["projectcode"])
		}
	}
	return strings.Join(out, ",")
}

func rowOf(t *testing.T, report Report, rowtype, code, department, project string) map[string]string {
	t.Helper()
	for _, row := range report.Rows {
		if row["rowtype"] == rowtype && row["accountcode"] == code && row["departmentcode"] == department && row["projectcode"] == project {
			return row
		}
	}
	t.Fatalf("no %s row %s/%s/%s in %s", rowtype, code, department, project, trialRows(report))
	return nil
}

// Champ 5506/5499/5535 "แยกแผนก" / "แยกโครงการ" (GLRepTrialBalanceDlg.cpp:687-715) and "มียอดรวม" (GLRepTrialBalanceView.cpp:706-717).
func TestTrialBalanceBreakdownByDepartmentAndProject(t *testing.T) {
	p, scope := detailsFixture(t)
	plain := runDetailsReport(t, p, scope, "trialbalance", ReportQuery{ControlTotals: true})
	if got := trialRows(plain); got != "1000::,[4],4000::,4100::,4*,5000::" {
		t.Fatalf("plain trial balance = %s", got)
	}

	byDepartment := runDetailsReport(t, p, scope, "trialbalance", ReportQuery{ControlTotals: true, ByDepartment: true})
	if got := trialRows(byDepartment); got != "1000::,1000:D1:,1000:D2:,[4],4000::,4000:D1:,4000:D2:,4100::,4*,5000:D2:" {
		t.Fatalf("by department = %s", got)
	}
	// 1000: no department 75 Dr, D1 100 Dr, D2 20 movement Cr (20 in, 40 out)
	glpgDecimalEqual(t, rowOf(t, byDepartment, "account", "1000", "", "")["endingdebit"], "75")
	glpgDecimalEqual(t, rowOf(t, byDepartment, "account", "1000", "D1", "")["endingdebit"], "100")
	glpgDecimalEqual(t, rowOf(t, byDepartment, "account", "1000", "D2", "")["endingcredit"], "20")
	glpgDecimalEqual(t, rowOf(t, byDepartment, "account", "1000", "D2", "")["endingdebit"], "0")
	glpgDecimalEqual(t, rowOf(t, byDepartment, "account", "4000", "D1", "")["endingcredit"], "100")
	// each row is netted by itself (Champ), so the report totals differ from the plain report but still balance
	glpgDecimalEqual(t, byDepartment.Totals["endingdebit"], "215")
	glpgDecimalEqual(t, byDepartment.Totals["endingcredit"], "215")
	glpgDecimalEqual(t, byDepartment.Totals["difference"], "0")
	glpgDecimalEqual(t, plain.Totals["endingdebit"], "195")
	// the control total of "4" adds up every breakdown row
	for _, row := range byDepartment.Rows {
		if row["rowtype"] == "control" && row["accountcode"] == "4" {
			glpgDecimalEqual(t, row["endingcredit"], "195")
			if row["departmentcode"] != "" {
				t.Errorf("control row names a department: %v", row)
			}
		}
	}

	totals := runDetailsReport(t, p, scope, "trialbalance", ReportQuery{ControlTotals: true, ByDepartment: true, DimTotals: true})
	// only the department-separate account (4000) gets the "มียอดรวม" row, right after its department rows
	if got := trialRows(totals); got != "1000::,1000:D1:,1000:D2:,[4],4000::,4000:D1:,4000:D2:,4000=,4100::,4*,5000:D2:" {
		t.Fatalf("with totals = %s", got)
	}
	var total map[string]string
	for _, row := range totals.Rows {
		if row["rowtype"] == "dimtotal" {
			total = row
		}
	}
	glpgDecimalEqual(t, total["endingcredit"], "195")
	glpgDecimalEqual(t, total["endingdebit"], "0")
	if total["departmentcode"] != "" || total["accountlevel"] != "2" {
		t.Errorf("total row = %v", total)
	}
	// the total row never counts again in the report totals
	glpgDecimalEqual(t, totals.Totals["endingcredit"], "215")
	glpgDecimalEqual(t, totals.Totals["difference"], "0")
	// "มียอดรวม" belongs to the department breakdown only (m_iSumDepart is read when m_iDepart == 1)
	if got := trialRows(runDetailsReport(t, p, scope, "trialbalance", ReportQuery{ControlTotals: true, DimTotals: true})); got != "1000::,[4],4000::,4100::,4*,5000::" {
		t.Errorf("totals without a breakdown = %s", got)
	}
	if got := trialRows(runDetailsReport(t, p, scope, "trialbalance", ReportQuery{ByProject: true, DimTotals: true})); strings.Contains(got, "=") {
		t.Errorf("project breakdown has a total row: %s", got)
	}

	byProject := runDetailsReport(t, p, scope, "trialbalance", ReportQuery{ByProject: true})
	if got := trialRows(byProject); got != "1000::,1000::P1,1000::P2,[4],4000::,4000::P1,4100::,5000::P2" {
		t.Fatalf("by project = %s", got)
	}
	glpgDecimalEqual(t, rowOf(t, byProject, "account", "4000", "", "P1")["endingcredit"], "120")
	both := runDetailsReport(t, p, scope, "trialbalance", ReportQuery{ByDepartment: true, ByProject: true, HideZero: true})
	if got := trialRows(both); got != "1000::,1000:D1:P1,1000:D2:P1,1000:D2:P2,[4],4000::,4000:D1:P1,4000:D2:P1,5000:D2:P2" {
		t.Fatalf("by department and project, zero rows hidden = %s", got)
	}
	keys := make([]string, len(both.Columns))
	for i, column := range both.Columns {
		keys[i] = column.Key
	}
	if !strings.HasPrefix(strings.Join(keys, ","), "accountcode,accountname,departmentcode,projectcode,openingdebit") {
		t.Errorf("columns = %v", keys)
	}
	// the department filter still narrows the breakdown
	filtered := runDetailsReport(t, p, scope, "trialbalance", ReportQuery{ByDepartment: true, DepartmentCode: "D1", HideZero: true})
	if got := trialRows(filtered); got != "1000:D1:,[4],4000:D1:" {
		t.Errorf("filtered breakdown = %s", got)
	}
}

func TestBreakdownReachesWorkingPaperAndLedgerSummary(t *testing.T) {
	p, scope := detailsFixture(t)
	paper := runDetailsReport(t, p, scope, "workingpaper", ReportQuery{ByDepartment: true, DimTotals: true, HideZero: true})
	if got := trialRows(paper); got != "1000::,1000:D1:,1000:D2:,[4],4000::,4000:D1:,4000:D2:,4000=,5000:D2:" {
		t.Fatalf("working paper = %s", got)
	}
	if len(paper.Footer) != 2 {
		t.Errorf("working paper footer = %v", paper.Footer)
	}
	// 4000 is income: its movement lands in the profit-and-loss credit column, the total row adds the three rows
	glpgDecimalEqual(t, rowOf(t, paper, "dimtotal", "4000", "", "")["pnlcredit"], "195")

	summary := runDetailsReport(t, p, scope, "ledger", ReportQuery{Summary: true, ByDepartment: true, DimTotals: true, HideZero: true})
	if got := trialRows(summary); got != "1000::,1000:D1:,1000:D2:,[4],4000::,4000:D1:,4000:D2:,4000=,4100::,5000:D2:" {
		t.Fatalf("ledger summary = %s", got)
	}
	glpgDecimalEqual(t, rowOf(t, summary, "account", "4000", "D1", "")["credit"], "100")
	glpgDecimalEqual(t, rowOf(t, summary, "dimtotal", "4000", "", "")["credit"], "195")
	// the detailed ledger ignores the breakdown options
	if _, err := p.Report(context.Background(), scope, "ledger", ReportQuery{FiscalYear: "2026", ByDepartment: true, DimTotals: true}); err != nil {
		t.Fatalf("ledger with breakdown options: %v", err)
	}
}
