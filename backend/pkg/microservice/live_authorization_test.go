package microservice

import (
	"context"
	"errors"
	"testing"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/pkg/microservice/models"
)

func TestLiveAuthorizationRejectsNonUUIDIdentityBeforeQuerying(t *testing.T) {
	// A nil database proves the check happens before any query.
	authorization := &liveAuthorization{}
	for _, uid := range []string{"", "  ", "member@example.com"} {
		if _, err := authorization.Authorize(context.Background(), models.UserInfo{UID: uid}); !errors.Is(err, ErrLiveUserAccess) {
			t.Fatalf("Authorize(%q) error = %v, want ErrLiveUserAccess", uid, err)
		}
	}
}

func TestNewLiveAuthorizationWithoutDatabaseIsDisabled(t *testing.T) {
	if newLiveAuthorization(nil) != nil {
		t.Fatal("nil database must not build a live authorization")
	}
}

func TestCompanyWorkspaceRequiresCompanyLevelScope(t *testing.T) {
	scopes := []authmodels.AccessScope{
		{ScopeType: "company", CompanyUID: "01"},
		{ScopeType: "branch", CompanyUID: "02", BranchUID: "00001"},
	}
	if !allowsCompanyWorkspace(scopes, "01") {
		t.Fatal("company scope must open the company workspace")
	}
	if allowsCompanyWorkspace(scopes, "02") {
		t.Fatal("branch-only scope must not open the whole company")
	}
	if allowsCompanyWorkspace(nil, "01") {
		t.Fatal("empty scopes must fail closed")
	}
}
