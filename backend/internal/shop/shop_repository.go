package shop

import (
	"context"
	"errors"
	organization "smlcloudplatform/internal/organization"
	"smlcloudplatform/internal/shop/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IShopRepository interface {
	Transaction(ctx context.Context, queryFunc func(context.Context) error) error
	EnsureBootstrapIndexes(ctx context.Context) error
	Create(ctx context.Context, shop models.ShopDoc) (string, error)
	CreateCodeClaim(ctx context.Context, claim organization.OrganizationCodeClaim) error
	CreateAudit(ctx context.Context, audit organization.OrganizationAudit) error
	CreateOutbox(ctx context.Context, event organization.OrganizationOutboxEvent) error
	Update(ctx context.Context, guid string, expectedVersion int64, expectedActive bool, shop models.ShopDoc) error
	FindByGuid(ctx context.Context, guid string) (models.ShopDoc, error)
	FindByHoldingCode(ctx context.Context, holdingCode string) (models.ShopDoc, error)
	FindPage(ctx context.Context, pageable micromodels.Pageable) ([]models.ShopInfo, mongopagination.PaginationData, error)
	Delete(ctx context.Context, guid string, username string) error
}

func (repo ShopRepository) Transaction(ctx context.Context, queryFunc func(context.Context) error) error {
	return repo.pst.Transaction(ctx, queryFunc)
}

func (repo ShopRepository) EnsureBootstrapIndexes(ctx context.Context) error {
	_, err := repo.pst.CreateIndex(ctx, organization.OrganizationCodeClaim{}, "uniq_organizationcodeclaims_scope_code", bson.D{
		{Key: "entitytype", Value: 1},
		{Key: "scopeuid", Value: 1},
		{Key: "normalizedcode", Value: 1},
	})
	return err
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

func (repo ShopRepository) CreateCodeClaim(ctx context.Context, claim organization.OrganizationCodeClaim) error {
	_, err := repo.pst.Create(ctx, organization.OrganizationCodeClaim{}, claim)
	return err
}

func (repo ShopRepository) CreateAudit(ctx context.Context, audit organization.OrganizationAudit) error {
	_, err := repo.pst.Create(ctx, organization.OrganizationAudit{}, audit)
	return err
}

func (repo ShopRepository) CreateOutbox(ctx context.Context, event organization.OrganizationOutboxEvent) error {
	_, err := repo.pst.Create(ctx, organization.OrganizationOutboxEvent{}, event)
	return err
}

func (repo ShopRepository) Update(ctx context.Context, guid string, expectedVersion int64, expectedActive bool, shop models.ShopDoc) error {
	collection, err := repo.pst.Exec(ctx, &models.ShopDoc{})
	if err != nil {
		return err
	}
	result, err := collection.UpdateOne(ctx, bson.M{
		"guidfixed": guid,
		"__v":       expectedVersion,
		"isactive":  expectedActive,
	}, bson.M{
		"$set": bson.M{
			"holdingcode":          shop.HoldingCode,
			"profilepicture":       shop.ProfilePicture,
			"name1":                shop.Name1,
			"names":                shop.Names,
			"telephone":            shop.Telephone,
			"branchcode":           shop.BranchCode,
			"ismainshop":           shop.IsMainShop,
			"posproductcentertype": shop.PosProductCenterType,
			"productcentertype":    shop.ProductCenterType,
			"debtorcentertype":     shop.DebtorCenterType,
			"mainholdingcode":      shop.MainHoldingCode,
			"address":              shop.Address,
			"images":               shop.Images,
			"logo":                 shop.Logo,
			"settings":             shop.Settings,
			"isbcmember":           shop.IsBcMember,
			"promptshopinfo":       shop.PromptShopInfo,
			"updatedby":            shop.UpdatedBy,
			"updatedat":            shop.UpdatedAt,
		},
		"$inc": bson.M{"__v": 1},
	})
	if err != nil {
		return err
	}
	if result.MatchedCount != 1 {
		return organization.ErrStatusChangeConflict
	}
	return nil
}

func (repo ShopRepository) FindByGuid(ctx context.Context, guid string) (models.ShopDoc, error) {
	findShop := &models.ShopDoc{}
	err := repo.pst.FindOne(ctx, &models.ShopDoc{}, bson.M{"guidfixed": guid, "deletedat": bson.M{"$exists": false}}, findShop)

	if err != nil {
		return repo.FindByHoldingCode(ctx, guid)
	}
	if strings.TrimSpace(findShop.GuidFixed) == "" {
		return repo.FindByHoldingCode(ctx, guid)
	}
	return *findShop, err
}

func (repo ShopRepository) FindByHoldingCode(ctx context.Context, holdingCode string) (models.ShopDoc, error) {
	findShop := &models.ShopDoc{}
	err := repo.pst.FindOne(ctx, &models.ShopDoc{}, bson.M{"holdingcode": holdingCode, "deletedat": bson.M{"$exists": false}}, findShop)

	if err != nil {
		return models.ShopDoc{}, err
	}
	if strings.TrimSpace(findShop.GuidFixed) == "" {
		return models.ShopDoc{}, errors.New("holding not found")
	}
	return *findShop, nil
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
	err := repo.pst.SoftDelete(ctx, &models.ShopDoc{}, username, bson.M{"guidfixed": guid, "deletedat": bson.M{"$exists": false}})
	if err != nil {
		return err
	}
	return nil
}
