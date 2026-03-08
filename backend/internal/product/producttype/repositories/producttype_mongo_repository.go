package repositories

import (
	"context"
	"smlcloudplatform/internal/product/producttype/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IProductTypeRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ProductTypeDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ProductTypeDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ProductTypeDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductTypeInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ProductTypeDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.ProductTypeDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ProductTypeItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ProductTypeDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductTypeInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ProductTypeInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductTypeDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductTypeActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductTypeDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductTypeActivity, error)
}

type ProductTypeRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ProductTypeDoc]
	repositories.SearchRepository[models.ProductTypeInfo]
	repositories.GuidRepository[models.ProductTypeItemGuid]
	repositories.ActivityRepository[models.ProductTypeActivity, models.ProductTypeDeleteActivity]
}

func NewProductTypeRepository(pst microservice.IPersisterMongo) *ProductTypeRepository {

	insRepo := &ProductTypeRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ProductTypeDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ProductTypeInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ProductTypeItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ProductTypeActivity, models.ProductTypeDeleteActivity](pst)

	return insRepo
}
