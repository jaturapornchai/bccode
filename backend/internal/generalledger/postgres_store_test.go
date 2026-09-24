package generalledger_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	gl "smlcloudplatform/internal/generalledger"
)

func getTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("BC_GL_TEST_POSTGRES_DSN")
	if dsn == "" {
		// ห้าม fallback ไป localhost:5432 — เครื่อง dev มี postgres ของ tenant จริงรันอยู่ ต้องใช้ฐานทดสอบแยกเท่านั้น
		t.Skip("set BC_GL_TEST_POSTGRES_DSN to isolated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("Skipping PostgreSQL test: unable to open DB: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("Skipping PostgreSQL test: unable to ping DB at %s: %v", dsn, err)
	}
	return db
}

func TestPostgresStore_FullLedgerFlow(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()
	reqNonce := time.Now().UnixNano()
	scope := gl.Scope{
		Holding: fmt.Sprintf("test_holding_%d", reqNonce),
		Company: fmt.Sprintf("test_company_%d", reqNonce),
		Actor:   "test_user_jead",
	}

	pg := gl.NewPostgres(func(holding string) (*sql.DB, error) {
		return db, nil
	})
	store := gl.NewPostgresStore(pg)

	// 1. Create Chart of Accounts
	accounts := []struct {
		code, name, accType, normal string
	}{
		{"1101", "เงินสดในมือ", "asset", "debit"},
		{"1102", "เงินฝากธนาคาร", "asset", "debit"},
		{"1103", "ลูกหนี้การค้า", "asset", "debit"},
		{"2101", "เจ้าหนี้การค้า", "liability", "credit"},
		{"3101", "ทุนจดทะเบียน", "equity", "credit"},
		{"3200", "กำไรสะสม", "equity", "credit"},
		{"3300", "กำไรสุทธิประจำปี", "equity", "credit"},
		{"4101", "รายได้จากการขาย", "income", "credit"},
		{"5101", "ต้นทุนขาย", "expense", "debit"},
		{"5201", "ค่าใช้จ่ายทั่วไป", "expense", "debit"},
	}

	for i, acc := range accounts {
		cmd := gl.Command{
			RequestID: fmt.Sprintf("req-acc-create-%d-%d", reqNonce, i),
			Resource:  "accounts",
			Action:    "create",
			Account: &gl.Account{
				AccountCode:   acc.code,
				Names:         []gl.Name{{Code: "th", Name: acc.name}},
				AccountType:   acc.accType,
				NormalBalance: acc.normal,
				AllowPosting:  true,
				IsActive:      true,
				Level:         1,
			},
		}
		res, err := store.Execute(ctx, scope, cmd)
		if err != nil {
			t.Fatalf("Failed to create account %s: %v", acc.code, err)
		}
		if res.ID == "" {
			t.Fatalf("Expected account ID to be returned for %s", acc.code)
		}
	}

	// 2. Create Fiscal Year (2026)
	fyCmd := gl.Command{
		RequestID: fmt.Sprintf("req-fy-create-%d", reqNonce),
		Resource:  "fiscal-years",
		Action:    "create",
		FiscalYear: &gl.FiscalYear{
			Code:                    "2026",
			StartDate:               "2026-01-01",
			EndDate:                 "2026-12-31",
			Scale:                   2,
			IsActive:                true,
			ProfitLossAccount:       "3300",
			RetainedEarningsAccount: "3200",
		},
	}
	fyRes, err := store.Execute(ctx, scope, fyCmd)
	if err != nil {
		t.Fatalf("Failed to create fiscal year: %v", err)
	}
	if fyRes.ID == "" {
		t.Fatalf("Expected fiscal year ID")
	}

	// 3. ปีบัญชีแรกสร้างสมุดรายวันมาตรฐานให้ (JV ทั่วไป ... UV ซื้อ) พร้อมประเภทสมุด
	books, err := store.List(ctx, scope, "journal-books", "", 1, 50, gl.ListFilter{})
	if err != nil {
		t.Fatalf("list journal books: %v", err)
	}
	gotTypes := map[string]int{}
	for _, raw := range books.Items {
		var m gl.Master
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("decode journal book: %v", err)
		}
		gotTypes[m.Code] = m.BookType
	}
	for _, want := range gl.DefaultJournalBooks() {
		if gotTypes[want.Code] != want.BookType {
			t.Fatalf("default book %s type=%d want %d (all=%v)", want.Code, gotTypes[want.Code], want.BookType, gotTypes)
		}
	}

	// 4. Create Unbalanced Journal (Must Fail)
	unbalancedCmd := gl.Command{
		RequestID: fmt.Sprintf("req-jnl-unbalanced-%d", reqNonce),
		Resource:  "journals",
		Action:    "create",
		Journal: &gl.Journal{
			DocNo:       "JV2026-0001",
			Date:        "2026-03-01",
			BookCode:    "JV",
			FiscalYear:  "2026",
			Description: "รายการไม่สมดุล",
			Kind:        "manual",
			Lines: []gl.Line{
				{AccountCode: "1101", Description: "รับเงิน", Debit: gl.Amount("1000.00"), Credit: gl.Amount("0")},
				{AccountCode: "4101", Description: "รายได้ขาย", Debit: gl.Amount("0"), Credit: gl.Amount("500.00")},
			},
		},
	}
	_, err = store.Execute(ctx, scope, unbalancedCmd)
	if err == nil {
		t.Fatalf("Expected unbalanced journal to fail, but it succeeded")
	}

	// 5. Create Balanced Journal (JV2026-0001)
	journalCmd := gl.Command{
		RequestID: fmt.Sprintf("req-jnl-create-%d", reqNonce),
		Resource:  "journals",
		Action:    "create",
		Journal: &gl.Journal{
			DocNo:       "JV2026-0001",
			Date:        "2026-03-01",
			BookCode:    "JV",
			FiscalYear:  "2026",
			Description: "ขายสินค้าเป็นเงินสด",
			Kind:        "manual",
			Lines: []gl.Line{
				{AccountCode: "1101", Description: "รับเงินสด", Debit: gl.Amount("10700.00"), Credit: gl.Amount("0")},
				{AccountCode: "4101", Description: "รายได้ขาย", Debit: gl.Amount("0"), Credit: gl.Amount("10000.00")},
				{AccountCode: "2101", Description: "เจ้าหนี้/ภาษี", Debit: gl.Amount("0"), Credit: gl.Amount("700.00")},
			},
		},
	}
	jnlRes, err := store.Execute(ctx, scope, journalCmd)
	if err != nil {
		t.Fatalf("Failed to create balanced journal: %v", err)
	}
	journalID := jnlRes.ID
	if journalID == "" {
		t.Fatalf("Expected valid journal ID")
	}

	// 6. Post Journal (ผ่านรายการ -> บันทึกลง gl_lines ทันที)
	postCmd := gl.Command{
		RequestID: fmt.Sprintf("req-jnl-post-%d", reqNonce),
		Resource:  "journals",
		Action:    "post",
		ID:        journalID,
		Version:   jnlRes.Version,
		Reason:    "ผ่านรายการปกติโดยนักบัญชี",
	}
	postRes, err := store.Execute(ctx, scope, postCmd)
	if err != nil {
		t.Fatalf("Failed to post journal: %v", err)
	}
	if postRes.ID != journalID {
		t.Fatalf("Expected posted journal ID %s, got %s", journalID, postRes.ID)
	}

	// 7. Verify gl_lines in PostgreSQL
	var lineCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM gl_lines WHERE company = $1 AND journal_id = $2`, scope.Company, journalID).Scan(&lineCount)
	if err != nil {
		t.Fatalf("Failed to query gl_lines: %v", err)
	}
	if lineCount != 3 {
		t.Fatalf("Expected 3 gl_lines, got %d", lineCount)
	}

	// 8. Report: Trial Balance
	tbReport, err := store.Report(ctx, scope, "trialbalance", gl.ReportQuery{
		FiscalYear: "2026",
	})
	if err != nil {
		t.Fatalf("Failed to generate trialbalance report: %v", err)
	}
	tbDebit, err := decimal.NewFromString(tbReport.Totals["debit"])
	if err != nil {
		t.Fatalf("Invalid debit total in trialbalance: %v", err)
	}
	tbCredit, err := decimal.NewFromString(tbReport.Totals["credit"])
	if err != nil {
		t.Fatalf("Invalid credit total in trialbalance: %v", err)
	}
	if !tbDebit.Equal(tbCredit) {
		t.Fatalf("Trial balance is not balanced! Debit: %s, Credit: %s", tbDebit.String(), tbCredit.String())
	}
	if !tbDebit.Equal(decimal.NewFromInt(10700)) {
		t.Fatalf("Expected TotalDebit to be 10700, got %s", tbDebit.String())
	}

	// 9. Report: General Ledger
	glReport, err := store.Report(ctx, scope, "ledger", gl.ReportQuery{
		FiscalYear: "2026",
	})
	if err != nil {
		t.Fatalf("Failed to generate ledger report: %v", err)
	}
	if len(glReport.Rows) == 0 {
		t.Fatalf("Expected rows in ledger report")
	}

	// 10. Reverse Journal (กลับรายการ)
	revCmd := gl.Command{
		RequestID: fmt.Sprintf("req-jnl-rev-%d", reqNonce),
		Resource:  "journals",
		Action:    "reverse",
		ID:        journalID,
		Version:   postRes.Version,
		DocNo:     "JV2026-0001R",
		Date:      "2026-03-02",
		Reason:    "ยกเลิกรายการเนื่องจากคีย์ผิด",
	}
	revRes, err := store.Execute(ctx, scope, revCmd)
	if err != nil {
		t.Fatalf("Failed to reverse journal: %v", err)
	}
	if revRes.ID == "" {
		t.Fatalf("Expected reversal journal ID")
	}

	// Verify that the reversal posted reverse lines
	var totalLinesAfterRev int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM gl_lines WHERE company = $1`, scope.Company).Scan(&totalLinesAfterRev)
	if err != nil {
		t.Fatalf("Failed to count lines after reversal: %v", err)
	}
	if totalLinesAfterRev != 6 { // Original 3 lines + 3 reversal lines
		t.Fatalf("Expected 6 total gl_lines after reversal, got %d", totalLinesAfterRev)
	}

	// Verify Net Balance in Trial Balance after reversal is balanced
	tbAfterRev, err := store.Report(ctx, scope, "trialbalance", gl.ReportQuery{
		FiscalYear: "2026",
	})
	if err != nil {
		t.Fatalf("Failed to generate trialbalance after reversal: %v", err)
	}
	tbRevDebit, _ := decimal.NewFromString(tbAfterRev.Totals["debit"])
	tbRevCredit, _ := decimal.NewFromString(tbAfterRev.Totals["credit"])
	if !tbRevDebit.Equal(tbRevCredit) {
		t.Fatalf("Trial balance after reversal is not balanced! Debit: %s, Credit: %s", tbRevDebit.String(), tbRevCredit.String())
	}
	// Total turnover should be 10700 + 10700 = 21400
	if !tbRevDebit.Equal(decimal.NewFromInt(21400)) {
		t.Fatalf("Expected TotalDebit after reversal to be 21400, got %s", tbRevDebit.String())
	}

	t.Logf("Pure PostgreSQL General Ledger Flow Passed 100%% with ACID direct transactions!")
}
