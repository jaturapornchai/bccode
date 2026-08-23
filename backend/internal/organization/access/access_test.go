package access

import (
	"context"
	"errors"
	"testing"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	shopmodels "smlcloudplatform/internal/shop/models"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type membershipFinderStub struct {
	membership authmodels.ShopUser
	err        error
	filters    []interface{}
}

func (stub *membershipFinderStub) FindOne(_ context.Context, _ interface{}, filter interface{}, decode interface{}, _ ...*options.FindOneOptions) error {
	stub.filters = append(stub.filters, filter)
	if stub.err != nil {
		return stub.err
	}
	*(decode.(*authmodels.ShopUser)) = stub.membership
	return nil
}

func TestFindActiveMembershipFailsClosed(t *testing.T) {
	now := time.Date(2026, 8, 13, 8, 0, 0, 0, time.UTC)
	active := authmodels.ShopUser{ID: primitive.NewObjectID()}

	tests := []struct {
		name       string
		membership authmodels.ShopUser
		finderErr  error
		wantDenied bool
		wantLookup bool
	}{
		{name: "active", membership: active},
		{name: "missing", wantDenied: true},
		{name: "no documents", finderErr: mongo.ErrNoDocuments, wantDenied: true},
		{name: "soft deleted", membership: authmodels.ShopUser{ID: primitive.NewObjectID(), IsDeleted: true}, wantDenied: true},
		{name: "disabled", membership: authmodels.ShopUser{ID: primitive.NewObjectID(), IsAccessDisabled: true}, wantDenied: true},
		{name: "expired", membership: authmodels.ShopUser{ID: primitive.NewObjectID(), AccessExpiryDate: now}, wantDenied: true},
		{name: "future expiry", membership: authmodels.ShopUser{ID: primitive.NewObjectID(), AccessExpiryDate: now.Add(time.Minute)}},
		{name: "lookup failure", finderErr: errors.New("database unavailable"), wantLookup: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			finder := &membershipFinderStub{membership: test.membership, err: test.finderErr}
			_, err := FindActiveMembership(context.Background(), finder, micromodels.UserInfo{
				HoldingCode: "HOLDING",
				Username:    " User@Example.com ",
				UID:         "USER-ID",
			}, now)
			if test.wantDenied != errors.Is(err, ErrActiveMembershipRequired) {
				t.Fatalf("denied = %v, error = %v", errors.Is(err, ErrActiveMembershipRequired), err)
			}
			if test.wantLookup != (err != nil && !errors.Is(err, ErrActiveMembershipRequired)) {
				t.Fatalf("lookup failure mismatch: %v", err)
			}
			if test.wantLookup && len(finder.filters) != 1 {
				t.Fatalf("database failure must not fall back; lookup count = %d", len(finder.filters))
			}
		})
	}
}

func TestFindActiveMembershipRejectsMissingTenantOrIdentityWithoutLookup(t *testing.T) {
	for _, userInfo := range []micromodels.UserInfo{
		{Username: "user@example.com", UID: "USER-ID"},
		{HoldingCode: "HOLDING"},
		{HoldingCode: "HOLDING", Username: "user@example.com"},
	} {
		finder := &membershipFinderStub{}
		_, err := FindActiveMembership(context.Background(), finder, userInfo, time.Now())
		if !errors.Is(err, ErrActiveMembershipRequired) {
			t.Fatalf("error = %v, want active membership required", err)
		}
		if len(finder.filters) != 0 {
			t.Fatalf("unexpected membership lookup: %#v", finder.filters)
		}
	}
}

