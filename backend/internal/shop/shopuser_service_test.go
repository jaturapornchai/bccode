package shop_test

import (
	"context"
	"smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/shop"
	"testing"
	"time"

	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShopUserDeleteCannotDeleteCreator(t *testing.T) {
	shopUserRepo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	holdingCode := "holdingcode"
	authUsername := "owner@example.com"
	creatorUsername := "creator@example.com"

	shopUserRepo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, authUsername).Return(testShopUser(holdingCode, authUsername, models.ROLE_OWNER), nil)
	shopUserRepo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, creatorUsername).Return(testShopUser(holdingCode, creatorUsername, models.ROLE_OWNER), nil)
	shopUserRepo.On("FindShopCreatedBy", ctx, holdingCode).Return(creatorUsername, nil)

	shopUserSvc := shop.NewShopUserService(shopUserRepo)

	err := shopUserSvc.DeleteUserPermissionShop(holdingCode, authUsername, creatorUsername)

	require.EqualError(t, err, "creator_cannot_delete")
	shopUserRepo.AssertNotCalled(t, "Delete", ctx, holdingCode, creatorUsername)
}

func TestShopUserSaveFullProfileCreatorCannotBeDisabled(t *testing.T) {
	shopUserRepo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	holdingCode := "holdingcode"
	authUsername := "owner@example.com"
	creatorUsername := "creator@example.com"

	shopUserRepo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, authUsername).Return(testShopUser(holdingCode, authUsername, models.ROLE_OWNER), nil)
	shopUserRepo.On("FindShopCreatedBy", ctx, holdingCode).Return(creatorUsername, nil)
	shopUserRepo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, creatorUsername).Return(testShopUser(holdingCode, creatorUsername, models.ROLE_OWNER), nil)

	shopUserSvc := shop.NewShopUserService(shopUserRepo)

	err := shopUserSvc.SaveUserFullProfile(holdingCode, authUsername, &models.UserRoleRequest{
		Username:         creatorUsername,
		Role:             models.ROLE_OWNER,
		IsAccessDisabled: true,
	})

	require.EqualError(t, err, "creator_access_cannot_be_disabled")
	shopUserRepo.AssertNotCalled(t, "SaveFullProfile", ctx, holdingCode)
}

func TestShopUserSaveFullProfileDisablesMember(t *testing.T) {
	shopUserRepo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	holdingCode := "holdingcode"
	authUsername := "owner@example.com"
	targetUsername := "member@example.com"

	shopUserRepo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, authUsername).Return(testShopUser(holdingCode, authUsername, models.ROLE_OWNER), nil)
	shopUserRepo.On("FindShopCreatedBy", ctx, holdingCode).Return(authUsername, nil)
	shopUserRepo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, targetUsername).Return(testShopUser(holdingCode, targetUsername, models.ROLE_ADMIN), nil)
	shopUserRepo.On("SaveFullProfile", ctx, holdingCode, requireAccessDisabledRequest(t, targetUsername, authUsername)).Return(nil)

	shopUserSvc := shop.NewShopUserService(shopUserRepo)

	err := shopUserSvc.SaveUserFullProfile(holdingCode, authUsername, &models.UserRoleRequest{
		EditUsername:     targetUsername,
		Username:         targetUsername,
		Role:             models.ROLE_ADMIN,
		IsAccessDisabled: true,
	})

	require.NoError(t, err)
}

// Adding (no editusername/useruid) a username that is already a member must not silently
// overwrite that member's role and permissions (save-audit finding 14).
func TestShopUserAddRejectsExistingMember(t *testing.T) {
	repo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	holdingCode := "rungrueng"
	owner := "owner@example.com"
	existing := testShopUser(holdingCode, "somchai01", models.ROLE_ADMIN)
	existing.UserUID = "55555555-5555-4555-8555-555555555555"
	repo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, owner).Return(testShopUser(holdingCode, owner, models.ROLE_OWNER), nil)
	repo.On("FindShopCreatedBy", ctx, holdingCode).Return(owner, nil)
	repo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, "somchai01").Return(existing, nil)

	err := shop.NewShopUserService(repo).SaveUserFullProfile(holdingCode, owner, &models.UserRoleRequest{
		Username: "SomChai01", Role: models.ROLE_USER,
	})

	require.EqualError(t, err, "member already exists")
	repo.AssertNotCalled(t, "SaveFullProfile", mock.Anything, mock.Anything, mock.Anything)
}

