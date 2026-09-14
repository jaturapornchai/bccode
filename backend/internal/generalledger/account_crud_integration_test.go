//go:build integration

package generalledger

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestLedgerMongoPostgresAccountLevelCRUD(t *testing.T) {
	ctx, store, p, db := glAuditStore(t)
	scope := Scope{Holding: "H", Company: "C", Actor: "audit-account-level-20260911"}
	serial := 0
	command := func(resource, action, id string, version int64) Command {
		serial++
		return Command{Resource: resource, Action: action, ID: id, Version: version, RequestID: fmt.Sprintf("audit-account-level-%06d", serial)}
	}

	// 1. Create Parent Account (Level 1)
	parent := Account{
		AccountCode:   "1000",
		Names:         []Name{{Code: "th", Name: "สินทรัพย์"}},
		AccountType:   "asset",
		NormalBalance: "debit",
		AllowPosting:  false,
		IsActive:      true,
		Level:         1,
	}
	createParent := command("accounts", "create", "", 0)
	createParent.Account = &parent
	parentResult, err := store.Execute(ctx, scope, createParent)
	if err != nil || parentResult.ProjectionPending {
		t.Fatalf("create parent account failed: %+v %v", parentResult, err)
	}

	// 1.1 Verify Parent in MongoDB
	var mongoParent Account
	if err := db.Collection("chart_of_accounts").FindOne(ctx, scopedID(scope, parentResult.ID)).Decode(&mongoParent); err != nil {
		t.Fatalf("parent account not found in Mongo: %v", err)
	}
	if mongoParent.Level != 1 || mongoParent.AccountCode != "1000" || mongoParent.IsDeleted {
		t.Fatalf("mongo parent mismatch: %+v", mongoParent)
	}

	// 1.2 Verify Parent in PostgreSQL
	projectedParentRaw, err := p.Get(ctx, scope, "accounts", parentResult.ID)
	if err != nil {
		t.Fatalf("parent account not found in PostgreSQL projection: %v", err)
	}
	var pgParent Account
	if err := json.Unmarshal(projectedParentRaw, &pgParent); err != nil || pgParent.Level != 1 || pgParent.AccountCode != "1000" {
		t.Fatalf("pg projected parent mismatch: %s %v", projectedParentRaw, err)
	}

	// 2. Create Child Account with auto-calculated or specified Level 2
	child := Account{
		AccountCode:       "1100",
		Names:             []Name{{Code: "th", Name: "สินทรัพย์หมุนเวียน"}},
		AccountType:       "asset",
		ParentAccountCode: "1000",
		NormalBalance:     "debit",
		AllowPosting:      true,
		IsActive:          true,
		Level:             2,
	}
	createChild := command("accounts", "create", "", 0)
	createChild.Account = &child
	childResult, err := store.Execute(ctx, scope, createChild)
	if err != nil || childResult.ProjectionPending {
		t.Fatalf("create child account failed: %+v %v", childResult, err)
	}

	// 2.1 Verify Child in MongoDB
	var mongoChild Account
	if err := db.Collection("chart_of_accounts").FindOne(ctx, scopedID(scope, childResult.ID)).Decode(&mongoChild); err != nil {
		t.Fatalf("child account not found in Mongo: %v", err)
	}
	if mongoChild.Level != 2 || mongoChild.ParentAccountCode != "1000" || mongoChild.IsDeleted {
		t.Fatalf("mongo child mismatch: %+v", mongoChild)
	}

	// 2.2 Verify Child in PostgreSQL
	projectedChildRaw, err := p.Get(ctx, scope, "accounts", childResult.ID)
	if err != nil {
		t.Fatalf("child account not found in PostgreSQL projection: %v", err)
	}
	var pgChild Account
	if err := json.Unmarshal(projectedChildRaw, &pgChild); err != nil || pgChild.Level != 2 || pgChild.ParentAccountCode != "1000" {
		t.Fatalf("pg projected child mismatch: %s %v", projectedChildRaw, err)
	}

	// 3. Verify Read / List via PostgreSQL
	page, err := p.List(ctx, scope, "accounts", "", 1, 10, ListFilter{})
	if err != nil || page.Total < 2 {
		t.Fatalf("p.List accounts failed: total=%d %v", page.Total, err)
	}

	// 4. Update Child Account (Update Thai Name and Level to 3)
	updatedChild := mongoChild
	updatedChild.Names = []Name{{Code: "th", Name: "สินทรัพย์หมุนเวียน (แก้ไข)"}}
	updatedChild.Level = 3
	updateCmd := command("accounts", "update", childResult.ID, childResult.Version)
	updateCmd.Account = &updatedChild
	updateResult, err := store.Execute(ctx, scope, updateCmd)
	if err != nil || updateResult.ProjectionPending {
		t.Fatalf("update child account failed: %+v %v", updateResult, err)
	}

	// 4.1 Verify Update in MongoDB
	if err := db.Collection("chart_of_accounts").FindOne(ctx, scopedID(scope, childResult.ID)).Decode(&mongoChild); err != nil {
		t.Fatalf("reloading child after update in Mongo failed: %v", err)
	}
	if mongoChild.ThaiName() != "สินทรัพย์หมุนเวียน (แก้ไข)" || mongoChild.Level != 3 || mongoChild.Version != updateResult.Version {
		t.Fatalf("mongo child update mismatch: %+v", mongoChild)
	}

	// 4.2 Verify Update in PostgreSQL
	projectedChildRaw, err = p.Get(ctx, scope, "accounts", childResult.ID)
	if err != nil {
		t.Fatalf("updated child not found in PG: %v", err)
	}
	if err := json.Unmarshal(projectedChildRaw, &pgChild); err != nil || pgChild.ThaiName() != "สินทรัพย์หมุนเวียน (แก้ไข)" || pgChild.Level != 3 || pgChild.Version != updateResult.Version {
		t.Fatalf("pg projected child update mismatch: %s %v", projectedChildRaw, err)
	}

	// 5. Delete Guard: Attempt to delete parent while child exists (must fail)
	deleteParentCmd := command("accounts", "delete", parentResult.ID, parentResult.Version)
	if _, err := store.Execute(ctx, scope, deleteParentCmd); err == nil {
		t.Fatal("expected error deleting parent with existing children, but got nil")
	}

	// 6. Delete Child Account
	deleteChildCmd := command("accounts", "delete", childResult.ID, updateResult.Version)
	deleteChildResult, err := store.Execute(ctx, scope, deleteChildCmd)
	if err != nil || deleteChildResult.ProjectionPending {
		t.Fatalf("delete child account failed: %+v %v", deleteChildResult, err)
	}

	// 6.1 Verify Child deleted in Mongo
	if err := db.Collection("chart_of_accounts").FindOne(ctx, scopedID(scope, childResult.ID)).Decode(&mongoChild); err != nil || !mongoChild.IsDeleted {
		t.Fatalf("mongo child should be soft-deleted: isdeleted=%v err=%v", mongoChild.IsDeleted, err)
	}

	// 6.2 Verify Child deleted in PostgreSQL (p.Get returns ErrNotFound)
	if _, err := p.Get(ctx, scope, "accounts", childResult.ID); err != ErrNotFound {
		t.Fatalf("p.Get for deleted child expected ErrNotFound, got: %v", err)
	}

	// 7. Delete Parent Account
	deleteParentCmd = command("accounts", "delete", parentResult.ID, parentResult.Version)
	deleteParentResult, err := store.Execute(ctx, scope, deleteParentCmd)
	if err != nil || deleteParentResult.ProjectionPending {
		t.Fatalf("delete parent account failed: %+v %v", deleteParentResult, err)
	}

	// 7.1 Verify Parent deleted in Mongo
	if err := db.Collection("chart_of_accounts").FindOne(ctx, scopedID(scope, parentResult.ID)).Decode(&mongoParent); err != nil || !mongoParent.IsDeleted {
		t.Fatalf("mongo parent should be soft-deleted: isdeleted=%v err=%v", mongoParent.IsDeleted, err)
	}

	// 7.2 Verify Parent deleted in PostgreSQL
	if _, err := p.Get(ctx, scope, "accounts", parentResult.ID); err != ErrNotFound {
		t.Fatalf("p.Get for deleted parent expected ErrNotFound, got: %v", err)
	}

	// 8. Journal Reference Delete Guard: Strictly forbid deleting account referenced in journals
	profitLossAcc := Account{
		AccountCode:   "3998",
		Names:         []Name{{Code: "th", Name: "กำไรขาดทุน"}},
		AccountType:   "equity",
		NormalBalance: "credit",
		AllowPosting:  true,
		IsActive:      true,
		Level:         1,
	}
	createPL := command("accounts", "create", "", 0)
	createPL.Account = &profitLossAcc
	_, err = store.Execute(ctx, scope, createPL)
	if err != nil {
		t.Fatalf("create profitloss account failed: %v", err)
	}

	retainedAcc := Account{
		AccountCode:   "3999",
		Names:         []Name{{Code: "th", Name: "กำไรสะสม"}},
		AccountType:   "equity",
		NormalBalance: "credit",
		AllowPosting:  true,
		IsActive:      true,
		Level:         1,
	}
	createRetained := command("accounts", "create", "", 0)
	createRetained.Account = &retainedAcc
	_, err = store.Execute(ctx, scope, createRetained)
	if err != nil {
		t.Fatalf("create retained account failed: %v", err)
	}

	yearCmd := command("fiscal-years", "create", "", 0)
	yearCmd.FiscalYear = &FiscalYear{
		Code:                    "2026",
		StartDate:               "2026-01-01",
		EndDate:                 "2026-12-31",
		Scale:                   2,
		IsActive:                true,
		ProfitLossAccount:       "3998",
		RetainedEarningsAccount: "3999",
	}
	_, err = store.Execute(ctx, scope, yearCmd)
	if err != nil {
		t.Fatalf("create fiscal year failed: %v", err)
	}

	periodCmd := command("periods", "create", "", 0)
	periodCmd.Master = &Master{
		Code:       "2026-09",
		Name:       "กันยายน 2026",
		FiscalYear: "2026",
		StartDate:  "2026-09-01",
		EndDate:    "2026-09-30",
		IsActive:   true,
	}
	_, err = store.Execute(ctx, scope, periodCmd)
	if err != nil {
		t.Fatalf("create period failed: %v", err)
	}

	accJournal := Account{
		AccountCode:   "1200",
		Names:         []Name{{Code: "th", Name: "เงินฝากธนาคาร"}},
		AccountType:   "asset",
		NormalBalance: "debit",
		AllowPosting:  true,
		IsActive:      true,
		Level:         1,
	}
	createAccJournal := command("accounts", "create", "", 0)
	createAccJournal.Account = &accJournal
	accJournalResult, err := store.Execute(ctx, scope, createAccJournal)
	if err != nil || accJournalResult.ProjectionPending {
		t.Fatalf("create account for journal test failed: %+v %v", accJournalResult, err)
	}

	// Create journal referencing 1200
	journalCmd := command("journals", "create", "", 0)
	journalCmd.Journal = &Journal{
		DocNo:       "JV-REF-001",
		Date:        "2026-09-11",
		BookCode:    "JV",
		FiscalYear:  "2026",
		Description: "ทดสอบรายการอ้างอิงผังบัญชี",
		BranchCode:  "B1",
		Kind:        "manual",
		Lines: []Line{
			{AccountCode: "1200", Debit: "500", CashFlow: "operating"},
			{AccountCode: "3999", Credit: "500"},
		},
	}
	journalResult, err := store.Execute(ctx, scope, journalCmd)
	if err != nil || journalResult.ProjectionPending {
		t.Fatalf("create journal failed: %+v %v", journalResult, err)
	}

	// Attempt to delete account 1200: MUST be strictly forbidden
	deleteAccJournalCmd := command("accounts", "delete", accJournalResult.ID, accJournalResult.Version)
	_, err = store.Execute(ctx, scope, deleteAccJournalCmd)
	if err == nil {
		t.Fatal("expected error deleting account referenced by journal, got nil")
	}
	expectedErrMsg := "บัญชีนี้มีข้อมูลอ้างอิงจากสมุดรายวัน ห้ามลบผังบัญชีเด็ดขาด กรุณาปิดใช้งานแทนการลบ"
	if err.Error() != expectedErrMsg {
		t.Fatalf("unexpected error message: got %q, want %q", err.Error(), expectedErrMsg)
	}

	// Verify account 1200 in Mongo: still active and NOT deleted
	var mongoAcc1200 Account
	if err := db.Collection("chart_of_accounts").FindOne(ctx, scopedID(scope, accJournalResult.ID)).Decode(&mongoAcc1200); err != nil {
		t.Fatalf("account 1200 not found in Mongo: %v", err)
	}
	if mongoAcc1200.IsDeleted {
		t.Fatal("account 1200 should NOT be deleted in Mongo")
	}

	// Verify account 1200 in PostgreSQL: still readable
	if _, err := p.Get(ctx, scope, "accounts", accJournalResult.ID); err != nil {
		t.Fatalf("account 1200 should still exist in PostgreSQL projection: %v", err)
	}

	// 8.1 Even if the journal is voided/soft-deleted, deleting account 1200 MUST STILL FAIL
	voidJournalCmd := command("journals", "delete", journalResult.ID, journalResult.Version)
	voidJournalCmd.Reason = "ยกเลิกรายวันทดสอบ"
	voidResult, err := store.Execute(ctx, scope, voidJournalCmd)
	if err != nil || voidResult.ProjectionPending {
		t.Fatalf("void journal failed: %+v %v", voidResult, err)
	}

	// Attempt delete account 1200 again after journal void: MUST STILL FAIL
	_, err = store.Execute(ctx, scope, deleteAccJournalCmd)
	if err == nil {
		t.Fatal("expected error deleting account with historical journal reference even after void, got nil")
	}
	if err.Error() != expectedErrMsg {
		t.Fatalf("unexpected error message after journal void: got %q, want %q", err.Error(), expectedErrMsg)
	}
}
