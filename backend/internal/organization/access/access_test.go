package access

import (
	"testing"

	authmodels "smlcloudplatform/internal/authentication/models"
)

func TestScopesFailClosedAndMatchOnlyCompanyOrBranch(t *testing.T) {
	if AllowsCompany(nil, "company-a-uid") || AllowsBranch(nil, "company-a-uid", "branch-a-uid") {
		t.Fatal("empty scopes must deny company and branch access")
	}

	scopes := []authmodels.AccessScope{
		{ScopeType: "holding"},
		{ScopeType: "company", CompanyUID: " company-a-uid ", AllBranches: true},
		{ScopeType: "branch", CompanyUID: "company-b-uid", BranchUID: " branch-b-uid "},
		{ScopeType: "company", BusinessCode: "LEGACY-COMPANY"},
		{ScopeType: "branch", BusinessCode: "LEGACY-COMPANY", BranchCode: "00001"},
	}
	if !AllowsCompany(scopes, "company-a-uid") || !AllowsCompany(scopes, "company-b-uid") || AllowsCompany(scopes, "company-c-uid") {
		t.Fatal("company visibility must come from a matching stable company UID")
	}
	if !AllowsBranch(scopes, "company-a-uid", "any-branch-uid") {
		t.Fatal("company scope with allbranches must allow its branches")
	}
	if !AllowsBranch(scopes, "company-b-uid", "branch-b-uid") || AllowsBranch(scopes, "company-b-uid", "branch-b-other") {
		t.Fatal("branch scope must match both stable company and branch UIDs")
	}
	if AllowsCompany([]authmodels.AccessScope{{ScopeType: "holding"}}, "company-a-uid") {
		t.Fatal("a raw holding rule is expanded by FindActiveMembership, never matched here")
	}
	if AllowsCompany([]authmodels.AccessScope{{ScopeType: "branch", CompanyUID: "company-a-uid"}}, "company-a-uid") {
		t.Fatal("branch scope without a branch UID must not imply company access")
	}
	if AllowsCompany(scopes, "LEGACY-COMPANY") || AllowsBranch(scopes, "LEGACY-COMPANY", "00001") {
		t.Fatal("legacy business codes must never authorize access")
	}
	if AllowsAllBranches([]authmodels.AccessScope{{ScopeType: "company", AllBranches: true}}, "") || AllowsBranch([]authmodels.AccessScope{{ScopeType: "company", AllBranches: true}}, "", "branch-a-uid") {
		t.Fatal("empty stable company UID must never authorize branch access")
	}
	companyOnly := []authmodels.AccessScope{{ScopeType: "company", CompanyUID: "company-a-uid"}}
	if !AllowsCompany(companyOnly, "company-a-uid") || AllowsBranch(companyOnly, "company-a-uid", "branch-a-uid") {
		t.Fatal("company scope without allbranches must not imply branch access")
	}
}
