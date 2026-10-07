//go:build integration

package generalledger

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

func (f *pgIntegrityFixture) newAccount(code, nameTh string) Result {
	f.t.Helper()
	return f.run(Command{Resource: "accounts", Action: "create", Account: &Account{AccountCode: code, AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: nameTh}}}})
}

// Champ 5502 (GLFrmChartofAccountFormView.cpp:158-196): no second account with the same Name1.
func TestAccountDuplicateNameIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.newAccount("1100", "Petty Cash")
	f.newAccount("1101", "เงินสดย่อย")

	// SQL Server `name1 = '…'`: case and trailing blanks do not matter; leading blanks are trimmed by BC too.
	for _, name := range []string{"Petty Cash", "petty cash", "  PETTY CASH  ", "เงินสดย่อย"} {
		user := f.failCode(Command{Resource: "accounts", Action: "create", Account: &Account{AccountCode: "1200", AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: name}}}}, "account_name_duplicate", "names")
		if len(user.Args) != 2 || (user.Args[1] != "1100" && user.Args[1] != "1101") {
			t.Fatalf("args = %v, want [name, code of the other account]", user.Args)
		}
	}
	// A Latin letter typed as one code point or as base + combining mark is the same name.
	f.newAccount("1110", "caf\u00e9")
	f.failCode(Command{Resource: "accounts", Action: "create", Account: &Account{AccountCode: "1111", AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "cafe\u0301"}}}}, "account_name_duplicate", "names")

	// Name2 (English) is not unique in Champ; only the Thai name is checked.
	f.run(Command{Resource: "accounts", Action: "create", Account: &Account{AccountCode: "1120", AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "เงินสดสำรอง"}, {Code: "en", Name: "Petty Cash"}}}})

	// Edit: the account's own name never collides with itself, another account's name does.
	same := f.accountRecord("1100")
	same.IsActive = false
	f.run(Command{Resource: "accounts", Action: "update", ID: same.ID, Version: same.Version, Account: &same})
	renamed := f.accountRecord("1100")
	renamed.Names = []Name{{Code: "th", Name: "เงินสดย่อย "}}
	f.failCode(Command{Resource: "accounts", Action: "update", ID: renamed.ID, Version: renamed.Version, Account: &renamed}, "account_name_duplicate", "names")

	// An older duplicate (made before the rule existed) must stay editable: deactivating it is the way to fix it.
	if _, err := f.db.Exec(`UPDATE gl_records SET payload=jsonb_set(payload,'{names,0,name}','"Petty Cash"') WHERE company='C' AND kind='accounts' AND code='1101'`); err != nil {
		t.Fatal(err)
	}
	legacy := f.accountRecord("1101")
	legacy.IsActive = false
	f.run(Command{Resource: "accounts", Action: "update", ID: legacy.ID, Version: legacy.Version, Account: &legacy})

	// A deleted account frees its name.
	deleted := f.accountRecord("1110")
	f.run(Command{Resource: "accounts", Action: "delete", ID: deleted.ID, Version: deleted.Version})
	f.newAccount("1130", "cafe\u0301")
}

func (f *pgIntegrityFixture) accountRecord(code string) Account {
	f.t.Helper()
	var id string
	if err := f.db.QueryRow(`SELECT id FROM gl_records WHERE company='C' AND kind='accounts' AND code=$1`, code).Scan(&id); err != nil {
		f.t.Fatal(err)
	}
	raw, err := f.store.Get(f.ctx, f.scope, "accounts", id)
	if err != nil {
		f.t.Fatal(err)
	}
	var a Account
	if err = json.Unmarshal(raw, &a); err != nil {
		f.t.Fatal(err)
	}
	return a
}

func (f *pgIntegrityFixture) references(code string) []AccountReference {
	f.t.Helper()
	r := f.run(Command{Resource: "accounts", Action: "references", ID: f.accountRecord(code).ID})
	if r.References == nil || r.References.Items == nil {
		f.t.Fatalf("references of %s: result has no items", code)
	}
	return r.References.Items
}

