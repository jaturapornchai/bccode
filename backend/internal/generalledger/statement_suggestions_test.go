package generalledger

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func suggestAccount(code, accountType, parent string, posting bool) Account {
	return Account{AccountCode: code, AccountType: accountType, ParentAccountCode: parent, AllowPosting: posting, IsActive: true, Names: []Name{{Code: "th", Name: "บัญชี " + code}}}
}

func suggestChart(accounts ...Account) map[string]Account {
	chart := map[string]Account{}
	for _, account := range accounts {
		chart[account.AccountCode] = account
	}
	return chart
}

func suggestEvidence(source string, counts map[string]int) statementEvidence {
	ev := statementEvidence{}
	for code, count := range counts {
		ev.add(source, code, count)
	}
	return ev
}

func findSuggestTarget(t *testing.T, out StatementSuggestions, id string) StatementSuggestionTarget {
	t.Helper()
	for _, target := range out.Targets {
		if target.ID == id {
			return target
		}
	}
	t.Fatalf("target %s not in %+v", id, out.Targets)
	return StatementSuggestionTarget{}
}

func suggestedCodes(accounts []StatementSuggestedAccount) []string {
	codes := []string{}
	for _, account := range accounts {
		codes = append(codes, account.AccountCode)
	}
	return codes
}

func TestStatementSuggestRulesMatchKeys(t *testing.T) {
	if len(statementSuggestRules) != len(statementSuggestKeys) {
		t.Fatalf("rules %d != keys %d", len(statementSuggestRules), len(statementSuggestKeys))
	}
	for _, key := range statementSuggestKeys {
		rule, ok := statementSuggestRules[key]
		if !ok {
			t.Fatalf("key %s has no rule", key)
		}
		if !contains([]string{"asset", "liability", "equity", "income", "expense"}, rule.accountType) {
			t.Fatalf("%s: account type %q", key, rule.accountType)
		}
		recorded := 0
		for _, source := range rule.sources {
			if !contains(statementSuggestSourceOrder, source) {
				t.Fatalf("%s: unknown source %s", key, source)
			}
			if source != suggestOtherTemplates {
				recorded++
			}
		}
		if recorded == 0 {
			t.Fatalf("%s: needs at least one recorded-data source besides other templates", key)
		}
	}
	if statementSuggestSourceOrder[len(statementSuggestSourceOrder)-1] != suggestOtherTemplates {
		t.Fatal("other_templates must be the last reason")
	}
}

func TestStatementCodeTargetsEngineRead(t *testing.T) {
	bs := Master{StatementType: "balance_sheet", Rows: []StatementRow{
		{ID: "h", RowNo: 10, RowType: "header", Title: "หัวข้อ", AccountCodes: []string{"1000"}},
		{ID: "a", RowNo: 20, RowType: "account", Title: "เงินสด", AccountCodes: []string{"1100"}, SuggestKey: "cash_and_equivalents"},
		{ID: "o", RowNo: 30, RowType: "account", Title: "ต้นงวด", AmountBasis: "opening"},
		{ID: "m", RowNo: 40, RowType: "account", Title: "เคลื่อนไหว", AmountBasis: "movement"},
		{ID: "f", RowNo: 50, RowType: "formula", Title: "รวม", Formula: "R20", AccountCodes: []string{"1200"}},
	}, Columns: []StatementColumn{{ID: "c", Title: "ไม่ใช้", AccountCodes: []string{"3000"}}}}
	got := statementCodeTargets(bs)
	if len(got) != 3 || got[0].id != "a" || got[0].basis != "closing" || got[0].index != 1 || got[0].suggestKey != "cash_and_equivalents" || got[1].basis != "opening" || got[2].basis != "movement" {
		t.Fatalf("balance sheet targets = %+v", got)
	}
	pnl := Master{StatementType: "pnl", Rows: []StatementRow{{ID: "r", RowNo: 10, RowType: "account"}, {ID: "c", RowNo: 20, RowType: "account", AmountBasis: "closing"}}}
	if got = statementCodeTargets(pnl); len(got) != 2 || got[0].basis != "movement" || got[1].basis != "closing" {
		t.Fatalf("pnl targets = %+v", got)
	}
	custom := Master{StatementType: "custom", Rows: []StatementRow{{ID: "r", RowNo: 10, RowType: "account"}}}
	if got = statementCodeTargets(custom); len(got) != 1 || got[0].basis != "closing" {
		t.Fatalf("custom targets = %+v", got)
	}
	equity := Master{StatementType: "equity", Rows: []StatementRow{
		{ID: "open", RowNo: 10, RowType: "account", AmountBasis: "opening", AccountCodes: []string{"3000"}},
		{ID: "close", RowNo: 20, RowType: "account", AmountBasis: "closing"},
		{ID: "other", RowNo: 30, RowType: "account", AmountBasis: "other"},
		{ID: "move", RowNo: 40, RowType: "account", AccountCodes: []string{"3100"}},
		{ID: "move2", RowNo: 50, RowType: "account", AmountBasis: "movement"},
	}, Columns: []StatementColumn{{ID: "c1", Title: "ทุน", AccountCodes: []string{"3000"}}, {ID: "c4", Title: "กำไรสะสม", SuggestKey: "retained_earnings"}}}
	got = statementCodeTargets(equity)
	if len(got) != 4 || got[0].id != "move" || got[0].basis != "movement" || got[1].id != "move2" || got[2].target != "column" || got[2].basis != "column" || got[3].id != "c4" || got[3].index != 1 || got[3].suggestKey != "retained_earnings" || got[3].title != "กำไรสะสม" {
		t.Fatalf("equity targets = %+v", got)
	}
}

