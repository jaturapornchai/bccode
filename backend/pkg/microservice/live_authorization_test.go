package microservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"smlcloudplatform/pkg/microservice/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type liveAuthorizationFinderStub struct {
	user       liveUserRecord
	membership liveMembershipRecord
	holding    liveHoldingRecord
	company    liveCompanyRecord
	branch     liveBranchRecord
}

func (s *liveAuthorizationFinderStub) FindOne(_ context.Context, model interface{}, _ interface{}, decode interface{}, _ ...*options.FindOneOptions) error {
	switch model.(type) {
	case *liveUserRecord:
		if s.user.UID == "" {
			return mongo.ErrNoDocuments
		}
		*decode.(*liveUserRecord) = s.user
	case *liveMembershipRecord:
		if s.membership.UserUID == "" {
			return mongo.ErrNoDocuments
		}
		*decode.(*liveMembershipRecord) = s.membership
	case *liveHoldingRecord:
		if s.holding.HoldingCode == "" {
			return mongo.ErrNoDocuments
		}
		*decode.(*liveHoldingRecord) = s.holding
	case *liveCompanyRecord:
		if s.company.CompanyUID == "" {
			return mongo.ErrNoDocuments
		}
		*decode.(*liveCompanyRecord) = s.company
	case *liveBranchRecord:
		if s.branch.BranchUID == "" {
			return mongo.ErrNoDocuments
		}
		*decode.(*liveBranchRecord) = s.branch
	default:
		return mongo.ErrNoDocuments
	}
	return nil
}

func activeAuthorizationFixture() (*liveAuthorization, *liveAuthorizationFinderStub, models.UserInfo) {
	finder := &liveAuthorizationFinderStub{
		user: liveUserRecord{UID: "user-1"},
		membership: liveMembershipRecord{
			MembershipUID:     "membership-1",
			HoldingUID:        "holding-1",
			HoldingCode:       "HOLDING-A",
			UserUID:           "user-1",
			Role:              1,
			PermissionVersion: 7,
			AccessScopes: []liveAccessScope{{
				ScopeType:  "company",
				CompanyUID: "company-1",
			}},
		},
		holding: liveHoldingRecord{HoldingCode: "HOLDING-A", IsActive: true},
		company: liveCompanyRecord{
			HoldingCode: "HOLDING-A",
			CompanyUID:  "company-1",
			Code:        "COMP-A",
			IsActive:    true,
		},
	}
	authorization := newLiveAuthorization(finder)
	authorization.timeNow = func() time.Time { return time.Date(2026, 8, 13, 3, 0, 0, 0, time.UTC) }
	selected := models.UserInfo{
		UID:               "user-1",
		HoldingCode:       "HOLDING-A",
		BusinessCode:      "COMP-A",
		Role:              1,
		PermissionVersion: 7,
	}
	return authorization, finder, selected
}

func TestLiveAuthorizationUsesCurrentMembershipAndStableScope(t *testing.T) {
	authorization, _, selected := activeAuthorizationFixture()

	got, err := authorization.Authorize(context.Background(), selected)
	if err != nil {
		t.Fatalf("Authorize returned error: %v", err)
	}
	if got.MembershipUID != "membership-1" || got.HoldingUID != "holding-1" || got.CompanyUID != "company-1" {
		t.Fatalf("stable workspace IDs = %#v", got)
	}
}

func TestLiveAuthorizationAllowsExactBranchWithoutCompanyWideScope(t *testing.T) {
	authorization, finder, selected := activeAuthorizationFixture()
	finder.membership.AccessScopes = []liveAccessScope{{
		ScopeType: "branch", CompanyUID: "company-1", BranchUID: "branch-1",
	}}
	finder.branch = liveBranchRecord{
		HoldingCode: "HOLDING-A", CompanyUID: "company-1", BranchUID: "branch-1", IsActive: true,
	}
	selected.BranchUID = "branch-1"

	got, err := authorization.Authorize(context.Background(), selected)
	if err != nil {
		t.Fatalf("Authorize returned error: %v", err)
	}
	if got.CompanyUID != "company-1" || got.BranchUID != "branch-1" {
		t.Fatalf("branch workspace IDs = %#v", got)
	}
}

func TestLiveAuthorizationRejectsChangedPermissionVersion(t *testing.T) {
	authorization, finder, selected := activeAuthorizationFixture()
	finder.membership.PermissionVersion++

	_, err := authorization.Authorize(context.Background(), selected)
	if !errors.Is(err, ErrLiveWorkspaceAccess) {
		t.Fatalf("error = %v, want ErrLiveWorkspaceAccess", err)
	}
}

func TestLiveAuthorizationRejectsRevokedMembershipAndInactiveCompany(t *testing.T) {
	t.Run("soft deleted membership", func(t *testing.T) {
		authorization, finder, selected := activeAuthorizationFixture()
		finder.membership.IsDeleted = true
		_, err := authorization.Authorize(context.Background(), selected)
		if !errors.Is(err, ErrLiveWorkspaceAccess) {
			t.Fatalf("error = %v, want ErrLiveWorkspaceAccess", err)
		}
	})

	t.Run("inactive company", func(t *testing.T) {
		authorization, finder, selected := activeAuthorizationFixture()
		finder.company.IsActive = false
		_, err := authorization.Authorize(context.Background(), selected)
		if !errors.Is(err, ErrLiveWorkspaceAccess) {
			t.Fatalf("error = %v, want ErrLiveWorkspaceAccess", err)
		}
	})

	t.Run("unknown role", func(t *testing.T) {
		authorization, finder, selected := activeAuthorizationFixture()
		finder.membership.Role = 255
		selected.Role = 255
		_, err := authorization.Authorize(context.Background(), selected)
		if !errors.Is(err, ErrLiveWorkspaceAccess) {
			t.Fatalf("error = %v, want ErrLiveWorkspaceAccess", err)
		}
	})
}

func TestLiveAuthorizationRejectsDisabledUserWithoutWorkspace(t *testing.T) {
	authorization, finder, selected := activeAuthorizationFixture()
	finder.user.DisabledAt = time.Now().UTC()
	selected = loginOnlyUserInfo(selected)

	_, err := authorization.Authorize(context.Background(), selected)
	if !errors.Is(err, ErrLiveUserAccess) {
		t.Fatalf("error = %v, want ErrLiveUserAccess", err)
	}
}
