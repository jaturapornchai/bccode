package repositories

import (
	"context"
	"smlcloudplatform/internal/product/productgroup/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IProductGroupRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ProductGroupDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ProductGroupDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ProductGroupDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductGroupInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ProductGroupDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.ProductGroupDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ProductGroupItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ProductGroupDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductGroupInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ProductGroupInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductGroupDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductGroupActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductGroupDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductGroupActivity, error)
}

type ProductGroupRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ProductGroupDoc]
	repositories.SearchRepository[models.ProductGroupInfo]
	repositories.GuidRepository[models.ProductGroupItemGuid]
	repositories.ActivityRepository[models.ProductGroupActivity, models.ProductGroupDeleteActivity]
}

func NewProductGroupRepository(pst microservice.IPersisterMongo) *ProductGroupRepository {

	insRepo := &ProductGroupRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ProductGroupDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ProductGroupInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ProductGroupItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ProductGroupActivity, models.ProductGroupDeleteActivity](pst)

	return insRepo
}