func TestBuildStatementSuggestionsBalanceSheet(t *testing.T) {
	inactiveCash := suggestAccount("1120", "asset", "1000", true)
	inactiveCash.IsCash, inactiveCash.IsActive = true, false
	cash := suggestAccount("1100", "asset", "1000", true)
	cash.IsCash = true
	chart := suggestChart(suggestAccount("1000", "asset", "", false), cash, suggestAccount("1110", "asset", "1000", true), inactiveCash,
		suggestAccount("1130", "asset", "1000", true), suggestAccount("2100", "liability", "", true), suggestAccount("1900", "asset", "", false))
	// หลักฐานผิดหมวด (2100), บัญชีหัวข้อ (1900), บัญชีที่ลบแล้ว/ไม่มีในผัง (1999) ต้องไม่ถูกแนะนำ
	ev := suggestEvidence(suggestBank, map[string]int{"1110": 2, "1130": 1, "2100": 1, "1900": 1, "1999": 1})
	m := Master{StatementType: "balance_sheet", Rows: []StatementRow{
		{ID: "bs-30", RowNo: 30, RowType: "account", Title: "เงินสดและรายการเทียบเท่าเงินสด", AccountCodes: []string{"1100"}, SuggestKey: "cash_and_equivalents"},
		{ID: "bs-140", RowNo: 140, RowType: "account", Title: "สินทรัพย์หมุนเวียนอื่น", AccountCodes: []string{"1110"}},
		{ID: "bs-open", RowNo: 150, RowType: "account", Title: "ยอดต้นงวด", AccountCodes: []string{"1130"}, AmountBasis: "opening"},
		{ID: "bs-x", RowNo: 160, RowType: "account", Title: "ชนิดที่ไม่รู้จัก", SuggestKey: "bogus"},
		{ID: "bs-h", RowNo: 170, RowType: "header", Title: "หัวข้อ", SuggestKey: "trade_receivables"},
	}}
	out, tooLarge := buildStatementSuggestions(m, chart, ev)
	if tooLarge || len(out.Targets) != 1 {
		t.Fatalf("targets = %+v tooLarge=%v (unknown key and non-account rows must be ignored)", out.Targets, tooLarge)
	}
	target := findSuggestTarget(t, out, "bs-30")
	if target.Target != "row" || target.RowNo != 30 || target.AccountType != "asset" || target.SuggestKey != "cash_and_equivalents" {
		t.Fatalf("target = %+v", target)
	}
	// 1100 อยู่ในบรรทัดนี้แล้ว; 1130 อยู่แค่บรรทัดยอดต้นงวด (คนละฐาน) จึงยังแนะนำ; 1120 ปิดใช้งานแต่ยังแนะนำพร้อมธง
	if got := suggestedCodes(target.Accounts); !reflect.DeepEqual(got, []string{"1120", "1130"}) {
		t.Fatalf("accounts = %v", got)
	}
	if a := target.Accounts[0]; a.IsActive || a.AccountName != "บัญชี 1120" || !reflect.DeepEqual(a.Reasons, []StatementSuggestionReason{{Source: suggestIsCash, Count: 1}}) {
		t.Fatalf("inactive cash = %+v", a)
	}
	if a := target.Accounts[1]; !a.IsActive || !reflect.DeepEqual(a.Reasons, []StatementSuggestionReason{{Source: suggestBank, Count: 1}}) {
		t.Fatalf("bank account = %+v", a)
	}
	// 1110 อยู่ในบรรทัดที่ไม่มีชนิด (bs-140) ฐานเดียวกัน → ไม่แนะนำซ้ำ แต่บอกว่าอยู่บรรทัดไหน
	want := []StatementAccountInUse{{AccountCode: "1110", AccountName: "บัญชี 1110", Target: "row", ID: "bs-140", RowNo: 140, Title: "สินทรัพย์หมุนเวียนอื่น", Reasons: []StatementSuggestionReason{{Source: suggestBank, Count: 2}}}}
	if !reflect.DeepEqual(target.InUse, want) {
		t.Fatalf("inuse = %+v", target.InUse)
	}
	if len(out.Fixes) != 0 {
		t.Fatalf("no header/unknown codes in engine-read rows, fixes = %+v", out.Fixes)
	}
}

