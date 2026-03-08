package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/storefront/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type IStorefrontRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.StorefrontDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.StorefrontDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.StorefrontDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.StorefrontInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.StorefrontDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.StorefrontItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.StorefrontDoc, error)
}

type StorefrontRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.StorefrontDoc]
	repositories.SearchRepository[models.StorefrontInfo]
	repositories.GuidRepository[models.StorefrontItemGuid]
}

func NewStorefrontRepository(pst microservice.IPersisterMongo) *StorefrontRepository {

	insRepo := &StorefrontRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.StorefrontDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.StorefrontInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.StorefrontItemGuid](pst)

	return insRepo
}
