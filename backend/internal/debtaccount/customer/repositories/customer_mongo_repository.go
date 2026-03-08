package repositories

import (
	"context"
	"smlcloudplatform/internal/debtaccount/customer/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ICustomerRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.CustomerDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.CustomerDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.CustomerDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.CustomerInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.CustomerDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.CustomerItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.CustomerDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.CustomerInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.CustomerInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CustomerDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CustomerActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CustomerDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CustomerActivity, error)
}

type CustomerRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.CustomerDoc]
	repositories.SearchRepository[models.CustomerInfo]
	repositories.GuidRepository[models.CustomerItemGuid]
	repositories.ActivityRepository[models.CustomerActivity, models.CustomerDeleteActivity]
}

func NewCustomerRepository(pst microservice.IPersisterMongo) *CustomerRepository {

	insRepo := &CustomerRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.CustomerDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.CustomerInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.CustomerItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.CustomerActivity, models.CustomerDeleteActivity](pst)

	return insRepo
}