func TestBuildStatementSuggestionsCashFlowAndEquity(t *testing.T) {
	chart := suggestChart(suggestAccount("1200", "asset", "", true), suggestAccount("1210", "asset", "", true),
		suggestAccount("3000", "equity", "", true), suggestAccount("3100", "equity", "", true), suggestAccount("3150", "equity", "", true), suggestAccount("3200", "equity", "", true))
	ev := suggestEvidence(suggestAR, map[string]int{"1200": 3, "1210": 1})
	ev.add(suggestFiscalRetained, "3100", 2)
	ev.add(suggestFiscalRetained, "3150", 1)
	ev.add(suggestFiscalRetained, "3200", 1)
	// งบกระแสเงินสดวิธีทางอ้อม: บัญชีเดียวกันอยู่หลายบรรทัดได้ จึงไม่มี inuse
	cash := Master{StatementType: "cash_flow", Rows: []StatementRow{
		{ID: "cf-2", RowNo: 2, RowType: "account", Title: "กำไร", AccountCodes: []string{"1200"}},
		{ID: "cf-4", RowNo: 4, RowType: "account", Title: "ลูกหนี้การค้า", SuggestKey: "trade_receivables"},
	}}
	out, _ := buildStatementSuggestions(cash, chart, ev)
	target := findSuggestTarget(t, out, "cf-4")
	if got := suggestedCodes(target.Accounts); !reflect.DeepEqual(got, []string{"1200", "1210"}) || len(target.InUse) != 0 {
		t.Fatalf("cash flow accounts=%v inuse=%+v", got, target.InUse)
	}
	if !reflect.DeepEqual(target.Accounts[0].Reasons, []StatementSuggestionReason{{Source: suggestAR, Count: 3}}) {
		t.Fatalf("ar reasons = %+v", target.Accounts[0].Reasons)
	}
	// งบส่วนของผู้ถือหุ้น: คอลัมน์ชนกับคอลัมน์เท่านั้น (แถวความเคลื่อนไหวคนละฐาน)
	equity := Master{StatementType: "equity", Rows: []StatementRow{
		{ID: "eq-r", RowNo: 10, RowType: "account", Title: "จ่ายเงินปันผล", AccountCodes: []string{"3200"}},
	}, Columns: []StatementColumn{
		{ID: "eq-c1", Title: "ทุนที่ชำระแล้ว", AccountCodes: []string{"3000", "3150"}},
		{ID: "eq-c4", Title: "กำไรสะสม", SuggestKey: "retained_earnings"},
	}}
	out, _ = buildStatementSuggestions(equity, chart, ev)
	target = findSuggestTarget(t, out, "eq-c4")
	if target.Target != "column" || target.RowNo != 0 || target.AccountType != "equity" {
		t.Fatalf("column target = %+v", target)
	}
	if got := suggestedCodes(target.Accounts); !reflect.DeepEqual(got, []string{"3100", "3200"}) {
		t.Fatalf("equity accounts = %v", got)
	}
	if len(target.InUse) != 1 || target.InUse[0].AccountCode != "3150" || target.InUse[0].Target != "column" || target.InUse[0].ID != "eq-c1" || target.InUse[0].Title != "ทุนที่ชำระแล้ว" || target.InUse[0].RowNo != 0 {
		t.Fatalf("equity inuse = %+v", target.InUse)
	}
}