// An edit whose member disappeared (deleted in another tab) must answer "not found", not
// re-create the membership.
func TestShopUserEditOfMissingMemberIsNotFound(t *testing.T) {
	repo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	holdingCode := "rungrueng"
	owner := "owner@example.com"
	repo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, owner).Return(testShopUser(holdingCode, owner, models.ROLE_OWNER), nil)
	repo.On("FindShopCreatedBy", ctx, holdingCode).Return(owner, nil)
	repo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, "gone").Return(models.ShopUser{}, shop.ErrShopUserNotFound)
	repo.On("FindByHoldingCodeAndUserUID", ctx, holdingCode, "gone").Return(models.ShopUser{}, shop.ErrShopUserNotFound)

	err := shop.NewShopUserService(repo).SaveUserFullProfile(holdingCode, owner, &models.UserRoleRequest{
		EditUsername: "gone", Username: "gone", Role: models.ROLE_USER,
	})

	require.EqualError(t, err, "user not found")
	repo.AssertNotCalled(t, "SaveFullProfile", mock.Anything, mock.Anything, mock.Anything)
}

// The Holding must always keep a working owner: its creator cannot be given an expiry date.
func TestShopUserCreatorCannotGetAccessExpiry(t *testing.T) {
	repo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	holdingCode := "rungrueng"
	owner := "owner@example.com"
	creator := testShopUser(holdingCode, "creator@example.com", models.ROLE_OWNER)
	repo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, owner).Return(testShopUser(holdingCode, owner, models.ROLE_OWNER), nil)
	repo.On("FindShopCreatedBy", ctx, holdingCode).Return("creator@example.com", nil)
	repo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, "creator@example.com").Return(creator, nil)

	err := shop.NewShopUserService(repo).SaveUserFullProfile(holdingCode, owner, &models.UserRoleRequest{
		EditUsername: "creator@example.com", Username: "creator@example.com", Role: models.ROLE_OWNER, AccessExpiryDate: "2026-12-31",
	})

	require.EqualError(t, err, "creator access expiry not allowed")
	repo.AssertNotCalled(t, "SaveFullProfile", mock.Anything, mock.Anything, mock.Anything)
}

func TestShopUserSaveFullProfilePreservesUserUIDWhenUsernameChanges(t *testing.T) {
	shopUserRepo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	holdingCode := "holdingcode"
	authUsername := "owner@example.com"
	oldUsername := "old@example.com"
	newUsername := "new@example.com"
	userUID := "stable-user-uid"

	target := testShopUser(holdingCode, oldUsername, models.ROLE_ADMIN)
	target.UserUID = userUID

	shopUserRepo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, authUsername).Return(testShopUser(holdingCode, authUsername, models.ROLE_OWNER), nil)
	shopUserRepo.On("FindShopCreatedBy", ctx, holdingCode).Return(authUsername, nil)
	shopUserRepo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, oldUsername).Return(target, nil)
	shopUserRepo.On("SaveFullProfile", ctx, holdingCode, mock.MatchedBy(func(req *models.UserRoleRequest) bool {
		return req.Username == newUsername &&
			req.EditUsername == oldUsername &&
			req.UserUID == userUID
	})).Return(nil)

	shopUserSvc := shop.NewShopUserService(shopUserRepo)

	err := shopUserSvc.SaveUserFullProfile(holdingCode, authUsername, &models.UserRoleRequest{
		EditUsername: oldUsername,
		Username:     newUsername,
		Role:         models.ROLE_ADMIN,
	})

	require.NoError(t, err)
}

