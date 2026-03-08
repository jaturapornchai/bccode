package api

import (
	"context"
	"smlcloudplatform/internal/shop"

	auth_model "smlcloudplatform/internal/authentication/models"
	shop_model "smlcloudplatform/internal/shop/models"
	"smlcloudplatform/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
)

func (m *MigrationService) ImportShop(shops []shop_model.ShopDoc) error {
	shopRepo := shop.NewShopRepository(m.mongoPersister)
	for _, shop := range shops {

		findShopDoc, err := shopRepo.FindByGuid(context.Background(), shop.GuidFixed)

		if err != nil {

		}

		if findShopDoc.GuidFixed == "" {
			_, err = shopRepo.Create(context.Background(), shop)
			if err != nil {
				return err
			}
		}

		err = m.ImportShopUser(shop)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MigrationService) ImportShopUser(shop shop_model.ShopDoc) error {

	username := shop.GuidFixed
	findUser := &auth_model.UserDoc{}

	err := m.mongoPersister.FindOne(context.Background(), &auth_model.UserDoc{}, bson.M{"username": username}, findUser)

	if err != nil {
		return err
	}

	if findUser.Username == "" {
		userPassword, err := utils.HashPassword(username)
		if err != nil {
			return err
		}
		newUser := auth_model.UserDoc{
			UserPassword: auth_model.UserPassword{
				Password: userPassword,
			},
		}
		newUser.Username = username

		_, err = m.mongoPersister.Create(context.Background(), &auth_model.UserDoc{}, newUser)
		if err != nil {
			return err
		}
	}

	findUserShop, err := m.userShopRepo.FindByShopIDAndUsername(context.Background(), shop.GuidFixed, username)

	if findUserShop.Username == "" {

		err = m.userShopRepo.Save(context.Background(), shop.GuidFixed, username, auth_model.ROLE_OWNER)

		if err != nil {
			return err
		}
	}

	return nil
}
