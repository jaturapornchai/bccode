package shop

import (
	"context"
	"smlcloudplatform/internal/shop/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IShopRepository interface {
	Create(ctx context.Context, shop models.ShopDoc) (string, error)
	Update(ctx context.Context, guid string, shop models.ShopDoc) error
	FindByGuid(ctx context.Context, guid string) (models.ShopDoc, error)
	FindPage(ctx context.Context, pageable micromodels.Pageable) ([]models.ShopInfo, mongopagination.PaginationData, error)
	Delete(ctx context.Context, guid string, username string) error
}

type ShopRepository struct {
	pst microservice.IPersisterMongo
}

func NewShopRepository(pst microservice.IPersisterMongo) ShopRepository {
	return ShopRepository{
		pst: pst,
	}
}

func (repo ShopRepository) Create(ctx context.Context, shop models.ShopDoc) (string, error) {
	idx, err := repo.pst.Create(ctx, &models.ShopDoc{}, shop)
	if err != nil {
		return "", err
	}
	return idx.Hex(), nil
}

func (repo ShopRepository) Update(ctx context.Context, guid string, shop models.ShopDoc) error {
	filterDoc := map[string]interface{}{
		"guidfixed": guid,
	}
	err := repo.pst.UpdateOne(ctx, &models.ShopDoc{}, filterDoc, shop)

	if err != nil {
		return err
	}

	return nil
}

func (repo ShopRepository) FindByGuid(ctx context.Context, guid string) (models.ShopDoc, error) {
	findShop := &models.ShopDoc{}
	err := repo.pst.FindOne(ctx, &models.ShopDoc{}, bson.M{"guidfixed": guid, "deletedat": bson.M{"$exists": false}}, findShop)

	if err != nil {
		return models.ShopDoc{}, err
	}
	return *findShop, err
}

func (repo ShopRepository) FindPage(ctx context.Context, pageable micromodels.Pageable) ([]models.ShopInfo, mongopagination.PaginationData, error) {
	filterQueries := bson.M{
		"deletedat": bson.M{"$exists": false},
		"name1": bson.M{"$regex": primitive.Regex{
			Pattern: ".*" + pageable.Query + ".*",
			Options: "",
		}}}

	shopList := []models.ShopInfo{}

	pagination, err := repo.pst.FindPage(ctx, &models.ShopInfo{}, filterQueries, pageable, &shopList)

	if err != nil {
		return []models.ShopInfo{}, mongopagination.PaginationData{}, err
	}

	return shopList, pagination, nil
}

func (repo ShopRepository) Delete(ctx context.Context, guid string, username string) error {
	err := repo.pst.SoftDeleteByID(ctx, &models.ShopInfo{}, guid, username)
	if err != nil {
		return err
	}
	return nil
}
