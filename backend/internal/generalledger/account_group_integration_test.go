//go:build integration

package generalledger

import (
	"encoding/json"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestLedgerMongoPostgresAccountGroupReferences(t *testing.T) {
	ctx, store, p, db := glAuditStore(t)
	scope := Scope{Holding: "H", Company: "C", Actor: "audit-groups-20260911"}
	serial := 0
	command := func(resource, action, id string, version int64) Command {
		serial++
		return Command{Resource: resource, Action: action, ID: id, Version: version, RequestID: fmt.Sprintf("audit-groups-20260911-%06d", serial)}
	}
	var activeGroup Result
	for _, spec := range []struct {
		code, company string
		active        bool
	}{{"ACTIVE", "C", true}, {"INACTIVE", "C", false}, {"OTHER-COMPANY", "OTHER", true}} {
		groupScope := scope
		groupScope.Company = spec.company
		cmd := command("account-groups", "create", "", 0)
		cmd.Master = &Master{Code: spec.code, Name: "กลุ่มผังบัญชีทดสอบ", IsActive: spec.active}
		result, err := store.Execute(ctx, groupScope, cmd)
		if err != nil || result.ProjectionPending {
			t.Fatalf("create group %s: %+v %v", spec.code, result, err)
		}
		var source Master
		if err := db.Collection("gl_account_groups").FindOne(ctx, scopedID(groupScope, result.ID)).Decode(&source); err != nil || source.Code != spec.code || source.IsActive != spec.active || source.BusinessCode != spec.company {
			t.Fatalf("group fixture not persisted: %+v %v", source, err)
		}
		if spec.code == "ACTIVE" {
			activeGroup = result
		}
	}
	account := Account{AccountCode: "101", AccountGroup: "ACTIVE", Names: []Name{{Code: "th", Name: "บัญชีอ้างอิงกลุ่ม"}}, AccountType: "asset", NormalBalance: "debit", IsActive: true, AllowPosting: true}
	create := command("accounts", "create", "", 0)
	create.Account = &account
	created, err := store.Execute(ctx, scope, create)
	if err != nil || created.ProjectionPending {
		t.Fatalf("valid group rejected: %+v %v", created, err)
	}
	checkAccount := func() {
		t.Helper()
		var source Account
		if err := db.Collection("chart_of_accounts").FindOne(ctx, scopedID(scope, created.ID)).Decode(&source); err != nil || source.AccountGroup != "ACTIVE" || source.Version != created.Version || source.IsDeleted {
			t.Fatalf("account reference changed after rejection: %+v %v", source, err)
		}
		if count, err := db.Collection("chart_of_accounts").CountDocuments(ctx, scopeFilter(scope)); err != nil || count != 1 {
			t.Fatalf("rejected create left orphan account: %d %v", count, err)
		}
	}
	checkAccount()
	projected, err := p.Get(ctx, scope, "accounts", created.ID)
	var projectedAccount Account
	if err != nil || json.Unmarshal(projected, &projectedAccount) != nil || projectedAccount.AccountGroup != "ACTIVE" {
		t.Fatalf("valid account group missing from projection: %s %v", projected, err)
	}
	for _, groupCode := range []string{"MISSING", "OTHER-COMPANY", "INACTIVE"} {
		t.Run(groupCode, func(t *testing.T) {
			for _, action := range []string{"create", "update"} {
				next := account
				next.AccountGroup = groupCode
				cmd := command("accounts", action, created.ID, created.Version)
				if action == "create" {
					cmd.ID, cmd.Version, next.AccountCode = "", 0, "REJECTED"
				}
				cmd.Account = &next
				if _, err := store.Execute(ctx, scope, cmd); err == nil {
					t.Fatalf("accepted %s using invalid group %s", action, groupCode)
				}
				checkAccount()
			}
		})
	}
	remove := command("account-groups", "delete", activeGroup.ID, activeGroup.Version)
	if _, err := store.Execute(ctx, scope, remove); err == nil {
		t.Fatal("deleted group referenced by an active account")
	}
	var group Master
	if err := db.Collection("gl_account_groups").FindOne(ctx, scopedID(scope, activeGroup.ID)).Decode(&group); err != nil || group.IsDeleted || group.Version != activeGroup.Version {
		t.Fatalf("referenced group changed after rejected delete: %+v %v", group, err)
	}
	checkAccount()
	if sequence, err := p.Version(ctx, scope); err != nil || sequence != created.Sequence {
		t.Fatalf("rejected reference mutations advanced projection: %d != %d %v", sequence, created.Sequence, err)
	}
	if count, err := db.Collection("gl_account_groups").CountDocuments(ctx, bson.M{"holdingcode": "H", "businesscode": "OTHER", "code": "OTHER-COMPANY", "isdeleted": false}); err != nil || count != 1 {
		t.Fatalf("cross-company reference attempt changed original group: %d %v", count, err)
	}
}
