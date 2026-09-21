package httpapi

import (
	"context"
	"database/sql"
	"smlcloudplatform/pkg/microservice/models"
	"testing"
)

func TestCompanyScopeUsesCurrentGrantNotSelectedBranch(t *testing.T) {
	for _, tc := range []struct {
		name, role, scopes string
		wide               bool
	}{
		{"manager selected branch", "OWNER", "{}", true},
		{"explicit owner branch restriction", "OWNER", `[{"scopetype":"branch","companyuid":"C","branchuid":"B"}]`, false},
		{"member company grant", "USER", `[{"scopetype":"company","companyuid":"C","allbranches":true}]`, true},
		{"member branch grant", "USER", `[{"scopetype":"branch","companyuid":"C","branchuid":"B"}]`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := &scopeDriver{role: tc.role, sets: "[]", scopes: tc.scopes, grants: map[string]string{"USER": `["financial-close","financial-close:update"]`}}
			db := scopeDB(t, d)
			connect := func(string) (*sql.DB, error) { return db, nil }
			request := scopeContext{user: models.UserInfo{UID: "U", HoldingCode: "H", BusinessCode: "C", BranchUID: "B"}}
			got, err := resolveScope(context.Background(), request, connect)
			if err != nil || got.CompanyWide != tc.wide || got.Scope.Branch != "B" {
				t.Fatalf("scope=%+v err=%v", got, err)
			}
			// Process execution and report preview use the same explicit promotion; normal
			// journal authorization retains the selected branch in the original value.
			promoted, err := got.companyScope()
			if tc.wide {
				if err != nil || promoted.Scope.Branch != "" || got.Scope.Branch != "B" {
					t.Fatal("incorrect promotion", err)
				}
			} else if err == nil {
				t.Fatal("branch-only grant promoted")
			}
			d.role = "USER"
			d.scopes = `[{"scopetype":"branch","companyuid":"C","branchuid":"B"}]`
			next, err := resolveScope(context.Background(), request, connect)
			if err != nil || next.CompanyWide {
				t.Fatal("scope revocation not applied", err)
			}
			d.grantErrors = map[string]error{"USER": sql.ErrNoRows}
			next, err = resolveScope(context.Background(), request, connect)
			if err == nil || next.CompanyWide || len(next.Permissions) != 0 {
				t.Fatal("role revocation leaked grant")
			}
		})
	}
}

func TestCompanyScopePreservesLegacyTokenBranch(t *testing.T) {
	db := scopeDB(t, &scopeDriver{role: "OWNER", sets: "[]", scopes: "{}"})
	connect := func(string) (*sql.DB, error) { return db, nil }
	for _, branch := range []string{"B", ""} {
		request := &mcpGLContext{IContext: scopeContext{user: models.UserInfo{UID: "U", HoldingCode: "H", BusinessCode: "C", BranchUID: branch}}, tokenKind: "mcp", tokenID: "verified"}
		got, err := resolveScope(context.Background(), request, connect)
		if err != nil || got.CompanyWide != (branch == "") {
			t.Fatalf("branch=%q scope=%+v err=%v", branch, got, err)
		}
		promoted, err := got.companyScope()
		if branch != "" {
			if err == nil || promoted.Scope.Branch != "B" {
				t.Fatal("legacy branch token expanded")
			}
		} else if err != nil || promoted.Scope.Branch != "" {
			t.Fatal("company token denied", err)
		}
	}
}