func TestBuildStatementSuggestionsOtherTemplates(t *testing.T) {
	chart := suggestChart(suggestAccount("4100", "income", "4000", true), suggestAccount("4000", "income", "", false), suggestAccount("4200", "income", "", true))
	ev := statementEvidence{}
	for i := 7; i >= 1; i-- {
		ev.templates = append(ev.templates, Master{Code: fmt.Sprintf("T%d", i), StatementType: "pnl", Rows: []StatementRow{
			{ID: "pnl-10", RowNo: 10, RowType: "account", AccountCodes: []string{"4100", "4100", "4000"}, SuggestKey: "sales_revenue"},
			{ID: "pnl-90", RowNo: 90, RowType: "account", AccountCodes: []string{"4200"}},                             // ไม่มีชนิด: ไม่นับ
			{ID: "pnl-91", RowNo: 91, RowType: "header", AccountCodes: []string{"4200"}, SuggestKey: "sales_revenue"}, // เครื่องคำนวณไม่อ่าน: ไม่นับ
		}})
	}
	m := Master{StatementType: "pnl", Rows: []StatementRow{{ID: "pnl-10", RowNo: 10, RowType: "account", Title: "รายได้จากการขาย", SuggestKey: "sales_revenue"}}}
	out, _ := buildStatementSuggestions(m, chart, ev)
	target := findSuggestTarget(t, out, "pnl-10")
	want := []StatementSuggestedAccount{{AccountCode: "4100", AccountName: "บัญชี 4100", IsActive: true, Reasons: []StatementSuggestionReason{{Source: suggestOtherTemplates, Count: 7, Templates: []string{"T1", "T2", "T3", "T4", "T5"}}}}}
	if !reflect.DeepEqual(target.Accounts, want) {
		t.Fatalf("other templates = %+v", target.Accounts)
	}
}

