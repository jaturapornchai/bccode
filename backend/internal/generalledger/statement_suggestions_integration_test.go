//go:build integration

package generalledger

import (
	"encoding/json"
	"reflect"
	"testing"
)

// suggestAccountFixture สร้างบัญชีผ่านคำสั่งจริง (ด้านปกติตามหมวด) แล้วคืนพร้อม id/version สำหรับแก้ไข/ลบต่อ
func (f *pgIntegrityFixture) suggestAccountFixture(code, accountType, parent, name string, posting bool, edit func(*Account)) Account {
	f.t.Helper()
	normal := "credit"
	if accountType == "asset" || accountType == "expense" {
		normal = "debit"
	}
	a := Account{AccountCode: code, AccountType: accountType, NormalBalance: normal, ParentAccountCode: parent, AllowPosting: posting, IsActive: true, Names: []Name{{Code: "th", Name: name}}}
	if edit != nil {
		edit(&a)
	}
	r := f.run(Command{Resource: "accounts", Action: "create", Account: &a})
	a.ID, a.Version = r.ID, r.Version
	return a
}

func (f *pgIntegrityFixture) countRows(query string, args ...any) int {
	f.t.Helper()
	var n int
	if err := f.db.QueryRow(query, args...).Scan(&n); err != nil {
		f.t.Fatalf("%s: %v", query, err)
	}
	return n
}

func suggestBalanceSheet(code string, rows ...StatementRow) *Master {
	return &Master{Code: code, Name: "งบแสดงฐานะการเงิน", StatementType: "balance_sheet", IsActive: true, GlobalStyle: &StatementGlobalStyle{Scale: 2}, Rows: rows}
}

