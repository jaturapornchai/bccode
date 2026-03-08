package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/shopcoupon/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type IShopCouponRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ShopCouponDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ShopCouponDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ShopCouponDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ShopCouponInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ShopCouponDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ShopCouponInfo, mongopagination.PaginationData, error)
}

type ShopCouponRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ShopCouponDoc]
	repositories.SearchRepository[models.ShopCouponInfo]
	repositories.GuidRepository[models.ShopCouponItemGuid]
}

func NewShopCouponRepository(pst microservice.IPersisterMongo) *ShopCouponRepository {

	insRepo := &ShopCouponRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ShopCouponDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ShopCouponInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ShopCouponItemGuid](pst)

	return insRepo
}