func TestStatementAccountFixesSkipConflicts(t *testing.T) {
	chart := suggestChart(suggestAccount("1000", "asset", "", false), suggestAccount("1100", "asset", "1000", true), suggestAccount("1110", "asset", "1000", true),
		suggestAccount("1120", "asset", "1000", true), suggestAccount("1200", "asset", "1000", true), suggestAccount("1300", "asset", "", false))
	m := Master{StatementType: "balance_sheet", Rows: []StatementRow{
		{ID: "a", RowNo: 10, RowType: "account", Title: "สินทรัพย์หมุนเวียน", AccountCodes: []string{"1000", "9999", "1100", statementCurrentEarnings}},
		{ID: "b", RowNo: 20, RowType: "account", Title: "เงินฝากธนาคาร", AccountCodes: []string{"1110"}},
		{ID: "c", RowNo: 30, RowType: "account", Title: "ต้นงวด", AccountCodes: []string{"1000"}, AmountBasis: "opening"},
		{ID: "d", RowNo: 40, RowType: "account", Title: "ปกติ", AccountCodes: []string{"1200"}},
		{ID: "e", RowNo: 50, RowType: "account", Title: "หัวข้อไม่มีลูก", AccountCodes: []string{"1300"}},
		{ID: "h", RowNo: 60, RowType: "header", Title: "หัวข้อ", AccountCodes: []string{"1000", "8888"}},
	}}
	out, tooLarge := buildStatementSuggestions(m, chart, statementEvidence{})
	if tooLarge || len(out.Fixes) != 3 {
		t.Fatalf("fixes = %+v", out.Fixes)
	}
	a := out.Fixes[0]
	if a.ID != "a" || a.RowNo != 10 || a.Title != "สินทรัพย์หมุนเวียน" || !reflect.DeepEqual(a.Removed, []string{"9999"}) || len(a.Headers) != 1 {
		t.Fatalf("fix a = %+v", a)
	}
	// 1110 อยู่บรรทัด b (ฐานเดียวกัน) แล้ว → ข้าม; 1100 อยู่ในบรรทัด a เองไม่นับว่าชน
	if h := a.Headers[0]; h.AccountCode != "1000" || h.AccountName != "บัญชี 1000" || !reflect.DeepEqual(h.Descendants, []string{"1100", "1120"}) ||
		!reflect.DeepEqual(h.Skipped, []StatementAccountInUse{{AccountCode: "1110", AccountName: "บัญชี 1110", Target: "row", ID: "b", RowNo: 20, Title: "เงินฝากธนาคาร"}, {AccountCode: "1200", AccountName: "บัญชี 1200", Target: "row", ID: "d", RowNo: 40, Title: "ปกติ"}}) {
		t.Fatalf("header a = %+v", h)
	}
	// บรรทัดยอดต้นงวดคนละฐาน: ได้บัญชีย่อยครบ
	if c := out.Fixes[1]; c.ID != "c" || !reflect.DeepEqual(c.Headers[0].Descendants, []string{"1100", "1110", "1120", "1200"}) || len(c.Headers[0].Skipped) != 0 || len(c.Removed) != 0 {
		t.Fatalf("fix c = %+v", c)
	}
	// บัญชีหัวข้อที่ยังไม่มีบัญชีย่อยก็ต้องบอก; บรรทัด header ไม่ถูกแก้ (เครื่องคำนวณไม่อ่าน)
	if e := out.Fixes[2]; e.ID != "e" || len(e.Headers) != 1 || e.Headers[0].Descendants == nil || len(e.Headers[0].Descendants) != 0 || e.Removed == nil {
		t.Fatalf("fix e = %+v", e)
	}
}

func TestStatementAccountFixesSameHeaderTwice(t *testing.T) {
	chart := suggestChart(suggestAccount("1000", "asset", "", false), suggestAccount("1100", "asset", "1000", true), suggestAccount("1110", "asset", "1000", true))
	rows := []StatementRow{
		{ID: "x", RowNo: 10, RowType: "account", Title: "แรก", AccountCodes: []string{"1000"}},
		{ID: "y", RowNo: 20, RowType: "account", Title: "สอง", AccountCodes: []string{"1000"}},
	}
	out, _ := buildStatementSuggestions(Master{StatementType: "balance_sheet", Rows: rows}, chart, statementEvidence{})
	if len(out.Fixes) != 2 || !reflect.DeepEqual(out.Fixes[0].Headers[0].Descendants, []string{"1100", "1110"}) || len(out.Fixes[1].Headers[0].Descendants) != 0 || len(out.Fixes[1].Headers[0].Skipped) != 2 || out.Fixes[1].Headers[0].Skipped[0].ID != "x" {
		t.Fatalf("same-basis header twice = %+v", out.Fixes)
	}
	out, _ = buildStatementSuggestions(Master{StatementType: "cash_flow", Rows: rows}, chart, statementEvidence{})
	if len(out.Fixes) != 2 || len(out.Fixes[1].Headers[0].Descendants) != 2 || len(out.Fixes[1].Headers[0].Skipped) != 0 {
		t.Fatalf("cash flow must not skip = %+v", out.Fixes)
	}
}

func TestStatementAccountFixesTooLarge(t *testing.T) {
	accounts := []Account{suggestAccount("H", "asset", "", false)}
	for i := 0; i <= statementSuggestMaxFixCodes; i++ {
		accounts = append(accounts, suggestAccount(fmt.Sprintf("C%05d", i), "asset", "H", true))
	}
	m := Master{StatementType: "balance_sheet", Rows: []StatementRow{{ID: "a", RowNo: 10, RowType: "account", AccountCodes: []string{"H"}}}}
	if _, tooLarge := buildStatementSuggestions(m, suggestChart(accounts...), statementEvidence{}); !tooLarge {
		t.Fatal("fix output past the limit must be refused")
	}
	if !statementSuggestTooLarge(Master{Rows: make([]StatementRow, statementSuggestMaxTargets+1)}) {
		t.Fatal("too many rows accepted")
	}
	big := Master{StatementType: "balance_sheet", Rows: []StatementRow{{RowType: "account", AccountCodes: make([]string, statementSuggestMaxCodes+1)}}}
	if !statementSuggestTooLarge(big) {
		t.Fatal("too many codes accepted")
	}
	big.Rows[0].AccountCodes = big.Rows[0].AccountCodes[:statementSuggestMaxCodes]
	if statementSuggestTooLarge(big) {
		t.Fatal("limit must be inclusive")
	}
	user, ok := AsUserError(statementSuggestTooLargeError())
	if !ok || user.Code != "statement_template_suggest_too_large" || user.Field != "master" || !reflect.DeepEqual(user.Args, []string{"500", "2000"}) {
		t.Fatalf("too large error = %+v", user)
	}
}