func TestShopUserSaveFullProfileRejectsScopeBeyondAdmin(t *testing.T) {
	shopUserRepo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	holdingCode := "holdingcode"
	adminUsername := "admin@example.com"
	targetUsername := "staff@example.com"

	admin := testShopUser(holdingCode, adminUsername, models.ROLE_ADMIN)
	admin.AccessScopes = []models.AccessScope{{ScopeType: "company", CompanyUID: "01", BusinessCode: "01", AllBranches: true}}
	shopUserRepo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, adminUsername).Return(admin, nil)
	shopUserRepo.On("FindShopCreatedBy", ctx, holdingCode).Return("owner@example.com", nil)
	shopUserRepo.On("FindByHoldingCodeAndUsername", ctx, holdingCode, targetUsername).Return(testShopUser(holdingCode, targetUsername, models.ROLE_USER), nil)

	err := shop.NewShopUserService(shopUserRepo).SaveUserFullProfile(holdingCode, adminUsername, &models.UserRoleRequest{
		Username:     targetUsername,
		Role:         models.ROLE_USER,
		AccessScopes: []models.AccessScope{{ScopeType: "holding"}},
	})

	require.EqualError(t, err, "ss_err_access_scope_exceeds_grantor")
	shopUserRepo.AssertNotCalled(t, "SaveFullProfile", ctx, holdingCode, mock.Anything)
}

func TestEnsureHoldingManagerRejectsInactiveUser(t *testing.T) {
	tests := []struct {
		name string
		user models.ShopUser
	}{
		{name: "disabled", user: models.ShopUser{IsAccessDisabled: true}},
		{name: "expired", user: models.ShopUser{AccessExpiryDate: time.Now().Add(-time.Minute)}},
		// Expiry is the start of the expiry date: reaching it exactly already closes access.
		{name: "expires now", user: models.ShopUser{AccessExpiryDate: time.Now()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(ShopUserRepositoryMock)
			tt.user.HoldingCode = "holdingcode"
			tt.user.Username = "admin@example.com"
			tt.user.Role = models.ROLE_ADMIN
			repo.On("FindByHoldingCodeAndUsername", context.Background(), tt.user.HoldingCode, tt.user.Username).Return(tt.user, nil)

			err := shop.NewShopUserService(repo).EnsureHoldingManager(tt.user.HoldingCode, tt.user.Username)

			require.EqualError(t, err, "permission denied")
		})
	}
}

