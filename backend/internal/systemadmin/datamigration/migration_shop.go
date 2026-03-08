package datamigration

import (
	"context"
	"smlcloudplatform/internal/shop"
	shopModel "smlcloudplatform/internal/shop/models"
	"smlcloudplatform/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
)

func (m *MigrationService) ImportShop(shops []shopModel.ShopDoc) error {
	shopRepo := shop.NewShopRepository(m.mongoPersister)
	for _, shop := range shops {

		findShopDoc, err := shopRepo.FindByGuid(context.TODO(), shop.GuidFixed)

		if err != nil {
			return err
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

func (m *MigrationService) ImportShopUser(shop shopModel.ShopDoc) error {

	username := shop.GuidFixed
	findUser := &shopModel.UserDoc{}

	err := m.mongoPersister.FindOne(context.TODO(), &shopModel.UserDoc{}, bson.M{"username": username}, findUser)

	if err != nil {
		return err
	}

	if findUser.Username == "" {
		userPassword, err := utils.HashPassword(username)
		if err != nil {
			return err
		}
		newUser := shopModel.UserDoc{
			UsernameCode: shopModel.UsernameCode{
				Username: username,
			},
			UserPassword: shopModel.UserPassword{
				Password: userPassword,
			},
		}
		_, err = m.mongoPersister.Create(context.Background(), &shopModel.UserDoc{}, newUser)
		if err != nil {
			return err
		}
	}

	findUserShop, err := m.userShopRepo.FindByShopIDAndUsername(context.TODO(), shop.GuidFixed, username)
	if err != nil {
		return err
	}

	if findUserShop.Username == "" {

		err = m.userShopRepo.Save(context.Background(), shop.GuidFixed, username, shopModel.ROLE_OWNER)

		if err != nil {
			return err
		}
	}

	return nil
}