// Champ 5502 "รายการที่ใช้บัญชีนี้": counts per kind of the records that use the account — the same set that blocks deleting it.
func TestAccountReferencesIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t)
	f.draft("JV-REF-1", "B1") // lines on 1000 and 4000
	f.run(Command{Resource: "budgets", Action: "create", Budget: &Budget{Code: "BG-REF", Name: "งบประมาณรายได้", FiscalYear: "2026", Lines: []BudgetLine{{AccountCode: "4000", Periods: budgetMonths("100")}}}})
	f.run(Command{Resource: "account-groups", Action: "create", Master: &Master{Code: "G-REF", Name: "กลุ่มรายได้", IsActive: true, AccountEntries: []AccountEntry{{From: "4000"}}}})
	f.run(Command{Resource: "accounts", Action: "create", Account: &Account{AccountCode: "1900", AccountType: "asset", NormalBalance: "debit", AllowPosting: false, IsActive: true, Names: []Name{{Code: "th", Name: "สินทรัพย์อื่น"}}}})
	f.run(Command{Resource: "accounts", Action: "create", Account: &Account{AccountCode: "1910", AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, ParentAccountCode: "1900", Names: []Name{{Code: "th", Name: "เงินมัดจำจ่าย"}}}})

	// Allocation rule targets and journal template lines name accounts too (stored rows, as the screens save them).
	f.run(Command{Resource: "accounts", Action: "create", Account: &Account{AccountCode: "5100", AccountType: "expense", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "ค่าเช่าสาขา"}}}})
	f.run(Command{Resource: "accounts", Action: "create", Account: &Account{AccountCode: "5200", AccountType: "expense", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "ค่าน้ำประปา"}}}})
	for _, row := range []struct{ kind, id, payload string }{
		{"allocations", "alloc-ref", `{"code":"AL-REF","name":"ปันส่วนค่าเช่า","accountcode":"5000","allocaterules":[{"accountcode":"5100","rate":"100"}]}`},
		{"journal-templates", "tpl-ref", `{"code":"TP-REF","name":"ค่าน้ำประปารายเดือน","template":{"lines":[{"accountcode":"5200"},{"accountcode":"1000"}]}}`},
	} {
		if _, err := f.db.ExecContext(f.ctx, `INSERT INTO gl_records(company,kind,id,code,version,payload) VALUES ($1,$2,$3,$4,1,$5::jsonb)`, f.scope.Company, row.kind, row.id, row.id, row.payload); err != nil {
			t.Fatal(err)
		}
	}
	want := map[string][]AccountReference{
		"4000": {{Kind: "account-groups", Count: 1}, {Kind: "budgets", Count: 1}, {Kind: "journals", Count: 1}},
		"1000": {{Kind: "journal-templates", Count: 1}, {Kind: "journals", Count: 1}},
		"3200": {{Kind: "fiscal-years", Count: 2}}, // profit/loss account of 2026 and 2027
		"3100": {{Kind: "fiscal-years", Count: 2}}, // retained earnings account of both years
		"1900": {{Kind: "accounts", Count: 1}},     // child account
		"3000": {},
		"5000": {{Kind: "allocations", Count: 1}}, // cost account of the allocation
		"5100": {{Kind: "allocations", Count: 1}}, // allocation rule target
		"5200": {{Kind: "journal-templates", Count: 1}},
	}
	for code, items := range want {
		got := f.references(code)
		sort.Slice(got, func(i, j int) bool { return got[i].Kind < got[j].Kind })
		if !reflect.DeepEqual(got, items) {
			t.Errorf("references of %s = %+v, want %+v", code, got, items)
		}
	}

	// The panel and the delete block agree: every account with references refuses deletion, the others delete.
	for code, items := range want {
		a := f.accountRecord(code)
		_, err := f.execute(f.scope, Command{Resource: "accounts", Action: "delete", ID: a.ID, Version: a.Version})
		if len(items) > 0 && err == nil {
			t.Errorf("account %s has references %+v but delete was accepted", code, items)
		}
		if len(items) == 0 && err != nil {
			t.Errorf("account %s has no references but delete failed: %v", code, err)
		}
	}
}
