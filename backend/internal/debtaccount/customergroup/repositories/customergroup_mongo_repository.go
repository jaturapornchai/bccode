package repositories

import (
	"context"
	"smlcloudplatform/internal/debtaccount/customergroup/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ICustomerGroupRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.CustomerGroupDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.CustomerGroupDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.CustomerGroupDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.CustomerGroupInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.CustomerGroupDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.CustomerGroupDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.CustomerGroupItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.CustomerGroupDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.CustomerGroupInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.CustomerGroupInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CustomerGroupDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CustomerGroupActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CustomerGroupDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CustomerGroupActivity, error)
}

type CustomerGroupRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.CustomerGroupDoc]
	repositories.SearchRepository[models.CustomerGroupInfo]
	repositories.GuidRepository[models.CustomerGroupItemGuid]
	repositories.ActivityRepository[models.CustomerGroupActivity, models.CustomerGroupDeleteActivity]
}

func NewCustomerGroupRepository(pst microservice.IPersisterMongo) *CustomerGroupRepository {

	insRepo := &CustomerGroupRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.CustomerGroupDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.CustomerGroupInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.CustomerGroupItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.CustomerGroupActivity, models.CustomerGroupDeleteActivity](pst)

	return insRepo
}
