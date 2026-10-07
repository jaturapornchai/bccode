//go:build integration

package generalledger

import (
	"testing"
)

// Champ 5502 "ผังบัญชีตัวอย่าง" as BC "ใช้ผังบัญชีมาตรฐาน": loads only into a company with no account row at all.
func TestLoadStandardChartIntegration(t *testing.T) {
	f := newPGIntegrityFixture(t) // company "C" already has accounts; it must not count for another company
	chart, err := StandardChart()
	if err != nil {
		t.Fatal(err)
	}
	roots := 0
	for _, a := range chart {
		if a.ParentAccountCode == "" {
			roots++
		}
	}
	empty := Scope{Holding: "H", Company: "EMPTY", Actor: "tester"}
	load := Command{Resource: "accounts", Action: "load-standard"}

	result, err := f.execute(empty, load)
	if err != nil {
		t.Fatalf("load into an empty company: %v", err)
	}
	if result.CreatedAccounts != len(chart) {
		t.Errorf("createdaccounts = %d, want %d", result.CreatedAccounts, len(chart))
	}
	var count, rootRows int
	if err = f.db.QueryRow(`SELECT count(*) FROM gl_records WHERE company='EMPTY' AND kind='accounts'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != len(chart) {
		t.Errorf("gl_records accounts = %d, want %d", count, len(chart))
	}
	if err = f.db.QueryRow(`SELECT count(*) FROM gl_records WHERE company='EMPTY' AND kind='accounts' AND payload->>'level'='1' AND payload->>'parentaccountcode' IS NULL`).Scan(&rootRows); err != nil {
		t.Fatal(err)
	}
	if rootRows != roots {
		t.Errorf("level-1 rows without parent = %d, want %d", rootRows, roots)
	}
	var events int
	if err = f.db.QueryRow(`SELECT count(*) FROM gl_events WHERE company='EMPTY' AND payload->>'action'='accounts:load-standard'`).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if events != 1 {
		t.Errorf("audit events = %d, want 1 (one event for the whole load)", events)
	}

	// The loaded chart is a normal chart: a new account under one of its headers passes the usual account rules.
	var header Account
	for _, a := range chart {
		if !a.AllowPosting && a.Level >= 2 {
			header = a
			break
		}
	}
	if header.AccountCode == "" {
		t.Fatal("standard chart has no header below level 1")
	}
	child := Account{AccountCode: "NEW-1", AccountType: header.AccountType, NormalBalance: header.NormalBalance, ParentAccountCode: header.AccountCode, AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "บัญชีที่ผู้ใช้เพิ่มเอง"}}}
	if _, err = f.execute(empty, Command{Resource: "accounts", Action: "create", Account: &child}); err != nil {
		t.Fatalf("create under a loaded header: %v", err)
	}

	// A second load is refused and adds nothing.
	_, err = f.execute(empty, load)
	user, ok := AsUserError(err)
	if !ok || user.Code != "standard_chart_not_empty" || user.Message == "" {
		t.Fatalf("second load: got %#v, want standard_chart_not_empty", err)
	}
	if err = f.db.QueryRow(`SELECT count(*) FROM gl_records WHERE company='EMPTY' AND kind='accounts'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != len(chart)+1 {
		t.Errorf("accounts after refused load = %d, want %d", count, len(chart)+1)
	}

	// A company that already has an account (the fixture's company) is refused too.
	_, err = f.execute(f.scope, load)
	if user, ok = AsUserError(err); !ok || user.Code != "standard_chart_not_empty" {
		t.Fatalf("load into a company with accounts: got %#v, want standard_chart_not_empty", err)
	}

	// A soft-deleted account still counts: its code would collide and the company is not new.
	gone := Scope{Holding: "H", Company: "GONE", Actor: "tester"}
	created, err := f.execute(gone, Command{Resource: "accounts", Action: "create", Account: &Account{AccountCode: "9000", AccountType: "asset", NormalBalance: "debit", AllowPosting: true, IsActive: true, Names: []Name{{Code: "th", Name: "เงินสดย่อยสาขา"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.execute(gone, Command{Resource: "accounts", Action: "delete", ID: created.ID, Version: created.Version}); err != nil {
		t.Fatal(err)
	}
	_, err = f.execute(gone, load)
	if user, ok = AsUserError(err); !ok || user.Code != "standard_chart_not_empty" {
		t.Fatalf("load after the only account was deleted: got %#v, want standard_chart_not_empty", err)
	}

	// load-standard belongs to accounts only.
	if _, err = f.execute(empty, Command{Resource: "budgets", Action: "load-standard"}); err == nil {
		t.Error("load-standard accepted for budgets")
	}
}
