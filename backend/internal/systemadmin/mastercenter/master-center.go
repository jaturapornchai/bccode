package mastercenter

import (
	"context"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/models"
	"smlcloudplatform/internal/shop"
	shopModel "smlcloudplatform/internal/shop/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IMasterCenter interface {
	InitMasterCenter()
}
type MasterCenter struct {
	mongoPersister microservice.IPersisterMongo
	shopRepo       shop.IShopRepository
	userShopRepo   shop.IShopUserRepository
	ms             *microservice.Microservice
}

func NewMasterCenter(ms *microservice.Microservice, cfg config.IConfig) *MasterCenter {

	mongoPersister := ms.MongoPersister(cfg.MongoPersisterConfig())

	return &MasterCenter{
		ms:             ms,
		mongoPersister: mongoPersister,
		shopRepo:       shop.NewShopRepository(mongoPersister),
		userShopRepo:   shop.NewShopUserRepository(mongoPersister),
	}
}

func (m *MasterCenter) InitMasterCenter() error {
	m.ms.Logger.Debug("Init Master Shop")
	err := m.CheckMasterShop()
	if err != nil {
		m.ms.Logger.Errorf("Check Master Shop Failed %v", err)
	}

	return nil
}

func (m *MasterCenter) CheckMasterShop() error {

	// MasterCenterUserName : vfmastercenter
	// MasterCenterPassword : L8Fs4rvXaTbD

	masterShop := shopModel.ShopDoc{
		ShopInfo: shopModel.ShopInfo{
			DocIdentity: models.DocIdentity{
				GuidFixed: "999999999",
			},
			Shop: shopModel.Shop{
				Name1: "Master Shop",
			},
		},
	}

	findShopDoc, err := m.shopRepo.FindByGuid(context.TODO(), masterShop.GuidFixed)
	if err != nil {
		return err
	}
	if findShopDoc.GuidFixed == "" {
		// create master
		_, err = m.shopRepo.Create(context.Background(), masterShop)
		if err != nil {
			return err
		}
	}

	masterUser := &shopModel.UserDoc{
		UsernameCode: shopModel.UsernameCode{
			Username: "vfmastercenter",
		},
		// UserPassword: shopModel.UserPassword{
		// 	Password: userPassword,
		// },
	}

	findUser := &shopModel.UserDoc{}
	err = m.mongoPersister.FindOne(context.TODO(), &shopModel.UserDoc{}, bson.M{"username": masterUser.Username}, findUser)
	if err != nil {
		return err
	}

	if findUser.Username == "" {
		masterUser.Password, _ = utils.HashPassword("L8Fs4rvXaTbD")
		_, err = m.mongoPersister.Create(context.Background(), &shopModel.UserDoc{}, masterUser)
		if err != nil {
			return err
		}

		findUserShop, err := m.userShopRepo.FindByShopIDAndUsername(context.TODO(), masterShop.GuidFixed, masterUser.Username)
		if err != nil {
			return err
		}

		if findUserShop.Username == "" {
			err = m.userShopRepo.Save(context.Background(), masterShop.GuidFixed, masterUser.Username, shopModel.ROLE_OWNER)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
