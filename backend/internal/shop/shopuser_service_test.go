package shop_test

import (
	"context"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/shop"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShopUserSave(t *testing.T) {
	shopUserRepo := new(ShopUserRepositoryMock)

	mockShopID := "MockShopID"

	authUser := "auth_user"
	updateUser := "update_user"

	ctx := context.Background()
	shopUserRepo.On("Update", ctx, 1, mockShopID, "user_update", models.ROLE_OWNER).Return(nil)

	mockShopUser := models.ShopUser{}
	mockShopUser.ShopID = mockShopID
	mockShopUser.Username = "user_create"
	mockShopUser.Role = models.ROLE_OWNER

	shopUserRepo.On("Create", ctx, mockShopUser).Return(nil)

	mockFindShopUser := models.ShopUser{}
	mockFindShopUser.ShopID = mockShopID
	mockFindShopUser.Username = updateUser
	mockFindShopUser.Role = models.ROLE_OWNER

	shopUserRepo.On("FindByShopIDAndUsername", ctx, mockShopID, updateUser).Return(mockFindShopUser, nil)

	mockShopUserAuth := models.ShopUser{}
	mockShopUserAuth.ShopID = mockShopID
	mockShopUserAuth.Username = updateUser
	mockShopUserAuth.Role = models.ROLE_OWNER

	shopUserRepo.On("FindByShopIDAndUsername", ctx, mockShopID, updateUser).Return(mockShopUserAuth, nil)

	shopUserSvc := shop.NewShopUserService(shopUserRepo)

	err := shopUserSvc.SaveUserPermissionShop(mockShopID, authUser, "", "user_create", models.ROLE_OWNER)

	require.NoError(t, err)

}