func TestExpandStatementAccountCodes(t *testing.T) {
	accounts := []Account{
		suggestAccount("1000", "asset", "", false), suggestAccount("1001", "asset", "1000", true),
		suggestAccount("1100", "asset", "1000", false), suggestAccount("1101", "asset", "1100", true),
		suggestAccount("EMPTY", "asset", "", false),
		// วนซ้ำ X ↔ Y
		suggestAccount("X", "asset", "Y", false), suggestAccount("Y", "asset", "X", false), suggestAccount("Z", "asset", "Y", true),
		// ลึก 12 ระดับได้ ระดับ 13 ไม่นับ
		suggestAccount("D0", "asset", "", false),
	}
	parent := "D0"
	for level := 1; level <= 11; level++ {
		code := fmt.Sprintf("D%d", level)
		accounts = append(accounts, suggestAccount(code, "asset", parent, false))
		parent = code
	}
	accounts = append(accounts, suggestAccount("P12", "asset", "D11", true), suggestAccount("D12", "asset", "D11", false), suggestAccount("P13", "asset", "D12", true))
	inactive := suggestAccount("1002", "asset", "1000", true)
	inactive.IsActive = false
	accounts = append(accounts, inactive)
	tree := newStatementTree(suggestChart(accounts...))
	headers, removed := expandStatementAccountCodes(tree, []string{"UNK", "1000", statementCurrentEarnings, "1001", "UNK", "1000", "EMPTY", "X", "D0"})
	if !reflect.DeepEqual(removed, []string{"UNK"}) || len(headers) != 4 {
		t.Fatalf("headers=%+v removed=%v", headers, removed)
	}
	if headers[0].code != "1000" || headers[0].name != "บัญชี 1000" || !reflect.DeepEqual(headers[0].descendants, []string{"1001", "1002", "1101"}) {
		t.Fatalf("nested = %+v", headers[0])
	}
	if headers[1].code != "EMPTY" || headers[1].descendants == nil || len(headers[1].descendants) != 0 {
		t.Fatalf("no children = %+v", headers[1])
	}
	if headers[2].code != "X" || !reflect.DeepEqual(headers[2].descendants, []string{"Z"}) {
		t.Fatalf("cycle = %+v", headers[2])
	}
	if headers[3].code != "D0" || !reflect.DeepEqual(headers[3].descendants, []string{"P12"}) {
		t.Fatalf("depth limit = %+v", headers[3])
	}
	// memo: เพิ่มบัญชีลูกทีหลังแล้วเรียกซ้ำต้องได้ผลเดิมที่จำไว้
	tree.children["1000"] = append(tree.children["1000"], "LATE")
	tree.chart["LATE"] = suggestAccount("LATE", "asset", "1000", true)
	if again, _ := expandStatementAccountCodes(tree, []string{"1000"}); !reflect.DeepEqual(again[0].descendants, []string{"1001", "1002", "1101"}) {
		t.Fatalf("memo = %+v", again)
	}
}