// บันทึกรูปแบบงบ: รหัสที่เครื่องคำนวณอ่านแล้วได้ศูนย์เสมอ (บัญชีหัวข้อ / ถูกลบ / ไม่มีในผัง) ต้องถูกปฏิเสธพร้อมบอกบรรทัด;
// บัญชีปิดใช้งานยังใช้ได้ และรหัสที่ค้างในบรรทัดชนิดอื่นไม่ถูกตรวจ
func TestStatementTemplateSaveRejectsZeroAccountsIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.suggestAccountFixture("1900", "asset", "", "สินทรัพย์หมุนเวียนอื่น", false, nil)
	f.suggestAccountFixture("1910", "asset", "1900", "เงินมัดจำจ่าย", true, nil)
	gone := f.suggestAccountFixture("1800", "asset", "", "เงินทดรองจ่ายพนักงาน", true, nil)
	f.run(Command{Resource: "accounts", Action: "delete", ID: gone.ID, Version: gone.Version})
	f.suggestAccountFixture("1700", "asset", "", "เงินฝากประจำ", true, func(a *Account) { a.IsActive = false })

	cash := StatementRow{ID: "bs-30", RowNo: 30, RowType: "account", Title: "เงินสดและรายการเทียบเท่าเงินสด", AccountCodes: []string{"1000"}, SuggestKey: "cash_and_equivalents"}
	user := f.failCode(Command{Resource: "statement-templates", Action: "create", Master: suggestBalanceSheet("BS-HEADER", cash,
		StatementRow{ID: "bs-140", RowNo: 140, RowType: "account", Title: "สินทรัพย์หมุนเวียนอื่น", AccountCodes: []string{"1900"}})}, "statement_template_row_accounts_invalid", "rows[1].accountcodes")
	if user.Status != 400 || !reflect.DeepEqual(user.Args, []string{"140", "สินทรัพย์หมุนเวียนอื่น", "1900"}) {
		t.Fatalf("header row error = %+v", user)
	}
	user = f.failCode(Command{Resource: "statement-templates", Action: "create", Master: suggestBalanceSheet("BS-DELETED",
		StatementRow{ID: "bs-40", RowNo: 40, RowType: "account", Title: "เงินทดรองจ่าย", AccountCodes: []string{"1800", "NOPE"}})}, "statement_template_row_accounts_invalid", "rows[0].accountcodes")
	if user.Args[2] != "1800, NOPE" {
		t.Fatalf("deleted/unknown codes = %+v", user.Args)
	}
	equity := &Master{Code: "EQ-HEADER", Name: "งบแสดงการเปลี่ยนแปลงส่วนของผู้ถือหุ้น", StatementType: "equity", IsActive: true,
		Rows:    []StatementRow{{ID: "eq-1", RowNo: 10, RowType: "account", Title: "ยอดต้นงวด", AmountBasis: "opening", AccountCodes: []string{"1900"}}},
		Columns: []StatementColumn{{ID: "eq-c1", Title: "ทุนที่ออกและชำระแล้ว", AccountCodes: []string{"1900"}}}}
	user = f.failCode(Command{Resource: "statement-templates", Action: "create", Master: equity}, "statement_template_column_accounts_invalid", "columns[0].accountcodes")
	if !reflect.DeepEqual(user.Args, []string{"ทุนที่ออกและชำระแล้ว", "1900"}) {
		t.Fatalf("column error = %+v", user)
	}
	f.failCode(Command{Resource: "statement-templates", Action: "create", Master: suggestBalanceSheet("BS-KEY",
		StatementRow{ID: "bs-1", RowNo: 10, RowType: "account", Title: "เงินสด", AccountCodes: []string{"1000"}, SuggestKey: "cash"})}, "statement_template_suggest_key_invalid", "rows[0].suggestkey")
	if n := f.countRows(`SELECT count(*) FROM gl_records WHERE company='C' AND kind='statement-templates'`); n != 0 {
		t.Fatalf("rejected templates were saved: %d", n)
	}

	// บัญชีลงรายการได้ + บัญชีปิดใช้งาน + กำไรที่ยังไม่ปิดบัญชี ผ่าน; บัญชีหัวข้อในบรรทัดชนิดหัวข้อไม่ตรวจ
	valid := suggestBalanceSheet("BS-OK", StatementRow{ID: "bs-10", RowNo: 10, RowType: "header", Title: "สินทรัพย์หมุนเวียน", AccountCodes: []string{"1900"}},
		StatementRow{ID: "bs-30", RowNo: 30, RowType: "account", Title: "เงินสดและรายการเทียบเท่าเงินสด", AccountCodes: []string{"1000", "1700", statementCurrentEarnings}, SuggestKey: "cash_and_equivalents"})
	r := f.run(Command{Resource: "statement-templates", Action: "create", Master: valid})
	var saved Master
	raw, err := f.store.Get(f.ctx, f.scope, "statement-templates", r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Rows[1].SuggestKey != "cash_and_equivalents" || !reflect.DeepEqual(saved.Rows[0].AccountCodes, []string{"1900"}) {
		t.Fatalf("saved template = %+v", saved.Rows)
	}
	// แก้ไข: เส้นทาง update ตรวจเหมือนกัน
	next := *valid
	next.Rows = append(append([]StatementRow{}, valid.Rows...), StatementRow{ID: "bs-140", RowNo: 140, RowType: "account", Title: "สินทรัพย์หมุนเวียนอื่น", AccountCodes: []string{"1900"}})
	f.failCode(Command{Resource: "statement-templates", Action: "update", ID: r.ID, Version: r.Version, Master: &next}, "statement_template_row_accounts_invalid", "rows[2].accountcodes")
	next.Rows[2].AccountCodes = []string{"1910"}
	f.run(Command{Resource: "statement-templates", Action: "update", ID: r.ID, Version: r.Version, Master: &next})
	if n := f.countRows(`SELECT count(*) FROM gl_records WHERE company='C' AND kind='statement-templates' AND payload->'rows'->2->'accountcodes' ? '1910'`); n != 1 {
		t.Fatalf("valid update not saved: %d", n)
	}
}