func TestListShopByUserUsesUIDForCreatorWithoutAccessExemption(t *testing.T) {
	repo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	pageable := micromodels.Pageable{Page: 1, Limit: 20}
	docList := []models.ShopUserInfo{
		{CreatedBy: "stable-user-uid", IsAccessDisabled: true},
		{CreatedBy: "renamed@example.com"},
	}
	repo.On("FindByUserUIDPage", ctx, "stable-user-uid", pageable).Return(docList, common.PaginationData{}, nil)

	got, _, err := shop.NewShopUserService(repo).ListShopByUser("renamed@example.com", "stable-user-uid", pageable)

	require.NoError(t, err)
	require.True(t, got[0].IsCreator)
	require.True(t, got[0].IsAccessDisabled)
	require.False(t, got[1].IsCreator)
	repo.AssertNotCalled(t, "FindByUsernamePage", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func testShopUser(holdingCode string, username string, role models.UserRole) models.ShopUser {
	shopUser := models.ShopUser{}
	shopUser.HoldingCode = holdingCode
	shopUser.Username = username
	shopUser.Role = role
	return shopUser
}

func requireAccessDisabledRequest(t *testing.T, username string, disabledBy string) interface{} {
	t.Helper()
	return mock.MatchedBy(func(req *models.UserRoleRequest) bool {
		return req.Username == username &&
			req.IsAccessDisabled &&
			req.AccessDisabledBy == disabledBy &&
			!req.AccessDisabledAt.IsZero() &&
			req.AccessEnabledAt.Equal(time.Time{}) &&
			req.AccessEnabledBy == ""
	})
}

// TestInfoShopByUserResolvesUserUID - the settings screen opens a member by useruid; the PostgreSQL
// repo returns ErrShopUserNotFound for the username lookup, so the service must fall back to the uid
// (regression: "find failed" when editing any login account, 2026-09-23).
func TestInfoShopByUserResolvesUserUID(t *testing.T) {
	repo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	member := testShopUser("rungrueng", "uat_admin", models.ROLE_ADMIN)
	member.UserUID = "acceddcf-f223-4e5a-a863-f9259b0b4298"
	member.PermissionSets = []string{}
	profile := models.UserProfile{}
	profile.Name = "วิภาวดี ผู้ดูแลบัญชี"
	repo.On("FindByHoldingCodeAndUsername", ctx, "rungrueng", member.UserUID).Return(models.ShopUser{}, shop.ErrShopUserNotFound)
	repo.On("FindByHoldingCodeAndUserUID", ctx, "rungrueng", member.UserUID).Return(member, nil)
	repo.On("FindUserProfileByUsernames", ctx, []string{"uat_admin"}).Return([]models.UserProfile{profile}, nil)
	repo.On("FindShopCreatedBy", ctx, "rungrueng").Return("demo", nil)

	got, err := shop.NewShopUserService(repo).InfoShopByUser("rungrueng", member.UserUID)

	require.NoError(t, err)
	require.Equal(t, "uat_admin", got.Username)
	require.Equal(t, member.UserUID, got.UserUID)
	require.Equal(t, models.ROLE_ADMIN, got.Role)
	require.Equal(t, "วิภาวดี ผู้ดูแลบัญชี", got.UserProfileName)
	require.False(t, got.IsCreator)
	repo.AssertExpectations(t)
}

// A company-restricted ADMIN may only change or remove members whose CURRENT scope already fits
// inside their own (review 2026-09-24): checking only the requested scope let them narrow a
// holding-wide ADMIN or pull another company's staff into their company.
func companyAdminFixture(t *testing.T) (*ShopUserRepositoryMock, string, string) {
	t.Helper()
	repo := new(ShopUserRepositoryMock)
	holdingCode := "rungrueng"
	adminUsername := "admin-a@example.com"
	admin := testShopUser(holdingCode, adminUsername, models.ROLE_ADMIN)
	admin.UserUID = "11111111-1111-4111-8111-111111111111"
	admin.AccessScopes = []models.AccessScope{companyScope("A")}
	repo.On("FindByHoldingCodeAndUsername", context.Background(), holdingCode, adminUsername).Return(admin, nil)
	repo.On("FindShopCreatedBy", context.Background(), holdingCode).Return("owner@example.com", nil)
	return repo, holdingCode, adminUsername
}

func companyScope(company string) models.AccessScope {
	return models.AccessScope{ScopeType: "company", CompanyUID: company, BusinessCode: company, AllBranches: true}
}

func memberWithScope(holdingCode, username, uid string, role models.UserRole, scopes ...models.AccessScope) models.ShopUser {
	member := testShopUser(holdingCode, username, role)
	member.UserUID = uid
	member.AccessScopes = scopes
	return member
}

func TestCompanyAdminCannotNarrowHoldingWideAdmin(t *testing.T) {
	repo, holdingCode, adminUsername := companyAdminFixture(t)
	target := memberWithScope(holdingCode, "group-admin", "22222222-2222-4222-8222-222222222222", models.ROLE_ADMIN, models.AccessScope{ScopeType: "holding"})
	repo.On("FindByHoldingCodeAndUsername", context.Background(), holdingCode, "group-admin").Return(target, nil)

	for name, req := range map[string]*models.UserRoleRequest{
		"narrow to own company": {Username: "group-admin", Role: models.ROLE_ADMIN, AccessScopes: []models.AccessScope{companyScope("A")}},
		"demote to staff":       {Username: "group-admin", Role: models.ROLE_USER, AccessScopes: []models.AccessScope{companyScope("A")}},
		"strip every scope":     {Username: "group-admin", Role: models.ROLE_USER},
	} {
		t.Run(name, func(t *testing.T) {
			err := shop.NewShopUserService(repo).SaveUserFullProfile(holdingCode, adminUsername, req)
			require.EqualError(t, err, "ss_err_access_scope_exceeds_grantor")
		})
	}
	repo.AssertNotCalled(t, "SaveFullProfile", mock.Anything, mock.Anything, mock.Anything)
}

func TestCompanyAdminCannotModifyOrDeleteOtherCompanyStaff(t *testing.T) {
	repo, holdingCode, adminUsername := companyAdminFixture(t)
	target := memberWithScope(holdingCode, "staff-b", "33333333-3333-4333-8333-333333333333", models.ROLE_USER, companyScope("B"))
	repo.On("FindByHoldingCodeAndUsername", context.Background(), holdingCode, "staff-b").Return(target, nil)
	svc := shop.NewShopUserService(repo)

	err := svc.SaveUserFullProfile(holdingCode, adminUsername, &models.UserRoleRequest{
		Username: "staff-b", Role: models.ROLE_USER, AccessScopes: []models.AccessScope{companyScope("A")},
	})
	require.EqualError(t, err, "ss_err_access_scope_exceeds_grantor")

	err = svc.DeleteUserPermissionShop(holdingCode, adminUsername, "staff-b")
	require.EqualError(t, err, "ss_err_access_scope_exceeds_grantor")

	repo.AssertNotCalled(t, "SaveFullProfile", mock.Anything, mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
}

// A bogus editusername must not skip the checks of the member the upsert would overwrite:
// the repository falls back to the username, so the service must check that member too.
func TestCompanyAdminCannotBypassTargetCheckWithUnknownEditID(t *testing.T) {
	repo, holdingCode, adminUsername := companyAdminFixture(t)
	target := memberWithScope(holdingCode, "group-admin", "22222222-2222-4222-8222-222222222222", models.ROLE_ADMIN, models.AccessScope{ScopeType: "holding"})
	repo.On("FindByHoldingCodeAndUsername", context.Background(), holdingCode, "ghost").Return(models.ShopUser{}, shop.ErrShopUserNotFound)
	repo.On("FindByHoldingCodeAndUserUID", context.Background(), holdingCode, "ghost").Return(models.ShopUser{}, shop.ErrShopUserNotFound)
	repo.On("FindByHoldingCodeAndUsername", context.Background(), holdingCode, "group-admin").Return(target, nil)

	err := shop.NewShopUserService(repo).SaveUserFullProfile(holdingCode, adminUsername, &models.UserRoleRequest{
		EditUsername: "ghost", Username: "group-admin", Role: models.ROLE_USER, AccessScopes: []models.AccessScope{companyScope("A")},
	})

	require.EqualError(t, err, "ss_err_access_scope_exceeds_grantor")
	repo.AssertNotCalled(t, "SaveFullProfile", mock.Anything, mock.Anything, mock.Anything)
}

func TestCompanyAdminCanStillEditAndDeleteOwnCompanyStaff(t *testing.T) {
	repo, holdingCode, adminUsername := companyAdminFixture(t)
	target := memberWithScope(holdingCode, "staff-a", "44444444-4444-4444-8444-444444444444", models.ROLE_USER, companyScope("A"))
	repo.On("FindByHoldingCodeAndUsername", context.Background(), holdingCode, "staff-a").Return(target, nil)
	repo.On("SaveFullProfile", context.Background(), holdingCode, mock.MatchedBy(func(req *models.UserRoleRequest) bool {
		return req.UserUID == target.UserUID && len(req.AccessScopes) == 1 && req.AccessScopes[0].ScopeType == "branch"
	})).Return(nil)
	// Delete addresses the member that was checked (its uid), not the raw path id, and records the admin.
	repo.On("Delete", context.Background(), holdingCode, target.UserUID, "11111111-1111-4111-8111-111111111111").Return(nil)
	svc := shop.NewShopUserService(repo)

	err := svc.SaveUserFullProfile(holdingCode, adminUsername, &models.UserRoleRequest{
		EditUsername: "staff-a", Username: "staff-a", Role: models.ROLE_USER,
		AccessScopes: []models.AccessScope{{ScopeType: "branch", CompanyUID: "A", BusinessCode: "A", BranchUID: "00001"}},
	})
	require.NoError(t, err)
	require.NoError(t, svc.DeleteUserPermissionShop(holdingCode, adminUsername, "staff-a"))
	repo.AssertExpectations(t)
}