func TestValidateStatementTemplateAccounts(t *testing.T) {
	deleted := suggestAccount("1999", "asset", "", true)
	deleted.IsDeleted = true
	inactive := suggestAccount("1120", "asset", "", true)
	inactive.IsActive = false
	accounts := suggestChart(suggestAccount("1000", "asset", "", false), suggestAccount("1100", "asset", "1000", true), inactive, deleted)
	ok := Master{StatementType: "balance_sheet", Rows: []StatementRow{
		{ID: "h", RowNo: 10, RowType: "header", AccountCodes: []string{"1000", "NOPE"}},
		{ID: "a", RowNo: 20, RowType: "account", AccountCodes: []string{"1100", "1120", statementCurrentEarnings}},
		{ID: "f", RowNo: 30, RowType: "formula", Formula: "R20", AccountCodes: []string{"1000"}},
	}}
	if err := validateStatementTemplateAccounts(ok, accounts); err != nil {
		t.Fatalf("valid template rejected: %v", err)
	}
	bad := ok
	bad.Rows = append(append([]StatementRow{}, ok.Rows...),
		StatementRow{ID: "b", RowNo: 40, RowType: "account", Title: "สินทรัพย์หมุนเวียน", AccountCodes: []string{"1000", "9999", "1000", "1100"}},
		StatementRow{ID: "c", RowNo: 50, RowType: "account", Title: "ลบแล้ว", AccountCodes: []string{"1999"}})
	user, isUser := AsUserError(validateStatementTemplateAccounts(bad, accounts))
	if !isUser || user.Code != "statement_template_row_accounts_invalid" || user.Field != "rows[3].accountcodes" || user.Status != 400 ||
		!reflect.DeepEqual(user.Args, []string{"40", "สินทรัพย์หมุนเวียน", "1000, 9999"}) || !strings.Contains(user.Message, "บรรทัด 40 “สินทรัพย์หมุนเวียน”: บัญชี 1000, 9999") {
		t.Fatalf("row error = %+v", user)
	}
	deletedOnly := ok
	deletedOnly.Rows = []StatementRow{{ID: "c", RowNo: 50, RowType: "account", Title: "ลบแล้ว", AccountCodes: []string{"1999"}}}
	if user, _ := AsUserError(validateStatementTemplateAccounts(deletedOnly, accounts)); user == nil || user.Args[2] != "1999" {
		t.Fatalf("deleted account accepted: %+v", user)
	}
	equity := Master{StatementType: "equity", Rows: []StatementRow{
		{ID: "o", RowNo: 10, RowType: "account", AmountBasis: "opening", AccountCodes: []string{"1000"}},
		{ID: "m", RowNo: 20, RowType: "account", AccountCodes: []string{"1100"}},
	}, Columns: []StatementColumn{{ID: "c1", Title: "ทุน", AccountCodes: []string{"1100"}}, {ID: "c2", Title: "กำไรสะสม", AccountCodes: []string{statementCurrentEarnings, "1000"}}}}
	user, isUser = AsUserError(validateStatementTemplateAccounts(equity, accounts))
	if !isUser || user.Code != "statement_template_column_accounts_invalid" || user.Field != "columns[1].accountcodes" || !reflect.DeepEqual(user.Args, []string{"กำไรสะสม", "1000"}) {
		t.Fatalf("column error = %+v", user)
	}
}

func TestCheckStatementSuggestKeys(t *testing.T) {
	m := Master{Rows: []StatementRow{{RowNo: 10, SuggestKey: "cash_and_equivalents"}, {RowNo: 20}, {RowNo: 30, SuggestKey: "cash"}}, Columns: []StatementColumn{{Title: "กำไรสะสม", SuggestKey: "retained_earnings"}}}
	user, ok := AsUserError(checkStatementSuggestKeys(m))
	if !ok || user.Code != "statement_template_suggest_key_invalid" || user.Field != "rows[2].suggestkey" || !reflect.DeepEqual(user.Args, []string{"30", "cash"}) {
		t.Fatalf("row key error = %+v", user)
	}
	m.Rows[2].SuggestKey = ""
	m.Columns = append(m.Columns, StatementColumn{Title: "อื่น", SuggestKey: "equity"})
	user, ok = AsUserError(checkStatementSuggestKeys(m))
	if !ok || user.Code != "statement_template_column_suggest_key_invalid" || user.Field != "columns[1].suggestkey" || !reflect.DeepEqual(user.Args, []string{"อื่น", "equity"}) {
		t.Fatalf("column key error = %+v", user)
	}
	m.Columns = m.Columns[:1]
	if err := checkStatementSuggestKeys(m); err != nil {
		t.Fatalf("known keys rejected: %v", err)
	}
}
