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

	mockShopID := "MockShopID"

	authUser := "auth_user"

	ctx := context.Background()

	mockShopUserAuth := models.ShopUser{}
	mockShopUserAuth.ShopID = mockShopID
	mockShopUserAuth.Username = authUser
	mockShopUserAuth.Role = models.ROLE_OWNER

	shopUserRepo.On("FindByShopIDAndUsername", ctx, mockShopID, authUser).Return(mockShopUserAuth, nil)
	shopUserRepo.On("Save", ctx, mockShopID, "user_create", models.ROLE_OWNER).Return(nil)

	shopUserSvc := shop.NewShopUserService(shopUserRepo)

	err := shopUserSvc.SaveUserPermissionShop(mockShopID, authUser, "", "user_create", models.ROLE_OWNER)

	require.NoError(t, err)

}

func TestShopUserDeleteCannotDeleteCreator(t *testing.T) {
	shopUserRepo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	shopID := "shop_id"
	authUsername := "owner@example.com"
	creatorUsername := "creator@example.com"

	shopUserRepo.On("FindByShopIDAndUsername", ctx, shopID, authUsername).Return(testShopUser(shopID, authUsername, models.ROLE_OWNER), nil)
	shopUserRepo.On("FindByShopIDAndUsername", ctx, shopID, creatorUsername).Return(testShopUser(shopID, creatorUsername, models.ROLE_OWNER), nil)
	shopUserRepo.On("FindShopCreatedBy", ctx, shopID).Return(creatorUsername, nil)

	shopUserSvc := shop.NewShopUserService(shopUserRepo)

	err := shopUserSvc.DeleteUserPermissionShop(shopID, authUsername, creatorUsername)

	require.EqualError(t, err, "creator_cannot_delete")
	shopUserRepo.AssertNotCalled(t, "Delete", ctx, shopID, creatorUsername)
}

func TestShopUserSaveFullProfileCreatorCannotBeDisabled(t *testing.T) {
	shopUserRepo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	shopID := "shop_id"
	authUsername := "owner@example.com"
	creatorUsername := "creator@example.com"

	shopUserRepo.On("FindByShopIDAndUsername", ctx, shopID, authUsername).Return(testShopUser(shopID, authUsername, models.ROLE_OWNER), nil)
	shopUserRepo.On("FindShopCreatedBy", ctx, shopID).Return(creatorUsername, nil)
	shopUserRepo.On("FindByShopIDAndUsername", ctx, shopID, creatorUsername).Return(testShopUser(shopID, creatorUsername, models.ROLE_OWNER), nil)

	shopUserSvc := shop.NewShopUserService(shopUserRepo)

	err := shopUserSvc.SaveUserFullProfile(shopID, authUsername, &models.UserRoleRequest{
		Username:         creatorUsername,
		Role:             models.ROLE_OWNER,
		IsAccessDisabled: true,
	})

	require.EqualError(t, err, "creator_access_cannot_be_disabled")
	shopUserRepo.AssertNotCalled(t, "SaveFullProfile", ctx, shopID)
}

func TestShopUserSaveFullProfileDisablesMember(t *testing.T) {
	shopUserRepo := new(ShopUserRepositoryMock)
	ctx := context.Background()
	shopID := "shop_id"
	authUsername := "owner@example.com"
	targetUsername := "member@example.com"

	shopUserRepo.On("FindByShopIDAndUsername", ctx, shopID, authUsername).Return(testShopUser(shopID, authUsername, models.ROLE_OWNER), nil)
	shopUserRepo.On("FindShopCreatedBy", ctx, shopID).Return(authUsername, nil)
	shopUserRepo.On("FindByShopIDAndUsername", ctx, shopID, targetUsername).Return(testShopUser(shopID, targetUsername, models.ROLE_ADMIN), nil)
	shopUserRepo.On("FindByShopIDAndLineUserID", ctx, shopID, "line-id").Return(models.ShopUser{}, errors.New("not found"))
	shopUserRepo.On("SaveFullProfile", ctx, shopID, requireAccessDisabledRequest(t, targetUsername, authUsername)).Return(nil)

	shopUserSvc := shop.NewShopUserService(shopUserRepo)

	err := shopUserSvc.SaveUserFullProfile(shopID, authUsername, &models.UserRoleRequest{
		Username:         targetUsername,
		Role:             models.ROLE_ADMIN,
		LineUserID:       "line-id",
		IsAccessDisabled: true,
	})

	require.NoError(t, err)
}

func testShopUser(shopID string, username string, role models.UserRole) models.ShopUser {
	shopUser := models.ShopUser{}
	shopUser.ShopID = shopID
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