// ลบบัญชี/เปลี่ยนเป็นบัญชีหัวข้อ: กันเฉพาะรหัสในตำแหน่งที่เครื่องคำนวณงบอ่าน (รวมคอลัมน์งบส่วนของผู้ถือหุ้น)
func TestStatementTemplateAccountGuardsIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	column := f.suggestAccountFixture("3600", "equity", "", "ส่วนเกินมูลค่าหุ้น", true, nil)
	hidden := f.suggestAccountFixture("1500", "asset", "", "เงินให้กู้ยืมระยะสั้น", true, nil)
	inRow := f.suggestAccountFixture("1400", "asset", "", "ลูกหนี้อื่น", true, nil)
	hiddenFlip := f.suggestAccountFixture("1300", "asset", "", "ค่าใช้จ่ายจ่ายล่วงหน้า", true, nil)
	movement := f.suggestAccountFixture("3700", "equity", "", "ทุนที่ออกระหว่างงวด", true, nil)
	f.run(Command{Resource: "statement-templates", Action: "create", Master: &Master{Code: "EQ-01", Name: "งบแสดงการเปลี่ยนแปลงส่วนของผู้ถือหุ้น", StatementType: "equity", IsActive: true,
		Rows: []StatementRow{{ID: "eq-1", RowNo: 10, RowType: "account", Title: "ยอดต้นงวด", AmountBasis: "opening", AccountCodes: []string{"1500"}},
			// ไม่มี amountbasis = ความเคลื่อนไหวที่เครื่องคำนวณอ่านจากรหัสของบรรทัดเอง
			{ID: "eq-2", RowNo: 20, RowType: "account", Title: "เพิ่มทุนระหว่างงวด", AccountCodes: []string{"3700"}}},
		Columns: []StatementColumn{{ID: "eq-c2", Title: "ส่วนเกินมูลค่าหุ้น", AccountCodes: []string{"3600"}}}}})
	f.run(Command{Resource: "statement-templates", Action: "create", Master: suggestBalanceSheet("BS-01",
		StatementRow{ID: "bs-10", RowNo: 10, RowType: "header", Title: "สินทรัพย์หมุนเวียน", AccountCodes: []string{"1500", "1300"}},
		StatementRow{ID: "bs-60", RowNo: 60, RowType: "account", Title: "ลูกหนี้หมุนเวียนอื่น", AccountCodes: []string{"1400"}},
		StatementRow{ID: "bs-70", RowNo: 70, RowType: "formula", Title: "รวม", Formula: "R60", AccountCodes: []string{"1300"}})})

	f.failCode(Command{Resource: "accounts", Action: "delete", ID: column.ID, Version: column.Version}, CodeReferenced, "")
	f.failCode(Command{Resource: "accounts", Action: "delete", ID: movement.ID, Version: movement.Version}, CodeReferenced, "")
	movementHeader := movement
	movementHeader.AllowPosting = false
	if user := f.failCode(Command{Resource: "accounts", Action: "update", ID: movement.ID, Version: movement.Version, Account: &movementHeader}, "account_in_statement_template", "allowposting"); !reflect.DeepEqual(user.Args, []string{"EQ-01"}) {
		t.Fatalf("equity movement row flip error = %+v", user)
	}
	f.run(Command{Resource: "accounts", Action: "delete", ID: hidden.ID, Version: hidden.Version})
	if n := f.countRows(`SELECT count(*) FROM gl_records WHERE company='C' AND kind='accounts' AND code='1500' AND (payload->>'isdeleted')::boolean`); n != 1 {
		t.Fatal("hidden-code account not deleted")
	}

	flip := inRow
	flip.AllowPosting = false
	user := f.failCode(Command{Resource: "accounts", Action: "update", ID: inRow.ID, Version: inRow.Version, Account: &flip}, "account_in_statement_template", "allowposting")
	if user.HTTPStatus() != 409 || !reflect.DeepEqual(user.Args, []string{"BS-01"}) {
		t.Fatalf("flip error = %+v", user)
	}
	renamed := inRow
	renamed.Names = []Name{{Code: "th", Name: "ลูกหนี้อื่น-พนักงาน"}}
	r := f.run(Command{Resource: "accounts", Action: "update", ID: inRow.ID, Version: inRow.Version, Account: &renamed})
	renamed.IsActive = false
	r = f.run(Command{Resource: "accounts", Action: "update", ID: inRow.ID, Version: r.Version, Account: &renamed})
	// บัญชีในแถว account ของงบฐานะการเงิน (bs-60) ลบไม่ได้ — ต่อจาก SQL statementTemplateReadsAccount ใน deletePGAccountGuard
	f.failCode(Command{Resource: "accounts", Action: "delete", ID: inRow.ID, Version: r.Version}, CodeReferenced, "")

	toHeader := hiddenFlip
	toHeader.AllowPosting = false
	f.run(Command{Resource: "accounts", Action: "update", ID: hiddenFlip.ID, Version: hiddenFlip.Version, Account: &toHeader})
	if n := f.countRows(`SELECT count(*) FROM gl_records WHERE company='C' AND kind='accounts' AND code='1300' AND NOT (payload->>'allowposting')::boolean`); n != 1 {
		t.Fatal("hidden-code account not flipped to header")
	}
	if n := f.countRows(`SELECT count(*) FROM gl_records WHERE company='C' AND kind='accounts' AND code='1400' AND (payload->>'allowposting')::boolean AND NOT (payload->>'isactive')::boolean`); n != 1 {
		t.Fatal("account in a template row must stay postable while name/isactive updates succeed")
	}
}

