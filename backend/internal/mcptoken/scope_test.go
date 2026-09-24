package mcptoken

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/pkg/microservice"
)

// withinScope must agree with browser session selection: a company-wide token needs a
// company-wide scope, a legacy branch token a scope covering that branch.
func TestWithinScope(t *testing.T) {
	company := func(code string, allBranches bool) authmodels.AccessScope {
		return authmodels.AccessScope{ScopeType: "company", CompanyUID: code, AllBranches: allBranches}
	}
	branch := authmodels.AccessScope{ScopeType: "branch", CompanyUID: "C", BranchUID: "B"}
	for _, tc := range []struct {
		name            string
		scopes          []authmodels.AccessScope
		company, branch string
		want            bool
	}{
		{"no scopes fail closed", nil, "C", "", false},
		{"own company", []authmodels.AccessScope{company("C", false)}, "C", "", true},
		{"another company", []authmodels.AccessScope{company("C", true)}, "C2", "", false},
		{"branch scope not promoted to company", []authmodels.AccessScope{branch}, "C", "", false},
		{"legacy branch token inside branch scope", []authmodels.AccessScope{branch}, "C", "B", true},
		{"legacy token for another branch", []authmodels.AccessScope{branch}, "C", "B2", false},
		{"company rule without all branches", []authmodels.AccessScope{company("C", false)}, "C", "B", false},
		{"company rule with all branches", []authmodels.AccessScope{company("C", true)}, "C", "B", true},
		{"unexpanded holding rule", []authmodels.AccessScope{{ScopeType: "holding"}}, "C9", "", true},
		{"empty company", []authmodels.AccessScope{company("C", true)}, "", "", false},
	} {
		if got := withinScope(tc.scopes, tc.company, tc.branch); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

type languageRequest struct {
	microservice.IContext
	request *http.Request
}

func (r languageRequest) Request() *http.Request { return r.request }

func TestCompanyOutsideScopeMessageFollowsLanguage(t *testing.T) {
	request := httptest.NewRequest("POST", "/mcp-tokens", nil)
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if got := requestLanguage(languageRequest{request: request}); got != "en-US,en;q=0.9" {
		t.Fatalf("Accept-Language ignored: %q", got)
	}
	request.URL.RawQuery = "lang=th"
	if got := requestLanguage(languageRequest{request: request}); got != "th" {
		t.Fatalf("lang query must win: %q", got)
	}
	th := language.Text("mcp_err_company_outside_scope", "th")
	en := language.Text("mcp_err_company_outside_scope", "en-US,en;q=0.9")
	if th == "mcp_err_company_outside_scope" || !strings.Contains(th, "บริษัท") || !strings.HasPrefix(en, "You can allow only companies") {
		t.Fatalf("missing languages.tsv row: th=%q en=%q", th, en)
	}
}
