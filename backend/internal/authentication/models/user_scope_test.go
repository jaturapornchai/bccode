package models

import "testing"

func TestScopesAllowCompanySelection(t *testing.T) {
	tests := []struct {
		name   string
		scopes []AccessScope
		want   bool
	}{
		{name: "empty scopes keep full access", want: true},
		{name: "holding scope allows company", scopes: []AccessScope{{ScopeType: "holding"}}, want: true},
		{name: "matching company scope allows company", scopes: []AccessScope{{ScopeType: "company", BusinessCode: "COMP-A"}}, want: true},
		{name: "different company scope is denied", scopes: []AccessScope{{ScopeType: "company", BusinessCode: "COMP-B"}}},
		{name: "matching branch-only scope is denied", scopes: []AccessScope{{ScopeType: "branch", BusinessCode: "COMP-A", BranchCode: "00001"}}},
		{name: "company scope carrying a branch is denied", scopes: []AccessScope{{ScopeType: "company", BusinessCode: "COMP-A", BranchCode: "00001"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ScopesAllowCompanySelection(test.scopes, "comp-a"); got != test.want {
				t.Fatalf("ScopesAllowCompanySelection() = %v, want %v", got, test.want)
			}
		})
	}
}