// แนะนำบัญชีจากข้อมูลที่บันทึกไว้จริงใน PostgreSQL และไม่เขียนอะไรลงฐานข้อมูล
func TestStatementTemplateSuggestIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.suggestAccountFixture("1110", "asset", "", "เงินสดในมือ", true, func(a *Account) { a.IsCash = true })
	f.suggestAccountFixture("1120", "asset", "", "เงินฝากออมทรัพย์", true, nil)
	f.suggestAccountFixture("1200", "asset", "", "ลูกหนี้การค้า", true, nil)
	f.suggestAccountFixture("1300", "asset", "", "สินค้าคงเหลือ-ปูนซีเมนต์", true, nil)
	f.suggestAccountFixture("1900", "asset", "", "สินทรัพย์หมุนเวียนอื่น", false, nil)
	f.suggestAccountFixture("1910", "asset", "1900", "เงินมัดจำจ่าย", true, nil)
	f.suggestAccountFixture("1920", "asset", "1900", "เงินทดรองจ่าย", true, nil)
	f.suggestAccountFixture("2000", "liability", "", "เจ้าหนี้การค้า", true, nil)

	sale := Journal{DocNo: "IV6907-001", Date: "2026-07-05", BookCode: "JV", FiscalYear: "2026", Kind: "manual", Description: "ขายวัสดุก่อสร้าง ให้ลูกค้าโครงการก่อสร้าง A เป็นเงินเชื่อ", BranchCode: "B1",
		Lines: []Line{{AccountCode: "1200", Debit: "10700", Credit: "0"}, {AccountCode: "4000", Debit: "0", Credit: "10700"}},
		Details: &JournalDetails{
			Partners:     []SubledgerPartner{{Code: "C101", Name: "บริษัท ก่อสร้างไทย จำกัด", IsCustomer: true, IsActive: true}},
			Documents:    []SubledgerDocument{{ID: "AR-D1", Ledger: "ar", PartnerCode: "C101", DocumentNo: "IV6907-001", Date: "2026-07-05", BranchCode: "B1", Kind: 1, Side: 1, Amount: "10700", ControlAccountCode: "1200"}},
			Allocations:  []SubledgerAllocation{{ID: "AR-A1", Ledger: "ar", DocumentID: "AR-D1", LineNumber: 1, Amount: "10700"}},
			BankAccounts: []SubledgerBankAccount{{Code: "KBANK-01", BankName: "ธนาคารกสิกรไทย", AccountNumber: "123-4-56789-0", AccountName: "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด", GLAccountCode: "1120", IsActive: true}},
		}}
	f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &sale}).ID))
	purchase := Journal{DocNo: "RR6907-001", Date: "2026-07-06", BookCode: "JV", FiscalYear: "2026", Kind: "manual", Description: "ซื้อปูนซีเมนต์ปอร์ตแลนด์ 50 กก. เป็นเงินเชื่อ", BranchCode: "B1",
		Lines: []Line{{AccountCode: "5000", Debit: "5350", Credit: "0"}, {AccountCode: "2000", Debit: "0", Credit: "5350"}},
		Details: &JournalDetails{
			Partners:    []SubledgerPartner{{Code: "S101", Name: "บริษัท ปูนไทย จำกัด", IsSupplier: true, IsActive: true}},
			Documents:   []SubledgerDocument{{ID: "AP-D1", Ledger: "ap", PartnerCode: "S101", DocumentNo: "INV-8812", Date: "2026-07-06", BranchCode: "B1", Kind: 1, Side: 1, Amount: "5350", ControlAccountCode: "2000"}},
			Allocations: []SubledgerAllocation{{ID: "AP-A1", Ledger: "ap", DocumentID: "AP-D1", LineNumber: 2, Amount: "5350"}},
		}}
	f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &purchase}).ID))
	f.run(Command{Resource: "product-account-groups", Action: "create", Master: &Master{Code: "PG-CEMENT", Name: "ปูนซีเมนต์", IsActive: true, ItemAccount: "1300", RevenueAccount: "4000", CostAccount: "5000"}})
	f.run(Command{Resource: "statement-templates", Action: "create", Master: suggestBalanceSheet("BS-OTHER",
		StatementRow{ID: "bs-50", RowNo: 50, RowType: "account", Title: "ลูกหนี้การค้า", AccountCodes: []string{"1200"}, SuggestKey: "trade_receivables"})})
	current := f.run(Command{Resource: "statement-templates", Action: "create", Master: suggestBalanceSheet("BS-MAIN",
		StatementRow{ID: "bs-50", RowNo: 50, RowType: "account", Title: "ลูกหนี้การค้า", AccountCodes: []string{"1200"}, SuggestKey: "trade_receivables"})})
	if n := f.countRows(`SELECT count(*) FROM gl_subledger_bank_accounts WHERE company='C' AND payload->>'gl_account_code'='1120'`); n != 1 {
		t.Fatalf("bank account evidence not recorded: %d", n)
	}
	if n := f.countRows(`SELECT count(*) FROM gl_subledger_documents WHERE company='C'`); n != 2 {
		t.Fatalf("AR/AP documents not recorded: %d", n)
	}
	if n := f.countRows(`SELECT count(*) FROM (SELECT to_regclass('fa_records') AS t) x WHERE t IS NULL`); n != 1 {
		t.Fatal("test premise: this schema must have no fa_records table")
	}

	counts := func() [3]int {
		return [3]int{f.countRows(`SELECT count(*) FROM gl_events`), f.countRows(`SELECT count(*) FROM gl_records`), f.countRows(`SELECT count(*) FROM gl_lines`)}
	}
	before := counts()
	// ส่งแม่แบบที่กำลังแก้จากจอ (ยังไม่บันทึก): แถว bs-50 ว่างแล้ว, มีบัญชีหัวข้อและรหัสที่ไม่มีในผัง
	request := suggestBalanceSheet("BS-MAIN",
		StatementRow{ID: "bs-30", RowNo: 30, RowType: "account", Title: "เงินสดและรายการเทียบเท่าเงินสด", SuggestKey: "cash_and_equivalents"},
		StatementRow{ID: "bs-50", RowNo: 50, RowType: "account", Title: "ลูกหนี้การค้า", SuggestKey: "trade_receivables"},
		StatementRow{ID: "bs-80", RowNo: 80, RowType: "account", Title: "สินค้าคงเหลือ", SuggestKey: "inventories"},
		StatementRow{ID: "bs-140", RowNo: 140, RowType: "account", Title: "สินทรัพย์หมุนเวียนอื่น", AccountCodes: []string{"1120", "1910"}},
		StatementRow{ID: "bs-230", RowNo: 230, RowType: "account", Title: "ที่ดิน อาคารและอุปกรณ์", SuggestKey: "property_plant_equipment"},
		StatementRow{ID: "bs-350", RowNo: 350, RowType: "account", Title: "เจ้าหนี้การค้า", SuggestKey: "trade_payables"},
		StatementRow{ID: "bs-650", RowNo: 650, RowType: "account", Title: "กำไรสะสม - ยังไม่ได้จัดสรร", SuggestKey: "retained_earnings"},
		StatementRow{ID: "bs-900", RowNo: 900, RowType: "account", Title: "สินทรัพย์อื่น", AccountCodes: []string{"1900", "ZZZ"}})
	result := f.run(Command{Resource: "statement-templates", Action: "suggest", ID: current.ID, Master: request})
	if after := counts(); after != before {
		t.Fatalf("suggest wrote to the database: before %v after %v", before, after)
	}
	if result.Suggestions == nil || len(result.Suggestions.Targets) != 6 {
		t.Fatalf("suggestions = %+v", result.Suggestions)
	}
	byID := map[string]StatementSuggestionTarget{}
	for _, target := range result.Suggestions.Targets {
		byID[target.ID] = target
	}
	reason := func(source string, count int, templates ...string) StatementSuggestionReason {
		return StatementSuggestionReason{Source: source, Count: count, Templates: templates}
	}
	wantAccounts := map[string][]StatementSuggestedAccount{
		"bs-30":  {{AccountCode: "1110", AccountName: "เงินสดในมือ", IsActive: true, Reasons: []StatementSuggestionReason{reason(suggestIsCash, 1)}}},
		"bs-50":  {{AccountCode: "1200", AccountName: "ลูกหนี้การค้า", IsActive: true, Reasons: []StatementSuggestionReason{reason(suggestAR, 1), reason(suggestOtherTemplates, 1, "BS-OTHER")}}},
		"bs-80":  {{AccountCode: "1300", AccountName: "สินค้าคงเหลือ-ปูนซีเมนต์", IsActive: true, Reasons: []StatementSuggestionReason{reason(suggestProductItem, 1)}}},
		"bs-230": {},
		"bs-350": {{AccountCode: "2000", AccountName: "เจ้าหนี้การค้า", IsActive: true, Reasons: []StatementSuggestionReason{reason(suggestAP, 1)}}},
		"bs-650": {{AccountCode: "3100", AccountName: "3100", IsActive: true, Reasons: []StatementSuggestionReason{reason(suggestFiscalRetained, 2)}}},
	}
	for id, want := range wantAccounts {
		if got := byID[id].Accounts; !reflect.DeepEqual(got, want) {
			t.Fatalf("%s accounts = %+v, want %+v", id, got, want)
		}
	}
	wantInUse := []StatementAccountInUse{{AccountCode: "1120", AccountName: "เงินฝากออมทรัพย์", Target: "row", ID: "bs-140", RowNo: 140, Title: "สินทรัพย์หมุนเวียนอื่น", Reasons: []StatementSuggestionReason{reason(suggestBank, 1)}}}
	if got := byID["bs-30"].InUse; !reflect.DeepEqual(got, wantInUse) {
		t.Fatalf("bs-30 inuse = %+v", got)
	}
	wantFixes := []StatementAccountFix{{Target: "row", ID: "bs-900", RowNo: 900, Title: "สินทรัพย์อื่น", Removed: []string{"ZZZ"}, Headers: []StatementHeaderExpansion{{
		AccountCode: "1900", AccountName: "สินทรัพย์หมุนเวียนอื่น", Descendants: []string{"1920"},
		Skipped: []StatementAccountInUse{{AccountCode: "1910", AccountName: "เงินมัดจำจ่าย", Target: "row", ID: "bs-140", RowNo: 140, Title: "สินทรัพย์หมุนเวียนอื่น"}}}}}}
	if !reflect.DeepEqual(result.Suggestions.Fixes, wantFixes) {
		t.Fatalf("fixes = %+v", result.Suggestions.Fixes)
	}

	pnl := &Master{Code: "PL-NEW", Name: "งบกำไรขาดทุน", StatementType: "pnl", IsActive: true, Rows: []StatementRow{
		{ID: "pnl-10", RowNo: 10, RowType: "account", Title: "รายได้จากการขาย", SuggestKey: "sales_revenue"},
		{ID: "pnl-20", RowNo: 20, RowType: "account", Title: "ต้นทุนขาย", SuggestKey: "cost_of_sales"},
	}}
	result = f.run(Command{Resource: "statement-templates", Action: "suggest", Master: pnl})
	if got := result.Suggestions.Targets; len(got) != 2 || len(got[0].Accounts) != 1 || got[0].Accounts[0].AccountCode != "4000" || got[0].Accounts[0].Reasons[0].Source != suggestProductRevenue ||
		len(got[1].Accounts) != 1 || got[1].Accounts[0].AccountCode != "5000" || got[1].Accounts[0].Reasons[0].Source != suggestProductCost || len(result.Suggestions.Fixes) != 0 {
		t.Fatalf("pnl suggestions = %+v", got)
	}
	f.failCode(Command{Resource: "statement-templates", Action: "suggest"}, "statement_template_suggest_payload_required", "master")
	if after := counts(); after != before {
		t.Fatalf("suggest wrote to the database: before %v after %v", before, after)
	}
}

