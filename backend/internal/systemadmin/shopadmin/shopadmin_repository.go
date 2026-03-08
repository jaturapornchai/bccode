package shopadmin

import (
	"context"
	shopModels "smlcloudplatform/internal/shop/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IShopAdminRepository interface {
	ListAllShop(ctx context.Context) ([]ShopDoc, error)
	CreateShop(ctx context.Context, doc shopModels.ShopDoc) error
	FindShopByProjectNo(ctx context.Context, projectNo string) (shopModels.ShopDoc, error)

	ListShopUsersAll(ctx context.Context) ([]ShopUserDoc, error)
	ListShopUsersByShopId(ctx context.Context, shopId string) ([]ShopUserDoc, error)
}

type ShopAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewShopAdminRepository(mongoPersister microservice.IPersisterMongo) IShopAdminRepository {
	return &ShopAdminRepository{
		pst: mongoPersister,
	}
}

func (repo *ShopAdminRepository) ListAllShop(ctx context.Context) ([]ShopDoc, error) {

	shopList := []ShopDoc{}

	err := repo.pst.Find(ctx, ShopDoc{}, bson.M{}, &shopList)

	if err != nil {
		return []ShopDoc{}, err
	}

	return shopList, nil
}

func (repo *ShopAdminRepository) CreateShop(ctx context.Context, doc shopModels.ShopDoc) error {
	_, err := repo.pst.Create(ctx, &shopModels.ShopDoc{}, doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo *ShopAdminRepository) FindShopByProjectNo(ctx context.Context, projectNo string) (shopModels.ShopDoc, error) {
	shopDoc := shopModels.ShopDoc{}
	err := repo.pst.FindOne(ctx, &shopModels.ShopDoc{}, bson.M{"branchcode": projectNo}, &shopDoc)
	if err != nil {
		return shopModels.ShopDoc{}, err
	}
	return shopDoc, nil
}

func (repo *ShopAdminRepository) ListShopUsersAll(ctx context.Context) ([]ShopUserDoc, error) {

	shopUserList := []ShopUserDoc{}

	err := repo.pst.Find(ctx, ShopUserDoc{}, bson.M{}, &shopUserList)

	if err != nil {
		return []ShopUserDoc{}, err
	}

	return shopUserList, nil
}

func (repo *ShopAdminRepository) ListShopUsersByShopId(ctx context.Context, shopId string) ([]ShopUserDoc, error) {

	shopUserList := []ShopUserDoc{}

	err := repo.pst.Find(ctx, ShopUserDoc{}, bson.M{"shopid": shopId}, &shopUserList)

	if err != nil {
		return []ShopUserDoc{}, err
	}

	return shopUserList, nil
}