func TestFindActiveMembershipPrefersStableUID(t *testing.T) {
	finder := &membershipFinderStub{membership: authmodels.ShopUser{ID: primitive.NewObjectID()}}
	_, err := FindActiveMembership(context.Background(), finder, micromodels.UserInfo{
		HoldingCode: "HOLDING",
		Username:    " User@Example.com ",
		UID:         "USER-ID",
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if len(finder.filters) != 1 {
		t.Fatalf("lookup count = %d, want 1", len(finder.filters))
	}
	filter := finder.filters[0].(bson.M)
	if filter["useruid"] != "USER-ID" || filter["username"] != nil || filter["isdeleted"] != false {
		t.Fatalf("unexpected stable identity filter: %#v", filter)
	}
}

func TestFindActiveMembershipWithUIDOnlyDoesNotMatchEmptyUsername(t *testing.T) {
	finder := &membershipFinderStub{membership: authmodels.ShopUser{ID: primitive.NewObjectID()}}
	_, err := FindActiveMembership(context.Background(), finder, micromodels.UserInfo{
		HoldingCode: "HOLDING",
		UID:         "USER-ID",
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	filter := finder.filters[0].(bson.M)
	if filter["useruid"] != "USER-ID" || filter["username"] != nil || filter["$or"] != nil {
		t.Fatalf("unexpected UID-only identity filter: %#v", filter)
	}
}

func TestFindActiveMembershipDoesNotFallbackToUsername(t *testing.T) {
	finder := &membershipFinderStub{}
	_, err := FindActiveMembership(context.Background(), finder, micromodels.UserInfo{
		HoldingCode: "HOLDING",
		Username:    "user@example.com",
		UID:         "USER-ID",
	}, time.Now())
	if !errors.Is(err, ErrActiveMembershipRequired) {
		t.Fatalf("error = %v, want active membership required", err)
	}
	if len(finder.filters) != 1 || finder.filters[0].(bson.M)["username"] != nil {
		t.Fatalf("username fallback must not be used: %#v", finder.filters)
	}
}

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
		t.Fatal("holding role/scope must not imply company transaction access")
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

func TestFindActiveHoldingManagerAndOwner(t *testing.T) {
	userInfo := micromodels.UserInfo{HoldingCode: "HOLDING", UID: "USER-ID"}
	now := time.Now()

	for _, test := range []struct {
		name       string
		role       authmodels.UserRole
		managerErr error
		ownerErr   error
	}{
		{name: "user", role: authmodels.ROLE_USER, managerErr: ErrHoldingManagerRequired, ownerErr: ErrHoldingOwnerRequired},
		{name: "admin", role: authmodels.ROLE_ADMIN, ownerErr: ErrHoldingOwnerRequired},
		{name: "owner", role: authmodels.ROLE_OWNER},
	} {
		t.Run(test.name, func(t *testing.T) {
			membership := authmodels.ShopUser{ID: primitive.NewObjectID()}
			membership.Role = test.role
			_, managerErr := FindActiveHoldingManager(context.Background(), &membershipFinderStub{membership: membership}, userInfo, now)
			if !errors.Is(managerErr, test.managerErr) || (test.managerErr == nil && managerErr != nil) {
				t.Fatalf("manager error = %v, want %v", managerErr, test.managerErr)
			}
			_, ownerErr := FindActiveHoldingOwner(context.Background(), &membershipFinderStub{membership: membership}, userInfo, now)
			if !errors.Is(ownerErr, test.ownerErr) || (test.ownerErr == nil && ownerErr != nil) {
				t.Fatalf("owner error = %v, want %v", ownerErr, test.ownerErr)
			}
		})
	}
}

type holdingFinderStub struct {
	holding shopmodels.ShopDoc
	err     error
	filter  bson.M
}

func (stub *holdingFinderStub) FindOne(_ context.Context, _ interface{}, filter interface{}, decode interface{}, _ ...*options.FindOneOptions) error {
	stub.filter = filter.(bson.M)
	*(decode.(*shopmodels.ShopDoc)) = stub.holding
	return stub.err
}

func TestRequireActiveHoldingFailsClosed(t *testing.T) {
	active := shopmodels.ShopDoc{ID: primitive.NewObjectID()}
	active.IsActive = true

	for _, test := range []struct {
		name    string
		holding shopmodels.ShopDoc
		err     error
		want    error
	}{
		{name: "active", holding: active},
		{name: "missing", want: ErrActiveHoldingRequired},
		{name: "inactive", holding: shopmodels.ShopDoc{ID: primitive.NewObjectID()}, want: ErrActiveHoldingRequired},
		{name: "no documents", err: mongo.ErrNoDocuments, want: ErrActiveHoldingRequired},
		{name: "database failure", err: errors.New("database unavailable")},
	} {
		t.Run(test.name, func(t *testing.T) {
			finder := &holdingFinderStub{holding: test.holding, err: test.err}
			err := RequireActiveHolding(context.Background(), finder, " HOLDING ")
			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if test.want == nil && test.err == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.err != nil && !errors.Is(test.err, mongo.ErrNoDocuments) && (err == nil || errors.Is(err, ErrActiveHoldingRequired)) {
				t.Fatalf("database failure must remain distinct: %v", err)
			}
			if finder.filter != nil && finder.filter["holdingcode"] != "HOLDING" {
				t.Fatalf("holding code was not trimmed: %#v", finder.filter)
			}
		})
	}
}
