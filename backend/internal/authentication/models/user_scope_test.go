package models

import "testing"

func TestScopesAllowCompanySelection(t *testing.T) {
	tests := []struct {
		name   string
		scopes []AccessScope
		want   bool
	}{
		{name: "empty scopes deny transaction access"},
		{name: "holding scope covers every company", scopes: []AccessScope{{ScopeType: " Holding "}}, want: true},
		{name: "holding scope ignores its company fields", scopes: []AccessScope{{ScopeType: "holding", CompanyUID: "company-b", BranchUID: "branch-9"}}, want: true},
		{name: "matching company scope allows company", scopes: []AccessScope{{ScopeType: "company", CompanyUID: "company-a"}}, want: true},
		{name: "different company scope is denied", scopes: []AccessScope{{ScopeType: "company", CompanyUID: "company-b"}}},
		{name: "legacy business code is denied", scopes: []AccessScope{{ScopeType: "company", BusinessCode: "COMP-A"}}},
		{name: "matching branch-only scope is denied", scopes: []AccessScope{{ScopeType: "branch", CompanyUID: "company-a", BranchUID: "branch-1"}}},
		{name: "company scope carrying a branch is denied", scopes: []AccessScope{{ScopeType: "company", CompanyUID: "company-a", BranchUID: "branch-1"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ScopesAllowCompanySelection(test.scopes, "company-a"); got != test.want {
				t.Fatalf("ScopesAllowCompanySelection() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestScopesAllowBranchSelection(t *testing.T) {
	tests := []struct {
		name   string
		scopes []AccessScope
		want   bool
	}{
		{name: "empty scopes deny branch"},
		{name: "exact branch allows", scopes: []AccessScope{{ScopeType: "branch", CompanyUID: "company-a", BranchUID: "branch-1"}}, want: true},
		{name: "different branch denies", scopes: []AccessScope{{ScopeType: "branch", CompanyUID: "company-a", BranchUID: "branch-2"}}},
		{name: "different company denies", scopes: []AccessScope{{ScopeType: "branch", CompanyUID: "company-b", BranchUID: "branch-1"}}},
		{name: "company all branches allows", scopes: []AccessScope{{ScopeType: "company", CompanyUID: "company-a", AllBranches: true}}, want: true},
		{name: "company without all branches denies", scopes: []AccessScope{{ScopeType: "company", CompanyUID: "company-a"}}},
		{name: "holding scope covers every branch", scopes: []AccessScope{{ScopeType: "HOLDING"}}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ScopesAllowBranchSelection(test.scopes, "company-a", "branch-1"); got != test.want {
				t.Fatalf("ScopesAllowBranchSelection() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestHoldingScopeNeedsIdentifiers(t *testing.T) {
	holding := []AccessScope{{ScopeType: "holding"}}
	if ScopesAllowCompanySelection(holding, " ") || ScopesAllowBranchSelection(holding, "company-a", "") || ScopesAllowBranchSelection(holding, "", "branch-1") {
		t.Fatal("holding scope must still require a company (and branch) uid")
	}
}

func TestHasHoldingScope(t *testing.T) {
	tests := []struct {
		name   string
		scopes []AccessScope
		want   bool
	}{
		{name: "empty"},
		{name: "company only", scopes: []AccessScope{{ScopeType: "company", CompanyUID: "company-a", AllBranches: true}}},
		{name: "unknown type", scopes: []AccessScope{{ScopeType: "holdings"}}},
		{name: "holding among others", scopes: []AccessScope{{ScopeType: "branch", CompanyUID: "company-a", BranchUID: "branch-1"}, {ScopeType: " HOLDING "}}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := HasHoldingScope(test.scopes); got != test.want {
				t.Fatalf("HasHoldingScope() = %v, want %v", got, test.want)
			}
		})
	}
}
