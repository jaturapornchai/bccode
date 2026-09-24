//go:build integration

package generalledger_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	gl "smlcloudplatform/internal/generalledger"
)

// TestFreshInstall_FullStackSubledgerUAT พิสูจน์วงจรบัญชีเต็มรูปบนบริษัททดสอบแยก
// (THAI_HOLDING/02) ตามขอบเขตตรวจรับ: ใบแจ้งหนี้ AR/AP หลายบรรทัด ชำระบางส่วน
// นำเข้า Statement ซ้ำ จับคู่/ถอน/ย้ายคู่ตัดยอด กลับรายการ rebuild และขอบเขตสาขา
// พร้อมตรวจทั้งยอดรายงานและ decimal ที่ persist จริงใน PostgreSQL
func TestFreshInstall_FullStackSubledgerUAT(t *testing.T) {
	pg, db := uatFreshInstallDB(t)
	ctx := context.Background()
	scope := gl.Scope{Holding: "THAI_HOLDING", Company: "02", Branch: "00000", Actor: "uat-fullstack"}
	store := gl.NewPostgresStore(pg)
	nonce := time.Now().UnixNano()

	// อำนวยความสะดวก: สร้างผ่านรายการ แล้วคืนผลล่าสุด (ID/Version)
	exec := func(label string, cmd gl.Command) gl.Result {
		t.Helper()
		res, err := store.Execute(ctx, scope, cmd)
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
		return res
	}
	newCmd := func(action string) gl.Command {
		return gl.Command{Resource: "journals", Action: action, RequestID: fmt.Sprintf("req-uat2-%s-%d", uuid.NewString()[:8], nonce)}
	}
	post := func(label string, res gl.Result, reason string) gl.Result {
		t.Helper()
		cmd := newCmd("post")
		cmd.ID, cmd.Version, cmd.Reason = res.ID, res.Version, reason
		return exec(label, cmd)
	}
	mustErr := func(label string, cmd gl.Command) {
		t.Helper()
		if _, err := store.Execute(ctx, scope, cmd); err == nil {
			t.Fatalf("%s: ระบบรับคำสั่งที่ควรถูกปฏิเสธ", label)
		}
	}

	// ---------------- ตั้งค่าเริ่มต้น: ผังบัญชี ปีบัญชี สมุดรายวัน ----------------
	accounts := []struct {
		code, name, typ, normal string
		cash                    bool
	}{
		{"1110", "เงินฝากธนาคารกรุงไทย สาขาใหญ่", "asset", "debit", true},
		{"1120", "ลูกหนี้การค้า", "asset", "debit", false},
		{"1130", "สินค้าคงเหลือ", "asset", "debit", false},
		{"2110", "เจ้าหนี้การค้า", "liability", "credit", false},
		{"2116", "ภาษีซื้อ", "asset", "debit", false},
		{"2118", "ภาษีขาย", "liability", "credit", false},
		{"3200", "กำไรสะสมยกมา", "equity", "credit", false},
		{"3300", "กำไร(ขาดทุน)สุทธิประจำปี", "equity", "credit", false},
		{"4100", "รายได้จากการขนส่งสินค้า", "income", "credit", false},
		{"5100", "ต้นทุนบริการขนส่ง", "expense", "debit", false},
	}
	for i, acc := range accounts {
		cmd := gl.Command{Resource: "accounts", Action: "create", RequestID: fmt.Sprintf("req-uat2-acc-%d-%d", i, nonce),
			Account: &gl.Account{AccountCode: acc.code, Names: []gl.Name{{Code: "th", Name: acc.name}}, AccountType: acc.typ, NormalBalance: acc.normal, AllowPosting: true, IsActive: true, IsCash: acc.cash, Level: 1}}
		exec("สร้างบัญชี "+acc.code, cmd)
	}
	exec("สร้างปีบัญชี 2026", gl.Command{Resource: "fiscal-years", Action: "create", RequestID: fmt.Sprintf("req-uat2-fy-%d", nonce),
		FiscalYear: &gl.FiscalYear{Code: "2026", StartDate: "2026-01-01", EndDate: "2026-12-31", Scale: 2, IsActive: true, ProfitLossAccount: "3300", RetainedEarningsAccount: "3200"}})
	// สมุดรายวันมาตรฐาน (JV PV RV SV UV พร้อมประเภท) ถูกสร้างพร้อมปีบัญชีแรกแล้ว

	// ---------------- UAT-F1: ใบแจ้งหนี้ลูกหนี้หลายบรรทัด (UV-2701) ----------------
	arCmd := newCmd("create")
	arCmd.Journal = &gl.Journal{DocNo: "UV-2701", Date: "2026-02-05", BookCode: "UV", BranchCode: "00000", FiscalYear: "2026",
		Description: "บริการขนส่งสินค้าโครงการ A เป็นเงินเชื่อ", Kind: "manual",
		Lines: []gl.Line{
			{AccountCode: "1120", Description: "ลูกหนี้การค้า", Debit: gl.Amount("26750.00"), Credit: gl.Amount("0.00")},
			{AccountCode: "4100", Description: "รายได้ค่าขนส่ง", Debit: gl.Amount("0.00"), Credit: gl.Amount("25000.00")},
			{AccountCode: "2118", Description: "ภาษีมูลค่าเพิ่ม 7%", Debit: gl.Amount("0.00"), Credit: gl.Amount("1750.00")},
		},
		Details: &gl.JournalDetails{
			Partners:    []gl.SubledgerPartner{{Code: "CUST-TH-001", Name: "บริษัท ซีพี โลจิสติกส์ จำกัด", TaxID: "0105558000006", IsCustomer: true, IsSupplier: false, IsActive: true}},
			Documents:   []gl.SubledgerDocument{{ID: "AR-INV-001", Ledger: "ar", PartnerCode: "CUST-TH-001", DocumentNo: "INV-2701-001", Date: "2026-02-05", DueDate: "2026-03-07", BranchCode: "00000", Kind: 1, Side: 1, Amount: gl.Amount("26750.00"), Currency: "THB", ControlAccountCode: "1120"}},
			Allocations: []gl.SubledgerAllocation{{ID: "ALLOC-001", Ledger: "ar", DocumentID: "AR-INV-001", LineNumber: 1, Amount: gl.Amount("26750.00")}},
		},
	}
	arRes := exec("สร้าง UV-2701", arCmd)
	post("ผ่านรายการ UV-2701", arRes, "ผ่านใบแจ้งหนี้ลูกหนี้")

	// ตรวจ decimal persist จริงใน gl_lines (NUMERIC ไม่ใช่ float)
	var drString string
	if err := db.QueryRowContext(ctx, `SELECT debit::text FROM gl_lines WHERE company=$1 AND journal_id=$2 AND line_no=1`, "02", arRes.ID).Scan(&drString); err != nil {
		t.Fatalf("อ่าน gl_lines ลูกหนี้ไม่ได้: %v", err)
	}
	if !strings.HasPrefix(drString, "26750.00000000") {
		t.Fatalf("debit ลูกหนี้ persist ผิด: %s", drString)
	}

	// ---------------- UAT-F2: ใบซื้อเจ้าหนี้หลายบรรทัด (SV-2701) ----------------
	apCmd := newCmd("create")
	apCmd.Journal = &gl.Journal{DocNo: "SV-2701", Date: "2026-02-06", BookCode: "SV", BranchCode: "00000", FiscalYear: "2026",
		Description: "ซื้อบริการขนส่งจากคู่ค้าเป็นเงินเชื่อ", Kind: "manual",
		Lines: []gl.Line{
			{AccountCode: "5100", Description: "ต้นทุนค่าขนส่ง", Debit: gl.Amount("10000.00"), Credit: gl.Amount("0.00")},
			{AccountCode: "2116", Description: "ภาษีซื้อ 7%", Debit: gl.Amount("700.00"), Credit: gl.Amount("0.00")},
			{AccountCode: "2110", Description: "เจ้าหนี้การค้า", Debit: gl.Amount("0.00"), Credit: gl.Amount("10700.00")},
		},
		Details: &gl.JournalDetails{
			Partners:    []gl.SubledgerPartner{{Code: "SUPP-TH-001", Name: "บริษัท ไทยซัพพลาย เน็ตเวิร์ค จำกัด", TaxID: "0105558000014", IsCustomer: false, IsSupplier: true, IsActive: true}},
			Documents:   []gl.SubledgerDocument{{ID: "AP-INV-001", Ledger: "ap", PartnerCode: "SUPP-TH-001", DocumentNo: "BILL-2701-001", Date: "2026-02-06", DueDate: "2026-03-08", BranchCode: "00000", Kind: 1, Side: 1, Amount: gl.Amount("10700.00"), Currency: "THB", ControlAccountCode: "2110"}},
			Allocations: []gl.SubledgerAllocation{{ID: "ALLOC-002", Ledger: "ap", DocumentID: "AP-INV-001", LineNumber: 3, Amount: gl.Amount("10700.00")}},
		},
	}
	apRes := exec("สร้าง SV-2701", apCmd)
	post("ผ่านรายการ SV-2701", apRes, "ผ่านใบซื้อเจ้าหนี้")

	// ---------------- UAT-F3: รับชำระบางส่วน 10,000 พร้อมตัดยอด (RV-2701) ----------------
	rvCmd := newCmd("create")
	rvCmd.Journal = &gl.Journal{DocNo: "RV-2701", Date: "2026-02-20", BookCode: "RV", BranchCode: "00000", FiscalYear: "2026",
		Description: "รับชำระบางส่วนใบ INV-2701-001 ทางธนาคาร", Kind: "manual",
		Lines: []gl.Line{
			{AccountCode: "1110", Description: "รับเงินเข้าธนาคาร", Debit: gl.Amount("10000.00"), Credit: gl.Amount("0.00")},
			{AccountCode: "1120", Description: "ตัดลูกหนี้บางส่วน", Debit: gl.Amount("0.00"), Credit: gl.Amount("10000.00")},
		},
		Details: &gl.JournalDetails{
			BankAccounts: []gl.SubledgerBankAccount{{Code: "BANK-KBANK", BankName: "ธนาคารกรุงไทย", AccountNumber: "1234567890", AccountName: "บริษัท รุ่งเรืองขนส่งและโลจิสติกส์ จำกัด", GLAccountCode: "1110", Currency: "THB", IsActive: true}},
			BankLines:    []gl.SubledgerBankLine{{LineNumber: 1, BankAccountCode: "BANK-KBANK", Direction: 1}},
			Documents:    []gl.SubledgerDocument{{ID: "AR-REC-001", Ledger: "ar", PartnerCode: "CUST-TH-001", DocumentNo: "REC-2701-001", Date: "2026-02-20", BranchCode: "00000", Kind: 2, Side: 2, Amount: gl.Amount("10000.00"), Currency: "THB", ControlAccountCode: "1120"}},
			Allocations:  []gl.SubledgerAllocation{{ID: "ALLOC-003", Ledger: "ar", DocumentID: "AR-REC-001", LineNumber: 2, Amount: gl.Amount("10000.00")}},
			Settlements:  []gl.SubledgerSettlement{{ID: "SETTLE-001", Ledger: "ar", PartnerCode: "CUST-TH-001", DebtDocumentID: "AR-INV-001", PaymentDocumentID: "AR-REC-001", Date: "2026-02-20", Amount: gl.Amount("10000.00")}},
		},
	}
	rvRes := exec("สร้าง RV-2701", rvCmd)
	rvPosted := post("ผ่านรายการ RV-2701", rvRes, "รับชำระบางส่วน")

	// เทียบยอดแบบ decimal (รายงานคืนเป็นข้อความตาม scale ต้นทาง เช่น 16750.00000000)
	amountEq := func(got, want string) bool {
		g, err1 := gl.ParseAmount(got)
		w, err2 := gl.ParseAmount(want)
		return err1 == nil && err2 == nil && g.Decimal().Equal(w.Decimal())
	}
	assertRemaining := func(label, docID, want string) {
		t.Helper()
		page, err := pg.SubledgerList(ctx, scope, "documents", docID, 1, 50, "")
		if err != nil {
			t.Fatalf("%s: อ่านเอกสาร %s ไม่ได้: %v", label, docID, err)
		}
		for _, raw := range page.Items {
			var row gl.SubledgerOpenDocument
			if err := json.Unmarshal(raw, &row); err != nil {
				t.Fatalf("%s: %v", label, err)
			}
			if row.ID != docID {
				continue
			}
			if got, _ := row.Remaining.Decimal().StringFixed(2), 0; got != want {
				t.Fatalf("%s: %s คงเหลือ %s คาด %s", label, docID, got, want)
			}
			return
		}
		t.Fatalf("%s: ไม่พบเอกสาร %s", label, docID)
	}
	assertRemaining("หลังรับชำระบางส่วน", "AR-INV-001", "16750.00")

	// ยอดรายงาน AR outstanding ต้องตรงฐานจริง
	arReport, err := pg.SubledgerReport(ctx, scope, "ar-outstanding", gl.ReportQuery{To: "2026-12-31"})
	if err != nil {
		t.Fatalf("รายงาน ar-outstanding ไม่ได้: %v", err)
	}
	if !amountEq(arReport.Totals["remaining_amount"], "16750.00") {
		t.Fatalf("ar-outstanding คงเหลือรวม %s คาด 16750.00", arReport.Totals["remaining_amount"])
	}

	// ---------------- UAT-F4: จ่ายเจ้าหนี้บางส่วน 5,000 (PV-2701) ----------------
	pvCmd := newCmd("create")
	pvCmd.Journal = &gl.Journal{DocNo: "PV-2701", Date: "2026-02-21", BookCode: "PV", BranchCode: "00000", FiscalYear: "2026",
		Description: "จ่ายชำระบางส่วนบิล BILL-2701-001 ทางธนาคาร", Kind: "manual",
		Lines: []gl.Line{
			{AccountCode: "2110", Description: "ตัดเจ้าหนี้บางส่วน", Debit: gl.Amount("5000.00"), Credit: gl.Amount("0.00")},
			{AccountCode: "1110", Description: "จ่ายเงินออกจากธนาคาร", Debit: gl.Amount("0.00"), Credit: gl.Amount("5000.00")},
		},
		Details: &gl.JournalDetails{
			BankAccounts: []gl.SubledgerBankAccount{{Code: "BANK-KBANK", BankName: "ธนาคารกรุงไทย", AccountNumber: "1234567890", AccountName: "บริษัท รุ่งเรืองขนส่งและโลจิสติกส์ จำกัด", GLAccountCode: "1110", Currency: "THB", IsActive: true}},
			BankLines:    []gl.SubledgerBankLine{{LineNumber: 2, BankAccountCode: "BANK-KBANK", Direction: 2}},
			Documents:    []gl.SubledgerDocument{{ID: "AP-PAY-001", Ledger: "ap", PartnerCode: "SUPP-TH-001", DocumentNo: "PAY-2701-001", Date: "2026-02-21", BranchCode: "00000", Kind: 2, Side: 2, Amount: gl.Amount("5000.00"), Currency: "THB", ControlAccountCode: "2110"}},
			Allocations:  []gl.SubledgerAllocation{{ID: "ALLOC-004", Ledger: "ap", DocumentID: "AP-PAY-001", LineNumber: 1, Amount: gl.Amount("5000.00")}},
			Settlements:  []gl.SubledgerSettlement{{ID: "SETTLE-002", Ledger: "ap", PartnerCode: "SUPP-TH-001", DebtDocumentID: "AP-INV-001", PaymentDocumentID: "AP-PAY-001", Date: "2026-02-21", Amount: gl.Amount("5000.00")}},
		},
	}
	pvRes := exec("สร้าง PV-2701", pvCmd)
	post("ผ่านรายการ PV-2701", pvRes, "จ่ายเจ้าหนี้บางส่วน")
	assertRemaining("หลังจ่ายเจ้าหนี้บางส่วน", "AP-INV-001", "5700.00")

	// ---------------- UAT-F5: กระทบยอด Statement + จับคู่ + dedup ซ้ำ ----------------
	reconcile := func(label string, version int64, d *gl.JournalDetails) gl.Result {
		t.Helper()
		cmd := newCmd("reconcile")
		cmd.ID, cmd.Version, cmd.Reason, cmd.Journal = rvPosted.ID, version, "กระทบยอดหลักฐานธนาคาร", &gl.Journal{Details: d}
		return exec(label, cmd)
	}
	rvPosted = reconcile("จับคู่ Statement", rvPosted.Version, &gl.JournalDetails{
		StatementLines: []gl.SubledgerStatementLine{{ID: "STMT-001", BankAccountCode: "BANK-KBANK", SourceKey: "uat-stmt-key-1", Date: "2026-02-21", Reference: "KBANK-000001", Description: "รับเงินเข้าบัญชี", Direction: 1, Amount: gl.Amount("10000.00")}},
		Matches:        []gl.SubledgerMatch{{ID: "MATCH-001", StatementLineID: "STMT-001", LineNumber: 1, Amount: gl.Amount("10000.00")}},
	})
	// นำเข้า Statement แถวเดิมซ้ำ (ID/SourceKey เดิม) ต้องไม่เพิ่มหลักฐาน
	var stmtCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM gl_subledger_statements WHERE company='02' AND bank_account_code='BANK-KBANK'`).Scan(&stmtCount); err != nil {
		t.Fatalf("อ่าน gl_subledger_statements ไม่ได้: %v", err)
	}
	rvPosted = reconcile("นำเข้า Statement ซ้ำ", rvPosted.Version, &gl.JournalDetails{
		StatementLines: []gl.SubledgerStatementLine{{ID: "STMT-001", BankAccountCode: "BANK-KBANK", SourceKey: "uat-stmt-key-1", Date: "2026-02-21", Reference: "KBANK-000001", Description: "รับเงินเข้าบัญชี", Direction: 1, Amount: gl.Amount("10000.00")}},
	})
	var stmtCountAfter int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM gl_subledger_statements WHERE company='02' AND bank_account_code='BANK-KBANK'`).Scan(&stmtCountAfter); err != nil {
		t.Fatalf("อ่าน gl_subledger_statements รอบสองไม่ได้: %v", err)
	}
	if stmtCountAfter != stmtCount {
		t.Fatalf("นำเข้า Statement ซ้ำเพิ่มหลักฐาน %d -> %d", stmtCount, stmtCountAfter)
	}
	// เพิ่มแถว Statement ใหม่ที่ยังไม่จับคู่ ต้องโผล่ใน bank-unmatched
	rvPosted = reconcile("เพิ่มแถวยังไม่จับคู่", rvPosted.Version, &gl.JournalDetails{
		StatementLines: []gl.SubledgerStatementLine{{ID: "STMT-002", BankAccountCode: "BANK-KBANK", SourceKey: "uat-stmt-key-2", Date: "2026-02-22", Reference: "KBANK-000002", Description: "ค่าธรรมเนียมโอน", Direction: 2, Amount: gl.Amount("50.00")}},
	})
	unmatched, err := pg.SubledgerReport(ctx, scope, "bank-unmatched", gl.ReportQuery{To: "2026-12-31"})
	if err != nil {
		t.Fatalf("รายงาน bank-unmatched ไม่ได้: %v", err)
	}
	if !amountEq(unmatched.Totals["remaining_amount"], "50.00") {
		t.Fatalf("bank-unmatched คงเหลือ %s คาด 50.00", unmatched.Totals["remaining_amount"])
	}

	// ---------------- UAT-F6: ถอนการจับคู่ + ย้ายคู่ตัดยอดในคำสั่งเดียว ----------------
	rvPosted = reconcile("ถอนการจับคู่", rvPosted.Version, &gl.JournalDetails{
		Withdrawals: []gl.SubledgerWithdrawal{{Kind: "match", ID: "MATCH-001", Reason: "จับคู่ผิดบัญชีธนาคารต้นทาง"}},
	})
	var matchReversed int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM gl_subledger_matches WHERE company='02' AND id='MATCH-001' AND reversed_at IS NULL`).Scan(&matchReversed); err != nil {
		t.Fatalf("อ่าน gl_subledger_matches ไม่ได้: %v", err)
	}
	if matchReversed != 0 {
		t.Fatal("ถอนการจับคู่แล้ว MATCH-001 ยัง active ในฐาน")
	}
	// ย้ายคู่ตัดยอด: ถอน SETTLE-001 พร้อมสร้าง SETTLE-001B ยอดเดิมใน transaction เดียว
	rvPosted = reconcile("ย้ายคู่ตัดยอด", rvPosted.Version, &gl.JournalDetails{
		Withdrawals: []gl.SubledgerWithdrawal{{Kind: "settlement", ID: "SETTLE-001", Reason: "ย้ายคู่ตัดยอดเป็นรายการใหม่"}},
		Settlements: []gl.SubledgerSettlement{{ID: "SETTLE-001B", Ledger: "ar", PartnerCode: "CUST-TH-001", DebtDocumentID: "AR-INV-001", PaymentDocumentID: "AR-REC-001", Date: "2026-02-21", Amount: gl.Amount("10000.00")}},
	})
	assertRemaining("หลังย้ายคู่ตัดยอด", "AR-INV-001", "16750.00")

	// ---------------- UAT-F7: retry idempotent + double-click + source identity ----------------
	// retry คำขอเดิมด้วย requestid เดิม: ระบบต้องคืนผลเดิม ไม่สร้างซ้ำ
	{
		retryCmd := newCmd("reconcile")
		retryCmd.ID, retryCmd.Version, retryCmd.Reason = rvPosted.ID, rvPosted.Version, "นำเข้า Statement ซ้ำสำหรับทดสอบ retry"
		retryCmd.Journal = &gl.Journal{Details: &gl.JournalDetails{
			StatementLines: []gl.SubledgerStatementLine{{ID: "STMT-001", BankAccountCode: "BANK-KBANK", SourceKey: "uat-stmt-key-1", Date: "2026-02-21", Reference: "KBANK-000001", Description: "รับเงินเข้าบัญชี", Direction: 1, Amount: gl.Amount("10000.00")}},
		}}
		first, err := store.Execute(ctx, scope, retryCmd)
		if err != nil {
			t.Fatalf("retry: คำขอแรกล้ม %v", err)
		}
		retrySame := retryCmd
		again, err := store.Execute(ctx, scope, retrySame)
		if err != nil || again.ID != first.ID || again.Version != first.Version {
			t.Fatalf("retry คำขอเดิมไม่ idempotent: got=%+v first=%+v err=%v", again, first, err)
		}
	}
	// source identity: requestid ต่างหลายทาง payload เดิม ต้องได้ใบเดียว
	{
		importJournal := gl.Journal{DocNo: "UV-2702", Date: "2026-02-25", BookCode: "UV", BranchCode: "00000", FiscalYear: "2026",
			Description: "นำเข้าจากระบบขายหน้าร้าน สาขาลาดหลุมแก้ว", Kind: "manual",
			Lines: []gl.Line{
				{AccountCode: "1120", Description: "ลูกหนี้การค้า POS", Debit: gl.Amount("1070.00"), Credit: gl.Amount("0.00")},
				{AccountCode: "4100", Description: "รายได้ค่าขนส่ง POS", Debit: gl.Amount("0.00"), Credit: gl.Amount("1070.00")},
			},
			SourceType: 2, SourceSystem: "pos-branch", SourceRecordID: "pos-2701-000042",
		}
		importCmd := gl.Command{Resource: "journals", Action: "create", RequestID: fmt.Sprintf("req-uat2-src-%d", nonce), Journal: &importJournal}
		first, err := store.Execute(ctx, scope, importCmd)
		if err != nil {
			t.Fatalf("นำเข้า source identity: %v", err)
		}
		var wg sync.WaitGroup
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				retry := importCmd
				retry.RequestID = fmt.Sprintf("req-uat2-src-%d-%d", nonce, i)
				got, e := store.Execute(ctx, scope, retry)
				if e != nil || got.ID != first.ID {
					t.Errorf("source retry got=%+v err=%v", got, e)
				}
			}()
		}
		wg.Wait()
		var importCount int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM gl_source_journals WHERE company='02' AND source_record_id='pos-2701-000042'`).Scan(&importCount); err != nil {
			t.Fatalf("อ่าน gl_source_journals ไม่ได้: %v", err)
		}
		if importCount != 1 {
			t.Fatalf("source registry มี %d แถว คาด 1", importCount)
		}
	}
	// double-click: ยิง post ใบเดียวสองทางพร้อมกัน ต้องผ่านครั้งเดียว
	{
		dcCmd := newCmd("create")
		dcCmd.Journal = &gl.Journal{DocNo: "JV-2703", Date: "2026-02-26", BookCode: "JV", BranchCode: "00000", FiscalYear: "2026",
			Description: "ตั้งค่าเจ้าหนี้ค่าเช่าคลังสินค้า", Kind: "manual",
			Lines: []gl.Line{
				{AccountCode: "5100", Description: "ค่าเช่าคลัง", Debit: gl.Amount("800.00"), Credit: gl.Amount("0.00")},
				{AccountCode: "2110", Description: "เจ้าหนี้ค่าเช่า", Debit: gl.Amount("0.00"), Credit: gl.Amount("800.00")},
			},
		}
		dcRes, err := store.Execute(ctx, scope, dcCmd)
		if err != nil {
			t.Fatalf("สร้างใบ double-click: %v", err)
		}
		postOnce := func() error {
			cmd := newCmd("post")
			cmd.ID, cmd.Version, cmd.Reason = dcRes.ID, dcRes.Version, "ผ่านรายการค่าเช่า"
			_, err := store.Execute(ctx, scope, cmd)
			return err
		}
		var wg sync.WaitGroup
		errs := make([]error, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(i int) { defer wg.Done(); errs[i] = postOnce() }(i)
		}
		wg.Wait()
		okCount := 0
		for _, e := range errs {
			if e == nil {
				okCount++
			}
		}
		if okCount != 1 {
			t.Fatalf("double-click post ผ่าน %d ครั้ง คาด 1 (errs=%v)", okCount, errs)
		}
	}

	// ---------------- UAT-F8: กลับรายการต้องถอนการตัดยอดก่อน ----------------
	// อ่าน version ล่าสุดจากระบบเสมอ เพราะการกระทบหลักฐานเพิ่ม revision ของใบที่เกี่ยวข้องด้วย
	currentVersion := func(label, id string) int64 {
		t.Helper()
		raw, err := store.Get(ctx, scope, "journals", id)
		if err != nil {
			t.Fatalf("%s: อ่านใบสำคัญไม่ได้: %v", label, err)
		}
		var j gl.Journal
		if err := json.Unmarshal(raw, &j); err != nil {
			t.Fatalf("%s: %v", label, err)
		}
		return j.Version
	}
	revBlocked := newCmd("reverse")
	revBlocked.ID, revBlocked.Version, revBlocked.DocNo, revBlocked.Date, revBlocked.Reason = apRes.ID, currentVersion("ก่อนกลับรายการ", apRes.ID), "SV-2701R", "2026-02-28", "ยกเลิกใบซื้อที่ผิดคู่ค้า"
	mustErr("กลับรายการที่มีตัดยอดค้าง", revBlocked)
	pvRecon := newCmd("reconcile")
	pvRecon.ID, pvRecon.Version, pvRecon.Reason = pvRes.ID, currentVersion("ก่อนถอนตัดยอด", pvRes.ID), "ถอนตัดยอดก่อนกลับรายการ"
	pvRecon.Journal = &gl.Journal{Details: &gl.JournalDetails{Withdrawals: []gl.SubledgerWithdrawal{{Kind: "settlement", ID: "SETTLE-002", Reason: "ถอนตัดยอดก่อนกลับรายการใบซื้อ"}}}}
	exec("ถอนตัดยอดเจ้าหนี้", pvRecon)
	revOK := newCmd("reverse")
	revOK.ID, revOK.Version, revOK.DocNo, revOK.Date, revOK.Reason = apRes.ID, currentVersion("ก่อนกลับรายการจริง", apRes.ID), "SV-2701R", "2026-02-28", "ยกเลิกใบซื้อที่ผิดคู่ค้า"
	exec("กลับรายการ SV-2701", revOK)

	// ---------------- UAT-F9: rebuild ต้องรักษาหลักฐานและยอดเดิม ----------------
	countOf := func(table string) int {
		t.Helper()
		var n int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM `+table+` WHERE company='02'`).Scan(&n); err != nil {
			t.Fatalf("นับ %s ไม่ได้: %v", table, err)
		}
		return n
	}
	sumLines := func() string {
		t.Helper()
		var s string
		if err := db.QueryRowContext(ctx, `SELECT coalesce(sum(debit),0)::text FROM gl_lines WHERE company='02'`).Scan(&s); err != nil {
			t.Fatalf("รวม gl_lines ไม่ได้: %v", err)
		}
		return s
	}
	beforeDocs, beforeSettles, beforeStatements := countOf("gl_subledger_documents"), countOf("gl_subledger_settlements"), countOf("gl_subledger_statements")
	beforeSum := sumLines()
	// process ใช้ขอบเขตบริษัท (ไม่ระบุสาขา) และระบุ ID เป้าหมาย
	companyScope := scope
	companyScope.Branch = ""
	if _, err := store.Execute(ctx, companyScope, gl.Command{Resource: "processes", Action: "recalculate", ID: "2026", Reason: "UAT rebuild", RequestID: fmt.Sprintf("req-uat2-rebuild-%d", nonce)}); err != nil {
		t.Fatalf("rebuild projection: %v", err)
	}
	if afterSum := sumLines(); afterSum != beforeSum {
		t.Fatalf("rebuild เปลี่ยนยอด gl_lines %s -> %s", beforeSum, afterSum)
	}
	if n := countOf("gl_subledger_documents"); n != beforeDocs {
		t.Fatalf("rebuild กระทบหลักฐาน documents %d -> %d", beforeDocs, n)
	}
	if n := countOf("gl_subledger_settlements"); n != beforeSettles {
		t.Fatalf("rebuild กระทบ settlements %d -> %d", beforeSettles, n)
	}
	if n := countOf("gl_subledger_statements"); n != beforeStatements {
		t.Fatalf("rebuild กระทบ statements %d -> %d", beforeStatements, n)
	}
	// งบทดลองหลังทุกขั้นตอนต้องสมดุล
	tb, err := store.Report(ctx, scope, "trialbalance", gl.ReportQuery{FiscalYear: "2026"})
	if err != nil {
		t.Fatalf("งบทดลองไม่ได้: %v", err)
	}
	if tb.Totals["debit"] != tb.Totals["credit"] {
		t.Fatalf("งบทดลองไม่สมดุล: DR %s CR %s", tb.Totals["debit"], tb.Totals["credit"])
	}

	// ---------------- UAT-F10: สอง session ต่างสาขา ต้องมองไม่เห็นของกัน ----------------
	branchScope := gl.Scope{Holding: "THAI_HOLDING", Company: "02", Branch: "00001", Actor: "uat-branch"}
	otherCmd := newCmd("create")
	otherCmd.Journal = &gl.Journal{DocNo: "UV-2704", Date: "2026-02-27", BookCode: "UV", BranchCode: "00001", FiscalYear: "2026",
		Description: "รายการสาขาลาดหลุมแก้ว", Kind: "manual",
		Lines: []gl.Line{
			{AccountCode: "1110", Description: "รับเงินสดหน้าสาขา", Debit: gl.Amount("500.00"), Credit: gl.Amount("0.00")},
			{AccountCode: "4100", Description: "รายได้ค่าขนส่งสาขา", Debit: gl.Amount("0.00"), Credit: gl.Amount("500.00")},
		},
	}
	if _, err := store.Execute(ctx, branchScope, otherCmd); err != nil {
		t.Fatalf("สร้างใบสาขา 00001 ไม่ได้: %v", err)
	}
	mainList, err := store.List(ctx, scope, "journals", "", 1, 200, gl.ListFilter{})
	if err != nil {
		t.Fatalf("อ่านรายการสาขาหลักไม่ได้: %v", err)
	}
	for _, raw := range mainList.Items {
		var j gl.Journal
		if err := json.Unmarshal(raw, &j); err != nil {
			t.Fatal(err)
		}
		if j.BranchCode != "00000" {
			t.Fatalf("สาขาหลักเห็นใบของสาขาอื่น: %s (%s)", j.DocNo, j.BranchCode)
		}
	}
	branchList, err := store.List(ctx, branchScope, "journals", "", 1, 200, gl.ListFilter{})
	if err != nil {
		t.Fatalf("อ่านรายการสาขา 00001 ไม่ได้: %v", err)
	}
	if branchList.Total == 0 {
		t.Fatal("สาขา 00001 ต้องเห็นใบของตัวเอง")
	}
	for _, raw := range branchList.Items {
		var j gl.Journal
		if err := json.Unmarshal(raw, &j); err != nil {
			t.Fatal(err)
		}
		if j.BranchCode != "00001" {
			t.Fatalf("สาขา 00001 เห็นใบของสาขาอื่น: %s (%s)", j.DocNo, j.BranchCode)
		}
	}
	if _, err := store.Get(ctx, branchScope, "journals", arRes.ID); err == nil {
		t.Fatal("สาขา 00001 อ่านใบของสาขา 00000 ได้ (scope รั่ว)")
	}

	t.Log("🎉 FULL-STACK SUBLEDGER UAT PASSED — AR/AP partial, statement dedup, withdraw/reallocate, reverse guard, rebuild, branch scope ทั้งหมดพิสูจน์บน PostgreSQL จริง")
}