// ผังบัญชีและหลักฐานกรองตามบริษัท: บริษัท D ใช้รหัสเดียวกันและมีหลักฐานครบ แต่บริษัท C ต้องไม่ได้อะไรจาก D
func TestStatementSuggestCompanyIsolationIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.suggestAccountFixture("1300", "asset", "", "สินค้าคงเหลือ", true, nil)
	other := f.scope
	other.Company = "D"
	runD := func(cmd Command) {
		t.Helper()
		if _, err := f.execute(other, cmd); err != nil {
			t.Fatalf("company D %s/%s failed: %v", cmd.Resource, cmd.Action, err)
		}
	}
	for _, a := range []Account{
		{AccountCode: "1300", AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, IsCash: true, Names: []Name{{Code: "th", Name: "เงินสดในมือ"}}},
		{AccountCode: "1350", AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "สินค้าคงเหลือ-เหล็กเส้น"}}},
		{AccountCode: "4000", AccountType: "income", NormalBalance: "credit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "รายได้จากการขาย"}}},
		{AccountCode: "5000", AccountType: "expense", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "ต้นทุนขาย"}}},
	} {
		account := a
		runD(Command{Resource: "accounts", Action: "create", Account: &account})
	}
	runD(Command{Resource: "product-account-groups", Action: "create", Master: &Master{Code: "PG-STEEL", Name: "เหล็กเส้น", IsActive: true, ItemAccount: "1350", RevenueAccount: "4000", CostAccount: "5000"}})
	runD(Command{Resource: "product-account-groups", Action: "create", Master: &Master{Code: "PG-CEMENT", Name: "ปูนซีเมนต์", IsActive: true, ItemAccount: "1300", RevenueAccount: "4000", CostAccount: "5000"}})
	runD(Command{Resource: "statement-templates", Action: "create", Master: suggestBalanceSheet("BS-D",
		StatementRow{ID: "bs-80", RowNo: 80, RowType: "account", Title: "สินค้าคงเหลือ", AccountCodes: []string{"1300"}, SuggestKey: "inventories"})})

	request := func() *Master {
		return suggestBalanceSheet("BS-NEW",
			StatementRow{ID: "bs-30", RowNo: 30, RowType: "account", Title: "เงินสดและรายการเทียบเท่าเงินสด", SuggestKey: "cash_and_equivalents"},
			StatementRow{ID: "bs-80", RowNo: 80, RowType: "account", Title: "สินค้าคงเหลือ", SuggestKey: "inventories"})
	}
	// premise: บริษัท D เองได้คำแนะนำจากหลักฐานของตัวเอง
	own, err := f.execute(other, Command{Resource: "statement-templates", Action: "suggest", Master: request()})
	if err != nil || own.Suggestions == nil || len(own.Suggestions.Targets) != 2 || len(own.Suggestions.Targets[0].Accounts) != 1 || len(own.Suggestions.Targets[1].Accounts) != 2 {
		t.Fatalf("company D suggestions = %+v, %v", own.Suggestions, err)
	}
	result := f.run(Command{Resource: "statement-templates", Action: "suggest", Master: request()})
	if result.Suggestions == nil || len(result.Suggestions.Targets) != 2 {
		t.Fatalf("suggestions = %+v", result.Suggestions)
	}
	for _, target := range result.Suggestions.Targets {
		if len(target.Accounts) != 0 || len(target.InUse) != 0 {
			t.Fatalf("company C got company D evidence or chart for %s: %+v", target.ID, target)
		}
	}
}

