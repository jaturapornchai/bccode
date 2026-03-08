package repositories

import (
	"context"
	"smlcloudplatform/internal/product/ordertype/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IOrderTypeRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.OrderTypeDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.OrderTypeDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.OrderTypeDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.OrderTypeInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.OrderTypeDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.OrderTypeDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.OrderTypeItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.OrderTypeDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.OrderTypeInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.OrderTypeInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.OrderTypeDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.OrderTypeActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.OrderTypeDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.OrderTypeActivity, error)
}

type OrderTypeRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.OrderTypeDoc]
	repositories.SearchRepository[models.OrderTypeInfo]
	repositories.GuidRepository[models.OrderTypeItemGuid]
	repositories.ActivityRepository[models.OrderTypeActivity, models.OrderTypeDeleteActivity]
}

func NewOrderTypeRepository(pst microservice.IPersisterMongo) *OrderTypeRepository {

	insRepo := &OrderTypeRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.OrderTypeDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.OrderTypeInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.OrderTypeItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.OrderTypeActivity, models.OrderTypeDeleteActivity](pst)

	return insRepo
}
