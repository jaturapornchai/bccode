package shop_test

import (
	"context"
	"errors"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/shop"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShopUserSave(t *testing.T) {
	shopUserRepo := new(ShopUserRepositoryMock)

	mockHoldingCode := "MockHoldingCode"

	authUser := "auth_user"

	ctx := context.Background()

	mockShopUserAuth := models.ShopUser{}
	mockShopUserAuth.HoldingCode = mockHoldingCode
	mockShopUserAuth.Username = authUser
	mockShopUserAuth.Role = models.ROLE_OWNER

	shopUserRepo.On("FindByHoldingCodeAndUsername", ctx, mockHoldingCode, authUser).Return(mockShopUserAuth, nil)
	shopUserRepo.On("Save", ctx, mockHoldingCode, "user_create", models.ROLE_OWNER).Return(nil)

	shopUserSvc := shop.NewShopUserService(shopUserRepo)

	err := shopUserSvc.SaveUserPermissionShop(mockHoldingCode, authUser, "", "user_create", models.ROLE_OWNER)

	require.NoError(t, err)

}

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
	shopUserRepo.On("FindByHoldingCodeAndLineUserID", ctx, holdingCode, "line-id").Return(models.ShopUser{}, errors.New("not found"))
	shopUserRepo.On("SaveFullProfile", ctx, holdingCode, requireAccessDisabledRequest(t, targetUsername, authUsername)).Return(nil)

	shopUserSvc := shop.NewShopUserService(shopUserRepo)

	err := shopUserSvc.SaveUserFullProfile(holdingCode, authUsername, &models.UserRoleRequest{
		Username:         targetUsername,
		Role:             models.ROLE_ADMIN,
		LineUserID:       "line-id",
		IsAccessDisabled: true,
	})

	require.NoError(t, err)
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

func TestEnsureHoldingManagerRejectsInactiveUser(t *testing.T) {
	tests := []struct {
		name string
		user models.ShopUser
	}{
		{name: "disabled", user: models.ShopUser{IsAccessDisabled: true}},
		{name: "expired", user: models.ShopUser{AccessExpiryDate: time.Now().Add(-time.Minute)}},
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
