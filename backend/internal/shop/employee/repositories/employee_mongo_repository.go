package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/shop/employee/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IEmployeeRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.EmployeeDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.EmployeeDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.EmployeeDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.EmployeeInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.EmployeeDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.EmployeeItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.EmployeeDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.EmployeeInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.EmployeeInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.EmployeeDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.EmployeeActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.EmployeeDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.EmployeeActivity, error)
}

type EmployeeRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.EmployeeDoc]
	repositories.SearchRepository[models.EmployeeInfo]
	repositories.GuidRepository[models.EmployeeItemGuid]
	repositories.ActivityRepository[models.EmployeeActivity, models.EmployeeDeleteActivity]
}

func NewEmployeeRepository(pst microservice.IPersisterMongo) *EmployeeRepository {

	insRepo := &EmployeeRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.EmployeeDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.EmployeeInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.EmployeeItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.EmployeeActivity, models.EmployeeDeleteActivity](pst)

	return insRepo
}