// บัญชีที่มียอดแต่ไม่อยู่ในบรรทัดใด: งบฐานะการเงิน/งบกำไรขาดทุน ปีนี้ + ปีก่อน, ชุดงบ และหลังปิดบัญชีสิ้นปี
func TestStatementUnassignedIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	post := func(doc, date, year, debitAccount, creditAccount, amount string) {
		j := Journal{DocNo: doc, Date: date, BookCode: "JV", FiscalYear: year, Description: "รายการทดสอบบัญชีที่ไม่อยู่ในงบ", Kind: "manual", BranchCode: "B1",
			Lines: []Line{{AccountCode: debitAccount, Debit: Amount(amount), Credit: Amount("0")}, {AccountCode: creditAccount, Debit: Amount("0"), Credit: Amount(amount)}}}
		f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &j}).ID))
	}
	post("JV26-1", "2026-01-05", "2026", "1000", "3000", "500")
	post("JV26-2", "2026-02-10", "2026", "1000", "4000", "100")
	post("JV26-3", "2026-03-10", "2026", "5000", "1000", "30.25")
	create := func(m *Master) { f.run(Command{Resource: "statement-templates", Action: "create", Master: m}) }
	create(suggestBalanceSheet("FS-BS1", StatementRow{ID: "a", RowNo: 10, RowType: "account", Title: "เงินสด", AccountCodes: []string{"1000"}}))
	create(&Master{Code: "FS-PL1", Name: "งบกำไรขาดทุน", StatementType: "pnl", IsActive: true, GlobalStyle: &StatementGlobalStyle{Scale: 2},
		Rows: []StatementRow{{ID: "a", RowNo: 10, RowType: "account", Title: "รายได้จากการขาย", AccountCodes: []string{"4000"}}}})
	report := func(template, year string) Report {
		t.Helper()
		r, err := f.store.Report(f.ctx, f.scope, "statement", ReportQuery{FiscalYear: year, From: year + "-01-01", To: year + "-12-31", Template: template})
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	item := func(key, year, code, name, accountType, basis, amount string) StatementUnassigned {
		return StatementUnassigned{Key: key, FiscalYear: year, AccountCode: code, AccountName: name, AccountType: accountType, Basis: basis, Amount: amount}
	}
	bsWant := []StatementUnassigned{item("amount", "2026", "3000", "3000", "equity", "closing", "500.00"), item("amount", "2026", statementCurrentEarnings, "", "equity", "closing", "69.75")}
	if got := report("FS-BS1", "2026").Unassigned; !reflect.DeepEqual(got, bsWant) {
		t.Fatalf("balance sheet unassigned = %+v", got)
	}
	plWant := []StatementUnassigned{item("amount", "2026", "5000", "5000", "expense", "movement", "30.25")}
	if got := report("FS-PL1", "2026").Unassigned; !reflect.DeepEqual(got, plWant) {
		t.Fatalf("pnl unassigned = %+v", got)
	}
	set, err := f.store.pg.StatementSet(f.ctx, f.scope, StatementSetQuery{Report: ReportQuery{FiscalYear: "2026"}, Templates: []string{"FS-BS1", "FS-PL1"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, section := range set.Sections {
		want := map[string][]StatementUnassigned{"FS-BS1": bsWant, "FS-PL1": plWant}[section.Code]
		if section.Report == nil || !reflect.DeepEqual(section.Report.Unassigned, want) {
			t.Fatalf("statement set %s unassigned = %+v", section.Code, section.Report)
		}
	}
	// ตัวกรองสาขา/แผนก/โครงการของรายงานใช้กับยอดที่หลุดงบด้วย: รายการของสาขา B2 ไม่ปนในรายงานสาขา B1 และกลับกัน
	b2 := Journal{DocNo: "JV26-B2", Date: "2026-04-01", BookCode: "JV", FiscalYear: "2026", Description: "รับเงินเพิ่มทุน สาขาบางนา", Kind: "manual", BranchCode: "B2",
		Lines: []Line{{AccountCode: "1000", Debit: Amount("40"), Credit: Amount("0"), DepartmentCode: "D2", ProjectCode: "P2"}, {AccountCode: "3000", Debit: Amount("0"), Credit: Amount("40"), DepartmentCode: "D2", ProjectCode: "P2"}}}
	f.post(f.journal(f.run(Command{Resource: "journals", Action: "create", Journal: &b2}).ID))
	filtered := func(q ReportQuery) []StatementUnassigned {
		t.Helper()
		q.FiscalYear, q.From, q.To, q.Template = "2026", "2026-01-01", "2026-12-31", "FS-BS1"
		r, err := f.store.Report(f.ctx, f.scope, "statement", q)
		if err != nil {
			t.Fatal(err)
		}
		return r.Unassigned
	}
	if got := filtered(ReportQuery{BranchCode: "B1"}); !reflect.DeepEqual(got, bsWant) {
		t.Fatalf("branch B1 unassigned = %+v", got)
	}
	b2Want := []StatementUnassigned{item("amount", "2026", "3000", "3000", "equity", "closing", "40.00")}
	for _, q := range []ReportQuery{{BranchCode: "B2"}, {DepartmentCode: "D2"}, {ProjectCode: "P2"}} {
		if got := filtered(q); !reflect.DeepEqual(got, b2Want) {
			t.Fatalf("%+v unassigned = %+v", q, got)
		}
	}
	if got := filtered(ReportQuery{ProjectCode: "P9"}); len(got) != 0 {
		t.Fatalf("project without entries unassigned = %+v", got)
	}

	year := f.years["2026"]
	f.post(f.journal(f.run(Command{Resource: "processes", Action: "close", ID: "2026", Version: year.Version, DocNo: "CLOSE26", Date: "2026-12-31", Reason: "ปิดบัญชีสิ้นปี"}).ID))
	opening := f.journal(f.run(Command{Resource: "processes", Action: "year-end", ID: "2026", Version: year.Version, TargetYear: "2027", DocNo: "OPEN27", Date: "2027-01-01", Reason: "ยกยอดไปปีถัดไป"}).ID)
	if opening.Status == "draft" {
		f.post(opening)
	}
	post("JV27-1", "2027-03-10", "2027", "1000", "4000", "200")
	post("JV27-2", "2027-06-10", "2027", "5000", "1000", "50")
	compare := &StatementGlobalStyle{Scale: 2, ComparisonType: "previous_year"}
	create(&Master{Code: "FS-BS2", Name: "งบแสดงฐานะการเงิน", StatementType: "balance_sheet", IsActive: true, GlobalStyle: compare,
		Rows: []StatementRow{{ID: "a", RowNo: 10, RowType: "account", Title: "เงินสด", AccountCodes: []string{"1000"}}, {ID: "b", RowNo: 20, RowType: "account", Title: "ทุน", AccountCodes: []string{"3000", statementCurrentEarnings}}}})
	// หลังปิดบัญชี: กำไรปี 2026 อยู่ในกำไรสะสม (3100) ไม่มีรายการรายได้/ค่าใช้จ่ายหรือบัญชีกำไรขาดทุน (3200 ยอดศูนย์)
	want := []StatementUnassigned{item("amount", "2027", "3100", "3100", "equity", "closing", "69.75"), item("prioramount", "2026", "3100", "3100", "equity", "closing", "69.75")}
	if got := report("FS-BS2", "2027").Unassigned; !reflect.DeepEqual(got, want) {
		t.Fatalf("after year-end = %+v", got)
	}
	create(&Master{Code: "FS-BS3", Name: "งบแสดงฐานะการเงิน", StatementType: "balance_sheet", IsActive: true, GlobalStyle: compare,
		Rows: []StatementRow{{ID: "a", RowNo: 10, RowType: "account", Title: "เงินสด", AccountCodes: []string{"1000"}}, {ID: "b", RowNo: 20, RowType: "account", Title: "ส่วนของผู้ถือหุ้น", AccountCodes: []string{"3000", "3100", statementCurrentEarnings}}}})
	if got := report("FS-BS3", "2027").Unassigned; len(got) != 0 {
		t.Fatalf("fully assigned balance sheet = %+v", got)
	}
	// งบกำไรขาดทุนปีที่ปิดแล้ว: ความเคลื่อนไหวไม่นับรายการปิดบัญชี ค่าใช้จ่ายยังไม่อยู่ในงบ
	if got := report("FS-PL1", "2026").Unassigned; !reflect.DeepEqual(got, plWant) {
		t.Fatalf("closed-year pnl unassigned = %+v", got)
	}
}
